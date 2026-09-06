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
//
// The airborne group -- TAKEOFF, LANDING, AIR -- sits after the block group
// rather than interleaved with it, which is the same layout the on-screen
// table uses and for the same reason: off/on block and takeoff/landing are
// four times in the same format, and the aircraft's logbook is filled from
// only one of the pairs. Interleaving them chronologically would read better
// and be easier to misread, which is the wrong trade in a legal record. The
// two block columns are headed OFF BLOCK and ON BLOCK rather than OFF and ON
// now that they have neighbours they could be confused with.
//
// The EASA form has cells for the block pair only, so this is the only
// document that carries the airborne times at all.
var tableColumns = []column{
	{"SEQ", 11, "R"},
	{"DATE", 15, "C"},
	{"TYPE", 11, "C"},
	{"REG", 14, "C"},
	{"CLASS", 14, "C"},
	{"FROM", 20, "L"},
	{"TO", 20, "L"},
	{"OFF\nBLOCK", 11, "C"},
	{"ON\nBLOCK", 11, "C"},
	{"TOTAL", 12, "R"},
	{"TAKEOFF", 12, "C"},
	{"LANDING", 12, "C"},
	{"AIR", 11, "R"},
	{"PIC", 11, "R"},
	{"DUAL", 11, "R"},
	{"NIGHT", 11, "R"},
	{"INSTR", 11, "R"},
	{"INSTRUCTOR", 14, "R"},
	{"LDG D/N", 13, "C"},
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
			fmt.Sprintf("%s -- %d flight(s). All times UTC; durations H:MM. "+
				"AIR is wheels-up to wheels-down, derived from TAKEOFF and LANDING; "+
				"it is blank on the rows that never recorded them.",
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
				pdfmodel.Clock(f.TakeoffUTC),
				pdfmodel.Clock(f.LandingUTC),
				pdfmodel.AirTime(f),
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
