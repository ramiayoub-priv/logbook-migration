// Package hhmm converts between the H:MM durations written in the paper
// logbook and the integer minutes the application carries internally.
//
// Durations are minutes everywhere inside the app. H:MM is a display and
// interchange format only, parsed at the edges. Hours are unbounded: the
// book's running totals run to four digits.
package hhmm

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse reads a H:MM duration and returns whole minutes.
//
// An empty or all-whitespace string means "blank cell" and yields 0. Anything
// else that cannot be read exactly is an error -- never a silent zero, because
// a swallowed value would corrupt a total (CLAUDE.md rule 0.2).
func Parse(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	h, m, ok := strings.Cut(s, ":")
	if !ok {
		return 0, fmt.Errorf("hhmm: %q has no ':' separator", s)
	}
	// Minutes are always two digits in the book; a single digit is ambiguous
	// (is "1:2" two minutes or twenty?) so we refuse to guess.
	if len(m) != 2 {
		return 0, fmt.Errorf("hhmm: %q minute field must be exactly two digits", s)
	}
	if h == "" {
		return 0, fmt.Errorf("hhmm: %q has an empty hour field", s)
	}

	hours, err := strconv.Atoi(h)
	if err != nil {
		return 0, fmt.Errorf("hhmm: %q has a non-numeric hour field", s)
	}
	mins, err := strconv.Atoi(m)
	if err != nil {
		return 0, fmt.Errorf("hhmm: %q has a non-numeric minute field", s)
	}
	// Atoi accepts a leading sign; a duration may not be negative.
	if hours < 0 || mins < 0 {
		return 0, fmt.Errorf("hhmm: %q is negative", s)
	}
	if mins > 59 {
		return 0, fmt.Errorf("hhmm: %q has %d minutes, must be 0-59", s, mins)
	}
	return hours*60 + mins, nil
}

// Format renders whole minutes as H:MM, zero-padding only the minutes. It is
// the exact inverse of Parse.
func Format(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}
	return fmt.Sprintf("%d:%02d", minutes/60, minutes%60)
}

// FormatBlank is Format, except that zero renders as an empty string.
//
// The paper logbook leaves a cell blank rather than writing 0:00, and the EASA
// PDF clone has to reproduce that to look like the real book.
func FormatBlank(minutes int) string {
	if minutes <= 0 {
		return ""
	}
	return Format(minutes)
}
