package pdfbook

import (
	"fmt"

	"github.com/ramiayoub/logbook/backend/internal/pdfmodel"
	"github.com/ramiayoub/logbook/backend/internal/stats"
)

// Statistics renders the statistics sheet for a range.
func Statistics(s stats.Summary, rng stats.Range, opts Options) ([]byte, error) {
	pdf, tr := newDoc("P", opts)
	pdf.SetTitle(tr("Pilot logbook -- statistics"), false)

	const width = 210 - 2*marginX

	pdf.AddPage()
	heading := "FLIGHT STATISTICS"
	if opts.HolderName != "" {
		heading = "FLIGHT STATISTICS -- " + opts.HolderName
	}
	title(pdf, tr, heading, pdfmodel.DescribeRange(rng)+". Durations H:MM.", width)

	lines := pdfmodel.StatisticsLines(s)
	group := ""
	for _, l := range lines {
		if l.Group != group {
			group = l.Group
			pdf.Ln(3)
			pdf.SetFont(fontFamily, "B", 9)
			pdf.SetFillColor(228, 228, 228)
			pdf.CellFormat(width, 6, tr(group), "1", 1, "L", true, 0, "")
		}
		pdf.SetFont(fontFamily, "", 9)
		pdf.CellFormat(width*0.65, 6, "  "+tr(l.Label), "1", 0, "L", false, 0, "")
		pdf.SetFont(fontFamily, "B", 9)
		pdf.CellFormat(width*0.35, 6, tr(l.Value), "1", 1, "R", false, 0, "")
	}

	// The caveat is part of the document, not a footnote to be dropped. Thirty
	// rows in the books carry night time whose day/night landing split was
	// inferred rather than read off the page, and a statistics sheet that
	// stated the night landing figure without saying so would be claiming a
	// verification that has not happened (rule 0.2, Task 8).
	if s.LandingsUnverified > 0 {
		pdf.Ln(4)
		pdf.SetFont(fontFamily, "I", 8)
		pdf.MultiCell(width, 4, tr(fmt.Sprintf(
			"Note: %d flight(s) in this range carry a day/night landing split that was inferred "+
				"from the presence of night time rather than read from the paper page. The landing "+
				"total is unaffected; the split between day and night is provisional.",
			s.LandingsUnverified)), "", "L", false)
	}

	footer(pdf, tr,
		fmt.Sprintf("Generated %s", opts.Generated.UTC().Format("2006-01-02 15:04 MST")),
		"", 285, width)
	return out(pdf)
}
