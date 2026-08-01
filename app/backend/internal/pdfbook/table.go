package pdfbook

import (
	"fmt"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/pdfmodel"
	"github.com/ramiayoub/logbook/backend/internal/stats"
)

// tableColumns is the detailed export: everything the application knows about
// a flight, including the provenance and the audit fields the EASA form has
// nowhere to put.
//
// This is the document for the owner rather than for an authority, which is
// why it carries source book and row, the time origin, and the flag saying
// whether a landing split was read off the page or inferred.
var tableColumns = []column{
	{"SEQ", 13, "R"},
	{"DATE", 17, "C"},
	{"TYPE", 13, "C"},
	{"REG", 16, "C"},
	{"CLASS", 17, "C"},
	{"FROM", 20, "L"},
	{"TO", 20, "L"},
	{"OFF", 11, "C"},
	{"ON", 11, "C"},
	{"TOTAL", 12, "R"},
	{"PIC", 12, "R"},
	{"DUAL", 12, "R"},
	{"NIGHT", 12, "R"},
	{"INSTR", 12, "R"},
	{"INSTRUCTOR", 15, "R"},
	{"LDG D/N", 14, "C"},
	{"TIME ORIGIN", 22, "C"},
	{"SRC", 12, "C"},
}

const (
	tableRowHeight  = 5.0
	tableRowsPerPDF = 34
)

// Table renders the detailed flight table over a date range.
func Table(flights []csvbook.Flight, rng stats.Range, opts Options) ([]byte, error) {
	pdf, tr := newDoc("L", opts)
	pdf.SetTitle(tr("Pilot logbook -- flight table"), false)

	width := 0.0
	for _, c := range tableColumns {
		width += c.width
	}

	total := (len(flights) + tableRowsPerPDF - 1) / tableRowsPerPDF
	if total == 0 {
		total = 1
	}

	for page := 0; page < total; page++ {
		pdf.AddPage()
		title(pdf, tr,
			"FLIGHT TABLE",
			fmt.Sprintf("%s -- %d flight(s). All times UTC; durations H:MM.",
				pdfmodel.DescribeRange(rng), len(flights)),
			width)
		drawHeader(pdf, tr, tableColumns, 6)

		pdf.SetFont(fontFamily, "", 5.8)
		start := page * tableRowsPerPDF
		end := min(start+tableRowsPerPDF, len(flights))
		for _, f := range flights[start:end] {
			r := pdfmodel.EASARowOf(f)
			cells := []string{
				fmt.Sprintf("%d", f.Seq),
				f.Date,
				f.AircraftType,
				f.AircraftReg,
				string(f.Class),
				f.DepPlace,
				f.ArrPlace,
				r.OffBlock,
				r.OnBlock,
				r.Total,
				r.PIC,
				r.Dual,
				r.Night,
				pdfmodel.BlankZero(f.InstrumentMinutes),
				r.Instructor,
				pdfmodel.LandingSplit(f),
				string(f.TimeOrigin),
				pdfmodel.SourceRef(f),
			}
			for i, c := range tableColumns {
				pdf.CellFormat(c.width, tableRowHeight, clip(pdf, tr(cells[i]), c.width-1),
					"1", 0, c.align, false, 0, "")
			}
			pdf.Ln(-1)
		}

		footer(pdf, tr,
			fmt.Sprintf("Generated %s", opts.Generated.UTC().Format("2006-01-02 15:04 MST")),
			fmt.Sprintf("Page %d of %d", page+1, total),
			200, width)
	}
	return out(pdf)
}
