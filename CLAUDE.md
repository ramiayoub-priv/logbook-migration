# CLAUDE.md

Digitizing a series of paper pilot logbooks into CSV.

## Rule #1 — Resume seamlessly
Every fresh session, **read `claude-docs/` first**. All state, workflow, and history live there so
any new session can pick up exactly where the last left off. Keep them current as you work.

- **`claude-docs/resume.md`** — START HERE. Current checkpoint, what's next, project overview.
- **`claude-docs/workflow.md`** — how to process one logbook page, step by step.
- **`claude-docs/reference.md`** — CSV schema, seaplane regs, aircraft regs, sanity checks.
- **`claude-docs/drift.md`** — corrections log and known discrepancies.

## The essentials
- **`logbook_2.csv`** is the active file we are building. `logbook_1_final.csv` is completed
  Book 1 (do not edit; it seeds Book 2's cumulatives). More books follow after Book 2.
- **Claude transcribes the page images directly** (`logbook-2/IMG_XXXX.jpg`); **the user
  verifies** before anything is appended. (No more ollama.)
- **Hybrid-batch pace (user-approved):** transcribe up to **3 pages per pass**, cross-check each
  with `logbook_tools.py`, and present a report that greenlights clean pages and surfaces only the
  ambiguous/flagged rows. The user still verifies the flags before append — never fully autonomous.
- After appending, **update the checkpoint in `resume.md`** and log any corrections in `drift.md`.
- **Session-length heads-up:** when the session grows long (context filling up, many pages
  processed in one sitting), proactively tell the user and nudge them to start a fresh session — a
  clean session re-reads `claude-docs/` and resumes seamlessly.

Details in `claude-docs/`. When in doubt, that directory wins over this summary.
