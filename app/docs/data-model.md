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
| `source_book`, `source_row` | INTEGER | Provenance: which CSV and which line. Makes any figure traceable back to paper. |

### `users`, `sessions`
See `security.md`.

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

Every figure the statistics page reports, over a `From`–`To` range on `flight_date`:

| Metric | Derivation |
|---|---|
| Seaplane PIC | `SUM(pic_minutes)` where `class` ends `_SEA` |
| Seaplane instructor | `SUM(instructor_minutes)` where `class` ends `_SEA` |
| Landplane PIC | `SUM(pic_minutes)` where `class` ends `_LAND` |
| Landplane instructor | `SUM(instructor_minutes)` where `class` ends `_LAND` |
| Dual | `SUM(dual_minutes)` |
| Total | `SUM(total_minutes)` |
| Night | `SUM(night_minutes)` |
| Instrument | `SUM(instrument_minutes)` |
| Landings sea | `SUM(landings_day + landings_night)` where `_SEA` |
| Landings land | `SUM(landings_day + landings_night)` where `_LAND` |
| Landings day | `SUM(landings_day)` |
| Landings night | `SUM(landings_night)` |

Note the two landing splits partition the same total on different axes — sea+land and day+night must
each sum to the same grand total. That is a cheap and effective invariant: **assert it in a test.**

## Known data-quality items

The importer surfaces these and **never auto-fixes them** (rule §0.2). Every one is stored in the
`discrepancies` table and printed by `logbookctl import`, with a book and line number so it is
traceable to a paper page.

**61 discrepancies over 1293 flights, in eight kinds** (56 when first written; the 2026-08-01
night reconciliation added six night rows, which each raise a `landings_unverified` flag, and
fixing `OK-PDP` removed one `registration_format`). The counts are asserted in
`internal/csvbook/realdata_test.go`, so a new occurrence in a future Book 3 batch becomes a failing
test rather than something that slips through.

| kind | n | what it is |
|---|---:|---|
| `landings_unverified` | 22 | rows carrying `Night_Time`; the day/night landing split was inferred (Task 8) |
| `registration_format` | 16 | not Finnish `OH-xxx`: `SE-GKT` ×14 (real), `SE-LWI` ×1 (real), `OK-PDP` ×1 (a slip) |
| `date_format` | 8 | Book 2 lines 83–90 transcribed `DD.MM.YYYY` — see below |
| `unknown_aircraft_type` | 4 | type `C192` on `OH-CTL` ×2 and `OH-GKT` ×2; not a real Cessna type |
| `type_conflict` | 3 | one registration written with two types: `OH-CTL`, `OH-GKT`, `OH-CMU` |
| `cumulative_break` | 1 | Book 1 line 28 — see below |
| `component_exceeds_total` | 1 | the same row |
| `block_total_mismatch` | 1 | 08/09/2025, block 0:45 vs total 0:38 (already known and correct) |

Notes on the individual items:

- **Book 1 line 28 (28/09/2011, `OH-COF`, EFHF local) — NEW, found building the importer.** The row
  has `Total_Time` **1:12** but `Instrument_Time` **1:21** — more instrument time than flight time,
  which is impossible — while its `Cumulative_Instrument` advances by exactly the 1:12 the flight
  lasted. `1:21` is almost certainly a transposition of `1:12`. **Not corrected.** Consequence:
  summing the rows gives instrument **107:14**, while the CSV's own `Cumulative_Instrument` ends at
  **107:05**. The app's totals follow the rows, so it reports 107:14 until the owner rules.
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
- **28 rows with `Night_Time`** (was 22 before the 2026-08-01 night reconciliation) — day/night
  landing split unverified (Task 8). ⚠ **The split inked at p.62 (59 night / 3335 day) is now
  stale and must be recomputed**; `Cumulative_Landings` is unaffected, only the split.

### Night time: reconciled against the paper 2026-08-01 — 1:55 still open

The `Night_Time` column summed to **16:47** across all three books against **22:45** inked at page 62
— a **5:58** gap `claude-docs/drift.md` had recorded as *"supplied but not read back"*.

**Resolved down to 1:55 the same day.** The gap was entirely inherited from Books 1–2: the EASA book
carried **18:42** into Book 3 against our 12:44, and Book 3 itself reconciled exactly
(18:42 + 4:03 = 22:45), so 22:45 was never a mis-add. The owner then read the paper's `Yölentoaika`
column back and photographed seven Book-1 spreads; because the column's `Siirto` figures chain
continuously, it becomes a page-by-page ledger that pins each entry to a row. Six values were added
and one was found sitting on the wrong row. **Night is now 20:50**, and our running total matches the
paper's `Siirto` at every checkpoint through 30/11/2013.

⏸ **The residual 1:55 is one unphotographed page range** — pp. 52–69 (Mar–Aug 2014), where the book
runs 9:12 → 11:07 and our CSV has nothing. Tracked as item E at the top of `claude-docs/drift.md`.

This remains a migration question about the paper, not an import question: the import's job is
fidelity to the CSV, and it reports whatever the rows say.

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
   because of it would leave the owner with no application at all. Exactly one break survives over
   1293 rows (Book 1 line 28, above).

Row-by-row rather than end-totals, on both: an end-total can be passed by two cancelling errors, and
a break with no line number is not actionable.

## The landings gap

The EASA book splits LANDINGS into DAY and NIGHT columns; the CSV only ever captured the sum, on the
assumption the split could be "inferred later from `Night_Time`". It cannot be inferred reliably — a
flight with night time may still have landed by day, and vice versa.

The importer therefore seeds `landings_day = Landings`, `landings_night = 0`, and sets
`landings_verified = 0` on the 22 rows carrying night time. Those get read off the page images in
Task 8. Everything else is `landings_verified = 1` — a day flight's landings are unambiguously day
landings.
