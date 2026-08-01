// Package pdfmodel computes what goes in every cell of the three exported
// PDFs. It is pure: no fpdf, no fonts, no io.
//
// It is a separate package from internal/pdfbook, which draws the documents,
// for one reason. CLAUDE.md rule 0.6 names "PDF totals" as part of the
// calculation core that must be at 100% coverage -- a wrong figure in the
// authority's copy of a legal record is the worst failure this application
// has. A figure produced by a pure function can be asserted exactly and held
// to that bar; the same figure computed inside a drawing call can only be
// eyeballed in a viewer, and drags a pile of untestable fpdf error branches
// into the same coverage number. Splitting them lets the Makefile enforce
// what the rule actually asks for.
//
// Nothing here is stored (rule 0.5). Every running total on every page is
// derived from the flights at render time, walking Seq, which is the book's
// own order.
package pdfmodel

import (
	"fmt"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/hhmm"
	"github.com/ramiayoub/logbook/backend/internal/stats"
)

// EASARowsPerPage is the paper book's own page size, counted off the
// photograph of page 33 (logbook-3/IMG_6025.JPEG).
const EASARowsPerPage = 15

// EASARow is one line of the EASA form, already turned into the strings that
// go in its cells. Empty means the cell is left blank, which is what the paper
// does for a zero -- a page of "0:00" is unreadable and, worse, reads as a
// measurement rather than an absence.
type EASARow struct {
	// GENERAL
	Date     string
	DepPlace string
	OffBlock string
	ArrPlace string
	OnBlock  string
	Type     string
	Reg      string
	PICName  string

	// FLIGHT TIME
	Total         string
	Night         string
	SEVFR         string
	SEIFR         string
	MEVFR         string
	MEIFR         string
	PIC           string
	Copilot       string
	MultiPilot    string
	Instructor    string
	Dual          string
	InstructorSTD string

	// OTHER
	LandingsDay   string
	LandingsNight string
	Remarks       string
}

// EASARowOf renders one flight.
//
// Two mappings in here are judgement calls rather than transcription, and both
// are made in the conservative direction:
//
//   - All single-engine time goes in SINGLE ENGINE VFR and SINGLE ENGINE IFR
//     is always blank. The CSVs carry no flight-rules column, and instrument
//     time is not a substitute for one: OH-COF and OH-CTH are C152s with
//     instrument time logged under the hood. Splitting a figure out of it
//     would be manufacturing a fact about a legal record (rule 0.2). This is
//     also exactly what the owner's own pages do.
//   - The TOTAL column carries the flight's own time, not a running total.
//     The owner writes a running total there by hand, but the form's TOTAL
//     THIS PAGE / PREVIOUS / TOTAL block below is where a cumulative belongs,
//     and an authority reads the column as per-flight. The two conventions are
//     reconciled on the page: the block still shows the running figure.
func EASARowOf(f csvbook.Flight) EASARow {
	return EASARow{
		Date:     shortDate(f.Date),
		DepPlace: f.DepPlace,
		OffBlock: clock(f.OffBlockUTC),
		ArrPlace: f.ArrPlace,
		OnBlock:  clock(f.OnBlockUTC),
		Type:     f.AircraftType,
		Reg:      f.AircraftReg,
		PICName:  f.PICName,

		Total: BlankZero(f.TotalMinutes),
		Night: BlankZero(f.NightMinutes),
		SEVFR: singleEngine(f),
		SEIFR: "",
		MEVFR: multiEngine(f),
		MEIFR: "",

		PIC:           BlankZero(f.PICMinutes),
		Copilot:       BlankZero(f.CopilotMinutes),
		MultiPilot:    BlankZero(f.MultiPilotMinutes),
		Instructor:    BlankZero(f.InstructorMinutes),
		Dual:          BlankZero(f.DualMinutes),
		InstructorSTD: "",

		LandingsDay:   BlankZeroCount(f.LandingsDay),
		LandingsNight: BlankZeroCount(f.LandingsNight),
		Remarks:       f.Remarks,
	}
}

// singleEngine returns the flight time if this is a single-engine class.
func singleEngine(f csvbook.Flight) string {
	if isMulti(f.Class) {
		return ""
	}
	return BlankZero(f.TotalMinutes)
}

// multiEngine is the mirror of singleEngine. No multi-engine aircraft appears
// anywhere in the three books, so this is empty on all 1293 rows today -- but
// the column exists on the form and the classification must not be a
// hard-coded "always single".
func multiEngine(f csvbook.Flight) string {
	if !isMulti(f.Class) {
		return ""
	}
	return BlankZero(f.TotalMinutes)
}

func isMulti(c csvbook.Class) bool {
	return c == csvbook.ClassMEPLand || c == csvbook.ClassMEPSea
}

// EASAPage is one page of the reproduction: its rows and the three-row totals
// block the paper carries at the bottom.
//
// Both forms of each total are kept. The strings are what gets drawn; the
// stats.Summary values are what the tests assert the arithmetic on, and what
// lets the next page start from this one's running figure.
type EASAPage struct {
	Number int
	Of     int
	Rows   []EASARow

	ThisPage EASARow // TOTAL THIS PAGE
	Previous EASARow // TOTAL PREVIOUS PAGES
	Total    EASARow // TOTAL

	ThisPageSummary stats.Summary
	PreviousSummary stats.Summary
	TotalSummary    stats.Summary
}

// EASAPages splits the logbook into pages and computes each page's totals.
//
// The pagination itself comes from stats.Paginate, which is the one place that
// knows a cumulative walks Seq and never the date. This function turns its
// output into cells.
func EASAPages(flights []csvbook.Flight, rowsPerPage int) ([]EASAPage, error) {
	paged, err := stats.Paginate(flights, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("pdfbook: %w", err)
	}

	out := make([]EASAPage, 0, len(paged))
	for _, p := range paged {
		rows := make([]EASARow, 0, len(p.Flights))
		for _, f := range p.Flights {
			rows = append(rows, EASARowOf(f))
		}
		out = append(out, EASAPage{
			Number: p.Number,
			Of:     len(paged),
			Rows:   rows,

			ThisPage: totalsRow(p.ThisPage),
			Previous: totalsRow(p.Previous),
			Total:    totalsRow(p.Total),

			ThisPageSummary: p.ThisPage,
			PreviousSummary: p.Previous,
			TotalSummary:    p.Total,
		})
	}
	return out, nil
}

// totalsRow renders a summary as the cells of one row in the totals block.
//
// Only the columns the block actually totals are filled: the general block has
// nothing to add up, and neither does a remark.
func totalsRow(s stats.Summary) EASARow {
	return EASARow{
		Total: BlankZero(s.Total),
		Night: BlankZero(s.Night),
		SEVFR: BlankZero(s.SeaTotal + s.LandTotal - multiTotal(s)),
		MEVFR: BlankZero(multiTotal(s)),

		PIC:        BlankZero(s.PIC),
		Instructor: BlankZero(s.Instructor),
		Dual:       BlankZero(s.Dual),

		LandingsDay:   BlankZeroCount(s.LandingsDay),
		LandingsNight: BlankZeroCount(s.LandingsNight),
	}
}

// multiTotal is the multi-engine time in a summary.
//
// stats.Summary splits on sea/land rather than on engine count, because that
// is the split the owner's ratings need. No multi-engine aircraft appears in
// the books, so this is zero -- stated as a named function rather than a
// literal 0 so that the day a twin is flown, this is the one place to fix.
func multiTotal(stats.Summary) int { return 0 }

// shortDate renders YYYY-MM-DD as the form's dd/mm/yy.
//
// An unparseable date is passed through rather than replaced or dropped: on a
// legal record, showing what is actually stored beats showing a blank that
// hides it.
func shortDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("02/01/06")
}

// clock renders an instant as HH:MM UTC, blank if there is none. The paper
// column is headed "(UTC)".
func clock(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("15:04")
}

// blankZero renders minutes as H:MM, or blank for zero.
func BlankZero(minutes int) string {
	if minutes == 0 {
		return ""
	}
	return hhmm.Format(minutes)
}

func BlankZeroCount(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

// landingSplit renders the day/night landings, marking the ones the
// application inferred rather than read off the page.
//
// The asterisk is the honest half of Task 8: 30 rows carry night time whose
// day/night split was seeded rather than transcribed, and a table that printed
// them like verified figures would be asserting something nobody checked
// (rule 0.2).
func LandingSplit(f csvbook.Flight) string {
	s := fmt.Sprintf("%d/%d", f.LandingsDay, f.LandingsNight)
	if !f.LandingsVerified {
		s += "*"
	}
	return s
}

// sourceRef is where this row came from: which paper book and which line, or
// "app" for a flight typed into the application.
func SourceRef(f csvbook.Flight) string {
	if f.SourceBook == 0 {
		return "app"
	}
	return fmt.Sprintf("b%d:%d", f.SourceBook, f.SourceRow)
}
func DescribeRange(r stats.Range) string {
	switch {
	case r.From == "" && r.To == "":
		return "Whole logbook"
	case r.From == "":
		return "Up to " + r.To
	case r.To == "":
		return "From " + r.From
	default:
		return r.From + " to " + r.To
	}
}

// StatisticsLine is one figure on the statistics sheet.
type StatisticsLine struct {
	Label string
	Value string
	// Group starts a new block on the page.
	Group string
}

// StatisticsLines is the statistics sheet as data: every figure app/APP.md
// section 2 asks for, in the order it is printed.
//
// It is a pure function so the numbers on the page can be asserted exactly
// rather than read out of a rendered document -- these are the figures that
// evidence a rating, so they are calculation core (rule 0.6).
func StatisticsLines(s stats.Summary) []StatisticsLine {
	dur := func(m int) string { return hhmm.Format(m) }
	count := func(n int) string { return fmt.Sprintf("%d", n) }

	return []StatisticsLine{
		{Group: "TOTALS", Label: "Flights", Value: count(s.Flights)},
		{Group: "TOTALS", Label: "Total time", Value: dur(s.Total)},
		{Group: "TOTALS", Label: "Pilot in command", Value: dur(s.PIC)},
		{Group: "TOTALS", Label: "Dual", Value: dur(s.Dual)},
		{Group: "TOTALS", Label: "Night", Value: dur(s.Night)},
		{Group: "TOTALS", Label: "Instrument", Value: dur(s.Instrument)},
		{Group: "TOTALS", Label: "Instructor", Value: dur(s.Instructor)},

		{Group: "SEAPLANE", Label: "Seaplane total", Value: dur(s.SeaTotal)},
		{Group: "SEAPLANE", Label: "Seaplane PIC", Value: dur(s.SeaPIC)},
		{Group: "SEAPLANE", Label: "Seaplane instructor", Value: dur(s.SeaInstructor)},

		{Group: "LANDPLANE", Label: "Landplane total", Value: dur(s.LandTotal)},
		{Group: "LANDPLANE", Label: "Landplane PIC", Value: dur(s.LandPIC)},
		{Group: "LANDPLANE", Label: "Landplane instructor", Value: dur(s.LandInstructor)},

		{Group: "LANDINGS", Label: "Landings day", Value: count(s.LandingsDay)},
		{Group: "LANDINGS", Label: "Landings night", Value: count(s.LandingsNight)},
		{Group: "LANDINGS", Label: "Landings seaplane", Value: count(s.LandingsSea)},
		{Group: "LANDINGS", Label: "Landings landplane", Value: count(s.LandingsLand)},
	}
}
