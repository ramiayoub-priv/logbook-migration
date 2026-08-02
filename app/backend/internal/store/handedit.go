package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/entry"
)

// Editing and deleting a flight.
//
// Added 2026-08-02, when the transcription effort closed and the application
// became the only way the record grows (CLAUDE.md rule 0.8). Until then a
// mistyped flight could only be corrected by opening SQLite by hand.
//
// Two rules govern everything in this file, and both are enforced here rather
// than in a handler, so that no HTTP route -- present or future -- can get
// round them:
//
//  1. ONLY HAND-ENTERED ROWS. A row from the paper books is closed data. It is
//     also owned by the importer, which replaces every row with
//     source_book <> 0 on each run, so an edit to one would be discarded
//     without warning at the next re-import. Refusing is both the honest answer
//     and the only durable one.
//
//  2. NOTHING CHANGES WITHOUT AN AUDIT ROW. The owner asked for a standard
//     in-place edit, which is the right shape for the form; the safety it lacks
//     on a legal record is added underneath, where it costs the reader nothing.
//     The audit write is in the same transaction as the change, so a change
//     cannot commit without it.

// ErrFlightNotFound means no flight carries that seq.
var ErrFlightNotFound = errors.New("store: no flight with that number")

// ErrNotHandEntered means the flight came from a paper book. It is returned
// rather than silently ignored: a caller has asked for something the record
// does not allow, and saying so is the whole point.
var ErrNotHandEntered = errors.New(
	"store: that flight was transcribed from a paper logbook and cannot be changed here")

// AuditEntry is one recorded change to a flight.
type AuditEntry struct {
	At     string
	UserID int64
	Action string
	Seq    int
	// Before is the complete row as it stood before the change, as JSON.
	Before string
}

// UpdateFlight replaces a hand-entered flight in place, keeping its seq.
//
// The flight must already have been validated by internal/entry -- the same
// validation a new flight goes through, because an edited flight is subject to
// exactly the same rules about what may be written. What this function decides
// is whether the row may be touched at all, and that the change is recorded.
//
// seq is not editable. It is the book's own order, the key every cumulative
// computation walks, and it is not something a pilot corrects by editing a
// flight; source_book and source_row are likewise preserved.
func (db *DB) UpdateFlight(seq int, f csvbook.Flight, userID int64) (csvbook.Flight, error) {
	tx, err := db.sql.Begin()
	if err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: starting the flight edit: %w", err)
	}
	defer tx.Rollback()

	before, err := handEnteredForChange(tx, seq)
	if err != nil {
		return csvbook.Flight{}, err
	}

	// The same key AddFlight refuses on, minus this row: two identical flights
	// inflate a licence total whether they arrive by insert or by edit, but a
	// row is always a duplicate of itself and saving one unchanged has to work.
	var clashes int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM flights
		  WHERE flight_date = ? AND aircraft_reg = ? AND off_block_raw = ? AND seq <> ?`,
		f.Date, f.AircraftReg, f.OffBlockRaw, seq,
	).Scan(&clashes); err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: checking for a duplicate flight: %w", err)
	}
	if clashes > 0 {
		return csvbook.Flight{}, ErrDuplicateFlight
	}

	if err := db.audit(tx, "update", seq, userID, before); err != nil {
		return csvbook.Flight{}, err
	}

	// Looked up again rather than carried over: an edit can change the
	// registration, and a stale aircraft_id would point the flight at the
	// aeroplane it used to be. Absent is fine -- an unknown registration
	// leaves the link NULL, exactly as AddFlight and the importer do.
	link, err := aircraftLink(tx, f.AircraftReg)
	if err != nil {
		return csvbook.Flight{}, err
	}

	if _, err := tx.Exec(
		`UPDATE flights SET
		    flight_date = ?, aircraft_id = ?, aircraft_reg = ?, aircraft_type = ?, class = ?,
		    dep_place = ?, arr_place = ?,
		    off_block_utc = ?, on_block_utc = ?, off_block_raw = ?, on_block_raw = ?,
		    time_origin = ?, takeoff_utc = ?, landing_utc = ?,
		    block_minutes = ?, total_minutes = ?, night_minutes = ?, instrument_minutes = ?,
		    pic_minutes = ?, dual_minutes = ?, instructor_minutes = ?,
		    copilot_minutes = ?, multipilot_minutes = ?,
		    pic_name = ?, landings_day = ?, landings_night = ?, landings_verified = ?, remarks = ?
		  WHERE seq = ?`,
		f.Date, link, f.AircraftReg, f.AircraftType, string(f.Class),
		f.DepPlace, f.ArrPlace,
		nullTime(f.OffBlockUTC), nullTime(f.OnBlockUTC), f.OffBlockRaw, f.OnBlockRaw,
		string(f.TimeOrigin), nullTime(f.TakeoffUTC), nullTime(f.LandingUTC),
		f.BlockMinutes, f.TotalMinutes, f.NightMinutes, f.InstrumentMinutes,
		f.PICMinutes, f.DualMinutes, f.InstructorMinutes,
		f.CopilotMinutes, f.MultiPilotMinutes,
		f.PICName, f.LandingsDay, f.LandingsNight, boolToInt(f.LandingsVerified), f.Remarks,
		seq,
	); err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: updating flight %d: %w", seq, err)
	}

	if err := tx.Commit(); err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: committing the flight edit: %w", err)
	}

	// The identity fields are the stored row's, not the caller's: a caller
	// cannot move a flight in the book by sending a different seq.
	f.Seq, f.SourceBook, f.SourceRow = seq, before.SourceBook, before.SourceRow
	return f, nil
}

// DeleteFlight removes a hand-entered flight and returns what was removed.
//
// The row goes, and every total follows immediately, because totals are
// computed and never stored (rule 0.5). Its complete contents survive in
// flight_audit, which is what makes a mistaken delete recoverable -- the owner
// chose "remove it, keep an audit copy" over a soft delete precisely so that
// nothing lingers in the logbook itself.
func (db *DB) DeleteFlight(seq int, userID int64) (csvbook.Flight, error) {
	tx, err := db.sql.Begin()
	if err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: starting the flight delete: %w", err)
	}
	defer tx.Rollback()

	before, err := handEnteredForChange(tx, seq)
	if err != nil {
		return csvbook.Flight{}, err
	}
	if err := db.audit(tx, "delete", seq, userID, before); err != nil {
		return csvbook.Flight{}, err
	}
	if _, err := tx.Exec(`DELETE FROM flights WHERE seq = ?`, seq); err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: deleting flight %d: %w", seq, err)
	}
	if err := tx.Commit(); err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: committing the flight delete: %w", err)
	}
	return before, nil
}

// FlightBySeq returns one flight, hand-entered or not.
func (db *DB) FlightBySeq(seq int) (csvbook.Flight, error) {
	f, err := flightBySeq(db.sql, seq)
	if err != nil {
		return csvbook.Flight{}, err
	}
	return f, nil
}

// FlightAudit returns every recorded change to one flight, oldest first.
//
// Nothing in the application calls this; it exists so the question "what did
// this row say before?" has an answer, and so the tests can assert that the
// answer is being written.
func (db *DB) FlightAudit(seq int) ([]AuditEntry, error) {
	rows, err := db.sql.Query(
		`SELECT at, user_id, action, seq, before FROM flight_audit WHERE seq = ? ORDER BY id`, seq)
	if err != nil {
		return nil, fmt.Errorf("store: reading the flight audit: %w", err)
	}
	defer rows.Close()

	var out []AuditEntry
	for rows.Next() {
		var a AuditEntry
		if err := rows.Scan(&a.At, &a.UserID, &a.Action, &a.Seq, &a.Before); err != nil {
			return nil, fmt.Errorf("store: reading the flight audit: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// FlightAircraftLinked reports whether a flight resolves to a row in the
// aircraft table. Used by the tests that pin the relink behaviour.
func (db *DB) FlightAircraftLinked(seq int) (bool, error) {
	var linked int
	err := db.sql.QueryRow(
		`SELECT aircraft_id IS NOT NULL FROM flights WHERE seq = ?`, seq).Scan(&linked)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrFlightNotFound
	}
	if err != nil {
		return false, fmt.Errorf("store: checking the aircraft link of flight %d: %w", seq, err)
	}
	return linked == 1, nil
}

// handEnteredForChange loads the row a change is about and applies the one
// rule both paths share.
func handEnteredForChange(q querier, seq int) (csvbook.Flight, error) {
	f, err := flightBySeq(q, seq)
	if err != nil {
		return csvbook.Flight{}, err
	}
	if f.SourceBook != entry.HandEnteredBook {
		return csvbook.Flight{}, fmt.Errorf("%w (book %d, row %d)",
			ErrNotHandEntered, f.SourceBook, f.SourceRow)
	}
	return f, nil
}

// audit records the state a row was in before a change, in the caller's
// transaction, so the two commit or roll back together.
func (db *DB) audit(tx *sql.Tx, action string, seq int, userID int64, before csvbook.Flight) error {
	blob, err := json.Marshal(auditRow(before))
	if err != nil {
		return fmt.Errorf("store: recording the audit copy of flight %d: %w", seq, err)
	}
	if _, err := tx.Exec(
		`INSERT INTO flight_audit (at, user_id, action, seq, before) VALUES (?, ?, ?, ?, ?)`,
		db.clock().UTC().Format(timeFormat), userID, action, seq, string(blob),
	); err != nil {
		return fmt.Errorf("store: recording the audit copy of flight %d: %w", seq, err)
	}
	return nil
}

// auditRow is the flight as the audit table stores it.
//
// Written out explicitly rather than marshalling csvbook.Flight directly: this
// JSON has to stay readable years from now, by someone recovering a deleted
// flight, and it must not change shape because a domain struct was refactored.
func auditRow(f csvbook.Flight) map[string]any {
	instant := func(t time.Time) any {
		if t.IsZero() {
			return nil
		}
		return t.UTC().Format(timeFormat)
	}
	return map[string]any{
		"seq":                f.Seq,
		"date":               f.Date,
		"aircraft_reg":       f.AircraftReg,
		"aircraft_type":      f.AircraftType,
		"class":              string(f.Class),
		"dep_place":          f.DepPlace,
		"arr_place":          f.ArrPlace,
		"off_block_utc":      instant(f.OffBlockUTC),
		"on_block_utc":       instant(f.OnBlockUTC),
		"off_block_raw":      f.OffBlockRaw,
		"on_block_raw":       f.OnBlockRaw,
		"time_origin":        string(f.TimeOrigin),
		"takeoff_utc":        instant(f.TakeoffUTC),
		"landing_utc":        instant(f.LandingUTC),
		"block_minutes":      f.BlockMinutes,
		"total_minutes":      f.TotalMinutes,
		"night_minutes":      f.NightMinutes,
		"instrument_minutes": f.InstrumentMinutes,
		"pic_minutes":        f.PICMinutes,
		"dual_minutes":       f.DualMinutes,
		"instructor_minutes": f.InstructorMinutes,
		"copilot_minutes":    f.CopilotMinutes,
		"multipilot_minutes": f.MultiPilotMinutes,
		"pic_name":           f.PICName,
		"landings_day":       f.LandingsDay,
		"landings_night":     f.LandingsNight,
		"landings_verified":  f.LandingsVerified,
		"remarks":            f.Remarks,
		"source_book":        f.SourceBook,
		"source_row":         f.SourceRow,
	}
}

// aircraftLink resolves a registration to an aircraft id, or NULL.
func aircraftLink(q querier, reg string) (any, error) {
	var id int64
	switch err := q.QueryRow(`SELECT id FROM aircraft WHERE registration = ?`, reg).Scan(&id); {
	case err == nil:
		return id, nil
	case errors.Is(err, sql.ErrNoRows):
		// A first flight in a new aeroplane must not cost the pilot the edit.
		return nil, nil
	default:
		return nil, fmt.Errorf("store: looking up aircraft %s: %w", reg, err)
	}
}
