// Command server is the logbook HTTP API.
//
// The design is in app/docs/security.md. Two things about this file are
// structural rather than conventional, and both exist to make the security
// posture something the compiler and the tests enforce rather than something a
// future change has to remember:
//
//   - Routes are registered through a table, not straight onto a ServeMux, and
//     the auth wrapper is applied by the registration function. A handler
//     cannot be mounted without passing through it. `public` is a field the
//     author has to set, so a new endpoint is private by default and by
//     construction (rule 0.3, default deny).
//   - Routes() exposes that table so a test can enumerate what is actually
//     mounted and assert 401 on everything not on a short allow-list. An
//     endpoint added without thinking about authentication fails the suite.
//
// The API is served under /logbook/api/ rather than /api/, because /api/ on
// ayoub.fi is already taken by a stale transit proxy (app/docs/deploy.md).
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/ratelimit"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

// basePath is where the API lives. Everything below is relative to it.
const basePath = "/logbook/api"

// cookiePath scopes the session cookie to this application. The box serves the
// owner's other sites, and a cookie at "/" would be sent to all of them.
const cookiePath = "/logbook"

const sessionCookieName = "logbook_session"

// maxBodyBytes bounds every request body. The largest thing this API accepts is
// a login or a flight, both of which are well under a kilobyte; 64 KiB is
// generous and still means an unauthenticated caller cannot make the server
// allocate anything interesting.
const maxBodyBytes = 64 << 10

// Config is everything the server needs, all of it from the environment. There
// are no secrets here: passwords are hashed in the database and sessions are
// database rows, so there is no signing key to leak (rule 0.3, no secrets in
// the repo).
type Config struct {
	Addr string
	// Origin is the exact scheme+host the frontend is served from, e.g.
	// "https://ayoub.fi". It is what the CSRF check compares against.
	Origin string
	// SecureCookie marks the cookie TLS-only. True everywhere real; the only
	// reason to turn it off is a plain-HTTP local development server.
	SecureCookie bool
	Logger       *slog.Logger
	// Now is the source of "today" for the new-flight form's future-date
	// check. Injectable so that the check is tested against a fixed instant
	// rather than against the day the suite happens to run -- a test that
	// starts failing when the calendar catches up is a test nobody trusts.
	// Defaults to time.Now.
	Now func() time.Time
	// HolderName is the licence holder's name, printed on the exported PDFs.
	// It is presentation only -- no figure depends on it -- so an empty value
	// simply leaves the documents unnamed rather than failing anything.
	HolderName string
}

// RouteInfo describes one mounted endpoint. Exported so the default-deny test
// can enumerate the router rather than a hand-maintained list.
type RouteInfo struct {
	Method  string
	Pattern string
	Public  bool
}

// Server is the API. It is an http.Handler, so tests drive it directly with
// httptest and never bind a port.
type Server struct {
	db      *store.DB
	cfg     Config
	limiter *ratelimit.Limiter
	mux     *http.ServeMux
	routes  []RouteInfo
	log     *slog.Logger
}

// NewServer builds the API and mounts every route.
func NewServer(db *store.DB, cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	s := &Server{
		db:      db,
		cfg:     cfg,
		limiter: ratelimit.New(ratelimit.Config{}),
		mux:     http.NewServeMux(),
		log:     cfg.Logger,
	}

	// Public. Both entries are deliberate and both are named in the
	// default-deny test's allow-list; adding a third means editing that test.
	s.public("POST", "/login", s.handleLogin)
	s.public("GET", "/health", s.handleHealth)

	// Private: everything else.
	s.private("POST", "/logout", s.handleLogout)
	s.private("GET", "/me", s.handleMe)
	s.private("GET", "/flights", s.handleFlights)
	s.private("POST", "/flights", s.handleCreateFlight)
	s.private("GET", "/flights/{seq}", s.handleFlight)
	s.private("PUT", "/flights/{seq}", s.handleUpdateFlight)
	s.private("DELETE", "/flights/{seq}", s.handleDeleteFlight)
	s.private("GET", "/aircraft", s.handleAircraft)
	s.private("POST", "/aircraft", s.handleCreateAircraft)
	// No DELETE, deliberately. Owner ruling 2026-08-02: an aeroplane once added
	// stays, a wrong one is corrected with a PUT, and the list is kept usable by
	// being filterable rather than by hiding rows. A test asserts the absence,
	// because a later session adding one "for symmetry" is how a ruling is lost.
	s.private("PUT", "/aircraft/{reg}", s.handleUpdateAircraft)
	s.private("GET", "/pilots", s.handlePilots)
	// Create only. No PUT and no DELETE: a roster entry is a spelling, and
	// renaming one could not rename the flights that carry it -- they are the
	// record. A wrong name is corrected on the flight itself. Asserted by a test
	// for the same reason the aircraft ruling is. See internal/store/pilots.go.
	s.private("POST", "/pilots", s.handleCreatePilot)
	s.private("GET", "/stats", s.handleStats)
	s.private("GET", "/aircraft-time", s.handleAircraftTime)
	s.private("GET", "/discrepancies", s.handleDiscrepancies)
	s.private("GET", "/export/easa.pdf", s.handleExportEASA)
	s.private("GET", "/export/table.pdf", s.handleExportTable)
	s.private("GET", "/export/statistics.pdf", s.handleExportStatistics)
	s.private("GET", "/sessions", s.handleSessions)
	s.private("DELETE", "/sessions/{id}", s.handleRevokeSession)

	return s
}

// now is the current instant, from the injectable clock.
func (s *Server) now() time.Time { return s.cfg.Now() }

// Routes reports what is mounted. Used by the default-deny test.
func (s *Server) Routes() []RouteInfo {
	out := make([]RouteInfo, len(s.routes))
	copy(out, s.routes)
	return out
}

func (s *Server) public(method, path string, h http.HandlerFunc)  { s.route(method, path, true, h) }
func (s *Server) private(method, path string, h http.HandlerFunc) { s.route(method, path, false, h) }

// route mounts a handler, wrapping it in the checks every endpoint gets. The
// wrapping happens here and nowhere else, so it cannot be forgotten at a call
// site.
func (s *Server) route(method, path string, public bool, h http.HandlerFunc) {
	pattern := basePath + path
	s.routes = append(s.routes, RouteInfo{Method: method, Pattern: pattern, Public: public})

	wrapped := h
	if !public {
		wrapped = s.requireSession(wrapped)
	}
	wrapped = s.requireJSONBody(wrapped)
	wrapped = s.checkOrigin(wrapped)
	s.mux.Handle(method+" "+pattern, wrapped)
}

// ServeHTTP applies the checks that must hold for every response, including
// 404s and method mismatches, then dispatches.
//
// Security headers and the body cap go on out here rather than in route()
// because a request that never matches a route still gets a response, and a
// 404 without nosniff is as much of a hole as a 200 without it.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.setSecurityHeaders(w)
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	s.mux.ServeHTTP(w, r)
}

// setSecurityHeaders applies the header set from app/docs/security.md.
//
// The CSP is as tight as an API can be: this server returns only JSON, so
// nothing may be loaded or executed from any of its responses at all. The
// frontend's own CSP is set by Apache on the static files it serves.
func (s *Server) setSecurityHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "same-origin")
	// The logbook is personal data and every response is user-specific. No
	// intermediary, and no browser cache on a shared device, should keep it.
	h.Set("Cache-Control", "no-store")
}

// checkOrigin is the CSRF control for state-changing methods.
//
// SameSite=Lax on the cookie already stops the cross-site form post, but that
// is the browser's promise rather than ours. This check does not depend on the
// browser behaving, and it fails closed: a request with no Origin at all is
// refused, which costs nothing because our own frontend always sends one.
//
// Safe methods are exempt. A cross-origin GET cannot change anything, and the
// same-origin policy already stops the attacker from reading the response.
func (s *Server) checkOrigin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next(w, r)
			return
		}
		if origin := r.Header.Get("Origin"); origin == "" || origin != s.cfg.Origin {
			s.log.Warn("cross-origin request refused",
				"method", r.Method, "path", r.URL.Path, "origin", origin)
			writeError(w, http.StatusForbidden, "cross-origin request refused")
			return
		}
		next(w, r)
	}
}

// requireJSONBody rejects anything that is not JSON on a request that carries a
// body.
//
// This is a CSRF control as much as a parsing one. A cross-origin HTML form can
// only send three content types, none of them application/json, so an API that
// accepts nothing else cannot be the target of a classic form-post CSRF at all
// -- regardless of what the Origin header says.
func (s *Server) requireJSONBody(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "" {
			media, _, _ := strings.Cut(ct, ";")
			if strings.TrimSpace(media) != "application/json" {
				writeError(w, http.StatusUnsupportedMediaType, "this API accepts application/json only")
				return
			}
		}
		next(w, r)
	}
}

// requireSession is the default-deny gate. Everything private goes through it.
func (s *Server) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		sess, user, err := s.db.LookupSession(c.Value)
		if err != nil {
			// Unknown, expired, or a disabled account -- all one answer. The
			// distinction is logged for the operator, never returned.
			s.log.Info("session rejected", "path", r.URL.Path, "reason", err)
			s.clearCookie(w)
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next(w, r.WithContext(withCaller(r.Context(), caller{User: user, Session: sess})))
	}
}

// writeJSON sends a value as JSON.
//
// Durations are minutes, as everywhere inside this application; H:MM is a
// presentation concern and the frontend formats it. One representation on the
// wire means there is no second figure that can disagree with the first.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// The status line is already sent, so there is nothing to tell the
		// client. Log it and move on.
		slog.Error("writing response", "error", err)
	}
}

// writeError sends a JSON error.
//
// The messages here are deliberately dull. "authentication required" is
// returned for a missing cookie, an expired session and a disabled account
// alike: the client gets a yes or a no, and the operator gets the reason in the
// log.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decodeJSON reads a request body into v, rejecting anything surprising.
//
// DisallowUnknownFields is on purpose: a field the server does not understand
// means the client and the server disagree about the request, and on a legal
// record guessing which of them is right is not acceptable. It also catches a
// typo in a field name that would otherwise silently do nothing.
func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("malformed request body: %w", err)
	}
	// Exactly one JSON value, so a second object smuggled after the first
	// cannot be interpreted by anything downstream.
	if dec.More() {
		return fmt.Errorf("malformed request body: unexpected trailing data")
	}
	return nil
}
