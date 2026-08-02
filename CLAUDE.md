# CLAUDE.md

This repo holds **two efforts** that share one dataset. As of **2026-08-02** only one of them is
live work:

1. **The migration** — digitizing the paper pilot logbooks into CSV. **COMPLETE AND CLOSED.**
   Docs: **`claude-docs/`**, now a historical record. See §0.9 below — this is a hard rule.
2. **The app** — the logbook web application at **`ayoub.fi/logbook`**. **The only active effort.**
   Docs: **`app/APP.md`**.

Read the rules below every session, then **`app/APP.md`**.

---

## 0. Non-negotiable rules (read every session)

**These are hard invariants. Breaking one is a defect, not a trade-off.**

1. **EVERYTHING MUST BE IN THE REPO AND PUSHED — work is not "recorded" until `git push` lands it on
   `origin`.** The Claude Code session log and machine-local `~/.claude/.../memory/` files **do NOT
   transfer** to another machine or a fresh clone. So:
   - **The repo is the single source of truth.** Every decision, task-board move, feature, gotcha and
     "why" goes into a **committed repo file** — `CLAUDE.md`, `app/APP.md`, `claude-docs/*` — **in the
     same change** that makes it true. Not only in chat, not only in a memory file. If it isn't in a
     repo file, it does not exist.
   - **Commit AND push before you stop.** Run `git status` + `git log origin/master..HEAD`, then
     **`git push`**. "Committed" ≠ "saved"; only pushed is saved.
   - A fresh session on a clean clone of `origin/master` must reconstruct where we are and why from
     repo files alone. That is the bar.

2. **THE LOGBOOK IS A LEGAL RECORD — NEVER LOSE OR SILENTLY CORRUPT A ROW.** This data backs licence
   privileges and currency; a wrong total is a real-world problem, not a bug report.
   - **⛔ THE PRODUCTION DATABASE IS THE SOURCE OF TRUTH.** *(Owner ruling, 2026-08-02: "we should
     start treating the production database now as the source of truth. We don't need the importer
     anymore.")* The migration is finished; the CSVs are frozen; re-importing them reproduces data
     that cannot have changed. **The importer is retired from production** — `update.sh` no longer
     imports, so no deploy runs a destructive operation on the live record to achieve a no-op. What
     protects the data is **the off-box backup** (`app/deploy/backup.sh`, daily, proven restorable on
     2026-08-02), not the ability to rebuild from CSVs — which stopped being a complete answer the
     moment the first flight was entered in the app.
     - **`logbookctl verify` STAYS, and is now the point.** It is read-only and compares the database
       against the CSVs on nine checksums, so it is a **drift and tamper check on the 1296 frozen
       historical rows** rather than a rebuild. Keep it in the deploy.
     - **`logbookctl import` still exists for dev scratch databases and tests only.** Never point it
       at production. **`logbookctl check`** is the restore check (no CSVs, no `sqlite3`).
     - The three CSVs stay in the repo as the frozen provenance record behind `claude-docs/drift.md`.
       They are simply no longer *loaded*.
   - **Every import/migration is idempotent, reversible, and verified.** It must be safe to re-run.
     Verify with **row counts AND total-time checksums** against the source CSVs, and refuse to
     complete on a mismatch. Schema migrations (`store.migrate`) run on the live file at every
     service start, so they must be **additive only** — never a rewrite, never a drop.
   - **Back up before any destructive operation.** The SQLite file is copied (or `VACUUM INTO`) before
     a migration runs, on both dev and prod.
   - **Discrepancies get surfaced, never silently fixed.** This rule is inherited from the migration
     (`claude-docs/`) and applies just as hard in the app: when the data disagrees with expectation,
     report it to the user and let them rule. Never "clean up" a value on your own initiative.
   - **The paper logbook is authoritative.** Electronic cross-references (Aviatron, laskukierros) are
     checks, not sources of truth.

3. **SECURITY IS KEY.** (User's explicit rule.) This app is on the public internet on a box that also
   serves other sites.
   - **Default deny.** Every endpoint requires an authenticated session unless it is explicitly and
     deliberately public. A new route is private until proven otherwise — never the reverse.
   - **No secrets in the repo, ever.** Not passwords, session keys, cookies, or tokens. Secrets come
     from environment or a root-owned file outside the web root. If a secret is ever pasted into chat
     or a file, treat it as compromised and rotate it.
   - **Passwords**: Argon2id only. Sessions are server-side and revocable, cookies `HttpOnly` +
     `Secure` + `SameSite=Lax`. Login is rate-limited.
   - **Keep the dependency tree near-empty.** Prefer the Go stdlib. Every new dependency is a supply-
     chain decision that must be justified in `app/APP.md`.
   - **Do not disturb the rest of the server.** Other sites and services share this box. Changes to
     Apache, ufw, systemd or Docker are additive, reversible, and verified from a second connection
     before the first is closed. **Never risk port 22.**

4. **TIME IS UTC, AND THERE IS EXACTLY ONE AUTHORITY FOR IT.** The paper books mix local and UTC (a
   `Z` suffix marks already-UTC rows). In the app, **every stored and displayed instant is UTC**.
   - Conversion from local uses **`Europe/Helsinki` with correct historical DST**, in **one** function.
     Do not re-implement time conversion anywhere else.
   - **Never lose the source.** Alongside the canonical UTC value, keep the raw string as written on
     paper and a `time_origin` flag (`utc_as_written` / `converted_from_local` / `unknown`). An
     ambiguous row surfaces for review rather than being guessed at silently.
   - **`tzdata` is embedded in the binary** so behaviour never depends on the server's zoneinfo.

5. **CUMULATIVE TOTALS ARE COMPUTED, NEVER STORED.** The CSVs carry seven `Cumulative_*` columns and
   they are the single largest source of drift in this project's history. In the app they are derived
   on demand from an explicit `seq` ordering. Do not add a stored running total, ever — if a page needs
   one (the EASA PDF does), compute it at render time.

6. **TESTS ARE THE BURDEN OF PROOF.**
   - **Failing test first, every change.** Write the test, watch it go **red**, then write the code,
     then watch it go **green**. A test you never saw fail proves nothing.
   - **Coverage bar: ≥80% on the backend overall, and 100% on the calculation core** — time
     conversion, aggregations, cumulative computation, PDF totals. That is the code where a bug means
     a wrong legal record. `go test -cover` is wired into the Makefile.
   - **Before every deploy, state why it will not break**: what changed, which tests cover it, which
     edge cases they assert. "It compiles" is not proof.
   - Enumerate failure modes *first* on shared paths — DST boundaries, midnight-crossing flights,
     empty ranges, missing aircraft, zero-flight months, rows with `unknown` time origin.

7. **NEVER BURN THE USER'S USAGE LIMIT WITH SUBAGENTS.** Max 1–2 at a time, and do not chain many in a
   row. Prefer doing the work inline. Only delegate when it genuinely needs breadth. **Do not spawn
   subagents or workflows unless the user asks for them.**

8. **THE HISTORICAL DATA IS CLOSED. DO NOT TOUCH IT.** *(Owner ruling, 2026-08-02, verbatim: "we
   will no longer touch historical data. This is the truth now. From now on the focus is on
   developing the logbook app.")*
   - **The three CSVs are frozen artefacts.** `logbook_1_final.csv`, `logbook_2_final.csv` and
     `logbook_3.csv` are **read-only inputs to the app**. Do not edit a cell, do not append a row,
     do not re-transcribe a page, do not "improve" a value — not to fix a typo, not to close a
     known discrepancy, not because a new photograph turned up. Every paper page that exists has
     been transcribed (`logbook-3/IMG_6007`–`IMG_6037`, book pages 1–62); there is no backlog.
   - **Only the OWNER lifts this freeze, explicitly, and only for named cells.** It has happened
     twice, both recorded in `app/APP.md`'s decision log: the three missing 28/08/2025 flights
     (2026-08-01) and the five aircraft-type cells (2026-08-02, below). A session **never** decides
     on its own that a value looks wrong enough to change — it surfaces it and stops. What the
     owner rules is then applied to the named cells only, and the freeze resumes at the new figures.
   - **New flights are entered in the app**, through `POST /flights`. That is now the only way the
     record grows. It lands with `source_book = 0` in its own `seq` band and survives everything.
   - **The numbers are final, and the tests now mean something stronger.** 1296 flights, total
     1222:10, PIC 1054:45, dual 167:25, instrument 107:58, night 22:45, instructor 189:41, seaplane
     407:39, landings 3444, 38 aircraft, **54 discrepancies**. `realdata_test.go` used to have one
     legitimate reason to go red — the CSVs growing. **That reason is gone. A change in any of
     those figures is now a defect, full stop**, and the fix is never to update the constant.
   - **Discrepancies were 61 until 2026-08-02**, when the owner ruled the aircraft-type slips:
     **"C192" is a typo for C172** (there is no such Cessna; the next flight of the same day in the
     same aeroplane is written C172) on book 2 lines 132, 133, 137, 138 — OH-GKT ×2, OH-CTL ×2 —
     and **OH-CMU is a C152 on every flight**, on book 3 line 434. Five cells, one column.
     `unknown_aircraft_type` 4 → 0 and `type_conflict` 3 → 0, both **kept in the map at zero**.
     **No time, landing, class or licence figure moved**: the sea/land split comes from the
     registration, never from the type. Guarded permanently and by name in
     `TestEveryRegistrationNamesOneRealAircraftType`.
   - **The open data questions stay open, permanently, and stay visible.** The 30
     `landings_unverified` rows keep their flag and the statistics page keeps saying so; the
     `logbook_2_final.csv` lines 89–90 date ambiguity stands unresolved — that one turns on a
     physical page nobody will re-read, which is exactly what the type slips did *not* turn on.
     Surfacing them is honest; closing them would require touching the data. Do not offer to
     finish them.
   - **`logbook_2.csv` (no `_final`) is a superseded transcription artefact and still reads
     `C192`.** Nothing loads it — `csvbook.DefaultSources` names the three files above and only
     those. Left as it was found; it is not a fourth book and it is not a bug to fix.
   - `claude-docs/` becomes a **historical record**: read it to understand why the data is what it
     is, never as a task list.

9. **NUDGE THE USER TO START A FRESH SESSION WHEN THIS ONE GETS LONG — and run the handoff first.**
   Say it out loud; don't wait to be asked. Finish the current unit of work, then: update `app/APP.md`
   (task board honest + a dated decision-log entry), rewrite the "NEXT SESSION STARTS HERE" block as a
   cold-start brief assuming zero memory of the conversation, then commit **and push**.

---

## 1. The migration (`claude-docs/`) — **CLOSED 2026-08-02**

**This effort is finished. Nothing below is a task.** Every photographed spread of every book is
transcribed, verified and reconciled, and the owner has closed the dataset (rule §0.8). What follows
is kept because the app depends on these conventions and because `drift.md` is the audit trail
behind every number in the record — read it to understand the data, never to change it.

If you are here because something looks wrong in the historical data: **surface it to the owner and
stop.** Do not prepare a fix.

- **`claude-docs/resume.md`** — the final state of the data, and the rulings that produced it.
- **`claude-docs/reference.md`** — CSV schema, seaplane regs, aircraft regs. **Still load-bearing
  for the app**: the importer reads these files and these conventions.
- **`claude-docs/drift.md`** — the corrections log. 106 KB of why the numbers are what they are.
- **`claude-docs/workflow.md`** — how a page *was* transcribed. Of historical interest only.

### What survived the closure, because the app depends on it
- **Three CSVs, 26 columns, 1298 rows → 1296 flights** (Books 2 and 3 each open with the previous
  book's final row as a cumulative seed; those two are skipped). **Read-only.**
- **A `Z` suffix means the time is already UTC**; its absence means Helsinki local. The app's single
  conversion authority is built on exactly this.
- **`Total_Time` is the figure the book totals on**, and it is block time on 478 of Book 3's 479
  rows.
- **The paper was authoritative.** Electronic cross-references (Aviatron, laskukierros) were checks,
  never sources. That question is now settled for good.

Details in `claude-docs/`. When in doubt about *why the data is what it is*, that directory wins.
When it reads like an instruction to do more transcription, rule §0.8 wins.

---

## 2. The app (`app/`)

**Read `app/APP.md` first** — it is the working tracker (task board + decision log + next-session
brief), the same role `claude-docs/resume.md` plays for the migration.

- **`app/APP.md`** — START HERE. Task board, decision log, cold-start brief.
- **`app/docs/data-model.md`** — schema, the CSV→DB mapping, and the domain rules behind it.
- **`app/docs/security.md`** — the threat model and every security control, with its test.
- **`app/docs/deploy.md`** — how it reaches `ayoub.fi/logbook`, and the server's shared-tenant map.

### Stack
- **Backend**: Go (stdlib `net/http`), SQLite via `modernc.org/sqlite` (pure Go, no CGO), embedded
  `tzdata`, PDFs via `go-pdf/fpdf`. Builds to a single static binary — deploy is rsync + restart.
- **Frontend**: React + TypeScript + Vite, built to static files. Mobile-first; it is used in the
  field. Node is a **build-time dependency only** — it never runs on the server.
- **Database**: one SQLite file. Backup = copy the file.

### Layout
```
app/
├── APP.md              # working tracker — task board, decision log, next-session brief
├── docs/               # data-model.md, security.md, deploy.md
├── backend/            # Go module: API, auth, import, stats, PDF
└── frontend/           # React + TS + Vite
```
