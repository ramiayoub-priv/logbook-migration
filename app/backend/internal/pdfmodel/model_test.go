package pdfmodel_test

import (
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/pdfmodel"
	"github.com/ramiayoub/logbook/backend/internal/stats"
)

func at(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// seaFlight is a seaplane instructing row: PIC and instructor time, day
// landings, no night and no instrument.
func seaFlight(t *testing.T) csvbook.Flight {
	t.Helper()
	return csvbook.Flight{
		Seq: 1, Date: "2021-06-01",
		AircraftType: "C172", AircraftReg: "OH-CTL", Class: csvbook.ClassSEPSea,
		DepPlace: "Tuusulanjärvi", ArrPlace: "Kelvenne",
		OffBlockUTC: at(t, "2021-06-01T15:13:00Z"), OnBlockUTC: at(t, "2021-06-01T16:34:00Z"),
		BlockMinutes: 81, TotalMinutes: 81, PICMinutes: 81, InstructorMinutes: 81,
		PICName: "Ayoub", LandingsDay: 7, LandingsVerified: true,
		SourceBook: 3, SourceRow: 3,
	}
}

func TestEASARowRendersTheGeneralBlock(t *testing.T) {
	r := pdfmodel.EASARowOf(seaFlight(t))

	// The paper column header is dd/mm/yy.
	if r.Date != "01/06/21" {
		t.Errorf("Date = %q, want 01/06/21", r.Date)
	}
	if r.DepPlace != "Tuusulanjärvi" || r.ArrPlace != "Kelvenne" {
		t.Errorf("places = %q -> %q", r.DepPlace, r.ArrPlace)
	}
	// The column is headed "OFF BLOCK (UTC)", so it shows the UTC instant.
	if r.OffBlock != "15:13" || r.OnBlock != "16:34" {
		t.Errorf("times = %q/%q, want 15:13/16:34", r.OffBlock, r.OnBlock)
	}
	if r.Type != "C172" || r.Reg != "OH-CTL" {
		t.Errorf("aircraft = %q %q", r.Type, r.Reg)
	}
	if r.PICName != "Ayoub" {
		t.Errorf("PICName = %q", r.PICName)
	}
}

// A single-engine flight's time goes in SINGLE ENGINE VFR, which is what the
// owner's own pages do, and the multi-engine columns stay empty because no
// multi-engine aircraft appears anywhere in the three books.
func TestEASARowPutsSingleEngineTimeInTheVFRColumn(t *testing.T) {
	r := pdfmodel.EASARowOf(seaFlight(t))

	if r.SEVFR != "1:21" {
		t.Errorf("SEVFR = %q, want 1:21", r.SEVFR)
	}
	if r.Total != "1:21" {
		t.Errorf("Total = %q, want the flight's own 1:21", r.Total)
	}
	if r.MEVFR != "" || r.MEIFR != "" {
		t.Errorf("multi-engine columns = %q/%q, want both empty", r.MEVFR, r.MEIFR)
	}
}

// The CSVs carry no flight-rules column, and instrument time is not the same
// thing as an IFR flight -- instrument time is logged under the hood in C152s
// that are not IFR certified. Deriving an IFR figure from it would be
// inventing data on a legal record.
func TestEASARowLeavesTheIFRColumnEmpty(t *testing.T) {
	f := seaFlight(t)
	f.InstrumentMinutes = 45
	r := pdfmodel.EASARowOf(f)

	if r.SEIFR != "" {
		t.Errorf("SEIFR = %q, want empty; the books record no flight rules", r.SEIFR)
	}
}

// The paper leaves a cell blank rather than writing 0:00, and a page of zeros
// is unreadable.
func TestEASARowLeavesZeroDurationsBlank(t *testing.T) {
	r := pdfmodel.EASARowOf(seaFlight(t))

	for _, c := range []struct {
		name string
		got  string
	}{
		{"Night", r.Night}, {"Dual", r.Dual}, {"Copilot", r.Copilot},
		{"MultiPilot", r.MultiPilot}, {"InstructorSTD", r.InstructorSTD},
		{"LandingsNight", r.LandingsNight},
	} {
		if c.got != "" {
			t.Errorf("%s = %q, want empty for a zero", c.name, c.got)
		}
	}
	if r.LandingsDay != "7" {
		t.Errorf("LandingsDay = %q, want 7", r.LandingsDay)
	}
}

func TestEASARowRendersTheFunctionColumns(t *testing.T) {
	f := seaFlight(t)
	f.PICMinutes = 0
	f.DualMinutes = 90
	f.InstructorMinutes = 0
	f.NightMinutes = 25
	f.LandingsDay = 2
	f.LandingsNight = 1
	r := pdfmodel.EASARowOf(f)

	if r.Dual != "1:30" {
		t.Errorf("Dual = %q, want 1:30", r.Dual)
	}
	if r.PIC != "" {
		t.Errorf("PIC = %q, want empty", r.PIC)
	}
	if r.Night != "0:25" {
		t.Errorf("Night = %q, want 0:25", r.Night)
	}
	if r.LandingsDay != "2" || r.LandingsNight != "1" {
		t.Errorf("landings = %q/%q, want 2/1", r.LandingsDay, r.LandingsNight)
	}
}

// A blank off-block cell on paper must stay blank, not become 00:00 -- which
// would read as a real time nobody flew.
func TestEASARowKeepsABlankTimeBlank(t *testing.T) {
	f := seaFlight(t)
	f.OffBlockUTC = time.Time{}
	f.OnBlockUTC = time.Time{}
	r := pdfmodel.EASARowOf(f)

	if r.OffBlock != "" || r.OnBlock != "" {
		t.Errorf("blank times rendered as %q/%q, want both empty", r.OffBlock, r.OnBlock)
	}
}

// The pagination and its three-row totals block are the part of this document
// that a bug turns into a wrong legal record, so the arithmetic is asserted
// rather than eyeballed in a viewer.
func TestEASAPagesCarryTheThreeRowTotalsBlock(t *testing.T) {
	var flights []csvbook.Flight
	for i := 0; i < 20; i++ {
		f := seaFlight(t)
		f.Seq = i + 1
		f.TotalMinutes = 60
		f.PICMinutes = 60
		f.InstructorMinutes = 0
		f.LandingsDay = 1
		flights = append(flights, f)
	}

	pages, err := pdfmodel.EASAPages(flights, 15)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 {
		t.Fatalf("got %d pages, want 2 for 20 flights at 15 a page", len(pages))
	}

	first, second := pages[0], pages[1]
	if first.ThisPage.Total != "15:00" {
		t.Errorf("page 1 TOTAL THIS PAGE = %q, want 15:00", first.ThisPage.Total)
	}
	if first.Previous.Total != "" {
		t.Errorf("page 1 TOTAL PREVIOUS PAGES = %q, want empty on the first page", first.Previous.Total)
	}
	if first.Total.Total != "15:00" {
		t.Errorf("page 1 TOTAL = %q, want 15:00", first.Total.Total)
	}

	if second.ThisPage.Total != "5:00" {
		t.Errorf("page 2 TOTAL THIS PAGE = %q, want 5:00", second.ThisPage.Total)
	}
	if second.Previous.Total != "15:00" {
		t.Errorf("page 2 TOTAL PREVIOUS PAGES = %q, want 15:00", second.Previous.Total)
	}
	// The figure the authority reads: the running total after every flight.
	if second.Total.Total != "20:00" {
		t.Errorf("page 2 TOTAL = %q, want 20:00", second.Total.Total)
	}
	if second.Total.LandingsDay != "20" {
		t.Errorf("page 2 TOTAL landings = %q, want 20", second.Total.LandingsDay)
	}
}

// TOTAL must always equal TOTAL PREVIOUS PAGES + TOTAL THIS PAGE. It is the
// one identity on the page an authority can check by eye, and the whole reason
// cumulatives are computed rather than stored (rule 0.5).
func TestEASAPageTotalsAlwaysReconcile(t *testing.T) {
	var flights []csvbook.Flight
	for i := 0; i < 47; i++ {
		f := seaFlight(t)
		f.Seq = i + 1
		f.TotalMinutes = 37 + i // deliberately uneven
		f.PICMinutes = f.TotalMinutes
		f.NightMinutes = i % 5
		f.LandingsDay = i % 3
		f.LandingsNight = i % 2
		flights = append(flights, f)
	}

	pages, err := pdfmodel.EASAPages(flights, 15)
	if err != nil {
		t.Fatal(err)
	}

	var running stats.Summary
	for _, p := range pages {
		running = running.Add(p.ThisPageSummary)
		if p.TotalSummary != running {
			t.Fatalf("page %d TOTAL does not equal every flight up to it", p.Number)
		}
		if p.TotalSummary != p.PreviousSummary.Add(p.ThisPageSummary) {
			t.Fatalf("page %d: TOTAL != PREVIOUS + THIS PAGE", p.Number)
		}
	}

	// And the last page's TOTAL is the whole logbook.
	if got, want := pages[len(pages)-1].TotalSummary, stats.Summarize(flights); got != want {
		t.Errorf("final TOTAL = %+v, want the summary of every flight %+v", got, want)
	}
}

func TestEASAPagesRejectsANonsensePageSize(t *testing.T) {
	if _, err := pdfmodel.EASAPages(nil, 0); err == nil {
		t.Error("EASAPages accepted a page size of 0")
	}
}

// An empty logbook is a real case, not a degenerate one -- a fresh install.
func TestEASAPagesOfAnEmptyLogbook(t *testing.T) {
	pages, err := pdfmodel.EASAPages(nil, 15)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 0 {
		t.Errorf("got %d pages for an empty logbook, want 0", len(pages))
	}
}

// Every page reports its position, because the paper form has "Page _ of 128"
// printed on it and an authority checks that a submission is complete.
func TestEASAPagesAreNumbered(t *testing.T) {
	var flights []csvbook.Flight
	for i := 0; i < 31; i++ {
		f := seaFlight(t)
		f.Seq = i + 1
		flights = append(flights, f)
	}
	pages, err := pdfmodel.EASAPages(flights, 15)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 3 {
		t.Fatalf("got %d pages, want 3", len(pages))
	}
	for i, p := range pages {
		if p.Number != i+1 {
			t.Errorf("page at index %d is numbered %d", i, p.Number)
		}
		if p.Of != 3 {
			t.Errorf("page %d says 'of %d', want 'of 3'", p.Number, p.Of)
		}
	}
}

// The figures on the statistics sheet are what evidence a rating, so they are
// asserted exactly rather than trusted to the renderer.
func TestStatisticsLinesCarryEveryRequiredFigure(t *testing.T) {
	s := stats.Summary{
		Flights: 3, Total: 200, PIC: 150, Dual: 50, Night: 25, Instrument: 30, Instructor: 60,
		SeaTotal: 120, SeaPIC: 100, SeaInstructor: 60,
		LandTotal: 80, LandPIC: 50, LandInstructor: 0,
		LandingsDay: 7, LandingsNight: 2, LandingsSea: 5, LandingsLand: 4,
	}
	lines := pdfmodel.StatisticsLines(s)

	got := map[string]string{}
	for _, l := range lines {
		got[l.Label] = l.Value
	}
	for _, want := range []struct{ label, value string }{
		{"Flights", "3"},
		{"Total time", "3:20"},
		{"Pilot in command", "2:30"},
		{"Dual", "0:50"},
		{"Night", "0:25"},
		{"Instrument", "0:30"},
		{"Instructor", "1:00"},
		{"Seaplane total", "2:00"},
		{"Seaplane PIC", "1:40"},
		{"Seaplane instructor", "1:00"},
		{"Landplane total", "1:20"},
		{"Landplane PIC", "0:50"},
		{"Landplane instructor", "0:00"},
		{"Landings day", "7"},
		{"Landings night", "2"},
		{"Landings seaplane", "5"},
		{"Landings landplane", "4"},
	} {
		if got[want.label] != want.value {
			t.Errorf("%s = %q, want %q", want.label, got[want.label], want.value)
		}
	}
}

// No multi-engine aircraft appears in the three books, so this branch has no
// real data behind it -- which is exactly why it is asserted. The day a twin
// is flown, its time must land in the multi-engine column and not silently in
// the single-engine one that evidences a different rating.
func TestAMultiEngineFlightUsesTheMultiEngineColumn(t *testing.T) {
	f := seaFlight(t)
	f.Class = csvbook.ClassMEPLand
	r := pdfmodel.EASARowOf(f)

	if r.MEVFR != "1:21" {
		t.Errorf("MEVFR = %q, want 1:21", r.MEVFR)
	}
	if r.SEVFR != "" {
		t.Errorf("SEVFR = %q, want empty for a twin", r.SEVFR)
	}
}

// A date the application cannot parse is shown as stored rather than blanked.
// Hiding it would remove the only clue that something is wrong with the row.
func TestAnUnparseableDateIsPassedThrough(t *testing.T) {
	f := seaFlight(t)
	f.Date = "not-a-date"
	if got := pdfmodel.EASARowOf(f).Date; got != "not-a-date" {
		t.Errorf("Date = %q, want it passed through unchanged", got)
	}
}

// The asterisk is the honest half of Task 8: 30 rows carry a day/night split
// that was seeded from the presence of night time, not read off the page.
func TestLandingSplitMarksAnInferredRow(t *testing.T) {
	f := seaFlight(t)
	f.LandingsDay, f.LandingsNight = 2, 1

	f.LandingsVerified = true
	if got := pdfmodel.LandingSplit(f); got != "2/1" {
		t.Errorf("verified split = %q, want 2/1", got)
	}
	f.LandingsVerified = false
	if got := pdfmodel.LandingSplit(f); got != "2/1*" {
		t.Errorf("inferred split = %q, want 2/1* so the page does not claim a verification", got)
	}
}

// Every figure has to stay traceable to the page it came from, or to the app
// if nobody wrote it on paper.
func TestSourceRefNamesTheOrigin(t *testing.T) {
	f := seaFlight(t)
	if got := pdfmodel.SourceRef(f); got != "b3:3" {
		t.Errorf("SourceRef = %q, want b3:3", got)
	}
	f.SourceBook, f.SourceRow = 0, 4
	if got := pdfmodel.SourceRef(f); got != "app" {
		t.Errorf("SourceRef of a hand-entered row = %q, want app", got)
	}
}

// The range is printed on all three documents, and "Whole logbook" versus a
// bounded range is the difference between a complete record and an extract.
func TestDescribeRange(t *testing.T) {
	for _, tc := range []struct {
		rng  stats.Range
		want string
	}{
		{stats.Range{}, "Whole logbook"},
		{stats.Range{To: "2024-12-31"}, "Up to 2024-12-31"},
		{stats.Range{From: "2024-01-01"}, "From 2024-01-01"},
		{stats.Range{From: "2024-01-01", To: "2024-12-31"}, "2024-01-01 to 2024-12-31"},
	} {
		if got := pdfmodel.DescribeRange(tc.rng); got != tc.want {
			t.Errorf("DescribeRange(%+v) = %q, want %q", tc.rng, got, tc.want)
		}
	}
}

// Clock is exported for the flight table's takeoff and landing cells, so its
// two behaviours are asserted directly rather than only through EASARowOf.
// The blank is the one that matters: an instant nobody recorded must not
// print as a time somebody measured.
func TestClockRendersUTCAndBlanksAnAbsentInstant(t *testing.T) {
	if got := pdfmodel.Clock(time.Time{}); got != "" {
		t.Errorf("Clock(zero) = %q, want the empty string", got)
	}
	at := time.Date(2026, 9, 3, 14, 36, 0, 0, time.UTC)
	if got := pdfmodel.Clock(at); got != "14:36" {
		t.Errorf("Clock(%s) = %q, want 14:36", at, got)
	}
	// A non-UTC instant is converted, never printed as its local wall clock:
	// rule 0.4 says every displayed instant is UTC.
	helsinki := time.FixedZone("EEST", 3*3600)
	if got := pdfmodel.Clock(at.In(helsinki)); got != "14:36" {
		t.Errorf("Clock(same instant in +03) = %q, want 14:36", got)
	}
}

// AirTime is wheels-up to wheels-down. It is derived at render time and never
// stored (rule 0.5), and it must be blank rather than 0:00 whenever it cannot
// be known -- a zero would say the aeroplane never left the ground.
func TestAirTimeIsDerivedAndBlankWhenItCannotBeKnown(t *testing.T) {
	at := func(h, m int) time.Time { return time.Date(2026, 9, 3, h, m, 0, 0, time.UTC) }

	for _, tc := range []struct {
		name             string
		takeoff, landing time.Time
		want             string
	}{
		{"neither recorded", time.Time{}, time.Time{}, ""},
		{"takeoff only", at(14, 36), time.Time{}, ""},
		{"landing only", time.Time{}, at(15, 28), ""},
		{"a normal flight", at(14, 36), at(15, 28), "0:52"},
		{"a long flight", at(6, 0), at(9, 5), "3:05"},
		// The pair is built by timeutil.BlockPair, which rolls the landing
		// forward a day, so the instants carry their own dates and midnight
		// needs no special case here.
		{"across midnight", at(23, 50), at(23, 50).Add(40 * time.Minute), "0:40"},
		// Cannot arise from stored data. If it ever does, a negative air time
		// is not a figure to print -- same ruling as the on-screen table.
		{"landing before takeoff", at(15, 28), at(14, 36), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := csvbook.Flight{TakeoffUTC: tc.takeoff, LandingUTC: tc.landing}
			if got := pdfmodel.AirTime(f); got != tc.want {
				t.Errorf("AirTime = %q, want %q", got, tc.want)
			}
		})
	}
}
