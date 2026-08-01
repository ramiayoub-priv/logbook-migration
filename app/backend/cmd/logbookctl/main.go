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
