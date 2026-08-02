package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/backup"
	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

// repoRoot is where the three logbook CSVs live.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	// .../app/backend/cmd/logbookctl/main_test.go -> repo root
	return filepath.Clean(filepath.Join(filepath.Dir(self), "..", "..", "..", ".."))
}

func TestImportEndToEnd(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "logbook.db")

	var out bytes.Buffer
	err := run([]string{"import", "-db", dbPath, "-csv", repoRoot(t), "-note", "test"}, &out)
	if err != nil {
		t.Fatalf("import: %v\n%s", err, out.String())
	}

	got := out.String()
	for _, want := range []string{
		"Read 1296 flights",
		// 54, was 61: the owner ruled the "C192" typo and OH-CMU's type on
		// 2026-08-02, which closed unknown_aircraft_type (4) and type_conflict
		// (3) together. No time, landing or class figure moved with them --
		// sea/land comes from the registration, never from the type.
		"Imported 1296 flights, 38 aircraft, 54 discrepancies",
		"Aircraft linkage: 1296 linked, 0 unlinked",
		"Verified: every checksum matches the source CSVs.",
		// The report must say plainly that nothing was fixed, every time.
		"NOTHING HAS BEEN CORRECTED",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output is missing %q\n---\n%s", want, got)
		}
	}

	// A backup must exist before the import touched anything.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var backups int
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".bak") {
			backups++
		}
	}
	if backups != 1 {
		t.Errorf("found %d backups, want exactly 1", backups)
	}

	// And verify must agree afterwards.
	out.Reset()
	if err := run([]string{"verify", "-db", dbPath, "-csv", repoRoot(t)}, &out); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !strings.Contains(out.String(), "matches the CSVs on all nine checksums") {
		t.Errorf("verify output:\n%s", out.String())
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "logbook.db")

	var out bytes.Buffer
	if err := run([]string{"import", "-dry-run", "-db", dbPath, "-csv", repoRoot(t)}, &out); err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if !strings.Contains(out.String(), "Dry run: nothing was written.") {
		t.Errorf("output:\n%s", out.String())
	}
	if entries, err := os.ReadDir(dir); err != nil {
		t.Fatal(err)
	} else if len(entries) != 0 {
		t.Errorf("a dry run created %d files, want none", len(entries))
	}
}

func TestVerifyFailsAgainstAnEmptyDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "empty.db")
	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	var out bytes.Buffer
	err = run([]string{"verify", "-db", dbPath, "-csv", repoRoot(t)}, &out)
	if err == nil {
		t.Fatal("verifying an empty database against 1296 flights must fail")
	}
	if !strings.Contains(err.Error(), "flight count") {
		t.Errorf("err = %v, want it to name the failing checksum", err)
	}
}

// --- check: the command a restore is verified with --------------------------
//
// It exists because RESTORE.md's verification step used to be a `sqlite3`
// invocation, and sqlite3 is not installed on the server and is not a
// dependency of this project. See internal/backup for the full argument.

// backedUp imports the real books, adds the kind of flight that exists in no
// CSV, and takes a backup of it -- the exact artefact a restore starts from.
func backedUp(t *testing.T) (dbPath, backupDir string) {
	t.Helper()
	dir := t.TempDir()
	dbPath = filepath.Join(dir, "logbook.db")

	var out bytes.Buffer
	if err := run([]string{"import", "-db", dbPath, "-csv", repoRoot(t)}, &out); err != nil {
		t.Fatalf("import: %v\n%s", err, out.String())
	}

	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.AddFlight(csvbook.Flight{
		Date: "2026-08-02", AircraftType: "P28A", AircraftReg: "OH-PDP",
		Class: csvbook.ClassSEPLand, DepPlace: "EFHV", ArrPlace: "EFRY",
		OffBlockUTC: time.Date(2026, 8, 2, 11, 18, 0, 0, time.UTC),
		OnBlockUTC:  time.Date(2026, 8, 2, 11, 49, 0, 0, time.UTC),
		OffBlockRaw: "14:18", OnBlockRaw: "14:49", TimeOrigin: "converted_from_local",
		BlockMinutes: 31, TotalMinutes: 31, PICMinutes: 31,
		PICName: "self", LandingsDay: 1, LandingsVerified: true,
	}); err != nil {
		t.Fatal(err)
	}
	db.Close()

	backupDir = t.TempDir()
	out.Reset()
	if err := run([]string{"backup", "-db", dbPath, "-out", backupDir}, &out); err != nil {
		t.Fatalf("backup: %v\n%s", err, out.String())
	}
	return dbPath, backupDir
}

func TestCheckPassesOnAFreshlyRestoredBackup(t *testing.T) {
	_, backupDir := backedUp(t)

	var out bytes.Buffer
	err := run([]string{"check",
		"-db", filepath.Join(backupDir, backup.DBName),
		"-manifest", filepath.Join(backupDir, backup.ManifestName)}, &out)
	if err != nil {
		t.Fatalf("check on a good restore: %v\n%s", err, out.String())
	}

	got := out.String()
	// It must PRINT the figures, not merely bless them. The operator is
	// standing in front of a restored legal record and needs to see them.
	for _, want := range []string{"1297", "flights", "landings", "total time", "users"} {
		if !strings.Contains(got, want) {
			t.Errorf("check output is missing %q\n---\n%s", want, got)
		}
	}
	if !strings.Contains(got, "matches") {
		t.Errorf("check does not say plainly that it matched\n---\n%s", got)
	}
}

// The hand-entered flight is the whole reason the backup exists, so check has
// to account for it by name -- 1296 came from the CSVs, the 1297th did not.
func TestCheckCountsTheFlightThatExistsNowhereElse(t *testing.T) {
	_, backupDir := backedUp(t)

	var out bytes.Buffer
	if err := run([]string{"check",
		"-db", filepath.Join(backupDir, backup.DBName),
		"-manifest", filepath.Join(backupDir, backup.ManifestName)}, &out); err != nil {
		t.Fatalf("check: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "hand-entered") {
		t.Errorf("check never mentions the hand-entered flights, which are the only "+
			"rows in the file that exist nowhere else\n---\n%s", out.String())
	}
}

func TestCheckRefusesADatabaseThatDisagreesWithItsManifest(t *testing.T) {
	dbPath, backupDir := backedUp(t)

	// The live database has since grown; its manifest describes the snapshot.
	// Checking the wrong file against a manifest is exactly the mistake this
	// command exists to catch -- a restore from the wrong day looks fine.
	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.AddFlight(csvbook.Flight{
		Date: "2026-08-03", AircraftType: "P28A", AircraftReg: "OH-PDP",
		Class: csvbook.ClassSEPLand, DepPlace: "EFRY", ArrPlace: "EFHV",
		OffBlockUTC: time.Date(2026, 8, 3, 12, 28, 0, 0, time.UTC),
		OnBlockUTC:  time.Date(2026, 8, 3, 12, 50, 0, 0, time.UTC),
		OffBlockRaw: "15:28", OnBlockRaw: "15:50", TimeOrigin: "converted_from_local",
		BlockMinutes: 22, TotalMinutes: 22, PICMinutes: 22,
		PICName: "self", LandingsDay: 1, LandingsVerified: true,
	}); err != nil {
		t.Fatal(err)
	}
	db.Close()

	var out bytes.Buffer
	err = run([]string{"check", "-db", dbPath,
		"-manifest", filepath.Join(backupDir, backup.ManifestName)}, &out)
	if err == nil {
		t.Fatal("check accepted a database holding a flight its manifest does not know about")
	}
	// The operator must be told WHICH figures disagree and by how much, or the
	// only available response is to guess.
	for _, want := range []string{"flights", "1297", "1298"} {
		if !strings.Contains(err.Error()+out.String(), want) {
			t.Errorf("check's refusal never mentions %q\nerr: %v\n---\n%s", want, err, out.String())
		}
	}
}

// A manifest it cannot read must stop the check, never silently compare
// against a pile of zeroes.
func TestCheckRefusesAManifestItCannotRead(t *testing.T) {
	_, backupDir := backedUp(t)
	junk := filepath.Join(t.TempDir(), "MANIFEST.txt")
	if err := os.WriteFile(junk, []byte("this is not a manifest\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := run([]string{"check",
		"-db", filepath.Join(backupDir, backup.DBName), "-manifest", junk}, &out)
	if err == nil {
		t.Fatal("check accepted a manifest it could not parse")
	}
}

// check must need the database and NOTHING else -- no CSVs. A restored server
// has the data; it does not necessarily have the three source books.
func TestCheckNeedsNoCSVs(t *testing.T) {
	_, backupDir := backedUp(t)

	var out bytes.Buffer
	if err := run([]string{"check",
		"-db", filepath.Join(backupDir, backup.DBName),
		"-manifest", filepath.Join(backupDir, backup.ManifestName)}, &out); err != nil {
		t.Fatalf("check: %v", err)
	}
	// -csv is not even a flag it accepts, so nobody can come to depend on it.
	out.Reset()
	if err := run([]string{"check", "-db", "x.db", "-csv", repoRoot(t)}, &out); err == nil {
		t.Error("check accepts a -csv flag; it must not, or the restore check " +
			"grows a dependency on books a restored server may not have")
	}
}

// Without -manifest it still reports the figures, which is what you want when
// the manifest is lost but the database survived.
func TestCheckWithoutAManifestStillReportsTheFigures(t *testing.T) {
	_, backupDir := backedUp(t)

	var out bytes.Buffer
	if err := run([]string{"check", "-db", filepath.Join(backupDir, backup.DBName)}, &out); err != nil {
		t.Fatalf("check without -manifest: %v\n%s", err, out.String())
	}
	got := out.String()
	if !strings.Contains(got, "1297") {
		t.Errorf("check printed no flight count\n---\n%s", got)
	}
	// And it must say it proved nothing, rather than looking like a pass.
	if !strings.Contains(got, "nothing was compared") {
		t.Errorf("check without a manifest must not read as a successful "+
			"verification\n---\n%s", got)
	}
}

func TestBadInvocations(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"no command", nil, "no command"},
		{"unknown command", []string{"frobnicate"}, "unknown command"},
		{"import without -db", []string{"import", "-csv", "."}, "-db is required"},
		{"verify without -db", []string{"verify", "-csv", "."}, "-db is required"},
		{"import with a bad flag", []string{"import", "-nope"}, "flag provided but not defined"},
		{"verify with a bad flag", []string{"verify", "-nope"}, "flag provided but not defined"},
		{"missing csv directory", []string{"import", "-dry-run", "-csv", "/nonexistent"}, "book 1"},
		{"verify against a missing csv directory",
			[]string{"verify", "-db", "x.db", "-csv", "/nonexistent"}, "book 1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			err := run(c.args, &out)
			if err == nil {
				t.Fatalf("got no error, want one mentioning %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to mention %q", err, c.want)
			}
		})
	}
}

func TestHelp(t *testing.T) {
	for _, arg := range []string{"-h", "--help", "help"} {
		var out bytes.Buffer
		if err := run([]string{arg}, &out); err != nil {
			t.Errorf("%s: %v", arg, err)
		}
		if !strings.Contains(out.String(), "logbookctl <command>") {
			t.Errorf("%s printed no usage", arg)
		}
	}
}

func TestImportRefusesToOverwriteADatabaseItCannotBackUp(t *testing.T) {
	// The backup is the reversibility half of rule 0.2. If it cannot be
	// written, the import must not proceed.
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "logbook.db")

	var out bytes.Buffer
	err := run([]string{"import", "-db", dbPath, "-csv", repoRoot(t),
		"-backup-dir", filepath.Join(dir, "no-such-dir")}, &out)
	if err == nil {
		t.Fatal("an unwritable backup path must abort the import")
	}

	db, oerr := store.Open(dbPath)
	if oerr != nil {
		t.Fatal(oerr)
	}
	defer db.Close()
	if n, _ := db.CountFlights(); n != 0 {
		t.Errorf("%d flights were written despite the failed backup", n)
	}
}

func TestReportOnACleanLogbookSaysSoExplicitly(t *testing.T) {
	// Silence is the wrong way to report "no problems found" on a legal
	// record: the operator has to be able to tell it apart from a report that
	// was never produced.
	var out bytes.Buffer
	report(&out, &csvbook.Logbook{Totals: csvbook.Totals{Flights: 0}})

	if !strings.Contains(out.String(), "No discrepancies.") {
		t.Errorf("output:\n%s", out.String())
	}
}
