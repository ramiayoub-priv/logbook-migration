// Package stats computes every aggregate the application reports.
//
// Nothing here is stored. CLAUDE.md rule 0.5 forbids a persisted running total
// anywhere in this app: the seven Cumulative_* columns in the source CSVs are
// the single largest source of drift in this project's history, and the fix was
// to derive them instead. So the EASA PDF's "TOTAL PREVIOUS PAGES" block, the
// statistics page and the table page all call into this package at render time
// and none of them read a stored figure.
//
// Two ordering rules apply and they are not interchangeable:
//
//   - Anything cumulative walks Seq, the explicit book order. Eighteen rows
//     across the three books are genuinely out of date order on paper, so
//     ordering on the date would silently reshuffle the pilot's logbook.
//   - Anything filtered by the user walks the date, because "my 2015" means the
//     flights flown in 2015 whatever order they were written down in.
//
// Durations are whole minutes throughout, as everywhere inside this app;
// internal/hhmm converts at the edges.
package stats

import (
	"fmt"
	"slices"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
)

// Range is an inclusive date filter over YYYY-MM-DD strings. An empty bound is
// open, so the zero Range is the whole logbook -- which is what the statistics
// page opens on.
//
// The dates are compared as strings. YYYY-MM-DD is fixed-width and
// zero-padded, so lexicographic order is chronological order, and the same
// property is what lets the SQL layer range-query the column directly.
type Range struct {
	From string // inclusive; empty means no lower bound
	To   string // inclusive; empty means no upper bound
}

// Contains reports whether a flight on the given date falls in the range.
//
// An inverted range (From after To) contains nothing. That is deliberate
// rather than an error: the UI has two date pickers and the user can set them
// in either order, and an empty result set is the honest answer to "flights
// between December and January".
func (r Range) Contains(date string) bool {
	if r.From != "" && date < r.From {
		return false
	}
	if r.To != "" && date > r.To {
		return false
	}
	return true
}

// Filter returns the flights whose date falls in the range, in the order they
// were given. It always allocates a fresh slice, so the result never shares a
// backing array with the caller's: this is a legal record and an aliased slice
// is a way to corrupt one at a distance.
func Filter(flights []csvbook.Flight, r Range) []csvbook.Flight {
	out := make([]csvbook.Flight, 0, len(flights))
	for _, f := range flights {
		if r.Contains(f.Date) {
			out = append(out, f)
		}
	}
	return out
}

// Summary is every figure the statistics page reports, over whatever set of
// flights it was computed from. All durations are minutes.
//
// It is a comparable value type on purpose: tests assert on whole summaries
// rather than field by field, so a field added here without being accumulated
// in Summarize fails the partition tests immediately.
type Summary struct {
	Flights int

	// The whole-logbook columns.
	Total      int
	PIC        int
	Dual       int
	Night      int
	Instrument int
	Instructor int

	// The sea/land split. SeaTotal+LandTotal == Total, and likewise for PIC and
	// Instructor. The owner holds a seaplane rating and an instructor rating,
	// and these are the figures that evidence them.
	SeaTotal       int
	SeaPIC         int
	SeaInstructor  int
	LandTotal      int
	LandPIC        int
	LandInstructor int

	// Landings, split on two independent axes over the same total:
	// LandingsDay+LandingsNight == LandingsSea+LandingsLand.
	LandingsDay   int
	LandingsNight int
	LandingsSea   int
	LandingsLand  int

	// How many of those flights had their day/night landing split inferred
	// rather than read off the page. Surfaced rather than hidden (rule 0.2):
	// the statistics page must be able to say the night landing figure is not
	// yet fully verified. Task 8 drives this to zero.
	LandingsUnverified int
}

// Add returns the sum of two summaries. It is what builds the EASA page block's
// "TOTAL PREVIOUS PAGES" without re-walking every earlier flight, and it agrees
// exactly with summarizing the concatenation -- asserted in the tests, because
// a page total that disagrees with its own rows is precisely the drift this
// package exists to prevent.
func (s Summary) Add(o Summary) Summary {
	return Summary{
		Flights:            s.Flights + o.Flights,
		Total:              s.Total + o.Total,
		PIC:                s.PIC + o.PIC,
		Dual:               s.Dual + o.Dual,
		Night:              s.Night + o.Night,
		Instrument:         s.Instrument + o.Instrument,
		Instructor:         s.Instructor + o.Instructor,
		SeaTotal:           s.SeaTotal + o.SeaTotal,
		SeaPIC:             s.SeaPIC + o.SeaPIC,
		SeaInstructor:      s.SeaInstructor + o.SeaInstructor,
		LandTotal:          s.LandTotal + o.LandTotal,
		LandPIC:            s.LandPIC + o.LandPIC,
		LandInstructor:     s.LandInstructor + o.LandInstructor,
		LandingsDay:        s.LandingsDay + o.LandingsDay,
		LandingsNight:      s.LandingsNight + o.LandingsNight,
		LandingsSea:        s.LandingsSea + o.LandingsSea,
		LandingsLand:       s.LandingsLand + o.LandingsLand,
		LandingsUnverified: s.LandingsUnverified + o.LandingsUnverified,
	}
}

// IsSea reports whether a class is flown on floats.
//
// The class comes from the registration, not the aircraft type: the paper book
// only starts writing "C172sea" from IMG_6022 and is inconsistent after that,
// whereas the registration rule reproduces the Cumulative_SEP_Sea column at
// every one of the 1293 rows. See the 2026-08-01 decision-log entry in
// app/APP.md.
//
// Land is the default, so a class added to the vocabulary later cannot silently
// inflate the seaplane figures that a rating depends on -- it shows up as a
// land discrepancy instead, which is the safer direction to be wrong in.
func IsSea(c csvbook.Class) bool {
	return c == csvbook.ClassSEPSea || c == csvbook.ClassMEPSea
}

// Summarize aggregates a set of flights. Order does not matter: addition is
// commutative and nothing here is cumulative.
//
// An empty set gives the zero Summary. That is a real case, not a degenerate
// one -- a winter month in Finland has no flights in it.
func Summarize(flights []csvbook.Flight) Summary {
	var s Summary
	for _, f := range flights {
		s.Flights++
		s.Total += f.TotalMinutes
		s.PIC += f.PICMinutes
		s.Dual += f.DualMinutes
		s.Night += f.NightMinutes
		s.Instrument += f.InstrumentMinutes
		s.Instructor += f.InstructorMinutes

		s.LandingsDay += f.LandingsDay
		s.LandingsNight += f.LandingsNight
		if !f.LandingsVerified {
			s.LandingsUnverified++
		}

		n := f.LandingsDay + f.LandingsNight
		if IsSea(f.Class) {
			s.SeaTotal += f.TotalMinutes
			s.SeaPIC += f.PICMinutes
			s.SeaInstructor += f.InstructorMinutes
			s.LandingsSea += n
			continue
		}
		s.LandTotal += f.TotalMinutes
		s.LandPIC += f.PICMinutes
		s.LandInstructor += f.InstructorMinutes
		s.LandingsLand += n
	}
	return s
}

// Page is one page of the EASA logbook reproduction: its rows plus the
// three-row totals block the paper page carries at the bottom.
type Page struct {
	Number  int // 1-based
	Flights []csvbook.Flight

	// The paper page's three totals rows, in its own words.
	ThisPage Summary // TOTAL THIS PAGE
	Previous Summary // TOTAL PREVIOUS PAGES
	Total    Summary // TOTAL -- always Previous + ThisPage
}

// Paginate splits the flights into fixed-size pages in Seq order and computes
// each page's totals block. The EASA book this reproduces runs 15 rows to a
// page.
//
// The caller's slice is never reordered: sorting it in place would renumber the
// pilot's logbook as a side effect of rendering it.
func Paginate(flights []csvbook.Flight, rows int) ([]Page, error) {
	if rows <= 0 {
		return nil, fmt.Errorf("stats: page size must be positive, got %d", rows)
	}

	// Sorted copy: Seq is the explicit book order and the only ordering a
	// cumulative may walk. Sorting on the date would reshuffle the 18 rows that
	// are genuinely out of date order on paper.
	ordered := slices.Clone(flights)
	slices.SortFunc(ordered, func(a, b csvbook.Flight) int { return a.Seq - b.Seq })

	var pages []Page
	var previous Summary
	for start := 0; start < len(ordered); start += rows {
		end := min(start+rows, len(ordered))

		// The three-arg slice caps the result at its own length, so a caller
		// appending to one page's Flights cannot overwrite the next page's
		// rows.
		page := ordered[start:end:end]
		thisPage := Summarize(page)
		pages = append(pages, Page{
			Number:   len(pages) + 1,
			Flights:  page,
			ThisPage: thisPage,
			Previous: previous,
			Total:    previous.Add(thisPage),
		})
		previous = pages[len(pages)-1].Total
	}
	return pages, nil
}
