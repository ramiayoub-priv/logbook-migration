# Resume Here — Current State

**Read this first every fresh session.** It is the single source of truth for where we are.

## The project in one paragraph
We are digitizing a series of paper pilot logbooks into CSV. **Book 1 complete** →
`logbook_1_final.csv`. **Book 2 complete** → `logbook_2_final.csv` (do not edit; both are finished
prior logbooks; `logbook_2_final.csv` is the seed for Book 3's cumulatives). **Book 3 is now in
progress** → `logbook_3.csv` — this is the active file we extend page by page. Book 3 is a **newer
EASA-format logbook** (different, wider layout) but we keep the **same 26-column CSV schema** so the
future app ingests all books uniformly.

## How pages get processed now (IMPORTANT — changed)
There is **no more ollama**. **Claude transcribes the logbook page images directly** (Read the
JPG in `logbook-3/IMG_XXXX.JPEG`), presents the rows to the user, the **user verifies**, and only
then do we append to `logbook_3.csv` with recomputed cumulative columns.
**Book-3 photos are rotated 90°** — rotate clockwise before reading, e.g.
`python3 -c "from PIL import Image; Image.open('logbook-3/IMG_6007.JPEG').rotate(-90, expand=True).save('/tmp/r.jpg')"`.
The old per-image files in `logbook-2-csv/` are stale ollama output — untrusted. See `workflow.md`.

## Current checkpoint — BOOK 3 IN PROGRESS (EASA logbook)
- **Book 2 is COMPLETE & finalized** → `logbook_2_final.csv` (421 rows, do not edit; seed for Book 3).
- **Book 3 started 2026-07-31.** Source photos live in **`logbook-3/IMG_6007–6037.JPEG`** (31 images,
  extracted from `logbook3_photos.zip`; each image is one two-page EASA spread; **rotate CW to read**).
  Book pages are numbered "Page N of 128" — order by page number + dates (image numbers ≈ order but
  confirm). `logbook_3.csv` seeded from `logbook_2_final.csv`'s last row, then extended.
- **First spread done:** `IMG_6007` (pages 1–2), 15 flights **01/06/2021–29/07/2021**, verified &
  appended. All UTC (Z). Club/float instructing: **OH-CTL** & **OH-GKT** seaplane hops on the lakes
  (Tuusulanjärvi, Salonsaari, Kahvisaari/Hillosensalmi) + **OH-PDP/OH-COK/OH-CAY** landplane flights.
  Instructing rows carry PIC=Instructor. **All 4 page cross-checks reconciled exactly**
  (Δtotal 12:10, Δpic 12:10, Δinstr 7:11, Δland 57 — read straight off the book's "TOTAL THIS PAGE").
- **Second spread done:** `IMG_6008` (pages 3–4), 15 flights **29/07/2021–10/09/2021**, verified &
  appended 2026-07-31. All UTC (Z). Clean page — **all 4 cross-checks exact, zero block warnings**
  (Δtotal 13:30, Δpic 13:30, Δinstr 2:45, Δland 44). No student/night this page. Instructing rows
  (OH-CAY, OH-CTL, OH-COK). Two instrument rows (OH-TIL P28A: 1:56 + 1:18). Lake ops on OH-CTL/OH-GKT.
- **Third spread done:** `IMG_6009` (pages 5–6), 15 flights **20/09/2021–26/12/2021**, verified &
  appended 2026-07-31. All UTC (Z). Time Δ all exact (Δtotal 13:56, Δpic 12:16, Δstudent 1:40,
  Δinstr 4:56). **Student row:** 20/09 OH-TIL EFLA→EFIM 1:40, PIC=**Tarhanen**, user was student
  (IR reval) → Student_Time 1:40 + Instrument 1:40 (user-confirmed). Instructing rows 29/10 OH-PDP +
  three 30/10 **OH-CAM** (new C172 reg). **First night flight of Book 3:** 17/12 OH-CAM 0:30 night,
  3 night landings. New reg **OH-CMV** (C152, 26/12). **LANDINGS CONVENTION CHANGED (user 2026-07-31):
  count ALL landings (day+night)** in Cumulative_Landings → now runs AHEAD of the book's day-only
  printed count. See drift.md.
- **Fourth & fifth spreads done:** `IMG_6010` (pages 7–8, 15 flights **10/02/2022–12/05/2022**) and
  `IMG_6011` (pages 9–10, 15 flights **12/05/2022–10/06/2022**), verified & appended 2026-07-31. All
  UTC (Z). Both fully reconciled (6010: Δtotal 14:18/Δpic 11:37/Δstu 2:41/Δinstr 4:02/Δland 54; 6011:
  Δtotal 14:23/Δstu 2:04/Δinstr 7:22/Δland 66). **New type SR20 (OH-ESR)** — Cirrus type-rating
  training with instructor **Stude** (3 student rows in 6010 + checkride "PASSED" 12/05 in 6011), then
  flown **PIC from 18/05** (SR20 instrument 1:00). **Seaplane student** row 13/05 OH-CTL with **Sinervä**.
  **Night flights** 23/02 (OH-CGX, 3 night ldg) & 03/03 (OH-CAM, 5 night ldg). 6011 row 5 (15/05 OH-CTL)
  on-block **inferred 08:24** (book wrote 07:24, a 60-min slip — user-confirmed). 6011 paper PIC-column
  slip 14:17 vs correct 12:19 (=total−student; our figure right).
- **Sixth spread done:** `IMG_6012` (pages 11–12), 15 flights **10/06/2022–21/07/2022**, verified &
  appended 2026-07-31. All UTC (Z). **All 5 cross-checks exact** (Δtotal 12:57, Δpic 12:35, Δstudent
  0:22, Δinstr 1:20, Δland 33). **NOTE: this image was already UPRIGHT in the original — no rotation
  needed** (unlike 6007–6011). Check each image's orientation; don't assume. Instructing row 22/06
  OH-CTL seaplane (Instructor 1:20). Student row 05/07 OH-TIL P28A EFTP local, PIC=**Salo**, Dual 0:22.
  Three instrument rows (SR20 OH-ESR: 30/06 0:30, 06/07 0:44 + 0:47 — EFNU↔EFTU Turku). New reg
  **OH-CAM** already seen; SR20 OH-ESR now routine PIC. Seaplanes: OH-CTL ×2 + **OH-CDK C185 floatplane**
  (21/07 Papinluoto→Astuvansalmi, Saimaa). Two inferred rows — see drift.md (17/07 on-block; 21/07 block).
- **Seventh spread done:** `IMG_6013` (pages 13–14), 15 flights **21/07/2022–04/09/2022**, verified &
  appended 2026-07-31. All UTC (Z). Image **already UPRIGHT — no rotation** (like 6012). No night this page.
  **All 5 cross-checks exact — but only after correcting a 1:00 book slip** (Δtotal 14:49, Δpic 12:54,
  Δstudent 1:55, Δinstr 2:17, Δland 29). **⚠ TARHANEN student row (10/08 OH-PIF EFLA local) = 1:55**
  (Student 1:55 + Instrument 1:55, pic_name=Tarhanen, another IR-reval): the book's running-Total column
  undercounted it by 1:00 (wrote 875:28), so the book's printed page/cumulative TOTAL & PIC are 1:00 low
  from here on; directly-summed cols (SE-IFR, Dual, Instructor) are correct. See drift.md. Missing landing
  on that row counted as **1** (user). **Row 7 (07/08 OH-CTL)** on-block inferred **17:10** (book's "12:10"
  impossible). **Instructing:** 01/08 & 07/08 OH-CTL Tuusula seaplane (Instr 1:11+1:06). **Lapland floatplane
  trip** 23/08 OH-CTL Inari→Kemijärvi→Sodankylä. **User-corrected reads:** Leikonvesi (float), Tuusula↔Hiidenvesi
  (24/09), row-15 date **04/09/2022** (book digits looked like 04/07).
- **Eighth spread done:** `IMG_6014` (pages 15–16), 15 flights **04/09/2022–11/02/2023**, verified &
  appended 2026-07-31. Image **already UPRIGHT — no rotation** (like 6012/6013). **All 4 cross-checks
  exact, zero block warnings** (Δtotal 14:07, Δpic 13:17, Δstudent 0:50, Δinstr 3:24). No night, no
  SE-IFR this page. **⚠ FIRST MIXED-TIMEZONE SPREAD:** the three 19/10/2022 OH-CAY rows looked
  overlapping — user confirmed **row 10 is UTC, rows 8 & 9 are local** (row 10 → 15:21 local slots
  cleanly after row 9's 15:04). Rows 8 & 9 stored **without `Z`**; the other 12 rows stay `Z`.
  **Don't assume a whole spread shares one time zone.** Row 8 on-block **inferred 13:51** (digit
  overwritten; 12:43 + 1:08). **Instructing ×3** (12/10 OH-PDP 1:07; 19/10 OH-CAY 1:08 + 1:09).
  **Student row** 11/02/2023 SR20 OH-ESR EFNU local 0:50, pic_name=**Stude** (second SR20 dual).
  Book left the per-page landings cell blank → no paper check; ours sums to 58.
- **Ninth spread done:** `IMG_6015` (pages 17–18), 15 flights **14/02/2023–06/05/2023**, verified &
  appended 2026-07-31. Image **already upright**. **All 3 available cross-checks exact, zero block
  warnings** (Δtotal 15:17, Δpic 15:17, Δinstr 3:36); no Dual this page. **Book slip: printed page
  SE-VFR total 12:35 should be 12:07** — the pilot self-corrected it downstream (struck 837:55 →
  837:27); we don't store SE-VFR so nothing propagates. **Row 7 = EFNU → ESNU (Umeå, Sweden)**,
  2:00 / Instrument 1:50 — a genuine one-way leg, **someone else flew the aircraft back** (user-confirmed;
  not a missing entry). **Night VFR** 10/03 OH-CGX EFHV local 0:37, **2 night landings** (day cell struck).
  **Instructing ×5** — all five 06/05 **Kabböle** OH-CTL seaplane locals (3:36); the 05/05
  Räyskälä→Kabböle ferry leg is PIC only. New airport **EFPO (Pori)** — 22/04 SR20 IFR day-return.
- **Last row in `logbook_3.csv`:** `06/05/2023 · C172 · OH-CTL · Kabböle → Kabböle ·
  17:28Z–18:08Z · Total 0:40 · PIC 0:40 · Instructor 0:40 · 5 landings`
- **Cumulative totals at that row (our continuous series, seeded from Book 2):**
  - Cumulative_Total **912:58** · Cumulative_PIC **753:18** · Cumulative_Student **159:40**
  - Cumulative_Instrument **78:08** · Cumulative_SEP_Sea **245:59**
  - Cumulative_Landings **2350** (= day+night; runs ahead of book's day-only count — see drift.md)
  - Cumulative_Instructor **84:32**
- **`logbook_3.csv` has 135 data rows** (+ header + seed row = 137 lines).

### Book-3 conventions (locked 2026-07-31 with user)
- **Same 26-col schema.** EASA→our-schema mapping: **Dual (Oppilas) → Student_Time**;
  **Single-Engine IFR → Instrument_Time**; **Flight Instructor → Instructor_Time**; **Co-pilot** ignored
  (none yet); **Multi-engine / Multi-pilot** none yet; **night landings** not stored (inferred later from
  Night_Time). The book's TOTAL column is a *running cumulative* — the per-flight time is in the
  SE-VFR / SE-IFR column.
- **Times are UTC → suffix `Z`** on Off/On block, UNLESS the entry is annotated **`LT`** (then plain
  local, no Z). The book is **not 100% consistent** — flag anything suspicious (out-of-order or
  colliding times) to the user. (Book-2 rows stayed local as written; do not retro-convert.)
- **We continue OUR internally-consistent cumulative series** (seeded from `logbook_2_final.csv`),
  NOT the EASA book's printed "previous pages" totals. The book's own per-page "TOTAL THIS PAGE"
  values are gold for cross-checking (offset-independent Δ): feed them as `d_total/d_pic/d_land/d_instr`.
- **Tooling:** `logbook_tools.py <batch.json> --csv logbook_3.csv [--append]` (new `--csv` flag targets
  Book 3; defaults to Book 2). Block-vs-total diffs ≤5 min now warn instead of blocking append.

## Next action — Book 3, IMG_6016 (pages 19–20)
Process **`IMG_6016`** next (continues from 06/05/2023). **Check orientation first** — image
orientation is NOT consistent across Book 3: 6007–6011 needed CCW `rotate(90)`, but **6012, 6013 and 6014
were already upright** in the original (no rotation). Quickest check: `Image.open(p).resize((1024,768))` and eyeball
which way is up; then crop the two pages at high res (the original is ~2048×1536). Transcribe all rows,
cross-check via the book's "TOTAL THIS PAGE" using `--csv logbook_3.csv`, and surface flags.
**Time zones can be mixed within one spread** (proved on IMG_6014) — when rows appear to overlap in
time, suspect a local-vs-UTC mix and ask the user *which rows*, not just whether.
**Hybrid-batch pace works well:** transcribe 2–3 spreads/pass, tool-reconcile each, present ONE digest
that greenlights clean pages and surfaces only flagged rows (student/instructing/night/odd-time/landing
anomalies) for user sign-off before append. Watch for: SR20 OH-ESR (now a PIC type), Stude (SR20 instr),
Sinervä + **Salo** (instructors), night landings, and the day+night landings convention.
Then continue IMG_6016…6037 at that pace (each spread is
~15 flights — sizeable, so 1–2 spreads per pass is plenty).
- **Paper-vs-ours drift, refreshed at end of page 18 (06/05/2023 boundary; EASA "TOTAL" bottom-of-page):**
  book Total **911:33** vs ours **912:58** (**+1:25**, steady since the IMG_6013 running-Total slip — our
  value correct); book PIC **754:43** vs ours **753:18** (**−1:25**); book SE-IFR **74:06** vs our
  Instrument **78:08** (**+4:02**, steady); book Dual **159:17** vs our Student **159:40** (**+0:23**, steady);
  book Flight-Instructor **85:52** vs ours **84:32** (**−1:20**, steady). **All five deltas have now held
  exactly across both IMG_6014 and IMG_6015** — strong confirmation of the transcription and that no new
  book slip has crept into a column we store.
  Landings: ours **2350** (day+night) runs ahead of the book's day-only cumulative (see drift.md).
  **Always cross-check on offset-independent per-page Δ ("TOTAL THIS PAGE"), never absolute totals — and
  note the book's Total column is itself 1:00 low from p.14 on.**
- **Remote:** `origin` = git@github.com:ramiayoub-priv/logbook-migration.git. master is pushed &
  up-to-date through IMG_6014 (`646cf85`) as of 2026-07-31; IMG_6015 appended locally (not yet committed);
  images/HEIC/zip are gitignored (not pushed). Git identity is now set repo-locally
  (`Rami Ayoub <rami.ayoub@gmail.com>`) — it was missing and blocked a commit.
  Push only when asked.

## Conventions locked in this session
- **Zulu times:** keep the `Z` suffix in Off/On block (e.g. `07:56Z`) for UTC-logged rows; plain
  `HH:MM` for local. Preserves which rows the future app must NOT re-convert.
- **IR instructor:** `Autere` (pilot flew as student in OH-PIF for the CB-IR rating).
- **Batch helper:** `logbook_tools.py <batch.json> --csv logbook_3.csv [--append]` computes cumulatives
  + cross-checks each page against the paper's Δtotal/Δpic/Δland/Δinstr before writing. Use it for every
  page. (Omit `--csv` and it targets Book 2. Block-vs-total diffs ≤5 min warn, don't block.)

> Note: the old ollama output in `logbook-2-csv/` shows obvious OCR errors (e.g. a date
> "41/09/2018" and reg "OK-PDP" in `logbook_IMG_4920.csv`) — exactly why we now transcribe with
> Claude and verify.

## Known open items
- Landings drift: the paper book's cumulative landing count has historically run ahead of the
  true count. Cross-check landing sums when in doubt. See `drift.md`.
