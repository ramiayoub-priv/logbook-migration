package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/entry"
)

// HandEnteredSeqBase is where the seq numbers of app-entered flights start.
//
// seq is the explicit book order and the key every cumulative computation
// walks, and the importer reassigns it 1..N over the CSVs on every single run.
// A hand-entered row therefore cannot share that range: Book 3 is still being
// transcribed page by page, so any seq below N is a collision waiting for the
// migration to catch up to it.
//
// Two disjoint bands solve it outright. Imported rows own 1..N and are
// renumbered freely; hand-entered rows own 1000000.. and keep their seq for
// life. The books would have to grow by three orders of magnitude to meet in
// the middle, which is 1293 flights over 14 years away from happening.
//
// The bands also give the right sort order for free: a flight flown today
// belongs after every page of the paper books, and that is exactly where a
// higher seq puts it.
const HandEnteredSeqBase = 1_000_000

// ErrDuplicateFlight is returned when a flight matching one already stored is
// submitted again -- the double-tapped submit button on a phone. The logbook
// is a legal record and the same flight appearing twice inflates a licence
// total, so the second write is refused rather than merged.
var ErrDuplicateFlight = errors.New("store: this flight is already in the logbook")

// AddFlight writes one hand-entered flight and returns it with the seq and
// source_row the database assigned.
//
// The flight must already have been validated by internal/entry; this function
// is about allocation and storage, not about whether the numbers make sense.
// It runs in a transaction because the seq allocation is a read followed by a
// write, and two submits racing on that would otherwise both read the same
// maximum and collide on the UNIQUE constraint.
func (db *DB) AddFlight(f csvbook.Flight) (csvbook.Flight, error) {
	if f.SourceBook != entry.HandEnteredBook {
		// A caller reaching this with a real book number is trying to insert a
		// row the importer believes it owns, which the next import would
		// delete. Refuse rather than write a row with a lifespan.
		return csvbook.Flight{}, fmt.Errorf(
			"store: AddFlight is for hand-entered rows only, got source_book %d", f.SourceBook)
	}

	tx, err := db.sql.Begin()
	if err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: starting the flight insert: %w", err)
	}
	defer tx.Rollback()

	dup, err := hasDuplicate(tx, f)
	if err != nil {
		return csvbook.Flight{}, err
	}
	if dup {
		return csvbook.Flight{}, ErrDuplicateFlight
	}

	// Both allocations come from the hand-entered band only, so neither can be
	// disturbed by the importer renumbering its own rows.
	var maxSeq, maxRow sql.NullInt64
	if err := tx.QueryRow(
		`SELECT MAX(seq), MAX(source_row) FROM flights WHERE source_book = ?`,
		entry.HandEnteredBook,
	).Scan(&maxSeq, &maxRow); err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: allocating a seq: %w", err)
	}

	f.Seq = HandEnteredSeqBase
	if maxSeq.Valid && maxSeq.Int64 >= HandEnteredSeqBase {
		f.Seq = int(maxSeq.Int64) + 1
	}
	f.SourceRow = int(maxRow.Int64) + 1

	// The link is looked up rather than required: an aircraft flown for the
	// first time must not cost the pilot the flight, exactly as the importer
	// treats a registration its seed list does not know.
	var link any
	var id int64
	switch err := tx.QueryRow(
		`SELECT id FROM aircraft WHERE registration = ?`, f.AircraftReg,
	).Scan(&id); {
	case err == nil:
		link = id
	case errors.Is(err, sql.ErrNoRows):
		// Left NULL on purpose.
	default:
		return csvbook.Flight{}, fmt.Errorf("store: looking up aircraft %s: %w", f.AircraftReg, err)
	}

	if _, err := tx.Exec(
		`INSERT INTO flights (
		    seq, flight_date, aircraft_id, aircraft_reg, aircraft_type, class,
		    dep_place, arr_place,
		    off_block_utc, on_block_utc, off_block_raw, on_block_raw, time_origin,
		    takeoff_utc, landing_utc,
		    block_minutes, total_minutes, night_minutes, instrument_minutes,
		    pic_minutes, dual_minutes, instructor_minutes,
		    copilot_minutes, multipilot_minutes,
		    pic_name, landings_day, landings_night, landings_verified, remarks,
		    source_book, source_row
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.Seq, f.Date, link, f.AircraftReg, f.AircraftType, string(f.Class),
		f.DepPlace, f.ArrPlace,
		nullTime(f.OffBlockUTC), nullTime(f.OnBlockUTC), f.OffBlockRaw, f.OnBlockRaw, string(f.TimeOrigin),
		nullTime(f.TakeoffUTC), nullTime(f.LandingUTC),
		f.BlockMinutes, f.TotalMinutes, f.NightMinutes, f.InstrumentMinutes,
		f.PICMinutes, f.DualMinutes, f.InstructorMinutes,
		f.CopilotMinutes, f.MultiPilotMinutes,
		f.PICName, f.LandingsDay, f.LandingsNight, boolToInt(f.LandingsVerified), f.Remarks,
		f.SourceBook, f.SourceRow,
	); err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: inserting the hand-entered flight: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: committing the hand-entered flight: %w", err)
	}
	return f, nil
}

// relinkHandEntered restores the aircraft link on rows the importer does not
// own.
//
// The import replaces the aircraft table wholesale, and the flights table
// declares ON DELETE SET NULL, so every hand-entered row that referenced an
// aircraft has its aircraft_id nulled the moment the old rows go -- silently,
// as a side effect of a run that was otherwise leaving those flights alone.
// The registration is denormalized on the flight so nothing is actually lost,
// but the link would never come back on its own, and it would degrade a
// little more with every page the migration transcribes.
//
// Matching on the registration is the same rule insertFlights uses, so an
// aircraft that has since disappeared from the seed list simply stays
// unlinked rather than failing the import.
func relinkHandEntered(tx *sql.Tx) error {
	if _, err := tx.Exec(
		`UPDATE flights
		    SET aircraft_id = (SELECT id FROM aircraft WHERE aircraft.registration = flights.aircraft_reg)
		  WHERE source_book = ?`, entry.HandEnteredBook,
	); err != nil {
		return fmt.Errorf("store: relinking hand-entered flights to the aircraft list: %w", err)
	}
	return nil
}

// hasDuplicate reports whether this flight is already stored.
//
// The key is the date, the aircraft and the off-block time as written. That is
// what distinguishes two flights: the same aeroplane cannot leave the blocks
// twice at the same minute on the same day, while two genuinely different
// flights on one day differ in their off-block. Matching on the raw string
// rather than the converted instant means a resubmission of the identical form
// is caught even if the conversion were ever to change underneath it.
func hasDuplicate(q querier, f csvbook.Flight) (bool, error) {
	var n int
	if err := q.QueryRow(
		`SELECT COUNT(*) FROM flights
		 WHERE flight_date = ? AND aircraft_reg = ? AND off_block_raw = ?`,
		f.Date, f.AircraftReg, f.OffBlockRaw,
	).Scan(&n); err != nil {
		return false, fmt.Errorf("store: checking for a duplicate flight: %w", err)
	}
	return n > 0, nil
}
