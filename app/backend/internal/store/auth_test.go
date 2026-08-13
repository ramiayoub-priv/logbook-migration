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

	// Expressed against the constant rather than a number of days, so that
	// changing the window (90 days -> 14 on 2026-08-03) changes what this test
	// exercises rather than breaking it. What is being asserted is the ROLLING,
	// which is a property; the length of the window is a ruling, and it is
	// pinned by TestSessionLifetimeIsTheOwnersRuling next door.
	const almost = store.SessionLifetime - 24*time.Hour

	// Used just inside the window: still valid, and it rolls forward from there.
	clock = base.Add(almost)
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

	// Another near-window gap keeps it alive indefinitely -- that is the point
	// of a rolling window, and it is what keeps the owner from logging in on
	// every visit.
	clock = clock.Add(almost)
	if _, _, err := db.LookupSession(raw); err != nil {
		t.Fatalf("a session used inside its window every time should never expire: %v", err)
	}

	// Then leave it alone for longer than the window. Dead.
	clock = clock.Add(store.SessionLifetime + time.Hour)
	if _, _, err := db.LookupSession(raw); !errors.Is(err, store.ErrNoSession) {
		t.Errorf("an idle session past its window was accepted: %v", err)
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

// --- The window, and the rows a longer one left behind (Task 20) -------------
//
// These need to write a session row that current code would never produce --
// one carrying the 90-day expires_at that shipped until 2026-08-03 -- and that
// is the whole point of them. The owner's real database has thirteen such rows,
// and a fix that only reached sessions created after the deploy would leave
// every one of them in the Devices list for another three months, which is the
// bug that was reported.

// storedTimeFormat is how store writes an instant. Spelled out here because
// these tests write rows directly.
const storedTimeFormat = "2006-01-02T15:04:05Z"

// TestSessionLifetimeIsTheOwnersRuling pins the number, because it is a ruling
// rather than a tuning knob.
//
// 2026-08-03: the cookie carries no Max-Age, so it dies when the phone's
// browser restarts -- the owner sees that re-login and calls it correct. The
// row then had 90 idle days left to live, unreachable by any device, and every
// login made another. Offered a persistent cookie instead, the owner chose to
// keep logging in and shorten the window. Raising this back towards the
// browser's own lifetime is what re-opens the bug.
func TestSessionLifetimeIsTheOwnersRuling(t *testing.T) {
	if want := 14 * 24 * time.Hour; store.SessionLifetime != want {
		t.Errorf("store.SessionLifetime = %v, want %v (owner ruling 2026-08-03)", store.SessionLifetime, want)
	}
}

// TestShorteningTheWindowReachesSessionsAlreadyStored is the fix for the rows
// that already exist.
//
// The window is evaluated from last_used_at against the CONSTANT, never from
// the expires_at written into the row when it was created. So changing the
// constant changes every session's fate at once, including the ones a previous
// version of this code stamped with a date three months out.
func TestShorteningTheWindowReachesSessionsAlreadyStored(t *testing.T) {
	db := openTemp(t)
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	restore := db.SetClockForTest(func() time.Time { return now })
	defer restore()

	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// A row exactly as the 90-day code left it: last used 20 days ago, stamped
	// to expire 70 days from now. The phone that made it dropped the cookie at
	// its next restart and can never present this token again.
	raw, hash, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken: %v", err)
	}
	lastUsed := now.AddDate(0, 0, -20)
	legacyExpiry := lastUsed.Add(90 * 24 * time.Hour)
	if _, err := db.SQLForTest().Exec(
		`INSERT INTO sessions (user_id, token_hash, created_at, last_used_at, expires_at, user_agent, ip)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.ID, hash, lastUsed.Format(storedTimeFormat), lastUsed.Format(storedTimeFormat),
		legacyExpiry.Format(storedTimeFormat), "an old iPhone", "192.0.2.1"); err != nil {
		t.Fatalf("inserting a legacy session: %v", err)
	}

	// Twenty days idle is past the fourteen-day window, whatever the row says.
	if _, _, err := db.LookupSession(raw); !errors.Is(err, store.ErrNoSession) {
		t.Errorf("a 20-day-idle session was accepted because its row claimed 70 days left: %v", err)
	}

	// And it is gone, rather than merely refused: LookupSession drops an
	// expired row on sight so the table is bounded even if the sweep never
	// runs. This is what empties the owner's Devices list on the first request
	// after the deploy.
	left, err := db.Sessions(u.ID)
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}
	if len(left) != 0 {
		t.Errorf("the dead session is still listed: %+v", left)
	}
}

// TestPurgeReadsTheWindowNotTheStoredExpiry is the same fix on the sweep that
// runs hourly, for rows nobody will ever present again -- which, since the
// browser dropped their cookies, is all of them.
func TestPurgeReadsTheWindowNotTheStoredExpiry(t *testing.T) {
	db := openTemp(t)
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	restore := db.SetClockForTest(func() time.Time { return now })
	defer restore()

	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	insert := func(idleDays int) {
		t.Helper()
		_, hash, err := auth.NewSessionToken()
		if err != nil {
			t.Fatalf("NewSessionToken: %v", err)
		}
		used := now.AddDate(0, 0, -idleDays)
		if _, err := db.SQLForTest().Exec(
			`INSERT INTO sessions (user_id, token_hash, created_at, last_used_at, expires_at, user_agent, ip)
			 VALUES (?, ?, ?, ?, ?, '', '')`,
			u.ID, hash, used.Format(storedTimeFormat), used.Format(storedTimeFormat),
			used.Add(90*24*time.Hour).Format(storedTimeFormat)); err != nil {
			t.Fatalf("inserting: %v", err)
		}
	}
	insert(20) // dead under the new window
	insert(30) // dead
	insert(13) // still inside it, by a day

	n, err := db.PurgeExpiredSessions()
	if err != nil {
		t.Fatalf("PurgeExpiredSessions: %v", err)
	}
	if n != 2 {
		t.Errorf("purged %d sessions, want 2 (the 20- and 30-day-idle ones)", n)
	}
	left, err := db.Sessions(u.ID)
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}
	if len(left) != 1 {
		t.Fatalf("%d sessions left, want 1", len(left))
	}
}

// TestSessionsReportTheWindowFromLastUse keeps the Devices page honest.
//
// A legacy row's stored expires_at is three months out and wrong. Reporting it
// would tell the owner a session is good until November when the server will
// refuse it in a fortnight -- and this page exists precisely so they can see
// what is live.
func TestSessionsReportTheWindowFromLastUse(t *testing.T) {
	db := openTemp(t)
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	restore := db.SetClockForTest(func() time.Time { return now })
	defer restore()

	u, err := db.CreateUser("rami", pw)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	_, hash, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken: %v", err)
	}
	used := now.AddDate(0, 0, -2)
	if _, err := db.SQLForTest().Exec(
		`INSERT INTO sessions (user_id, token_hash, created_at, last_used_at, expires_at, user_agent, ip)
		 VALUES (?, ?, ?, ?, ?, '', '')`,
		u.ID, hash, used.Format(storedTimeFormat), used.Format(storedTimeFormat),
		used.Add(90*24*time.Hour).Format(storedTimeFormat)); err != nil {
		t.Fatalf("inserting: %v", err)
	}

	list, err := db.Sessions(u.ID)
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("%d sessions, want 1", len(list))
	}
	if want := used.Add(store.SessionLifetime); !list[0].ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v (last use + the window, not the stored column)",
			list[0].ExpiresAt, want)
	}
}
