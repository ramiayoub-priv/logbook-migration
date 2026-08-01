import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AuthProvider } from '../auth'
import { App } from '../App'
import { api, ApiError, type Flight, type Summary } from '../api'

const USER = { user_id: 1, username: 'rami' }

function flight(over: Partial<Flight> = {}): Flight {
  return {
    seq: 1, date: '2021-06-01',
    aircraft_type: 'C172', aircraft_reg: 'OH-CTL', class: 'SEP_SEA',
    dep_place: 'EFHF', arr_place: 'EFHF',
    off_block_utc: '2021-06-01T15:13:00Z', on_block_utc: '2021-06-01T16:34:00Z',
    off_block_raw: '18:13', on_block_raw: '19:34', time_origin: 'converted_from_local',
    block_minutes: 81, total_minutes: 81, night_minutes: 0, instrument_minutes: 0,
    pic_minutes: 81, dual_minutes: 0, instructor_minutes: 81,
    pic_name: 'self', landings_day: 7, landings_night: 0, landings_verified: true,
    remarks: '', source_book: 3, source_row: 3,
    ...over,
  }
}

function summary(over: Partial<Summary> = {}): Summary {
  return {
    flights: 1293, total: 73175, pic: 63183, dual: 9992, night: 1365,
    instrument: 6425, instructor: 11381,
    sea_total: 24459, sea_pic: 22652, sea_instructor: 9042,
    land_total: 48716, land_pic: 40531, land_instructor: 2339,
    landings_day: 3439, landings_night: 0, landings_sea: 1693, landings_land: 1746,
    landings_unverified: 30,
    ...over,
  }
}

beforeEach(() => {
  vi.spyOn(api, 'me').mockResolvedValue(USER)
  vi.spyOn(api, 'flights').mockResolvedValue({ flights: [flight()], count: 1 })
  vi.spyOn(api, 'stats').mockResolvedValue({ summary: summary(), range: { from: '', to: '' } })
  vi.spyOn(api, 'aircraft').mockResolvedValue({
    aircraft: [
      { registration: 'OH-CTL', type: 'C172', default_class: 'SEP_SEA', ifr_capable: false, active: true, notes: '' },
      { registration: 'OH-CAM', type: 'C172', default_class: 'SEP_LAND', ifr_capable: true, active: true, notes: '' },
    ],
  })
  vi.spyOn(api, 'discrepancies').mockResolvedValue({ discrepancies: [], count: 0 })
  vi.spyOn(api, 'sessions').mockResolvedValue({ sessions: [] })
  window.history.pushState({}, '', '/logbook/')
})

afterEach(() => vi.restoreAllMocks())

function renderApp() {
  return render(
    <AuthProvider>
      <App />
    </AuthProvider>,
  )
}

describe('signing in', () => {
  // Nothing may be shown until the server has said who this is. A cached
  // "signed in" guess would flash the logbook at somebody whose session ended.
  it('shows the login page when there is no session', async () => {
    vi.spyOn(api, 'me').mockRejectedValue(new ApiError(401, 'authentication required'))
    renderApp()
    expect(await screen.findByLabelText('Username')).toBeInTheDocument()
    expect(screen.queryByRole('navigation')).not.toBeInTheDocument()
  })

  // The server answers a wrong username and a wrong password identically and
  // in the same time. The page must not undo that by being more specific.
  it('gives one uninformative message for any bad credentials', async () => {
    vi.spyOn(api, 'me').mockRejectedValue(new ApiError(401, 'authentication required'))
    vi.spyOn(api, 'login').mockRejectedValue(new ApiError(401, 'invalid username or password'))
    const user = userEvent.setup()
    renderApp()

    await user.type(await screen.findByLabelText('Username'), 'nosuchuser')
    await user.type(screen.getByLabelText('Password'), 'wrong')
    await user.click(screen.getByRole('button', { name: 'Sign in' }))

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent('Wrong username or password.')
    expect(alert.textContent).not.toMatch(/no such|unknown user|does not exist/i)
  })

  it('tells the pilot how long a throttled login has to wait', async () => {
    vi.spyOn(api, 'me').mockRejectedValue(new ApiError(401, 'authentication required'))
    vi.spyOn(api, 'login').mockRejectedValue(new ApiError(429, 'too many attempts', [], 45))
    const user = userEvent.setup()
    renderApp()

    await user.type(await screen.findByLabelText('Username'), 'rami')
    await user.type(screen.getByLabelText('Password'), 'wrong')
    await user.click(screen.getByRole('button', { name: 'Sign in' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('45 seconds')
  })

  it('shows the logbook once signed in', async () => {
    renderApp()
    expect(await screen.findByRole('navigation', { name: 'Sections' })).toBeInTheDocument()
    // The username shares its element with the devices link.
    expect(screen.getByText(/rami/)).toBeInTheDocument()
  })
})

describe('the flights table', () => {
  it('renders durations as H:MM rather than minutes', async () => {
    renderApp()
    // 81 minutes appears in the total, PIC and instructor columns alike, so
    // all three are expected -- what matters is that none of them says "81".
    expect(await screen.findAllByText('1:21')).toHaveLength(3)
    expect(screen.queryByText('81')).not.toBeInTheDocument()
  })

  // A row whose day/night split was inferred must not look like one somebody
  // checked (rule 0.2, Task 8).
  it('marks a row whose landing split was inferred', async () => {
    vi.spyOn(api, 'flights').mockResolvedValue({
      flights: [flight({ landings_verified: false, landings_night: 2 })],
      count: 1,
    })
    renderApp()
    const flag = await screen.findByTitle(/inferred, not read from the paper page/i)
    expect(flag).toBeInTheDocument()
  })

  it('does not mark a verified row', async () => {
    renderApp()
    await screen.findAllByText('1:21')
    expect(screen.queryByTitle(/inferred, not read/i)).not.toBeInTheDocument()
  })

  // A failed read must never render as "no flights". That is the silent
  // corruption rule 0.2 forbids.
  it('reports a failure instead of showing an empty logbook', async () => {
    vi.spyOn(api, 'flights').mockRejectedValue(new ApiError(500, 'could not read the logbook'))
    renderApp()
    expect(await screen.findByRole('alert')).toHaveTextContent('could not read the logbook')
    expect(screen.queryByText('No flights in this range.')).not.toBeInTheDocument()
  })

  it('says so plainly when the range really is empty', async () => {
    vi.spyOn(api, 'flights').mockResolvedValue({ flights: [], count: 0 })
    renderApp()
    expect(await screen.findByText('No flights in this range.')).toBeInTheDocument()
  })

  it('asks the server again when the range changes', async () => {
    const user = userEvent.setup()
    renderApp()
    await screen.findAllByText('1:21')

    await user.type(screen.getByLabelText('From'), '2021-01-01')
    await waitFor(() =>
      expect(api.flights).toHaveBeenCalledWith(expect.objectContaining({ from: '2021-01-01' })),
    )
  })
})

describe('the statistics page', () => {
  it('shows the twelve figures as H:MM', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'Statistics' }))

    // 73175 minutes is the frozen whole-logbook total of 1219:35.
    expect(await screen.findByText('1219:35')).toBeInTheDocument()
    expect(screen.getByText('1053:03')).toBeInTheDocument() // PIC
    expect(screen.getByText('407:39')).toBeInTheDocument()  // seaplane
  })

  // The page must say that the night landing figure is partly inferred rather
  // than presenting it as verified.
  it('discloses the unverified landing split', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'Statistics' }))
    expect(await screen.findByText(/30 flights in this range carry a day\/night landing split/i))
      .toBeInTheDocument()
  })

  it('omits the caveat when nothing in range is unverified', async () => {
    vi.spyOn(api, 'stats').mockResolvedValue({
      summary: summary({ landings_unverified: 0 }),
      range: { from: '', to: '' },
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'Statistics' }))
    await screen.findByText('1219:35')
    expect(screen.queryByText(/landing split that was inferred/i)).not.toBeInTheDocument()
  })
})

describe('the new-flight form', () => {
  it('fills the type and class from the chosen aircraft', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    await user.selectOptions(await screen.findByLabelText('Aircraft'), 'OH-CTL')
    expect(screen.getByLabelText('Type')).toHaveValue('C172')
    // OH-CTL is on floats, so the class defaults to sea -- which is what
    // decides whether the flight counts towards the seaplane rating.
    expect(screen.getByLabelText('Class')).toHaveValue('SEP_SEA')
  })

  // The default is a default, not a constraint: the same registration can be
  // flown on wheels.
  it('lets the class be overridden after a preselect', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.selectOptions(await screen.findByLabelText('Aircraft'), 'OH-CTL')
    await user.selectOptions(screen.getByLabelText('Class'), 'SEP_LAND')
    expect(screen.getByLabelText('Class')).toHaveValue('SEP_LAND')
  })

  it('puts the server field errors on the controls they name', async () => {
    vi.spyOn(api, 'createFlight').mockRejectedValue(
      new ApiError(400, 'this flight cannot be logged as written', [
        { field: 'total_time', message: 'a total time is required' },
        { field: 'off_block', message: 'an off-block time is required' },
      ]),
    )
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(await screen.findByRole('button', { name: 'Log this flight' }))

    expect(await screen.findByText('a total time is required')).toBeInTheDocument()
    expect(screen.getByText('an off-block time is required')).toBeInTheDocument()
    expect(screen.getByLabelText('Total time')).toHaveAttribute('aria-invalid', 'true')
  })

  it('reports a duplicate rather than pretending the flight was logged', async () => {
    vi.spyOn(api, 'createFlight').mockRejectedValue(
      new ApiError(409, 'this flight is already in the logbook -- same date, aircraft and off-block time'),
    )
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(await screen.findByRole('button', { name: 'Log this flight' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('already in the logbook')
  })

  it('confirms a saved flight with what was actually stored', async () => {
    vi.spyOn(api, 'createFlight').mockResolvedValue({
      flight: flight({ date: '2026-07-30', aircraft_reg: 'OH-CAM', total_minutes: 75 }),
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(await screen.findByRole('button', { name: 'Log this flight' }))

    expect(await screen.findByRole('status')).toHaveTextContent('Logged 2026-07-30, OH-CAM, 1:15.')
  })

  // The date input is capped at today because a flight that has not happened
  // cannot be logged; the server refuses it too.
  it('will not offer a future date', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    const date = await screen.findByLabelText('Date')
    expect(date).toHaveAttribute('max', new Date().toISOString().slice(0, 10))
  })
})

describe('the export page', () => {
  it('links to all three documents', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'Export' }))

    expect(await screen.findByRole('button', { name: /EASA logbook/i }))
      .toHaveAttribute('href', '/logbook/api/export/easa.pdf')
    expect(screen.getByRole('link', { name: /flight table/i }))
      .toHaveAttribute('href', '/logbook/api/export/table.pdf')
    expect(screen.getByRole('link', { name: /statistics sheet/i }))
      .toHaveAttribute('href', '/logbook/api/export/statistics.pdf')
  })

  // A partial EASA logbook would understate a licence total, so the range
  // must not reach it however the picker is set.
  it('keeps the range off the EASA link but puts it on the others', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'Export' }))
    await user.type(await screen.findByLabelText('From'), '2024-01-01')

    await waitFor(() =>
      expect(screen.getByRole('link', { name: /flight table/i }))
        .toHaveAttribute('href', '/logbook/api/export/table.pdf?from=2024-01-01'),
    )
    expect(screen.getByRole('button', { name: /EASA logbook/i }))
      .toHaveAttribute('href', '/logbook/api/export/easa.pdf')
  })
})

describe('session handling', () => {
  // A session can end at any moment: it expires, it is revoked from another
  // device, or the account is disabled. Whatever the page was doing, the app
  // returns to the login screen rather than showing an error it cannot fix.
  it('falls back to the login page when a request comes back 401', async () => {
    vi.spyOn(api, 'flights').mockRejectedValue(new ApiError(401, 'authentication required'))
    renderApp()
    expect(await screen.findByLabelText('Username')).toBeInTheDocument()
  })
})
