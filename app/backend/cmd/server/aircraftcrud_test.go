package main

import (
	"fmt"
	"net/http"
	"testing"
)

// The aircraft write path (owner ruling, 2026-08-02).
//
// The list used to be derived from the flights alone, so the only aeroplanes
// that could exist were the ones already flown -- and the form's registration
// is a dropdown fed by this list. A first flight in a new aeroplane was
// unenterable.
//
// There is deliberately NO DELETE route, and there is no active/retired flag.
// Both are owner rulings: an aeroplane once added stays, a wrong one is
// corrected with a PUT, and the list is kept usable by being filterable and
// ordered by what was flown most recently.

type aircraftItem struct {
	Registration string `json:"registration"`
	Type         string `json:"type"`
	DefaultClass string `json:"default_class"`
	IFRCapable   bool   `json:"ifr_capable"`
	Notes        string `json:"notes"`
	UserAdded    bool   `json:"user_added"`
	LastFlown    string `json:"last_flown"`
	Flights      int    `json:"flights"`
}

type aircraftList struct {
	Aircraft []aircraftItem `json:"aircraft"`
}

// Single writes come back wrapped, the same shape POST /flights uses.
type aircraftEnvelope struct {
	Aircraft aircraftItem `json:"aircraft"`
}

func (h *harness) fleet(auth func(*http.Request)) aircraftList {
	h.t.Helper()
	var got aircraftList
	decodeStatus(h.t, h.do("GET", "/logbook/api/aircraft", "", auth), http.StatusOK, &got)
	return got
}

func TestCreateAircraftMakesANeverFlownAeroplaneSelectable(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"OH-XYZ","type":"C152","default_class":"SEP_LAND"}`, auth)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}
	var env aircraftEnvelope
	decodeStatus(t, w, http.StatusCreated, &env)
	created := env.Aircraft
	if created.Registration != "OH-XYZ" || !created.UserAdded {
		t.Errorf("created %+v", created)
	}
	if created.Flights != 0 || created.LastFlown != "" {
		t.Errorf("a never-flown aeroplane reports %d flights, last flown %q",
			created.Flights, created.LastFlown)
	}

	// And it is now in the list the form builds its dropdown from -- first,
	// because it has never been flown and was added in order to be.
	list := h.fleet(auth)
	if len(list.Aircraft) == 0 || list.Aircraft[0].Registration != "OH-XYZ" {
		t.Errorf("the new aeroplane is not at the head of the list: %+v", list.Aircraft)
	}
}

// The whole point: a flight in it must now be enterable end to end.
func TestAFlightCanBeLoggedInAnAircraftAddedThroughTheAPI(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"OH-XYZ","type":"C152","default_class":"SEP_LAND"}`, auth); w.Code != http.StatusCreated {
		t.Fatalf("creating the aircraft: %d %s", w.Code, w.Body.String())
	}

	w := h.do("POST", "/logbook/api/flights", `{
		"date":"2026-07-30","aircraft_reg":"OH-XYZ","aircraft_type":"C152",
		"class":"SEP_LAND","dep_place":"EFHV","arr_place":"EFHV",
		"off_block":"09:15Z","on_block":"10:30Z","total_time":"1:15","pic_time":"1:15",
		"pic_name":"self","landings_day":1}`, auth)
	if w.Code != http.StatusCreated {
		t.Fatalf("logging a flight in the new aeroplane: %d %s", w.Code, w.Body.String())
	}

	// And the aeroplane now reports that flight, derived rather than stored.
	for _, a := range h.fleet(auth).Aircraft {
		if a.Registration != "OH-XYZ" {
			continue
		}
		if a.Flights != 1 || a.LastFlown != "2026-07-30" {
			t.Errorf("after one flight OH-XYZ reports %d flights, last flown %q",
				a.Flights, a.LastFlown)
		}
		return
	}
	t.Error("OH-XYZ vanished from the list after being flown")
}

func TestCreateAircraftRefusesADuplicateRegistration(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	body := `{"registration":"OH-XYZ","type":"C152","default_class":"SEP_LAND"}`
	if w := h.do("POST", "/logbook/api/aircraft", body, auth); w.Code != http.StatusCreated {
		t.Fatalf("first create: %d %s", w.Code, w.Body.String())
	}
	w := h.do("POST", "/logbook/api/aircraft", body, auth)
	if w.Code != http.StatusConflict {
		t.Fatalf("second create: status %d, want 409; body %s", w.Code, w.Body.String())
	}
}

// One already in the books is a duplicate too, and the message has to say so
// rather than reading as a mysterious failure.
func TestCreateAircraftRefusesOneAlreadyInTheBooks(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"OH-CTL","type":"C172","default_class":"SEP_SEA"}`, auth)
	if w.Code != http.StatusConflict {
		t.Fatalf("status %d, want 409; body %s", w.Code, w.Body.String())
	}
}

func TestCreateAircraftValidatesItsFields(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	for _, c := range []struct{ name, body string }{
		{"no registration", `{"registration":"","type":"C152","default_class":"SEP_LAND"}`},
		{"no type", `{"registration":"OH-XYZ","type":"","default_class":"SEP_LAND"}`},
		{"bad class", `{"registration":"OH-XYZ","type":"C152","default_class":"BOAT"}`},
		{"not json", `{`},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := h.do("POST", "/logbook/api/aircraft", c.body, auth)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status %d, want 400; body %s", w.Code, w.Body.String())
			}
		})
	}
}

// A registration is stored upper-cased and trimmed, so that "oh-xyz" typed on a
// phone keyboard cannot become a second aeroplane.
func TestCreateAircraftNormalizesTheRegistration(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"  oh-xyz  ","type":"c152","default_class":"SEP_LAND"}`, auth)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}
	var env aircraftEnvelope
	decodeStatus(t, w, http.StatusCreated, &env)
	created := env.Aircraft
	if created.Registration != "OH-XYZ" || created.Type != "C152" {
		t.Errorf("created %+v, want the registration and type normalized", created)
	}

	// And the lower-case spelling is now a duplicate, not a new aeroplane.
	if w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"oh-xyz","type":"C152","default_class":"SEP_LAND"}`, auth); w.Code != http.StatusConflict {
		t.Errorf("the same registration in lower case was accepted as new: %d", w.Code)
	}
}

func TestUpdateAircraftCorrectsIt(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"OH-XZY","type":"C152","default_class":"SEP_LAND"}`, auth); w.Code != http.StatusCreated {
		t.Fatal(w.Body.String())
	}

	w := h.do("PUT", "/logbook/api/aircraft/OH-XZY",
		`{"registration":"OH-XYZ","type":"C152","default_class":"SEP_LAND","notes":"typo fixed"}`, auth)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}
	var env aircraftEnvelope
	decode(t, w, &env)
	got := env.Aircraft
	if got.Registration != "OH-XYZ" || got.Notes != "typo fixed" {
		t.Errorf("updated to %+v", got)
	}
}

// Editing an aeroplane that came from the paper books is ALLOWED -- this table
// seeds a form, it is not the record. What must never happen is a flight
// moving, so that is what the test asserts.
func TestUpdateAircraftFromTheBooksMovesNoFlight(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	before := h.do("GET", "/logbook/api/flights", "", auth).Body.String()

	w := h.do("PUT", "/logbook/api/aircraft/OH-CTL",
		`{"registration":"OH-CTL","type":"PA18","default_class":"SEP_LAND"}`, auth)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}

	after := h.do("GET", "/logbook/api/flights", "", auth).Body.String()
	if before != after {
		t.Error("editing an aeroplane changed the flights.\n" +
			"The flights record what was written on paper; the aircraft row seeds a form.\n" +
			"before: " + before + "\nafter:  " + after)
	}
}

func TestUpdateAircraftRefusesToCollide(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"OH-XYZ","type":"C152","default_class":"SEP_LAND"}`, auth); w.Code != http.StatusCreated {
		t.Fatal(w.Body.String())
	}
	w := h.do("PUT", "/logbook/api/aircraft/OH-XYZ",
		`{"registration":"OH-CTL","type":"C172","default_class":"SEP_SEA"}`, auth)
	if w.Code != http.StatusConflict {
		t.Fatalf("status %d, want 409; body %s", w.Code, w.Body.String())
	}
}

func TestUpdateAircraftThatIsNotThere(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("PUT", "/logbook/api/aircraft/OH-NOPE",
		`{"registration":"OH-NOPE","type":"C152","default_class":"SEP_LAND"}`, auth)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404; body %s", w.Code, w.Body.String())
	}
}

// Owner ruling: there is no delete. Asserted rather than assumed, because a
// later session adding one "for symmetry" is exactly how a ruling gets lost.
func TestThereIsNoAircraftDeleteRoute(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("DELETE", "/logbook/api/aircraft/OH-CTL", "", auth)
	if w.Code == http.StatusOK || w.Code == http.StatusNoContent {
		t.Fatalf("DELETE /aircraft/{reg} succeeded with %d; the owner ruled that "+
			"aircraft are never deleted, only corrected", w.Code)
	}

	for _, r := range h.Routes() {
		if r.Method == "DELETE" && r.Pattern == basePath+"/aircraft/{reg}" {
			t.Error("a DELETE route for aircraft is mounted")
		}
	}
}

// Default deny covers the new routes. The router's table makes this checkable
// rather than hopeful.
func TestTheAircraftWriteRoutesAreNotPublic(t *testing.T) {
	h := newHarness(t)

	for _, c := range []struct{ method, target, body string }{
		{"POST", "/logbook/api/aircraft",
			`{"registration":"OH-XYZ","type":"C152","default_class":"SEP_LAND"}`},
		{"PUT", "/logbook/api/aircraft/OH-CTL",
			`{"registration":"OH-CTL","type":"C172","default_class":"SEP_SEA"}`},
	} {
		w := h.do(c.method, c.target, c.body)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without a session: status %d, want 401",
				c.method, c.target, w.Code)
		}
	}

	// And nothing was written by the attempt.
	auth := h.login()
	for _, a := range h.fleet(auth).Aircraft {
		if a.Registration == "OH-XYZ" {
			t.Error("an unauthenticated POST created an aircraft")
		}
	}
}

// The list is ordered for the dropdown: never-flown first, then most recently
// flown. That ordering is what the owner asked for in place of a retired flag.
func TestTheAircraftListIsOrderedForTheDropdown(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"OH-XYZ","type":"C152","default_class":"SEP_LAND"}`, auth); w.Code != http.StatusCreated {
		t.Fatal(w.Body.String())
	}

	list := h.fleet(auth).Aircraft
	if len(list) < 3 {
		t.Fatalf("expected at least 3 aircraft, got %d", len(list))
	}
	if list[0].Registration != "OH-XYZ" {
		t.Errorf("never-flown aeroplane is not first: %+v", list)
	}
	for i := 2; i < len(list); i++ {
		if list[i-1].LastFlown < list[i].LastFlown {
			t.Errorf("out of order at %d: %s (%s) before %s (%s)", i,
				list[i-1].Registration, list[i-1].LastFlown,
				list[i].Registration, list[i].LastFlown)
		}
	}
}

// A rejected write must say which field, in the same shape POST /flights uses,
// or the form cannot put the message next to the control.
func TestAircraftValidationNamesTheField(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"OH-XYZ","type":"","default_class":"SEP_LAND"}`, auth)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", w.Code)
	}
	var got struct {
		Fields map[string]string `json:"fields"`
	}
	decodeStatus(t, w, http.StatusBadRequest, &got)
	if _, ok := got.Fields["type"]; !ok {
		t.Errorf("errors were %v, want one against \"type\"", got.Fields)
	}
}

func TestAircraftRoutesRejectAnUnknownRegistrationShape(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	// A registration is normalized before lookup, so a lower-case path segment
	// finds the aeroplane rather than 404ing confusingly.
	if w := h.do("POST", "/logbook/api/aircraft",
		`{"registration":"OH-XYZ","type":"C152","default_class":"SEP_LAND"}`, auth); w.Code != http.StatusCreated {
		t.Fatal(w.Body.String())
	}
	w := h.do("PUT", "/logbook/api/aircraft/oh-xyz",
		fmt.Sprintf(`{"registration":"OH-XYZ","type":"C152","default_class":"SEP_LAND","notes":%q}`,
			"found by the lower-case path"), auth)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200; body %s", w.Code, w.Body.String())
	}
}
