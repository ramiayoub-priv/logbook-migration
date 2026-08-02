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

// --- HHMM: what the pilot actually types ----------------------------------
//
// Every time field on the new-flight form holds four digits and nothing else:
// "0915", never "09:15" and never "09:15Z". A phone's number pad has no colon
// key and no Z, so anything else is either untypeable or a picker the pilot has
// to tap through. The colon and the Z are put back on the wire by the form --
// the API contract and the server's single conversion authority are unchanged.
//
// EXACTLY four, always. "915" is not read as 09:15: it is equally readable as
// 91:5, and guessing at a time that goes into a legal record is the silent
// corruption rule 0.2 forbids. Four digits or nothing.

/** digits keeps at most four numerals and throws away everything else. */
export function digits(raw: string): string {
  return raw.replace(/\D/g, '').slice(0, 4)
}

/** parseClockDigits reads "0915" as minutes since midnight, or null. */
export function parseClockDigits(raw: string): number | null {
  const m = /^(\d{2})(\d{2})$/.exec(raw)
  if (!m) return null
  const h = Number(m[1])
  const min = Number(m[2])
  if (h > 23 || min > 59) return null
  return h * 60 + min
}

/**
 * parseDurationDigits reads "0115" as 75 minutes, or null.
 *
 * Separate from the clock reading because the hours are not a time of day:
 * 24:00 is a legitimate duration where 24:00 is not a legitimate clock. The
 * server still bounds a single flight at 24 hours.
 */
export function parseDurationDigits(raw: string): number | null {
  const m = /^(\d{2})(\d{2})$/.exec(raw)
  if (!m) return null
  const min = Number(m[2])
  if (min > 59) return null
  return Number(m[1]) * 60 + min
}

/** clockWire turns four digits into the "HH:MM" the API expects, or "". */
export function clockWire(raw: string): string {
  const m = parseClockDigits(raw)
  if (m === null) return ''
  return `${String(Math.floor(m / 60)).padStart(2, '0')}:${String(m % 60).padStart(2, '0')}`
}

/** minutesToDigits writes minutes back into a four-digit field: 75 -> "0115". */
export function minutesToDigits(minutes: number): string {
  const total = Math.max(0, Math.trunc(minutes))
  return `${String(Math.floor(total / 60)).padStart(2, '0')}${String(total % 60).padStart(2, '0')}`
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

/**
 * airMinutes is wheels-up to wheels-down, from the two stored instants.
 *
 * DERIVED at render, never stored and never sent. Rule 0.5 forbids a persisted
 * running total because the seven Cumulative_* columns were this project's
 * largest source of drift, and the reasoning is about any figure that can be
 * computed from others rather than about cumulatives specifically: a second
 * copy of a number is a second thing that can disagree with the first.
 *
 * The instants carry their dates, so a flight that lands after midnight comes
 * out right by plain subtraction. That is NOT the same problem the form solves
 * with blockFrom(), where the pilot types a bare clock with no date and the
 * roll has to be applied by hand.
 *
 * null means "not recorded", which is most rows -- 19 of the 1296 transcribed
 * flights carry airborne times. It must render as blank rather than 0:00: a
 * zero is a claim that the aeroplane never left the ground.
 */
export function airMinutes(takeoff: string | null, landing: string | null): number | null {
  if (!takeoff || !landing) return null
  const a = new Date(takeoff).getTime()
  const b = new Date(landing).getTime()
  if (Number.isNaN(a) || Number.isNaN(b)) return null
  const minutes = Math.round((b - a) / 60000)
  // Cannot arise from stored data: the server builds the pair through
  // timeutil.BlockPair, which rolls forward. If it ever does, a negative air
  // time is not a figure to print.
  return minutes < 0 ? null : minutes
}
