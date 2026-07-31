# Resume Here — Current State

**Read this first every fresh session.** It is the single source of truth for where we are.

## The project in one paragraph
We are digitizing a series of paper pilot logbooks into CSV. **Book 1 is complete** →
`logbook_1_final.csv` (do not edit; it is the finished prior logbook and the seed for Book 2's
cumulative totals). **Book 2 is in progress** → `logbook_2.csv` — this is the active file we
extend page by page. More logbooks (Book 3+) will follow after Book 2 is finished.

## How pages get processed now (IMPORTANT — changed)
There is **no more ollama**. **Claude transcribes the logbook page images directly** (Read the
JPG in `logbook-2/IMG_XXXX.jpg`), presents the rows to the user, the **user verifies**, and only
then do we append to `logbook_2.csv` with recomputed cumulative columns.
The old per-image files in `logbook-2-csv/` are stale ollama output — treat them as untrusted
hints at best, not source of truth. See `workflow.md` for the step-by-step.

## Current checkpoint
- **Last processed page:** `IMG_4924.jpg` (book pages 52/53), verified & appended 2026-07-31.
  First training page: IR flights with pilot as student (OH-PIF, instructor **Autere**), Zulu times.
- **Last row in `logbook_2.csv`:** `12/04/2019 · P28A · OH-PDP · EFHF → EFPR ·
  09:11Z–09:55Z · Total 0:44 · PIC self · 1 landing`
- **Cumulative totals at that row:**
  - Cumulative_Total **559:53**
  - Cumulative_PIC **464:10**
  - Cumulative_Student **95:43**
  - Cumulative_Instrument **5:20**  ← first change since 3:12 (IR training instrument time)
  - Cumulative_SEP_Sea **128:09**
  - Cumulative_Landings **1307**
  - Cumulative_Instructor **9:17**
- **`logbook_2.csv` has 192 data rows** (+ header).
- **PIC vs paper:** our Cumulative_PIC runs **+1:19** ahead of the paper's *corrected* value (pure
  seed drift). The paper's *written* PIC is a further 1:20 low from an arithmetic slip it flags on
  IMG_4921 ("pic added +1:20"); our PIC is the more correct figure.

## Next action
Process **IMG_4925.jpg** (pages 54/55) — still in the training section: DA40 (OH-SIL/OH-STL)
appears, more IR/student flights, seaplane lake ops, and **strike-throughs**. Do these **one page at
a time** with explicit user classification; batching resumes once pages are clean PIC.

## Conventions locked in this session
- **Zulu times:** keep the `Z` suffix in Off/On block (e.g. `07:56Z`) for UTC-logged rows; plain
  `HH:MM` for local. Preserves which rows the future app must NOT re-convert.
- **IR instructor:** `Autere` (pilot flew as student in OH-PIF for the CB-IR rating).
- **Batch helper:** `logbook_tools.py <batch.json> [--append]` computes cumulatives + cross-checks
  each page against the paper's Δtotal/Δpic/Δlandings before writing. Use it for every page.

> Note: the old ollama output in `logbook-2-csv/` shows obvious OCR errors (e.g. a date
> "41/09/2018" and reg "OK-PDP" in `logbook_IMG_4920.csv`) — exactly why we now transcribe with
> Claude and verify.

## Known open items
- Landings drift: the paper book's cumulative landing count has historically run ahead of the
  true count. Cross-check landing sums when in doubt. See `drift.md`.
