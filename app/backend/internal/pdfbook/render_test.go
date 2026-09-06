package pdfbook_test

import (
	"bytes"
	"compress/zlib"
	"io"
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/pdfbook"
	"github.com/ramiayoub/logbook/backend/internal/stats"
)

var testOpts = pdfbook.Options{
	HolderName: "Rami Ayoub",
	Generated:  time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
}

// seaFlight is a seaplane instructing row, with a Finnish place name so that
// the Latin-1 translation is exercised by every document these tests render.
func seaFlight(t *testing.T) csvbook.Flight {
	t.Helper()
	off, err := time.Parse(time.RFC3339, "2021-06-01T15:13:00Z")
	if err != nil {
		t.Fatal(err)
	}
	return csvbook.Flight{
		Seq: 1, Date: "2021-06-01",
		AircraftType: "C172", AircraftReg: "OH-CTL", Class: csvbook.ClassSEPSea,
		DepPlace: "Tuusulanjärvi", ArrPlace: "Kelvenne",
		OffBlockUTC: off, OnBlockUTC: off.Add(81 * time.Minute),
		BlockMinutes: 81, TotalMinutes: 81, PICMinutes: 81, InstructorMinutes: 81,
		PICName: "Ayoub", LandingsDay: 7, LandingsVerified: true,
		SourceBook: 3, SourceRow: 3,
	}
}

func manyFlights(t *testing.T, n int) []csvbook.Flight {
	t.Helper()
	var out []csvbook.Flight
	for i := 0; i < n; i++ {
		f := seaFlight(t)
		f.Seq = i + 1
		out = append(out, f)
	}
	return out
}

func isPDF(t *testing.T, b []byte) {
	t.Helper()
	if len(b) == 0 {
		t.Fatal("rendered no bytes at all")
	}
	if !bytes.HasPrefix(b, []byte("%PDF-")) {
		t.Fatalf("output does not start with a PDF header: %q", b[:min(12, len(b))])
	}
	if !bytes.Contains(b, []byte("%%EOF")) {
		t.Error("output has no end-of-file marker; the document is truncated")
	}
}

func TestEASARendersAValidDocument(t *testing.T) {
	b, err := pdfbook.EASA(manyFlights(t, 20), testOpts)
	if err != nil {
		t.Fatalf("EASA: %v", err)
	}
	isPDF(t, b)
}

// Two renders of the same logbook must be byte-identical. Without it there is
// no way to tell "the export changed" from "the export was regenerated", which
// matters when the document is a copy of a legal record.
func TestRenderingIsDeterministic(t *testing.T) {
	flights := manyFlights(t, 20)
	for _, tc := range []struct {
		name   string
		render func() ([]byte, error)
	}{
		{"EASA", func() ([]byte, error) { return pdfbook.EASA(flights, testOpts) }},
		{"Table", func() ([]byte, error) { return pdfbook.Table(flights, stats.Range{}, testOpts) }},
		{"Statistics", func() ([]byte, error) {
			return pdfbook.Statistics(stats.Summarize(flights), stats.Range{}, testOpts)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first, err := tc.render()
			if err != nil {
				t.Fatal(err)
			}
			second, err := tc.render()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(first, second) {
				t.Error("two renders of the same data produced different bytes")
			}
		})
	}
}

// One PDF page per 15 flights, so that a page number in the export means the
// same thing as a page of the book.
func TestEASAProducesOnePagePerFifteenFlights(t *testing.T) {
	for _, tc := range []struct{ flights, pages int }{
		{1, 1}, {15, 1}, {16, 2}, {30, 2}, {31, 3},
	} {
		b, err := pdfbook.EASA(manyFlights(t, tc.flights), testOpts)
		if err != nil {
			t.Fatal(err)
		}
		if got := countPages(t, b); got != tc.pages {
			t.Errorf("%d flights rendered %d pages, want %d", tc.flights, got, tc.pages)
		}
	}
}

// An empty logbook must still be a readable document rather than a zero-byte
// file that looks like a failed download.
func TestEASAOfAnEmptyLogbookIsStillADocument(t *testing.T) {
	b, err := pdfbook.EASA(nil, testOpts)
	if err != nil {
		t.Fatalf("EASA of an empty logbook: %v", err)
	}
	isPDF(t, b)
	if countPages(t, b) != 1 {
		t.Errorf("empty logbook rendered %d pages, want 1", countPages(t, b))
	}
}

func TestTableAndStatisticsRender(t *testing.T) {
	flights := manyFlights(t, 40)

	b, err := pdfbook.Table(flights, stats.Range{From: "2021-01-01", To: "2021-12-31"}, testOpts)
	if err != nil {
		t.Fatalf("Table: %v", err)
	}
	isPDF(t, b)
	if got := countPages(t, b); got != 2 {
		t.Errorf("40 flights rendered %d table pages, want 2 at 34 rows a page", got)
	}

	b, err = pdfbook.Statistics(stats.Summarize(flights), stats.Range{}, testOpts)
	if err != nil {
		t.Fatalf("Statistics: %v", err)
	}
	isPDF(t, b)
}

func TestTableOfNoFlightsIsStillADocument(t *testing.T) {
	b, err := pdfbook.Table(nil, stats.Range{From: "2030-01-01"}, testOpts)
	if err != nil {
		t.Fatal(err)
	}
	isPDF(t, b)
	if countPages(t, b) != 1 {
		t.Errorf("rendered %d pages for an empty range, want 1", countPages(t, b))
	}
}

// The statistics sheet must say when a night-landing figure is inferred rather
// than read off the page. Printing it silently would claim a verification that
// has not happened.
func TestStatisticsDisclosesUnverifiedLandings(t *testing.T) {
	s := stats.Summary{Flights: 1, Total: 60, LandingsNight: 2, LandingsUnverified: 1}
	b, err := pdfbook.Statistics(s, stats.Range{}, testOpts)
	if err != nil {
		t.Fatal(err)
	}
	if !containsText(t, b, "inferred") {
		t.Error("the sheet reports an inferred landing split without saying so")
	}
}

func TestStatisticsOmitsTheCaveatWhenEverythingIsVerified(t *testing.T) {
	s := stats.Summary{Flights: 1, Total: 60, LandingsDay: 2}
	b, err := pdfbook.Statistics(s, stats.Range{}, testOpts)
	if err != nil {
		t.Fatal(err)
	}
	if containsText(t, b, "inferred") {
		t.Error("the caveat is printed even though nothing in the range is unverified")
	}
}

// Finnish place names must survive into the document. The PDF core fonts are
// Latin-1 and an untranslated "Tuusulanjärvi" is silently mangled -- changing
// what the record says a place was called.
func TestFinnishPlaceNamesSurvive(t *testing.T) {
	b, err := pdfbook.Table(manyFlights(t, 1), stats.Range{}, testOpts)
	if err != nil {
		t.Fatal(err)
	}
	// cp1252 encodes ä as the single byte 0xE4.
	if !containsText(t, b, "Tuusulanj\xe4rvi") {
		t.Error("the place name did not reach the document in Latin-1")
	}
}

// --- helpers ----------------------------------------------------------------

// countPages counts /Type /Page objects, which is the page count as a reader
// sees it rather than as the writer believes it.
func countPages(t *testing.T, b []byte) int {
	t.Helper()
	return bytes.Count(b, []byte("/Type /Page\n")) + bytes.Count(b, []byte("/Type /Page/")) +
		bytes.Count(b, []byte("/Type /Page "))
}

// containsText looks for a string in the document's content streams.
//
// The streams are Flate-compressed in the real output, so they are inflated
// here rather than turning compression off in the renderer: a test that needs
// production code to behave differently is testing something else.
func containsText(t *testing.T, b []byte, want string) bool {
	t.Helper()
	if bytes.Contains(b, []byte(want)) {
		return true
	}
	for _, s := range inflateStreams(b) {
		if bytes.Contains(s, []byte(want)) {
			return true
		}
	}
	return false
}

// inflateStreams returns every zlib stream in the document that inflates.
// Anything that does not is skipped: not every stream object is compressed
// text, and this is a search helper rather than a PDF parser.
func inflateStreams(b []byte) [][]byte {
	var out [][]byte
	const open, close = "stream\n", "\nendstream"
	for rest := b; ; {
		i := bytes.Index(rest, []byte(open))
		if i < 0 {
			return out
		}
		rest = rest[i+len(open):]
		j := bytes.Index(rest, []byte(close))
		if j < 0 {
			return out
		}
		raw := rest[:j]
		rest = rest[j+len(close):]

		zr, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			continue
		}
		dec, err := io.ReadAll(zr)
		zr.Close()
		if err != nil {
			continue
		}
		out = append(out, dec)
	}
}

// The flight table is the document that carries "everything the application
// knows" about a row, and takeoff and landing are two of the things it knows:
// 36 rows in the live logbook record them, including every one of the 17
// entered through the app. Printing only OFF and ON drops them silently --
// the reader cannot tell a flight that never had airborne times recorded from
// one whose times the export threw away.
func TestTableCarriesTakeoffAndLandingTimes(t *testing.T) {
	f := seaFlight(t)
	f.TakeoffUTC = f.OffBlockUTC.Add(7 * time.Minute) // 15:20, seven after off-block
	f.LandingUTC = f.OnBlockUTC.Add(-6 * time.Minute) // 16:28, six before on-block

	b, err := pdfbook.Table([]csvbook.Flight{f}, stats.Range{}, testOpts)
	if err != nil {
		t.Fatalf("Table: %v", err)
	}
	// "1:08" is the derived airborne time, 15:20 to 16:28. It is distinct from
	// this flight's 1:21 of block time, so finding it proves the AIR column
	// carries the derived figure rather than repeating TOTAL.
	for _, want := range []string{"TAKEOFF", "LANDING", "AIR", "15:20", "16:28", "1:08"} {
		if !containsText(t, b, want) {
			t.Errorf("the flight table does not carry %q", want)
		}
	}
}

// A row with no airborne times recorded -- 1277 of the 1313 -- must leave the
// two cells empty rather than print a zero instant. "00:00" in a legal record
// reads as a measured midnight, not as an absence.
func TestTableLeavesAirborneTimesBlankWhenTheyWereNeverRecorded(t *testing.T) {
	f := seaFlight(t) // no TakeoffUTC, no LandingUTC
	b, err := pdfbook.Table([]csvbook.Flight{f}, stats.Range{}, testOpts)
	if err != nil {
		t.Fatalf("Table: %v", err)
	}
	if containsText(t, b, "00:00") {
		t.Error("an unrecorded takeoff or landing was printed as 00:00")
	}
}
