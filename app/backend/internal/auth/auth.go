// Package auth holds the credential primitives: password hashing and session
// token generation. The threat model and the reasoning behind each choice are
// in app/docs/security.md; this file is the implementation of that document.
//
// It is deliberately small and has no knowledge of HTTP or of the database.
// Storage lives in internal/store and the cookie handling in cmd/server, so
// the part that must be right in a cryptographic sense can be tested
// exhaustively on its own.
//
// Two structural properties are worth stating, because they are enforced by
// the shape of the API rather than by remembering to do the right thing:
//
//   - NewSessionToken hands back the raw token and its hash together. A caller
//     writing a session row is given the hash, so storing the raw value is not
//     something it can do by accident.
//   - Verification always reads its cost parameters out of the stored hash, so
//     raising DefaultParams never invalidates an existing password.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Params are the Argon2id cost parameters. They are written into every hash so
// that a hash made under one set still verifies after the defaults move.
type Params struct {
	Memory  uint32 // KiB
	Time    uint32 // iterations
	Threads uint8
	SaltLen uint32 // bytes
	KeyLen  uint32 // bytes
}

// DefaultParams is an OWASP-recommended Argon2id configuration: 19 MiB, two
// iterations, one lane.
//
// The lighter of OWASP's two recommended sets is chosen on purpose. This app
// shares a 1 vCPU / 2 GB droplet with the owner's other sites, and CLAUDE.md
// rule 0.3 says a fault here must not reach them -- 64 MiB per login attempt
// would turn the login form into a memory-pressure lever against the whole box.
// One lane matches one vCPU, where extra parallelism buys nothing.
//
// Raising these later is a non-breaking change: existing hashes keep verifying
// under their own parameters and NeedsRehash flags them for upgrade at the next
// successful login.
var DefaultParams = Params{
	Memory:  19456, // 19 MiB
	Time:    2,
	Threads: 1,
	SaltLen: 16,
	KeyLen:  32,
}

// Password length bounds. The minimum is high because this single account
// guards a legal record and there is no second factor. The maximum exists
// because Argon2 will happily chew on a megabyte of input, and an unbounded
// password field is free CPU and memory for an attacker.
const (
	MinPasswordLen = 12
	MaxPasswordLen = 1024
)

// ErrPasswordLength is returned by HashPassword for a password outside the
// bounds above.
var ErrPasswordLength = fmt.Errorf("auth: password must be between %d and %d characters",
	MinPasswordLen, MaxPasswordLen)

// randRead is the entropy source. It is a variable solely so the tests can
// prove that an entropy failure refuses to issue a credential rather than
// falling back to something predictable.
var randRead io.Reader = rand.Reader

// SetRandReaderForTest swaps the entropy source and returns a function that
// restores it. Test-only.
func SetRandReaderForTest(r io.Reader) (restore func()) {
	prev := randRead
	randRead = r
	return func() { randRead = prev }
}

// HashPassword hashes a password with DefaultParams, returning a PHC-encoded
// string suitable for the users.password_hash column.
func HashPassword(password string) (string, error) {
	return HashPasswordWith(password, DefaultParams)
}

// HashPasswordWith hashes a password with explicit parameters. Used by the
// tests and by any future re-hashing at a raised cost.
func HashPasswordWith(password string, p Params) (string, error) {
	if len(password) < MinPasswordLen || len(password) > MaxPasswordLen {
		return "", ErrPasswordLength
	}
	salt := make([]byte, p.SaltLen)
	if _, err := io.ReadFull(randRead, salt); err != nil {
		return "", fmt.Errorf("auth: reading salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	return encode(p, salt, key), nil
}

// VerifyPassword reports whether the password matches the encoded hash.
//
// A non-nil error means the stored hash could not be decoded -- database
// corruption, a truncated column, or a hash from some other scheme. The caller
// must deny access in that case just as it would for a wrong password; the
// error exists so the operator finds out, not so the outcome differs.
func VerifyPassword(encoded, password string) (bool, error) {
	p, salt, want, err := decode(encoded)
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	// Constant time: a length-dependent or early-exit comparison leaks how much
	// of a guess was right.
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// NeedsRehash reports whether a stored hash was made with weaker parameters
// than the current defaults, so the login handler can transparently upgrade it
// while it has the plaintext in hand.
//
// An undecodable hash needs replacing more urgently than a weak one, so it
// answers true rather than surfacing the error again.
func NeedsRehash(encoded string) bool {
	p, _, _, err := decode(encoded)
	if err != nil {
		return true
	}
	return p.Memory < DefaultParams.Memory ||
		p.Time < DefaultParams.Time ||
		p.KeyLen < DefaultParams.KeyLen
}

// SessionTokenLen is the size of a session identifier. The cookie is the whole
// credential for the life of the session, so it carries a full 256 bits.
const SessionTokenLen = 32

// NewSessionToken mints a session identifier and returns it twice: raw, for the
// cookie, and hashed, for the database.
//
// Returning both is the control from security.md -- the database stores only
// the hash, so reading the sessions table does not yield usable tokens. Making
// it the single way to mint a session means a caller has the hash to hand and
// no reason to reach for the raw value.
func NewSessionToken() (raw, hash string, err error) {
	b := make([]byte, SessionTokenLen)
	if _, err := io.ReadFull(randRead, b); err != nil {
		return "", "", fmt.Errorf("auth: reading session token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, HashSessionToken(raw), nil
}

// HashSessionToken maps a cookie value to its database key.
//
// Plain SHA-256 rather than a password hash, deliberately: the input is 256
// bits of uniform randomness, so there is no dictionary to attack and no work
// factor worth paying on every single request.
func HashSessionToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// encode renders a hash in the PHC string format, the same layout the
// reference Argon2 implementation uses:
//
//	$argon2id$v=19$m=19456,t=2,p=1$<salt>$<key>
func encode(p Params, salt, key []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Time, p.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key))
}

// decode parses a PHC string back into its parameters, salt and key.
//
// Every failure here is a potential authentication bypass if it were to return
// a usable zero value, so each one returns an error and no caller may proceed
// on one.
func decode(encoded string) (p Params, salt, key []byte, err error) {
	parts := strings.Split(encoded, "$")
	// A leading "$" makes the first field empty, so a well-formed string has
	// six fields: "", algorithm, version, parameters, salt, key.
	if len(parts) != 6 || parts[0] != "" {
		return p, nil, nil, errors.New("auth: hash is not in PHC format")
	}
	if parts[1] != "argon2id" {
		return p, nil, nil, fmt.Errorf("auth: hash algorithm is %q, want argon2id", parts[1])
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return p, nil, nil, fmt.Errorf("auth: unreadable argon2 version: %w", err)
	}
	if version != argon2.Version {
		return p, nil, nil, fmt.Errorf("auth: argon2 version %d, want %d", version, argon2.Version)
	}

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Threads); err != nil {
		return p, nil, nil, fmt.Errorf("auth: unreadable argon2 parameters: %w", err)
	}

	if salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil {
		return p, nil, nil, fmt.Errorf("auth: unreadable salt: %w", err)
	}
	if key, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil {
		return p, nil, nil, fmt.Errorf("auth: unreadable key: %w", err)
	}

	p.SaltLen = uint32(len(salt))
	p.KeyLen = uint32(len(key))
	return p, salt, key, nil
}
