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
- The **first data row** of `logbook_2.csv` is the final row carried over from
  `logbook_1_final.csv` — the seed for all Book 2 cumulative totals.

## Seaplane registrations (count toward Cumulative_SEP_Sea)
These regs are seaplanes; their `Total_Time` adds to `Cumulative_SEP_Sea`. Pay extra attention:
`OH-CTL, SE-GKT, OH-GKT, OH-PAX, OH-MIL, OH-CTE, OH-CDK`

## Aircraft registrations seen so far in Book 2
Common: `OH-PDP` (P28A), `OH-CTL` (C172, seaplane), `OH-GKT`, `OH-CDK`, `OH-CWB`, `OH-CAV`,
`OH-DBS`, `OH-PAX`. Rarer: `OH-TIL, OH-SPH, OH-CTH, OH-CRA, OH-CMO`.
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
