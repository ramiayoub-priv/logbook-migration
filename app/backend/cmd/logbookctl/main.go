// Command logbookctl is the operator's tool for the logbook database.
//
// Today it does one thing: import the three transcribed CSVs into SQLite,
// verified. It is deliberately a separate binary from the server, so that a
// destructive operation on a legal record can never be triggered by an HTTP
// request.
//
//	logbookctl import -db /var/lib/logbook/logbook.db -csv /path/to/repo
//	logbookctl verify -db /var/lib/logbook/logbook.db -csv /path/to/repo
//
// Import always takes a backup first, always verifies before it commits, and
// always prints the discrepancy report. See CLAUDE.md rule 0.2.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/backup"
	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/hhmm"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "logbookctl:", err)
		os.Exit(1)
	}
}

const usage = `logbookctl <command> [flags]

Commands:
  import   Load the logbook CSVs into the database, verified. Backs up first.
  verify   Re-check an existing database against the CSVs. Writes nothing.
  backup   Write a verified off-box copy (database + CSV + manifest + restore
           instructions) into a directory. Safe to run against a live server.
  check    Check a RESTORED database against a backup's MANIFEST.txt. Needs no
           CSVs and no sqlite3. Writes nothing. This is the restore check.
`

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(out, usage)
		return fmt.Errorf("no command given")
	}
	switch args[0] {
	case "import":
		return runImport(args[1:], out)
	case "verify":
		return runVerify(args[1:], out)
	case "backup":
		return runBackup(args[1:], out)
	case "check":
		return runCheck(args[1:], out)
	case "-h", "--help", "help":
		fmt.Fprint(out, usage)
		return nil
	default:
		fmt.Fprint(out, usage)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

// flagSet declares the two flags both commands share.
func flagSet(name string) (*flag.FlagSet, *string, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	db := fs.String("db", "", "path to the SQLite database (required)")
	csv := fs.String("csv", ".", "directory holding the three logbook CSVs")
	return fs, db, csv
}

func runImport(args []string, out io.Writer) error {
	fs, dbPath, csvDir := flagSet("import")
	backupDir := fs.String("backup-dir", "",
		"where to write the pre-import backup (default: alongside the database)")
	note := fs.String("note", "", "free text recorded in the import_runs audit trail")
	dryRun := fs.Bool("dry-run", false,
		"parse and report, but write nothing. Use this first.")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dbPath == "" && !*dryRun {
		return fmt.Errorf("-db is required")
	}

	lb, err := csvbook.Load(csvbook.DefaultSources(*csvDir))
	if err != nil {
		return err
	}
	report(out, lb)

	if *dryRun {
		fmt.Fprintf(out, "\nDry run: nothing was written.\n")
		return nil
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Back up before touching anything (rule 0.2). A first import has nothing
	// to lose, but taking the backup unconditionally keeps the procedure the
	// same on dev and on prod, which is what makes it reliable.
	dir := *backupDir
	if dir == "" {
		dir = filepath.Dir(*dbPath)
	}
	stamp := time.Now().UTC().Format("20060102T150405Z")
	backup := filepath.Join(dir, fmt.Sprintf("%s.%s.bak",
		filepath.Base(*dbPath), stamp))
	if err := db.Backup(backup); err != nil {
		return err
	}
	db.NoteBackup(backup)
	fmt.Fprintf(out, "\nBacked up to %s\n", backup)

	res, err := db.Import(lb, *note)
	if err != nil {
		return err
	}

	linked, unlinked, err := db.AircraftLinkage()
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Imported %d flights, %d aircraft, %d discrepancies into %s\n",
		res.Flights, res.Aircraft, res.Discrepancies, *dbPath)
	fmt.Fprintf(out, "Aircraft linkage: %d linked, %d unlinked\n", linked, unlinked)
	fmt.Fprintf(out, "Verified: every checksum matches the source CSVs.\n")
	return nil
}

func runVerify(args []string, out io.Writer) error {
	fs, dbPath, csvDir := flagSet("verify")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dbPath == "" {
		return fmt.Errorf("-db is required")
	}

	lb, err := csvbook.Load(csvbook.DefaultSources(*csvDir))
	if err != nil {
		return err
	}
	db, err := store.Open(*dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Verify(lb.Totals); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s matches the CSVs on all nine checksums:\n%s",
		*dbPath, totalsBlock(lb.Totals))
	return nil
}

// runBackup writes the off-box copy that Task 14 exists for.
//
// It reads the database and writes only into -out, so it is safe to run against
// a live server -- VACUUM INTO is transactionally consistent in WAL mode and the
// service never has to stop. That matters: a backup that requires downtime is a
// backup that gets skipped.
//
// It is here in logbookctl rather than in the server for the same reason import
// is: nothing that copies the whole legal record off the machine should be
// reachable over HTTP.
func runBackup(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	dbPath := fs.String("db", "", "path to the SQLite database (required)")
	outDir := fs.String("out", "", "directory to write the backup into (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dbPath == "" {
		return fmt.Errorf("-db is required")
	}
	if *outDir == "" {
		return fmt.Errorf("-out is required")
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	snap, err := backup.Run(db, *outDir, time.Now)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Backed up %d flights (%d entered in the app) to %s\n",
		snap.Flights, snap.HandEntered, *outDir)
	fmt.Fprintf(out, "  total       %s (%d minutes)\n",
		hhmm.Format(snap.TotalMinutes), snap.TotalMinutes)
	fmt.Fprintf(out, "  landings    %d\n", snap.Landings)
	fmt.Fprintf(out, "  aircraft    %d\n", snap.Aircraft)
	fmt.Fprintf(out, "  users       %d (%d sessions dropped, not carried off the box)\n",
		snap.Users, snap.SessionsDropped)
	fmt.Fprintf(out, "  sha256      %s  %s\n", snap.SHA256DB, backup.DBName)
	fmt.Fprintf(out, "  sha256      %s  %s\n", snap.SHA256CSV, backup.CSVName)
	fmt.Fprintf(out, "Verified against the live database before writing.\n")
	return nil
}

// runCheck verifies a RESTORED database against the manifest of the backup it
// came from. It writes nothing.
//
// WHY THIS IS NOT `verify`. Verify compares a database against the three
// transcribed CSVs, and its checksums are scoped `source_book <> 0`. On a
// restored server that is both unavailable (the books may not be there) and
// insufficient (it would pass while every app-entered flight was missing --
// precisely the rows that exist nowhere else and that the backup exists for).
// Check asks the only question a restore raises: is this the data the backup
// recorded?
//
// It needs nothing but the database file and a manifest. RESTORE.md used to
// send the reader to `sqlite3` here, which is not installed on the server and
// is not a dependency of this project, so the mandatory verification step of a
// restore was `command not found` on the day it was needed.
func runCheck(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	dbPath := fs.String("db", "", "path to the restored SQLite database (required)")
	manifestPath := fs.String("manifest", "",
		"path to the backup's MANIFEST.txt to check against")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dbPath == "" {
		return fmt.Errorf("-db is required")
	}

	// The file's own checksum first, BEFORE anything opens it. sha256 is the
	// strongest statement available -- a file that hashes correctly is the
	// backup, byte for byte -- and it is the one check that costs nothing.
	sum := backup.SHA256File(*dbPath)

	var want backup.Snapshot
	haveManifest := *manifestPath != ""
	if haveManifest {
		text, err := os.ReadFile(*manifestPath)
		if err != nil {
			return fmt.Errorf("reading the manifest: %w", err)
		}
		if want, err = backup.ParseManifest(string(text)); err != nil {
			return err
		}
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	got, err := backup.Figures(db)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "%s\n", *dbPath)
	fmt.Fprintf(out, "  %-16s %d\n", "flights", got.Flights)
	fmt.Fprintf(out, "  %-16s %d   (source_book = 0; these exist in NO CSV)\n",
		"hand-entered", got.HandEntered)
	fmt.Fprintf(out, "  %-16s %d\n", "aircraft", got.Aircraft)
	fmt.Fprintf(out, "  %-16s %d\n", "discrepancies", got.Discrepancies)
	fmt.Fprintf(out, "  %-16s %d\n", "users", got.Users)
	fmt.Fprintf(out, "  %-16s %s   (%d minutes)\n",
		"total time", hhmm.Format(got.TotalMinutes), got.TotalMinutes)
	fmt.Fprintf(out, "  %-16s %d\n", "landings", got.Landings)
	fmt.Fprintf(out, "  %-16s %s\n", "sha256", sum)

	if !haveManifest {
		// Never let this read as a pass. It reported figures; it compared them
		// to nothing.
		fmt.Fprintf(out, "\nNo -manifest given, so nothing was compared. These are the figures\n"+
			"this database holds; whether they are the RIGHT ones is unproven.\n")
		return nil
	}

	if sum != "" && want.SHA256DB != "" && sum != want.SHA256DB {
		fmt.Fprintf(out, "\n  !! sha256 does not match the manifest\n"+
			"     manifest %s\n     file     %s\n", want.SHA256DB, sum)
	}

	bad := backup.Compare(want, got)
	if len(bad) > 0 {
		fmt.Fprintf(out, "\nTHIS RESTORE DOES NOT MATCH ITS MANIFEST:\n")
		for _, m := range bad {
			fmt.Fprintf(out, "  %-16s manifest %s, database %s\n", m.Field, m.Want, m.Got)
		}
		// Rule 0.2: a legal record that disagrees with its own record of itself
		// is not something to work around.
		return fmt.Errorf("REFUSING: %d figure(s) disagree with %s (%v) -- "+
			"do not let the application write to this file until you know why",
			len(bad), *manifestPath, bad)
	}

	if sum != want.SHA256DB {
		return fmt.Errorf("REFUSING: every figure matches but the file's sha256 does not "+
			"match the manifest (%s vs %s)", sum, want.SHA256DB)
	}
	fmt.Fprintf(out, "\nEvery figure matches %s, taken %s.\n",
		*manifestPath, want.At.Format(time.RFC3339))
	return nil
}

// report prints what was read and everything the source data disagrees about.
// The discrepancy list is the point of the exercise: it is what the owner rules
// on (rule 0.2), so it is printed in full rather than summarised away.
func report(out io.Writer, lb *csvbook.Logbook) {
	fmt.Fprintf(out, "Read %d flights from %d books.\n\n", lb.Totals.Flights, 3)
	fmt.Fprint(out, totalsBlock(lb.Totals))

	if len(lb.Discrepancies) == 0 {
		fmt.Fprintf(out, "\nNo discrepancies.\n")
		return
	}

	byKind := map[csvbook.Kind]int{}
	for _, d := range lb.Discrepancies {
		byKind[d.Kind]++
	}
	kinds := make([]string, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, string(k))
	}
	sort.Strings(kinds)

	fmt.Fprintf(out, "\n%d discrepancies. NOTHING HAS BEEN CORRECTED -- these are for you to rule on.\n",
		len(lb.Discrepancies))
	for _, k := range kinds {
		fmt.Fprintf(out, "  %-24s %d\n", k, byKind[csvbook.Kind(k)])
	}

	fmt.Fprintln(out)
	for _, d := range lb.Discrepancies {
		fmt.Fprintf(out, "  book %d line %-4d %-11s %-24s %s\n",
			d.Book, d.Row, d.Date, d.Kind, d.Detail)
	}
}

func totalsBlock(t csvbook.Totals) string {
	var b strings.Builder
	fmt.Fprintf(&b, "  %-16s %d\n", "flights", t.Flights)
	for _, row := range []struct {
		name string
		min  int
	}{
		{"total", t.Total},
		{"pic", t.PIC},
		{"dual", t.Dual},
		{"instrument", t.Instrument},
		{"night", t.Night},
		{"instructor", t.Instructor},
		{"seaplane", t.SEPSea},
	} {
		fmt.Fprintf(&b, "  %-16s %s\n", row.name, hhmm.Format(row.min))
	}
	fmt.Fprintf(&b, "  %-16s %d\n", "landings", t.Landings)
	return b.String()
}
