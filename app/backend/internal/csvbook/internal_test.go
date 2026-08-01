package csvbook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The helpers below are defensive: through the CSV path their fallback arms
// cannot be reached, because the header is validated and every date is
// produced by time.Format. They are tested directly rather than deleted,
// because "unreachable" is a property of today's callers, not of the code.

func TestRecordGetIsSafeForColumnsTheBookDoesNotHave(t *testing.T) {
	// Book 1 has no Cumulative_Instructor column at all. An absent column must
	// read as empty, never as another column's value.
	r := record{fields: []string{"a", "b"}, col: map[string]int{"A": 0, "Beyond": 9}}

	if got := r.get("Cumulative_Instructor"); got != "" {
		t.Errorf("get(absent) = %q, want empty", got)
	}
	if got := r.get("Beyond"); got != "" {
		t.Errorf("get(index past the row) = %q, want empty", got)
	}
	if got := r.get("A"); got != "a" {
		t.Errorf("get(present) = %q, want a", got)
	}
	if r.has("Cumulative_Instructor") {
		t.Errorf("has(absent) must be false")
	}
}

func TestMinusYears(t *testing.T) {
	cases := []struct{ in, want string }{
		{"2026-07-30", "2024-07-30"},
		{"2011-05-25", "2009-05-25"},
		{"0001-01-01", "-001-01-01"}, // nonsense in, nonsense out; never a panic
		{"", ""},
		{"20x6-07-30", "20x6-07-30"},
	}
	for _, c := range cases {
		if got := minusYears(c.in, activeWithinYears); got != c.want {
			t.Errorf("minusYears(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDominantTypeBreaksTiesAlphabetically(t *testing.T) {
	if got := dominantType(map[string]int{"C172": 2, "C192": 1}); got != "C172" {
		t.Errorf("got %q, want the most-flown type C172", got)
	}
	// A tie must resolve the same way on every run, or the seed list would
	// change between two imports of identical data.
	if got := dominantType(map[string]int{"C172": 1, "C152": 1}); got != "C152" {
		t.Errorf("got %q, want the alphabetically first type C152", got)
	}
	if got := dominantType(nil); got != "" {
		t.Errorf("got %q, want empty for no observations", got)
	}
}

func TestSortDiscrepanciesOrdersByBookThenRowThenKind(t *testing.T) {
	ds := []Discrepancy{
		{Book: 2, Row: 5, Kind: KindUnknownType},
		{Book: 1, Row: 9, Kind: KindCumulativeBreak},
		{Book: 1, Row: 9, Kind: KindComponentOverTotal},
		{Book: 1, Row: 2, Kind: KindUnknownType},
	}
	sortDiscrepancies(ds)

	want := []Discrepancy{
		{Book: 1, Row: 2, Kind: KindUnknownType},
		{Book: 1, Row: 9, Kind: KindComponentOverTotal},
		{Book: 1, Row: 9, Kind: KindCumulativeBreak},
		{Book: 2, Row: 5, Kind: KindUnknownType},
	}
	for i := range want {
		if ds[i] != want[i] {
			t.Fatalf("position %d = %+v, want %+v", i, ds[i], want[i])
		}
	}
}

func TestAnchorHandlesABookWithoutTheInstructorColumn(t *testing.T) {
	// Book 1's 25-column layout has no Cumulative_Instructor. A seed row in
	// that layout must anchor the columns it does have and skip the one it
	// does not, rather than anchoring instructor time at zero.
	body := csvOf(header25,
		`"23/04/2017","P28A","OH-PDP","EFLA","EFHF","15:00","15:47","","","0:47","0:47","","","0:47","","","self","1","","395:49","312:42","83:07","3:12","58:39","889"`,
		`"24/04/2017","P28A","OH-PDP","EFHF","EFHF","10:00","11:00","","","1:00","1:00","","","1:00","","0:30","self","2","","396:49","313:42","83:07","3:12","58:39","891"`)

	lb := loadOne(t, body, Source{Book: 2, SkipSeedRow: true})
	if n := len(discrepanciesOf(lb, KindCumulativeBreak)); n != 0 {
		t.Errorf("got %d cumulative breaks, want 0: %+v", n, lb.Discrepancies)
	}
	if lb.Totals.Instructor != 30 {
		t.Errorf("Instructor = %d, want 30", lb.Totals.Instructor)
	}
}

func TestSeedRowWithUnreadableCumulativesIsAnError(t *testing.T) {
	cases := []struct{ name, seed, want string }{
		{
			name: "unreadable cumulative minutes",
			seed: `"23/04/2017","P28A","OH-PDP","EFLA","EFHF","15:00","15:47","","","0:47","0:47","","","0:47","","","self","1","","395","312:42","83:07","3:12","58:39","889",""`,
			want: "Cumulative_Total",
		},
		{
			name: "unreadable cumulative landings",
			seed: `"23/04/2017","P28A","OH-PDP","EFLA","EFHF","15:00","15:47","","","0:47","0:47","","","0:47","","","self","1","","395:49","312:42","83:07","3:12","58:39","many",""`,
			want: "Cumulative_Landings",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parseAll([]reader{{
				src:  Source{Book: 2, SkipSeedRow: true},
				body: strings.NewReader(csvOf(header26, c.seed)),
			}})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want one mentioning %q", err, c.want)
			}
		})
	}
}

func TestNegativeCountsAreRejected(t *testing.T) {
	// strconv.Atoi accepts a leading sign, so "-1" parses cleanly and would
	// otherwise subtract a landing from a legal total.
	body := csvOf(header25,
		`"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","M","-1","","0:57","","0:57","","","1"`)

	_, err := parseAll([]reader{{src: Source{Book: 1}, body: strings.NewReader(body)}})
	if err == nil || !strings.Contains(err.Error(), "Landings") {
		t.Fatalf("err = %v, want a Landings error", err)
	}
}

func TestBlankLandingCellCountsAsZero(t *testing.T) {
	// A blank cell in the book means "none", which is a real value. Only an
	// unreadable cell is an error.
	body := csvOf(header25,
		`"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","M","","","0:57","","0:57","","",""`)

	lb := loadOne(t, body, Source{Book: 1})
	if lb.Flights[0].LandingsDay != 0 {
		t.Errorf("LandingsDay = %d, want 0", lb.Flights[0].LandingsDay)
	}
	if n := len(discrepanciesOf(lb, KindCumulativeBreak)); n != 0 {
		t.Errorf("a blank landings column is consistent with a blank total, got %d breaks", n)
	}
}

func TestLandingCountBreakIsReported(t *testing.T) {
	// The landings series is a plain integer count, checked on the same terms
	// as the duration series.
	body := csvOf(header25,
		`"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","1:00","1:00","","","","1:00","","M","1","","1:00","","1:00","","","1"`,
		`"26/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","1:00","1:00","","","","1:00","","M","3","","2:00","","2:00","","","9"`,
		`"27/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","1:00","1:00","","","","1:00","","M","1","","3:00","","3:00","","","10"`)

	lb := loadOne(t, body, Source{Book: 1})
	breaks := discrepanciesOf(lb, KindCumulativeBreak)
	if len(breaks) != 1 {
		t.Fatalf("got %d breaks, want 1: %+v", len(breaks), breaks)
	}
	if breaks[0].Row != 3 || !strings.Contains(breaks[0].Detail, "Cumulative_Landings") {
		t.Errorf("break = %+v, want row 3 naming Cumulative_Landings", breaks[0])
	}
	// Totals count what the rows say, not what the column claims.
	if lb.Totals.Landings != 5 {
		t.Errorf("Landings = %d, want 5", lb.Totals.Landings)
	}
}

func TestDefaultSourcesNamesTheThreeBooks(t *testing.T) {
	got := DefaultSources("/repo")
	want := []Source{
		{Book: 1, Path: "/repo/logbook_1_final.csv"},
		{Book: 2, Path: "/repo/logbook_2_final.csv", SkipSeedRow: true},
		{Book: 3, Path: "/repo/logbook_3.csv", SkipSeedRow: true},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d sources, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("source %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	// Book 1 opens the series and has no seed row to skip.
	if got[0].SkipSeedRow {
		t.Errorf("book 1 has no carried-over row")
	}
}

func TestLoadReadsFromDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.csv")
	body := csvOf(header25,
		`"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","M","1","","0:57","","0:57","","","1"`)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	lb, err := Load([]Source{{Book: 1, Path: path}})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if lb.Totals.Flights != 1 {
		t.Errorf("Flights = %d, want 1", lb.Totals.Flights)
	}
}

func TestLoadReportsAMissingBook(t *testing.T) {
	_, err := Load([]Source{{Book: 3, Path: filepath.Join(t.TempDir(), "absent.csv")}})
	if err == nil {
		t.Fatal("a missing book must be an error, not an empty logbook")
	}
	if !strings.Contains(err.Error(), "book 3") {
		t.Errorf("err = %v, want it to name the book", err)
	}
}
