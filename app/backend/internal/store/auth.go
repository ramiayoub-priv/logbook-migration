// Users and sessions. The design and the reasoning are in
// app/docs/security.md; the cryptographic primitives are in internal/auth.
//
// The split is deliberate. internal/auth knows how to hash and how to mint a
// token but nothing about storage; this file knows about storage but never
// invents a credential rule of its own. Anything that decides whether a request
// is authenticated lives in exactly one of the two.
//
// Sessions are rows rather than JWTs for one reason: revocation. A signed token
// cannot be withdrawn before it expires, which would mean a fortnight's
// liability on a stolen cookie and no way to end it. A row can be deleted.
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/auth"
)

// SessionLifetime is the rolling window: a session dies this long after it was
// last used, not after it was created. That is what delivers the user's
// requirement of not having to log in on every visit, while still making an
// abandoned or stolen cookie expire on its own.
//
// IT IS 14 DAYS AND NOT 90, BY OWNER RULING (2026-08-03), AND THE REASON IS
// THE COOKIE. setCookie deliberately sends no Max-Age, so the cookie is a
// browser-session cookie that dies when the phone's browser or PWA restarts --
// the owner sees that re-login and considers it correct. With a 90-day row
// behind a cookie that lives a few days, every login left behind a session no
// device could ever present again: unreachable, but alive and listed on the
// Devices page for another three months. Thirteen of them had accumulated for
// one user.
//
// So the two lifetimes are now roughly matched, from the shorter end. Offered
// the alternative -- a persistent cookie, so the row and the device agree by
// staying logged in -- the owner chose to keep logging in. Raising this back
// towards 90 days without also giving the cookie a Max-Age re-opens the bug.
const SessionLifetime = 14 * 24 * time.Hour

// ErrAuthFailed is returned for every authentication failure, whatever the
// cause: no such user, wrong password, disabled account.
//
// One error value for all of them is the control, not laziness. A caller cannot
// accidentally tell an attacker which usernames exist if it is never given the
// distinction in the first place.
var ErrAuthFailed = errors.New("store: authentication failed")

// ErrNoSession means the cookie presented does not identify a usable session --
// unknown, expired, or belonging to a disabled account. Same reasoning as
// ErrAuthFailed: the caller gets a yes or a no, never a reason it could leak.
var ErrNoSession = errors.New("store: no such session")

// User is an account. It deliberately has no password field: the hash never
// leaves this file.
type User struct {
	ID        int64
	Username  string
	CreatedAt time.Time
	Disabled  bool
}

// Session is one logged-in device, as shown in the user's revocable session
// list. It carries no token: the raw value exists only in the cookie and the
// stored value is a hash of it.
type Session struct {
	ID         int64
	UserID     int64
	CreatedAt  time.Time
	LastUsedAt time.Time
	ExpiresAt  time.Time
	UserAgent  string
	IP         string
}

// CreateUser adds an account.
//
// There is no HTTP route to this. Accounts are made by the operator CLI on the
// server (security.md), so the self-service registration endpoint that would
// need defending simply does not exist.
func (db *DB) CreateUser(username, password string) (User, error) {
	if username == "" {
		return User{}, errors.New("store: username must not be empty")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return User{}, err
	}
	created := db.at()
	res, err := db.sql.Exec(
		`INSERT INTO users (username, password_hash, created_at, disabled) VALUES (?, ?, ?, 0)`,
		username, hash, created.Format(timeFormat))
	if err != nil {
		return User{}, fmt.Errorf("store: creating user %q: %w", username, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("store: creating user %q: %w", username, err)
	}
	return User{ID: id, Username: username, CreatedAt: created}, nil
}

// Authenticate checks a username and password, returning the user on success.
//
// Every failure path returns ErrAuthFailed and, crucially, every failure path
// still pays for one Argon2 hash. Returning early for an unknown username would
// answer in microseconds where a real account takes tens of milliseconds, which
// is a username oracle anybody can read with a stopwatch.
func (db *DB) Authenticate(username, password string) (User, error) {
	var (
		u        User
		hash     string
		created  string
		disabled int
	)
	err := db.sql.QueryRow(
		`SELECT id, username, password_hash, created_at, disabled FROM users WHERE username = ?`,
		username).Scan(&u.ID, &u.Username, &hash, &created, &disabled)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// No such user. Hash anyway, against a decoy of the same cost, so this
		// path takes as long as a real one. The result is discarded.
		auth.VerifyPassword(decoyHash, password)
		return User{}, ErrAuthFailed
	case err != nil:
		return User{}, fmt.Errorf("store: looking up user %q: %w", username, err)
	}

	ok, err := auth.VerifyPassword(hash, password)
	if err != nil {
		// An undecodable hash is database corruption. Deny, but say so: the
		// operator needs to know, even though the caller must not.
		return User{}, fmt.Errorf("store: stored hash for %q is unusable: %w", username, err)
	}
	if !ok || disabled != 0 {
		return User{}, ErrAuthFailed
	}
	if u.CreatedAt, err = parseStamp(created); err != nil {
		return User{}, err
	}
	return u, nil
}

// decoyHash is a real Argon2id hash at DefaultParams, used to give the
// unknown-user path the same cost as the real one. Its plaintext is irrelevant
// and unknown; it is never compared against anything that could succeed.
//
// Generated once at package init rather than checked in, so it cannot be
// mistaken for a credential and cannot go stale if DefaultParams is raised.
var decoyHash = mustDecoy()

func mustDecoy() string {
	h, err := auth.HashPasswordWith("decoy-not-a-credential", auth.DefaultParams)
	if err != nil {
		// Unreachable: the literal above satisfies the length policy and the
		// only other failure is an entropy source that cannot read, in which
		// case nothing else in this program works either.
		panic("store: cannot build the timing decoy: " + err.Error())
	}
	return h
}

// SetPassword changes a password and revokes every session the user has.
//
// The revocation is the point: a password change is what someone does when they
// think a credential is compromised, and leaving live sessions alive would make
// the change cosmetic. Both happen in one transaction.
func (db *DB) SetPassword(username, password string) error {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	tx, err := db.sql.Begin()
	if err != nil {
		return fmt.Errorf("store: setting password: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE users SET password_hash = ? WHERE username = ?`, hash, username)
	if err != nil {
		return fmt.Errorf("store: setting password for %q: %w", username, err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("store: setting password for %q: %w", username, err)
	} else if n == 0 {
		return fmt.Errorf("store: no user %q", username)
	}
	if _, err := tx.Exec(
		`DELETE FROM sessions WHERE user_id = (SELECT id FROM users WHERE username = ?)`,
		username); err != nil {
		return fmt.Errorf("store: revoking sessions for %q: %w", username, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("store: setting password for %q: %w", username, err)
	}
	return nil
}

// SetUserDisabled locks or unlocks an account. A disabled account cannot log in
// and its existing sessions stop working immediately (see LookupSession) --
// locking someone out must take effect now, not at their next login.
func (db *DB) SetUserDisabled(username string, disabled bool) error {
	res, err := db.sql.Exec(`UPDATE users SET disabled = ? WHERE username = ?`,
		boolToInt(disabled), username)
	if err != nil {
		return fmt.Errorf("store: updating user %q: %w", username, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: updating user %q: %w", username, err)
	}
	if n == 0 {
		return fmt.Errorf("store: no user %q", username)
	}
	return nil
}

// Users lists the accounts, sorted by name so two runs of the operator CLI
// print the same thing.
func (db *DB) Users() ([]User, error) {
	rows, err := db.sql.Query(
		`SELECT id, username, created_at, disabled FROM users ORDER BY username`)
	if err != nil {
		return nil, fmt.Errorf("store: listing users: %w", err)
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var (
			u        User
			created  string
			disabled int
		)
		if err := rows.Scan(&u.ID, &u.Username, &created, &disabled); err != nil {
			return nil, fmt.Errorf("store: listing users: %w", err)
		}
		if u.CreatedAt, err = parseStamp(created); err != nil {
			return nil, err
		}
		u.Disabled = disabled != 0
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: listing users: %w", err)
	}
	return out, nil
}

// CreateSession issues a session and returns the raw token for the cookie
// alongside the stored row.
//
// The raw token is returned and never kept: what goes into the table is its
// hash, so reading the sessions table yields nothing that can be presented as
// a credential.
func (db *DB) CreateSession(userID int64, userAgent, ip string) (string, Session, error) {
	raw, hash, err := auth.NewSessionToken()
	if err != nil {
		return "", Session{}, err
	}
	nowT := db.at()
	s := Session{
		UserID:     userID,
		CreatedAt:  nowT,
		LastUsedAt: nowT,
		ExpiresAt:  nowT.Add(SessionLifetime),
		UserAgent:  truncate(userAgent, 255),
		IP:         truncate(ip, 64),
	}
	res, err := db.sql.Exec(
		`INSERT INTO sessions (user_id, token_hash, created_at, last_used_at, expires_at, user_agent, ip)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.UserID, hash, s.CreatedAt.Format(timeFormat), s.LastUsedAt.Format(timeFormat),
		s.ExpiresAt.Format(timeFormat), s.UserAgent, s.IP)
	if err != nil {
		return "", Session{}, fmt.Errorf("store: creating session: %w", err)
	}
	if s.ID, err = res.LastInsertId(); err != nil {
		return "", Session{}, fmt.Errorf("store: creating session: %w", err)
	}
	return raw, s, nil
}

// LookupSession resolves a cookie value to its session and user, rolling the
// expiry window forward as a side effect.
//
// The roll-forward is what makes SessionLifetime a window of inactivity rather
// than a hard deadline. It is also why this is not a pure read: every
// authenticated request writes one row. At one user and a handful of requests
// that is irrelevant, and the alternative -- a session that dies mid-use a fixed
// time after login -- is the behaviour the owner explicitly did not want.
//
// Expiry is evaluated in Go against db.clock rather than in SQL against
// CURRENT_TIMESTAMP, so the window is testable and so there is one authority
// for time (rule 0.4).
func (db *DB) LookupSession(raw string) (Session, User, error) {
	if raw == "" {
		return Session{}, User{}, ErrNoSession
	}
	var (
		s              Session
		u              User
		created        string
		lastUsed       string
		expires        string
		userCreated    string
		disabled       int
		tokenHashValue = auth.HashSessionToken(raw)
	)
	err := db.sql.QueryRow(
		`SELECT s.id, s.user_id, s.created_at, s.last_used_at, s.expires_at, s.user_agent, s.ip,
		        u.username, u.created_at, u.disabled
		   FROM sessions s JOIN users u ON u.id = s.user_id
		  WHERE s.token_hash = ?`, tokenHashValue).
		Scan(&s.ID, &s.UserID, &created, &lastUsed, &expires, &s.UserAgent, &s.IP,
			&u.Username, &userCreated, &disabled)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Session{}, User{}, ErrNoSession
	case err != nil:
		return Session{}, User{}, fmt.Errorf("store: looking up session: %w", err)
	}
	if disabled != 0 {
		return Session{}, User{}, ErrNoSession
	}
	if s.CreatedAt, err = parseStamp(created); err != nil {
		return Session{}, User{}, err
	}
	if s.LastUsedAt, err = parseStamp(lastUsed); err != nil {
		return Session{}, User{}, err
	}
	if s.ExpiresAt, err = parseStamp(expires); err != nil {
		return Session{}, User{}, err
	}
	if u.CreatedAt, err = parseStamp(userCreated); err != nil {
		return Session{}, User{}, err
	}
	u.ID = s.UserID

	nowT := db.at()
	// THE WINDOW IS COMPUTED, NOT READ BACK. The deadline is last_used_at plus
	// the constant above, never the expires_at stamped into the row when it was
	// written -- so SessionLifetime is the single authority and shortening it
	// takes effect on every session that already exists, rather than only on
	// the ones created after the change. That is what empties a Devices list
	// full of rows a previous version stamped with a date three months out.
	if !nowT.Before(s.LastUsedAt.Add(SessionLifetime)) {
		// Expired. Delete it on sight rather than leaving it to the purge: the
		// row is useless and removing it here bounds the table even if the
		// periodic sweep never runs.
		if _, err := db.sql.Exec(`DELETE FROM sessions WHERE id = ?`, s.ID); err != nil {
			return Session{}, User{}, fmt.Errorf("store: removing expired session: %w", err)
		}
		return Session{}, User{}, ErrNoSession
	}

	s.LastUsedAt = nowT
	s.ExpiresAt = nowT.Add(SessionLifetime)
	if _, err := db.sql.Exec(
		`UPDATE sessions SET last_used_at = ?, expires_at = ? WHERE id = ?`,
		s.LastUsedAt.Format(timeFormat), s.ExpiresAt.Format(timeFormat), s.ID); err != nil {
		return Session{}, User{}, fmt.Errorf("store: rolling session forward: %w", err)
	}
	return s, u, nil
}

// RevokeSession ends one session. It is scoped to its owner: passing another
// user's session id is an error rather than a revocation, so an id guessed off
// the wire cannot be used to log somebody else out. There is one user today and
// security.md says more are a supported future.
func (db *DB) RevokeSession(userID, sessionID int64) error {
	res, err := db.sql.Exec(`DELETE FROM sessions WHERE id = ? AND user_id = ?`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("store: revoking session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: revoking session: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("store: no session %d for this user", sessionID)
	}
	return nil
}

// RevokeAllSessions logs a user out everywhere and reports how many sessions
// ended.
func (db *DB) RevokeAllSessions(userID int64) (int, error) {
	res, err := db.sql.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	if err != nil {
		return 0, fmt.Errorf("store: revoking sessions: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("store: revoking sessions: %w", err)
	}
	return int(n), nil
}

// Sessions lists one user's live sessions, newest first. This is what backs the
// visible session list with per-device revoke.
func (db *DB) Sessions(userID int64) ([]Session, error) {
	rows, err := db.sql.Query(
		`SELECT id, user_id, created_at, last_used_at, expires_at, user_agent, ip
		   FROM sessions WHERE user_id = ? ORDER BY last_used_at DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("store: listing sessions: %w", err)
	}
	defer rows.Close()

	var out []Session
	for rows.Next() {
		var s Session
		var created, lastUsed, expires string
		if err := rows.Scan(&s.ID, &s.UserID, &created, &lastUsed, &expires,
			&s.UserAgent, &s.IP); err != nil {
			return nil, fmt.Errorf("store: listing sessions: %w", err)
		}
		if s.CreatedAt, err = parseStamp(created); err != nil {
			return nil, err
		}
		if s.LastUsedAt, err = parseStamp(lastUsed); err != nil {
			return nil, err
		}
		// Parsed to prove the row is readable, then DISCARDED in favour of the
		// computed window. A row written under the old 90-day rule carries a
		// date the server will no longer honour, and this page is the one place
		// the owner looks to see what is still live -- telling them a session
		// is good until November when it dies in a fortnight is the same class
		// of lie as the stale rows themselves.
		if _, err = parseStamp(expires); err != nil {
			return nil, err
		}
		s.ExpiresAt = s.LastUsedAt.Add(SessionLifetime)
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: listing sessions: %w", err)
	}
	return out, nil
}

// PurgeExpiredSessions deletes sessions past their window and reports how many
// went. Run periodically by the server so the table stays bounded.
//
// It asks the same question LookupSession does, in the same terms: idle for
// longer than SessionLifetime. Sweeping on the stored expires_at instead would
// mean the hourly sweep and the request path could disagree about which
// sessions are alive -- and it would never reach the rows written under a
// longer window, which are precisely the ones this sweep exists to remove.
func (db *DB) PurgeExpiredSessions() (int, error) {
	cutoff := db.at().Add(-SessionLifetime).Format(timeFormat)
	res, err := db.sql.Exec(`DELETE FROM sessions WHERE last_used_at <= ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("store: purging sessions: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("store: purging sessions: %w", err)
	}
	return int(n), nil
}

// parseStamp reads a timestamp column back. A column that will not parse means
// the file was written by something other than this code, which is worth an
// error rather than a zero time silently standing in for an expiry date.
func parseStamp(s string) (time.Time, error) {
	t, err := time.Parse(timeFormat, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("store: unreadable timestamp %q: %w", s, err)
	}
	return t, nil
}

// truncate bounds a string written from a request header. A User-Agent is
// attacker-controlled and unbounded; the session list is the only thing that
// reads these and it does not need a megabyte.
func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
