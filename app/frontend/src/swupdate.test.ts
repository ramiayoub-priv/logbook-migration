import { describe, it, expect, vi } from 'vitest'
import { reloadWhenUpdated } from './swupdate'

/**
 * A fake ServiceWorkerContainer: just enough of one to say whether this page
 * is already controlled, and to fire the event that says a new worker has
 * taken over.
 */
function container(controlled: boolean) {
  const listeners: Array<() => void> = []
  return {
    controller: controlled ? {} : null,
    addEventListener: (name: string, fn: () => void) => {
      if (name === 'controllerchange') listeners.push(fn)
    },
    fire: () => listeners.forEach((fn) => fn()),
    listenerCount: () => listeners.length,
  }
}

describe('reloadWhenUpdated', () => {
  // The reason this exists: the app is installed to a phone home screen, where
  // there is no address bar and no reload button. Without this, a deploy is
  // invisible until the operating system decides to kill the tab.
  it('reloads the page when a new worker takes over', () => {
    const reload = vi.fn()
    const c = container(true)
    reloadWhenUpdated(c as never, reload)
    c.fire()
    expect(reload).toHaveBeenCalledTimes(1)
  })

  // The first registration on a device also fires controllerchange -- there
  // was no controller and now there is one. Reloading there would reload every
  // first-ever visit for nothing.
  it('does nothing on the first registration, when there was no controller', () => {
    const reload = vi.fn()
    const c = container(false)
    reloadWhenUpdated(c as never, reload)
    c.fire()
    expect(reload).not.toHaveBeenCalled()
  })

  // A reload loop on a phone at an airfield would make the app unusable
  // exactly where it is needed, so the reload happens at most once per page.
  it('reloads at most once', () => {
    const reload = vi.fn()
    const c = container(true)
    reloadWhenUpdated(c as never, reload)
    c.fire()
    c.fire()
    expect(reload).toHaveBeenCalledTimes(1)
  })
})
