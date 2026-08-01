# Reference

## `logbook_2.csv` schema (26 columns, in order)
`Date, Aircraft_Type, Aircraft_Reg, Departure, Arrival, Off_Block, On_Block, Takeoff, Landing,
Block_Time, Total_Time, Instrument_Time, Night_Time, PIC_Time, Student_Time, Instructor_Time,
pic_name, Landings, Remarks, Cumulative_Total, Cumulative_PIC, Cumulative_Student,
Cumulative_Instrument, Cumulative_SEP_Sea, Cumulative_Landings, Cumulative_Instructor`

- All values are quoted. Dates are `DD/MM/YYYY`. Times are `H:MM`/`HH:MM`.
- **Time zone: transcribe times exactly as written in the paper book (a mix of local/UTC).**
  Do NOT convert while digitizing. A future app (to be built after all logbooks are entered) will
  normalize everything to UTC. So an apparent clock collision across two rows (e.g. a water landing
  at 13:40 then an airport departure at 13:55) is usually a local-vs-UTC artifact, not a real error.
  - **Zulu rows:** when the book writes a time with a `Z`/`z` (UTC), keep the **`Z` suffix** in
    `Off_Block`/`On_Block` (e.g. `07:56Z`). Local rows stay plain `HH:MM`. This is how the app tells
    already-UTC rows apart from local ones. `logbook_tools.py` handles the `Z` suffix.
- The **first data row** of each book's CSV is the final row carried over from the previous book's
  `_final.csv` — the seed for that book's cumulative totals (Book 2 ← Book 1; Book 3 ← Book 2).

## Book 3 — EASA logbook (active, `logbook_3.csv`)
Book 3 is a **newer EASA-format paper logbook** with a wider layout, but we keep the **same 26-column
CSV schema**. Photos: `logbook-3/IMG_6007–6037.JPEG` (each = one two-page spread). **Orientation varies —
check each image:** 6007–6011 are stored sideways (`rotate(90, expand=True)`); **6012 onward are
already upright**. Book pages numbered "Page N of 128".

**EASA columns → our schema mapping:**
- GENERAL: Date · Dep(place) · **Off-block (UTC)** · Arr(place) · **On-block (UTC)** · Type · Reg · PIC name.
- FLIGHT TIME: **Total** (written as a *running cumulative* — the per-flight time sits in the
  SE-VFR/SE-IFR column) · **Night** → `Night_Time` · **SE-VFR / SE-IFR / ME-VFR / ME-IFR** (single vs
  multi × VFR vs IFR; **SE-IFR → `Instrument_Time`**; ME cols unused so far) · **PIC → `PIC_Time`** ·
  **Co-pilot** (ignored; none yet) · **Multi-pilot** (none yet) · **Flight Instructor → `Instructor_Time`** ·
  **Dual (from student appendix) → `Student_Time`** (dual == Oppilas/student).
- OTHER: Instructor-in-STD (ignored) · **Landings Day / Night** — store the *sum* in `Landings`
  (night-landing split is inferred later from `Night_Time`, not stored) · Remarks.

**Conventions (locked 2026-07-31):**
- **Times are UTC → suffix `Z`** on Off/On block, UNLESS an entry is annotated **`LT`** (then plain
  local). The book is not 100% consistent; **flag suspicious times** (out-of-order/colliding).
  **A single spread can mix local and UTC rows** — confirmed on IMG_6014, where three same-day rows
  looked overlapping because one was UTC and two were local. When times collide, ask the user *which
  rows* are which; don't assume the whole page shares one zone.
- We continue **our own internally-consistent cumulative series** (seeded from `logbook_2_final.csv`),
  NOT the EASA book's printed "previous pages" totals. The book prints a per-page **"TOTAL THIS PAGE"**
  for every column — use those as the offset-independent cross-check (`d_total/d_pic/d_land/d_instr`).
- Tool: `logbook_tools.py <batch.json> --csv logbook_3.csv [--append]`.

## Seaplane registrations (count toward Cumulative_SEP_Sea)
These regs are seaplanes; their `Total_Time` adds to `Cumulative_SEP_Sea`. Pay extra attention:
`OH-CTL, SE-GKT, OH-GKT, OH-PAX, OH-MIL, OH-CTE, OH-CDK`

## Aircraft registrations seen so far in Book 2
Common: `OH-PDP` (P28A), `OH-CTL` (C172, seaplane), `OH-GKT` (C172 seaplane, ex-`SE-GKT`,
now owned by the user; in Aviatron), `OH-CDK`, `OH-CWB` (C172),
`OH-CAV`, `OH-DBS`, `OH-PAX`. C152: `OH-CRA`, `OH-COF`, `OH-CKO`, `OH-KLS`, `OH-NEU` (all landplanes; OH-NEU appears Mar 2020).
C172 landplanes: `OH-CAY`, `OH-CGX` (both appear 19/04/2020 at EFHV; OH-CGX flown as a student checkout).
P28A IR trainer: `OH-PIF` (used for CB-IR instrument training from Apr 2019, instructor Autere).
DA40: `OH-STL` (Diamond DA40, appears Apr 2019).
C185 floatplane: `OH-CDK` (Cessna 185 on floats, seaplane; Saimaa-lakes trip Jun–Jul 2019).

## Aircraft added in Book 3
- `OH-CAM` (C172), `OH-CMV` (C152) — from IMG_6009. **`OH-CAM` is IFR-certified** (user-confirmed
  2026-07-31) — it logs single-engine IFR time, e.g. 30/04/2024 EFTP→EFHV 1:01 = 0:36 VFR + 0:25 IFR.
  Don't treat a C172 with an SE-IFR entry as a misread.
- **`SR20` / `OH-ESR`** (Cirrus SR20) — type-rating training Apr–May 2022 (instructor Stude), flown
  **PIC from 18/05/2022**, routinely IFR on cross-countries.
- **`OH-MIL`** — **Maule on floats, type written `M6(sea)`; always on floats** (user-confirmed
  2026-07-31). First appears 24/08/2023 (IMG_6019) on a Sinervä seaplane dual. Counts SEP_Sea.
- **`OH-AWB`** (C152) and **`OH-CMU`** (C152) — appear Dec 2023 / Mar 2024 at EFHV (IMG_6020).
  ⚠ **`OH-CMU` and `OH-CMV` are two different aircraft** (user-confirmed) — both C152s, differing only
  in the last letter. Read that letter carefully.
- ⚠ **There is no `OH-CGT`.** On IMG_6020 the last letter of `OH-CGX` sometimes looks like a `T`;
  user confirmed all such rows are **`OH-CGX`**.

## laskukierros — second electronic cross-reference (NOT the source of truth)
**Canonical files (repo root, tracked): `laskukierros_flights.json` + `laskukierros_flights.csv`.**
**228 flights, 19/04/2020 → 25/07/2026** (2026 included: 13 rows, last 25/07/2026). Pulled
2026-08-01 from **`GET /api/v1/flights`** on laskukierros.fi, a Finnish club booking/billing system.
`laskukierros_to_csv.py` regenerates the CSV from the JSON — the JSON is the raw dump, edit neither
by hand. Never touch the site's add/edit/delete endpoints; GET only.

⚠ **`laskukierros_export.csv` (128 rows) is SUPERSEDED — do not use it for coverage questions.**
It is the older `GET /export/pilotFlights` dump and is kept only because `laskukierros_zflags.md`
was computed from it. **It returns only flights where the user is the PRIMARY pilot**, so all
**100** flights he *instructed* — filed under the pupil's account — are missing from it. That is the
whole 128-vs-228 gap, and it is why 06.08.2024 and 14.08.2024 appeared "absent" during IMG_6026/6027.
It is also **UTF-16** as served (the committed copy was converted to UTF-8).

**Fetching it again:** needs the user's `PHPSESSID` cookie, so ask them for it; then
`curl -H "Cookie: PHPSESSID=…" https://laskukierros.fi/api/v1/flights`. Never commit the cookie.

- **Coverage: club aircraft only.** **No `OH-GKT`, no `OH-MIL`, no `OH-PDP`/`OH-PIF`/`OH-ESR`/
  `OH-CDK`** — the pilot's own aircraft and the Blue Skies fleet are absent, so float rows on OH-GKT
  and the Maule still need the user. Within the club fleet the 228-row dump is *good* coverage
  (e.g. 71 rows postdate 19/09/2024: `OH-CTL` ×46, `OH-CAM` ×17, `OH-CAY` ×4, `OH-CMU` ×2,
  `OH-CGX`/`OH-TIL` ×1).
- **`laskukierros_flights.csv` columns:** `date, reg, model, dep, arr, block_start, block_stop,
  takeoff, landing, block_min, air_min, ldg_day, ldg_night, night_min, ifr_min, rami_role,
  other_name, other_role, func_*, flight_type, notes, id`.
- **⚠ Times are LOCAL, not UTC**, despite the `+00:00` the API stamps on every timestamp. Proved
  against the paper: 20/08/2024 `block_start 18:50` == paper `15:50Z`. See `drift.md`.
- **`rami_role`** is the useful column: `pilot` (128), `instructor` (97), `copilot` (3).
  **`other_name`** gives the counterpart — on instructing rows that is the **pupil**, which is how we
  now recover pupil names without asking the user (Nirkkonen, Puhakka, Romanov-Chernigovsky,
  Järvenpää, Korkiakoski, …; 104 rows carry a name).
- ⚠ **The `func_*` flags describe the PRIMARY pilot, not the user.** On an instructing row the
  primary pilot is the *student*, so `func_student=true` even though our pilot flew as instructor.
  **Never map `func_dual`/`func_student` straight onto our `Student_Time`** — read `rami_role`.
- **Use it for:** confirming a doubtful block time, settling whether a paper row is local or UTC,
  checking landing counts, and naming the pupil on an instructing row. **The paper logbook remains
  authoritative** (same rule as Aviatron); flag discrepancies to the user, don't silently rewrite.
- **71 of its rows postdate 19/09/2024**, so it forward-cross-checks everything still to do
  (IMG_6028 onward) — club regs only.
- ⚠ A few rows are junk: e.g. `2026-05-16` Kabböle→Kabböle has `00:00-00:00`, `block_min 0`. Sanity
  check before leaning on a row.

## Aviatron.pdf — the OTHER electronic cross-reference (⚠ covers `OH-GKT` to 07/2026)
`Aviatron.pdf` (repo root, tracked) is an electronic-logbook export of the **Blue Skies aviation
aircraft only**: `OH-PIF` (IR trainer), **`OH-GKT`** (= re-registered SE-GKT seaplane), `OH-DBS`,
`OH-TIL`, `OH-DBE`. 126 flights, 05/2018→07/2026, **all times UTC**.

⚠⚠ **This is the counterpart to `laskukierros_flights.csv`, not a legacy artefact.** The two cover
**disjoint fleets** — laskukierros has the club aircraft, Aviatron has the pilot's own. Between them
they cover nearly every Book-3 float row. **`OH-GKT` rows appear ONLY here**, and they run all the
way to July 2026, so Aviatron forward-checks every remaining OH-GKT row in the book.
**Grep both references before transcribing a spread; never tell the user "there is no record" for an
OH-GKT row without checking here first.** (This was missed through IMG_6022–6031.)

**Extracting it:** `pdftotext -layout Aviatron.pdf out.txt`. Each flight is **two stacked records**:
- the **header** line — `ID · LÄHTÖP · LASKUP · OFF · ON · BLOCK · LASK · SYLLABUS · TYPE · … · PIC` —
  this is the **block** record and the one to compare against (`BLOCK` = block time, `LASK` = landings);
- a **`RIVI`** line beneath it — `RIVI · ILMA-ALUS · OUT · IN · AIR · HLÖM` — this is **airborne**
  time and will read ~5 min shorter. Do not cross-check against it.

- **The paper logbook is authoritative by default** (user decision) — Aviatron is a cross-check, and
  discrepancies get flagged, not silently applied. **But the user may rule a row a genuine book error
  and direct us to take Aviatron's figures** (done for 08/07/2025 OH-GKT — see `drift.md` IMG_6032).
- Use it to: resolve unreadable/overwritten Zulu times, settle landing counts on OH-GKT float rows,
  confirm instructor names, and enrich `Remarks`.
- **Known to differ from paper:** block times by ~1–2 min, and Aviatron omits solo/non-course flights.

## Instructors / other-pilot names (pic_name for non-PIC rows)
- **Autere** — IR instructor for the pilot's OH-PIF student flights (student time + instrument).
- **Stude** — instructor for DA40 OH-STL student flights (and the 2017 student flights in drift.md).
- **Sinervä** — instructor for a 30/04/2019 OH-CTL seaplane student flight (IMG_4926).
- **Jansson, Korjula** — PIC/instructor on early seaplane student flights (2017). **Jansson** is
  also the instructor on the 19/06/2019 C185 OH-CDK floatplane student flight (IMG_4930).
- **Härkönen** — instructor on the 19/04/2020 C172 OH-CGX student checkout flight (IMG_4940).
- **Ravantti** (FI.FCL.34041) — examiner on the **04/09/2023 CRI(A) Assessment of Competence**
  (OH-GKT, Kahvisaari local, PASSED). Logged under Dual → Student_Time. See IMG_6018 in `drift.md`.
- **Tarhanen** — PIC on the pilot's IR-revalidation student rows (20/09/2021 OH-TIL, 10/08/2022 OH-PIF).
- **Salo** — instructor on the 05/07/2022 OH-TIL P28A dual row (IMG_6012).
- **Ignaty Romanov-Chernigovsky** — a *student* of the pilot's (Oppilas) on the OH-CTL Tuusulanjärvi
  seaplane instruction, e.g. 01/07/2024 (IMG_6024) and 06/08/2024 (IMG_6026). Not stored in the CSV:
  on instructing rows `pic_name` stays `self`. Recorded here only as context for who the pupils were.
- **Other summer-2024 float pupils** (Oppilas; same convention — not stored, context only):
  **Tommi Nirkkonen** (28/07 + three 06/08 OH-CTL legs), **Ilkka Korkiakoski** (14/08 OH-CTL),
  **Ivan Siragusa** (19/09 OH-GKT Vääksy→Kahvisaari).
- **Tarhanen** also examines the **02/08/2024** SR20 OH-ESR IR revalidation (IMG_6026 row 9, remark
  `(TAR)`) — a third Tarhanen student row.

## ⚠ `LT` subscripts in Book 3 — present but NOT applied consistently
From IMG_6024 the pilot writes a tiny **`LT`** subscript beside some off/on-block minute cells
(6024 rows 2–6, all OH-CTL). That is the book's documented "this row is local" marker and it is the
first time we have seen it actually used. **But its absence proves nothing** — IMG_6023 carries no
`LT` marks at all and the user confirmed those rows are local too. Treat `LT` as positive evidence
only; for unmarked rows fall back on the per-aircraft rule and ask the user.
⚠ Watch for the `LT` subscript blurring the preceding digit: on IMG_6024 rows 2 and 5 an off-block
`20` + `LT` read convincingly as `29`, putting both rows 9 minutes out (see `drift.md`).

Seaplane places seen: Laajasalo, Kalkkiranta, Inkoo, **Kabböle** (near Loviisa), Kubböle,
Tuusulanjärvi, Hiidenvesi, Räyskälä, Gumbostrand, Salonsaari, Leikonvesi, Savonlinna, Kuohijärvi,
Kahvisaari, Kelvenne, Iso-Jälä (Saimaa-lakes region, Jun–Jul 2019).
**Book 3 float season 2023 adds:** Anttola, Siltasaari, Pellinki, **Halsholmen** (all user-confirmed
2026-07-31; the book's handwriting for Halsholmen is partly obscured and reads as "Halsholm").
Rarer: `OH-TIL, OH-SPH, OH-CTH, OH-CMO`.
Airports seen: EFHF (Malmi), EFLA (Vesivehmaa), EFHV (Hyvinkää), EFKI, EFNU (Nummela),
EFHV, EFTU (Turku), EFMA (Mariehamn/Åland), EFHA (Halli), EFJY (Jyväskylä), EFLP (Lappeenranta),
EFTP (Tampere-Pirkkala), EFRY (Räyskälä — C185 floatplane dep 29/04/2020); Swedish: ESSP (Norrköping)
— the Aug–Sep 2019 CB-IR cross-countries. **Book 3 adds: EFPO (Pori)** and **ESNU (Umeå, Sweden —
15/04/2023 SR20 IFR one-way; someone else ferried the aircraft back, so no return leg is logged)**.
Estonian: **EETN** (Tallinn — 31/12/2019 DA40 day-trip, first non-Nordic field).
**Also in Book 3: EFVP (Vampula)** — user-confirmed; 08/09/2023 OH-PDP EFHV↔EFVP day-return (IMG_6019).
**EFIK (Kiikala)** — 29/04/2024 SR20 dual round-trip with Stude (IMG_6021).
**2025 float season adds three lakes near Tuusula/Hyvinkää: `Hirvijärvi`, `Loppijärvi`,
`Kytäjärvi`** (IMG_6030/6031, all corroborated by `laskukierros_flights.csv`).
`OH-COF` (C152, EFNU) reappears 16/05/2025 as the pilot's first Book-3 OH-COF row — it is in the
Book-1/2 C152 list above, not a new registration.
**IMG_6029 (Jan–Apr 2025) adds three foreign fields, all SR20 OH-ESR:** **ESMG** (Feringe/Ljungby,
Sweden — the aircraft wintered there, flown out 12/01/2025 and back 08/03/2025), **EHGG**
(Groningen Eelde, NL) and **EDWF** (Leer-Papenburg, DE). ⚠ The 07/03/2025 `EHGG→EDWF` leg does not
connect to the ESMG legs either side: **it was a two-pilot ferry and he logged only his own legs**
(user-confirmed). EFJO (Joensuu) also appears in Book 3 (17/10/2024 OH-CAM day-return).
**Book 3 float season 2024 adds: Lieso, Lietsaari** (Päijänne/Vääksy area, OH-GKT — all three names
user-confirmed 2026-08-01) and **Padasjoki** (Päijänne, IMG_6023), **Pulkkilanharju** (Päijänne,
IMG_6024) and **Mäntyharju** (IMG_6025).
⚠ **`Kahvisaari` is OH-GKT's HOME BASE, near Lahti** (user-confirmed 2026-08-01) — not a Saimaa /
Hillosensalmi lake as earlier notes implied. That is why Tuusulanjärvi↔Kahvisaari (~0:40) and
Padasjoki→Kahvisaari (0:27) legs are routine ferry hops, not anomalies.
⚠ **From IMG_6022 the book writes the type as `C172sea` for float C172s.** That is the pilot's own
informal seaplane marker, not a type — **store plain `C172`**; the seaplane flag comes from the
registration (user-confirmed 2026-08-01).
Plus lakes for seaplane ops (Tuusulanjärvi, Hiidenvesi, Räyskälä, Gumbostrand, Salonsaari,
Keilaniemi, Vuosaari, Kelvenne, Kahvisaari).
Finnish regs are `OH-xxx`. **Flag anything that breaks this pattern** (e.g. `OK-PDP` is almost
certainly an OCR error for `OH-PDP`).

⚠⚠ **The place / airport / registration lists above are hand-maintained running notes — they are NOT
derived from the CSVs and they have gaps.** `EFSA` (Savonlinna) was absent from them despite being
flown 6 times since 2012 (`15/08/2012` OH-CWB EFLP↔EFSA; `12/05/2018` and `18–19/07/2020` OH-PDP
EFHF↔EFSA). **Before calling any place, airport or registration "new", grep all three CSVs**
(`logbook_1_final.csv`, `logbook_2_final.csv`, `logbook_3.csv`) — those are the only authority on
what has been flown before.

## Sanity checks (report, don't silently fix)
- On-block minus off-block should ≈ Block_Time ≈ Total_Time.
- Impossible/typo'd dates (e.g. day > 31).
- Registration typos vs the known set above.
- Landing counts (see `drift.md` — paper cumulative landings has historically run high).

## Files & directories
- `logbook_2.csv` — **active output**, the file we extend. Tracked in git.
- `logbook_1_final.csv` — completed Book 1, seed for Book 2. Tracked, do not edit.
- `logbook-2/IMG_XXXX.jpg` — page images Claude transcribes. **Not tracked** (gitignored).
- `logbook2-20260327.../logbook2/*.HEIC` — original HEIC images. Not tracked.
- `logbook-2-csv/*.csv` — stale ollama per-image output. Not tracked; untrusted.
- `convert_heic_to_jpg.py`, `run_processing.py`, `run_patch_7rows.py`,
  `logbook_yearly_totals.py` — processing scripts. Tracked.
