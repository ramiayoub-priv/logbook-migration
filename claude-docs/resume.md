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
- **Last processed pages:** `IMG_4940/4941/4942.jpg` (book pages 84–89), verified & appended
  2026-07-31 as a 3-page hybrid batch. 24 flights, **19/04/2020–30/05/2020**. Contents: club/DA40/
  PA28 flying (OH-PDP, OH-STL DA40 incl. **2 instructing**), a big run of **OH-CTL C172 seaplane**
  lake-hopping (Joensuu/Anttola/Salonsaari/Tuusulanjärvi/Kabböle/Pirttisaari; several instructing),
  OH-CDK C185 floatplane, and **OH-PIF returns twice** (post-CB-IR): 22/04 a *Kertauslento* (refresher,
  student, Autere) and 30/05 a VUO-MAT proficiency (PIC). **All page Δ cross-checks (total/PIC/
  student/instructor/landings) reconciled exactly** on all three pages. New regs: **OH-CAY** & **OH-CGX**
  (C172 landplanes, EFHV). New seaplane field: **EFRY** (Räyskälä). See drift.md for details.
- **Landing drift still CLOSED:** our Cumulative_Landings **= paper's printed count** (1661 at 30/05/2020).
- **Timezone (this batch):** only the two OH-PIF rows are UTC (`Z`, Aviatron-confirmed to the minute):
  22/04 EFLA-EFLA 12:56–14:51 and 30/05 EFLA-EFLA 10:53–12:46. Everything else plain local. OH-PIF can
  still reappear (Aviatron shows it again 23/09/2020) — cross-check Aviatron + mark `Z` when it does.
- **Last row in `logbook_2.csv`:** `30/05/2020 · P28A · OH-PIF · EFLA → EFLA · 10:53Z–12:46Z ·
  Total 1:53 · PIC · instrument 1:30 · 1 landing`
- **Cumulative totals at that row:**
  - Cumulative_Total **699:09**
  - Cumulative_PIC **556:39**
  - Cumulative_Student **142:30**
  - Cumulative_Instrument **53:13**
  - Cumulative_SEP_Sea **169:29**
  - Cumulative_Landings **1661** (equals paper's printed count — drift closed)
  - Cumulative_Instructor **32:28**
- **`logbook_2.csv` has 333 data rows** (+ header).
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
Process **IMG_4943.jpg** (next spread after 30/05/2020). Hybrid-batch pace is approved at **3 pages
per pass**: transcribe, cross-check each with `logbook_tools.py`, and surface only flagged rows for
the user. If a UTC-logged OH-PIF/OH-GKT row appears (OH-PIF recurs 23/09/2020 per Aviatron),
cross-check Aviatron and mark `Z`. Confirm each page by its dates (image numbers aren't chronological).
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
