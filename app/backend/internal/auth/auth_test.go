package auth_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/auth"
)

const good = "a-sufficiently-long-passphrase"

// --- Password hashing -------------------------------------------------------

// TestHashIsArgon2idEncoded is the control from app/docs/security.md: stored
// hashes must be Argon2id with their parameters embedded, so the cost can be
// raised later without invalidating hashes made today.
func TestHashIsArgon2idEncoded(t *testing.T) {
	h, err := auth.HashPassword(good)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(h, "$argon2id$v=19$") {
		t.Errorf("hash %q is not a PHC-encoded argon2id string", h)
	}
	if n := len(strings.Split(h, "$")); n != 6 {
		t.Errorf("hash has %d $-separated fields, want 6", n)
	}
	// The parameters must be readable back out of the hash, which is what
	// makes raising them later a non-breaking change.
	if !strings.Contains(h, fmt.Sprintf("m=%d,t=%d,p=%d",
		auth.DefaultParams.Memory, auth.DefaultParams.Time, auth.DefaultParams.Threads)) {
		t.Errorf("hash %q does not carry its own parameters", h)
	}
}

// TestIdenticalPasswordsHashDifferently is the salt control: two accounts with
// the same password must not be visibly the same in the database, and a
// precomputed table must not work.
func TestIdenticalPasswordsHashDifferently(t *testing.T) {
	a, err := auth.HashPassword(good)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	b, err := auth.HashPassword(good)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if a == b {
		t.Error("the same password hashed twice produced identical output; the salt is not random")
	}
	// Both must still verify -- a random salt is only useful if it round-trips.
	for _, h := range []string{a, b} {
		ok, err := auth.VerifyPassword(h, good)
		if err != nil || !ok {
			t.Errorf("VerifyPassword(%q) = %v, %v; want true, nil", h, ok, err)
		}
	}
}

// TestThePlaintextNeverAppearsInTheHash guards the obvious catastrophe.
func TestThePlaintextNeverAppearsInTheHash(t *testing.T) {
	h, err := auth.HashPassword(good)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if strings.Contains(h, good) {
		t.Errorf("the encoded hash contains the plaintext password")
	}
}

func TestVerifyAcceptsOnlyTheRightPassword(t *testing.T) {
	h, err := auth.HashPassword(good)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	for _, c := range []struct {
		name string
		try  string
		want bool
	}{
		{"the password", good, true},
		{"a different password", "some-other-passphrase", false},
		{"one character short", good[:len(good)-1], false},
		{"one character extra", good + "x", false},
		{"case changed", strings.ToUpper(good), false},
		{"empty", "", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			ok, err := auth.VerifyPassword(h, c.try)
			if err != nil {
				t.Fatalf("VerifyPassword: %v", err)
			}
			if ok != c.want {
				t.Errorf("VerifyPassword(_, %q) = %v, want %v", c.try, ok, c.want)
			}
		})
	}
}

// TestPasswordLengthIsBounded. A minimum because this account guards a legal
// record and there is no second factor; a maximum because an unbounded input
// is free CPU and memory for an attacker on a 2 GB shared box.
func TestPasswordLengthIsBounded(t *testing.T) {
	for _, c := range []struct {
		name string
		pw   string
	}{
		{"empty", ""},
		{"one below the minimum", strings.Repeat("x", auth.MinPasswordLen-1)},
		{"one above the maximum", strings.Repeat("x", auth.MaxPasswordLen+1)},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := auth.HashPassword(c.pw); !errors.Is(err, auth.ErrPasswordLength) {
				t.Errorf("HashPassword(%d chars) error = %v, want ErrPasswordLength", len(c.pw), err)
			}
		})
	}
	for _, c := range []struct {
		name string
		pw   string
	}{
		{"exactly the minimum", strings.Repeat("x", auth.MinPasswordLen)},
		{"exactly the maximum", strings.Repeat("x", auth.MaxPasswordLen)},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := auth.HashPassword(c.pw); err != nil {
				t.Errorf("HashPassword(%d chars) = %v, want it accepted", len(c.pw), err)
			}
		})
	}
}

// TestVerifyRejectsAMalformedHash. A corrupted or truncated hash column must
// deny access and say why, never accidentally authenticate. Every branch of the
// decoder is a potential auth bypass, so each one is asserted.
func TestVerifyRejectsAMalformedHash(t *testing.T) {
	valid, err := auth.HashPassword(good)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	parts := strings.Split(valid, "$")

	for _, c := range []struct {
		name    string
		encoded string
	}{
		{"empty", ""},
		{"not PHC at all", "hunter2"},
		{"too few fields", "$argon2id$v=19$m=1,t=1,p=1$c2FsdA"},
		{"too many fields", valid + "$extra"},
		{"a different algorithm", strings.Replace(valid, "argon2id", "argon2i", 1)},
		{"bcrypt", "$2y$10$abcdefghijklmnopqrstuv"},
		{"an unknown version", strings.Replace(valid, "v=19", "v=20", 1)},
		{"a malformed version", strings.Replace(valid, "v=19", "v=xx", 1)},
		{"malformed parameters", strings.Replace(valid, parts[3], "m=x,t=y,p=z", 1)},
		{"missing a parameter", strings.Replace(valid, parts[3], "m=19456,t=2", 1)},
		{"a corrupt salt", strings.Replace(valid, parts[4], "!!!not base64!!!", 1)},
		{"a corrupt hash", strings.Replace(valid, parts[5], "!!!not base64!!!", 1)},
	} {
		t.Run(c.name, func(t *testing.T) {
			ok, err := auth.VerifyPassword(c.encoded, good)
			if ok {
				t.Fatalf("a malformed hash authenticated the user")
			}
			if err == nil {
				t.Errorf("VerifyPassword returned no error for %q", c.encoded)
			}
		})
	}
}

// TestVerifyHonoursTheParametersInTheHash is what makes raising the cost a
// non-breaking change: a hash written with the old parameters must still
// verify after DefaultParams moves.
func TestVerifyHonoursTheParametersInTheHash(t *testing.T) {
	weak := auth.Params{Memory: 8, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}
	h, err := auth.HashPasswordWith(good, weak)
	if err != nil {
		t.Fatalf("HashPasswordWith: %v", err)
	}
	if !strings.Contains(h, "m=8,t=1,p=1") {
		t.Fatalf("hash %q did not record the parameters it was made with", h)
	}
	ok, err := auth.VerifyPassword(h, good)
	if err != nil || !ok {
		t.Errorf("VerifyPassword on a weak-parameter hash = %v, %v; want true, nil", ok, err)
	}
	// And it must still reject the wrong password at those parameters.
	if ok, _ := auth.VerifyPassword(h, "wrong-passphrase-here"); ok {
		t.Error("a weak-parameter hash accepted the wrong password")
	}
}

// TestNeedsRehash tells the login handler when to transparently upgrade a hash
// that was written under weaker parameters.
func TestNeedsRehash(t *testing.T) {
	weak, err := auth.HashPasswordWith(good, auth.Params{Memory: 8, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32})
	if err != nil {
		t.Fatalf("HashPasswordWith: %v", err)
	}
	current, err := auth.HashPassword(good)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !auth.NeedsRehash(weak) {
		t.Error("a hash below the current parameters should be flagged for rehashing")
	}
	if auth.NeedsRehash(current) {
		t.Error("a hash at the current parameters should not be flagged")
	}
	// An undecodable hash needs replacing more than any other.
	if !auth.NeedsRehash("garbage") {
		t.Error("an undecodable hash should be flagged for rehashing")
	}
}

// --- Session tokens ---------------------------------------------------------

// TestNewSessionTokenReturnsTheHashSoTheRawIsNeverStored. The API hands back
// both halves precisely so a caller cannot accidentally write the raw token to
// the database: the value that goes in the row is the only one it is given.
func TestNewSessionTokenReturnsTheHashSoTheRawIsNeverStored(t *testing.T) {
	raw, hash, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken: %v", err)
	}
	if raw == "" || hash == "" {
		t.Fatal("NewSessionToken returned an empty half")
	}
	if raw == hash {
		t.Error("the stored hash equals the cookie value; a database read would yield live sessions")
	}
	if strings.Contains(hash, raw) {
		t.Error("the stored hash contains the cookie value")
	}
	if got := auth.HashSessionToken(raw); got != hash {
		t.Errorf("HashSessionToken(raw) = %q, want the hash returned alongside it (%q)", got, hash)
	}
}

// TestSessionTokensCarry256BitsOfEntropy. The cookie is the whole credential
// for 90 days, so it must not be guessable.
func TestSessionTokensCarry256BitsOfEntropy(t *testing.T) {
	raw, _, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken: %v", err)
	}
	// base64url without padding: 32 bytes -> 43 characters.
	if len(raw) != 43 {
		t.Errorf("token %q is %d characters, want 43 (32 random bytes, base64url)", raw, len(raw))
	}
	if strings.ContainsAny(raw, "+/=") {
		t.Errorf("token %q is not URL-safe; it goes in a cookie", raw)
	}

	// Distinctness, as a smoke test against a fixed or reused token.
	seen := map[string]bool{}
	for range 100 {
		tok, _, err := auth.NewSessionToken()
		if err != nil {
			t.Fatalf("NewSessionToken: %v", err)
		}
		if seen[tok] {
			t.Fatalf("token %q was issued twice", tok)
		}
		seen[tok] = true
	}
}

func TestHashSessionTokenIsDeterministicAndSpecific(t *testing.T) {
	a := auth.HashSessionToken("token-one")
	if a != auth.HashSessionToken("token-one") {
		t.Error("hashing the same token twice gave different results; session lookup would fail")
	}
	if a == auth.HashSessionToken("token-two") {
		t.Error("two different tokens hashed to the same value")
	}
	if a == "" {
		t.Error("HashSessionToken returned an empty string")
	}
}

// TestEntropyFailuresPropagate. If the system's randomness source fails we must
// refuse to issue a credential, never fall back to something predictable.
func TestEntropyFailuresPropagate(t *testing.T) {
	restore := auth.SetRandReaderForTest(failingReader{})
	defer restore()

	if _, _, err := auth.NewSessionToken(); err == nil {
		t.Error("NewSessionToken succeeded with a broken entropy source")
	}
	if _, err := auth.HashPassword(good); err == nil {
		t.Error("HashPassword succeeded with a broken entropy source, so the salt was not random")
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("no entropy") }
