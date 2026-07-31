#!/usr/bin/env python3
"""Batch-append helper for logbook_2.csv.

Given a JSON batch (one entry per book page) of transcribed flights plus the
paper page's own totals, this:
  1. seeds cumulatives from the current last row of logbook_2.csv,
  2. computes every cumulative column for the new rows,
  3. cross-checks each page against the paper's stated per-page deltas
     (total time, PIC time, landings) — these are offset-independent, so they
     catch transcription errors regardless of the known paper-vs-ours drift,
  4. re-derives Total_Time from Off/On block as an extra sanity check,
  5. prints a review table + the exact CSV lines, and with --append writes them.

Usage:
  python3 logbook_tools.py batch.json           # dry run: checks + preview
  python3 logbook_tools.py batch.json --append   # append verified rows

batch.json schema (list of pages):
  [{"img": "IMG_4924",
    "paper": {"d_total": "6:20", "d_pic": "6:20", "d_land": 15},
    "flights": [
      {"date":"12/02/2019","type":"C152","reg":"OH-CRA","dep":"EFHF","arr":"EFHF",
       "off":"13:14","on":"14:32","total":"1:18","land":4}
       # optional per-flight: "student": true, "pic_name": "Name",
       #   "instructor": "0:30", "instrument": "0:12", "night": "0:20"
    ]}]
By default every flight is PIC/self. Seaplane SEP_Sea is auto-detected by reg.
"""
import csv, json, os, sys

CSV = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'logbook_2.csv')
SEAPLANES = {'OH-CTL', 'SE-GKT', 'OH-GKT', 'OH-PAX', 'OH-MIL', 'OH-CTE', 'OH-CDK'}
HEADER = ["Date","Aircraft_Type","Aircraft_Reg","Departure","Arrival","Off_Block",
    "On_Block","Takeoff","Landing","Block_Time","Total_Time","Instrument_Time",
    "Night_Time","PIC_Time","Student_Time","Instructor_Time","pic_name","Landings",
    "Remarks","Cumulative_Total","Cumulative_PIC","Cumulative_Student",
    "Cumulative_Instrument","Cumulative_SEP_Sea","Cumulative_Landings","Cumulative_Instructor"]

def m(s):
    s = (s or '').strip()
    if not s: return 0
    h, mm = s.split(':'); return int(h) * 60 + int(mm)

def fmt(x): return f'{x // 60}:{x % 60:02d}'

def clock(s):
    """Parse a clock time that may carry a 'Z' (Zulu/UTC) suffix, e.g. '07:56Z'."""
    return m((s or '').rstrip('Zz'))

def diff_hhmm(off, on):
    d = clock(on) - clock(off)
    return d + 24 * 60 if d < 0 else d   # tolerate midnight wrap

def main():
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    args = [a for a in sys.argv[1:] if not a.startswith('--')]
    batch = json.load(open(args[0]))
    do_append = '--append' in sys.argv
    # --csv PATH targets a different logbook (e.g. logbook_3.csv); default Book 2.
    target = CSV
    if '--csv' in sys.argv:
        p = sys.argv[sys.argv.index('--csv') + 1]
        target = p if os.path.isabs(p) else os.path.join(os.path.dirname(CSV), p)
    rows = list(csv.reader(open(target)))
    idx = {c: i for i, c in enumerate(rows[0])}
    last = rows[-1]
    ct, cp, cs, ci, csea, cl, cins = (m(last[idx['Cumulative_Total']]),
        m(last[idx['Cumulative_PIC']]), m(last[idx['Cumulative_Student']]),
        m(last[idx['Cumulative_Instrument']]), m(last[idx['Cumulative_SEP_Sea']]),
        int(last[idx['Cumulative_Landings']]), m(last[idx['Cumulative_Instructor']]))

    out_lines, problems, warnings = [], [], []
    BLOCK_TOL = 5  # minutes: hand-logged block vs logged Lentoaika routinely differ 1-3 min
    print(f"seed: {last[idx['Date']]} {last[idx['Aircraft_Reg']]}  "
          f"Total {fmt(ct)} PIC {fmt(cp)} Land {cl}\n")
    for page in batch:
        pt = pp = pl = ps = pins = 0
        print(f"=== {page['img']} ===")
        for f in page['flights']:
            tot = m(f['total'])
            # sanity: off/on block should ≈ total. Small diffs (<=BLOCK_TOL) are
            # ordinary hand-rounding (logged Lentoaika is authoritative) -> warn only;
            # a larger gap likely signals a real transcription error -> block.
            if f.get('off') and f.get('on'):
                bdiff = diff_hhmm(f['off'], f['on'])
                if bdiff != tot:
                    msg = (f"{page['img']} {f['date']} {f['reg']}: off/on {f['off']}-{f['on']} "
                           f"= {fmt(bdiff)} != total {f['total']}")
                    (warnings if abs(bdiff - tot) <= BLOCK_TOL else problems).append(msg)
            student = f.get('student', False)
            pic_t = 0 if student else tot
            stu_t = tot if student else 0
            ins_t = m(f.get('instructor'))
            ct += tot; cp += pic_t; cs += stu_t; ci += m(f.get('instrument'))
            cins += ins_t
            if f['reg'] in SEAPLANES: csea += tot
            cl += int(f['land'])
            pt += tot; pp += pic_t; pl += int(f['land']); ps += stu_t; pins += ins_t
            row = [f['date'], f['type'], f['reg'], f['dep'], f['arr'], f.get('off',''),
                f.get('on',''), '', '', f['total'], f['total'], f.get('instrument',''),
                f.get('night',''), fmt(pic_t) if pic_t else '', fmt(stu_t) if stu_t else '',
                f.get('instructor',''), f.get('pic_name','self'), str(f['land']), '',
                fmt(ct), fmt(cp), fmt(cs), fmt(ci), fmt(csea), str(cl), fmt(cins)]
            out_lines.append(row)
            print(f"  {f['date']}  {f['reg']:7} {f['total']:>4}  land {f['land']}  "
                  f"-> cumT {fmt(ct)} cumPIC {fmt(cp)} cumLand {cl}"
                  + ('  [SEAPLANE]' if f['reg'] in SEAPLANES else ''))
        # cross-check page deltas vs paper
        pap = page.get('paper', {})
        checks = [('total', pt, m(pap.get('d_total','')) if pap.get('d_total') else None),
                  ('pic', pp, m(pap.get('d_pic','')) if pap.get('d_pic') else None),
                  ('student', ps, m(pap.get('d_student','')) if pap.get('d_student') else None),
                  ('instr', pins, m(pap.get('d_instr','')) if pap.get('d_instr') else None),
                  ('land', pl, pap.get('d_land'))]
        for name, got, exp in checks:
            if exp is None: continue
            gv = got if name != 'land' else got
            ev = exp
            ok = (got == exp)
            disp = fmt(got) if name != 'land' else got
            edisp = fmt(exp) if name != 'land' else exp
            print(f"  paper Δ{name}: computed {disp} vs paper {edisp}  {'OK' if ok else 'MISMATCH'}")
            if not ok:
                problems.append(f"{page['img']} page Δ{name}: computed {disp} != paper {edisp}")
        print()

    if warnings:
        print("WARNINGS (block vs total within tolerance; logged time authoritative):")
        for w in warnings: print("  -", w)
        print()
    if problems:
        print("PROBLEMS:")
        for p in problems: print("  -", p)
        print("\nNot appending (fix batch first)." if do_append else "")
        if do_append: sys.exit(1)
    else:
        print("All page cross-checks OK.")

    print(f"\nnew checkpoint: {out_lines[-1][0]} {out_lines[-1][2]}  "
          f"Total {out_lines[-1][19]} PIC {out_lines[-1][20]} Land {out_lines[-1][24]}  "
          f"(+{len(out_lines)} rows)")

    if do_append and not problems:
        with open(target, 'a', newline='') as fh:
            w = csv.writer(fh, quoting=csv.QUOTE_ALL)
            for r in out_lines: w.writerow(r)
        print(f"\nAppended {len(out_lines)} rows to {target}")

if __name__ == '__main__':
    main()
