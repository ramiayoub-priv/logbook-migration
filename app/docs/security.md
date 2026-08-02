# Security

Rule §0.3 in the repo root `CLAUDE.md` is the standing instruction: **security is key**. This file is
the threat model and the control list. **Every control below must have a test** — a control without a
test is a claim, not a control.

## Threat model

A single-user private flight logbook on a public IP, on a box shared with the owner's other sites.
What actually matters, in order:

1. **Unauthenticated read of the logbook.** It is personal data: names, dates, locations, a movement
   history spanning 14 years. This is the primary asset.
2. **Unauthenticated write.** The logbook is a legal record backing licence privileges. A silently
   injected or altered flight is worse than a leak.
3. **Credential compromise** — brute force, or reuse of a password seen elsewhere.
4. **Session theft** — especially because the user explicitly wants long-lived sessions so they never
   have to log in. Convenience here directly raises the value of a stolen cookie.
5. **Collateral damage to the rest of the server.** A compromise or a resource exhaustion in this app
   must not reach the other sites.

Explicitly *out* of scope for v1: multi-tenant isolation (one user), and DoS beyond what fail2ban and
Apache already absorb.

## Controls

> **Status (2026-08-01): every control in this file is implemented and tested.** The code is
> `internal/auth` (primitives), `internal/store/auth.go` (users and sessions),
> `internal/ratelimit` (login throttling) and `cmd/server` (the router and middleware). The
> control-to-test map at the bottom names the test that fails if a control is removed. The three
> **action items** below are still open — they are operational, not code.

### Authentication
- **Argon2id** password hashing. Parameters recorded in the encoded hash so they can be raised later
  without invalidating existing hashes. `auth.NeedsRehash` flags a hash written under weaker
  parameters so the login handler can upgrade it while it has the plaintext in hand.
  - **Parameters: m=19456 (19 MiB), t=2, p=1** — OWASP's *lighter* recommended set, chosen
    deliberately over the 64 MiB one. This app shares a 1 vCPU / 2 GB droplet with the owner's other
    sites and rule §0.3 says a fault here must not reach them; 64 MiB per login attempt turns the
    login form into a memory-pressure lever against the whole box. One lane matches one vCPU.
- **The unknown-user path pays for a full hash.** `store.Authenticate` verifies against a decoy hash
  when the username does not exist, so a missing account does not answer in microseconds where a
  real one takes tens of milliseconds. Without this the login form is a username oracle readable
  with a stopwatch.
- **No self-service registration.** Users are created by a CLI subcommand on the server
  (`logbook-server createuser`). The endpoint does not exist, so it cannot be abused. Adding users
  later is supported by design; opening registration would be a deliberate future decision.
- **Login rate limiting**, per-IP and per-account, with exponential backoff. fail2ban already guards
  SSH; this guards the app.
- **Uniform failure responses and timing.** A wrong username and a wrong password are indistinguishable.

### Sessions
- **Server-side sessions in SQLite**, not JWTs. The deciding factor is revocation: a JWT cannot be
  withdrawn before expiry, and a 90-day JWT is a 90-day liability. A row in a table can be deleted.
- The cookie carries a 256-bit random identifier; the **database stores only its hash**, so a DB read
  does not yield usable session tokens. `auth.NewSessionToken` returns the raw value and its hash
  *together*, so the code that writes the session row is handed the hash and has no reason to reach
  for the raw one — the property is structural, not remembered.
- Cookie flags: `HttpOnly` (no JS access), `Secure` (TLS only), `SameSite=Lax`, `Path=/logbook`.
  The path matters: at `/` the cookie would be sent to the owner's other sites on this box.
  No `Expires`/`Max-Age` — it is a browser session cookie, and the real 90-day life is enforced
  server-side where it can be revoked.
- **90-day rolling expiry** — this is what delivers "I don't want to log in every time". Each use
  extends the window; an unused session dies. Evaluated in Go against an injectable clock rather
  than in SQL against `CURRENT_TIMESTAMP`: one authority for time (rule §0.4), and the window gets
  tested in milliseconds instead of taken on trust.
- **A disabled account's live sessions stop working immediately**, not at the next login.
- A visible session list with individual revoke, plus revoke-all-on-password-change. Revocation is
  scoped to the owning user in the query itself, so a session id guessed off the wire cannot log
  somebody else out.

### Request handling
- **Default deny.** Routes are registered through a table rather than straight onto the `ServeMux`,
  and the registration function applies the auth wrapper — a handler cannot be mounted without
  passing through it. `public` is a field the author has to set, so a new endpoint is private by
  construction even if the author forgets. `Server.Routes()` exposes the table so the test can
  enumerate what is *actually mounted* rather than a hand-maintained list.
  - The only public endpoints are **`POST /login`** and **`GET /health`**. Health returns exactly
    `{"status":"ok"}` — no version, no counts, no paths. Static assets are served by Apache, not
    by this process.
- **CSRF**: `SameSite=Lax` plus an `Origin` check on every state-changing method, failing closed —
  a request with **no** `Origin` is refused, which costs nothing because our own frontend always
  sends one. `SameSite` is the browser's promise; this check does not depend on the browser
  behaving. Safe methods are exempt: a cross-origin GET changes nothing and the same-origin policy
  already stops the attacker reading the response.
- **JSON only.** A non-`application/json` content type is a 415. This is a CSRF control as much as a
  parsing one: a cross-origin HTML form can only send three content types, none of them JSON, so
  the classic form-post vector cannot reach this API at all.
- **Input validation at the boundary**, into typed structs, with `DisallowUnknownFields` — a field
  the server does not understand means client and server disagree, and guessing which is right is
  not acceptable on a legal record. Date parameters are validated for shape *and* reality;
  an unparseable one is a 400, never a silently ignored filter that would answer a question the
  user did not ask.
- **Body size limits on every request** (64 KiB), applied in `ServeHTTP` so they also cover requests
  that match no route. Server-side read/write/idle timeouts and a 16 KiB header cap.
- **Parameterized SQL everywhere.** No string-built queries, ever.
- **Errors are dull.** "authentication required" covers a missing cookie, an expired session and a
  disabled account alike; a read failure says "could not read the logbook" and never the driver,
  the table or the query. The operator gets the reason in the log; the client gets a yes or a no.
- **A read failure is a 500, never an empty 200.** A 200 carrying an empty list reads as "you have
  flown nothing", which is precisely the silent corruption rule §0.2 forbids. The server also
  refuses to start against a database with no flights in it.
- **Security headers**: `Content-Security-Policy` (no inline script; the PWA is built to comply),
  `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: same-origin`, HSTS.

### Data
- **The SQLite file lives outside the web root**, at `/var/lib/logbook/`, owned by the service user
  and mode `0600`. Apache serves `/var/www/logbook` and can never reach it.
- **Automated backup before every migration and on a schedule**, with restores actually tested. Rule
  §0.2: this is a legal record.
- Logs record authentication outcomes but **never** passwords, session tokens, or cookie values.

### The off-box backup — what leaves the machine, and why that is acceptable

Added 2026-08-02 (Task 14). A copy of the whole database is pushed daily to a **private** GitHub
repository. That is deliberately a copy of a legal record leaving a machine we control, so the
trade is written down rather than assumed.

**What goes.** `logbook.db` (the whole schema and its contents) and `logbook.csv` (every flight).
The flights themselves are the point: since the transcription effort closed, rows entered in the app
exist in no CSV, so the pre-import backups on the same disk protect against a bad import and against
nothing else — not a dead disk, not a lost server, not a mistaken `rm`.

**`sessions` is stripped before the copy leaves the box** (`store.RedactForBackup`, then `VACUUM` so
the rows are not merely unlinked pages still readable in the file). They are the nearest thing in
the schema to a live credential. The column is a **hash** of the cookie rather than the cookie
itself, so even an unredacted copy would yield nothing usable — but a backup should carry the
smallest set that still restores, and sessions restore nothing: the expiry has passed and the
addresses are stale.

**`users` deliberately survives, Argon2id hash and all.** This is the one real trade. A restored
logbook nobody can sign in to is not a restored logbook, and the alternative — reconstructing the
account by hand under the pressure of an outage — is worse. The hash is Argon2id at 19 MiB, which is
what makes an offline attack against a private repository an acceptable risk rather than a
theoretical one. **If that repository is ever made public, or the GitHub account is compromised,
treat the logbook password as compromised and rotate it.**

**Credentials for the push.** An **ed25519 SSH key registered on the dedicated `ramiayoub-priv`
GitHub account**, generated on the server so the private half never travels, stored at
`/var/lib/logbook/.ssh/backup_ed25519` (`0600`, owned by `logbook`) — outside the web root and
outside this repository (rule §0.3). An SSH key rather than a personal access token, because the
private half is created on the box and never leaves it.

⚠ **Account-level, not a repository deploy key — OWNER RULING 2026-08-02**, and it is a real
trade-off rather than an oversight. A deploy key reaches exactly one repository; **this key reaches
every repository `ramiayoub-priv` owns.** The owner's ruling is that the account is dedicated to this
purpose and holds nothing else, so the account boundary already supplies the scope. **The control
that makes that true is therefore a policy one: that account must keep owning nothing but the backup
repository.** If it ever acquires another repository, this key's blast radius grows silently and the
decision should be revisited. Revocation is still one click, from the account's key list.

The kind of key in use is otherwise invisible, so `install-backup.sh` step 5 reports it: GitHub
answers `Hi ramiayoub-priv!` for an account key and `Hi ramiayoub-priv/logbook-backup!` for a deploy
key.

`StrictHostKeyChecking=yes` against a `known_hosts` pinned at install time, so a substituted GitHub
host key fails the push instead of being trusted silently.

**Least privilege.** The timer runs `logbook-backup.service` as the **`logbook` user, not root** —
it reads a database it already owns and writes into a directory it already owns — with the same
systemd hardening as the server and `ReadWritePaths=/var/lib/logbook` as its only writable path.
Nothing in the backup path is reachable over HTTP: the snapshot is taken by `logbookctl`, which is a
separate binary from the server for exactly this reason.

### Process and supply chain
- **Keep the dependency tree near-empty.** Prefer the Go stdlib. Every dependency added must be
  justified in `APP.md`. The intended total is single digits: `modernc.org/sqlite`,
  `golang.org/x/crypto` (Argon2id), `go-pdf/fpdf`.
  - **Direct backend dependencies as of 2026-08-01 — four**: `modernc.org/sqlite`,
    `golang.org/x/crypto` (Argon2id), `golang.org/x/term` (reading a password without echoing it —
    the alternative is a password in the shell history and the process list of a shared box), and
    **`github.com/go-pdf/fpdf` v0.9.0**, added with Task 6. Everything else in `go.mod` is indirect,
    pulled in by `modernc.org/sqlite`.
  - **Why `fpdf` is justified**: the EASA export is a fixed 15-row grid with absolutely positioned
    cells, which is what the format is. The alternative that renders HTML — headless Chrome — is a
    300 MB+ browser on a 2 GB box shared with the owner's other sites, executing a rendering engine
    on request; that is a far larger attack surface and a far larger memory lever than a pure-Go
    drawing library. `fpdf` is pure Go with **no dependencies of its own**, so it adds exactly one
    node to the tree, and it keeps `CGO_ENABLED=0` and the single static binary. It also embeds the
    cp1252 map, so no font or encoding file has to ship beside the binary.
- **The frontend's dependencies are build-time only.** Node never runs on the server: `npm run
  build` produces static files that Apache serves. The runtime bundle is React and React DOM and
  nothing else — routing is ~40 lines in `src/router.tsx` rather than a routing library, because a
  six-page app with no nested routes would be buying a supply-chain decision for a switch statement.
- **No secrets in the repo, ever** — enforced by review and by keeping config in environment
  variables read at startup.
- The service runs as a **dedicated unprivileged user**, never root, with systemd hardening
  (`NoNewPrivileges`, `ProtectSystem=strict`, `PrivateTmp`, `MemoryMax`). `MemoryMax` is a
  *containment* control as much as a resource one: this app must not be able to OOM the other sites.

## Action items

- [ ] ⚠ **Rotate the `rami` sudo password. THE MOST URGENT OPEN ITEM — exposed twice.**
      - **2026-08-01**: pasted into a chat session to allow server setup.
      - **2026-08-02**: handed over in-session again ("`… <-- for sudo, you deploy`") so that the
        Tasks 16/17/18 deploy could run its `sudo` steps without the owner present. The owner
        accepted the trade explicitly and undertook to rotate immediately afterwards; **that deploy
        is complete**, so the condition attached to the exposure has been met.
      - It therefore exists in **two transcripts**, and it is `sudo` on a box serving eight sites,
        including the logbook's live legal record.
      - **Rotation is not enough on its own if either transcript is ever shared.** Treat the
        credential as burned: change it, and do not re-enter the new one into a session. The
        deploy runbook's `sudo` steps are the owner's to run for exactly this reason.
      - The rule this violates is `CLAUDE.md` §0.3: *"If a secret is ever pasted into chat or a
        file, treat it as compromised and rotate it."*
- [ ] Decide the fate of the publicly-exposed `:8000` container (see `deploy.md`).
- [ ] Prune stale ufw rules (`30814`, `19132`) and the duplicated Apache profile rules.

## Testing the controls

Each control maps to a test that fails if the control is removed. **All of these exist and pass**
(2026-08-01); `internal/auth` and `internal/ratelimit` are held to **100%** coverage by the
Makefile's `CORE` list, because in a credential primitive an untested branch is an authentication
bypass.

| Control | Test | Where |
|---|---|---|
| Default deny | Enumerates every *registered* route from `Server.Routes()` and asserts 401 without a session unless on a two-entry allow-list. A new endpoint fails the suite until someone edits the allow-list on purpose. | `TestEveryRouteIsPrivateUnlessExplicitlyPublic` |
| Session expiry | Injected clock: 89 days of use rolls forward forever, 91 days idle is dead. | `TestSessionExpiryIsRollingAndEnforced` |
| Token hashing | Reads `token_hash` **out of the file** and asserts it is not the cookie value. | `TestTheRawTokenIsNeverStored` |
| CSRF | Every mutating method × three hostile origins → 403; a missing `Origin` → 403; a cross-origin GET still works. | `TestCrossOriginMutationsAreRejected`, `TestAMutationWithNoOriginHeaderIsRejected` |
| Form-post vector | Three form/text content types → 415. | `TestFormContentTypesAreRefused` |
| Rate limiting | Nth failed login → 429 with `Retry-After`; the correct password is *also* refused while the penalty stands; success resets. | `TestLoginIsRateLimited`, `TestSuccessResets` |
| Backoff soundness | Doubling, capped, and the shift clamped before it is taken — at 63 doublings a `Duration` wraps negative and would read as "no penalty". | `TestBackoffIsExponentialAndCapped`, `TestTheBackoffShiftCannotOverflow` |
| Limiter memory bound | Flooding with fresh keys must not evict the attacked key's penalty. | `TestTheTableIsBoundedAndEvictsTheStalest` |
| Cookie flags | `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/logbook`, and the token absent from the body. | `TestLoginSetsAHardenedCookie` |
| SQL injection | Hostile-input table against every string parameter reaching a query, then asserts the database still works. | `TestHostileInputIsParameterised` |
| Password hashing | Argon2id encoding with embedded parameters, distinct salts, no plaintext in the hash, twelve malformed-hash shapes that must each deny rather than bypass. | `internal/auth` suite |
| Login timing | Measures the unknown-user path against the real one and fails if it gets cheap. | `TestAuthenticateDoesTheWorkForAnUnknownUser` |
| Uniform failures | A wrong password and an unknown user must produce byte-identical responses. | `TestLoginFailsUniformly` |
| Entropy failure | An injected failing RNG must make both token minting and password hashing refuse, never fall back to something predictable. | `TestEntropyFailuresPropagate` |
| Session scoping | Revoke and list are scoped to the owner; another user's id is a 404, not a revocation. | `TestRevokeSession`, `TestSessionsListsTheUsersOwnOnly` |
| Password change | Revokes every session in the same transaction. | `TestSetPasswordRevokesEverySession` |
| Security headers | Asserted on a 200, a 401 **and a 404** — the places they are most often forgotten. | `TestSecurityHeadersAreOnEveryResponse` |
| No empty-logbook lie | A failed read is a 500 with no schema in the message, never a 200 with an empty list. | `TestAReadFailureIs500AndNotAnEmptyLogbook` |
| Handler can't skip auth | `callerOf` panics rather than serving a zero-value user, if a handler is ever mounted without `requireSession`. | `TestCallerOfPanicsWithoutASession` |
| **Write path is authenticated + CSRF-checked** | `POST /flights` with no `Origin`, and from a foreign origin, → 403. It is also covered automatically by the default-deny enumeration above. | `TestCreatingAFlightRequiresAnOrigin` |
| **Write path rejects unknown fields** | A body carrying a field the server does not model → 400, rather than a silently ignored value on a legal record. | `TestAnUnknownFieldIsRejectedRatherThanIgnored` |
| **No duplicate flights** | The same flight submitted twice → 409, and the count does not move. | `TestASubmittedFlightCannotBeSubmittedTwice` |
| **A rejected write stores nothing** | An invalid draft leaves the flight count exactly where it was. | `TestNothingIsStoredWhenAFlightIsRejected` |
| **Exports are private** | The three `/export/*.pdf` routes are in the default-deny enumeration and answer 401 without a session. | `TestEveryRouteIsPrivateUnlessExplicitlyPublic` |
| **The frontend never touches the cookie** | The fetch layer sends `credentials: 'same-origin'` on every call and no `Authorization` header; nothing reads `document.cookie`. | `src/api.test.ts` |
| **The service worker never caches the logbook** | Evaluates the shipped `public/sw.js` and asserts every `/logbook/api/` URL is passed through unstored, including a navigation to one and every mutating method. A worker ignores `Cache-Control: no-store` unless written not to, so caching an API response would quietly undo a control the server states explicitly and leave the logbook readable on the device after the session was revoked. Verified in a browser too: after signing in and browsing, the cache holds only the shell. | `src/sw.test.ts` |
| **The login page stays uninformative** | The page shows one message for every credential failure and must not name the cause — undoing the uniform-response control in the UI would be just as much of a username oracle. | `src/pages/pages.test.tsx` |
