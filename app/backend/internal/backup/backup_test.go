package backup_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/backup"
	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

// The off-box backup (Task 14). It is the ONLY protection for flights entered
// in the app: those exist in no CSV, so "rebuild it from the repo" stopped
// being a complete answer on 2026-08-02 when the owner logged the first two.
//
// Rule 0.2 governs every line of this: a backup is verified before it is
// believed, it is refused rather than half-written, and it must be restorable
// on its own -- a copy nobody can restore from is not a backup, it is a file.

const pw = "a-sufficiently-long-passphrase"

var fixedNow = time.Date(2026, 8, 2, 3, 0, 0, 0, time.UTC)

func newLogbook() *csvbook.Logbook {
	at := func(s string) time.Time {
		v, _ := time.Parse(time.RFC3339, s)
		return v
	}
	return &csvbook.Logbook{
		Flights: []csvbook.Flight{
			{
				Seq: 1, Date: "2021-06-01",
				AircraftType: "C172", AircraftReg: "OH-CTL", Class: csvbook.ClassSEPSea,
				DepPlace: "Tuusulanjärvi", ArrPlace: "Tuusulanjärvi",
				OffBlockUTC: at("2021-06-01T15:13:00Z"), OnBlockUTC: at("2021-06-01T16:34:00Z"),
				OffBlockRaw: "18:13", OnBlockRaw: "19:34", TimeOrigin: "converted_from_local",
				BlockMinutes: 81, TotalMinutes: 81, PICMinutes: 81,
				PICName: "self", LandingsDay: 7, LandingsVerified: true,
				SourceBook: 3, SourceRow: 3,
			},
			{
				Seq: 2, Date: "2022-06-02",
				AircraftType: "P28A", AircraftReg: "OH-PDP", Class: csvbook.ClassSEPLand,
				DepPlace: "EFHV", ArrPlace: "EFHV",
				OffBlockUTC: at("2022-06-02T06:00:00Z"), OnBlockUTC: at("2022-06-02T07:30:00Z"),
				OffBlockRaw: "06:00Z", OnBlockRaw: "07:30Z", TimeOrigin: "utc_as_written",
				BlockMinutes: 90, TotalMinutes: 90, DualMinutes: 90,
				PICName: "Sinervä", LandingsDay: 3, LandingsVerified: true,
				SourceBook: 3, SourceRow: 4,
			},
		},
		Aircraft: []csvbook.Aircraft{
			{Registration: "OH-CTL", Type: "C172", DefaultClass: csvbook.ClassSEPSea, Active: true},
			{Registration: "OH-PDP", Type: "P28A", DefaultClass: csvbook.ClassSEPLand, Active: true},
		},
		Totals: csvbook.Totals{
			Flights: 2, Total: 171, PIC: 81, Dual: 90, SEPSea: 81, Landings: 10,
		},
	}
}

// live returns a database standing in for production: the imported books, an
// account, an open session, and one flight typed into the app.
func live(t *testing.T) *store.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "logbook.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Import(newLogbook(), "test"); err != nil {
		t.Fatal(err)
	}
	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.CreateSession(u.ID, "phone", "10.0.0.1"); err != nil {
		t.Fatal(err)
	}

	// The row that exists nowhere else. Everything about this package is for
	// this row.
	handEntered := csvbook.Flight{
		Date: "2026-08-02", AircraftType: "C172", AircraftReg: "OH-CAM",
		Class: csvbook.ClassSEPLand, DepPlace: "EFHV", ArrPlace: "EFHV",
		OffBlockUTC: time.Date(2026, 8, 2, 9, 15, 0, 0, time.UTC),
		OnBlockUTC:  time.Date(2026, 8, 2, 10, 30, 0, 0, time.UTC),
		OffBlockRaw: "09:15Z", OnBlockRaw: "10:30Z", TimeOrigin: "utc_as_written",
		TakeoffUTC:   time.Date(2026, 8, 2, 9, 20, 0, 0, time.UTC),
		LandingUTC:   time.Date(2026, 8, 2, 10, 25, 0, 0, time.UTC),
		BlockMinutes: 75, TotalMinutes: 75, PICMinutes: 75,
		PICName: "self", LandingsDay: 1, LandingsVerified: true,
	}
	if _, err := db.AddFlight(handEntered); err != nil {
		t.Fatal(err)
	}
	return db
}

func run(t *testing.T, db *store.DB) (backup.Snapshot, string) {
	t.Helper()
	out := t.TempDir()
	snap, err := backup.Run(db, out, func() time.Time { return fixedNow })
	if err != nil {
		t.Fatalf("backup.Run: %v", err)
	}
	return snap, out
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}

// --- What a backup contains -------------------------------------------------

func TestRunWritesTheDatabaseTheCSVAndTheInstructions(t *testing.T) {
	_, out := run(t, live(t))

	for _, name := range []string{"logbook.db", "logbook.csv", "MANIFEST.txt", "RESTORE.md"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("%s is missing from the backup: %v", name, err)
		}
	}
	// Nothing else. A backup directory is a git repository that gets pushed,
	// so a stray temporary file becomes a commit.
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("backup holds %v, want exactly the four files", names)
	}
}

// The whole point: the flight that exists in no CSV has to be in the backup.
func TestBackupCarriesTheFlightsThatExistNowhereElse(t *testing.T) {
	snap, out := run(t, live(t))

	if snap.Flights != 3 || snap.HandEntered != 1 {
		t.Errorf("snapshot has %d flights (%d hand-entered), want 3 (1)",
			snap.Flights, snap.HandEntered)
	}
	csv := read(t, filepath.Join(out, "logbook.csv"))
	if !strings.Contains(csv, "OH-CAM") || !strings.Contains(csv, "2026-08-02") {
		t.Error("the hand-entered flight is not in the CSV")
	}
	// And its airborne times, which nothing else records.
	if !strings.Contains(csv, "2026-08-02T09:20:00Z") {
		t.Error("the hand-entered flight's takeoff time is not in the CSV")
	}
}

// A restore has to be able to log in, so the account survives -- but the
// sessions do not leave the box.
func TestBackupKeepsTheAccountAndDropsTheSessions(t *testing.T) {
	snap, out := run(t, live(t))

	if snap.Users != 1 {
		t.Errorf("snapshot has %d users, want 1 -- a logbook nobody can open "+
			"is not a restored logbook", snap.Users)
	}
	if snap.SessionsDropped != 1 {
		t.Errorf("dropped %d sessions, want 1", snap.SessionsDropped)
	}

	copied, err := store.Open(filepath.Join(out, "logbook.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer copied.Close()

	users, err := copied.Users()
	if err != nil || len(users) != 1 || users[0].Username != "rami" {
		t.Errorf("the account did not survive into the backup: %+v, %v", users, err)
	}
	// The restored copy can actually be signed in to. This is the assertion
	// that makes "restorable" mean something.
	if _, err := copied.Authenticate("rami", pw); err != nil {
		t.Errorf("cannot authenticate against the restored copy: %v", err)
	}
	sessions, err := copied.Sessions(users[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Errorf("%d sessions were carried off the box", len(sessions))
	}
}

// The backup is verified against the live database before it is believed. Every
// figure in the manifest is one a restore can be checked against.
func TestBackupIsVerifiedAgainstTheLiveDatabase(t *testing.T) {
	db := live(t)
	snap, out := run(t, db)

	liveCount, err := db.CountFlights()
	if err != nil {
		t.Fatal(err)
	}
	if snap.Flights != liveCount {
		t.Errorf("snapshot has %d flights, live database has %d", snap.Flights, liveCount)
	}
	// 81 + 90 + 75.
	if snap.TotalMinutes != 246 {
		t.Errorf("total = %d minutes, want 246", snap.TotalMinutes)
	}
	if snap.Landings != 11 {
		t.Errorf("landings = %d, want 11", snap.Landings)
	}

	manifest := read(t, filepath.Join(out, "MANIFEST.txt"))
	for _, want := range []string{"flights", "246", "sha256", snap.SHA256DB} {
		if !strings.Contains(manifest, want) {
			t.Errorf("MANIFEST.txt does not mention %q:\n%s", want, manifest)
		}
	}
}

// The checksum in the manifest must be the checksum of the file that shipped,
// or it proves nothing on the day it is needed.
func TestManifestChecksumMatchesTheFileOnDisk(t *testing.T) {
	snap, out := run(t, live(t))

	if got := backup.SHA256File(filepath.Join(out, "logbook.db")); got != snap.SHA256DB {
		t.Errorf("logbook.db hashes to %s, manifest says %s", got, snap.SHA256DB)
	}
	if got := backup.SHA256File(filepath.Join(out, "logbook.csv")); got != snap.SHA256CSV {
		t.Errorf("logbook.csv hashes to %s, manifest says %s", got, snap.SHA256CSV)
	}
}

// Same data, same bytes. This runs daily and commits to git: a CSV that
// reordered itself would produce a diff every night and bury the day something
// actually changed.
func TestBackupIsDeterministic(t *testing.T) {
	db := live(t)
	first, _ := run(t, db)
	second, _ := run(t, db)

	if first.SHA256CSV != second.SHA256CSV {
		t.Error("two backups of unchanged data produced different CSVs")
	}
}

// Overwriting in place is the normal case -- the backup repo holds ONE current
// copy, and git holds the history. A run that refused because yesterday's file
// was there would break on its second day.
func TestBackupOverwritesThepreviousOne(t *testing.T) {
	db := live(t)
	out := t.TempDir()

	if _, err := backup.Run(db, out, func() time.Time { return fixedNow }); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if _, err := backup.Run(db, out, func() time.Time { return fixedNow }); err != nil {
		t.Fatalf("second run must overwrite, not refuse: %v", err)
	}
}

// A backup that changes what it is backing up would be a defect of the worst
// kind -- and RedactForBackup is one wrong argument away from doing exactly
// that.
func TestBackupLeavesTheLiveDatabaseAlone(t *testing.T) {
	db := live(t)
	users, err := db.Users()
	if err != nil {
		t.Fatal(err)
	}
	before, err := db.Sessions(users[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	run(t, db)

	after, err := db.Sessions(users[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) || len(after) != 1 {
		t.Errorf("the live session count went from %d to %d; the backup signed "+
			"the owner out of their own logbook", len(before), len(after))
	}
	if n, err := db.CountFlights(); err != nil || n != 3 {
		t.Errorf("the live flight count is now %d (%v), want 3", n, err)
	}
}

// The CSV is a complete dump, not a summary: it must carry the provenance and
// the raw times, or it cannot stand in for the database it came from.
func TestCSVCarriesEveryStoredField(t *testing.T) {
	_, out := run(t, live(t))
	csv := read(t, filepath.Join(out, "logbook.csv"))
	header, _, _ := strings.Cut(csv, "\n")

	for _, col := range []string{
		"seq", "source_book", "source_row", "flight_date",
		"aircraft_reg", "aircraft_type", "class", "dep_place", "arr_place",
		"off_block_utc", "on_block_utc", "off_block_raw", "on_block_raw",
		"time_origin", "takeoff_utc", "landing_utc",
		"block_minutes", "total_minutes", "night_minutes", "instrument_minutes",
		"pic_minutes", "dual_minutes", "instructor_minutes",
		"pic_name", "landings_day", "landings_night", "landings_verified", "remarks",
	} {
		if !strings.Contains(header, col) {
			t.Errorf("the CSV header is missing %q: %s", col, header)
		}
	}

	// One header plus one line per flight, and nothing else.
	lines := strings.Split(strings.TrimRight(csv, "\n"), "\n")
	if len(lines) != 4 {
		t.Errorf("CSV has %d lines, want 1 header + 3 flights", len(lines))
	}

	// The raw string as written on paper survives, not only the converted
	// instant (rule 0.4: never lose the source).
	if !strings.Contains(csv, "18:13") || !strings.Contains(csv, "converted_from_local") {
		t.Error("the CSV lost the raw time or its origin")
	}
}

// Book order, which is the only order anything cumulative may be rebuilt from.
func TestCSVIsInBookOrder(t *testing.T) {
	_, out := run(t, live(t))
	csv := read(t, filepath.Join(out, "logbook.csv"))
	lines := strings.Split(strings.TrimRight(csv, "\n"), "\n")

	// seq is the first column: 1, 2, then the hand-entered band at 1000000.
	for i, wantSeq := range []string{"1,", "2,", "1000000,"} {
		if !strings.HasPrefix(lines[i+1], wantSeq) {
			t.Errorf("line %d starts %q, want seq %s", i+1, lines[i+1][:12], wantSeq)
		}
	}
}

// RESTORE.md travels WITH the data. Instructions that live only in the
// application repository are instructions you do not have on the day both the
// server and your memory of it are gone.
func TestRestoreInstructionsTravelWithTheData(t *testing.T) {
	snap, out := run(t, live(t))
	doc := read(t, filepath.Join(out, "RESTORE.md"))

	for _, want := range []string{
		"logbook.db",       // what to put back
		"/var/lib/logbook", // where it goes
		"logbook",          // the service user
		"verify",           // how to check it
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("RESTORE.md never mentions %q", want)
		}
	}
	// It states the figures this copy should restore to, so a restore is
	// checkable rather than hoped at.
	if !strings.Contains(doc, "3") || !strings.Contains(doc, snap.SHA256DB[:12]) {
		t.Error("RESTORE.md does not state the figures to check the restore against")
	}
}

func TestRunRefusesAMissingOutputDirectory(t *testing.T) {
	db := live(t)
	_, err := backup.Run(db, filepath.Join(t.TempDir(), "nope"), func() time.Time { return fixedNow })
	if err == nil {
		t.Error("backing up into a directory that does not exist must fail loudly")
	}
}

// The 30 landings_unverified rows are a permanently open data question that the
// UI is required to keep showing (CLAUDE.md rule 0.8). A backup that flattened
// the flag would quietly turn "inferred" into "checked" in the only copy that
// survives the server.
func TestCSVKeepsTheUnverifiedLandingFlag(t *testing.T) {
	db := live(t)
	unverified := csvbook.Flight{
		Date: "2026-08-01", AircraftType: "C172", AircraftReg: "OH-CAV",
		Class: csvbook.ClassSEPLand, DepPlace: "EFHV", ArrPlace: "EFHV",
		OffBlockUTC: time.Date(2026, 8, 1, 20, 0, 0, 0, time.UTC),
		OnBlockUTC:  time.Date(2026, 8, 1, 21, 0, 0, 0, time.UTC),
		OffBlockRaw: "20:00Z", OnBlockRaw: "21:00Z", TimeOrigin: "utc_as_written",
		BlockMinutes: 60, TotalMinutes: 60, PICMinutes: 60, NightMinutes: 30,
		PICName: "self", LandingsNight: 2, LandingsVerified: false,
	}
	if _, err := db.AddFlight(unverified); err != nil {
		t.Fatal(err)
	}

	_, out := run(t, db)
	csv := read(t, filepath.Join(out, "logbook.csv"))
	for _, line := range strings.Split(csv, "\n") {
		if strings.Contains(line, "OH-CAV") {
			// ...,landings_day,landings_night,landings_verified,remarks
			if !strings.Contains(line, ",0,2,0,") {
				t.Errorf("the unverified flag did not survive into the CSV: %s", line)
			}
			return
		}
	}
	t.Error("the flight is missing from the CSV entirely")
}

func TestSHA256FileOnSomethingUnreadable(t *testing.T) {
	if got := backup.SHA256File(filepath.Join(t.TempDir(), "not-there")); got != "" {
		t.Errorf("SHA256File on a missing file = %q, want empty", got)
	}
}

// Pointing -out at a file rather than a directory must fail before anything is
// read, not part-way through.
func TestRunRefusesAnOutputThatIsNotADirectory(t *testing.T) {
	file := filepath.Join(t.TempDir(), "a-file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := backup.Run(live(t), file, func() time.Time { return fixedNow }); err == nil {
		t.Error("backing up into a plain file must fail")
	}
}

// A real operational failure: the backup directory exists but the service user
// cannot write to it. It must fail loudly at the first step rather than leave a
// partial copy where yesterday's good one was.
func TestRunRefusesADirectoryItCannotWriteTo(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores the permission bits this test relies on")
	}
	out := t.TempDir()
	if err := os.Chmod(out, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(out, 0o700) })

	if _, err := backup.Run(live(t), out, func() time.Time { return fixedNow }); err == nil {
		t.Error("a read-only backup directory must fail the run")
	}
}
