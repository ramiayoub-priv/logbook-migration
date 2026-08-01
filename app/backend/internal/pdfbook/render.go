package pdfbook

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

// Options are the bits of a document that are not the flight data.
type Options struct {
	// HolderName is printed on the cover line of each document. Empty is fine;
	// the documents are still valid without it.
	HolderName string
	// Generated stamps the document and its PDF metadata. Passed in rather
	// than read from the clock so that rendering the same logbook twice
	// produces byte-identical output, which is what makes a diff meaningful
	// and a test able to assert on the bytes.
	Generated time.Time
}

// fontFamily is a PDF core font, so nothing is embedded and no font file has
// to ship next to the binary -- the deploy artefact stays one static file.
const fontFamily = "Helvetica"

// newDoc starts a document with the settings every one of the three shares.
func newDoc(orientation string, opts Options) (*fpdf.Fpdf, func(string) string) {
	pdf := fpdf.New(orientation, "mm", "A4", "")

	// Core PDF fonts are Latin-1. Finnish place names are full of ä and ö
	// ("Tuusulanjärvi"), and without this they would be dropped or mangled --
	// silently changing what a legal record says a place was called. cp1252 is
	// embedded in the fpdf module, so this needs no external file.
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Two things are needed for byte-identical output from identical input,
	// and both are easy to lose. The creation date would otherwise be the
	// wall clock; and fpdf writes its font and page objects by ranging over
	// Go maps, whose order is deliberately randomised, so without
	// SetCatalogSort two renders of the same logbook differ in the order the
	// font objects appear. Determinism is what makes a diff between two
	// exports mean "the record changed" rather than "it was regenerated".
	pdf.SetCatalogSort(true)
	pdf.SetCreationDate(opts.Generated)
	pdf.SetModificationDate(opts.Generated)
	pdf.SetProducer("logbook", false)
	pdf.SetAuthor(tr(opts.HolderName), false)

	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(marginX, marginY, marginX)
	pdf.SetLineWidth(0.15)
	pdf.SetDrawColor(120, 120, 120)
	return pdf, tr
}

const (
	marginX = 6.0
	marginY = 7.0
)

// out renders the document to bytes, turning fpdf's deferred error into a
// returned one.
//
// fpdf collects the first error and turns every later call into a no-op, so a
// failure surfaces here rather than at the call that caused it. That makes an
// unchecked Output a way to write a silently truncated PDF of a legal record.
func out(pdf *fpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdfbook: rendering: %w", err)
	}
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("pdfbook: rendering: %w", err)
	}
	return buf.Bytes(), nil
}

// column is one column of a table: its heading and its width in millimetres.
type column struct {
	title string
	width float64
	align string
}

// headerFooter draws the title line and the page footer shared by all three
// documents.
func title(pdf *fpdf.Fpdf, tr func(string) string, heading, subtitle string, width float64) {
	pdf.SetFont(fontFamily, "B", 11)
	pdf.CellFormat(width, 6, tr(heading), "", 1, "L", false, 0, "")
	if subtitle != "" {
		pdf.SetFont(fontFamily, "", 7.5)
		pdf.CellFormat(width, 4, tr(subtitle), "", 1, "L", false, 0, "")
	}
	pdf.Ln(1)
}

func footer(pdf *fpdf.Fpdf, tr func(string) string, left, right string, y, width float64) {
	pdf.SetXY(marginX, y)
	pdf.SetFont(fontFamily, "I", 6.5)
	pdf.CellFormat(width/2, 4, tr(left), "", 0, "L", false, 0, "")
	pdf.CellFormat(width/2, 4, tr(right), "", 0, "R", false, 0, "")
}

// drawHeader draws a grey table heading row.
//
// The headings are stacked over several lines because CellFormat does not
// honour a newline -- it draws the whole string on one baseline, so a title
// like "SINGLE\nENGINE VFR" runs straight through the neighbouring columns
// and the header becomes unreadable. Each line is therefore placed by hand
// inside a box drawn once.
func drawHeader(pdf *fpdf.Fpdf, tr func(string) string, cols []column, height float64) {
	const size = 4.6
	pdf.SetFont(fontFamily, "B", size)
	pdf.SetFillColor(228, 228, 228)

	x0, y := pdf.GetX(), pdf.GetY()
	lineHeight := size * 0.42 // points to millimetres, near enough for a heading

	x := x0
	for _, c := range cols {
		pdf.Rect(x, y, c.width, height, "FD")

		lines := strings.Split(c.title, "\n")
		// Vertically centre the block of lines in the box.
		top := y + (height-float64(len(lines))*lineHeight)/2
		for i, line := range lines {
			pdf.SetXY(x, top+float64(i)*lineHeight)
			pdf.CellFormat(c.width, lineHeight, clip(pdf, tr(line), c.width-0.6), "", 0, "C", false, 0, "")
		}
		x += c.width
	}
	pdf.SetXY(x0, y+height)
}
