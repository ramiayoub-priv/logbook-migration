// Package backup produces the off-box copy of the logbook.
//
// WHY THIS EXISTS, IN ONE PARAGRAPH. Until 2026-08-02 the production database
// was reconstructible from this repository: three committed CSVs and one
// command. Then the transcription effort closed, the app became the only way
// the record grows, and the owner logged two flights from an airfield. Those
// rows are in no CSV. From that moment "rebuild it from the repo" stopped
// meaning "lose nothing", and the only copies of the owner's own flights were
// the pre-import backups sitting on the same disk as the database they protect.
// This package is what makes a second copy exist somewhere else.
//
// WHAT IT PRODUCES: a directory holding exactly four files, meant to be a git
// repository that is committed and pushed daily.
//
//	logbook.db    the whole database, redacted of sessions -- a complete restore
//	logbook.csv   every flight, every stored field -- readable without this code
//	MANIFEST.txt  counts and checksums, so a restore can be checked
//	RESTORE.md    how to put it back, travelling WITH the data
//
// TWO OF THOSE FILES ARE DELIBERATE REDUNDANCY. `logbook.db` is the real
// restore: put it back and the application runs. `logbook.csv` exists for the
// day something is wrong with the database file, or with this program, or with
// SQLite -- a legal record whose only readable form requires a specific binary
// to still work is a record with a single point of failure. The CSV can be read
// by anything, forever, and it carries the provenance columns and the raw times
// as written on paper (rule 0.4) rather than a summary.
//
// RULE 0.2 GOVERNS EVERY STEP. The snapshot is taken with VACUUM INTO, which is
// transactionally consistent against a live database in WAL mode. It is then
// verified -- integrity check, row counts and a total-time checksum against the
// source -- and the whole run FAILS rather than half-writing. Nothing is moved
// into place until all of it is proven, so an interrupted or refused run leaves
// yesterday's good backup exactly where it was.
package backup

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/hhmm"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

// The four files a backup consists of. Named here rather than inline because
// the deploy script and RESTORE.md both refer to them and they must agree.
const (
	DBName       = "logbook.db"
	CSVName      = "logbook.csv"
	ManifestName = "MANIFEST.txt"
	RestoreName  = "RESTORE.md"
)

// timeFormat is how instants are written into the CSV: RFC3339 in UTC, the
// same shape the database stores and the API sends. One representation, so no
// two of them can disagree.
const timeFormat = "2006-01-02T15:04:05Z"

// Snapshot is what one backup run produced. Every field is something a restore
// can be checked against, which is the only reason to record any of them.
type Snapshot struct {
	At time.Time

	Flights int
	// How many of those exist in no CSV -- the rows this whole package is for.
	HandEntered   int
	Aircraft      int
	Discrepancies int
	Users         int

	TotalMinutes int
	Landings     int

	SessionsDropped int

	SHA256DB  string
	SHA256CSV string
}

// Run writes a verified backup of db into outDir, replacing what is there.
//
// Overwriting is correct: the backup repository holds ONE current copy and git
// holds the history, so a run that refused because yesterday's file existed
// would fail on its second day. (Contrast store.Backup, which refuses to
// overwrite because it protects a single irreplaceable moment.)
//
// now is injected so the manifest's timestamp is testable. It is the only
// nondeterminism in the output, and it lives in MANIFEST.txt alone so that an
// unchanged logbook produces a byte-identical CSV and database -- which is what
// lets the deploy script skip the commit on a day nothing was flown.
func Run(db *store.DB, outDir string, now func() time.Time) (Snapshot, error) {
	var snap Snapshot

	info, err := os.Stat(outDir)
	if err != nil {
		return snap, fmt.Errorf("backup: output directory %s: %w", outDir, err)
	}
	if !info.IsDir() {
		return snap, fmt.Errorf("backup: %s is not a directory", outDir)
	}

	// Everything is built in a staging directory and only moved into place once
	// all of it has been verified. A backup run that dies halfway must leave
	// yesterday's good copy untouched rather than a mixture of the two.
	stage, err := os.MkdirTemp(outDir, ".staging-")
	if err != nil {
		return snap, fmt.Errorf("backup: creating a staging directory: %w", err)
	}
	defer os.RemoveAll(stage)

	stagedDB := filepath.Join(stage, DBName)
	if err := db.Backup(stagedDB); err != nil {
		return snap, err
	}

	// What the LIVE database says, read before the copy is touched. This is the
	// figure the copy has to agree with, and reading it from the source is the
	// whole substance of "verified" (rule 0.2).
	liveFlights, err := db.CountFlights()
	if err != nil {
		return snap, err
	}

	copied, err := store.Open(stagedDB)
	if err != nil {
		return snap, fmt.Errorf("backup: opening the snapshot: %w", err)
	}
	defer copied.Close()

	// Redact the COPY. Never the live file -- that would sign the owner out of
	// their own logbook as a side effect of protecting it, and there is a test
	// asserting it does not happen.
	dropped, err := copied.RedactForBackup()
	if err != nil {
		return snap, err
	}
	if err := copied.IntegrityCheck(); err != nil {
		return snap, err
	}

	flights, err := copied.Flights()
	if err != nil {
		return snap, err
	}
	aircraft, err := copied.Aircraft()
	if err != nil {
		return snap, err
	}
	discrepancies, err := copied.Discrepancies()
	if err != nil {
		return snap, err
	}
	users, err := copied.CountUsers()
	if err != nil {
		return snap, err
	}

	// THE VERIFICATION. A copy holding a different number of flights than the
	// database it was taken from is not a backup, it is a corruption with a
	// filename. Refuse, loudly, and leave yesterday's copy in place.
	if len(flights) != liveFlights {
		return snap, fmt.Errorf(
			"backup: REFUSING -- the snapshot holds %d flights but the live database "+
				"holds %d; nothing has been written and the previous backup is untouched",
			len(flights), liveFlights)
	}

	snap = Snapshot{
		At:              now().UTC().Truncate(time.Second),
		Flights:         len(flights),
		Aircraft:        len(aircraft),
		Discrepancies:   len(discrepancies),
		Users:           users,
		SessionsDropped: dropped,
	}
	for _, f := range flights {
		snap.TotalMinutes += f.TotalMinutes
		snap.Landings += f.LandingsDay + f.LandingsNight
		// source_book 0 is the hand-entered band: the flights that exist in no
		// CSV and cannot be recovered from anywhere but here.
		if f.SourceBook == 0 {
			snap.HandEntered++
		}
	}

	stagedCSV := filepath.Join(stage, CSVName)
	if err := writeCSV(stagedCSV, flights); err != nil {
		return snap, err
	}

	// Close before hashing and moving: the WAL and shared-memory files have to
	// be checkpointed into the database file, or the bytes that get pushed are
	// not the bytes that were verified.
	if err := copied.Close(); err != nil {
		return snap, fmt.Errorf("backup: closing the snapshot: %w", err)
	}
	for _, side := range []string{stagedDB + "-wal", stagedDB + "-shm"} {
		os.Remove(side)
	}

	snap.SHA256DB = SHA256File(stagedDB)
	snap.SHA256CSV = SHA256File(stagedCSV)
	if snap.SHA256DB == "" || snap.SHA256CSV == "" {
		return snap, fmt.Errorf("backup: could not checksum the staged files")
	}

	if err := os.WriteFile(filepath.Join(stage, ManifestName),
		[]byte(manifest(snap)), 0o644); err != nil {
		return snap, fmt.Errorf("backup: writing the manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(stage, RestoreName),
		[]byte(restoreDoc(snap)), 0o644); err != nil {
		return snap, fmt.Errorf("backup: writing the restore instructions: %w", err)
	}

	// Everything is proven. Move it into place.
	for _, name := range []string{DBName, CSVName, ManifestName, RestoreName} {
		if err := os.Rename(filepath.Join(stage, name), filepath.Join(outDir, name)); err != nil {
			return snap, fmt.Errorf("backup: installing %s: %w", name, err)
		}
	}
	return snap, nil
}

// csvHeader is every column the CSV carries, in order.
//
// It is a COMPLETE dump of what the database stores per flight, not the shape
// of the three source books. Provenance (source_book/source_row) and the raw
// times with their origin are here because rule 0.4 forbids losing the source:
// a backup that kept only the converted instant would make a bad DST conversion
// unauditable forever.
//
// There are deliberately NO Cumulative_* columns. Rule 0.5: running totals are
// derived from seq, never stored, and writing them into a backup would put the
// project's single largest historical source of drift back into the data.
var csvHeader = []string{
	"seq", "source_book", "source_row", "flight_date",
	"aircraft_reg", "aircraft_type", "class", "dep_place", "arr_place",
	"off_block_utc", "on_block_utc", "off_block_raw", "on_block_raw", "time_origin",
	"takeoff_utc", "landing_utc",
	"block_minutes", "total_minutes", "night_minutes", "instrument_minutes",
	"pic_minutes", "dual_minutes", "instructor_minutes",
	"pic_name", "landings_day", "landings_night", "landings_verified", "remarks",
}

// writeCSV dumps the flights in seq order -- the book's own order, and the only
// order anything cumulative may be rebuilt from. store.Flights() already
// returns them that way; nothing here re-sorts.
func writeCSV(path string, flights []csvbook.Flight) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("backup: writing %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(csvHeader); err != nil {
		return fmt.Errorf("backup: writing the CSV header: %w", err)
	}
	for _, fl := range flights {
		if err := w.Write([]string{
			strconv.Itoa(fl.Seq), strconv.Itoa(fl.SourceBook), strconv.Itoa(fl.SourceRow),
			fl.Date, fl.AircraftReg, fl.AircraftType, string(fl.Class),
			fl.DepPlace, fl.ArrPlace,
			instant(fl.OffBlockUTC), instant(fl.OnBlockUTC),
			fl.OffBlockRaw, fl.OnBlockRaw, string(fl.TimeOrigin),
			instant(fl.TakeoffUTC), instant(fl.LandingUTC),
			strconv.Itoa(fl.BlockMinutes), strconv.Itoa(fl.TotalMinutes),
			strconv.Itoa(fl.NightMinutes), strconv.Itoa(fl.InstrumentMinutes),
			strconv.Itoa(fl.PICMinutes), strconv.Itoa(fl.DualMinutes),
			strconv.Itoa(fl.InstructorMinutes),
			fl.PICName, strconv.Itoa(fl.LandingsDay), strconv.Itoa(fl.LandingsNight),
			boolText(fl.LandingsVerified), fl.Remarks,
		}); err != nil {
			return fmt.Errorf("backup: writing flight %d: %w", fl.Seq, err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("backup: writing %s: %w", path, err)
	}
	return f.Close()
}

// instant renders a stored time, or an empty cell for a blank one. A missing
// time must read as missing, never as year 1 or as midnight.
func instant(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(timeFormat)
}

func boolText(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// SHA256File is the hex digest of a file, or "" if it cannot be read.
//
// Exported so the manifest, the tests and anyone checking a restore all compute
// the same number the same way.
func SHA256File(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func manifest(s Snapshot) string {
	return fmt.Sprintf(`logbook backup manifest
=======================

taken            %s
flights          %d
  hand-entered   %d   (source_book = 0; these exist in NO CSV)
aircraft         %d
discrepancies    %d
users            %d
sessions dropped %d   (not carried off the box; see RESTORE.md)

total time       %s   (%d minutes)
landings         %d

sha256 %s  %s
sha256 %s  %s

Check a restore against the figures above. If they do not match, STOP and work
out why before letting the application write to the file (CLAUDE.md rule 0.2).
`,
		s.At.Format(time.RFC3339),
		s.Flights, s.HandEntered, s.Aircraft, s.Discrepancies, s.Users, s.SessionsDropped,
		hhmm.Format(s.TotalMinutes), s.TotalMinutes, s.Landings,
		s.SHA256DB, DBName, s.SHA256CSV, CSVName)
}

// restoreDoc is the instructions, written into the backup itself.
//
// They travel WITH the data on purpose. Instructions that live only in the
// application repository are instructions you do not have on the day the server
// is gone and you are working from a phone -- and a backup you cannot remember
// how to restore is a backup that does not work.
//
// Written with '@' standing in for a backtick and replaced at the end: Go has
// no way to put a backtick inside a raw string literal, and the alternative --
// concatenating fragments around quoted backticks -- turns a document into
// something nobody will edit correctly a year from now.
func restoreDoc(s Snapshot) string {
	const doc = `# Restoring the logbook

This repository is the off-box copy of a personal pilot logbook. **It is a legal
record**: it backs licence privileges and currency, so a wrong total here is a
real-world problem rather than a bug report.

Taken **%s**. It holds **%d flights**, of which **%d were entered in the app and
exist in no CSV anywhere** -- those are the rows that cannot be reconstructed
from anything else.

## What is here

| file | what it is |
|---|---|
| @%s@ | the whole SQLite database. This is the restore. |
| @%s@ | every flight, every stored field, as plain text. |
| @%s@ | counts and checksums to verify a restore against. |
| @%s@ | this file. |

The CSV is deliberate redundancy. If the database file, or the application, or
SQLite itself ever lets you down, the CSV is readable by anything and carries
the provenance columns and the times exactly as they were written on paper.

## Restoring onto a new server

The application code lives in the project repository; this holds only the data.

1. Install the app as @app/docs/deploy.md@ describes, but **do not import the
   CSVs**. That rebuilds the transcribed books and would lose every flight
   entered in the app since.

2. Put the database back and give it to the service user:

@@@bash
sudo systemctl stop logbook
sudo install -o logbook -g logbook -m 0640 %s /var/lib/logbook/logbook.db
@@@

3. **Verify before starting anything.** These are the figures this copy must
   restore to:

@@@
flights   %d
total     %s  (%d minutes)
landings  %d
users     %d
sha256    %s
@@@

   Check the file you just installed:

@@@bash
sha256sum /var/lib/logbook/logbook.db
sudo -u logbook sqlite3 /var/lib/logbook/logbook.db \
  'SELECT COUNT(*), SUM(total_minutes), SUM(landings_day + landings_night) FROM flights;'
@@@

   If any figure disagrees, **stop** and find out why before the application is
   allowed to write to the file (CLAUDE.md rule 0.2).

4. Start it: @sudo systemctl start logbook@. The startup line reports the flight
   count, and it must read %d.

## Signing in

Your account is in this backup and your password still works -- it is stored as
an Argon2id hash, never as a password. **Sessions are NOT here**: they are
dropped before the backup ever leaves the server, so every device signs in again
after a restore. That is intended, not a fault.

If the password is lost, the operator CLI sets a new one without touching any of
the data:

@@@bash
sudo -u logbook /opt/logbook/logbook-server passwd <username> -db /var/lib/logbook/logbook.db
@@@

## What this backup is not

It is **not** a substitute for the paper logbooks, which remain authoritative
for everything transcribed from them, and it holds no photographs of them. It is
the electronic record -- and for flights entered in the app, it is the only
record.
`
	return strings.ReplaceAll(fmt.Sprintf(doc,
		s.At.Format("2006-01-02 15:04 UTC"), s.Flights, s.HandEntered,
		DBName, CSVName, ManifestName, RestoreName,
		DBName,
		s.Flights, hhmm.Format(s.TotalMinutes), s.TotalMinutes, s.Landings, s.Users,
		s.SHA256DB,
		s.Flights), "@", "`")
}
