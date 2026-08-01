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
