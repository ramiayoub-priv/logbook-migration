import { useAuth } from './auth'
import { editSeqOf, Link, useRoute, type Route } from './router'
import { LoginPage } from './pages/Login'
import { TablePage } from './pages/Table'
import { StatisticsPage } from './pages/Statistics'
import { AircraftTimePage } from './pages/AircraftTime'
import { NewFlightPage } from './pages/NewFlight'
import { EditFlightPage } from './pages/EditFlight'
import { ExportPage } from './pages/Export'
import { ReviewPage } from './pages/Review'
import { SessionsPage } from './pages/Sessions'

// Six tabs on a 390px phone, which is why "Statistics" became "Stats" when the
// aircraft-time page arrived (2026-08-02). The alternative was hiding a page
// the owner asked for behind another one; a shorter label costs nothing and
// every pilot reads it.
const TABS: { route: Route; label: string }[] = [
  { route: 'table', label: 'Flights' },
  { route: 'statistics', label: 'Stats' },
  { route: 'aircraft', label: 'Aircraft' },
  { route: 'new', label: 'New' },
  { route: 'export', label: 'Export' },
  { route: 'review', label: 'Review' },
]

export function App() {
  const { state } = useAuth()
  const route = useRoute()

  // Nothing is rendered until the server has answered "who am I". Guessing and
  // correcting would flash the logbook at someone whose session has ended.
  if (state.status === 'checking') {
    return <div className="center">Checking your session…</div>
  }
  if (state.status === 'anonymous') {
    return <LoginPage />
  }

  return (
    <div className="app">
      <header className="topbar">
        <h1>Logbook</h1>
        <span className="who">
          {state.user.username} · <Link to="sessions">devices</Link>
        </span>
      </header>

      <nav className="tabs" aria-label="Sections">
        {TABS.map((t) => (
          <Link key={t.route} to={t.route} current={route === t.route}>
            {t.label}
          </Link>
        ))}
      </nav>

      <main className="content">
        <Page route={route} />
      </main>
    </div>
  )
}

function Page({ route }: { route: Route }) {
  switch (route) {
    case 'statistics':
      return <StatisticsPage />
    case 'aircraft':
      return <AircraftTimePage />
    case 'new':
      return <NewFlightPage />
    case 'export':
      return <ExportPage />
    case 'review':
      return <ReviewPage />
    case 'sessions':
      return <SessionsPage />
    case 'edit': {
      // The one route with a parameter. Read from the URL at render time
      // rather than held in state: useRoute already re-renders on every
      // navigation, so a second source of truth for "which flight" could only
      // be a way for the two to disagree. A malformed one cannot reach here --
      // routeOf only answers 'edit' for a path ending in digits.
      const seq = editSeqOf(window.location.pathname)
      return seq === null ? <TablePage /> : <EditFlightPage seq={seq} />
    }
    case 'table':
      return <TablePage />
  }
}
