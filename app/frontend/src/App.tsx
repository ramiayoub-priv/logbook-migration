import { useAuth } from './auth'
import { Link, useRoute, type Route } from './router'
import { LoginPage } from './pages/Login'
import { TablePage } from './pages/Table'
import { StatisticsPage } from './pages/Statistics'
import { NewFlightPage } from './pages/NewFlight'
import { ExportPage } from './pages/Export'
import { ReviewPage } from './pages/Review'
import { SessionsPage } from './pages/Sessions'

const TABS: { route: Route; label: string }[] = [
  { route: 'table', label: 'Flights' },
  { route: 'statistics', label: 'Statistics' },
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
    case 'new':
      return <NewFlightPage />
    case 'export':
      return <ExportPage />
    case 'review':
      return <ReviewPage />
    case 'sessions':
      return <SessionsPage />
    case 'table':
      return <TablePage />
  }
}
