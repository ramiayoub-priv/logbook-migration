package stats_test

import (
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/stats"
)

// flight builds a minimal flight for a test. Only the fields a given assertion
// cares about are set at the call site; everything else stays zero, which is
// itself a case worth covering (a blank row must not panic or double-count).
func flight(seq int, date string, class csvbook.Class, opts ...func(*csvbook.Flight)) csvbook.Flight {
	f := csvbook.Flight{Seq: seq, Date: date, Class: class}
	for _, o := range opts {
		o(&f)
	}
	return f
}

func times(total, pic, dual, night, instrument, instructor int) func(*csvbook.Flight) {
	return func(f *csvbook.Flight) {
		f.TotalMinutes = total
		f.PICMinutes = pic
		f.DualMinutes = dual
		f.NightMinutes = night
		f.InstrumentMinutes = instrument
		f.InstructorMinutes = instructor
	}
}

func landings(day, night int, verified bool) func(*csvbook.Flight) {
	return func(f *csvbook.Flight) {
		f.LandingsDay = day
		f.LandingsNight = night
		f.LandingsVerified = verified
	}
}

// --- Range ------------------------------------------------------------------

func TestRangeContains(t *testing.T) {
	for _, c := range []struct {
		name string
		r    stats.Range
		date string
		want bool
	}{
		// An empty bound is open. Both empty is the whole logbook, which is
		// what the statistics page opens on.
		{"fully open", stats.Range{}, "2011-09-28", true},
		{"open start, inside", stats.Range{To: "2015-01-01"}, "2011-09-28", true},
		{"open start, outside", stats.Range{To: "2015-01-01"}, "2020-01-01", false},
		{"open end, inside", stats.Range{From: "2015-01-01"}, "2020-01-01", true},
		{"open end, outside", stats.Range{From: "2015-01-01"}, "2011-09-28", false},
		{"closed, inside", stats.Range{From: "2015-01-01", To: "2015-12-31"}, "2015-06-15", true},

		// Both bounds are inclusive. A pilot asking for 2015 means all of
		// 2015, including the flights on 1 January and 31 December.
		{"lower bound is inclusive", stats.Range{From: "2015-01-01", To: "2015-12-31"}, "2015-01-01", true},
		{"upper bound is inclusive", stats.Range{From: "2015-01-01", To: "2015-12-31"}, "2015-12-31", true},
		{"just below", stats.Range{From: "2015-01-01", To: "2015-12-31"}, "2014-12-31", false},
		{"just above", stats.Range{From: "2015-01-01", To: "2015-12-31"}, "2016-01-01", false},

		// A single-day range must select that day, not nothing.
		{"single day", stats.Range{From: "2015-06-15", To: "2015-06-15"}, "2015-06-15", true},

		// An inverted range selects nothing rather than erroring: the UI can
		// put the pickers in either order and an empty result is the honest
		// answer.
		{"inverted", stats.Range{From: "2015-12-31", To: "2015-01-01"}, "2015-06-15", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := c.r.Contains(c.date); got != c.want {
				t.Errorf("Range%+v.Contains(%q) = %v, want %v", c.r, c.date, got, c.want)
			}
		})
	}
}

func TestFilterSelectsOnDateAndPreservesOrder(t *testing.T) {
	// Deliberately out of date order: 18 rows across the three books genuinely
	// are, so Filter must not assume sortedness and must not reorder.
	in := []csvbook.Flight{
		flight(1, "2015-06-15", csvbook.ClassSEPLand),
		flight(2, "2014-01-01", csvbook.ClassSEPLand),
		flight(3, "2015-01-01", csvbook.ClassSEPLand),
		flight(4, "2016-01-01", csvbook.ClassSEPLand),
	}
	got := stats.Filter(in, stats.Range{From: "2015-01-01", To: "2015-12-31"})

	want := []int{1, 3}
	if len(got) != len(want) {
		t.Fatalf("got %d flights, want %d", len(got), len(want))
	}
	for i, seq := range want {
		if got[i].Seq != seq {
			t.Errorf("position %d has seq %d, want %d (Filter must preserve input order)",
				i, got[i].Seq, seq)
		}
	}
}

func TestFilterOnEmptyRangeReturnsNothingRatherThanEverything(t *testing.T) {
	in := []csvbook.Flight{flight(1, "2015-06-15", csvbook.ClassSEPLand)}
	if got := stats.Filter(in, stats.Range{From: "2020-01-01", To: "2020-12-31"}); len(got) != 0 {
		t.Errorf("got %d flights for a range with no flights in it, want 0", len(got))
	}
}

func TestFilterDoesNotAliasTheInput(t *testing.T) {
	// Filter returning a slice that shares the caller's backing array would let
	// a later append silently rewrite the source. This is a legal record.
	in := []csvbook.Flight{
		flight(1, "2015-01-01", csvbook.ClassSEPLand),
		flight(2, "2015-01-02", csvbook.ClassSEPLand),
	}
	got := stats.Filter(in, stats.Range{})
	got[0].Seq = 99
	if in[0].Seq != 1 {
		t.Errorf("mutating the filtered slice changed the input: in[0].Seq = %d", in[0].Seq)
	}
}

// --- Summarize --------------------------------------------------------------

func TestSummarizeOfNothingIsZero(t *testing.T) {
	// A zero-flight month is a normal thing to ask for in winter. It must
	// produce zeroes, not a panic and not a division by zero.
	if got := stats.Summarize(nil); got != (stats.Summary{}) {
		t.Errorf("Summarize(nil) = %+v, want the zero Summary", got)
	}
}

func TestSummarizeSplitsSeaAndLand(t *testing.T) {
	in := []csvbook.Flight{
		flight(1, "2024-06-01", csvbook.ClassSEPSea,
			times(60, 60, 0, 0, 0, 45), landings(6, 0, true)),
		flight(2, "2024-06-02", csvbook.ClassSEPLand,
			times(90, 30, 60, 20, 15, 10), landings(3, 2, true)),
	}
	got := stats.Summarize(in)

	want := stats.Summary{
		Flights:    2,
		Total:      150,
		PIC:        90,
		Dual:       60,
		Night:      20,
		Instrument: 15,
		Instructor: 55,

		SeaTotal:       60,
		SeaPIC:         60,
		SeaInstructor:  45,
		LandTotal:      90,
		LandPIC:        30,
		LandInstructor: 10,

		LandingsDay:   9,
		LandingsNight: 2,
		LandingsSea:   6,
		LandingsLand:  5,
	}
	if got != want {
		t.Errorf("Summarize mismatch\n got %+v\nwant %+v", got, want)
	}
}

// TestSummarizeClassifiesEveryClassAsSeaOrLand pins the sea/land partition for
// the whole EASA class vocabulary, not just the two classes these books happen
// to contain. A new class defaulting to the wrong side would misreport the
// seaplane figures that the owner's ratings depend on.
func TestSummarizeClassifiesEveryClassAsSeaOrLand(t *testing.T) {
	for _, c := range []struct {
		class csvbook.Class
		sea   bool
	}{
		{csvbook.ClassSEPSea, true},
		{csvbook.ClassMEPSea, true},
		{csvbook.ClassSEPLand, false},
		{csvbook.ClassMEPLand, false},
		{csvbook.ClassTMG, false},
	} {
		t.Run(string(c.class), func(t *testing.T) {
			got := stats.Summarize([]csvbook.Flight{
				flight(1, "2024-06-01", c.class, times(60, 60, 0, 0, 0, 0), landings(2, 0, true)),
			})
			if c.sea && (got.SeaTotal != 60 || got.LandTotal != 0 || got.LandingsSea != 2) {
				t.Errorf("%s should count as sea, got sea=%d land=%d ldgSea=%d",
					c.class, got.SeaTotal, got.LandTotal, got.LandingsSea)
			}
			if !c.sea && (got.LandTotal != 60 || got.SeaTotal != 0 || got.LandingsLand != 2) {
				t.Errorf("%s should count as land, got sea=%d land=%d ldgLand=%d",
					c.class, got.SeaTotal, got.LandTotal, got.LandingsLand)
			}
		})
	}
}

// TestSummarizeCountsUnverifiedLandings guards the honesty of the day/night
// split: 30 rows carry night time and their split was inferred rather than read
// off the page (Task 8). The statistics page must be able to say so.
func TestSummarizeCountsUnverifiedLandings(t *testing.T) {
	in := []csvbook.Flight{
		flight(1, "2024-06-01", csvbook.ClassSEPLand, landings(3, 0, true)),
		flight(2, "2024-06-02", csvbook.ClassSEPLand, landings(0, 2, false)),
		flight(3, "2024-06-03", csvbook.ClassSEPLand, landings(1, 1, false)),
	}
	if got := stats.Summarize(in).LandingsUnverified; got != 2 {
		t.Errorf("LandingsUnverified = %d, want 2", got)
	}
}

// TestSummaryPartitionsAgree is the invariant that makes the two splits
// trustworthy: sea+land and day+night slice the same landings on different
// axes, and sea+land time must reconstitute the total.
func TestSummaryPartitionsAgree(t *testing.T) {
	in := []csvbook.Flight{
		flight(1, "2024-06-01", csvbook.ClassSEPSea, times(60, 60, 0, 0, 0, 0), landings(6, 1, true)),
		flight(2, "2024-06-02", csvbook.ClassSEPLand, times(90, 90, 0, 30, 0, 0), landings(3, 2, true)),
		flight(3, "2024-06-03", csvbook.ClassSEPSea, times(45, 0, 45, 0, 0, 0), landings(0, 4, true)),
	}
	s := stats.Summarize(in)

	if s.SeaTotal+s.LandTotal != s.Total {
		t.Errorf("sea %d + land %d != total %d", s.SeaTotal, s.LandTotal, s.Total)
	}
	if s.SeaPIC+s.LandPIC != s.PIC {
		t.Errorf("sea PIC %d + land PIC %d != PIC %d", s.SeaPIC, s.LandPIC, s.PIC)
	}
	if s.SeaInstructor+s.LandInstructor != s.Instructor {
		t.Errorf("sea FI %d + land FI %d != FI %d", s.SeaInstructor, s.LandInstructor, s.Instructor)
	}
	if s.LandingsSea+s.LandingsLand != s.LandingsDay+s.LandingsNight {
		t.Errorf("sea+land landings %d != day+night landings %d",
			s.LandingsSea+s.LandingsLand, s.LandingsDay+s.LandingsNight)
	}
}

func TestSummaryAddIsAssociativeOverPartitions(t *testing.T) {
	// Paginate builds "total previous pages" by adding page summaries, so Add
	// must agree with summarizing the concatenation. If it does not, the EASA
	// PDF's running totals drift from its own rows.
	a := []csvbook.Flight{
		flight(1, "2024-06-01", csvbook.ClassSEPSea, times(60, 60, 0, 0, 0, 30), landings(6, 1, false)),
	}
	b := []csvbook.Flight{
		flight(2, "2024-06-02", csvbook.ClassSEPLand, times(90, 30, 60, 20, 15, 0), landings(3, 2, true)),
	}
	sum := stats.Summarize(a).Add(stats.Summarize(b))
	whole := stats.Summarize(append(append([]csvbook.Flight{}, a...), b...))
	if sum != whole {
		t.Errorf("Add mismatch\n got %+v\nwant %+v", sum, whole)
	}
}

// --- Paginate ---------------------------------------------------------------

func TestPaginateRejectsANonPositivePageSize(t *testing.T) {
	// Silently clamping would produce a plausible-looking PDF with the wrong
	// number of rows per page, which is exactly the sort of quiet corruption
	// rule 0.2 forbids.
	for _, rows := range []int{0, -1} {
		if _, err := stats.Paginate(nil, rows); err == nil {
			t.Errorf("Paginate(nil, %d) returned no error", rows)
		}
	}
}

func TestPaginateOfNothingIsNoPages(t *testing.T) {
	got, err := stats.Paginate(nil, 15)
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d pages for an empty logbook, want 0", len(got))
	}
}

func TestPaginateFillsWholeAndPartialPages(t *testing.T) {
	var in []csvbook.Flight
	for i := 1; i <= 7; i++ {
		in = append(in, flight(i, "2024-06-01", csvbook.ClassSEPLand, times(60, 60, 0, 0, 0, 0)))
	}
	pages, err := stats.Paginate(in, 3)
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}
	if len(pages) != 3 {
		t.Fatalf("got %d pages for 7 flights at 3 rows, want 3", len(pages))
	}
	for i, wantRows := range []int{3, 3, 1} {
		if got := len(pages[i].Flights); got != wantRows {
			t.Errorf("page %d has %d rows, want %d", i+1, got, wantRows)
		}
		if pages[i].Number != i+1 {
			t.Errorf("page at index %d is numbered %d, want %d", i, pages[i].Number, i+1)
		}
	}
}

// TestPaginateOrdersOnSeqAndNotOnDate is the ordering rule that the whole
// project turns on: the books are not in date order, so the EASA reproduction
// must walk seq.
func TestPaginateOrdersOnSeqAndNotOnDate(t *testing.T) {
	in := []csvbook.Flight{
		flight(3, "2014-01-01", csvbook.ClassSEPLand),
		flight(1, "2016-01-01", csvbook.ClassSEPLand),
		flight(2, "2015-01-01", csvbook.ClassSEPLand),
	}
	pages, err := stats.Paginate(in, 15)
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}
	for i, want := range []int{1, 2, 3} {
		if got := pages[0].Flights[i].Seq; got != want {
			t.Errorf("row %d has seq %d, want %d (pages must be ordered on seq)", i, got, want)
		}
	}
}

func TestPaginateDoesNotReorderTheCallersSlice(t *testing.T) {
	in := []csvbook.Flight{
		flight(3, "2014-01-01", csvbook.ClassSEPLand),
		flight(1, "2016-01-01", csvbook.ClassSEPLand),
	}
	if _, err := stats.Paginate(in, 15); err != nil {
		t.Fatalf("Paginate: %v", err)
	}
	if in[0].Seq != 3 || in[1].Seq != 1 {
		t.Errorf("Paginate sorted the caller's slice in place: %d, %d", in[0].Seq, in[1].Seq)
	}
}

// TestPaginateTotalsChainLikeTheEASABook reproduces the three-row totals block
// the paper page carries: TOTAL THIS PAGE, TOTAL PREVIOUS PAGES, TOTAL. Each
// page's Total must equal Previous + ThisPage, and the last page's Total must
// equal the whole logbook.
func TestPaginateTotalsChainLikeTheEASABook(t *testing.T) {
	var in []csvbook.Flight
	for i := 1; i <= 7; i++ {
		class := csvbook.ClassSEPLand
		if i%2 == 0 {
			class = csvbook.ClassSEPSea
		}
		in = append(in, flight(i, "2024-06-01", class,
			times(60+i, 60+i, 0, i, 0, 0), landings(i, 0, true)))
	}
	pages, err := stats.Paginate(in, 3)
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}

	var running stats.Summary
	for _, p := range pages {
		if p.Previous != running {
			t.Errorf("page %d Previous = %+v, want %+v", p.Number, p.Previous, running)
		}
		if want := p.Previous.Add(p.ThisPage); p.Total != want {
			t.Errorf("page %d Total != Previous + ThisPage", p.Number)
		}
		if got := stats.Summarize(p.Flights); got != p.ThisPage {
			t.Errorf("page %d ThisPage = %+v, want %+v", p.Number, p.ThisPage, got)
		}
		running = p.Total
	}
	if whole := stats.Summarize(in); pages[len(pages)-1].Total != whole {
		t.Errorf("the last page's Total = %+v, want the whole logbook %+v",
			pages[len(pages)-1].Total, whole)
	}
}

// TestPaginatePagesDoNotAliasTheInput guards against a page's Flights slice
// sharing a backing array with the sorted copy, which would let a caller
// writing to one page corrupt another.
func TestPaginatePagesDoNotAliasTheInput(t *testing.T) {
	in := []csvbook.Flight{
		flight(1, "2024-06-01", csvbook.ClassSEPLand),
		flight(2, "2024-06-02", csvbook.ClassSEPLand),
	}
	pages, err := stats.Paginate(in, 1)
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}
	pages[0].Flights = append(pages[0].Flights, flight(99, "2099-01-01", csvbook.ClassSEPLand))
	if pages[1].Flights[0].Seq != 2 {
		t.Errorf("appending to page 1 overwrote page 2: seq = %d", pages[1].Flights[0].Seq)
	}
}
