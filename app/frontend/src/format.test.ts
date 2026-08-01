import { describe, it, expect } from 'vitest'
import { hhmm, hhmmOrBlank, parseHHMM, todayISO, isoDate } from './format'

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
