package store_test

import (
	"errors"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/entry"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

// handEntered is a flight as internal/entry produces one: no Seq (the store
// allocates it) and SourceBook 0, which is what marks it as typed into the app
// rather than transcribed off paper.
func handEntered() csvbook.Flight {
	return csvbook.Flight{
		Date:         "2026-07-30",
		AircraftType: "C172", AircraftReg: "OH-CAM", Class: csvbook.ClassSEPLand,
		DepPlace: "EFHF", ArrPlace: "EFHF",
		OffBlockRaw: "09:15Z", OnBlockRaw: "10:30Z", TimeOrigin: "utc_as_written",
		BlockMinutes: 75, TotalMinutes: 75, PICMinutes: 75,
		PICName: "self", LandingsDay: 3, LandingsVerified: true,
		SourceBook: entry.HandEnteredBook,
	}
}

func mustAdd(t *testing.T, db *store.DB, f csvbook.Flight) csvbook.Flight {
	t.Helper()
	added, err := db.AddFlight(f)
	if err != nil {
		t.Fatalf("AddFlight: %v", err)
	}
	return added
}

func findBySeq(t *testing.T, db *store.DB, seq int) csvbook.Flight {
	t.Helper()
	all, err := db.Flights()
	if err != nil {
		t.Fatalf("Flights: %v", err)
	}
	for _, f := range all {
		if f.Seq == seq {
			return f
		}
	}
	t.Fatalf("no flight with seq %d among the %d in the database", seq, len(all))
	return csvbook.Flight{}
}

func TestAddFlightStoresEveryField(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}

	added := mustAdd(t, db, handEntered())
	got := findBySeq(t, db, added.Seq)

	if got.Date != "2026-07-30" || got.AircraftReg != "OH-CAM" {
		t.Errorf("stored %s %s, want 2026-07-30 OH-CAM", got.Date, got.AircraftReg)
	}
	if got.TotalMinutes != 75 || got.PICMinutes != 75 {
		t.Errorf("stored %d total / %d pic, want 75/75", got.TotalMinutes, got.PICMinutes)
	}
	if !got.LandingsVerified {
		t.Error("LandingsVerified did not survive the round trip")
	}
	if got.SourceBook != entry.HandEnteredBook {
		t.Errorf("SourceBook = %d, want %d", got.SourceBook, entry.HandEnteredBook)
	}
}

// THE test on this path. The importer replaces the flights table on every run
// and the migration effort re-imports every time logbook_3.csv grows, so a
// hand-entered row that the importer owns is a row that gets silently deleted
// the next time somebody transcribes a page. CLAUDE.md rule 0.2: never lose a
// row.
func TestHandEnteredFlightsSurviveAReimport(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "first"); err != nil {
		t.Fatal(err)
	}
	added := mustAdd(t, db, handEntered())

	// The migration appends a page and the CSVs are imported again.
	if _, err := db.Import(sample(t), "second"); err != nil {
		t.Fatalf("re-import: %v", err)
	}

	got := findBySeq(t, db, added.Seq)
	if got.TotalMinutes != 75 {
		t.Errorf("the hand-entered flight came back as %d minutes, want 75", got.TotalMinutes)
	}
	n, err := db.CountFlights()
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("flights after re-import = %d, want 3 (two imported plus one hand-entered)", n)
	}
}

// The import's checksums answer "is the database what the CSVs say". A
// hand-entered row is not in the CSVs, so counting it would make that question
// unanswerable -- the import would fail verification on its own correct work.
func TestImportVerificationIgnoresHandEnteredRows(t *testing.T) {
	db := openTemp(t)
	lb := sample(t)
	if _, err := db.Import(lb, "first"); err != nil {
		t.Fatal(err)
	}
	mustAdd(t, db, handEntered())

	res, err := db.Import(lb, "second")
	if err != nil {
		t.Fatalf("re-import with a hand-entered row present: %v", err)
	}
	if res.Totals.Total != lb.Totals.Total {
		t.Errorf("verified total = %d, want the CSVs' %d; the hand-entered 75 minutes leaked into the checksum",
			res.Totals.Total, lb.Totals.Total)
	}
	if res.Flights != lb.Totals.Flights {
		t.Errorf("verified flight count = %d, want the CSVs' %d", res.Flights, lb.Totals.Flights)
	}

	// Verify, which runs the same gate against a live database, must agree.
	if err := db.Verify(lb.Totals); err != nil {
		t.Errorf("Verify against the CSV totals failed with a hand-entered row present: %v", err)
	}
}

// seq is the key every cumulative computation walks and the importer reassigns
// it 1..N on every run. A hand-entered row therefore cannot live in that range
// without colliding the moment the books grow past it -- Book 3 is still being
// transcribed.
func TestHandEnteredSeqCannotCollideWithTheImporter(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}

	first := mustAdd(t, db, handEntered())
	second := handEntered()
	second.Date = "2026-07-31"
	addedSecond := mustAdd(t, db, second)

	if first.Seq < store.HandEnteredSeqBase {
		t.Errorf("first hand-entered seq = %d, want at least %d so the importer's 1..N can never reach it",
			first.Seq, store.HandEnteredSeqBase)
	}
	if addedSecond.Seq <= first.Seq {
		t.Errorf("second hand-entered seq = %d, want it after the first at %d", addedSecond.Seq, first.Seq)
	}
}

// Book order must put a flight flown today after every page of the paper
// books, and Flights() is what the EASA PDF and every cumulative walk.
func TestHandEnteredFlightsSortAfterTheBooks(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	mustAdd(t, db, handEntered())

	all, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("got %d flights, want 3", len(all))
	}
	if all[2].SourceBook != entry.HandEnteredBook {
		t.Errorf("last flight in book order came from book %d, want the hand-entered one", all[2].SourceBook)
	}
	for i := 1; i < len(all); i++ {
		if all[i].Seq <= all[i-1].Seq {
			t.Errorf("flights are not in ascending seq order at index %d", i)
		}
	}
}

// A double-tapped submit button on a phone must not put the same flight in a
// legal record twice.
func TestAddFlightRefusesAnObviousDuplicate(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	mustAdd(t, db, handEntered())

	_, err := db.AddFlight(handEntered())
	if !errors.Is(err, store.ErrDuplicateFlight) {
		t.Fatalf("second identical AddFlight returned %v, want ErrDuplicateFlight", err)
	}

	n, err := db.CountFlights()
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("flights = %d, want 3; the refused duplicate was written anyway", n)
	}
}

// The duplicate check keys on the flight, not on the pilot: two genuinely
// different flights on one day are ordinary and must both be accepted.
func TestAddFlightAcceptsASecondFlightTheSameDay(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	mustAdd(t, db, handEntered())

	second := handEntered()
	second.OffBlockRaw = "11:00Z"
	second.OnBlockRaw = "12:00Z"
	mustAdd(t, db, second)

	n, err := db.CountFlights()
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Errorf("flights = %d, want 4", n)
	}
}

// The link is what makes the aircraft page work; a missing link would also
// hide the flight from anything that joins on it.
func TestAddFlightLinksAKnownAircraft(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}

	f := handEntered()
	f.AircraftReg = "OH-CTL" // in the sample's aircraft list
	added := mustAdd(t, db, f)

	linked, _, err := db.AircraftLinkage()
	if err != nil {
		t.Fatal(err)
	}
	if linked != 3 {
		t.Errorf("linked flights = %d, want 3", linked)
	}
	if got := findBySeq(t, db, added.Seq); got.AircraftReg != "OH-CTL" {
		t.Errorf("registration = %q, want OH-CTL", got.AircraftReg)
	}
}

// An aircraft flown for the first time must not cost the pilot the flight. The
// row goes in unlinked, exactly as the importer treats an unknown registration.
func TestAddFlightAcceptsAnUnknownAircraft(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}

	f := handEntered()
	f.AircraftReg = "OH-NEW"
	added := mustAdd(t, db, f)

	if got := findBySeq(t, db, added.Seq); got.AircraftReg != "OH-NEW" {
		t.Errorf("registration = %q, want OH-NEW", got.AircraftReg)
	}
	_, unlinked, err := db.AircraftLinkage()
	if err != nil {
		t.Fatal(err)
	}
	if unlinked != 1 {
		t.Errorf("unlinked = %d, want 1", unlinked)
	}
}

// The statistics are computed from the rows, so a hand-entered flight has to
// move them -- that is the whole point of entering it.
func TestHandEnteredFlightsCountTowardTheTotals(t *testing.T) {
	db := openTemp(t)
	lb := sample(t)
	if _, err := db.Import(lb, "seed"); err != nil {
		t.Fatal(err)
	}
	mustAdd(t, db, handEntered())

	all, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	var total int
	for _, f := range all {
		total += f.TotalMinutes
	}
	if want := lb.Totals.Total + 75; total != want {
		t.Errorf("total over all flights = %d, want %d", total, want)
	}
}

// source_row is part of a UNIQUE constraint, so every hand-entered row needs
// its own. Sharing one would make the second insert fail at the database.
func TestHandEnteredRowsGetDistinctSourceRows(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}

	a := mustAdd(t, db, handEntered())
	second := handEntered()
	second.Date = "2026-07-31"
	b := mustAdd(t, db, second)

	if a.SourceRow == b.SourceRow {
		t.Errorf("both hand-entered rows carry source_row %d", a.SourceRow)
	}
}

// The importer replaces the aircraft table wholesale, which sets aircraft_id
// to NULL on every row that referenced the old ids -- including hand-entered
// flights, which the importer otherwise leaves alone. Left unfixed, a flight
// typed in the app quietly loses its aircraft link the next time somebody
// transcribes a page, and it never comes back.
func TestAReimportRelinksHandEnteredFlights(t *testing.T) {
	db := openTemp(t)
	lb := sample(t)
	if _, err := db.Import(lb, "first"); err != nil {
		t.Fatal(err)
	}

	f := handEntered()
	f.AircraftReg = "OH-CTL" // in the sample's aircraft list
	mustAdd(t, db, f)

	if _, err := db.Import(lb, "second"); err != nil {
		t.Fatal(err)
	}

	linked, unlinked, err := db.AircraftLinkage()
	if err != nil {
		t.Fatal(err)
	}
	if unlinked != 0 {
		t.Errorf("%d flights lost their aircraft link across a re-import, want 0", unlinked)
	}
	if linked != 3 {
		t.Errorf("linked = %d, want all 3", linked)
	}
}
