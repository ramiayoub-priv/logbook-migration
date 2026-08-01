package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/stats"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

// caller is who made an authenticated request. It is put on the context by
// requireSession and is the only way a handler learns who it is serving --
// there is no path by which a handler can read a user id straight off the wire.
type caller struct {
	User    store.User
	Session store.Session
}

type callerKey struct{}

func withCaller(ctx context.Context, c caller) context.Context {
	return context.WithValue(ctx, callerKey{}, c)
}

// callerOf returns the authenticated caller. It panics if there is none, which
// can only happen if a handler were mounted without requireSession -- a bug
// that must be loud rather than a nil user silently reading someone's logbook.
func callerOf(r *http.Request) caller {
	c, ok := r.Context().Value(callerKey{}).(caller)
	if !ok {
		panic("server: handler reached without a session; it was mounted without requireSession")
	}
	return c
}

// --- Authentication ---------------------------------------------------------

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "malformed request")
		return
	}

	// Two keys: the account and the source address. The account key stops
	// distributed guessing at one password; the address key stops one host
	// working through a list of usernames.
	ip := clientIP(r)
	keys := []string{"user:" + req.Username, "ip:" + ip}
	for _, k := range keys {
		if ok, retry := s.limiter.Allow(k); !ok {
			w.Header().Set("Retry-After", strconv.Itoa(int(retry.Round(time.Second).Seconds())))
			s.log.Warn("login throttled", "ip", ip, "retry_after", retry)
			writeError(w, http.StatusTooManyRequests, "too many attempts; try again later")
			return
		}
	}

	user, err := s.db.Authenticate(req.Username, req.Password)
	if err != nil {
		for _, k := range keys {
			s.limiter.Fail(k)
		}
		if !errors.Is(err, store.ErrAuthFailed) {
			// A corrupted password hash. The client still gets the ordinary
			// refusal; the operator needs to see this one.
			s.log.Error("authentication error", "username", req.Username, "error", err)
		} else {
			s.log.Info("login failed", "username", req.Username, "ip", ip)
		}
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	for _, k := range keys {
		s.limiter.Succeed(k)
	}

	raw, _, err := s.db.CreateSession(user.ID, r.UserAgent(), ip)
	if err != nil {
		s.log.Error("creating session", "error", err)
		writeError(w, http.StatusInternalServerError, "could not start a session")
		return
	}
	s.setCookie(w, raw)
	s.log.Info("login succeeded", "username", user.Username, "ip", ip)

	// The token goes in the cookie and nowhere else -- not in this body, where
	// it would end up in a log or in a second place in the browser's memory.
	writeJSON(w, http.StatusOK, map[string]any{
		"username": user.Username,
		"user_id":  user.ID,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	c := callerOf(r)
	if err := s.db.RevokeSession(c.User.ID, c.Session.ID); err != nil {
		// Already gone is a fine outcome for a logout; the caller wanted the
		// session not to exist and it does not.
		s.log.Info("logout on an already-revoked session", "user", c.User.Username)
	}
	s.clearCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	c := callerOf(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":  c.User.ID,
		"username": c.User.Username,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Public, so it discloses nothing: no version, no counts, no paths. It
	// answers "is the process up", which is all a health check may ask of an
	// endpoint that anyone on the internet can reach.
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Sessions ---------------------------------------------------------------

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	c := callerOf(r)
	list, err := s.db.Sessions(c.User.ID)
	if err != nil {
		s.log.Error("listing sessions", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list sessions")
		return
	}

	type item struct {
		ID         int64     `json:"id"`
		CreatedAt  time.Time `json:"created_at"`
		LastUsedAt time.Time `json:"last_used_at"`
		ExpiresAt  time.Time `json:"expires_at"`
		UserAgent  string    `json:"user_agent"`
		IP         string    `json:"ip"`
		// Which row is the device asking, so the UI can label it and warn
		// before revoking it.
		Current bool `json:"current"`
	}
	out := make([]item, 0, len(list))
	for _, sess := range list {
		out = append(out, item{
			ID: sess.ID, CreatedAt: sess.CreatedAt, LastUsedAt: sess.LastUsedAt,
			ExpiresAt: sess.ExpiresAt, UserAgent: sess.UserAgent, IP: sess.IP,
			Current: sess.ID == c.Session.ID,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": out})
}

func (s *Server) handleRevokeSession(w http.ResponseWriter, r *http.Request) {
	c := callerOf(r)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "session id must be a number")
		return
	}
	// Scoped to the caller in the query itself, so another user's session
	// cannot be ended by guessing its id.
	if err := s.db.RevokeSession(c.User.ID, id); err != nil {
		writeError(w, http.StatusNotFound, "no such session")
		return
	}
	if id == c.Session.ID {
		s.clearCookie(w)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// --- The logbook ------------------------------------------------------------

// dateParam matches YYYY-MM-DD and nothing else.
var dateParam = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// rangeOf reads the from/to query parameters.
//
// An unparseable date is an error rather than an ignored parameter. Silently
// dropping it would answer a question the user did not ask and hand back a
// total they might write into a licence application.
func rangeOf(r *http.Request) (stats.Range, error) {
	var rng stats.Range
	for _, p := range []struct {
		name string
		dst  *string
	}{{"from", &rng.From}, {"to", &rng.To}} {
		v := r.URL.Query().Get(p.name)
		if v == "" {
			continue
		}
		if !dateParam.MatchString(v) {
			return rng, errors.New(p.name + " must be a date as YYYY-MM-DD")
		}
		// The shape is right; check it is a real day. "2022-13-45" matches the
		// pattern and is not a date.
		if _, err := time.Parse("2006-01-02", v); err != nil {
			return rng, errors.New(p.name + " is not a real date")
		}
		*p.dst = v
	}
	return rng, nil
}

// flightsIn loads the flights in a range. The store returns them in seq order,
// which is the book's own order and the only one anything cumulative may use.
func (s *Server) flightsIn(rng stats.Range) ([]csvbook.Flight, error) {
	all, err := s.db.Flights()
	if err != nil {
		return nil, err
	}
	return stats.Filter(all, rng), nil
}

type flightJSON struct {
	Seq          int    `json:"seq"`
	Date         string `json:"date"`
	AircraftType string `json:"aircraft_type"`
	AircraftReg  string `json:"aircraft_reg"`
	Class        string `json:"class"`
	DepPlace     string `json:"dep_place"`
	ArrPlace     string `json:"arr_place"`

	// The canonical instants, and alongside them the strings exactly as
	// written on paper plus how the conversion was reached. Never just the
	// converted value: a bad DST guess must stay auditable (rule 0.4).
	OffBlockUTC *time.Time `json:"off_block_utc"`
	OnBlockUTC  *time.Time `json:"on_block_utc"`
	OffBlockRaw string     `json:"off_block_raw"`
	OnBlockRaw  string     `json:"on_block_raw"`
	TimeOrigin  string     `json:"time_origin"`

	// Minutes throughout. H:MM is a presentation concern; one representation
	// on the wire means no second figure can disagree with the first.
	BlockMinutes      int `json:"block_minutes"`
	TotalMinutes      int `json:"total_minutes"`
	NightMinutes      int `json:"night_minutes"`
	InstrumentMinutes int `json:"instrument_minutes"`
	PICMinutes        int `json:"pic_minutes"`
	DualMinutes       int `json:"dual_minutes"`
	InstructorMinutes int `json:"instructor_minutes"`

	PICName          string `json:"pic_name"`
	LandingsDay      int    `json:"landings_day"`
	LandingsNight    int    `json:"landings_night"`
	LandingsVerified bool   `json:"landings_verified"`
	Remarks          string `json:"remarks"`

	// Which CSV line, and so which page of which paper book, this row came
	// from. Every figure the app shows stays traceable to the paper.
	SourceBook int `json:"source_book"`
	SourceRow  int `json:"source_row"`
}

func toFlightJSON(f csvbook.Flight) flightJSON {
	return flightJSON{
		Seq: f.Seq, Date: f.Date,
		AircraftType: f.AircraftType, AircraftReg: f.AircraftReg, Class: string(f.Class),
		DepPlace: f.DepPlace, ArrPlace: f.ArrPlace,
		OffBlockUTC: nilIfZero(f.OffBlockUTC), OnBlockUTC: nilIfZero(f.OnBlockUTC),
		OffBlockRaw: f.OffBlockRaw, OnBlockRaw: f.OnBlockRaw, TimeOrigin: string(f.TimeOrigin),
		BlockMinutes: f.BlockMinutes, TotalMinutes: f.TotalMinutes,
		NightMinutes: f.NightMinutes, InstrumentMinutes: f.InstrumentMinutes,
		PICMinutes: f.PICMinutes, DualMinutes: f.DualMinutes,
		InstructorMinutes: f.InstructorMinutes,
		PICName:           f.PICName, LandingsDay: f.LandingsDay,
		LandingsNight: f.LandingsNight, LandingsVerified: f.LandingsVerified,
		Remarks:    f.Remarks,
		SourceBook: f.SourceBook, SourceRow: f.SourceRow,
	}
}

// nilIfZero maps a missing instant to JSON null rather than to year 1. A blank
// cell in the paper book must read as blank, not as a real time nobody flew.
func nilIfZero(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func (s *Server) handleFlights(w http.ResponseWriter, r *http.Request) {
	rng, err := rangeOf(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	flights, err := s.flightsIn(rng)
	if err != nil {
		s.log.Error("listing flights", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read the logbook")
		return
	}
	out := make([]flightJSON, 0, len(flights))
	for _, f := range flights {
		out = append(out, toFlightJSON(f))
	}
	writeJSON(w, http.StatusOK, map[string]any{"flights": out, "count": len(out)})
}

func (s *Server) handleAircraft(w http.ResponseWriter, r *http.Request) {
	list, err := s.db.Aircraft()
	if err != nil {
		s.log.Error("listing aircraft", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read the aircraft list")
		return
	}
	type item struct {
		Registration string `json:"registration"`
		Type         string `json:"type"`
		DefaultClass string `json:"default_class"`
		IFRCapable   bool   `json:"ifr_capable"`
		Active       bool   `json:"active"`
		Notes        string `json:"notes"`
	}
	out := make([]item, 0, len(list))
	for _, a := range list {
		out = append(out, item{
			Registration: a.Registration, Type: a.Type,
			DefaultClass: string(a.DefaultClass), IFRCapable: a.IFRCapable,
			Active: a.Active, Notes: a.Notes,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"aircraft": out})
}

// summaryJSON is the statistics page's payload: the twelve figures from
// app/APP.md section 2, plus the count of rows whose landing split is still
// inferred.
//
// Computed on every request from the rows, never stored (rule 0.5).
type summaryJSON struct {
	Flights int `json:"flights"`

	Total      int `json:"total"`
	PIC        int `json:"pic"`
	Dual       int `json:"dual"`
	Night      int `json:"night"`
	Instrument int `json:"instrument"`
	Instructor int `json:"instructor"`

	SeaTotal       int `json:"sea_total"`
	SeaPIC         int `json:"sea_pic"`
	SeaInstructor  int `json:"sea_instructor"`
	LandTotal      int `json:"land_total"`
	LandPIC        int `json:"land_pic"`
	LandInstructor int `json:"land_instructor"`

	LandingsDay   int `json:"landings_day"`
	LandingsNight int `json:"landings_night"`
	LandingsSea   int `json:"landings_sea"`
	LandingsLand  int `json:"landings_land"`

	// How many flights in this range still carry an inferred day/night landing
	// split. Reported rather than hidden: the page must be able to say that the
	// night landing figure is not yet read off the paper (Task 8).
	LandingsUnverified int `json:"landings_unverified"`
}

func toSummaryJSON(s stats.Summary) summaryJSON {
	return summaryJSON{
		Flights: s.Flights,
		Total:   s.Total, PIC: s.PIC, Dual: s.Dual, Night: s.Night,
		Instrument: s.Instrument, Instructor: s.Instructor,
		SeaTotal: s.SeaTotal, SeaPIC: s.SeaPIC, SeaInstructor: s.SeaInstructor,
		LandTotal: s.LandTotal, LandPIC: s.LandPIC, LandInstructor: s.LandInstructor,
		LandingsDay: s.LandingsDay, LandingsNight: s.LandingsNight,
		LandingsSea: s.LandingsSea, LandingsLand: s.LandingsLand,
		LandingsUnverified: s.LandingsUnverified,
	}
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	rng, err := rangeOf(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	flights, err := s.flightsIn(rng)
	if err != nil {
		s.log.Error("computing statistics", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read the logbook")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": toSummaryJSON(stats.Summarize(flights)),
		"range":   map[string]string{"from": rng.From, "to": rng.To},
	})
}

func (s *Server) handleDiscrepancies(w http.ResponseWriter, r *http.Request) {
	list, err := s.db.Discrepancies()
	if err != nil {
		s.log.Error("listing discrepancies", "error", err)
		writeError(w, http.StatusInternalServerError, "could not read the review list")
		return
	}
	type item struct {
		Kind   string `json:"kind"`
		Book   int    `json:"book"`
		Row    int    `json:"row"`
		Date   string `json:"date"`
		Detail string `json:"detail"`
	}
	out := make([]item, 0, len(list))
	for _, d := range list {
		out = append(out, item{
			Kind: string(d.Kind), Book: d.Book, Row: d.Row, Date: d.Date, Detail: d.Detail,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"discrepancies": out, "count": len(out)})
}

// --- Cookies ----------------------------------------------------------------

func (s *Server) setCookie(w http.ResponseWriter, raw string) {
	http.SetCookie(w, &http.Cookie{
		Name:  sessionCookieName,
		Value: raw,
		Path:  cookiePath,
		// No Expires or MaxAge: a session cookie in the browser, with the real
		// 90-day lifetime enforced server-side where it can be revoked. A
		// long-lived cookie the server has forgotten about is a credential
		// nobody can withdraw.
		HttpOnly: true,
		Secure:   s.cfg.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

// clientIP is the caller's address for logging and rate limiting.
//
// It reads X-Forwarded-For because the only way in is Apache's reverse proxy on
// this same box, so the header is set by us and cannot be spoofed from outside.
// If the server were ever exposed directly this would have to go -- noted here
// because it is the kind of assumption that silently stops being true.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// The left-most entry is the original client; the rest are proxies.
		first, _, _ := strings.Cut(fwd, ",")
		return strings.TrimSpace(first)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
