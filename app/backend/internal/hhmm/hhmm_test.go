package hhmm

import "testing"

// Durations in the paper logbook are written H:MM and can exceed 24 hours
// (the book's running totals reach four digits, e.g. "1040:33"). They are
// parsed once at the edge and carried as integer minutes everywhere else --
// a legal record cannot afford float rounding.

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"simple", "1:21", 81},
		{"zero padded hour", "01:21", 81},
		{"exact hour", "2:00", 120},
		{"zero", "0:00", 0},
		{"empty means zero", "", 0},
		{"whitespace only", "   ", 0},
		{"surrounding whitespace", " 1:21 ", 81},
		{"beyond a day", "27:45", 1665},
		{"four digit hours, a book running total", "1040:33", 62433},
		{"fifty nine minutes", "0:59", 59},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Parse(c.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("Parse(%q) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

func TestParseRejectsMalformed(t *testing.T) {
	// Anything we cannot read exactly is an error, never a silent zero.
	// Silently coercing a bad value to 0 would corrupt a total (rule 0.2).
	cases := []struct {
		name string
		in   string
	}{
		{"no colon", "121"},
		{"minutes out of range", "1:60"},
		{"minutes far out of range", "1:99"},
		{"negative", "-1:21"},
		{"non numeric hour", "a:21"},
		{"non numeric minute", "1:2b"},
		{"too many parts", "1:21:33"},
		{"empty minute", "1:"},
		{"empty hour", ":21"},
		{"single digit minute is ambiguous", "1:2"},
		{"decimal", "1.5"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, err := Parse(c.in); err == nil {
				t.Errorf("Parse(%q) = %d with no error; want an error", c.in, got)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{81, "1:21"},
		{0, "0:00"},
		{59, "0:59"},
		{60, "1:00"},
		{1665, "27:45"},
		{62433, "1040:33"},
	}
	for _, c := range cases {
		if got := Format(c.in); got != c.want {
			t.Errorf("Format(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatClampsNegative(t *testing.T) {
	// A negative duration is meaningless in a logbook. It should never reach
	// Format, but if a subtraction bug ever produces one we render 0:00 rather
	// than "-1:-21", which would be silently wrong on an exported PDF.
	if got := Format(-5); got != "0:00" {
		t.Errorf("Format(-5) = %q, want %q", got, "0:00")
	}
	if got := FormatBlank(-5); got != "" {
		t.Errorf("FormatBlank(-5) = %q, want empty string", got)
	}
}

func TestFormatParseRoundTrip(t *testing.T) {
	// The PDF exports re-parse what they format; the pair must be lossless.
	for _, m := range []int{0, 1, 59, 60, 81, 599, 1665, 62433} {
		got, err := Parse(Format(m))
		if err != nil {
			t.Fatalf("Parse(Format(%d)) errored: %v", m, err)
		}
		if got != m {
			t.Errorf("round trip of %d gave %d", m, got)
		}
	}
}

func TestFormatBlankRendersEmpty(t *testing.T) {
	// The logbook leaves a cell blank rather than writing 0:00. The PDF needs
	// that distinction to look like the real book.
	if got := FormatBlank(0); got != "" {
		t.Errorf("FormatBlank(0) = %q, want empty string", got)
	}
	if got := FormatBlank(81); got != "1:21" {
		t.Errorf("FormatBlank(81) = %q, want %q", got, "1:21")
	}
}
