package stats_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/stats"
)

// The unit tests prove the aggregation rules. This file proves they reproduce
// the pilot's actual logbook -- the same standard internal/csvbook is held to.
//
// The figures below are the ones inked in the paper book and frozen by the
// owner on 2026-08-01 (claude-docs/resume.md). If one of them fails here, the
// aggregation is wrong until proven otherwise. Do not adjust an expectation to
// make a test pass (CLAUDE.md rule 0.2).

func repoRoot(t *testing.T) string {
	t.Helper()
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	// .../app/backend/internal/stats/realdata_test.go -> repo root
	return filepath.Clean(filepath.Join(filepath.Dir(self), "..", "..", "..", ".."))
}

func loadReal(t *testing.T) *csvbook.Logbook {
	t.Helper()
	lb, err := csvbook.Load(csvbook.DefaultSources(repoRoot(t)))
	if err != nil {
		t.Fatalf("loading the real books: %v", err)
	}
	return lb
}

func hm(h, m int) int { return h*60 + m }

// TestSummarizeReproducesTheFrozenTotals is the checksum that matters: two
// independent code paths -- csvbook's accumulator and this package's -- must
// agree on the same nine figures, and both must equal the paper.
func TestSummarizeReproducesTheFrozenTotals(t *testing.T) {
	lb := loadReal(t)
	s := stats.Summarize(lb.Flights)

	for _, c := range []struct {
		name      string
		got, want int
	}{
		// Moved 2026-08-01 by the three 28/08/2025 flights the owner found
		// missing from the CSV entirely. Deltas: +3 flights, total +2:35,
		// PIC +1:42, dual +0:53, instrument +0:53, landings +5; night,
		// instructor and seaplane unmoved. The reasoning, and why this is the
		// one sanctioned exception to the frozen cumulatives, is on the same
		// constants in internal/csvbook/realdata_test.go.
		{"flights", s.Flights, 1296},
		{"total", s.Total, hm(1222, 10)},
		{"PIC", s.PIC, hm(1054, 45)},
		{"dual", s.Dual, hm(167, 25)},
		{"instrument", s.Instrument, hm(107, 58)},
		{"night", s.Night, hm(22, 45)},
		{"instructor", s.Instructor, hm(189, 41)},
		{"seaplane", s.SeaTotal, hm(407, 39)},
		{"landings", s.LandingsDay + s.LandingsNight, 3444},
	} {
		if c.got != c.want {
			t.Errorf("%s = %d minutes, want %d", c.name, c.got, c.want)
		}
	}

	// And it must agree with the loader's own accumulator, which computes the
	// same figures by a different route.
	if s.Total != lb.Totals.Total || s.PIC != lb.Totals.PIC || s.SeaTotal != lb.Totals.SEPSea {
		t.Errorf("stats disagrees with csvbook.Totals: %d/%d/%d vs %d/%d/%d",
			s.Total, s.PIC, s.SeaTotal, lb.Totals.Total, lb.Totals.PIC, lb.Totals.SEPSea)
	}
}

// TestRealPartitionsAgree runs the partition invariants over the whole corpus
// rather than a three-row fixture. A single misclassified row out of 1293 shows
// up here and nowhere else.
func TestRealPartitionsAgree(t *testing.T) {
	s := stats.Summarize(loadReal(t).Flights)

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

	// 30 rows carry night time and none of their splits has been read off the
	// page yet (Task 8). This is the figure the statistics page must disclose,
	// and it drops as the backfill lands.
	if s.LandingsUnverified != 30 {
		t.Errorf("LandingsUnverified = %d, want 30", s.LandingsUnverified)
	}
	if s.LandingsNight != 0 {
		t.Errorf("night landings = %d; the split has not been backfilled yet", s.LandingsNight)
	}
}

// TestRangeSlicesTheRealBooksWithoutLosingAFlight checks the property that
// makes the date filter safe on a legal record: any partition of the timeline
// must put every flight in exactly one part.
func TestRangeSlicesTheRealBooksWithoutLosingAFlight(t *testing.T) {
	lb := loadReal(t)
	whole := stats.Summarize(lb.Flights)

	// The books run 2011-2026. Cut at the start of 2018, which is inside the
	// range and lands mid-book.
	early := stats.Summarize(stats.Filter(lb.Flights, stats.Range{To: "2017-12-31"}))
	late := stats.Summarize(stats.Filter(lb.Flights, stats.Range{From: "2018-01-01"}))

	if got := early.Add(late); got != whole {
		t.Errorf("splitting the timeline lost or duplicated flights\n got %+v\nwant %+v", got, whole)
	}
	if early.Flights == 0 || late.Flights == 0 {
		t.Errorf("the split should be non-trivial: %d early, %d late", early.Flights, late.Flights)
	}

	// A range beyond the last flight is empty, not everything.
	if got := stats.Filter(lb.Flights, stats.Range{From: "2099-01-01"}); len(got) != 0 {
		t.Errorf("got %d flights after 2099, want 0", len(got))
	}
}

// TestPaginateTheRealBooksIntoEASAPages checks the page geometry the PDF
// depends on: 1296 flights at 15 rows to a page is 86 full pages and a partial
// 87th, and the running block must chain from the first page to the last
// without drifting from the whole-logbook figure.
func TestPaginateTheRealBooksIntoEASAPages(t *testing.T) {
	lb := loadReal(t)
	const rowsPerPage = 15

	pages, err := stats.Paginate(lb.Flights, rowsPerPage)
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}
	if len(pages) != 87 {
		t.Fatalf("got %d pages for 1296 flights at %d rows, want 87", len(pages), rowsPerPage)
	}
	// Still 87 pages: the three 28/08/2025 late entries landed on the already
	// partial last page, taking it from 3 rows to 6.
	if got := len(pages[86].Flights); got != 6 {
		t.Errorf("the last page has %d rows, want 6 (1296 = 86*15 + 6)", got)
	}

	var seen int
	var running stats.Summary
	for _, p := range pages {
		seen += len(p.Flights)
		if p.Previous != running {
			t.Fatalf("page %d: TOTAL PREVIOUS PAGES drifted from the running total", p.Number)
		}
		if p.Total != p.Previous.Add(p.ThisPage) {
			t.Fatalf("page %d: TOTAL != TOTAL PREVIOUS PAGES + TOTAL THIS PAGE", p.Number)
		}
		running = p.Total
	}
	if seen != 1296 {
		t.Errorf("the pages hold %d flights between them, want 1296", seen)
	}

	// The bottom line of the last page is the logbook total. This is the figure
	// an authority reads, and it must equal the frozen paper figure.
	last := pages[len(pages)-1].Total
	if last != stats.Summarize(lb.Flights) {
		t.Errorf("the last page's TOTAL != the whole logbook")
	}
	if last.Total != hm(1222, 10) {
		t.Errorf("the last page's TOTAL is %d minutes, want 1222:10", last.Total)
	}
}

// TestPaginateWalksSeqNotDate is the ordering rule stated against the real
// corpus: 18 rows are genuinely out of date order, so a page built on the date
// would differ from one built on seq. Assert that seq order is what comes out
// and that it is not accidentally the same as date order.
func TestPaginateWalksSeqNotDate(t *testing.T) {
	lb := loadReal(t)

	pages, err := stats.Paginate(lb.Flights, 15)
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}

	var prevSeq, outOfDateOrder int
	var prevDate string
	for _, p := range pages {
		for _, f := range p.Flights {
			if f.Seq <= prevSeq {
				t.Fatalf("seq went backwards: %d after %d", f.Seq, prevSeq)
			}
			if prevDate != "" && f.Date < prevDate {
				outOfDateOrder++
			}
			prevSeq, prevDate = f.Seq, f.Date
		}
	}
	// If this ever hits zero the two orderings have converged and this test has
	// stopped proving anything -- but it would also mean the books were
	// rewritten, which is not something that happens quietly.
	if outOfDateOrder == 0 {
		t.Error("no row is out of date order; the seq-vs-date distinction is untested")
	}
	t.Logf("%d rows are out of date order, ordered correctly on seq", outOfDateOrder)
}

// --- Aircraft time against the real books (Task 13, 2026-08-02) -------------

// The aircraft-time page's honesty depends on two facts about the corpus, and
// both are asserted here rather than assumed, because the page states them to
// the pilot in words: block time is known for EVERY flight, and air time for
// almost none of them.
//
// If the first ever stops being true the page is quietly under-billing, and if
// the second changes the coverage sentence needs rewording. Either way it must
// be a failing test rather than a page that has silently become wrong.
func TestRealBooksAircraftTimeCoverage(t *testing.T) {
	lb := loadReal(t)
	rows := stats.ByAircraft(lb.Flights)
	total := stats.TotalAircraftTime(rows)

	if len(rows) != 38 {
		t.Errorf("got %d aircraft, want 38", len(rows))
	}
	if total.Flights != 1296 {
		t.Errorf("got %d flights, want 1296", total.Flights)
	}

	// Block time is complete: every one of the 1296 rows carries one. This is
	// what lets the page print the block total without a caveat.
	var noBlock int
	for _, f := range lb.Flights {
		if f.BlockMinutes == 0 {
			noBlock++
		}
	}
	if noBlock != 0 {
		t.Errorf("%d flights carry no block time; the page bills on block and "+
			"states it without a caveat, so this must stay zero", noBlock)
	}

	// Air time is known on 19 rows. THIS is why the figure never travels
	// without its coverage, and why it is never mixed with the block total.
	if total.AirKnown != 19 || total.AirMissing != 1277 {
		t.Errorf("air coverage = %d known / %d missing, want 19 / 1277",
			total.AirKnown, total.AirMissing)
	}

	// The one flagged row in the corpus where block and total disagree:
	// 08/09/2025, block 0:45 vs total 0:38. It is counted, not reconciled.
	if total.BlockDiffersFromTotal != 1 {
		t.Errorf("block/total disagreements = %d, want 1 (08/09/2025)",
			total.BlockDiffersFromTotal)
	}

	// Every registration resolves to exactly one type, since the owner's
	// 2026-08-02 ruling. The page shows all of them precisely so that a future
	// second type is visible rather than chosen between.
	for _, r := range rows {
		if len(r.Types) != 1 {
			t.Errorf("%s has types %v; one airframe has one type", r.Registration, r.Types)
		}
	}

	// The block total across all aircraft must equal the logbook's own total
	// time, because Total_Time == Block_Time on every row but the one above
	// (which is 7 minutes shorter on total than on block).
	whole := stats.Summarize(lb.Flights)
	if total.BlockMinutes != whole.Total+7 {
		t.Errorf("block total %d vs logbook total %d; expected exactly 7 minutes "+
			"apart, the single 08/09/2025 discrepancy",
			total.BlockMinutes, whole.Total)
	}
}
