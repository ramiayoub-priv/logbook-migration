package timeutil

import (
	"testing"
	"time"
)

// The paper books mix local and UTC times. A "Z" suffix marks a row that was
// already written in UTC; everything else is Helsinki local. This package is
// the single authority for that conversion (CLAUDE.md rule 0.4).

func TestToUTC_ZSuffixIsTakenAsWritten(t *testing.T) {
	cases := []string{"07:56Z", "07:56z", " 07:56Z "}
	for _, raw := range cases {
		got, origin, err := ToUTC("2024-08-20", raw)
		if err != nil {
			t.Fatalf("ToUTC(%q) errored: %v", raw, err)
		}
		want := time.Date(2024, 8, 20, 7, 56, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("ToUTC(%q) = %v, want %v", raw, got, want)
		}
		if origin != OriginUTCAsWritten {
			t.Errorf("ToUTC(%q) origin = %q, want %q", raw, origin, OriginUTCAsWritten)
		}
	}
}

func TestToUTC_LocalSummerIsEEST(t *testing.T) {
	// This exact pair is the proof recorded in claude-docs/reference.md:
	// laskukierros block_start 18:50 local == the paper's 15:50Z. If this test
	// ever fails, the conversion is wrong against known-good real data.
	got, origin, err := ToUTC("2024-08-20", "18:50")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	want := time.Date(2024, 8, 20, 15, 50, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v (Helsinki is UTC+3 in August)", got, want)
	}
	if origin != OriginConvertedFromLocal {
		t.Errorf("origin = %q, want %q", origin, OriginConvertedFromLocal)
	}
}

func TestToUTC_LocalWinterIsEET(t *testing.T) {
	got, _, err := ToUTC("2024-03-12", "15:30")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	want := time.Date(2024, 3, 12, 13, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v (Helsinki is UTC+2 in March)", got, want)
	}
}

func TestToUTC_HistoricalDatesUseTheRulesOfTheirOwnEra(t *testing.T) {
	// The logbook starts in 2012. Conversion must not apply today's rules to
	// an old date; tzdata is embedded so this never depends on the server.
	cases := []struct {
		date, raw string
		want      time.Time
	}{
		{"2012-08-15", "14:00", time.Date(2012, 8, 15, 11, 0, 0, 0, time.UTC)},  // EEST +3
		{"2012-01-15", "14:00", time.Date(2012, 1, 15, 12, 0, 0, 0, time.UTC)},  // EET  +2
		{"2019-12-31", "09:15", time.Date(2019, 12, 31, 7, 15, 0, 0, time.UTC)}, // EET  +2
	}
	for _, c := range cases {
		got, _, err := ToUTC(c.date, c.raw)
		if err != nil {
			t.Fatalf("ToUTC(%q,%q) errored: %v", c.date, c.raw, err)
		}
		if !got.Equal(c.want) {
			t.Errorf("ToUTC(%q,%q) = %v, want %v", c.date, c.raw, got, c.want)
		}
	}
}

func TestToUTC_EmptyMeansNoTime(t *testing.T) {
	for _, raw := range []string{"", "   "} {
		got, origin, err := ToUTC("2024-08-20", raw)
		if err != nil {
			t.Fatalf("ToUTC(%q) errored: %v", raw, err)
		}
		if !got.IsZero() {
			t.Errorf("ToUTC(%q) = %v, want the zero time", raw, got)
		}
		if origin != OriginNone {
			t.Errorf("ToUTC(%q) origin = %q, want %q", raw, origin, OriginNone)
		}
	}
}

func TestToUTC_SpringForwardGapIsFlaggedUnknown(t *testing.T) {
	// Finland springs forward on the last Sunday of March: 03:00 EET becomes
	// 04:00 EEST, so local times 03:00-03:59 never happened that day. We must
	// not silently normalize them (Go's time.Date would quietly give 04:30).
	_, origin, err := ToUTC("2024-03-31", "03:30")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	if origin != OriginUnknown {
		t.Errorf("origin = %q, want %q for a time inside the DST gap", origin, OriginUnknown)
	}
}

func TestToUTC_AutumnFoldIsFlaggedUnknown(t *testing.T) {
	// Finland falls back on the last Sunday of October: 04:00 EEST becomes
	// 03:00 EET, so local 03:00-03:59 happens twice. Which one the pilot meant
	// is genuinely unknowable, so it is flagged for review rather than guessed.
	_, origin, err := ToUTC("2024-10-27", "03:30")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	if origin != OriginUnknown {
		t.Errorf("origin = %q, want %q for an ambiguous time", origin, OriginUnknown)
	}
}

func TestToUTC_UnambiguousTimesNearATransitionAreNotFlagged(t *testing.T) {
	// Guard against an over-eager ambiguity check flagging the whole day.
	for _, c := range []struct{ date, raw string }{
		{"2024-03-31", "01:30"},
		{"2024-03-31", "12:00"},
		{"2024-10-27", "12:00"},
		{"2024-10-27", "23:45"},
	} {
		_, origin, err := ToUTC(c.date, c.raw)
		if err != nil {
			t.Fatalf("ToUTC(%q,%q) errored: %v", c.date, c.raw, err)
		}
		if origin != OriginConvertedFromLocal {
			t.Errorf("ToUTC(%q,%q) origin = %q, want %q", c.date, c.raw, origin, OriginConvertedFromLocal)
		}
	}
}

func TestToUTC_RejectsMalformedInput(t *testing.T) {
	cases := []struct{ name, date, raw string }{
		{"bad date", "not-a-date", "12:00"},
		{"bad time", "2024-08-20", "abc"},
		{"hour out of range", "2024-08-20", "25:00"},
		{"minute out of range", "2024-08-20", "12:75"},
		{"day out of range", "2024-02-31", "12:00"},
		{"no colon", "2024-08-20", "1200"},
		{"non-numeric minute", "2024-08-20", "12:ab"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, _, err := ToUTC(c.date, c.raw); err == nil {
				t.Errorf("ToUTC(%q,%q) succeeded; want an error", c.date, c.raw)
			}
		})
	}
}

func TestMustLoadLocation(t *testing.T) {
	// A missing zone database must stop the binary, never fall back to UTC:
	// a silent fallback would shift every converted row by two or three hours
	// while still looking entirely plausible.
	if got := mustLoadLocation("Europe/Helsinki"); got == nil {
		t.Fatal("mustLoadLocation returned nil for a valid zone")
	}
	defer func() {
		if recover() == nil {
			t.Error("mustLoadLocation did not panic on an unknown zone")
		}
	}()
	mustLoadLocation("Not/AZone")
}

func TestEmbeddedTzdataIsPresent(t *testing.T) {
	// Guards the blank import of time/tzdata. If someone removes it, this
	// fails on any machine without system zoneinfo -- including the container
	// the binary is cross-compiled for.
	if _, err := time.LoadLocation("Europe/Helsinki"); err != nil {
		t.Fatalf("Europe/Helsinki unavailable: %v", err)
	}
}

func TestBlockPair_NormalFlight(t *testing.T) {
	off, on, origin, err := BlockPair("2024-08-20", "18:50", "20:05")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	if want := time.Date(2024, 8, 20, 15, 50, 0, 0, time.UTC); !off.Equal(want) {
		t.Errorf("off = %v, want %v", off, want)
	}
	if want := time.Date(2024, 8, 20, 17, 5, 0, 0, time.UTC); !on.Equal(want) {
		t.Errorf("on = %v, want %v", on, want)
	}
	if origin != OriginConvertedFromLocal {
		t.Errorf("origin = %q", origin)
	}
}

func TestBlockPair_CrossingMidnightRollsTheDateForward(t *testing.T) {
	// A flight that lands after midnight UTC must not produce a negative
	// duration. On-block earlier than off-block means the next day.
	off, on, _, err := BlockPair("2024-08-20", "23:30", "00:45")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	if !on.After(off) {
		t.Fatalf("on-block %v is not after off-block %v", on, off)
	}
	if got := on.Sub(off); got != 75*time.Minute {
		t.Errorf("duration = %v, want 1h15m", got)
	}
}

func TestBlockPair_MixedZoneMarkersDegradeToUnknown(t *testing.T) {
	// A spread can mix local and UTC rows (confirmed on IMG_6014). If the two
	// halves of one pair disagree about their zone we cannot trust either, so
	// the row is flagged rather than silently half-converted.
	_, _, origin, err := BlockPair("2024-08-20", "18:50", "20:05Z")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	if origin != OriginUnknown {
		t.Errorf("origin = %q, want %q for a mixed-zone pair", origin, OriginUnknown)
	}
}

func TestBlockPair_BothZuluStaysAsWritten(t *testing.T) {
	off, on, origin, err := BlockPair("2024-08-20", "07:56Z", "09:10Z")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	if origin != OriginUTCAsWritten {
		t.Errorf("origin = %q, want %q", origin, OriginUTCAsWritten)
	}
	if got := on.Sub(off); got != 74*time.Minute {
		t.Errorf("duration = %v, want 1h14m", got)
	}
}

func TestToUTC_RejectsMalformedZuluTime(t *testing.T) {
	// The Z-suffix path parses its own clock and must be just as strict.
	for _, raw := range []string{"xx:30Z", "07:99Z", "0730Z"} {
		if _, _, err := ToUTC("2024-08-20", raw); err == nil {
			t.Errorf("ToUTC(%q) succeeded; want an error", raw)
		}
	}
}

func TestToUTC_RejectsNegativeClockFields(t *testing.T) {
	for _, raw := range []string{"-1:30", "12:-5"} {
		if _, _, err := ToUTC("2024-08-20", raw); err == nil {
			t.Errorf("ToUTC(%q) succeeded; want an error", raw)
		}
	}
}

func TestBlockPair_PropagatesParseErrors(t *testing.T) {
	// Either half being unreadable must fail the whole pair, not half-convert it.
	if _, _, _, err := BlockPair("2024-08-20", "nonsense", "20:05"); err == nil {
		t.Error("bad off-block succeeded; want an error")
	}
	if _, _, _, err := BlockPair("2024-08-20", "18:50", "nonsense"); err == nil {
		t.Error("bad on-block succeeded; want an error")
	}
}

func TestBlockPair_OneHalfBlankTrustsTheOtherHalf(t *testing.T) {
	// Some rows record only one side of the block pair. The origin should
	// reflect the half that actually carries a time, not degrade to unknown.
	cases := []struct {
		name, off, on string
		want          Origin
	}{
		{"only off-block, local", "18:50", "", OriginConvertedFromLocal},
		{"only on-block, local", "", "20:05", OriginConvertedFromLocal},
		{"only off-block, zulu", "15:50Z", "", OriginUTCAsWritten},
		{"only on-block, zulu", "", "17:05Z", OriginUTCAsWritten},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, origin, err := BlockPair("2024-08-20", c.off, c.on)
			if err != nil {
				t.Fatalf("errored: %v", err)
			}
			if origin != c.want {
				t.Errorf("origin = %q, want %q", origin, c.want)
			}
		})
	}
}

func TestBlockPair_AnAmbiguousHalfPoisonsThePair(t *testing.T) {
	// Off-block sits in the autumn fold, so the pair's duration is unreliable
	// even though the on-block reads cleanly.
	_, _, origin, err := BlockPair("2024-10-27", "03:30", "05:00")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	if origin != OriginUnknown {
		t.Errorf("origin = %q, want %q", origin, OriginUnknown)
	}
}

func TestBlockPair_EmptyPair(t *testing.T) {
	off, on, origin, err := BlockPair("2024-08-20", "", "")
	if err != nil {
		t.Fatalf("errored: %v", err)
	}
	if !off.IsZero() || !on.IsZero() {
		t.Errorf("expected zero times, got %v / %v", off, on)
	}
	if origin != OriginNone {
		t.Errorf("origin = %q, want %q", origin, OriginNone)
	}
}
