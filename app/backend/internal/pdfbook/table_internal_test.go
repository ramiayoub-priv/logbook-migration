package pdfbook

import "testing"

// fpdf does not complain when a table row runs off the paper. It draws the
// cells anyway and the printer cuts them, so a column added without checking
// the arithmetic silently truncates the record instead of failing loudly.
// This is the check that adding TAKEOFF and LANDING did not do that.
func TestTableColumnsFitBetweenTheMargins(t *testing.T) {
	const a4LandscapeWidth = 297.0 // mm; newDoc builds every document on A4
	printable := a4LandscapeWidth - 2*marginX

	total := 0.0
	for _, c := range tableColumns {
		total += c.width
	}
	if total > printable {
		t.Errorf("the flight table is %.1f mm wide but only %.1f mm fits between the margins",
			total, printable)
	}
	t.Logf("flight table: %.1f mm of %.1f mm printable", total, printable)
}
