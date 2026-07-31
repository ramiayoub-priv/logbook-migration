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
- **Last row in `logbook_2.csv`:** `29/08/2018 · C172 · OH-CTL · Gumbostrand → Tuusulanjärvi ·
  13:12–13:45 · Total 0:33 · PIC self · 2 landings`
- **Cumulative totals at that row:**
  - Cumulative_Total **530:11**
  - Cumulative_PIC **436:36**
  - Cumulative_Student **93:35**
  - Cumulative_Instrument **3:12**
  - Cumulative_SEP_Sea **125:23**
  - Cumulative_Landings **1239**
  - Cumulative_Instructor **7:57**
- **`logbook_2.csv` has 153 data rows** (+ header).

## Next action
Identify and process the next page after 29/08/2018. Candidate image is **`IMG_4919.jpg` /
`IMG_4920.jpg`** (both contain 29/08/2018 flights; 4920 continues into 05–30/09/2018). Confirm
the correct next page with the user before transcribing — image file numbers are NOT in
chronological order, so always verify the page by its dates, not its filename.

> Note: `logbook-2-csv/logbook_IMG_4920.csv` (old ollama output) shows obvious OCR errors like a
> date "41/09/2018" and reg "OK-PDP" — exactly why we now transcribe with Claude and verify.

## Known open items
- Landings drift: the paper book's cumulative landing count has historically run ahead of the
  true count. Cross-check landing sums when in doubt. See `drift.md`.
