package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/store"
)

// --- Startup guards ---------------------------------------------------------

// TestRunRefusesAnEmptyDatabase. Serving an empty logbook would present "0:00
// total time" as though it were the record, which is exactly the silent
// corruption CLAUDE.md rule 0.2 forbids. Refusing to start is the honest
// failure.
func TestRunRefusesAnEmptyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	db.Close()

	err = run([]string{"-db", path, "-addr", "127.0.0.1:0"})
	if err == nil {
		t.Fatal("the server started on an empty database")
	}
	if !strings.Contains(err.Error(), "no flights") {
		t.Errorf("error = %v, want it to name the empty logbook", err)
	}
}

// TestRunRefusesAnInsecureCookieOverHTTPS. The flag exists for plain-HTTP local
// development. Combined with an https origin it can only be a mistake, and the
// consequence -- a session cookie that will travel over plain HTTP -- is too
// expensive to allow through as a warning.
func TestRunRefusesAnInsecureCookieOverHTTPS(t *testing.T) {
	path := populatedDB(t)
	err := run([]string{"-db", path, "-addr", "127.0.0.1:0", "-insecure-cookie",
		"-origin", "https://ayoub.fi"})
	if err == nil {
		t.Fatal("the server started with -insecure-cookie on an https origin")
	}
	if !strings.Contains(err.Error(), "insecure-cookie") {
		t.Errorf("error = %v, want it to name the flag", err)
	}
}

func TestRunRejectsUnknownFlags(t *testing.T) {
	if err := run([]string{"-nonsense"}); err == nil {
		t.Error("an unknown flag was accepted")
	}
}

// TestASubcommandHonoursTheDatabaseFlagOnEitherSide. The first version parsed
// no flags at all for subcommands, so `server createuser rami -db /tmp/x.db`
// silently ignored -db and reached for the production database. A smoke test
// caught it; this keeps it caught.
//
// Both orders must work, because Go's flag package stops parsing at the first
// non-flag argument and an operator writes the subcommand wherever it reads
// naturally.
func TestASubcommandHonoursTheDatabaseFlagOnEitherSide(t *testing.T) {
	for _, c := range []struct {
		name string
		args func(path string) []string
	}{
		{"flag after the subcommand", func(p string) []string { return []string{"users", "-db", p} }},
		{"flag before the subcommand", func(p string) []string { return []string{"-db", p, "users"} }},
		// The order that actually broke: a positional between the subcommand
		// and the flag, which makes flag.Parse stop before ever seeing -db.
		{"flag after a positional", func(p string) []string {
			return []string{"enable", "rami", "-db", p}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			path := populatedDB(t)
			db, err := store.Open(path)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			if _, err := db.CreateUser("rami", testPassword); err != nil {
				t.Fatalf("CreateUser: %v", err)
			}
			db.Close()

			// If -db were ignored this would reach for
			// /var/lib/logbook/logbook.db, which does not exist here.
			if err := run(c.args(path)); err != nil {
				t.Errorf("run(%v) = %v; the -db flag was not honoured", c.args(path), err)
			}
		})
	}
}

func TestAnUnknownSubcommandIsAnError(t *testing.T) {
	if err := run([]string{"-db", populatedDB(t), "frobnicate"}); err == nil {
		t.Error("an unknown subcommand was accepted")
	}
}

func TestEnvFallsBackWhenUnset(t *testing.T) {
	if got := env("LOGBOOK_TEST_UNSET_VAR", "fallback"); got != "fallback" {
		t.Errorf("env unset = %q, want the fallback", got)
	}
	t.Setenv("LOGBOOK_TEST_SET_VAR", "value")
	if got := env("LOGBOOK_TEST_SET_VAR", "fallback"); got != "value" {
		t.Errorf("env set = %q, want the value", got)
	}
	// An empty variable is treated as unset: `LOGBOOK_ADDR= ./server` must not
	// silently bind to nothing.
	t.Setenv("LOGBOOK_TEST_EMPTY_VAR", "")
	if got := env("LOGBOOK_TEST_EMPTY_VAR", "fallback"); got != "fallback" {
		t.Errorf("env empty = %q, want the fallback", got)
	}
}

// --- Operator subcommands ---------------------------------------------------

func TestSubcommands(t *testing.T) {
	db := openPopulated(t)
	if _, err := db.CreateUser("rami", testPassword); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	t.Run("users", func(t *testing.T) {
		if err := runSubcommand("users", nil, db); err != nil {
			t.Errorf("users: %v", err)
		}
	})

	t.Run("disable revokes every session", func(t *testing.T) {
		u, err := db.Authenticate("rami", testPassword)
		if err != nil {
			t.Fatalf("Authenticate: %v", err)
		}
		raw, _, err := db.CreateSession(u.ID, "phone", "192.0.2.1")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}
		if err := runSubcommand("disable", []string{"rami"}, db); err != nil {
			t.Fatalf("disable: %v", err)
		}
		if _, _, err := db.LookupSession(raw); err == nil {
			t.Error("disabling the account left a live session behind")
		}
		if _, err := db.Authenticate("rami", testPassword); err == nil {
			t.Error("a disabled account can still log in")
		}
	})

	t.Run("enable", func(t *testing.T) {
		if err := runSubcommand("enable", []string{"rami"}, db); err != nil {
			t.Fatalf("enable: %v", err)
		}
		if _, err := db.Authenticate("rami", testPassword); err != nil {
			t.Errorf("a re-enabled account cannot log in: %v", err)
		}
	})

	t.Run("errors", func(t *testing.T) {
		for _, c := range []struct {
			name string
			cmd  string
			args []string
		}{
			{"unknown subcommand", "frobnicate", nil},
			{"no username", "disable", nil},
			{"empty username", "disable", []string{""}},
			{"too many arguments", "enable", []string{"a", "b"}},
			{"no such user", "disable", []string{"nobody"}},
			{"enable a missing user", "enable", []string{"nobody"}},
		} {
			t.Run(c.name, func(t *testing.T) {
				if err := runSubcommand(c.cmd, c.args, db); err == nil {
					t.Error("no error")
				}
			})
		}
	})

	// createuser and passwd both read the password from the terminal. Under
	// `go test` there is no terminal, so they must refuse rather than read a
	// password from a pipe -- a piped password lands in the shell history and
	// the process list of a box shared with other people's sites.
	t.Run("password commands refuse a non-terminal", func(t *testing.T) {
		for _, cmd := range []string{"createuser", "passwd"} {
			err := runSubcommand(cmd, []string{"someone"}, db)
			if err == nil {
				t.Fatalf("%s read a password from a non-terminal", cmd)
			}
			if !strings.Contains(err.Error(), "terminal") {
				t.Errorf("%s error = %v, want it to explain the terminal requirement", cmd, err)
			}
		}
	})
}

func TestRevokeAllForAnUnknownUser(t *testing.T) {
	db := openPopulated(t)
	if _, err := revokeAllFor(db, "nobody"); err == nil {
		t.Error("revoking sessions for a user that does not exist reported success")
	}
}

// --- Housekeeping -----------------------------------------------------------

func TestHousekeepingStopsWhenTheContextIsCancelled(t *testing.T) {
	// A goroutine that outlives shutdown would hold the database open past
	// Close and turn a clean stop into a crash on the next start.
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		h.housekeeping(ctx)
		close(done)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("housekeeping did not stop when its context was cancelled")
	}
}

// --- clientIP ---------------------------------------------------------------

func TestClientIP(t *testing.T) {
	for _, c := range []struct {
		name, forwarded, remote, want string
	}{
		{"direct", "", "192.0.2.9:54321", "192.0.2.9"},
		{"behind the proxy", "198.51.100.7", "127.0.0.1:41234", "198.51.100.7"},
		{"a chain of proxies", "198.51.100.7, 10.0.0.1, 10.0.0.2", "127.0.0.1:41234", "198.51.100.7"},
		{"spaces are trimmed", "  198.51.100.7  ,10.0.0.1", "127.0.0.1:41234", "198.51.100.7"},
		{"an unsplittable remote address", "", "not-a-host-port", "not-a-host-port"},
		{"IPv6", "", "[2001:db8::1]:443", "2001:db8::1"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = c.remote
			if c.forwarded != "" {
				r.Header.Set("X-Forwarded-For", c.forwarded)
			}
			if got := clientIP(r); got != c.want {
				t.Errorf("clientIP = %q, want %q", got, c.want)
			}
		})
	}
}

// --- Failure paths ----------------------------------------------------------

// TestAReadFailureIs500AndNotAnEmptyLogbook. If a query fails, the handler must
// say so. The failure mode that matters is not a 500 -- it is a 200 carrying an
// empty list, which reads as "you have flown nothing" and is exactly the silent
// corruption of a legal record that rule 0.2 forbids.
//
// The table under test is dropped while the auth tables are left intact, so the
// request authenticates normally and fails where the handler reads the logbook.
func TestAReadFailureIs500AndNotAnEmptyLogbook(t *testing.T) {
	for _, c := range []struct{ table, path string }{
		{"flights", "/logbook/api/flights"},
		{"flights", "/logbook/api/stats"},
		{"aircraft", "/logbook/api/aircraft"},
		{"discrepancies", "/logbook/api/discrepancies"},
	} {
		t.Run(c.path, func(t *testing.T) {
			h := newHarness(t)
			auth := h.login()
			if _, err := h.db.SQLForTest().Exec("DROP TABLE " + c.table); err != nil {
				t.Fatalf("dropping %s: %v", c.table, err)
			}

			w := h.do("GET", c.path, "", auth)
			if w.Code != http.StatusInternalServerError {
				t.Errorf("status %d, want 500 -- an empty 200 would read as an empty logbook", w.Code)
			}
			// The message must not hand a stranger the schema or the driver.
			// Domain nouns are fine -- "could not read the aircraft list" says
			// nothing a caller of /aircraft did not already know.
			for _, leak := range []string{"sql", "sqlite", "select", "no such table"} {
				if strings.Contains(strings.ToLower(w.Body.String()), leak) {
					t.Errorf("the response leaks database detail (%q): %s", leak, w.Body.String())
				}
			}
		})
	}
}

// TestASessionListFailureIs500 covers the same rule for the sessions endpoint,
// which cannot be tested by dropping its own table -- authentication needs it.
// The row is made unreadable instead.
func TestASessionListFailureIs500(t *testing.T) {
	h := newHarness(t)
	auth := h.login()
	if _, err := h.db.SQLForTest().Exec(`UPDATE sessions SET created_at = 'not a timestamp'`); err != nil {
		t.Fatalf("corrupting the session row: %v", err)
	}
	w := h.do("GET", "/logbook/api/sessions", "", auth)
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 500 or 401", w.Code)
	}
}

// TestABlankBlockTimeSerialisesAsNull. A missing cell in the paper book must
// read as absent, not as midnight on 1 January year 1 -- which is what a zero
// time.Time renders as and what someone would eventually read as a real time.
func TestABlankBlockTimeSerialisesAsNull(t *testing.T) {
	h := newHarness(t)

	lb := sampleLogbook(t)
	lb.Flights[0].OffBlockUTC = time.Time{}
	lb.Flights[0].OnBlockUTC = time.Time{}
	lb.Flights[0].OffBlockRaw = ""
	lb.Flights[0].OnBlockRaw = ""
	lb.Flights[0].TimeOrigin = "none"
	if _, err := h.db.Import(lb, "blank block times"); err != nil {
		t.Fatalf("Import: %v", err)
	}

	auth := h.login()
	w := h.do("GET", "/logbook/api/flights", "", auth)
	var got struct {
		Flights []struct {
			Seq         int              `json:"seq"`
			OffBlockUTC *json.RawMessage `json:"off_block_utc"`
			TimeOrigin  string           `json:"time_origin"`
		} `json:"flights"`
	}
	decode(t, w, &got)
	if len(got.Flights) == 0 {
		t.Fatal("no flights returned")
	}
	if got.Flights[0].OffBlockUTC != nil {
		t.Errorf("a blank off-block serialised as %s, want null", *got.Flights[0].OffBlockUTC)
	}
	// The origin flag must survive so the row can be surfaced for review.
	if got.Flights[0].TimeOrigin != "none" {
		t.Errorf("time_origin = %q, want it preserved", got.Flights[0].TimeOrigin)
	}
}

// TestCallerOfPanicsWithoutASession is the backstop behind default deny: if a
// handler were ever mounted without requireSession, it must crash loudly rather
// than serve someone else's logbook to a zero-value user.
func TestCallerOfPanicsWithoutASession(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("callerOf returned normally with no session on the context")
		}
	}()
	callerOf(httptest.NewRequest("GET", "/", nil))
}

// --- helpers ----------------------------------------------------------------

func populatedDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "logbook.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if _, err := db.Import(sampleLogbook(t), "test"); err != nil {
		t.Fatalf("Import: %v", err)
	}
	return path
}

func openPopulated(t *testing.T) *store.DB {
	t.Helper()
	db, err := store.Open(populatedDB(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
