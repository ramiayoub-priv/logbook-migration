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
- **Last processed pages:** `IMG_4934/4935/4936.jpg` (book pages 72–77), verified & appended
  2026-07-31 as a 3-page hybrid batch. 24 flights, 10/08–27/09/2019 — the tail of the **CB-IR
  course** (OH-PIF Student/IR, Autere; course completed 27/09) plus OH-PDP/OH-CTL/OH-CDK PIC and two
  new **instructing** flights. All page Δ cross-checks (total/PIC/student/instructor/landings)
  reconciled exactly. New airports: **EFJY** (Jyväskylä), **EFLP** (Lappeenranta), **EFTP**
  (Tampere-Pirkkala). See drift.md for the IMG_4936 pilot correction and the 27.08→27.09 fix.
- **Timezone (this batch):** all OH-PIF rows + the 10/08 OH-GKT row are **Aviatron-confirmed UTC**
  and were written **with the `Z` suffix**; OH-PDP/OH-CTL/OH-CDK stay plain (local; one 17/09 OH-PDP
  row is explicitly marked "LT").
- **Z backfill DONE (2026-07-31):** added `Z` to the 12 earlier plain OH-PIF rows that were actually
  UTC — 18/04 (×3), 14/06, 28/06 (×2), 23/07, 01/08 (×2), 02/08 (×3). Eleven were confirmed against
  Aviatron to the minute; **18/04 05:44–06:35 was inferred** (not in Aviatron, but its two same-day
  sibling lessons are UTC and 05:44 local = 02:44 UTC is implausibly pre-dawn) — flag for the user if
  their memory differs. File now has 37 `Z`-marked time cells. Cumulatives unchanged (Z is
  computation-neutral). See drift.md.
- **Last row in `logbook_2.csv`:** `27/09/2019 · P28A · OH-PIF · EFLA → EFLA · 14:22Z–15:11Z ·
  Total 0:49 · Student/IR (Autere) · 1 landing` (CB-IR course completed)
- **Cumulative totals at that row:**
  - Cumulative_Total **650:28**
  - Cumulative_PIC **513:29**
  - Cumulative_Student **136:59**
  - Cumulative_Instrument **44:10** (runs +4:02 ahead of paper's Mittari 40:08 — pure seed drift; per-page Δ match)
  - Cumulative_SEP_Sea **157:25**
  - Cumulative_Landings **1529**
  - Cumulative_Instructor **23:22** (matches paper)
- **`logbook_2.csv` has 286 data rows** (+ header).
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
Process **IMG_4937.jpg** (next spread after 27/09/2019). Hybrid-batch pace is approved at **3 pages
per pass**: transcribe, cross-check each with `logbook_tools.py`, and surface only flagged rows for
the user. Use Aviatron cross-ref for OH-PIF/OH-GKT rows (also resolves local-vs-UTC — mark UTC rows
with `Z`). Confirm each page by its dates (image numbers aren't chronological).
- **Paper-vs-ours drift, post-IMG_4936 correction:** the pilot's 27.9.2019 "kokonaisaika korjattu"
  lump-corrected the paper Total column, so it now sits **~5 min AHEAD** of ours (was ~26 behind).
  PIC and instrument columns still lag ours (seed drift). Always cross-check on offset-independent
  per-page Δ, never absolute totals.
- **Z-backfill of earlier UTC OH-PIF rows is complete** (see checkpoint block above / drift.md).

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
