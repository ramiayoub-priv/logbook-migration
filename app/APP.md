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

**Status: feature-complete for v1 and green. Deployment (Task 7) is HALF DONE and is the live task —
the backend runs on the box, Apache has not been switched on, so `ayoub.fi/logbook` is still 404 to
the world. Read "Where the deploy actually stands" below before touching anything.**

### Done (2026-08-01)
- **Task 2** — `app/backend/` Go module, `internal/hhmm` and `internal/timeutil`. Both 100%.
- **Task 3** — the schema and importer. All **1293** flights import and verify.
- **Task 4** — the API and authentication. Every control in `docs/security.md` has the test that
  fails if it is removed.
- **Task 5b** — **`POST /flights`**, the only write path into the legal record. `internal/entry`
  validates (pure, 100%); `store.AddFlight` allocates book order. See the decision log: the load-
  bearing part is that a hand-entered flight **survives the next CSV re-import**.
- **Task 6** — the **three PDFs**. `internal/pdfmodel` (cells and totals, pure, 100%) +
  `internal/pdfbook` (rendering). Verified against the real logbook: 87 EASA pages, totals block
  reconciling, Finnish place names intact.
- **Task 5** — the **frontend**, `app/frontend/`. Six pages behind a login gate. 43 tests green,
  and driven in a real browser against the live API — including logging a flight end to end,
  watching the duplicate be refused, and confirming zero horizontal overflow on a 390px phone.

### The whole map
```
app/backend/
  cmd/logbookctl/    the operator CLI: `import` and `verify`. Separate binary from the server on
                     purpose, so a destructive op on a legal record cannot be reached over HTTP.
  cmd/server/        the API, the export handlers, and the operator CLI (createuser/passwd/users/
                     disable/enable). Table-driven router: a handler cannot be mounted without the
                     auth wrapper, and Routes() lets the test enumerate what is really there.  [76%]
  internal/hhmm/       H:MM <-> minutes. Minutes everywhere inside; H:MM at the edges.  [core, 100%]
  internal/timeutil/   THE single UTC-conversion authority. Do not re-implement time
                       conversion anywhere else (rule 0.4).                            [core, 100%]
  internal/csvbook/    CSV -> domain + the audit. Pure, no database.                   [core, 100%]
  internal/entry/      Validates a HAND-TYPED flight. Pure. The opposite posture to csvbook:
                       it refuses rather than surfaces -- see the decision log.        [core, 100%]
  internal/store/      schema.sql, the verified import, the read queries, auth.go, and
                       handentry.go (AddFlight + the seq band + the reimport relink).      [83%]
  internal/stats/      Summarize, Range/Filter, Paginate. Computed, never stored.      [core, 100%]
  internal/pdfmodel/   Every cell and every total of the three PDFs. Pure -- rule 0.6
                       names "PDF totals" as calculation core.                         [core, 100%]
  internal/pdfbook/    Draws them with go-pdf/fpdf.                                        [~95%]
  internal/auth/       Argon2id + session tokens. Knows nothing of HTTP or the DB.     [core, 100%]
  internal/ratelimit/  Login throttling, per-IP and per-account.                       [core, 100%]

app/frontend/
  src/api.ts         the fetch layer. credentials:'same-origin' and NOTHING else -- the cookie is
                     HttpOnly, so JavaScript cannot read it and must not try.
  src/auth.tsx       who is signed in (asked of the server, never cached) + useApi. Any 401
                     anywhere drops the app to the login page.
  src/router.tsx     ~40 lines instead of a routing library (rule 0.3).
  src/format.ts      H:MM and UTC dates. The ONLY place minutes become H:MM.
  src/pages/         Login, Table, Statistics, NewFlight, Export, Review, Sessions, RangePicker.
```
`make cover-core` enforces 100% on everything marked `[core]` — the code where a bug means a wrong
legal record, or an exposed one.

### How to run things
```bash
export PATH=$HOME/.local/go/bin:$PATH   # Go 1.26 lives here; the system had none

cd app/backend
make check      # vet + race tests + both coverage gates. This is the bar.
make build      # static binaries into dist/ (builds every cmd/*)

cd app/frontend
npm install
npm run check   # tsc --noEmit + vitest
npm run build   # static files into dist/

# --- Trying things out safely -------------------------------------------------
# THE ISOLATION BOUNDARY IS THE DATABASE FILE, NOT THE ACCOUNT. This app is
# single-tenant: `flights` has no owner column, so a second user account writes
# into the same logbook. Use a scratch file instead.
cd app/backend
make scratch                                    # rebuilds /tmp/logbook-scratch.db from the CSVs
./dist/server createuser ramitest -db /tmp/logbook-scratch.db    # needs a real terminal
./dist/server -db /tmp/logbook-scratch.db -addr 127.0.0.1:8099 \
              -origin http://localhost:5173 -insecure-cookie -holder "Rami Ayoub"
cd ../frontend && npm run dev                   # http://localhost:5173/logbook/
make scratch-clean                              # throw it away

# Import the real CSVs. -dry-run reports and writes nothing; use it first.
go run ./cmd/logbookctl import -dry-run -csv ../..
go run ./cmd/logbookctl verify  -db <path> -csv ../..
```

⚠ **`-origin` must match exactly what the browser sends.** Against `npm run dev` that is
`http://localhost:5173`, not `http://localhost`. Getting it wrong makes login fail with a 403 and
nothing else — the CSRF check doing its job. This cost real time on 2026-08-01.

⚠ **`go test` caches.** After the CSVs change, a green `make check` proves nothing until you have
run `go test -count=1 ./...` — this exact trap hid five real failures on 2026-08-01.

⚠ **Run the thing before calling a task done.** A green suite has now three times missed what
thirty seconds of running found: the `createuser -db` bug in Task 4; in Task 5/6 the broken PDF
column headers, the clipped totals labels, the date fields overflowing a phone, and the aircraft
relink lost on re-import; and on 2026-08-01 **the new-flight form asking for `09:15Z` in a field
whose keyboard has no colon key** — untypeable on a phone, invisible to 43 passing tests, a browser
run at 390px, and a live end-to-end flight entry, because all of those type with a desktop keyboard
or programmatically. **On a mobile form, test the keyboard, not just the field.**

**There is no committed database** — it is generated, and `app/.gitignore` keeps `*.db` and `*.bak`
out of the repo. Nothing you need is only in a database file; rebuild it from the CSVs in one
command.

### The numbers the import produces — memorise these
```
flights 1296 | total 1222:10 | pic 1054:45 | dual 167:25 | instrument 107:58
night 22:45  | instructor 189:41 | seaplane 407:39 | landings 3444 | aircraft 38
discrepancies 61 | EASA export 87 pages
```
All seven `Cumulative_*` series reconcile with **zero breaks**.

⚠ **These moved on 2026-08-01** and the previous values are still all over the git history: they
were `1293 / 1219:35 / 1053:03 / 166:32 / 107:05 / 3439`. Three flights of **28/08/2025** were
missing from `logbook_3.csv` entirely — see the decision log below and `claude-docs/drift.md`.
Outside that one owner-ruled exception the figures remain **frozen**: no change, migration or app,
may move them (`claude-docs/resume.md`).

Asserted in `internal/csvbook/realdata_test.go` and again, by a different code path, in
`internal/stats/realdata_test.go`. **If one of them changes unexpectedly, the import is wrong until
proven otherwise — do not adjust the expectation to make the test pass.**

⚠ **The one legitimate reason for those tests to fail: `logbook_3.csv` is still growing.** Book 3 is
**not finished**. When it grows, `realdata_test.go` fails — that is the test doing its job. The
correct response is:

1. Run `go run ./cmd/logbookctl import -dry-run -csv ../..` and read the report.
2. Confirm the new totals are the *expected* deltas — `drift.md` records a per-page Δ for every
   batch, so cross-check against that, not against a feeling.
3. Only then update the constants, **in the same commit as the CSV change**, with the delta stated
   in the commit message.

Never update a constant first and reconcile afterwards.

### ⏸ THE DATA IS RECONCILED — one paper-side item is left, and it is not an app task
`claude-docs/resume.md` is the authority; do **not** re-validate the books on spec.

- **`logbook_2_final.csv` lines 89–90** (`04.05.2018` ×2), dated `DD.MM.YYYY`. Affects **row order
  only**, moves no total, and **no electronic source can settle it**. Needs the physical page.
- Also paper-side: the **p.62 inked landing split** `59 night / 3335 day` recomputes to
  **`68 / 3326`**. The landing *sum* 3394 never moved. **Correct the paper, not the CSV.**

### The API surface, as built
All under **`/logbook/api/`** — not `/api/`, which on `ayoub.fi` is taken by a stale transit proxy.
Durations are **integer minutes**; the frontend formats H:MM.

```
POST   /login              public   {username,password} -> 200 + Set-Cookie; 401 uniform; 429 throttled
GET    /health             public   exactly {"status":"ok"} and nothing else
POST   /logout             private  revokes this session, clears the cookie
GET    /me                 private  {user_id, username}
GET    /flights   ?from&to private  {flights:[...], count} in seq order
POST   /flights            private  a hand-entered flight -> 201; 400 with per-field errors; 409 duplicate
                                    times are "HH:MM" (Helsinki local) or "HH:MMZ" (UTC)
                                    takeoff/landing are OPTIONAL, but all-or-nothing as a pair
GET    /aircraft           private  the derived seed list for the new-flight form
GET    /stats     ?from&to private  {summary:{...}, range}
GET    /discrepancies      private  the "needs review" list, 61 rows today
GET    /sessions           private  the revocable device list; `current` marks the caller
DELETE /sessions/{id}      private  revoke one, scoped to the owner
GET    /export/easa.pdf        private  the whole logbook, EASA format. IGNORES from/to on purpose.
GET    /export/table.pdf       ?from&to private
GET    /export/statistics.pdf  ?from&to private
```
`from`/`to` are inclusive `YYYY-MM-DD`. An unparseable one is a **400**, never an ignored filter.

**Operator CLI** (no HTTP route exists for any of it, by design):
`./dist/server createuser|passwd|users|disable|enable <name> -db <path>`.

### Where the deploy actually stands (2026-08-01)

**Before touching the box, re-read `CLAUDE.md` §0.3.** It is shared with the owner's other sites;
changes to Apache, ufw, systemd or Docker are additive, reversible, and verified before the first
connection is closed. **Never risk port 22.** Nothing in this deploy touches ufw, sshd or Docker —
the backend binds `127.0.0.1:9002`, so no firewall change is needed at any step.

**`rami` has no passwordless sudo on the box.** Every privileged step is a command the *owner* runs;
a session cannot do it unattended. Read-only survey over SSH works fine with the existing key.

**The site is LIVE**: `https://ayoub.fi/logbook` answers 200, the API answers 200 on `/health` and
401 unauthenticated, and all seven of the owner's other sites still answer 200. The owner ran
`install-apache.sh` and created an account.

⚠ **But the deployment is BEHIND the repo.** The box is running the pre-2026-08-01-evening binary
and a **1293-flight database**. Both are staged and waiting on one command — see below.

✅ **Done and verified live:**
- `logbook` system user; `/opt/logbook`, `/var/lib/logbook` (0750), `/var/www/logbook`.
- The static binary at `/opt/logbook/logbook-server`, cross-compiled `CGO_ENABLED=0`.
- `logbook.service` **enabled and running**, 21.4 MB RSS against a 192 MB `MemoryMax`.
- Health `200` returning exactly `{"status":"ok"}`, and `/flights` without a session `401` — default
  deny survived the deploy.
- The frontend build rsynced to `/var/www/logbook/` (`rami` owns it, so no sudo), assets carrying
  the `/logbook/` base.

✅ **Apache** — `a2enmod headers` plus the additive block from `docs/deploy.md`, configtest, reload.
✅ **The account** — created interactively with `createuser`. The password has never been in a file
or a chat session and must stay that way.

⏳ **THE ONE COMMAND STILL OUTSTANDING** — the owner must run it; there is no passwordless sudo:

```bash
ssh -t rami@ayoub.fi 'sudo /home/rami/logbook-deploy/update.sh'
# then, from the dev machine (no sudo -- rami owns the web root):
rsync -a --delete app/frontend/dist/ rami@ayoub.fi:/var/www/logbook/
```

`update.sh` installs the new binary (keeping `.prev` for rollback), backs the database up to
`/var/lib/logbook/backups/`, **re-imports** the corrected CSVs, verifies by a separate code path,
restarts, and re-checks the other six sites. **The startup log line must read `flights=1296`.**

⚠ **Re-import, never a file swap.** The `users` and `sessions` tables live in the same SQLite file,
so replacing it would delete the owner's account. The importer's `DELETE` is scoped to `aircraft`,
`discrepancies` and `flights WHERE source_book <> 0`, inside one transaction that rolls back on any
checksum mismatch — so users, sessions and app-entered flights all survive.

⚠ **The binary and the frontend must land together.** The reworked form sends `takeoff`/`landing`;
Go's JSON decoder ignores unknown fields, so the *old* binary would accept the flight and **silently
drop two of its times**. Shipping the pair together is what prevents that.

⚠ **`docs/deploy.md` said `ProxyPass /logbook/api/ → 127.0.0.1:9002/api/` and that was WRONG** — the
server mounts routes at the full public path (`basePath = "/logbook/api"`), so the backend 404s
`/api/health`. Fixed in the doc on 2026-08-01. The install script's health check caught it *before*
Apache was ever reloaded, which is the argument for phasing the deploy: the backend was proven on
`127.0.0.1` while the site was still untouched.

The staging directory `/home/rami/logbook-deploy/` on the box holds the binary, the database, the
unit, the Apache snippet and two idempotent install scripts (`install-backend.sh`,
`install-apache.sh`). `install-apache.sh` backs the vhost up to `/root/`, runs `configtest` **before**
any reload, restores the backup automatically if it fails, and then curls all seven of the owner's
other sites. Baseline before the change: all seven **200**, `/logbook/` **404**.

**The PWA half is done**: manifest, icons and `public/sw.js`, which caches the shell so the app opens
at an airfield with no signal and never caches a logbook response. Offline *writes* stay out of v1.

### Open questions awaiting the owner
- Is the `kraken-predictor-python-2` container on `:8000` still wanted? Publicly exposed, up 2 years,
  and the box's largest memory consumer (~759 MB / 38%).
- Prune the stale ufw rules for `30814` and `19132` (nothing listens on either)?
- **Rotate the `rami` sudo password** once deployment is done — it was pasted into a chat session on
  2026-08-01 and must be treated as compromised. Tracked in `docs/security.md`.
- Task 8 (backfill the 30 inferred landing splits from the page images) is still open and is the
  only thing standing between the app and a fully verified night-landing figure.

### Traps already paid for — do not rediscover these
- **It is 1296 flights, not 1298.** 1298 is the CSV *row* count; Books 2 and 3 each open with the
  previous book's final row as a cumulative seed, and those two are skipped.
- **A zero-break cumulative reconciliation does NOT mean the data is complete.** All seven series
  reconciled perfectly while three flights were missing from 28/08/2025, because absent rows are
  invisible to a consistency check. Only the owner, or an external record with a continuous counter,
  can find an omission.
- **This book totals on BLOCK time.** `Total_Time` == `Block_Time` on 478 of Book 3's 479 rows; the
  single exception (08/09/2025) is a flagged discrepancy, not a pattern. Do not read it as one —
  that misreading produced a wrong delta once already.
- **A hand-entered flight lives in a different `seq` band (1 000 000+) and carries `source_book = 0`.**
  Both are load-bearing: the importer keys on `source_book` to know which rows it may delete, and
  the bands are disjoint because the importer renumbers 1..N on every run. See `docs/data-model.md`.
- **Sea vs land comes from the registration, not the type.** Verified row by row at all 1293 rows.
- **The books are not in date order** — 18 rows go backwards. Order on `seq`, never on `flight_date`.
- **Go's `time.Date` is silent on both DST edges, in different ways.** `internal/timeutil` handles
  it; do not "simplify" that check.
- **fpdf is non-deterministic by default** — it writes font objects in Go map order. `SetCatalogSort(true)`
  plus a fixed creation date is what makes two exports of the same logbook byte-identical.
- **fpdf's `CellFormat` ignores `\n`.** A multi-line column heading must be placed line by line, or
  it draws straight through the neighbouring columns.
- **CSS grid items default to `min-width: auto`**, which is why two date inputs overflowed a 390px
  phone until `.row > * { min-width: 0 }`.
- **Docker bypasses ufw.** A published container port is not closed by a firewall rule.
- **`/api/` on `ayoub.fi` is already taken** by a stale transit proxy, hence `/logbook/api/`.
- **Port 22 is under constant attack** (fail2ban: 50,264 bans). Never risk it.
- **Go's `flag` package stops parsing at the first non-flag argument**, so a flag written after a
  positional is silently dropped. `cmd/server` parses in a loop for this reason.

### Where the reasoning lives
Do not re-derive these — they are argued out in the Decision Log below (§5), all dated 2026-08-01:
**a hand-entered flight surviving the next import** · **the write path refusing where the importer
surfaces** · **the EASA layout read off the page** · **the frontend and what it refuses to hide** ·
**a second account does not isolate test data** · the stale test cache · table-driven default deny ·
Argon2id at 19 MiB · the decoy hash · sessions as rows · the rate limiter's stalest-key eviction ·
minutes on the wire · the smoke test that caught the `-db` bug · stack choice · cumulatives computed
not stored · the time model · the EASA PDF covering all three books · the landings day/night gap ·
the server security findings · Go's `time.Date` DST behaviour · the two-verifications design ·
sea/land from the registration · the derived aircraft list · the three open source-data problems.

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
| 2 | Scaffold backend + frontend, test harness | **done** 2026-08-01 — backend `make check`, frontend `npm run check` (tsc + vitest) |
| 3 | Schema + importer for 1293 flights (verified) | **done** 2026-08-01 |
| 4 | API + authentication | **done** 2026-08-01 — `internal/stats`, `internal/auth`, `internal/ratelimit`, `store/auth.go`, `cmd/server`. Every `docs/security.md` control implemented with the test that fails if it is removed. Verified live against the real 1293 flights. |
| 5 | Four frontend pages (mobile-first) | **done** 2026-08-01 — plus the auth UI. Six pages: Flights, Statistics, New flight, Export, Review, Devices, behind a login gate. React + TS + Vite, `app/frontend/`. 43 frontend tests green. Verified in a real browser against the live API, including logging a flight end to end. |
| 5b | `POST /flights` — the write path | **done** 2026-08-01 — `internal/entry` (validation, pure, 100%), `store.AddFlight`, the hand-entered `seq` band, the duplicate guard, and the import scoping that stops a re-import deleting app-entered flights. |
| 6 | Three PDF exports (EASA clone + table + stats) | **done** 2026-08-01 — `internal/pdfmodel` (the cells and totals, pure, 100%) + `internal/pdfbook` (rendering, `go-pdf/fpdf`). Live against the real logbook: **87 EASA pages**, totals block reconciling, Finnish place names intact. |
| 7 | PWA + deploy to `ayoub.fi/logbook` | **PWA done** 2026-08-01 — manifest, icons, and a hand-written service worker that caches the shell and **never** an API response, proven in a browser with the HTTP cache disabled. **Deploy HALF DONE** 2026-08-01 — backend, binary, systemd unit and frontend assets are on the box and the service is running and healthy on `127.0.0.1:9002`; **Apache is not switched on, the database there is stale (1293, not 1296), and no user account exists**, so the site is still 404. See "Where the deploy actually stands". |
| 8 | Backfill landings day/night for the **30** night rows | not started — all 30 are flagged `landings_unverified` in the DB, listed by `logbookctl import`, and surfaced by the API as `landings_unverified` in the stats summary. `claude-docs/drift.md` has the analysis. The p.62 split recomputes to **68 night / 3326 day** (the sum 3394 is unchanged); six of those 68 are estimates from three multi-landing partial-night rows, range 65–72. |
| 9 | Rule on the open source-data problems | **mostly closed** 2026-08-01 — two of the three ruled and fixed. One item left (`logbook_2_final.csv` lines 89–90) and it needs the physical page; it moves no total. See the ⏸ block at the top. |

---

## 5. Decision Log

### 2026-08-01 — The form asked for a format the phone's keyboard cannot type

Reported from the field, and the sharpest lesson in the project so far: **a flight could not be
entered on the phone at all.** The clock fields asked for `09:15Z` with `inputMode="numeric"`, and an
iOS number pad has no colon key. The required format was untypeable on the only device this app
exists for. 43 frontend tests, a browser run at 390px and a live end-to-end flight entry all passed
over it, because every one of them typed into the field programmatically or on a desktop keyboard.
**Testing a form without testing its keyboard tests half the form.** The duration fields — PIC, dual,
night, instrument, instructor — had the identical defect and had not been noticed yet.

Clock fields become native `<input type="time">`. The interesting part is what that costs: a native
picker yields `HH:MM` and cannot carry the `Z`, and the `Z` is load-bearing under rule §0.4 — it is
the whole distinction between UTC and Helsinki local, and dropping it would make every hand-entered
time silently ambiguous. So **the zone becomes a control** — a UTC / Helsinki-local toggle over the
whole Times card, defaulting to UTC — instead of punctuation the pilot has to remember. That is
better than the old field even ignoring the keyboard: the zone is now always visible rather than
implied by a character at the end of a string. The wire format is unchanged and the server's single
conversion authority is untouched; the form composes `HH:MM` or `HH:MMZ` at submit.

Durations stay free text, because a duration is a judgement about the flight rather than a reading
off a clock, but move to `inputMode="text"` so the keyboard can produce a colon.

**Total time is now derived and read-only**, at the owner's instruction — and it is still *sent*, so
the server continues to require the total to be stated rather than inventing it (the Task 5b entry
below argues why that server rule stands). The form is simply what states it now. **The cost, stated
plainly: a flight whose total differs from its block clock can no longer be typed into the app.**
That is 1 row in 479 (`08/09/2025`, itself a flagged discrepancy), and the importer can still record
it — so the capability is narrowed, not lost.

Takeoff and landing are new, **optional**, and folded behind a disclosure because most rows in the
paper books have none; air time derives from them the same way. Almost nothing was needed underneath:
the schema, `csvbook.Flight` and `store.AddFlight` already carried `takeoff_utc`/`landing_utc` — only
`entry.Draft` did not accept them.

`entry.validateAirborneTimes` mirrors the block pair deliberately: optional **as a pair**, refusing
half a pair while naming the missing half, converting through `timeutil.BlockPair` so the midnight
roll and the DST refusal behave identically, and **refusing an airborne time longer than the block
time** — an aeroplane cannot be airborne longer than it is off blocks, and storing that would create
a flight whose own parts contradict each other.

One implementation note worth keeping: the derived figure is a read-only `<input>`, not an
`<output>`. `<output>` carries an implicit ARIA role of `status`, which collided with the page's
saved-flight live region and made the "flight logged" assertion ambiguous. The test caught it.

### 2026-08-01 — Three flights were missing, and the frozen totals were unfrozen once to add them

Found by the owner mid-deployment, which is why the deploy is staged in phases: the backend was
running but Apache had not yet been switched on, so the stale database never became reachable.

`logbook_3.csv` had **no 28/08/2025 rows at all** — line 411 is 27/08/2025, line 412 is 08/09/2025.
Three OH-ESR flights had never been written down, in the CSV or on paper, and one of them is a
**SEP/IR revalidation check flight**. Full reconstruction, sourcing and deltas in
`claude-docs/drift.md`; the three things worth arguing here are these.

**A zero-break reconciliation proved nothing about completeness.** All seven `Cumulative_*` series
reconciled row by row over 1293 rows with zero breaks — *while three flights were missing* — because
a consistency check compares the rows that exist to a column those same rows produced. An absent row
is absent from both sides. This is the structural blind spot of every check this project has, and no
amount of internal verification closes it; it took the owner and an external record with a continuous
airframe counter (2663:11 → 2663:51 → 2664:39 → 2665:31, each step exact) to find and bound it.

**The freeze governs corrections, not omissions.** The owner froze the end-of-book-3 cumulatives so
that nobody would keep re-litigating figures that now match the paper. Applying that to *missing
flights* would have inverted its purpose — it would have made the record permanently wrong in order
to keep it stable, and suppressed a licence-relevant currency item. So the freeze was lifted, by the
owner, explicitly, for these three rows only, and resumes at the new figures. Recorded because the
distinction is the whole reason the rule survives contact with new data.

**Late entries, not chronological insertion.** They append at the end of Book 3 rather than slotting
into August 2025, on paper and in the CSV alike. Inserting them in date order would mean re-inking
carried-forward totals on ~5 already-written pages; a dated late entry changes nothing already
written and reads as what it is. The CSV follows the paper because the paper is authoritative, and
the schema already orders on `seq` rather than `flight_date` — the books hold 21 out-of-date-order
rows now, up from 18.

The guard tests did their job loudly: **six assertions across four packages** went red on the CSV
change (`csvbook`, `stats`, `store`, `cmd/logbookctl`), including the EASA pagination geometry. Each
constant was moved only after the importer's independent recomputation confirmed the delta — never
the other way round.

### 2026-08-01 — Task 5b: a flight typed into the app must survive the next import

The importer replaces the flights table on every run, and the migration effort re-imports every
time a page is appended to `logbook_3.csv`. So the first design question for `POST /flights` was not
validation — it was **how a hand-entered row avoids being deleted by the next transcription batch**.
Left unsolved, the app would have silently destroyed the owner's own entries within a week, which is
precisely the loss rule §0.2 forbids.

The answer is two disjoint populations in one table, keyed on `source_book`: paper rows carry 1–3,
app rows carry **0**. Three things follow, and each is a test rather than a convention.
The import's `DELETE` is scoped to `source_book <> 0`. The import's **checksums** are scoped the
same way — they answer "is the database what the CSVs say", and counting a flight that is in no CSV
would make the import fail verification on its own correct work, where the only way to pass would be
to delete the pilot's flight. And `seq`, which the importer reassigns 1..N on every run, is
allocated to app rows from a separate band at **1 000 000**; Book 3 is still being transcribed, so
any hand-entered `seq` inside 1..N is a collision waiting for the migration to catch up to it. The
high band also sorts app-entered flights after every page of the paper books, which is where a
flight flown today belongs.

One repair was found by running it rather than by reasoning: replacing the `aircraft` table nulls
`aircraft_id` on the hand-entered rows through `ON DELETE SET NULL`, so a flight typed in the app
lost its aircraft link the first time a page was transcribed and never got it back. The importer now
re-links by registration. The live re-import reported `1293 linked, 1 unlinked` — that one row is
what exposed it.

### 2026-08-01 — The write path refuses where the importer surfaces

`internal/entry` validates a hand-entered flight and its posture is the deliberate **opposite** of
`internal/csvbook`'s. The importer surfaces a problem and imports the row anyway, because the paper
is authoritative and nobody can be asked about a 14-year-old page. Nothing on the write path is
authoritative yet and the pilot is standing at the form — so a draft that does not make sense is
refused, with the field named, rather than stored with a flag on it. Surfacing a discrepancy and
creating one are different acts.

The sharpest case is an ambiguous local time. On an imported row a DST gap or fold is stored with
`time_origin = unknown` and surfaces for review; here it is **refused with a message asking for a
Zulu time**, because manufacturing an unaudited instant when the true one is one question away would
be inventing a fact about a legal record.

Three other decisions worth keeping. **`total_time` is required, not derived** from the off/on-block
clock: it is the figure the whole logbook adds up and a licence application is written from, so the
form prefills it from the clock as a one-tap suggestion but the server never invents it. **Every
problem is reported at once**, each naming its field, because a twenty-field form that reveals one
mistake per submission is a form that gets abandoned — and an abandoned form means the flight is not
logged at all. And a resubmission of the same `(date, aircraft, off-block)` is a **409**: that is the
double-tapped submit button on a phone, and two identical rows inflate a licence total.

### 2026-08-01 — Task 6: the EASA layout was read off the page, not off the standard

The layout came from photographing `logbook-3/IMG_6025.JPEG` and reading it, which corrected two
assumptions this file previously recorded. The paper is a **two-page spread** of 15 rows — GENERAL
plus TOTAL/NIGHT/SE-VFR/SE-IFR on the left, the remaining function columns plus LANDINGS and REMARKS
on the right — rendered here onto one A4 landscape sheet so that one PDF page is one logbook page.

Two mappings are judgement calls, both made conservatively:

- **SINGLE ENGINE IFR is always blank.** The CSVs carry no flight-rules column, and instrument time
  is not a substitute for one — `OH-COF` and `OH-CTH` are C152s with instrument time logged under
  the hood. Deriving an IFR figure from it would be manufacturing data. This is also exactly what
  the owner's own pages do: all single-engine time goes in SE-VFR.
- **The per-row TOTAL column carries the flight's own time**, not a running total. The owner writes
  a running total there by hand (1027:29 → 1028:14 → …, which is how the page was decoded), but an
  authority reads that column as per-flight, and the TOTAL THIS PAGE / PREVIOUS / TOTAL block below
  is where a cumulative belongs. Both conventions are satisfied on the page.

The package is split in two so the coverage rule can mean something: `internal/pdfmodel` computes
every cell and every total and is **pure and at 100%**, because rule §0.6 names "PDF totals" as
calculation core; `internal/pdfbook` draws, and lives at the 80% bar with fpdf's error paths.

Rendering is **deterministic** — same flights, same bytes — which took a fixed creation date *and*
`SetCatalogSort(true)`, because fpdf writes its font objects by ranging over a Go map and two
renders otherwise differ in object order. Without that, a diff between two exports cannot
distinguish "the record changed" from "it was regenerated".

`go-pdf/fpdf` is the fourth direct dependency and is justified in `docs/security.md`: it is pure Go
with no dependencies of its own, versus a 300 MB headless browser executing a rendering engine on a
2 GB shared box.

### 2026-08-01 — Task 5: the frontend, and what it refuses to hide

Six pages behind a login gate: Flights, Statistics, New flight, Export, Review, Devices. React +
TypeScript + Vite, built to static files; Node never runs on the server.

The session is an HttpOnly cookie the page cannot read, so **"am I signed in?" is answered by asking
the server** (`GET /me`) rather than by any local flag. That costs one request at startup and it is
the honest answer — a cached "signed in" boolean would survive a revoked session and show an empty
logbook with no explanation. Any 401 from anywhere drops the whole app back to the login page.

Three places where the UI is deliberately not smoother than the truth:

- **A failed read is never an empty list.** "You have flown nothing" because the phone lost signal is
  the silent corruption rule §0.2 forbids, so a network failure renders as an error.
- **The 30 inferred landing splits are marked** — an asterisk on the row, and a paragraph on the
  statistics page naming the count. A page that printed the night-landing figure plainly would be
  claiming a verification nobody has done.
- **The login page stays uninformative.** The server answers a wrong username and a wrong password
  identically and in the same time; a UI that said "no such user" would undo that control, so there
  is a test asserting the message does not.

Routing is ~40 lines rather than a routing library (rule §0.3): six pages, no nested routes, no
route parameters. Real `<a href>` links, so middle-click and long-press work.

### 2026-08-01 — The service worker caches the shell and never the logbook

The PWA half of Task 7. The app is used at an airfield on a phone with poor signal, so the shell
should open without waiting for the network — but a service worker is a cache that **ignores
`Cache-Control: no-store` unless it is written not to**, and every `/logbook/api/` response is
personal data the server explicitly marks no-store.

So the policy is one function, checked first, before any other rule can catch an API URL: anything
under `/logbook/api/`, and anything that is not a GET, is passed through untouched and unstored —
the worker does not call `respondWith` at all, so the browser does exactly what it would have done
without it. Navigations are network-first with the cached shell as a fallback; the content-hashed
build assets are cache-first, which is safe because the filename changes when the bytes do.

Caching a logbook response would have left the owner's flights readable on the device after the
session was revoked — a control the server states in a header, undone silently on the client.

Written by hand rather than with a build plugin: the whole policy is forty lines, it is the kind of
policy that has to be read to be trusted, and a PWA plugin is a supply-chain decision (rule §0.3).

Tested against **the shipped `public/sw.js`** rather than a copy of its logic — the test evaluates
the real file in a fake worker global and pulls `policy` back off it. Deleting the API guard turns
three of those tests red, which is how it was confirmed rather than assumed. Then verified in a
browser: after signing in and browsing, the cache holds the shell and nothing else; with the HTTP
cache disabled and the network off, the shell still opens and the flights are **not** readable.

One deploy consequence, recorded in `docs/deploy.md`: `sw.js` must be served `Cache-Control:
no-cache`, or a stale worker outlives a deploy and keeps serving the previous bundle.

### 2026-08-01 — A second account does NOT isolate test data; a second file does

Raised by the owner while Task 5 was being verified: keep `rami` for the real logbook and use a
`ramitest` account for experimenting with flight entry.

**A second account would not have isolated anything.** This app is single-tenant by design (§2, "not
in v1: multi-user sharing"): `flights` has no owner column and `AddFlight` does not record who wrote
a row, so a flight entered by `ramitest` lands in the same logbook as `rami`'s and is
indistinguishable from it. Authentication here is a gate on the front door, not a partition.

The isolation boundary is **the database file**. `make scratch` therefore rebuilds a throwaway
database from the CSVs at `/tmp/logbook-scratch.db` and prints the two commands to put an account on
it and run the API against it; `make scratch-clean` removes it. The scratch file is rebuilt from the
CSVs in one command, so it can be deleted at any moment without losing anything.

Recorded because the instinct — "make another user" — is the natural one and is wrong here, and
because the right answer stops being obvious the moment this app grows a second real user.


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
