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
- **Last processed page:** `IMG_4923.jpg` (book pages 50/51), verified & appended 2026-07-31.
- **Last row in `logbook_2.csv`:** `03/04/2019 · P28A · OH-PDP · EFHF → EFHV ·
  13:06–13:26 · Total 0:20 · PIC self · 2 landings`
- **Cumulative totals at that row:**
  - Cumulative_Total **553:53**
  - Cumulative_PIC **460:18**
  - Cumulative_Student **93:35**
  - Cumulative_Instrument **3:12**
  - Cumulative_SEP_Sea **128:09**
  - Cumulative_Landings **1298**
  - Cumulative_Instructor **9:17**  ← corrected 2026-07-31 (was 7:57; see drift.md)
- **`logbook_2.csv` has 184 data rows** (+ header).
- **PIC vs paper:** our Cumulative_PIC runs **+1:19** ahead of the paper's *corrected* value (pure
  seed drift). The paper's *written* PIC is a further 1:20 low from an arithmetic slip it flags on
  IMG_4921 ("pic added +1:20"); our PIC is the more correct figure.

## Next action
Process **IMG_4924.jpg** (pages 52/53) — the training section starts here: IR-training flights
where the pilot is the **student** (Student_Time + pic_name = instructor, PIC blank), instructing
flights (PIC **and** Instructor both filled), Zulu times, and strike-throughs. Do these **one page
at a time** with explicit user classification; batching resumes once pages are clean PIC.

> Note: the old ollama output in `logbook-2-csv/` shows obvious OCR errors (e.g. a date
> "41/09/2018" and reg "OK-PDP" in `logbook_IMG_4920.csv`) — exactly why we now transcribe with
> Claude and verify.

## Known open items
- Landings drift: the paper book's cumulative landing count has historically run ahead of the
  true count. Cross-check landing sums when in doubt. See `drift.md`.
