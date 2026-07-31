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
- **Last processed page:** `IMG_4928.jpg` (book pages 60/61), verified & appended 2026-07-31.
  DA40 OH-STL (row1 PIC+instructor, row2 PIC), OH-PIF IR student (Autere, Zulu, Aviatron-confirmed),
  OH-CTL seaplane (row6 PIC, row7 instructing), OH-GKT seaplane (C172, ex-SE-GKT) PIC.
- **Last row in `logbook_2.csv`:** `06/06/2019 · C172 · OH-GKT · Kahvisaari → Kelvenne ·
  13:10–13:43 · Total 0:33 · PIC self · 4 landings`
- **Cumulative totals at that row:**
  - Cumulative_Total **586:41**
  - Cumulative_PIC **481:28**
  - Cumulative_Student **105:13**
  - Cumulative_Instrument **13:18**
  - Cumulative_SEP_Sea **136:04**
  - Cumulative_Landings **1385**
  - Cumulative_Instructor **14:19** (matches paper)
- **`logbook_2.csv` has 222 data rows** (+ header).
- **Open item:** IMG_4928 rows 6/7 landings entered as 4/1 from a hard-to-read column (page total
  +21 is certain); if the book shows 1/4 instead, only those two intermediate Cumulative_Landings
  change — trivial fix.
- **Cross-check note:** our Cumulative_Student runs ~+0:24 above the paper's written column — the
  paper under-counted student by 0:27 on IMG_4925 (an arithmetic slip; our value is correct).
- **PIC vs paper:** our Cumulative_PIC runs **+1:19** ahead of the paper's *corrected* value (pure
  seed drift). The paper's *written* PIC is a further 1:20 low from an arithmetic slip it flags on
  IMG_4921 ("pic added +1:20"); our PIC is the more correct figure.

## Next action
Process **IMG_4929.jpg** (next spread after 06/06/2019). Continue one page at a time with user
classification; use `logbook_tools.py` + Aviatron cross-ref for OH-PIF/OH-GKT rows. Confirm the
next page by its dates (image numbers aren't chronological).

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
