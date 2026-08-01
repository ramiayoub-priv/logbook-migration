package store_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

func openTemp(t *testing.T) *store.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "logbook.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// sample is a two-flight logbook: one seaplane instructing row and one dual
// row, with an aircraft list and one discrepancy.
func sample(t *testing.T) *csvbook.Logbook {
	t.Helper()
	flights := []csvbook.Flight{
		{
			Seq: 1, Date: "2021-06-01",
			AircraftType: "C172", AircraftReg: "OH-CTL", Class: csvbook.ClassSEPSea,
			DepPlace: "Tuusulanjärvi", ArrPlace: "Tuusulanjärvi",
			OffBlockUTC: mustTime(t, "2021-06-01T15:13:00Z"),
			OnBlockUTC:  mustTime(t, "2021-06-01T16:34:00Z"),
			OffBlockRaw: "18:13", OnBlockRaw: "19:34",
			TimeOrigin:   "converted_from_local",
			BlockMinutes: 81, TotalMinutes: 81, PICMinutes: 81, InstructorMinutes: 81,
			PICName: "self", LandingsDay: 7, LandingsVerified: true,
			SourceBook: 3, SourceRow: 3,
		},
		{
			Seq: 2, Date: "2021-06-02",
			AircraftType: "P28A", AircraftReg: "OH-PDP", Class: csvbook.ClassSEPLand,
			DepPlace: "EFHV", ArrPlace: "EFHV",
			OffBlockUTC: mustTime(t, "2021-06-02T06:00:00Z"),
			OnBlockUTC:  mustTime(t, "2021-06-02T07:30:00Z"),
			TakeoffUTC:  mustTime(t, "2021-06-02T06:05:00Z"),
			LandingUTC:  mustTime(t, "2021-06-02T07:25:00Z"),
			OffBlockRaw: "06:00Z", OnBlockRaw: "07:30Z",
			TimeOrigin:   "utc_as_written",
			BlockMinutes: 90, TotalMinutes: 90, DualMinutes: 90,
			NightMinutes: 10, InstrumentMinutes: 20,
			PICName: "Autere", LandingsDay: 3, LandingsVerified: false,
			Remarks:    "night circuits",
			SourceBook: 3, SourceRow: 4,
		},
	}
	lb := &csvbook.Logbook{
		Flights: flights,
		Aircraft: []csvbook.Aircraft{
			{Registration: "OH-CTL", Type: "C172", DefaultClass: csvbook.ClassSEPSea, Active: true},
			{Registration: "OH-PDP", Type: "P28A", DefaultClass: csvbook.ClassSEPLand, Active: true, Notes: "n"},
		},
		Discrepancies: []csvbook.Discrepancy{
			{Kind: csvbook.KindLandingsUnverified, Book: 3, Row: 4, Date: "2021-06-02", Detail: "d"},
		},
	}
	for _, f := range flights {
		lb.Totals.Flights++
		lb.Totals.Total += f.TotalMinutes
		lb.Totals.PIC += f.PICMinutes
		lb.Totals.Dual += f.DualMinutes
		lb.Totals.Instrument += f.InstrumentMinutes
		lb.Totals.Night += f.NightMinutes
		lb.Totals.Instructor += f.InstructorMinutes
		if f.Class == csvbook.ClassSEPSea {
			lb.Totals.SEPSea += f.TotalMinutes
		}
		lb.Totals.Landings += f.LandingsDay + f.LandingsNight
	}
	return lb
}

func TestOpenCreatesTheSchemaAndIsSafeToRepeat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logbook.db")

	db, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := db.Import(sample(t), "first"); err != nil {
		t.Fatalf("Import: %v", err)
	}
	db.Close()

	// Opening an existing database must not migrate, reset or lose anything.
	db2, err := store.Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db2.Close()

	n, err := db2.CountFlights()
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("flights after reopen = %d, want 2", n)
	}
}

func TestOpenRejectsAnUnusablePath(t *testing.T) {
	_, err := store.Open(filepath.Join(t.TempDir(), "no-such-dir", "logbook.db"))
	if err == nil {
		t.Fatal("opening under a missing directory must fail loudly")
	}
}

func TestImportWritesEveryFieldBack(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "test"); err != nil {
		t.Fatalf("Import: %v", err)
	}

	got, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	want := sample(t).Flights
	if len(got) != len(want) {
		t.Fatalf("read back %d flights, want %d", len(got), len(want))
	}
	for i := range want {
		if !got[i].OffBlockUTC.Equal(want[i].OffBlockUTC) {
			t.Errorf("flight %d off-block = %v, want %v", i, got[i].OffBlockUTC, want[i].OffBlockUTC)
		}
		// Compare the rest field-by-field via the struct, with the instants
		// normalised to UTC so a location difference is not a false failure.
		g, w := got[i], want[i]
		g.OffBlockUTC, w.OffBlockUTC = g.OffBlockUTC.UTC(), w.OffBlockUTC.UTC()
		g.OnBlockUTC, w.OnBlockUTC = g.OnBlockUTC.UTC(), w.OnBlockUTC.UTC()
		g.TakeoffUTC, w.TakeoffUTC = g.TakeoffUTC.UTC(), w.TakeoffUTC.UTC()
		g.LandingUTC, w.LandingUTC = g.LandingUTC.UTC(), w.LandingUTC.UTC()
		if g != w {
			t.Errorf("flight %d round-tripped as\n %+v\nwant\n %+v", i, g, w)
		}
	}
}

func TestImportLinksFlightsToAircraftAndToleratesAMissingOne(t *testing.T) {
	lb := sample(t)
	// Drop OH-PDP from the seed list. Its flights must still import, with a
	// null aircraft_id -- a flight is never lost because an aircraft row is.
	lb.Aircraft = lb.Aircraft[:1]

	db := openTemp(t)
	if _, err := db.Import(lb, "test"); err != nil {
		t.Fatalf("Import: %v", err)
	}

	linked, unlinked, err := db.AircraftLinkage()
	if err != nil {
		t.Fatal(err)
	}
	if linked != 1 || unlinked != 1 {
		t.Errorf("linkage = %d linked / %d unlinked, want 1/1", linked, unlinked)
	}
}

func TestImportIsIdempotent(t *testing.T) {
	db := openTemp(t)

	first, err := db.Import(sample(t), "first")
	if err != nil {
		t.Fatalf("first import: %v", err)
	}
	before, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}

	second, err := db.Import(sample(t), "second")
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	after, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}

	if first.Totals != second.Totals || first.Flights != second.Flights {
		t.Errorf("re-running changed the result:\n %+v\n %+v", first, second)
	}
	if len(before) != len(after) {
		t.Fatalf("re-running changed the row count: %d then %d", len(before), len(after))
	}
	for i := range before {
		if before[i].Seq != after[i].Seq || before[i].SourceRow != after[i].SourceRow {
			t.Errorf("flight %d moved between runs", i)
		}
	}

	// The audit trail is the one thing that legitimately grows.
	runs, err := db.ImportRuns()
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Errorf("import_runs = %d, want 2 -- every run must be recorded", len(runs))
	}
}

func TestImportReplacesEarlierDataRatherThanAppending(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "first"); err != nil {
		t.Fatal(err)
	}

	// A second, shorter logbook: the database must end up describing it and
	// nothing else, or a removed row would live on as a phantom total.
	lb := sample(t)
	lb.Flights = lb.Flights[:1]
	lb.Totals = csvbook.Totals{
		Flights: 1, Total: 81, PIC: 81, Instructor: 81, SEPSea: 81, Landings: 7,
	}
	lb.Discrepancies = nil
	if _, err := db.Import(lb, "second"); err != nil {
		t.Fatalf("second import: %v", err)
	}

	n, err := db.CountFlights()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("flights = %d, want 1", n)
	}
	ds, err := db.Discrepancies()
	if err != nil {
		t.Fatal(err)
	}
	if len(ds) != 0 {
		t.Errorf("discrepancies = %d, want 0 -- a resolved item must disappear", len(ds))
	}
}

// TestImportRefusesOnAChecksumMismatch is the rule-0.2 gate. The logbook below
// is internally inconsistent: its Totals claim more time than its flights
// carry. Nothing may be committed.
func TestImportRefusesOnAChecksumMismatch(t *testing.T) {
	db := openTemp(t)

	bad := sample(t)
	bad.Totals.Total += 1 // one minute is enough; there is no tolerance

	if _, err := db.Import(bad, "test"); err == nil {
		t.Fatal("a checksum mismatch must abort the import")
	} else if !strings.Contains(err.Error(), "total") {
		t.Errorf("err = %v, want it to name the failing checksum", err)
	}

	n, err := db.CountFlights()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("flights = %d after a refused import, want 0 -- it must roll back", n)
	}
	runs, err := db.ImportRuns()
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 0 {
		t.Errorf("a refused import must not be recorded as a run, got %d", len(runs))
	}
}

func TestImportRefusesOnEveryChecksum(t *testing.T) {
	// Each checksum is checked independently, so a compensating pair of errors
	// cannot slip through on a single grand total.
	cases := []struct {
		name string
		bend func(*csvbook.Totals)
		want string
	}{
		{"flight count", func(x *csvbook.Totals) { x.Flights++ }, "flight count"},
		{"total", func(x *csvbook.Totals) { x.Total++ }, "total"},
		{"pic", func(x *csvbook.Totals) { x.PIC++ }, "pic"},
		{"dual", func(x *csvbook.Totals) { x.Dual++ }, "dual"},
		{"instrument", func(x *csvbook.Totals) { x.Instrument++ }, "instrument"},
		{"night", func(x *csvbook.Totals) { x.Night++ }, "night"},
		{"instructor", func(x *csvbook.Totals) { x.Instructor++ }, "instructor"},
		{"seaplane", func(x *csvbook.Totals) { x.SEPSea++ }, "seaplane"},
		{"landings", func(x *csvbook.Totals) { x.Landings++ }, "landings"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := openTemp(t)
			bad := sample(t)
			c.bend(&bad.Totals)

			_, err := db.Import(bad, "test")
			if err == nil {
				t.Fatalf("bending %s must abort the import", c.name)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to mention %q", err, c.want)
			}
			if n, _ := db.CountFlights(); n != 0 {
				t.Errorf("rolled forward instead of back: %d flights", n)
			}
		})
	}
}

func TestImportRefusesADuplicatedSourceRow(t *testing.T) {
	// Provenance is unique by constraint. Two flights claiming the same CSV
	// line would mean a row was read twice, which the schema must reject.
	lb := sample(t)
	lb.Flights[1].SourceBook = lb.Flights[0].SourceBook
	lb.Flights[1].SourceRow = lb.Flights[0].SourceRow

	db := openTemp(t)
	if _, err := db.Import(lb, "test"); err == nil {
		t.Fatal("a duplicated source row must abort the import")
	}
	if n, _ := db.CountFlights(); n != 0 {
		t.Errorf("flights = %d, want 0", n)
	}
}

func TestImportRefusesAnInvalidClass(t *testing.T) {
	lb := sample(t)
	lb.Flights[0].Class = "SEAPLANE"

	db := openTemp(t)
	if _, err := db.Import(lb, "test"); err == nil {
		t.Fatal("the class vocabulary is fixed by a CHECK and must be enforced")
	}
}

func TestBackupCopiesTheDatabase(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "test"); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), "backup.db")
	if err := db.Backup(dest); err != nil {
		t.Fatalf("Backup: %v", err)
	}

	// A backup that cannot be opened and read is not a backup.
	restored, err := store.Open(dest)
	if err != nil {
		t.Fatalf("opening the backup: %v", err)
	}
	defer restored.Close()

	n, err := restored.CountFlights()
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("backup holds %d flights, want 2", n)
	}
}

func TestBackupRefusesToOverwrite(t *testing.T) {
	db := openTemp(t)
	dest := filepath.Join(t.TempDir(), "backup.db")
	if err := db.Backup(dest); err != nil {
		t.Fatal(err)
	}
	// Silently replacing an existing backup would destroy the only copy of the
	// state someone is trying to preserve.
	if err := db.Backup(dest); err == nil {
		t.Fatal("backing up over an existing file must fail")
	}
}

func TestDiscrepanciesRoundTrip(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "test"); err != nil {
		t.Fatal(err)
	}

	ds, err := db.Discrepancies()
	if err != nil {
		t.Fatal(err)
	}
	if len(ds) != 1 {
		t.Fatalf("got %d discrepancies, want 1", len(ds))
	}
	want := csvbook.Discrepancy{
		Kind: csvbook.KindLandingsUnverified, Book: 3, Row: 4, Date: "2021-06-02", Detail: "d",
	}
	if ds[0] != want {
		t.Errorf("got %+v, want %+v", ds[0], want)
	}
}

func TestAircraftRoundTrip(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "test"); err != nil {
		t.Fatal(err)
	}

	got, err := db.Aircraft()
	if err != nil {
		t.Fatal(err)
	}
	want := sample(t).Aircraft
	if len(got) != len(want) {
		t.Fatalf("got %d aircraft, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("aircraft %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestImportRunRecordsTheBackupPath(t *testing.T) {
	db := openTemp(t)
	dest := filepath.Join(t.TempDir(), "backup.db")
	if err := db.Backup(dest); err != nil {
		t.Fatal(err)
	}
	db.NoteBackup(dest)

	if _, err := db.Import(sample(t), "with backup"); err != nil {
		t.Fatal(err)
	}
	runs, err := db.ImportRuns()
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 {
		t.Fatalf("got %d runs, want 1", len(runs))
	}
	if runs[0].BackupPath != dest {
		t.Errorf("BackupPath = %q, want %q", runs[0].BackupPath, dest)
	}
	if runs[0].Note != "with backup" {
		t.Errorf("Note = %q", runs[0].Note)
	}
	if runs[0].Flights != 2 || runs[0].TotalMinutes != 171 || runs[0].Landings != 10 {
		t.Errorf("run = %+v, want the imported tally", runs[0])
	}
	if runs[0].RanAt == "" {
		t.Errorf("a run with no timestamp is not an audit trail")
	}
}

func TestUsingAClosedDatabaseIsAnError(t *testing.T) {
	db := openTemp(t)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Import(sample(t), "test"); err == nil {
		t.Error("Import on a closed database must fail")
	}
	if _, err := db.Flights(); err == nil {
		t.Error("Flights on a closed database must fail")
	}
	if _, err := db.Aircraft(); err == nil {
		t.Error("Aircraft on a closed database must fail")
	}
	if _, err := db.Discrepancies(); err == nil {
		t.Error("Discrepancies on a closed database must fail")
	}
	if _, err := db.ImportRuns(); err == nil {
		t.Error("ImportRuns on a closed database must fail")
	}
	if _, err := db.CountFlights(); err == nil {
		t.Error("CountFlights on a closed database must fail")
	}
	if _, _, err := db.AircraftLinkage(); err == nil {
		t.Error("AircraftLinkage on a closed database must fail")
	}
	if err := db.Backup(filepath.Join(t.TempDir(), "b.db")); err == nil {
		t.Error("Backup on a closed database must fail")
	}
}

func TestVerifyChecksACommittedDatabase(t *testing.T) {
	// The same gate Import applies, runnable against a live database so a
	// production file can be re-checked against the CSVs at any time.
	path := filepath.Join(t.TempDir(), "logbook.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if db.Path() != path {
		t.Errorf("Path = %q, want %q", db.Path(), path)
	}

	lb := sample(t)
	if _, err := db.Import(lb, "test"); err != nil {
		t.Fatal(err)
	}
	if err := db.Verify(lb.Totals); err != nil {
		t.Errorf("Verify on freshly imported data: %v", err)
	}

	bent := lb.Totals
	bent.SEPSea++
	if err := db.Verify(bent); err == nil {
		t.Error("Verify must fail against the wrong checksums")
	}
}

func TestEmptyLogbookImportsCleanly(t *testing.T) {
	// An empty range is a real case and must not be a special one.
	db := openTemp(t)
	res, err := db.Import(&csvbook.Logbook{}, "empty")
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if res.Flights != 0 || res.Totals != (csvbook.Totals{}) {
		t.Errorf("res = %+v, want an empty tally", res)
	}
}
