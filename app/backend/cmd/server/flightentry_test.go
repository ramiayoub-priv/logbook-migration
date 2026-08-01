package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/store"
)

// goodFlight is a submission that must be accepted. The date is fixed in the
// past relative to the harness clock, which is pinned so that the suite does
// not start failing on a future date the day the calendar catches up.
const goodFlight = `{
	"date": "2026-07-30",
	"aircraft_reg": "OH-CTL",
	"aircraft_type": "C172",
	"class": "SEP_SEA",
	"dep_place": "EFHF",
	"arr_place": "EFHF",
	"off_block": "09:15Z",
	"on_block": "10:30Z",
	"total_time": "1:15",
	"pic_time": "1:15",
	"pic_name": "self",
	"landings_day": 3
}`

func TestCreatingAFlightStoresItAndReturnsIt(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/flights", goodFlight, auth)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d, want 201; body %s", w.Code, w.Body.String())
	}

	var got struct {
		Flight struct {
			Seq          int    `json:"seq"`
			Date         string `json:"date"`
			AircraftReg  string `json:"aircraft_reg"`
			TotalMinutes int    `json:"total_minutes"`
			SourceBook   int    `json:"source_book"`
			Verified     bool   `json:"landings_verified"`
		} `json:"flight"`
	}
	decodeStatus(t, w, http.StatusCreated, &got)

	if got.Flight.TotalMinutes != 75 {
		t.Errorf("total_minutes = %d, want 75 (the H:MM was not parsed to minutes)", got.Flight.TotalMinutes)
	}
	if got.Flight.Seq < store.HandEnteredSeqBase {
		t.Errorf("seq = %d, want it allocated from the hand-entered band at %d",
			got.Flight.Seq, store.HandEnteredSeqBase)
	}
	if got.Flight.SourceBook != 0 {
		t.Errorf("source_book = %d, want 0 so the next CSV import does not delete it", got.Flight.SourceBook)
	}
	if !got.Flight.Verified {
		t.Error("landings_verified = false; the pilot typed the split, so it is not an inference")
	}
}

func TestACreatedFlightIsVisibleAndCounted(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("POST", "/logbook/api/flights", goodFlight, auth); w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}

	var list struct {
		Count int `json:"count"`
	}
	decode(t, h.do("GET", "/logbook/api/flights", "", auth), &list)
	if list.Count != 3 {
		t.Errorf("flight count = %d, want 3", list.Count)
	}

	// The statistics are computed from the rows on every request, so the new
	// flight has to move them -- that is the point of entering it.
	var st struct {
		Summary struct {
			Total    int `json:"total"`
			PIC      int `json:"pic"`
			SeaTotal int `json:"sea_total"`
		} `json:"summary"`
	}
	decode(t, h.do("GET", "/logbook/api/stats", "", auth), &st)
	if st.Summary.Total != 171+75 {
		t.Errorf("total = %d, want %d", st.Summary.Total, 171+75)
	}
	if st.Summary.PIC != 81+75 {
		t.Errorf("pic = %d, want %d", st.Summary.PIC, 81+75)
	}
	if st.Summary.SeaTotal != 81+75 {
		t.Errorf("sea_total = %d, want %d; SEP_SEA did not reach the seaplane column", st.Summary.SeaTotal, 81+75)
	}
}

// A rejected submission has to say which field to fix. A 400 with prose in it
// leaves the pilot guessing on a twenty-field form.
func TestARejectedFlightNamesTheFields(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/flights",
		`{"date":"","aircraft_reg":"","aircraft_type":"C172","class":"SEP_LAND",
		  "off_block":"09:15Z","on_block":"10:30Z","total_time":"","landings_day":1}`, auth)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400; body %s", w.Code, w.Body.String())
	}

	var got struct {
		Error  string `json:"error"`
		Fields []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"fields"`
	}
	decodeStatus(t, w, http.StatusBadRequest, &got)

	if len(got.Fields) < 3 {
		t.Fatalf("got %d field errors, want at least date, aircraft_reg and total_time: %s",
			len(got.Fields), w.Body.String())
	}
	seen := map[string]bool{}
	for _, f := range got.Fields {
		seen[f.Field] = true
		if f.Message == "" {
			t.Errorf("field %q came back with no message", f.Field)
		}
	}
	for _, want := range []string{"date", "aircraft_reg", "total_time"} {
		if !seen[want] {
			t.Errorf("no error reported for %q", want)
		}
	}
}

func TestNothingIsStoredWhenAFlightIsRejected(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	h.do("POST", "/logbook/api/flights", `{"date":"","total_time":""}`, auth)

	var list struct {
		Count int `json:"count"`
	}
	decode(t, h.do("GET", "/logbook/api/flights", "", auth), &list)
	if list.Count != 2 {
		t.Errorf("flight count = %d, want the original 2", list.Count)
	}
}

// The double-tapped submit button. Two identical flights in a legal record
// inflate a licence total.
func TestASubmittedFlightCannotBeSubmittedTwice(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("POST", "/logbook/api/flights", goodFlight, auth); w.Code != http.StatusCreated {
		t.Fatalf("first create: %d", w.Code)
	}
	w := h.do("POST", "/logbook/api/flights", goodFlight, auth)
	if w.Code != http.StatusConflict {
		t.Fatalf("second create: status %d, want 409; body %s", w.Code, w.Body.String())
	}

	var list struct {
		Count int `json:"count"`
	}
	decode(t, h.do("GET", "/logbook/api/flights", "", auth), &list)
	if list.Count != 3 {
		t.Errorf("flight count = %d, want 3; the duplicate was stored", list.Count)
	}
}

// DisallowUnknownFields is on for a reason: a field the server does not
// understand means the client and the server disagree about what is being
// written, and guessing which is right is not acceptable on a legal record.
func TestAnUnknownFieldIsRejectedRatherThanIgnored(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/flights",
		`{"date":"2026-07-30","aircraft_reg":"OH-CTL","aircraft_type":"C172",
		  "class":"SEP_LAND","off_block":"09:15Z","on_block":"10:30Z",
		  "total_time":"1:15","landings_day":1,"cumulative_total":"999:00"}`, auth)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400 for an unknown field", w.Code)
	}
}

// Writing to the logbook is the highest-value action in the app, so the CSRF
// control matters more here than anywhere else.
func TestCreatingAFlightRequiresAnOrigin(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/flights", goodFlight, auth, func(r *http.Request) {
		r.Header.Del("Origin")
	})
	if w.Code != http.StatusForbidden {
		t.Errorf("status %d, want 403 with no Origin header", w.Code)
	}

	w = h.do("POST", "/logbook/api/flights", goodFlight, auth, func(r *http.Request) {
		r.Header.Set("Origin", "https://evil.example")
	})
	if w.Code != http.StatusForbidden {
		t.Errorf("status %d, want 403 from a foreign origin", w.Code)
	}
}

// An ambiguous local time is refused rather than stored as a guess, and the
// refusal has to be actionable.
func TestAnAmbiguousTimeIsRefusedWithAWayOut(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/flights", `{
		"date": "2025-10-26", "aircraft_reg": "OH-CTL", "aircraft_type": "C172",
		"class": "SEP_LAND", "off_block": "03:30", "on_block": "04:45",
		"total_time": "1:15", "landings_day": 1}`, auth)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400 for a time inside the autumn fold; body %s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); !strings.Contains(body, "Z") {
		t.Errorf("the refusal %s does not tell the pilot to write UTC", body)
	}
}
