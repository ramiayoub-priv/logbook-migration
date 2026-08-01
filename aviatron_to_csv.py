#!/usr/bin/env python3
"""Extract Aviatron.pdf's flight records into a flat CSV for cross-checking.

Aviatron is the electronic-logbook export for the pilot's own / Blue Skies aircraft
(OH-GKT, OH-PIF, OH-DBS, OH-TIL, OH-DBE) — the fleet `laskukierros_flights.csv` does
NOT cover. All its times are UTC. See claude-docs/reference.md.

Each flight is two stacked records in the PDF text:

    <blank>                          DD.MM.YYYY   DD.MM.YYYY                    Ayoub,
    35472 Kahvisaari     Kelvenne                       0:37:00 4      OMA HAR
                                     15:53:00 UTC 16:30:00 UTC                  Rami
    RIVI  ILMA-ALUS      OUT          IN           AIR     HLÖM ...
                                     DD.MM.YYYY   DD.MM.YYYY
    105   OH-GKT                                          0:29:00 1
                                     15:57:00 UTC 16:26:00 UTC

The first is the BLOCK record (block time + landings) — the one to cross-check the
paper against. The second (`RIVI`) is AIRBORNE time and reads ~5 min shorter; it also
carries the registration. We pair them up and emit one row per flight.

Usage:
  python3 aviatron_to_csv.py                     # -> aviatron_flights.csv
  python3 aviatron_to_csv.py --from 2025-08-01   # filter, print to stdout
"""
import csv, os, re, subprocess, sys

HERE = os.path.dirname(os.path.abspath(__file__))
PDF = os.path.join(HERE, 'Aviatron.pdf')
OUT = os.path.join(HERE, 'aviatron_flights.csv')

DATE = re.compile(r'(\d{2}\.\d{2}\.\d{4})')
TIME = re.compile(r'(\d{2}:\d{2}):\d{2} UTC')
# Block record: id, departure, arrival, h:mm:ss, landings.
# Arrival is the last whitespace-run before the wide gap to the block time — a long
# departure name can crowd the column so the two places end up one space apart
# (e.g. "19161 Hillosensalmk Kahvisaari"), so don't require a 2-space split there.
BLOCK = re.compile(r'^\s*(\d{4,6})\s+(\S.*?)\s+(\S+)\s{2,}(\d+:\d{2}:\d{2})\s+(\d+)\b')
# rivi record: seq, registration, h:mm:ss, pax
RIVI = re.compile(r'^\s*(\d{1,4})\s+(OH-\w+)\s+(\d+:\d{2}:\d{2})\s+(\d+)\b')


def hhmm(s):
    h, m, _ = s.split(':')
    return f'{int(h)}:{m}'


def parse(lines):
    """Walk the text and pair each block record with the RIVI record beneath it."""
    flights, pending = [], None
    for i, ln in enumerate(lines):
        mb = BLOCK.match(ln)
        if mb:
            fid, dep, arr, block, land = mb.groups()
            # date sits on the line above, times on the line below
            date = DATE.search(lines[i - 1]) if i else None
            times = TIME.findall(lines[i + 1]) if i + 1 < len(lines) else []
            if not (date and len(times) == 2):
                continue
            d, m_, y = date.group(1).split('.')
            if pending:
                flights.append(pending)
            pending = {
                'date': f'{d}/{m_}/{y}', 'reg': '', 'dep': dep.strip(), 'arr': arr.strip(),
                'off': times[0], 'on': times[1], 'block': hhmm(block), 'ldg': land,
                'air': '', 'id': fid,
            }
            continue
        mr = RIVI.match(ln)
        if mr and pending and not pending['reg']:
            _, reg, air, _pax = mr.groups()
            pending['reg'], pending['air'] = reg, hhmm(air)
    if pending:
        flights.append(pending)
    return flights


def main():
    txt = subprocess.run(['pdftotext', '-layout', PDF, '-'],
                         capture_output=True, text=True, check=True).stdout
    flights = parse(txt.splitlines())
    # PDF order is chronological; sort defensively by ISO date
    def iso(f):
        d, m_, y = f['date'].split('/')
        return f'{y}-{m_}-{d}'
    flights.sort(key=iso)

    lo = sys.argv[sys.argv.index('--from') + 1] if '--from' in sys.argv else None
    cols = ['date', 'reg', 'dep', 'arr', 'off', 'on', 'block', 'air', 'ldg', 'id']
    sel = [f for f in flights if not lo or iso(f) >= lo]

    if lo:
        w = csv.DictWriter(sys.stdout, cols, extrasaction='ignore')
        w.writeheader(); w.writerows(sel)
    else:
        with open(OUT, 'w', newline='') as fh:
            w = csv.DictWriter(fh, cols, extrasaction='ignore')
            w.writeheader(); w.writerows(sel)
        regs = {}
        for f in sel:
            regs[f['reg']] = regs.get(f['reg'], 0) + 1
        print(f'{len(sel)} flights -> {OUT}')
        print(f"  {sel[0]['date']} .. {sel[-1]['date']}")
        print('  ' + '  '.join(f'{r or "?"} x{n}' for r, n in sorted(regs.items())))


if __name__ == '__main__':
    main()
