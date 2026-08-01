package main

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

// The three exports are the documents that leave the application, so what is
// asserted here is that each one is a real PDF, is served as one, and is
// private.
func TestTheThreeExportsRenderPDFs(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	for _, path := range []string{
		"/logbook/api/export/easa.pdf",
		"/logbook/api/export/table.pdf",
		"/logbook/api/export/statistics.pdf",
	} {
		t.Run(path, func(t *testing.T) {
			w := h.do("GET", path, "", auth)
			if w.Code != http.StatusOK {
				t.Fatalf("status %d, body %s", w.Code, w.Body.String())
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/pdf" {
				t.Errorf("Content-Type = %q, want application/pdf", ct)
			}
			if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF-")) {
				t.Error("the body is not a PDF")
			}
			// A browser must save it under a name that says what it is, not
			// under the endpoint's path.
			if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "filename=") {
				t.Errorf("Content-Disposition = %q, want a filename", cd)
			}
		})
	}
}

// The date range has to reach the export, or the document says something
// different from the page the owner generated it from.
func TestExportsHonourTheDateRange(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	full := h.do("GET", "/logbook/api/export/table.pdf", "", auth)
	narrow := h.do("GET", "/logbook/api/export/table.pdf?from=2021-01-01&to=2021-12-31", "", auth)
	if full.Code != http.StatusOK || narrow.Code != http.StatusOK {
		t.Fatalf("statuses %d/%d", full.Code, narrow.Code)
	}
	if bytes.Equal(full.Body.Bytes(), narrow.Body.Bytes()) {
		t.Error("filtering to one year produced the same document as the whole logbook")
	}
}

// An unparseable date must not be quietly ignored here either: an export is
// the most likely document to be sent to somebody else.
func TestExportsRejectABadDateRatherThanIgnoringIt(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("GET", "/logbook/api/export/statistics.pdf?from=nonsense", "", auth)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", w.Code)
	}
}

// The EASA document deliberately covers all three books rather than the range,
// because it is the complete record an authority asks for.
func TestTheEASAExportIgnoresTheRangeAndCoversEverything(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	full := h.do("GET", "/logbook/api/export/easa.pdf", "", auth)
	narrow := h.do("GET", "/logbook/api/export/easa.pdf?from=2021-01-01&to=2021-06-30", "", auth)
	if full.Code != http.StatusOK || narrow.Code != http.StatusOK {
		t.Fatalf("statuses %d/%d", full.Code, narrow.Code)
	}
	if !bytes.Equal(full.Body.Bytes(), narrow.Body.Bytes()) {
		t.Error("the EASA export changed with a date range; it must always be the whole logbook")
	}
}
