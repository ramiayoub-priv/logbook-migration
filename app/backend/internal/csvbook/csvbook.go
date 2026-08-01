// Package csvbook turns the three transcribed logbook CSVs into the domain
// records the application stores, and audits them while doing so.
//
// It is deliberately free of any database: parsing, classification and
// verification are pure functions over the CSV text, so the rules that decide
// what a legal record says are testable without a schema in the way.
//
// Two things this package will never do (CLAUDE.md rule 0.2):
//
//   - drop a row. A row that fails a consistency check is still imported,
//     exactly as written on paper, and reported alongside.
//   - correct a value. Every disagreement between the CSV and what the numbers
//     ought to say comes out as a Discrepancy for the owner to rule on.
//
// It also never imports the seven Cumulative_* columns (rule 0.5). It reads
// them, reconciles them row by row against a running total recomputed from the
// flights themselves, and then discards them.
package csvbook

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/hhmm"
	"github.com/ramiayoub/logbook/backend/internal/timeutil"
)

// Class is the airframe configuration a flight was actually flown in. It is
// what the seaplane/landplane statistics split on.
//
// Only single-engine piston appears in these books; the wider EASA set is
// declared so the column has a fixed vocabulary from the start.
type Class string

const (
	ClassSEPLand Class = "SEP_LAND"
	ClassSEPSea  Class = "SEP_SEA"
	ClassMEPLand Class = "MEP_LAND"
	ClassMEPSea  Class = "MEP_SEA"
	ClassTMG     Class = "TMG"
)

// seaplaneRegistrations are the registrations flown on floats, from
// claude-docs/reference.md.
//
// The classification is per registration rather than per type because the book
// writes float C172s as "C172sea" only from IMG_6022 onward, and inconsistently
// even then -- the registration is the reliable signal (user-confirmed
// 2026-08-01). This set is verified, not assumed: reconciling it row by row
// against the Cumulative_SEP_Sea column reproduces all 407:39 of seaplane time
// over 1293 flights with no break, which pins every individual row.
var seaplaneRegistrations = map[string]bool{
	"OH-CTL": true,
	"SE-GKT": true,
	"OH-GKT": true,
	"OH-PAX": true,
	"OH-MIL": true,
	"OH-CTE": true,
	"OH-CDK": true,
}

// ifrCapableRegistrations drives whether the new-flight form offers an
// instrument-time field. It is a UI hint and never a constraint on the data.
//
// It is a curated list rather than "any aircraft that has logged instrument
// time", because instrument time is also logged under the hood in aircraft that
// are not IFR certified -- OH-COF and OH-CTH are C152s with instrument rows.
// These three are confirmed in claude-docs/reference.md.
var ifrCapableRegistrations = map[string]bool{
	"OH-CAM": true, // C172, IFR certified (user-confirmed 2026-07-31)
	"OH-ESR": true, // SR20, routinely IFR on cross-countries
	"OH-PIF": true, // P28A, the CB-IR instrument trainer
}

// knownTypes is the set of aircraft types these books legitimately contain.
// Anything outside it is surfaced rather than corrected: "C192" is not a Cessna
// and appears on four rows, almost certainly for C172.
var knownTypes = map[string]bool{
	"C150": true, "C152": true, "C172": true, "C185": true, "C206": true,
	"DA40": true, "M6": true, "P28A": true, "P32R": true, "SR20": true,
}

// activeWithinYears is how recently an aircraft must have been flown, counting
// back from the last flight in the books, to stay in the new-flight form's
// default list. Retired aircraft keep their rows and their history.
//
// It counts back from the newest flight rather than from today so that the seed
// list is reproducible: importing the same CSVs must give the same DB whenever
// it is run (CLAUDE.md rule 0.2, idempotence).
const activeWithinYears = 2

// Flight is one row of the paper logbook as the application stores it.
// Durations are whole minutes; instants are UTC. See app/docs/data-model.md.
type Flight struct {
	Seq          int    // explicit book order, 1..n over all three books
	Date         string // YYYY-MM-DD, the date as written in the book
	AircraftType string // as written on paper, not as the aircraft table says
	AircraftReg  string
	Class        Class
	DepPlace     string
	ArrPlace     string

	OffBlockUTC time.Time
	OnBlockUTC  time.Time
	OffBlockRaw string // exactly as written: "15:30" or "07:56Z"
	OnBlockRaw  string
	TimeOrigin  timeutil.Origin
	TakeoffUTC  time.Time
	LandingUTC  time.Time

	BlockMinutes      int
	TotalMinutes      int
	NightMinutes      int
	InstrumentMinutes int
	PICMinutes        int
	DualMinutes       int
	InstructorMinutes int
	CopilotMinutes    int // not in the CSV; the EASA PDF has the column
	MultiPilotMinutes int

	PICName          string
	LandingsDay      int
	LandingsNight    int
	LandingsVerified bool
	Remarks          string

	SourceBook int // 1, 2 or 3
	SourceRow  int // 1-based line number in that CSV, header included
}

// Aircraft is a row of the seed list that makes the new-flight form smart. It
// is derived from the flights, so it can never drift from what was flown.
type Aircraft struct {
	Registration string
	Type         string // the most-flown type for this registration
	DefaultClass Class
	IFRCapable   bool
	Active       bool
	Notes        string
}

// Totals is the checksum set the importer verifies against after writing. Every
// figure is whole minutes except Flights and Landings.
type Totals struct {
	Flights    int
	Total      int
	PIC        int
	Dual       int
	Instrument int
	Night      int
	Instructor int
	SEPSea     int
	Landings   int
}

// Kind names a class of problem found in the source data.
type Kind string

const (
	// KindCumulativeBreak - a Cumulative_* column disagrees with the running
	// total recomputed from the rows. Either the row or the column is wrong.
	KindCumulativeBreak Kind = "cumulative_break"
	// KindComponentOverTotal - a component time exceeds the flight's total,
	// which is impossible.
	KindComponentOverTotal Kind = "component_exceeds_total"
	// KindBlockTotalMismatch - Block_Time and Total_Time disagree.
	KindBlockTotalMismatch Kind = "block_total_mismatch"
	// KindUnknownTimeOrigin - a clock time could not be resolved to UTC with
	// confidence (a DST gap or fold, or a pair that mixes zones).
	KindUnknownTimeOrigin Kind = "unknown_time_origin"
	// KindRegistrationFormat - a registration that is not Finnish OH-xxx.
	KindRegistrationFormat Kind = "registration_format"
	// KindUnknownType - an aircraft type outside the known set.
	KindUnknownType Kind = "unknown_aircraft_type"
	// KindTypeConflict - one registration written with two different types.
	KindTypeConflict Kind = "type_conflict"
	// KindDateFormat - the date was written DD.MM.YYYY rather than DD/MM/YYYY.
	KindDateFormat Kind = "date_format"
	// KindLandingsUnverified - the day/night landing split was inferred rather
	// than read off the page, because the row carries night time.
	KindLandingsUnverified Kind = "landings_unverified"
)

// Discrepancy is something the source data says that the importer will not act
// on by itself. Every one of these is traceable to a paper page via Book/Row.
type Discrepancy struct {
	Kind   Kind
	Book   int
	Row    int
	Date   string
	Detail string
}

// Logbook is everything one load produced.
type Logbook struct {
	Flights       []Flight
	Aircraft      []Aircraft
	Totals        Totals
	Discrepancies []Discrepancy
}

// Source is one book's CSV.
type Source struct {
	Book int
	Path string
	// SkipSeedRow drops the first data row. In Books 2 and 3 that row is the
	// previous book's final flight, carried over to seed the cumulative
	// columns; importing it would enter three flights twice.
	SkipSeedRow bool
}

// DefaultSources are the three books as they sit in the repository root.
func DefaultSources(dir string) []Source {
	return []Source{
		{Book: 1, Path: filepath.Join(dir, "logbook_1_final.csv")},
		{Book: 2, Path: filepath.Join(dir, "logbook_2_final.csv"), SkipSeedRow: true},
		{Book: 3, Path: filepath.Join(dir, "logbook_3.csv"), SkipSeedRow: true},
	}
}

// Load reads the given books in order and returns one continuous logbook.
func Load(sources []Source) (*Logbook, error) {
	readers := make([]reader, 0, len(sources))
	defer func() {
		for _, r := range readers {
			if c, ok := r.body.(io.Closer); ok {
				c.Close()
			}
		}
	}()
	for _, src := range sources {
		f, err := os.Open(src.Path)
		if err != nil {
			return nil, fmt.Errorf("csvbook: book %d: %w", src.Book, err)
		}
		readers = append(readers, reader{src: src, body: f})
	}
	return parseAll(readers)
}

type reader struct {
	src  Source
	body io.Reader
}

// required lists the columns every book must carry. Cumulative_Instructor is
// absent from Book 1 and therefore not required.
var required = []string{
	"Date", "Aircraft_Type", "Aircraft_Reg", "Departure", "Arrival",
	"Off_Block", "On_Block", "Takeoff", "Landing", "Block_Time", "Total_Time",
	"Instrument_Time", "Night_Time", "PIC_Time", "Student_Time",
	"Instructor_Time", "pic_name", "Landings", "Remarks",
	"Cumulative_Total", "Cumulative_PIC", "Cumulative_Student",
	"Cumulative_Instrument", "Cumulative_SEP_Sea", "Cumulative_Landings",
}

func parseAll(readers []reader) (*Logbook, error) {
	lb := &Logbook{}
	rec := newReconciler()
	seen := newRegistry()

	for _, r := range readers {
		if err := parseBook(lb, rec, seen, r); err != nil {
			return nil, err
		}
	}

	lb.Aircraft = seen.aircraft(lb.Flights)
	sortDiscrepancies(lb.Discrepancies)
	return lb, nil
}

func parseBook(lb *Logbook, rec *reconciler, seen *registry, r reader) error {
	cr := csv.NewReader(r.body)
	cr.FieldsPerRecord = 0 // the header fixes the width; ragged rows are errors

	header, err := cr.Read()
	if err != nil {
		return fmt.Errorf("csvbook: book %d: reading header: %w", r.src.Book, err)
	}
	col := map[string]int{}
	for i, name := range header {
		col[strings.TrimSpace(name)] = i
	}
	for _, name := range required {
		if _, ok := col[name]; !ok {
			return fmt.Errorf("csvbook: book %d: missing required column %q", r.src.Book, name)
		}
	}

	line := 1 // the header was line 1
	rows := 0
	for {
		rowStrings, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("csvbook: book %d: %w", r.src.Book, err)
		}
		line++
		rows++

		row := record{fields: rowStrings, col: col, book: r.src.Book, line: line}
		if rows == 1 && r.src.SkipSeedRow {
			// The seed row is not a flight, but its cumulative values anchor
			// the running series that the rest of the book is checked against.
			if err := rec.anchor(row); err != nil {
				return err
			}
			continue
		}

		f, err := row.flight(len(lb.Flights) + 1)
		if err != nil {
			return err
		}
		lb.Flights = append(lb.Flights, f)
		lb.Totals.add(f)
		seen.observe(f)

		lb.Discrepancies = append(lb.Discrepancies, seen.conflicts(f)...)
		lb.Discrepancies = append(lb.Discrepancies, rowChecks(f, row)...)
		ds, err := rec.step(row, f)
		if err != nil {
			return err
		}
		lb.Discrepancies = append(lb.Discrepancies, ds...)
	}

	if rows == 0 {
		return fmt.Errorf("csvbook: book %d: no data rows", r.src.Book)
	}
	return nil
}

func (t *Totals) add(f Flight) {
	t.Flights++
	t.Total += f.TotalMinutes
	t.PIC += f.PICMinutes
	t.Dual += f.DualMinutes
	t.Instrument += f.InstrumentMinutes
	t.Night += f.NightMinutes
	t.Instructor += f.InstructorMinutes
	if f.Class == ClassSEPSea {
		t.SEPSea += f.TotalMinutes
	}
	t.Landings += f.LandingsDay + f.LandingsNight
}

// record is one CSV line with the column index it was read under.
type record struct {
	fields []string
	col    map[string]int
	book   int
	line   int
}

func (r record) get(name string) string {
	i, ok := r.col[name]
	if !ok || i >= len(r.fields) {
		return ""
	}
	return strings.TrimSpace(r.fields[i])
}

// has reports whether the book carries the column at all. Book 1 has no
// Cumulative_Instructor, and an absent column is not a zero.
func (r record) has(name string) bool {
	_, ok := r.col[name]
	return ok
}

func (r record) fail(name, why string) error {
	return fmt.Errorf("csvbook: book %d line %d: %s: %s", r.book, r.line, name, why)
}

func (r record) minutes(name string) (int, error) {
	m, err := hhmm.Parse(r.get(name))
	if err != nil {
		return 0, r.fail(name, err.Error())
	}
	return m, nil
}

func (r record) count(name string) (int, error) {
	s := r.get(name)
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, r.fail(name, fmt.Sprintf("%q is not a whole non-negative count", s))
	}
	return n, nil
}

// date reads the Date cell. The schema is DD/MM/YYYY.
//
// Eight consecutive Book 2 rows (lines 83-90) were transcribed as DD.MM.YYYY
// instead. Those are read day-first, which is the Finnish convention and is
// corroborated inside the batch itself: six of the eight have a day above 12
// and so cannot be read any other way. Every dotted row is reported regardless
// (see dateDoubt), and the two whose day field does not settle the question on
// its own are reported as needing confirmation against the paper.
//
// Day-first is applied rather than refused because refusing would block the
// import of 1291 sound rows over a separator; reported rather than corrected
// because the CSV belongs to the transcription effort (CLAUDE.md rule 0.2).
func (r record) date() (string, error) {
	raw := r.get("Date")
	for _, layout := range []string{"02/01/2006", "02.01.2006"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", r.fail("Date", fmt.Sprintf("%q is not a DD/MM/YYYY date", raw))
}

// dateDoubt describes what is wrong with how the Date cell was written, or ""
// if it followed the schema.
func (r record) dateDoubt() string {
	raw := r.get("Date")
	if !strings.Contains(raw, ".") {
		return ""
	}
	// A day above 12 cannot be a month, so the reading is certain.
	if t, err := time.Parse("02.01.2006", raw); err == nil && t.Day() > 12 {
		return fmt.Sprintf("date written %q; the schema is DD/MM/YYYY. Read day-first, "+
			"which the day field settles on its own", raw)
	}
	return fmt.Sprintf("date written %q; the schema is DD/MM/YYYY. Read day-first, but the "+
		"day field does not settle DD.MM vs MM.DD -- CONFIRM AGAINST THE PAPER", raw)
}

// flight maps one record onto the domain. Anything unreadable is an error, not
// a zero: a swallowed cell would silently shrink a legal total.
func (r record) flight(seq int) (Flight, error) {
	var f Flight
	var err error

	if f.Date, err = r.date(); err != nil {
		return f, err
	}

	f.Seq = seq
	f.AircraftType = r.get("Aircraft_Type")
	f.AircraftReg = r.get("Aircraft_Reg")
	f.Class = classOf(f.AircraftReg)
	f.DepPlace = r.get("Departure")
	f.ArrPlace = r.get("Arrival")
	f.PICName = r.get("pic_name")
	f.Remarks = r.get("Remarks")
	f.SourceBook = r.book
	f.SourceRow = r.line

	f.OffBlockRaw = r.get("Off_Block")
	f.OnBlockRaw = r.get("On_Block")
	f.OffBlockUTC, f.OnBlockUTC, f.TimeOrigin, err = timeutil.BlockPair(f.Date, f.OffBlockRaw, f.OnBlockRaw)
	if err != nil {
		return f, r.fail("Off_Block/On_Block", err.Error())
	}

	var airborne timeutil.Origin
	f.TakeoffUTC, f.LandingUTC, airborne, err = timeutil.BlockPair(f.Date, r.get("Takeoff"), r.get("Landing"))
	if err != nil {
		return f, r.fail("Takeoff/Landing", err.Error())
	}
	// An airborne pair we could not resolve taints the whole row: the flight
	// carries one time_origin, and it must reflect the least certain value on
	// the row rather than the most convenient one.
	if airborne == timeutil.OriginUnknown {
		f.TimeOrigin = timeutil.OriginUnknown
	}

	for _, m := range []struct {
		col string
		dst *int
	}{
		{"Block_Time", &f.BlockMinutes},
		{"Total_Time", &f.TotalMinutes},
		{"Night_Time", &f.NightMinutes},
		{"Instrument_Time", &f.InstrumentMinutes},
		{"PIC_Time", &f.PICMinutes},
		{"Student_Time", &f.DualMinutes}, // Oppilas/student is EASA dual
		{"Instructor_Time", &f.InstructorMinutes},
	} {
		if *m.dst, err = r.minutes(m.col); err != nil {
			return f, err
		}
	}

	// The CSV only ever captured the landing sum. Seed it as day and mark the
	// split unverified wherever night time makes the assumption unsafe -- a
	// flight with night time may still have landed by day. Backfilled from the
	// page images in Task 8.
	if f.LandingsDay, err = r.count("Landings"); err != nil {
		return f, err
	}
	f.LandingsVerified = f.NightMinutes == 0

	return f, nil
}

func classOf(reg string) Class {
	if seaplaneRegistrations[strings.ToUpper(reg)] {
		return ClassSEPSea
	}
	return ClassSEPLand
}

// rowChecks are the invariants that need only the row itself. It takes the
// record as well as the flight because a few checks are about how the cell was
// written, which the domain record deliberately no longer remembers.
func rowChecks(f Flight, r record) []Discrepancy {
	var out []Discrepancy
	at := func(kind Kind, format string, args ...any) {
		out = append(out, Discrepancy{
			Kind: kind, Book: f.SourceBook, Row: f.SourceRow, Date: f.Date,
			Detail: fmt.Sprintf(format, args...),
		})
	}

	for _, c := range []struct {
		name string
		val  int
	}{
		{"Instrument_Time", f.InstrumentMinutes},
		{"Night_Time", f.NightMinutes},
		{"PIC_Time", f.PICMinutes},
		{"Student_Time", f.DualMinutes},
		{"Instructor_Time", f.InstructorMinutes},
	} {
		if c.val > f.TotalMinutes {
			at(KindComponentOverTotal, "%s %s exceeds Total_Time %s on %s %s",
				c.name, hhmm.Format(c.val), hhmm.Format(f.TotalMinutes), f.AircraftReg, f.Date)
		}
	}

	if f.BlockMinutes != f.TotalMinutes {
		at(KindBlockTotalMismatch, "Block_Time %s but Total_Time %s; totals follow Total_Time",
			hhmm.Format(f.BlockMinutes), hhmm.Format(f.TotalMinutes))
	}

	if f.TimeOrigin == timeutil.OriginUnknown {
		at(KindUnknownTimeOrigin, "%q-%q could not be resolved to UTC with confidence",
			f.OffBlockRaw, f.OnBlockRaw)
	}

	if !isFinnishRegistration(f.AircraftReg) {
		at(KindRegistrationFormat, "registration %q is not Finnish OH-xxx", f.AircraftReg)
	}

	if !knownTypes[f.AircraftType] {
		at(KindUnknownType, "aircraft type %q is not a known type", f.AircraftType)
	}

	if doubt := r.dateDoubt(); doubt != "" {
		at(KindDateFormat, "%s", doubt)
	}

	if !f.LandingsVerified {
		at(KindLandingsUnverified, "%d landing(s) seeded as day on a row carrying %s of night time",
			f.LandingsDay, hhmm.Format(f.NightMinutes))
	}

	return out
}

func isFinnishRegistration(reg string) bool {
	return len(reg) == 6 && strings.HasPrefix(reg, "OH-")
}

// reconciler recomputes each Cumulative_* series from the flights and compares
// it to the column the CSV carries, row by row.
//
// Row by row rather than only at the end, because an end-total check can be
// passed by two errors that cancel; and because a break needs a line number to
// be worth anything.
type reconciler struct {
	running  map[string]int
	landings int
}

// series pairs each cumulative column with the per-flight value that feeds it.
var series = []struct {
	col string
	of  func(Flight) int
}{
	{"Cumulative_Total", func(f Flight) int { return f.TotalMinutes }},
	{"Cumulative_PIC", func(f Flight) int { return f.PICMinutes }},
	{"Cumulative_Student", func(f Flight) int { return f.DualMinutes }},
	{"Cumulative_Instrument", func(f Flight) int { return f.InstrumentMinutes }},
	{"Cumulative_Instructor", func(f Flight) int { return f.InstructorMinutes }},
	{"Cumulative_SEP_Sea", func(f Flight) int {
		if f.Class == ClassSEPSea {
			return f.TotalMinutes
		}
		return 0
	}},
}

func newReconciler() *reconciler {
	return &reconciler{running: map[string]int{}}
}

// anchor takes the running series from a seed row without counting it as a
// flight, so Book 2 and Book 3 continue Book 1's numbers rather than restarting.
func (rc *reconciler) anchor(r record) error {
	for _, s := range series {
		if !r.has(s.col) {
			continue
		}
		v, err := r.minutes(s.col)
		if err != nil {
			return err
		}
		rc.running[s.col] = v
	}
	n, err := r.count("Cumulative_Landings")
	if err != nil {
		return err
	}
	rc.landings = n
	return nil
}

func (rc *reconciler) step(r record, f Flight) ([]Discrepancy, error) {
	var out []Discrepancy
	break_ := func(col, ours, theirs string) {
		out = append(out, Discrepancy{
			Kind: KindCumulativeBreak, Book: f.SourceBook, Row: f.SourceRow, Date: f.Date,
			Detail: fmt.Sprintf("%s: rows sum to %s but the column says %s", col, ours, theirs),
		})
	}

	for _, s := range series {
		if !r.has(s.col) {
			continue
		}
		rc.running[s.col] += s.of(f)
		want, err := r.minutes(s.col)
		if err != nil {
			return nil, err
		}
		if got := rc.running[s.col]; got != want {
			break_(s.col, hhmm.Format(got), hhmm.Format(want))
			// Re-anchor on the CSV so one bad row reports once instead of
			// making every later row look broken.
			rc.running[s.col] = want
		}
	}

	rc.landings += f.LandingsDay + f.LandingsNight
	want, err := r.count("Cumulative_Landings")
	if err != nil {
		return nil, err
	}
	if rc.landings != want {
		break_("Cumulative_Landings", strconv.Itoa(rc.landings), strconv.Itoa(want))
		rc.landings = want
	}

	return out, nil
}

// registry accumulates what has been seen per registration, so the aircraft
// seed list is derived from the flights rather than hand-maintained.
type registry struct {
	types map[string]map[string]int
	order []string
}

func newRegistry() *registry {
	return &registry{types: map[string]map[string]int{}}
}

func (g *registry) observe(f Flight) {
	if _, ok := g.types[f.AircraftReg]; !ok {
		g.types[f.AircraftReg] = map[string]int{}
		g.order = append(g.order, f.AircraftReg)
	}
	g.types[f.AircraftReg][f.AircraftType]++
}

// conflicts reports a registration written with a second type, once, on the row
// that introduces it. OH-CMU appears as both C152 and C172 and reference.md
// warns that OH-CMU and OH-CMV are genuinely different aircraft, so this needs
// the owner's eye rather than a majority vote.
func (g *registry) conflicts(f Flight) []Discrepancy {
	seenTypes := g.types[f.AircraftReg]
	if len(seenTypes) < 2 || seenTypes[f.AircraftType] != 1 {
		return nil
	}
	others := make([]string, 0, len(seenTypes))
	for t := range seenTypes {
		if t != f.AircraftType {
			others = append(others, t)
		}
	}
	sort.Strings(others)
	return []Discrepancy{{
		Kind: KindTypeConflict, Book: f.SourceBook, Row: f.SourceRow, Date: f.Date,
		Detail: fmt.Sprintf("registration %s is also written as %s elsewhere; now written %s",
			f.AircraftReg, strings.Join(others, ", "), f.AircraftType),
	}}
}

// aircraft builds the seed list. Active follows recent use rather than a
// hand-kept flag, counting back from the last flight in the books so the answer
// does not change with the wall clock.
func (g *registry) aircraft(flights []Flight) []Aircraft {
	if len(flights) == 0 {
		return nil
	}

	lastFlown := map[string]string{}
	newest := ""
	for _, f := range flights {
		if f.Date > lastFlown[f.AircraftReg] {
			lastFlown[f.AircraftReg] = f.Date
		}
		if f.Date > newest {
			newest = f.Date
		}
	}
	cutoff := minusYears(newest, activeWithinYears)

	out := make([]Aircraft, 0, len(g.order))
	for _, reg := range g.order {
		out = append(out, Aircraft{
			Registration: reg,
			Type:         dominantType(g.types[reg]),
			DefaultClass: classOf(reg),
			IFRCapable:   ifrCapableRegistrations[reg],
			Active:       lastFlown[reg] >= cutoff,
			Notes:        aircraftNotes[reg],
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Registration < out[j].Registration })
	return out
}

// minusYears shifts a YYYY-MM-DD date back by whole years.
//
// It works on the string rather than through time.Parse because a whole-year
// shift never changes the month or the day, and because these dates were all
// produced by time.Format and are well-formed. Anything that is not a
// four-digit year is returned unchanged, which keeps every aircraft active --
// the safe direction for a list that only decides what a form offers.
func minusYears(date string, n int) string {
	if len(date) < 4 {
		return date
	}
	y, err := strconv.Atoi(date[:4])
	if err != nil {
		return date
	}
	return fmt.Sprintf("%04d%s", y-n, date[4:])
}

// dominantType picks the type flown most often under a registration, breaking
// ties alphabetically so the seed list is reproducible. The flights themselves
// always keep the type as written on paper.
func dominantType(counts map[string]int) string {
	best, bestN := "", -1
	for t, n := range counts {
		if n > bestN || (n == bestN && t < best) {
			best, bestN = t, n
		}
	}
	return best
}

// aircraftNotes carry the facts about an airframe that no CSV column records.
var aircraftNotes = map[string]string{
	"OH-GKT": "ex-SE-GKT, re-registered; same airframe. Owned by the pilot.",
	"SE-GKT": "re-registered OH-GKT; same airframe.",
	"OH-MIL": "Maule on floats; the book writes the type as M6(sea). Always on floats.",
	"OH-CDK": "Cessna 185 on floats.",
	"OK-PDP": "Almost certainly a transcription slip for OH-PDP. Not corrected; see APP.md.",
	"OH-CMU": "Distinct aircraft from OH-CMV -- the registrations differ only in the last letter.",
	"OH-CMV": "Distinct aircraft from OH-CMU -- the registrations differ only in the last letter.",
}

// sortDiscrepancies puts the report in book order so it reads like the paper.
func sortDiscrepancies(ds []Discrepancy) {
	sort.SliceStable(ds, func(i, j int) bool {
		if ds[i].Book != ds[j].Book {
			return ds[i].Book < ds[j].Book
		}
		if ds[i].Row != ds[j].Row {
			return ds[i].Row < ds[j].Row
		}
		return ds[i].Kind < ds[j].Kind
	})
}
