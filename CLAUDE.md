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
- **Never batch-process pages.** One page at a time, with the user's confirmation.
- After appending, **update the checkpoint in `resume.md`** and log any corrections in `drift.md`.

Details in `claude-docs/`. When in doubt, that directory wins over this summary.
