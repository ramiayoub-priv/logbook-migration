# CLAUDE.md

This repo holds **two efforts** that share one dataset:

1. **The migration** — digitizing a series of paper pilot logbooks into CSV. Docs: **`claude-docs/`**.
2. **The app** — a logbook web application served at **`ayoub.fi/logbook`**. Docs: **`app/APP.md`**.

Both are live work. Read the rules below every session, then the doc set for whichever you are touching.

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
   - **Every import/migration is idempotent, reversible, and verified.** It must be safe to re-run.
     Verify with **row counts AND total-time checksums** against the source CSVs, and refuse to
     complete on a mismatch.
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

8. **NUDGE THE USER TO START A FRESH SESSION WHEN THIS ONE GETS LONG — and run the handoff first.**
   Say it out loud; don't wait to be asked. Finish the current unit of work, then: update `app/APP.md`
   (task board honest + a dated decision-log entry), rewrite the "NEXT SESSION STARTS HERE" block as a
   cold-start brief assuming zero memory of the conversation, then commit **and push**.

---

## 1. The migration (`claude-docs/`)

Every fresh session touching the migration, **read `claude-docs/` first**.

- **`claude-docs/resume.md`** — START HERE. Current checkpoint, what's next, project overview.
- **`claude-docs/workflow.md`** — how to process one logbook page, step by step.
- **`claude-docs/reference.md`** — CSV schema, seaplane regs, aircraft regs, sanity checks.
- **`claude-docs/drift.md`** — corrections log and known discrepancies.

### The essentials
- **`logbook_3.csv`** is the active file we are building (Book 3, a newer **EASA-format** logbook).
  `logbook_1_final.csv` and `logbook_2_final.csv` are completed Books 1 & 2 (do not edit;
  `logbook_2_final.csv` seeds Book 3's cumulatives). Same 26-col schema across all books.
- **Claude transcribes the page images directly** (`logbook-3/IMG_XXXX.JPEG`, **rotate CW** — they're
  sideways); **the user verifies** before anything is appended. (No more ollama.)
- **Hybrid-batch pace (user-approved):** transcribe up to **3 pages per pass**, cross-check each
  with `logbook_tools.py`, and present a report that greenlights clean pages and surfaces only the
  ambiguous/flagged rows. The user still verifies the flags before append — never fully autonomous.
- After appending, **update the checkpoint in `resume.md`** and log any corrections in `drift.md`.

Details in `claude-docs/`. When in doubt, that directory wins over this summary.

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
