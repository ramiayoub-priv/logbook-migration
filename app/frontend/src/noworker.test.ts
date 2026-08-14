import { describe, it, expect, beforeAll } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

/**
 * The app must not register a service worker, ever again.
 *
 * Owner instruction, 2026-08-14: nothing is to be cached. `public/sw.js` is now
 * a kill switch that deletes every cache and unregisters itself — but that only
 * cleans up the devices that already have a worker. If the app went on calling
 * `register()`, it would install the kill switch afresh on every load, which
 * unregisters, which navigates, which registers again: a reload loop on the
 * phone this was meant to fix.
 *
 * Asserted against the source of `main.tsx` rather than a mock, because the
 * thing being guarded is that a line does not exist.
 */
let main: string

beforeAll(() => {
  main = readFileSync(resolve(__dirname, 'main.tsx'), 'utf8')
})

describe('the app registers no service worker', () => {
  it('never calls serviceWorker.register', () => {
    expect(main).not.toMatch(/serviceWorker\s*\n?\s*\.?\s*register/)
    expect(main).not.toContain('.register(')
  })

  it('does not reach for navigator.serviceWorker at all', () => {
    expect(main).not.toContain('serviceWorker')
  })

  // sw.js itself must stay shipped: deleting it from the server would strand
  // every device that still has the OLD worker, which would go on serving its
  // cached shell forever. See the comment at the top of public/sw.js.
  it('still ships public/sw.js, which is what retires the old worker', () => {
    const sw = readFileSync(resolve(__dirname, '../public/sw.js'), 'utf8')
    expect(sw).toContain('unregister')
  })
})
