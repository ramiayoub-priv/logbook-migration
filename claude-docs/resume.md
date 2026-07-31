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
- **Last processed page:** `IMG_4921.jpg` (book pages 46/47), verified & appended 2026-07-31.
  Its 15/10/2018 OH-CWB line was struck through in the book and deliberately excluded (see
  `drift.md`).
- **Last row in `logbook_2.csv`:** `28/10/2018 · C172 · OH-CWB · EFHV → EFHF ·
  13:10–13:40 · Total 0:30 · PIC self · 1 landing`
- **Cumulative totals at that row:**
  - Cumulative_Total **540:49**
  - Cumulative_PIC **447:14**
  - Cumulative_Student **93:35**
  - Cumulative_Instrument **3:12**
  - Cumulative_SEP_Sea **128:09**
  - Cumulative_Landings **1268**
  - Cumulative_Instructor **7:57**
- **`logbook_2.csv` has 168 data rows** (+ header).

## Next action
Identify and process the next page after 28/10/2018. Confirm the correct next image with the user
before transcribing — image file numbers are NOT in chronological order, so always verify the page
by its dates, not its filename.

> Note: the old ollama output in `logbook-2-csv/` shows obvious OCR errors (e.g. a date
> "41/09/2018" and reg "OK-PDP" in `logbook_IMG_4920.csv`) — exactly why we now transcribe with
> Claude and verify.

## Known open items
- Landings drift: the paper book's cumulative landing count has historically run ahead of the
  true count. Cross-check landing sums when in doubt. See `drift.md`.
