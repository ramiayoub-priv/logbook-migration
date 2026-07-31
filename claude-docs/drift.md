# Drift Notes — Corrections & Discrepancies

Log every correction that alters a row later rows built on. Fix cumulatives from that row forward.

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
- **Book-3 landing drift now open & growing (day+night vs book's day-only).** At 10/06/2022 our
  Cumulative_Landings = **2172** (day+night). The book's printed day-only cumulative is lower by the running
  night-landing total (3 on 17/12/2021 + 8 on IMG_6010 = 11 so far) plus the ~2-landing IMG_6009 struck-cell
  slip. Cross-check on per-page Δland (day+night sum) only; do not expect our cumulative to equal the book's.

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
