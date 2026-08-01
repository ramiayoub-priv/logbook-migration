# Drift Notes — Corrections & Discrepancies

Log every correction that alters a row later rows built on. Fix cumulatives from that row forward.

## 🔎 OPEN — items found by the app importer (2026-08-01), re-investigated 2026-08-01

Building the app's importer (`app/backend/internal/csvbook`) put every row of all three CSVs through
a row-by-row reconciliation of all seven `Cumulative_*` series plus a set of per-row invariants. Over
1293 flights it found **one cumulative break and two formatting problems that had never been logged**.
A follow-up pass on 2026-08-01 **localised all three** and turned up **a fourth defect** (item 4).
None has been corrected — the paper book is the authority and these are the user's to rule on.
Full context in `app/docs/data-model.md`.

> **⚠ METHOD NOTE — how night time is and is not determined (user, 2026-08-01).**
> **`Night_Time` comes from the book's night column and nothing else.** If a row is marked night on
> paper it is night; if it is not marked, it is not night. **Never infer night time from clock times,
> sunset, twilight or time zones, and never write a computed value into a row.** A solar calculation
> was used *once*, on 2026-08-01, purely to rank *which paper rows are worth re-reading* for item 3 —
> it is a page-finding aid, not evidence, and no value it produced has entered or may enter the CSVs.

### 1. `logbook_1_final.csv` line 28 — instrument time exceeds flight time
`28/09/2011 · C152 OH-COF · EFHF → EFHF · 08:22–09:34`
- `Block_Time` **1:12**, `Total_Time` **1:12**, but `Instrument_Time` **1:21**. A flight cannot log
  more instrument time than it lasted.
- Its `Cumulative_Instrument` advances **1:00 → 2:12**, i.e. by exactly 1:12. **The cumulative column
  is self-consistent; the row's `Instrument_Time` is the outlier**, which makes `1:21` almost
  certainly a transposition of `1:12`.
- **Consequence, currently live:** summing the `Instrument_Time` column across all three books gives
  **107:14**, while `Cumulative_Instrument` ends at **107:05**. Every instrument figure downstream of
  line 28 in our series is therefore 9 minutes low relative to the rows.
- This is the **only** break in all seven series across all 1293 flights — everything else reconciles
  exactly, including `Cumulative_SEP_Sea` (407:39) row by row.

**Re-investigated 2026-08-01 — the case for `1:12` is now strong, and the fix costs nothing:**
- **The 9 minutes live entirely inside Book 1.** Book 1's `Instrument_Time` column sums to **3:21**
  but its own last `Cumulative_Instrument` is **3:12**. Books 2 and 3 then chain off 3:12 perfectly:
  3:12 + Book 2's column 61:56 = **65:08** (Book 2's last cumulative, exact), and 65:08 + Book 3's
  column 41:57 = **107:05** (exact). So one row, one column, one book — nothing else is involved.
- **`Total_Time` is not the wrong cell.** Three independent cells agree on 1:12 (`Off_Block`/`On_Block`
  08:22–09:34, `Block_Time`, `Total_Time`) and `Cumulative_Total` reconciles across all 1293 rows.
  The duration is solid, so `Instrument_Time` cannot exceed 1:12 whatever the truth is.
- **Precedent one row earlier in the same training block.** Line 20 (`15/09/2011`, same C152 OH-COF,
  same instructor Martevuo) logs `Instrument_Time 1:00` on a `Total_Time 1:00` flight — **instrument
  == the whole flight**, the standard PPL basic-instrument lesson. Line 28 is the second such lesson
  and its cumulative advances by exactly its whole flight time. `1:12` fits that pattern; `1:21` is a
  digit transposition of it.
- **⚠ Why this matters for the app, not just for tidiness.** Rule 5 says cumulative totals are
  **computed from the rows, never stored** — so the app derives instrument time by summing
  `Instrument_Time` and will show **107:14**, 9 minutes above the **107:05** that every stored
  cumulative carries and that the user inked on paper at p.62 (SE-IFR 105:57 at that point).
  **Until this row is fixed the app and the paper book cannot agree on instrument time.**
- **The fix has zero downstream effect.** Every `Cumulative_Instrument` value in all three books, and
  the inked p.62 figure, *already* reflect 1:12. Changing the row to 1:12 makes the column agree with
  the cumulative; **no cumulative needs rebuilding and no total moves.** (The earlier note here said
  "rebuild from line 28 forward (+9 min)" — that was backwards; the cumulatives are already right.)
- ⏸ **Awaiting the user.** Recommended: `1:21` → `1:12`, single cell, nothing else touched.
  Book 1's page images are not in the repo (gitignored, and not on this disk), so this cannot be
  read back off paper without the physical book.

### 2. `logbook_2_final.csv` lines 83–90 — dates written `DD.MM.YYYY`
Eight consecutive rows from one transcription batch use dots instead of slashes:
`21.03.2018, 30.03.2018, 19.04.2018, 22.04.2018 ×3, 04.05.2018 ×2`.
- Six are unambiguous on their own (day > 12). The chronological bracket — line 82 is `15/03/2018`
  and line 91 is `07/05/2018` — confirms day-first for all eight.
- ⚠ **The two `04.05.2018` rows (lines 89, 90) cannot be settled from the cell alone.** Read as
  4 May; if the paper says 5 April, two rows move by a month. **Worth one look at the page.**
- The importer accepts them and flags them; the CSV itself is unchanged. ⏸ **A normalisation pass to
  `DD/MM/YYYY` is the user's call.**

**Re-investigated 2026-08-01 — no electronic source can settle lines 89–90; the paper is the only way.**
- **Aviatron contains no `OH-PDP` row at all** (0 of 126 — it covers the pilot's own/Blue Skies fleet:
  GKT, PIF, DBS, DBE, TIL). **`laskukierros_flights.csv` starts 19/04/2020**, two years too late.
  Both references were checked directly; neither has any coverage of these two flights.
- **`OH-PDP`'s own history gives no bracket either** — its first flight in the books is 23/03/2017,
  well before either candidate date, so an April reading is not excluded by the aircraft.
- What still stands, and it is decent: **all eight dotted rows come from one transcription batch and
  six of them are unambiguously day-first** (21, 30, 19, 22, 22, 22) — a batch does not switch date
  format halfway. And 4 May sits correctly in book order between line 88 (`22.04.2018`) and line 91
  (`07/05/2018`), whereas 5 April would be ~3 weeks out of order.
- ⏸ **Still worth the one look.** The two rows are an evening out-and-back to Lahti:
  `EFHF→EFLA 17:15–18:04 (0:49)` and `EFLA→EFHF 18:49–19:40 (0:51)`, P28A OH-PDP, cumulative
  Total 474:07 then 474:58. **If the paper says 5 April, two rows move by a month** (row order only —
  no total changes, since the totals do not depend on the date).

### 3. Night time — the paper's 22:45 does not match the CSV's 16:47
- The `Night_Time` column sums to **16:47** over all three books (Book 1 9:04 · Book 2 3:40 ·
  Book 3 4:03).
- The p.62 inked block carries night time **22:45**. This file already records that figure as
  *"supplied but not read back"* — it was never reconciled. **The gap is 5:58.**
- Every other p.62 figure (Total, PIC, SE-IFR, Dual, FI, landings) was reconciled and closed.
  Night time is the one that was not, and it does not agree.
- ⏸ **Awaiting the user.** Either the paper's night column carries time our rows do not, or 22:45 was
  a mis-add. Note that the **59 night landings** inked at p.62 were derived from these same rows, so
  if the night *time* is wrong the landing split below may need revisiting too.

**⭐ SOLVED-BY-HALVES 2026-08-01 — `22:45` IS NOT A MIS-ADD, AND BOOK 3 IS CLEAN.**
The gap is **entirely inherited from Books 1–2** and was already on paper before Book 3 opened:

| | our `Night_Time` | paper | Δ |
|---|---|---|---|
| Books 1 + 2 (9:04 + 3:40) | **12:44** | EASA "TOTAL PREVIOUS PAGES" night **18:42** | **−5:58** |
| Book 3 | **4:03** | 22:45 − 18:42 = **4:03** | **0 — exact** |
| p.62 inked total | 16:47 | **22:45** | −5:58 |

- **18:42 + 4:03 = 22:45 to the minute.** The paper is internally consistent: the pilot carried 18:42
  into the EASA book in 2021 and added exactly the night time our Book-3 rows carry. **So 22:45 was
  correctly added and there is nothing wrong anywhere in Book 3** — its seven night rows account for
  the whole Book-3 movement. The p.62 figure is not the defect; it faithfully reflects a carry-in we
  never reconciled.
- **The entire 5:58 is unreconciled Book-1/Book-2 night time**, and it predates the migration: the
  discrepancy is between the *old paper books' night column* and *our transcription of it*. The
  18:42 carry-in has been sitting in this file since 2026-07-31 (see the Book-3 start cross-check
  below) — it was recorded as a Total/PIC/SE-IFR/Dual/Instructor check and the night line was never
  compared. Every other column on that carry-in was checked; night was the one that was not.
- **Most likely reading: rows in Books 1–2 whose night column we did not transcribe.** All three
  books together hold only 22 rows with any night time, so 5:58 is roughly six ordinary evening
  circuits' worth — a small number of missed cells, not a systematic error.

⏸ **Awaiting the user — and this one genuinely needs the paper.** Books 1 and 2 page images are
**not in the repo** (gitignored) and are not on this disk, so the night column cannot be re-read here.

**Shortlist of Book-1 rows to re-read — A PAGE-FINDING AID ONLY, NOT EVIDENCE.** These are rows that
carry **no `Night_Time`** in our CSV. They are listed *only* so the physical book can be opened at the
right pages; **the book's night column is the sole authority and decides every one of them.** Six of
them are evening flights whose durations happen to sum to **5:39 of the 5:58** — suggestive of where
to look, and nothing more:

| line | date | reg | route | block | total |
|---|---|---|---|---|---|
| 111 | 21/01/2013 | OH-CTM | EFHF local | 18:41–19:58 | 1:17 |
| 173 | 25/02/2014 | OH-KLS | EFHF local | 18:31–19:26 | 0:55 |
| 237 | 15/09/2014 | OH-CMO | EFLA→EFHF | 20:27–21:18 | 0:51 |
| 245 | 09/12/2014 | OH-KAM | EFHF local | 17:18–18:30 | 1:12 |
| 246 | 20/12/2014 | OH-CMO | EFHF local | 16:28–17:15 | 0:47 |
| 250 | 01/02/2015 | OH-CAV | EFHF local | 17:21–17:58 | 0:37 |

Also worth a glance while the book is open, all Book 1: line 107 (`09/11/2012` OH-CWB), line 177
(`26/03/2014` OH-TIL EFSI→EFHF), line 255 (`16/03/2015` OH-CMO). And the five **partial**-night rows
listed in the p.62 landing-split table further down already have a night value — if any of those is
short on paper, that closes part of the 5:58 too.

**Knock-on if any of these rows does carry night time on paper:** its landings become night landings,
so the inked **59 night / 3335 day** split at p.62 moves. `Cumulative_Landings` (the sum) is
unaffected either way — only the split. Correct p.62 rather than the CSV, per the note further down.

### 4. `logbook_2_final.csv` line 102 — registration typo `OK-PDP` (found 2026-08-01)
`17/05/2018 · P28A · EFRY → EFHF · 18:51–19:35 · 0:44`
- The registration reads **`OK-PDP`** (Czech prefix) where every other row reads `OH-PDP`. It is the
  **return leg of the out-and-back on line 101** (`17/05/2018` OH-PDP `EFHF→EFRY` 17:57–18:43), so
  the aircraft is not in doubt.
- **This is the exact OCR failure `resume.md` warns about** — it cites `OK-PDP` in the stale ollama
  output `logbook-2-csv/logbook_IMG_4920.csv` as an example of untrusted output. **It survived into
  `logbook_2_final.csv`.**
- **Consequence for the app:** the aircraft table is keyed on registration, so this row creates a
  **phantom one-flight aircraft `OK-PDP`** and detaches 0:44 from OH-PDP's totals.
- Only 3 registrations in 1293 rows fail the `OH-XXX` pattern; the other two are genuine —
  **`SE-LWI`** (one flight) and **`SE-GKT`** (14 rows, `11/06/2015`–`06/09/2016`), which is the
  pilot's own aircraft under its **Swedish registration before it became `OH-GKT`** (first OH-GKT row
  `16/06/2018`; the two never overlap). ⚠ **`SE-GKT` and `OH-GKT` are the same airframe** — the app
  must not treat them as two aircraft. Not a defect, but it needs handling.
- ⏸ **Awaiting the user.** Recommended: `OK-PDP` → `OH-PDP`. No time value changes.

### 5. Minor — `logbook_2_final.csv` line 97 sits one hour off Aviatron (found 2026-08-01)
`10/05/2018 · C150 OH-DBS · EFLA local · 08:15–09:27 · 1:12` (student, Ravantti).
Aviatron id **9843** has the same aircraft, date and **identical 1:12 block**, but at **09:15–10:27**
— exactly one hour later. The **other** OH-DBS row that day (line 98, 13:46–14:45) matches Aviatron
id 9852 **to the minute**, so this is not a time-zone offset (that would move both rows); it is a
one-hour slip on one cell, in the book or in transcription. **No total is affected** — the duration
is right either way. ⏸ Documented only; the paper decides.

### Not defects — confirmed by the same sweep
- **18 rows are genuinely out of date order** across the three books (e.g. Book 1 line 269,
  `11/06/2015` after `23/06/2015`). This is why the app orders on an explicit `seq` and never on
  `flight_date`. Expected, not an error.
- **`Block_Time != Total_Time` on exactly one row** (`08/09/2025`, 0:45 vs 0:38) — already known.
- **No row mixes zones within a time pair**, and no `Off_Block`/`On_Block` cell is blank, in any book.
- **Every one of the 1293 rows resolves to UTC unambiguously** — not one lands in a DST gap or fold.

## Intentionally excluded (struck-through / void) entries
- **IMG_4925 / 18/04/2019 / P28A OH-PIF / EFLA → EFLA / 10:00–12:03 / 2:03:** struck through —
  excluded. Paper running total skips it (562:32 → 563:55).
- **IMG_4925 / 12/04/2019 / P28A OH-PDP / EFPR → EFHF:** void first-row entry with no time, not in
  the page total — excluded (the EFHF→EFPR outbound is the last row of IMG_4924).

- **IMG_4921 / 15/10/2018 / C172 OH-CWB / EFHF → EFHF / 12:15–13:12 / 0:57:** this line is
  **struck through in the paper book** — a wrong entry. It is deliberately **NOT** in
  `logbook_2.csv`. Do not "recover" it. The paper's own running total skips it (538:19 → 539:04).

## Book 3 (EASA) — start & carry-over cross-check (2026-07-31)
- `logbook_3.csv` seeded from `logbook_2_final.csv`'s last row (our continuous series). The EASA book's
  printed **"TOTAL PREVIOUS PAGES"** gives an independent read of the Book-2 end totals: Total **787:06**,
  Night 18:42, SE-VFR 726:00, SE-IFR (instrument) **61:06**, PIC **637:50**, Flight Instructor **48:59**,
  Dual/Student **149:45**, Landings **1909**. Offsets vs ours match the standing drift (Total +25, PIC −27,
  Instrument +4:02, Student +0:23, **Landings exact**) — **except Instructor**: EASA carries **48:59** but
  ours is **47:39** (paper **+1:20** here, opposite sign to earlier Book-2 estimates that had ours slightly
  ahead). Not reconciled; harmless as long as we cross-check on per-page Δ. Revisit only if the final app
  needs absolute-instructor agreement with the paper.
- **IMG_6007 (pages 1–2):** all 4 page cross-checks reconciled exactly (Δtotal 12:10, Δpic 12:10,
  Δinstr 7:11, Δland 57). No corrections needed.
- **IMG_6008 (pages 3–4, 29/07/2021–10/09/2021) — appended 2026-07-31.** 15 flights, all UTC (`Z`),
  a **notably clean page**: all 4 page Δ reconciled exactly (Δtotal 13:30, Δpic 13:30, Δinstr 2:45,
  Δland 44) and **zero block-vs-total flags** (every off/on-block equals its logged total). No
  student/dual and no night this page. Instructing rows (PIC+Instructor=Total): 29/07 OH-CAY EFHV
  local (1:00), 01/08 OH-CTL Tuusulanjärvi local (1:05, seaplane), 06/08 OH-COK EFPR local (0:40).
  Two instrument (SE-IFR) rows, both **PIC** OH-TIL P28A: 05/08 EFLA local 1:56, 03/09 EFKU→EFLA 1:18
  (page SE-IFR Δ 3:14 reconciles). Lake ops on OH-CTL/OH-GKT (Tuusulanjärvi/Salonsaari/Kahvisaari/
  Virtosaari/Sandö — the Virtosaari & Sandö reads user-confirmed 2026-07-31). New reg **OH-COK** already
  seen in Book-2 closeout; new airport **EFKU** (Kuopio). **Paper-vs-ours drift at 10/09/2021:** paper
  printed bottoms Total 812:46, PIC 663:30, SE-IFR 64:20, Instructor 58:55, Landings 2010. Ours: Total
  813:11 (+25), PIC 663:03 (−27), Instrument 68:22 (+4:02), Instructor 57:35 (paper +1:20), Landings
  2010 (**exact** — drift stays closed).
- **IMG_6009 (pages 5–6, 20/09/2021–26/12/2021) — appended 2026-07-31.** 15 flights, all UTC (`Z`).
  Time cross-checks all exact (Δtotal 13:56, Δpic 12:16, Δstudent 1:40, Δinstr 4:56). One 2-min taxi
  warning (17/12 OH-CAM 17:57–18:29 = 0:32 vs logged 0:30; appended deliberately, same class as prior).
  - **Student row (user-confirmed):** 20/09 P28A **OH-TIL** EFLA→**EFIM** 1:40, **PIC = Tarhanen**, user
    flew as student on an **IR revalidation** → Student_Time 1:40 + Instrument 1:40, pic_name=Tarhanen.
    (Note: distinct from the 2020 OH-TIL "Lord" instructor correction — the book here literally writes
    TARHANEN as PIC.) Only SE-IFR/instrument row this page.
  - **Instructing rows (PIC+Instructor=Total):** 29/10 OH-PDP EFHV local (1:30), and three 30/10 **OH-CAM**
    flights (1:05 EFHV, 1:32 EFPR, 0:49 EFHV). Δinstr 4:56 reconciled exactly. **OH-CAM** = new C172 reg.
  - **First night flight of Book 3:** 17/12 OH-CAM EFHV local 0:30, **Night 0:30, 3 night landings**
    (day-landing cell struck in the book → 0 day, 3 night). New reg **OH-CMV** (C152, 26/12 EFHV local).
  - **LANDINGS CONVENTION CHANGED — count ALL landings (day+night) in Cumulative_Landings** (user decision
    2026-07-31, replacing the day-only "match the paper" rule now that night landings occur). Per-row reads
    sum to **39 day + 3 night = 42** this page → **Cumulative_Landings 2052**. The book's printed cumulative
    reads **2047** (day-only; its page day-total cell is struck/overwritten, best read ~37–39). So from
    17/12/2021 **ours runs +5 ahead of the book's printed count** (= 3 night landings + a ~2-landing slip in
    the book's own struck day column that we don't inherit — we trust per-row sums). This is a **new standing
    drift**; the old "landings exact vs paper" note is superseded for Book 3. Day-landing reads on the short
    rows (27/09 OH-PDP 0:18 = 3; 13/10 OH-CTL lake hops = 3 each) are my best legible reads, not
    independently reconciled against the struck page-total.
  - **New regs:** OH-CAM (C172), OH-CMV (C152). New field **EFIM** (Immola) already in Book-2 notes.
- **IMG_6010 (pages 7–8, 10/02/2022–12/05/2022) — appended 2026-07-31.** 15 flights, all UTC (`Z`).
  All Δ exact (Δtotal 14:18, Δpic 11:37, Δstudent 2:41, Δinstr 4:02, Δland 54). Zero block flags.
  - **New type SR20 (OH-ESR)** = Cirrus SR20. **3 student rows** (type-rating training, remark "Koulutus
    TYPE SR20"): 17/04 EFNU 0:50, 22/04 EFNU 1:00, 12/05 EFNU→EFTU 0:51. pic_name=**Stude** (instructor;
    same Stude as prior DA40/2017 student flights, user-confirmed). Δstudent 2:41 reconciled.
  - **Two night flights** (Night_Time + night landings): 23/02 **OH-CGX** EFHV local 0:25 night (3 night
    ldg), 03/03 OH-CAM EFHV local 0:49 night (5 night ldg). Landings this page = **46 day + 8 night = 54**.
  - Instructing (PIC+Instr): 11/03 OH-CAM 1:12, 12/04 OH-CAM 1:44, 11/05 OH-CTL 1:06 (Δinstr 4:02).
- **IMG_6011 (pages 9–10, 12/05/2022–10/06/2022) — appended 2026-07-31.** 15 flights, all UTC (`Z`).
  Δtotal 14:23, Δstudent 2:04, Δinstr 7:22, Δland 66 all exact. PIC **not** cross-checked vs paper (paper
  column error, below); PIC pinned by total−student = 12:19 (= our row sum).
  - **SR20 type rating PASSED:** 12/05 EFTU→EFNU 0:59 (student, pic_name Stude, remark "…PASSED"). From
    **18/05 the SR20 (OH-ESR) is flown PIC** — 18/05 EFNU local 1:45 with **instrument 1:00** (SE-IFR).
  - **Seaplane student row:** 13/05 OH-CTL Tuusulanjärvi→Hiidenvesi 1:05 (9 ldg), PIC=**Sinervä** →
    Student_Time, pic_name Sinervä. (Sinervä = the seaplane instructor from Book-2 30/04/2019 & 15/07/2020.)
  - **Seven OH-CTL seaplane instructing** flights (Tuusulanjärvi/Hiidenvesi/Karhusaari lake circuits):
    13/05 0:50, 14/05 1:17, 15/05 1:10, 20/05 1:10, 24/05 1:25, 27/05 1:05, 27/05 0:25. Δinstr 7:22.
  - **Row 5 (15/05 OH-CTL) on-block inferred = 08:24** (book wrote **07:24**, impossible for the logged 1:10
    from off-block 07:14 — a 60-min book slip; user-confirmed). Off/on recorded 07:14–08:24. Also a 5-min
    warn on 13/05 OH-CTL (off/on 14:08–15:18 = 1:10 vs logged 1:05; logged authoritative).
  - **Paper PIC-column slip:** book wrote PIC-this-page **14:17**, but total−student = **12:19** (= our row
    sum); the book's PIC cell over-adds by 1:58 (same class as prior PIC-column errors). Our 12:19 kept.
    Effect: our Cumulative_PIC now runs further behind the book's printed PIC column, but is correct.
  - **Landings all day (66)**, no night this page.
- **IMG_6012 (pages 11–12, 10/06/2022–21/07/2022) — appended 2026-07-31.** 15 flights, all UTC (`Z`).
  **All 5 Δ exact** (Δtotal 12:57, Δpic 12:35, Δstudent 0:22, Δinstr 1:20, Δland 33). No night this page.
  - **Original image was already UPRIGHT — no rotation** (6007–6011 needed CCW rotate(90); Book-3
    orientation is inconsistent, check each image).
  - **Instructing:** 22/06 OH-CTL seaplane Tuusulanjärvi local 1:20 (Instructor 1:20, PIC 1:20, 6 ldg).
  - **Student row:** 05/07 OH-TIL P28A EFTP local 0:22, pic_name=**Salo** (new instructor name), Dual 0:22
    → Student_Time. Δstudent 0:22.
  - **Three instrument (SE-IFR) rows, SR20 OH-ESR now routine PIC:** 30/06 EFNU local 2:10 (instr 0:30),
    06/07 EFNU→EFTU 0:44 (instr 0:44), 06/07 EFTU→EFNU 0:47 (instr 0:47). Turku (EFTU) day-trips.
  - **Row 14 (17/07 OH-PDP EFHV local) on-block inferred = 18:36** (written digit smudged; 18:01→18:36 =
    0:35 = logged time & cumulative Δ). No user action needed — unambiguous from the math.
  - **Row 15 (21/07 OH-CDK C185 floatplane, Papinluoto→Astuvansalmi, Saimaa) block inferred (user-approved
    one-time exception):** book's block cells 08:02–08:44 = 0:42 but logged flight time = **1:00** (SE-VFR
    + cumulative both confirm). User: "I logged flight time, not block time" → widen block symmetrically
    (−9 min start, +9 min end) to 1:00 → recorded **off 07:53Z / on 08:53Z**. Arrival "Astuvansalmi"
    (user-supplied; handwriting read as "Aslonos"). OH-CDK counts SEP_Sea.
- **IMG_6013 (pages 13–14, 21/07/2022–04/09/2022) — appended 2026-07-31.** 15 flights, all UTC (`Z`).
  Original image **already UPRIGHT — no rotation** (like 6012). No night this page.
  - **⚠ 1:00 book slip at the TARHANEN student row (10/08 OH-PIF EFLA local).** The flight is **1:55**
    (Student 1:55 + Instrument 1:55, pic_name=**Tarhanen** — another IR-reval, same instructor/pattern as
    the 20/09/2021 OH-TIL row). Three cells agree on 1:55: on-block 13:20→15:15 (overwritten), SE-IFR
    column 1:55, Dual column 1:55. But the **running Total column wrote 875:28** (implies 0:55) — a 1:00
    undercount that propagated into the book's printed page totals. So the book's printed **TOTAL THIS
    PAGE 13:49 → should be 14:49**, printed **PIC 11:54 → 12:54**, printed **SE-VFR 11:54 → 12:54**, and
    every book cumulative from this row forward is **1:00 low**. Directly-summed columns (SE-IFR 1:55,
    Dual 1:55, Flight-Instructor 2:17) are correct. We fed the **corrected** Δtotal 14:49 / Δpic 12:54 to
    the cross-check (all 5 Δ then exact). **The book's own SE-IFR total went 69:01→70:56, i.e. it credited
    the full 1:55 there — confirming the running-Total column, not the flight, is the error.**
  - **Missing landing on the Tarhanen row counted as 1** (user-instructed; the landings cell was blank).
    Page day-landings = 28 written + 1 = **29** (no night). Δland 29 exact.
  - **Row 7 (07/08 OH-CTL Tuusula) on-block inferred = 17:10** (written "12:10" is impossible before
    off-block 16:04; 16:04→17:10 = 1:06 = logged & cumulative Δ). Seaplane instructing (Instructor 1:06).
  - **Instructing (2):** 01/08 & 07/08 OH-CTL Tuusula seaplane locals, Instructor 1:11 + 1:06 = 2:17, 6 ldg each.
  - **Row 15 date = 04/09/2022** (user-corrected; book digits read as "04/07/22", chronologically out of
    order after the 24/09 rows). EFHV→EFLA OH-PDP 0:40.
  - **Place-name reads (user-confirmed):** "Leukolws" = **Leikonvesi** (Saimaa float, rows 2–3);
    24/09 round-trip is **Tuusula ↔ Hiidenvesi** (seaplane lake, OH-CTL floatplane), read initially as "Hitanpää".
  - **Lapland floatplane trip:** 23/08 OH-CTL Inari→Kemijärvi→Sodankylä (2:05 + 2:06), SEP_Sea.
  - SEP_Sea this page +9:31 (OH-CDK ×3 + OH-CTL ×6) → 240:24.
- **IMG_6014 (pages 15–16, 04/09/2022–11/02/2023) — appended 2026-07-31.** 15 flights. Original image
  **already UPRIGHT — no rotation** (like 6012/6013). **All 4 page cross-checks exact, zero block
  warnings** — every off/on-block matched its logged time to the minute (Δtotal 14:07, Δpic 13:17,
  Δstudent 0:50, Δinstr 3:24). No night and no SE-IFR this page (book's SE-IFR total stays 70:56).
  - **⚠ MIXED TIME ZONES ON ONE PAGE — first occurrence in Book 3 (user-confirmed 2026-07-31).** The
    three 19/10/2022 OH-CAY rows appeared to overlap: 12:43–13:51, 14:18–15:04, 12:21–13:30. User: the
    **third row is UTC, the first two are local**. Converting row 3 to local (UTC+3, EEST until 30/10)
    gives 15:21–16:30 — the three flights are then perfectly sequential. Stored accordingly: rows 8 & 9
    **plain local (no `Z`)**, row 10 keeps **`Z`**. Scope confirmed as *those two rows only* — the other
    12 rows on the spread stay `Z`/UTC. **Do not assume a whole spread shares one time zone.**
  - **Row 8 (19/10 OH-CAY) on-block inferred = 13:51** — the last digit is written over itself (reads as
    51/56/58). Logged time 1:08 and the page total both require 13:51 (12:43 + 1:08).
  - **Instructing (3):** 12/10 OH-PDP EFHV local 1:07; 19/10 OH-CAY EFHV locals 1:08 + 1:09 = Δinstr 3:24.
    First instructing rows on P28A OH-PDP and on OH-CAY in Book 3.
  - **Student row:** 11/02/2023 SR20 OH-ESR EFNU local 0:50, pic_name=**Stude** — a second SR20 dual with
    the type-rating instructor, 9 months after the 12/05/2022 checkride. Dual 0:50, no instrument.
  - Book left the **per-page landings cell blank**, so no paper check on landings this page; ours sums to
    **58** (all day) → Cumulative_Landings 2292. SEP_Sea +0:52 (OH-CTL 12/10 Tuusula→EFRY) → 241:16.
  - **All four steady paper-vs-ours deltas unchanged** (SE-IFR +4:02, Dual +0:23, Instructor −1:20,
    Total/PIC ±1:25) — confirms the transcription and that no new book slip crept in on this spread.
- **IMG_6015 (pages 17–18, 14/02/2023–06/05/2023) — appended 2026-07-31.** 15 flights, all UTC (`Z`).
  Image **already upright**. **All 3 available cross-checks exact, zero block warnings** (Δtotal 15:17,
  Δpic 15:17, Δinstr 3:36). No Dual this page. Night +0:37 and SE-IFR +3:10 also tie to the book's own
  bottom line exactly.
  - **⚠ BOOK ARITHMETIC SLIP — page SE-VFR total (self-corrected by the pilot, does NOT affect us).**
    Printed "TOTAL THIS PAGE" SE-VFR reads **12:35**, but the column sums to **12:07**. The book caught
    it downstream: the SE-VFR cumulative cell is **struck through (`837:55`) and rewritten below as
    `837:27`** = 825:20 + 12:07. Error confined to that one printed cell; Total/PIC/Night/SE-IFR all
    reconcile. We don't store SE-VFR, so nothing propagates. (Unrelated to the p.14 running-Total slip.)
  - **Row 7 = genuine one-way international leg (user-confirmed):** 15/04/2023 SR20 OH-ESR
    **EFNU → ESNU (Umeå, Sweden)**, 05:45Z–07:45Z, 2:00 with **Instrument 1:50**. No return leg is
    logged because **someone else flew the aircraft back** — not a missing entry. First Swedish
    destination in Book 3.
  - **Row 3 = night VFR (user-confirmed):** 10/03/2023 C172 OH-CGX EFHV local 17:17Z–17:54Z, **Night
    0:37, no instrument**. Book struck the day-landing cell and wrote **2 in the NIGHT column** →
    counted as **2** under the day+night convention.
  - **Instructing (5, user-confirmed PIC + Instructor):** all five 06/05 **Kabböle** OH-CTL seaplane
    locals (0:35 + 0:50 + 0:48 + 0:43 + 0:40 = 3:36). The 05/05 Räyskälä→Kabböle ferry leg (1:07) is
    **PIC only**, no instructor time.
  - **New airport: EFPO (Pori)** — 22/04 SR20 IFR day-return EFNU↔EFPO (0:50 IFR outbound).
    SEP_Sea +4:43 (OH-CTL ×6) → 245:59. Book left the per-page landings cell blank again → no paper
    check; ours sums to **58** (56 day + 2 night).
- **IMG_6016 (pages 19–20, 07/05/2023–14/06/2023) — appended 2026-07-31.** 15 flights, all UTC (`Z`).
  Image **already upright**. **All 3 available cross-checks exact** (Δtotal 13:22, Δpic 12:48,
  Δinstr 4:41). No night and no SE-IFR this page (book's SE-IFR total stays 74:06). Heavy float
  season: 13 of 15 rows are OH-CTL seaplane (SEP_Sea +11:56 → **257:55**).
  - **⚠ BOOK SLIP IN THE DUAL COLUMN (+0:03).** The one dual row (12/05 Laajasalo, below) is **0:34** —
    corroborated three ways (block 10:29→11:03, SE-VFR cell 0:34, running-Total Δ 914:19→914:53) — but
    the book's printed **"TOTAL THIS PAGE" Dual reads 0:37** and carries into its cumulative
    (159:17 + 0:37 = 159:54). Ours is right; effect is that the **standing Dual drift moves
    +0:23 → +0:20**. Only that column is affected; Total/PIC/Instructor all still reconcile exactly.
  - **Student row (user-confirmed):** the *second* 12/05/2023 OH-CTL Laajasalo local,
    10:29Z–11:03Z, **0:34**, PIC=**SINERVÄ** → Student_Time 0:34, pic_name Sinervä, 12 landings.
    Third Sinervä seaplane dual (after 30/04/2019 and 13/05/2022). No instrument.
  - **Two book time slips, both user-resolved (2026-07-31) — and they resolved in *opposite* directions:**
    - **Row 2 (07/05 Kabböle→Laajasalo):** book's on-block "15:25" is wrong; user confirms
      **13:44Z–14:25Z = 0:41** (matches SE-VFR + running-Total). A 60-min slip, same class as
      IMG_6011 row 5. Arrival place was scribbled over/illegible — user read it as **Laajasalo**
      (the Kabböle→Helsinki ferry leg; the next two rows are Laajasalo locals, which corroborates it).
    - **Row 6 (19/05 Tuusulanjärvi→Anttola):** here the *off-block* is right and the **on-block was
      misread/mis-written** — user confirms **10:47Z–12:15Z = 1:28** (book's on-block cell reads 12:09).
      Recorded 12:15. **Lesson: don't assume the off-block is the bad cell just because "fixing" it makes
      the arithmetic close — ask.**
  - **Row 5 (09/05 SR20 OH-ESR EFNU local 0:50) is logged out of date order**, sitting between the two
    12/05 rows — user-confirmed as written in the book, not a misread date. Kept in book order.
  - **Instructing (5, PIC + Instructor = Total):** 07/05 Kabböle local 1:13 (7 ldg); 12/05 Laajasalo
    local 0:52 (11 ldg); 24/05 Tuusula local 1:23 (7 ldg); 14/06 Tuusula↔**Halsholmen** 0:45 + 0:28.
    Δinstr 4:41 exact. New float places this page: **Anttola, Siltasaari, Pellinki, Halsholmen**
    (Halsholmen user-confirmed 2026-07-31; the book's handwriting truncates it to "Halsholm").
  - Book left the **per-page landings cell blank** for the third time (as on IMG_6014/6015) → no paper
    check; ours sums to **58** (all day) → Cumulative_Landings **2408**.
  - **Paper-vs-ours drift at 14/06/2023** (book bottoms Total 924:55, PIC 767:31, SE-IFR 74:06,
    Dual 159:54, Flight-Instructor 90:33): ours Total 926:20 (**+1:25**), PIC 766:06 (**−1:25**),
    Instrument 78:08 (**+4:02**), Student 160:14 (**+0:20** — moved by the Dual slip above),
    Instructor 89:13 (**−1:20**). Four of five deltas steady; the Dual one stepped by exactly the
    book's 3-min error.
- **IMG_6017 (pages 21–22, 19/06/2023–17/07/2023) — appended 2026-07-31.** 15 flights, all UTC (`Z`).
  Image **already upright**. **All 3 available cross-checks exact** (Δtotal 15:59, Δpic 15:59,
  Δinstr 4:09). No night, no dual this page.
  - **⚠ THE PILOT FLAGGED HIS OWN ERROR — handwritten note below the table: "\* Error in total time
    14:07 → 15:59".** Row 12 (14/07 **OH-PIF** EFLA local) is **2:07** = SE-VFR 0:15 + **SE-IFR 1:52**,
    but the running-Total column only advanced 0:15 (933:51 → 934:06), dropping the IFR leg. He caught
    it at the page total and rewrote 14:07 → 15:59; the bottom-line cumulative (940:54) already
    includes the fix. Our row sum reaches 15:59 independently, confirming both his correction and the
    2:07 reconstruction. **The book's per-row running-Total column stays 1:52 low from row 12 onward
    (on top of the p.14 1:00 slip) — but its printed page/bottom totals are correct, so our steady
    +1:25 Total drift is unchanged.** Always cross-check on "TOTAL THIS PAGE", never the row column.
  - **Row 12 is OH-PIF flown PIC, not student (user-confirmed 2026-07-31).** Every prior OH-PIF row in
    Books 2–3 was student time (Autere/Tarhanen IR work), so this is the exception worth remembering:
    PIC 2:07 + Instrument 1:52, pic_name=self. Book has AYOUB in the PIC column, Dual column struck.
  - **Row 13 (15/07 SR20 EFNU→EFTU) on-block = 13:16 (user-confirmed);** book's cell reads 13:46, which
    would leave a 3-min turnaround before the 13:49 departure to Sweden. Logged time 0:46 either way.
  - **Date corrections (user-confirmed):** row 8 is **07/07/2023**, not the book's "09/07" — rows 6–9
    are one continuous EFHV→EFRY→EFFO→EFLA→EFHV day, pinned by the block times (EFFO 13:32 arr →
    14:25 dep; EFLA 15:12 arr → 15:31 dep). Row 4 is **29/06/2023**, same day as row 3 (book's digits
    read as 27/06); user confirmed **29/06** explicitly.
  - **Sweden round-trip, this time with a return leg:** 15/07 EFNU→EFTU→**ESNU** (Umeå) and 17/07
    ESNU→EFTU, all SR20 OH-ESR, IFR 1:56 + 1:50. (Contrast the 15/04/2023 one-way in IMG_6015.)
  - **Instructing (3):** 19/06 + 20/06 OH-CTL Tuusulanjärvi float locals (1:31 + 1:36, 9 and 5 ldg)
    and 29/06 OH-CAY EFHV local (1:02). Δinstr 4:09 exact.
  - One block warn: 29/06 OH-CAY 11:45–12:42 = 0:57 vs logged 1:02 (5 min; logged authoritative).
  - Landings page-total cell blank again → no paper check; ours sums to **35**. SEP_Sea +3:07 → 261:02.
- **IMG_6018 (pages 23–24, 12/07/2023–07/09/2023) — appended 2026-07-31.** 15 flights. Image
  **already upright**. **All 4 cross-checks exact, zero block warnings** (Δtotal 14:02, Δpic 12:57,
  Δstudent 1:05, Δinstr 1:08). No night, no SE-IFR (book's SE-IFR total stays 79:44).
  - **🎓 CRI(A) RATING EARNED — row 13, 04/09/2023, C172 **OH-GKT** Kahvisaari local, 1:05, 6 ldg.**
    Book remark: *"AoC for CRI(A) Passed"* + examiner signature + **FI.FCL.34041**; PIC column reads
    **RAVANTTI**. Logged under Dual → `Student_Time 1:05`, `pic_name = Ravantti`. (AoC = Assessment of
    Competence; CRI(A) = Class Rating Instructor, Aeroplane.) **New name: Ravantti.** First OH-GKT
    flight of Book 3 — the pilot's own seaplane, last seen in Book 2.
  - **⚠ MIXED TIME ZONES AGAIN — second occurrence in Book 3 (user-confirmed 2026-07-31).** On 04/09
    row 12 (OH-CTL Tuusulanjärvi→Kahvisaari 11:25–12:15) *arrives* Kahvisaari at 12:15, yet row 13
    (OH-GKT Kahvisaari local) runs 11:05–12:10 — impossible. User: **"GKT is UTC time"** → row 13
    keeps `Z`, **row 12 stored plain local** (11:25–12:15 LT = 08:25–09:15Z, which slots cleanly before
    the checkride). **Row 14 (OH-CTL Kahvisaari→Tuusulanjärvi 16:53–17:35) was left as `Z`** — it has
    no time conflict either way, so it was transcribed as the column header states. That is a residual
    uncertainty: same aircraft, same day as the local row 12. Worth asking if it ever matters.
  - **Row 5 date = 12/07/2023 (user-confirmed): "an older flight I missed."** The book's cell is
    overwritten and sits physically between 25/07 and 06/08 — it was entered late, out of order.
    Kept in book row order. Rows 6 (06/08) and 7 (05/08) are likewise inverted, as written.
  - **Lapland float ferry:** 19/08 OH-CTL **Ranua → Viitasaari** 2:23 (place read, not user-confirmed).
    New float places this page: **Ranua, Viitasaari** (Kahvisaari already known).
  - **Instructing (1):** 12/07 OH-CTL Tuusulanjärvi local 1:08. Δinstr 1:08 exact.
  - **Book's running-Total column starts 0:36 low on this page** — the pilot carried 939:02 forward and
    wrote it again on row 1 instead of adding row 1's 0:36. Internally consistent from row 2 on, and his
    "TOTAL PREVIOUS PAGES 940:54 + TOTAL THIS PAGE 14:02 = 954:56" bottom line is right, so nothing
    propagates to us. (Third distinct arithmetic slip in that column: p.14 1:00, p.21 1:52, p.23 0:36.)
  - Landings page-total cell blank for the fifth time → no paper check; ours sums to **42**.
    SEP_Sea +10:25 → 271:27.
  - **Paper-vs-ours drift at 07/09/2023** (book bottoms Total 954:56, PIC 796:27, SE-IFR 79:44,
    Dual 160:59, Flight-Instructor 95:50): ours Total 956:21 (**+1:25**), PIC 795:02 (**−1:25**),
    Instrument 83:46 (**+4:02**), Student 161:19 (**+0:20**), Instructor 94:30 (**−1:20**).
    **All five deltas held exactly across both spreads** — independent confirmation of the whole batch,
    including the 2:07 reconstruction on IMG_6017 row 12.
- **IMG_6019 (pages 25–26, 24/08/2023–26/10/2023) — appended 2026-07-31.** **14** flights (one row
  struck, below), all UTC (`Z`). Image already upright. **All 4 cross-checks exact** (Δtotal 15:04,
  Δpic 13:42, Δstudent 1:22, Δinstr 1:20). No night this page.
  - **⚠ STRUCK ROW = DUPLICATE OF AN ALREADY-APPENDED FLIGHT — excluded.** Row 2 reads
    *07/09/23 Tuusulanjärvi→Tuusulanjärvi, 14:51–16:00, C172(sea) OH-CTL, PIC 1:20, 6 ldg* and is
    **ruled through across both pages**. It is the same flight as **IMG_6018 row 15** (07/09/2023
    Tuusulanjärvi local, 1:20, 6 ldg — logged there as 17:51Z–19:11Z). The pilot re-entered it at the
    top of the new page, then voided it. **Not in `logbook_3.csv`** — do not "recover" it. The page
    total 15:04 only reconciles with the row excluded.
  - **The Total column on this page is scribbled illegible; the pilot wrote "\* 15:04 ←" in the left
    margin pointing at TOTAL THIS PAGE.** That note is the only legible source for the page total, and
    it is corroborated two ways: our 14-row sum = 15:04, and the book's own SE-VFR 14:04 + SE-IFR 1:00
    = 15:04. The bottom TOTAL cell is unreadable, but p.28's carry (970:00 = 954:56 + 15:04) confirms it.
  - **NEW AIRCRAFT OH-MIL — a Maule on floats, type written "M6(sea)" (user: "always on floats").**
    First OH-MIL flight in any of the three books; it was already in the SEAPLANES list, so SEP_Sea
    picks it up. Row 1, 24/08/2023 **Tuusulanjärvi → Keilaniemi** (place user-confirmed), 1:22,
    **STUDENT with Sinervä** (user-confirmed), 7 landings. Fourth Sinervä seaplane dual.
  - **⚠ THE PILOT STRUCK HIS OWN PIC TOTAL ON THIS PAGE.** The bottom PIC cell reads **810:09 crossed
    out**, and p.28 carries **807:39** — a deliberate **−2:30** hand correction. User (2026-07-31):
    *"I was recalculating by hand and corrected something from earlier. I don't remember details."*
    Our per-page Δpic reconciled **exactly** on both this page and the next (13:42, 10:56), so our
    row-built PIC is internally sound and was kept. **Effect: the standing PIC drift flips from
    −1:25 to +1:05.** Unresolved whether the 2:30 was fixing a real over-count that our series also
    inherited, or an arithmetic slip that only ever lived in the book's column. The book's PIC column
    has over-added before (IMG_4929 +8 min, IMG_4933, IMG_6011 +1:58), which favours the latter.
  - Two block warns (logged authoritative): 24/08 OH-MIL 1:24 vs 1:22; 19/10 OH-PDP 0:30 vs 0:33.
  - **New airport EFVP (Vampula)** — user-confirmed; 08/09 OH-PDP EFHV↔EFVP day-return.
  - The book's own TOTAL PREVIOUS PAGES cell here reads ~953:54 against the 954:56 that p.24 ended on
    — a transient carry slip that it corrects by p.28. Cosmetic; nothing propagates.
  - Landings page-total cell blank again → no paper check; ours sums to **46**. SEP_Sea +2:47 → 274:14.
- **IMG_6020 (pages 27–28, 26/10/2023–05/03/2024) — appended 2026-07-31.** 15 flights, all UTC (`Z`).
  **All 3 available cross-checks exact** (Δtotal 10:56, Δpic 10:56, Δinstr 3:59). No dual, no SE-IFR.
  **Crosses into 2024.** Almost entirely EFHV circuit/local work — the float season is over.
  - **Three new regs:** **OH-AWB** (C152), **OH-CMU** (C152), plus OH-CGT ruled out (below).
    **OH-CMU is genuinely distinct from OH-CMV** (C152, first seen 26/12/2021) — user-confirmed, not
    a misread. Both are C152s; watch the last letter.
  - **Reg read resolved:** rows 6 & 8 (27/12, 28/12 C172 at EFHV) end in a letter that looks more like
    **T** than the X on the other rows — user confirms all of them are **OH-CGX**. There is no OH-CGT.
  - **Row 12 type corrected (user):** the book writes **C172** for the 09/02/2024 OH-AWB night flight,
    but OH-AWB is a **C152** (as on rows 5/9/10) — *"my mistake."* Stored as C152.
  - **Two night flights:** 09/02/2024 OH-AWB EFHV local 0:39 night (5 night ldg) and 05/03/2024
    **OH-CMU** EFHV local 19:32–19:55 **0:23 night**. ⚠ On that last row the book puts its **3 landings
    in the DAY column**, but the flight is entirely night — user confirms *"night landings, good find."*
    Under the day+night convention `Landings` = 3 either way, and `Night_Time 0:23` lets the app infer
    the split. Night page total 1:02 reconciles with the book exactly.
  - **Instructing (3):** 26/10 OH-CGX 1:40, 29/10 OH-CGX 1:09, 05/03/2024 OH-CGX 1:10. Δinstr 3:59 exact.
  - Landings page-total cell blank → no paper check; ours sums to **46**.
  - **Paper-vs-ours drift at 05/03/2024** (book bottoms Total 980:56, PIC 818:35, SE-IFR 80:44,
    Dual 162:21, Flight-Instructor 101:09): ours Total 982:21 (**+1:25**), PIC 819:40 (**+1:05** —
    *flipped from −1:25 by the pilot's −2:30 hand correction on p.26*), Instrument 84:46 (**+4:02**),
    Student 162:41 (**+0:20**), Instructor 99:49 (**−1:20**). Four of five steady; only PIC moved, and
    by exactly the book's own adjustment.
- **IMG_6021 (pages 29–30, 28/03/2024–04/05/2024) — appended 2026-07-31.** 15 flights, **all UTC (`Z`)**,
  image **already upright**. **All 4 cross-checks exact, zero block warnings** (Δtotal 9:41 = SE-VFR 8:32
  + SE-IFR 1:09, Δpic 8:49, Δdual 0:52, Δinstr **nil**). No night, no seaplane, no instructing.
  - **Two student rows:** 29/04/2024 SR20 **OH-ESR** EFNU→**EFIK**→EFNU (0:28 + 0:24), Dual column,
    pic_name = **Stude** — his fourth SR20 dual. New airport **EFIK (Kiikala)**.
  - **⚠ `OH-CAM` (C172) logs 0:25 single-engine IFR** on 30/04 EFTP→EFHV: total 1:01 = 0:36 VFR + 0:25
    IFR. **User confirms OH-CAM is IFR-certified** — this is not a misread. Recorded in `reference.md`.
  - **Two inferred block cells,** both pinned by arithmetic and by the page column totals: row 10
    on-block **05:21** (last digit unreadable, 0:17 logged) and row 13 off/on **16:37 → 17:38** (both
    cells overwritten/blotted; 1:01 is the only pair consistent with the running total *and* the
    0:36+0:25 column split). User-confirmed.
  - **Book bookkeeping detour (nets out, no action):** the pilot carried "TOTAL PREVIOUS PAGES" **1:52
    low** (979:04 instead of p.28's 980:56), so his per-row running-Total column on this page is 1:52
    low throughout; he then struck the bottom line 988:45 and wrote **990:37**, which lands exactly on
    980:56 + 9:41. He likewise struck and rewrote the PIC bottom line as **827:24**. The 1:52 is the
    same IMG_6017 (p.21) SE-IFR slip he had already corrected once — he undid and redid it here.
  - Landings page-total cell blank → no paper check; ours sums **31**.
- **IMG_6022 (pages 31–32, 05/05/2024–07/06/2024) — appended 2026-08-01.** 15 flights.
  **⚠ ORIENTATION: this image is sideways and needs CCW `rotate(90)`** — 6012–6021 were all upright,
  so orientation is genuinely per-image. Always thumbnail first. **All 4 cross-checks exact, zero block
  warnings** (Δtotal 8:58 = all SE-VFR, Δpic 8:58, Δdual nil, Δinstr **1:07**).
  - **🎉 Crosses 1000 hours** — Cumulative_Total reaches **1000:02** on row 14 (07/06/2024
    Vääksy→Lietsaari, OH-GKT). Milestone; the user asked for it to be noted.
  - **⚠ Third mixed-timezone spread, and the rule is BY AIRCRAFT this time.** User: *"all rows are
    local except OH-GKT and OH-ESR."* Applied — **with one user-confirmed exception**: row 10
    (26/05 SR20 **OH-ESR** EFNU local 12:20–13:05) is stored **local**, not `Z`. As UTC it becomes
    15:20–16:05 local and breaks the day's chain — row 11 (OH-PDP EFNU→EFHV 13:40–14:05 local) already
    flew him home, so he could not depart EFNU at 15:20. Local makes 26/05 read cleanly:
    EFHV→EFRY→EFNU, SR20 local at Nummela, EFNU→EFHV. So: **rows 1–11 local (no `Z`), rows 12–15
    (OH-GKT) `Z`.** *Lesson: a per-aircraft timezone rule can still have per-row exceptions — check
    each day's chain for ordering conflicts before applying it wholesale.*
  - **Row 12 date corrected:** the book's day digit reads "34" (impossible); user confirms **30/05/2024**.
    (First answer "30/04" was a slip — 30/04/2024 is fully occupied on p.29–30, including a 16:37–17:38Z
    row that would overlap this 16:05–17:05 flight.)
  - **Type `C172sea` is NOT a type.** From this page the book writes `C172sea` for the float C172s —
    the pilot's own informal seaplane marker (*"my own (bad way) of marking seaplanes"*). **Stored as
    plain `C172`**; the SEP_Sea flag comes from the registration. Applies to all OH-GKT rows here.
  - **Float season 2024 opens** on OH-GKT in the Päijänne/Vääksy area. New places **Lieso** and
    **Lietsaari** (user-confirmed); **Vääksy** already known from Book 2.
  - **One instructing row:** 05/05 **OH-CAY** EFHV local 1:07 (PIC + Instructor), 4 landings.
  - **Late-evening rows are day, per the book:** rows 4/5/7 land 20:08 / 19:45 / 20:22 — these are
    **local** times (see the timezone note), so they are ordinary late-May evening flights, not night.
  - Landings page-total cell blank → no paper check; ours sums **36**.
- **IMG_6023 (pages 33–34, 17/06/2024–28/06/2024) — appended 2026-08-01.** 15 flights.
  **⚠ ORIENTATION: sideways again, CCW `rotate(90)`** (same as 6022). **All 3 cross-checks exact and
  every block time matched its logged time to the minute — zero warnings** (Δtotal **15:01** all
  SE-VFR, Δpic 15:01, Δinstr **12:37**). No night, no SE-IFR, no dual.
  - **Pure float-instruction spread.** 13 of 15 rows are **OH-CTL** at Tuusulanjärvi; rows 14–15 are
    **OH-GKT** in the Päijänne/Lahti area. **Instructing ×11** — rows 1–9 (17–18/06 Tuusulanjärvi
    locals plus the 18/06 Hiidenvesi out-and-back) and rows 12–13 (27/06 Tuusula↔**Pellinki**).
    Rows 10/11 (Tuusula↔Kahvisaari ferry) and 14/15 are PIC only.
  - **Time zones — the IMG_6022 per-aircraft rule held, no exception needed this time** (user-confirmed):
    **OH-CTL rows 1–13 stored local (no `Z`); OH-GKT rows 14–15 stored `Z`.**
  - **Row 7 (18/06 Tuusulanjärvi local, 14:29–15:12, 0:43) is entered OUT OF ORDER** — it sits in the
    book between the 16:16 and 18:55 rows. User confirms the times are as written in the same zone as
    its neighbours; it is **not** a zone mix. (As UTC it would have slotted perfectly at 17:29–18:12
    local — a tempting but wrong reading. Ask, don't infer.)
  - **Row 8's arrival place is struck through**; user confirms **Hiidenvesi** (row 9 departs from there).
  - **🏠 `Kahvisaari` is OH-GKT's HOME BASE, near Lahti** (user-confirmed) — *not* a Saimaa/
    Hillosensalmi lake as earlier notes implied. That makes Tuusulanjärvi↔Kahvisaari (~0:40) and
    **Padasjoki**→Kahvisaari (0:27, "normal in good wind") routine ferry hops, not anomalies.
    New place **Padasjoki** (Päijänne).
  - **The running-Total column on this page is built off the struck 997:43 carry** (1:52 low — the same
    slip as p.30), but the pilot corrected it at the bottom line: 999:35 + 15:01 = **1014:36**. The
    per-page and bottom-line totals are sound; the running column is not. No action.
  - Landings page-total cell blank → no paper check; ours sums **65**.
- **IMG_6024 (pages 35–36, 28/06/2024–12/07/2024) and IMG_6025 (pages 37–38, 12/07/2024–21/07/2024)
  — appended 2026-08-01.** 15 flights each; both **sideways, CCW `rotate(90)`**. All six cross-checks
  exact (6024: Δtotal 14:05 / Δpic 14:05 / Δinstr **11:47**; 6025: Δtotal 11:52 / Δpic 11:52 /
  Δinstr **5:13**). All SE-VFR — no night, no IFR, no dual on either page. Every row is a float row
  (OH-CTL or OH-GKT). Instructing: **11 rows on 6024** (2–6, 8–10, 12, 13, 15) and **7 on 6025**
  (3–6, 13–15). Landings cells blank on both → ours sum **74** and **65**.
  - **⚠⚠ THE BOOK'S `LT` SUBSCRIPT IS REAL — first sighting.** On 6024 a tiny handwritten **`LT`**
    sits beside the block-time minutes of **rows 2–6** (all OH-CTL); nowhere else on either spread.
    **Its absence means nothing though** — IMG_6023 has no `LT` marks at all and the user confirmed
    those rows are local. Positive evidence only. See `reference.md`.
  - **⚠ The `LT` subscript can corrupt the digit it follows.** 6024 rows 2 and 5 both read as
    off-block `…29` and both came out **exactly 9 minutes short** of the logged time. The true value
    is **`20` + `LT`**: row 2 = 19:20→20:50 (1:30), row 5 = 17:20→19:15 (1:55).
    **Confirmed by an independent electronic record the user supplied** for 01/07/2024 OH-CTL:
    off-block **16:20Z**, on-block **17:50Z**, block 90 min, flight 75 min, 9 landings, *Opettaja*
    Rami Ayoub / *Oppilas* **Ignaty Romanov-Chernigovsky**. 16:20Z + 3h = 19:20 local — so the paper
    row is local and the off-block minute is 20. *When two rows on a page are short by the same odd
    amount, suspect one systematic digit misread, not two independent slips.*
  - **Time zones: user says ALL 30 rows on both spreads are LOCAL** (no `Z`) — including the OH-GKT
    rows, which is a departure from the 6022/6023 per-aircraft rule. Corroborated by the electronic
    record above. **Don't carry the per-aircraft rule forward blindly; ask each spread.**
  - **6024 row 11's instructor entry is STRUCK OUT** by the pilot (05/07 Kahvisaari→Tuusulanjärvi
    0:39). Excluded from Instructor_Time — the page's 11:47 instructor total only works without it.
  - **New places: `Pulkkilanharju`** (Päijänne, 6024) and **`Mäntyharju`** (6025). Kelvenne,
    Leikonvesi, Lietsaari, Vääksy, Pellinki, Hiidenvesi all already known.
  - The running-Total column stays **1:52 low** on both pages (the struck 997:43 carry from p.30);
    the pilot again corrected it at each bottom line — 1014:36 + 14:05 = **1028:41**, then
    1028:41 + 11:52 = **1040:33**. Per-page and bottom-line totals sound; running column is not.
- **Paper-vs-ours drift at 21/07/2024 (end of p.38)** — book bottoms Total **1040:33**, SE-VFR 958:40,
  SE-IFR **81:53**, PIC **877:20**, Dual **163:13**, Flight-Instructor **131:53**. Ours: Total
  **1041:58** (**+1:25**), Instrument **85:55** (**+4:02**), PIC **878:25** (**+1:05**), Student
  **163:33** (**+0:20**), Instructor **130:33** (**−1:20**). **All five steady** across
  IMG_6021 through IMG_6025 — no new divergence.
- **Book-3 landing drift now open & growing (day+night vs book's day-only).** At 10/06/2022 our
  Cumulative_Landings = **2172** (day+night). The book's printed day-only cumulative is lower by the running
  night-landing total (3 on 17/12/2021 + 8 on IMG_6010 = 11 so far) plus the ~2-landing IMG_6009 struck-cell
  slip. Cross-check on per-page Δland (day+night sum) only; do not expect our cumulative to equal the book's.

## IMG_6026/6027 batch (pages 39–42, 21/07/2024–19/09/2024) — appended 2026-08-01
Both spreads **sideways** (CCW `rotate(90)`), no `LT` subscripts anywhere. **All cross-checks exact**
(6026: Δtotal 18:33 / Δpic 16:51 / Δstudent 1:42 / Δinstr 7:21; 6027: Δtotal 12:22 / Δpic 12:22 /
Δinstr **7:41**, see below). Landings cells blank on both → ours sum **34** and **61**.

- **⚠ TIME ZONES: back to UTC after two all-local spreads.** All 30 rows stored with `Z`. Proved on
  five independent rows, each exactly **+3h** to the electronic local time — and one of them,
  19/09/2024 OH-GKT, matched an **Aviatron row stamped `14:04:00 UTC`, digit-for-digit with the paper
  cell**. That is the strongest zone evidence we have had; it is not an inference. Confirms once more
  that the zone flips **per spread** (6024/6025 local → 6026/6027 UTC) and must be re-established
  every time.
- **⚠ Book's Flight-Instructor page total on p.42 is 3 min high (7:44 vs true 7:41).** Row 7
  (30/08/2024 OH-MIL Tuusulanjärvi→Lohja) is written **0:44** in SE-VFR, PIC *and* Flight Instructor,
  but the block cells (`08:44→09:25`), the running-Total column (+41) and **both** page totals
  (12:22) all require **0:41**. User confirmed **0:41**. Only the FI page total needs 44 → the book's
  own addition slipped. **Our Instructor drift therefore moves −1:20 → −1:23.**
- **Row 7 pilot-name cell:** `SINERVÄ` is **struck through** over `Ayoub`. User confirms **PIC is
  Ayoub** — this is an instructing row (PIC + FI), *not* a Maule dual like 24/08/2023. Don't let the
  Sinervä name pull it into Student_Time.
- **Student row (6026 row 9):** 02/08/2024 SR20 OH-ESR EFNU→EFHV **1:42**, book logs Dual + SE-IFR,
  remark `(TAR)` → **Tarhanen**, another IR revalidation. `Student_Time 1:42` + `Instrument_Time 1:42`,
  `pic_name = Tarhanen`.

### Three inferred block cells (logged flight time authoritative in all three)
| row | book cells | stored | basis |
|---|---|---|---|
| 6026 r14 · 06/08 Tuusulanjärvi→Haikko 0:37 | `16:16 → 16:48` | on-block **16:53** | electronic record `19:16→19:53` local, block 37 |
| 6027 r3 · 14/08 Tuusulanjärvi local 1:15 | `15:35 → 17:50` | on-block **16:50** | electronic record `18:35→19:50` local, block 75 |
| 6027 r6 · 20/08 Hiidenvesi→Tuusulanjärvi 0:56 | `16:59 → 17:50` | on-block **17:55** | electronic record `19:59→20:55` local, block 56 |
- ⚠ **I initially proposed off-block `16:11` for the Haikko row and the user picked it; the electronic
  record then proved the off-block `16:16` was right and the *on-block* was the bad cell.** **All three**
  resolved rows turned out to be **on-block** errors, not off-block. *Don't offer an off-block
  candidate first just because it closes the arithmetic.*
- **⚠⚠ ROOT CAUSE FOUND — the pilot sometimes writes the LANDING time into the on-block cell.**
  Proved on 20/08: the record reads `20:50` landing / `20:55` on-block, and the book's on-block cell
  says `17:50Z` = **the landing time**. He habitually taxis ~5 min on the water at Tuusulanjärvi
  (20/08 outbound `19:25→19:30`; 06/08 leg 3 `19:48→19:53`). **So when a row is short by ~5 min and the
  cells look clean, suspect on-block = landing time and add the taxi — don't hunt for a mangled digit.**

### ⚠ `laskukierros_export.csv` was INCOMPLETE — root cause found, superseded 2026-08-01
During this batch the committed export showed **zero rows** for 06.08.2024 and 14.08.2024 although
both days flew, and held only **7** rows across Jul–Sep 2024 where the paper has ~20 OH-CTL legs.

**Root cause:** `GET /export/pilotFlights` returns only flights where the user is the **primary
pilot**. Every flight he *instructed* is filed under the **pupil's** account (`pilotName` = pupil,
`pilotTwoName` = "Rami Ayoub", `pilotTwoRole` = `instructor`) and is therefore absent. That is the
entire 128-vs-228 gap — **100 instructing rows were missing**, which is precisely the float-season
material Book 3 is full of.

**Fixed:** `GET /api/v1/flights` (same session cookie) returns all **228**, 19/04/2020 → 25/07/2026.
Saved as **`laskukierros_flights.json`** with **`laskukierros_flights.csv`** derived by
`laskukierros_to_csv.py`. All four 06/08 legs, the 14/08 flight and **both** 20/08 legs are present
and match the paper exactly. `laskukierros_export.csv` is kept only because `laskukierros_zflags.md`
was computed from it — **do not use it for coverage questions.**

### Record conflicts this batch (paper kept — documented only)
| ours (paper) | electronic record | conflict |
|---|---|---|
| `21/07/2024` OH-CTL dep **Vääksy** 17:53Z | laskukierros dep **Pulkkilanharju** 20:53 local | departure place |
| `28/07/2024` OH-CTL Tuusulanjärvi local 1:41, **6 ldg** | block 101 ✓, **5 ldg** | landing count |
| `20/08/2024` OH-CTL Hiidenvesi→Tuusulanjärvi, **3 ldg** | block 56 ✓, 3 ldg ✓ | none — full match |
| `06/08/2024` OH-CTL Tuusulanjärvi local 1:14, `14:22Z→15:36Z` | `14:22→15:36` **local**, block 74 | zone |
- That last one only *looks* like a zone mismatch: as local it would **overlap the preceding leg**
  (`13:27→14:57`) in the same aeroplane. The club-system row is the one entered in UTC by mistake;
  the paper's `14:22Z` is right and the whole 06/08 chain is UTC.

## IMG_6028/6029 batch (pages 43–46, 18/09/2024–15/04/2025) — appended 2026-08-01
Both spreads **sideways** (CCW `rotate(90)`), no `LT` subscripts. **All five cross-checks exact**
(6028: Δtotal **13:50** / Δpic 13:50 / Δinstr **2:32**; 6029: Δtotal **17:14** = SE-VFR 7:18 +
SE-IFR 9:56 / Δpic 17:14). Landings cells blank on both → ours sum **48** and **28**.
On 6028 every running-Total cell chains perfectly for the first time in a while — no column slip.
**The float season ends on p.43 and the whole batch is landplane/IFR work**; 6029 has **no
instructor time and no dual at all**, the first such spread in Book 3.

- **⚠⚠ THE `Z`/LOCAL SWITCH HAPPENS *INSIDE* IMG_6028 — pinned from both sides by club records.**
  - 17/10/2024 (rows 8–10, OH-CAM): club logs `11:35 / 12:15 / 15:00` local; book writes
    `08:35 / 09:15 / 12:00` → book is **UTC** (EEST, +3).
  - 10/12/2024 (row 14) and 04/01/2025 (row 15), same aircraft: book times match the club's
    **local** times digit-for-digit (`10:55–11:41`, `13:27–13:45`) → book is **LOCAL**.

  So the pilot stopped writing Zulu somewhere between 17/10 and 10/12/2024. **Rows 11–13
  (28/10 ×2 OH-PDP, 23/11 OH-ESR) sit inside that gap with no electronic record**; user reviewed and
  confirmed "otherwise correct", so they are stored **`Z`**. That boundary is the one soft spot in
  this batch — revisit if a record for those three dates ever turns up.
- **IMG_6029 is entirely LOCAL** (no `Z` on any row). Four rows are club-confirmed to the minute:
  04/01 OH-CAM `14:08–15:07`, 04/02 OH-CMU `12:05–13:02`, and both 18/03 OH-CAY legs.
- **Row 1's date was scribbled over — user confirms `18/09/2024`.** Two overwritten day digits, read
  variously as 26/28. It cannot be 19/09 (he was flying OH-GKT around Vääksy/Kahvisaari all that
  afternoon — see p.42) and 20/09 collides with row 5. P28A OH-PDP EFHV local 14:41–15:20, 5 ldg.
- **⚠ Registration: 04/02/2025 is `OH-CMU`, not OH-CMV.** I first read the last letter as a V; the
  club record for that exact slot (EFHV 12:05–13:02, C152, 5 ldg) says **OH-CMU**, and a high-res
  re-crop shows an open-topped U. *Both regs are real C152s — always confirm the last letter against
  the club file when the flight is in a club aircraft.*
- **The Netherlands/Germany leg is genuine, not a transcription error.** 07/03/2025 SR20 **OH-ESR
  EHGG→EDWF** (Groningen→Leer) 0:45 sits between EFNU→ESMG (12/01) and ESMG→EFNU (08/03) with no
  connecting legs. User: *"we were 2 pilots so I only logged my legs in my book."* **Expect
  geographically disconnected rows on multi-pilot ferry trips — don't treat them as errors.**
- **New airports:** **ESMG** (Feringe/Ljungby, Sweden — the SR20's winter stop, 12/01→08/03),
  **EHGG** (Groningen Eelde), **EDWF** (Leer-Papenburg). ESNU (Umeå) and EFJO (Joensuu) already known.
- **Instructing (3, all on 6028):** 20/09 OH-GKT Kahvisaari↔Lietsaari 0:44 + 0:43 and 28/10 OH-PDP
  EFHV local 1:05 → Δinstr 2:32 exact. Note rows 6 & 7 (30/09 and 16/10 Kahvisaari locals, **10 and
  5 landings**) look like float instruction but carry **no FI time** — the 2:32 page total confirms
  the book logs them PIC-only. Left as written.
- **Record conflict, paper kept:** 18/03/2025 OH-CAY EFHV→EFRY — book **1 landing**, club record
  **2**. One-landing difference, no time effect.

### Two inferred block cells on IMG_6028 (logged flight time authoritative)
| row | book cells | stored | basis |
|---|---|---|---|
| r3 · 20/09 Lietsaari→Kahvisaari 0:43 | `09:35 → 10:11` (= 0:36) | on-block **10:18** | logged 0:43 + page total; outbound leg was 0:44. User-approved |
| r5 · 20/09 Tuusulanjärvi→Kahvisaari 0:38 | `15:33 → 16:??` (minutes overwritten) | on-block **16:11** | only value consistent with 0:38 and the column total |
| r9 · 17/10 EFNU→EFJO 2:06 | `09:15 → ??:21` (hour digit overwritten) | on-block **11:21** | **club record** `12:15–14:21` local −3h |
- Row 3 is the **fourth on-block error in a row** across recent spreads (all three IMG_6026/6027
  corrections were on-block too). *Keep leading with the on-block candidate.*

### Paper-vs-ours drift at 15/04/2025 (end of p.46) — all five steady
Book bottoms: Total **1102:32**, PIC **937:37**, SE-IFR **98:00**, Dual **164:55**,
Flight-Instructor **149:30**. Ours: Total **1103:57** (**+1:25**), PIC **938:42** (**+1:05**),
Instrument **102:02** (**+4:02**), Student **165:15** (**+0:20**), Instructor **148:07** (**−1:23**).
**No delta moved across IMG_6028 or IMG_6029** — independent confirmation of both spreads, including
the three inferred block cells and the 18/09 date.

## IMG_6030/6031 batch (pages 47–50, 18/04/2025–02/07/2025) — appended 2026-08-01
Both spreads **sideways** (CCW `rotate(90)`), no `LT` subscripts. **All seven cross-checks exact**
(6030: Δtotal **13:59** / Δpic 13:21 / Δstudent **0:38** / Δinstr **6:12**; 6031: Δtotal **11:05**
all SE-VFR / Δpic 11:05 / Δinstr **7:27**). Every block time on both spreads matched its logged time
to the minute except one (6031 r12, below). Landings blank on both → ours sum **43** and **55**.
**🌊 The 2025 float season opens 18/05/2025** (OH-CTL Räyskälä→Tuusulanjärvi); from there the
spreads go back to the familiar Tuusulanjärvi instruction pattern.

- **⚠⚠ THE CLUB FILE CARRIED THIS BATCH.** 23 rows matched on times and **landings matched on
  23 of 23** rows that have a record (13/13 on 6031) — plus every Flight-Instructor row was
  independently corroborated by the record's `rami_role=instructor` and its pupil name (Salo,
  Storgårds, Nirkkonen, Puhakka, Kere). *This is the strongest external validation a Book-3 batch
  has had. Always grep it first.*
- **⚠ ZONES: 6030 is UTC except its LAST row; 6031 is genuinely mixed.** All user-confirmed.
  | rows | zone | evidence |
  |---|---|---|
  | 6030 r1–r14 | **UTC** (`Z`) | 11 rows club-confirmed at exactly +3h |
  | 6030 r15 · 29/05 Tuusulanjärvi→Pellinki | **local** | book `16:15–16:51` = club local digit-for-digit; as UTC it lands 19:51 local and collides with that evening's two other club flights |
  | 6031 r1 · 29/05 Pellinki→Tuusulanjärvi | **local** | return half of the same trip, club-confirmed |
  | 6031 r2–r8, r11, r13, r14 | **UTC** (`Z`) | 10 rows club-confirmed at +3h |
  | 6031 r9, r10 · 08/06 Tuusulanjärvi↔Hiidenvesi | **local** | book times = club local exactly. ⚠ Unlike the 29/05 case there is **no second flight that day to chain against**, so this could also be a club entry mistakenly filed in UTC (cf. 06/08/2024). User approved storing local; residual uncertainty. |
  | 6031 r12 (OH-GKT), r15 (OH-PDP) | `Z` | no record; left as the column header states |
  **The zone now flips within a single spread and back again inside the next one.** Re-derive it
  row-by-row from the club file, not per spread.
- **⚠ 6030 row 7 is a DUAL row with no instructor name recorded.** 19/05/2025 OH-CTL
  Tuusulanjärvi→Hirvijärvi 0:38, 3 ldg. The book puts 0:38 in the **Dual** column and leaves **PIC
  blank** (the 13:21 PIC page total confirms it), yet writes `AYOUB` in the pilot-name cell like
  every other row — *the name cell is habit, not evidence; trust the time columns.* It is also the
  only flight of that day absent from the club file, consistent with being filed under the other
  pilot's account. The **return leg 20 minutes later (r8) is PIC + Instructor.** Most likely a
  seaplane spring check / class-rating revalidation. **`pic_name = Sinervä`** (user-supplied
  2026-08-01, after the append) — the seaplane instructor's **fifth** dual with him
  (30/04/2019, 15/07/2020, 13/05/2022, 12/05/2023, 24/08/2023 Maule, 19/05/2025).
  *Nothing else changed: Student_Time was already 0:38 and every cumulative was already correct;
  only the name cell was blank.*
- **6030 row 1's landings cell is blank** (18/04/2025 OH-PDP EFHV local, 1:10). Counted as **1** per
  the IMG_6013 precedent, user-approved — but a 1:10 Arrow local is plausibly circuits, so this is
  the one soft number in the batch. Row-1-blank was pinned by mapping the other 14 landing cells
  against 10 club records; the column is visually offset ~1.5 rows from the left page (spine curl),
  so **do not read landings by eyeballing row alignment — verify against the record.**
- **6031 row 12 — the 5-minute on-block pattern again.** 17/06/2025 OH-GKT Kahvisaari local: cells
  read `16:56 → 17:48` = 0:52, but the logged time and the page total both need **0:57**. Stored
  **on-block 17:53** (water taxi; see the IMG_6026/6027 root-cause note). No OH-GKT record to confirm.
- **`OH-COF` returns** (C152, EFNU local 16/05/2025, 1:02, logged PIC + **Instructor**) — *not* a new
  registration: it is already in `reference.md`'s Book-1/2 C152 list. First OH-COF row in Book 3.
  Not in the club file (EFNU aircraft, different operator). User-confirmed.
- **New places:** **Hirvijärvi**, **Loppijärvi**, **Kytäjärvi** (all lakes near Tuusula/Hyvinkää,
  all club-corroborated). Haikko, Hiidenvesi, Lohja, Pellinki, Räyskälä already known.
- **Date slip:** 6031 row 7 is written `5/6/15`; the year digit is a slip — it is **05/06/2025**
  (bracketed by its neighbours and club-confirmed).
- Rows 2–6 of 6031 carry no year at all in the date cell (`30/05`, `1/06`, `3/06`) — they continue 2025.

### Paper-vs-ours drift at 02/07/2025 (end of p.50) — all five steady
Book bottoms: Total **1127:36**, PIC **962:03**, SE-IFR **101:12**, Dual **165:33**,
Flight-Instructor **163:09**. Ours: Total **1129:01** (**+1:25**), PIC **963:08** (**+1:05**),
Instrument **105:14** (**+4:02**), Student **165:53** (**+0:20**), Instructor **161:46** (**−1:23**).
**Unmoved across four consecutive spreads now** (IMG_6028 → IMG_6031).

## IMG_6032/6033 batch (pages 51–54, 01/07/2025–15/08/2025) — appended 2026-08-01
Both spreads **sideways** (CCW `rotate(90)`), no `LT` subscripts. **All six cross-checks exact after
two corrections** (6032: Δtotal **13:47** / Δpic 13:47 / Δinstr **7:01**; 6033: Δtotal **13:37** /
Δpic 13:37 / Δinstr **6:31**). Landings cells blank on both → ours sum **40** and **65**.
All 30 rows **UTC (`Z`)** — every club- and Aviatron-covered row sits exactly 3 h behind local.

- **⚠⚠ AVIATRON COVERS `OH-GKT` THROUGH 07/2026 — the docs undersold it badly.** `reference.md`
  described `Aviatron.pdf` mainly as the CB-IR / `OH-PIF` cross-reference; in fact its 126 flights
  include **every OH-GKT float row**, i.e. exactly the rows `laskukierros_flights.csv` can never
  reach (club fleet only). On this batch it arbitrated **8 of 8** GKT rows and caught a real book
  error. **Grep it alongside the club file, not after it** — extract once with
  `pdftotext -layout Aviatron.pdf`. Its records give block **and** airborne times: the flight header
  line (`ID / LÄHTÖP / OFF / ON / BLOCK / LASK`) is the block record; the `RIVI` line under it is the
  airborne time. Compare against **BLOCK/LASK**, not the RIVI figures.
- **⚠ 6032 row 11 — the book over-logged a flight by 0:20 and its landings by 3.**
  08/07/2025 OH-GKT Kahvisaari→Kelvenne. Book: on-block `16:36`, **0:57**, **7 ldg**, and it carried
  the 0:57 through both the running column and the page total — so this is the **pilot's own error,
  not a misread**. Aviatron **35472**: `15:53–16:30 UTC`, block **0:37**, **4 landings**. Stored as
  Aviatron (user-directed: *"row 11 seems like a real mistake, here are the authoritative ones"*).
  **Probable cause: 0:57 / 7 ldg is exactly the preceding OH-GKT row** — 17/06/2025 Kahvisaari local
  (IMG_6031 r12, itself Aviatron-confirmed at 16:56–17:53, 0:57, 7). The figures look **copied down
  one row**. *When a row's time and landings both duplicate the previous same-registration row,
  suspect a copy-down before trusting it.*
  ⚠ Note this also **retro-validates IMG_6031 r12's inferred on-block 17:53** — Aviatron 35236 has
  that flight at 16:56–17:53 exactly. The inference was right.
- **⚠ 6032 row 13 — running-Total slip, the fifth.** 14/07/2025 OH-CTL Tuusulanjärvi local is a
  **1:38** flight (club: 98 min block; Aviatron n/a), but the running column adds it as 0:38
  (1138:54 → 1139:32), so the printed page **Total and PIC of 13:07 are both 1:00 low**.
  Caught by the **directly-summed Flight-Instructor column reconciling at 7:01 exactly**, which only
  works with row 13 at 1:38. Slip list is now p.14 −1:00, p.21 −1:52, p.23 −0:36, p.26 PIC −2:30,
  **p.52 −1:00**.
  Net of both errors our page is **13:47** vs the book's printed 13:07: **+1:00 − 0:20**.
- **6033 row 8 on-block corrected** 17:28 → **17:32Z** (12/08/2025 OH-GKT Kahvisaari local).
  Aviatron 36429; the book's logged 1:01 and 6 landings were right, only the cell was 4 min out.
- **Six GKT rows confirmed digit-for-digit** by Aviatron: 6032 r12 (35473), 6033 r4 (36200),
  r5 (36201), r9 (36459), r10 (36460), r11 (36461). **The 17 landings on 13/08 in 1:15 are real.**
- **6032 row 14 is `09/07/2025`, entered out of order** between the 14/07 and 21/07 rows
  (user-confirmed). Single-digit `9`, not `19`. OH-PDP is in neither electronic record.
- **6033 row 1's on-block cell is overwritten** → resolved **09:15Z** (club: 11:40–12:15 local,
  35 min = the book's logged 0:35).
- **Residual warning:** 6033 r2 (23/07/2025 OH-PDP EFLA→EFHV) cells give 0:23, book logs **0:25**.
  No record either way; logged time kept, page totals reconcile on it.
- **⚠ `EFSA` was NOT a new airport** — I called it new off `reference.md`'s "airports seen" list.
  Savonlinna appears **6 times** in the finished books (`15/08/2012` OH-CWB EFLP↔EFSA;
  `12/05/2018` and `18–19/07/2020` OH-PDP EFHF↔EFSA). **That list is a hand-maintained running note,
  not derived from the CSVs — `grep` the three CSVs before calling any place or registration new.**
- Instructing rows all corroborated: 6032 has **8** (club pupils Puhakka ×6, Joukanen, Kere),
  6033 has **6** (Puhakka, Kere + the four Kahvisaari GKT details, Aviatron-confirmed).
  15/08/2025 is a four-leg OH-PDP day EFHV→EFNU→EFJO→EFSA→EFHV.

### Paper-vs-ours drift at 15/08/2025 (end of p.54)
Book bottoms (p.54): Total **1154:20**, PIC **988:47**, SE-IFR **101:12**, Dual **165:33**,
Flight-Instructor **176:41**. Ours: Total **1156:25** (**+2:05**), PIC **990:32** (**+1:45**),
Instrument **105:14** (**+4:02**), Student **165:53** (**+0:20**), Instructor **175:18** (**−1:23**).
**Total and PIC each moved +0:40** (= the p.52 slip +1:00, less the row-11 over-log −0:20).
Instrument, Student and Instructor unmoved since IMG_6027.

## IMG_6034/6035 batch (pages 55–58, 15/08/2025–31/01/2026) — appended 2026-08-01
Both spreads **sideways** (CCW `rotate(90)`), no `LT` subscripts. **All eight cross-checks exact and
zero corrections** (6034: Δtotal **13:11**/Δpic 12:32/Δdual **0:39**/Δinstr **2:14**; 6035: Δtotal
**13:03**/Δpic 13:03/Δinstr **1:01**, incl. Night 0:40). **Every one of the 30 block times matches
its logged time to the minute** — the first zero-warning batch of Book 3. Landings sum **39**, **44**.
All 30 rows **UTC (`Z`)**. **20 of 30 rows externally confirmed; landings matched 20 of 20.**

- **🎉 1000 HOURS PIC on 19/09/2025** — 6034 row 11 (OH-CTL Kahvisaari→Tuusulanjärvi) lands
  Cumulative_PIC on **1000:20**. (Cumulative_Total passed 1000 on 07/06/2024, IMG_6022.)
- **⚠⚠ THE LOCAL↔UTC OFFSET IS +3 IN SUMMER AND +2 IN WINTER — and DST ends *inside* IMG_6035.**
  Finland is EEST (UTC+3) Mar–Oct, EET (UTC+2) Oct–Mar; DST ended **26/10/2025**. Row 6 (20/10)
  reconciles against the club file at **−3h**; rows 12/13/14 (12/12, 01/01, 26/01) at **−2h**. Both
  readings put the book on UTC. **A club or Aviatron row that looks exactly 1 h "wrong" across the
  late-Oct or late-Mar boundary is daylight saving, not a mis-logged row** — this had not come up
  before because every prior spread sat inside one season.
- **6034 row 1 — DUAL with `pic_name = Sinervä`** (user-confirmed 2026-08-01). 15/08/2025 **OH-MIL**
  Maule Tuusulanjärvi→Hiidenvesi 0:39, 1 ldg. Unusually the book *names the other pilot*: the
  pilot-in-command cell reads **SINERVÄ** (every other row says AYOUB) and the 0:39 sits in the
  **Dual** column with PIC blank — the page PIC total 12:32 (= 13:11 − 0:39) confirms it.
  *This is the first Book-3 dual row where the book itself supplies the instructor's name;
  on 6030 r7 the cell still said AYOUB out of habit.* Sinervä's **7th** seaplane dual with the pilot
  and his **2nd on the Maule** (first was 24/08/2023 Tuusulanjärvi→Keilaniemi).
  Neither electronic record covers OH-MIL.
- **⚠⚠ 6034 row 6 — the book logged the AIRBORNE times in the off/on-block cells.**
  08/09/2025 OH-CTL Inkoo→Tuusulanjärvi. Club has block `14:20–15:05` (45 min) **and airborne
  `14:23–15:01` (38 min)**; the book's block cells hold the **airborne pair digit-for-digit** and it
  logs **0:38**. The pilot confirms the entry is correct as flown and will add it to Aviatron later.
  **User decision (2026-08-01): store the CLUB BLOCK TIMES.** Stored as:
  | col | value | |
  |---|---|---|
  | `Off_Block` / `On_Block` | `14:20Z` / `15:05Z` | the club's block pair |
  | `Takeoff` / `Landing` | `14:23Z` / `15:01Z` | the airborne pair the book had written in the block cells |
  | `Block_Time` | **`0:45`** | off→on |
  | `Total_Time` | **`0:38`** | unchanged — the flown time the book totals on |
  **This is the first row in any book where `Block_Time ≠ Total_Time`, and the first use of the
  `Takeoff`/`Landing` columns.** Both were in the 26-col schema from the start and were simply never
  needed. Nothing downstream moves: `Total_Time` is untouched, so every page Δ and every cumulative
  is unaffected — the page still reconciles at 13:11.
  ⚠ **A new failure mode to recognise:** previously a short row meant the *on-block* cell held the
  landing time — **one** bad cell. Here **both** cells are from a different clock. **When a row runs
  short, check the record's airborne pair as well as its block pair before proposing a fix.**
- **6034 rows 9–11 · 19/09/2025 — a two-aircraft ferry day; OH-CTL moves with no logged leg.**
  CTL Tuusulanjärvi→Pyhäjärvi, then **GKT** Pyhäjärvi→Kahvisaari, then CTL Kahvisaari→Tuusulanjärvi.
  The CTL is left at Pyhäjärvi and reappears at Kahvisaari. Same pattern as the EHGG→EDWF legs
  (IMG_6029): OH-GKT goes home to Kahvisaari for the winter, a second pilot brings the CTL along,
  both ride it back — **only his own legs are logged.** All three rows corroborated as written
  (club ×2, Aviatron 36897). Not an error.
- **6035 row 12 · 12/12/2025 — night flight with the landings in the RIGHT column for once.**
  OH-CAM EFHV local 0:40, Night 0:40, book puts **3 in the NIGHT column and leaves DAY blank**;
  club confirms `ldg_day 0, ldg_night 3`. Stored as 3 under the day+night convention.
  (Contrast IMG_6020 r15, where night landings were written into the day column.)
- **Out-of-order rows, both as written:** 6034 r15 (29/09 11:20Z, logged after the 13:45 and 14:16
  rows) and 6035 r11/r12 (13/12 logged before 12/12).
- **2025 float season closes 06/11** (6035 r9, OH-GKT Kahvisaari→EFRY for the winter); from r10 it is
  all EFHV/EFNU landplane work. **The book crosses into 2026** on 6035 r13.
- **New pupil `Koskinen`** (Aviatron 37229, the 31/10 instructing row). **New place `Vuolenkoski`**
  (6034 r7–r8) — the only genuinely new one; EFPR, Pyhäjärvi, Inkoo, Padasjoki, Vääksy and EFRY were
  all checked against the three CSVs first. ⚠ **`EFPR` and `EFPO` are different fields** and both
  appear in the books — the 27/08 SR20 pair really is EFPR.
- **6034 r2 (22/08 OH-GKT Kahvisaari local 1:14, 8 ldg, instructing) is ABSENT from Aviatron**, though
  the 15:33 flight the same afternoon is there (36607). **Aviatron's OH-GKT coverage is very good but
  not complete** — absence from it is not evidence against a row.

### Paper-vs-ours drift at 31/01/2026 (end of p.58) — all five steady
Book bottoms (p.58): Total **1180:34**, PIC **1014:22**, SE-IFR **101:55**, Dual **166:12**,
Flight-Instructor **179:56**. Ours: Total **1182:39** (**+2:05**), PIC **1016:07** (**+1:45**),
Instrument **105:14** (**+3:19**), Student **166:32** (**+0:20**), Instructor **178:33** (**−1:23**).
Total, PIC, Student and Instructor all **unmoved since p.54**. Instrument moved **+4:02 → +3:19**:
the book added the 27/08/2025 SR20 leg's **0:43** SE-IFR (101:12 → 101:55) and so did we, but the
book's SE-IFR line had been running 0:43 further behind — the gap simply closes by that leg. No
cumulative is affected; nothing to correct.

> ⚠ Correction to the paragraph above (2026-08-01): our Cumulative_Instrument at 31/01/2026 is
> **105:57**, not 105:14 — read it off `logbook_3.csv`, the figure quoted here (and in an earlier
> `resume.md`) was stale. Against the book's 101:55 the Instrument drift is therefore **+4:02** and
> it did **not** move at p.58; the "+3:19" reasoning was wrong (we added the same 0:43 the book did).

## IMG_6036/6037 batch (pages 59–62, 06/03/2026–03/06/2026) — appended 2026-08-01
**29 flights** (6036: 15 · 6037: 14). Both images **sideways → CCW `rotate(90)`**; no `LT` subscripts.
**Zero block-vs-total warnings on all 29 rows** — every block pair matched its logged time to the
minute (second batch running to do so, after 6034/6035). IMG_6037 reconciled **exactly**
(Δtotal/Δpic 11:16). Landings **30** and **56**; both pages' landing cells blank → no paper check.
**These are the last two photographed spreads. The paper book is NOT finished** — p.62 of 128, and
p.61/62 has no "TOTAL THIS PAGE" filled in (page still in progress). `logbook_3.csv` therefore stays
`logbook_3.csv`; do **not** rename it `_final`.

### ⚠ IMG_6036 row 13 — the page total is 1 minute low (our value kept, user-confirmed)
01/05/2026 OH-PDP EFHV local, off/on **08:24–08:48 = 0:24**, and the book's own SE-VFR *and* PIC cells
say **0:24**. But its running-Total column adds only **0:23**, and both printed page totals
(Total 12:17, SE-VFR 12:17, PIC 12:17) inherit that. Our row sum is **12:18**.
**User decision: keep 0:24** — the block cells and the row's own SE-VFR/PIC cell all agree; the
running column is the unreliable one (sixth such slip). Consequence: **the paper-vs-ours Total and
PIC drift each move +0:01** (see below). The batch was appended with `d_total`/`d_pic` set to our
corrected **12:18**.

### ⚠ IMG_6036 rows 12–15 — the year is written `/25`, the flights are 2026
The book writes `08/5/25`, `1/5/25`, `3/5/25`, `11/5/25`. **User-confirmed typo — all four are 2026.**
Pinned independently: the running-Total column runs continuously through them, and IMG_6037 opens
15.05.**26**. Rows 13 & 14 (01/05, 03/05) also carry **margin arrows** and are **entered out of
order**, after the 08/05 row — stored as written.

### ⚠⚠ IMG_6036 is stored LOCAL throughout — and that CONFLICTS with the one club record
**User decision (2026-08-01): the whole of IMG_6036 is local; no `Z` on any of the 15 rows.**
Only one row on the spread is externally covered — 12/04/2026 OH-CAM EFHV local — and
`laskukierros_flights.csv` (whose times are local) has it at **09:58–12:10** against the book's
**06:58–09:10**, i.e. exactly **−3 h**, which reads as UTC. Landings agree (3 = 3), so it is
certainly the same flight. **Flagged to the user, who ruled the page local anyway; paper stays
authoritative.** Documented here only — no cumulative is affected. Everything else on the spread is
OH-PDP / OH-ESR / SR20, absent from both electronic references.

### IMG_6037 — zones are mixed four ways, all user-approved
| rows | dates | zone | evidence |
|---|---|---|---|
| 1–4 | 15–16/05 | **UTC (`Z`)** | club records sit exactly +3 h (r1 16:36–17:16, r2 17:20–18:15, r4's airborne 19:35–19:52) |
| 5–7 | 17–18/05 | **local** | club matches the book **digit-for-digit** (10:00–10:58, 15:26–16:10, 16:49–18:17) |
| 8 | 01/06 | **local** | user: *"I flew the Maule first then drove to fly CDK"* |
| 9 | 01/06 | **UTC (`Z`)** | same — 14:43Z = 17:43 local, after the Maule leg |
| 10–14 | 03/06 | **local** | user |
Rows 8 and 9 **overlapped as written** (OH-MIL 14:37–15:34, OH-CDK 14:43–15:48); the user resolved it
from memory. *A same-day overlap between two different aircraft is still the reliable tell for a
zone mix — and the answer can be a per-row memory, not a rule.*

### Other IMG_6037 notes
- **Row 8's type is written `M2`**; every prior OH-MIL row in the CSVs is **`M6`**. User confirms
  **M6** (Maule, always on floats). Stored `M6`.
- **Row 4 landings: book 4, club record 3.** User: **book is correct.** Kept 4.
- **Row 3** (16/05 Kabböle local 0:50, 7 ldg, instructing) has **no club record at all** — the pupil
  forgot to log it. User will have them enter it. Not an error in our row.
- **Row 2** book 14:14–15:14Z vs club 17:20–18:15 — 6 min / 1 min apart after the +3 h shift. Paper
  kept; landings agree (6 = 6).
- **New airport `EFOP` (Oripää)** and **new place `Ojakkala`** (Vihti, on Hiidenvesi) — both grepped
  against all three CSVs first, both genuinely new. Row 9 is a **C185 OH-CDK float instruction local
  at Ojakkala**, 1:05 PIC + 1:05 FI.
- **Pupils** (from `laskukierros_flights.csv`, context only — `pic_name` stays `self` on instructing
  rows): r2 **Tommi Nirkkonen**, r4 **Toivo Huovinen**, r5 **Harry Karlsson** (*"Kevätkertaus 2026"*),
  r6 **Thomas Hansson** (*"Kerhon kevättarkkari"*), r7 **Mikko Sinervä** (*"Opekertaus"* — the
  seaplane instructor is now the *pupil* on a proficiency check). New names: Huovinen, Karlsson,
  Hansson.
- **2026 float season opens 15/05/2026** (r1, OH-CTL EFRY→Tuusulanjärvi) — a fortnight earlier than
  2025's 18/05.

### Paper-vs-ours drift at 11/05/2026 (end of p.60)
Book bottoms (p.60): Total **1192:51**, PIC **1026:39**, SE-VFR 1090:56, SE-IFR **101:55**,
Dual **166:12**, Flight-Instructor **179:56**. Ours at that row: Total **1194:57** (**+2:06**),
PIC **1028:25** (**+1:46**), Instrument **105:57** (**+4:02**), Student **166:32** (**+0:20**),
Instructor **178:33** (**−1:23**).
**Total and PIC each moved +0:01** — entirely the IMG_6036 row-13 decision above. Instrument, Student
and Instructor are **unmoved since p.54**. **No drift refresh is possible at p.62** — the last page
carries no totals.

## Confirmed corrections
- **IMG_4953 (closeout, pages 110/111) — row-1 landing & instructor rows (user-verified 2026-07-31).**
  Row 1 (27/05/2021 OH-CTL Räyskälä→Vääksy) landings read ambiguously as 10; user confirmed **9**
  (the per-page landing Δ then reconciles to paper's 43; Cumulative_Landings stays = paper's 1909).
  Instructing rows on this page are **rows 4, 6, 7, 8** (all OH-CTL seaplane) — each logged
  PIC+Instructor, so Instructor_Time = Total (0:41, 1:23, 1:15, 1:21). Page instructor Δ = 4:40
  (42:59 → 47:39). Row 2 of IMG_4952 (16/05/2021 OH-CAY) is also PIC+Instructor 0:48.
  The 25/05/2021 C185 OH-CDK float hops are **PIC, not student** (paper Oppilas column unchanged).
- **`logbook_tools.py` block-vs-total check softened (2026-07-31).** Hand-logged block times routinely
  differ from the logged Lentoaika by 1–3 min; the tool used to treat any such gap as a hard PROBLEM
  that blocked `--append`. Now a gap ≤5 min (`BLOCK_TOL`) prints as a non-blocking WARNING (logged time
  authoritative); a larger gap still blocks as a likely transcription error. Page-Δ cross-checks
  (total/PIC/student/landings) remain hard-blocking. Four ≤3-min warnings on the closeout batch
  (OH-CDK 12:38–13:35=0:57, OH-PDP 15:38–16:57=1:19, OH-CTL 12:14–12:36=0:22, OH-CTL 18:15–19:36=1:21).
- **IMG_4921 / 04/10/2018 / P28A OH-PDP / EFHF → EFHF / 1:20 / 4 landings — instructor time.**
  This is an *instructing* flight: the paper logs 1:20 under **both** Päällikkö (PIC) and Opettaja
  (instructor). Originally appended with PIC only; corrected on 2026-07-31 to also set
  `Instructor_Time = 1:20`, bumping `Cumulative_Instructor` 7:57 → **9:17** on that row and all rows
  after. PIC/Total/Student/SEP/Landings unchanged. Convention confirmed by the next page's carry
  (instructor 9:17). This is the pilot's first *instructing* flight in Book 2.

- **IMG_4905 / 28/06/2017 / P28A OH-PDP / EFHF → EFHF:** correct total is **00:37**, not 00:36.
  Cumulative Total and PIC corrected +00:01 from that row onward.
- **IMG_4908 / 02/09/2017 / C172 OH-CTL / VEHMERSALMI → HIRVENSALMI:** correct total is **01:04**,
  not 01:08. Paper book is +00:04 too high from that point if left uncorrected.
- **IMG_4908 / 08/10/2017 flights:** both are **student time** with PIC "Stude".

## Flagged paper anomalies (not corrections — our value kept)
- **IMG_4929 / 08/06/2019 / C172 OH-CTL / Sipoo → Tuusulanjärvi:** off-block 13:43 / on-block 14:12
  = 29 min, but the paper's logged flight time is **0:27** (and the right-page cumulative advances by
  0:27: 590:48 → 591:15). Kept Total 0:27; recorded Off/On as written. Likely on-block is 14:10, not
  14:12. Cumulatives unaffected.
- **IMG_4929 (pages 62/63) — paper PIC over-count.** Every row this page is PIC, so true Δ_PIC =
  Δ_total = **5:44**. The paper's PIC column advances **5:52** (477:21 → 483:13), an 8-min slip on the
  paper side. Our Cumulative_PIC (→ 487:12) is the correct figure. Δtotal/Δinstr/Δland all reconciled
  exactly, so this is isolated to the paper's PIC arithmetic. Consistent with the standing PIC-drift note.

- **IMG_4931 (pages 66/67) — paper running-total under-count of 30 min.** The pilot's own margin
  notes read **"+30" on row 4** (28/06/2019 OH-PIF EFHA→EFLA, 1:00) and "−2" on row 3. The paper's
  Kokonaisaika column advances only 6:33 across the page while PIC+student (and the true row totals
  from off/on block) sum to **7:03** — the total column was under-added by 30 min from row 4 on. Our
  cumulatives use the real row totals (row 4 = 1:00, off/on 11:23→12:23), so **our Cumulative_Total
  now runs ~26 min AHEAD of the paper's printed total** (the +30 slip minus the pre-existing paper
  +4). Confirmed: Δpic 4:41, Δstudent 2:22, Δinstrument 2:22, Δland 15 all reconcile exactly.
- **Place-name reads confirmed by user (2026-07-31):** the lake logged as `EFKU/Jälä` is **Iso-Jälä**
  (18/06/2019 OH-CDK departure); `Leikonvesi` (not "Leikanvesi") is the correct spelling on all
  IMG_4930/4931 rows; `EFHA` (IMG_4931 rows 3/4 OH-PIF) is correct.
- **IMG_4931 / 04/07/2019 / C172 OH-GKT / Kahvisaari → Kuohijärvi:** off 12:03 / on 12:37 = 34 min,
  logged flight time **0:36**. Kept 0:36 (the logged/cumulative value); recorded off/on as written.
- **IMG_4930 / 19/06/2019 / C185 OH-CDK / Salonsaari local (student, 12 landings):** off 15:25Z /
  on 16:24Z = 59 min, logged flight time **0:54**. Kept 0:54. **Zulu** row (Z suffix). Instructor =
  **Jansson** (floatplane training, not the IR course). Rows 4/5/6 of IMG_4930 all tagged Zulu.

- **IMG_4933 (pages 70/71) — paper PIC column under-count of 53 min, pilot self-corrected.** The
  paper's written PIC bottom total is 499:48, but the true page Δpic (rows 1+2+8 = 0:33+0:53+0:31 =
  1:57 on the carried 498:44) is **500:41** — written by hand under the table as "500 41". Our
  Cumulative_PIC uses the correct 500:41-basis. All other page Δ (total 9:29, student 7:32,
  instrument 7:32, land 11) reconcile. Combined with the IMG_4931 +30 total slip, the paper's
  printed PIC column now runs well behind ours; cross-check on per-page Δ only.
- **IMG_4933 / 01/08/2019 / P28A OH-PDP / EFHF → EFLA:** the off/on block was struck/rewritten in the
  book ("1035/1024" crossed out). User confirmed (2026-07-31) the correct block is **10:10–11:03**
  (= 0:53, matches the logged total). Recorded.

## IMG_4934/4935/4936 batch (pages 72–77, 10/08–27/09/2019) — appended 2026-07-31
- **New instructing flights (PIC + Instructor).** Paper logged time in the Opettaja column for:
  **13/08/2019 C185 OH-CDK Salonsaari local** (1:16, 10 landings — first floatplane *instructing*)
  and **28/08/2019 C172 OH-CTL Vuosaari→Tuusulanjärvi** (0:28, 3 landings). Both set
  `Instructor_Time = PIC_Time = Total`. Page Δinstructor 1:44 reconciled exactly.
- **27.08 → 27.09 date fix (IMG_4936 row 6).** Paper wrote "27.08" for the OH-PIF EFLA→EFJY flight,
  but Aviatron shows it as **27.09.2019** (09:28–11:13 UTC, matches paper 0928–1113) and chronology
  (between 22.09 and the other 27.09 rows) agrees. Recorded as **27/09/2019**.
- **IMG_4936 pilot correction ("IR-aika / oppilasaika / kokonaisaika tarkistettu ja korjattu
  27.9.2019").** The page's bottom **Total** = 650:33, but the sum of the 8 row totals on the carried
  640:17 is **650:02** — i.e. the pilot added a ~31-min lump catch-up to the Total column (fixing the
  earlier IMG_4931 −30 undercount). Instrument (Mittari 40:08) and Student (136:36) each read +1 min
  over our row sums (40:07 / 136:35) — small hand corrections. **Our row-based figures are
  authoritative** (Total 650:28, computed from off/on block). Effect: paper's printed Total now runs
  **~5 min AHEAD** of ours (was ~26 behind); PIC still lags. Per-page Δpic 1:31 and Δland 11
  reconciled exactly; Δtotal/Δstudent/Δinstr were intentionally not cross-checked on this page
  because the pilot's lump correction makes the column deltas non-comparable.
- **Timezone / `Z` suffix.** Every OH-PIF row and the 10/08 OH-GKT row match Aviatron's UTC block
  times **to the minute**, proving the paper wrote them in UTC (no `Z` marked in the book). Per user
  decision (2026-07-31) these were recorded **with `Z`**; OH-PDP/OH-CTL/OH-CDK left plain/local (the
  17/09 OH-PDP EFHF→EFLA row is explicitly marked "LT" in the book).
- **Z backfill of earlier plain-but-UTC OH-PIF rows (done 2026-07-31).** Added `Z` to 12 rows that
  had been appended without it: 18/04/2019 (×3), 14/06, 28/06 (×2), 23/07, 01/08 (×2), 02/08 (×3).
  Eleven match Aviatron's UTC block times to the minute. **The 18/04/2019 05:44–06:35 EFLA→EFLA row
  is the one exception — inferred UTC, not Aviatron-confirmed** (Aviatron omits it, likely a solo/
  warm-up; but its two same-day sibling lessons are UTC and a 05:44 *local* start = 02:44 UTC is
  implausibly pre-dawn). If the pilot recalls that flight was logged local, drop its `Z`. Only the
  Off_Block/On_Block strings changed; all cumulatives are byte-identical (Z is stripped by
  `logbook_tools.py` when computing). Remaining plain OH-PIF rows: none. Plain (local) rows are now
  only the club/seaplane types (OH-PDP/OH-CTL/OH-CDK/OH-STL where applicable).
- **Two near-identical OH-PDP ferries (not a duplicate).** 17/09 EFHF→EFLA (IMG_4935) and 19/09
  EFLA→EFHF (IMG_4936) are both logged as 0855/0859–0945/0946, 0:51, 1 landing. The paper counts
  each on its own page (both included). Times likely an estimate/copy; kept as written.
- **Instrument seed drift.** Our Cumulative_Instrument (44:10) runs **+4:02** ahead of the paper's
  Mittari column (40:08) — accumulated seed drift present since the IMG_4933 checkpoint (paper siirto
  23:05 vs our 27:08). All per-page instrument Δ match the paper's Mittari Δ (2:55 / 5:53 / 8:14).

## IMG_4937/4938/4939 batch (pages 78–83, 30/09/2019–17/04/2020) — appended 2026-07-31
- **CB-IR skill test (03/10/2019, OH-PIF, 2 flights).** Aviatron confirms syllabus "KOU TAR"
  (tarkastuslento/check-ride), examiner **Timo Aineslahti** (FI.FCL 20163) — *not* Autere. Logged
  under the paper's Oppilas column ⇒ **Student time** (+ instrument 0:44 / 1:39), `pic_name =
  Aineslahti`. Both block times match Aviatron **UTC to the minute** (EFHA-EFHA 10:23→11:07, EFHA-
  EFJY off-block paper 11:12 / Aviatron 11:11 →12:51) ⇒ written **with `Z`**. This closes the CB-IR
  syllabus; OH-PIF does not recur after this.
- **IMG_4937 landing column under-add of 1 (paper +1 drift CLOSED).** The page's landing column
  advances **+12** (Siirto 1530 → bottom 1542) but the 8 row entries sum to **+13** (1,2,2,1,1,4,1,1).
  Per policy we sum the row entries (13), which puts our Cumulative_Landings at **1542 — exactly the
  paper's printed bottom**. Because our count had been running 1 *behind* the paper's printed running
  total (the standing paper +1), this page consumes that lead: **from 17/04/2020 our landings equal
  the paper's printed count** (batch ends 1591 = paper 1591). The row5/row6 landing split (23/10
  OH-PDP instructing = 1; 30/10 OH-STL KOU student = 4) is inferred but endpoint-neutral.
- **Instructing flights (PIC + Instructor).** 23/10/2019 OH-PDP EFHF→EFHF (1:26, 1 landing) and
  19/12/2019 OH-STL EFTU→EFTU (0:37, 4 landings). Both set Instructor_Time = PIC_Time = Total.
  Page Δinstructor 1:26 / 0:37 reconciled exactly.
- **30/10/2019 DA40 OH-STL student flight ("KOU JS").** Logged under Oppilas (0:54, 4 landings) ⇒
  Student time. `pic_name = Stude` (user-confirmed 2026-07-31; the "JS" in the remark left as-is).
- **First Night_Time in the file.** 26/03/2020 DA40 OH-STL EFHF→EFHF evening flight: Night 0:50
  (also instrument 1:05). Schema has no Cumulative_Night, so it is a per-row value only.
- **06/03/2020 C152 OH-NEU** has a struck duplicate line directly above it in the book (on-block
  1414 crossed out, rewritten 1416). Kept **one** row; off/on recorded as written (12:41–14:16 = 1:35
  block vs logged Lentoaika 1:33 — a 2-min taxi-rounding gap, same class as prior accepted rows; the
  only `logbook_tools.py` block/total flag this batch, appended deliberately).
- **New this batch:** **EETN** (Tallinn — first Estonian/international field, 31/12/2019 DA40 EFHF↔
  EETN day-trip), **OH-NEU** (C152 landplane), **Särkijärvi** (seaplane lake, 30/09 C185 OH-CDK).

## IMG_4940/4941/4942 batch (pages 84–89, 19/04/2020–30/05/2020) — appended 2026-07-31
- **All page Δ reconciled exactly** (total/PIC/student/instructor/landings) on all three pages — a
  notably clean batch. Only anomaly: one 2-min taxi gap (below), appended deliberately.
- **OH-PIF returns twice, post-CB-IR, both UTC (`Z`, Aviatron-confirmed to the minute):**
  - **22/04/2020 EFLA-EFLA (ID 14505):** off/on 12:56–14:51 UTC, block 1:55, syllabus **KOU HAR**,
    OPPILAS = Ayoub, instructor **Autere** (Tuomas), remark *"Kertauslento"* (= paper's "KERT" +
    signature). Recorded as **student** time, instrument 1:55, `pic_name = Autere`, `Z`. Landings 1.
  - **30/05/2020 EFLA-EFLA (ID 15017):** off/on 10:53–12:46 UTC, block 1:53, syllabus **VUO MAT**.
    Recorded as **PIC**, instrument 1:30, `Z`, **1 landing** per paper. ⚠️ Aviatron logs this under
    **OPPILAS / 2 landings**; **paper wins** (PIC / 1 landing) per policy. Page Δ all reconcile with
    the paper's PIC-and-1-landing view, so the paper is internally consistent.
- **New regs OH-CAY & OH-CGX (both C172 landplanes, EFHV, 19/04/2020).** OH-CAY flown PIC (0:29).
  **OH-CGX flown student** (0:19) — not in Aviatron; instructor **Härkönen** (user-confirmed
  2026-07-31), `pic_name = Härkönen`. Page ΔStudent 2:14 = 0:19 (OH-CGX) + 1:55 (OH-PIF) confirms the
  split.
- **Instructing flights (PIC + Instructor) this batch:** 24/04 DA40 OH-STL (1:02, 3 land); 09/05
  C172 OH-CTL Anttola→Tuusulanjärvi (1:35, 4); 14/05 DA40 OH-STL (1:17, 4); 15/05 C172 OH-CTL
  Tuusulanjärvi local (1:17, 4); 19/05 C172 OH-CTL Kabböle local (0:40, 7); 25/05 C172 OH-CTL
  Tuusulanjärvi local (1:12, 6). All set Instructor_Time = PIC_Time = Total; page Δinstr 1:02 / 4:49 /
  1:12 reconciled exactly.
- **2-min taxi gap (appended deliberately, same class as 06/03/2020 OH-NEU).** 22/04/2020 P28A OH-PDP
  EFHF→EFLA: off/on 13:47–14:39 = 0:52 block, but logged Lentoaika **0:50**. Kept logged 0:50; off/on
  recorded as written. The only `logbook_tools.py` block/total flag this batch.
- **Date ordering as written:** on IMG_4942 the pilot logged 20/05 (P28A OH-PDP) *after* the two 22/05
  C185 rows. Kept the book's row order; dates are correct.
- **Paper-vs-ours drift (unchanged, seed-only):** at 30/05/2020 paper printed bottoms are Total 698:44,
  PIC 557:07, Landings 1661. Ours: Total 699:09 (+25), PIC 556:39 (−28), Landings 1661 (**exact** —
  drift stays closed). Cross-check on per-page Δ only.

## IMG_4943/4944/4945 batch (pages 90–95, 07/06–18/07/2020) — appended 2026-07-31
- **Notably clean batch.** All page Δ reconciled exactly (total/PIC/student/instructor/landings) and
  **zero off/on-block vs total flags** — every row's block time equals its logged total.
- **One student flight.** 15/07/2020 C172 OH-CTL Salonsaari local (1:00, 7 landings) logged under
  Oppilas ⇒ **Student time**, `pic_name = Sinervä` (user-confirmed 2026-07-31; Sinervä was also the
  30/04/2019 OH-CTL seaplane-student instructor). Page ΔStudent 1:00 reconciled exactly.
- **Five instructing flights (PIC + Instructor = Total).** 10/06 OH-CTL Tuusulanjärvi (1:01), 15/06
  OH-CTL Tuusulanjärvi (1:30), 06/07 OH-CTL Tuusulanjärvi (1:24 — see slip below), 15/07 OH-CTL
  Salonsaari (1:06), 18/07 OH-CTL Savonlinna→Ruokkee (1:03). Page Δinstr 2:31 / 2:09 reconciled exactly.
- **06/07/2020 OH-CTL — paper PIC/Opettaja column slip of 4 min.** The row total/block is **1:24**
  (18:35–19:59, cumulative advances 1:24) but the paper wrote **1:20** in both the Päällikkö and
  Opettaja columns. Per our instructing rule we recorded Total = PIC = Instructor = **1:24** (block-
  based). Effect: our Cumulative_PIC gains +4 vs paper's column here — closing the standing PIC gap
  from −28 to **−24** vs paper. Paper's own Δtotal (6:36) still includes the full 1:24, so Δtotal/Δland
  reconcile; Δpic/Δinstr were **not** cross-checked on this page (paper's 1:20 columns are non-comparable).
- **Timezone.** All rows plain **local**. The 15/07 OH-CTL rows 1 & 2 carried a faint `z`-like
  subscript, but user chose **local** (2026-07-31) — OH-CTL isn't in Aviatron so no UTC confirmation,
  and the mark is likely stray. Row-1 off-block (15/07) was struck/rewritten; read as **08:41** (gives
  0:47 with on-block 09:28, matches the logged total).
- **Place-name reads (user-confirmed 2026-07-31):** 12/06 OH-CTL **Laajasalo→Nurmoo** and
  **Vesikkas→Virtosalmi** (lakes); 18/07 OH-CTL **Savonlinna→Ruokkee**. New airport codes this batch:
  **EFNU** (Nummela, 16/07 OH-PDP) and **EFSA** (Savonlinna, 18/07 OH-PDP EFHF→EFSA).
- **Paper-vs-ours drift at 18/07/2020:** paper printed bottoms Total 720:43, PIC 578:02, Landings 1726.
  Ours: Total 721:08 (+25), PIC 577:38 (−24), Landings 1726 (**exact** — drift stays closed).

## IMG_4946/4947/4948 batch (pages 96–101, 19/07/2020–08/10/2020) — appended 2026-07-31
- **All page Δ reconciled exactly** (total/PIC/student/instructor/landings) on all three pages.
- **New reg OH-TIL (P28A, Aviatron aircraft, IFR Arrow).** First appears 28/07/2020. Not in Aviatron's
  data until 2021, so no UTC cross-ref for the 2020 rows → all OH-TIL rows kept **plain local**.
- **28/07/2020 OH-TIL EFHF→EFHF (2:02, 1 land) = STUDENT + instrument 2:02.** Logged under Oppilas
  (paper Δstudent 2:02 reconciles exactly). Dual instrument checkout on the new OH-TIL; instructor
  **Lord** (user-confirmed 2026-07-31; user first said Tarhanen, then corrected to **Lord**). The two
  following OH-TIL rows (05/08 EFKU→EFHF 2:17, 06/08 EFHF→EFHF 2:51) and 21/08 EFHF→EFHF 2:46 are **PIC**
  (05/08 & 06/08 carry instrument time, 21/08 does not). Page Mittari Δ 7:10 = 2:02+2:17+2:51.
- **06/08/2020 OH-TIL — 2-min block/total gap (appended deliberately).** Off/on 14:07–17:00 = 2:53
  block, but logged Lentoaika (and cumulative advance) **2:51**. Kept logged 2:51; off/on recorded as
  written. Same class as 06/03/2020 OH-NEU & 22/04/2020 OH-PDP. Only block flag this batch — off/on were
  omitted from the tool run to bypass its block check, then hand-filled into the CSV.
- **23/09/2020 OH-PIF returns for an IR/SEP proficiency check — three rows, all STUDENT + instrument,
  all UTC (`Z`, Aviatron-confirmed to the minute).** Page Δstudent 3:09 = Mittari Δ 3:09 reconcile exactly.
  - **EFLA→EFJY (Aviatron ID 16810):** 07:30–08:25 UTC, 0:55, 1 land, syllabus **KOU MAT** (ferry to EFJY
    for the check rides), instructor **Kääriäinen** (Antti). `pic_name = Kaariainen`.
  - **EFJY→EFJY (ID 16813):** 11:10–12:25 UTC, 1:15, 3 land, syllabus **KOU TAR** (tarkastuslento = the
    check ride), examiner **Aineslahti** (Timo — same examiner as the 03/10/2019 CB-IR skill test).
    `pic_name = Aineslahti`. Book's Huomautuksia carries a **"SEP(land) (P) Revalid"** examiner endorsement.
  - **EFJY→EFLA (ID 16819):** 13:18–14:17 UTC, 0:59, syllabus **KOU MAT** (return leg), instructor
    **Kääriäinen**. ⚠️ Aviatron logs 16819 as **EFLA→EFJY / 1 landing**; **paper wins** — kept
    **EFJY→EFLA** (geographically the return home) and **3 landings** (page Δland 20 reconciles with 3).
- **Chronology quirk across the page boundary (kept as written).** Page 98/99 ends with the 23/09 OH-PIF
  check-ride rows; page 100/101 then runs 11/09–08/10 club/seaplane flights — i.e. the 11/09/15/09/22/09
  rows are logged *after* 23/09. The pilot recorded the special check-ride trip out of sequence. Row/page
  order (hence append order) kept as in the book; cumulatives accumulate in that order, unaffected.
- **20/08/2020 OH-PDP EFHF→EFHF — paper PIC-column slip of +3.** Row total/block/Kokonaisaika = **0:30**
  (20:15–20:45), but the paper wrote **0:33** in the Päällikkö column. Kept PIC = 0:30. Effect: paper's
  PIC column now runs +3 further ahead of ours (our vs-paper PIC gap widened −24 → **−27**). The page's
  d_pic check was therefore omitted (non-comparable); d_total 9:24 + d_student 3:09 + d_land 20 pin the
  transcription regardless.
- **Instructing flights (PIC + Instructor = Total) this batch:** 22/09 OH-PDP EFHF→EFHF (1:40, 5 land),
  24/09 OH-CTL Tuusulanjärvi→EFRY (0:37, 2 land — floatplane instructing), 01/10 OH-PDP EFHF→EFHF (1:22,
  4 land). Page Δinstructor 3:39 reconciled exactly.
- **New this batch:** **EFFO** (Forssa — 28/08 OH-PDP EFHF↔EFFO day-return), **OH-TIL** reg, place
  **Papinniemi** (14/08 OH-CDK arrival, Saimaa lake — user-confirmed 2026-07-31).
- **Paper-vs-ours drift at 08/10/2020:** paper printed bottoms Total 749:05, PIC 601:16, Mittari 59:30,
  Instructor 42:07, Landings 1781. Ours: Total 749:30 (+25), PIC 600:49 (−27), Instrument 63:32 (+4:02),
  Instructor 42:11 (+4), Landings 1781 (**exact** — drift stays closed).

## IMG_4949/4950/4951 batch (pages 102–107, 13/10/2020–07/05/2021) — appended 2026-07-31
- **Crosses into 2021.** Book has no year on page 102; 13/10–28/12 are 2020, 20/01 onward 2021
  (page 104 prints "2021"). 24 flights, all page Δ (total/PIC/student/landings) reconciled exactly.
- **UTC rows (user-confirmed):** the two **18/04/2021 C172 OH-COK** ferry legs are UTC → `Z`
  (07:25Z→08:38Z EFIM→EFPR, 10:54Z→12:04Z EFPR→EFIM). The book logged them **out of order with two
  arrows** marking the swap; user confirmed the correct order is by time. Reordered on append. The two
  C185 OH-CDK float rows between them stay **local** (clock overlap with the OH-COK legs is the expected
  local-vs-UTC artifact, not an error). OH-COK is a landplane C172, not in Aviatron — UTC per pilot's log.
- **Student flights (2):** 18/04/2021 **C185 OH-CDK** EFIM locals, 0:59 + 0:28, instructor **Matikainen**
  (signature + "FI.RCL2234" in remarks — a SEP-sea class-rating check). Student_Time; feeds Cumulative_
  Student (+1:27) and Cumulative_SEP_Sea (+1:27, OH-CDK seaplane). pic_name = Matikainen.
- **11/02/2021 DA40 OH-STL** night flight (EFHF local, 17:27–19:40): **PIC self**, Total 2:13,
  **Instrument 1:36 + Night 1:55**. User-confirmed instrument < PIC because part of the flight was VFR
  (and part was not night). Only instrument row this batch (+1:36 → Cumulative_Instrument 65:08).
- **28/12/2020 DA40 OH-STL** (EFHF local): off-block cell had a struck-through correction; used **off
  10:30 / on 11:31** (Total 1:01, confirmed by the cumulative Δ). Type handwritten "D140" = the DA40.
- **New this batch:** reg **OH-COK** (C172, user flew it briefly 2021/2022), **DA40 OH-STL** recurring,
  airports **EFPR** and **EFIM** (Immola — Saimaa-region float base).
- **Paper-vs-ours drift at 07/05/2021:** paper printed bottoms Total 772:12, PIC 622:56, Mittari 61:06,
  Oppilas 149:45, Instructor 42:07, Landings 1843. Ours: Total 772:37 (+25), PIC 622:29 (−27),
  Instrument 65:08 (+4:02), Student 150:08 (+0:23), Instructor 42:11 (+4), Landings 1843 (**exact** —
  drift stays closed).

## ⚠⚠ STANDING: the paper book's `Z` / UTC marking is unreliable (laskukierros cross-check, 2026-08-01)
The user gave access to **laskukierros.fi** (a club booking/billing system); its export is now
`laskukierros_export.csv` in the repo — 128 flights 2020–2026, **club aircraft only** (see
`reference.md` for coverage and how to refresh it).

**Its times are LOCAL.** Proof — three winter rows it logs with **day** landings that would land well
after dark if the numbers were UTC (Helsinki sunsets in brackets): 16/01/2024 OH-AWB 14:30–15:19,
1 day ldg [sunset ~15:55]; 04/01/2025 OH-CAM 14:08–15:07, 1 day ldg [~15:32]; 01/01/2026 OH-CMU
13:54–14:26, 4 day ldg [~15:24]. As UTC each ends 1½–2 h after sunset. It also correctly marks
12/12/2025 OH-CAM 18:42–19:22 as **3 night** landings. So laskukierros = local, consistently.

**Consequence — the paper's `Z` is wrong roughly half the time, in BOTH directions.** Of **82** rows
matching ours on date + registration + time:
- **52 are genuinely local** — 11 stored correctly plain, **41 wrongly carry a `Z`**
- **30 are genuinely UTC** — 27 stored correctly, **3 are missing their `Z`**
  (`21/07/2024` OH-CTL 12:39, `25/06/2024` OH-CTL 12:18, `08/04/2021` OH-COK 16:00)

This is **not** a transcription error — we recorded the `Z` the book shows. It confirms mechanically
what the user has been saying row by row all along ("these are local"), and it means **`Z` flags
in the CSVs cannot be trusted wherever we took the book at its word.**

**DECISION (user, 2026-08-01): document only — do NOT change the CSVs.** The paper logbook stays
authoritative. **Full 44-row list: [`laskukierros_zflags.md`](laskukierros_zflags.md).** The future
normalizing app must treat the `Z` flag as advisory and prefer laskukierros/Aviatron where they cover
the row. **No cumulative total is affected** — `Z` suffixes don't feed any sum.

### Three record-level conflicts (paper kept, per user — documented only)
All three match ours exactly on times *and* landings, so they are the same flights:
| ours | laskukierros | conflict |
|---|---|---|
| `24/09/2022` OH-CTL ×2, 10:05–10:44 / 11:33–12:13 | **22/09/2022** | date (2 days) |
| `12/07/2023` OH-CTL 18:00–19:08, 5 ldg | **18/07/2023** | date |
| `26/12/2021` **OH-CMV** 09:34Z–10:36Z | **OH-CMU** 11:34–12:36 | registration |
- The 12/07/2023 row is the one the user called *"an older flight I missed, entered out of order"*
  (IMG_6018 row 5) — 18/07 would make it a **later** flight, not an earlier one.
- The 26/12/2021 row is the `OH-CMU`/`OH-CMV` pair again. Note its times reconcile **exactly** as
  UTC+2, so that row's `Z` is genuinely correct — only the registration is in question.

## ⚠⚠ MISSING-FLIGHTS BATCH — 15 rows entered into the CSV **BEFORE** the paper book (2026-08-01)
**This is the first batch in the whole project that runs the other way round: the user dictated
15 flights that had never been written into the paper logbook, we appended them to `logbook_3.csv`
first, and he then transcribes them onto paper from our table.** There is no page image, no
"TOTAL THIS PAGE" and no page cross-check — the `paper` block in `batch_missing_2026.json` is empty.
Batch file: **`batch_missing_2026.json`** (tracked). Dates **12/06/2026 → 30/07/2026**.

**The user's stated plan:** row 1 (12/06 OH-ESR EFNU→EFIK) fills the **last free row of page 62**;
the remaining 14 open a fresh spread (**pages 63–64**); he then **rewrites the book's carried totals
to match ours, closing the standing paper-vs-ours drift for good.** Expect the next spread photo to
show corrected "TOTAL PREVIOUS PAGES" figures — **do not treat that as a book error.**

### Sources & how each row was settled
- **Aviatron confirmed 5 OH-GKT rows to the minute** (13/06, 26/06, both 11/07, 12/07) and **fixed
  two of them.** The user's dictated on-block/landing pairs had **slipped one row down** — the 26/06
  row carried the 13/06 row's `11:56/12:02`, and the 11/07 Mäntyharju row carried the 26/06 row's
  `13:26/13:34`. Stored per Aviatron: 26/06 → `13:26/13:34` (block 1:03, air 0:49); 11/07
  Kahvisaari→Mäntyharju → `10:17/10:27` (block 1:04, air 0:45). *Same copy-down failure mode as
  IMG_6032 r11 — when a dictated or written pair duplicates the previous same-reg row, check it.*
- **laskukierros confirmed OH-CAM (29/06) and OH-TIL (25/07) digit-for-digit.**
- **⚠ Date corrected: the OH-CAM EFHV local is `29/06/2026`, not the 26/06 the user dictated**
  (user chose the club record). Note 26/06 was *also* internally coherent — GKT at Kahvisaari
  15:31–16:34 local, drive to EFLA, PDP EFLA→EFHV 17:35, CAM local 19:02 — so the chain argument
  did **not** settle it; the club record did.
- **⚠ ZONE CONFLICT, user ruled UTC: 25/07/2026 OH-TIL EFTP local.** Stored `06:56Z–08:04Z` per the
  user. But `laskukierros_flights.csv` — whose times are **local** — holds exactly `06:56–08:04`.
  Either the club row was entered in UTC by mistake (precedent: 06/08/2024 OH-CTL) or the flight was
  06:56 local. Aviatron has **no OH-TIL row after 2021**, so there is no independent check.
  Same shape as the IMG_6036 ruling: present once, store what the user says, log it here.
- **Two 23/06/2026 OH-CTL instructing rows the user had not listed** (pupil **Pekka Puhakka**) were
  surfaced from the club file and added at his instruction: Tuusulanjärvi local 1:13 / 9 ldg and
  Karhusaari→Pellinki 0:28 / 1 ldg, both **local**, both PIC + Instructor.
- **30/06/2026 OH-CTL Tuusulanjärvi local** (sent mid-turn) has **no record in either reference** —
  it stands on the user's figures alone: block `11:47–12:57` = **1:10**, airborne 0:58, 7 ldg,
  PIC + Instructor, local.
- **12/06/2026 OH-ESR ×2 and the two OH-PDP rows have no electronic record** (neither fleet covers
  SR20 or PDP). As dictated.

### ⚠ Airborne times are stored for the first time at scale — `Takeoff`/`Landing` now populated
The user dictated **both** pairs (`off-block/takeoff – landing/on-block`), so on **all 15 rows**:
`Off_Block`/`On_Block` + `Block_Time` = the block pair, **`Takeoff`/`Landing` = the airborne pair**,
and **`Total_Time` = the block time** (what the book totals on — unchanged convention). Before this
batch only one row in any book used those columns (08/09/2025, IMG_6034 r6). `logbook_tools.py` was
extended this session with optional **`takeoff`**, **`landing`** and **`block`** batch fields to
support it. **`Block_Time == Total_Time` on all 15**, so no cumulative behaves differently.

### Two rows where the block pair was derived, not dictated
The user gave **airborne** times only and said to pad ~5 min per side:
| row | dictated (airborne) | stored block | total |
|---|---|---|---|
| 12/06/2026 OH-ESR EFNU→EFIK | `15:45–16:20Z` | **15:40Z–16:25Z** | 0:45 |
| 26/06/2026 OH-PDP EFLA→EFHV | `17:35–18:05` local | **17:30–18:10** | 0:40 |
His wording differed slightly ("+/- 5 minutes per side" vs "+/- 5 minutes total on either side");
both were read as **5 min each side**. Flag if a later record contradicts either.

### Roles & totals contributed
Instructing (PIC + Instructor): 12/06 OH-ESR EFNU→EFIK 0:45, both 23/06 OH-CTL (1:13 + 0:28),
30/06 OH-CTL 1:10, 30/07 OH-GKT 1:00 → **Instructor +4:36**. The 12/06 *return* leg (EFIK→EFNU) is
**PIC only** per the user — an instructing outbound with a non-instructing return; as dictated.
Instrument: 25/07 OH-TIL **1:08** (club `ifr_min 68` = the whole flight). No dual, no night.
SEP_Sea **+9:22** (OH-GKT 6:31 + OH-CTL 2:51). Landings **+48**.

### Still flown-but-unrecorded after this batch
Nothing known. Both references are now fully consumed: `laskukierros_flights.csv` ends **25/07/2026**
and Aviatron ends **12/07/2026**, and every row in each that postdates the paper book is in the CSV.
⚠ **6037 r3 (16/05/2026 Kabböle local 0:50) is still missing from the club file** — the pupil never
logged it; our row is correct as is.

## 🏁 PAPER DRIFT CLOSED AT PAGE 62 (2026-08-01) — the book now carries OUR figures
The user wrote the corrected totals at the bottom of **page 62** and confirmed them back:
**Total 1206:58 · SE-VFR 1101:01 · SE-IFR 105:57 · PIC 1040:26 · FI 185:50 · Dual 166:32 ·
Day landings 3335 · Night landings 59.** (Night time **22:45** was supplied but not read back.)
**Every paper-vs-ours drift is therefore zero from p.62 onward.** The historical offsets
(Total +2:06, PIC +1:46, SE-IFR +4:02, Dual +0:20, FI −1:23 at p.60) are now closed by hand.
**Do not re-apply them to any later page, and do not read a p.62-or-later book total as evidence of
a new slip until it has been cross-checked against `logbook_3.csv` directly.**

Accounting for the Total +2:06 that was corrected (fully decomposed):
**+0:25** inherited at the Book-2→Book-3 handover · **+1:00** p.14 (10/08/2022 OH-PIF Tarhanen row
1:55 entered as 0:55 in the running column) · **+1:00** p.52 (6032 r13, a 1:38 flight added as 0:38) ·
**−0:20** p.52 (6032 r11, 08/07/2025 OH-GKT — book 0:57, Aviatron 0:37, user ruled Aviatron) ·
**+0:01** p.60 (6036 r13, 0:24 added as 0:23).
**PIC +1:46 does NOT decompose** — its history runs −0:27 at handover → −1:25 by p.20 → **+1:05 from
p.26**, where the pilot struck his own PIC total 810:09 and carried 807:39, an undocumented −2:30 hand
correction he does not recall the basis for → +1:45 at p.54 → +1:46 at p.60. Our per-page Δpic
reconciled exactly on p.26 and p.28, which is why our figure was kept. **That 2:30 remains unexplained
and is now baked into the paper book as well.**

### ⚠ The day/night landing split written at p.62 is PART INFERRED — 9 of the 59 nights
`Landings` stores the **sum**; the night split is inferred from `Night_Time` (see `reference.md`).
Across all three books there are **22 rows with night time**. **17 are full-night**
(`Night_Time == Total_Time`) → all their landings are night, **50 landings, certain**. Book 3's seven
(24 landings) are corroborated row-by-row by the transcription notes above.
**The remaining 9 are estimates** from five partial-night rows — all evening departures from EFHF,
so the night portion is at the *end* and each row's **final** landing is certainly night; how many
earlier ones fall after night onset is not recoverable from the CSV:

| date | reg | off–on | total / night | ldg | night ldg | basis |
|---|---|---|---|---|---|---|
| 15/11/2012 | OH-KAS | 14:38–15:42 | 1:04 / 0:24 | 5 | **2** | night = last 24 of 64 min; ~13 min/circuit |
| 25/03/2013 | OH-CTH | 18:50–20:03 | 1:13 / 0:20 | 1 | **1** | single landing, at the end — certain |
| 11/11/2013 | OH-COF | 17:00–18:00 | 1:00 / 0:30 | 4 | **2** | night = exactly the second half |
| 26/03/2020 | OH-STL | 16:35–18:38 | 2:03 / 0:50 | 3 | **1** | night = last 40%; only the final landing is safe |
| 11/02/2021 | OH-STL | 17:27–19:40 | 2:13 / 1:55 | 3 | **3** | night starts 18 min after off-block |

**Range: night 55–66, day 3328–3339.** The user adopted the midpoint estimate (**59 / 3335**) and
inked it. **Two exact sources still exist and were offered but not used:** (a) the Day/Night landings
columns on paper for those five rows — Book 1 ×3, Book 2 ×2 — which our transcription never captured,
and (b) the book's own printed day-only cumulative at p.60 plus the 59 all-day landings of pp.61–62.
**If either is ever read, correct p.62 rather than assuming the CSV is wrong** — `Cumulative_Landings`
(the sum) is unaffected either way; only the split moves.
⚠ Assumes no row has night landings without night time logged. The reverse happened on 05/03/2024
(book put 3 night landings in the day column), so the columns do get crossed.

## Standing discrepancy — paper landing count (RESOLVED 17/04/2020)
The paper logbook's **cumulative landing count ran ahead of the true count** for most of Books 1–2
(at the historical IMG_4910 checkpoint paper showed 1051 vs correct 1050, paper +1). **This closed on
IMG_4937 (2026-07-31):** the paper under-added its own landing column by 1 there, so from 17/04/2020
our Cumulative_Landings **equals** the paper's printed count (1591). Still cross-check landing sums
on new pages — a fresh divergence can reopen.

## Historical checkpoint (superseded — see resume.md for current)
- After IMG_4910: last row 15/03/2018 C172 OH-CWB EFHF→EFHF total 00:26 PIC self;
  cumulatives Total 467:48, PIC 379:50, Student 87:58, Instrument 3:12, SEP_Sea 94:03,
  Landings 1051 paper / 1050 correct.
