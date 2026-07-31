# Workflow — Processing One Page

Work **one page at a time, with the user**. Never batch-process future pages.

> **Book 3 (active, EASA) note:** images are `logbook-3/IMG_XXXX.JPEG`. **Orientation varies — always
> check first** (thumbnail it): 6007–6011 are stored sideways and need `Image.open(p).rotate(90, expand=True)`,
> but **6012 onward are already upright**. One image = one two-page spread (~15 flights). The EASA book
> prints a **"TOTAL THIS PAGE"** row per column — read it straight off
> for the cross-check deltas (`d_total/d_pic/d_land/d_instr`). Times are **UTC → `Z`** unless marked
> `LT` — but **a single spread can mix local and UTC rows** (proved on IMG_6014): if rows appear to
> overlap in time, suspect a zone mix and ask the user *which rows*. Append with
> `logbook_tools.py <batch.json> --csv logbook_3.csv --append`. See `reference.md`
> "Book 3" for the full EASA→schema column mapping (Dual→Student, SE-IFR→Instrument, etc.).

## Steps
1. **Confirm the page.** With the user, identify the next unprocessed page by its *dates*, not
   its image filename (filenames are not chronological). Images are in `logbook-2/IMG_XXXX.jpg`.
   If only a `.HEIC` exists, convert with `convert_heic_to_jpg.py` first.
2. **Transcribe.** Claude reads the JPG directly and extracts every flight row: Date, Aircraft
   Type, Reg, Departure, Arrival, Off/On block, Takeoff/Landing, times, PIC/Student/Instructor
   time, pic_name, Landings, Remarks.
3. **Present for verification.** Show the transcribed rows to the user. **Do not append until the
   user confirms.** The user verifies against the paper book / image.
4. **Sanity-check before appending** (report discrepancies, don't silently fix):
   - On-block − off-block should ≈ Block_Time ≈ Total_Time. Flag mismatches.
   - Watch aircraft registrations — flag anything outside the known set (`reference.md`).
   - Flag OCR-style oddities (impossible dates, transposed reg letters).
5. **Append & recompute.** Append verified rows to `logbook_2.csv` and recompute the cumulative
   columns from the last existing row (per-image cumulative values, if any, are placeholders).

## Flight role classification (PIC / Student / Instructor)
Three cases — read the paper's Päällikkö (PIC) / Oppilas (Student) / Opettaja (Instructor) columns:
- **Normal flight (default):** PIC time only. `PIC_Time = Total_Time`, `pic_name = self`.
- **Pilot is the STUDENT** (e.g. IR training, dual received): `Student_Time = Total_Time`,
  `PIC_Time` blank, `pic_name = the instructor's name`. Feeds `Cumulative_Student`. Often marked
  with Zulu times. Do NOT infer this — the user must confirm the row and give the instructor name.
- **Pilot is the INSTRUCTOR** (instructing a student): the flight is logged as PIC **and**
  Instructor simultaneously — `PIC_Time = Instructor_Time = Total_Time`, `pic_name = self`. Feeds
  both `Cumulative_PIC` and `Cumulative_Instructor`. Confirm with the user.
- **Never infer** Student or Instructor time from blanks or names. Default to PIC unless the user
  explicitly says otherwise (with the row number and the other pilot's name).

## Cumulative recompute rules
Seed each run from the **last row already in `logbook_2.csv`**, then per new row:
- `Cumulative_Total` += `Total_Time`
- `Cumulative_PIC` += `PIC_Time`; if `PIC_Time` is blank, treat the row as PIC and add
  `Total_Time` — **unless** the user explicitly marked it student time.
- `Cumulative_Student` += `Student_Time` (only for rows the user explicitly marked student).
- `Cumulative_Instrument` += `Instrument_Time`
- `Cumulative_Instructor` += `Instructor_Time`
- `Cumulative_SEP_Sea`: carry forward, **plus** `Total_Time` when the reg is a seaplane
  (see `reference.md` for the seaplane reg list).
- `Cumulative_Landings` += `Landings`

Times are `H:MM` / `HH:MM` (60-minute carry). Add as minutes, then reformat.

## After appending
- If a correction changes a row that later rows already built on, record it in `drift.md` and
  fix cumulatives from that row forward.
- Update the checkpoint block in `resume.md` (last row + all cumulative totals + row count).
