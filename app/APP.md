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

**Status: foundations landed, calculation core done and green. Nothing is deployed. No frontend yet.**

### Done (2026-08-01)
- Recon of `ayoub.fi` → `docs/deploy.md` has the full shared-tenant map. **The box is shared with the
  owner's other sites; do not disturb them.**
- Stack chosen: Go + SQLite + React/Vite (Decision Log §5).
- `CLAUDE.md` §0 rules written, adapted from the neighbouring `transit` project.
- Server cleanup, both reversible: transit's orphaned Quarkus killed (it runs on its own VM now);
  OpenVPN stopped + disabled at the owner's request.
- `app/backend/` Go module with **`internal/hhmm`** (H:MM ↔ integer minutes) and **`internal/timeutil`**
  (the single UTC-conversion authority). **Both at 100% coverage**, written failing-test-first.

### How to run things
```bash
export PATH=$HOME/.local/go/bin:$PATH   # Go 1.26 lives here; the system had none
cd app/backend
make check      # vet + race tests + both coverage gates
make cover-core # the 100% gate on the calculation core
```
The server's own Go is 1.13 and irrelevant — we cross-compile (`make build`, `CGO_ENABLED=0`).

### Next task: #3, the schema + importer
**This is the riskiest piece in the project.** It must reproduce **1295 flights** and their totals
exactly. Before writing code, read **`docs/data-model.md`** — the schema, the full CSV→DB column
mapping, and the domain rules are already specified there. Key points that will bite otherwise:

- **Skip each book's first data row.** It is the previous book's carried-over final row (a cumulative
  seed), and importing it would double-count three flights.
- **Do not import the seven `Cumulative_*` columns.** Use them as a verification checksum, then drop
  them. Cumulatives are computed in this app, never stored (rule §0.5).
- **Verify and refuse on mismatch** (rule §0.2): row counts *and* total-time checksums against the
  CSVs. An importer that "mostly worked" is a corrupted legal record.
- **Surface the known data-quality items, never auto-fix them**: `OK-PDP` (1 row), type `C192`
  (4 rows), `OH-CMU` typed as both C152 and C172. Listed with context in `docs/data-model.md`.
- Source files are at the repo root: `logbook_1_final.csv`, `logbook_2_final.csv`, `logbook_3.csv`
  (26 columns, all values quoted, dates `DD/MM/YYYY`).

### Open questions awaiting the owner
- Is the `kraken-predictor-python-2` container on `:8000` still wanted? Publicly exposed, up 2 years,
  and now the box's largest memory consumer (~759 MB / 38%).
- Prune the stale ufw rules for `30814` and `19132` (nothing listens on either)?
- **Rotate the `rami` sudo password** once deployment is done — it was pasted into a chat session on
  2026-08-01 and must be treated as compromised. Tracked in `docs/security.md`.

### Traps already paid for — do not rediscover these
- **Go's `time.Date` is silent on both DST edges, in different ways.** See the 2026-08-01 entry in the
  Decision Log. `internal/timeutil` already handles it; do not "simplify" that check.
- **Docker bypasses ufw.** A published container port is not closed by a firewall rule. See
  `docs/deploy.md`.
- **`/api/` on `ayoub.fi` is already taken** by a stale transit proxy, which is why our API lives at
  `/logbook/api/`.
- **Port 22 is under constant attack** (fail2ban: 50,264 bans). Never risk it.

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
| DB | SQLite (`modernc.org/sqlite`, pure Go) | No CGO ⇒ trivial cross-compile. One file; backup = copy it. 1295 rows is nothing. |
| Time | embedded `tzdata` | Behaviour must not depend on the server's zoneinfo. |
| PDF | `go-pdf/fpdf` | Absolute positioning, which a fixed 15-row EASA grid needs. Headless Chrome would cost 300 MB+. |
| Frontend | React + TS + Vite | Builds to static files. Node is build-time only, never on the server. |

## 4. Task Board

| # | Task | Status |
|---|---|---|
| 1 | Project rules + app docs | **done** 2026-08-01 |
| 2 | Scaffold backend + frontend, test harness | **backend done** 2026-08-01; frontend not started |
| 3 | Schema + importer for 1295 flights (verified) | not started |
| 4 | API + authentication | not started |
| 5 | Four frontend pages (mobile-first) | not started |
| 6 | Three PDF exports (EASA clone + table + stats) | not started |
| 7 | PWA + deploy to `ayoub.fi/logbook` | not started |
| 8 | Backfill landings day/night for the 22 night rows | not started |

---

## 5. Decision Log

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
