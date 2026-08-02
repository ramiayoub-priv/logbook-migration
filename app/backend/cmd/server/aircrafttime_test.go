package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

// GET /aircraft-time is the endpoint behind the aircraft-time page (Task 13).
// The owner pays for aeroplanes by the hour and some owners charge block time
// while some charge air time, so this is money -- the aggregation itself lives
// in internal/stats at 100%, and what is tested here is the contract: what
// reaches the client, what the range does, and that the air figure never
// travels without the coverage that makes it readable.

// aircraftTimeResponse mirrors what the handler sends.
type aircraftTimeResponse struct {
	Range    map[string]string `json:"range"`
	Reg      string            `json:"reg"`
	Aircraft []struct {
		Registration          string   `json:"registration"`
		Types                 []string `json:"types"`
		Flights               int      `json:"flights"`
		BlockMinutes          int      `json:"block_minutes"`
		AirMinutes            int      `json:"air_minutes"`
		AirKnown              int      `json:"air_known"`
		AirMissing            int      `json:"air_missing"`
		BlockDiffersFromTotal int      `json:"block_differs_from_total"`
	} `json:"aircraft"`
	Total struct {
		Flights      int `json:"flights"`
		BlockMinutes int `json:"block_minutes"`
		AirMinutes   int `json:"air_minutes"`
		AirKnown     int `json:"air_known"`
		AirMissing   int `json:"air_missing"`
	} `json:"total"`
	Flights []flightJSON `json:"flights"`
}

func (h *harness) aircraftTime(target string, auth func(*http.Request)) aircraftTimeResponse {
	h.t.Helper()
	w := h.do("GET", target, "", auth)
	if w.Code != http.StatusOK {
		h.t.Fatalf("GET %s: status %d, body %s", target, w.Code, w.Body.String())
	}
	var got aircraftTimeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		h.t.Fatalf("decoding %s: %v", target, err)
	}
	return got
}

func TestAircraftTimeRequiresASession(t *testing.T) {
	h := newHarness(t)
	if w := h.do("GET", "/logbook/api/aircraft-time", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", w.Code)
	}
}

// Minutes on the wire, as everywhere in this API. The invoice is checked in
// H:MM and computed in minutes, so the client formats and the server never
// sends two representations of one figure.
func TestAircraftTimeReportsBlockAndAirSeparately(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	got := h.aircraftTime("/logbook/api/aircraft-time", auth)

	if len(got.Aircraft) != 2 {
		t.Fatalf("got %d aircraft, want 2: %+v", len(got.Aircraft), got.Aircraft)
	}
	// Most time first: OH-PDP flew 90 minutes to OH-CTL's 81.
	if got.Aircraft[0].Registration != "OH-PDP" {
		t.Errorf("first row is %s, want OH-PDP (most block time first)",
			got.Aircraft[0].Registration)
	}
	if got.Total.BlockMinutes != 171 || got.Total.Flights != 2 {
		t.Errorf("total = %d flights / %d block; want 2 / 171",
			got.Total.Flights, got.Total.BlockMinutes)
	}

	// Neither sample flight carries airborne times, so the air figure is zero
	// AND its coverage says so. A page reading these two fields cannot mistake
	// "nobody wrote it down" for "the aeroplane never left the ground".
	if got.Total.AirMinutes != 0 || got.Total.AirKnown != 0 || got.Total.AirMissing != 2 {
		t.Errorf("air = %d over %d known / %d missing; want 0 / 0 / 2",
			got.Total.AirMinutes, got.Total.AirKnown, got.Total.AirMissing)
	}
}

// The airborne pair reaching the aggregation is the whole point of Task 12's
// having put it on the wire, so it is asserted end to end through the write
// path rather than by fabricating a row in the store.
func TestAircraftTimeCountsAFlightWithAirborneTimes(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/flights", `{
		"date":"2026-07-30","aircraft_reg":"OH-CAM","aircraft_type":"C172","class":"SEP_LAND",
		"dep_place":"EFHV","arr_place":"EFHV",
		"off_block":"09:15Z","on_block":"10:30Z",
		"takeoff":"09:20Z","landing":"10:25Z",
		"total_time":"1:15","pic_time":"1:15",
		"night_time":"","instrument_time":"","dual_time":"","instructor_time":"",
		"pic_name":"self","landings_day":1,"landings_night":0,"remarks":""
	}`, auth)
	if w.Code != http.StatusCreated {
		t.Fatalf("creating the flight: status %d, body %s", w.Code, w.Body.String())
	}

	got := h.aircraftTime("/logbook/api/aircraft-time", auth)
	for _, a := range got.Aircraft {
		if a.Registration != "OH-CAM" {
			continue
		}
		if a.BlockMinutes != 75 {
			t.Errorf("OH-CAM block = %d, want 75", a.BlockMinutes)
		}
		if a.AirMinutes != 65 || a.AirKnown != 1 || a.AirMissing != 0 {
			t.Errorf("OH-CAM air = %d over %d known / %d missing; want 65 / 1 / 0",
				a.AirMinutes, a.AirKnown, a.AirMissing)
		}
		if len(a.Types) != 1 || a.Types[0] != "C172" {
			t.Errorf("OH-CAM types = %v, want [C172]", a.Types)
		}
		return
	}
	t.Fatalf("OH-CAM is missing from %+v", got.Aircraft)
}

// The range filters the aggregation the same way it filters the statistics.
func TestAircraftTimeHonoursTheRange(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	got := h.aircraftTime("/logbook/api/aircraft-time?from=2022-01-01", auth)

	if len(got.Aircraft) != 1 || got.Aircraft[0].Registration != "OH-PDP" {
		t.Fatalf("got %+v, want only OH-PDP", got.Aircraft)
	}
	if got.Range["from"] != "2022-01-01" {
		t.Errorf("range echoed as %v, want from=2022-01-01", got.Range)
	}
}

// An unparseable date is a 400, never an ignored filter. Silently widening the
// range would hand back a bill for flights the pilot did not ask about.
func TestAircraftTimeRefusesABadDate(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	w := h.do("GET", "/logbook/api/aircraft-time?from=yesterday", "", auth)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", w.Code)
	}
}

// "The list of flights behind the figure so a disputed line can be traced to a
// flight rather than argued against a single number" (owner, 2026-08-02).
func TestAircraftTimeReturnsTheFlightsBehindTheFigure(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	got := h.aircraftTime("/logbook/api/aircraft-time?reg=OH-CTL", auth)

	if got.Reg != "OH-CTL" {
		t.Errorf("reg echoed as %q, want OH-CTL", got.Reg)
	}
	if len(got.Flights) != 1 || got.Flights[0].AircraftReg != "OH-CTL" {
		t.Fatalf("got %d flights, want the one OH-CTL flight: %+v", len(got.Flights), got.Flights)
	}
	// The summary rows still describe the WHOLE range, so the page can show
	// one aircraft's flights without losing what it is being compared against.
	if len(got.Aircraft) != 2 {
		t.Errorf("got %d summary rows, want all 2 -- asking for one aircraft's "+
			"flights must not narrow the comparison", len(got.Aircraft))
	}
}

// Asking for an aeroplane with nothing in the range is an empty list, not an
// error and not a silent fallback to every flight.
func TestAircraftTimeForAnAircraftWithNoFlightsInRange(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	got := h.aircraftTime("/logbook/api/aircraft-time?reg=OH-NONE", auth)

	if got.Reg != "OH-NONE" {
		t.Errorf("reg echoed as %q, want OH-NONE", got.Reg)
	}
	if len(got.Flights) != 0 {
		t.Errorf("got %d flights, want none: %+v", len(got.Flights), got.Flights)
	}
}

// Without a reg there is no flight list at all -- 1296 flight objects is not
// something to send to a phone that only asked for totals.
func TestAircraftTimeSendsNoFlightsWithoutAReg(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	got := h.aircraftTime("/logbook/api/aircraft-time", auth)

	if len(got.Flights) != 0 {
		t.Errorf("got %d flights without a reg, want none", len(got.Flights))
	}
}
