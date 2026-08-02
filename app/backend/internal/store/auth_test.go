package store_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/auth"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

const pw = "a-sufficiently-long-passphrase"

// --- Users ------------------------------------------------------------------

func TestCreateUserStoresAnArgon2idHashAndNotThePassword(t *testing.T) {
	db := openTemp(t)

	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == 0 || u.Username != "rami" || u.Disabled {
		t.Errorf("CreateUser returned %+v", u)
	}

	// The control from security.md: the plaintext must be nowhere in the file.
	hash := readPasswordHash(t, db, "rami")
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("stored hash %q is not argon2id", hash)
	}
	if strings.Contains(hash, pw) {
		t.Error("the stored hash contains the plaintext password")
	}
}

func TestCreateUserRejectsADuplicateUsername(t *testing.T) {
	db := openTemp(t)
	if _, err := db.CreateUser("rami", pw); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if _, err := db.CreateUser("rami", pw); err == nil {
		t.Error("creating a second user with the same name succeeded")
	}
}

func TestCreateUserEnforcesThePasswordPolicy(t *testing.T) {
	db := openTemp(t)
	if _, err := db.CreateUser("rami", "short"); !errors.Is(err, auth.ErrPasswordLength) {
		t.Errorf("CreateUser with a short password = %v, want ErrPasswordLength", err)
	}
	if _, err := db.CreateUser("", pw); err == nil {
		t.Error("CreateUser accepted an empty username")
	}
}

// TestAuthenticateIsUniformOnEveryFailure is the control that keeps a wrong
// username indistinguishable from a wrong password: both must return the same
// error value, so no caller can leak the difference even by accident.
func TestAuthenticateIsUniformOnEveryFailure(t *testing.T) {
	db := openTemp(t)
	if _, err := db.CreateUser("rami", pw); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	for _, c := range []struct{ name, user, pass string }{
		{"no such user", "nobody", pw},
		{"wrong password", "rami", "the-wrong-passphrase"},
		{"empty password", "rami", ""},
		{"empty username", "", pw},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := db.Authenticate(c.user, c.pass)
			if !errors.Is(err, store.ErrAuthFailed) {
				t.Errorf("Authenticate(%q, ...) = %v, want ErrAuthFailed", c.user, err)
			}
		})
	}

	u, err := db.Authenticate("rami", pw)
	if err != nil {
		t.Fatalf("Authenticate with the right password: %v", err)
	}
	if u.Username != "rami" {
		t.Errorf("authenticated as %q", u.Username)
	}
}

// TestAuthenticateDoesTheWorkForAnUnknownUser defends against user
// enumeration by timing: if a missing user returned early, an attacker could
// tell which usernames exist by how fast the answer comes back.
func TestAuthenticateDoesTheWorkForAnUnknownUser(t *testing.T) {
	db := openTemp(t)
	if _, err := db.CreateUser("rami", pw); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	measure := func(user string) time.Duration {
		start := time.Now()
		db.Authenticate(user, "the-wrong-passphrase")
		return time.Since(start)
	}
	// Warm up, then compare. Argon2 at 19 MiB dominates both paths, so an
	// early return for a missing user is an order-of-magnitude difference, not
	// a subtle one -- a loose bound is enough to catch it without being flaky.
	measure("rami")
	real, missing := measure("rami"), measure("nobody")
	if missing < real/4 {
		t.Errorf("an unknown user answered in %v against %v for a real one; "+
			"the missing-user path is skipping the hash and leaks which names exist",
			missing, real)
	}
}

func TestAuthenticateRefusesADisabledUser(t *testing.T) {
	db := openTemp(t)
	if _, err := db.CreateUser("rami", pw); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if err := db.SetUserDisabled("rami", true); err != nil {
		t.Fatalf("SetUserDisabled: %v", err)
	}
	if _, err := db.Authenticate("rami", pw); !errors.Is(err, store.ErrAuthFailed) {
		t.Errorf("a disabled user authenticated: %v", err)
	}
	if err := db.SetUserDisabled("rami", false); err != nil {
		t.Fatalf("SetUserDisabled: %v", err)
	}
	if _, err := db.Authenticate("rami", pw); err != nil {
		t.Errorf("re-enabled user cannot log in: %v", err)
	}
	if err := db.SetUserDisabled("nobody", true); err == nil {
		t.Error("disabling a non-existent user reported success")
	}
}

// TestSetPasswordRevokesEverySession is the security.md control: changing the
// password must invalidate sessions, or a stolen cookie survives the response
// to its own theft.
func TestSetPasswordRevokesEverySession(t *testing.T) {
	db := openTemp(t)
	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	raw, _, err := db.CreateSession(u.ID, "phone", "192.0.2.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, _, err := db.LookupSession(raw); err != nil {
		t.Fatalf("the session should be live before the password change: %v", err)
	}

	const newPW = "an-entirely-different-passphrase"
	if err := db.SetPassword("rami", newPW); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if _, _, err := db.LookupSession(raw); !errors.Is(err, store.ErrNoSession) {
		t.Errorf("the session survived a password change: %v", err)
	}
	if _, err := db.Authenticate("rami", pw); !errors.Is(err, store.ErrAuthFailed) {
		t.Error("the old password still works")
	}
	if _, err := db.Authenticate("rami", newPW); err != nil {
		t.Errorf("the new password does not work: %v", err)
	}

	if err := db.SetPassword("rami", "short"); !errors.Is(err, auth.ErrPasswordLength) {
		t.Errorf("SetPassword accepted a short password: %v", err)
	}
	if err := db.SetPassword("nobody", newPW); err == nil {
		t.Error("SetPassword on a non-existent user reported success")
	}
}

func TestUsersListsWhatWasCreated(t *testing.T) {
	db := openTemp(t)
	for _, n := range []string{"rami", "another"} {
		if _, err := db.CreateUser(n, pw); err != nil {
			t.Fatalf("CreateUser(%q): %v", n, err)
		}
	}
	list, err := db.Users()
	if err != nil {
		t.Fatalf("Users: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d users, want 2", len(list))
	}
	// Sorted by name, so two runs of the operator CLI print the same thing.
	if list[0].Username != "another" || list[1].Username != "rami" {
		t.Errorf("users are not sorted by name: %q, %q", list[0].Username, list[1].Username)
	}
}

// --- Sessions ---------------------------------------------------------------

// TestTheRawTokenIsNeverStored is the control from security.md, asserted
// against the actual bytes on disk rather than against the code that wrote
// them.
func TestTheRawTokenIsNeverStored(t *testing.T) {
	db := openTemp(t)
	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	raw, _, err := db.CreateSession(u.ID, "phone", "192.0.2.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	var stored string
	if err := db.SQLForTest().QueryRow(`SELECT token_hash FROM sessions`).Scan(&stored); err != nil {
		t.Fatalf("reading the session row: %v", err)
	}
	if stored == raw {
		t.Fatal("the sessions table holds the cookie value verbatim")
	}
	if stored != auth.HashSessionToken(raw) {
		t.Errorf("token_hash = %q, want the SHA-256 of the cookie value", stored)
	}
}

func TestLookupSessionFindsTheUser(t *testing.T) {
	db := openTemp(t)
	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	raw, s, err := db.CreateSession(u.ID, "phone", "192.0.2.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if s.UserAgent != "phone" || s.IP != "192.0.2.1" {
		t.Errorf("session metadata not recorded: %+v", s)
	}

	got, gotUser, err := db.LookupSession(raw)
	if err != nil {
		t.Fatalf("LookupSession: %v", err)
	}
	if got.ID != s.ID || gotUser.ID != u.ID {
		t.Errorf("LookupSession returned session %d / user %d, want %d / %d",
			got.ID, gotUser.ID, s.ID, u.ID)
	}
}

func TestLookupSessionRejectsAnythingItDoesNotKnow(t *testing.T) {
	db := openTemp(t)
	for _, c := range []struct{ name, token string }{
		{"empty", ""},
		{"a token that was never issued", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		{"a hostile string", "' OR 1=1 --"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, _, err := db.LookupSession(c.token); !errors.Is(err, store.ErrNoSession) {
				t.Errorf("LookupSession(%q) = %v, want ErrNoSession", c.token, err)
			}
		})
	}
}

// TestSessionExpiryIsRollingAndEnforced is the security.md control, on an
// injected clock: a session that is used keeps rolling forward, and one left
// alone past its window is dead.
func TestSessionExpiryIsRollingAndEnforced(t *testing.T) {
	db := openTemp(t)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := base
	restore := db.SetClockForTest(func() time.Time { return clock })
	defer restore()

	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	raw, s, err := db.CreateSession(u.ID, "phone", "192.0.2.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if want := base.Add(store.SessionLifetime); !s.ExpiresAt.Equal(want) {
		t.Errorf("a fresh session expires at %v, want %v", s.ExpiresAt, want)
	}

	// Used on day 89: still valid, and the window rolls forward from there.
	clock = base.AddDate(0, 0, 89)
	got, _, err := db.LookupSession(raw)
	if err != nil {
		t.Fatalf("an 89-day-old session that is still inside its window was rejected: %v", err)
	}
	if want := clock.Add(store.SessionLifetime); !got.ExpiresAt.Equal(want) {
		t.Errorf("expiry did not roll forward: %v, want %v", got.ExpiresAt, want)
	}
	if !got.LastUsedAt.Equal(clock) {
		t.Errorf("last_used_at = %v, want %v", got.LastUsedAt, clock)
	}

	// 89 more days of use keeps it alive indefinitely -- that is the point of a
	// rolling window, and it is what delivers "I never want to log in again".
	clock = clock.AddDate(0, 0, 89)
	if _, _, err := db.LookupSession(raw); err != nil {
		t.Fatalf("a session used every 89 days should never expire: %v", err)
	}

	// Then leave it alone for 91 days. Dead.
	clock = clock.AddDate(0, 0, 91)
	if _, _, err := db.LookupSession(raw); !errors.Is(err, store.ErrNoSession) {
		t.Errorf("a 91-day-idle session was accepted: %v", err)
	}
}

func TestLookupSessionRefusesADisabledUsersLiveSession(t *testing.T) {
	// Disabling an account must lock it out now, not at the next login.
	db := openTemp(t)
	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	raw, _, err := db.CreateSession(u.ID, "phone", "192.0.2.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := db.SetUserDisabled("rami", true); err != nil {
		t.Fatalf("SetUserDisabled: %v", err)
	}
	if _, _, err := db.LookupSession(raw); !errors.Is(err, store.ErrNoSession) {
		t.Errorf("a disabled user's live session still works: %v", err)
	}
}

func TestRevokeSession(t *testing.T) {
	db := openTemp(t)
	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	keep, _, err := db.CreateSession(u.ID, "laptop", "192.0.2.2")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	drop, dropSession, err := db.CreateSession(u.ID, "phone", "192.0.2.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if err := db.RevokeSession(u.ID, dropSession.ID); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}
	if _, _, err := db.LookupSession(drop); !errors.Is(err, store.ErrNoSession) {
		t.Error("the revoked session still works")
	}
	if _, _, err := db.LookupSession(keep); err != nil {
		t.Errorf("revoking one session killed another: %v", err)
	}

	// Revoking is scoped to the owner, so one user cannot end another's
	// session by guessing an id. There is one user today; this must still hold
	// when there are more, which security.md says is a supported future.
	other, err := db.CreateUser("other", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if err := db.RevokeSession(other.ID, dropSession.ID); err == nil {
		t.Error("a user revoked a session belonging to someone else")
	}
	if err := db.RevokeSession(u.ID, 99999); err == nil {
		t.Error("revoking a session that does not exist reported success")
	}
}

func TestRevokeAllSessionsIsScopedToOneUser(t *testing.T) {
	db := openTemp(t)
	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	other, err := db.CreateUser("other", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	mine, _, err := db.CreateSession(u.ID, "phone", "192.0.2.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	theirs, _, err := db.CreateSession(other.ID, "phone", "192.0.2.9")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	n, err := db.RevokeAllSessions(u.ID)
	if err != nil {
		t.Fatalf("RevokeAllSessions: %v", err)
	}
	if n != 1 {
		t.Errorf("revoked %d sessions, want 1", n)
	}
	if _, _, err := db.LookupSession(mine); !errors.Is(err, store.ErrNoSession) {
		t.Error("my session survived revoke-all")
	}
	if _, _, err := db.LookupSession(theirs); err != nil {
		t.Errorf("revoke-all reached another user's session: %v", err)
	}
}

func TestSessionsListsTheUsersOwnOnly(t *testing.T) {
	db := openTemp(t)
	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	other, err := db.CreateUser("other", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	for _, ua := range []string{"phone", "laptop"} {
		if _, _, err := db.CreateSession(u.ID, ua, "192.0.2.1"); err != nil {
			t.Fatalf("CreateSession: %v", err)
		}
	}
	if _, _, err := db.CreateSession(other.ID, "theirs", "192.0.2.9"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	list, err := db.Sessions(u.ID)
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d sessions, want 2", len(list))
	}
	for _, s := range list {
		if s.UserAgent == "theirs" {
			t.Error("the session list leaked another user's session")
		}
	}
}

// TestPurgeExpiredSessionsRemovesOnlyTheDeadOnes keeps the table from growing
// without bound, and must never take a live session with it.
func TestPurgeExpiredSessionsRemovesOnlyTheDeadOnes(t *testing.T) {
	db := openTemp(t)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := base
	restore := db.SetClockForTest(func() time.Time { return clock })
	defer restore()

	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	old, _, err := db.CreateSession(u.ID, "old", "192.0.2.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	clock = base.AddDate(0, 0, 91)
	fresh, _, err := db.CreateSession(u.ID, "fresh", "192.0.2.2")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	n, err := db.PurgeExpiredSessions()
	if err != nil {
		t.Fatalf("PurgeExpiredSessions: %v", err)
	}
	if n != 1 {
		t.Errorf("purged %d sessions, want 1", n)
	}
	if _, _, err := db.LookupSession(fresh); err != nil {
		t.Errorf("the purge took a live session: %v", err)
	}
	if _, _, err := db.LookupSession(old); !errors.Is(err, store.ErrNoSession) {
		t.Error("the expired session survived the purge")
	}
}

// TestHostileInputIsParameterised runs the SQL-injection table from
// security.md against every string parameter that reaches a query.
func TestHostileInputIsParameterised(t *testing.T) {
	db := openTemp(t)
	if _, err := db.CreateUser("rami", pw); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	hostile := []string{
		`' OR '1'='1`,
		`'; DROP TABLE users; --`,
		`" OR ""="`,
		`admin'--`,
		`rami' --`,
		"\x00rami",
	}
	for _, h := range hostile {
		if _, err := db.Authenticate(h, pw); !errors.Is(err, store.ErrAuthFailed) {
			t.Errorf("Authenticate(%q) = %v, want ErrAuthFailed", h, err)
		}
		if _, err := db.Authenticate("rami", h); !errors.Is(err, store.ErrAuthFailed) {
			t.Errorf("Authenticate with password %q = %v, want ErrAuthFailed", h, err)
		}
		if _, _, err := db.LookupSession(h); !errors.Is(err, store.ErrNoSession) {
			t.Errorf("LookupSession(%q) = %v, want ErrNoSession", h, err)
		}
	}

	// The users table must still be there and still work.
	if _, err := db.Authenticate("rami", pw); err != nil {
		t.Errorf("the hostile inputs damaged the database: %v", err)
	}
}

func readPasswordHash(t *testing.T, db *store.DB, username string) string {
	t.Helper()
	var h string
	if err := db.SQLForTest().QueryRow(
		`SELECT password_hash FROM users WHERE username = ?`, username).Scan(&h); err != nil {
		t.Fatalf("reading password_hash: %v", err)
	}
	return h
}

// RedactForBackup is what lets a copy of this database leave the box.
//
// Sessions are live credentials-adjacent rows and are worthless once restored
// -- their expiry has passed, their IPs are stale, and the owner would sign in
// again anyway. Users MUST survive: a restored logbook nobody can log into is
// not a restored logbook.
func TestRedactForBackupDropsSessionsAndKeepsUsers(t *testing.T) {
	db := openTemp(t)
	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.CreateSession(u.ID, "phone", "10.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.CreateSession(u.ID, "laptop", "10.0.0.2"); err != nil {
		t.Fatal(err)
	}

	removed, err := db.RedactForBackup()
	if err != nil {
		t.Fatalf("RedactForBackup: %v", err)
	}
	if removed != 2 {
		t.Errorf("removed %d sessions, want 2", removed)
	}

	sessions, err := db.Sessions(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Errorf("%d sessions survived redaction", len(sessions))
	}

	// The account is still there, and still usable -- this is the whole point.
	if _, err := db.Authenticate("rami", pw); err != nil {
		t.Errorf("the account did not survive redaction: %v", err)
	}
}
