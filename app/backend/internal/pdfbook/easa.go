package pdfbook

import (
	"fmt"

	"github.com/go-pdf/fpdf"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/pdfmodel"
)

// The EASA grid, in millimetres, laid out to match the paper book photographed
// at logbook-3/IMG_6025.JPEG. The widths sum to the printable width of A4
// landscape at a 6 mm margin (297 - 12 = 285).
//
// The paper form is a two-page spread: GENERAL and the first four FLIGHT TIME
// columns on the left page, the remaining FLIGHT TIME columns and OTHER on the
// right. This renders the whole spread on one landscape sheet, so that one PDF
// page is one logbook page of 15 flights, which is what makes the page numbers
// mean the same thing as the book's.
var easaColumns = []column{
	// GENERAL
	{"DATE\n(dd/mm/yy)", 14, "C"},
	{"DEP\n(place)", 18, "L"},
	{"OFF BLK\n(UTC)", 10, "C"},
	{"ARR\n(place)", 18, "L"},
	{"ON BLK\n(UTC)", 10, "C"},
	{"TYPE OF\nAIRCRAFT", 14, "C"},
	{"REGIS-\nTRATION", 15, "C"},
	{"NAME OF PILOT\nIN COMMAND", 18, "L"},

	// FLIGHT TIME
	{"TOTAL", 11, "C"},
	{"NIGHT", 10, "C"},
	{"SINGLE\nENGINE VFR", 11, "C"},
	{"SINGLE\nENGINE IFR", 10, "C"},
	{"MULTI\nENGINE VFR", 10, "C"},
	{"MULTI\nENGINE IFR", 10, "C"},
	{"PILOT IN\nCOMMAND", 11, "C"},
	{"CO-PILOT", 10, "C"},
	{"MULTI\nPILOT", 10, "C"},
	{"FLIGHT\nINSTR.", 11, "C"},
	{"DUAL", 10, "C"},
	{"INSTR.\nSTD", 10, "C"},

	// OTHER
	{"LDG\nDAY", 8, "C"},
	{"LDG\nNIGHT", 8, "C"},
	{"REMARKS AND ENDORSEMENTS", 28, "L"},
}

const (
	easaRowHeight    = 7.4
	easaHeaderHeight = 8.5
)

// cells returns one row's values in the same order as easaColumns. Keeping the
// order in exactly one place is what stops a column and its data drifting
// apart -- the failure mode where a PDF looks perfect and puts dual time in
// the PIC column.
func cells(r pdfmodel.EASARow) []string {
	return []string{
		r.Date, r.DepPlace, r.OffBlock, r.ArrPlace, r.OnBlock, r.Type, r.Reg, r.PICName,
		r.Total, r.Night, r.SEVFR, r.SEIFR, r.MEVFR, r.MEIFR,
		r.PIC, r.Copilot, r.MultiPilot, r.Instructor, r.Dual, r.InstructorSTD,
		r.LandingsDay, r.LandingsNight, r.Remarks,
	}
}

// EASA renders the whole logbook as an EASA-format reproduction.
//
// It covers all three paper books, not just the EASA-format Book 3: Books 1
// and 2 are an older layout, and what an authority wants is one complete
// record in the current standard format (decision log, 2026-08-01).
func EASA(flights []csvbook.Flight, opts Options) ([]byte, error) {
	pages, err := pdfmodel.EASAPages(flights, pdfmodel.EASARowsPerPage)
	if err != nil {
		return nil, err
	}

	pdf, tr := newDoc("L", opts)
	pdf.SetTitle(tr("Pilot logbook -- EASA format"), false)

	width := 0.0
	for _, c := range easaColumns {
		width += c.width
	}

	for _, page := range pages {
		pdf.AddPage()
		drawEASAPage(pdf, tr, page, opts, width)
	}

	// A logbook with no flights still has to produce a valid document rather
	// than an empty file that looks like a failed download.
	if len(pages) == 0 {
		pdf.AddPage()
		title(pdf, tr, "PILOT LOGBOOK", "No flights recorded.", width)
	}
	return out(pdf)
}

func drawEASAPage(pdf *fpdf.Fpdf, tr func(string) string, page pdfmodel.EASAPage, opts Options, width float64) {
	heading := "PILOT LOGBOOK"
	if opts.HolderName != "" {
		heading = "PILOT LOGBOOK -- " + opts.HolderName
	}
	title(pdf, tr, heading, "EASA format. All times UTC; durations H:MM.", width)

	drawGroupHeader(pdf, tr)
	drawHeader(pdf, tr, easaColumns, easaHeaderHeight)

	pdf.SetFont(fontFamily, "", 5.6)
	for _, row := range page.Rows {
		drawEASARow(pdf, tr, cells(row), false)
	}
	// Short pages keep the grid: the paper form has 15 ruled lines whether or
	// not they are all used, and a half-drawn table reads as a truncated
	// document.
	for i := len(page.Rows); i < pdfmodel.EASARowsPerPage; i++ {
		drawEASARow(pdf, tr, make([]string, len(easaColumns)), false)
	}

	// The three-row totals block. This is the arithmetic an authority checks
	// by eye, and it is computed from the flights on every render -- never
	// stored (rule 0.5).
	pdf.SetFont(fontFamily, "B", 5.6)
	for _, block := range []struct {
		label string
		row   pdfmodel.EASARow
	}{
		{"TOTAL THIS PAGE", page.ThisPage},
		{"TOTAL PREVIOUS PAGES", page.Previous},
		{"TOTAL", page.Total},
	} {
		drawEASATotals(pdf, tr, block.label, cells(block.row))
	}

	y := pdf.GetY() + 3
	pdf.SetXY(marginX, y)
	pdf.SetFont(fontFamily, "", 6.5)
	pdf.CellFormat(60, 4, tr("Certified true and correct: ______________________"), "", 0, "L", false, 0, "")

	footer(pdf, tr,
		fmt.Sprintf("Generated %s", opts.Generated.UTC().Format("2006-01-02 15:04 MST")),
		fmt.Sprintf("Page %d of %d", page.Number, page.Of),
		200, width)
}

// drawGroupHeader draws the GENERAL / FLIGHT TIME / OTHER banner the paper
// form carries above the column headings.
func drawGroupHeader(pdf *fpdf.Fpdf, tr func(string) string) {
	groups := []struct {
		label string
		from  int
		to    int
	}{
		{"GENERAL", 0, 8},
		{"FLIGHT TIME", 8, 20},
		{"OTHER", 20, len(easaColumns)},
	}
	pdf.SetFont(fontFamily, "B", 6)
	pdf.SetFillColor(210, 210, 210)
	for _, g := range groups {
		var w float64
		for _, c := range easaColumns[g.from:g.to] {
			w += c.width
		}
		pdf.CellFormat(w, 4.5, tr(g.label), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
}

func drawEASARow(pdf *fpdf.Fpdf, tr func(string) string, cells []string, fill bool) {
	if fill {
		pdf.SetFillColor(238, 238, 238)
	}
	for i, c := range easaColumns {
		v := ""
		if i < len(cells) {
			v = clip(pdf, tr(cells[i]), c.width-1)
		}
		pdf.CellFormat(c.width, easaRowHeight, v, "1", 0, c.align, fill, 0, "")
	}
	pdf.Ln(-1)
}

// drawEASATotals draws one row of the totals block.
//
// The label spans the whole GENERAL group, exactly as the paper form does.
// Squeezing "TOTAL PREVIOUS PAGES" into the last general column instead
// truncates it to "TOTAL PREVIOU." -- which on a document going to an
// authority reads as a broken export rather than as a total.
func drawEASATotals(pdf *fpdf.Fpdf, tr func(string) string, label string, cells []string) {
	const generalColumns = 8

	pdf.SetFillColor(238, 238, 238)
	var span float64
	for _, c := range easaColumns[:generalColumns] {
		span += c.width
	}
	pdf.CellFormat(span, easaRowHeight, tr(label)+"  ", "1", 0, "R", true, 0, "")

	for i, c := range easaColumns[generalColumns:] {
		v := ""
		if j := i + generalColumns; j < len(cells) {
			v = clip(pdf, tr(cells[j]), c.width-1)
		}
		pdf.CellFormat(c.width, easaRowHeight, v, "1", 0, c.align, true, 0, "")
	}
	pdf.Ln(-1)
}

// clip shortens a value that would overflow its column.
//
// Truncating is the least-bad option for the two free-text columns: a place
// name or a remark that overflows would otherwise be drawn across the next
// cell's contents, which on a table of figures makes a number unreadable and
// could make it look like a different number. Durations and dates never reach
// this -- they are fixed width by construction.
func clip(pdf *fpdf.Fpdf, s string, max float64) string {
	if s == "" || pdf.GetStringWidth(s) <= max {
		return s
	}
	for len(s) > 1 {
		s = s[:len(s)-1]
		if pdf.GetStringWidth(s+".") <= max {
			return s + "."
		}
	}
	return s
}
