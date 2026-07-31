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
- **Last processed pages:** `IMG_4949/4950/4951.jpg` (book pages 102–107), verified & appended
  2026-07-31 as a 3-page hybrid batch. 24 flights, **13/10/2020–07/05/2021** (crosses into 2021).
  Contents: club **OH-PDP P28A** hops (EFHF/EFHV/EFLA), a **07/12/2020 OH-TIL P28A** night flight,
  a new **DA40 OH-STL** night-IFR flight (11/02/2021: instrument 1:36 + night 1:55), C152 club flying
  (**OH-COF**, **OH-CRA** at EFNU), a new **C172 OH-COK** at **EFPR**, and an **18/04/2021 float trip
  to EFIM (Immola)**: two OH-COK ferry legs (UTC) + two **C185 OH-CDK seaplane student** flights
  (instructor **Matikainen**, SEP-sea class-rating check, "FI.RCL2234").
  **All page Δ reconciled exactly** (total/PIC/student/landings). Two student flights, no instructing.
  New: **OH-COK** (C172), **DA40 OH-STL** recurring, airports **EFPR** and **EFIM** (Immola). See drift.md.
- **Landing drift still CLOSED:** our Cumulative_Landings **= paper's printed count** (1843 at 07/05/2021).
- **Timezone (this batch):** all rows plain local **except** the two **18/04/2021 OH-COK** ferry legs,
  which the user confirmed are **UTC → `Z`** (07:25Z→08:38Z EFIM→EFPR, 10:54Z→12:04Z EFPR→EFIM; the
  book had them in the wrong order with two arrows — reordered by time). C185 float rows stay local.
- **Last row in `logbook_2.csv`:** `07/05/2021 · P28A · OH-PDP · EFHV → EFHV · 20:05–20:32 ·
  Total 0:27 · PIC 0:27 · 1 landing`
- **Cumulative totals at that row:**
  - Cumulative_Total **772:37**
  - Cumulative_PIC **622:29**
  - Cumulative_Student **150:08** (+1:27 this batch: two OH-CDK student flights)
  - Cumulative_Instrument **65:08** (+1:36 this batch: 11/02/2021 DA40 OH-STL night-IFR)
  - Cumulative_SEP_Sea **191:59** (+1:27 this batch: OH-CDK 0:59 + 0:28)
  - Cumulative_Landings **1843** (equals paper's printed count — drift closed)
  - Cumulative_Instructor **42:11** (unchanged — no instructing this batch)
- **`logbook_2.csv` has 405 data rows** (+ header).
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
Process **IMG_4952.jpg** then **IMG_4953.jpg** — these are the **last two images of Book 2**
(only 4952 & 4953 remain in `logbook-2/`; likely the final spread(s), so this is the **closeout
batch**). Transcribe both, cross-check each with `logbook_tools.py`, and surface only flagged rows
for the user. If a UTC-logged OH-PIF/OH-GKT/OH-TIL/OH-COK row appears, cross-check Aviatron and mark
`Z`. Confirm each page by its dates (image numbers aren't chronological).
- **After Book 2 is complete:** rename/finalize `logbook_2.csv` per the project's Book-1 pattern
  (`logbook_1_final.csv`) if the user wants, then Book 3 begins with a fresh source set.
- **Paper-vs-ours drift (at 07/05/2021):** landings MATCH paper's printed count (1843, drift closed
  17/04/2020). Total runs **+25** ahead of paper (772:12 printed → ours 772:37; seed residue). PIC runs
  **−27** vs paper's column (paper 622:56). Instrument runs **+4:02** ahead of paper's Mittari (61:06
  printed; steady seed drift). Instructor runs **+4** ahead (paper 42:07). Student runs **+0:23**
  ahead of paper's Oppilas column (149:45 printed → ours 150:08; the long-standing IMG_4925 slip).
  Always cross-check on offset-independent per-page Δ, never absolute totals.
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
