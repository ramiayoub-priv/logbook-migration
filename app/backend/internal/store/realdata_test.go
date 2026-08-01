package store_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

// End-to-end over the real books. The unit tests above prove the store honours
// its contract; this proves the contract is enough to carry the actual legal
// record into SQLite without changing a minute of it.

func repoRoot(t *testing.T) string {
	t.Helper()
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(self), "..", "..", "..", ".."))
}

func loadRealBooks(t *testing.T) *csvbook.Logbook {
	t.Helper()
	lb, err := csvbook.Load(csvbook.DefaultSources(repoRoot(t)))
	if err != nil {
		t.Fatalf("loading the real books: %v", err)
	}
	return lb
}

func TestRealImportRoundTripsEveryFlight(t *testing.T) {
	lb := loadRealBooks(t)
	db := openTemp(t)

	res, err := db.Import(lb, "real books")
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if res.Flights != 1293 {
		t.Errorf("imported %d flights, want 1293", res.Flights)
	}
	if res.Totals != lb.Totals {
		t.Errorf("stored totals %+v differ from the source %+v", res.Totals, lb.Totals)
	}

	// Every aircraft in the books resolved to a row, so no flight lost its
	// link. A non-zero unlinked count would mean the seed list missed a
	// registration that was actually flown.
	linked, unlinked, err := db.AircraftLinkage()
	if err != nil {
		t.Fatal(err)
	}
	if linked != 1293 || unlinked != 0 {
		t.Errorf("linkage = %d/%d, want 1293 linked and 0 unlinked", linked, unlinked)
	}

	// Field-by-field, all 1293 rows. A checksum can be passed by two errors
	// that cancel; this cannot.
	got, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(lb.Flights) {
		t.Fatalf("read back %d flights, want %d", len(got), len(lb.Flights))
	}
	for i := range lb.Flights {
		g, w := got[i], lb.Flights[i]
		g.OffBlockUTC, w.OffBlockUTC = g.OffBlockUTC.UTC(), w.OffBlockUTC.UTC()
		g.OnBlockUTC, w.OnBlockUTC = g.OnBlockUTC.UTC(), w.OnBlockUTC.UTC()
		g.TakeoffUTC, w.TakeoffUTC = g.TakeoffUTC.UTC(), w.TakeoffUTC.UTC()
		g.LandingUTC, w.LandingUTC = g.LandingUTC.UTC(), w.LandingUTC.UTC()
		if g != w {
			t.Fatalf("flight seq %d (book %d line %d) changed in the database:\n got %+v\nwant %+v",
				w.Seq, w.SourceBook, w.SourceRow, g, w)
		}
	}

	ds, err := db.Discrepancies()
	if err != nil {
		t.Fatal(err)
	}
	if len(ds) != len(lb.Discrepancies) {
		t.Errorf("stored %d discrepancies, want %d", len(ds), len(lb.Discrepancies))
	}
}

// TestRealImportIsIdempotent runs the import twice into two independent
// databases and requires the results to be indistinguishable. Re-running an
// import on a legal record must never be a gamble (rule 0.2).
func TestRealImportIsIdempotent(t *testing.T) {
	lb := loadRealBooks(t)

	first := openTemp(t)
	if _, err := first.Import(lb, "run one"); err != nil {
		t.Fatal(err)
	}
	// Same database, second run: the replace must land on the same rows.
	if _, err := first.Import(lb, "run two"); err != nil {
		t.Fatal(err)
	}

	second := openTemp(t)
	if _, err := second.Import(lb, "fresh database"); err != nil {
		t.Fatal(err)
	}

	a, err := first.Flights()
	if err != nil {
		t.Fatal(err)
	}
	b, err := second.Flights()
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) {
		t.Fatalf("%d flights vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("flight %d differs between a re-import and a fresh import", i)
		}
	}

	ax, err := first.Aircraft()
	if err != nil {
		t.Fatal(err)
	}
	bx, err := second.Aircraft()
	if err != nil {
		t.Fatal(err)
	}
	if len(ax) != len(bx) {
		t.Fatalf("%d aircraft vs %d", len(ax), len(bx))
	}
	for i := range ax {
		if ax[i] != bx[i] {
			t.Fatalf("aircraft %d differs: %+v vs %+v", i, ax[i], bx[i])
		}
	}
}

// TestRealBackupRestoresTheWholeRecord is the reversibility half of rule 0.2:
// a backup taken before an import must still hold the pre-import record after
// the import has replaced it.
func TestRealBackupRestoresTheWholeRecord(t *testing.T) {
	lb := loadRealBooks(t)
	db := openTemp(t)
	if _, err := db.Import(lb, "the real books"); err != nil {
		t.Fatal(err)
	}

	backup := filepath.Join(t.TempDir(), "before.db")
	if err := db.Backup(backup); err != nil {
		t.Fatalf("Backup: %v", err)
	}

	// Now destroy the live database with a much smaller import.
	small := &csvbook.Logbook{
		Flights:  lb.Flights[:1],
		Aircraft: lb.Aircraft[:1],
	}
	f := small.Flights[0]
	small.Totals = csvbook.Totals{
		Flights: 1, Total: f.TotalMinutes, PIC: f.PICMinutes, Dual: f.DualMinutes,
		Instrument: f.InstrumentMinutes, Night: f.NightMinutes,
		Instructor: f.InstructorMinutes, Landings: f.LandingsDay + f.LandingsNight,
	}
	if f.Class == csvbook.ClassSEPSea {
		small.Totals.SEPSea = f.TotalMinutes
	}
	if _, err := db.Import(small, "destructive"); err != nil {
		t.Fatal(err)
	}
	if n, _ := db.CountFlights(); n != 1 {
		t.Fatalf("expected the live database to be replaced, got %d flights", n)
	}

	restored, err := store.Open(backup)
	if err != nil {
		t.Fatalf("opening the backup: %v", err)
	}
	defer restored.Close()

	if err := restored.Verify(lb.Totals); err != nil {
		t.Errorf("the backup no longer matches the record it was taken from: %v", err)
	}
}
