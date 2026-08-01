package csvbook

import (
	"strings"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/timeutil"
)

// header25 is the Book 1 header: no Cumulative_Instructor column.
const header25 = `"Date","Aircraft_Type","Aircraft_Reg","Departure","Arrival","Off_Block","On_Block","Takeoff","Landing","Block_Time","Total_Time","Instrument_Time","Night_Time","PIC_Time","Student_Time","Instructor_Time","pic_name","Landings","Remarks","Cumulative_Total","Cumulative_PIC","Cumulative_Student","Cumulative_Instrument","Cumulative_SEP_Sea","Cumulative_Landings"`

// header26 is the Book 2/3 header, which adds Cumulative_Instructor.
const header26 = header25 + `,"Cumulative_Instructor"`

func csvOf(header string, rows ...string) string {
	return header + "\n" + strings.Join(rows, "\n") + "\n"
}

func loadOne(t *testing.T, body string, src Source) *Logbook {
	t.Helper()
	lb, err := parseAll([]reader{{src: src, body: strings.NewReader(body)}})
	if err != nil {
		t.Fatalf("parseAll: %v", err)
	}
	return lb
}

func TestParseMapsOneRowOntoTheDomain(t *testing.T) {
	// A dual training row from Book 1: local times, student time, no PIC.
	body := csvOf(header25,
		`"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","Martevuo","1","Siirto","0:57","","0:57","","","1"`)

	lb := loadOne(t, body, Source{Book: 1})
	if len(lb.Flights) != 1 {
		t.Fatalf("got %d flights, want 1", len(lb.Flights))
	}
	f := lb.Flights[0]

	if f.Seq != 1 {
		t.Errorf("Seq = %d, want 1", f.Seq)
	}
	if f.Date != "2011-05-25" {
		t.Errorf("Date = %q, want 2011-05-25", f.Date)
	}
	if f.AircraftType != "C152" || f.AircraftReg != "OH-KLS" {
		t.Errorf("aircraft = %q/%q, want C152/OH-KLS", f.AircraftType, f.AircraftReg)
	}
	if f.Class != ClassSEPLand {
		t.Errorf("Class = %q, want %q", f.Class, ClassSEPLand)
	}
	if f.DepPlace != "EFHF" || f.ArrPlace != "EFHF" {
		t.Errorf("places = %q/%q, want EFHF/EFHF", f.DepPlace, f.ArrPlace)
	}
	// 13:07 Helsinki on 2011-05-25 is EEST (UTC+3) -> 10:07Z.
	if got := f.OffBlockUTC.Format("2006-01-02T15:04:05Z"); got != "2011-05-25T10:07:00Z" {
		t.Errorf("OffBlockUTC = %s, want 2011-05-25T10:07:00Z", got)
	}
	if got := f.OnBlockUTC.Format("2006-01-02T15:04:05Z"); got != "2011-05-25T11:04:00Z" {
		t.Errorf("OnBlockUTC = %s, want 2011-05-25T11:04:00Z", got)
	}
	if f.OffBlockRaw != "13:07" || f.OnBlockRaw != "14:04" {
		t.Errorf("raw = %q/%q, want the strings exactly as written", f.OffBlockRaw, f.OnBlockRaw)
	}
	if f.TimeOrigin != timeutil.OriginConvertedFromLocal {
		t.Errorf("TimeOrigin = %q, want %q", f.TimeOrigin, timeutil.OriginConvertedFromLocal)
	}
	if !f.TakeoffUTC.IsZero() || !f.LandingUTC.IsZero() {
		t.Errorf("airborne pair should be zero when the cells are blank")
	}
	if f.BlockMinutes != 57 || f.TotalMinutes != 57 {
		t.Errorf("block/total = %d/%d, want 57/57", f.BlockMinutes, f.TotalMinutes)
	}
	if f.DualMinutes != 57 {
		t.Errorf("DualMinutes = %d, want 57 (Student_Time is dual)", f.DualMinutes)
	}
	if f.PICMinutes != 0 || f.InstructorMinutes != 0 || f.NightMinutes != 0 || f.InstrumentMinutes != 0 {
		t.Errorf("blank duration cells must be zero, got %+v", f)
	}
	if f.CopilotMinutes != 0 || f.MultiPilotMinutes != 0 {
		t.Errorf("co-pilot and multi-pilot are not in the CSV and must be zero")
	}
	if f.PICName != "Martevuo" {
		t.Errorf("PICName = %q, want Martevuo", f.PICName)
	}
	if f.LandingsDay != 1 || f.LandingsNight != 0 {
		t.Errorf("landings = %d/%d, want 1/0 (seeded as day)", f.LandingsDay, f.LandingsNight)
	}
	if !f.LandingsVerified {
		t.Errorf("a day flight's landings are unambiguously day landings")
	}
	if f.Remarks != "Siirto" {
		t.Errorf("Remarks = %q", f.Remarks)
	}
	if f.SourceBook != 1 || f.SourceRow != 2 {
		t.Errorf("provenance = book %d row %d, want book 1 row 2 (1-based CSV line)", f.SourceBook, f.SourceRow)
	}
}

func TestParseZuluRowKeepsTheRawAndDoesNotConvert(t *testing.T) {
	body := csvOf(header26,
		`"30/07/2026","C172","OH-GKT","Kahvisaari","Kahvisaari","15:40Z","16:40Z","15:45Z","16:31Z","1:00","1:00","","","1:00","","1:00","self","7","","1:00","1:00","","","1:00","7","1:00"`)

	f := loadOne(t, body, Source{Book: 3}).Flights[0]

	if got := f.OffBlockUTC.Format("15:04"); got != "15:40" {
		t.Errorf("OffBlockUTC = %s, want 15:40 unconverted", got)
	}
	if f.OffBlockRaw != "15:40Z" {
		t.Errorf("OffBlockRaw = %q, want the Z kept exactly as written", f.OffBlockRaw)
	}
	if f.TimeOrigin != timeutil.OriginUTCAsWritten {
		t.Errorf("TimeOrigin = %q, want %q", f.TimeOrigin, timeutil.OriginUTCAsWritten)
	}
	if got := f.TakeoffUTC.Format("15:04"); got != "15:45" {
		t.Errorf("TakeoffUTC = %s, want 15:45", got)
	}
	if got := f.LandingUTC.Format("15:04"); got != "16:31" {
		t.Errorf("LandingUTC = %s, want 16:31", got)
	}
	if f.Class != ClassSEPSea {
		t.Errorf("OH-GKT is a seaplane registration; Class = %q", f.Class)
	}
	if f.InstructorMinutes != 60 || f.PICMinutes != 60 {
		t.Errorf("instructor/PIC = %d/%d, want 60/60", f.InstructorMinutes, f.PICMinutes)
	}
}

func TestParseRollsOnBlockForwardAcrossMidnight(t *testing.T) {
	body := csvOf(header26,
		`"31/12/2019","DA40","OH-STL","EETN","EFHF","23:30Z","00:20Z","","","0:50","0:50","","","0:50","","","self","1","","0:50","0:50","","","","1",""`)

	f := loadOne(t, body, Source{Book: 2}).Flights[0]

	if got := f.OnBlockUTC.Format("2006-01-02T15:04Z"); got != "2020-01-01T00:20Z" {
		t.Errorf("OnBlockUTC = %s, want the date rolled to 2020-01-01", got)
	}
	if d := f.OnBlockUTC.Sub(f.OffBlockUTC).Minutes(); d != 50 {
		t.Errorf("elapsed = %v minutes, want 50", d)
	}
}

func TestSeedRowIsSkippedButStillSeedsTheCumulativeSeries(t *testing.T) {
	// Row 2 is Book 1's final row carried over. It must not become a flight,
	// and it must not be reported as a cumulative break either -- the running
	// series continues from it.
	body := csvOf(header26,
		`"23/04/2017","P28A","OH-PDP","EFLA","EFHF","15:00","15:47","","","0:47","0:47","","","0:47","","","self","1","","395:49","312:42","83:07","3:12","58:39","889",""`,
		`"24/04/2017","P28A","OH-PDP","EFHF","EFHF","10:00","11:00","","","1:00","1:00","","","1:00","","","self","2","","396:49","313:42","83:07","3:12","58:39","891",""`)

	lb := loadOne(t, body, Source{Book: 2, SkipSeedRow: true})

	if len(lb.Flights) != 1 {
		t.Fatalf("got %d flights, want 1 (the seed row is not a flight)", len(lb.Flights))
	}
	if lb.Flights[0].Date != "2017-04-24" {
		t.Errorf("kept the wrong row: %q", lb.Flights[0].Date)
	}
	if lb.Flights[0].SourceRow != 3 {
		t.Errorf("SourceRow = %d, want 3 -- provenance is the real CSV line", lb.Flights[0].SourceRow)
	}
	if n := len(lb.Discrepancies); n != 0 {
		t.Fatalf("got %d discrepancies, want 0: %+v", n, lb.Discrepancies)
	}
	if lb.Totals.Flights != 1 || lb.Totals.Total != 60 {
		t.Errorf("totals count only imported rows, got %+v", lb.Totals)
	}
}

func TestTotalsSumOnlyImportedRows(t *testing.T) {
	body := csvOf(header26,
		`"01/06/2021","C172","OH-CTL","Tuusulanjärvi","Tuusulanjärvi","18:13","19:34","","","1:21","1:21","","","1:21","","1:21","self","7","","1:21","1:21","","","1:21","7","1:21"`,
		`"02/06/2021","P28A","OH-PDP","EFHV","EFHV","09:00","10:30","","","1:30","1:30","0:20","0:10","","1:30","","Autere","3","","2:51","1:21","1:30","0:20","1:21","10","1:21"`)

	lb := loadOne(t, body, Source{Book: 3})
	want := Totals{
		Flights: 2, Total: 171, PIC: 81, Dual: 90, Instrument: 20,
		Night: 10, Instructor: 81, SEPSea: 81, Landings: 10,
	}
	if lb.Totals != want {
		t.Errorf("Totals = %+v, want %+v", lb.Totals, want)
	}
}

func TestNightRowLeavesTheLandingSplitUnverified(t *testing.T) {
	body := csvOf(header26,
		`"02/06/2021","P28A","OH-PDP","EFHV","EFHV","09:00","10:30","","","1:30","1:30","","0:10","1:30","","","self","3","","1:30","1:30","","","","3",""`)

	lb := loadOne(t, body, Source{Book: 3})
	f := lb.Flights[0]
	if f.LandingsDay != 3 || f.LandingsNight != 0 {
		t.Errorf("landings = %d/%d, want the sum seeded as day", f.LandingsDay, f.LandingsNight)
	}
	if f.LandingsVerified {
		t.Errorf("a row carrying night time cannot have its landing split inferred")
	}
	assertDiscrepancy(t, lb, KindLandingsUnverified, 2)
}

func TestCumulativeBreakIsReportedNotCorrected(t *testing.T) {
	// The real defect in logbook_1_final.csv line 28, reduced: the row claims
	// 1:21 of instrument time but the cumulative column advances only 1:12.
	body := csvOf(header25,
		`"26/09/2011","C152","OH-COF","EFHF","EFTP","08:38","10:02","","","1:24","1:24","","","","1:24","","Martevuo","1","","1:24","","1:24","1:00","","1"`,
		`"28/09/2011","C152","OH-COF","EFHF","EFHF","08:22","09:34","","","1:12","1:12","1:21","","","1:12","","Martevuo","1","","2:36","","2:36","2:12","","2"`)

	lb := loadOne(t, body, Source{Book: 1})

	if len(lb.Flights) != 2 {
		t.Fatalf("a discrepancy must never drop a row: got %d flights", len(lb.Flights))
	}
	if lb.Flights[1].InstrumentMinutes != 81 {
		t.Errorf("InstrumentMinutes = %d, want 81 -- the row is kept as written",
			lb.Flights[1].InstrumentMinutes)
	}
	d := assertDiscrepancy(t, lb, KindCumulativeBreak, 3)
	if !strings.Contains(d.Detail, "Cumulative_Instrument") {
		t.Errorf("Detail = %q, want it to name the column", d.Detail)
	}
	// The same row also breaks the component<=total invariant.
	assertDiscrepancy(t, lb, KindComponentOverTotal, 3)
}

func TestCumulativeSeriesResumesAfterABreak(t *testing.T) {
	// After a break the running series must re-anchor on the CSV, otherwise one
	// bad row would report every later row as broken too.
	body := csvOf(header25,
		`"26/09/2011","C152","OH-COF","EFHF","EFTP","08:38","10:02","","","1:00","1:00","","","","1:00","","M","1","","1:00","","1:00","","","1"`,
		`"27/09/2011","C152","OH-COF","EFHF","EFTP","08:38","10:02","","","1:00","1:00","","","","1:00","","M","1","","9:00","","2:00","","","2"`,
		`"28/09/2011","C152","OH-COF","EFHF","EFTP","08:38","10:02","","","1:00","1:00","","","","1:00","","M","1","","10:00","","3:00","","","3"`)

	lb := loadOne(t, body, Source{Book: 1})
	breaks := discrepanciesOf(lb, KindCumulativeBreak)
	if len(breaks) != 1 {
		t.Fatalf("got %d cumulative breaks, want 1: %+v", len(breaks), breaks)
	}
	if breaks[0].Row != 3 {
		t.Errorf("break reported on row %d, want 3", breaks[0].Row)
	}
}

func TestBlockAndTotalDisagreementIsReported(t *testing.T) {
	body := csvOf(header26,
		`"08/09/2025","C172","OH-GKT","Kahvisaari","Kahvisaari","15:40Z","16:25Z","","","0:45","0:38","","","0:38","","","self","1","","0:38","0:38","","","0:38","1",""`)

	lb := loadOne(t, body, Source{Book: 3})
	f := lb.Flights[0]
	if f.BlockMinutes != 45 || f.TotalMinutes != 38 {
		t.Errorf("both values are kept as written, got %d/%d", f.BlockMinutes, f.TotalMinutes)
	}
	assertDiscrepancy(t, lb, KindBlockTotalMismatch, 2)
}

func TestNonFinnishAndConflictingRegistrationsAreFlagged(t *testing.T) {
	body := csvOf(header26,
		`"17/05/2018","P28A","OK-PDP","EFHF","EFHF","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","1:00","1:00","","","","1",""`,
		`"05/03/2024","C152","OH-CMU","EFHV","EFHV","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","2:00","2:00","","","","2",""`,
		`"06/03/2024","C172","OH-CMU","EFHV","EFHV","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","3:00","3:00","","","","3",""`)

	lb := loadOne(t, body, Source{Book: 3})

	reg := assertDiscrepancy(t, lb, KindRegistrationFormat, 2)
	if !strings.Contains(reg.Detail, "OK-PDP") {
		t.Errorf("Detail = %q, want it to name OK-PDP", reg.Detail)
	}
	// The conflict is reported once, against the row that introduces the
	// second type -- not once per row.
	conflicts := discrepanciesOf(lb, KindTypeConflict)
	if len(conflicts) != 1 {
		t.Fatalf("got %d type conflicts, want 1: %+v", len(conflicts), conflicts)
	}
	if conflicts[0].Row != 4 {
		t.Errorf("conflict reported on row %d, want 4", conflicts[0].Row)
	}
}

func TestUnknownAircraftTypeIsFlagged(t *testing.T) {
	// C192 is not a Cessna type; four real rows carry it.
	body := csvOf(header26,
		`"10/06/2015","C192","OH-CTL","Tuusulanjärvi","Tuusulanjärvi","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","1:00","1:00","","","1:00","1",""`)

	lb := loadOne(t, body, Source{Book: 3})
	d := assertDiscrepancy(t, lb, KindUnknownType, 2)
	if !strings.Contains(d.Detail, "C192") {
		t.Errorf("Detail = %q, want it to name C192", d.Detail)
	}
	if lb.Flights[0].AircraftType != "C192" {
		t.Errorf("the type is kept as written on paper, got %q", lb.Flights[0].AircraftType)
	}
	// A conflicting type must not change the seaplane classification, which
	// comes from the registration.
	if lb.Flights[0].Class != ClassSEPSea {
		t.Errorf("Class = %q, want %q", lb.Flights[0].Class, ClassSEPSea)
	}
}

func TestDottedDateIsAcceptedAndFlagged(t *testing.T) {
	// Eight consecutive rows in Book 2 were transcribed as DD.MM.YYYY. The day
	// field is unambiguous there (21 > 12) and the rows sit in chronological
	// order between 15/03 and 07/05, so the reading is certain -- but the
	// inconsistency is still reported so the CSV can be normalised.
	body := csvOf(header26,
		`"21.03.2018","C172","OH-CWB","EFHF","EFHF","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","1:00","1:00","","","","1",""`)

	lb := loadOne(t, body, Source{Book: 2})
	if lb.Flights[0].Date != "2018-03-21" {
		t.Errorf("Date = %q, want 2018-03-21", lb.Flights[0].Date)
	}
	d := assertDiscrepancy(t, lb, KindDateFormat, 2)
	if !strings.Contains(d.Detail, "21.03.2018") {
		t.Errorf("Detail = %q, want it to quote the raw cell", d.Detail)
	}
}

func TestDottedDateWithAnAmbiguousDayIsFlaggedLouder(t *testing.T) {
	// "04.05.2018" is 4 May day-first and 5 April month-first. It is read
	// day-first like the rest of its batch, but the report has to say so
	// plainly -- reading it wrongly would move a flight by a month.
	body := csvOf(header26,
		`"04.05.2018","C172","OH-CWB","EFHF","EFHF","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","1:00","1:00","","","","1",""`)

	lb := loadOne(t, body, Source{Book: 2})
	if lb.Flights[0].Date != "2018-05-04" {
		t.Errorf("Date = %q, want 2018-05-04", lb.Flights[0].Date)
	}
	d := assertDiscrepancy(t, lb, KindDateFormat, 2)
	if !strings.Contains(d.Detail, "CONFIRM AGAINST THE PAPER") {
		t.Errorf("Detail = %q, want it to demand confirmation", d.Detail)
	}
}

func TestUnreadableDottedDateIsStillRejected(t *testing.T) {
	body := csvOf(header26,
		`"32.13.2018","C172","OH-CWB","EFHF","EFHF","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","1:00","1:00","","","","1",""`)

	_, err := parseAll([]reader{{src: Source{Book: 2}, body: strings.NewReader(body)}})
	if err == nil || !strings.Contains(err.Error(), "Date") {
		t.Fatalf("err = %v, want a Date error", err)
	}
}

func TestAmbiguousLocalTimeSurfacesForReview(t *testing.T) {
	// 03:30 on 2019-03-31 never existed in Helsinki: the clocks jumped
	// 03:00 -> 04:00. The row is kept, flagged, and marked unknown.
	body := csvOf(header26,
		`"31/03/2019","P28A","OH-PDP","EFHF","EFHF","03:30","04:30","","","1:00","1:00","","","1:00","","","self","1","","1:00","1:00","","","","1",""`)

	lb := loadOne(t, body, Source{Book: 2})
	if lb.Flights[0].TimeOrigin != timeutil.OriginUnknown {
		t.Errorf("TimeOrigin = %q, want %q", lb.Flights[0].TimeOrigin, timeutil.OriginUnknown)
	}
	assertDiscrepancy(t, lb, KindUnknownTimeOrigin, 2)
}

func TestAirborneAmbiguityAlsoSurfaces(t *testing.T) {
	// Block times are Zulu and unambiguous; the airborne pair is a local time
	// inside the spring gap. The row must still surface.
	body := csvOf(header26,
		`"31/03/2019","P28A","OH-PDP","EFHF","EFHF","01:00Z","02:00Z","03:30","04:00","1:00","1:00","","","1:00","","","self","1","","1:00","1:00","","","","1",""`)

	lb := loadOne(t, body, Source{Book: 2})
	if lb.Flights[0].TimeOrigin != timeutil.OriginUnknown {
		t.Errorf("TimeOrigin = %q, want %q -- an unresolved airborne pair taints the row",
			lb.Flights[0].TimeOrigin, timeutil.OriginUnknown)
	}
	assertDiscrepancy(t, lb, KindUnknownTimeOrigin, 2)
}

func TestBlankBlockCellsYieldOriginNone(t *testing.T) {
	body := csvOf(header26,
		`"02/06/2021","P28A","OH-PDP","EFHV","EFHV","","","","","1:30","1:30","","","1:30","","","self","3","","1:30","1:30","","","","3",""`)

	f := loadOne(t, body, Source{Book: 3}).Flights[0]
	if f.TimeOrigin != timeutil.OriginNone {
		t.Errorf("TimeOrigin = %q, want %q", f.TimeOrigin, timeutil.OriginNone)
	}
	if !f.OffBlockUTC.IsZero() || !f.OnBlockUTC.IsZero() {
		t.Errorf("blank cells must not invent an instant")
	}
}

func TestSeqIsContinuousAcrossBooks(t *testing.T) {
	book1 := csvOf(header25,
		`"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","M","1","","0:57","","0:57","","","1"`)
	book2 := csvOf(header26,
		`"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","M","1","","0:57","","0:57","","","1",""`,
		`"26/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","1:00","1:00","","","","1:00","","M","1","","1:57","","1:57","","","2",""`)

	lb, err := parseAll([]reader{
		{src: Source{Book: 1}, body: strings.NewReader(book1)},
		{src: Source{Book: 2, SkipSeedRow: true}, body: strings.NewReader(book2)},
	})
	if err != nil {
		t.Fatalf("parseAll: %v", err)
	}
	if len(lb.Flights) != 2 {
		t.Fatalf("got %d flights, want 2", len(lb.Flights))
	}
	if lb.Flights[0].Seq != 1 || lb.Flights[1].Seq != 2 {
		t.Errorf("seq = %d,%d -- must be one continuous series over all books",
			lb.Flights[0].Seq, lb.Flights[1].Seq)
	}
	if lb.Flights[1].SourceBook != 2 || lb.Flights[1].SourceRow != 3 {
		t.Errorf("provenance = %d/%d, want book 2 row 3",
			lb.Flights[1].SourceBook, lb.Flights[1].SourceRow)
	}
	if n := len(discrepanciesOf(lb, KindCumulativeBreak)); n != 0 {
		t.Errorf("the series must carry across the book boundary, got %d breaks", n)
	}
}

func TestAircraftListIsDerivedFromTheFlights(t *testing.T) {
	body := csvOf(header26,
		`"10/06/2015","C172","OH-CTL","Tuusulanjärvi","Tuusulanjärvi","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","1:00","1:00","","","1:00","1",""`,
		`"11/06/2015","C192","OH-CTL","Tuusulanjärvi","Tuusulanjärvi","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","2:00","2:00","","","2:00","2",""`,
		`"04/04/2019","P28A","OH-PIF","EFNU","EFNU","09:00","10:00","","","1:00","1:00","0:30","","1:00","","","self","1","","3:00","3:00","","0:30","2:00","3",""`)

	lb := loadOne(t, body, Source{Book: 3})
	byReg := map[string]Aircraft{}
	for _, a := range lb.Aircraft {
		byReg[a.Registration] = a
	}
	if len(byReg) != 2 {
		t.Fatalf("got %d aircraft, want 2: %+v", len(byReg), lb.Aircraft)
	}
	ctl := byReg["OH-CTL"]
	if ctl.Type != "C172" {
		t.Errorf("OH-CTL type = %q, want the most-flown type C172, not the C192 outlier", ctl.Type)
	}
	if ctl.DefaultClass != ClassSEPSea {
		t.Errorf("OH-CTL class = %q, want %q", ctl.DefaultClass, ClassSEPSea)
	}
	if ctl.IFRCapable {
		t.Errorf("OH-CTL is not on the IFR list")
	}
	if !byReg["OH-PIF"].IFRCapable {
		t.Errorf("OH-PIF is the IR trainer and must be marked IFR capable")
	}
	if byReg["OH-PIF"].DefaultClass != ClassSEPLand {
		t.Errorf("OH-PIF class = %q, want %q", byReg["OH-PIF"].DefaultClass, ClassSEPLand)
	}
	if lb.Aircraft[0].Registration != "OH-CTL" || lb.Aircraft[1].Registration != "OH-PIF" {
		t.Errorf("aircraft must come out in a stable sorted order, got %+v", lb.Aircraft)
	}
}

func TestAircraftActiveFlagFollowsRecentUse(t *testing.T) {
	body := csvOf(header26,
		`"25/05/2011","C152","OH-KLS","EFHF","EFHF","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","1:00","1:00","","","","1",""`,
		`"30/07/2026","C172","OH-GKT","Kahvisaari","Kahvisaari","09:00","10:00","","","1:00","1:00","","","1:00","","","self","1","","2:00","2:00","","","1:00","2",""`)

	lb := loadOne(t, body, Source{Book: 3})
	for _, a := range lb.Aircraft {
		want := a.Registration == "OH-GKT"
		if a.Active != want {
			t.Errorf("%s Active = %v, want %v", a.Registration, a.Active, want)
		}
	}
}

func TestMalformedInputIsRejectedRatherThanGuessed(t *testing.T) {
	cases := []struct {
		name, body, want string
	}{
		{
			name: "missing required column",
			body: `"Date","Aircraft_Type"` + "\n" + `"25/05/2011","C152"` + "\n",
			want: "missing required column",
		},
		{
			name: "unreadable date",
			body: csvOf(header25, `"2011-05-25","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","M","1","","0:57","","0:57","","","1"`),
			want: "not a DD/MM/YYYY date",
		},
		{
			name: "unreadable duration",
			body: csvOf(header25, `"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","57","","","","0:57","","M","1","","0:57","","0:57","","","1"`),
			want: "Total_Time",
		},
		{
			name: "unreadable landings",
			body: csvOf(header25, `"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","M","one","","0:57","","0:57","","","1"`),
			want: "Landings",
		},
		{
			name: "unreadable clock",
			body: csvOf(header25, `"25/05/2011","C152","OH-KLS","EFHF","EFHF","1307","14:04","","","0:57","0:57","","","","0:57","","M","1","","0:57","","0:57","","","1"`),
			want: "Off_Block",
		},
		{
			name: "unreadable airborne clock",
			body: csvOf(header25, `"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","1345","14:00","0:57","0:57","","","","0:57","","M","1","","0:57","","0:57","","","1"`),
			want: "Takeoff",
		},
		{
			name: "unreadable cumulative",
			body: csvOf(header25, `"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","M","1","","57","","0:57","","","1"`),
			want: "Cumulative_Total",
		},
		{
			name: "unreadable cumulative landings",
			body: csvOf(header25, `"25/05/2011","C152","OH-KLS","EFHF","EFHF","13:07","14:04","","","0:57","0:57","","","","0:57","","M","1","","0:57","","0:57","","","x"`),
			want: "Cumulative_Landings",
		},
		{
			name: "ragged row",
			body: header25 + "\n" + `"25/05/2011","C152"` + "\n",
			want: "record",
		},
		{
			name: "no data rows",
			body: header25 + "\n",
			want: "no data rows",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseAll([]reader{{src: Source{Book: 1}, body: strings.NewReader(tc.body)}})
			if err == nil {
				t.Fatalf("got no error, want one mentioning %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestEmptyHeaderIsRejected(t *testing.T) {
	_, err := parseAll([]reader{{src: Source{Book: 1}, body: strings.NewReader("")}})
	if err == nil {
		t.Fatal("an empty file must be an error, not an empty logbook")
	}
}

func TestSeedRowSkipOnAFileThatOnlyHasASeedRow(t *testing.T) {
	body := csvOf(header26,
		`"23/04/2017","P28A","OH-PDP","EFLA","EFHF","15:00","15:47","","","0:47","0:47","","","0:47","","","self","1","","395:49","312:42","83:07","3:12","58:39","889",""`)

	lb := loadOne(t, body, Source{Book: 2, SkipSeedRow: true})
	if len(lb.Flights) != 0 {
		t.Fatalf("got %d flights, want 0", len(lb.Flights))
	}
	if lb.Totals.Flights != 0 {
		t.Errorf("Totals = %+v, want an empty tally", lb.Totals)
	}
	if len(lb.Aircraft) != 0 {
		t.Errorf("no flights means no aircraft, got %+v", lb.Aircraft)
	}
}

// --- helpers ---

func discrepanciesOf(lb *Logbook, kind Kind) []Discrepancy {
	var out []Discrepancy
	for _, d := range lb.Discrepancies {
		if d.Kind == kind {
			out = append(out, d)
		}
	}
	return out
}

func assertDiscrepancy(t *testing.T, lb *Logbook, kind Kind, row int) Discrepancy {
	t.Helper()
	for _, d := range lb.Discrepancies {
		if d.Kind == kind && d.Row == row {
			return d
		}
	}
	t.Fatalf("no %s discrepancy on row %d; got %+v", kind, row, lb.Discrepancies)
	return Discrepancy{}
}

// ValidClass guards the new-flight form against writing a class the schema's
// CHECK constraint would reject. The vocabulary lives in this package because a
// second copy elsewhere is one that eventually disagrees with the database.
func TestValidClass(t *testing.T) {
	for _, c := range []Class{ClassSEPLand, ClassSEPSea, ClassMEPLand, ClassMEPSea, ClassTMG} {
		if !ValidClass(c) {
			t.Errorf("ValidClass(%q) = false; it is in the schema's CHECK list", c)
		}
	}
	for _, c := range []Class{"", "SEP_AMPHIB", "sep_land", "GLIDER"} {
		if ValidClass(c) {
			t.Errorf("ValidClass(%q) = true; the insert would fail the CHECK constraint", c)
		}
	}
}
