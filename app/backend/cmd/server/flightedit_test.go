package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// The edit and delete endpoints, added 2026-08-02.
//
// The rule these exist to enforce over HTTP is that only a HAND-ENTERED flight
// can be touched. The 1296 imported rows are closed data (CLAUDE.md rule 0.8)
// and the importer owns them; an edit to one would be discarded at the next
// re-import even if it were allowed.

// createOne posts goodFlight and returns the seq it was given.
func createOne(t *testing.T, h *harness, auth func(*http.Request)) int {
	t.Helper()
	w := h.do("POST", "/logbook/api/flights", goodFlight, auth)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var got struct {
		Flight struct {
			Seq int `json:"seq"`
		} `json:"flight"`
	}
	decodeStatus(t, w, http.StatusCreated, &got)
	return got.Flight.Seq
}

// editedFlight is goodFlight with a corrected on-block time and total: the
// ordinary case, a mistyped clock noticed after saving.
const editedFlight = `{
	"date": "2026-07-30",
	"aircraft_reg": "OH-CTL",
	"aircraft_type": "C172",
	"class": "SEP_SEA",
	"dep_place": "EFHF",
	"arr_place": "EFNU",
	"off_block": "09:15Z",
	"on_block": "10:45Z",
	"total_time": "1:30",
	"pic_time": "1:30",
	"pic_name": "self",
	"landings_day": 2,
	"remarks": "corrected"
}`

// The edit page is reachable by URL, so it must be able to load one flight
// without pulling all 1296 down to find it.
func TestOneFlightCanBeReadByItsNumber(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	seq := createOne(t, h, auth)

	w := h.do("GET", fmt.Sprintf("/logbook/api/flights/%d", seq), "", auth)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200; body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"aircraft_reg":"OH-CTL"`) {
		t.Errorf("body = %s", w.Body.String())
	}

	// An imported flight reads back too -- the page shows it and explains that
	// it cannot be changed, which is better than a 404 the pilot cannot place.
	if w := h.do("GET", "/logbook/api/flights/1", "", auth); w.Code != http.StatusOK {
		t.Errorf("reading an imported flight: %d, want 200", w.Code)
	}
	if w := h.do("GET", "/logbook/api/flights/9999999", "", auth); w.Code != http.StatusNotFound {
		t.Errorf("reading a missing flight: %d, want 404", w.Code)
	}
}

// The airborne times have to travel back to the client, or the edit form
// cannot show them -- and a form that cannot show a field it submits ERASES it
// on the next save. That is the silent corruption rule 0.2 forbids, and it was
// only visible once there was something reading a flight back to edit it.
func TestAFlightCarriesItsAirborneTimesBackToTheClient(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	withAirborne := strings.Replace(goodFlight,
		`"off_block": "09:15Z",`,
		`"off_block": "09:15Z", "takeoff": "09:20Z", "landing": "10:25Z",`, 1)
	w := h.do("POST", "/logbook/api/flights", withAirborne, auth)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Flight struct {
			Seq        int     `json:"seq"`
			TakeoffUTC *string `json:"takeoff_utc"`
			LandingUTC *string `json:"landing_utc"`
		} `json:"flight"`
	}
	decodeStatus(t, w, http.StatusCreated, &created)
	if created.Flight.TakeoffUTC == nil || created.Flight.LandingUTC == nil {
		t.Fatalf("the create response dropped the airborne times: %s", w.Body.String())
	}

	read := h.do("GET", fmt.Sprintf("/logbook/api/flights/%d", created.Flight.Seq), "", auth)
	if !strings.Contains(read.Body.String(), `"takeoff_utc":"2026-07-30T09:20:00Z"`) {
		t.Errorf("reading the flight back lost the takeoff time: %s", read.Body.String())
	}

	// A flight with no airborne times still reports them as null, not as a
	// zero instant that reads as year 1.
	plain := h.do("GET", "/logbook/api/flights/1", "", auth).Body.String()
	if !strings.Contains(plain, `"takeoff_utc":null`) {
		t.Errorf("a blank airborne time is not null: %s", plain)
	}
}

func TestEditingAFlightStoresTheCorrection(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	seq := createOne(t, h, auth)

	w := h.do("PUT", fmt.Sprintf("/logbook/api/flights/%d", seq), editedFlight, auth)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200; body %s", w.Code, w.Body.String())
	}

	var got struct {
		Flight struct {
			Seq          int    `json:"seq"`
			ArrPlace     string `json:"arr_place"`
			TotalMinutes int    `json:"total_minutes"`
			Remarks      string `json:"remarks"`
		} `json:"flight"`
	}
	decodeStatus(t, w, http.StatusOK, &got)

	if got.Flight.Seq != seq {
		t.Errorf("seq moved on an edit: %d -> %d", seq, got.Flight.Seq)
	}
	if got.Flight.TotalMinutes != 90 || got.Flight.ArrPlace != "EFNU" || got.Flight.Remarks != "corrected" {
		t.Errorf("the correction did not land: %+v", got.Flight)
	}
}

// The edited flight must be what the list and the totals report, not just what
// the write returned.
func TestAnEditedFlightIsWhatTheLogbookReports(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	seq := createOne(t, h, auth)

	if w := h.do("PUT", fmt.Sprintf("/logbook/api/flights/%d", seq), editedFlight, auth); w.Code != http.StatusOK {
		t.Fatalf("edit: %d %s", w.Code, w.Body.String())
	}

	w := h.do("GET", "/logbook/api/flights", "", auth)
	body := w.Body.String()
	if !strings.Contains(body, `"remarks":"corrected"`) {
		t.Error("the flight list does not show the correction")
	}
	if strings.Contains(body, `"total_minutes":75`) {
		t.Error("the flight list still shows the pre-edit total")
	}
}

func TestEditingIsRefusedOnAFlightFromThePaperBooks(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	// seq 1 is the first row of the imported books.
	w := h.do("PUT", "/logbook/api/flights/1", editedFlight, auth)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status %d, want 403; body %s", w.Code, w.Body.String())
	}
	// The message has to say why, or the pilot reads it as a bug.
	if !strings.Contains(strings.ToLower(w.Body.String()), "paper") {
		t.Errorf("the refusal does not explain itself: %s", w.Body.String())
	}
}

func TestDeletingIsRefusedOnAFlightFromThePaperBooks(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("DELETE", "/logbook/api/flights/1", "", auth); w.Code != http.StatusForbidden {
		t.Fatalf("status %d, want 403; body %s", w.Code, w.Body.String())
	}
	// And it is still there.
	if w := h.do("GET", "/logbook/api/flights", "", auth); !strings.Contains(w.Body.String(), `"seq":1,`) {
		t.Error("an imported flight disappeared")
	}
}

func TestEditingAndDeletingReportAMissingFlight(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("PUT", "/logbook/api/flights/9999999", editedFlight, auth); w.Code != http.StatusNotFound {
		t.Errorf("PUT on a missing flight: %d, want 404", w.Code)
	}
	if w := h.do("DELETE", "/logbook/api/flights/9999999", "", auth); w.Code != http.StatusNotFound {
		t.Errorf("DELETE on a missing flight: %d, want 404", w.Code)
	}
}

// A seq that is not a number is a 404, never a 500 and never a query.
func TestAnUnreadableFlightNumberIsRefused(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("PUT", "/logbook/api/flights/not-a-number", editedFlight, auth); w.Code != http.StatusNotFound {
		t.Errorf("PUT with a non-numeric seq: %d, want 404", w.Code)
	}
}

// An edit is validated exactly as a new flight is: the same rules decide what
// may be written, whichever door it comes through.
func TestAnEditIsValidatedLikeANewFlight(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	seq := createOne(t, h, auth)

	bad := `{
		"date": "2026-07-30",
		"aircraft_reg": "OH-CTL",
		"aircraft_type": "C172",
		"class": "SEP_SEA",
		"off_block": "09:15Z",
		"on_block": "10:30Z",
		"total_time": "1:15",
		"pic_time": "9:00",
		"pic_name": "self",
		"landings_day": 1
	}`
	w := h.do("PUT", fmt.Sprintf("/logbook/api/flights/%d", seq), bad, auth)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400; body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"field":"pic_time"`) {
		t.Errorf("the field error does not name the control: %s", w.Body.String())
	}
}

func TestAnEditOntoAnotherFlightIsADuplicate(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	first := createOne(t, h, auth)

	// A second, different flight the same day.
	second := strings.Replace(goodFlight, `"off_block": "09:15Z"`, `"off_block": "14:00Z"`, 1)
	second = strings.Replace(second, `"on_block": "10:30Z"`, `"on_block": "15:15Z"`, 1)
	w := h.do("POST", "/logbook/api/flights", second, auth)
	if w.Code != http.StatusCreated {
		t.Fatalf("second create: %d %s", w.Code, w.Body.String())
	}
	var got struct {
		Flight struct {
			Seq int `json:"seq"`
		} `json:"flight"`
	}
	decodeStatus(t, w, http.StatusCreated, &got)

	// Editing the second one onto the first one's key.
	if w := h.do("PUT", fmt.Sprintf("/logbook/api/flights/%d", got.Flight.Seq), goodFlight, auth); w.Code != http.StatusConflict {
		t.Errorf("status %d, want 409; body %s", w.Code, w.Body.String())
	}
	_ = first
}

func TestDeletingAFlightRemovesItFromTheLogbook(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	seq := createOne(t, h, auth)

	w := h.do("DELETE", fmt.Sprintf("/logbook/api/flights/%d", seq), "", auth)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200; body %s", w.Code, w.Body.String())
	}
	// What was deleted comes back, so the page can name the flight it removed.
	if !strings.Contains(w.Body.String(), `"aircraft_reg":"OH-CTL"`) {
		t.Errorf("the delete response does not say what was removed: %s", w.Body.String())
	}

	list := h.do("GET", "/logbook/api/flights", "", auth).Body.String()
	if strings.Contains(list, fmt.Sprintf(`"seq":%d,`, seq)) {
		t.Error("the deleted flight is still in the logbook")
	}
}

// Deleting the same flight twice is a 404, not a second success. A phone with
// a slow connection double-taps.
func TestDeletingTwiceIsNotFound(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	seq := createOne(t, h, auth)

	if w := h.do("DELETE", fmt.Sprintf("/logbook/api/flights/%d", seq), "", auth); w.Code != http.StatusOK {
		t.Fatalf("first delete: %d", w.Code)
	}
	if w := h.do("DELETE", fmt.Sprintf("/logbook/api/flights/%d", seq), "", auth); w.Code != http.StatusNotFound {
		t.Errorf("second delete: %d, want 404", w.Code)
	}
}
