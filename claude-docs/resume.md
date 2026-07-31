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
- **Last processed pages:** `IMG_4946/4947/4948.jpg` (book pages 96–101), verified & appended
  2026-07-31 as a 3-page hybrid batch. 24 flights, **19/07/2020–08/10/2020**. Contents: club **OH-PDP
  P28A** hops (EFHF/EFSA/EFHV/EFFO), a new **OH-TIL P28A** (Aviatron aircraft, IFR-capable Arrow) run
  of instrument flights (28/07 dual checkout + 05/08 & 06/08 & 21/08 PIC instrument), **OH-CDK C185
  floatplane** Saimaa (Salonsaari/Hillosensalmi/Papinniemi), **OH-CTL C172 seaplane** Tuusulanjärvi
  locals + EFRY, and **OH-PIF P28A** returning 23/09 for an **IR/SEP proficiency check** at EFJY.
  **All page Δ reconciled exactly** (total/PIC/student/instructor/landings). Two student flights, three
  instructing flights (22/09 OH-PDP 1:40, 24/09 OH-CTL 0:37, 01/10 OH-PDP 1:22). New: **EFFO** (Forssa),
  **OH-TIL** reg, place **Papinniemi**. See drift.md.
- **Landing drift still CLOSED:** our Cumulative_Landings **= paper's printed count** (1781 at 08/10/2020).
- **Timezone (this batch):** club/seaplane rows plain local; the three **23/09 OH-PIF** rows are **UTC →
  `Z`** (Aviatron-confirmed to the minute). OH-TIL rows plain local (OH-TIL not in Aviatron until 2021).
- **Last row in `logbook_2.csv`:** `08/10/2020 · P28A · OH-PDP · EFHF → EFHF · 17:00–17:37 ·
  Total 0:37 · PIC 0:37 · 1 landing`
- **Cumulative totals at that row:**
  - Cumulative_Total **749:30**
  - Cumulative_PIC **600:49**
  - Cumulative_Student **148:41**
  - Cumulative_Instrument **63:32** (+10:19 this batch: OH-TIL 7:10 + OH-PIF 3:09)
  - Cumulative_SEP_Sea **190:32**
  - Cumulative_Landings **1781** (equals paper's printed count — drift closed)
  - Cumulative_Instructor **42:11**
- **`logbook_2.csv` has 381 data rows** (+ header).
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
Process **IMG_4949.jpg** (next spread after 08/10/2020). Hybrid-batch pace is approved at **3 pages
per pass**: transcribe, cross-check each with `logbook_tools.py`, and surface only flagged rows for
the user. If a UTC-logged OH-PIF/OH-GKT/OH-TIL row appears, cross-check Aviatron and mark `Z`.
Confirm each page by its dates (image numbers aren't chronological — note the 23/09 OH-PIF rows sat
on page 98/99 *before* the 11/09–08/10 club rows on page 100/101).
- **Paper-vs-ours drift (at 08/10/2020):** landings MATCH paper's printed count (1781, drift closed
  17/04/2020). Total runs **+25** ahead of paper (749:05 printed → ours 749:30; seed residue). PIC runs
  **−27** vs paper's column (paper 601:16; the gap widened 3 min from the 20/08 paper PIC slip below).
  Instrument runs **+4:02** ahead of paper's Mittari (59:30 printed; steady seed drift). Instructor runs
  **+4** ahead (paper 42:07). Always cross-check on offset-independent per-page Δ, never absolute totals.
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
