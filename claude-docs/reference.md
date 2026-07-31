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

## Aviatron.pdf — electronic cross-reference (NOT the source of truth)
`Aviatron.pdf` (repo root, tracked) is an electronic-logbook export of the **Blue Skies aviation
aircraft only**: `OH-PIF` (IR trainer), `OH-GKT` (= re-registered SE-GKT seaplane), `OH-DBS`,
`OH-TIL`, `OH-DBE`. 126 flights, 05/2018→07/2026, **all times UTC**. Columns include exact
off/on-block, landings, syllabus codes (`CB-IR/n`), instructor (PIC column), student (OPPILAS =
"Ayoub, Rami"), and remarks (approach types, "CB-IR course completed").

- **The paper logbook is authoritative** (user decision). Aviatron is a cross-check only.
- Use it to: resolve unreadable Zulu times, confirm instructor names, and enrich `Remarks`.
- **Known to differ from paper:** block times by ~1–2 min, landings (Aviatron often counts more),
  and Aviatron omits solo/non-course flights. Do NOT rewrite paper-based rows to match Aviatron;
  flag material discrepancies to the user instead.

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
**Book 3 float season 2024 adds: Lieso, Lietsaari** (Päijänne/Vääksy area, OH-GKT — all three names
user-confirmed 2026-08-01) and **Padasjoki** (Päijänne, IMG_6023).
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
