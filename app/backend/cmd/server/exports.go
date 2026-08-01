package main

import (
	"fmt"
	"net/http"

	"github.com/ramiayoub/logbook/backend/internal/pdfbook"
	"github.com/ramiayoub/logbook/backend/internal/stats"
)

// The three exported documents. All private, like everything else.
//
// They are separate endpoints rather than one with a ?kind= parameter so that
// each has its own URL a browser can download, its own filename, and its own
// line in the default-deny test.

func (s *Server) handleExportEASA(w http.ResponseWriter, r *http.Request) {
	// Deliberately not filtered by the date range. This is the document an
	// authority reads and it is a reproduction of the whole logbook -- all
	// three paper books as one continuous EASA-format record. A partial one
	// would understate a licence total, which is the dangerous direction.
	flights, err := s.db.Flights()
	if err != nil {
		s.log.Error("reading flights for the EASA export", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read the logbook")
		return
	}

	doc, err := pdfbook.EASA(flights, s.exportOptions())
	if err != nil {
		s.log.Error("rendering the EASA export", "error", err)
		writeError(w, http.StatusInternalServerError, "could not render the document")
		return
	}
	s.servePDF(w, r, doc, "logbook-easa.pdf")
}

func (s *Server) handleExportTable(w http.ResponseWriter, r *http.Request) {
	rng, err := rangeOf(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	flights, err := s.flightsIn(rng)
	if err != nil {
		s.log.Error("reading flights for the table export", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read the logbook")
		return
	}

	doc, err := pdfbook.Table(flights, rng, s.exportOptions())
	if err != nil {
		s.log.Error("rendering the table export", "error", err)
		writeError(w, http.StatusInternalServerError, "could not render the document")
		return
	}
	s.servePDF(w, r, doc, filenameFor("logbook-table", rng))
}

func (s *Server) handleExportStatistics(w http.ResponseWriter, r *http.Request) {
	rng, err := rangeOf(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	flights, err := s.flightsIn(rng)
	if err != nil {
		s.log.Error("reading flights for the statistics export", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read the logbook")
		return
	}

	doc, err := pdfbook.Statistics(stats.Summarize(flights), rng, s.exportOptions())
	if err != nil {
		s.log.Error("rendering the statistics export", "error", err)
		writeError(w, http.StatusInternalServerError, "could not render the document")
		return
	}
	s.servePDF(w, r, doc, filenameFor("logbook-statistics", rng))
}

// exportOptions is the non-flight content of a document.
func (s *Server) exportOptions() pdfbook.Options {
	return pdfbook.Options{
		HolderName: s.cfg.HolderName,
		Generated:  s.now().UTC(),
	}
}

// servePDF writes a rendered document.
//
// Content-Disposition is "attachment" rather than "inline": these are records
// to be filed, and a PDF rendered inside the page would be showing personal
// data in a viewer that is not ours. The no-store header from
// setSecurityHeaders already applies -- an exported logbook must not sit in a
// shared device's cache.
func (s *Server) servePDF(w http.ResponseWriter, r *http.Request, doc []byte, filename string) {
	h := w.Header()
	h.Set("Content-Type", "application/pdf")
	h.Set("Content-Length", fmt.Sprintf("%d", len(doc)))
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if _, err := w.Write(doc); err != nil {
		// The headers are already out, so there is nothing to tell the client.
		s.log.Error("writing an export", "error", err, "path", r.URL.Path)
	}
}

// filenameFor names a download after what is in it, so that two exports of
// different ranges do not overwrite each other in a downloads folder.
func filenameFor(base string, rng stats.Range) string {
	switch {
	case rng.From == "" && rng.To == "":
		return base + ".pdf"
	case rng.From == "":
		return fmt.Sprintf("%s-to-%s.pdf", base, rng.To)
	case rng.To == "":
		return fmt.Sprintf("%s-from-%s.pdf", base, rng.From)
	default:
		return fmt.Sprintf("%s-%s-to-%s.pdf", base, rng.From, rng.To)
	}
}
