package entry_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/entry"
	"github.com/ramiayoub/logbook/backend/internal/timeutil"
)

// now is the instant every test validates against. Fixed, so "not in the
// future" is a property of the input rather than of the day the suite runs.
var now = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

// valid is a draft that must pass. Every negative test below is this value with
// exactly one field broken, so a test that fails proves the field it names and
// nothing else.
func valid() entry.Draft {
	return entry.Draft{
		Date:         "2026-07-30",
		AircraftReg:  "OH-CAM",
		AircraftType: "C172",
		Class:        "SEP_LAND",
		DepPlace:     "EFHF",
		ArrPlace:     "EFHF",
		OffBlock:     "09:15Z",
		OnBlock:      "10:30Z",
		TotalTime:    "1:15",
		PICTime:      "1:15",
		PICName:      "SELF",
		LandingsDay:  3,
	}
}

func mustValidate(t *testing.T, d entry.Draft) csvbook.Flight {
	t.Helper()
	f, err := entry.Validate(d, now)
	if err != nil {
		t.Fatalf("Validate() returned %v, want no error", err)
	}
	return f
}

// fieldsInError returns the field names a rejection names.
func fieldsInError(t *testing.T, err error) []string {
	t.Helper()
	var errs entry.Errors
	if !errors.As(err, &errs) {
		t.Fatalf("error %v is not entry.Errors; the form cannot highlight a field it is not told about", err)
	}
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		out = append(out, e.Field)
	}
	return out
}

func assertRejects(t *testing.T, d entry.Draft, field string) {
	t.Helper()
	_, err := entry.Validate(d, now)
	if err == nil {
		t.Fatalf("Validate() accepted a draft with a bad %s; on a legal record that is a silently corrupt row", field)
	}
	got := fieldsInError(t, err)
	for _, g := range got {
		if g == field {
			return
		}
	}
	t.Fatalf("Validate() rejected the draft but blamed %v, want the error to name %q", got, field)
}

func TestValidateAcceptsAGoodDraft(t *testing.T) {
	f := mustValidate(t, valid())

	if f.Date != "2026-07-30" {
		t.Errorf("Date = %q, want 2026-07-30", f.Date)
	}
	if f.TotalMinutes != 75 {
		t.Errorf("TotalMinutes = %d, want 75", f.TotalMinutes)
	}
	if f.PICMinutes != 75 {
		t.Errorf("PICMinutes = %d, want 75", f.PICMinutes)
	}
	if f.LandingsDay != 3 || f.LandingsNight != 0 {
		t.Errorf("landings = %d day / %d night, want 3/0", f.LandingsDay, f.LandingsNight)
	}
}

// A hand-entered row is the one case where the day/night landing split was
// typed by the pilot rather than inferred, so it is verified by construction.
// Task 8 exists because the imported rows are not.
func TestHandEnteredLandingsAreVerified(t *testing.T) {
	f := mustValidate(t, valid())
	if !f.LandingsVerified {
		t.Error("LandingsVerified = false; a split the pilot typed is not an inference and must not appear in the review list")
	}
}

// source_book 0 is what distinguishes a row entered in the app from a row
// transcribed off paper. The importer keys on it to know which rows it owns,
// so if this ever stopped being 0 the next CSV import would delete the row.
func TestHandEnteredRowsAreMarkedAsSuch(t *testing.T) {
	f := mustValidate(t, valid())
	if f.SourceBook != 0 {
		t.Errorf("SourceBook = %d, want 0 to mark a row that came from the app and not from a paper book", f.SourceBook)
	}
}

func TestValidateConvertsZuluTimes(t *testing.T) {
	f := mustValidate(t, valid())

	wantOff := time.Date(2026, 7, 30, 9, 15, 0, 0, time.UTC)
	if !f.OffBlockUTC.Equal(wantOff) {
		t.Errorf("OffBlockUTC = %v, want %v", f.OffBlockUTC, wantOff)
	}
	if f.TimeOrigin != timeutil.OriginUTCAsWritten {
		t.Errorf("TimeOrigin = %q, want %q", f.TimeOrigin, timeutil.OriginUTCAsWritten)
	}
	// The raw strings survive exactly as typed (rule 0.4): a conversion that
	// cannot be audited is a conversion that cannot be corrected.
	if f.OffBlockRaw != "09:15Z" || f.OnBlockRaw != "10:30Z" {
		t.Errorf("raw times = %q/%q, want them kept exactly as written", f.OffBlockRaw, f.OnBlockRaw)
	}
}

func TestValidateConvertsLocalTimes(t *testing.T) {
	d := valid()
	d.OffBlock = "12:15" // Helsinki summer time is UTC+3
	d.OnBlock = "13:30"
	f := mustValidate(t, d)

	wantOff := time.Date(2026, 7, 30, 9, 15, 0, 0, time.UTC)
	if !f.OffBlockUTC.Equal(wantOff) {
		t.Errorf("OffBlockUTC = %v, want %v (Helsinki 12:15 in July is 09:15Z)", f.OffBlockUTC, wantOff)
	}
	if f.TimeOrigin != timeutil.OriginConvertedFromLocal {
		t.Errorf("TimeOrigin = %q, want %q", f.TimeOrigin, timeutil.OriginConvertedFromLocal)
	}
}

// A flight that lands after midnight. timeutil rolls the on-block date; the
// duration must come out positive rather than negative.
func TestValidateHandlesTheMidnightRoll(t *testing.T) {
	d := valid()
	d.OffBlock = "23:30Z"
	d.OnBlock = "00:45Z"
	f := mustValidate(t, d)

	if f.BlockMinutes != 75 {
		t.Errorf("BlockMinutes = %d, want 75; the on-block did not roll to the next day", f.BlockMinutes)
	}
}

// The whole point of refusing here rather than storing OriginUnknown: on an
// imported row nobody is around to ask, but a hand-entered row has the pilot
// right there. Storing a guess when the answer is available is the silent
// corruption rule 0.2 forbids.
func TestValidateRefusesAnAmbiguousLocalTime(t *testing.T) {
	d := valid()
	// 2025-10-26 03:30 Helsinki happens twice: the autumn fold. It is in the
	// past relative to `now`, because a future date is refused for its own
	// reason and would prove nothing about the fold.
	d.Date = "2025-10-26"
	d.OffBlock = "03:30"
	d.OnBlock = "04:45"

	_, err := entry.Validate(d, now)
	if err == nil {
		t.Fatal("Validate() accepted an ambiguous local time; it must ask for a Zulu time instead of guessing")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "z") {
		t.Errorf("error %q does not tell the pilot to write the time as UTC; an unactionable refusal is a bug report", err)
	}
}

func TestValidateAcceptsTheSameInstantWrittenAsZulu(t *testing.T) {
	d := valid()
	d.Date = "2025-10-26"
	d.OffBlock = "03:30Z"
	d.OnBlock = "04:45Z"
	mustValidate(t, d) // the escape hatch the refusal above points at must work
}

func TestValidateRejectsBadFields(t *testing.T) {
	for _, tc := range []struct {
		name   string
		field  string
		break_ func(*entry.Draft)
	}{
		{"empty date", "date", func(d *entry.Draft) { d.Date = "" }},
		{"malformed date", "date", func(d *entry.Draft) { d.Date = "30/07/2026" }},
		{"impossible date", "date", func(d *entry.Draft) { d.Date = "2026-02-30" }},
		{"future date", "date", func(d *entry.Draft) { d.Date = "2026-08-02" }},

		{"no registration", "aircraft_reg", func(d *entry.Draft) { d.AircraftReg = "  " }},
		{"no type", "aircraft_type", func(d *entry.Draft) { d.AircraftType = "" }},
		{"unknown class", "class", func(d *entry.Draft) { d.Class = "SEP_AMPHIB" }},
		{"empty class", "class", func(d *entry.Draft) { d.Class = "" }},

		{"no off block", "off_block", func(d *entry.Draft) { d.OffBlock = "" }},
		{"no on block", "on_block", func(d *entry.Draft) { d.OnBlock = "" }},
		{"unparseable off block", "off_block", func(d *entry.Draft) { d.OffBlock = "0915" }},
		// Blamed on the on-block specifically: the pair is parsed as a unit, so
		// a fault in the second half must not send the pilot to the first.
		{"unparseable on block", "on_block", func(d *entry.Draft) { d.OnBlock = "10.30" }},
		{"hour out of range", "off_block", func(d *entry.Draft) { d.OffBlock = "25:00" }},
		{"minute out of range", "on_block", func(d *entry.Draft) { d.OnBlock = "10:75" }},

		{"no total", "total_time", func(d *entry.Draft) { d.TotalTime = "" }},
		{"zero total", "total_time", func(d *entry.Draft) { d.TotalTime = "0:00" }},
		{"unparseable total", "total_time", func(d *entry.Draft) { d.TotalTime = "1.25" }},
		{"absurd total", "total_time", func(d *entry.Draft) { d.TotalTime = "25:00" }},

		{"negative day landings", "landings_day", func(d *entry.Draft) { d.LandingsDay = -1 }},
		{"negative night landings", "landings_night", func(d *entry.Draft) { d.LandingsNight = -2 }},
		{"no landings at all", "landings_day", func(d *entry.Draft) { d.LandingsDay = 0; d.LandingsNight = 0 }},

		{"pic over total", "pic_time", func(d *entry.Draft) { d.PICTime = "2:00" }},
		{"night over total", "night_time", func(d *entry.Draft) { d.NightTime = "9:99" }},
		{"instrument over total", "instrument_time", func(d *entry.Draft) { d.InstrumentTime = "1:30" }},
		{"dual over total", "dual_time", func(d *entry.Draft) { d.DualTime = "1:16" }},
		{"instructor over total", "instructor_time", func(d *entry.Draft) { d.InstructorTime = "3:00" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := valid()
			tc.break_(&d)
			assertRejects(t, d, tc.field)
		})
	}
}

// Every problem at once, not the first one. A form that reveals one error per
// submission on a twenty-field flight entry is a form nobody finishes.
func TestValidateReportsEveryProblemAtOnce(t *testing.T) {
	d := valid()
	d.Date = ""
	d.AircraftReg = ""
	d.TotalTime = ""

	_, err := entry.Validate(d, now)
	if err == nil {
		t.Fatal("Validate() accepted a draft with three broken fields")
	}
	got := fieldsInError(t, err)
	if len(got) < 3 {
		t.Errorf("Validate() reported %v, want all three of date, aircraft_reg and total_time", got)
	}
}

// A registration typed on a phone arrives however the keyboard felt. It is
// stored uppercase so that "oh-cam" and "OH-CAM" are the same aircraft rather
// than two rows in the statistics.
func TestValidateNormalizesInput(t *testing.T) {
	d := valid()
	d.AircraftReg = "  oh-cam "
	d.AircraftType = " c172 "
	d.DepPlace = "efhf"
	d.ArrPlace = " efhk"
	f := mustValidate(t, d)

	if f.AircraftReg != "OH-CAM" {
		t.Errorf("AircraftReg = %q, want OH-CAM", f.AircraftReg)
	}
	if f.AircraftType != "C172" {
		t.Errorf("AircraftType = %q, want C172", f.AircraftType)
	}
	if f.DepPlace != "EFHF" || f.ArrPlace != "EFHK" {
		t.Errorf("places = %q/%q, want EFHF/EFHK", f.DepPlace, f.ArrPlace)
	}
}

// The components are allowed to equal the total -- a 1:15 flight logged 1:15
// PIC is the ordinary case, and an off-by-one here would reject nearly every
// real flight.
func TestValidateAllowsAComponentEqualToTheTotal(t *testing.T) {
	d := valid()
	d.PICTime = "1:15"
	d.NightTime = "1:15"
	d.InstrumentTime = "1:15"
	mustValidate(t, d)
}

// Optional durations left blank are zero, not an error. Most flights log
// neither night nor instrument time.
func TestValidateTreatsBlankDurationsAsZero(t *testing.T) {
	f := mustValidate(t, valid())
	if f.NightMinutes != 0 || f.InstrumentMinutes != 0 || f.InstructorMinutes != 0 {
		t.Errorf("blank durations became %d/%d/%d, want zeros",
			f.NightMinutes, f.InstrumentMinutes, f.InstructorMinutes)
	}
}

// Block time is chocks-to-chocks and Total is what the book totals on. They are
// usually equal but genuinely differ on one row of the paper books, so a
// mismatch is recorded rather than rejected -- while a total the pilot did not
// type is derived from the clock rather than left at zero.
func TestBlockTimeComesFromTheClock(t *testing.T) {
	d := valid()
	d.TotalTime = "1:10" // five minutes of taxi not counted as flight time
	d.PICTime = "1:10"   // the components follow the total, not the clock
	f := mustValidate(t, d)

	if f.BlockMinutes != 75 {
		t.Errorf("BlockMinutes = %d, want 75 from the off/on-block clock", f.BlockMinutes)
	}
	if f.TotalMinutes != 70 {
		t.Errorf("TotalMinutes = %d, want the 70 the pilot typed", f.TotalMinutes)
	}
}

func TestErrorsMessageNamesTheFields(t *testing.T) {
	d := valid()
	d.Date = ""
	_, err := entry.Validate(d, now)
	if err == nil {
		t.Fatal("expected a rejection")
	}
	if !strings.Contains(err.Error(), "date") {
		t.Errorf("error text %q does not mention the field at fault", err.Error())
	}
}

// --- Airborne times -------------------------------------------------------
//
// Takeoff and landing are OPTIONAL: most rows in the paper books have none,
// and a form that demanded them would block the common case to serve the rare
// one. When they are given they are converted by the same authority as the
// block pair, and the airborne duration is derived rather than typed.

func TestAirborneTimesAreOptional(t *testing.T) {
	f := mustValidate(t, valid()) // valid() carries neither
	if !f.TakeoffUTC.IsZero() || !f.LandingUTC.IsZero() {
		t.Errorf("a draft with no airborne times produced %v/%v, want both zero",
			f.TakeoffUTC, f.LandingUTC)
	}
}

func TestValidateConvertsAirborneTimes(t *testing.T) {
	d := valid()
	d.Takeoff, d.Landing = "09:20Z", "10:25Z"

	f := mustValidate(t, d)
	if got, want := f.TakeoffUTC, time.Date(2026, 7, 30, 9, 20, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("takeoff = %v, want %v", got, want)
	}
	if got, want := f.LandingUTC, time.Date(2026, 7, 30, 10, 25, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("landing = %v, want %v", got, want)
	}
}

// Half a pair is not a time, it is a typo. Saying so beats storing a takeoff
// with no landing, which would silently make the airborne duration unknowable.
func TestValidateRefusesHalfAnAirbornePair(t *testing.T) {
	for _, c := range []struct{ takeoff, landing, field string }{
		{"09:20Z", "", "landing"},
		{"", "10:25Z", "takeoff"},
	} {
		d := valid()
		d.Takeoff, d.Landing = c.takeoff, c.landing
		_, err := entry.Validate(d, now)
		if err == nil {
			t.Fatalf("takeoff=%q landing=%q was accepted; want a rejection naming %s",
				c.takeoff, c.landing, c.field)
		}
		if got := fieldsInError(t, err); len(got) != 1 || got[0] != c.field {
			t.Errorf("takeoff=%q landing=%q named %v, want exactly [%s]",
				c.takeoff, c.landing, got, c.field)
		}
	}
}

func TestValidateRejectsAnUnparseableAirborneTime(t *testing.T) {
	d := valid()
	d.Takeoff, d.Landing = "0920", "10:25Z"
	_, err := entry.Validate(d, now)
	if err == nil {
		t.Fatal("takeoff \"0920\" was accepted, want a rejection")
	}
	if got := fieldsInError(t, err); len(got) != 1 || got[0] != "takeoff" {
		t.Errorf("named %v, want exactly [takeoff]", got)
	}
}

// An aeroplane cannot be airborne longer than it is off blocks. This catches
// the realistic typo -- a landing time later than the on-block time -- which
// would otherwise be stored as a flight whose parts contradict each other.
func TestValidateRefusesAirborneTimeLongerThanBlockTime(t *testing.T) {
	d := valid() // 09:15Z -> 10:30Z, so 1:15 of block
	d.Takeoff, d.Landing = "09:10Z", "10:35Z"

	_, err := entry.Validate(d, now)
	if err == nil {
		t.Fatal("an 1:25 airborne time inside 1:15 of block was accepted, want a rejection")
	}
	if got := fieldsInError(t, err); len(got) != 1 || got[0] != "takeoff" {
		t.Errorf("named %v, want exactly [takeoff]", got)
	}
}

// Equal is fine: no taxi at all is unusual, not impossible, and refusing it
// would be inventing a rule the paper books do not keep.
func TestValidateAllowsAirborneTimeEqualToBlockTime(t *testing.T) {
	d := valid()
	d.Takeoff, d.Landing = "09:15Z", "10:30Z"
	mustValidate(t, d)
}

func TestAirborneTimesRollPastMidnight(t *testing.T) {
	d := valid()
	d.OffBlock, d.OnBlock = "23:30Z", "00:40Z"
	d.Takeoff, d.Landing = "23:35Z", "00:35Z"
	d.TotalTime, d.PICTime = "1:10", "1:10"

	f := mustValidate(t, d)
	if got := f.LandingUTC.Sub(f.TakeoffUTC); got != time.Hour {
		t.Errorf("airborne = %v, want 1h -- the landing must roll to the next day", got)
	}
}

// The block pair has this test too: the message must send the pilot to the
// half that is actually wrong, or the form highlights a field that is fine.
func TestAnUnparseableLandingNamesTheLandingField(t *testing.T) {
	d := valid()
	d.Takeoff, d.Landing = "09:20Z", "quarter past"

	_, err := entry.Validate(d, now)
	if err == nil {
		t.Fatal("landing \"quarter past\" was accepted, want a rejection")
	}
	if got := fieldsInError(t, err); len(got) != 1 || got[0] != "landing" {
		t.Errorf("named %v, want exactly [landing]", got)
	}
}

// Same posture as the block pair on a DST fold: refuse and ask for Zulu rather
// than manufacture an unaudited instant (rule 0.4).
func TestValidateRefusesAnAmbiguousLocalAirborneTime(t *testing.T) {
	d := valid()
	d.Date = "2025-10-26" // the autumn fold; 03:30 Helsinki happens twice
	d.OffBlock, d.OnBlock = "03:25Z", "04:45Z"
	d.Takeoff, d.Landing = "03:30", "04:40"

	_, err := entry.Validate(d, now)
	if err == nil {
		t.Fatal("an ambiguous local takeoff was accepted; it must ask for a Zulu time")
	}
	if got := fieldsInError(t, err); len(got) != 1 || got[0] != "takeoff" {
		t.Errorf("named %v, want exactly [takeoff]", got)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "z") {
		t.Errorf("error %q does not tell the pilot to write the time as UTC", err)
	}
}
