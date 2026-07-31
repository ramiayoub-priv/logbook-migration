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
- **Last processed pages:** `IMG_4937/4938/4939.jpg` (book pages 78–83), verified & appended
  2026-07-31 as a 3-page hybrid batch. 23 flights, **30/09/2019–17/04/2020**. Contents: the **CB-IR
  skill test** (03/10/2019, OH-PIF, examiner **Timo Aineslahti** FI.FCL 20163 — CB-IR fully done,
  OH-PIF era ends here), then a long run of club/DA40/PA28 flying (OH-STL DA40, C152 OH-COF/OH-CRA/
  OH-NEU/OH-KLS, OH-PDP), incl. two **instructing** flights and one DA40 **student (KOU)** flight.
  **All page Δ cross-checks (total/PIC/student/instructor/landings) reconciled exactly** on all three
  pages. New: **EETN** (Tallinn — first Estonian/intl field, 31/12/2019 DA40 day-trip), **OH-NEU**
  (C152), **Särkijärvi** (seaplane lake). See drift.md for details.
- **Landing drift CLOSED (2026-07-31):** on IMG_4937 the paper's landing column advanced +12 while
  its row entries summed to +13; following the rows brought our cumulative to **exactly the paper's
  printed bottoms** (1542, then 1591). The historical paper **+1 landing lead is now consumed** —
  our Cumulative_Landings equals the paper's printed count as of 17/04/2020.
- **First Night_Time in the file:** 26/03/2020 DA40 evening flight, Night 0:50 (no Cumulative_Night
  column in schema; per-row only).
- **Timezone (this batch):** only the two 03/10 OH-PIF skill-test rows are UTC (`Z`, Aviatron-
  confirmed to the minute); everything else (OH-PDP/DA40/C152/C185) is plain local. OH-PIF era is
  over, so future pages are expected to be all-local unless a new UTC-logged type appears.
- **Last row in `logbook_2.csv`:** `17/04/2020 · DA40 · OH-STL · EFHF → EFHF · 06:38–08:50 ·
  Total 2:12 · PIC · instrument 1:00 · 3 landings`
- **Cumulative totals at that row:**
  - Cumulative_Total **675:50**
  - Cumulative_PIC **535:34**
  - Cumulative_Student **140:16**
  - Cumulative_Instrument **49:48**
  - Cumulative_SEP_Sea **158:10**
  - Cumulative_Landings **1591** (now equals paper's printed count — drift closed)
  - Cumulative_Instructor **25:25**
- **`logbook_2.csv` has 309 data rows** (+ header).
- **Workflow note:** user approved a **hybrid-batch** pace (2026-07-31) — transcribe several pages
  per pass, auto-cross-check each with `logbook_tools.py`, present a report that greenlights clean
  pages and flags only ambiguous rows for user verification. Not fully autonomous.
- **Open item:** IMG_4928 rows 6/7 landings entered as 4/1 from a hard-to-read column (page total
  +21 is certain); if the book shows 1/4 instead, only those two intermediate Cumulative_Landings
  change — trivial fix.
- **Cross-check note:** our Cumulative_Student runs ~+0:24 above the paper's written column — the
  paper under-counted student by 0:27 on IMG_4925 (an arithmetic slip; our value is correct).
- **PIC vs paper:** our Cumulative_PIC runs **+1:19** ahead of the paper's *corrected* value (pure
  seed drift). The paper's *written* PIC is a further 1:20 low from an arithmetic slip it flags on
  IMG_4921 ("pic added +1:20"); our PIC is the more correct figure.

## Next action
Process **IMG_4940.jpg** (next spread after 17/04/2020). Hybrid-batch pace is approved at **3 pages
per pass**: transcribe, cross-check each with `logbook_tools.py`, and surface only flagged rows for
the user. OH-PIF/OH-GKT (Aviatron) rows are unlikely from here (CB-IR done); if a UTC-logged row
appears, cross-check Aviatron and mark `Z`. Confirm each page by its dates (image numbers aren't
chronological).
- **Paper-vs-ours drift:** landings now MATCH the paper's printed count (drift closed 17/04/2020).
  Total runs ~a few min either side of paper depending on the pilot's lump corrections; PIC ~+0:28
  and Instrument still run ahead of paper (seed drift). Always cross-check on offset-independent
  per-page Δ, never absolute totals.
- **Remote:** `origin` = git@github.com:ramiayoub-priv/logbook-migration.git. User pushed master
  (13+ commits) on 2026-07-31; images/HEIC are gitignored (not pushed). Push only when asked.

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
