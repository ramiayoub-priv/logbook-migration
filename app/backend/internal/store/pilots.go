package store

import (
	"errors"
	"fmt"
	"strings"
)

// The pilot roster: the names that may go in a flight's "name of pilot in
// command" field.
//
// WHY IT EXISTS. The field was free text, and 1143 of the 1296 transcribed
// flights carry the single word `self`. The owner asked on 2026-08-03 for the
// same treatment the aircraft registrations got: "I could have a typo when I
// write `self`, it could be `sself` or `SELF` or `seeelf` and I need it to be
// consistent." One stray keystroke would split a licence figure's provenance
// across two spellings of the same word, and nothing would ever flag it.
//
// IT IS MOSTLY DERIVED, like the aircraft list was before it gained a write
// path. Every distinct pic_name already in the record is in the roster with its
// count and the date it was last flown with; the `pilots` table holds only the
// names that have not been used yet -- the ones added because a flight with
// somebody new is about to be logged.
//
// WHAT IT DELIBERATELY DOES NOT DO IS TIDY THE PAPER (rule 0.8). The books
// contain `Sinervä` six times and `Sinerva` once, and `Stude` eighteen times,
// which reads as a word rather than a surname. Both are surfaced to the owner
// and neither is merged, renamed or hidden here. This roster stops a NEW
// spelling being invented; it makes no claim about the ones that exist.
//
// THERE IS NO UPDATE AND NO DELETE, and that is a smaller surface than the
// aircraft got on purpose. Renaming a roster entry could not rename the flights
// that use it -- they are the record -- so the two would simply disagree, and
// the derived half would go on reporting the old spelling anyway. A name typed
// wrongly is corrected where it actually lives: on the flight, through the
// ordinary edit path, after which the wrong spelling has no flights and sorts
// to the top of the list as never-used. Recorded in APP.md as a deliberate
// limitation rather than an oversight.

// ErrDuplicatePilot is returned when the roster already knows that name --
// including under a different case, which is the whole point of the check.
var ErrDuplicatePilot = errors.New("store: that name is already in the pilot list")

// PilotRow is one name plus what the flights say about it.
//
// Flights and LastFlown are DERIVED at query time and never stored, for the
// same reason cumulative totals are (rule 0.5): a stored count is one more
// thing to drift the moment a flight is edited or deleted.
type PilotRow struct {
	Name string

	// UserAdded is true for a name that exists only in the pilots table, i.e.
	// one added in the app and not yet flown with. It is provenance for the
	// reader, not a permission: nothing treats the two kinds differently.
	UserAdded bool

	// LastFlown is YYYY-MM-DD, empty if no flight names this pilot yet.
	LastFlown string
	Flights   int
}

// pilotSelect unions the two halves of the roster.
//
// The ordering is the aircraft ordering, for the same reason: a name never used
// comes FIRST, because the only reason to have added one is that it is about to
// be flown with, and everything else runs most-recent to least. It lives in SQL
// so there is one authority for it -- the frontend deliberately does not
// re-sort what it is sent.
//
// Names are grouped EXACTLY, not case-insensitively. If two spellings are both
// in the record, both appear: merging them here would be this code deciding
// that `Sinerva` is a typo for `Sinervä`, which is the owner's call and nobody
// else's (rule 0.2).
const pilotSelect = `
	SELECT name,
	       MAX(last_flown) AS last_flown,
	       SUM(flights)    AS flights,
	       MIN(user_added) AS user_added
	  FROM (
	        SELECT p.name AS name, '' AS last_flown, 0 AS flights, 1 AS user_added
	          FROM pilots p
	        UNION ALL
	        SELECT f.pic_name, MAX(f.flight_date), COUNT(f.id), 0
	          FROM flights f
	         WHERE TRIM(f.pic_name) <> ''
	         GROUP BY f.pic_name
	       )
	 GROUP BY name
	 ORDER BY (last_flown = '') DESC, last_flown DESC, name`

// PilotList returns every name the form may offer, ordered for its picker.
func (db *DB) PilotList() ([]PilotRow, error) {
	rows, err := db.sql.Query(pilotSelect)
	if err != nil {
		return nil, fmt.Errorf("store: reading pilots: %w", err)
	}
	defer rows.Close()

	var out []PilotRow
	for rows.Next() {
		var (
			p    PilotRow
			user int
		)
		if err := rows.Scan(&p.Name, &p.LastFlown, &p.Flights, &user); err != nil {
			return nil, fmt.Errorf("store: reading pilots: %w", err)
		}
		p.UserAdded = user == 1
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: reading pilots: %w", err)
	}
	return out, nil
}

// AddPilot puts a name on the roster.
//
// It is refused if the roster already knows the name in any case -- against the
// `pilots` table by the UNIQUE ... COLLATE NOCASE index, and against the names
// already on flights by the query below. Both halves are needed: `self` exists
// only on flights, and it is the exact name the owner was worried about.
//
// ⚠ The case-insensitivity is SQLite's NOCASE, which folds ASCII only. Adding
// `SINERVÄ` alongside `Sinervä` would therefore be allowed. Not worth a
// dependency to fix for a roster of eighteen names that one person types, but
// it is a real limit and it is written down rather than assumed away.
func (db *DB) AddPilot(name string) (PilotRow, error) {
	name = normalizeName(name)
	if name == "" {
		return PilotRow{}, errors.New("store: a name is required")
	}

	// The flights half, which no constraint can cover: pic_name is written on
	// the record as it was on paper and is not unique there.
	var used int
	if err := db.sql.QueryRow(
		`SELECT COUNT(*) FROM flights WHERE pic_name = ? COLLATE NOCASE`, name).Scan(&used); err != nil {
		return PilotRow{}, fmt.Errorf("store: adding pilot %q: %w", name, err)
	}
	if used > 0 {
		return PilotRow{}, ErrDuplicatePilot
	}

	if _, err := db.sql.Exec(`INSERT INTO pilots (name) VALUES (?)`, name); err != nil {
		if isUniqueViolation(err) {
			return PilotRow{}, ErrDuplicatePilot
		}
		return PilotRow{}, fmt.Errorf("store: adding pilot %q: %w", name, err)
	}
	return db.PilotByName(name)
}

// PilotByName returns the roster entry spelled EXACTLY like this.
//
// Exact, not case-insensitive, and the difference is the whole feature. AddPilot
// compares case-insensitively so that `SELF` cannot join `self` as a second
// entry; this asks the opposite question -- "is this precise spelling one the
// roster offers?" -- and it is what the flight write path checks before putting
// a name on the record. A case-insensitive answer here would cheerfully accept
// the `SELF` the roster was built to prevent.
func (db *DB) PilotByName(name string) (PilotRow, error) {
	list, err := db.PilotList()
	if err != nil {
		return PilotRow{}, err
	}
	want := normalizeName(name)
	for _, p := range list {
		if p.Name == want {
			return p, nil
		}
	}
	return PilotRow{}, ErrPilotNotFound
}

// ErrPilotNotFound is returned when no roster entry carries that name.
var ErrPilotNotFound = errors.New("store: no pilot with that name")

// normalizeName trims and does nothing else.
//
// A NAME IS NOT A REGISTRATION. Registrations are upper-cased because that is
// the one true form of an identifier; names are not, and `self` is lower case
// on all 1143 rows that carry it while `Martevuo` is capitalised. Re-casing
// would invent a spelling the books do not use, which is precisely the thing
// this file exists to prevent.
func normalizeName(s string) string { return strings.TrimSpace(s) }
