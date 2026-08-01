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

### Authentication
- **Argon2id** password hashing. Parameters recorded in the encoded hash so they can be raised later
  without invalidating existing hashes.
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
  does not yield usable session tokens.
- Cookie flags: `HttpOnly` (no JS access), `Secure` (TLS only), `SameSite=Lax`, `Path=/logbook`.
- **90-day rolling expiry** — this is what delivers "I don't want to log in every time". Each use
  extends the window; an unused session dies.
- A visible session list with individual revoke, plus revoke-all-on-password-change.

### Request handling
- **Default deny.** The router wraps every route in the auth middleware; a handler must *opt out* to
  be public. A new endpoint is therefore private by construction, even if the author forgets. The two
  deliberate exceptions are login and the PWA's static assets.
- **CSRF**: `SameSite=Lax` plus an `Origin`/`Referer` check on every state-changing method. The API is
  JSON-only and rejects form content types, which removes the classic cross-origin form vector.
- **Input validation at the boundary**, into typed structs. Body size limits on every request.
- **Parameterized SQL everywhere.** No string-built queries, ever.
- **Security headers**: `Content-Security-Policy` (no inline script; the PWA is built to comply),
  `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: same-origin`, HSTS.

### Data
- **The SQLite file lives outside the web root**, at `/var/lib/logbook/`, owned by the service user
  and mode `0600`. Apache serves `/var/www/logbook` and can never reach it.
- **Automated backup before every migration and on a schedule**, with restores actually tested. Rule
  §0.2: this is a legal record.
- Logs record authentication outcomes but **never** passwords, session tokens, or cookie values.

### Process and supply chain
- **Keep the dependency tree near-empty.** Prefer the Go stdlib. Every dependency added must be
  justified in `APP.md`. The intended total is single digits: `modernc.org/sqlite`,
  `golang.org/x/crypto` (Argon2id), `go-pdf/fpdf`.
- **No secrets in the repo, ever** — enforced by review and by keeping config in environment
  variables read at startup.
- The service runs as a **dedicated unprivileged user**, never root, with systemd hardening
  (`NoNewPrivileges`, `ProtectSystem=strict`, `PrivateTmp`, `MemoryMax`). `MemoryMax` is a
  *containment* control as much as a resource one: this app must not be able to OOM the other sites.

## Action items

- [ ] **Rotate the `rami` sudo password.** It was pasted into a chat session on 2026-08-01 to allow
      server setup, which means it must be treated as compromised. Rotate it once deployment is done.
- [ ] Decide the fate of the publicly-exposed `:8000` container (see `deploy.md`).
- [ ] Prune stale ufw rules (`30814`, `19132`) and the duplicated Apache profile rules.

## Testing the controls

Each control maps to a test that fails if the control is removed:

| Control | Test |
|---|---|
| Default deny | Enumerate every registered route; assert each returns 401 without a session unless explicitly allow-listed. This catches a forgotten endpoint automatically. |
| Session expiry | Time-injected clock; assert a 91-day-old session is rejected and a used one rolls forward. |
| Token hashing | Assert the raw cookie value never appears in the DB. |
| CSRF | Assert a cross-origin `Origin` header is rejected on every mutating method. |
| Rate limiting | Assert the Nth failed login is throttled and that a success resets the counter. |
| Cookie flags | Assert `HttpOnly`, `Secure`, `SameSite` on the `Set-Cookie` header. |
| SQL injection | Table-driven hostile inputs against every filter parameter. |
| Password hashing | Assert stored hashes are Argon2id-encoded and that identical passwords yield different hashes. |
