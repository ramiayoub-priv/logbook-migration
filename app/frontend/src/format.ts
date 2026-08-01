// Formatting at the edge of the application.
//
// Durations cross the API as integer minutes and are minutes everywhere inside
// the backend too. H:MM exists only in this file and only for display, because
// two representations of one figure are two things that can disagree -- and
// this is a legal record (decision log, 2026-08-01).

/** hhmm renders whole minutes as H:MM. Hours are not capped at 24. */
export function hhmm(minutes: number): string {
  const negative = minutes < 0
  const total = Math.abs(Math.trunc(minutes))
  const h = Math.floor(total / 60)
  const m = total % 60
  return `${negative ? '-' : ''}${h}:${String(m).padStart(2, '0')}`
}

/** hhmmOrBlank is hhmm, but a zero renders as nothing. */
export function hhmmOrBlank(minutes: number): string {
  return minutes === 0 ? '' : hhmm(minutes)
}

/**
 * parseHHMM turns what the pilot typed into minutes, or null if it is not a
 * duration.
 *
 * The form uses this to refuse an impossible entry without a round trip. It is
 * never the authority: internal/entry on the server decides what may be
 * written, and this must not be the only thing standing between a typo and the
 * record.
 */
export function parseHHMM(raw: string): number | null {
  const m = /^\s*(\d{1,4}):([0-5]\d)\s*$/.exec(raw)
  if (!m) return null
  return Number(m[1]) * 60 + Number(m[2])
}

/**
 * isoDate formats an instant as the YYYY-MM-DD the API expects.
 *
 * It reads the UTC calendar day rather than the browser's. Every instant in
 * this application is UTC (rule 0.4), and in Helsinki a late-evening local
 * date is already the next day -- which would file a flight under tomorrow.
 */
export function isoDate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

export function todayISO(): string {
  return isoDate(new Date())
}

/** clock renders an RFC3339 instant from the API as HH:MM UTC. */
export function clock(instant: string | null): string {
  if (!instant) return ''
  const d = new Date(instant)
  if (Number.isNaN(d.getTime())) return ''
  return d.toISOString().slice(11, 16)
}
