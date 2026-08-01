# APP.md — the logbook application working tracker

Source of truth for **what we are building, what is done, and why**. Resume from here.
Rules live in the repo root **`CLAUDE.md` §0** — read those first, they are non-negotiable.

- Schema + domain rules: **`docs/data-model.md`**
- Threat model + controls: **`docs/security.md`**
- Server map + deployment: **`docs/deploy.md`**

> When you finish or change a task, update the **Task Board** and **Decision Log** below in the same
> change. Convert relative dates to absolute. A fix is "done" only when a test or a live check proves it.

---

## ★★ NEXT SESSION STARTS HERE

*Assume you remember nothing. This block is the whole brief.*

**What this is.** A private, mobile-first pilot logbook web app for one user (the repo owner, a
Finnish pilot), to be served at `ayoub.fi/logbook`. It replaces three CSVs in this repo that were
transcribed from paper logbooks by the *other* effort in this repo (`claude-docs/`, still ongoing and
not replaced by the app). Read `CLAUDE.md` §0 first — those rules are non-negotiable and were written
for this work specifically.

**Status: the whole backend is done and green — data, calculation core, API and authentication.
Nothing is deployed. There is no frontend yet, and that is the next task.**

### Done (2026-08-01)
- **Task 4 — the API and authentication.** `internal/stats` (aggregations + EASA pagination),
  `internal/auth` (Argon2id + session tokens), `internal/ratelimit` (login throttling),
  `internal/store/auth.go` (users + sessions), `cmd/server` (the router, middleware and operator
  CLI). **Every control in `docs/security.md` is implemented and has the test that fails if it is
  removed** — see the map at the bottom of that file. `make check` is green at **88.9%** overall
  with **six packages at 100%**.
  Verified live, not only in tests: the running server serves all nine frozen paper figures exactly.
- Recon of `ayoub.fi` → `docs/deploy.md` has the full shared-tenant map. **The box is shared with the
  owner's other sites; do not disturb them.**
- Stack chosen: Go + SQLite + React/Vite (Decision Log §5).
- `CLAUDE.md` §0 rules written, adapted from the neighbouring `transit` project.
- Server cleanup, both reversible: transit's orphaned Quarkus killed (it runs on its own VM now);
  OpenVPN stopped + disabled at the owner's request.
- **Task 2** — `app/backend/` Go module with `internal/hhmm` (H:MM ↔ integer minutes) and
  `internal/timeutil` (the single UTC-conversion authority). Both 100%, failing-test-first.
- **Task 3 — the schema and the importer.** `internal/csvbook` (CSV → domain, 100%),
  `internal/store` (SQLite schema + verified import, 85%), `cmd/logbookctl` (the operator CLI).
  **All 1293 flights import and verify.** `make check` is green at 94.5% overall.

### What exists in `app/backend/` — the whole map
```
cmd/logbookctl/      the operator CLI: `import` and `verify`. A separate binary from the
                     server on purpose, so a destructive op on a legal record can never be
                     reached over HTTP.
internal/hhmm/       H:MM <-> integer minutes. Durations are minutes everywhere inside the
                     app; H:MM is parsed at the edges only.            [core, 100%]
internal/timeutil/   THE single UTC-conversion authority. Do not re-implement time
                     conversion anywhere else (rule 0.4). Handles the Z suffix, the
                     midnight roll, and both DST edges.                [core, 100%]
internal/csvbook/    CSV -> domain records + the audit. Pure, no database. Skips the seed
                     rows, derives sea/land, reconciles all seven Cumulative_* series row
                     by row, emits Discrepancy values.                 [core, 100%]
internal/store/      schema.sql (embedded), the verified import, the read queries, and
                     auth.go: users and sessions.                          [81%]
internal/stats/      Summarize (the statistics page's twelve figures), Range/Filter, and
                     Paginate (the EASA page-totals block). Computed at call time,
                     never stored -- rule 0.5.                            [core, 100%]
internal/auth/       Argon2id passwords and session tokens. Knows nothing of HTTP or the
                     database, so the part that must be cryptographically right is
                     tested exhaustively on its own.                      [core, 100%]
internal/ratelimit/  Login throttling, per-IP and per-account, exponential backoff.
                     In memory, bounded, evicts the stalest key.          [core, 100%]
cmd/server/          The API and the operator CLI (createuser/passwd/users/disable/
                     enable). Table-driven router: a handler cannot be mounted without
                     the auth wrapper, and Routes() lets the test enumerate what is
                     really there.                                              [74%]
```
`make cover-core` enforces 100% on everything marked `[core]`. That list is the code where a
bug means a wrong legal record — **or an exposed one**: `internal/auth` and `internal/ratelimit`
are on it because in a credential primitive an untested branch is an authentication bypass.
`internal/store` is held to the 80% bar, not 100%, because it is I/O; `cmd/server` likewise,
where the untested remainder is `main()` and the listen loop.

### How to run things
```bash
export PATH=$HOME/.local/go/bin:$PATH   # Go 1.26 lives here; the system had none
cd app/backend
make check      # vet + race tests + both coverage gates. This is the bar.
make build      # cross-compiled static binaries into dist/ (builds every cmd/*)

# Import the CSVs. -dry-run reports and writes nothing; use it first.
go run ./cmd/logbookctl import -dry-run -csv ../..
go run ./cmd/logbookctl import -db /tmp/logbook.db -csv ../.. -note "why"
go run ./cmd/logbookctl verify -db /tmp/logbook.db -csv ../..

# Run the API locally, end to end.
./dist/server createuser rami -db /tmp/logbook.db        # prompts; needs a real terminal
./dist/server -db /tmp/logbook.db -addr 127.0.0.1:8099 \
              -origin http://localhost -insecure-cookie  # -insecure-cookie is dev-only
curl -s -c ck.txt -X POST localhost:8099/logbook/api/login \
     -H 'Content-Type: application/json' -H 'Origin: http://localhost' \
     -d '{"username":"rami","password":"..."}'
curl -s -b ck.txt localhost:8099/logbook/api/stats
```
⚠ **`go test` caches.** After the CSVs change, a green `make check` proves nothing until you have
run `go test -count=1 ./...` — this exact trap hid five real failures on 2026-08-01 (Decision Log).

⚠ **Run the server before calling a backend task done.** The whole suite was green while
`createuser -db` silently pointed at the production database. See the Decision Log entry.

The server's own Go is 1.13 and irrelevant — we cross-compile (`CGO_ENABLED=0`), producing an
11 MB static binary. Three direct dependencies, all justified in `docs/security.md`:
`modernc.org/sqlite` (pure Go, no CGO — that is what makes the static binary possible),
`golang.org/x/crypto` (Argon2id) and `golang.org/x/term` (no-echo password prompt). Keep it that
way (rule §0.3).

**The import is idempotent, backs up first (`VACUUM INTO`), and refuses to commit on any checksum
mismatch.** Re-run it freely. **There is no committed database** — it is generated, and
`app/.gitignore` keeps `*.db` and `*.bak` out of the repo. Nothing you need is only in a database
file; rebuild it from the CSVs in one command.

### The numbers the import produces — memorise these
```
flights 1293 | total 1219:35 | pic 1053:03 | dual 166:32 | instrument 107:05
night 22:45  | instructor 189:41 | seaplane 407:39 | landings 3439 | aircraft 38
```
**These now equal the figures inked at paper page 62, which the owner has frozen** — no change,
migration or app, may move them (`claude-docs/resume.md`). All seven `Cumulative_*` series reconcile
with **zero breaks**.

Asserted in `internal/csvbook/realdata_test.go` and again, by a different code path, in
`internal/stats/realdata_test.go` — along with the exact count of each discrepancy kind. **If one of them changes unexpectedly, the import is wrong until proven otherwise —
do not adjust the expectation to make the test pass.**

⚠ **The one legitimate reason for those tests to fail: `logbook_3.csv` is still growing.** The
migration effort (`claude-docs/`) appends flights to it page by page, and Book 3 is **not finished**.
When it grows, `realdata_test.go` fails — that is the test doing its job, not a regression. The
correct response is:

1. Run `go run ./cmd/logbookctl import -dry-run -csv ../..` and read the report.
2. Confirm the new totals are the *expected* deltas for the rows that were appended — `drift.md`
   records a per-page Δ for every batch, so cross-check against that, not against a feeling.
3. Only then update the constants, **in the same commit as the CSV change**, with the delta stated
   in the commit message.

Never update a constant first and reconcile afterwards. The whole value of these tests is that they
fail before anyone notices a wrong total.

### ⏸ THE DATA IS RECONCILED — one paper-side item is left, and it is not an app task
`claude-docs/resume.md` is the authority on this; do **not** re-validate the books on spec.

- **`logbook_2_final.csv` lines 89–90** (`04.05.2018` ×2), dated `DD.MM.YYYY`. Affects **row order
  only**, moves no total, and **no electronic source can settle it** (Aviatron holds zero OH-PDP
  rows; the club file starts 19/04/2020). Needs the physical page.
- Also paper-side: the **p.62 inked landing split** `59 night / 3335 day` recomputes to
  **`68 / 3326`**. The landing *sum* 3394 never moved. **Correct the paper, not the CSV.**

Everything else the importer surfaced on 2026-08-01 is closed and the owner has **frozen the
end-of-book-3 cumulatives**: night `22:45`, instrument `107:05`, and all seven `Cumulative_*` series
reconciling with zero breaks. Nothing in the app may move those figures.

### The API surface, as built
All under **`/logbook/api/`** — not `/api/`, which on `ayoub.fi` is taken by a stale transit proxy.
Durations are **integer minutes**; the frontend formats H:MM.

```
POST   /login              public   {username,password} -> 200 + Set-Cookie; 401 uniform; 429 throttled
GET    /health             public   exactly {"status":"ok"} and nothing else
POST   /logout             private  revokes this session, clears the cookie
GET    /me                 private  {user_id, username}
GET    /flights   ?from&to private  {flights:[...], count} in seq order
GET    /aircraft           private  the derived seed list for the new-flight form
GET    /stats     ?from&to private  {summary:{...twelve figures + landings_unverified}, range}
GET    /discrepancies      private  the "needs review" list, 61 rows today
GET    /sessions           private  the revocable device list; `current` marks the caller
DELETE /sessions/{id}      private  revoke one, scoped to the owner
```
`from`/`to` are inclusive `YYYY-MM-DD`. An unparseable one is a **400**, never an ignored filter.

**Operator CLI** (no HTTP route exists for any of it, by design):
`./dist/server createuser|passwd|users|disable|enable <name> -db <path>`.

### Next task: #5, the four frontend pages
Nothing in the backend blocks it. Read `docs/security.md` for the cookie contract before writing the
fetch layer — the session cookie is `HttpOnly`, so **JavaScript cannot read it and must not try**;
send `credentials: 'same-origin'` and let the browser carry it. Every mutating request needs an
`Origin` header, which the browser sets automatically for real fetches.

The four pages are in §2 above: **Table**, **Statistics**, **New flight**, **Export**. Two of the
four are read-only against endpoints that already exist. Note the gaps:

- **New flight has no endpoint yet.** `POST /flights` does not exist — writing to a legal record
  needs its own design pass (validation, the seq assignment, and how a hand-entered row is
  distinguished from an imported one in `source_book`/`source_row`). Do that deliberately.
- **Export (Task 6) has no endpoint yet** either. `stats.Paginate` is written and tested — it
  produces the EASA page blocks (15 rows, TOTAL THIS PAGE / PREVIOUS / TOTAL, 87 pages over the
  1293 flights) — but nothing renders a PDF.
- The frontend should surface **`landings_unverified`** wherever it shows night landings. It is 30
  today and it is how the app tells the truth about Task 8 rather than implying a verified split.

**The `discrepancies` table is already populated** (**61 rows**) and is what the frontend's "needs
review" list reads. It is rewritten on every import, so an item that gets resolved in the CSV simply
disappears — there is no second place to update.

### Open questions awaiting the owner
- Is the `kraken-predictor-python-2` container on `:8000` still wanted? Publicly exposed, up 2 years,
  and now the box's largest memory consumer (~759 MB / 38%).
- Prune the stale ufw rules for `30814` and `19132` (nothing listens on either)?
- **Rotate the `rami` sudo password** once deployment is done — it was pasted into a chat session on
  2026-08-01 and must be treated as compromised. Tracked in `docs/security.md`.

### Traps already paid for — do not rediscover these
- **It is 1293 flights, not 1295.** 1295 is the CSV *row* count; Books 2 and 3 each open with the
  previous book's final row as a cumulative seed, and those two are skipped. Earlier drafts of this
  file said 1295.
- **Sea vs land comes from the registration, not the type.** The book only started writing `C172sea`
  from IMG_6022 and is inconsistent after that. Verified: the registration rule reproduces
  `Cumulative_SEP_Sea` row by row at all 1293 rows.
- **The books are not in date order** — 18 rows go backwards. Order on `seq`, never on `flight_date`.
- **Go's `time.Date` is silent on both DST edges, in different ways.** See the 2026-08-01 entry in the
  Decision Log. `internal/timeutil` already handles it; do not "simplify" that check.
- **Docker bypasses ufw.** A published container port is not closed by a firewall rule. See
  `docs/deploy.md`.
- **`/api/` on `ayoub.fi` is already taken** by a stale transit proxy, which is why our API lives at
  `/logbook/api/`.
- **Port 22 is under constant attack** (fail2ban: 50,264 bans). Never risk it.
- **`go test` caches, so a green `make check` can be a lie** after the CSVs move. Use `-count=1`.
  This hid five real failures on 2026-08-01.
- **Go's `flag` package stops parsing at the first non-flag argument**, so a flag written after a
  positional is silently dropped. `cmd/server` parses in a loop (`parseFlagsAnywhere`) because
  `createuser rami -db /tmp/x.db` was otherwise aiming at the production database.
- **A green test suite is not a substitute for running the server.** The bug above survived 22
  passing HTTP tests and died in the first thirty seconds of a live smoke test.

### Where the reasoning lives
Do not re-derive these — they are argued out in the Decision Log below (§5), all dated 2026-08-01:
the stale test cache · **table-driven default deny** · **Argon2id at 19 MiB and why not 64** ·
**the decoy hash on the unknown-user path** · **sessions as rows, token returned with its hash** ·
**the rate limiter's stalest-key eviction** · **minutes on the wire, and 500 over an empty 200** ·
**the smoke test that caught the `-db` bug** ·
stack choice · cumulatives computed not stored · the time model · the EASA PDF covering all three
books · the landings day/night gap · the server security findings · Go's `time.Date` DST behaviour ·
**the two-verifications design** · **sea/land from the registration** · **the derived aircraft list**
· **the three open source-data problems**.

---

## 1. Product in one paragraph

A private, mobile-first pilot logbook for one user (the owner), replacing a stack of paper books and
three CSVs. It stores every flight as structured data in UTC, computes all totals on demand, and
produces three PDFs: an **EASA-format clone** of the whole logbook for the authority, a detailed
table export, and a statistics export. It is served at `ayoub.fi/logbook` behind authentication, and
installs to the phone home screen as a PWA because it is used in the field.

## 2. Scope — v1

Four pages:
1. **Table** — all flights, selectable date range.
2. **Statistics** — every aggregate below, over a From–To range.
3. **New flight** — form with aircraft preselect driving the sea/land default (user-overridable).
4. **Export** — the three PDFs.

Statistics must report, for the selected range: seaplane PIC · seaplane instructor · landplane PIC ·
landplane instructor · dual · total · night · instrument · landings sea · landings land · landings
day · landings night.

**Not in v1**: self-service registration (the user is created via CLI), multi-user sharing, offline
*writes*. Authentication supports adding more users later by design.

## 3. Stack

| Layer | Choice | Why |
|---|---|---|
| Backend | Go, stdlib `net/http` | Single static binary; deploy is rsync + restart. Tiny dependency tree. |
| DB | SQLite (`modernc.org/sqlite`, pure Go) | No CGO ⇒ trivial cross-compile. One file; backup = `VACUUM INTO`. 1293 rows is nothing. |
| Time | embedded `tzdata` | Behaviour must not depend on the server's zoneinfo. |
| PDF | `go-pdf/fpdf` | Absolute positioning, which a fixed 15-row EASA grid needs. Headless Chrome would cost 300 MB+. |
| Frontend | React + TS + Vite | Builds to static files. Node is build-time only, never on the server. |

## 4. Task Board

| # | Task | Status |
|---|---|---|
| 1 | Project rules + app docs | **done** 2026-08-01 |
| 2 | Scaffold backend + frontend, test harness | **backend done** 2026-08-01; frontend not started |
| 3 | Schema + importer for 1293 flights (verified) | **done** 2026-08-01 |
| 4 | API + authentication | **done** 2026-08-01 — `internal/stats`, `internal/auth`, `internal/ratelimit`, `store/auth.go`, `cmd/server`. Every `docs/security.md` control implemented with the test that fails if it is removed. Verified live against the real 1293 flights. |
| 5 | Four frontend pages (mobile-first) | not started |
| 6 | Three PDF exports (EASA clone + table + stats) | not started |
| 7 | PWA + deploy to `ayoub.fi/logbook` | not started |
| 8 | Backfill landings day/night for the **30** night rows | not started — all 30 are flagged `landings_unverified` in the DB, listed by `logbookctl import`, and surfaced by the API as `landings_unverified` in the stats summary. `claude-docs/drift.md` has the analysis. The p.62 split recomputes to **68 night / 3326 day** (the sum 3394 is unchanged); six of those 68 are estimates from three multi-landing partial-night rows, range 65–72. |
| 9 | Rule on the open source-data problems | **mostly closed** 2026-08-01 — two of the three ruled and fixed. One item left (`logbook_2_final.csv` lines 89–90) and it needs the physical page; it moves no total. See the ⏸ block at the top. |

---

## 5. Decision Log

### 2026-08-01 — `make check` was green on a stale test cache, and the suite was actually red

Worth recording as a process trap, not just a fix. `make check` reported green; forcing
`go test -count=1` showed **five failures**. The data-validation effort had changed the CSVs and the
expectation constants in `internal/csvbook/realdata_test.go` were never re-synced.

Every delta turned out to be an expected, owner-ruled CSV correction: instrument 107:14 → **107:05**
(line 28, `1:21`→`1:12`), night 20:50 → **22:45** (the p.52/53 photograph, +0:55 +1:00),
`cumulative_break` and `component_exceeds_total` both 1 → **0** (the same line-28 fix), and
`landings_unverified` 28 → **30** (the two new night rows; 30 is now every night row in the books,
20+3+7). Nothing was unexplained and no expectation was relaxed to make a test pass.

Two things changed as a result. The cumulative-break test now asserts **zero breaks over 1293 rows
and seven series** rather than "exactly the one known defect" — the corrected data supports the
stronger claim. And the two closed discrepancy kinds stay in the map **at 0** rather than being
deleted, because the "unexpected kind" sweep only catches kinds that were never listed.

**The lesson: `go test` caches, and a green `make check` is not evidence on its own after the CSVs
move.** Run `-count=1` after any migration batch. `logbookctl import -dry-run` is the other half.

### 2026-08-01 — Task 4: default deny is enforced by the router's shape, not by discipline

The obvious way to write this is `mux.HandleFunc("GET /x", authRequired(handleX))`, and the obvious
failure mode is the day someone writes `mux.HandleFunc("GET /y", handleY)`. Nothing catches it: the
endpoint works, the tests pass, and the logbook is public.

So routes are registered through a table and the registration function applies the wrapper. There is
no way to mount a handler without going through it, and `public` is a field the author has to set —
private is what you get by doing nothing. `Server.Routes()` then exposes the table, and the test
enumerates **what is actually mounted** and asserts 401 on everything not on a two-entry allow-list.
Adding a public endpoint means editing that allow-list, in the test file, deliberately.

The backstop behind it: `callerOf` **panics** if a handler runs without a session on the context. If
default deny were ever circumvented, the process crashes rather than serving a zero-value user
somebody's logbook. A loud failure beats a quiet one on a legal record.

Rejected: a middleware that inspects the path with a prefix rule (`/public/...`). Path conventions
are exactly the kind of thing that gets refactored by someone who does not know they are load-bearing.

### 2026-08-01 — Argon2id at 19 MiB, not 64 MiB, because the box is shared

OWASP recommends two Argon2id parameter sets. The heavier is m=64 MiB, t=3; the lighter is
m=19 MiB, t=2, p=1. We take the lighter one, which is the opposite of the usual instinct.

The reason is rule §0.3: this app shares a 1 vCPU / 2 GB droplet with the owner's other sites and a
fault here must not reach them. At 64 MiB an attacker turns the login form into a memory-pressure
lever against the whole box — and the login form is the one endpoint that is reachable without
credentials. `p=1` because one lane matches one vCPU; extra parallelism buys nothing here.

This is safe to be wrong about later: the parameters are encoded in every hash, so raising them
never invalidates an existing password, and `NeedsRehash` upgrades old hashes at the next successful
login while the plaintext is in hand.

### 2026-08-01 — The unknown-user login path pays for a full hash

`Authenticate` returns one `ErrAuthFailed` for every cause. That is necessary but not sufficient: if
a missing username returns before hashing, it answers in microseconds where a real account takes
tens of milliseconds. Uniform text, non-uniform timing — a username oracle readable with a stopwatch.

So the no-such-user path verifies against a decoy hash generated at package init and throws the
result away. There is a test that measures both paths and fails if the missing-user path gets cheap.
The decoy is generated rather than checked in, so it cannot be mistaken for a credential and cannot
go stale when `DefaultParams` is raised.

### 2026-08-01 — Sessions are rows, and the raw token is returned *with* its hash

Rows rather than JWTs, for revocation: the owner wants 90-day sessions, and a signed token that
cannot be withdrawn makes a stolen cookie a 90-day liability. A row can be deleted.

The API shape is the interesting part. `auth.NewSessionToken()` returns **both** the raw token and
its hash. The caller writing the session row is handed the hash and has no reason to reach for the
raw value, so "the database never stores a usable token" is a property of the interface rather than
something a future change has to remember. The test asserts it against the bytes on disk, not
against the code that wrote them.

Expiry is evaluated in Go against an injectable clock rather than in SQL against `CURRENT_TIMESTAMP`.
One authority for time (rule §0.4), and the 90-day window gets tested in milliseconds instead of
being taken on trust.

### 2026-08-01 — The rate limiter evicts the *stalest* key, which is never the attacker's

An in-memory table keyed by IP and account is an obvious memory-exhaustion target: rotate source
addresses and grow it without bound. So it is capped — but the eviction policy is the part worth
recording. Evicting the **least recently active** entry means an attacker can never evict their own
penalty, because the key they are hammering is by definition the most recently touched. Flooding the
table cannot flush your own lockout.

Two other details. The backoff shift is clamped **before** it is taken: at 63 doublings a
`time.Duration` wraps to a negative number, which reads as "no penalty" — the exact inversion of the
function's purpose. And a throttled account is refused **even with the correct password**, or the
limiter is bypassed by guessing right.

In memory rather than in SQLite because a row per failed login lets an attacker drive disk I/O for
free, and the state is worthless across a restart they cannot cause.

### 2026-08-01 — The wire format is minutes, and a 500 beats an empty 200

Durations cross the API as **integer minutes**, the same representation used everywhere inside the
app; the frontend formats H:MM. Returning both minutes and a preformatted string was rejected: two
representations of one figure is two things that can disagree, and this is a legal record.

Blank paper cells serialise as JSON `null`, not as a zero `time.Time` — which renders as year 1 and
would eventually be read as a real time. The raw string as written and the `time_origin` flag travel
alongside the converted instant, so a bad DST guess stays auditable (rule §0.4).

And a failed read is a **500**, never a 200 with an empty list. An empty list reads as "you have
flown nothing", which is the silent corruption rule §0.2 forbids — the honest failure is louder than
the convenient one. The server likewise refuses to start against a database with no flights in it.

### 2026-08-01 — A smoke test caught what 22 HTTP tests did not

The full suite was green and the server had never been run. Running it found that
`logbook-server createuser rami -db /tmp/x.db` **silently ignored `-db`** and reached for the
production database: subcommands were dispatched with `fs.Parse(nil)`, and even after a first fix,
Go's `flag` package stops parsing at the first non-flag argument, so a flag written *after* a
positional is never seen. On this project that is the wrong direction to be wrong in — an operator
command aimed at the live legal record instead of a scratch file.

Fixed with a parse loop that pulls flags from any position, and a regression test covering all three
orderings including the one that broke. **The lesson is the general one: a green suite is not a
substitute for running the thing.** The live check against the real 1293 flights is now part of what
"done" means for a backend task.

### 2026-08-01 — Task 3: the import verifies twice, on two different questions

The importer answers two questions that were tempting to conflate, and treats them differently.

**Fidelity — is the database what the CSVs say?** Nine checksums (flights, total, PIC, dual,
instrument, night, instructor, seaplane, landings) plus the row count are *read back out of SQLite*
after writing and compared to what the CSVs produced. One minute of disagreement rolls the whole
transaction back. Read back rather than trusted, because a CHECK constraint, a type coercion or a
truncated value would otherwise pass unnoticed. Checked per figure rather than as one grand total,
because two errors of opposite sign cancel in a combined number.

**Consistency — does the source agree with itself?** All seven `Cumulative_*` series are recomputed
row by row and compared to the columns the transcription maintained. A break here is **reported, not
fatal.** Refusing to import over a pre-existing property of the paper record would leave the owner
with no application at all, and rule §0.2 asks for discrepancies to be surfaced for the owner to
rule on — not for the importer to have a veto.

The row-by-row form matters on both: an end-total can be passed by two cancelling errors, and a break
with no line number is not actionable.

**Result: 1293 flights, 39 aircraft, 56 discrepancies, all nine checksums matching.** Exactly one
cumulative break survives across 1293 rows and seven series.
*(Figures as of this entry. After the 2026-08-01 owner rulings they are **38 aircraft, 61
discrepancies, and ZERO cumulative breaks** — the one break was Book 1 line 28, which the owner
ruled on and the CSV was corrected. The test now asserts zero.)*

### 2026-08-01 — Sea/land comes from the registration, and it is verified rather than assumed

The CSVs have no class column. `reference.md` gives a seaplane registration list and warns that the
book's own `C172sea` marker only appears from IMG_6022 and is inconsistent after that, so the type is
not usable.

Classifying on the registration turns out to be provable: recomputing `Cumulative_SEP_Sea` row by row
from that rule reproduces the column **exactly at every one of the 1293 rows**, ending on 407:39. A
per-row match over 1293 rows pins each individual row's class, not merely the total. That is a
stronger guarantee than the rule started with, and it is asserted in the tests.

### 2026-08-01 — The aircraft seed list is derived, never hand-maintained

`reference.md` warns, in its own words, that its hand-kept registration and place lists "are NOT
derived from the CSVs and they have gaps" — `EFSA` was missing despite six flights. So the app's
`aircraft` table is built from the flights on import: `type` is the most-flown type for that
registration, `default_class` from the seaplane list, `active` = flown within two years.

Two deliberate details. **`active` counts back from the last flight in the books, not from today** —
otherwise the same CSVs would import differently next year and idempotence would be a lie. And
**`ifr_capable` is a curated set (`OH-CAM`, `OH-ESR`, `OH-PIF`) rather than "has logged instrument
time"**, because instrument time is also logged under the hood: `OH-COF` and `OH-CTH` are C152s with
instrument rows. It is a hint for the form and never constrains what a flight may record.

### 2026-08-01 — Three source-data problems found, none corrected

The reconciliation swept all 1293 rows and found three things nobody had logged. All are recorded in
`claude-docs/drift.md` and `docs/data-model.md`, and all are the owner's to rule on (rule §0.2).

1. **`logbook_1_final.csv` line 28** — `Instrument_Time` 1:21 on a flight totalling 1:12. Impossible;
   the cumulative column advances by 1:12, so the row is the outlier.
   **✅ Closed the same day:** the owner ruled 1:12 and the CSV was fixed. Instrument 107:14 →
   **107:05**, which is what the column always said, so no cumulative moved. This was the corpus's
   only `cumulative_break` and only `component_exceeds_total`; both are now **zero**.
2. **`logbook_2_final.csv` lines 83–90** — dates written `DD.MM.YYYY`. Read day-first, which six of
   the eight settle themselves and the chronological bracket confirms; the two `04.05.2018` rows are
   flagged for a look at the paper.
3. **Night time 16:47 (ours) vs 22:45 (inked at p.62)** — a 5:58 gap on the one p.62 figure that
   `drift.md` records as never having been read back.
   **✅ Closed the same day.** The importer's job here was only to surface the gap; the owner then
   read the paper's night column back and photographed seven Book-1 spreads, which turned it into a
   page-by-page ledger (16:47 → 20:50), and the p.52/53 photograph closed the last 1:55.
   **Night is 22:45 = the paper, Δ 0:00.** *The flag was worth raising precisely because nobody had
   ever compared that column.*

The dotted dates are the interesting judgement call. Refusing would have blocked 1291 sound rows over
a separator; silently normalising would have hidden a real inconsistency. Accepting with a loud,
per-row flag — and a louder one on the two ambiguous rows — is what "surface, never fix" means when
the alternative is delivering nothing.

Rejected: using chronology to disambiguate. **18 rows across the three books are genuinely out of
date order**, so "later than its predecessor" is not an invariant these books have. That is also why
the schema orders on `seq` and never on `flight_date`.

### 2026-08-01 — Stack chosen: Go + SQLite + React/Vite

Recon of `ayoub.fi` found a 1 vCPU / 2 GB droplet with **141 MB available** — transit's orphaned
Quarkus process (started 2026-05-23, PPID 1, no systemd unit) was holding 605 MB. The user confirmed
transit now runs on its own VM and authorized killing it; memory went **141 MB → 738 MB**. Stopping
OpenVPN later took it to ~722 MB with more headroom in cache.

That removed the hard constraint, so the stack is a choice on merit rather than desperation — but the
reasoning is unchanged: a Go binary with SQLite sits at ~25 MB RSS against ~100–150 MB for a Node
backend, the deploy artifact is one file, and the near-empty dependency tree is the single biggest
security lever available (rule §0.3). It also matches the user's stated preference.

**Rejected**: Node backend (memory, supply chain); Java/Quarkus (600 MB — what we just removed);
the box's existing Postgres — *which turned out not to exist at all*, see below.

### 2026-08-01 — Cumulatives become computed, not stored

The CSVs carry seven `Cumulative_*` columns and they are the single largest source of drift in this
project's history (`claude-docs/drift.md` is 106 KB, much of it cumulative corrections). The app
derives them from an explicit `seq` ordering instead. The EASA PDF needs per-page running totals —
those are computed at render time. Locked as rule §0.5.

### 2026-08-01 — Time model: convert to UTC, keep the raw, flag the origin

The paper books mix local and UTC (`Z` suffix = already UTC; an `LT` subscript sometimes marks local,
but its absence proves nothing — see `claude-docs/reference.md`). The user's rule is that everything
is UTC from now on.

Chosen: store canonical `*_utc`, **plus** the raw string exactly as written on paper, **plus** a
`time_origin` flag (`utc_as_written` / `converted_from_local` / `unknown`). Conversion is
`Europe/Helsinki` with correct historical DST, in exactly one function. Ambiguous rows surface in a
"needs review" list rather than being guessed at. Rejected "convert and discard the raw" — a bad DST
guess would then be unauditable and unrecoverable on a legal record.

### 2026-08-01 — EASA PDF covers all 1293 flights, not just Book 3

Books 1 and 2 are an older, non-EASA paper format; only Book 3 is EASA. The user chose a single
continuous EASA-format logbook over all three books (~87 pages at 15 rows/page), because that is what
an authority actually wants: one complete record in the current standard format.

**Layout confirmed from `logbook-3/IMG_6025.JPEG` and by the user**: **15 flight rows per page**, then
a 3-row totals block (TOTAL THIS PAGE / TOTAL PREVIOUS PAGES / TOTAL), a "Certified true and correct"
signature line, and "Page _ of 128". Columns: GENERAL (date, dep+off-block, arr+on-block, type, reg,
PIC name) · FLIGHT TIME (total, night, SE-VFR, SE-IFR, ME-VFR, ME-IFR, PIC, co-pilot, multi-pilot,
flight instructor, dual, instructor-STD) · OTHER (landings day/night, remarks).

### 2026-08-01 — Landings day/night is a real data gap

The EASA book **does** split LANDINGS into DAY and NIGHT; our 26-column CSV only ever stored the sum
(`claude-docs/reference.md` says the split would be "inferred later from `Night_Time`"). The stats
page needs both. Schema therefore carries `landings_day` + `landings_night` + a `landings_verified`
flag; the importer seeds everything as day and flags the rows carrying `Night_Time` for backfill
from the page images (Task 8). That was 22 rows when this was written and is **30** now, after the
night reconciliation of the same day. Bounded and small.

### 2026-08-01 — Server security: findings, and one correction

Initial recon reported a "publicly exposed Postgres on 5432". **That was wrong** — there is no
Postgres on the box. Port 5432 was the user's **OpenVPN server**, deliberately bound there (`local
164.90.195.106 / port 5432 / proto tcp`) to disguise VPN traffic as database traffic. Closing it
would have broken the VPN. Investigating before acting is what caught it.

Actual posture is sound: ufw active with default-deny incoming; fail2ban active with an `sshd` jail
plus four apache jails, and it is doing real work — **412,570 failed SSH attempts, 50,264 bans**.

The user then asked for OpenVPN to be stopped until they next travel: `systemctl stop` +
`disable openvpn-server@server`, config left intact. Re-enable with
`sudo systemctl enable --now openvpn-server@server`. Its ufw rule for 5432 was deliberately left in
place so re-enabling is a single command.

Outstanding, not yet actioned: the publicly-exposed `:8000` container, and stale ufw rules for
`30814` / `19132` (nothing listens on either).

**Gotcha recorded for the future**: Docker publishes ports by writing its own iptables `DOCKER` chain,
which is evaluated *before* ufw's INPUT rules — so `ufw deny` does **not** block a published container
port. The `:8000` container is published `0.0.0.0:8000`; if it ever needs closing, fix the port
binding, not the firewall.

### 2026-08-01 — Go's `time.Date` does not signal DST trouble, so we detect it ourselves

Building `internal/timeutil` surfaced a trap worth recording. Go's `time.Date` handles the two DST
edge cases silently and differently:

- **Spring gap** (a wall clock that never existed): it *normalizes*. Asking for 03:30 on 2024-03-31
  returns 04:30 with no error. Detected by checking whether the returned value still reads back as
  what we asked for.
- **Autumn fold** (a wall clock that happened twice): it returns one of the two instants and the
  documentation explicitly does not guarantee which. An empirical probe showed it picking the
  **later** offset (EET), the opposite of this implementation's first assumption — which is exactly
  how the first version of the fold check passed review and still failed its test.

The fold check therefore probes an hour in **both** directions and flags ambiguity if either shift
reads back as the same wall clock. That is correct regardless of which offset Go picks, so the code
does not depend on undocumented behaviour.

Both cases yield `time_origin = unknown` and surface for review rather than being guessed at.
