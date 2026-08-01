# Resume Here — Current State

**Read this first every fresh session.** It is the single source of truth for where we are.

> ⚠ **This repo now holds a second effort.** As of 2026-08-01 we are also building the **logbook
> application** (`ayoub.fi/logbook`) in **`app/`** — its tracker is **`app/APP.md`**. This file
> (`claude-docs/`) remains the source of truth for **the CSV migration only**. The non-negotiable
> rules that govern both live in the repo root **`CLAUDE.md` §0** — read those first.
>
> The two efforts share the dataset: the app imports `logbook_1_final.csv`, `logbook_2_final.csv`
> and `logbook_3.csv`. So the migration's conventions — the `Z` suffix, `Total_Time` being the
> figure the book totals on, the paper being authoritative — are now load-bearing for the app too.
> **The migration is still the way flights get digitized; the app does not replace it.**
>
> ## ✅ STATE OF THE DATA IN ONE TABLE (2026-08-01) — read this before re-verifying anything
>
> All 1293 flight rows, checked row by row against the paper's inked page-62 block:
>
> | column | ours | paper p.62 | |
> |---|---|---|---|
> | Total | 1206:58 | 1206:58 | ✅ |
> | PIC | 1040:26 | 1040:26 | ✅ |
> | SE-IFR / Instrument | 105:57 | 105:57 | ✅ |
> | Dual / Student | 166:32 | 166:32 | ✅ |
> | FI / Instructor | 185:50 | 185:50 | ✅ |
> | Landings (sum) | 3394 | 3394 | ✅ |
> | **Night** | **22:45** | **22:45** | ✅ **closed 2026-08-01** |
>
> ### 🎯 ALL SEVEN `Cumulative_*` SERIES RECONCILE ROW-BY-ROW WITH **ZERO BREAKS** OVER 1293 ROWS.
> Every time column's own sum now equals its stored final cumulative, exactly:
> Total **1219:35** · PIC **1053:03** · Student **166:32** · Instrument **107:05** ·
> SEP_Sea **407:39** · Landings **3439** · Instructor **189:41** · Night **22:45**.
>
> **THE TIMES MATCH. Do not re-validate the book on spec.** Re-check with the command at the end of
> this block if you want proof; do not re-transcribe pages that already reconcile.
>
> ### ⛔ THE END-OF-BOOK-3 CUMULATIVES ARE NOW FINAL AND FROZEN (user ruling, 2026-08-01)
> > *"At this point, the FINAL and cumulative times at the end of book 3 are the final values, I will
> > not correct them again as they are now corrected on paper too and match the csvs… The app and
> > paper logbook will need to match with the cumulative numbers we have now."*
>
> **No future change — migration or app — may move those seven figures.** Any proposed fix must be
> shown not to, by diffing every `Cumulative_*` cell before and after. ⚠ **Note the direction of the
> logic**: the line-28 fix below *looked* like it would move a total and in fact was the only way to
> *preserve* one. Test the claim, don't assume it.
>
> **What is still open** (full detail in `drift.md`; **none of it moves any total**):
> 1. **`logbook_2_final.csv` lines 89–90** — two `04.05.2018` dates. Affects row *order* only, and
>    only if the paper says 5 April. No electronic source can settle it; needs the physical page.
> 2. **`logbook_2_final.csv` line 97** — sits one hour off Aviatron. Duration is right either way.
> 3. **The p.62 inked landing split `59 night / 3335 day` is stale → recomputes to `68 / 3326`.**
>    The landing *sum* (3394) is unaffected. **Correct the paper, not the CSV.**
> 4. **`logbook_1_final.csv` line 173** has a stray `Siirto` in its `Remarks` cell — cosmetic.
>
> **Closed 2026-08-01:** the 5:58 night gap (now 22:45, Δ 0:00, via the p.52/53 photo); the line-28
> instrument break (`1:21` → `1:12`, no cumulative moved); three p.52 airfield codes; the `OK-PDP`
> typo. *(The importer originally surfaced three unlogged problems — line 28, the dotted dates and
> the night gap. Two are now closed; the dotted dates remain as item 1.)*
>
> **Re-investigated the same day — each one is now localised, and a fourth and fifth turned up:**
> - **Line 28 (instrument):** the 9 min is entirely inside Book 1 (its column sums 3:21, its own
>   cumulative says 3:12; Books 2–3 chain off 3:12 exactly). The preceding instrument lesson
>   (line 20, same aircraft & instructor) also logs instrument == the whole flight, so **1:12** is
>   the reading. **Fixing it moves no total** — every cumulative already reflects 1:12. It matters
>   because the app *computes* cumulatives from rows (rule 5) and would show 107:14 vs the paper's
>   107:05.
> - **Lines 89–90 (dates):** **no electronic source can settle it** — Aviatron has zero OH-PDP rows,
>   `laskukierros_flights.csv` starts 19/04/2020. Needs the physical page.
> - **⭐ Night 5:58:** **`22:45` is NOT a mis-add and Book 3 is clean.** The EASA book's carry-in
>   night is **18:42** vs our Books 1+2 **12:44** — the whole gap is inherited. Book 3 reconciles
>   exactly: 18:42 + our 4:03 = **22:45**. So the missing night time is in the **old paper books'
>   night column**, not in Book 3 and not in the addition.
>   **The user then read the paper night column back and photographed seven Book-1 spreads**
>   (pp. 34/35, 36/37, 38/39, 48/49, 50/51, 70/71, 74/75 — the first Book-1 images in the project).
>   The book's `Yölentoaika` `Siirto` figures chain continuously, which turns it into a page-by-page
>   ledger; full table in `drift.md`. Night **16:47 → 20:50**, and **our running night now equals the
>   paper's `Siirto` at every checkpoint through 30/11/2013**.
>   - ✳️ Applied: **21.01.2013 OH-CTM 1:17** (line 111), **15.09.2014 OH-CMO 0:51** (line 237),
>     **01.02.2015 OH-CAV 0:37** (line 250), **25.02.2015 OH-KAM 0:10** (line 253, the later leg).
>   - ✳️ **⚠ A night value was on the WRONG ROW:** the `0:24` belongs to **09/11/2012 OH-CWB**
>     (line 107), not 15/11/2012 OH-KAS (line 108). Moved; total unchanged. *First known one-row
>     slip — suspect it when a night value sits on a row whose clock times make no sense.*
>   - ✅ **`26.01.2015` was a DATE ERROR, not a missing flight** — line 249 read `28/01/2015`; the
>     paper says `26.1`. Fixed. Page 74/75 cross-checks exactly (`Cumulative_Total` 264:27 and
>     landings 571 both equal the paper), so no Book-1 flight is missing.
>   - ✳️ Applied on the user's ruling *"go with the photo read"*: **05.01.2015 OH-KAM 0:30**
>     (line 247) and **26.01.2015 OH-STL 0:38** (line 248, full night). This supersedes his earlier
>     dictated list, which put those two values one row apart; totals identical either way.
>   - ✅ **CLOSED by the p.52/53 photograph (`IMG_6048`, 2026-08-01).** The last 1:55 was **all on one
>     spread**: **25.02.2014 C152 OH-KLS** EFHF local 18:31–19:26 **0:55, full night** (line 173) and
>     **26.03.2014 P28A OH-TIL** EFJY→EFHF 18:05–20:06 **1:00 of 2:01** (line 177). The page's own
>     `Yölentoaika` runs `Siirto` **9:12** → bottom **11:07**, and **11:07 is exactly p.71's `Siirto`**
>     — so pages 54–69 carry no night at all and the column is closed, not merely sampled.
>     Night **20:50 → 22:45 = the paper, Δ 0:00.** All eight rows on the spread matched our CSV on
>     reg, block times and duration; `Kokonaisaika` 186:08→194:00 and landings 410→420 both equal ours.
>   - ✅ **The p.62 day/night landing split recomputed: `59/3335` → `68/3326`** (62 certain + 6
>     estimated from three multi-landing partial-night rows; range 65–72). **The landing sum 3394 is
>     unchanged.** ⏸ The *paper* needs correcting, not the CSV.
>   - **Reading method for Book-1/2 pages:** `Yölentoaika` = night, `Kokonaisaika` = running total,
>     `Päällikkö` = PIC, `Oppilas` = student, `Opettaja` = instructor, `Siirto` = carried forward.
>     Values straddle the dotted row separators and the photos are taken at an angle — **pin a row by
>     `Kokonaisaika`, and pair the night entry with the `Päällikkö` value on its visual line.**
> - **NEW, ✅ FIXED — line 102 of `logbook_2_final.csv` read `OK-PDP`**, a one-off OCR typo for
>   `OH-PDP` (would have created a phantom aircraft in the app). User: **any `OK-` reg in these books
>   is a typo for `OH-`.** Also: **`SE-GKT` (Book 1) and `OH-GKT` are the same airframe**
>   re-registered — the app must not split them.
> - **NEW — line 97** (`10/05/2018` OH-DBS) sits exactly one hour off Aviatron while the same day's
>   other row matches to the minute. No total affected.
>
> **⚠ `Night_Time` comes from the book's night column and nothing else — never infer it from clock
> times, sunset or time zones** (user, 2026-08-01). A solar calculation was used once, only to rank
> which paper rows are worth re-reading, and no computed value went into any CSV.
>
> This also means the CSVs now have a machine-checked invariant: run
> `cd app/backend && go run ./cmd/logbookctl import -dry-run -csv ../..` after any append. It re-runs
> every reconciliation and prints anything that no longer adds up.

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
- **Tenth spread done:** `IMG_6016` (pages 19–20), 15 flights **07/05/2023–14/06/2023**, verified &
  appended 2026-07-31. All UTC (Z). Image **already upright**. **All 3 available cross-checks exact**
  (Δtotal 13:22, Δpic 12:48, Δinstr 4:41); no night, no SE-IFR. Peak float season — 13 of 15 rows are
  OH-CTL seaplane (SEP_Sea +11:56). **Student row** 12/05 OH-CTL Laajasalo local 0:34, pic_name=**Sinervä**
  (his third seaplane dual). **Book slip: printed page Dual total 0:37 vs the row's true 0:34** → our
  standing Dual drift moves +0:23 → **+0:20**. **Two book time slips, resolved in OPPOSITE directions
  by the user** — row 2 (07/05) the *on-block* "15:25" was wrong → 14:25 (0:41); row 6 (19/05) the
  on-block "12:09" was wrong → **12:15** (off-block 10:47 was right, 1:28). *Don't assume the off-block
  is the bad cell just because fixing it closes the arithmetic — ask.* Row 2 arrival illegible, user
  read it **Laajasalo**. **Row 5 (09/05 SR20) logged out of date order** between two 12/05 rows —
  user-confirmed as written. **Instructing ×5** (07/05 Kabböle 1:13; 12/05 Laajasalo 0:52; 24/05 Tuusula
  1:23; 14/06 Tuusula↔Halsholmen 0:45+0:28). New float places: **Anttola, Siltasaari, Pellinki, Halsholmen**
  (all user-confirmed). Landings cell blank again → no paper check; ours sums 58.
- **Eleventh & twelfth spreads done:** `IMG_6017` (pages 21–22, 15 flights **19/06/2023–17/07/2023**)
  and `IMG_6018` (pages 23–24, 15 flights **12/07/2023–07/09/2023**), verified & appended 2026-07-31.
  Both **already upright**. Both fully reconciled (6017: Δtotal 15:59/Δpic 15:59/Δinstr 4:09;
  6018: Δtotal 14:02/Δpic 12:57/Δstudent 1:05/Δinstr 1:08). No night, no dual on 6017.
  **⚠ 6017 carries a handwritten pilot correction: "* Error in total time 14:07 → 15:59"** — row 12
  (14/07 **OH-PIF** EFLA local) is 2:07 = 0:15 VFR + **1:52 SE-IFR**, but the running-Total column
  dropped the IFR leg. He fixed it at the page total; our row sum hits 15:59 independently. **That row
  is PIC, not student (user-confirmed) — the first non-student OH-PIF row in either book.**
  **Sweden round-trip with a return leg this time:** 15/07 EFNU→EFTU→**ESNU** (IFR 1:56) and 17/07
  ESNU→EFTU (IFR 1:50), SR20. **🎓 6018 row 13 = CRI(A) RATING EARNED:** 04/09 **OH-GKT**
  Kahvisaari local 1:05, remark *"AoC for CRI(A) Passed"* + **FI.FCL.34041**, PIC=**RAVANTTI** (new
  name) → Student_Time 1:05. **⚠ Second mixed-timezone spread:** 04/09 rows 12 & 13 collided; user
  says **"GKT is UTC"** → row 13 keeps `Z`, **row 12 stored local**; row 14 left as `Z` (no conflict —
  residual uncertainty, see drift.md). **User date fixes:** 6017 row 8 → 07/07 (book said 09/07),
  row 4 → 29/06 (same day as row 3); 6018 row 5 → **12/07/2023**, "an older flight I missed", entered
  out of order. New float places **Ranua, Viitasaari** (19/08 Lapland ferry 2:23 — my read, unconfirmed).
  Landings cells blank on both pages → no paper check; ours sum 35 and 42.
- **Thirteenth & fourteenth spreads done:** `IMG_6019` (pages 25–26, **14** flights
  **24/08/2023–26/10/2023**) and `IMG_6020` (pages 27–28, 15 flights **26/10/2023–05/03/2024**),
  verified & appended 2026-07-31. Both upright, all UTC. Fully reconciled (6019: Δtotal 15:04/Δpic
  13:42/Δstudent 1:22/Δinstr 1:20; 6020: Δtotal 10:56/Δpic 10:56/Δinstr 3:59).
  **⚠ 6019 row 2 is STRUCK THROUGH and is a duplicate of IMG_6018's last row** (07/09/2023
  Tuusulanjärvi local, OH-CTL, 1:20, 6 ldg) — excluded; the page total only works without it.
  **6019's Total column is scribbled illegible** — the pilot's margin note **"* 15:04 ←"** is what
  pins the page total (corroborated by SE-VFR 14:04 + SE-IFR 1:00 and by p.28's carry).
  **NEW AIRCRAFT OH-MIL** — a Maule on floats, type **"M6(sea)"**, always on floats (user). 24/08
  Tuusulanjärvi→**Keilaniemi** 1:22 **STUDENT with Sinervä**, 7 ldg. **New airport EFVP (Vampula).**
  **⚠ The pilot struck his own PIC total on p.26** (810:09 → carried 807:39, a hand **−2:30**
  correction; he doesn't recall the details). Our per-page Δpic reconciled exactly on both pages so
  our figure was kept — **but the PIC drift flips −1:25 → +1:05.** See drift.md.
  **6020 crosses into 2024**, almost all EFHV circuits. New regs **OH-AWB** (C152) and **OH-CMU**
  (C152 — genuinely distinct from OH-CMV, user-confirmed). Book wrote C172 for the 09/02/24 OH-AWB
  row; user says C152, corrected. **Two night flights** (09/02/24 0:39 + 5 night ldg; 05/03/24 0:23) —
  on the second the book put its 3 landings in the DAY column by mistake; they are **night** landings.
- **Fifteenth spread done:** `IMG_6021` (pages 29–30), 15 flights **28/03/2024–04/05/2024**, verified &
  appended 2026-07-31. All UTC (Z). Image **already upright**. **All 4 cross-checks exact, zero block
  warnings** (Δtotal 9:41, Δpic 8:49, Δstudent 0:52, Δinstr nil). No night, no seaplane, no instructing.
  **Two student rows** 29/04 SR20 OH-ESR EFNU→**EFIK (Kiikala, new)**→EFNU (0:28 + 0:24), pic_name=**Stude**
  — his fourth SR20 dual. **⚠ OH-CAM (C172) logs 0:25 SE-IFR** on 30/04 EFTP→EFHV (1:01 = 0:36 VFR +
  0:25 IFR) — **user confirms OH-CAM is IFR-certified**, not a misread. Two inferred block cells
  (row 10 on-block 05:21; row 13 off/on 16:37→17:38), both pinned by the column totals. The pilot
  carried his "TOTAL PREVIOUS PAGES" 1:52 low then struck/rewrote the bottom line back to 990:37 —
  nets out, no action. Landings blank → ours sums 31.
- **Sixteenth spread done:** `IMG_6022` (pages 31–32), 15 flights **05/05/2024–07/06/2024**, verified &
  appended 2026-08-01. **⚠ THIS IMAGE IS SIDEWAYS — needs CCW `rotate(90)`** (6012–6021 were all
  upright; orientation is genuinely per-image, always thumbnail first). **All 4 cross-checks exact,
  zero block warnings** (Δtotal 8:58 all SE-VFR, Δpic 8:58, Δdual nil, Δinstr 1:07).
  **🎉 CROSSES 1000 HOURS** — Cumulative_Total hits **1000:02** on row 14 (07/06/2024 Vääksy→Lietsaari).
  **⚠ Third mixed-timezone spread — and the rule is BY AIRCRAFT:** user says *"all rows are local
  except OH-GKT and OH-ESR"*, **with one confirmed exception** — row 10 (26/05 SR20 OH-ESR EFNU local)
  is stored **local**, because as UTC it lands 15:20–16:05 local and breaks the 26/05 chain (row 11
  already flew him home to EFHV at 14:05). Net: **rows 1–11 no `Z`, rows 12–15 (OH-GKT) `Z`**.
  **Row 12 date = 30/05/2024** (book's day digit reads "34"; user-confirmed). **`C172sea` is not a
  type** — it's the pilot's own informal seaplane marker; **store plain `C172`** (user). Float season
  2024 opens on OH-GKT in the Päijänne area — new places **Lieso, Lietsaari** (Vääksy already known).
  One instructing row 05/05 OH-CAY EFHV local 1:07. Landings blank → ours sums 36.
- **Seventeenth spread done:** `IMG_6023` (pages 33–34), 15 flights **17/06/2024–28/06/2024**, verified &
  appended 2026-08-01. **⚠ SIDEWAYS AGAIN — CCW `rotate(90)`** (same as 6022). **All 3 cross-checks
  exact, zero block warnings** — every block time matched its logged time to the minute
  (Δtotal **15:01**, all SE-VFR; Δpic 15:01; Δinstr 12:37). No night, no IFR, no dual.
  Pure float-instruction spread: **13 of 15 rows are OH-CTL** at Tuusulanjärvi, rows 14–15 OH-GKT.
  **Instructing ×11** — rows 1–9 (17–18/06 Tuusulanjärvi locals + the Hiidenvesi out-and-back) and
  rows 12–13 (27/06 Tuusula↔**Pellinki**). Rows 10/11 (Tuusula↔Kahvisaari ferry) and 14/15 PIC only.
  **Time zones (user-confirmed): the IMG_6022 per-aircraft rule held — OH-CTL rows 1–13 stored
  local (no `Z`), OH-GKT rows 14–15 stored `Z`.**
  **Row 7** (18/06 14:29–15:12) sits in the book between the 16:16 and 18:55 rows — user says it is
  simply **entered out of order**, times as written, same zone as its neighbours (not a zone mix).
  **Row 8's arrival is struck through**; user confirms **Hiidenvesi** (row 9 departs from there).
  **🏠 `Kahvisaari` is OH-GKT's HOME BASE near Lahti** (user) — so Tuusula↔Kahvisaari ~0:40 and
  Padasjoki→Kahvisaari 0:27 are normal ferry hops. New place **Padasjoki** (Päijänne).
  The pilot's running-Total column on this page is built off the struck **997:43** carry (1:52 low,
  the same p.30 slip) but he fixed it at the bottom line: 999:35 + 15:01 = **1014:36**. No action.
  Landings cell blank → ours sums **65**.
- **Eighteenth & nineteenth spreads done:** `IMG_6024` (pages 35–36, 15 flights **28/06/2024–12/07/2024**)
  and `IMG_6025` (pages 37–38, 15 flights **12/07/2024–21/07/2024**), verified & appended 2026-08-01.
  **Both SIDEWAYS — CCW `rotate(90)`** (6022–6025 now all sideways). **All six cross-checks exact**
  (6024: Δtotal 14:05/Δpic 14:05/Δinstr **11:47**; 6025: Δtotal 11:52/Δpic 11:52/Δinstr **5:13**).
  All SE-VFR; no night, no IFR, no dual. Every row a float row (OH-CTL or OH-GKT). Instructing:
  **11 rows on 6024**, **7 on 6025**. Landings blank on both → ours sum **74** and **65**.
  **⚠⚠ THE BOOK'S `LT` SUBSCRIPT IS REAL — first sighting.** Tiny handwritten `LT` beside the block
  minutes of **6024 rows 2–6** (all OH-CTL). **Absence still means nothing** — 6023 has none and its
  rows are local. Positive evidence only. **⚠ And the `LT` mark corrupts the digit before it:**
  6024 rows 2 and 5 both read `…29` and were both **exactly 9 min short**; the truth is `20`+`LT`
  (19:20→20:50 = 1:30; 17:20→19:15 = 1:55), **confirmed by an electronic record the user supplied**
  (01/07/2024 OH-CTL off **16:20Z** / on **17:50Z**, 90 min block, 9 ldg, *Opettaja* Rami Ayoub /
  *Oppilas* **Ignaty Romanov-Chernigovsky**). 16:20Z + 3h = 19:20 local.
  **Time zones: user says ALL 30 rows on both spreads are LOCAL** (no `Z`) — including OH-GKT, a
  departure from the 6022/6023 per-aircraft rule. **Ask per spread; don't carry the rule forward.**
  **6024 row 11's instructor entry is struck out** by the pilot (excluded; the 11:47 total needs it out).
  New places **Pulkkilanharju** (Päijänne) and **Mäntyharju**.
- **Twentieth & twenty-first spreads done:** `IMG_6026` (pages 39–40, 15 flights **21/07/2024–06/08/2024**)
  and `IMG_6027` (pages 41–42, 15 flights **07/08/2024–19/09/2024**), verified & appended 2026-08-01.
  **Both SIDEWAYS — CCW `rotate(90)`** (6022–6027 now all sideways). No `LT` subscripts on either.
  All cross-checks exact (6026: Δtotal 18:33/Δpic 16:51/Δstudent 1:42/Δinstr 7:21;
  6027: Δtotal 12:22/Δpic 12:22/Δinstr **7:41**). Landings blank on both → ours sum **34** and **61**.
  **⚠ TIME ZONES SWITCH BACK TO UTC — all 30 rows carry `Z`.** Five rows cross-checked at exactly +3h
  against the user's electronic records, and the 19/09 OH-GKT row matched an **Aviatron entry stamped
  `14:04:00 UTC` digit-for-digit with the paper cell**. The zone genuinely flips per spread
  (6024/6025 local → 6026/6027 UTC) — *always re-establish it.*
  **Student row** 02/08 SR20 OH-ESR EFNU→EFHV 1:42, Dual + SE-IFR, remark `(TAR)` = **Tarhanen**,
  a third IR revalidation. **⚠ 6027 row 7** (30/08 **OH-MIL** Maule Tuusulanjärvi→Lohja): the book writes
  0:44 in SE-VFR/PIC/FI but block, running column and both page totals need **0:41** (user-confirmed) —
  so the book's **FI page total 7:44 is 3 min high** and our Instructor drift moves **−1:20 → −1:23**.
  That row's pilot-name cell has **`SINERVÄ` struck through** over Ayoub — it is an **instructing** row
  (PIC + FI), not a dual. **Three on-block cells corrected** from the user's electronic records
  (06/08 Haikko →`16:53`, 14/08 →`16:50`, 20/08 →`17:55`) — see drift.md; **all three were on-block,
  none off-block.** New places **Haikko** (coast nr Porvoo), **Lohja**, **Asikkala**.
  New pupils: Tommi Nirkkonen, Ilkka Korkiakoski, Ivan Siragusa.
- **Twenty-second & twenty-third spreads done:** `IMG_6028` (pages 43–44, 15 flights
  **18/09/2024–04/01/2025**) and `IMG_6029` (pages 45–46, 15 flights **04/01/2025–15/04/2025**),
  verified & appended 2026-08-01. **Both SIDEWAYS — CCW `rotate(90)`** (6022–6029 now all sideways).
  No `LT` subscripts. All five cross-checks exact (6028: Δtotal 13:50/Δpic 13:50/Δinstr **2:32**;
  6029: Δtotal 17:14 = SE-VFR 7:18 + SE-IFR 9:56 /Δpic 17:14). Landings blank → ours sum **48**, **28**.
  **🍂 THE FLOAT SEASON ENDS ON p.43** — from 17/10/2024 it is all EFHV/EFRY landplane circuits plus
  SR20 IFR cross-countries; **6029 has zero instructor time and zero dual**, a first for Book 3.
  **⚠⚠ THE `Z`→LOCAL SWITCH HAPPENS INSIDE IMG_6028**, pinned from both sides by club records:
  17/10/2024 (rows 8–10) is **UTC** (club local −3h), but 10/12/2024 and 04/01/2025 (rows 14–15) are
  **LOCAL** (book = club times digit-for-digit). Rows 11–13 (28/10 ×2, 23/11) fall in the gap with no
  record → left `Z` per user. **All of IMG_6029 is LOCAL** (four rows club-confirmed to the minute).
  **Row 1's date was scribbled over — user confirms `18/09/2024`.**
  **⚠ 04/02/2025 is `OH-CMU`, not OH-CMV** — I misread the last letter; the club record settled it.
  **⚠ 07/03/2025 SR20 `EHGG→EDWF` (Groningen→Leer) is genuine** even though it doesn't connect to the
  ESMG legs either side — user: *"we were 2 pilots so I only logged my legs."* **Multi-pilot ferry
  trips produce geographically disconnected rows; don't treat them as errors.**
  Three inferred on-block cells (r3 →`10:18`, r5 →`16:11`, r9 →`11:21` from the club record) — see
  drift.md; **all on-block again**. New airports **ESMG** (Feringe, Sweden), **EHGG**, **EDWF**.
- **Twenty-fourth & twenty-fifth spreads done:** `IMG_6030` (pages 47–48, 15 flights
  **18/04/2025–29/05/2025**) and `IMG_6031` (pages 49–50, 15 flights **29/05/2025–02/07/2025**),
  verified & appended 2026-08-01. **Both SIDEWAYS — CCW `rotate(90)`** (6022–6031 all sideways).
  **All seven cross-checks exact** (6030: Δtotal 13:59/Δpic 13:21/Δstudent **0:38**/Δinstr **6:12**;
  6031: Δtotal 11:05 all SE-VFR/Δpic 11:05/Δinstr **7:27**). Landings sum **43** and **55**.
  **🌊 THE 2025 FLOAT SEASON OPENS 18/05/2025** (OH-CTL Räyskälä→Tuusulanjärvi) — back to the
  familiar Tuusulanjärvi instruction pattern; 6031 is 10/15 OH-CTL float rows.
  **⚠⚠ THE CLUB FILE CARRIED THIS BATCH** — 23 rows matched on times and **landings matched 23 of
  23** rows that have a record; every FI row was corroborated by `rami_role=instructor` + pupil name
  (Salo, Storgårds, Nirkkonen, Puhakka, Kere). **Grep it first, always.**
  **⚠ ZONES FLIP WITHIN A SPREAD AND BACK IN THE NEXT:** 6030 rows 1–14 `Z`, **row 15 local**;
  6031 **row 1 local**, rows 2–8/11/13/14 `Z`, **rows 9–10 (08/06) local**, rows 12/15 `Z`.
  All user-approved. Derive the zone **row by row from the club file**, not per spread.
  **⚠ 6030 row 7 is a DUAL row with `pic_name` left BLANK** — 19/05 OH-CTL Tuusulanjärvi→Hirvijärvi
  0:38: Dual column filled, **PIC blank**, but the name cell still says `AYOUB` (habit, not
  evidence). Absent from the club file → filed under the other pilot: a seaplane spring
  check with **PIC = Sinervä** (user-supplied 2026-08-01, filled in after the append — his fifth
  seaplane dual with the user). **`OH-COF` returns** (C152, EFNU, instructing —
  an existing Book-1/2 reg, not a new one).
  6031 r12 stored **on-block 17:53** (the 5-min water-taxi pattern). New places **Hirvijärvi,
  Loppijärvi, Kytäjärvi**.
- **Twenty-sixth & twenty-seventh spreads done:** `IMG_6032` (pages 51–52, 15 flights
  **01/07/2025–21/07/2025**) and `IMG_6033` (pages 53–54, 15 flights **21/07/2025–15/08/2025**),
  verified & appended 2026-08-01. **Both SIDEWAYS — CCW `rotate(90)`** (6022–6033 all sideways).
  No `LT` subscripts. **All 30 rows UTC (`Z`)** — every covered row sits exactly 3 h behind local.
  All six cross-checks exact *after two corrections* (6032: Δtotal **13:47**/Δpic 13:47/Δinstr
  **7:01**; 6033: Δtotal **13:37**/Δpic 13:37/Δinstr **6:31**). Landings sum **40** and **65**.
  **⚠⚠ AVIATRON COVERS `OH-GKT` THROUGH 07/2026 — the docs badly undersold it.** It is *not* just
  the CB-IR/OH-PIF reference: it holds **every OH-GKT float row**, exactly what the club file can
  never reach. It arbitrated **8 of 8** GKT rows here. **Grep it alongside `laskukierros_flights.csv`,
  not after it** (`pdftotext -layout Aviatron.pdf`; compare on the header line's **BLOCK/LASK**, the
  `RIVI` line under it is airborne time).
  **⚠ 6032 row 11 — a real book error, not a misread:** 08/07 OH-GKT Kahvisaari→Kelvenne written
  `16:36`/**0:57**/**7 ldg** and carried through the running column *and* page total; Aviatron 35472
  says `15:53–16:30`, **0:37**, **4 ldg** (user: *"a real mistake, Aviatron is authoritative"*).
  **0:57 / 7 ldg is exactly the preceding OH-GKT row** (17/06/2025, IMG_6031 r12) — *figures copied
  down a row; suspect that whenever time AND landings both duplicate the previous same-reg row.*
  That same Aviatron record **retro-validates IMG_6031 r12's inferred on-block 17:53.**
  **⚠ 6032 row 13 — fifth running-Total slip:** a 1:38 flight added as 0:38, so the book's printed
  page Total/PIC 13:07 are 1:00 low. Caught because the **directly-summed FI column hit 7:01 exactly**.
  **6033 row 8 on-block corrected** 17:28 → **17:32Z** (Aviatron 36429; its 1:01/6 ldg were right).
  **6032 row 14 = `09/07/2025`, entered out of order** (user-confirmed; single-digit 9).
  **⚠ `EFSA` was NOT new** — I called it new off `reference.md`'s airport list; Savonlinna is in the
  books 6 times since 2012. **That list is a hand note, not derived from the CSVs — grep the three
  CSVs before calling anything new.** Instructing: 8 rows on 6032, 6 on 6033.
- **Twenty-eighth & twenty-ninth spreads done:** `IMG_6034` (pages 55–56, 15 flights
  **15/08/2025–29/09/2025**) and `IMG_6035` (pages 57–58, 15 flights **30/09/2025–31/01/2026**),
  verified & appended 2026-08-01. **Both SIDEWAYS — CCW `rotate(90)`** (6022–6035 all sideways).
  **All eight cross-checks exact, ZERO corrections, ZERO block warnings** — every one of the 30 block
  times matched its logged time to the minute, a first for Book 3 (6034: Δtotal **13:11**/Δpic 12:32/
  Δdual **0:39**/Δinstr **2:14**; 6035: Δtotal **13:03**/Δpic 13:03/Δinstr **1:01** + Night 0:40).
  All 30 rows **UTC (`Z`)**. Landings **39** and **44**; **20 of 30 rows externally confirmed and
  landings matched 20 of 20.**
  **🎉 1000 HOURS PIC on 19/09/2025** (6034 r11, OH-CTL Kahvisaari→Tuusulanjärvi → **1000:20**).
  **⚠⚠ THE UTC OFFSET IS +3 IN SUMMER, +2 IN WINTER — and DST ends INSIDE IMG_6035** (26/10/2025).
  6035 r6 (20/10) reconciles at club-local −3h; r12/r13/r14 (Dec–Jan) at −2h. **A record that looks
  exactly 1 h "wrong" across the late-Oct or late-Mar boundary is daylight saving, not a bad row.**
  **6034 r1 = DUAL, `pic_name = Sinervä`** (user-confirmed): 15/08 **OH-MIL** Maule
  Tuusulanjärvi→Hiidenvesi 0:39 — the book actually *names the other pilot* in the PIC cell
  (SINERVÄ, not AYOUB) and fills Dual; his 7th seaplane dual, 2nd on the Maule.
  **⚠ 6034 r6 — the book wrote the AIRBORNE times into the off/on-block cells** (08/09 OH-CTL
  Inkoo→Tuusulanjärvi). User: entry is correct as flown; **store the club BLOCK times.** Now the
  only row in any book with `Block_Time (0:45) ≠ Total_Time (0:38)` and the first to use
  `Takeoff`/`Landing`. **Total_Time untouched, so no Δ or cumulative moved.**
  **6034 r9–r11 (19/09) = another two-pilot ferry day** — OH-CTL moves Pyhäjärvi→Kahvisaari with no
  logged leg while he flies OH-GKT home for the winter; all three rows corroborated as written.
  **6035 r12 (12/12) night 0:40 with the 3 landings correctly in the NIGHT column** (club-confirmed).
  **2025 float season closes 06/11**; the book crosses into **2026** on 6035 r13.
  New pupil **Koskinen**; new place **Vuolenkoski** (only genuinely new one — the rest were grepped
  against all three CSVs first). ⚠ **`EFPR` ≠ `EFPO`** — both are real and both are in the books.
- **Thirtieth & thirty-first spreads done — ALL PHOTOGRAPHED SPREADS ARE NOW TRANSCRIBED:**
  `IMG_6036` (pages 59–60, 15 flights **06/03/2026–11/05/2026**) and `IMG_6037` (pages 61–62,
  **14** flights **15/05/2026–03/06/2026**), verified & appended 2026-08-01. **Both SIDEWAYS — CCW
  `rotate(90)`** (6022–6037 all sideways). No `LT` subscripts. **Zero block-vs-total warnings on all
  29 rows** — every block pair matched to the minute (second such batch running). 6037 reconciled
  **exactly** (Δtotal/Δpic **11:16**, FI **6:32**); landings **30** and **56**, both landing cells blank.
  **⚠ 6036 is 1 min short in the book:** row 13 (01/05 OH-PDP EFHV local) is **0:24** by its block
  cells *and* its own SE-VFR/PIC cell, but the running column adds 0:23 and both page totals inherit
  it (12:17 vs our 12:18). **User: keep 0:24** → Total & PIC drift each move +0:01. Sixth running-Total slip.
  **⚠ 6036 rows 12–15 are written `/25` — user-confirmed typo, they are 2026**; rows 13/14 (01/05,
  03/05) also carry margin arrows and are **entered out of order** after the 08/05 row.
  **⚠⚠ 6036 is stored LOCAL throughout (user), which CONFLICTS with its one club record** —
  12/04/2026 OH-CAM is club `09:58–12:10` vs book `06:58–09:10` (exactly −3 h, i.e. UTC), landings
  3 = 3. Flagged; user ruled the page local anyway. Documented in drift.md, no total affected.
  **6037 zones are mixed FOUR ways:** rows 1–4 `Z` (club +3 h), rows **5–7 local** (club matches
  digit-for-digit), row **8 local**, row **9 `Z`**, rows **10–14 local**. Rows 8/9 overlapped as
  written; user resolved from memory — *"I flew the Maule first then drove to fly CDK."*
  **🌊 2026 float season opens 15/05/2026** (OH-CTL EFRY→Tuusulanjärvi), a fortnight earlier than 2025.
  **New airport `EFOP` (Oripää)**, **new place `Ojakkala`** (Vihti, on Hiidenvesi — 01/06 C185 OH-CDK
  float-instruction local). New pupils **Toivo Huovinen, Harry Karlsson, Thomas Hansson**; 18/05 has
  **Mikko Sinervä as the *pupil*** on an *Opekertaus* proficiency check. 6037 r3 has **no club record**
  — the pupil forgot to log it; the user will have them enter it. 6037 r8's type is written `M2`;
  user confirms **M6**. 6037 r4 landings book 4 vs club 3 — **book correct** (user).
- **⚠⚠ MISSING-FLIGHTS BATCH APPENDED 2026-08-01 — 15 rows that are NOT YET IN THE PAPER BOOK.**
  First time the project ran in reverse: the user **dictated** 15 flights (**12/06/2026 → 30/07/2026**)
  that he had never written down, we appended them to `logbook_3.csv` first, and **he transcribes them
  onto paper from our table.** No page image, no "TOTAL THIS PAGE", no page cross-check — the `paper`
  block in **`batch_missing_2026.json`** is empty. **His plan: row 1 fills the last free row of
  page 62, the other 14 open pages 63–64, and he then rewrites the book's carried totals to match
  ours to close the standing drift.** So the next spread photo should show *corrected* "TOTAL
  PREVIOUS PAGES" figures — **that is intentional, not a book error.** Full account in `drift.md`.
  - **Aviatron confirmed 5 OH-GKT rows to the minute and fixed 2** — the user's dictated on-block
    pairs had **slipped one row down** (26/06 carried 13/06's, 11/07 carried 26/06's).
  - **⚠ OH-CAM EFHV local is `29/06/2026`, not the 26/06 dictated** (user took the club record;
    26/06 was *also* a coherent chain, so only the record settled it).
  - **⚠ 25/07/2026 OH-TIL stored `Z` per the user, but the club file — whose times are LOCAL — holds
    `06:56–08:04` digit-for-digit.** No Aviatron OH-TIL row after 2021. Conflict logged, not resolved.
  - **Two 23/06/2026 OH-CTL instructing rows** (pupil **Pekka Puhakka**) came from the club file, not
    the user's list; added at his instruction. **30/06/2026 OH-CTL has no record in either reference.**
  - **⚠ `Takeoff`/`Landing` are now populated on all 15 rows** (the user dictated both pairs).
    `Off_Block`/`On_Block` + `Block_Time` = block pair, `Takeoff`/`Landing` = airborne,
    **`Total_Time` = the block time** as always. `logbook_tools.py` gained optional
    **`takeoff` / `landing` / `block`** batch fields for this.
- **Last row in `logbook_3.csv`:** `30/07/2026 · C172 · OH-GKT · Kahvisaari → Kahvisaari ·
  15:40Z–16:40Z (block) · 15:45Z–16:31Z (airborne) · Total 1:00 · PIC 1:00 · Instructor 1:00 ·
  7 landings` — **not yet on paper.**
- **Cumulative totals at that row (our continuous series, seeded from Book 2):**
  - Cumulative_Total **1219:35** · Cumulative_PIC **1053:03** · Cumulative_Student **166:32**
  - Cumulative_Instrument **107:05** · Cumulative_SEP_Sea **407:39**
  - Cumulative_Landings **3439** (= day+night; runs ahead of book's day-only count — see drift.md)
  - Cumulative_Instructor **189:41**
- **At the end of paper page 62** (i.e. after only the first missing row, 12/06 OH-ESR EFNU→EFIK):
  Total **1206:58** · PIC **1040:26** · Instrument **105:57** · Student **166:32** ·
  Instructor **185:50** · Landings **3394**.
- **`logbook_3.csv` has 478 data rows** (+ header + seed row = 480 lines).

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

## 🏁 PAPER DRIFT CLOSED AT PAGE 62 (2026-08-01)
The user inked our figures at the bottom of p.62: **Total 1206:58 · SE-VFR 1101:01 · SE-IFR 105:57 ·
PIC 1040:26 · FI 185:50 · Dual 166:32 · Day ldg 3335 · Night ldg 59** (night time 22:45).
**All paper-vs-ours drift is zero from p.62 on** — every historical offset (Total +2:06, PIC +1:46,
SE-IFR +4:02, Dual +0:20, FI −1:23) was corrected by hand. **Do not re-apply them.** Full
decomposition in `drift.md`.
⚠ **Two of those inked cells are now stale — the day/night LANDING split only.** Since the night
column closed, the split recomputes to **68 night / 3326 day**. **Night *time* 22:45 is confirmed
correct and our column now equals it exactly**, and the landing **sum** (3394) never moved. So the
p.62 ink needs two numbers changed, and nothing in the CSV does.

## Next action — TWO THINGS, BOTH WAITING ON THE USER

**1. ✅ THE DATA IS RECONCILED — nothing to photograph, nothing to re-validate.** Night = **22:45**,
equal to the paper; all seven cumulative series reconcile with **zero breaks**. Closed 2026-08-01 by
the p.52/53 photograph (`IMG_6048`) plus the line-28 fix. ⚠ **Never infer night from clock times or
sunset** — the book's night column is the only authority (user, 2026-08-01).
Only two paper-side items remain, and neither moves a total:
- **`logbook_2_final.csv` lines 89–90** — the two `04.05.2018` dates. Needs the physical page.
- **The p.62 inked landing split** `59/3335` → recomputed **`68/3326`**. **Correct the paper.**

Both are written up in `drift.md` (item 2 and the landing-split section).
**Do not re-validate the book on spec — the totals match.** To re-prove it in one command:
`cd app/backend && go run ./cmd/logbookctl import -dry-run -csv ../..`
*(⚠ `go` was not installed on the user's machine as of 2026-08-01 — the same seven-series
reconciliation can be done directly off the CSVs in Python if the toolchain is missing.)*

**2. ⏸ THE CSV IS AHEAD OF THE PAPER (Book 3).**
**Every photographed spread is transcribed** (`IMG_6007`–`IMG_6037`, pages 1–62) **and the CSV now runs
15 flights *ahead* of the book** (the missing-flights batch above, 12/06/2026–30/07/2026). The user is
writing those 15 onto paper — 1 row closes page 62, 14 open pages 63–64 — and **rewriting the book's
carried totals to our figures to close the drift.** Nothing to process until he photographs the result.

**When the pages 63–64 photo arrives it is a VERIFICATION pass, not a transcription pass.** Compare the
image against the 15 rows already in `logbook_3.csv`; the paper is copied *from* us this time, so a
mismatch means a writing slip on paper, not a bad CSV row. Expect the page totals to be:
**Δtotal 12:37** (= SE-VFR 11:29 + SE-IFR 1:08) · **Δpic 12:37** · **Δinstr 3:51** · **Δland 45**
for the 14 rows on the new spread (page 62's single added row is Total/PIC/FI 0:45, 3 ldg).

**⚠ Book 3 is NOT finished — do not rename `logbook_3.csv` to `logbook_3_final.csv`.** The paper book
runs to page 128 and we are at **p.62**. The Book-1/2 `_final` rename only applies to a *closed* book.

**Nothing else is known-flown-but-unrecorded.** Both references are fully consumed:
`laskukierros_flights.csv` ends **25/07/2026**, Aviatron ends **12/07/2026**, and every row in either
that postdates the paper book is now in the CSV. Re-pull them before assuming that still holds.
- **⚠ 6037 r3 (16/05/2026 Kabböle local 0:50) is missing from the club file** — the pupil forgot to
  log it and the user is chasing them; it may appear in a future pull. Our row is correct as is.

### When new photos arrive
**Check orientation first** — it is NOT consistent across Book 3: 6007–6011 needed CCW `rotate(90)`,
6012–6021 were already upright, **6022–6037 have all been sideways (CCW `rotate(90)`)**. Never assume —
thumbnail first (`Image.open(p).resize((1024,768))`), then crop the two pages at high res
(original ~2048×1536). **Also scan the block-minute cells for the tiny `LT` subscript** (see below).
Transcribe all rows, cross-check via the book's "TOTAL THIS PAGE" using `--csv logbook_3.csv`, and
surface flags. **If the newest page has no page totals filled in, cross-check on the running-Total
column's first→last delta instead** (that worked exactly on IMG_6037: Δ 11:16 over 14 rows) — but
remember the running column has slipped six times, so treat a mismatch as a row to re-read, not proof.
**⚠ DST: late-Mar 2026 the club/Aviatron local↔UTC offset goes +2 → +3.**
**Time zones can be mixed within one spread** (IMG_6014, 6018, 6022, 6030, 6031, and **6037 four ways**)
— when rows overlap or run out of order, suspect a local-vs-UTC mix and ask the user *which rows*.
On IMG_6037 the answer was **per-row memory** ("I flew the Maule first then drove to fly CDK"), not a
rule; on IMG_6022 it was a per-aircraft rule that still needed a per-row exception. Always walk each
day's flight chain for ordering conflicts before applying anything wholesale.
**Hybrid-batch pace works well:** transcribe 2–3 spreads/pass, tool-reconcile each, present ONE digest
that greenlights clean pages and surfaces only flagged rows (student/instructing/night/odd-time/landing
anomalies) for user sign-off before append. Watch for: SR20 OH-ESR (a PIC type, flown IFR on
cross-countries), Stude (SR20 instr), **Sinervä** (seaplane instr — 7 duals, incl. two on the Maule;
but on 18/05/2026 he is the *pupil* on an *Opekertaus* check), Salo, night landings, and the day+night
landings convention.
**Seasonality:** May–Sep the pilot float-instructs (Kabböle/Tuusulanjärvi/Kahvisaari/Ojakkala on
OH-CTL/OH-GKT/OH-MIL/OH-CDK) — clusters of short lake locals with high landing counts, logged PIC +
Instructor; **Oct–Mar it switches to EFHV/EFNU landplane circuits** (OH-CGX/OH-CAY/OH-AWB/OH-CMU/
OH-PDP/SR20), with occasional night flights. Float seasons so far: **2025** 18/05 → 06/11;
**2026** opens **15/05**. The next spread should be deep in the 2026 float season.

### Hard-won reading lessons (all cost a round-trip; don't relearn them)
- **⚠⚠ THERE ARE TWO ELECTRONIC REFERENCES AND THEY COVER DISJOINT FLEETS. GREP BOTH.**
  `laskukierros_flights.csv` = **club** aircraft (CTL/CAM/CAY/CGX/CMU/COK/AWB/TIL).
  `Aviatron.pdf` = the **pilot's own / Blue Skies** aircraft — **`OH-GKT`**, `OH-PIF`, `OH-DBS`,
  `OH-TIL`, `OH-DBE` — **126 flights running to 07/2026**. Together they cover nearly every row of
  Book 3's float season. I spent IMG_6022–6031 treating Aviatron as a stale CB-IR artefact and
  telling the user "OH-GKT isn't in the club file, so there's no record" — it was in the repo the
  whole time. On IMG_6032/6033 it arbitrated **8 of 8** GKT rows and exposed a real book error.
  Extract with `pdftotext -layout Aviatron.pdf`; each flight is a header line
  (`ID / LÄHTÖP / LASKUP / OFF / ON / BLOCK / LASK / … / PIC`) followed by a `RIVI` line —
  **compare on the header's BLOCK/LASK; the RIVI line is airborne time and will read ~5 min short.**
- **⚠ THE USER CAN OVERRULE THE ELECTRONIC RECORD ON TIME ZONES — ASK, THEN DO AS TOLD.** On
  IMG_6036 the one club-covered row (12/04/2026 OH-CAM) sits exactly −3 h from the club's local
  times, which reads as UTC; the user ruled **the whole page local** anyway. Paper stays
  authoritative, the conflict goes in `drift.md`, and no total is affected. **Present the evidence
  once, then store what the user says** — don't re-litigate it.
- **⚠⚠ THE LOCAL↔UTC OFFSET IS SEASONAL: +3 (EEST) SUMMER, +2 (EET) WINTER.** DST flips in late
  March and late October, and it flipped *inside* IMG_6035 (26/10/2025). **A club or Aviatron row
  that looks exactly 1 hour "wrong" near those boundaries is daylight saving, not a mis-logged row.**
  Every prior spread happened to sit inside a single season, which is why this only surfaced now.
- **⚠ A SHORT ROW MAY MEAN THE BOOK USED THE AIRBORNE CLOCK FOR BOTH CELLS.** The known pattern was
  a single bad cell (the on-block holding the landing time). IMG_6034 r6 is different: **both** cells
  are the record's *takeoff/landing* pair, not its off/on-block pair. **Check the record's airborne
  times as well as its block times before proposing a fix.** Storage convention for such a row is in
  `reference.md` (block pair → `Off_Block`/`On_Block`, airborne pair → `Takeoff`/`Landing`,
  **`Total_Time` never changes**).
- **⚠ THE BOOK ITSELF CAN BE WRONG ON A ROW, NOT JUST MISREAD — AND IT COPIES DOWN.** IMG_6032 r11
  was written 0:57 / 7 landings and carried through the running column *and* the page total, so it
  was internally consistent and passed every arithmetic check; Aviatron says 0:37 / 4. **0:57 / 7 is
  exactly the previous OH-GKT row's figures.** *When a row's time and its landing count both
  duplicate the previous same-registration row, suspect a copy-down before trusting it —
  self-consistency inside the book proves nothing.*
- **⚠ `reference.md`'s "airports/places seen" and registration lists are hand-maintained running
  notes, NOT derived from the CSVs.** I announced EFSA as a new airport off that list; Savonlinna is
  in the finished books **6 times since 2012**. **`grep` `logbook_1_final.csv`, `logbook_2_final.csv`
  and `logbook_3.csv` before calling any place, airport or registration new.**
- **When a block time doesn't match the logged flight time, do NOT assume which cell is wrong.** On
  IMG_6016 one row's off-block was correct and one row's on-block was correct — guessing would have
  gotten one backwards. Present both readings and let the user pick.
- **⚠⚠ CHECK `laskukierros_flights.csv` BEFORE ASKING THE USER TO GUESS.** On IMG_6026/6027 three
  rows had block-vs-total gaps. I offered candidate readings instead; the user pasted the club-system
  records and **all three were settled outright** — and one (06/08 Haikko) proved the candidate the
  user had already picked was **wrong**. Every one of those records is now in the repo
  (228 flights, through 25/07/2026). **Grep it first; offer readings only if the row isn't there.**
- **⚠ The pilot sometimes writes the LANDING time into the on-block cell** (proved 20/08/2024: record
  says landing 20:50 / on-block 20:55, book cell says 17:50Z = the landing time). He taxis ~5 min on
  the water at Tuusulanjärvi. **A row short by ~5 min with clean-looking cells is usually this, not a
  mangled digit.** All three IMG_6026/6027 corrections were **on-block**, none off-block.
- **Use `laskukierros_flights.csv` (228 rows), NOT the old `laskukierros_export.csv` (128).** The old
  export returns only flights where the user is the *primary* pilot, so all **100 instructing rows**
  — the bulk of Book 3's float season — were missing. `GET /api/v1/flights` has them; they are filed
  under the pupil's account. **`rami_role`** tells you which; **`other_name`** gives the pupil.
  ⚠ The `func_*` flags describe the *primary* pilot, so `func_student=true` on rows our pilot
  *instructed* — never map them straight onto `Student_Time`.
- **Check every row for a strike-through before transcribing.** IMG_6019 row 2 was ruled through and
  was a *duplicate of the previous spread's last row*. The page total only reconciles with it excluded.
- **The book's per-row running-Total column is unreliable** (four slips so far: p.14 −1:00, p.21 −1:52,
  p.23 −0:36, p.26 PIC −2:30). Its **"TOTAL THIS PAGE"** and bottom-line totals are still good.
  **Cross-check only on those.** When a total cell is scribbled illegible, look for a **margin note** —
  on IMG_6019 the pilot wrote "\* 15:04 ←" and that was the only legible source.
- **Registrations: read the last letter carefully.** `OH-CMU` ≠ `OH-CMV` (both C152, both real).
  `OH-CGX`'s X often looks like a T — there is no OH-CGT. **I got this wrong on IMG_6029** (read
  `OH-CMV` for a row the club file proves is `OH-CMU`) — when the flight is in a **club** aircraft,
  grep `laskukierros_flights.csv` by date+time and let it arbitrate the letter.
- **⚠ A row that doesn't connect geographically is not automatically an error.** IMG_6029 has
  `EHGG→EDWF` (Netherlands→Germany) stranded between two Sweden legs. User: *"we were 2 pilots so I
  only logged my legs in my book."* **On multi-pilot ferry trips only his own legs are in the book** —
  ask before hunting for a misread airport code.
- **⚠ The UTC-vs-local convention can switch *mid-spread*, not just between spreads.** On IMG_6028
  the club file proves rows 8–10 (17/10/2024) are UTC and rows 14–15 (10/12/2024, 04/01/2025) are
  local. **Cross-check the first and last club-covered row of every spread**, not just one.
- **A C172 logging single-engine IFR is not a misread — `OH-CAM` is IFR-certified** (user, IMG_6021).
- **`C172sea` in the type column is not a type** — it is the pilot's own informal seaplane marker
  (from IMG_6022 on). Store plain **`C172`**; the SEP_Sea flag comes from the registration.
- **Impossible day digits happen.** IMG_6022 row 12 read "34/05/24" → user says **30/05/2024**. Bracket
  the row by its neighbours' dates and offer the candidates rather than guessing silently.
- **A night flight's landings may still be written in the DAY column** (IMG_6020 row 15). If the clock
  times make daylight impossible, they're night landings — say so.
- **⚠⚠ THE BOOK'S `Z`/UTC MARKING IS UNRELIABLE — proved 2026-08-01 against `laskukierros_export.csv`**
  (new second electronic cross-reference in the repo; club aircraft only — CTL/CAM/CGX/CAY/COK/CMU/
  AWB/TIL — **no OH-GKT, no Maule, no PDP/PIF/ESR**; see `reference.md`). **Its times are LOCAL.**
  Of 82 rows matching ours: **52 are genuinely local (41 of them wrongly carry a `Z` in our CSVs)**
  and **30 are genuinely UTC (3 missing their `Z`)**. **User decision: document only, do NOT change
  the CSVs** — paper stays authoritative; full list in `laskukierros_zflags.md`. No total affected.
  **Practical upshot: keep asking the user about time zones per spread — the book's `Z` is advisory,
  not evidence.** Also documented: 3 record conflicts (2 dates, 1 registration) — see `drift.md`.
  **~35 rows of `laskukierros_flights.csv` postdate 02/07/2025, so it forward-checks IMG_6032 onward**
  (club regs only — `OH-CTL`, `OH-CAM`, `OH-CAY`, `OH-CGX`, `OH-CMU`, `OH-TIL`; no GKT/ESR/PDP/MIL/COF);
  the next two are `01/07/2025 OH-CTL Tuusula↔Hiidenvesi` (instructing, Puhakka) and the `02/07/2025
  OH-CAM EFHV→EFJO→EFNU→EFHV` day. **It is the single most valuable tool in this project** — on
  IMG_6028/6029 it settled the UTC→local switch date and corrected a registration I misread
  (`OH-CMV` → `OH-CMU`); on IMG_6030/6031 it confirmed **23 of 23** landing counts and every
  instructing row. **Grep it before offering the user candidate readings.**
- **⚠ The book's tiny `LT` subscript (first seen IMG_6024) is positive evidence a row is LOCAL — but
  its absence proves nothing** (IMG_6023 has none and is local throughout). **And it corrupts the digit
  it follows:** `20`+`LT` read as `29` on two IMG_6024 rows, each putting the row 9 min out. *When two
  rows on one page are short by the same odd amount, suspect one systematic digit misread.*
- **Paper-vs-ours drift, refreshed at end of page 60 (11/05/2026 boundary; EASA "TOTAL" bottom-of-page):**
  book Total **1192:51** vs ours **1194:57** (**+2:06**); book PIC **1026:39** vs ours **1028:25** (**+1:46**);
  book SE-IFR **101:55** vs our Instrument **105:57** (**+4:02**); book Dual **166:12** vs our Student
  **166:32** (**+0:20**); book Flight-Instructor **179:56** vs ours **178:33** (**−1:23**).
  **Total and PIC each moved +0:01 at p.60** (the IMG_6036 row-13 decision — the book's page total is
  1 min low); Instrument, Student and Instructor **unmoved since p.54**. **No drift refresh is
  possible at p.62** — the last page carries no totals. (The base +1:25 dates from the IMG_6013
  running-Total slip, our value correct;
  the +1:05 flipped from −1:25 on IMG_6019 where the pilot struck his own PIC total 810:09 → 807:39,
  a hand −2:30 correction he doesn't recall the reason for.) The book's SE-IFR line was itself a
  pilot correction written *below* the p.40 total box (struck 86:55 → 88:04); it carries forward correctly.
  Landings: ours **3117** (day+night) runs ahead of the book's day-only cumulative (see drift.md).
- **The book's per-row running-Total column is now unreliable — three separate slips** (p.14 **−1:00**,
  p.21 **−1:52** which the pilot caught and fixed at the page total, p.23 **−0:36**). Its printed
  "TOTAL THIS PAGE" and bottom-line totals are still trustworthy. **Cross-check only on those.**
  **Always cross-check on offset-independent per-page Δ ("TOTAL THIS PAGE"), never absolute totals — and
  note the book's Total column is itself 1:00 low from p.14 on.**
- **Remote:** `origin` = git@github.com:ramiayoub-priv/logbook-migration.git. **All work through the
  missing-flights batch and the p.62 drift close is committed AND pushed at `46235ec`
  (2026-08-01)** — master is clean and up to date with origin. (Book-3 work through IMG_6037 was
  `3acb6a7`.) Images/HEIC/zip are gitignored (not pushed). Git identity is set repo-locally
  (`Rami Ayoub <rami.ayoub@gmail.com>`) — it was missing once and blocked a commit.
  **Commit and push only when asked.**

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
- **Night column: 1:55 outstanding, in Book-1 pages 52–69** (Mar–Aug 2014, not yet photographed).
  See "Next action" above and `drift.md` item E. Everything else in the night column is reconciled.
- **`logbook_1_final.csv` line 28** — `Instrument_Time` 1:21 on a 1:12 flight; recommended fix 1:12,
  moves no total, but the app cannot agree with the paper on instrument time until it is made.
- **`logbook_2_final.csv` lines 89–90** — the two `04.05.2018` dates; needs the paper.
- **⚠ The p.62 day/night landing split (59 night / 3335 day) is STALE** — it was computed from the
  night rows before the 2026-08-01 reconciliation. Recompute once the night column closes;
  `Cumulative_Landings` (the sum) is unaffected, only the split.
- **⚠ `SE-GKT` (Book 1, 2015–16) and `OH-GKT` are the same airframe** re-registered — never treat
  them as two aircraft.
- Landings drift: the paper book's cumulative landing count has historically run ahead of the
  true count. Cross-check landing sums when in doubt. See `drift.md`.
