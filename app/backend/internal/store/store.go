// Package store owns the SQLite database: the schema, the import, and the
// verification that makes the import trustworthy.
//
// The import is a full replace inside one transaction. That is what delivers
// the three properties CLAUDE.md rule 0.2 demands of it:
//
//   - idempotent -- re-running the same CSVs yields an identical database, so
//     there is never a question of whether it has been run already;
//   - reversible -- the caller takes a backup first (see Backup) and the
//     transaction rolls back untouched if anything fails;
//   - verified -- after writing, every figure is read back out of SQLite and
//     compared against the checksums the CSVs produced. A mismatch of one
//     minute aborts the whole thing.
//
// The verification deliberately reads back rather than trusting what was
// written: a CHECK constraint, a type conversion or a silently truncated value
// would otherwise pass unnoticed, and this is a legal record.
package store

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/entry"
	"github.com/ramiayoub/logbook/backend/internal/timeutil"

	// Pure-Go SQLite. No CGO, which is what makes the server a single static
	// binary that cross-compiles from anywhere (app/APP.md, Stack).
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// schemaVersion is bumped when schema.sql changes in a way that needs a
// migration. Version 1 is the initial schema; version 2 adds flight_audit
// (2026-08-02), which needs no migration -- it is a new, empty table created
// by an idempotent statement, so an existing database gains it on the next
// open and loses nothing.
const schemaVersion = 2

// timeFormat is how instants are stored: RFC3339 in UTC, second precision.
// Lexical order equals chronological order, which every range query relies on.
const timeFormat = "2006-01-02T15:04:05Z"

// DB is an open logbook database.
type DB struct {
	sql  *sql.DB
	path string
	// backup is the path of the most recent backup, recorded on the next
	// import run so a figure can be traced to the state it replaced.
	backup string
	// clock is the source of "now" for session expiry. It is a field rather
	// than a call to time.Now so that the 90-day rolling window can be tested
	// in milliseconds instead of being taken on trust -- see SetClockForTest
	// and the expiry control in app/docs/security.md.
	clock func() time.Time
}

// at is the current instant in UTC, from this database's clock.
func (db *DB) at() time.Time { return db.clock().UTC().Truncate(time.Second) }

// stamp formats the current instant for a TEXT column.
func (db *DB) stamp() string { return db.at().Format(timeFormat) }

// SetClockForTest replaces the clock and returns a function restoring it.
// Test-only; nothing in the server calls it.
func (db *DB) SetClockForTest(c func() time.Time) (restore func()) {
	prev := db.clock
	db.clock = c
	return func() { db.clock = prev }
}

// SQLForTest exposes the underlying handle so tests can assert on the bytes
// actually written -- that the sessions table holds no usable token, that a
// password hash is argon2id. Test-only; production code goes through the
// methods on DB.
func (db *DB) SQLForTest() *sql.DB { return db.sql }

// Open opens (or creates) the database at path and brings the schema up.
//
// Opening an existing database is a no-op: every statement in schema.sql is
// idempotent, so this is safe to call on a live production file.
func Open(path string) (*DB, error) {
	// busy_timeout keeps a concurrent reader from turning into an instant
	// SQLITE_BUSY; foreign_keys is off by default in SQLite and must be asked
	// for on every connection.
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: opening %s: %w", path, err)
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("store: opening %s: %w", path, err)
	}
	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("store: applying schema to %s: %w", path, err)
	}
	if err := migrate(sqlDB); err != nil {
		sqlDB.Close()
		return nil, err
	}
	if _, err := sqlDB.Exec(
		`INSERT INTO schema_version (version, applied_at)
		 SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM schema_version WHERE version = ?)`,
		schemaVersion, now(), schemaVersion,
	); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("store: recording schema version: %w", err)
	}
	return &DB{sql: sqlDB, path: path, clock: time.Now}, nil
}

// migrate applies the schema changes that cannot be written idempotently in
// SQL. SQLite has no ADD COLUMN IF NOT EXISTS, and a bare ALTER would fail on
// the second Open -- so the column is added only if PRAGMA table_info says it
// is absent.
//
// EVERY MIGRATION HERE MUST BE ADDITIVE AND SAFE ON A LIVE PRODUCTION FILE.
// Open() runs on the real logbook at every service start (CLAUDE.md rule 0.2);
// this is not a place for anything that rewrites or drops.
func migrate(sqlDB *sql.DB) error {
	for _, m := range []struct{ table, column, spec string }{
		// Which rows the importer owns. Without it, `DELETE FROM aircraft`
		// takes every hand-added aeroplane with it -- the same trap that
		// source_book = 0 solves for flights.
		{"aircraft", "user_added", "INTEGER NOT NULL DEFAULT 0 CHECK (user_added IN (0,1))"},
	} {
		has, err := hasColumn(sqlDB, m.table, m.column)
		if err != nil {
			return err
		}
		if has {
			continue
		}
		if _, err := sqlDB.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
			m.table, m.column, m.spec)); err != nil {
			return fmt.Errorf("store: adding %s.%s: %w", m.table, m.column, err)
		}
	}
	return nil
}

func hasColumn(sqlDB *sql.DB, table, column string) (bool, error) {
	rows, err := sqlDB.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("store: reading %s columns: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid, notnull, pk int
			name, ctype      string
			dflt             any
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, fmt.Errorf("store: reading %s columns: %w", table, err)
		}
		if name == column {
			return true, rows.Err()
		}
	}
	return false, rows.Err()
}

// Close releases the database.
func (db *DB) Close() error { return db.sql.Close() }

// Path is where the database lives on disk.
func (db *DB) Path() string { return db.path }

// Backup writes a consistent copy of the database to dest.
//
// VACUUM INTO rather than a file copy: it is transactionally consistent even
// while the database is open and in WAL mode, where the .db file on its own is
// not the whole story.
//
// It refuses to overwrite. Replacing an existing backup would destroy the only
// copy of the state someone was trying to preserve, and the caller can always
// pick another name.
func (db *DB) Backup(dest string) error {
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("store: backup %s already exists; refusing to overwrite it", dest)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("store: checking backup path %s: %w", dest, err)
	}
	if _, err := db.sql.Exec("VACUUM INTO ?", dest); err != nil {
		return fmt.Errorf("store: backing up to %s: %w", dest, err)
	}
	return nil
}

// NoteBackup records which backup protects the next import.
func (db *DB) NoteBackup(path string) { db.backup = path }

// RedactForBackup strips what a copy of this database must not carry off the
// box, and reports how many rows went.
//
// CALL THIS ON A COPY, NEVER ON THE LIVE DATABASE. On the live file it would
// sign every device out. It exists for internal/backup, which takes a
// VACUUM INTO snapshot and then redacts that.
//
// Only `sessions` goes. They are the closest thing in this file to a live
// credential -- the column is a hash of the cookie rather than the cookie
// itself (docs/security.md), so a leaked backup yields nothing usable, but a
// backup that leaves the machine should still carry the smallest useful set.
// They are also worthless on restore: the expiry has passed, the addresses are
// stale, and the owner signs in again regardless.
//
// `users` deliberately SURVIVES. A restored logbook nobody can log into is not
// a restored logbook, and the password is an Argon2id hash at 19 MiB, which is
// what makes shipping it to a private repository an acceptable trade for being
// able to come back from a dead server. Recorded in docs/security.md.
//
// VACUUM afterwards so the deleted rows are not merely unlinked pages still
// readable in the file that gets pushed.
func (db *DB) RedactForBackup() (int, error) {
	res, err := db.sql.Exec("DELETE FROM sessions")
	if err != nil {
		return 0, fmt.Errorf("store: redacting sessions: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("store: redacting sessions: %w", err)
	}
	if _, err := db.sql.Exec("VACUUM"); err != nil {
		return 0, fmt.Errorf("store: vacuuming after redaction: %w", err)
	}
	return int(n), nil
}

// IntegrityCheck asks SQLite whether the file is sound.
//
// Used by internal/backup before a snapshot is allowed to be committed and
// pushed. A backup nobody checked is a backup nobody can rely on, and the point
// at which to find out is now rather than on the day the server is gone.
func (db *DB) IntegrityCheck() error {
	var result string
	if err := db.sql.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil {
		return fmt.Errorf("store: integrity check: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("store: integrity check on %s failed: %s", db.path, result)
	}
	return nil
}

// CountUsers is how many accounts the database holds. Reported in the backup
// manifest so that a restore can tell at a glance whether it can be logged in
// to at all.
func (db *DB) CountUsers() (int, error) {
	var n int
	if err := db.sql.QueryRow("SELECT COUNT(*) FROM users").Scan(&n); err != nil {
		return 0, fmt.Errorf("store: counting users: %w", err)
	}
	return n, nil
}

// Result is what one import wrote.
type Result struct {
	Flights       int
	Aircraft      int
	Discrepancies int
	Totals        csvbook.Totals
	BackupPath    string
}

// Import replaces the contents of the database with lb, verifying before it
// commits.
//
// note is free text recorded in import_runs -- typically what was imported and
// why.
func (db *DB) Import(lb *csvbook.Logbook, note string) (Result, error) {
	var res Result

	tx, err := db.sql.Begin()
	if err != nil {
		return res, fmt.Errorf("store: starting the import transaction: %w", err)
	}
	// Rollback after a successful Commit is a no-op, so this is the safe
	// default: any path out of this function that is not the happy one leaves
	// the database exactly as it was found.
	defer tx.Rollback()

	// Replace rather than merge. A merge would leave a row that disappeared
	// from the CSVs alive in the database as a phantom total.
	if _, err := tx.Exec("DELETE FROM discrepancies"); err != nil {
		return res, fmt.Errorf("store: clearing discrepancies: %w", err)
	}
	// Aircraft are cleared only where the importer owns them. This DELETE was
	// unqualified until 2026-08-02, which would have destroyed every aeroplane
	// added by hand -- the same trap source_book = 0 solves for flights, and it
	// was wide open the moment aircraft got a write path.
	if _, err := tx.Exec("DELETE FROM aircraft WHERE user_added = 0"); err != nil {
		return res, fmt.Errorf("store: clearing the imported aircraft: %w", err)
	}
	// Flights are cleared only where the importer owns them. Rows typed into
	// the app carry source_book 0, are not in any CSV, and would be destroyed
	// by an unqualified DELETE -- and this import runs again every time the
	// migration effort appends a page to logbook_3.csv. Losing them is exactly
	// what CLAUDE.md rule 0.2 forbids.
	if _, err := tx.Exec("DELETE FROM flights WHERE source_book <> ?", entry.HandEnteredBook); err != nil {
		return res, fmt.Errorf("store: clearing the imported flights: %w", err)
	}

	aircraftID, err := insertAircraft(tx, lb.Aircraft)
	if err != nil {
		return res, err
	}
	if err := insertFlights(tx, lb.Flights, aircraftID); err != nil {
		return res, err
	}
	if err := insertDiscrepancies(tx, lb.Discrepancies); err != nil {
		return res, err
	}
	if err := relinkHandEntered(tx); err != nil {
		return res, err
	}

	// The gate. Everything above is provisional until this passes.
	written, err := readBackTotals(tx)
	if err != nil {
		return res, err
	}
	if err := compare(written, lb.Totals); err != nil {
		return res, err
	}

	if _, err := tx.Exec(
		`INSERT INTO import_runs
		   (ran_at, flights, aircraft, discrepancies, total_minutes, landings, backup_path, note)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		now(), written.Flights, len(lb.Aircraft), len(lb.Discrepancies),
		written.Total, written.Landings, db.backup, note,
	); err != nil {
		return res, fmt.Errorf("store: recording the import run: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return res, fmt.Errorf("store: committing the import: %w", err)
	}

	return Result{
		Flights:       written.Flights,
		Aircraft:      len(lb.Aircraft),
		Discrepancies: len(lb.Discrepancies),
		Totals:        written,
		BackupPath:    db.backup,
	}, nil
}

func insertAircraft(tx *sql.Tx, list []csvbook.Aircraft) (map[string]int64, error) {
	ids := make(map[string]int64, len(list))
	stmt, err := tx.Prepare(
		`INSERT INTO aircraft (registration, type, default_class, ifr_capable, active, notes)
		 VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, fmt.Errorf("store: preparing the aircraft insert: %w", err)
	}
	defer stmt.Close()

	for _, a := range list {
		r, err := stmt.Exec(a.Registration, a.Type, string(a.DefaultClass),
			boolToInt(a.IFRCapable), boolToInt(a.Active), a.Notes)
		if err != nil {
			return nil, fmt.Errorf("store: inserting aircraft %s: %w", a.Registration, err)
		}
		id, err := r.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("store: inserting aircraft %s: %w", a.Registration, err)
		}
		ids[a.Registration] = id
	}
	return ids, nil
}

func insertFlights(tx *sql.Tx, flights []csvbook.Flight, aircraftID map[string]int64) error {
	stmt, err := tx.Prepare(
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
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("store: preparing the flight insert: %w", err)
	}
	defer stmt.Close()

	for _, f := range flights {
		// A missing aircraft row must never cost us a flight, so the link is
		// nullable and an unknown registration simply goes in unlinked.
		var link any
		if id, ok := aircraftID[f.AircraftReg]; ok {
			link = id
		}
		if _, err := stmt.Exec(
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
			return fmt.Errorf("store: inserting flight seq %d (book %d row %d): %w",
				f.Seq, f.SourceBook, f.SourceRow, err)
		}
	}
	return nil
}

func insertDiscrepancies(tx *sql.Tx, list []csvbook.Discrepancy) error {
	stmt, err := tx.Prepare(
		`INSERT INTO discrepancies (kind, source_book, source_row, flight_date, detail)
		 VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("store: preparing the discrepancy insert: %w", err)
	}
	defer stmt.Close()

	for _, d := range list {
		if _, err := stmt.Exec(string(d.Kind), d.Book, d.Row, d.Date, d.Detail); err != nil {
			return fmt.Errorf("store: inserting discrepancy %s: %w", d.Kind, err)
		}
	}
	return nil
}

// querier is satisfied by both *sql.DB and *sql.Tx, so verification can run
// inside the import transaction and afterwards from outside it.
type querier interface {
	Query(string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
}

// totalsQuery recomputes every checksum from what is actually in the table.
//
// COALESCE because SUM over no rows is NULL, and an empty range is a real
// case: importing an empty logbook must give zeros, not an error.
//
// Scoped to the imported rows. The question these checksums answer is "is the
// database what the CSVs say", and a flight typed into the app is in no CSV --
// counting it would make the import fail verification on its own correct work,
// and the only way to pass would be to delete the pilot's flight. The
// hand-entered rows are guarded instead by AddFlight's own constraints.
const totalsQuery = `
	SELECT COUNT(*),
	       COUNT(DISTINCT seq),
	       COALESCE(SUM(total_minutes), 0),
	       COALESCE(SUM(pic_minutes), 0),
	       COALESCE(SUM(dual_minutes), 0),
	       COALESCE(SUM(instrument_minutes), 0),
	       COALESCE(SUM(night_minutes), 0),
	       COALESCE(SUM(instructor_minutes), 0),
	       COALESCE(SUM(CASE WHEN class = 'SEP_SEA' THEN total_minutes ELSE 0 END), 0),
	       COALESCE(SUM(landings_day + landings_night), 0)
	FROM flights WHERE source_book <> ?`

func readBackTotals(q querier) (csvbook.Totals, error) {
	var t csvbook.Totals
	var distinctSeq int
	err := q.QueryRow(totalsQuery, entry.HandEnteredBook).Scan(
		&t.Flights, &distinctSeq, &t.Total, &t.PIC, &t.Dual,
		&t.Instrument, &t.Night, &t.Instructor, &t.SEPSea, &t.Landings,
	)
	if err != nil {
		return t, fmt.Errorf("store: reading back the totals: %w", err)
	}
	// seq is UNIQUE, so this cannot currently differ -- but seq is the key
	// every cumulative computation walks, and a duplicate would silently
	// corrupt every running total downstream.
	if distinctSeq != t.Flights {
		return t, fmt.Errorf("store: %d flights carry only %d distinct seq values",
			t.Flights, distinctSeq)
	}
	return t, nil
}

// compare checks each figure separately rather than a single grand total,
// because two errors of opposite sign would cancel in a combined figure and
// pass a check that means nothing.
func compare(got, want csvbook.Totals) error {
	for _, c := range []struct {
		name      string
		got, want int
	}{
		{"flight count", got.Flights, want.Flights},
		{"total time", got.Total, want.Total},
		{"pic time", got.PIC, want.PIC},
		{"dual time", got.Dual, want.Dual},
		{"instrument time", got.Instrument, want.Instrument},
		{"night time", got.Night, want.Night},
		{"instructor time", got.Instructor, want.Instructor},
		{"seaplane time", got.SEPSea, want.SEPSea},
		{"landings", got.Landings, want.Landings},
	} {
		if c.got != c.want {
			return fmt.Errorf(
				"store: import verification failed on %s: the database holds %d but the "+
					"source CSVs total %d. Nothing was committed", c.name, c.got, c.want)
		}
	}
	return nil
}

// Verify recomputes the checksums from the committed data and compares them to
// want. It is the same gate Import applies, available to run against a live
// database without writing anything.
func (db *DB) Verify(want csvbook.Totals) error {
	got, err := readBackTotals(db.sql)
	if err != nil {
		return err
	}
	return compare(got, want)
}

// CountFlights returns how many flights the database holds.
func (db *DB) CountFlights() (int, error) {
	var n int
	if err := db.sql.QueryRow("SELECT COUNT(*) FROM flights").Scan(&n); err != nil {
		return 0, fmt.Errorf("store: counting flights: %w", err)
	}
	return n, nil
}

// AircraftLinkage reports how many flights resolved to an aircraft row and how
// many did not. A non-zero unlinked count is worth looking at: it means a
// registration was flown that the seed list does not know.
func (db *DB) AircraftLinkage() (linked, unlinked int, err error) {
	err = db.sql.QueryRow(
		`SELECT COUNT(aircraft_id), COUNT(*) - COUNT(aircraft_id) FROM flights`,
	).Scan(&linked, &unlinked)
	if err != nil {
		return 0, 0, fmt.Errorf("store: counting aircraft linkage: %w", err)
	}
	return linked, unlinked, nil
}

// flightColumns is the one column list every flight read uses. Shared so that
// the list query and the single-row read cannot drift apart -- a mismatch
// between them would be a field silently missing from one of the two.
const flightColumns = `seq, flight_date, aircraft_reg, aircraft_type, class, dep_place, arr_place,
	        off_block_utc, on_block_utc, off_block_raw, on_block_raw, time_origin,
	        takeoff_utc, landing_utc,
	        block_minutes, total_minutes, night_minutes, instrument_minutes,
	        pic_minutes, dual_minutes, instructor_minutes,
	        copilot_minutes, multipilot_minutes,
	        pic_name, landings_day, landings_night, landings_verified, remarks,
	        source_book, source_row`

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface{ Scan(...any) error }

// scanFlight reads one row of flightColumns into a domain flight.
func scanFlight(s rowScanner) (csvbook.Flight, error) {
	var f csvbook.Flight
	var class, origin string
	var off, on, takeoff, landing sql.NullString
	var verified int
	if err := s.Scan(
		&f.Seq, &f.Date, &f.AircraftReg, &f.AircraftType, &class, &f.DepPlace, &f.ArrPlace,
		&off, &on, &f.OffBlockRaw, &f.OnBlockRaw, &origin,
		&takeoff, &landing,
		&f.BlockMinutes, &f.TotalMinutes, &f.NightMinutes, &f.InstrumentMinutes,
		&f.PICMinutes, &f.DualMinutes, &f.InstructorMinutes,
		&f.CopilotMinutes, &f.MultiPilotMinutes,
		&f.PICName, &f.LandingsDay, &f.LandingsNight, &verified, &f.Remarks,
		&f.SourceBook, &f.SourceRow,
	); err != nil {
		return csvbook.Flight{}, err
	}
	f.Class = csvbook.Class(class)
	f.TimeOrigin = timeutil.Origin(origin)
	f.LandingsVerified = verified == 1
	for _, p := range []struct {
		src sql.NullString
		dst *time.Time
	}{{off, &f.OffBlockUTC}, {on, &f.OnBlockUTC}, {takeoff, &f.TakeoffUTC}, {landing, &f.LandingUTC}} {
		if !p.src.Valid {
			continue
		}
		t, err := time.Parse(timeFormat, p.src.String)
		if err != nil {
			return csvbook.Flight{}, fmt.Errorf("store: flight seq %d has an unreadable instant %q: %w",
				f.Seq, p.src.String, err)
		}
		*p.dst = t
	}
	return f, nil
}

// flightBySeq reads one flight, reporting ErrFlightNotFound if there is none.
func flightBySeq(q querier, seq int) (csvbook.Flight, error) {
	f, err := scanFlight(q.QueryRow(`SELECT `+flightColumns+` FROM flights WHERE seq = ?`, seq))
	if errors.Is(err, sql.ErrNoRows) {
		return csvbook.Flight{}, fmt.Errorf("%w: %d", ErrFlightNotFound, seq)
	}
	if err != nil {
		return csvbook.Flight{}, fmt.Errorf("store: reading flight %d: %w", seq, err)
	}
	return f, nil
}

// Flights returns every flight in book order.
func (db *DB) Flights() ([]csvbook.Flight, error) {
	rows, err := db.sql.Query(`SELECT ` + flightColumns + ` FROM flights ORDER BY seq`)
	if err != nil {
		return nil, fmt.Errorf("store: reading flights: %w", err)
	}
	defer rows.Close()

	var out []csvbook.Flight
	for rows.Next() {
		f, err := scanFlight(rows)
		if err != nil {
			return nil, fmt.Errorf("store: reading flights: %w", err)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// Aircraft returns the seed list in registration order.
func (db *DB) Aircraft() ([]csvbook.Aircraft, error) {
	rows, err := db.sql.Query(
		`SELECT registration, type, default_class, ifr_capable, active, notes
		 FROM aircraft ORDER BY registration`)
	if err != nil {
		return nil, fmt.Errorf("store: reading aircraft: %w", err)
	}
	defer rows.Close()

	var out []csvbook.Aircraft
	for rows.Next() {
		var a csvbook.Aircraft
		var class string
		var ifr, active int
		if err := rows.Scan(&a.Registration, &a.Type, &class, &ifr, &active, &a.Notes); err != nil {
			return nil, fmt.Errorf("store: reading aircraft: %w", err)
		}
		a.DefaultClass = csvbook.Class(class)
		a.IFRCapable = ifr == 1
		a.Active = active == 1
		out = append(out, a)
	}
	return out, rows.Err()
}

// Discrepancies returns the review list in book order.
func (db *DB) Discrepancies() ([]csvbook.Discrepancy, error) {
	rows, err := db.sql.Query(
		`SELECT kind, source_book, source_row, flight_date, detail
		 FROM discrepancies ORDER BY source_book, source_row, kind`)
	if err != nil {
		return nil, fmt.Errorf("store: reading discrepancies: %w", err)
	}
	defer rows.Close()

	var out []csvbook.Discrepancy
	for rows.Next() {
		var d csvbook.Discrepancy
		var kind string
		if err := rows.Scan(&kind, &d.Book, &d.Row, &d.Date, &d.Detail); err != nil {
			return nil, fmt.Errorf("store: reading discrepancies: %w", err)
		}
		d.Kind = csvbook.Kind(kind)
		out = append(out, d)
	}
	return out, rows.Err()
}

// ImportRun is one entry in the audit trail.
type ImportRun struct {
	ID            int
	RanAt         string
	Flights       int
	Aircraft      int
	Discrepancies int
	TotalMinutes  int
	Landings      int
	BackupPath    string
	Note          string
}

// ImportRuns returns every recorded import, oldest first.
func (db *DB) ImportRuns() ([]ImportRun, error) {
	rows, err := db.sql.Query(
		`SELECT id, ran_at, flights, aircraft, discrepancies, total_minutes,
		        landings, backup_path, note
		 FROM import_runs ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("store: reading import runs: %w", err)
	}
	defer rows.Close()

	var out []ImportRun
	for rows.Next() {
		var r ImportRun
		if err := rows.Scan(&r.ID, &r.RanAt, &r.Flights, &r.Aircraft, &r.Discrepancies,
			&r.TotalMinutes, &r.Landings, &r.BackupPath, &r.Note); err != nil {
			return nil, fmt.Errorf("store: reading import runs: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC().Format(timeFormat)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func now() string { return time.Now().UTC().Format(timeFormat) }
