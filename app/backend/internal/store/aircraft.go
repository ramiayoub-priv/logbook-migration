package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
)

// The aircraft write path.
//
// WHY IT EXISTS. The aircraft list used to be purely derived: rebuilt from the
// flights on every import, so the only aeroplanes that could exist were the
// ones already flown. The new-flight form's registration is a dropdown fed by
// this list, which made the first flight in an aeroplane you had never flown
// unenterable. The owner asked for CRUD on 2026-08-02.
//
// WHAT IS DELIBERATELY MISSING: a delete. Owner ruling, same day -- an
// aeroplane, once added, stays. A wrong one is corrected with an update, not
// removed, and the list is kept usable by a filterable dropdown ordered by what
// was flown most recently rather than by hiding rows. There is also no
// active/retired concept for the same reason: not flying something for a year
// does not retire it.
//
// WHY EDITING IS SAFE HERE AND IS NOT SAFE FOR A FLIGHT. Every flight carries
// its own registration, type and class denormalized, exactly as written on
// paper. This table seeds a form; it is not the record. Editing an aeroplane
// therefore cannot move one minute of the legal record, and there is a test
// that reads every flight back and asserts precisely that.

// ErrDuplicateAircraft is returned when a registration is already in the list.
// Registrations are unique: overwriting one would silently change the type or
// class of an aeroplane already in the books.
var ErrDuplicateAircraft = errors.New("store: that registration is already in the aircraft list")

// ErrAircraftNotFound is returned when no aeroplane carries that registration.
var ErrAircraftNotFound = errors.New("store: no aircraft with that registration")

// AircraftRow is one aeroplane plus what the flights say about it.
//
// LastFlown and Flights are DERIVED at query time and never stored -- the same
// reason cumulative totals are (rule 0.5). A stored "last flown" would be one
// more thing to drift the moment a flight is edited or deleted.
type AircraftRow struct {
	csvbook.Aircraft

	// UserAdded records provenance: false means it came from the paper books
	// via the importer, true means it was typed into the app.
	UserAdded bool

	// LastFlown is YYYY-MM-DD, empty if this aeroplane has no flights yet.
	LastFlown string
	Flights   int
}

// aircraftSelect reads the list with its derived columns.
//
// The ordering is the feature. A never-flown aeroplane comes FIRST: the only
// reason to have added one is that it is about to be flown. Everything else
// runs most-recently-flown to least, because the aeroplane you flew last week
// is the one you are most likely to be logging now. This is what replaces the
// retired/active idea the owner dropped.
const aircraftSelect = `
	SELECT a.registration, a.type, a.default_class, a.ifr_capable, a.notes,
	       a.user_added,
	       COALESCE(MAX(f.flight_date), '') AS last_flown,
	       COUNT(f.id)                      AS flights
	  FROM aircraft a
	  LEFT JOIN flights f ON f.aircraft_reg = a.registration
	 %s
	 GROUP BY a.id
	 ORDER BY (last_flown = '') DESC, last_flown DESC, a.registration`

func scanAircraft(rows *sql.Rows) ([]AircraftRow, error) {
	defer rows.Close()
	var out []AircraftRow
	for rows.Next() {
		var (
			a         AircraftRow
			class     string
			ifr, user int
		)
		if err := rows.Scan(&a.Registration, &a.Type, &class, &ifr, &a.Notes,
			&user, &a.LastFlown, &a.Flights); err != nil {
			return nil, fmt.Errorf("store: reading aircraft: %w", err)
		}
		a.DefaultClass = csvbook.Class(class)
		a.IFRCapable = ifr == 1
		a.UserAdded = user == 1
		out = append(out, a)
	}
	return out, rows.Err()
}

// AircraftList returns every aeroplane, ordered for the form's dropdown.
func (db *DB) AircraftList() ([]AircraftRow, error) {
	rows, err := db.sql.Query(fmt.Sprintf(aircraftSelect, ""))
	if err != nil {
		return nil, fmt.Errorf("store: reading aircraft: %w", err)
	}
	return scanAircraft(rows)
}

// AircraftByReg returns one aeroplane.
func (db *DB) AircraftByReg(reg string) (AircraftRow, error) {
	rows, err := db.sql.Query(
		fmt.Sprintf(aircraftSelect, "WHERE a.registration = ?"), normalizeReg(reg))
	if err != nil {
		return AircraftRow{}, fmt.Errorf("store: reading aircraft: %w", err)
	}
	list, err := scanAircraft(rows)
	if err != nil {
		return AircraftRow{}, err
	}
	if len(list) == 0 {
		return AircraftRow{}, ErrAircraftNotFound
	}
	return list[0], nil
}

// AddAircraft writes a new aeroplane and marks it as user-added.
//
// The mark is load-bearing: the importer's DELETE is scoped by it, so a
// hand-added aeroplane survives an import exactly as a hand-entered flight does.
func (db *DB) AddAircraft(a csvbook.Aircraft) (AircraftRow, error) {
	a = normalizeAircraft(a)
	if err := validateAircraft(a); err != nil {
		return AircraftRow{}, err
	}

	_, err := db.sql.Exec(
		`INSERT INTO aircraft (registration, type, default_class, ifr_capable, notes, user_added)
		 VALUES (?, ?, ?, ?, ?, 1)`,
		a.Registration, a.Type, string(a.DefaultClass), boolInt(a.IFRCapable), a.Notes)
	if err != nil {
		if isUniqueViolation(err) {
			return AircraftRow{}, ErrDuplicateAircraft
		}
		return AircraftRow{}, fmt.Errorf("store: adding aircraft %s: %w", a.Registration, err)
	}
	return db.AircraftByReg(a.Registration)
}

// UpdateAircraft replaces the details of the aeroplane currently registered as
// reg. A full replacement, like UpdateFlight: a partial update is a merge, and
// a merge is where a field silently keeps a value nobody meant to keep.
//
// It does NOT rewrite the flights that reference this aeroplane, and that is
// deliberate. They record what was written on paper. If the two now disagree,
// that is a discrepancy to surface, never to auto-correct (rule 0.2) -- and it
// is exactly the shape of the owner's own SE-GKT -> OH-GKT case, one airframe
// whose registration changed.
func (db *DB) UpdateAircraft(reg string, a csvbook.Aircraft) (AircraftRow, error) {
	reg = normalizeReg(reg)
	a = normalizeAircraft(a)
	if err := validateAircraft(a); err != nil {
		return AircraftRow{}, err
	}

	res, err := db.sql.Exec(
		`UPDATE aircraft
		    SET registration = ?, type = ?, default_class = ?, ifr_capable = ?, notes = ?
		  WHERE registration = ?`,
		a.Registration, a.Type, string(a.DefaultClass), boolInt(a.IFRCapable), a.Notes, reg)
	if err != nil {
		if isUniqueViolation(err) {
			return AircraftRow{}, ErrDuplicateAircraft
		}
		return AircraftRow{}, fmt.Errorf("store: updating aircraft %s: %w", reg, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return AircraftRow{}, fmt.Errorf("store: updating aircraft %s: %w", reg, err)
	}
	if n == 0 {
		return AircraftRow{}, ErrAircraftNotFound
	}
	return db.AircraftByReg(a.Registration)
}

// normalizeReg is how a registration is stored and compared: upper case, no
// surrounding space. One representation, so "oh-pdp" and "OH-PDP " cannot
// become two aeroplanes.
func normalizeReg(reg string) string {
	return strings.ToUpper(strings.TrimSpace(reg))
}

func normalizeAircraft(a csvbook.Aircraft) csvbook.Aircraft {
	a.Registration = normalizeReg(a.Registration)
	a.Type = strings.ToUpper(strings.TrimSpace(a.Type))
	a.DefaultClass = csvbook.Class(strings.ToUpper(strings.TrimSpace(string(a.DefaultClass))))
	a.Notes = strings.TrimSpace(a.Notes)
	return a
}

func validateAircraft(a csvbook.Aircraft) error {
	if a.Registration == "" {
		return errors.New("store: a registration is required")
	}
	if a.Type == "" {
		return errors.New("store: an aircraft type is required")
	}
	if !csvbook.ValidClass(a.DefaultClass) {
		return fmt.Errorf("store: %q is not a known class", a.DefaultClass)
	}
	return nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// isUniqueViolation reports whether err is SQLite's UNIQUE constraint failure.
// Matched on the message because modernc.org/sqlite does not export a typed
// constraint error, and the alternative -- SELECT then INSERT -- has a race
// between the two statements.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
