// The fetch layer.
//
// The session cookie is HttpOnly, Secure, SameSite=Lax and scoped to /logbook
// (docs/security.md). JavaScript therefore cannot read it and must not try:
// every request just sends `credentials: 'same-origin'` and lets the browser
// carry it. There is no token in localStorage, no Authorization header, and
// nothing to steal from the page.

export const API_BASE = '/logbook/api'

/** ApiError carries the status so a caller can tell 401 from 409 from 500. */
export class ApiError extends Error {
  readonly status: number
  /** Field-level messages from the new-flight endpoint, if it sent any. */
  readonly fields: FieldError[]
  /** Seconds the server asked us to wait, from Retry-After on a 429. */
  readonly retryAfter: number | null

  constructor(status: number, message: string, fields: FieldError[] = [], retryAfter: number | null = null) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.fields = fields
    this.retryAfter = retryAfter
  }

  /** The session is gone: expired, revoked, or the account was disabled. */
  get isUnauthenticated(): boolean {
    return this.status === 401
  }
}

export interface FieldError {
  field: string
  message: string
}

/**
 * request performs one API call.
 *
 * A mutating request sends Content-Type: application/json, which is itself a
 * CSRF control: a cross-origin HTML form cannot send that content type, so the
 * classic form-post attack cannot reach this API at all. The browser sets
 * Origin automatically and the server checks it.
 */
async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${API_BASE}${path}`, {
      ...init,
      // The whole authentication scheme. Without it the cookie is not sent and
      // every private endpoint answers 401.
      credentials: 'same-origin',
      headers: {
        ...(init.body ? { 'Content-Type': 'application/json' } : {}),
        ...init.headers,
      },
    })
  } catch (cause) {
    // A network failure is not an empty result. Saying "you have flown
    // nothing" because the phone lost signal is the silent-corruption failure
    // rule 0.2 forbids, so it surfaces as an error the page has to render.
    throw new ApiError(0, 'Could not reach the logbook. Check your connection.', [], null)
  }

  if (res.status === 204) return undefined as T

  const body = await readJSON(res)

  if (!res.ok) {
    const retry = Number(res.headers.get('Retry-After'))
    throw new ApiError(
      res.status,
      (body?.error as string) || `The server refused the request (${res.status}).`,
      (body?.fields as FieldError[]) ?? [],
      Number.isFinite(retry) && retry > 0 ? retry : null,
    )
  }
  return body as T
}

async function readJSON(res: Response): Promise<Record<string, unknown> | null> {
  const text = await res.text()
  if (!text) return null
  try {
    return JSON.parse(text) as Record<string, unknown>
  } catch {
    return null
  }
}

// --- Types, mirroring what cmd/server actually sends ------------------------

export interface Flight {
  seq: number
  date: string
  aircraft_type: string
  aircraft_reg: string
  class: string
  dep_place: string
  arr_place: string
  off_block_utc: string | null
  on_block_utc: string | null
  off_block_raw: string
  on_block_raw: string
  time_origin: string
  /**
   * The optional airborne pair, null on most rows. Present on the wire because
   * the edit form has to be able to SHOW them: a form that submits a field it
   * cannot display erases that field on the next save.
   */
  takeoff_utc: string | null
  landing_utc: string | null
  block_minutes: number
  total_minutes: number
  night_minutes: number
  instrument_minutes: number
  pic_minutes: number
  dual_minutes: number
  instructor_minutes: number
  pic_name: string
  landings_day: number
  landings_night: number
  landings_verified: boolean
  remarks: string
  source_book: number
  source_row: number
}

export interface Summary {
  flights: number
  total: number
  pic: number
  dual: number
  night: number
  instrument: number
  instructor: number
  sea_total: number
  sea_pic: number
  sea_instructor: number
  land_total: number
  land_pic: number
  land_instructor: number
  landings_day: number
  landings_night: number
  landings_sea: number
  landings_land: number
  /** How many flights in range still carry an inferred day/night split. */
  landings_unverified: number
}

export interface Aircraft {
  registration: string
  type: string
  default_class: string
  ifr_capable: boolean
  active: boolean
  notes: string
}

/**
 * AircraftTime is one aeroplane's time over a range, in the two currencies it
 * gets charged in.
 *
 * Block and air are separate fields with separate coverage, and that is the
 * point rather than an accident of the schema: some owners charge block time
 * and some charge air time, block is known for every flight, and air is known
 * for 19 of the 1296 transcribed ones. `air_minutes` must never be shown
 * without `air_known` beside it.
 */
export interface AircraftTime {
  registration: string
  /** Every distinct type written for this registration. Normally one. */
  types: string[]
  flights: number
  block_minutes: number
  air_minutes: number
  /** How many of those flights recorded both airborne times. */
  air_known: number
  air_missing: number
  /** Flights whose block time differs from the total the licence runs on. */
  block_differs_from_total: number
}

export interface Discrepancy {
  kind: string
  book: number
  row: number
  date: string
  detail: string
}

export interface SessionInfo {
  id: number
  created_at: string
  last_used_at: string
  expires_at: string
  user_agent: string
  ip: string
  current: boolean
}

export interface User {
  user_id: number
  username: string
}

export interface DateRange {
  // Explicitly `| undefined` rather than just optional: with
  // exactOptionalPropertyTypes an unset bound has to be assignable, and
  // clearing a date field genuinely means "no bound", which the API reads as
  // an open range.
  from?: string | undefined
  to?: string | undefined
}

function query(range: DateRange): string {
  const p = new URLSearchParams()
  if (range.from) p.set('from', range.from)
  if (range.to) p.set('to', range.to)
  const s = p.toString()
  return s ? `?${s}` : ''
}

// --- The API ----------------------------------------------------------------

export const api = {
  login: (username: string, password: string) =>
    request<User>('/login', { method: 'POST', body: JSON.stringify({ username, password }) }),

  logout: () => request<{ status: string }>('/logout', { method: 'POST', body: '{}' }),

  me: () => request<User>('/me'),

  flights: (range: DateRange = {}) =>
    request<{ flights: Flight[]; count: number }>(`/flights${query(range)}`),

  /** One flight by its number. The edit page is reachable by URL. */
  flight: (seq: number) => request<{ flight: Flight }>(`/flights/${seq}`),

  createFlight: (draft: FlightDraft) =>
    request<{ flight: Flight }>('/flights', { method: 'POST', body: JSON.stringify(draft) }),

  /**
   * Corrects a flight entered in the app. A full replacement, not a patch:
   * the form holds every field, and merging "whatever happened to be sent"
   * into a legal record is a difference nobody can see in a diff.
   *
   * The server refuses (403) for any flight transcribed from the paper books.
   */
  updateFlight: (seq: number, draft: FlightDraft) =>
    request<{ flight: Flight }>(`/flights/${seq}`, { method: 'PUT', body: JSON.stringify(draft) }),

  /** Deletes a flight entered in the app, returning what was removed. */
  deleteFlight: (seq: number) =>
    request<{ flight: Flight }>(`/flights/${seq}`, { method: 'DELETE' }),

  stats: (range: DateRange = {}) =>
    request<{ summary: Summary; range: { from: string; to: string } }>(`/stats${query(range)}`),

  aircraft: () => request<{ aircraft: Aircraft[] }>('/aircraft'),

  /**
   * What each aeroplane cost over a range.
   *
   * `reg` adds the flights behind that one aeroplane's figure, so a disputed
   * invoice line can be traced to a flight rather than argued against a single
   * number. The summary rows always cover the whole range either way — asking
   * about one aeroplane must not narrow what it is compared against.
   */
  aircraftTime: (range: DateRange = {}, reg?: string) => {
    const p = new URLSearchParams()
    if (range.from) p.set('from', range.from)
    if (range.to) p.set('to', range.to)
    if (reg) p.set('reg', reg)
    const s = p.toString()
    return request<{
      range: { from: string; to: string }
      reg: string
      aircraft: AircraftTime[]
      total: AircraftTime
      flights: Flight[]
    }>(`/aircraft-time${s ? `?${s}` : ''}`)
  },

  discrepancies: () =>
    request<{ discrepancies: Discrepancy[]; count: number }>('/discrepancies'),

  sessions: () => request<{ sessions: SessionInfo[] }>('/sessions'),

  revokeSession: (id: number) =>
    request<{ status: string }>(`/sessions/${id}`, { method: 'DELETE' }),

  /**
   * exportURL is a plain link rather than a fetch. The browser downloads it
   * with the session cookie attached, which is what Content-Disposition needs
   * to produce a real "save file" -- turning a PDF into a blob in memory just
   * to re-offer it would cost nothing but complexity.
   */
  exportURL: (kind: 'easa' | 'table' | 'statistics', range: DateRange = {}) =>
    `${API_BASE}/export/${kind}.pdf${kind === 'easa' ? '' : query(range)}`,
}

/** FlightDraft is the new-flight form's payload; the server validates it. */
export interface FlightDraft {
  date: string
  aircraft_reg: string
  aircraft_type: string
  class: string
  dep_place: string
  arr_place: string
  off_block: string
  on_block: string
  /** Airborne times. Optional and usually blank -- but if one is sent, both must be. */
  takeoff: string
  landing: string
  total_time: string
  night_time: string
  instrument_time: string
  pic_time: string
  dual_time: string
  instructor_time: string
  pic_name: string
  landings_day: number
  landings_night: number
  remarks: string
}
