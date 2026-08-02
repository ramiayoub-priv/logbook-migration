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
Finnish pilot), live at `https://ayoub.fi/logbook`. It holds 1296 flights transcribed from three
paper logbooks, computes every total on demand, and exports an EASA-format PDF for the authority.
Read `CLAUDE.md` §0 first — those rules are non-negotiable and were written for this work.

**⛔ FIRST, THE RULING THAT CHANGES WHAT YOU MAY DO (2026-08-02).** The transcription effort in
`claude-docs/` is **CLOSED**. The three CSVs are **read-only inputs**: do not edit a cell, append a
row, re-transcribe a page or close a known discrepancy — see `CLAUDE.md` §0.8. Every paper page that
exists is already transcribed; there is no backlog. **New flights are entered in the app**, and
`app/` is the only active effort. If the old migration docs read like a task list, they are stale
and the rule wins.

**Where to start.** This file, then `docs/data-model.md` / `docs/security.md` / `docs/deploy.md` as
the work requires.

**THE QUEUE, in the order the owner wants it** — all three came out of the first real day of use on
2026-08-02, when the owner flew a there-and-back and logged both flights on the phone. Every design
decision is already ruled and written up in the decision-log entry **"The first real day of use"**;
read that before touching any of them, then build.

| | | |
|---|---|---|
| **Task 11** | Saving must be unmissable | The confirmation is off-screen on a phone. Success takes over the screen. |
| **Task 12** | Takeoff / landing / air time in the table, fields un-collapsed | The aircraft's own logbook is filled from airborne times. **Needs the staged binary deployed.** |
| **Task 13** | Aircraft time page: block vs air, by aircraft and range | He pays by the hour, some owners by block time and some by air time. Honesty about coverage is the load-bearing part. |

⚠ **DEPLOY FIRST — IT IS ALREADY BUILT AND STAGED.** The box is running the morning's binary and
frontend; the repo is two features ahead (edit/delete, and `takeoff_utc`/`landing_utc` on the wire).
The binary is staged at `/home/rami/logbook-deploy/logbook-server`, md5
`64c47992cc8d949aa0e84fdf4ae2ccaf`, matching a build of this repo's HEAD. Shipping is the owner's
`sudo /home/rami/logbook-deploy/update.sh` followed by the frontend rsync — **in that order**,
because the edit page calls `PUT`/`DELETE /flights/{seq}`, which the deployed binary does not have.
**Task 12 cannot even be seen to work until this lands.** See the runbook below.

⚠ **THERE ARE NOW REAL FLIGHTS IN PRODUCTION THAT EXIST NOWHERE ELSE.** The owner logged two on
2026-08-02. They carry `source_book = 0`, so the re-import inside `update.sh` leaves them alone (its
`DELETE` is scoped, and there is a test) — but they are no longer reconstructible from the CSVs, and
"rebuild it from the CSVs in one command" no longer means "lose nothing". **The backups under
`/var/lib/logbook/backups/` are the only copy of those rows.**

**Status: LIVE at `https://ayoub.fi/logbook` and in real use — the owner logged two flights from the
field on 2026-08-02. The morning's deploy (binary, frontend, the three 28/08/2025 flights, the
no-cache headers) is on the box and verified. Since then the repo has moved TWO features ahead of it
and the queue above is what the first day of use asked for.**

### Done (2026-08-01)
- **Task 2** — `app/backend/` Go module, `internal/hhmm` and `internal/timeutil`. Both 100%.
- **Task 3** — the schema and importer. All **1296** flights import and verify.
- **Task 4** — the API and authentication. Every control in `docs/security.md` has the test that
  fails if it is removed.
- **Task 5b** — **`POST /flights`**, the only write path into the legal record. `internal/entry`
  validates (pure, 100%); `store.AddFlight` allocates book order. See the decision log: the load-
  bearing part is that a hand-entered flight **survives the next CSV re-import**.
- **Task 6** — the **three PDFs**. `internal/pdfmodel` (cells and totals, pure, 100%) +
  `internal/pdfbook` (rendering). Verified against the real logbook: 87 EASA pages, totals block
  reconciling, Finnish place names intact.
- **Task 5** — the **frontend**, `app/frontend/`. Six pages behind a login gate. **60 tests green**,
  and driven in a real browser against the live API — including logging a flight end to end,
  watching the duplicate be refused, and confirming zero horizontal overflow on a 390px phone.
  Reworked on the evening of 2026-08-01 after the owner found the new-flight form unusable on an
  actual phone; see the decision log.

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
                       handentry.go (AddFlight + the seq band + the reimport relink) and
                       handedit.go (UpdateFlight/DeleteFlight + the append-only audit).   [83%]
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
  src/format.ts      H:MM, UTC dates, and the HHMM four-digit entry helpers. The ONLY place
                     minutes become H:MM, and the only place four digits become a time.
  src/swupdate.ts    reloads the page once when a new service worker claims it -- a home-screen
                     install has no reload button.
  src/pages/         Login, Table, Statistics, Export, Review, Sessions, RangePicker, and
                     FlightForm -- the ONE form, wrapped by NewFlight and EditFlight. Two copies
                     would drift at the first fix applied to only one of them.

app/deploy/          the box's staging scripts, IN THE REPO as of 2026-08-01 (rule 0.1 -- they
                     lived only in /home/rami/logbook-deploy/ until then, which meant a fresh
                     clone could not reconstruct the deploy). update.sh, install-backend.sh,
                     install-apache.sh, apache-logbook.conf, logbook.service. Edit them HERE and
                     rsync them to the box; never edit them on the box.
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

⚠ **And fix the whole class, not the reported instance.** That same evening the keyboard fix was
applied to the clock fields and *not* to the duration fields one card below, which had the identical
defect and had even been noted as having it. The owner had to report it a second time. **If a defect
appears twice on one page, it is a rule — sweep the page.**

**There is no committed database** — it is generated, and `app/.gitignore` keeps `*.db` and `*.bak`
out of the repo. A *development* database is rebuilt from the CSVs in one command (`make scratch`).

⚠ **That stopped being true of PRODUCTION on 2026-08-02.** Flights entered in the app exist in no
CSV, and there are real ones now. The production database is no longer derivable from the repo, so
`/var/lib/logbook/backups/` — written by `update.sh` before every import — is the only copy of those
rows. Treat it accordingly: this is the first thing in the project whose loss cannot be undone by
re-running something.

### The numbers the import produces — memorise these
```
flights 1296 | total 1222:10 | pic 1054:45 | dual 167:25 | instrument 107:58
night 22:45  | instructor 189:41 | seaplane 407:39 | landings 3444 | aircraft 38
discrepancies 61 | EASA export 87 pages
```
All seven `Cumulative_*` series reconcile with **zero breaks**.

⚠ **These moved once, on 2026-08-01**, and the previous values are still all over the git history:
they were `1293 / 1219:35 / 1053:03 / 166:32 / 107:05 / 3439`. Three flights of **28/08/2025** were
missing from `logbook_3.csv` entirely — see the decision log and `claude-docs/drift.md`. **They will
not move again**: the dataset was closed on 2026-08-02 (`CLAUDE.md` §0.8).

Asserted in `internal/csvbook/realdata_test.go` and again, by a different code path, in
`internal/stats/realdata_test.go`. **If one of them changes unexpectedly, the import is wrong until
proven otherwise — do not adjust the expectation to make the test pass.**

⚠ **These tests no longer have ANY legitimate reason to fail.** Until 2026-08-02 there was one —
`logbook_3.csv` growing as pages were transcribed — and this file carried a procedure for updating
the constants when it did. **That procedure is void.** The CSVs are closed, so a red
`realdata_test.go` now means the importer, the store or the stats code has broken, and the fix is
never to touch the expectation. If you find yourself editing a number in `realdata_test.go`, stop
and re-read `CLAUDE.md` §0.8.

### ⛔ THE DATA IS CLOSED — these items stay open forever, and that is deliberate
Do **not** re-validate the books on spec, and do **not** offer to finish these. Closing either one
would mean touching the historical data (`CLAUDE.md` §0.8). They are surfaced in the UI because a
record that hides what it has not verified is worse than one that says so.

- **The 30 `landings_unverified` rows.** Flagged in the DB, counted by the API, asterisked in the
  table and named in a paragraph on the statistics page. **Keep every one of those signals.**
- **`logbook_2_final.csv` lines 89–90** (`04.05.2018` ×2), dated `DD.MM.YYYY`. Affects **row order
  only**, moves no total, and no electronic source can settle it.
- Paper-side only: the **p.62 inked landing split** `59 night / 3335 day` recomputes to **`68 /
  3326`** (the sum 3394 never moved). Nothing to do — the CSV was always right.

### The API surface, as built
All under **`/logbook/api/`** — not `/api/`, which on `ayoub.fi` is taken by a stale transit proxy.
Durations are **integer minutes**; the frontend formats H:MM.

```
POST   /login              public   {username,password} -> 200 + Set-Cookie; 401 uniform; 429 throttled
GET    /health             public   exactly {"status":"ok"} and nothing else
POST   /logout             private  revokes this session, clears the cookie
GET    /me                 private  {user_id, username}
GET    /flights   ?from&to private  {flights:[...], count} in seq order (the table reverses for display)
POST   /flights            private  a hand-entered flight -> 201; 400 with per-field errors; 409 duplicate
                                    times are "HH:MM" (Helsinki local) or "HH:MMZ" (UTC)
                                    takeoff/landing are OPTIONAL, but all-or-nothing as a pair
GET    /flights/{seq}      private  one flight, imported or hand-entered. 404 if there is none.
PUT    /flights/{seq}      private  corrects a HAND-ENTERED flight. Full replacement, same
                                    validation as POST. 403 on an imported row, 404 missing,
                                    409 if it would duplicate another flight's key.
DELETE /flights/{seq}      private  deletes a HAND-ENTERED flight, returns what was removed.
                                    403 on an imported row, 404 missing (so a double tap is safe).
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

✅ **The deployment caught up on 2026-08-02** — new binary, re-imported database, four-digit form and
cache headers. What was verified is listed under the runbook below.

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

✅ **RAN 2026-08-02, and the box is now level with the repo.** What was verified from here afterwards,
each by an independent check rather than by trusting the script's own output:

- `/opt/logbook/logbook-server` md5 **`d22d8a39b456560e0f76ba1f28fbb821`** — byte-identical to the
  binary built from this repo's HEAD. `logbook-server.prev` is in place for rollback.
- `logbook.service` active since 21:05:37 UTC, **28.0 MB** against the 192 MB `MemoryMax`.
- `/logbook/api/health` **200**, `/logbook/api/flights` without a session **401** — default deny
  survived the deploy.
- `index.html` serves `Cache-Control: no-cache, no-store, must-revalidate` + `Pragma` + `Expires: 0`;
  `sw.js` serves `no-cache`; the live `sw.js` is `logbook-shell-v2` with the `no-store` shell fetch.
- The live bundle is the four-digit form (`index-C1WjdtsT.js`), and `index.html` points at it.
- **All seven of the owner's other sites still answer 200.**

⏳ **The one item a session cannot confirm: the `flights=1296` startup line.** The database and the
unit's journal are readable only by root and the service user, so this must be read off the owner's
own `update.sh` output, or from the app once signed in. Everything else above was checked from here.

**The runbook, for next time** — four steps, in this order. The two `sudo` ones are the owner's;
there is no passwordless sudo, and a session cannot run them. Before running it, note that the staged
CSVs on the box were confirmed **byte-identical (md5) to the repo's**, which is what put the three
28/08/2025 flights in front of the importer.

```bash
# 1. Stage the current build. rami owns /home/rami/logbook-deploy, so no sudo.
#    Binaries are cross-compiled CGO_ENABLED=0 from this repo's HEAD.
cd app/backend && export PATH=$HOME/.local/go/bin:$PATH
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /tmp/logbook-server ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /tmp/logbookctl   ./cmd/logbookctl
cd ../.. && rsync -a /tmp/logbook-server /tmp/logbookctl \
    logbook_1_final.csv logbook_2_final.csv logbook_3.csv \
    rami@ayoub.fi:/home/rami/logbook-deploy/
rsync -a logbook_1_final.csv logbook_2_final.csv logbook_3.csv rami@ayoub.fi:/home/rami/logbook-deploy/csv/
rsync -a app/deploy/ rami@ayoub.fi:/home/rami/logbook-deploy/   # scripts live in the repo now

# 2. OWNER, sudo: new binary + backup + re-import + verify + restart.
#    The startup log line must read flights=1296.
ssh -t rami@ayoub.fi 'sudo /home/rami/logbook-deploy/update.sh'

# 3. The frontend, AFTER step 2 (see the pairing warning below). No sudo.
cd app/frontend && npm run build && cd ../..
rsync -a --delete app/frontend/dist/ rami@ayoub.fi:/var/www/logbook/

# 4. OWNER, sudo: the Cache-Control headers, so the phone stops serving the old build.
ssh -t rami@ayoub.fi 'sudo /home/rami/logbook-deploy/install-apache.sh'
```

⚠ **`install-apache.sh` now REPLACES its block rather than skipping it.** The first version refused
to touch a vhost that already had a `BEGIN logbook` block, which meant a changed snippet could never
reach the server. It now strips its own block, re-inserts the current `apache-logbook.conf`, and
**refuses to write if stripping the block from the before and after files does not produce identical
text** — the proof that nothing outside our block moved. Backup, `configtest` and auto-restore are
unchanged. Rehearsed against a copy of the vhost before it was ever run as root, which caught two
bugs: an inserted blank line made every run differ from the last, breaking both that safety check
and idempotence.

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

**The one that blocks work — Task 10, editing a flight.** The app is now the only way the record
grows, and a flight typed on a phone cannot currently be corrected without opening SQLite. It needs
a ruling before code, because on a legal record the obvious implementation is the wrong one:

- **Correct in place, or append a correction?** An in-place `UPDATE` makes a licence total change
  with no trace of what it was — the drift this project spent 106 KB of `drift.md` on. The
  alternative is a superseded row, kept and excluded from totals, which is auditable but doubles the
  work of every read path.
- **What may be edited?** A typo in a remark is not a typo in `total_time`. A field-by-field answer
  is possible; a blanket one is what gets regretted.
- **Delete, or void?** Real deletion of a flight that was never flown is legitimate; so is the fear
  of a fat-fingered swipe on a phone. The `source_book = 0` band means app-entered rows are
  distinguishable from paper ones — a delete restricted to them is a much smaller decision.
- Whatever is chosen, **the imported 1296 stay untouchable** (`CLAUDE.md` §0.8).

**Security and the box** (unchanged, all pre-dating today):
- **Rotate the `rami` sudo password.** It was pasted into a chat session on 2026-08-01 and must be
  treated as compromised. Tracked in `docs/security.md`. **This is the oldest open item.**
- Is the `kraken-predictor-python-2` container on `:8000` still wanted? Publicly exposed, up 2 years,
  and the box's largest memory consumer (~759 MB / 38%).
- Prune the stale ufw rules for `30814` and `19132` (nothing listens on either)?

**Smaller app items, none blocking:** offline *writes* are still out of scope (§2); the new-flight
form's four-digit layout has been proven in tests and on the owner's phone but never inspected in a
desktop browser at 390px; and `MemoryMax` is still set from an estimate — the real peak during a
full-logbook PDF export has not been measured (`docs/deploy.md` asks for it).

### Traps already paid for — do not rediscover these
- **A cache fix cannot reach a device that already has the old service worker.** Right after the
  2026-08-02 deploy the owner's phone still showed the *pre-rework* form, and the live bundle
  provably did not contain it (the old page's own text greps to 0 in `index-C1WjdtsT.js`). The three
  no-cache layers stop this recurring; they do nothing for a worker installed before they existed.
  The one-time escape is a query string (`/logbook/?v=2`) plus fully closing and reopening the
  home-screen app. **When the owner reports a stale page, check the live bundle's contents before
  suspecting the deploy** — server-side and device-side staleness look identical from the phone.
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
| DB | SQLite (`modernc.org/sqlite`, pure Go) | No CGO ⇒ trivial cross-compile. One file; backup = `VACUUM INTO`. 1296 rows is nothing. |
| Time | embedded `tzdata` | Behaviour must not depend on the server's zoneinfo. |
| PDF | `go-pdf/fpdf` | Absolute positioning, which a fixed 15-row EASA grid needs. Headless Chrome would cost 300 MB+. |
| Frontend | React + TS + Vite | Builds to static files. Node is build-time only, never on the server. |

## 4. Task Board

| # | Task | Status |
|---|---|---|
| 1 | Project rules + app docs | **done** 2026-08-01 |
| 2 | Scaffold backend + frontend, test harness | **done** 2026-08-01 — backend `make check`, frontend `npm run check` (tsc + vitest) |
| 3 | Schema + importer for 1296 flights (verified) | **done** 2026-08-01 |
| 4 | API + authentication | **done** 2026-08-01 — `internal/stats`, `internal/auth`, `internal/ratelimit`, `store/auth.go`, `cmd/server`. Every `docs/security.md` control implemented with the test that fails if it is removed. Verified live against the real flights. |
| 5 | Four frontend pages (mobile-first) | **done** 2026-08-01 — plus the auth UI. Six pages: Flights, Statistics, New flight, Export, Review, Devices, behind a login gate. React + TS + Vite, `app/frontend/`. **75 frontend tests green.** Reworked twice the same evening after the owner found the new-flight form unusable on a real phone: the table now lists newest first, and **every time on the form is four digits on a number pad** (`0915`, `0115`) with the totals derived live. Verified in a real browser against the live API, including logging a flight end to end, and the four-digit form's exact payload accepted `201` by a scratch server. |
| 5b | `POST /flights` — the write path | **done** 2026-08-01 — `internal/entry` (validation, pure, 100%), `store.AddFlight`, the hand-entered `seq` band, the duplicate guard, and the import scoping that stops a re-import deleting app-entered flights. |
| 6 | Three PDF exports (EASA clone + table + stats) | **done** 2026-08-01 — `internal/pdfmodel` (the cells and totals, pure, 100%) + `internal/pdfbook` (rendering, `go-pdf/fpdf`). Live against the real logbook: **87 EASA pages**, totals block reconciling, Finnish place names intact. |
| 7 | PWA + deploy to `ayoub.fi/logbook` | **PWA done** 2026-08-01 — manifest, icons, and a hand-written service worker that caches the shell and **never** an API response, proven in a browser with the HTTP cache disabled; the shell is now fetched `no-store` and a new worker reloads the page (`src/swupdate.ts`). **Deployed and LIVE** 2026-08-01 at `https://ayoub.fi/logbook` — service user, binary, systemd unit, frontend, the additive Apache block and the account are all in place, and the owner's other seven sites still answer 200. The deploy scripts now live in **`app/deploy/`** instead of only on the box. **NOT finished**: the box runs an older binary and a **1293**-flight database while the repo is at **1296** (the three 28/08/2025 flights). Current binaries and scripts are staged on the box and md5-matched; two owner `sudo` commands remain. See the runbook in "Where the deploy actually stands". |
| 8 | Backfill landings day/night for the **30** night rows | **WILL NOT DO** — closed 2026-08-02 by the owner's ruling that historical data is not to be touched (`CLAUDE.md` §0.8). The 30 rows keep their `landings_unverified` flag **permanently**: the API reports the count, the table asterisks the row and the statistics page names it in a paragraph. That is the honest state and it must not be quietly dropped to make a page look tidier. |
| 9 | Rule on the open source-data problems | **closed** 2026-08-02 — two of three were ruled and fixed on 2026-08-01. The third (`logbook_2_final.csv` lines 89–90, `04.05.2018` ×2) stands **unresolved forever**: it affects row order only, moves no total, and settling it needs a physical page that will not be re-read. Recorded, not fixed. |
| 11 | **Saving a flight must be unmissable** | **planned, not started** — from the first real day of use: the owner logged two flights on the phone and could not tell whether either had saved. Ruled: **the success takes over the screen**. See the 2026-08-02 "first day of use" decision-log entry for the full design. |
| 12 | **Takeoff / landing / air time in the table, and out of the disclosure** | **planned, not started** — the aircraft's own logbook is filled from airborne times, not block times, so they have to be readable straight off the flights table. Includes promoting the two fields out of the collapsed section, ruled by the owner. Needs the **staged binary deployed**: `takeoff_utc`/`landing_utc` only reach the client from the new build. |
| 13 | **Aircraft time page (block vs air, by aircraft and date range)** | **planned, not started** — the owner pays for aeroplanes by the hour, some owners charging block time and some air time. Pick an aircraft and a range, get both totals in H:MM **and** in minutes, plus the flights behind the figure. The load-bearing part is honesty about coverage — see the decision-log entry. |
| 10 | **Edit / delete a flight** | **done** 2026-08-02 — owner ruled: app-entered flights only, real delete with an audit copy, double confirmation. `PUT`/`DELETE`/`GET /flights/{seq}`, `store.UpdateFlight`/`DeleteFlight`, the append-only `flight_audit` table, and the shared `FlightForm` behind both the new and edit pages. **83 frontend tests, backend 88.3%**, and driven live against a scratch server: edit, refusal on a paper row, delete, totals following, audit rows written. |

---

## 5. Decision Log

### 2026-08-02 — The first real day of use, and the three things it asked for

The owner flew a there-and-back and logged **both flights on the phone, in the field** — the first
time the app was used for what it is for. Everything below comes from that hour, and none of it was
visible from a test suite, a desktop browser or a scratch server. All three are **planned here and
implemented in a later session**; the rulings are the owner's and are already made.

**1 · "I get no feedback when I save, so I wasn't sure it was entered."** The confirmation exists —
`<p role="status">` at the top of the form — and on a phone it is **off-screen**: the submit button
is at the bottom of a form three cards long, and the message renders somewhere above the fold. A
screen-reader user was told; the pilot looking at the screen was not. **Ruled: the success takes over
the screen.** After a save the form is replaced by a large confirmation naming the flight — date,
registration, total — scrolled into view, offering *Log another flight* and *See it in the table*. A
failure gets the same prominence in red, scrolled to the field that caused it.

Two details that must not be lost in the build. The **draft has to survive a failed save** — a phone
that empties a twenty-field form because the server said 400 is a phone that does not get the flight
logged at all. And there must remain **exactly one live region**: the page already learned this the
hard way when an `<output>` element's implicit `role="status"` collided with the saved-flight
announcement.

This is the third time a defect has been invisible to every test and obvious in thirty seconds of
real use, after the untypeable colon and the stale service worker. **The pattern is not "test more",
it is "the phone, in the field, is a different machine".**

**2 · Takeoff, landing and air time belong in the table.** The aircraft's own logbook — a separate,
legally required document the owner fills after flying — records **airborne** times, not block times.
Reading them off the app instead of the paper is the whole point of having the app in the field. So
the flights table gains **Takeoff**, **Landing** and **Air**, and air time is *computed* at render
from the two instants (rolling past midnight the way the server does), never stored — rule §0.5's
reasoning applies to any derived figure, not only to cumulatives.

The consequence the owner also ruled on: **the airborne pair comes out of the collapsed "optional"
section** and sits in the Times card next to off/on block. It was folded away because most rows in the
*paper books* have none — but that is a fact about 1296 historical rows, not about the flights being
flown now. A field you have to remember to expand is a field that ends up empty, and an empty
airborne time is what makes an air-time total unusable a year later, when it is being billed from.

⚠ This one **needs the staged binary deployed**: `takeoff_utc`/`landing_utc` reach the client only
from the build made this morning, which is still sitting in `/home/rami/logbook-deploy/`.

**3 · An aircraft-time page, because the money is real.** The owner rents aeroplanes and **some
owners charge block time, some charge air time**. Pick an aircraft and a date range; get both totals
in **H:MM and in whole minutes** (an invoice is checked in one and computed in the other), and the
list of flights behind the figure so a disputed line can be traced to a flight rather than argued
against a single number.

The aggregation belongs in **`internal/stats`**, which is pure and held at 100% — this is money and
it is the same class of code as the licence totals.

**The load-bearing decision is what the page does about missing airborne times.** Air time is known
only for flights carrying both, and today that is a small minority. A page that adds up what it has
and prints "Air time: 3:20" is claiming a completeness it does not have, and the owner would bill or
be billed on it. So the figure is always shown **with its coverage** — air time known for N of M
flights in this range — and the block total, which is known for every flight, is never mixed with it.
That is rule §0.2 applied to a figure nobody has computed before: surface the gap, never paper over
it.

Naming and one open question: the tab bar already holds five entries and a sixth is tight on a 390px
phone, so where this page hangs is a layout decision to make with the page in front of us, not now.

### 2026-08-02 — Editing a flight: a plain form over an append-only trail

Asked for by the owner the same day the data closed, and for the reason the closure created: with
transcription finished, a flight typed on a phone is the record, and until now the only way to fix a
typo in one was to open SQLite on the server. Two rulings shaped it, both the owner's:
**app-entered flights only**, and **a real delete with an audit copy** rather than a soft-deleted row.

**Scope: `source_book = 0` and nothing else.** Refusing to edit an imported row is not conservatism,
it is the only durable answer — the importer replaces every row with `source_book <> 0` on each run,
so an edit to one would be discarded at the next re-import without anyone being told. It is enforced
in `internal/store`, not in a handler, so no route present or future can get round it. An imported
flight still *loads* on the edit page and is explained there; a page that 404'd a flight visible in
the table would read as a broken link rather than as "this row is closed".

**In place, over an append-only trail.** The owner asked for a standard edit and that is what the
form is. What sits underneath is `flight_audit`: the complete previous row as JSON, with a timestamp
and a user, written **in the same transaction** so a change cannot commit without its record. It
costs the read paths nothing — nothing in the app reads it — and it is what makes an in-place `UPDATE`
defensible on a record that backs licence privileges. A delete is recoverable from that copy alone,
which is what made "remove the row" the better of the two delete options: nothing lingers in the
logbook, and nothing is actually lost.

**A full replacement, not a patch.** The form holds every field, so it sends every field. Merging
"whatever happened to be sent" into a legal record is a class of bug invisible in a diff — the pilot
reads a form saying one thing while the database holds another. It also means an edit runs through
the *same* `entry.Validate` as a new flight: the rules about what may be written do not depend on
which door it came through.

**The bug this surfaced, which is the most valuable part.** The API had never returned
`takeoff_utc`/`landing_utc` — nothing needed them until something read a flight back to edit it. A
form that submits a field it cannot display **erases that field**, so the first edit of a flight with
airborne times would have silently dropped them: rule §0.2's silent corruption, introduced by a
feature meant to make corrections possible. They are on the wire now, with a test that fails if they
leave. **The general lesson: adding a read path is what audits a write path. Until something reads a
record back, "we store it" and "we can show it" are untested assumptions.**

**Deleting asks twice and the second question names the flight** — date, registration, both clock
times, the total, the landings, and how far the totals will drop. "Are you sure?" about an unnamed
row is how the wrong flight gets deleted. It is a `role="alertdialog"` region rather than
`window.confirm`, which cannot say which flight it is about, cannot be styled for a phone, and cannot
be tested.

**One form, not two.** `FlightForm` is shared by both pages. Two copies would have started identical
and drifted at the first fix applied to only one of them — which is precisely how the duration fields
kept their untypeable colon after the clock fields lost theirs, twelve hours earlier.

The router grew its first parameterised route (`/logbook/edit/1000123`) — one regex, still no routing
library (rule §0.3). A real URL rather than component state, because an edit should survive the
reload that happens every time a phone swaps the app out.

### 2026-08-02 — The historical data is closed, and the app is the only effort left

Owner ruling, verbatim: *"we will no longer touch historical data. This is the truth now. From now
on the focus is on developing the logbook app."* Written into `CLAUDE.md` as **rule §0.8** and
banner-headed on `claude-docs/resume.md`, because the old docs read as a standing instruction to
keep transcribing and a fresh session would have followed them.

**It closes cleanly, which is why now.** Every photographed spread — `IMG_6007`–`IMG_6037`, book
pages 1–62 — is transcribed, verified and reconciled; all seven `Cumulative_*` series match row by
row with zero breaks; and the last known omission (the three 28/08/2025 flights) was found, sourced
and imported to production the same night. There is no backlog being abandoned here. The paper book
has blank pages left, and the flights that would have filled them go into the app instead.

Three consequences worth stating, because each one changes how a future session should behave.

**The guard tests get stronger, not weaker.** `realdata_test.go` has had exactly one legitimate
reason to go red — `logbook_3.csv` growing — and appending was routine enough that the correct
response was written into this file as a procedure. That procedure is now void: the CSVs will not
grow again, so **any movement in 1296 / 1222:10 / 1054:45 / … is a defect**, and updating the
constant is never the fix. A test that could previously fail for a good reason can now only fail for
a bad one, which makes it a much better test.

**The two open data questions stay open, permanently, and stay visible.** The 30
`landings_unverified` rows keep their flag, the asterisk on the row and the paragraph on the
statistics page; the `logbook_2_final.csv` lines 89–90 date ambiguity stands. Closing either would
mean touching the data. This is not an oversight to tidy up later — it is the honest state of a
record whose paper source is no longer being consulted, and the UI already says so. Task 8 is closed
as **will not do**, not as done.

**The write path stops being a convenience and becomes the system of record.** `POST /flights` was
designed on the assumption that the importer owned the data and hand entry was the exception — which
is why a hand-entered row lives in its own `seq` band with `source_book = 0` and survives a
re-import. That design now carries more weight than it was built for, and the gap it leaves is
sharper: **there is no way to edit or delete a flight.** A typo in a flight logged on a phone is
currently permanent unless someone opens SQLite. That is the first thing to weigh next session — see
the brief at the top of this file.

A re-import is now only ever a scratch-database rebuild, not a production operation.

### 2026-08-01 — Four digits, and nothing to punctuate

Asked for by the owner, in the plainest possible terms: *"no need to force the user to write a colon,
no need to write Z — just hhmm, exactly four numbers always, then calculate the total times
dynamically."* This is the third pass over the same form, and it is the one that finally states a
single rule instead of a rule per field.

**The morning's fix was half a fix.** The clock fields became native `<input type="time">`, which
solved the colon by removing typing altogether — but the durations (PIC, dual, night, instrument,
instructor) were left as free text still wanting `1:15`, on the same keyboard, one card further down
the same page. The lesson from the morning was written down as "test the keyboard, not just the
field" and then applied to only the fields that had been complained about. **A defect that appears
twice on one page is a rule, not two bugs.**

So every time on the form — clock and duration alike — is now four digits on a number pad: `0915` is
09:15, `0115` is 1:15. `inputMode="numeric"`, `maxLength=4`, a digits-only filter (so a pasted
`09:15Z` becomes `0915` instead of a field to clean up by hand), and an echo underneath reading the
digits back as the time they mean, because four unpunctuated numerals are quick to type and easy to
transpose.

**Exactly four, never three.** `915` is as readable as 91:5 as it is as 09:15, and this is a legal
record — a form that guesses is the silent corruption rule §0.2 forbids. A half-typed field is
refused by the form itself, naming the control, which is the one thing this form is allowed to decide
on its own; everything else still belongs to the server.

Two things deliberately did **not** change. The **wire format is still `HH:MM` / `HH:MMZ`**, composed
at submit, so `internal/timeutil` remains the single conversion authority (rule §0.4) and never
learns that the form's fields changed shape. And the **zone stays a toggle** rather than a typed `Z`,
for the reason argued this morning: the `Z` is load-bearing and a number pad cannot produce it.

The **total and the air time were already derived**; what changed is that they now recompute off
four-digit fields, so the figure appears the moment the fourth digit of the on-block time lands.

### 2026-08-01 — The phone would not pick up a new build, and that needed three layers

Reported by the owner in the same breath as the form ("do some pragma no cache so my phone will
reload the page"), and it is a deploy-correctness problem rather than a convenience one: the frontend
and the backend **ship together on purpose**, and a phone holding a stale `index.html` is precisely
how they come apart.

`index.html` is the only file under `/logbook/` whose **name stays the same while its bytes change on
every deploy** — everything else is content-hashed, which is why the assets can be cached for a year
and this one file cannot be cached at all. Three layers, each covering a device the others do not:

1. **`Cache-Control: no-store` + `Pragma` + `Expires`** on `index.html`, in the Apache block *and* as
   `<meta http-equiv>` in the document. The meta tags travel with the file, so a device served by
   anything other than this vhost is still covered.
2. **`fetch(request, {cache: 'no-store'})` for the shell in `sw.js`.** "Network first" was only ever
   as fresh as the HTTP cache underneath it — the worker would faithfully serve a stale document the
   browser handed it. The worker also no longer treats `/logbook/index.html` as an immutable asset,
   which it did through the catch-all rule.
3. **`reloadWhenUpdated`** (`src/swupdate.ts`): a home-screen PWA has no address bar and no reload
   button, so when a new worker claims the page, the page reloads itself onto it. Once — the latch is
   its own flag rather than `{once: true}`, because a reload loop on a phone at an airfield would
   break the app exactly where it is needed.

The cache name is bumped to `logbook-shell-v2`, so `activate` deletes the old shell outright.

### 2026-08-01 — The deploy scripts move into the repo, and the Apache installer stops refusing to update

Two rule-§0.1 defects found while getting the three 28/08/2025 flights to production.

**The scripts existed only on the box.** `update.sh`, `install-backend.sh`, `install-apache.sh`,
`apache-logbook.conf` and `logbook.service` lived in `/home/rami/logbook-deploy/` and nowhere else —
so a fresh clone of `origin/master` could not reconstruct the deploy, which is the bar §0.1 sets.
They are now in **`app/deploy/`**, edited there and rsynced to the box; never edited on the box.

**`install-apache.sh` could not deliver a changed snippet.** It skipped the insert entirely if a
`BEGIN logbook` block was already present — sound as a re-run guard, useless as a way to ship the new
cache headers. It now strips its own block, re-inserts the current snippet, and **refuses to write
unless stripping the block from the before and after files yields byte-identical text**: the proof
that nothing outside our block moved, on a vhost that serves seven other sites (rule §0.3). Backup,
`configtest`-before-reload and auto-restore are unchanged.

Rehearsed against a **copy** of the vhost before it was ever run as root, which is the part worth
keeping: it caught a blank line the inserter added on every pass, making each run differ from the
last — breaking both idempotence and the safety check meant to catch exactly that.

### 2026-08-01 — The table shows newest first, and that reversal lives in the view

Asked for by the owner: the flight list should open on the most recent flight, not on 2011. The
EASA export was explicitly noted as already correct, which is the constraint that decided where the
change goes.

There is **exactly one `ORDER BY seq`** in the store, and it feeds the table, the statistics and all
three PDFs. Reversing it — or adding a descending variant and pointing the list handler at it —
would have put the reversal one careless refactor away from the export, whose page geometry and
`TOTAL PREVIOUS PAGES` chain are built on ascending book order (rule §0.5). So the API still returns
the book's own order and **the reversal is a view concern in `TablePage`**, on a copy of the array
rather than in place, because `stats.Paginate` has a test asserting it does not reorder its caller's
slice and the same courtesy is owed here.

The subtle part is *what* gets reversed: **book order, not date order.** 21 rows across the three
books are genuinely out of date order, and three of them are the 28/08/2025 late entries now sitting
at the end of Book 3. Sorting the table by date would move those rows out of the order the paper
keeps them in — and a logbook that disagrees with the paper about row order is the beginning of
exactly the drift this project spent 106 KB of `drift.md` on. Both properties have a test.

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
