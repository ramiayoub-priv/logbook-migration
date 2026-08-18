import { describe, it, expect, beforeAll } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

/**
 * A guard on the ONE CSS mistake this stylesheet has actually made.
 *
 * The aircraft picker and the pilot picker share a block of rules, and the
 * sharing was introduced (Task 21) by inserting `.pilotpicker` as a second
 * selector — but on several rules the split landed in the wrong place:
 *
 *     .aircraftpicker .options,          <- the LIST, styled as an option
 *     .pilotpicker .options button { display: grid; ... }
 *
 * so the aircraft options got no layout at all (registration, type and flight
 * count ran together as `OH-CTLC172287 flights`) while the dropdown itself
 * became a three-column grid, laying the list out in columns. Below 22rem the
 * same slip applied `display: none` to the whole aircraft list.
 *
 * Nothing in a jsdom test renders CSS, so this asserts on the stylesheet's
 * text, the way `noworker.test.ts` asserts on `main.tsx`: the thing being
 * guarded is a selector shape, and it is invisible to every other test.
 */
type Rule = { selectors: string[]; body: string }
let rules: Rule[]

beforeAll(() => {
  // Comments carry both braces and commas, so they go first.
  const css = readFileSync(resolve(__dirname, 'styles.css'), 'utf8').replace(
    /\/\*[\s\S]*?\*\//g,
    '',
  )
  // Innermost blocks only: an @media wrapper cannot match, because its body
  // contains the braces this pattern forbids.
  rules = [...css.matchAll(/([^{}]+)\{([^{}]*)\}/g)].map((m) => ({
    selectors: (m[1] ?? '').split(',').map((s) => s.trim()).filter(Boolean),
    body: m[2] ?? '',
  }))
})

const strip = (s: string) => s.replace(/^\.(aircraft|pilot)picker/, '').trim()

describe('the shared picker CSS', () => {
  it('styles the same thing in both pickers when a rule names both', () => {
    for (const rule of rules) {
      const air = rule.selectors.filter((s) => s.startsWith('.aircraftpicker'))
      const pilot = rule.selectors.filter((s) => s.startsWith('.pilotpicker'))
      if (air.length === 0 || pilot.length === 0) continue
      expect(air.map(strip).sort(), rule.selectors.join(', ')).toEqual(
        pilot.map(strip).sort(),
      )
    }
  })

  it('never lays out the dropdown itself, or hides it', () => {
    for (const rule of rules) {
      if (!rule.selectors.includes('.aircraftpicker .options')) continue
      expect(rule.body, rule.selectors.join(', ')).not.toMatch(/display:\s*(grid|none)/)
    }
  })

  it('lays out the option button in each picker', () => {
    for (const picker of ['.aircraftpicker', '.pilotpicker']) {
      const rule = rules.find((r) => r.selectors.includes(`${picker} .options button`))
      expect(rule?.body, `${picker} .options button`).toMatch(/display:\s*grid/)
    }
  })
})
