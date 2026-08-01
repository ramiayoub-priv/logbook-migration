# Data model

The schema the app owns, how the three CSVs map onto it, and the domain rules behind the choices.
Read `CLAUDE.md` §0 rules 2, 4 and 5 first — they constrain everything here.

## Principles

1. **Durations are stored as INTEGER minutes.** Never as `"1:21"` strings or floats. A legal record
   needs exact arithmetic; `H:MM` is a display format, parsed at the edges and nowhere else.
2. **Instants are stored as RFC3339 UTC strings.** SQLite has no date type; RFC3339 sorts
   lexicographically, which is what we rely on.
3. **No stored cumulatives.** (Rule §0.5.)
4. **Provenance is never dropped.** Every flight remembers which book and row it came from, and every
   converted time remembers what the paper actually said.

## Where this lives

The schema is `app/backend/internal/store/schema.sql`, embedded in the binary. The CSV→domain mapping
and every check below is `app/backend/internal/csvbook` (100% test coverage, it is calculation core).
The operator runs it with `app/backend/cmd/logbookctl`:

```bash
logbookctl import -db /var/lib/logbook/logbook.db -csv /path/to/repo -dry-run   # report only
logbookctl import -db /var/lib/logbook/logbook.db -csv /path/to/repo            # backs up, then imports
logbookctl verify -db /var/lib/logbook/logbook.db -csv /path/to/repo            # re-check, writes nothing
```

It is a separate binary from the server on purpose: a destructive operation on a legal record must
never be reachable from an HTTP request.

## Tables

Besides the four below the schema carries `discrepancies` (the review list, rewritten on every
import), `import_runs` (an append-only audit trail: when, how many, which backup) and
`schema_version`.

### `aircraft`
The seed list that makes the new-flight form smart. **Derived from the flights on import**, so it can
never drift from what was actually flown: `type` is the most-flown type for that registration (ties
broken alphabetically, so two imports produce identical rows), `default_class` comes from the
seaplane registration list, `active` means flown within two years of the **last flight in the books**
(not of today, so the import stays reproducible), and `ifr_capable` is a curated set — `OH-CAM`,
`OH-ESR`, `OH-PIF` — because instrument time is also logged under the hood in aircraft that are not
IFR certified (`OH-COF` and `OH-CTH` are C152s with instrument rows). `ifr_capable` is a form hint
and never constrains what a flight may record.

| Column | Type | Notes |
|---|---|---|
| `id` | INTEGER PK | |
| `registration` | TEXT UNIQUE | `OH-CTL`. Finnish regs are `OH-xxx`; anything else is flagged, not rejected (`SE-GKT`, `SE-LWI` are real). |
| `type` | TEXT | `C172`, `P28A`, `M6` … |
| `default_class` | TEXT | `SEP_LAND` · `SEP_SEA` · `MEP_LAND` · `MEP_SEA` · `TMG` |
| `ifr_capable` | INTEGER | Drives whether the form offers an instrument-time field. `OH-CAM` and `OH-ESR` are IFR; most C172s are not. |
| `active` | INTEGER | Retired aircraft stay in the DB for history but drop out of the form's default list. |
| `notes` | TEXT | e.g. "ex-`SE-GKT`, re-registered", "Maule on floats, always". |

**`default_class` is a default, not a fact.** An aircraft can fly floats one season and wheels the
next, so the *flight* carries the authoritative class (below). This is exactly the behaviour the user
asked for: preselect sea, allow override when the configuration changed.

### `flights`

| Column | Type | Notes |
|---|---|---|
| `id` | INTEGER PK | |
| `seq` | INTEGER UNIQUE | **Explicit book order.** Many flights share a date, and the paper order is meaningful, so ordering must never rely on date alone. This is the key every cumulative computation walks. |
| `flight_date` | TEXT | `YYYY-MM-DD`, the date as written in the book. |
| `aircraft_id` | INTEGER FK | Nullable — a flight is never lost because its aircraft row is missing. |
| `aircraft_reg`, `aircraft_type` | TEXT | Denormalized **as written on paper**. The paper is authoritative (rule §0.2); if the aircraft table later disagrees, that is a discrepancy to surface, not to auto-correct. |
| `class` | TEXT | Per-flight, seeded from `aircraft.default_class`, user-overridable. **This is what the sea/land statistics split on.** |
| `dep_place`, `arr_place` | TEXT | Airport ICAO (`EFHV`) or lake name (`Tuusulanjärvi`). Free text by necessity. |
| `off_block_utc`, `on_block_utc` | TEXT | Canonical RFC3339 UTC. |
| `off_block_raw`, `on_block_raw` | TEXT | **Exactly as on paper**: `"15:30"` or `"07:56Z"`. Never rewritten. |
| `time_origin` | TEXT | `utc_as_written` · `converted_from_local` · `unknown` · `none` |
| `takeoff_utc`, `landing_utc` | TEXT | The airborne pair. Populated on 16 rows only (see `claude-docs/reference.md`). |
| `block_minutes` | INTEGER | Gate-to-gate. |
| `total_minutes` | INTEGER | **The flown time the book totals on.** Equals `block_minutes` on all but one row. Never adjust it to match a block time — every downstream total depends on it. |
| `night_minutes`, `instrument_minutes` | INTEGER | `Night_Time`, `Instrument_Time` (EASA SE-IFR). |
| `pic_minutes`, `dual_minutes`, `instructor_minutes` | INTEGER | `PIC_Time`, `Student_Time` (= dual/Oppilas), `Instructor_Time`. |
| `copilot_minutes`, `multipilot_minutes` | INTEGER | Zero everywhere so far; the EASA book has the columns and the PDF must render them. |
| `pic_name` | TEXT | `self` on PIC and instructing rows; an instructor's name on dual rows. |
| `landings_day`, `landings_night` | INTEGER | See the gap note below. |
| `landings_verified` | INTEGER | `0` = the day/night split was inferred, not read from paper. Drives the review list. |
| `remarks` | TEXT | |
| `source_book`, `source_row` | INTEGER | Provenance: which CSV and which line. Makes any figure traceable back to paper. **`source_book = 0` marks a flight typed into the app** — see below. |

### `users`, `sessions`
See `security.md`.

## Hand-entered flights: two disjoint bands (2026-08-01, Task 5)

`POST /flights` writes flights that were never on paper. They share the `flights` table with the
imported rows, and the whole design turns on keeping the two populations apart, because **the
importer replaces its own rows on every run** and the migration effort re-imports every time a page
is appended to `logbook_3.csv`.

| | Imported | Hand-entered |
|---|---|---|
| `source_book` | 1, 2 or 3 | **0** |
| `source_row` | line in that CSV | an app-local counter, 1, 2, 3… |
| `seq` | 1..N, **reassigned on every import** | **from 1 000 000 up**, allocated once and never changed |
| Written by | `logbookctl import` | `POST /flights` → `store.AddFlight` |
| `landings_verified` | `0` where night time forced an inference | always `1` — the pilot typed the split |

Three consequences, each of which is a test:

1. **The import's `DELETE` is scoped to `source_book <> 0`.** An unqualified delete would destroy
   every app-entered flight the next time somebody transcribed a page — the exact loss rule §0.2
   forbids. `TestHandEnteredFlightsSurviveAReimport`.
2. **The import's checksums are scoped the same way.** They answer "is the database what the CSVs
   say", and a flight that is in no CSV would make that question unanswerable — the import would
   fail verification on its own correct work, and the only way to pass would be to delete the
   pilot's flight. `TestImportVerificationIgnoresHandEnteredRows`.
3. **The seq bands cannot collide.** Book 3 is still being transcribed, so any hand-entered `seq`
   inside 1..N is a collision waiting for the migration to catch up to it. The bands are disjoint by
   three orders of magnitude, and the higher band also sorts app-entered flights after every page of
   the paper books, which is where a flight flown today belongs.
   `TestHandEnteredSeqCannotCollideWithTheImporter`.

One repair runs on every import: replacing the `aircraft` table sets `aircraft_id` to NULL on the
hand-entered rows that referenced it (`ON DELETE SET NULL`), so the importer **re-links them by
registration** afterwards. Without it a flight typed in the app quietly loses its aircraft link the
first time a page is transcribed, and never gets it back. `TestAReimportRelinksHandEnteredFlights`.

**Validation lives in `internal/entry`**, which is pure and held to 100%. Its posture is the
opposite of the importer's, deliberately: the importer surfaces a problem and imports the row anyway
because the paper is authoritative and nobody can be asked, whereas nothing on the write path is
authoritative yet and the pilot is standing at the form. So a draft that does not make sense is
**refused with the field named**, not stored with a flag. In particular an ambiguous local time — a
DST gap or fold, or a pair that mixes zones — is refused with a message asking for a Zulu time,
rather than being stored as `time_origin = unknown`.

The one duplicate guard: a flight matching an existing `(flight_date, aircraft_reg, off_block_raw)`
is refused with **409**. That is the double-tapped submit button on a phone, and two identical rows
in a legal record inflate a licence total.

## CSV → DB mapping

The 26 CSV columns map as follows. The seven `Cumulative_*` columns are **not imported** — they are
used only as a verification checksum and then discarded.

| CSV column | Destination |
|---|---|
| `Date` | `flight_date` (parse `DD/MM/YYYY`) |
| `Aircraft_Type`, `Aircraft_Reg` | `aircraft_type`, `aircraft_reg` (+ `aircraft_id` lookup) |
| `Departure`, `Arrival` | `dep_place`, `arr_place` |
| `Off_Block`, `On_Block` | `off_block_raw`/`on_block_raw` → `off_block_utc`/`on_block_utc` + `time_origin` |
| `Takeoff`, `Landing` | `takeoff_utc`, `landing_utc` |
| `Block_Time`, `Total_Time` | `block_minutes`, `total_minutes` |
| `Instrument_Time`, `Night_Time` | `instrument_minutes`, `night_minutes` |
| `PIC_Time`, `Student_Time`, `Instructor_Time` | `pic_minutes`, `dual_minutes`, `instructor_minutes` |
| `pic_name` | `pic_name` |
| `Landings` | `landings_day` (seeded) — see below |
| `Remarks` | `remarks` |
| `Cumulative_*` (×7) | **verification only, then dropped** |

**Book seeding**: the first data row of Books 2 and 3 is the carried-over final row of the previous
book. Those two seed rows are **skipped** on import — importing them would count two flights twice.
Book 1 has no seed row; its first row is a real first flight.

**The count is 1293 flights, not 1295.** The three CSVs hold 1295 data rows (395 + 421 + 479); minus
the two seed rows that is 1293. Earlier drafts of `APP.md` said 1295, which was the row count.

| Book | CSV | data rows | seed row | flights |
|---|---|---:|---|---:|
| 1 | `logbook_1_final.csv` | 395 | — | 395 |
| 2 | `logbook_2_final.csv` | 421 | line 2 | 420 |
| 3 | `logbook_3.csv` | 479 | line 2 | 478 |
| | | **1295** | | **1293** |

## Class: how sea and land are decided

The CSVs have no class column. The classification comes from the **registration**, from the seaplane
list in `claude-docs/reference.md` (`OH-CTL, SE-GKT, OH-GKT, OH-PAX, OH-MIL, OH-CTE, OH-CDK`) — not
from the type, because the book only started writing `C172sea` from IMG_6022 and is inconsistent even
after that.

**This is verified, not assumed.** Recomputing `Cumulative_SEP_Sea` row by row from that rule
reproduces the column exactly at every one of the 1293 rows, ending on 407:39. A per-row match over
1293 rows pins each individual row's class, not just the total.

## Time conversion

One function, `timeutil.ToUTC(date, raw)`. Rules, in order:

1. Raw ends in `Z`/`z` → strip it, treat as UTC. `time_origin = utc_as_written`.
2. Raw is plain `HH:MM` → interpret in `Europe/Helsinki` on that date and convert.
   `time_origin = converted_from_local`. Go's `tzdata` gives correct historical EET/EEST transitions;
   it is **embedded in the binary** so this never depends on the server.
3. Raw is empty → `time_origin = none`.
4. The local time is ambiguous (inside the autumn DST fold) or nonexistent (inside the spring gap) →
   `time_origin = unknown`, and the row surfaces for review. Do not guess silently.

**On-block before off-block** means the flight crossed midnight UTC; roll the date forward one day.
This must be a test case.

## Statistics

**Implemented in `internal/stats` (100% covered), not in SQL.** The rows are loaded in `seq` order
and aggregated in Go, so there is one place where a figure is derived and it is the same code the
PDF will use. Nothing is stored (rule §0.5).

Every figure the statistics page reports, over a `From`–`To` range on `flight_date`, with the JSON
field `GET /logbook/api/stats` returns it as. **All durations are integer minutes**; the frontend
formats H:MM.

| Metric | JSON | Derivation |
|---|---|---|
| Seaplane PIC | `sea_pic` | `pic_minutes` where the class is `_SEA` |
| Seaplane instructor | `sea_instructor` | `instructor_minutes` where `_SEA` |
| Landplane PIC | `land_pic` | `pic_minutes` where not `_SEA` |
| Landplane instructor | `land_instructor` | `instructor_minutes` where not `_SEA` |
| Dual | `dual` | `dual_minutes` |
| Total | `total` | `total_minutes` |
| Night | `night` | `night_minutes` |
| Instrument | `instrument` | `instrument_minutes` |
| Landings sea | `landings_sea` | `landings_day + landings_night` where `_SEA` |
| Landings land | `landings_land` | `landings_day + landings_night` where not `_SEA` |
| Landings day | `landings_day` | `landings_day` |
| Landings night | `landings_night` | `landings_night` |

Also returned: `flights`, `sea_total`, `land_total`, `instructor`, `pic`, and **`landings_unverified`**
— how many flights in the range still carry an inferred day/night landing split. That last one is
how the app tells the truth about Task 8 instead of presenting `0` night landings as verified.

**Land is the default**, not sea: a class added to the vocabulary later cannot silently inflate the
seaplane figures a rating depends on. Being wrong in that direction shows up as a land discrepancy.

Note the two landing splits partition the same total on different axes — sea+land and day+night must
each sum to the same grand total, and sea+land time must reconstitute `total`. Those invariants are
asserted both on fixtures and **against all 1293 real flights**
(`internal/stats/realdata_test.go`), which is where a single misclassified row would show up.

### Cumulative totals for the EASA PDF

`stats.Paginate(flights, 15)` walks **`seq`** — never `flight_date`, because 18 rows across the three
books are genuinely out of date order — and returns each page's rows plus the paper's three-row
block: `ThisPage` (TOTAL THIS PAGE), `Previous` (TOTAL PREVIOUS PAGES) and `Total`. The 1293 flights
give **87 pages**, the last holding 3 rows, and the last page's `Total` equals the whole logbook.
Asserted against the real books. This is the rule §0.5 replacement for the CSVs' `Cumulative_*`
columns.

## Known data-quality items

The importer surfaces these and **never auto-fixes them** (rule §0.2). Every one is stored in the
`discrepancies` table and printed by `logbookctl import`, with a book and line number so it is
traceable to a paper page.

**61 discrepancies over 1293 flights, in six live kinds** (56 when first written). The 2026-08-01
reconciliation moved several: eight new night rows each raise a `landings_unverified` flag, fixing
`OK-PDP` removed one `registration_format`, and the line-28 ruling closed the only
`cumulative_break` and the only `component_exceeds_total`. The counts are asserted exactly in
`internal/csvbook/realdata_test.go`, so a new occurrence in a future Book 3 batch becomes a failing
test rather than something that slips through. **The two closed kinds stay in that assertion at
zero** rather than being deleted, so a regression fails loudly.

| kind | n | what it is |
|---|---:|---|
| `landings_unverified` | 30 | rows carrying `Night_Time`; the day/night landing split was inferred (Task 8). Every night row in the books: 20 in Book 1, 3 in Book 2, 7 in Book 3 |
| `registration_format` | 15 | not Finnish `OH-xxx`: `SE-GKT` ×14 (real), `SE-LWI` ×1 (real). Was 16 before the `OK-PDP` fix |
| `date_format` | 8 | Book 2 lines 83–90 transcribed `DD.MM.YYYY` — see below |
| `unknown_aircraft_type` | 4 | type `C192` on `OH-CTL` ×2 and `OH-GKT` ×2; not a real Cessna type |
| `type_conflict` | 3 | one registration written with two types: `OH-CTL`, `OH-GKT`, `OH-CMU` |
| `block_total_mismatch` | 1 | 08/09/2025, block 0:45 vs total 0:38 (already known and correct) |
| `cumulative_break` | **0** | was 1 (Book 1 line 28) until the owner ruled on it — see below |
| `component_exceeds_total` | **0** | the same row |

Notes on the individual items:

- ~~**Book 1 line 28 (28/09/2011, `OH-COF`, EFHF local)**~~ — **✅ closed 2026-08-01.** The row had
  `Total_Time` **1:12** but `Instrument_Time` **1:21** — more instrument time than flight time,
  which is impossible — while its `Cumulative_Instrument` advanced by exactly the 1:12 the flight
  lasted. The importer surfaced it rather than correcting it (rule §0.2); the owner ruled that
  **1:12** is the reading and the CSV was fixed. Instrument **107:14 → 107:05**, which is what the
  `Cumulative_Instrument` column always said, so **no cumulative moved.** This was the only
  `cumulative_break` and the only `component_exceeds_total` in the corpus, and both are now zero.
- **Book 2 lines 83–90 — dates written `DD.MM.YYYY`, NEW.** Eight consecutive rows from a single
  transcription batch. Read day-first, which six of the eight prove on their own (day > 12) and which
  the chronological bracket 15/03 → … → 07/05 confirms. **The two `04.05.2018` rows (89, 90) cannot be
  settled from the cell alone** and are flagged `CONFIRM AGAINST THE PAPER`.
- ~~**`OK-PDP`** (1 row)~~ — **fixed 2026-08-01.** It was a transcription typo for `OH-PDP` and was
  seeding a phantom one-flight aircraft; the owner ruled that any `OK-` registration in these books
  is `OH-`. Aircraft count dropped 39 → 38.
- **Type `C192`** (4 rows) — almost certainly `C172`. The flight keeps the type as written; the
  derived `aircraft` row takes the most-flown type, so `OH-CTL` and `OH-GKT` are seeded `C172`.
- **`OH-CMU` typed as both `C152` (×2) and `C172` (×1)** — reference.md warns `OH-CMU` and `OH-CMV`
  are genuinely different aircraft whose registrations differ only in the last letter, so this needs
  the user's eye rather than a guess.
- **`SE-GKT` → `OH-GKT`** — the same airframe re-registered. Two `aircraft` rows, linked via `notes`.
- **30 rows with `Night_Time`** (was 22, then 28, then 30 as the 2026-08-01 night reconciliation
  progressed) — day/night landing split unverified (Task 8). ⚠ **The split inked at p.62
  (59 night / 3335 day) is stale and recomputes to `68 / 3326`**; the landing *sum* 3394 never
  moved, so this is a correction to the paper, not to the CSV.

### Night time: ✅ closed against the paper, 2026-08-01

The `Night_Time` column summed to **16:47** across all three books against **22:45** inked at page 62
— a **5:58** gap `claude-docs/drift.md` had recorded as *"supplied but not read back"*.

**It is now 22:45, equal to the paper, Δ 0:00.** The gap was entirely inherited from Books 1–2: the
EASA book carried **18:42** into Book 3 against our 12:44, and Book 3 itself reconciled exactly
(18:42 + 4:03 = 22:45), so 22:45 was never a mis-add. The owner read the paper's `Yölentoaika`
column back and photographed seven Book-1 spreads; because the column's `Siirto` figures chain
continuously, it becomes a page-by-page ledger that pins each entry to a row. Six values were added
and one was found sitting on the wrong row, taking night to **20:50**.

The last **1:55** closed with the p.52/53 photograph (`IMG_6048`): **25/02/2014 OH-KLS 0:55**
(full night) and **26/03/2014 OH-TIL 1:00 of 2:01**. That spread's own `Yölentoaika` runs `Siirto`
**9:12 → 11:07**, and 11:07 is exactly p.71's `Siirto` — so pages 54–69 carry no night at all and the
**column is closed, not merely sampled.**

⚠ **Never infer night time from clock times, sunset or time zones** (owner, 2026-08-01). The book's
night column is the only authority.

## Verification: what "verified" means here

Two separate checks, deliberately not conflated:

1. **Fidelity** — does the database hold exactly what the CSVs say? Row count plus nine independent
   checksums (flights, total, PIC, dual, instrument, night, instructor, seaplane, landings) are read
   *back out of SQLite* after writing and compared. A mismatch of one minute rolls the whole
   transaction back. Checked per figure rather than as one grand total, because two errors of
   opposite sign would cancel in a combined number. This is a hard gate.
2. **Consistency** — does the source data agree with itself? All seven `Cumulative_*` series are
   recomputed row by row and compared to the columns the transcription maintained. A break is
   *reported*, not fatal: it is a pre-existing property of the paper record, and refusing to import
   because of it would leave the owner with no application at all. **All seven series now reconcile
   with zero breaks over 1293 rows** — the single break (Book 1 line 28) closed when the owner ruled
   on it. The test asserts zero, which is the stronger claim the corrected data supports.

Row-by-row rather than end-totals, on both: an end-total can be passed by two cancelling errors, and
a break with no line number is not actionable.

## The landings gap

The EASA book splits LANDINGS into DAY and NIGHT columns; the CSV only ever captured the sum, on the
assumption the split could be "inferred later from `Night_Time`". It cannot be inferred reliably — a
flight with night time may still have landed by day, and vice versa.

The importer therefore seeds `landings_day = Landings`, `landings_night = 0`, and sets
`landings_verified = 0` on the **30** rows carrying night time. Those get read off the page images in
Task 8. Everything else is `landings_verified = 1` — a day flight's landings are unambiguously day
landings.

The API surfaces this as `landings_unverified` in the statistics summary, so the app can say the
night-landing figure is not yet read off the paper rather than presenting `0` as a verified truth.
