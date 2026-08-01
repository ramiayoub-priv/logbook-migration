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
  | 'new'
  | 'export'
  | 'review'
  | 'sessions'

const ROUTES: Record<string, Route> = {
  '': 'table',
  '/': 'table',
  '/table': 'table',
  '/statistics': 'statistics',
  '/new': 'new',
  '/export': 'export',
  '/review': 'review',
  '/sessions': 'sessions',
}

/** pathOf strips the mount point off a browser path. */
function pathOf(pathname: string): string {
  return pathname.startsWith(BASE) ? pathname.slice(BASE.length) : pathname
}

export function routeOf(pathname: string): Route {
  return ROUTES[pathOf(pathname).replace(/\/$/, '')] ?? 'table'
}

export function hrefFor(route: Route): string {
  return route === 'table' ? `${BASE}/` : `${BASE}/${route}`
}

export function navigate(route: Route): void {
  window.history.pushState({}, '', hrefFor(route))
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
  children,
  className,
  current,
}: {
  to: Route
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
      navigate(to)
    },
    [to],
  )
  return (
    <a
      href={hrefFor(to)}
      onClick={onClick}
      className={className}
      aria-current={current ? 'page' : undefined}
    >
      {children}
    </a>
  )
}
