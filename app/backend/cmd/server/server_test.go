package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

const (
	testPassword = "a-sufficiently-long-passphrase"
	testOrigin   = "https://ayoub.fi"
)

// harness is a server wired to a fresh database with one user and two flights.
type harness struct {
	*Server
	db *store.DB
	t  *testing.T
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "logbook.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Import(sampleLogbook(t), "test"); err != nil {
		t.Fatalf("Import: %v", err)
	}
	if _, err := db.CreateUser("rami", testPassword); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	srv := NewServer(db, Config{
		Origin:       testOrigin,
		SecureCookie: true,
	})
	return &harness{Server: srv, db: db, t: t}
}

// do issues a request against the server. Requests default to a same-origin
// POST so that the CSRF check is satisfied unless a test is exercising it.
func (h *harness) do(method, target string, body string, opts ...func(*http.Request)) *httptest.ResponseRecorder {
	h.t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	r.Header.Set("Origin", testOrigin)
	for _, o := range opts {
		o(r)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// login authenticates and returns the option that carries the session cookie.
func (h *harness) login() func(*http.Request) {
	h.t.Helper()
	w := h.do("POST", "/logbook/api/login",
		fmt.Sprintf(`{"username":"rami","password":%q}`, testPassword))
	if w.Code != http.StatusOK {
		h.t.Fatalf("login: status %d, body %s", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		h.t.Fatal("login set no cookie")
	}
	c := cookies[0]
	return func(r *http.Request) { r.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value}) }
}

func sampleLogbook(t *testing.T) *csvbook.Logbook {
	t.Helper()
	at := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	return &csvbook.Logbook{
		Flights: []csvbook.Flight{
			{
				Seq: 1, Date: "2021-06-01",
				AircraftType: "C172", AircraftReg: "OH-CTL", Class: csvbook.ClassSEPSea,
				DepPlace: "Tuusulanjärvi", ArrPlace: "Tuusulanjärvi",
				OffBlockUTC: at("2021-06-01T15:13:00Z"), OnBlockUTC: at("2021-06-01T16:34:00Z"),
				OffBlockRaw: "18:13", OnBlockRaw: "19:34", TimeOrigin: "converted_from_local",
				BlockMinutes: 81, TotalMinutes: 81, PICMinutes: 81, InstructorMinutes: 81,
				PICName: "self", LandingsDay: 7, LandingsVerified: true,
				SourceBook: 3, SourceRow: 3,
			},
			{
				Seq: 2, Date: "2022-06-02",
				AircraftType: "P28A", AircraftReg: "OH-PDP", Class: csvbook.ClassSEPLand,
				DepPlace: "EFHV", ArrPlace: "EFHV",
				OffBlockUTC: at("2022-06-02T06:00:00Z"), OnBlockUTC: at("2022-06-02T07:30:00Z"),
				OffBlockRaw: "06:00Z", OnBlockRaw: "07:30Z", TimeOrigin: "utc_as_written",
				BlockMinutes: 90, TotalMinutes: 90, DualMinutes: 90,
				NightMinutes: 10, InstrumentMinutes: 20,
				PICName: "Sinervä", LandingsDay: 3, LandingsVerified: false,
				SourceBook: 3, SourceRow: 4,
			},
		},
		Aircraft: []csvbook.Aircraft{
			{Registration: "OH-CTL", Type: "C172", DefaultClass: csvbook.ClassSEPSea, Active: true},
			{Registration: "OH-PDP", Type: "P28A", DefaultClass: csvbook.ClassSEPLand, Active: true},
		},
		Totals: csvbook.Totals{
			Flights: 2, Total: 171, PIC: 81, Dual: 90, Instrument: 20, Night: 10,
			Instructor: 81, SEPSea: 81, Landings: 10,
		},
		Discrepancies: []csvbook.Discrepancy{
			{Kind: csvbook.KindLandingsUnverified, Book: 3, Row: 4, Date: "2022-06-02",
				Detail: "night time logged; the day/night landing split was inferred"},
		},
	}
}

// --- Default deny -----------------------------------------------------------

// TestEveryRouteIsPrivateUnlessExplicitlyPublic is the control from
// app/docs/security.md, and the reason the router keeps a route table at all:
// the test enumerates what is actually registered rather than a list someone
// maintains by hand, so an endpoint added without a thought about auth fails
// here automatically.
func TestEveryRouteIsPrivateUnlessExplicitlyPublic(t *testing.T) {
	h := newHarness(t)

	// The only endpoints that may be reachable without a session. Anything
	// else appearing here is a deliberate decision that has to be made in this
	// test, in this file, on purpose.
	allowed := map[string]bool{
		"POST /logbook/api/login": true,
		"GET /logbook/api/health": true,
	}

	routes := h.Routes()
	if len(routes) < 8 {
		t.Fatalf("only %d routes registered; the enumeration is not seeing the router", len(routes))
	}
	for _, rt := range routes {
		key := rt.Method + " " + rt.Pattern
		t.Run(key, func(t *testing.T) {
			if rt.Public != allowed[key] {
				t.Fatalf("%s has Public=%v; the allow-list says %v. A new public endpoint "+
					"must be added to the allow-list in this test deliberately.",
					key, rt.Public, allowed[key])
			}
			if rt.Public {
				return
			}
			// A concrete path for patterns with a wildcard segment.
			path := strings.ReplaceAll(rt.Pattern, "{id}", "1")
			w := h.do(rt.Method, path, "{}")
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s without a session returned %d, want 401", key, w.Code)
			}
		})
	}
}

func TestAnUnknownPathIs404NotAnAccidentalPass(t *testing.T) {
	h := newHarness(t)
	if w := h.do("GET", "/logbook/api/nope", ""); w.Code != http.StatusNotFound {
		t.Errorf("unknown path returned %d, want 404", w.Code)
	}
}

func TestAGarbageCookieIsRejected(t *testing.T) {
	h := newHarness(t)
	for _, v := range []string{"", "not-a-token", "' OR 1=1 --"} {
		w := h.do("GET", "/logbook/api/me", "", func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: v})
		})
		if w.Code != http.StatusUnauthorized {
			t.Errorf("cookie %q returned %d, want 401", v, w.Code)
		}
	}
}

// --- Login ------------------------------------------------------------------

func TestLoginSetsAHardenedCookie(t *testing.T) {
	h := newHarness(t)
	w := h.do("POST", "/logbook/api/login",
		fmt.Sprintf(`{"username":"rami","password":%q}`, testPassword))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("login set %d cookies, want 1", len(cookies))
	}
	c := cookies[0]
	if c.Name != sessionCookieName {
		t.Errorf("cookie name %q, want %q", c.Name, sessionCookieName)
	}
	if !c.HttpOnly {
		t.Error("the session cookie is readable from JavaScript")
	}
	if !c.Secure {
		t.Error("the session cookie is not marked Secure")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", c.SameSite)
	}
	if c.Path != "/logbook" {
		t.Errorf("cookie path %q, want /logbook -- it must not be sent to the other sites "+
			"sharing this box", c.Path)
	}
	if c.Value == "" {
		t.Error("the cookie has no value")
	}
	// The response body must not echo the token: it would land in logs and in
	// the browser's memory in a second place.
	if strings.Contains(w.Body.String(), c.Value) {
		t.Error("the response body contains the session token")
	}
}

func TestLoginFailsUniformly(t *testing.T) {
	h := newHarness(t)
	var bodies []string
	for _, c := range []struct{ name, body string }{
		{"wrong password", `{"username":"rami","password":"the-wrong-passphrase"}`},
		{"no such user", `{"username":"nobody","password":"a-sufficiently-long-pass"}`},
	} {
		w := h.do("POST", "/logbook/api/login", c.body)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s: status %d, want 401", c.name, w.Code)
		}
		if len(w.Result().Cookies()) != 0 {
			t.Errorf("%s: a failed login set a cookie", c.name)
		}
		bodies = append(bodies, w.Body.String())
	}
	if bodies[0] != bodies[1] {
		t.Errorf("a wrong password and an unknown user gave different responses:\n%q\n%q",
			bodies[0], bodies[1])
	}
}

func TestLoginRejectsMalformedInput(t *testing.T) {
	h := newHarness(t)
	for _, c := range []struct {
		name, body string
		want       int
	}{
		{"not JSON", `<xml/>`, http.StatusBadRequest},
		{"empty object", `{}`, http.StatusUnauthorized},
		{"unknown field", `{"username":"rami","password":"x","admin":true}`, http.StatusBadRequest},
		{"wrong type", `{"username":123,"password":"x"}`, http.StatusBadRequest},
	} {
		t.Run(c.name, func(t *testing.T) {
			if w := h.do("POST", "/logbook/api/login", c.body); w.Code != c.want {
				t.Errorf("status %d, want %d (body %s)", w.Code, c.want, w.Body.String())
			}
		})
	}
}

// TestLoginIsRateLimited is the control from security.md: the Nth failed login
// is throttled, and a success clears the count.
func TestLoginIsRateLimited(t *testing.T) {
	h := newHarness(t)
	bad := `{"username":"rami","password":"the-wrong-passphrase"}`

	var throttled bool
	for range 20 {
		w := h.do("POST", "/logbook/api/login", bad)
		if w.Code == http.StatusTooManyRequests {
			throttled = true
			if w.Header().Get("Retry-After") == "" {
				t.Error("a throttled login carries no Retry-After header")
			}
			break
		}
	}
	if !throttled {
		t.Fatal("twenty failed logins were never throttled")
	}

	// Even the correct password is refused while the penalty stands --
	// otherwise the limiter is trivially bypassed by guessing correctly.
	w := h.do("POST", "/logbook/api/login",
		fmt.Sprintf(`{"username":"rami","password":%q}`, testPassword))
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status %d while throttled, want 429", w.Code)
	}
}

func TestLogoutRevokesTheSession(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	if w := h.do("GET", "/logbook/api/me", "", auth); w.Code != http.StatusOK {
		t.Fatalf("me before logout: %d", w.Code)
	}
	w := h.do("POST", "/logbook/api/logout", "", auth)
	if w.Code != http.StatusOK {
		t.Fatalf("logout: %d", w.Code)
	}
	// The cookie must be cleared in the browser too, not only server-side.
	cleared := w.Result().Cookies()
	if len(cleared) != 1 || cleared[0].MaxAge >= 0 {
		t.Errorf("logout did not expire the cookie: %+v", cleared)
	}
	if w := h.do("GET", "/logbook/api/me", "", auth); w.Code != http.StatusUnauthorized {
		t.Errorf("the session still works after logout: %d", w.Code)
	}
}

// --- CSRF and headers -------------------------------------------------------

// TestCrossOriginMutationsAreRejected is the CSRF control. SameSite=Lax already
// stops most of it; this is the belt to that pair of braces, and it is the part
// that does not depend on the browser behaving.
func TestCrossOriginMutationsAreRejected(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	mutating := []struct{ method, path string }{
		{"POST", "/logbook/api/logout"},
		{"DELETE", "/logbook/api/sessions/1"},
	}
	for _, m := range mutating {
		for _, origin := range []string{"https://evil.example", "null", "http://ayoub.fi"} {
			name := m.method + " " + m.path + " from " + origin
			t.Run(name, func(t *testing.T) {
				w := h.do(m.method, m.path, "{}", auth, func(r *http.Request) {
					r.Header.Set("Origin", origin)
				})
				if w.Code != http.StatusForbidden {
					t.Errorf("status %d, want 403", w.Code)
				}
			})
		}
	}

	// A GET is not state-changing and must still work cross-origin-labelled;
	// blocking it would break nothing an attacker cares about and would break
	// legitimate navigation.
	w := h.do("GET", "/logbook/api/me", "", auth, func(r *http.Request) {
		r.Header.Set("Origin", "https://evil.example")
	})
	if w.Code != http.StatusOK {
		t.Errorf("a cross-origin GET returned %d; SameSite already protects the cookie", w.Code)
	}
}

func TestAMutationWithNoOriginHeaderIsRejected(t *testing.T) {
	// Curl and old browsers send no Origin. Failing closed costs the API
	// nothing -- our own frontend always sends one.
	h := newHarness(t)
	auth := h.login()
	w := h.do("POST", "/logbook/api/logout", "{}", auth, func(r *http.Request) {
		r.Header.Del("Origin")
	})
	if w.Code != http.StatusForbidden {
		t.Errorf("status %d, want 403", w.Code)
	}
}

func TestSecurityHeadersAreOnEveryResponse(t *testing.T) {
	h := newHarness(t)
	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "same-origin",
		"Cache-Control":          "no-store",
	}
	// Including on failures and on 404s, which is where they are most often
	// forgotten.
	for _, target := range []string{"/logbook/api/health", "/logbook/api/me", "/logbook/api/nope"} {
		w := h.do("GET", target, "")
		for k, v := range want {
			if got := w.Header().Get(k); got != v {
				t.Errorf("%s: header %s = %q, want %q", target, k, got, v)
			}
		}
		if w.Header().Get("Content-Security-Policy") == "" {
			t.Errorf("%s: no Content-Security-Policy", target)
		}
	}
}

func TestOversizedBodiesAreRefused(t *testing.T) {
	h := newHarness(t)
	huge := `{"username":"rami","password":"` + strings.Repeat("x", maxBodyBytes) + `"}`
	w := h.do("POST", "/logbook/api/login", huge)
	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 413 or 400", w.Code)
	}
}

func TestFormContentTypesAreRefused(t *testing.T) {
	// The classic CSRF vector is a cross-origin form post. A JSON-only API that
	// rejects form encodings cannot be the target of one at all.
	h := newHarness(t)
	for _, ct := range []string{
		"application/x-www-form-urlencoded",
		"multipart/form-data; boundary=x",
		"text/plain",
	} {
		w := h.do("POST", "/logbook/api/login", `{"username":"rami","password":"x"}`,
			func(r *http.Request) { r.Header.Set("Content-Type", ct) })
		if w.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Content-Type %q returned %d, want 415", ct, w.Code)
		}
	}
}

// --- Reads ------------------------------------------------------------------

func TestMeReportsTheUserAndNoSecrets(t *testing.T) {
	h := newHarness(t)
	w := h.do("GET", "/logbook/api/me", "", h.login())
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var got struct {
		Username string `json:"username"`
		UserID   int64  `json:"user_id"`
	}
	decode(t, w, &got)
	if got.Username != "rami" || got.UserID == 0 {
		t.Errorf("me = %+v", got)
	}
	for _, leak := range []string{"password", "argon2", "token"} {
		if strings.Contains(strings.ToLower(w.Body.String()), leak) {
			t.Errorf("the /me response mentions %q: %s", leak, w.Body.String())
		}
	}
}

func TestFlightsAreListedAndFilteredByDate(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	var all struct {
		Flights []struct {
			Seq          int    `json:"seq"`
			Date         string `json:"date"`
			TotalMinutes int    `json:"total_minutes"`
		} `json:"flights"`
		Count int `json:"count"`
	}
	decode(t, h.do("GET", "/logbook/api/flights", "", auth), &all)
	if all.Count != 2 || len(all.Flights) != 2 {
		t.Fatalf("got %d flights, want 2", all.Count)
	}
	if all.Flights[0].Seq != 1 {
		t.Errorf("flights are not in seq order: first is seq %d", all.Flights[0].Seq)
	}

	var filtered struct {
		Count int `json:"count"`
	}
	decode(t, h.do("GET", "/logbook/api/flights?from=2022-01-01&to=2022-12-31", "", auth), &filtered)
	if filtered.Count != 1 {
		t.Errorf("the 2022 range holds %d flights, want 1", filtered.Count)
	}
}

func TestABadDateParameterIsRejectedRatherThanIgnored(t *testing.T) {
	// Silently ignoring an unparseable date would answer a question the user
	// did not ask, with a total they might write down.
	h := newHarness(t)
	auth := h.login()
	for _, q := range []string{
		"?from=yesterday",
		"?to=2022-13-45",
		"?from=2022-01-01T00:00:00Z",
		"?from='+OR+1=1",
	} {
		if w := h.do("GET", "/logbook/api/flights"+q, "", auth); w.Code != http.StatusBadRequest {
			t.Errorf("%s returned %d, want 400", q, w.Code)
		}
	}
}

func TestStatsReportsTheRangeItWasAsked(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	var got struct {
		Summary struct {
			Flights       int `json:"flights"`
			Total         int `json:"total"`
			PIC           int `json:"pic"`
			Dual          int `json:"dual"`
			Night         int `json:"night"`
			Instrument    int `json:"instrument"`
			Instructor    int `json:"instructor"`
			SeaTotal      int `json:"sea_total"`
			LandTotal     int `json:"land_total"`
			LandingsDay   int `json:"landings_day"`
			LandingsNight int `json:"landings_night"`
			Unverified    int `json:"landings_unverified"`
		} `json:"summary"`
	}
	decode(t, h.do("GET", "/logbook/api/stats", "", auth), &got)

	if got.Summary.Flights != 2 || got.Summary.Total != 171 {
		t.Errorf("summary = %+v, want 2 flights / 171 minutes", got.Summary)
	}
	if got.Summary.SeaTotal+got.Summary.LandTotal != got.Summary.Total {
		t.Error("the sea/land split does not reconstitute the total")
	}
	// The honesty field: one of the two rows has an inferred landing split.
	if got.Summary.Unverified != 1 {
		t.Errorf("landings_unverified = %d, want 1", got.Summary.Unverified)
	}

	// A range with no flights is an empty summary, not an error.
	decode(t, h.do("GET", "/logbook/api/stats?from=2030-01-01", "", auth), &got)
	if got.Summary.Flights != 0 || got.Summary.Total != 0 {
		t.Errorf("an empty range summarised to %+v", got.Summary)
	}
}

func TestAircraftAndDiscrepanciesAreServed(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	var ac struct {
		Aircraft []struct {
			Registration string `json:"registration"`
		} `json:"aircraft"`
	}
	decode(t, h.do("GET", "/logbook/api/aircraft", "", auth), &ac)
	if len(ac.Aircraft) != 2 {
		t.Errorf("got %d aircraft, want 2", len(ac.Aircraft))
	}

	var ds struct {
		Discrepancies []struct {
			Kind string `json:"kind"`
		} `json:"discrepancies"`
		Count int `json:"count"`
	}
	decode(t, h.do("GET", "/logbook/api/discrepancies", "", auth), &ds)
	if ds.Count != 1 || ds.Discrepancies[0].Kind != string(csvbook.KindLandingsUnverified) {
		t.Errorf("discrepancies = %+v", ds)
	}
}

// --- Session management -----------------------------------------------------

func TestTheSessionListNeverCarriesAToken(t *testing.T) {
	h := newHarness(t)
	auth := h.login()

	w := h.do("GET", "/logbook/api/sessions", "", auth)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var got struct {
		Sessions []struct {
			ID      int64  `json:"id"`
			Current bool   `json:"current"`
			IP      string `json:"ip"`
		} `json:"sessions"`
	}
	decode(t, w, &got)
	if len(got.Sessions) != 1 {
		t.Fatalf("got %d sessions, want 1", len(got.Sessions))
	}
	if !got.Sessions[0].Current {
		t.Error("the session being used is not marked as the current one")
	}
	for _, k := range []string{"token", "hash"} {
		if strings.Contains(strings.ToLower(w.Body.String()), k) {
			t.Errorf("the session list mentions %q: %s", k, w.Body.String())
		}
	}
}

func TestASessionCanBeRevokedFromTheList(t *testing.T) {
	h := newHarness(t)
	first := h.login()
	second := h.login() // a second device

	var got struct {
		Sessions []struct {
			ID      int64 `json:"id"`
			Current bool  `json:"current"`
		} `json:"sessions"`
	}
	decode(t, h.do("GET", "/logbook/api/sessions", "", second), &got)
	if len(got.Sessions) != 2 {
		t.Fatalf("got %d sessions, want 2", len(got.Sessions))
	}
	var other int64
	for _, s := range got.Sessions {
		if !s.Current {
			other = s.ID
		}
	}
	if other == 0 {
		t.Fatal("both sessions claim to be the current one")
	}

	w := h.do("DELETE", fmt.Sprintf("/logbook/api/sessions/%d", other), "", second)
	if w.Code != http.StatusOK {
		t.Fatalf("revoke: status %d, body %s", w.Code, w.Body.String())
	}
	if w := h.do("GET", "/logbook/api/me", "", first); w.Code != http.StatusUnauthorized {
		t.Error("the revoked session still works")
	}
	if w := h.do("GET", "/logbook/api/me", "", second); w.Code != http.StatusOK {
		t.Error("revoking one session killed the other")
	}
}

func TestRevokingANonexistentSessionIs404(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	for _, id := range []string{"99999", "abc", "-1"} {
		w := h.do("DELETE", "/logbook/api/sessions/"+id, "", auth)
		if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
			t.Errorf("revoking session %q returned %d, want 404 or 400", id, w.Code)
		}
	}
}

func TestHealthSaysNothingUseful(t *testing.T) {
	// It is public, so it must not disclose the version, the flight count, the
	// database path or anything else worth knowing.
	h := newHarness(t)
	w := h.do("GET", "/logbook/api/health", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if body := strings.TrimSpace(w.Body.String()); body != `{"status":"ok"}` {
		t.Errorf("health said %q, want exactly {\"status\":\"ok\"}", body)
	}
}

func decode(t *testing.T, w *httptest.ResponseRecorder, into any) {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type %q, want application/json", ct)
	}
	if err := json.Unmarshal(w.Body.Bytes(), into); err != nil {
		t.Fatalf("decoding %s: %v", w.Body.String(), err)
	}
}
