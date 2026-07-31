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

## Confirmed corrections
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
