// Package timeutil is the single authority for turning the times written in
// the paper logbooks into UTC instants (CLAUDE.md rule 0.4).
//
// The books mix zones: a "Z" suffix marks a row already written in UTC,
// everything else is Helsinki local. Conversion must therefore be explicit,
// must use the DST rules in force on the flight's own date, and must never
// guess when the local time is genuinely ambiguous.
//
// Do not re-implement any of this elsewhere.
package timeutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	// Embed the timezone database in the binary so behaviour never depends on
	// the server having up-to-date zoneinfo. The deploy target is an Ubuntu
	// box that has been up for 844 days.
	_ "time/tzdata"
)

// Origin records how a stored UTC instant was arrived at, so that a conversion
// can always be audited or revisited. The raw string as written on paper is
// kept alongside it by the caller.
type Origin string

const (
	// OriginUTCAsWritten - the book wrote a "Z" time; no conversion happened.
	OriginUTCAsWritten Origin = "utc_as_written"
	// OriginConvertedFromLocal - a Helsinki local time we converted.
	OriginConvertedFromLocal Origin = "converted_from_local"
	// OriginUnknown - the value could not be resolved with confidence. The row
	// carries a best-effort instant and must surface for human review.
	OriginUnknown Origin = "unknown"
	// OriginNone - the cell was blank.
	OriginNone Origin = "none"
)

// helsinki is the zone every local time in these logbooks was written in.
var helsinki = mustLoadLocation("Europe/Helsinki")

// mustLoadLocation panics rather than returning an error.
//
// A missing zone database means every converted row would silently fall back
// to UTC and be two or three hours wrong -- a corrupted legal record that
// looks entirely plausible. Failing at startup is the only safe response.
func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("timeutil: zone " + name + " unavailable: " + err.Error())
	}
	return loc
}

// ToUTC converts one clock time written on the given date into a UTC instant.
//
// date is YYYY-MM-DD. raw is the cell exactly as written: "15:30" for local,
// "07:56Z" for an already-UTC row, or empty for a blank cell.
func ToUTC(date, raw string) (time.Time, Origin, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, OriginNone, nil
	}

	y, mo, d, err := parseDate(date)
	if err != nil {
		return time.Time{}, OriginUnknown, err
	}

	// A trailing Z (either case) means the pilot already wrote UTC.
	if u := strings.TrimSuffix(strings.TrimSuffix(raw, "Z"), "z"); u != raw {
		hh, mi, err := parseClock(strings.TrimSpace(u))
		if err != nil {
			return time.Time{}, OriginUnknown, err
		}
		return time.Date(y, mo, d, hh, mi, 0, 0, time.UTC), OriginUTCAsWritten, nil
	}

	hh, mi, err := parseClock(raw)
	if err != nil {
		return time.Time{}, OriginUnknown, err
	}

	local := time.Date(y, mo, d, hh, mi, 0, 0, helsinki)

	// Spring forward: the wall clock we were given never existed. time.Date
	// silently normalizes it (03:30 becomes 04:30), so detect that by checking
	// whether the value we got back still reads as what we asked for.
	if local.Hour() != hh || local.Minute() != mi {
		return local.UTC(), OriginUnknown, nil
	}

	// Autumn fold: the wall clock happened twice, so two distinct instants
	// share it and we cannot know which the pilot meant.
	//
	// Detected by shifting the instant an hour each way: if either shift still
	// reads as the same wall clock, the clock reading is ambiguous. Both
	// directions are checked because time.Date explicitly does not guarantee
	// which of the two offsets it picks, and we must not depend on that.
	if sameWallClock(local.Add(time.Hour), hh, mi) || sameWallClock(local.Add(-time.Hour), hh, mi) {
		return local.UTC(), OriginUnknown, nil
	}

	return local.UTC(), OriginConvertedFromLocal, nil
}

// BlockPair converts an off-block/on-block pair written on the same date.
//
// It handles the two things a single-value conversion cannot: a flight that
// crosses midnight, and a pair whose two halves disagree about their zone.
func BlockPair(date, offRaw, onRaw string) (off, on time.Time, origin Origin, err error) {
	off, offOrigin, err := ToUTC(date, offRaw)
	if err != nil {
		return time.Time{}, time.Time{}, OriginUnknown, err
	}
	on, onOrigin, err := ToUTC(date, onRaw)
	if err != nil {
		return time.Time{}, time.Time{}, OriginUnknown, err
	}

	origin = combineOrigins(offOrigin, onOrigin)

	// On-block at or before off-block means the flight landed after midnight.
	// Without this the row would carry a negative duration.
	if !off.IsZero() && !on.IsZero() && !on.After(off) {
		on = on.AddDate(0, 0, 1)
	}
	return off, on, origin, nil
}

// combineOrigins reduces the two halves of a block pair to one verdict.
//
// A pair where one half is Zulu and the other is not cannot be trusted: a
// single spread really does mix zones (confirmed on IMG_6014), so a
// half-converted pair would produce a plausible but wrong duration.
func combineOrigins(a, b Origin) Origin {
	switch {
	case a == OriginNone && b == OriginNone:
		return OriginNone
	case a == OriginUnknown || b == OriginUnknown:
		return OriginUnknown
	// One half blank: trust whichever half actually carries a time.
	case a == OriginNone:
		return b
	case b == OriginNone:
		return a
	case a != b:
		return OriginUnknown
	default:
		return a
	}
}

func sameWallClock(t time.Time, hh, mi int) bool {
	return t.Hour() == hh && t.Minute() == mi
}

func parseDate(s string) (int, time.Month, int, error) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("timeutil: %q is not a YYYY-MM-DD date", s)
	}
	y, mo, d := t.Date()
	return y, mo, d, nil
}

func parseClock(s string) (int, int, error) {
	h, m, ok := strings.Cut(s, ":")
	if !ok {
		return 0, 0, fmt.Errorf("timeutil: %q has no ':' separator", s)
	}
	hh, err := strconv.Atoi(h)
	if err != nil {
		return 0, 0, fmt.Errorf("timeutil: %q has a non-numeric hour", s)
	}
	mi, err := strconv.Atoi(m)
	if err != nil {
		return 0, 0, fmt.Errorf("timeutil: %q has a non-numeric minute", s)
	}
	if hh < 0 || hh > 23 {
		return 0, 0, fmt.Errorf("timeutil: %q hour %d out of range 0-23", s, hh)
	}
	if mi < 0 || mi > 59 {
		return 0, 0, fmt.Errorf("timeutil: %q minute %d out of range 0-59", s, mi)
	}
	return hh, mi, nil
}
