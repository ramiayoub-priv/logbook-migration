// Package entry validates a flight typed into the application by hand and
// turns it into the same domain record the CSV importer produces.
//
// It exists as its own package, with no database and no HTTP in it, for the
// same reason internal/csvbook does: the rules that decide whether something
// may be written into a legal record are the part that has to be right, and
// they should be testable without a schema or a request in the way.
//
// The posture here is deliberately the opposite of the importer's. CLAUDE.md
// rule 0.2 tells the importer to surface a problem and import the row anyway,
// because the paper book is authoritative and refusing to load it would leave
// the owner with no application. Nothing on this path is authoritative yet:
// the pilot is standing at the form and can be asked. So a draft that does not
// make sense is refused, with the field named, rather than stored with a flag
// on it. Surfacing a discrepancy and creating one are different acts.
package entry

import (
	"fmt"
	"strings"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/hhmm"
	"github.com/ramiayoub/logbook/backend/internal/timeutil"
)

// maxFlightMinutes bounds a single flight at 24 hours.
//
// The longest flight in 14 years of these books is 4:34, so this is not a
// domain rule so much as a guard against a typo: "25:00" for "2:50" would
// otherwise add a day to a licence total.
const maxFlightMinutes = 24 * 60

// HandEnteredBook is the source_book value marking a row that was typed into
// the application rather than transcribed from a paper book.
//
// It is load-bearing in two places and must not be changed casually: the
// importer keys on it to know which rows it owns and may replace, and the
// store allocates these rows their seq from a separate band. Books are
// numbered from 1, so 0 is free and reads correctly as "no paper book".
const HandEnteredBook = 0

// FieldError is one problem with one field. The field name is the wire name of
// the form input, so the page can highlight the control the pilot has to fix
// rather than printing a sentence at the top.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e FieldError) Error() string { return e.Field + ": " + e.Message }

// Errors is every problem found with one draft.
//
// Validation collects rather than returning at the first fault: a flight entry
// has twenty fields, and a form that reveals one mistake per submission is a
// form that gets abandoned halfway through -- which on this app means the
// flight does not get logged at all.
type Errors []FieldError

func (e Errors) Error() string {
	parts := make([]string, 0, len(e))
	for _, fe := range e {
		parts = append(parts, fe.Error())
	}
	return strings.Join(parts, "; ")
}

// Draft is a flight exactly as the form submitted it: strings, unvalidated,
// uncorrected. Durations are H:MM because that is what a pilot writes; they
// become minutes the moment they are validated and are minutes everywhere
// afterwards.
type Draft struct {
	Date         string `json:"date"`
	AircraftReg  string `json:"aircraft_reg"`
	AircraftType string `json:"aircraft_type"`
	Class        string `json:"class"`
	DepPlace     string `json:"dep_place"`
	ArrPlace     string `json:"arr_place"`

	// As written: "14:05" for Helsinki local, "14:05Z" for UTC.
	OffBlock string `json:"off_block"`
	OnBlock  string `json:"on_block"`

	TotalTime      string `json:"total_time"`
	NightTime      string `json:"night_time"`
	InstrumentTime string `json:"instrument_time"`
	PICTime        string `json:"pic_time"`
	DualTime       string `json:"dual_time"`
	InstructorTime string `json:"instructor_time"`
	CopilotTime    string `json:"copilot_time"`
	MultiPilotTime string `json:"multi_pilot_time"`

	PICName       string `json:"pic_name"`
	LandingsDay   int    `json:"landings_day"`
	LandingsNight int    `json:"landings_night"`
	Remarks       string `json:"remarks"`
}

// Validate turns a draft into a flight, or reports everything wrong with it.
//
// now is passed in rather than read from the clock so that "not in the future"
// is testable, and so this package keeps the project's one-authority-for-time
// discipline (rule 0.4) instead of quietly acquiring a second one.
//
// The returned flight has no Seq: allocating book order is the store's job,
// because it is the only thing that knows what order the book is already in.
func Validate(d Draft, now time.Time) (csvbook.Flight, error) {
	var errs Errors
	bad := func(field, format string, args ...any) {
		errs = append(errs, FieldError{Field: field, Message: fmt.Sprintf(format, args...)})
	}

	f := csvbook.Flight{
		AircraftReg:  strings.ToUpper(strings.TrimSpace(d.AircraftReg)),
		AircraftType: strings.ToUpper(strings.TrimSpace(d.AircraftType)),
		DepPlace:     strings.ToUpper(strings.TrimSpace(d.DepPlace)),
		ArrPlace:     strings.ToUpper(strings.TrimSpace(d.ArrPlace)),
		PICName:      strings.TrimSpace(d.PICName),
		Remarks:      strings.TrimSpace(d.Remarks),

		LandingsDay:   d.LandingsDay,
		LandingsNight: d.LandingsNight,

		// The pilot typed the day/night split, so it is read from a person
		// rather than inferred from night time. That is exactly the property
		// the imported rows lack and Task 8 exists to fix.
		LandingsVerified: true,

		SourceBook: HandEnteredBook,
	}

	f.Date = validateDate(strings.TrimSpace(d.Date), now, bad)

	if f.AircraftReg == "" {
		bad("aircraft_reg", "a registration is required")
	}
	if f.AircraftType == "" {
		bad("aircraft_type", "an aircraft type is required")
	}

	class := csvbook.Class(strings.ToUpper(strings.TrimSpace(d.Class)))
	if !csvbook.ValidClass(class) {
		bad("class", "%q is not a known class", d.Class)
	} else {
		f.Class = class
	}

	validateDurations(d, &f, bad)
	validateLandings(d, bad)

	// Times are converted only if the date parsed: a conversion needs the
	// flight's own date to know which DST rules were in force, and reporting
	// "unparseable time" for what is really a bad date sends the pilot to the
	// wrong field.
	if f.Date != "" {
		validateTimes(d, &f, bad)
	}

	// Block time is chocks-to-chocks, taken from the clock. Total is what the
	// pilot typed and is what the book totals on; the two genuinely differ on
	// one row of the paper books, so a mismatch between them is recorded
	// rather than rejected.
	if !f.OffBlockUTC.IsZero() && !f.OnBlockUTC.IsZero() {
		f.BlockMinutes = int(f.OnBlockUTC.Sub(f.OffBlockUTC).Minutes())
	}

	if len(errs) > 0 {
		return csvbook.Flight{}, errs
	}
	return f, nil
}

// validateDate checks the one field everything else is interpreted against.
func validateDate(date string, now time.Time, bad func(string, string, ...any)) string {
	if date == "" {
		bad("date", "a date is required")
		return ""
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		// time.Parse rejects both a wrong shape and an impossible day, and the
		// pilot needs the same thing said about either.
		bad("date", "%q is not a real date written as YYYY-MM-DD", date)
		return ""
	}
	// A flight that has not happened yet cannot be logged. Compared on the
	// calendar day in UTC, which is the zone every instant in this app is in.
	if parsed.After(now.UTC().Truncate(24 * time.Hour)) {
		bad("date", "%s is in the future", date)
		return ""
	}
	return date
}

// validateDurations parses every H:MM field and enforces the one arithmetic
// rule that is impossible rather than merely unusual: no component of a flight
// can exceed the flight.
func validateDurations(d Draft, f *csvbook.Flight, bad func(string, string, ...any)) {
	// The total is required rather than derived from the off/on-block clock.
	// It is the figure the whole logbook totals on and the one a licence
	// application is written from, so it is the pilot's to state -- the form
	// can prefill it from the clock, but the server inventing it would put a
	// number nobody typed into a legal record. It is also why Block and Total
	// are allowed to differ at all.
	total, totalOK := parseDuration("total_time", d.TotalTime, bad)
	switch {
	case !totalOK:
		// Already reported as unparseable; do not also call it missing.
	case strings.TrimSpace(d.TotalTime) == "":
		bad("total_time", "a total time is required")
		totalOK = false
	case total == 0:
		bad("total_time", "a flight cannot last no time at all")
	case total > maxFlightMinutes:
		bad("total_time", "%s is longer than a day; check for a typo", d.TotalTime)
	}
	f.TotalMinutes = total

	// Every component is measured against the total. A component larger than
	// the flight it belongs to is arithmetically impossible, and the importer
	// has a whole discrepancy kind for the one paper row that does it.
	for _, c := range []struct {
		field string
		raw   string
		dst   *int
	}{
		{"night_time", d.NightTime, &f.NightMinutes},
		{"instrument_time", d.InstrumentTime, &f.InstrumentMinutes},
		{"pic_time", d.PICTime, &f.PICMinutes},
		{"dual_time", d.DualTime, &f.DualMinutes},
		{"instructor_time", d.InstructorTime, &f.InstructorMinutes},
		{"copilot_time", d.CopilotTime, &f.CopilotMinutes},
		{"multi_pilot_time", d.MultiPilotTime, &f.MultiPilotMinutes},
	} {
		v, ok := parseDuration(c.field, c.raw, bad)
		if !ok {
			continue
		}
		if totalOK && v > total {
			bad(c.field, "%s exceeds the flight's total of %s", hhmm.Format(v), hhmm.Format(total))
			continue
		}
		*c.dst = v
	}
}

// parseDuration reads one optional H:MM field. Blank is zero, not a fault:
// most flights log neither night nor instrument time.
func parseDuration(field, raw string, bad func(string, string, ...any)) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, true
	}
	v, err := hhmm.Parse(raw)
	if err != nil {
		bad(field, "%q is not a duration written as H:MM", raw)
		return 0, false
	}
	return v, true
}

func validateLandings(d Draft, bad func(string, string, ...any)) {
	if d.LandingsDay < 0 {
		bad("landings_day", "landings cannot be negative")
	}
	if d.LandingsNight < 0 {
		bad("landings_night", "landings cannot be negative")
	}
	// Every one of the 1295 rows in the paper books has at least one landing,
	// because every flight ends in one. Zero is a forgotten field.
	if d.LandingsDay == 0 && d.LandingsNight == 0 {
		bad("landings_day", "a flight has at least one landing")
	}
}

// validateTimes converts the off/on-block pair through the single conversion
// authority and refuses anything it cannot resolve.
//
// Refusing is the whole point. On an imported row an ambiguous clock time gets
// stored with time_origin=unknown, because the page is 14 years old and nobody
// can be asked. Here the pilot is at the form, so storing a guess would be
// manufacturing an unaudited instant when the true one is a question away.
func validateTimes(d Draft, f *csvbook.Flight, bad func(string, string, ...any)) {
	off := strings.TrimSpace(d.OffBlock)
	on := strings.TrimSpace(d.OnBlock)
	if off == "" {
		bad("off_block", "an off-block time is required")
	}
	if on == "" {
		bad("on_block", "an on-block time is required")
	}
	if off == "" || on == "" {
		return
	}

	// BlockPair is the only thing that parses these, so that the midnight roll
	// and the mixed-zone verdict stay in the one conversion authority (rule
	// 0.4) instead of being reimplemented here.
	offUTC, onUTC, origin, err := timeutil.BlockPair(f.Date, off, on)
	if err != nil {
		// It does not say which half it choked on, and the form has to
		// highlight the control the pilot must retype -- so ask each half.
		field, raw := "off_block", off
		if _, _, e := timeutil.ToUTC(f.Date, off); e == nil {
			field, raw = "on_block", on
		}
		bad(field, "%q is not a time written as HH:MM, optionally with a Z for UTC", raw)
		return
	}
	if origin == timeutil.OriginUnknown {
		// The two cases this catches are a daylight-saving gap or fold, and a
		// pair where one half is Zulu and the other is not. Both are resolved
		// by writing UTC, so the message says so.
		bad("off_block", "this local time is ambiguous on %s because the clocks change; "+
			"write both times in UTC with a Z, for example %sZ", f.Date, off)
		return
	}
	f.OffBlockUTC, f.OnBlockUTC = offUTC, onUTC
	// Kept exactly as typed, alongside the converted instant, so that a
	// conversion can always be audited or revisited (rule 0.4).
	f.OffBlockRaw, f.OnBlockRaw = off, on
	f.TimeOrigin = origin
}
