import { describe, it, expect, vi, afterEach } from 'vitest'
import { api, ApiError } from './api'

function mockFetch(res: Partial<Response> & { jsonBody?: unknown }) {
  const body = res.jsonBody === undefined ? '' : JSON.stringify(res.jsonBody)
  const impl = vi.fn(async () => ({
    ok: res.ok ?? true,
    status: res.status ?? 200,
    headers: res.headers ?? new Headers(),
    text: async () => body,
  })) as unknown as typeof fetch
  vi.stubGlobal('fetch', impl)
  return impl as unknown as ReturnType<typeof vi.fn>
}

afterEach(() => vi.unstubAllGlobals())

describe('the fetch layer', () => {
  // The whole authentication scheme. The cookie is HttpOnly so nothing in the
  // page can attach it by hand; if this option is ever dropped, every private
  // endpoint starts answering 401 and the app looks logged out forever.
  it('sends credentials on every request', async () => {
    const f = mockFetch({ jsonBody: { flights: [], count: 0 } })
    await api.flights()
    expect(f).toHaveBeenCalledWith(
      '/logbook/api/flights',
      expect.objectContaining({ credentials: 'same-origin' }),
    )
  })

  // JavaScript must never read or write the session cookie: it is HttpOnly
  // precisely so a script cannot. Anything setting an Authorization header
  // would mean a token had been copied somewhere readable.
  it('sends no Authorization header and touches no token', async () => {
    const f = mockFetch({ jsonBody: {} })
    await api.me()
    const init = f.mock.calls[0]![1] as RequestInit
    const headers = (init.headers ?? {}) as Record<string, string>
    expect(Object.keys(headers).map((k) => k.toLowerCase())).not.toContain('authorization')
  })

  // application/json is a CSRF control, not just a parsing convenience: a
  // cross-origin HTML form cannot send it.
  it('marks a mutation as JSON', async () => {
    const f = mockFetch({ status: 201, jsonBody: { flight: {} } })
    await api.createFlight({} as never)
    const init = f.mock.calls[0]![1] as RequestInit
    expect(init.method).toBe('POST')
    expect((init.headers as Record<string, string>)['Content-Type']).toBe('application/json')
  })

  it('turns a failure into an ApiError carrying the status', async () => {
    mockFetch({ ok: false, status: 409, jsonBody: { error: 'already in the logbook' } })
    await expect(api.createFlight({} as never)).rejects.toMatchObject({
      status: 409,
      message: 'already in the logbook',
    })
  })

  // The form needs to know which control to highlight, not just that something
  // was wrong.
  it('keeps the field errors from a rejected flight', async () => {
    mockFetch({
      ok: false,
      status: 400,
      jsonBody: {
        error: 'this flight cannot be logged as written',
        fields: [{ field: 'date', message: 'a date is required' }],
      },
    })
    try {
      await api.createFlight({} as never)
      expect.unreachable('should have thrown')
    } catch (e) {
      const err = e as ApiError
      expect(err.fields).toHaveLength(1)
      expect(err.fields[0]?.field).toBe('date')
    }
  })

  it('reads Retry-After off a throttled login', async () => {
    mockFetch({
      ok: false,
      status: 429,
      headers: new Headers({ 'Retry-After': '45' }),
      jsonBody: { error: 'too many attempts' },
    })
    try {
      await api.login('rami', 'wrong')
      expect.unreachable('should have thrown')
    } catch (e) {
      expect((e as ApiError).retryAfter).toBe(45)
    }
  })

  it('flags a 401 so the app can fall back to the login page', async () => {
    mockFetch({ ok: false, status: 401, jsonBody: { error: 'authentication required' } })
    try {
      await api.stats()
      expect.unreachable('should have thrown')
    } catch (e) {
      expect((e as ApiError).isUnauthenticated).toBe(true)
    }
  })

  // A dropped connection must not read as an empty logbook. "You have flown
  // nothing" is the silent corruption rule 0.2 forbids.
  it('turns a network failure into an error, never an empty result', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => { throw new TypeError('Failed to fetch') }))
    await expect(api.flights()).rejects.toBeInstanceOf(ApiError)
  })

  it('puts the date range on the query string', async () => {
    const f = mockFetch({ jsonBody: { summary: {}, range: {} } })
    await api.stats({ from: '2024-01-01', to: '2024-12-31' })
    expect(f.mock.calls[0]![0]).toBe('/logbook/api/stats?from=2024-01-01&to=2024-12-31')
  })

  it('omits an unset bound rather than sending an empty one', async () => {
    const f = mockFetch({ jsonBody: { flights: [], count: 0 } })
    await api.flights({ from: '2024-01-01' })
    expect(f.mock.calls[0]![0]).toBe('/logbook/api/flights?from=2024-01-01')
  })
})

describe('export links', () => {
  it('carries the range on the table and statistics exports', () => {
    expect(api.exportURL('table', { from: '2024-01-01' }))
      .toBe('/logbook/api/export/table.pdf?from=2024-01-01')
    expect(api.exportURL('statistics', { to: '2024-12-31' }))
      .toBe('/logbook/api/export/statistics.pdf?to=2024-12-31')
  })

  // The EASA document is the complete record an authority reads. A range on it
  // would produce a partial logbook that understates a licence total.
  it('never puts a range on the EASA export', () => {
    expect(api.exportURL('easa', { from: '2024-01-01', to: '2024-12-31' }))
      .toBe('/logbook/api/export/easa.pdf')
  })
})
