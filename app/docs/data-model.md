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

## Tables

### `aircraft`
The seed list that makes the new-flight form smart.

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

**Book seeding**: the first data row of each book's CSV is the carried-over final row of the previous
book. Those seed rows are **skipped** on import — importing them would double-count three flights.

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

Surfaced by the importer, **never auto-fixed** (rule §0.2):

- **`OK-PDP`** (1 row) — almost certainly a typo for `OH-PDP`; `claude-docs/reference.md` already
  flags it.
- **Type `C192`** (4 rows: `OH-GKT` ×2, `OH-CTL` ×2) — not a real Cessna type; almost certainly `C172`.
- **`OH-CMU` typed as both `C152` (×2) and `C172` (×1)** — reference.md warns `OH-CMU` and `OH-CMV`
  are genuinely different aircraft whose registrations differ only in the last letter, so this needs
  the user's eye rather than a guess.
- **`SE-GKT` → `OH-GKT`** — the same airframe re-registered. Two `aircraft` rows, linked via `notes`.
- **22 rows with `Night_Time`** — day/night landing split unverified.

## The landings gap

The EASA book splits LANDINGS into DAY and NIGHT columns; the CSV only ever captured the sum, on the
assumption the split could be "inferred later from `Night_Time`". It cannot be inferred reliably — a
flight with night time may still have landed by day, and vice versa.

The importer therefore seeds `landings_day = Landings`, `landings_night = 0`, and sets
`landings_verified = 0` on the 22 rows carrying night time. Those get read off the page images in
Task 8. Everything else is `landings_verified = 1` — a day flight's landings are unambiguously day
landings.
