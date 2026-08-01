import { describe, it, expect } from 'vitest'
import {
  hhmm,
  hhmmOrBlank,
  parseHHMM,
  todayISO,
  isoDate,
  digits,
  parseClockDigits,
  parseDurationDigits,
  clockWire,
  minutesToDigits,
} from './format'

// Durations cross the API as integer minutes -- one representation, so no two
// figures can disagree (decision log, 2026-08-01). H:MM exists only here, at
// the edge, and it is what the pilot reads off every page.
describe('hhmm', () => {
  it('renders whole minutes as H:MM', () => {
    expect(hhmm(0)).toBe('0:00')
    expect(hhmm(5)).toBe('0:05')
    expect(hhmm(60)).toBe('1:00')
    expect(hhmm(75)).toBe('1:15')
    expect(hhmm(81)).toBe('1:21')
  })

  // 1219:35 is the frozen whole-logbook total. Hours run past 24 and past 999
  // and must never wrap or be padded to two digits.
  it('does not wrap at 24 hours', () => {
    expect(hhmm(24 * 60)).toBe('24:00')
    expect(hhmm(73175)).toBe('1219:35')
  })

  // A negative duration is impossible in the data, but rendering one as
  // "-1:-30" would be worse than obvious nonsense -- it could be misread.
  it('renders a negative duration unambiguously', () => {
    expect(hhmm(-90)).toBe('-1:30')
  })

  it('blanks a zero only when asked', () => {
    expect(hhmmOrBlank(0)).toBe('')
    expect(hhmmOrBlank(75)).toBe('1:15')
  })
})

// The new-flight form sends H:MM strings and the server parses them. Parsing
// here too means the form can refuse an impossible entry before a round trip,
// but the server is still the authority -- this never decides what is stored.
describe('parseHHMM', () => {
  it('accepts what a pilot writes', () => {
    expect(parseHHMM('1:15')).toBe(75)
    expect(parseHHMM('0:05')).toBe(5)
    expect(parseHHMM('12:00')).toBe(720)
    expect(parseHHMM(' 1:15 ')).toBe(75)
  })

  it('rejects what is not a duration', () => {
    for (const bad of ['', '1.25', '115', '1:', ':15', 'x:yy', '1:60', '1:5']) {
      expect(parseHHMM(bad), `${bad} should not parse`).toBeNull()
    }
  })
})

// --- Four digits, and nothing else ----------------------------------------
//
// Every time on the form is typed as HHMM on a plain number pad: no colon, no
// Z, no picker. The colon was untypeable on a phone (decision log 2026-08-01)
// and a picker is three taps for something a pilot reads off a clock.
describe('digits', () => {
  it('keeps only the numbers, and only four of them', () => {
    expect(digits('0915')).toBe('0915')
    expect(digits('09:15')).toBe('0915')
    expect(digits('09:15Z')).toBe('0915')
    // Stripping is not padding: a half-typed field stays half-typed, and
    // parseClockDigits below is what refuses it.
    expect(digits(' 9 15 ')).toBe('915')
    expect(digits('091545')).toBe('0915')
    expect(digits('')).toBe('')
  })
})

describe('parseClockDigits', () => {
  it('reads HHMM as a time of day', () => {
    expect(parseClockDigits('0000')).toBe(0)
    expect(parseClockDigits('0915')).toBe(9 * 60 + 15)
    expect(parseClockDigits('2359')).toBe(23 * 60 + 59)
  })

  // Exactly four, always: three digits is ambiguous between 09:15 typed short
  // and 91:5 mistyped, and guessing at a legal record is what rule 0.2 forbids.
  it('refuses anything that is not four digits, or not a real clock time', () => {
    for (const bad of ['', '9', '91', '915', '09151', '2400', '0960', '09:15']) {
      expect(parseClockDigits(bad), `${bad} should not parse`).toBeNull()
    }
  })
})

describe('parseDurationDigits', () => {
  // A duration is HHMM too, but its hours are not a clock: 24:00 is a real
  // duration and 99:59 is a real typo guard rather than a real flight.
  it('reads HHMM as a duration', () => {
    expect(parseDurationDigits('0000')).toBe(0)
    expect(parseDurationDigits('0115')).toBe(75)
    expect(parseDurationDigits('1230')).toBe(750)
    expect(parseDurationDigits('2400')).toBe(1440)
  })

  it('refuses anything that is not four digits, or has impossible minutes', () => {
    for (const bad of ['', '115', '1:15', '0160', '01599']) {
      expect(parseDurationDigits(bad), `${bad} should not parse`).toBeNull()
    }
  })
})

describe('clockWire', () => {
  // The wire format is unchanged -- "HH:MM", with the Z added by the form's
  // zone toggle -- so the server's single conversion authority is untouched.
  // Only what the pilot types changed.
  it('turns four digits into what the API expects', () => {
    expect(clockWire('0915')).toBe('09:15')
    expect(clockWire('0005')).toBe('00:05')
  })

  it('leaves a blank field blank rather than inventing a time', () => {
    expect(clockWire('')).toBe('')
    expect(clockWire('091')).toBe('')
  })
})

describe('minutesToDigits', () => {
  it('writes minutes back as the four digits the field holds', () => {
    expect(minutesToDigits(75)).toBe('0115')
    expect(minutesToDigits(0)).toBe('0000')
    expect(minutesToDigits(600)).toBe('1000')
  })
})

describe('isoDate', () => {
  it('formats a date as the API wants it', () => {
    expect(isoDate(new Date(Date.UTC(2026, 6, 30, 15, 0, 0)))).toBe('2026-07-30')
  })

  // The date has to be the UTC calendar day, not the browser's local one:
  // every instant in this application is UTC (rule 0.4), and in Helsinki a
  // late-evening local date is already tomorrow.
  it('uses the UTC day, not the local one', () => {
    expect(isoDate(new Date(Date.UTC(2026, 6, 30, 23, 30, 0)))).toBe('2026-07-30')
  })

  it('todayISO returns a well-formed date', () => {
    expect(todayISO()).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })
})
