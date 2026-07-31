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
- **Last row in `logbook_3.csv`:** `05/03/2024 · C152 · OH-CMU · EFHV → EFHV ·
  19:32Z–19:55Z · Total 0:23 · Night 0:23 · PIC 0:23 · 3 landings (all night)`
- **Cumulative totals at that row (our continuous series, seeded from Book 2):**
  - Cumulative_Total **982:21** · Cumulative_PIC **819:40** · Cumulative_Student **162:41**
  - Cumulative_Instrument **84:46** · Cumulative_SEP_Sea **274:14**
  - Cumulative_Landings **2577** (= day+night; runs ahead of book's day-only count — see drift.md)
  - Cumulative_Instructor **99:49**
- **`logbook_3.csv` has 209 data rows** (+ header + seed row = 211 lines).

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

## Next action — Book 3, IMG_6021 (pages 29–30)
Process **`IMG_6021`** next (continues from 05/03/2024). **Check orientation first** — image
orientation is NOT consistent across Book 3: 6007–6011 needed CCW `rotate(90)`, but **6012–6020 have all
been already upright** in the original (no rotation) — still check, don't assume.
Quickest check: `Image.open(p).resize((1024,768))` and eyeball
which way is up; then crop the two pages at high res (the original is ~2048×1536). Transcribe all rows,
cross-check via the book's "TOTAL THIS PAGE" using `--csv logbook_3.csv`, and surface flags.
**Time zones can be mixed within one spread** (proved on IMG_6014) — when rows appear to overlap in
time, suspect a local-vs-UTC mix and ask the user *which rows*, not just whether.
**Hybrid-batch pace works well:** transcribe 2–3 spreads/pass, tool-reconcile each, present ONE digest
that greenlights clean pages and surfaces only flagged rows (student/instructing/night/odd-time/landing
anomalies) for user sign-off before append. Watch for: SR20 OH-ESR (now a PIC type, flown IFR on
cross-countries), Stude (SR20 instr), Sinervä + **Salo** (instructors), night landings, and the day+night
landings convention. **Seasonality:** May–Sep the pilot float-instructs (Kabböle/Tuusulanjärvi/Kahvisaari
on OH-CTL/OH-GKT/OH-MIL) — clusters of short lake locals with high landing counts, logged PIC +
Instructor; **Oct–Mar it switches to EFHV landplane circuits** (OH-CGX/OH-CAY/OH-AWB/OH-CMU) with
night flights. From 05/03/2024 we are at the start of a new float season.
Then continue IMG_6021…6037 at that pace (each spread is
~15 flights — sizeable, so 1–2 spreads per pass is plenty).

### Hard-won reading lessons (all cost a round-trip; don't relearn them)
- **When a block time doesn't match the logged flight time, do NOT assume which cell is wrong.** On
  IMG_6016 one row's off-block was correct and one row's on-block was correct — guessing would have
  gotten one backwards. Present both readings and let the user pick.
- **Check every row for a strike-through before transcribing.** IMG_6019 row 2 was ruled through and
  was a *duplicate of the previous spread's last row*. The page total only reconciles with it excluded.
- **The book's per-row running-Total column is unreliable** (four slips so far: p.14 −1:00, p.21 −1:52,
  p.23 −0:36, p.26 PIC −2:30). Its **"TOTAL THIS PAGE"** and bottom-line totals are still good.
  **Cross-check only on those.** When a total cell is scribbled illegible, look for a **margin note** —
  on IMG_6019 the pilot wrote "\* 15:04 ←" and that was the only legible source.
- **Registrations: read the last letter carefully.** `OH-CMU` ≠ `OH-CMV` (both C152, both real).
  `OH-CGX`'s X often looks like a T — there is no OH-CGT.
- **A night flight's landings may still be written in the DAY column** (IMG_6020 row 15). If the clock
  times make daylight impossible, they're night landings — say so.
- **Paper-vs-ours drift, refreshed at end of page 28 (05/03/2024 boundary; EASA "TOTAL" bottom-of-page):**
  book Total **980:56** vs ours **982:21** (**+1:25**, steady since the IMG_6013 running-Total slip — our
  value correct); book PIC **818:35** vs ours **819:40** (**+1:05** — ⚠ *flipped from −1:25 on IMG_6019*,
  where the pilot struck his own PIC total 810:09 and carried 807:39, a hand −2:30 correction he doesn't
  recall the reason for; our per-page Δpic reconciled exactly on both surrounding pages, so ours was kept);
  book SE-IFR **80:44** vs our Instrument **84:46** (**+4:02**, steady); book Dual **162:21** vs our
  Student **162:41** (**+0:20**, steady); book Flight-Instructor **101:09** vs ours **99:49** (**−1:20**,
  steady). Four of five steady across IMG_6019/6020; only PIC moved, by exactly the book's own adjustment.
  Landings: ours **2577** (day+night) runs ahead of the book's day-only cumulative (see drift.md).
- **The book's per-row running-Total column is now unreliable — three separate slips** (p.14 **−1:00**,
  p.21 **−1:52** which the pilot caught and fixed at the page total, p.23 **−0:36**). Its printed
  "TOTAL THIS PAGE" and bottom-line totals are still trustworthy. **Cross-check only on those.**
  **Always cross-check on offset-independent per-page Δ ("TOTAL THIS PAGE"), never absolute totals — and
  note the book's Total column is itself 1:00 low from p.14 on.**
- **Remote:** `origin` = git@github.com:ramiayoub-priv/logbook-migration.git. **master is pushed &
  fully up-to-date through IMG_6015 (`d851be5`) as of 2026-07-31 — working tree clean, nothing pending.**
  (IMG_6016 committed locally on top; push when asked.)
  Images/HEIC/zip are gitignored (not pushed). Git identity is now set repo-locally
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
