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

**Status: the data is in SQLite, verified. Backend calculation core done and green. Nothing is
deployed. No API and no frontend yet.**

### Done (2026-08-01)
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
internal/store/      schema.sql (embedded) + the verified import + read queries.  [85%]
internal/stats/      DOES NOT EXIST YET. The aggregations for the statistics page.
                     Already listed in the Makefile's CORE, so it prints SKIP today and
                     will be held to 100% the moment it exists.
```
`make cover-core` enforces 100% on everything marked `[core]`. That list is the code where a
bug means a wrong legal record. `internal/store` is held to the 80% bar, not 100%, because it
is I/O.

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
```
The server's own Go is 1.13 and irrelevant — we cross-compile (`CGO_ENABLED=0`). The one
dependency outside the stdlib is `modernc.org/sqlite` (pure Go, no CGO — that is what makes
the single static binary possible). Keep it that way (rule §0.3).

**The import is idempotent, backs up first (`VACUUM INTO`), and refuses to commit on any checksum
mismatch.** Re-run it freely. **There is no committed database** — it is generated, and
`app/.gitignore` keeps `*.db` and `*.bak` out of the repo. Nothing you need is only in a database
file; rebuild it from the CSVs in one command.

### The numbers the import produces — memorise these
```
flights 1293 | total 1219:35 | pic 1053:03 | dual 166:32 | instrument 107:14
night 20:50  | instructor 189:41 | seaplane 407:39 | landings 3439 | aircraft 38
```
Asserted in `internal/csvbook/realdata_test.go`, along with the exact count of each of the eight
discrepancy kinds. **If one of them changes unexpectedly, the import is wrong until proven otherwise —
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

### ⏸ TWO THINGS ARE STILL WAITING ON THE OWNER — ask about these first
Found by the importer on 2026-08-01 and logged in `claude-docs/drift.md` (top of file) and
`docs/data-model.md`. **The owner ruled on the rest the same day; see "Resolved" below.**

1. **`logbook_1_final.csv` line 28** (28/09/2011, OH-COF): `Instrument_Time` **1:21** on a flight
   whose total is **1:12** — more instrument time than flight time. The 9 minutes sit entirely inside
   Book 1 (its column sums 3:21, its own cumulative says 3:12; Books 2–3 chain off 3:12 exactly), and
   the preceding instrument lesson on the same aircraft logs instrument == the whole flight, so
   **1:12** is the reading. **This is why our instrument total is 107:14 while
   `Cumulative_Instrument` ends at 107:05 — the app computes from rows, so until it is fixed the app
   and the paper cannot agree.** Fixing it moves no total: every cumulative already reflects 1:12.
2. **`logbook_2_final.csv` lines 89–90** (`04.05.2018` ×2) — dated `DD.MM.YYYY`. Read day-first; six
   of the eight dotted rows prove it themselves. ⚠ These two cannot be settled from the cell, and
   **no electronic source can reach them** (Aviatron holds zero OH-PDP rows; the club file starts
   19/04/2020). **Needs the paper**, or they may be a month out. Row order only — no total moves.

#### ✅ Resolved 2026-08-01 (owner ruled; all verified, tests updated)
- **Night time is no longer 5:58 out.** `22:45` was never a mis-add — the gap was inherited from
  Books 1–2 (the EASA book carried **18:42** in against our 12:44). The owner read the paper's
  `Yölentoaika` column back and photographed seven Book-1 spreads, whose `Siirto` figures chain into
  a page-by-page ledger. Six night values were added and one moved onto its correct row.
  **Night 16:47 → 20:50**, matching the paper's `Siirto` at every checkpoint through 30/11/2013.
  ⏸ **1:55 remains, all of it in pages 52–69 (Mar–Aug 2014, not yet photographed)** — that is a
  migration task, not an app one. See `drift.md` item E.
- **`OK-PDP` → `OH-PDP`** (book 2 line 102) — a transcription typo that was seeding a phantom
  one-flight aircraft. Aircraft count **39 → 38**. Owner: any `OK-` reg in these books is `OH-`.
- **`28/01/2015` → `26/01/2015`** (book 1 line 249) — a date error, confirmed against the page.
- ⚠ **The p.62 day/night landing split (59 night / 3335 day) is now STALE** and must be recomputed
  once the night column closes. `Cumulative_Landings` is unaffected — only the split.

### Next task: #4, the API + authentication
Read **`docs/security.md`** first — the threat model, the Argon2id/session design and the
default-deny router are all specified there, and the `users`/`sessions` tables already exist in
`schema.sql` (declared, no logic behind them yet). Rule §0.3 governs: **default deny, no secrets in
the repo, stdlib only.** Argon2id means `golang.org/x/crypto/argon2` — the one dependency worth
adding, and it must be justified in this file when you do.

Everything the API needs to read already exists in `internal/store`: `Flights()`, `Aircraft()`,
`Discrepancies()`, `CountFlights()`, `Verify()`. Two things are missing:

- **`internal/stats`** — the aggregations for the statistics page, and the cumulative computation the
  EASA PDF needs. Computed from `seq`, never stored (rule §0.5). Held to 100%.
- **`cmd/server`** — does not exist. `make build` picks it up automatically once it does.

The API lives under **`/logbook/api/`**, not `/api/` — that path on `ayoub.fi` is already taken by a
stale transit proxy.

**The `discrepancies` table is already populated** (56 rows) and is what the frontend's "needs review"
list reads. It is rewritten on every import, so an item that gets resolved in the CSV simply
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

### Where the reasoning lives
Do not re-derive these — they are argued out in the Decision Log below (§5), all dated 2026-08-01:
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
| 4 | API + authentication | not started |
| 5 | Four frontend pages (mobile-first) | not started |
| 6 | Three PDF exports (EASA clone + table + stats) | not started |
| 7 | PWA + deploy to `ayoub.fi/logbook` | not started |
| 8 | Backfill landings day/night for the 22 night rows | not started — the 22 rows are already flagged `landings_unverified` in the DB and listed by `logbookctl import`. `claude-docs/drift.md` has the analysis: 17 rows are full-night and certain (50 landings); **9 of the 59 night landings are estimates** from five partial-night rows, and two exact sources exist on paper but were never read. |
| 9 | Rule on the three open source-data problems | **blocked on the owner** — see the ⏸ block at the top of this file |

---

## 5. Decision Log

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
discrepancies** — see the task board above; the single cumulative break is unchanged.)*

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
   the cumulative column advances by 1:12, so the row is the outlier. Our instrument total is
   therefore 107:14 against the column's 107:05.
2. **`logbook_2_final.csv` lines 83–90** — dates written `DD.MM.YYYY`. Read day-first, which six of
   the eight settle themselves and the chronological bracket confirms; the two `04.05.2018` rows are
   flagged for a look at the paper.
3. **Night time 16:47 (ours) vs 22:45 (inked at p.62)** — a 5:58 gap on the one p.62 figure that
   `drift.md` records as never having been read back.
   **Update, same day: resolved down to 1:55.** The importer's job here was only to surface the gap;
   the owner then read the paper's night column back and photographed seven Book-1 spreads, which
   turned it into a page-by-page ledger. Night is now **20:50** and the residual is one unphotographed
   page range. *The flag was worth raising precisely because nobody had ever compared that column.*

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

### 2026-08-01 — EASA PDF covers all 1295 flights, not just Book 3

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
flag; the importer seeds everything as day and flags the **22 rows carrying `Night_Time`** for
backfill from the page images (Task 8). Bounded and small.

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
