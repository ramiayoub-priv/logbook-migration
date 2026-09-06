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

⚠ **AS OF 2026-09-06 THERE IS WORK WAITING TO SHIP: Task 24, the flight-table PDF.** It touches
**both halves** — `app/backend/internal/pdfbook` + `internal/pdfmodel` and one copy string in
`app/frontend/src/pages/Export.tsx` — so it needs a **backend deploy**, which is the thing gated by
open item 2 below. Everything described in the table below was true and verified on **2026-09-04**;
it is still what is *live*, but the repo is now ahead of it in code, not only in documentation.

| | live | built from | checked |
|---|---|---|---|
| backend | `/opt/logbook/logbook-server` | **`6aed062`**, `vcs.modified=false` | route table + `vcs.revision` |
| frontend | **`index-D3Tqt5-U.js`** + `index-8vIbKNLy.css` | `ab32cee` | fetched back over HTTPS, **md5-matched** |

`origin/master` was at **`b153507`** when this was written, ahead of the deployed bundle by two
commits that were **documentation only**. **That is no longer true**: Task 24 (2026-09-06) changes
real code in both halves. Run the probe rather than trusting either sentence:

```bash
git log --oneline ab32cee..HEAD --stat | grep -E '^ app/(backend|frontend)/src'   # empty = docs only
```

**Health, checked 2026-09-04:** service `active`, **`NRestarts=0`**; `/logbook/api/health` **200**;
all eight private routes **401** without a session; the owner's **seven other sites all 200**; disk
57% used (21 G free), memory 312 MB of 1971 in use. **The daily off-box backup is alive** — timer
active, last run **2026-09-04 03:22 UTC, exit 0**, next 2026-09-05 03:23 UTC.

⚠ **One number this session could NOT read: the live flight count.** It needs either a session
cookie or `journalctl`, and `rami` has neither without sudo. It was **1298** on 2026-08-02 and only
goes up as flights are entered in the app. **Do not "fix" it to 1296** — see rule §0.8.

⚠ **THE BLOCK YOU ARE READING WAS WRONG FOR THREE WEEKS — read this before trusting any status
here.** It said Tasks 19/20/21 were "BUILT, TESTED, PUSHED — AND NOT DEPLOYED" and that production
was on the 2026-08-02 build. **They had been deployed on 2026-08-14**; the session that shipped them
recorded only the frontend half (Task 22) and left this brief describing a world that no longer
existed. A session on 2026-09-03 was told to deploy them and found there was nothing to deploy. That
is a **rule §0.1 failure** — the repo is the single source of truth, and for three weeks it lied.
**Verify status against the box, never against this file alone.** The cheapest three commands:

```bash
ssh rami@ayoub.fi 'strings -a /opt/logbook/logbook-server | grep -m1 vcs.revision'   # which commit is live
curl -s https://ayoub.fi/logbook/ | grep -oE 'assets/index-[A-Za-z0-9_-]+\.js'      # which bundle is live
git log --oneline <that-revision>..HEAD -- app/backend app/frontend                  # what is actually behind
```

Two probes that make this cheap, both learned the hard way (2026-09-03 decision log):

- **A 401 is not a missing route.** Under default deny nearly everything answers 401 without a
  session, which is indistinguishable from a route that does not exist — until you ask an obviously
  fake path (`/logbook/api/definitely-not-a-route-xyz`), which returns **404**. So a **401 proves the
  route exists**, and the whole route table reads without logging in.
- **`vcs.revision` answers "what is running"; md5 cannot.** Go stamps the commit into every binary.
  Two builds of the same source are the **same size with different hashes** — they differ only in
  that stamp and a build ID — so a hash mismatch proves nothing. Read the stamp, and read
  `vcs.modified` too: it says whether the live build was clean or came from a dirty tree.

**What is live**, full detail on the task board:

- **19 — the fleet page** at `/logbook/fleet`, reached from the Aircraft tab. Adds and corrects an
  aeroplane; `api.updateAircraft` finally has a caller. No delete, guarded on both sides.
- **20 — sessions stop piling up.** `SessionLifetime` 90d → **14d**, computed from `last_used_at`.
- **21 — the PIC name is picked, not spelled.** A `pilots` roster, a picker on the form, and
  `POST`/`PUT /flights` refusing a `pic_name` that is not on the roster exactly. `GET /pilots`
  answers **401** unauthenticated and `POST` **403**, so the route exists and default deny covers it.
- **23 — the aircraft picker's options ran together** (2026-08-18, deployed 2026-09-03). Task 21
  split a shared CSS selector at the wrong word, so the dropdown became a grid and its options lost
  their layout: `OH-CTLC172287 flights · 2026-08-14`, in two columns. Found by the owner **on a
  phone, from a screenshot** — the seventh time real use has beaten a green suite here, and the risk
  this very block had named. Guarded by a new `styles.test.ts`; the owner ruled the option down to
  **the registration alone**.

**Start here — there are three open items. Two of them are ONE terminal session (see below), and
after that a deploy never needs the owner again.**

**THE ONE SITTING**, owner ruling 2026-09-06 (*"yes paste it here fully so i can run it"*) — both
root items in one go, staged and waiting on the box:

```bash
ssh -t rami@ayoub.fi 'sudo /home/rami/logbook-deploy/install-deploy-privileges.sh && sudo /home/rami/logbook-deploy/install-apache.sh'
```

Then, from the dev machine, with **no password at all**: `app/deploy/deploy.sh`. That ships Task 24.

0. 🆕 **DEPLOY TASK 24 — the flight-table PDF carries the airborne times again.** Built, tested and
   pushed on **2026-09-06**, **not deployed**. The owner found it: *"There is a bug in the export
   (save as pdf) it only exports block times! not to and landing"*. Task 12 put Takeoff, Landing and
   Air on the **screen** on 2026-08-02 and never touched the **document**, so for five weeks the app
   knew four times per flight and the PDF it produced printed two. Backend + frontend, so it needs
   `sudo` on the box for the service restart — gated by item 2. `make check` **86.8%**, core
   **100%**; **128 frontend tests**. Nothing about the record itself changed: no figure, no schema,
   no stored value.

1. ⛔ **THE APACHE HALF OF TASK 22 IS STILL NOT APPLIED.** This is the one real piece of outstanding
   work. Verified live again 2026-09-04: `/logbook/assets/*` is still served
   `Cache-Control: public, max-age=31536000, immutable` **with an `ETag`**, which is exactly the rule
   the owner's *"make sure NOTHING is cached at all"* ruling abolished. The frontend half has been
   live since 2026-08-14 (no service worker), so the ruling is **half-honoured**. It is **not
   urgent** — Vite content-hashes every asset filename and `index.html` is already `no-store`, so a
   deploy still reaches the phone, which is exactly why this hid for three weeks — but the repo says
   one thing and the box does another. `deploy/install-apache.sh` is written and staged on the box;
   it needs **`sudo`**, so it needs the owner at a terminal.
2. ⚠ **Confirm the owner rotated the `rami` sudo password** — still the project's largest exposure,
   outstanding since 2026-08-02. It gates item 1 and every future backend deploy.

**Then, and this is the largest remaining risk:** open the app on the phone. See the warning below.

### How to work on this machine

- **Go is installed but NOT on a non-interactive shell's `PATH`.** It is at
  **`/home/havoc/.local/go/bin/go`** (go1.26.5, `GOPATH=/home/havoc/go`); nothing in `.bashrc` or
  `.profile` exports it, so a tool-run shell reports no `go` at all. **Prefix backend work with:**
  ```bash
  export PATH=$PATH:/home/havoc/.local/go/bin
  ```
  Then `make check` runs normally — **86.8%** overall, 100% on every `[core]` package, verified
  2026-09-04. **Backend deploys from this machine are perfectly possible.** An earlier version of
  this line claimed the opposite and was wrong; the 2026-09-04 decision-log entry records how, and
  the lesson (*silence is not a finding*) is worth more than the fact.
- ⚠ **`app/backend/dist/` holds a stale, DIRTY build** from `3921821` (`vcs.modified=true`). Never
  ship what is sitting there — rebuild.
- **A frontend-only deploy needs neither Go nor sudo.** `/var/www/logbook` is owned by `rami`:
  `npm ci && npm run check && npm run build`, tar the live directory for rollback, then
  `rsync -a --delete`. ⚠ Check `--delete` against the live listing first — `dist/` must produce all
  five entries the web root holds, including **`sw.js`**, the kill switch that retires service
  workers still installed on devices. No test catches its deletion, because the file is correct in
  the repo. Full procedure and rollback in `docs/deploy.md`.
- **`sudo` needs a password**, so nothing requiring root can be done from a session. That is correct
  and should stay.

⚠ **THIS WARNING HAS NOW BEEN PROVED RIGHT ONCE, AND STILL STANDS.** Almost nothing from 2026-08-03
has been looked at in a real browser. The one part that was — the owner opened the aircraft picker
on his phone on 2026-08-18 — **was broken**, and 127 green tests had nothing to say about it (Task
23). That makes **seven** times a green suite has loved what thirty seconds of real use exposed.
The fleet page and the pilot picker are still untouched by any finger, and `styles.test.ts` guards
one selector shape, not layout in general. **Two comboboxes and a list-with-inline-forms on a 390px
screen remain exactly that shape of risk.** Specifically worth checking: does the options list open
under the thumb, does the keyboard cover it, can you reach the "Add … as a new name" row, and does
the fleet page's edit form fit.

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

**✅ THE BOX IS LEVEL WITH THE REPO, AND THE IMPORTER IS RETIRED IN PRODUCTION.**

The three changes that were committed-but-undeployed all day on 2026-08-02 **shipped that evening**
and were verified from off-box:

1. **`logbookctl check`** — the restore check (Task 16). `RESTORE.md` used to tell the reader to run
   `sqlite3`, which is not on the box and is not a dependency of this project.
2. **Aircraft CRUD** (Task 17) — `POST /aircraft`, `PUT /aircraft/{reg}`, **no DELETE**, plus a
   schema migration adding `aircraft.user_added`, and a filterable aircraft picker in the form.
3. **`update.sh` no longer imports** (Task 18) — the owner's ruling below.

**The first production deploy that did not write to the legal record ran at 19:07 UTC** and is
described under "Deployed and verified" below. `update.sh` had never been executed in its rewritten
form until then; it is no longer the outstanding risk this block used to warn about.

**⛔ THE RULING THAT CHANGES THE DEPLOY (2026-08-02, owner, verbatim): "we should start treating the
production database now as the source of truth. We don't need the importer anymore."** See
`CLAUDE.md` §0.2. The migration is finished, the CSVs are frozen, and re-importing them could only
reproduce rows that cannot have changed — while running `DELETE` against a live legal record to do
it. **`update.sh` step 4 is now a READ-ONLY `verify`**, which turns the CSVs into a drift and tamper
check on the 1296 frozen historical rows instead of a rebuild. `logbookctl import` survives for dev
scratch databases and tests only. **The backup, not the repo, is what protects production.**

✅ **`update.sh` has now been run in its rewritten form (2026-08-02 19:07 UTC) and behaved exactly as
designed.** Step 4's read-only `verify` matched the live database against the frozen CSVs on **all
nine checksums** — 1296 / 1222:10 / 1054:45 / 167:25 / 107:58 / 22:45 / 189:41 / 407:39 / 3444 — which
is the first time that drift-and-tamper check has been exercised as the *point* of the deploy rather
than as a rebuild's afterthought. Nothing was written to the record. Step 4 failing still means *the
live historical rows disagree with the frozen books* — a defect to investigate, **never** a reason to
re-import.

### ✅ Deployed and verified 2026-08-02 (final session) — Tasks 16, 17, 18

The deploy a session could not previously finish: the owner supplied the sudo password in-session, so
steps 2–4 were run from here rather than by hand. **The password must be rotated as a result** —
`docs/security.md`, and it is now the second such exposure.

- **Backend**: `/opt/logbook/logbook-server` md5 **`f4b539aa9930419d820a81e6f45266a6`**, byte-equal to
  a `CGO_ENABLED=0` build of HEAD `c634556`. `logbookctl` **`59e089d3…`** installed alongside it —
  Task 16's fix, so a restore no longer arrives at step 3 without the tool.
  `logbook-server.prev` holds the previous `284a93fb…` for rollback.
- **Migration**: `aircraft.user_added` applied on first start with no incident; `flights=1298`
  unchanged across it, and a pre-migration copy sits in `/var/lib/logbook/backups/`.
- **Frontend**: `index-xgdC8L2o.js`, fetched back over HTTPS and **md5-identical** to the repo build
  (`fce1a22db8e0c0d69105b375128bbf81`), `index.html` pointing at it. Verified **by content**, not by
  filename: `Registration or type`, `as a new aircraft`, `No aircraft match.`, `New aircraft type`,
  `New aircraft class`, `route:"aircraft"`, `aircraft-time` and `Keep editing this flight` all
  present.
- **The new write routes are default-deny, and stricter than expected**: `POST /aircraft` and
  `PUT /aircraft/{reg}` answer **403**, not 401 — `checkOrigin` (`cmd/server/server.go:206`) wraps
  *outside* the auth check, so a state-changing request with no `Origin` header is refused as CSRF
  before authentication is even considered. Asserted in `flightentry_test.go:206`. **`DELETE
  /aircraft/{reg}` answers 405**: the no-delete ruling is live, not just tested.
- **Apache was deliberately NOT touched.** All three cache layers were already correct
  (`index.html` `no-cache, no-store, must-revalidate` + `Pragma` + `Expires: 0`; `sw.js` `no-cache`;
  hashed assets `immutable`) and `apache-logbook.conf` was byte-identical to the installed block, so
  `install-apache.sh` would have been a no-op reload on a box serving seven other sites (§0.3).
  **Skipping a no-op step is part of the runbook, not a shortcut.**
- **Service** active, **21.4 MB** against the 192 MB `MemoryMax`. **All seven other sites still 200.**

### ✅ Deployed and verified 2026-08-02 (late session)

Each line was checked from here by an independent probe, not read off a script's own output.

- **Backend**: `/opt/logbook/logbook-server` md5 **`284a93fb5d21a077118d8ed229cc0a04`** — byte-equal
  to a `CGO_ENABLED=0` build of this repo's HEAD. Tasks 11/12/13/14 are all in it.
- **Frontend**: `/var/www/logbook/assets/index-OExDQCHH.js`, fetched back over HTTPS and
  **md5-identical to the repo build**, with `index.html` pointing at it. Its *contents* were greped
  rather than trusted: `Log another flight`, `Keep editing this flight`, `See it in the table`,
  `Takeoff`, `route:"aircraft"` and `aircraft-time` all present.
- **API**: `/logbook/api/health` **200**; `/flights` and the new `/aircraft-time` both **401**
  unauthenticated — default deny survived, including on the route added today.
- **All seven of the owner's other sites still 200.**

⚠ **The frontend was two builds behind the backend for part of the day.** The live bundle was
`index-8hKHsSxu.js`, which greps **0** for every marker above, while the binary was already HEAD. So
the Aircraft tab existed in the API and not in the app. Deploying the pair in the right order —
**binary first, frontend second** — is what the pairing rule now means concretely; see the decision
log.

### ✅ The backup exists, off the box, and repeats

Task 14 is **installed and running**, which retires the item that stood as the project's largest
exposure for two days.

- First snapshot pushed: **`fc5cec9` — "Backup 2026-08-02: 1298 flights, 1223:03, 3446 landings"**,
  on branch `main` of the private `ramiayoub-priv/logbook-backup`.
- `logbook-backup.timer` is **enabled and active**, `OnCalendar=03:17:00 UTC` daily with
  `RandomizedDelaySec=600` and `Persistent=true` — so the exact minute moves (03:19, 03:22, …) and a
  box that was off at 03:17 still runs the backup when it comes back. **A shifting minute in
  `list-timers` is the jitter working, not a fault.**
- **1298, not 1296** — the production count includes the owner's two app-entered flights, which
  existed in exactly one place until this push. That is the whole point of the task.

✅ **AND IT COMES BACK.** Step 8 — the clone-back, the only check that is evidence about the *backup*
rather than about the push — **passed on 2026-08-02 17:20 UTC**. A fresh `git clone` of the remote
produced all four files, and `logbook.db` still hashed to what its own `MANIFEST.txt` claims:

```
logbook.db 421888 · logbook.csv 215012 · MANIFEST.txt 689 · RESTORE.md 2937 bytes
sha256 matches the manifest: 13a244e6f9042e64…
flights 1298 | users 1 | total time 1223:03 (73383 minutes) | landings 3446
```

`users 1` is load-bearing and deliberate: the account survives so the restored logbook can actually
be opened, which means the Argon2id hash leaves the box. That trade is argued in `docs/security.md`,
along with its consequence — **if that repository ever becomes public, the logbook password is
compromised.** Sessions are stripped.

`install-backup.sh` is idempotent and safe to re-run any time: it takes a fresh snapshot, pushes,
clones back, and re-enables an already-enabled timer.

⚠ **PRODUCTION IS NO LONGER REBUILDABLE FROM THIS REPOSITORY, AND THAT IS PERMANENT.** The owner
logged two flights in the app on 2026-08-02 and will keep logging more. They carry `source_book = 0`,
so the re-import inside `update.sh` leaves them alone (its `DELETE` is scoped, and there is a test) —
**confirmed against the live site rather than taken on trust**: the Flights page read 1298 after the
re-import. But they are in no CSV, so "rebuild it from the CSVs in one command" stopped meaning "lose
nothing". **The backup, not the repo, is what protects them.**

✅ **That exposure is now closed.** Those rows left the box for the first time on 2026-08-02 in
backup commit `fc5cec9` (**1298 flights**), and a timer repeats it daily. Before that push the only
copies were `/var/lib/logbook/backups/`, written by `update.sh` — **on the same disk as the database
they protect**, which defends against a bad import and against nothing else.

⚠ **THE LIVE COUNT AND THE TEST CONSTANTS ARE DIFFERENT NUMBERS, AND BOTH ARE RIGHT.**
`realdata_test.go` asserts **1296 / 1222:10 / 1054:45 / …** — those describe **the CSVs**, and they
are frozen (`CLAUDE.md` §0.8). Production is **1296 + every flight the owner has since entered in the
app**, so it read 1298 on 2026-08-02 and only goes up. A session that sees the live figure and the
test figure disagree **must not reconcile them**: they are answers to different questions, and
"fixing" the constant would silently unfreeze the historical record. The import's own checksums are
scoped `source_book <> 0` for exactly this reason.

**Status: LIVE at `https://ayoub.fi/logbook`, in real use, and the box is LEVEL WITH THE REPO** as of
2026-08-02 (final session) — binary **`f4b539aa…`**, `logbookctl` **`59e089d3…`**, bundle
**`index-xgdC8L2o.js`**, all three verified against HEAD from off-box. *(The figures below —
`284a93fb…` / `index-OExDQCHH.js` — are the previous evening's deploy, kept because the paragraph
records what was learned then; `284a93fb…` is now `logbook-server.prev`, the rollback target.)* The owner logged two flights from the field that day and both survived the re-import
(**1298** on the live Flights page, owner-confirmed) and are now in the off-box backup.

The re-import reported **54 discrepancies, not 61** — the 2026-08-02 aircraft-type ruling, expected
and correct — with the flight count and every time total unmoved. The frontend has a **sixth tab**
and "Statistics" is now "Stats".

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
  internal/stats/      Summarize, Range/Filter, Paginate, and aircraft.go (ByAircraft /
                       AirMinutes -- what each aeroplane COST, block vs air, with its
                       coverage). Computed, never stored.                             [core, 100%]
  internal/backup/     The off-box copy: VACUUM INTO -> redact sessions -> verify against
                       the live DB -> db + csv + manifest + RESTORE.md. Refuses rather
                       than half-writes. Driven by `logbookctl backup`.                    [78%]
                       NOT calculation core, so the 80% gate does not apply to it
                       individually; what is left uncovered is I/O error returns that
                       need a full disk to reach. Every behaviour that decides what the
                       backup CONTAINS is covered, including the restore actually being
                       signed in to.
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
  public/sw.js       a KILL SWITCH since 2026-08-14: deletes every cache, unregisters itself,
                     reloads the page. Nothing registers it -- it is deployed solely to retire
                     the workers already installed on devices. Do not delete it from the server.
                     (src/swupdate.ts is gone with it; nothing claims a page any more.)
  src/pages/         Login, Table, Statistics, Export, Review, Sessions, RangePicker, and
                     FlightForm -- the ONE form, wrapped by NewFlight and EditFlight. Two copies
                     would drift at the first fix applied to only one of them.

app/deploy/          the box's staging scripts, IN THE REPO as of 2026-08-01 (rule 0.1 -- they
                     lived only in /home/rami/logbook-deploy/ until then, which meant a fresh
                     clone could not reconstruct the deploy). update.sh, install-backend.sh,
                     install-apache.sh, apache-logbook.conf, logbook.service, and as of
                     2026-08-02 backup.sh + logbook-backup.service/.timer + install-backup.sh.
                     Edit them HERE and rsync them to the box; never edit them on the box.
```

`app/frontend/src/pages/AircraftTime.tsx` is the sixth page (Task 13). `src/format.ts` gained
`airMinutes`, which derives airborne time from two stored instants -- NOT the same function as
`FlightForm.blockFrom`, which rolls a bare four-digit clock that has no date attached.
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

⚠ **Run the thing before calling a task done.** A green suite has now SIX times missed what thirty
seconds of running found — most recently on 2026-08-03, when a `pic_name` refusal came back as a
field *map* where the form reads a field *array*, so the refusal rendered as a banner with nothing
marked below it and every test agreed it was fine. The earlier five: the `createuser -db` bug in Task 4; in Task 5/6 the broken PDF
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
discrepancies 54 | EASA export 87 pages
```
All seven `Cumulative_*` series reconcile with **zero breaks**.

⚠ **Discrepancies moved 61 → 54 on 2026-08-02** and that was an owner ruling, not drift: the five
aircraft-type cells (`C192` ×4, `OH-CMU` ×1). **No other figure above moved by a single minute** —
sea/land comes from the registration, not the type. See the decision-log entry "The five cells that
could not be true". Anything else moving is still a defect.

⚠ **These moved once, on 2026-08-01**, and the previous values are still all over the git history:
they were `1293 / 1219:35 / 1053:03 / 166:32 / 107:05 / 3439`. Three flights of **28/08/2025** were
missing from `logbook_3.csv` entirely — see the decision log and `claude-docs/drift.md`. **They will
not move again**: the dataset was closed on 2026-08-02 (`CLAUDE.md` §0.8).

Asserted in `internal/csvbook/realdata_test.go` and again, by a different code path, in
`internal/stats/realdata_test.go`. **If one of them changes unexpectedly, the import is wrong until
proven otherwise — do not adjust the expectation to make the test pass.**

⚠ **These tests have exactly ONE legitimate reason to fail, and it is not yours to invoke.** Until
2026-08-02 the reason was `logbook_3.csv` growing as pages were transcribed, and this file carried a
procedure for updating the constants when it did. **That procedure is void.** The only thing that may
move a constant now is an **explicit owner ruling on named cells** — it has happened twice, and both
times the owner said which cells and why (the three missing 28/08/2025 flights; the five
aircraft-type cells). Absent that, a red `realdata_test.go` means the importer, the store or the
stats code has broken. **If you find yourself editing a number in `realdata_test.go` without the
owner having named the cells, stop and re-read `CLAUDE.md` §0.8.**

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

✅ **What is NO LONGER on this list, closed 2026-08-02:** the `C192` type (4 rows) and `OH-CMU`'s
type (1 row). The owner ruled them transcription slips. **Do not reopen them and do not read the
line above as licence to close the other two** — the difference is that a Cessna 192 does not exist,
so no page could correctly say it, whereas lines 89–90 turn on a physical page nobody will re-read.

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
POST   /aircraft           private  adds an aeroplane never flown -> 201; 409 duplicate registration
PUT    /aircraft/{reg}     private  corrects one, including a rename. NO DELETE, by ruling (405).
GET    /pilots             private  the PIC roster: every distinct pic_name on a flight, counted,
                                    plus names added here and not yet flown with. Ordered
                                    never-flown-with first, then most recent.
POST   /pilots             private  adds a name -> 201; 409 if the roster knows it in ANY case,
                                    which is the feature. No PUT and no DELETE: a roster entry is
                                    a spelling, and a wrong name is corrected on the flight.
                                    POST/PUT /flights refuse a pic_name not on the roster EXACTLY.
GET    /stats     ?from&to private  {summary:{...}, range}
GET    /aircraft-time      private  ?from&to&reg -- what each aeroplane cost.
                                    {range, reg, aircraft:[...], total:{...}, flights:[...]}
                                    Block and air are SEPARATE fields with separate coverage
                                    counts (air_known/air_missing) and must never be merged:
                                    block is recorded on all 1296 rows, air on 19.
                                    `reg` adds that aeroplane's flights; the summary rows
                                    always cover the whole range either way.
GET    /discrepancies      private  the "needs review" list, 54 rows today (was 61 before the
                                    2026-08-02 aircraft-type ruling)
GET    /sessions           private  the revocable device list; `current` marks the caller
DELETE /sessions/{id}      private  revoke one, scoped to the owner
GET    /export/easa.pdf        private  the whole logbook, EASA format. IGNORES from/to on purpose.
GET    /export/table.pdf       ?from&to private
GET    /export/statistics.pdf  ?from&to private
```
`from`/`to` are inclusive `YYYY-MM-DD`. An unparseable one is a **400**, never an ignored filter.

**Operator CLI** (no HTTP route exists for any of it, by design):
`./dist/server createuser|passwd|users|disable|enable <name> -db <path>`, and
`logbookctl import|verify|backup|check`. **`verify` and `check` answer different questions**:
`verify` compares a database against the three CSVs (scoped `source_book <> 0`, so it passes while
every app-entered flight is missing), while **`check -db <db> -manifest MANIFEST.txt` is the RESTORE
check** — it needs no CSVs and no `sqlite3`, and covers the hand-entered rows that exist nowhere
else. Do not substitute one for the other.

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

✅ **That item is now confirmed, 2026-08-02 (final session).** The startup line was read directly off
`update.sh`'s output: **`flights=1298`**. It is no longer expected to say 1296 — see the warning under
step 2 below.

**The runbook, for next time** — five steps, in this order. The `sudo` ones are **the owner's unless
they hand over the password in session**, which happened on 2026-08-02 (and cost a rotation — see
`docs/security.md`). Before running it, note that the staged CSVs on the box are confirmed
**byte-identical (md5) to the repo's**; that is now a *verification* input rather than an import
input, but it is still checked, because step 4 compares the live database against those files.

```bash
# 1. Stage the current build. rami owns /home/rami/logbook-deploy, so no sudo.
#    Binaries are cross-compiled CGO_ENABLED=0 from this repo's HEAD.
#    Re-run it whenever you have rebuilt; then re-verify the md5s, do not assume.
#    ⚠ The `csv/` copy is the one update.sh reads. Staging only the top-level
#      copies leaves step 4's VERIFY comparing against whatever was there before.
cd app/backend && export PATH=$HOME/.local/go/bin:$PATH
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /tmp/logbook-server ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /tmp/logbookctl   ./cmd/logbookctl
cd ../.. && rsync -a /tmp/logbook-server /tmp/logbookctl \
    logbook_1_final.csv logbook_2_final.csv logbook_3.csv \
    rami@ayoub.fi:/home/rami/logbook-deploy/
rsync -a logbook_1_final.csv logbook_2_final.csv logbook_3.csv rami@ayoub.fi:/home/rami/logbook-deploy/csv/
rsync -a app/deploy/ rami@ayoub.fi:/home/rami/logbook-deploy/   # scripts live in the repo now

# 2. sudo: backup + new binaries + READ-ONLY verify + restart. NO IMPORT.
#    The startup log line reads the LIVE count: 1296 + every app-entered
#    flight. It was 1298 on 2026-08-02 and only goes up. NOT 1296.
ssh -t rami@ayoub.fi 'sudo /home/rami/logbook-deploy/update.sh'

# 3. The frontend, AFTER step 2 -- the ORDER matters, see the pairing warning
#    below. No sudo, so a session can do this one.
cd app/frontend && npm run build && cd ../..
rsync -a --delete app/frontend/dist/ rami@ayoub.fi:/var/www/logbook/
#    Then VERIFY BY CONTENT, not by filename -- fetch the live bundle back and
#    grep it. A matching hash in dist/ proves nothing about what Apache serves:
#    curl -s https://ayoub.fi/logbook/ | grep -o 'index-[A-Za-z0-9]*\.js'
#    curl -s https://ayoub.fi/logbook/assets/<that>.js | grep -c 'route:"aircraft"'

# 4. OWNER, sudo: the Cache-Control headers, so the phone stops serving the old build.
ssh -t rami@ayoub.fi 'sudo /home/rami/logbook-deploy/install-apache.sh'

# 5. OWNER, sudo: the daily off-box backup (Task 14). ALREADY INSTALLED and the
#    timer is enabled -- this is now only needed to re-verify, or after editing
#    backup.sh / the units. It is idempotent and takes a fresh snapshot each run.
ssh -t rami@ayoub.fi 'sudo /home/rami/logbook-deploy/install-backup.sh'
```

⚠ **Step 2 NO LONGER IMPORTS, so it no longer reports a discrepancy count at all.** It prints the
nine `verify` checksums instead. What they must read is the frozen-CSV figures — **1296 / 1222:10 /
1054:45 / 167:25 / 107:58 / 22:45 / 189:41 / 407:39 / 3444** — because `verify` is scoped
`source_book <> 0` and therefore describes *the historical rows only*. **The startup line, which
counts everything, read `flights=1298` on 2026-08-02.** Those two numbers disagreeing is the design,
not a fault; see the warning about it further up.

⚠ **A step-4 mismatch is NEVER a reason to re-import.** It means the live historical rows have
drifted from the frozen books — a defect to investigate before the service writes again (`CLAUDE.md`
§0.2). Note also that step 4 runs under `set -euo pipefail` **before** step 5 starts the service, so a
failed verify leaves the service **stopped**. That is deliberate, but it means you must not walk away
from a red step 4: the site is down until it is resolved or the service is started by hand.

⚠ **`install-apache.sh` now REPLACES its block rather than skipping it.** The first version refused
to touch a vhost that already had a `BEGIN logbook` block, which meant a changed snippet could never
reach the server. It now strips its own block, re-inserts the current `apache-logbook.conf`, and
**refuses to write if stripping the block from the before and after files does not produce identical
text** — the proof that nothing outside our block moved. Backup, `configtest` and auto-restore are
unchanged. Rehearsed against a copy of the vhost before it was ever run as root, which caught two
bugs: an inserted blank line made every run differ from the last, breaking both that safety check
and idempotence.

`update.sh` backs the database up to `/var/lib/logbook/backups/` **first, always**, installs the new
binaries (keeping `logbook-server.prev` for rollback), runs the **read-only `verify`**, restarts, and
re-checks the other six sites. The backup step survives the importer's retirement on purpose: the
server may apply an **additive schema migration** on first start, and a copy taken beforehand costs a
second and is the only way back.

⚠ **Never a file swap.** The `users` and `sessions` tables live in the same SQLite file, so replacing
it would delete the owner's account. This mattered when the importer ran, and it still governs any
future restore. The importer's `DELETE` was scoped to `aircraft`,
`discrepancies` and `flights WHERE source_book <> 0`, inside one transaction that rolls back on any
checksum mismatch — so users, sessions and app-entered flights all survive.

⚠ **The binary and the frontend must land together, BINARY FIRST.** The reworked form sends
`takeoff`/`landing`; Go's JSON decoder ignores unknown fields, so an *old* binary would accept the
flight and **silently drop two of its times**. That is the case for shipping the pair.

The case for the **order** is newer (2026-08-02): a new frontend against an old binary calls routes
that do not exist — the Aircraft tab would `404` on `GET /aircraft-time`. The reverse, a new binary
under an old frontend, is harmless: unused routes simply sit there, which is exactly what production
looked like for part of that day. **So when they cannot be simultaneous, the binary goes first.**

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

### Where to pick up — 2026-08-03 (built, not deployed)

⚠ **THIS SECTION'S HEADLINE IS NOW STALE BELOW THIS LINE.** As of 2026-08-03 the deploy is **not**
done: Tasks 19, 20 and 21 are committed and pushed at `7f54c26` and the box is still on the
2026-08-02 build. The brief at the top of this file is the current state; what follows is the
2026-08-02 pick-up list, kept because items 1 and 2 are still open and item 3 is the history behind
Task 19.

**(2026-08-02, still true except where noted.)** The two items below were the ones that mattered
then; item 3 is now **done** (Task 19).

1. ⚠ **ROTATE THE `rami` SUDO PASSWORD — this is now the top item, and it is overdue.** It was pasted
   into a chat session on 2026-08-01, typed repeatedly on 2026-08-02, and **handed over in-session
   again on 2026-08-02** so that this deploy's `sudo` steps could be run without the owner. The owner
   said they would rotate it "as soon as we are done". **We are done.** Until it is rotated, a
   credential to a box serving eight sites is sitting in two chat transcripts. Tracked in
   `docs/security.md`.
2. **OPEN THE AIRCRAFT PICKER IN A REAL BROWSER, ON THE PHONE.** It is now **live** and still has
   **111 green tests and has never been looked at**. On this project a green jsdom suite has now
   **five times** loved something that thirty seconds of real use exposed — the untypeable colon, the
   stale service worker, the off-screen save confirmation, the empty clone, the unrunnable restore
   instructions. A dropdown that opens under the thumb, on a form, on a phone, is exactly that shape
   of risk. Specifically: does the list close when you expect, does the keyboard cover it, can you
   reach the "Add …" row.
   ⚠ **The phone may still serve the old bundle** — a service worker installed before today survives
   a deploy. The live bundle is provably `index-xgdC8L2o.js`; if the picker is missing on the phone,
   check the device, not the deploy (see the trap below), and escape it with `/logbook/?v=3` plus a
   full close-and-reopen of the home-screen app.
3. **THE FLEET MANAGEMENT PAGE IS NOT BUILT — this is where tomorrow's work starts.** The owner
   asked on 2026-08-02 whether "the aircraft CRUD is live", and the honest answer is *partly*. The
   exact state, verified against the source rather than remembered:

   | | API | Reachable in the UI? |
   |---|---|---|
   | **R** — `GET /aircraft` | live | ✅ feeds the picker (`FlightForm.tsx:170`) |
   | **C** — `POST /aircraft` | live | ✅ the picker's "Add … as a new aircraft" row (`AircraftPicker.tsx:220`) |
   | **U** — `PUT /aircraft/{reg}` | live | ❌ **nothing calls it** |
   | **D** | does not exist, by ruling | n/a — enforced live (**405**) |

   **What the owner actually asked for works**: an aeroplane never flown before can be added inline
   and the flight logged, without leaving the form. **What is missing is the U.**
   **`api.updateAircraft` (`src/api.ts:301`) has exactly zero callers — not even a test.** It is dead
   code in the frontend sitting on a working, deployed, tested endpoint.

   **The consequence to state plainly: a typo'd registration or a wrong type cannot be corrected from
   the app, and cannot be deleted either (by design).** That combination is the gap — the no-delete
   ruling is only humane if editing exists, and right now it does not.

   It is a **frontend-only job**: the endpoint, its validation, the `user_added` scoping and the store
   tests all exist and are deployed. It needs a page that lists the fleet and lets a row be edited,
   built failing-test-first (§0.6). Shipping it is a frontend-only redeploy — **no `sudo`, and it does
   not touch the record.** ⚠ Do **not** add a delete route while you are in there; that ruling is
   asserted by a route-table test precisely because "symmetry" is how it would get lost.

**Then, in rough order of value (all pre-dating this session):**

1. ✅ **DONE 2026-08-02 — the restore drill.** The backup was cloned, verified, booted and read back
   without SQLite; it passed on every substantive claim. Its **instructions did not** — step 3 told
   the reader to run `sqlite3`, which is not on the box and is not a dependency of this project.
   Fixed with `logbookctl check`; see the decision log. **Still worth doing once by hand on the
   owner's side**: the one thing a session cannot test is that the *password* still works on a
   restored database, because it has never been in a file or a session.
2. **Open the Aircraft tab on the actual phone**, now that it is live. Task 13's figures are proven
   against the real books and the page is proven in jsdom, but **it has never been looked at in a
   real browser** — and six tabs plus a seven-column table are exactly the shapes that have broken
   on this project before. Four times now a green suite has loved something that thirty seconds of
   real use exposed. Same for Task 12's three new table columns.
3. **Rotate the `rami` sudo password** — see the top of this list; it was exposed a **second** time on
   2026-08-02 and is no longer merely the oldest open item, it is the most urgent one.

**Task 10's ruling is spent** — edit/delete shipped on 2026-08-02. The old open question recorded
here (correct in place vs. append a correction) was answered: app-entered flights only, real delete
with an append-only `flight_audit` copy, double confirmation. Nothing left to decide.

**Security and the box:**
- ⚠ **Rotate the `rami` sudo password. UPGRADED TO THE TOP OPEN ITEM 2026-08-02.** It was pasted into
  a chat session on 2026-08-01, typed repeatedly during that day's deploys, and **handed over
  in-session on 2026-08-02** to let this session run the deploy's `sudo` steps unattended. Two
  transcripts now contain it. The owner accepted the trade explicitly ("i will change the password as
  soon as we are done") — **and we are done.** Tracked in `docs/security.md`.
- Is the `kraken-predictor-python-2` container on `:8000` still wanted? Publicly exposed, up 2 years,
  and the box's largest memory consumer (~759 MB / 38%).
- Prune the stale ufw rules for `30814` and `19132` (nothing listens on either)?

**Smaller app items, none blocking:** offline *writes* are still out of scope (§2); the new-flight
form's four-digit layout has been proven in tests and on the owner's phone but never inspected in a
desktop browser at 390px; and `MemoryMax` is still set from an estimate — the real peak during a
full-logbook PDF export has not been measured (`docs/deploy.md` asks for it).

### Traps already paid for — do not rediscover these
- **`GIT_SSH_COMMAND` is read by `git`, and by nothing else.** Setting it and then running a bare
  `ssh` gets you the *default* identity, not the one in the variable. This made
  `install-backup.sh`'s key preflight fail identically whether the key was perfect or missing, and
  sent two sessions after GitHub. **A check that cannot fail for the reason it claims to test is
  worse than no check.**
- **`systemctl status` exits 3 for a finished oneshot** — "inactive (dead)" is its *success* state.
  Under `set -euo pipefail` that aborts the script. Never let a human-readable status page decide
  control flow; ask `systemctl show -p Result --value` instead.
- **`ls-remote` on an empty repo succeeds and prints nothing.** "No output" is not "cannot reach";
  branch on the exit status, or a normal first run gets reported as an auth failure.
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
| 7b | **Deploy Tasks 11–14 to production** | **done** 2026-08-02 (late) — binary `284a93fb…` and bundle `index-OExDQCHH.js` both live and verified from off-box: the bundle was fetched back over HTTPS and md5-matched to the repo build, then greped for the save-takeover strings, `Takeoff`, `route:"aircraft"` and `aircraft-time`. `/aircraft-time` answers **401** unauthenticated, so default deny covers the new route. All seven other sites still 200. For part of the day the box ran the **new binary with the old frontend** (`index-8hKHsSxu.js`, 0 markers) — harmless in that direction, and the reason the pairing rule now specifies an *order*. |
| 7 | PWA + deploy to `ayoub.fi/logbook` | **PWA done** 2026-08-01 — manifest, icons, and a hand-written service worker that caches the shell and **never** an API response, proven in a browser with the HTTP cache disabled; the shell is now fetched `no-store` and a new worker reloads the page (`src/swupdate.ts`). **Deployed and LIVE** 2026-08-01 at `https://ayoub.fi/logbook` — service user, binary, systemd unit, frontend, the additive Apache block and the account are all in place, and the owner's other seven sites still answer 200. The deploy scripts now live in **`app/deploy/`** instead of only on the box. ✅ **Finished.** The 1293-vs-1296 gap this row used to record was closed by the 2026-08-02 re-import, and the box has been level with the repo at every deploy since; the last one (Tasks 16/17/18, 19:07 UTC) put binary `f4b539aa…` and bundle `index-xgdC8L2o.js` live. See "Deployed and verified 2026-08-02 (final session)". |
| 8 | Backfill landings day/night for the **30** night rows | **WILL NOT DO** — closed 2026-08-02 by the owner's ruling that historical data is not to be touched (`CLAUDE.md` §0.8). The 30 rows keep their `landings_unverified` flag **permanently**: the API reports the count, the table asterisks the row and the statistics page names it in a paragraph. That is the honest state and it must not be quietly dropped to make a page look tidier. |
| 9 | Rule on the open source-data problems | **closed** 2026-08-02 — two of three were ruled and fixed on 2026-08-01. The third (`logbook_2_final.csv` lines 89–90, `04.05.2018` ×2) stands **unresolved forever**: it affects row order only, moves no total, and settling it needs a physical page that will not be re-read. Recorded, not fixed. |
| 15 | **The `C192` / `OH-CMU` aircraft types** | **done** 2026-08-02 — owner ruling, five cells in one column: `C192` → `C172` (book 2 lines 132, 133, 137, 138) and `OH-CMU` → `C152` (book 3 line 434). There is no Cessna 192, and book 2 line 139 is the same aeroplane the same day written `C172`. Discrepancies **61 → 54** (`unknown_aircraft_type` 4 → 0, `type_conflict` 3 → 0); **no other figure moved at all** — sea/land comes from the registration. Closed permanently by `TestEveryRegistrationNamesOneRealAircraftType`, watched red before it was accepted. Second and last lift of the §0.8 freeze to date. |
| 11 | **Saving a flight must be unmissable** | **done** 2026-08-02 — the success **takes over the screen**: the form is replaced by a confirmation naming what the SERVER stored (date, registration, route, both clock pairs, total, landings), scrolled to and focused, offering *Log another flight* / *Keep editing this flight* and *See it in the table*. A refusal gets the same weight in red and **jumps to the first failing control by page order**, not by the order the server listed them. The draft survives every failure path, and there is still **exactly one live region**. |
| 12 | **Takeoff / landing / air time in the table, and out of the disclosure** | **done** 2026-08-02 — the flights table gains **Takeoff, Landing and Air**; air time is `format.airMinutes`, computed at render from the two instants and **never stored**, blank (never `0:00`) on the 1277 rows that have none. The airborne pair is **out of the `<details>`** and sits in the Times card next to off/on block; the `details.airborne` CSS is gone. |
| 13 | **Aircraft time page (block vs air, by aircraft and date range)** | **done** 2026-08-02 — `internal/stats/aircraft.go` (`AirMinutes`, `ByAircraft`, `TotalAircraftTime`; pure, **100%**), `GET /aircraft-time?from&to&reg`, and the **Aircraft** tab. Block and air are separate fields with separate coverage and are never mixed; both totals in **H:MM and whole minutes**; `reg` adds the flights behind one figure without narrowing the comparison. The real books make the case: **OH-CTL has 267:16 of block time and 2:51 of air time from 4 of its 286 flights** — one merged "hours" figure would be catastrophically wrong. |
| 14 | **Daily off-box backup to a private git repo** | **INSTALLED AND RUNNING** 2026-08-02 — `logbookctl backup` (`internal/backup`) writes four files: `logbook.db` (sessions stripped), `logbook.csv` (every flight, every field), `MANIFEST.txt`, `RESTORE.md`. A systemd timer runs `backup.sh` as the **`logbook` user** → commit → push to `ramiayoub-priv/logbook-backup`. First snapshot pushed as **`fc5cec9` — 1298 flights, 1223:03, 3446 landings**; timer **enabled and active**, next **2026-08-03 03:22 UTC**. Auth is an **account-level key on the dedicated `ramiayoub-priv` account** (owner ruling — not a deploy key). Installing it took four attempts and exposed **four bugs in `install-backup.sh`'s own checks and none in the backup** — see the decision log. ✅ **Step 8, the clone-back, passed**: four files out of a fresh clone, `logbook.db` matching its own manifest sha256, 1298 flights / 1223:03 / 3446 landings / 1 user. |
| 16 | **The restore drill, and `logbookctl check`** | **done, and DEPLOYED 2026-08-02 19:07 UTC** (`logbookctl` `59e089d3…` now on the box, installed by `update.sh` step 3) — the backup was cloned and restored for real with no emergency running. It passed everything: both sha256s match, `logbook.db` is byte-identical across three snapshots, the server boots on it reading **`flights=1298`** with all six private routes still 401, and `logbook.csv` reconciles to 1298 / 1223:03 / 3446 / 38 with no SQLite involved. **Its instructions did not pass**: step 3 told the reader to run `sqlite3`, absent from the box and not a dependency of this project, so the mandatory rule-0.2 verification was `command not found` on a fresh server. New **`logbookctl check -db <db> [-manifest <file>]`** (no CSVs, no sqlite3, hashes before opening, shares `Figures` with the manifest writer so the two cannot drift), regenerated `RESTORE.md`, and **`install-backend.sh` now installs `logbookctl`** — step 1 of the restore never did, so fixing only the sqlite3 line would have swapped one missing command for another. Backend **87.6%**, core still 100%. |
| 17 | **Aircraft CRUD** | **backend + picker DEPLOYED, manage page NOT built** 2026-08-02 — owner ask: a first flight in an aeroplane never flown was unenterable, because the aircraft list was purely derived and the form's registration was a `<select>` fed by it. Now: `aircraft.user_added` (additive migration in `store.migrate`, proven safe on a copy of real production), `store/aircraft.go` (`AircraftList`/`AircraftByReg`/`AddAircraft`/`UpdateAircraft`, `last_flown` and flight counts **derived, never stored**), **`POST /aircraft`** and **`PUT /aircraft/{reg}`** — and **NO DELETE**, by ruling, asserted against the route table. The importer's unqualified `DELETE FROM aircraft` is scoped to `user_added = 0`. Frontend: `AircraftPicker.tsx`, a filterable combobox that also adds an aeroplane inline; **no retired/active concept**, nothing hidden, ordered never-flown-first then most-recently-flown. **111 frontend tests, backend 87.3%.** ✅ **Live since 2026-08-02 19:07 UTC** — `POST`/`PUT` both answer **403** unauthenticated (CSRF refused before auth, stricter than 401) and `DELETE` answers **405**, so the no-delete ruling is enforced in production. ⚠ **Still never opened in a real browser, and the U of CRUD has no UI at all**: `api.updateAircraft` (`src/api.ts:301`) has **zero callers, not even a test**, so a typo'd registration cannot be corrected from the app and cannot be deleted either. Create/read work end-to-end. See pick-up item 3 for the full table — that is where the next session starts. |
| 18 | **Retire the importer from production** | **done and DEPLOYED** 2026-08-02 — owner ruling: the production database is the source of truth. **The rewritten `update.sh` ran at 19:07 UTC**: step 4's read-only `verify` matched all nine checksums against the frozen CSVs and nothing was written to the record. That was the first deploy in this project's history that did not run a destructive operation on a live legal record. `update.sh` no longer imports; step 4 is a **read-only `verify`**, turning the frozen CSVs into a drift/tamper check on the 1296 historical rows rather than a rebuild. `CLAUDE.md` §0.2 rewritten. `logbookctl import` survives for dev scratch databases and tests only. Removes the stale-CSV class of failure entirely, and rests on the backup having been *proven* restorable the same day. |
| 22 | **Nothing is cached, anywhere** | **built and half-deployed 2026-08-14** — owner: *"make sure NOTHING is cached at all… the browser needs to forget (except the cookie for the session)."* `public/sw.js` becomes a **kill switch** (deletes every cache, unregisters itself, reloads the page); `src/main.tsx` registers no worker and `src/noworker.test.ts` fails if that returns; `swupdate.ts` deleted. Apache serves the whole directory `no-store` with no `ETag`, replacing three per-file rules. **Frontend is LIVE** (`index-qD3NNzOE.js`, superseded 2026-09-03 by `index-D3Tqt5-U.js`); ⛔ **the Apache half needs `sudo` and is STILL NOT APPLIED** — `install-apache.sh` is staged on the box. **Re-verified live 2026-09-03**: `/logbook/assets/*` still answers `Cache-Control: public, max-age=31536000, immutable` **with an `ETag`**, and `index.html` still carries an `ETag` alongside its `no-store`. The ruling is **half-honoured** — no service worker, but Apache still tells the browser to keep an asset for a year. Not urgent (content-hashed filenames plus a `no-store` index mean a deploy still reaches the phone) and not silently droppable either. Trade, stated: the app no longer opens without a network, and ~200 KB is fetched per cold start. The session cookie is untouched. |
| 19 | **The fleet management page** — the `U` of aircraft CRUD, plus create away from the form | **done, and DEPLOYED 2026-08-14** (recorded only on 2026-09-03 — see that day's decision-log entry; this row said "not yet deployed" for three weeks while it was live) — `pages/Fleet.tsx` at **`/logbook/fleet`**, reached from the Aircraft tab ("Manage the fleet") rather than a seventh tab: six already share a 390px phone. Lists every aeroplane in the server's order, adds one with all five fields (the form's inline panel asks only type and class, on purpose), and **corrects one — `api.updateAircraft` now has a caller**, keyed by the old registration so a rename works. **No delete**, guarded now on both sides of the wire: a route-table test on the backend and a frontend test that fails if any control ever says delete. **117 frontend tests** (was 111), six of them new and all watched red first. Frontend-only; no `sudo`, and it cannot touch the record. |
| 20 | **Stale sessions never die** — the Devices list is a graveyard | **done, and DEPLOYED 2026-08-14** (recorded only on 2026-09-03; this row said "not yet deployed" for three weeks while it was live) — `SessionLifetime` 90d → **14d** per the owner's ruling, and the window is now **computed from `last_used_at` against the constant instead of read back from the stored `expires_at`**. That second half is what reaches the rows that already exist: the owner's thirteen were each stamped with a date up to three months out, and a fix that only applied to new sessions would have left every one of them on the page. `LookupSession`, `PurgeExpiredSessions` and the Devices listing all ask the same question, so the sweep, the request path and the page cannot disagree. **No schema change.** Four new tests, three watched red on the old code and the fourth red the moment the constant moved; backend **87.2%**, core 100%. |
| 21 | **The PIC name becomes a picked object** | **done, and DEPLOYED 2026-08-14** (recorded only on 2026-09-03; verified live — `GET /pilots` answers **401** unauthenticated and `POST /pilots` **403**, so the route exists and default deny covers it) — a `pilots` roster: the distinct `pic_name` values already on flights (counted, dated) UNION names added in the app and not yet flown with, behind a filterable picker on the form. `GET`/`POST /pilots`, **no PUT and no DELETE** (a roster entry is only a spelling — a wrong name is corrected on the flight). **`SELF` cannot join `self`**: refused by a `UNIQUE … COLLATE NOCASE` index *and* by a check against the names already on flights, and the picker will not even offer to add a case variant. **The guarantee is at the write, not just in the form** — `POST`/`PUT /flights` refuse a `pic_name` that is not on the roster exactly, which can never block an edit because the roster is derived from the flights. Blank stays legal (one paper row has it). **Owner ruling: NO student field** ("no student field as it should be"). Historical values are read and never rewritten (§0.8) — `Sinervä`/`Sinerva` and `Stude` are all still offered exactly as written. Backend **86.8%**, core 100%; **124 frontend tests**. |
| 23 | **The aircraft picker's options ran together** | **done, and DEPLOYED 2026-09-03 20:33 UTC** (bundle `index-D3Tqt5-U.js`, css `index-8vIbKNLy.css`, both fetched back over HTTPS and md5-matched to the repo build) — the owner's phone screenshot showed the dropdown reading **`OH-CTLC172287 flights · 2026-08-14`** and laid out in **two columns**. Both came from one malformed selector list in `styles.css`: folding the pilot picker in (Task 21) split `.aircraftpicker .options button` at the wrong word, leaving `.aircraftpicker .options,` `.pilotpicker .options button { display: grid; … }`. So the aircraft **list** became the three-column grid and its **options** got no layout, no gap, nothing — on six rules, one of which set `display: none` on the whole list below 22rem. Fixed by pairing every selector properly. In the same change, **owner ruling: "It's enough to just show the registration"** — the type and the `N flights · date` tail are gone from the option; the type still **filters**. New **`src/styles.test.ts`** asserts the selector shape (three tests, all watched red), plus an exact-equality test on the option's text. **128 frontend tests.** Frontend-only. |
| 25 | **A deploy must not need root** | **built and rehearsed 2026-09-06, ONE ROOT COMMAND OUTSTANDING** — owner: *"we should find some solution for that, deploy should not need root"*. `app/deploy/deploy.sh` now runs the entire deploy **from the dev machine with no password**: clean-tree guard, `make check` + `npm run check`, build both halves, stage with a `SHA256SUMS`, `ssh … sudo -n /opt/logbook/logbook-apply`, then the frontend, then off-box verification. The only privileged step is `logbook-apply`, root-owned, named by a **one-user one-command no-arguments** rule in `/etc/sudoers.d/logbook-deploy`. It grants **nothing new** — `rami` is in the `sudo` group already — it removes the prompt from one audited operation, and the installer **asserts** that general `sudo -n` still fails afterwards. Polkit was ruled out on evidence: Ubuntu 20.04 ships **polkit 0.105**, `.pkla` only, which cannot scope manage-units to a single unit and would have handed rami restart rights over all seven other sites. `logbook-apply` refuses a missing/mismatched `SHA256SUMS` (a truncated rsync leaves a plausible, wrong binary) or a symlinked artefact, all **before** stopping the service, and **rolls back automatically** to `logbook-server.prev` if the new binary fails its health check — because unattended means nobody is reading the output. **Rehearsed against a fake tree with stubbed `systemctl`/`curl` before it was ever run as root**: all four refusals exit 1 with nothing installed, plus the happy path and the rollback. `update.sh` is now a shim that execs `logbook-apply`, so there is one implementation. ⛔ **Needs the owner once**: `sudo install-deploy-privileges.sh`. |
| 24 | **The flight-table PDF dropped the airborne times** | **built and pushed 2026-09-06, NOT DEPLOYED** — owner: *"There is a bug in the export (save as pdf) it only exports block times! not to and landing"*. Right: `tableColumns` in `internal/pdfbook/table.go` carried `OFF` and `ON` and nothing else, so **Task 12's Takeoff/Landing/Air went to the screen on 2026-08-02 and never to the document**. The export whose own description promised "everything the application knows about it" was throwing away two of the four times it knows — on **36 of the 1313 rows**, including **every one of the 17 entered through the app**, which is why the owner met it the first time they exported their own flying. Now **TAKEOFF, LANDING and AIR**, grouped *after* the block pair exactly as the on-screen table groups them and for the reason `Table.tsx` already gives — four times in one format, and the aircraft's logbook is filled from only one pair — with `OFF`/`ON` renamed **OFF BLOCK**/**ON BLOCK** now that they have neighbours to be confused with. `pdfmodel.AirTime` is pure, derived at render, **never stored** (§0.5), and mirrors `format.airMinutes` exactly: blank — never `0:00` — when either instant is missing or the interval is negative. Column widths **measured, not guessed** (`fpdf.GetStringWidth` over all 1296 rows at the real fonts: 166.3 mm of content, table now **279 mm of 285 mm printable**) and guarded by a new internal test watched red at 301 mm. **The EASA export is deliberately unchanged** — the form has no cell for an airborne time and its DEPARTURE/ARRIVAL TIME columns are block times; that document is what an authority reads. `make check` **86.8%**, `internal/pdfmodel` **100%**, 128 frontend tests. |
| 10 | **Edit / delete a flight** | **done** 2026-08-02 — owner ruled: app-entered flights only, real delete with an audit copy, double confirmation. `PUT`/`DELETE`/`GET /flights/{seq}`, `store.UpdateFlight`/`DeleteFlight`, the append-only `flight_audit` table, and the shared `FlightForm` behind both the new and edit pages. **83 frontend tests, backend 88.3%**, and driven live against a scratch server: edit, refusal on a paper row, delete, totals following, audit rows written. |

---

## 5. Decision Log

### 2026-09-06 — "Deploy should not need root", and what that actually costs

Asked immediately after Task 24 was pushed and could not be shipped: *"we should find some solution
for that, deploy should not need root."*

**The honest reframing came first.** `rami` is in the `sudo` group and always has been. So nothing
here could *grant* the account power it lacked — the ask was to remove an **interactive password**
from the deploy path, so that shipping does not require a human at a terminal. Every option below is
a way of spending privilege `rami` already holds; the question is only how narrowly, and how
auditably.

**Polkit was ruled out on evidence, not taste.** It is the textbook answer — let `rami` restart one
unit — and it does not work here: the box runs **Ubuntu 20.04 with polkit 0.105**, which reads
`.pkla` files and not JavaScript rules, and a `.pkla` cannot scope
`org.freedesktop.systemd1.manage-units` to a single unit. The tidy option would have handed `rami`
restart rights over **every service on a box that also runs seven other sites** (rule §0.3). Checked
with `pkaction --version` on the box before it was ruled out.

**A systemd `.path` drop-box was the other candidate** and was offered: rsync a binary plus a request
file naming its sha256, let a root unit notice and apply it, and the deploy contains no `sudo` at all.
The owner chose the scoped sudoers rule instead, and it is the better trade — the drop-box moves the
exit code into the journal, so a failed deploy looks exactly like a successful one from the terminal
that started it.

**What was built.**

- **`/etc/sudoers.d/logbook-deploy`** — `rami ALL=(root) NOPASSWD: /opt/logbook/logbook-apply ""`.
  The trailing `""` is sudo's syntax for *no arguments*; without it `NOPASSWD` on a path allows any
  arguments, and a script taking a file path would become a way to install arbitrary files as root.
  The target is **root-owned 0755**, so `rami` cannot rewrite what the grant points at — a `NOPASSWD`
  rule aimed at a user-writable script is passwordless root with extra steps.
- **`logbook-apply`** — the privileged half, and the only unattended-root command on the box. It
  refuses a missing or mismatched `SHA256SUMS`, and a symlinked artefact, **before** it stops the
  service. `deploy.sh` writes the checksums last, so their absence *is* the signal that an upload was
  truncated — and a truncated rsync leaves a binary that is executable, plausible and wrong.
- **Automatic rollback.** If the new binary does not answer its health check within fifteen seconds,
  `logbook-server.prev` goes back and the script exits non-zero; if the rollback is also unhealthy it
  says so in capitals. This is new, and it exists *because* the deploy is now unattended: nobody is
  reading the output, so "leave it broken and print a hint" stopped being an acceptable failure path.
- **`deploy.sh`** on the dev machine does everything else as an ordinary user — clean-tree guard
  (`app/backend/dist/` once held a DIRTY build that must never ship, and the `vcs.modified` stamp is
  permanent), `make check` + `npm run check`, both builds, the frontend rollback tar **before** the
  `--delete` rsync, a check that `dist/` really contains `sw.js` and `icons/`, **binary before
  frontend**, then off-box verification against the `vcs.revision` stamp.
- **`update.sh` is now a four-line shim** that execs `logbook-apply`. One implementation; the two
  paths cannot drift.

**It was rehearsed before it will ever run as root.** A fake tree with stubbed `systemctl`, `curl`
and `logger`, and the real script with only its paths rewritten: all four refusals exit **1** with
nothing installed and the service never stopped, plus the happy path (binary installed, `.prev` kept,
database copied first) and the rollback. The rehearsal also caught a fault in the *harness* — reading
`$?` after a pipe into `tail`, which reported every refusal as exit 0. That is precisely the failure
this whole design is trying to avoid, so it was worth measuring properly rather than eyeballing the
message.

**Also cleared:** `app/backend/dist/logbook-server`, the stale dirty 2026-08-01 artefact that three
APP.md revisions have warned must never be shipped. `deploy.sh` builds `dist/server` and stages it
under the deployed name, so the landmine had no remaining purpose.

**Outstanding: one root command**, which the owner asked for verbatim and which also carries the
Apache half of Task 22 in the same sitting.

### 2026-09-06 — The screen got the airborne times five weeks ago; the document never did

The owner, in full: *"There is a bug in the export (save as pdf) it only exports block times!"* and,
a moment later, *"not to and landing"*.

**They were right, and the gap was five weeks old.** Task 12 (2026-08-02) added **Takeoff, Landing
and Air** to the flights table on screen, and `pages/Table.tsx` carries a careful comment about how
those columns are laid out and why. `internal/pdfbook/table.go` was never opened. Its `tableColumns`
still read `{"OFF"}, {"ON"}` — so the application knew four times per flight, showed four on the
phone, and printed two in the PDF whose own description on the Export page promised *"everything the
application knows about it"*.

**Who it hit.** 36 rows of the live 1313 record an airborne pair: 19 transcribed from Book 3, and
**all 17 entered through the app**. Every flight the owner has logged themselves carries a takeoff
and a landing, which is precisely why they met this the first time they exported their own flying
rather than the paper history. The 1277 rows with no pair were unaffected and still print blank.

**What was built.** `TAKEOFF`, `LANDING` and `AIR`, plus `pdfmodel.AirTime`.

- **The airborne group sits after the block group, not interleaved with it.** Chronological order —
  OFF BLOCK, TAKEOFF, LANDING, ON BLOCK — was written first and then reversed, because `Table.tsx`
  had already ruled on exactly this question: *"it sits after the block group rather than
  interleaved, so the two pairs cannot be misread for one another at a glance — off/on block and
  takeoff/landing are four times in the same format, and the aircraft's logbook is filled from only
  one of the pairs."* The document must not disagree with the screen about a layout whose whole
  purpose is to stop a misreading. `OFF`/`ON` became **OFF BLOCK**/**ON BLOCK** in the same move.
- **`AIR` came along uninvited, and that was a decision.** The ask named takeoff and landing. The
  screen shows a third, derived column beside them, and a PDF of the same table that omits it is the
  same defect one column over. `AirTime` is pure, computed at render, **never stored** (§0.5), and
  mirrors `format.airMinutes` clause for clause — blank on a missing instant, blank on a negative
  interval, never `0:00`, because a zero claims the aeroplane never left the ground. The two
  implementations must never disagree about the same flight.
- **The EASA export was deliberately left alone.** The form has no cell for an airborne time and its
  DEPARTURE/ARRIVAL TIME columns are block times. That is the document an authority reads; adding
  columns to it to satisfy a complaint about a different document would be the wrong kind of helpful.

**The widths were measured, not guessed.** Three new columns had to fit a page already 261 mm wide
inside 285 mm of printable A4 landscape, and `fpdf` does not complain when a row runs off the paper —
it draws cells the printer silently cuts. So a throwaway program measured the widest real value in
every column with `GetStringWidth`, at the exact fonts the renderer uses, across all 1296 flights:
**166.3 mm of actual content** against 285 mm printable, i.e. 118.7 mm of slack that had never been
counted. Every column was resized from that measurement; the table is now **279 mm**, and
`TestTableColumnsFitBetweenTheMargins` guards it — **watched red at 301 mm** before being trusted,
because a guard nobody has seen fail proves nothing (§0.6).

**Tests, all red first.** `TestTableCarriesTakeoffAndLandingTimes` failed on all four of `TAKEOFF`,
`LANDING`, `15:20`, `16:28`; the `AIR` assertion asserts `1:08` specifically, which is the derived
15:20→16:28 figure and **not** the flight's 1:21 of block time, so the column cannot pass by echoing
`TOTAL`. `TestAirTimeIsDerivedAndBlankWhenItCannotBeKnown` covers the seven cases including a
midnight crossing and a landing before its takeoff — **and its own midnight case was wrong on the
first draft** (`+40 min` asserted as `0:30`), caught by reading it back before writing the function
it was meant to constrain. `make check`: **86.8%** overall, `internal/pdfmodel` **100%**.

**A thing the fix made visible, surfaced and not touched (§0.2).** With `AIR` printed next to
`TOTAL`, row `b3:412` (seq 1225, 08/09/2025) now openly shows `TOTAL 0:38` against a block time of
0:45 — its total is its **air** time. A sweep of all three CSVs says this is the **only** row in the
whole record where `Block_Time != Total_Time`, which matches `CLAUDE.md` §1 exactly: *"block time on
478 of Book 3's 479 rows"*. Nothing to fix; the export simply stopped hiding it.

**Not deployed.** It changes the backend, so it needs `sudo` on the box — the same gate as open
item 2. Verified only against the frozen CSVs on this machine, never against production.

### 2026-09-04 — "There is no Go toolchain" was false, and the owner caught it

**Yesterday's entry told the next session that a backend deploy from this machine was impossible.
That was wrong**, and the owner said so plainly: *"Then how have we built this app and done all
previous deploys. What you are saying makes no sense at all."* He was right — the whole backend, six
deploys and every coverage figure in this file were produced on this machine.

**Go is installed at `/home/havoc/.local/go/bin/go`** — go1.26.5, `GOPATH=/home/havoc/go`, a 795 MB
module cache sitting right there. It is simply **not on a non-interactive shell's `PATH`**: nothing
in `.bashrc` or `.profile` exports it, so a tool-run shell sees no `go` at all while the owner's
interactive terminal has it. Export it and everything works:

```bash
export PATH=$PATH:/home/havoc/.local/go/bin
make check          # 86.8% overall, 100% on every [core] package -- verified 2026-09-04
```

**How the wrong conclusion got made, because the reasoning error is more reusable than the fact.**
The probe was:

```bash
ls /usr/local/go/bin/ 2>/dev/null; ls ~/go/bin 2>/dev/null; which -a go 2>/dev/null; \
  find / -maxdepth 4 -name go -type f -perm -u+x 2>/dev/null
```

It printed **nothing at all**, and that was read as "Go is absent". Three faults, each on its own
enough:

1. **`-maxdepth 4` cannot reach `/home/havoc/.local/go/bin/go`**, which is at depth **6**. The search
   was structurally incapable of finding the thing it concluded was missing.
2. **`2>/dev/null` on every clause** meant a wrong path and a nonexistent path looked identical.
   Re-running with errors visible was what cracked it open in one command.
3. **Empty output was treated as evidence.** It is not. A probe that returns nothing has two
   readings — *the thing is absent*, or *the probe was bad* — and the second was never considered.
   `~/go` existed and held 795 MB of module cache, which was visible in that same output and
   contradicts "no Go here"; it was not noticed.

**The rule that follows: silence is not a finding.** Before writing "X does not exist" into this
file, the probe has to be shown capable of finding X — run it against something known present, or
drop the error suppression and read what it actually says. This one cost nothing because the backend
needed no deploy, but it was written into the **cold-start brief**, where its whole purpose was to
be believed by a session with no way to check.

**Also corrected:** the two places that carried the claim now describe the `PATH` gap instead, and
keep the half that was true — `app/backend/dist/` holds a **stale, dirty** build from `3921821`
(`vcs.modified=true`) that must never be shipped. Nothing about yesterday's deploy changes: it was
frontend-only, and remains correct and verified.

**State at handoff (end of 2026-09-04).** Re-verified from off-box after the correction: backend
live from `6aed062` clean, frontend `index-D3Tqt5-U.js`, service active with **`NRestarts=0`**,
health 200, all eight private routes 401, the owner's seven other sites 200, disk 57% used, memory
312 MB of 1971. **The daily off-box backup is alive** — timer active, last run 2026-09-04 03:22 UTC
**exit 0**, next 2026-09-05 03:23 UTC. `origin/master` at `b153507`, working tree clean, nothing
unpushed; the two commits ahead of the deployed bundle are **documentation only**. The one number
that could not be read is the **live flight count** — it needs a session cookie or `journalctl`, and
`rami` has neither without sudo.

**Open work is down to two items, and neither can be done from a session**: the Apache half of Task
22 (needs `sudo`) and confirming the sudo-password rotation. Everything else on the board is live.

### 2026-09-03 — Told to deploy, and the deploy was already done

**The owner said "you can deploy". There was nothing to deploy but one CSS fix**, and finding that
out was the whole of the work.

**The brief lied for three weeks.** `APP.md`'s NEXT-SESSION block opened with *"TASKS 19, 20 AND 21
ARE BUILT, TESTED, PUSHED — AND NOT DEPLOYED. The box is behind the repo for the first time since
2026-08-02"*, and named production as the 2026-08-02 build. Every word of that was false on
2026-08-14, when those three went live. The session that shipped them recorded **only the frontend
half** (Task 22, the caching work it was actually thinking about) and left the brief describing a
world that had stopped existing eight minutes earlier. Three task-board rows said "not yet deployed"
about live code.

This is a **rule §0.1 failure**, and worth naming precisely because §0.1 is the rule this project
takes most seriously: the repo is the single source of truth, so when it is wrong there is no second
place to look. A fresh session on a clean clone would have reconstructed exactly the wrong world —
and nearly did.

**What actually settled it was the box, not the file.** Four cheap questions, in this order:

| question | command | answer |
|---|---|---|
| does the route the new bundle needs exist? | `curl -o /dev/null -w %{http_code} …/api/pilots` | **401**, not 404 |
| is 401 just the auth catch-all? | same, on `…/api/definitely-not-a-route-xyz` | **404** — so 401 means *the route is real* |
| which commit is the live binary? | `strings -a /opt/logbook/logbook-server \| grep vcs.revision` | **`6aed062`, `vcs.modified=false`** |
| what is actually behind? | `git log 6aed062..HEAD -- app/backend` | **nothing** |

**The 401-vs-404 distinction is the reusable trick.** Under default deny almost every route answers
401 unauthenticated, which looks identical to a route that is missing — until you ask an obviously
fake path and get a 404. That one control probe turns the whole route table into a readable
inventory without a session.

**`vcs.revision` is the honest answer to "what is running".** Go stamps the commit into every binary
built in a repo. md5 had been useless here: the live binary and the local `dist/server` were the
**same size** and **different hashes**, because they differ only in that 40-character stamp and a
build ID. The size match said "same source"; the stamp said which commit, and that the live one was
built **clean** while the local one was **dirty** (`vcs.modified=true`, from `3921821`). A stale
dirty artefact sitting in `dist/` is exactly the thing a hurried session ships by accident.

**What was left to do, and was done.** Task 23 only — frontend, one commit. `npm ci`, `npm run
check` (128 green), a clean `npm run build`, a tar of the live directory into `/home/rami/` for
rollback, then `rsync -a --delete`. `--delete` was checked against the live listing first: `dist/`
produces all five entries the web root holds, including **`sw.js`**, which must keep being deployed
because it is the kill switch that retires workers still installed on devices. Verified after: the
bundle **fetched back over HTTPS and md5-matched** the repo build byte for byte, the paired selector
is present in the served CSS and zero offending rules survive, all eight private routes still answer
401, the other seven sites answer 200, and the service never restarted (**`NRestarts=0`** — a
frontend deploy does not touch it).

**Two constraints on this machine, written down because they are invisible and will cost the next
session an hour:**

1. **`go` is not on a non-interactive shell's `PATH`** — it is at `/home/havoc/.local/go/bin/go`,
   and exporting that makes everything work. ⚠ **This was first written down as "there is no Go
   toolchain here" and "a backend deploy from this machine is currently impossible", both false**;
   see the 2026-09-04 entry. The half that was true and is worth keeping: the only prebuilt binaries
   in `app/backend/dist/` are the **stale dirty** ones described above, and they must never be
   shipped.
2. **`sudo` needs a password**, so nothing requiring root can be done from a session. That is
   correct and should stay — but it means the Apache half of Task 22 cannot be finished without the
   owner at a terminal.

**The one thing still genuinely undone: Apache.** Re-verified live today — `/logbook/assets/*`
answers `public, max-age=31536000, immutable` **with an `ETag`**. The owner's *"make sure NOTHING is
cached at all"* has been **half-honoured since 2026-08-14**: the app registers no service worker,
but the server still tells browsers to hold an asset for a year. It is not urgent, and the reason is
worth stating so nobody panics about it: asset filenames are content-hashed and `index.html` is
`no-store`, so a new deploy is always fetched — which is precisely why this survived three weeks
unnoticed. It is still the repo saying one thing and the box doing another, and that is the failure
mode this whole day was about.

### 2026-08-18 — The seventh thing a green suite loved, and it was a comma

**The owner opened the app on a phone and sent a screenshot.** The aircraft dropdown read
`OH-CTLC172287 flights · 2026-08-14` — registration, type and flight count welded into one word —
and the list was laid out in **two columns** instead of one. 127 tests were green. This is the
**seventh** time on this project that real use has found what the suite loved, and the
NEXT-SESSION block had named this exact control as the risk.

**One character in the wrong place, six rules deep.** Task 21 folded the pilot picker into the
aircraft picker's CSS by adding a second selector to each rule. On six of them the comma landed one
word too early:

```css
.aircraftpicker .options,          /* the LIST */
.pilotpicker .options button { display: grid; grid-template-columns: auto auto 1fr; … }
```

So the **dropdown itself** became a three-column grid — that is the two-column list in the
screenshot — and the **options** inside it got no `display`, no `gap`, no columns, which is why
three spans ran together with nothing between them. The same slip put `justify-self`, the hover
background and, in the 22rem media query, **`display: none` on the entire aircraft list**. Nobody
had a 352px screen, so that one was never going to be found by looking.

**Why no test caught it, and what now does.** Nothing in a jsdom suite renders CSS, so a selector
can address the wrong element forever and every assertion still passes. **`src/styles.test.ts`**
now parses `styles.css` and asserts the shape rather than the rendering — the same tactic
`noworker.test.ts` uses on `main.tsx`:

1. when a rule names **both** pickers, the two selectors must end in the **same suffix**;
2. no rule may give the bare `.aircraftpicker .options` a `display: grid` or `display: none`;
3. each picker's `.options button` must actually be the grid.

All three were watched red. This is a **narrow guard on the one mistake this stylesheet has
actually made**, not a general CSS linter — it is worth having precisely because it is invisible to
every other kind of test here.

**Owner ruling in the same breath: "It's enough to just show the registration."** The option now
carries the registration and nothing else. The type and the `N flights · date` tail are gone from
the row — three things competing for one line on a 390px screen is what a pilot reads *past* to
find the aeroplane. **The type still filters**: typing `C172` or `P28` narrows the list exactly as
before, it just no longer takes up space. The picker's test asserts the option text by **exact
equality**, because the point of the ruling is what is *not* there.

Consequence, stated rather than hidden: with the type invisible, a type-filtered list gives no
visible reason why a row matched. The registration is what identifies the aeroplane, and the owner
knows which of his aeroplanes is a C172.

**Not verified in a browser.** The dev path needs `./dist/server createuser`, which needs a real
terminal, and the Chrome extension could not reach a local static server either. So this rests on
the tests and on the CSS now being provably paired — **it should be looked at on the phone at the
next deploy**, along with everything else from 2026-08-03 that never has been.

### 2026-08-03 — Three priorities from the owner, and two rulings that made them smaller

The owner named three things and said that after them "we should be solid": finish adding aircraft
(easily from the new-flight page, **and** from a management page of its own), fix a bug where the
Devices list shows sessions that are plainly dead, and stop the PIC name being free text — *"I could
have a typo when I write `self`, it could be `sself` or `SELF` or `seeelf`, and I need it to be
consistent (like the aircraft regs)"*. They are Tasks 19, 20 and 21.

**The session bug is not what it looks like, and naming the mechanism changed the fix.** The
symptom is a Devices list full of entries for what is really one phone. The cause is that two
lifetimes disagree: `setCookie` (`handlers.go:891`) deliberately sets **no `Max-Age`**, so the
cookie is a *browser-session* cookie that dies when the phone's browser or PWA restarts — that is
the re-login the owner sees and calls correct — while the row keeps `SessionLifetime = 90 * 24h` of
**idle** life (`store/auth.go:27`). The device can never present that token again, so the row is
**orphaned: unreachable, but listed and alive for another three months.** Every login makes one
more. The backup manifest showing `sessions dropped 13` for a single-user app is the same fact seen
from the other side. Nothing was broken in `PurgeExpiredSessions` or in the hourly sweep that calls
it (`main.go:184`) — they work exactly as written; they were simply never going to fire on a row
whose 90 days had not elapsed.

Offered the choice between making the cookie persistent (stay logged in, list becomes truthful) and
shortening the server's window, the owner **ruled for the shorter window: keep logging in after a
restart, drop 90 days to 14.** So the fix is to make the server's idea of an idle session roughly
match the browser's, not to extend the credential. Worth recording that the 90 days was itself an
earlier owner requirement — this supersedes it.

**The PIC roster is smaller than asked, by ruling.** The first plan was a `people` roster feeding
both a PIC picker and a new Student field. The owner cut the second half — *"my bad, no student
field as it should be"* — so Task 21 is the PIC name only, and students stay in Remarks where they
are today. No new column on `flights`.

**A census of the frozen data says the owner's worry is already real.** The 1296 historical rows
carry **18 distinct `pic_name` values** plus one blank cell: `self` ×**1143**, then `Martevuo` ×54,
`Autere` ×30, `Stude` ×18, `Jansson` ×16 and a tail. *(Counted off the CSV files first, which gave
`self` ×1145 — wrong by exactly the two cumulative seed rows that open Books 2 and 3, both of which
carry `self` and neither of which is imported. The store test caught it. **Count the record, not the
files.**)* Two of them are worth stating plainly, and **neither is being
touched**: **`Sinervä` ×6 and `Sinerva` ×1 are almost certainly one person**, and **`Stude` ×18
looks like a word, not a surname.** Both are exactly the class of thing rule 0.8 puts out of reach —
surfaced here, ruled on by nobody, changed by nothing. The roster will therefore be **derived from
these values as they are written**, variants and all, the same way the aircraft list was derived
before it gained hand-added rows. Task 21 stops the *next* spelling of `self` from being invented;
it does not tidy the paper.

### 2026-08-14 — "Make sure NOTHING is cached", and the stale build that was not cached at all

The owner's phone was showing the old app after a deploy, and they asked for the nuclear option:
*"Can you make sure NOTHING is cached at all? Like the browser needs to forget (except the cookie for
the session)."*

**The first thing to say is that the phone was innocent.** The backend had gone out — they had run
`update.sh` themselves — but **the frontend had never been uploaded**: `curl` showed
`index-xgdC8L2o.js` from 2 August still live, because the session had staged the bundle and was
waiting on the backend step before rsyncing it. The phone was faithfully serving what the server was
handing it. **This is the trap in `docs/deploy.md` read backwards** — that note warns "if the feature
is missing on the phone, check the device, not the deploy", and the honest version is: **check which
one it is before touching either.** Thirty seconds of `curl -sI` answered it; a session that had
started writing cache-busting code would have shipped a fix for a bug that did not exist, and the
frontend would *still* have been missing.

**Then the instruction was carried out anyway, because it was right on its own terms.** Two
incidents in a fortnight had been "is my phone stale?", and each one cost more than the cache ever
saved. What the shell cache actually bought was small: offline **writes** were never in scope (§2),
so an offline shell could only open an app that then failed every request it made. So:

- **`public/sw.js` is now a kill switch** — deletes every cache by name (not a filter on the names
  this project used), unregisters itself, then navigates the open page onto the network.
- **⛔ It must keep being deployed.** Deleting `sw.js` from the server does **not** remove a worker
  already on a device; the browser keeps running the copy it has, serving the shell it cached,
  forever. A new worker at the same URL is the *only* reliable way to retire one. There is no way to
  know when the last device has been cleaned, so the file stays indefinitely.
- **Nothing registers it.** `main.tsx` no longer calls `register()`, which is also what stops a
  register → unregister → navigate → register loop on the owner's phone. `noworker.test.ts` fails if
  the line comes back.
- **Apache serves the whole directory `no-store`, with no `ETag`** — one rule replacing three. The
  old arrangement was *correct per file* and still produced two incidents, because every exception is
  a place a future deploy can be stale in a way nobody thinks to look. A single rule cannot have that
  shape of bug. `ETag` goes because a 304 is the server saying "use your copy".

**Both costs are written down rather than buried**: the app no longer opens without a network, and
~200 KB is fetched on every cold start. If the second ever matters, the honest middle is `no-cache`
on `/logbook/assets` alone — content-hashed filenames cannot be stale — and that is a knowing trade
for a later session, not a quiet cleanup. **The session cookie is not affected by any of it**: it is
`HttpOnly` and lives in the cookie store, which is not a cache.

### 2026-08-03 — Running it found the sixth thing a green suite loved

The tally in "How to run things" said five. It is six.

Every test passed — 127 frontend, backend at 86.8% with the core at 100% — and then the three tasks
were driven against a **scratch server holding the real 1296 flights**. The roster came back correct
(18 names, `self` ×1143, ordered most-recent-first), the fleet rename worked, `DELETE /aircraft`
answered 405. And the `pic_name` refusal came back like this:

```
{"error":"...","fields":{"pic_name":"\"SELF\" is not in the pilot list ..."}}
```

**A map. Every other refusal from `POST`/`PUT /flights` answers with an array** of `{field, message}`
— that is Go's `entry.Errors` — and the form reads `err.fields.length` and iterates it. On a map,
`length` is `undefined`, so the whole per-field mechanism silently fell through: the pilot would have
seen a banner reading *"see the fields below"* with **nothing marked below**, and the form's
jump-to-the-first-bad-control would have had nothing to jump to. Backend tests passed because they
asserted the status code and that the body mentioned `pic_name`. Both were true. **The refusal was
unrenderable and every test agreed it was fine.**

Fixed in two places, because it is two bugs:

1. `refusePICName` now builds an **`entry.Errors` value** instead of hand-rolling the JSON, so the
   shape cannot drift from its siblings again, with a test that decodes into the array shape.
2. `api.ts` now **normalises both shapes** into `FieldError[]`. The map was not a mistake — the
   *aircraft* endpoints have answered that way since 2026-08-02 and there was no reason for them to
   change — but every page above the fetch layer was written expecting an array. `Fleet.tsx`, written
   this session, calls `.map` on it: a `TypeError` thrown **inside a catch block**, which would have
   swallowed the refusal and left the form apparently doing nothing. That one had never run, because
   nothing had ever refused it in a browser.

**The lesson is not "test harder", it is the one already in the runbook.** The shape of a refusal is
a contract between two programs, and a test that checks a status code and a substring is not reading
the contract. Thirty seconds of `curl` against a real server read it immediately.

### 2026-08-03 — Three tasks built, and the two places the work was bigger than the ask

**Task 19, the fleet page**, was the one already scoped: `PUT /aircraft` had been live and deployed
for a day with **zero callers**, so an aeroplane added at an airfield with a typo could never be
corrected and, by ruling, never deleted. It is a page now, at `/logbook/fleet`, reached from the
Aircraft tab rather than a seventh tab — six already share a 390px phone and "Statistics" was cut to
"Stats" to fit the sixth. The no-delete ruling is now guarded on **both** sides of the wire: the
backend's route-table test, and a frontend test that fails if any control anywhere says delete.

**Task 20 was a smaller change than the bug looked, once the mechanism was named** — and then a
bigger one than the ruling implied. Shortening `SessionLifetime` to 14 days fixes sessions created
from now on. It does **nothing** for the thirteen already in production, because each row carries an
`expires_at` stamped when it was written, up to three months out. So the window is now **computed
from `last_used_at` against the constant** and the stored column decides nothing: one authority, and
shortening it retires the rows that already exist. `PurgeExpiredSessions` asks the same question, and
the Devices page reports the computed date — otherwise it would promise access the server refuses.
**A fix that only applies to data created after the deploy is not a fix for a bug about old data.**

**Task 21 needed a decision the ask did not contain: where the guarantee lives.** A picker makes the
right thing easy; it does not make the wrong thing impossible. Typing `SELF` and submitting without
choosing would still have written `SELF`, which is exactly the outcome the owner asked to prevent. So
`POST`/`PUT /flights` now **refuse a `pic_name` that is not on the roster exactly**, naming the field.
Three properties make that safe rather than obstructive: the roster is *derived from the flights*, so
every name in the record is on it and **no existing flight can become uneditable**; a blank name stays
legal, because one transcribed row has a blank PIC cell; and the refusal **surfaces** rather than
silently re-spelling what was typed into the roster's version (rule 0.2).

The roster deliberately has **no update and no delete** — smaller than the aircraft surface. Renaming
an entry could not rename the flights that carry the name, so the two would just disagree while the
derived half went on reporting the old spelling. A wrong name is corrected where it lives, on the
flight. That leaves one rough edge, stated rather than hidden: **a hand-added name that was never
flown with cannot be removed**, and will sit at the top of the list as never-used. If that becomes
annoying it is a five-line delete scoped to `user_added` rows with no flights — but it is the owner's
call, not a session's.

**And the census in the entry above was wrong by two, which a test caught.** `self` is on **1143**
flights, not 1145. The first count was taken off the CSV files, which contain 1298 rows for 1296
flights: Books 2 and 3 each open with the previous book's final row as a cumulative seed, both of
those rows say `self`, and neither is imported. `TestTheRealBooksNameEighteenPeople` runs against the
store and reported 1143 against an expected 1145. **Count the record, not the files** — the same
distinction the whole 1296-vs-1298 architecture rests on, and it caught a session that knew that and
still reached for the CSVs because they were easier to grep.

### 2026-08-02 — The first deploy that did not write to the record

Tasks 16, 17 and 18 went to production. The mechanics were unremarkable — stage, md5, binary, verify,
restart, frontend, probe — which is itself the point: **this is the first deploy in the project's
history whose step 4 wrote nothing.** Every previous one ran `DELETE FROM flights WHERE source_book
<> 0` against a live legal record in order to reproduce rows that could not have changed. Today step 4
read the database, compared it to the frozen CSVs on nine checksums, printed them, and stopped.

**The check that used to be a side effect is now the deliverable.** `verify` matched on all nine —
1296 / 1222:10 / 1054:45 / 167:25 / 107:58 / 22:45 / 189:41 / 407:39 / 3444 — and because it is scoped
`source_book <> 0`, that is a positive statement about *the 1296 frozen historical rows specifically*:
they are byte-for-byte what the books say, three months of app use later. A re-import could never have
told us that. It would have overwritten the evidence and then reported success.

Meanwhile the startup line read **`flights=1298`**. Those two numbers coexisting — 1296 verified,
1298 served — is the whole architecture in one screen, and a session that "reconciles" them breaks it.

**Two smaller rulings, both about not doing things.**

*Apache was not touched.* The runbook's step 4 is `install-apache.sh`, and it was skipped: all three
cache layers were already correct live, and `apache-logbook.conf` in the repo was byte-identical to
the installed block. Running it would have been a no-op reload of a webserver fronting seven other
sites. §0.3 says changes to Apache are additive, reversible and verified — the cheapest way to satisfy
that is to establish there is no change to make. **Verifying that a step is unnecessary is part of
running the runbook, not a shortcut past it**, and the md5 comparison is what makes the difference
between the two.

*The 403 was not "fixed".* `POST /aircraft` and `PUT /aircraft/{reg}` answer **403**, not the 401 the
probe expected. That is `checkOrigin` wrapping *outside* the auth check: a state-changing request with
no `Origin` header is refused as CSRF before authentication is considered. It is stricter than 401 and
it is asserted in `flightentry_test.go:206`. Worth writing down because an unexpected status code on a
security route is exactly the shape of thing a later session "corrects" into a weaker one. **`DELETE`
answers 405** — the no-delete ruling of the previous session is enforced by the live route table, not
only by a test.

**And the cost, which is not technical.** The owner handed the sudo password over in-session so these
steps could run without them. That was their call and they attached a condition to it — rotate as soon
as the deploy is done. The deploy is done. The credential is now in two transcripts, and until it is
rotated it is the largest open exposure in the project, ahead of everything in `docs/security.md`.
**Convenience during a deploy is exactly when this rule gets bent**, which is why it is written down
here rather than only in the security doc: the next session should find it at the top of the pick-up
list, not buried.

### 2026-08-02 — Retiring the importer, and aircraft becoming records

Two owner rulings in one exchange, and they turn out to be the same decision seen from two sides.

**The ask was aircraft CRUD**: *"if I fly new aircraft that I have never flown yet, there needs to be
CRUD for aircraft."* The gap was real and worse than it looked. `GET /aircraft` was the only route,
and the list behind it was **purely derived** — rebuilt from the flights on every import — so the
only aeroplanes that could exist were the ones already flown. The form's registration was a
`<select>` fed by that list, which made **the first flight in a new aeroplane unenterable**. The API
would have taken it (`entry` only requires reg and type to be non-empty); the UI could not say it.

And the trap `source_book = 0` solves for flights was wide open next door: `store.Import` ran an
**unqualified `DELETE FROM aircraft`**. The moment aircraft got a write path, every hand-added
aeroplane was one import away from deletion. Scoped to `user_added = 0` now, with a test that
re-imports and looks for the row.

**What the owner ruled, and what it cost the design.** Asked whether imported aircraft should be
editable, the answer redirected the question: *"The retire note is totally obsolete for many
aircraft. You are assuming that some aircraft is retired because I haven't flown it in a year… I say
we drop this retired thing completely."* So:

- **No retired/active concept**, at all. The `active` column stays in the schema as **vestigial and
  documented as such** — dropping a column from a live legal-record schema to delete a feature is the
  worse trade — and nothing reads it.
- **No delete route.** An aeroplane once added stays; a wrong one is corrected with a `PUT`. Asserted
  by a test that inspects the route table, because a later session adding one "for symmetry" is
  exactly how a ruling gets lost.
- **What replaces both is ordering and filtering.** The list is ordered never-flown first (you added
  it *because* you are about to fly it), then most recently flown, and the picker filters as you type
  — on registration **and** type, since typing `C172` is as natural as typing `OH-`. Nothing is ever
  hidden. That ordering lives in the server's SQL and is deliberately not re-sorted in the component:
  two authorities for it would disagree.

**Editing is allowed on all 38 aeroplanes from the books, and that is not a hole in rule 0.8.** This
table seeds a form; it is not the record. Every flight carries its own registration, type and class
denormalized exactly as written on paper, so no edit here can move one minute — asserted twice, once
in the store by reading every flight back field by field and once at the API by diffing the whole
`GET /flights` body. It is also the shape of the owner's own observation in the same message:
**SE-GKT and OH-GKT are one airframe whose registration changed.** Recorded, not acted on — closing
it would mean touching frozen data.

**Then the larger ruling, which the owner raised and asked for an opinion on**: *"we should start
treating the production database now as the source of truth. We don't need the importer anymore… Do
you?"* **Yes — with one amendment.**

The importer had become pure risk. It re-imported *frozen* data on every deploy, so its only possible
outcomes were "reproduces exactly what is already there" or "something went wrong" — and to achieve
the first it ran `DELETE FROM flights WHERE source_book <> 0` against a live legal record. **A
recurring destructive operation whose best case is a no-op.** It had already nearly bitten: the
stale-CSV incident was one root command from writing `C192` back into production. Retiring it deletes
that entire class of failure. And the argument it used to rest on — "production is rebuildable from
the repo in one command" — **had been false since the owner logged two flights in the app**.

**The amendment: stop importing, do not stop verifying.** `verify` is read-only and checks nine
checksums against the CSVs. Kept in `update.sh`, it stops being "rebuild and hope" and becomes a
**drift and tamper check on the 1296 frozen historical rows** — which is what rule 0.2 actually wants
— without ever writing. The CSVs stay in the repo as the provenance record behind `drift.md`; they
are simply no longer loaded.

So the shape is now: **production database = source of truth · off-box backup = what protects it ·
CSVs = frozen provenance plus a read-only checksum.** That the backup can carry that weight is not an
assumption — it was cloned, restored, booted and reconciled earlier the same day.

The two rulings meet here: aircraft could only become editable records *because* the importer stopped
rebuilding them. Had the import stayed in the deploy, every aircraft edit would have been silently
reverted on the next one.

### 2026-08-02 — The restore drill: the backup was fine, its instructions were not runnable

The brief's top open item was *"read `RESTORE.md` from a real clone, once, while there is no
emergency — a backup nobody has restored from is still a backup nobody should trust."* Done, in full,
and the conclusion splits cleanly in two.

**The backup itself passed everything.** Cloned from the private remote, and:

- both files hash to exactly what `MANIFEST.txt` claims;
- `logbook.db` is **byte-identical across all three snapshots** taken that day — only the manifest's
  timestamp moves, which is the determinism claim proven rather than asserted, and is why git stores
  one blob;
- the database **boots**: the server started against the restored file and logged `flights=1298`,
  `/health` answered 200 and all six private routes answered 401, so default deny survives a restore;
- `logbook.csv` **independently reconciles** to the same figures with no SQLite involved at all —
  1298 flights, 1223:03 (73383 minutes), 3446 landings, 38 aircraft — carries `off_block_raw` and
  `time_origin` on every row (rule 0.4), has **no `Cumulative_*` columns** (rule 0.5), and both
  hand-entered flights are in it with their raw times. The deliberate redundancy is real redundancy.

**And its instructions could not be followed.** Step 3 — the step `RESTORE.md` calls mandatory, the
one rule 0.2 hangs on — told the reader to run `sqlite3`. **That binary is not installed on the
production box** (checked) **and is not a dependency of this project**: the whole point of
`modernc.org/sqlite` is that nothing outside the Go binary speaks SQLite. On a fresh server that line
is `command not found`, at the exact moment someone is deciding whether to let an application write
to a legal record. The choice it forces is "skip the verification" or "apt-install a database package
mid-emergency".

**This is the same species as the `GIT_SSH_COMMAND` preflight**, and the pattern is now three for
three on this project: *a verification step that cannot do the thing it claims to do*. That one could
not fail for the reason it named; this one could not run at all. Both read as protection. Both had
sat there since the day they were written, because **verification code gets no exercise on the happy
path** — and a backup's restore instructions are the most extreme case of that, since the happy path
is "never restore".

Also found, and the reason the fix is bigger than a one-line edit: `RESTORE.md` step 1 says *install
the app as `deploy.md` describes*, and step 3 then runs `/opt/logbook/logbookctl` — but
**`install-backend.sh` only ever installed `logbook-server`.** `logbookctl` reached the box solely as
a side effect of `install-backup.sh`. A restore performed exactly as documented arrived at step 3
without the tool. Fixing only the sqlite3 line would have replaced a command that does not exist with
a different command that does not exist.

**What was built.** `logbookctl check -db <db> [-manifest MANIFEST.txt]` — reads the figures out of a
restored database and compares them to the manifest, printing every one, exiting non-zero on any
disagreement, naming which figure and both values. It needs **nothing but the database file**: no
CSVs, no `sqlite3`, no network. It hashes the file **before** anything opens it. `Figures` is the
same code path `Run` uses to write the manifest, deliberately — a checker that computed its numbers
differently from the writer would drift, and the day it drifted would be the day of a restore.
`install-backend.sh` now installs `logbookctl`.

**`verify` is not this check, and the difference is load-bearing.** Verify compares against the three
CSVs and is scoped `source_book <> 0`. It would pass with a satisfied green message while **every
app-entered flight was missing** — the only rows in the file that exist nowhere else, and the entire
reason Task 14 exists. `RESTORE.md` now says so in as many words, because reaching for the
familiar-looking command is the obvious mistake.

Proven against the real artefact, not only in tests: `check` on the actual restored production
backup matches all seven figures and the sha256; pointed at a stale scratch database it refuses and
reports `flights 1298 vs 1296`, `hand-entered 2 vs 0`, `discrepancies 54 vs 61` — which is exactly
what "restored from the wrong day" looks like. The test that forbids the defect was confirmed to fire
against the **shipped** `RESTORE.md` before the fix went in.

**Fifth time on this project that running something found what reading it did not** — after the
untypeable colon, the stale service worker, the off-screen save confirmation, and the empty clone.

### 2026-08-02 — Installing the backup: three bugs in the checks, none in the thing being checked

The owner's summary of this task was *"You need to run git push periodically on a repo, what is so
complicated about that?"* — and it was the correct question. `backup.sh` is `git add`, `git commit
--allow-empty`, `git push`, and it worked on the first attempt and every attempt after. **Three
consecutive installs failed anyway, and all three failures were in the code that verifies, not in the
code that backs up.** That is worth recording precisely, because the verification exists on purpose
and the lesson is not "verify less".

**Bug 1 — the preflight never tested the key it was gatekeeping.** One helper set
`GIT_SSH_COMMAND` and ran the command as the `logbook` user. That is right for git, which spawns the
string itself. Step 5 then called `asuser ssh -T git@github.com`, and **a bare `ssh` ignores
`GIT_SSH_COMMAND` completely** — so it ran plain `ssh` with no `-i`, looked for a default identity at
`~/.ssh/id_*`, found none, and reported `Permission denied (publickey)`. It failed **identically
whether the key was perfect or absent**, and it accused a GitHub configuration that was correct while
`git ls-remote` was authenticating on that same key one step earlier. Two sessions' worth of
debugging went to GitHub's side of a problem entirely on ours. The options are now one array used two
ways — interpolated for git, passed as real argv by `asuser_ssh` for ssh — because the two forms are
not interchangeable and looked it.

**Bug 2 — a status display was allowed to decide control flow.** `systemctl status` **exits 3** for a
oneshot that has finished; "inactive (dead)" is its success state. Under `set -euo pipefail` that
killed the script at step 7 — *after* the snapshot, the commit and a successful push — so the owner
saw a screen full of success and got no timer, twice. Step 7 now displays status with `|| true` and
asks the question that matters separately, of `systemctl show -p Result`: a property with a defined
value rather than a page meant for humans.

**Bug 3 — an empty remote was reported as an authentication failure.** Branch discovery treated "no
output" as "could not reach the remote", but a brand-new empty repository *has* no default branch:
`ls-remote` succeeds and prints nothing. That is the normal first run, and the message pointed the
reader at a key problem one line above a step 5 that authenticated perfectly. It now branches on
ls-remote's **exit status**, and distinguishes reachable-and-empty from unreachable.

Bugs 2 and 3 are the same species as the empty-clone bug this script was written to prevent: **a step
reporting something it had not established.** Bug 1 is worse and worth naming on its own — *a check
that cannot fail for the reason it claims to test is not a weak check, it is a misleading one.* It
would have kept passing forever once the key worked, and nobody would have learned it was inert.

**What the owner ruled, and it is not a deploy key.** GitHub greets a deploy key as `Hi owner/repo!`
and an account key as `Hi owner!`, so the kind of key in use is otherwise invisible. The key here is
**account-level on the dedicated `ramiayoub-priv` account**, and the owner ruled that deliberately
after it was raised: the scoping a deploy key buys is already provided by an account that exists for
this and holds nothing else. Step 5 now *reports* which kind authenticated, as a fact rather than a
warning. The script's header had instructed otherwise, and the ruling wins. **Do not "fix" this
back.**

**Bug 4, found after the other three were fixed, and it is the one that best makes the point.** With
steps 1–7 finally all correct, step 8 died on `could not create work tree dir: Permission denied` —
`mktemp -d` runs as **root** and makes a `0700 root:root` directory, and the clone runs as
**logbook**, the only account holding the key. One `chown`. But note *when* it surfaced: this check
had been sitting downstream of a crash since the day it was written, so it had **never executed
once**. Three earlier bugs had to be fixed before the fourth could even be reached.

That is the real shape of the session. The backup worked from the first attempt and needed no
changes. Four separate defects sat in the scaffolding around it, in a strict chain, each one hidden
behind the last. **Verification code gets no exercise on the happy path, which is precisely when it
is written and precisely why it is wrong.** Run it, in the failing configuration, before trusting it.

✅ **Discharged the same day**: step 8 passed at 17:20 UTC — four files out of a fresh clone, and
`logbook.db` still hashing to what its own manifest claims. The backup is now proven to come back,
not merely to go out.

### 2026-08-02 — Staging the deploy, and the stale CSVs that were one root command from undoing a ruling

Runbook step 1 needs no sudo, so a session can do it, and doing it is what turns the owner's
remaining work into two commands instead of a procedure. Three things came out of it that are worth
keeping.

**The staged CSVs were stale, and nothing would have announced it.** The box was still holding the
Aug-1 20:27 files — from *before* the five aircraft-type cells were ruled on. `update.sh` re-imports
from `$STAGE/csv`, so running it would have written `C192` back into the production database and
reported **61 discrepancies**, and the only signal would have been a number in the middle of a long
root-command transcript that a tired reader would have to notice and know was wrong. The binary gets
all the attention in a deploy because it is the thing that obviously changed; **the data staged
beside it is what actually reaches the legal record.** Both halves are now md5-checked, and the
runbook says explicitly that `csv/` is the copy that matters.

**A dry run on the box is evidence; a green suite at home is not.** `logbookctl import -dry-run`
writes nothing, so it can be run as `rami` against the exact binary and the exact CSVs that the root
command will use. That is a different claim from `make check` passing locally: it tests the artefacts
*as staged*, after the cross-compile and the transfer, on the machine they will run on. It returned
1296 / 1222:10 / 54, with `unknown_aircraft_type` and `type_conflict` absent. Rule §0.6 asks for a
statement of why a deploy will not break **before** it runs, and this is the strongest form of it
available without root.

**The pairing rule got a sharper instance, and it now points in a specific direction.** The existing
warning is that the binary and frontend must land *together*; this deploy shows that the order within
"together" is not free. The Aircraft tab calls `GET /aircraft-time`, a route the live binary does not
have — so a frontend shipped first is a tab that 404s on the owner's phone. **Binary first, then
frontend**, and the frontend was built and verified but deliberately left unsent for that reason.
Checking the bundle's *contents* (rather than trusting that a build contains what was committed) is
the habit the stale-service-worker episode paid for, and it is cheap enough to keep doing.

### 2026-08-02 — Task 14: the backup, and the bug that reports success forever

Asked for by the owner: a daily copy of the data pushed to a private repository, restorable on its
own. The reason is sharper than "backups are good". Until today production was reconstructible from
this repository — three committed CSVs and one command. Now the app is the only way the record
grows, and **flights entered in it exist in no CSV**. The pre-import backups under
`/var/lib/logbook/backups/` sit on the same disk as the database they protect: they defend against a
bad import and against nothing else.

**The snapshot is `logbookctl backup`, not a shell script with `sqlite3` in it.** It is Go, so the
verification is the same code the import already trusts: `VACUUM INTO` (transactionally consistent
against a live database in WAL mode — the service never stops), then `PRAGMA integrity_check`, then
a row count against the **live** database, and it **refuses and writes nothing** on a mismatch. It
builds everything in a staging directory and moves it into place only once all of it is proven, so
an interrupted or refused run leaves yesterday's good backup exactly where it was.

**Four files, and two of them are deliberate redundancy.** `logbook.db` is the restore. `logbook.csv`
exists for the day something is wrong with the database file, or with this program, or with SQLite —
a legal record whose only readable form needs a specific binary to still work has a single point of
failure. The CSV carries the provenance columns and the raw times as written on paper (rule §0.4),
not a summary, and **no `Cumulative_*` columns** (rule §0.5): writing running totals into the backup
would put this project's largest historical source of drift back into the data.

**`RESTORE.md` travels *with* the data.** Instructions that live only in the application repository
are instructions you do not have on the day the server is gone.

**Sessions are stripped, users are kept, and that second half is the real decision.** Sessions are
the nearest thing in the schema to a live credential and restore nothing. Users must survive or the
restored logbook cannot be opened — which means the Argon2id hash leaves the box. That is written
down in `docs/security.md` rather than assumed, together with its consequence: **if that repository
ever becomes public, the logbook password is compromised.**

**It commits every day, even when nothing was flown.** Unchanged data produces byte-identical files
— verified — so git stores one blob and the only new bytes are the manifest's timestamp. What that
buys is a heartbeat: "nothing changed" and "this has been failing silently for a month" stop looking
identical in the log.

**⚠ THE BUG, AND IT IS THE MOST VALUABLE THING IN THIS TASK.** Rehearsing the script end to end
found a failure that *reports success forever* and surfaces only on the day the backup is needed:
**the push succeeded and a fresh clone came back EMPTY.** The remote's `HEAD` named a branch we had
never pushed to, so `git clone` warned "remote HEAD refers to nonexistent ref" and produced an empty
directory. Every run printed `done`.

The fix is not the interesting part; the lesson is. **A push that reports success is evidence about
the push. Only a clone is evidence about the backup.** So `install-backup.sh` now clones the
repository back, checks all four files are present and that `logbook.db` still hashes to what its own
manifest claims, and **refuses to enable the timer** otherwise — naming the exact fix. The push uses
an explicit refspec onto a branch discovered from the remote and stored in the repo's own config, so
the two cannot drift. Rehearsing also caught `git init -b main` needing git ≥ 2.28 against Ubuntu
20.04's 2.25.

This is the fourth time on this project that running something found what reading it did not, after
the untypeable colon, the stale service worker, and the off-screen save confirmation.

**Proven end to end before being called done**: 1296 real flights → snapshot → commit → push →
clone → `logbookctl verify` **green on all nine checksums** against the source CSVs. The broken
branch configuration was also reproduced deliberately, to confirm the check refuses it.

### 2026-08-02 — Task 13 as built: the coverage IS the feature

The plan (below) argued that the air-time figure must travel with its coverage. Running it against
the real books turned that from a principle into the obvious answer:

```
REG      TYPE   FLTS     BLOCK      AIR   air recorded
OH-CTL   C172    286    267:16     2:51   4 of 286
OH-PDP   P28A    275    193:19     0:45   2 of 275
OH-CWB   C172     65     71:57        —   0 of 65
TOTAL over 38 aircraft: 1296 flights, block 1222:17, air 13:07 from 19 of 1296
```

**A page that printed "OH-CTL — Air time: 2:51" next to 267 hours of flying would not be slightly
misleading, it would be worthless**, and it is the number an invoice gets checked against. So block
and air are separate fields with separate counts all the way through — `AircraftTime` carries
`AirKnown`/`AirMissing` beside `AirMinutes`, the API sends both, and the page's two sentences are
deliberately **different in kind**: block states a fact ("recorded on every flight"), air states a
fraction ("recorded on 1 of 2"). That asymmetry is the message; making them a matching pair of
figures is precisely the mistake.

Four decisions worth keeping.

**Block, not total, and the difference is counted rather than reconciled.** The licence figures run
on `TotalMinutes`; an owner charging by the hour charges chocks-to-chocks, which is `BlockMinutes`.
They agree on every row but one — 08/09/2025, block 0:45 vs total 0:38, a flagged discrepancy — so
the whole-logbook block total is **1222:17 against the licence total of 1222:10, exactly 7 minutes
apart**, and there is a real-data test asserting that relationship. `BlockDiffersFromTotal` counts
those rows and the page says so, because otherwise the first question is why two pages disagree.

**`Types` is a list, not a string.** One registration written with two types is a discrepancy, and
picking the more popular spelling would resolve it silently. The owner ruled the five historical
cases today and the CSVs are guarded, but that guard reads the *books* — a flight typed into the app
can still introduce one, and this page shows `C152 · C172` rather than choosing.

**`reg` adds the flights behind a figure without narrowing the summary.** Asking about one aeroplane
must not change what it is being compared against, so the rows always cover the whole range. Without
`reg` no flights are sent at all: 1296 flight objects is not a thing to hand a phone that asked for
totals.

**A sixth tab, and "Statistics" became "Stats".** The open question in the plan was where this page
hangs. Six tabs fit a 390px phone once the longest label is eight characters, with a font step down
under 400px. Hiding a page the owner asked for behind another page was the alternative, and a
shorter label costs nothing.

⚠ **Not yet done for this page: a run in a real browser.** The figures are proven against the real
books and the page is proven in jsdom, but this project has three times shipped something a green
suite loved and thirty seconds of real use exposed. **Open it on the phone before trusting the
layout** — six tabs and a seven-column table are exactly the shapes that have broken before.

### 2026-08-02 — Tasks 11 and 12 as built, and the two decisions that were not in the plan

Both were designed in the "first real day of use" entry below; this records what building them
actually settled.

**The confirmation replaces the form rather than joining it.** The old message was a
`<p role="status">` at the top of a three-card form — the design flaw was not that it was too quiet,
it was that it *shared the page with the thing it was about*, so on a phone the submit button and
the answer to "did that work?" could never be on screen together. Anything that stays on the same
page can be scrolled away from. So `FlightForm` returns the confirmation **instead of** the form,
which is also what makes "exactly one live region" true by construction rather than by care.

**It names what the server stored, not what was typed.** The panel is built from the `Flight` that
came back, which is why `onSave` now returns `{message, flight}` instead of a string. A screen that
read the draft back would agree with the pilot rather than with the logbook — and the whole reason
this panel exists is that the pilot could not tell what the logbook had.

**The refusal jumps by PAGE order, not server order.** `entry.Validate` deliberately reports every
problem at once, so a refusal routinely names three fields. Focusing whichever the server listed
first can scroll past two untouched errors to reach a third, and a form that behaves that way is a
form that gets abandoned — which means the flight is not logged at all. Hence `FIELD_ORDER`, and a
`refusals` counter rather than watching the error map: two submissions failing on the same field
must both move the pilot there, and comparing maps would call the second one "no change".

**Air time is derived in exactly two places and they solve different problems.** `format.airMinutes`
subtracts two stored **instants**, which carry their dates and therefore cross midnight on their
own; `FlightForm.blockFrom` rolls a bare four-digit **clock** by hand because the form has no date to
work with. Conflating them is how one of the two would quietly get the midnight case wrong, so both
are tested at midnight separately.

**Blank, never `0:00`.** 19 of the 1296 transcribed rows carry airborne times. A zero in the other
1277 is a claim that the aeroplane never left the ground — the same class of untruth as an empty
flight list after a failed fetch (rule §0.2).

Frontend tests **83 → 97**.

### 2026-08-02 — The five cells that could not be true, and how a frozen dataset gets corrected

Owner ruling, verbatim: *"C172 there is no C192. Also OH-CMU is always C152. We need to close this
permanently, now."* It came in while Task 11 was being built, after an aircraft-time question
surfaced that three registrations were flown as two types each.

**The finding is the argument.** Four rows gave `OH-GKT` and `OH-CTL` the type **`C192`**, and one
gave `OH-CMU` as a `C172`. *Cessna has never built a 192.* That is what separates this from every
other open item in the books: it is not a hard-to-read page or a disputed reading, it is a string
that no correct transcription of any page could have produced. The corroboration is one line away —
**book 2 line 139 is the same aeroplane on the same day**, `18/07/2018 · OH-CTL`, written `C172`.

**Five cells, one column, and not one figure moved.** `unknown_aircraft_type` 4 → 0 and
`type_conflict` 3 → 0, so discrepancies went **61 → 54**. Flights, total, PIC, dual, instrument,
night, instructor, seaplane, landings and the aircraft count are all **byte-for-byte unchanged**, and
that is structural rather than lucky: the seaplane/landplane split — the thing a rating is evidenced
by — derives from the **registration**, never the type (`stats.IsSea`, verified row by row against
`Cumulative_SEP_Sea` on 2026-08-01). The type column is provenance and display.

**This is the second time the freeze has been lifted, and the shape of both is the same**: the owner
names the cells, the correction is applied to those cells only, and the freeze resumes at the new
figures. The first was the three missing 28/08/2025 flights. `CLAUDE.md` §0.8 now says this
explicitly, because as written it forbade exactly this edit — *"not to fix a typo, not to close a
known discrepancy"* — and a rule that a session must quietly break is worse than no rule. **What a
session still may never do is decide on its own that a value looks wrong enough to change.** These
five sat surfaced and untouched for the whole project, which is the process working.

**Why this does not unfreeze the two items that stay open.** `logbook_2_final.csv` lines 89–90
(`04.05.2018` ×2) turn on a physical page nobody will re-read, and the 30 `landings_unverified` rows
need paper columns that were never photographed. Neither is settleable by inspection the way "there
is no Cessna 192" is. **The test is whether the data contradicts itself or merely disappoints us.**

**Closed with two guards that deliberately do not share a mechanism.**
`TestEveryRegistrationNamesOneRealAircraftType` asserts, in the language of the ruling, that no row
says `C192`, that **every registration maps to exactly one type**, and the three registrations by
name — and it was watched go red on all three axes by reintroducing one cell before it was accepted.
Separately, both discrepancy kinds stay in the `want` map **at zero** rather than being deleted,
because the "unexpected kind" sweep only catches kinds that were never listed. A guard living only
inside the discrepancy machinery would vanish the day somebody prunes that map.

`logbook_2.csv` (no `_final`) still reads `C192`. Nothing loads it — `csvbook.DefaultSources` names
only the three live books — and it is left as found rather than rewritten, because it is a superseded
artefact of the transcription workflow and not a fourth book.

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
> **⚠ SUPERSEDED 2026-08-14.** All three layers below are gone, replaced by one rule: nothing under
> `/logbook` is cached at all, and the app registers no service worker. The reasoning here is still
> worth reading — it is why the owner stopped trusting the cache — but the mechanism it describes no
> longer exists. See the 2026-08-14 entry and `docs/deploy.md`.

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
*(Figures as of this entry. After the 2026-08-01 owner rulings they were **38 aircraft, 61
discrepancies, and ZERO cumulative breaks** — the one break was Book 1 line 28, which the owner
ruled on and the CSV was corrected. The test now asserts zero. The 2026-08-02 aircraft-type ruling
then took discrepancies to **54**; the current figures are always the block near the top of this
file, not here.)*

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
