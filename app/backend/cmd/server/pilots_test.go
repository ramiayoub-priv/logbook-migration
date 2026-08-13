package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// The pilot roster over HTTP -- Task 21, owner ask 2026-08-03.
//
// The PIC field was free text. `self` appears on 1143 of the 1296 transcribed
// flights, and one stray keystroke -- `sself`, `SELF`, `seeelf` -- would split
// that across two spellings with nothing to flag it. The roster is the same
// answer the registrations got: a list to pick from.
//
// The surface is deliberately SMALLER than the aircraft one: read and create,
// with no update and no delete. A roster entry is only a spelling, and renaming
// one could not rename the flights that use it -- they are the record. A wrong
// name is corrected on the flight itself, through the ordinary edit path.

type pilotItem struct {
	Name      string `json:"name"`
	UserAdded bool   `json:"user_added"`
	LastFlown string `json:"last_flown"`
	Flights   int    `json:"flights"`
}

type pilotList struct {
	Pilots []pilotItem `json:"pilots"`
}

type pilotEnvelope struct {
	Pilot pilotItem `json:"pilot"`
}

func (h *harness) roster(auth func(*http.Request)) pilotList {
	h.t.Helper()
	var got pilotList
	decodeStatus(h.t, h.do("GET", "/logbook/api/pilots", "", auth), http.StatusOK, &got)
	return got
}

// The roster is derived: every name the seeded flights use is offered, with the
// count behind it, without anybody having maintained a list.
func TestThePilotListIsDerivedFromTheFlights(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	got := h.roster(auth)
	if len(got.Pilots) == 0 {
		t.Fatal("the roster is empty; it should name everybody on the seeded flights")
	}
	var found bool
	for _, p := range got.Pilots {
		if p.Name == "" {
			t.Error("the roster offers an empty name")
		}
		if p.Name == "self" {
			found = true
			if p.Flights == 0 {
				t.Error("self is offered with no flights behind it")
			}
			if p.UserAdded {
				t.Error("self came from the flights and must not read as user-added")
			}
		}
	}
	if !found {
		t.Error("the roster does not offer `self`")
	}
}

func TestCreatePilotMakesANeverFlownNamePickable(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/pilots", `{"name":"Jansson"}`, auth)
	var env pilotEnvelope
	decodeStatus(t, w, http.StatusCreated, &env)
	if env.Pilot.Name != "Jansson" || !env.Pilot.UserAdded || env.Pilot.Flights != 0 {
		t.Errorf("created %+v, want Jansson, user-added, no flights", env.Pilot)
	}

	// And it is now offered -- first, because it has not been flown with.
	list := h.roster(auth).Pilots
	if len(list) == 0 || list[0].Name != "Jansson" {
		t.Errorf("a never-flown name should come first; got %+v", list)
	}
}

// The owner's ask, asserted at the API: `SELF` must not become a second entry.
func TestCreatePilotRefusesACaseVariantOfAKnownName(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	before := len(h.roster(auth).Pilots)
	for _, body := range []string{`{"name":"SELF"}`, `{"name":"Self"}`, `{"name":" self "}`} {
		w := h.do("POST", "/logbook/api/pilots", body, auth)
		if w.Code != http.StatusConflict {
			t.Errorf("POST %s: status %d, want 409; body %s", body, w.Code, w.Body.String())
		}
	}
	if after := len(h.roster(auth).Pilots); after != before {
		t.Errorf("the roster grew from %d to %d names", before, after)
	}
}

func TestCreatePilotRefusesAnEmptyName(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/pilots", `{"name":"   "}`, auth)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400; body %s", w.Code, w.Body.String())
	}
}

// Adding a name must not be able to touch the legal record. The roster is a
// seed list for a form, exactly like the aircraft table (rule 0.8).
func TestCreatingAPilotChangesNoFlight(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	// The whole response body, byte for byte. Comparing parsed fields would
	// only assert the fields somebody remembered to compare.
	before := h.do("GET", "/logbook/api/flights", "", auth).Body.String()
	if w := h.do("POST", "/logbook/api/pilots", `{"name":"Ravantti"}`, auth); w.Code != http.StatusCreated {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}
	after := h.do("GET", "/logbook/api/flights", "", auth).Body.String()

	if before != after {
		t.Error("GET /flights changed after a name was added to the roster")
	}
}

// There is no update and no delete, and both absences are asserted rather than
// assumed -- the aircraft ruling was written down the same way for the same
// reason. See the file comment for why the roster's surface is smaller.
func TestThereAreNoPilotWriteRoutesBeyondCreate(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	for _, c := range []struct{ method, target, body string }{
		{"PUT", "/logbook/api/pilots/self", `{"name":"Self"}`},
		{"DELETE", "/logbook/api/pilots/self", ""},
	} {
		w := h.do(c.method, c.target, c.body, auth)
		if w.Code == http.StatusOK || w.Code == http.StatusNoContent {
			t.Errorf("%s %s succeeded with %d; the roster has no such route",
				c.method, c.target, w.Code)
		}
	}
	for _, r := range h.Routes() {
		if r.Pattern == basePath+"/pilots/{name}" {
			t.Errorf("a %s route for a single pilot is mounted", r.Method)
		}
	}
}

// Default deny covers the new routes, checked against the router's own table
// rather than hoped for.
func TestThePilotRoutesAreNotPublic(t *testing.T) {
	h := newHarness(t)

	for _, c := range []struct{ method, target, body string }{
		{"GET", "/logbook/api/pilots", ""},
		{"POST", "/logbook/api/pilots", `{"name":"Nobody"}`},
	} {
		w := h.do(c.method, c.target, c.body)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without a session: status %d, want 401", c.method, c.target, w.Code)
		}
	}

	auth := h.login()
	for _, p := range h.roster(auth).Pilots {
		if p.Name == "Nobody" {
			t.Error("an unauthenticated POST created a pilot")
		}
	}
}

// --- The roster is enforced where the record is actually written ------------
//
// The picker refuses to OFFER a case variant, but a form is not a guarantee:
// the field still accepts what is typed, because an edited flight may name
// somebody the roster no longer lists and blanking that would lose a name off a
// legal record. So the guarantee lives here, at the write, where it belongs.
//
// This can never block an edit. The roster is derived from the flights, so
// every name already in the record is in it by construction -- including the
// one on the flight being edited.

func flightWith(picName string) string {
	return strings.Replace(goodFlight, `"pic_name": "self"`,
		`"pic_name": "`+picName+`"`, 1)
}

func TestCreateFlightRefusesAPICNameTheRosterDoesNotKnow(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	for _, name := range []string{"SELF", "sself", "seeelf", "Marteuvo"} {
		w := h.do("POST", "/logbook/api/flights", flightWith(name), auth)
		if w.Code != http.StatusBadRequest {
			t.Errorf("pic_name %q: status %d, want 400; body %s", name, w.Code, w.Body.String())
			continue
		}
		if !strings.Contains(w.Body.String(), "pic_name") {
			t.Errorf("pic_name %q: the refusal does not name the field: %s", name, w.Body.String())
		}
	}
}

func TestCreateFlightAcceptsTheRostersOwnSpellingAndAnEmptyName(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("POST", "/logbook/api/flights", flightWith("self"), auth); w.Code != http.StatusCreated {
		t.Errorf("the roster's own spelling was refused: %d %s", w.Code, w.Body.String())
	}
	// One transcribed row has a blank PIC cell, so blank must stay writable --
	// refusing it would make a flight that exists on paper unenterable.
	body := strings.Replace(flightWith(""), `"date": "2026-07-30"`, `"date": "2026-07-31"`, 1)
	if w := h.do("POST", "/logbook/api/flights", body, auth); w.Code != http.StatusCreated {
		t.Errorf("an empty pic_name was refused: %d %s", w.Code, w.Body.String())
	}
}

// A name added to the roster is immediately usable, which is what makes the
// picker's "add" row worth having at an airfield.
func TestAPilotAddedToTheRosterCanBeFlownImmediately(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("POST", "/logbook/api/pilots", `{"name":"Tarhanen"}`, auth); w.Code != http.StatusCreated {
		t.Fatalf("adding the name: %d %s", w.Code, w.Body.String())
	}
	if w := h.do("POST", "/logbook/api/flights", flightWith("Tarhanen"), auth); w.Code != http.StatusCreated {
		t.Errorf("a just-added name was refused on a flight: %d %s", w.Code, w.Body.String())
	}
}

func TestUpdateFlightRefusesAPICNameTheRosterDoesNotKnow(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	var created struct {
		Flight struct {
			Seq int `json:"seq"`
		} `json:"flight"`
	}
	decodeStatus(t, h.do("POST", "/logbook/api/flights", goodFlight, auth),
		http.StatusCreated, &created)

	target := fmt.Sprintf("/logbook/api/flights/%d", created.Flight.Seq)
	if w := h.do("PUT", target, flightWith("SELF"), auth); w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400; body %s", w.Code, w.Body.String())
	}
	// And the flight kept the name it had.
	if w := h.do("GET", target, "", auth); !strings.Contains(w.Body.String(), `"pic_name":"self"`) {
		t.Errorf("the stored name changed: %s", w.Body.String())
	}
}

// TestThePICNameRefusalHasTheSameShapeAsEveryOtherFlightRefusal.
//
// Caught by running the thing, not by a test: the first version answered with
// a field MAP where every other refusal from these two endpoints answers with
// a field ARRAY. The frontend reads `fields.length` and iterates, so a map
// meant the pilot got a banner saying "see the fields below" and no field
// marked below. A refusal that cannot be rendered is barely a refusal.
func TestThePICNameRefusalHasTheSameShapeAsEveryOtherFlightRefusal(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	var got struct {
		Error  string `json:"error"`
		Fields []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"fields"`
	}
	decodeStatus(t, h.do("POST", "/logbook/api/flights", flightWith("SELF"), auth),
		http.StatusBadRequest, &got)

	if len(got.Fields) != 1 {
		t.Fatalf("fields = %+v, want exactly one entry as an ARRAY", got.Fields)
	}
	if got.Fields[0].Field != "pic_name" {
		t.Errorf("fields[0].field = %q, want pic_name", got.Fields[0].Field)
	}
	if !strings.Contains(got.Fields[0].Message, "SELF") {
		t.Errorf("the message does not quote what was typed: %q", got.Fields[0].Message)
	}
}
