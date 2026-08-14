import { describe, it, expect, beforeAll } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

/**
 * `public/sw.js` is now a KILL SWITCH, and these run against the file as
 * shipped rather than a copy of its logic.
 *
 * It used to cache the app shell so the logbook would open at an airfield with
 * no signal. The owner ended that on 2026-08-14: *"Can you make sure NOTHING is
 * cached at all? Like the browser needs to forget (except the cookie for the
 * session)."* The shell cache had twice put a stale build in front of them, and
 * offline writes were never in scope — so an offline shell could only ever open
 * an app that then failed every request.
 *
 * WHAT IS LEFT MUST STILL BE SHIPPED. Deleting sw.js from the server does not
 * remove a worker already installed on a device; the browser goes on using the
 * one it has. The only thing that reliably retires it is a NEW worker at the
 * same URL that deletes every cache and unregisters itself. That is this file,
 * and it must stay deployed for as long as any device might still be carrying
 * the old one.
 */
type Handler = (event: { waitUntil: (p: Promise<unknown>) => void }) => void

let source: string
let listeners: Record<string, Handler>
let deleted: string[]
let unregistered: boolean
let navigated: string[]

async function runWorker() {
  listeners = {}
  deleted = []
  unregistered = false
  navigated = []

  const fakeSelf: Record<string, unknown> = {
    location: { origin: 'https://ayoub.fi' },
    addEventListener: (name: string, fn: Handler) => {
      listeners[name] = fn
    },
    skipWaiting: () => {},
    registration: {
      unregister: async () => {
        unregistered = true
        return true
      },
    },
    clients: {
      claim: async () => {},
      matchAll: async () => [
        { url: 'https://ayoub.fi/logbook/', navigate: async (u: string) => void navigated.push(u) },
      ],
    },
  }
  const fakeCaches = {
    keys: async () => ['logbook-shell-v1', 'logbook-shell-v2'],
    delete: async (k: string) => {
      deleted.push(k)
      return true
    },
  }
  // eslint-disable-next-line @typescript-eslint/no-implied-eval
  new Function('self', 'caches', source)(fakeSelf, fakeCaches)

  // Drive activate the way the browser would, and wait for what it registered.
  const pending: Promise<unknown>[] = []
  listeners['activate']?.({ waitUntil: (p) => void pending.push(p) })
  await Promise.all(pending)
}

beforeAll(() => {
  source = readFileSync(resolve(__dirname, '../public/sw.js'), 'utf8')
})

describe('the service worker is a kill switch', () => {
  // The whole point. Nothing may be served from a cache this worker controls,
  // so it must not answer fetches at all -- an intercepted request is a request
  // that could be answered from storage.
  it('installs no fetch handler, so every request goes to the network', async () => {
    await runWorker()
    expect(listeners['fetch']).toBeUndefined()
  })

  it('deletes every cache, not just the ones it recognises', async () => {
    await runWorker()
    expect(deleted).toEqual(['logbook-shell-v1', 'logbook-shell-v2'])
  })

  it('unregisters itself, so the next load has no worker at all', async () => {
    await runWorker()
    expect(unregistered).toBe(true)
  })

  it('reloads the open page onto the network once it has cleaned up', async () => {
    await runWorker()
    expect(navigated).toEqual(['https://ayoub.fi/logbook/'])
  })

  // A worker that waited its turn would leave the old one serving the old
  // shell until every tab was closed -- on a home-screen app, possibly never.
  it('takes over immediately rather than waiting for the old worker to die', async () => {
    await runWorker()
    expect(source).toContain('skipWaiting')
  })

  // The guarantee that outlives this file: it must never grow a cache again.
  it('opens no cache and stores nothing', () => {
    expect(source).not.toContain('caches.open')
    expect(source).not.toContain('cache.put')
    expect(source).not.toContain('respondWith')
  })
})
