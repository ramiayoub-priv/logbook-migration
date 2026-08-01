package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

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
		"Read 1293 flights",
		"Imported 1293 flights, 39 aircraft, 56 discrepancies",
		"Aircraft linkage: 1293 linked, 0 unlinked",
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
		t.Fatal("verifying an empty database against 1293 flights must fail")
	}
	if !strings.Contains(err.Error(), "flight count") {
		t.Errorf("err = %v, want it to name the failing checksum", err)
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
