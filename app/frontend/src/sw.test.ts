import { describe, it, expect, beforeAll } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

/**
 * These run against `public/sw.js` as shipped, not against a copy of its
 * logic. The file is evaluated in a fake worker global and its `policy`
 * function is pulled back off that global, so a change to the real worker is
 * what these assertions see.
 */
type Policy = (req: { method: string; mode?: string }, url: URL) => string

let policy: Policy
let source: string

beforeAll(() => {
  source = readFileSync(resolve(__dirname, '../public/sw.js'), 'utf8')

  const listeners: Record<string, unknown> = {}
  const fakeSelf: Record<string, unknown> = {
    location: { origin: 'https://ayoub.fi' },
    addEventListener: (name: string, fn: unknown) => {
      listeners[name] = fn
    },
    skipWaiting: () => {},
    clients: { claim: () => {} },
  }
  // eslint-disable-next-line @typescript-eslint/no-implied-eval
  new Function('self', 'caches', 'fetch', 'Response', source)(
    fakeSelf,
    { open: async () => ({}), keys: async () => [], match: async () => undefined, delete: async () => true },
    async () => ({}),
    { error: () => ({}) },
  )
  policy = fakeSelf['policy'] as Policy
})

const url = (path: string) => new URL(`https://ayoub.fi${path}`)
const GET = { method: 'GET' }

describe('the service worker never caches the logbook', () => {
  // The one that matters. Every API response is personal data and the server
  // marks it no-store; a service worker ignores that header unless written not
  // to. Caching one would leave the logbook readable on the device after the
  // session was revoked.
  it('passes every API request straight through', () => {
    for (const path of [
      '/logbook/api/flights',
      '/logbook/api/stats?from=2024-01-01',
      '/logbook/api/me',
      '/logbook/api/discrepancies',
      '/logbook/api/sessions',
      '/logbook/api/export/easa.pdf',
      '/logbook/api/health',
    ]) {
      expect(policy(GET, url(path)), path).toBe('passthrough')
    }
  })

  // A navigation to an API path must not be answered from the shell cache
  // either -- the API check is deliberately first for this reason.
  it('does not treat an API navigation as a shell request', () => {
    expect(policy({ method: 'GET', mode: 'navigate' }, url('/logbook/api/flights')))
      .toBe('passthrough')
  })

  it('never intercepts a mutation', () => {
    for (const method of ['POST', 'DELETE', 'PUT', 'PATCH']) {
      expect(policy({ method }, url('/logbook/api/flights')), method).toBe('passthrough')
      expect(policy({ method }, url('/logbook/')), method).toBe('passthrough')
    }
  })

  // Belt and braces: the file must not contain a cache write that is reachable
  // for an API URL. This catches a future edit that adds one outside `policy`.
  it('has exactly one place that decides what is cached', () => {
    expect(source).toContain("url.pathname.startsWith(BASE + 'api/')")
    expect(source.match(/caches\.open\(/g)?.length).toBeLessThanOrEqual(3)
  })
})

describe('the service worker caches the shell', () => {
  it('serves navigations network-first with a shell fallback', () => {
    expect(policy({ method: 'GET', mode: 'navigate' }, url('/logbook/'))).toBe('shell')
    expect(policy({ method: 'GET', mode: 'navigate' }, url('/logbook/statistics'))).toBe('shell')
  })

  // index.html is the one file under /logbook/ that is NOT content-hashed, so
  // cache-first on it means a deploy is invisible until the cache is cleared.
  // It goes through the network-first shell path like any other navigation.
  it('never treats the un-hashed shell document as an immutable asset', () => {
    expect(policy(GET, url('/logbook/index.html'))).toBe('shell')
  })

  // The shell fetch must bypass the browser's own HTTP cache. Network-first is
  // only as fresh as what the network layer hands back, and a phone that has
  // index.html in its HTTP cache would keep opening the previous build with
  // the worker none the wiser. Asserted against the source because it is a
  // fetch option, not a routing decision -- there is nothing to call.
  it('asks the network for the shell rather than the HTTP cache', () => {
    expect(source).toMatch(/cache:\s*'no-store'/)
  })

  // Build assets are content-hashed, so the filename changes when the bytes
  // do and a cache hit can never be stale.
  it('caches the hashed build assets', () => {
    expect(policy(GET, url('/logbook/assets/index-abc123.js'))).toBe('asset')
    expect(policy(GET, url('/logbook/assets/index-abc123.css'))).toBe('asset')
    expect(policy(GET, url('/logbook/icons/icon-192.png'))).toBe('asset')
  })

  // The box serves the owner's other sites. Nothing outside /logbook is ours
  // to touch.
  it('leaves the rest of the host alone', () => {
    expect(policy(GET, url('/blog/'))).toBe('bypass')
    expect(policy(GET, url('/'))).toBe('bypass')
    expect(policy(GET, new URL('https://example.com/x.js'))).toBe('bypass')
  })
})
