package csvbook_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/hhmm"
	"github.com/ramiayoub/logbook/backend/internal/timeutil"
)

// This file checks the loader against the actual paper record rather than a
// fixture. The unit tests prove the rules; this proves the rules produce the
// logbook the pilot flew.
//
// Every number here was derived from the CSVs and confirmed against the
// Cumulative_* columns the transcription effort maintained. If one of these
// assertions ever fails, the import is wrong until proven otherwise -- do not
// adjust the expectation to make it pass (CLAUDE.md rule 0.2).

// repoRoot walks up from this source file. The books live at the repository
// root and are committed, so they are always present.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	// .../app/backend/internal/csvbook/realdata_test.go -> repo root
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

func TestRealBooksProduceTheExpectedTotals(t *testing.T) {
	lb := loadReal(t)

	// 1296 flights, not 1298: the three CSVs hold 1298 data rows, but the
	// first row of Books 2 and 3 is the previous book's final flight carried
	// over to seed the cumulative columns.
	//
	// 1296, was 1293. On 2026-08-01 the owner found three flights of
	// 28/08/2025 (OH-ESR, EFNU-EFPR-EFPR-EFNU) missing from the CSV entirely
	// -- line 411 is 27/08/2025 and line 412 jumps to 08/09/2025. They were
	// reconstructed from the Aviatron aircraft-logbook screens, whose airframe
	// counter chains exactly across all three (2663:11 -> 2663:51 -> 2664:39
	// -> 2665:31), and appended as LATE ENTRIES at the end of Book 3 rather
	// than inserted in date order, which is how the paper book records them
	// too -- see the decision in claude-docs/drift.md.
	//
	// This is the ONE sanctioned exception to the frozen end-of-book-3
	// cumulatives: those figures were frozen against *corrections*, and this
	// is missing data, not a correction. The owner lifted the freeze
	// explicitly on 2026-08-01 for these three rows only. Deltas below.
	want := csvbook.Totals{
		Flights: 1296,
		// +2:35 -- the three flights' block times, 0:45 + 0:53 + 0:57.
		// Aviatron records AIR time (0:40/0:48/0:52); the owner ruled +5 min
		// per flight for block, and this book totals on block time (only one
		// row in 479 has Block_Time != Total_Time, and it is a flagged
		// discrepancy).
		Total: hm(1222, 10),
		// +1:42 -- flights 1 and 3 only. The middle flight is the SEP/IR
		// revalidation check ride: Tarhanen was PIC, so it logs dual, not PIC.
		PIC: hm(1054, 45),
		// +0:53 -- the check flight, dual in full.
		Dual: hm(167, 25),
		// 107:05, was 107:14. Book 1 line 28 (28/09/2011, OH-COF) logged 1:21 of
		// instrument on a 1:12 flight -- more instrument than flight time. The
		// owner ruled on 2026-08-01 that 1:12 is the reading, and the CSV was
		// corrected. Delta -0:09, and it moves no cumulative: the
		// Cumulative_Instrument column always advanced by 1:12, which is exactly
		// why the row was the outlier rather than the column.
		//
		// +0:53 on 2026-08-01: the 28/08/2025 SEP/IR revalidation check flight,
		// logged instrument in full at the owner's ruling. 107:05 -> 107:58.
		Instrument: hm(107, 58),
		// Night was 16:47 until 2026-08-01, against 22:45 inked at page 62.
		// The owner then read the paper's Yolentoaika column back and
		// photographed seven Book-1 spreads; its Siirto figures chain
		// continuously, which pins every night entry to a row. Six values were
		// added and one moved onto its correct row, taking the column to 20:50
		// and matching the paper's Siirto at every checkpoint through
		// 30/11/2013.
		//
		// The last 1:55 closed the same day when pages 52/53 were photographed
		// (IMG_6048): 25/02/2014 OH-KLS 0:55 (book 1 line 173, full night) and
		// 26/03/2014 OH-TIL 1:00 of 2:01 (line 177). That spread's own
		// Yolentoaika column runs Siirto 9:12 -> 11:07, and 11:07 is exactly
		// p.71's Siirto, so pages 54-69 carry no night at all -- the column is
		// closed, not merely sampled.
		//
		// 22:45 now equals the figure inked at page 62, delta 0:00. The owner
		// has frozen the end-of-book-3 cumulatives (claude-docs/resume.md); this
		// figure must not move again. Do not edit it to make a test pass.
		Night: hm(22, 45),
		// Instructor and SEPSea are unmoved by the 28/08/2025 late entries:
		// nobody was instructed, and OH-ESR is an SR20 landplane.
		Instructor: hm(189, 41),
		SEPSea:     hm(407, 39),
		// +5 -- 1 + 1 + 3.
		Landings: 3444,
	}
	if lb.Totals != want {
		t.Errorf("totals mismatch\n got %s\nwant %s", show(lb.Totals), show(want))
	}
}

// TestRealBooksReconcileAgainstTheCumulativeColumns is the checksum that makes
// the import trustworthy: the seven Cumulative_* series, recomputed row by row
// from the flights, must reproduce what the transcription recorded.
//
// It reconciles with ZERO breaks over 1293 rows and seven series. It did not
// always: book 1 line 28 (28/09/2011, OH-COF, EFHF local) logged 1:12 of flight
// and claimed 1:21 of instrument -- more instrument than flight, impossible --
// while its Cumulative_Instrument column advanced by the 1:12 actually flown.
// The importer surfaced it rather than correcting it, the owner ruled on
// 2026-08-01 that 1:12 is the reading, and the CSV was fixed.
//
// Zero is the assertion, not "one known defect". A break appearing here means
// either a new transcription batch is inconsistent or this code has regressed;
// both need triage, and neither is fixed by relaxing this number.
func TestRealBooksReconcileAgainstTheCumulativeColumns(t *testing.T) {
	lb := loadReal(t)

	var breaks []csvbook.Discrepancy
	for _, d := range lb.Discrepancies {
		if d.Kind == csvbook.KindCumulativeBreak {
			breaks = append(breaks, d)
		}
	}
	if len(breaks) != 0 {
		t.Fatalf("got %d cumulative breaks over 1293 flights and seven series, want zero:\n%s",
			len(breaks), format(breaks))
	}
}

func TestRealBooksSurfaceEveryKnownDataQualityItem(t *testing.T) {
	lb := loadReal(t)

	counts := map[csvbook.Kind]int{}
	for _, d := range lb.Discrepancies {
		counts[d.Kind]++
	}

	// Each of these is documented in app/docs/data-model.md. The counts are
	// asserted, not just the presence, so that a new occurrence in a future
	// Book 3 batch shows up as a failing test rather than slipping through.
	want := map[csvbook.Kind]int{
		// Both were book 1 line 28, and both closed when the owner ruled its
		// 1:21 of instrument down to the 1:12 actually flown. Kept in the map at
		// zero rather than deleted: an impossible row or a broken cumulative
		// reappearing must fail loudly, and the "unexpected kind" sweep below
		// only catches kinds that were never listed.
		csvbook.KindCumulativeBreak:    0,
		csvbook.KindComponentOverTotal: 0,
		csvbook.KindBlockTotalMismatch: 1,  // 08/09/2025, block 0:45 vs total 0:38
		csvbook.KindRegistrationFormat: 15, // SE-GKT x14, SE-LWI x1 -- both genuine.
		// Was 16: OK-PDP at book 2 line 102 was a transcription typo and the
		// owner ruled on 2026-08-01 that any OK- reg in these books is OH-.
		// Both were the same five cells and both closed together on 2026-08-02,
		// when the owner ruled that "C192" is a transcription typo for C172 --
		// there is no such Cessna -- and that OH-CMU is a C152 on every flight.
		// Kept in the map at zero rather than deleted, like the two above: a
		// type that is not an aircraft, or one registration flown as two types,
		// must fail loudly if it ever comes back, and the "unexpected kind"
		// sweep below only catches kinds that were never listed.
		csvbook.KindUnknownType:  0, // was 4: "C192" on OH-CTL x2 and OH-GKT x2
		csvbook.KindTypeConflict: 0, // was 3: OH-CTL, OH-GKT, OH-CMU
		csvbook.KindDateFormat:   8, // book 2 lines 83-90, transcribed DD.MM.YYYY
		// One per row carrying night time, which is what Task 8 must backfill.
		// Was 22, then 28 when the night reconciliation of 2026-08-01 added six
		// values, then 30 when the p.52/53 photograph added the last two
		// (25/02/2014 OH-KLS 0:55 and 26/03/2014 OH-TIL 1:00). That is every
		// night row in the three books: 20 in book 1, 3 in book 2, 7 in book 3.
		csvbook.KindLandingsUnverified: 30,
	}
	for kind, n := range want {
		if counts[kind] != n {
			t.Errorf("%s: got %d, want %d", kind, counts[kind], n)
		}
	}
	for kind, n := range counts {
		if _, expected := want[kind]; !expected {
			t.Errorf("unexpected discrepancy kind %s (%d occurrences); a new class of "+
				"data problem must be triaged, not absorbed", kind, n)
		}
	}
}

// TestEveryRegistrationNamesOneRealAircraftType is the permanent close of the
// owner's ruling of 2026-08-02.
//
// Five cells across two books said something that could not be true. Four rows
// gave OH-GKT and OH-CTL the type "C192" -- Cessna has never built a 192, and
// the very next flight of the same day in the same aeroplane (book 2 line 139)
// is written C172. One row gave OH-CMU as a C172 when it is a C152 on all of
// its other flights. Both were transcription slips, and the owner ruled them
// closed rather than left standing as open questions.
//
// This is asserted here, in the language of the ruling, as well as through the
// two discrepancy kinds now held at zero above. Two independent statements of
// one fact, because a guard that lives only inside the discrepancy machinery
// disappears the day somebody prunes a kind from that map.
//
// NOTE this is NOT a licence figure and never was: the sea/land split -- the
// thing a seaplane rating is evidenced by -- comes from the REGISTRATION, not
// the type (see IsSea and the 2026-08-01 decision-log entry). Correcting these
// five cells moved no total, no landing count and no class. That is exactly why
// the freeze could be lifted for them.
func TestEveryRegistrationNamesOneRealAircraftType(t *testing.T) {
	lb := loadReal(t)

	typesOf := map[string]map[string]bool{}
	for _, f := range lb.Flights {
		// "C192" must never reappear, under any registration.
		if f.AircraftType == "C192" {
			t.Errorf("book %d line %d (%s, %s): type C192 is not an aircraft; "+
				"the owner ruled it a typo for C172 on 2026-08-02",
				f.SourceBook, f.SourceRow, f.Date, f.AircraftReg)
		}
		if typesOf[f.AircraftReg] == nil {
			typesOf[f.AircraftReg] = map[string]bool{}
		}
		typesOf[f.AircraftReg][f.AircraftType] = true
	}

	// One registration, one type. An aeroplane does not change model.
	for reg, types := range typesOf {
		if len(types) != 1 {
			t.Errorf("%s is flown as %d different types %v; one airframe has one type",
				reg, len(types), sortedKeys(types))
		}
	}

	// The specific rulings, named, so a future reader sees what was decided
	// rather than only that something was.
	if got := sortedKeys(typesOf["OH-CMU"]); len(got) != 1 || got[0] != "C152" {
		t.Errorf("OH-CMU is a C152 on every flight (owner, 2026-08-02); got %v", got)
	}
	for _, reg := range []string{"OH-GKT", "OH-CTL"} {
		if got := sortedKeys(typesOf[reg]); len(got) != 1 || got[0] != "C172" {
			t.Errorf("%s is a C172 on every flight (owner, 2026-08-02); got %v", reg, got)
		}
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestRealBooksResolveEveryTimeToUTC guards the conversion authority against
// the whole corpus at once: 1293 rows spanning 2011-2026, every DST changeover
// in between.
func TestRealBooksResolveEveryTimeToUTC(t *testing.T) {
	lb := loadReal(t)

	origins := map[timeutil.Origin]int{}
	for _, f := range lb.Flights {
		origins[f.TimeOrigin]++

		if f.OffBlockUTC.IsZero() != f.OnBlockUTC.IsZero() {
			t.Errorf("seq %d: half a block pair is missing", f.Seq)
		}
		if f.OffBlockUTC.IsZero() {
			continue
		}
		d := f.OnBlockUTC.Sub(f.OffBlockUTC).Minutes()
		if d <= 0 {
			t.Errorf("seq %d (%s): block pair spans %v minutes", f.Seq, f.Date, d)
		}
		// Every flight in these books is a light-aircraft leg. A block pair
		// more than a day long would mean a botched midnight roll.
		if d > 24*60 {
			t.Errorf("seq %d (%s): block pair spans %v minutes, which is implausible",
				f.Seq, f.Date, d)
		}
	}

	if origins[timeutil.OriginUnknown] != 0 {
		t.Errorf("%d rows could not be resolved to UTC with confidence",
			origins[timeutil.OriginUnknown])
	}
	if origins[timeutil.OriginNone] != 0 {
		t.Errorf("%d rows have no block times at all", origins[timeutil.OriginNone])
	}
	// Books 1 and 2 are local throughout; Book 3 switched to Zulu partway.
	if origins[timeutil.OriginUTCAsWritten] == 0 || origins[timeutil.OriginConvertedFromLocal] == 0 {
		t.Errorf("origins = %v, want both zones represented", origins)
	}
}

func TestRealBooksHaveOneContinuousSeq(t *testing.T) {
	lb := loadReal(t)

	for i, f := range lb.Flights {
		if f.Seq != i+1 {
			t.Fatalf("flight %d has seq %d; the ordering key must be dense and 1-based", i, f.Seq)
		}
	}
	// Provenance must be unique -- two flights claiming the same CSV line
	// would mean a row was read twice.
	seen := map[[2]int]bool{}
	for _, f := range lb.Flights {
		key := [2]int{f.SourceBook, f.SourceRow}
		if seen[key] {
			t.Fatalf("book %d row %d imported twice", f.SourceBook, f.SourceRow)
		}
		seen[key] = true
	}
	// The seed rows must be absent: Book 2 and Book 3 both start at line 3.
	for _, b := range []int{2, 3} {
		if seen[[2]int{b, 2}] {
			t.Errorf("book %d line 2 is the carried-over seed row and must not be a flight", b)
		}
		if !seen[[2]int{b, 3}] {
			t.Errorf("book %d line 3 is missing", b)
		}
	}
	if !seen[[2]int{1, 2}] {
		t.Errorf("book 1 line 2 is a real first flight and must be imported")
	}
}

// TestRealBooksSplitLandingsConsistently asserts the invariant from
// data-model.md: sea+land and day+night partition the same grand total on
// different axes, so they must agree.
func TestRealBooksSplitLandingsConsistently(t *testing.T) {
	lb := loadReal(t)

	var sea, land, day, night int
	for _, f := range lb.Flights {
		n := f.LandingsDay + f.LandingsNight
		switch f.Class {
		case csvbook.ClassSEPSea:
			sea += n
		case csvbook.ClassSEPLand:
			land += n
		default:
			t.Fatalf("seq %d has unclassified class %q", f.Seq, f.Class)
		}
		day += f.LandingsDay
		night += f.LandingsNight
	}
	if sea+land != day+night {
		t.Errorf("sea+land = %d but day+night = %d", sea+land, day+night)
	}
	if sea+land != lb.Totals.Landings {
		t.Errorf("landings partition = %d but Totals.Landings = %d", sea+land, lb.Totals.Landings)
	}
	// Nothing is read off paper yet, so the night column is still empty and
	// Task 8 has 30 rows of work waiting.
	if night != 0 {
		t.Errorf("night landings = %d; the split has not been backfilled yet", night)
	}
}

func TestRealBooksDeriveTheAircraftSeedList(t *testing.T) {
	lb := loadReal(t)

	// 38, not 39: OK-PDP was a typo for OH-PDP and was corrected in the CSV on
	// 2026-08-01, so it no longer seeds a phantom one-flight aircraft.
	if len(lb.Aircraft) != 38 {
		t.Errorf("got %d aircraft, want 38 distinct registrations", len(lb.Aircraft))
	}
	if !sort.SliceIsSorted(lb.Aircraft, func(i, j int) bool {
		return lb.Aircraft[i].Registration < lb.Aircraft[j].Registration
	}) {
		t.Errorf("the seed list must be sorted so two imports produce identical rows")
	}

	byReg := map[string]csvbook.Aircraft{}
	for _, a := range lb.Aircraft {
		byReg[a.Registration] = a
	}
	for _, c := range []struct {
		reg   string
		typ   string
		class csvbook.Class
		ifr   bool
	}{
		// Both were seeded C172 by dominantType even while the four C192 rows
		// existed; since the 2026-08-02 ruling every row agrees outright. The
		// mechanism is still exercised, on a fixture, in csvbook_test.go.
		{"OH-CTL", "C172", csvbook.ClassSEPSea, false},
		{"OH-GKT", "C172", csvbook.ClassSEPSea, false},
		{"SE-GKT", "C172", csvbook.ClassSEPSea, false},  // same airframe, earlier reg
		{"OH-CDK", "C185", csvbook.ClassSEPSea, false},  // floatplane
		{"OH-MIL", "M6", csvbook.ClassSEPSea, false},    // Maule, always on floats
		{"OH-ESR", "SR20", csvbook.ClassSEPLand, true},  // IFR
		{"OH-CAM", "C172", csvbook.ClassSEPLand, true},  // IFR
		{"OH-PIF", "P28A", csvbook.ClassSEPLand, true},  // the IR trainer
		{"OH-COF", "C152", csvbook.ClassSEPLand, false}, // has instrument time, not IFR certified
		{"OH-CMU", "C152", csvbook.ClassSEPLand, false}, // the disputed one
	} {
		a, ok := byReg[c.reg]
		if !ok {
			t.Errorf("%s missing from the seed list", c.reg)
			continue
		}
		if a.Type != c.typ || a.DefaultClass != c.class || a.IFRCapable != c.ifr {
			t.Errorf("%s = {%s %s ifr=%v}, want {%s %s ifr=%v}",
				c.reg, a.Type, a.DefaultClass, a.IFRCapable, c.typ, c.class, c.ifr)
		}
	}

	// The registrations the owner needs a note on must carry one.
	// OK-PDP is deliberately absent: it was corrected to OH-PDP in the CSV on
	// 2026-08-01, so there is no longer an aircraft row needing a note.
	for _, reg := range []string{"OH-GKT", "SE-GKT", "OH-CMU", "OH-CMV", "OH-MIL"} {
		if byReg[reg].Notes == "" {
			t.Errorf("%s should carry an explanatory note", reg)
		}
	}

	// The last flight in the books is 30/07/2026 in OH-GKT, so it is current;
	// OH-KLS was last flown in 2012 and must drop out of the form's list.
	if !byReg["OH-GKT"].Active {
		t.Errorf("OH-GKT is the pilot's current aircraft")
	}
	if byReg["OH-KLS"].Active {
		t.Errorf("OH-KLS has not been flown since 2012")
	}
}

func hm(h, m int) int { return h*60 + m }

func show(t csvbook.Totals) string {
	return fmt.Sprintf("flights=%d total=%s pic=%s dual=%s instr=%s night=%s fi=%s sea=%s ldg=%d",
		t.Flights, hhmm.Format(t.Total), hhmm.Format(t.PIC), hhmm.Format(t.Dual),
		hhmm.Format(t.Instrument), hhmm.Format(t.Night), hhmm.Format(t.Instructor),
		hhmm.Format(t.SEPSea), t.Landings)
}

func format(ds []csvbook.Discrepancy) string {
	out := ""
	for _, d := range ds {
		out += fmt.Sprintf("  book %d row %d (%s) %s: %s\n", d.Book, d.Row, d.Date, d.Kind, d.Detail)
	}
	return out
}
