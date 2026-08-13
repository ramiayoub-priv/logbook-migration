import { useCallback, useEffect, useState } from 'react'

// A ~40-line router instead of a dependency.
//
// CLAUDE.md rule 0.3 asks for the dependency tree to stay near-empty and for
// every addition to be justified. This app has six pages, no nested routes, no
// route parameters and no data loaders, so a routing library would be a
// supply-chain decision bought for a switch statement. Apache rewrites unknown
// paths under /logbook to index.html (docs/deploy.md), which is what makes a
// real URL work on a reload rather than a hash.

/** BASE is where the app is mounted, matching Vite's base and the Apache Alias. */
export const BASE = '/logbook'

export type Route =
  | 'table'
  | 'statistics'
  | 'aircraft'
  | 'fleet'
  | 'new'
  | 'export'
  | 'review'
  | 'sessions'
  | 'edit'

const ROUTES: Record<string, Route> = {
  '': 'table',
  '/': 'table',
  '/table': 'table',
  '/statistics': 'statistics',
  '/aircraft': 'aircraft',
  '/fleet': 'fleet',
  '/new': 'new',
  '/export': 'export',
  '/review': 'review',
  '/sessions': 'sessions',
}

/**
 * EDIT is the one route carrying a parameter: /logbook/edit/1000123.
 *
 * A real path rather than a query string, and a real URL rather than a piece
 * of component state, because editing a flight should be linkable, bookmarkable
 * and survive a reload -- which on a phone happens every time the app is
 * swapped out for long enough. Apache rewrites unknown paths to index.html
 * (docs/deploy.md), so the deep link works on a cold load.
 *
 * One regex is still a long way short of needing a routing library (rule 0.3).
 */
const EDIT = /^\/edit\/(\d+)$/

/** pathOf strips the mount point off a browser path. */
function pathOf(pathname: string): string {
  return pathname.startsWith(BASE) ? pathname.slice(BASE.length) : pathname
}

export function routeOf(pathname: string): Route {
  const path = pathOf(pathname).replace(/\/$/, '')
  if (EDIT.test(path)) return 'edit'
  return ROUTES[path] ?? 'table'
}

/**
 * editSeqOf reads the flight number out of an edit URL, or null.
 *
 * Read at render time rather than held in state: useRoute already re-renders
 * on every navigation, so a second source of truth for "which flight" could
 * only ever be a way for the two to disagree.
 */
export function editSeqOf(pathname: string): number | null {
  const m = EDIT.exec(pathOf(pathname).replace(/\/$/, ''))
  return m ? Number(m[1]) : null
}

export function hrefFor(route: Route, seq?: number): string {
  if (route === 'edit') return `${BASE}/edit/${seq ?? ''}`
  return route === 'table' ? `${BASE}/` : `${BASE}/${route}`
}

export function navigate(route: Route, seq?: number): void {
  window.history.pushState({}, '', hrefFor(route, seq))
  window.dispatchEvent(new PopStateEvent('popstate'))
}

/** useRoute reports the current route and re-renders on back/forward. */
export function useRoute(): Route {
  const [route, setRoute] = useState<Route>(() => routeOf(window.location.pathname))

  useEffect(() => {
    const onPop = () => setRoute(routeOf(window.location.pathname))
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])

  return route
}

/**
 * Link is an ordinary anchor that navigates without a reload.
 *
 * It stays an <a href> so the link is real: middle-click, long-press and
 * "open in new tab" all work, and it is reachable and announced by a screen
 * reader. Only a plain left click is intercepted.
 */
export function Link({
  to,
  seq,
  children,
  className,
  current,
}: {
  to: Route
  /** The flight number, for the one route that takes one. */
  seq?: number
  children: React.ReactNode
  className?: string
  /** Marks the link as the page being viewed, for styling and screen readers. */
  current?: boolean
}) {
  const onClick = useCallback(
    (e: React.MouseEvent<HTMLAnchorElement>) => {
      if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) {
        return
      }
      e.preventDefault()
      navigate(to, seq)
    },
    [to, seq],
  )
  return (
    <a
      href={hrefFor(to, seq)}
      onClick={onClick}
      className={className}
      aria-current={current ? 'page' : undefined}
    >
      {children}
    </a>
  )
}
