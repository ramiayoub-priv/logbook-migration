#!/usr/bin/env python3
"""Flatten laskukierros_flights.json (GET /api/v1/flights) into a grep-friendly CSV.

Why this exists: the site's CSV export (`GET /export/pilotFlights`) only returns
flights where the user is the PRIMARY pilot -- 128 rows. Every flight he
*instructed* is filed under the pupil's account and is absent from it. The JSON
API returns those too (228 rows), which is why it is the canonical dump.

  python3 laskukierros_to_csv.py            # rebuild laskukierros_flights.csv

⚠ Times are LOCAL wall-clock despite the '+00:00' suffix the API stamps on them.
  Proved against the paper book: 20/08/2024 blockStart 18:50 == paper 15:50Z.
⚠ The pilotFuncTime* flags describe the PRIMARY pilot, not necessarily the user.
  On an instructing row the primary pilot is the student, so func_student=true.
"""
import csv, json, os

HERE = os.path.dirname(os.path.abspath(__file__))
ME = 'Rami Ayoub'
COLS = ['date', 'reg', 'model', 'dep', 'arr', 'block_start', 'block_stop',
        'takeoff', 'landing', 'block_min', 'air_min', 'ldg_day', 'ldg_night',
        'night_min', 'ifr_min', 'rami_role', 'other_name', 'other_role',
        'func_pic', 'func_copilot', 'func_dual', 'func_instructor',
        'func_student', 'flight_type', 'notes', 'id']


def hhmm(ts):
    return ts[11:16] if ts else ''


def rows(flights):
    for r in sorted(flights, key=lambda x: (x['date'][:10], x.get('blockStartTime') or '')):
        primary = r.get('pilotName')
        if primary == ME:
            rami_role, other_name = 'pilot', r.get('pilotTwoName') or ''
            other_role = r.get('pilotTwoRole') or ''
        else:
            rami_role = r.get('pilotTwoRole') or 'pilotTwo'
            other_name, other_role = primary or '', 'pilot'
        yield {
            'date': r['date'][:10], 'reg': r.get('planeReg') or '',
            'model': r.get('planeModel') or '',
            'dep': r.get('departureField') or '', 'arr': r.get('arrivalField') or '',
            'block_start': hhmm(r.get('blockStartTime')),
            'block_stop': hhmm(r.get('blockStopTime')),
            'takeoff': hhmm(r.get('takeoffTime')), 'landing': hhmm(r.get('landingTime')),
            'block_min': r.get('flightTimeBlockOnOff'),
            'air_min': r.get('flightTimeTakeoffLanding'),
            'ldg_day': r.get('landingsDay'), 'ldg_night': r.get('landingsNight'),
            'night_min': r.get('opConditionTimeNight'),
            'ifr_min': r.get('opConditionTimeIFR'),
            'rami_role': rami_role, 'other_name': other_name, 'other_role': other_role,
            'func_pic': r.get('pilotFuncTimePIC'),
            'func_copilot': r.get('pilotFuncTimeCoPilot'),
            'func_dual': r.get('pilotFuncTimeDual'),
            'func_instructor': r.get('pilotFuncTimeInstructor'),
            'func_student': r.get('pilotFuncTimeStudent'),
            'flight_type': r.get('flightType') or '',
            'notes': (r.get('pilotComments') or '').replace('\n', ' ').strip(),
            'id': r.get('id'),
        }


def main():
    src = os.path.join(HERE, 'laskukierros_flights.json')
    dst = os.path.join(HERE, 'laskukierros_flights.csv')
    flights = json.load(open(src, encoding='utf-8'))
    with open(dst, 'w', newline='', encoding='utf-8') as fh:
        w = csv.DictWriter(fh, fieldnames=COLS)
        w.writeheader()
        for row in rows(flights):
            w.writerow(row)
    print(f'wrote {dst} ({len(flights)} flights)')


if __name__ == '__main__':
    main()
