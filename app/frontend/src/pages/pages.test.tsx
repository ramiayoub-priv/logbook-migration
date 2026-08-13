import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AuthProvider } from '../auth'
import { App } from '../App'
import { api, ApiError, type AircraftTime, type Flight, type Summary } from '../api'

const USER = { user_id: 1, username: 'rami' }

function flight(over: Partial<Flight> = {}): Flight {
  return {
    seq: 1, date: '2021-06-01',
    aircraft_type: 'C172', aircraft_reg: 'OH-CTL', class: 'SEP_SEA',
    dep_place: 'EFHF', arr_place: 'EFHF',
    off_block_utc: '2021-06-01T15:13:00Z', on_block_utc: '2021-06-01T16:34:00Z',
    off_block_raw: '18:13', on_block_raw: '19:34', time_origin: 'converted_from_local',
    takeoff_utc: null, landing_utc: null,
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

function acTime(over: Partial<AircraftTime> = {}): AircraftTime {
  return {
    registration: 'OH-CAM', types: ['C172'], flights: 2,
    block_minutes: 135, air_minutes: 65, air_known: 1, air_missing: 1,
    block_differs_from_total: 0,
    ...over,
  }
}

beforeEach(() => {
  vi.spyOn(api, 'me').mockResolvedValue(USER)
  vi.spyOn(api, 'aircraftTime').mockResolvedValue({
    range: { from: '', to: '' },
    reg: '',
    aircraft: [acTime()],
    total: acTime({ registration: '', types: [] }),
    flights: [],
  })
  vi.spyOn(api, 'flights').mockResolvedValue({ flights: [flight()], count: 1 })
  vi.spyOn(api, 'stats').mockResolvedValue({ summary: summary(), range: { from: '', to: '' } })
  vi.spyOn(api, 'aircraft').mockResolvedValue({
    // In the order the server sends: never-flown first, then most recently
    // flown. The picker must not re-sort it.
    aircraft: [
      { registration: 'OH-PDP', type: 'P28A', default_class: 'SEP_LAND', ifr_capable: false,
        notes: '', user_added: true, last_flown: '', flights: 0 },
      { registration: 'OH-CTL', type: 'C172', default_class: 'SEP_SEA', ifr_capable: false,
        notes: '', user_added: false, last_flown: '2026-07-30', flights: 286 },
      { registration: 'OH-CAM', type: 'C172', default_class: 'SEP_LAND', ifr_capable: true,
        notes: '', user_added: false, last_flown: '2026-07-01', flights: 12 },
    ],
  })
  // The roster the server sends: never flown with first, then most recent.
  // `self` carries 1143 of the 1296 transcribed flights.
  vi.spyOn(api, 'pilots').mockResolvedValue({
    pilots: [
      { name: 'Jansson', user_added: true, last_flown: '', flights: 0 },
      { name: 'self', user_added: false, last_flown: '2026-07-30', flights: 1143 },
      { name: 'Martevuo', user_added: false, last_flown: '2019-05-05', flights: 54 },
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

  // Newest first. The API returns the book's own seq order -- oldest first --
  // because that is the order everything cumulative must use, and the EASA
  // export depends on it. The reversal is therefore a VIEW concern and lives
  // here, not in the query: the most recent flight is the one being looked for.
  it('lists the newest flight first', async () => {
    vi.spyOn(api, 'flights').mockResolvedValue({
      flights: [
        flight({ seq: 1, date: '2021-06-01' }),
        flight({ seq: 2, date: '2023-04-15' }),
        flight({ seq: 3, date: '2026-07-30' }),
      ],
      count: 3,
    })
    renderApp()

    const rows = await screen.findAllByRole('row')
    // rows[0] is the header.
    const dates = rows.slice(1).map((r) => r.querySelectorAll('td')[0]?.textContent)
    expect(dates).toEqual(['2026-07-30', '2023-04-15', '2021-06-01'])
  })

  // Ordering on seq rather than on the date string is load-bearing: 21 rows
  // across the three books are genuinely out of date order, and three of them
  // are the 28/08/2025 late entries sitting at the end of Book 3. A table
  // sorted by date would silently move them out of book order.
  it('reverses book order, not date order', async () => {
    vi.spyOn(api, 'flights').mockResolvedValue({
      flights: [
        flight({ seq: 1294, date: '2026-07-30' }),
        // A late entry: last in the book, but an older date.
        flight({ seq: 1295, date: '2025-08-28' }),
      ],
      count: 2,
    })
    renderApp()

    const rows = await screen.findAllByRole('row')
    const dates = rows.slice(1).map((r) => r.querySelectorAll('td')[0]?.textContent)
    // The late entry is newest in the book, so it comes first -- even though
    // its date is older.
    expect(dates).toEqual(['2025-08-28', '2026-07-30'])
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

  // --- Task 12: the airborne times belong in the table ---------------------
  //
  // The aircraft's own logbook -- a separate, legally required document the
  // owner fills after flying -- records AIRBORNE times, not block times.
  // Reading them off the app instead of off the paper is the whole point of
  // having the app in the field.

  it('shows takeoff, landing and a derived air time', async () => {
    vi.spyOn(api, 'flights').mockResolvedValue({
      flights: [
        flight({
          takeoff_utc: '2021-06-01T15:20:00Z',
          landing_utc: '2021-06-01T16:25:00Z',
        }),
      ],
      count: 1,
    })
    renderApp()

    expect(await screen.findByRole('columnheader', { name: 'Takeoff' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Landing' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Air' })).toBeInTheDocument()

    const cells = screen.getAllByRole('cell').map((c) => c.textContent)
    expect(cells).toContain('15:20')
    expect(cells).toContain('16:25')
    // 65 minutes, computed here from the two instants -- never stored.
    expect(cells).toContain('1:05')
  })

  // 19 rows in 1296 carry airborne times. A 0:00 in the other 1277 would be a
  // claim that the aeroplane never left the ground.
  it('leaves the airborne columns blank when the row has none', async () => {
    renderApp()
    await screen.findAllByText('1:21')
    expect(screen.queryByText('0:00')).not.toBeInTheDocument()
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
    await user.click(await screen.findByRole('link', { name: 'Stats' }))

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
    await user.click(await screen.findByRole('link', { name: 'Stats' }))
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
    await user.click(await screen.findByRole('link', { name: 'Stats' }))
    await screen.findByText('1219:35')
    expect(screen.queryByText(/landing split that was inferred/i)).not.toBeInTheDocument()
  })
})

// --- Editing and deleting a flight (2026-08-02) -----------------------------
//
// The app became the only way the record grows when the transcription effort
// closed, so a mistyped flight needs to be correctable. The rule underneath
// every one of these: only a flight ENTERED IN THE APP (source_book === 0) can
// be touched. The 1296 imported rows are closed data.

const APP_FLIGHT: Partial<Flight> = {
  seq: 1000000, source_book: 0, source_row: 1,
  date: '2026-07-30', aircraft_reg: 'OH-CAM', aircraft_type: 'C172', class: 'SEP_LAND',
  dep_place: 'EFHF', arr_place: 'EFHF',
  off_block_utc: '2026-07-30T09:15:00Z', on_block_utc: '2026-07-30T10:30:00Z',
  off_block_raw: '09:15Z', on_block_raw: '10:30Z', time_origin: 'utc_as_written',
  block_minutes: 75, total_minutes: 75, pic_minutes: 75, night_minutes: 20,
  instructor_minutes: 0, remarks: 'circuits',
}

describe('editing a flight', () => {
  it('offers Edit on a flight entered in the app and not on one from the paper books', async () => {
    vi.spyOn(api, 'flights').mockResolvedValue({
      flights: [flight(), flight(APP_FLIGHT)],
      count: 2,
    })
    renderApp()

    const links = await screen.findAllByRole('link', { name: /^Edit/ })
    expect(links).toHaveLength(1)
    expect(links[0]).toHaveAttribute('href', '/logbook/edit/1000000')
  })

  it('prefills the form with the stored flight, as four digits', async () => {
    vi.spyOn(api, 'flight').mockResolvedValue({ flight: flight(APP_FLIGHT) })
    window.history.pushState({}, '', '/logbook/edit/1000000')
    renderApp()

    expect(await screen.findByLabelText('Off block')).toHaveValue('0915')
    expect(screen.getByLabelText('On block')).toHaveValue('1030')
    expect(screen.getByLabelText('PIC')).toHaveValue('0115')
    expect(screen.getByLabelText('Night')).toHaveValue('0020')
    expect(screen.getByLabelText('Total time')).toHaveValue('1:15')
    expect(screen.getByLabelText('Remarks')).toHaveValue('circuits')
    expect(screen.getByLabelText('Aircraft')).toHaveValue('OH-CAM')
  })

  it('sends the correction with the same wire format as a new flight', async () => {
    vi.spyOn(api, 'flight').mockResolvedValue({ flight: flight(APP_FLIGHT) })
    const update = vi.spyOn(api, 'updateFlight').mockResolvedValue({
      flight: flight({ ...APP_FLIGHT, total_minutes: 90 }),
    })
    const user = userEvent.setup()
    window.history.pushState({}, '', '/logbook/edit/1000000')
    renderApp()

    const on = await screen.findByLabelText('On block')
    await user.clear(on)
    await user.type(on, '1045')
    await user.click(screen.getByRole('button', { name: 'Save this flight' }))

    await waitFor(() =>
      expect(update).toHaveBeenCalledWith(
        1000000,
        expect.objectContaining({
          off_block: '09:15Z',
          on_block: '10:45Z',
          total_time: '1:30',
          pic_time: '1:15',
          night_time: '0:20',
        }),
      ),
    )
    expect(await screen.findByRole('status')).toHaveTextContent(/saved/i)
  })

  // The same takeover as a new flight, minus the offer to log another one --
  // there is nothing to log another of when you came here to correct one.
  it('takes over the screen after a correction too, and can go back to it', async () => {
    vi.spyOn(api, 'flight').mockResolvedValue({ flight: flight(APP_FLIGHT) })
    vi.spyOn(api, 'updateFlight').mockResolvedValue({ flight: flight(APP_FLIGHT) })
    const user = userEvent.setup()
    window.history.pushState({}, '', '/logbook/edit/1000000')
    renderApp()

    await user.click(await screen.findByRole('button', { name: 'Save this flight' }))
    await screen.findByRole('status')
    expect(screen.queryByLabelText('Off block')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Log another flight' })).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Keep editing this flight' }))
    expect(await screen.findByLabelText('Off block')).toHaveValue('0915')
  })

  // A flight the server will not let us touch has to say why. "Forbidden" on
  // your own logbook reads as a bug otherwise.
  it('explains a refusal to edit a paper-book flight', async () => {
    vi.spyOn(api, 'flight').mockResolvedValue({ flight: flight({ seq: 5, source_book: 3 }) })
    window.history.pushState({}, '', '/logbook/edit/5')
    renderApp()

    expect(await screen.findByText(/transcribed from a paper logbook/i)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Save this flight' })).not.toBeInTheDocument()
  })
})

describe('deleting a flight', () => {
  async function openEdit() {
    vi.spyOn(api, 'flight').mockResolvedValue({ flight: flight(APP_FLIGHT) })
    window.history.pushState({}, '', '/logbook/edit/1000000')
    renderApp()
    return await screen.findByRole('button', { name: 'Delete this flight' })
  }

  // The owner asked for a double confirmation. One tap must never delete.
  it('does not delete on the first tap', async () => {
    const del = vi.spyOn(api, 'deleteFlight')
    const user = userEvent.setup()
    await user.click(await openEdit())

    expect(del).not.toHaveBeenCalled()
    expect(await screen.findByRole('alertdialog')).toBeInTheDocument()
  })

  // The confirmation names the flight. "Are you sure?" about an unnamed row is
  // how the wrong flight gets deleted.
  it('names the flight it is about to delete', async () => {
    const user = userEvent.setup()
    await user.click(await openEdit())

    const dialog = await screen.findByRole('alertdialog')
    expect(dialog).toHaveTextContent('2026-07-30')
    expect(dialog).toHaveTextContent('OH-CAM')
    expect(dialog).toHaveTextContent('1:15')
  })

  it('deletes only after the second confirmation', async () => {
    const del = vi.spyOn(api, 'deleteFlight').mockResolvedValue({ flight: flight(APP_FLIGHT) })
    const user = userEvent.setup()
    await user.click(await openEdit())
    await user.click(await screen.findByRole('button', { name: 'Yes, delete this flight' }))

    await waitFor(() => expect(del).toHaveBeenCalledWith(1000000))
  })

  it('keeps the flight when the confirmation is dismissed', async () => {
    const del = vi.spyOn(api, 'deleteFlight')
    const user = userEvent.setup()
    await user.click(await openEdit())
    await user.click(await screen.findByRole('button', { name: 'Keep it' }))

    expect(del).not.toHaveBeenCalled()
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument()
    // And the form is still there to go on editing.
    expect(screen.getByRole('button', { name: 'Save this flight' })).toBeInTheDocument()
  })
})

/**
 * Choosing an aeroplane in the filterable picker.
 *
 * It is a combobox rather than a <select>: the owner asked for type-to-filter
 * ("if I write P it should filter OH-PDP, OH-PIF and so on") in place of the
 * retired/active idea, which was dropped -- an aeroplane you flew once in 2009
 * is not retired, so nothing is ever hidden. Filtering is what keeps a growing
 * list usable.
 */
async function pickAircraft(user: ReturnType<typeof userEvent.setup>, reg: string) {
  await user.click(await screen.findByLabelText('Aircraft'))
  await user.click(await screen.findByRole('option', { name: new RegExp(reg) }))
}

describe('the aircraft picker', () => {
  it('filters the list as the registration is typed', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    const input = await screen.findByLabelText('Aircraft')
    await user.click(input)
    // Scoped to the picker's own listbox: the Class <select> further down the
    // form has options too, and they carry the same ARIA role.
    const list = await screen.findByRole('listbox', { name: 'Aircraft' })
    // Everything is offered before anything is typed: nothing is ever hidden.
    expect(within(list).getAllByRole('option')).toHaveLength(3)

    await user.type(input, 'P')
    const shown = within(await screen.findByRole('listbox', { name: 'Aircraft' }))
      .getAllByRole('option')
      .map((o) => o.textContent)
    expect(shown.some((t) => t?.includes('OH-PDP'))).toBe(true)
    expect(shown.some((t) => t?.includes('OH-CTL'))).toBe(false)
  })

  // Typing the type is as natural as typing the registration.
  it('filters on the aircraft type too', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    await user.type(await screen.findByLabelText('Aircraft'), 'P28')
    const shown = within(await screen.findByRole('listbox', { name: 'Aircraft' }))
      .getAllByRole('option')
      .map((o) => o.textContent)
    expect(shown.some((t) => t?.includes('OH-PDP'))).toBe(true)
    expect(shown.some((t) => t?.includes('OH-CAM'))).toBe(false)
  })

  // The server orders the list -- never-flown first, then most recently flown.
  // The picker must present it as sent rather than re-sorting alphabetically.
  it('keeps the order the server sent', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    await user.click(await screen.findByLabelText('Aircraft'))
    const shown = within(await screen.findByRole('listbox', { name: 'Aircraft' }))
      .getAllByRole('option')
      .map((o) => o.textContent ?? '')
    expect(shown[0]).toContain('OH-PDP')
    expect(shown[1]).toContain('OH-CTL')
    expect(shown[2]).toContain('OH-CAM')
  })

  it('lets a brand-new aeroplane be added without leaving the form', async () => {
    const user = userEvent.setup()
    const create = vi.spyOn(api, 'createAircraft').mockResolvedValue({
      aircraft: {
        registration: 'OH-XYZ', type: 'C152', default_class: 'SEP_LAND',
        ifr_capable: false, notes: '', user_added: true, last_flown: '', flights: 0,
      },
    })
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    await user.type(await screen.findByLabelText('Aircraft'), 'OH-XYZ')
    // Nothing matches, so the way forward is to add it.
    await user.click(await screen.findByRole('option', { name: /Add OH-XYZ/ }))

    await user.type(screen.getByLabelText('New aircraft type'), 'C152')
    await user.click(screen.getByRole('button', { name: 'Save aircraft' }))

    await waitFor(() =>
      expect(create).toHaveBeenCalledWith(
        expect.objectContaining({ registration: 'OH-XYZ', type: 'C152' }),
      ),
    )
    // And it is now the selected aeroplane, with its type and class carried
    // into the flight -- the pilot does not type the type twice.
    await waitFor(() => expect(screen.getByLabelText('Aircraft')).toHaveValue('OH-XYZ'))
    expect(screen.getByLabelText('Type')).toHaveValue('C152')
    expect(screen.getByLabelText('Class')).toHaveValue('SEP_LAND')
  })

  // A refusal has to be visible and must not lose what was typed.
  it('shows why a new aeroplane was refused', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'createAircraft').mockRejectedValue(
      new ApiError(409, 'that registration is already in the aircraft list'),
    )
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    await user.type(await screen.findByLabelText('Aircraft'), 'OH-ZZZ')
    await user.click(await screen.findByRole('option', { name: /Add OH-ZZZ/ }))
    await user.type(screen.getByLabelText('New aircraft type'), 'C152')
    await user.click(screen.getByRole('button', { name: 'Save aircraft' }))

    expect(await screen.findByText(/already in the aircraft list/)).toBeInTheDocument()
    expect(screen.getByLabelText('New aircraft type')).toHaveValue('C152')
  })

  // An aeroplane cannot be saved without a type -- the flight needs one, and
  // asking for it later means asking for it on the phone at the airfield.
  it('will not add an aeroplane with no type', async () => {
    const user = userEvent.setup()
    const create = vi.spyOn(api, 'createAircraft')
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    await user.type(await screen.findByLabelText('Aircraft'), 'OH-ZZZ')
    await user.click(await screen.findByRole('option', { name: /Add OH-ZZZ/ }))
    await user.click(screen.getByRole('button', { name: 'Save aircraft' }))

    expect(create).not.toHaveBeenCalled()
    expect(await screen.findByText(/type is required/i)).toBeInTheDocument()
  })
})

describe('the new-flight form', () => {
  it('fills the type and class from the chosen aircraft', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    await pickAircraft(user, 'OH-CTL')
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
    await pickAircraft(user, 'OH-CTL')
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

  // --- Task 11: the save has to be unmissable ------------------------------
  //
  // From the first real day of use (2026-08-02): the owner logged two flights
  // on the phone and could not tell whether either had saved. The confirmation
  // existed and was rendered ABOVE a form three cards long -- off-screen, on
  // the only device this app is used on. Ruled: the success takes over the
  // screen.

  it('replaces the form with the confirmation, so the save cannot be missed', async () => {
    vi.spyOn(api, 'createFlight').mockResolvedValue({
      flight: flight({ date: '2026-07-30', aircraft_reg: 'OH-CAM', total_minutes: 75 }),
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(await screen.findByRole('button', { name: 'Log this flight' }))

    // The confirmation names the flight, in full.
    const panel = await screen.findByRole('status')
    expect(panel).toHaveTextContent('2026-07-30')
    expect(panel).toHaveTextContent('OH-CAM')
    expect(panel).toHaveTextContent('1:15')

    // And the form is GONE. A message rendered above three cards of form is a
    // message a pilot on a phone never sees.
    expect(screen.queryByLabelText('Off block')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Log this flight' })).not.toBeInTheDocument()
  })

  it('offers the next flight and the table from the confirmation', async () => {
    vi.spyOn(api, 'createFlight').mockResolvedValue({
      flight: flight({ date: '2026-07-30', aircraft_reg: 'OH-CAM', total_minutes: 75 }),
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await pickAircraft(user, 'OH-CAM')
    await user.click(screen.getByRole('button', { name: 'Log this flight' }))

    await user.click(await screen.findByRole('button', { name: 'Log another flight' }))

    // Back to a fresh form -- but still on the same aeroplane, because the next
    // entry after a day of circuits is usually the same one.
    expect(await screen.findByLabelText('Off block')).toHaveValue('')
    expect(screen.getByLabelText('Aircraft')).toHaveValue('OH-CAM')
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('leads to the table from the confirmation', async () => {
    vi.spyOn(api, 'createFlight').mockResolvedValue({
      flight: flight({ date: '2026-07-30', aircraft_reg: 'OH-CAM', total_minutes: 75 }),
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(await screen.findByRole('button', { name: 'Log this flight' }))
    await user.click(await screen.findByRole('link', { name: 'See it in the table' }))

    expect(await screen.findByRole('columnheader', { name: 'Date' })).toBeInTheDocument()
  })

  // A phone that empties a twenty-field form because the server said 400 is a
  // phone that does not get the flight logged at all.
  it('keeps every field when the save fails', async () => {
    vi.spyOn(api, 'createFlight').mockRejectedValue(
      new ApiError(400, 'this flight cannot be logged as written', [
        { field: 'pic_name', message: 'a name is required' },
      ]),
    )
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(await screen.findByLabelText('Off block'), '0915')
    await user.type(screen.getByLabelText('On block'), '1030')
    await user.type(screen.getByLabelText('Remarks'), 'circuits at Hyvinkää')
    await user.click(screen.getByRole('button', { name: 'Log this flight' }))

    expect(await screen.findByText('a name is required')).toBeInTheDocument()
    expect(screen.getByLabelText('Off block')).toHaveValue('0915')
    expect(screen.getByLabelText('On block')).toHaveValue('1030')
    expect(screen.getByLabelText('Remarks')).toHaveValue('circuits at Hyvinkää')
  })

  // "A failure gets the same prominence in red, scrolled to the field that
  // caused it." Focus is what scrolling means on a phone, and it is the part a
  // test can assert.
  it('moves to the field the server refused', async () => {
    vi.spyOn(api, 'createFlight').mockRejectedValue(
      new ApiError(400, 'this flight cannot be logged as written', [
        // Deliberately out of page order: the topmost failing control wins,
        // not whichever one the server happened to name first.
        { field: 'remarks', message: 'too long' },
        { field: 'off_block', message: 'an off-block time is required' },
      ]),
    )
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(await screen.findByRole('button', { name: 'Log this flight' }))

    await waitFor(() => expect(screen.getByLabelText('Off block')).toHaveFocus())
  })

  // The page learned this the hard way once already, when an <output>'s
  // implicit role="status" collided with the saved-flight announcement.
  it('has exactly one live region on the confirmation', async () => {
    vi.spyOn(api, 'createFlight').mockResolvedValue({
      flight: flight({ date: '2026-07-30', aircraft_reg: 'OH-CAM', total_minutes: 75 }),
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(await screen.findByRole('button', { name: 'Log this flight' }))

    await screen.findByRole('status')
    expect(screen.getAllByRole('status')).toHaveLength(1)
  })

  // --- The times: four digits, on a number pad -----------------------------
  //
  // Every time on this form is typed as HHMM and nothing else. The first
  // version asked for "09:15Z" on a keyboard with no colon key; the second
  // reached for native pickers and left the durations needing a colon. The
  // owner's instruction is one rule for all of them: four numbers, always,
  // with the zone carried by the toggle and the totals worked out for you.

  it('takes every time as four digits on a number pad', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    for (const label of [
      'Off block', 'On block', 'Takeoff', 'Landing',
      'PIC', 'Dual', 'Night', 'Instrument', 'Instructor',
    ]) {
      const field = await screen.findByLabelText(label)
      expect(field, label).toHaveAttribute('inputmode', 'numeric')
      expect(field, label).toHaveAttribute('maxlength', '4')
    }
  })

  // A pasted "09:15Z" -- from a message, or from the old habit -- must not
  // become a field the pilot has to clean up by hand.
  it('drops anything that is not a digit as it is typed', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    const off = await screen.findByLabelText('Off block')
    await user.type(off, '09:15Z')
    expect(off).toHaveValue('0915')
  })

  it('derives the total time from the clock and refuses to let it be typed', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    const total = await screen.findByLabelText('Total time')
    expect(total).toHaveAttribute('readonly')
    expect(total).toHaveValue('')

    await user.type(await screen.findByLabelText('Off block'), '0915')
    await user.type(screen.getByLabelText('On block'), '1030')
    expect(screen.getByLabelText('Total time')).toHaveValue('1:15')
  })

  it('derives the total across midnight rather than going negative', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(await screen.findByLabelText('Off block'), '2330')
    await user.type(screen.getByLabelText('On block'), '0040')
    expect(screen.getByLabelText('Total time')).toHaveValue('1:10')
  })

  // The airborne pair is NOT folded away any more (Task 12, owner-ruled). It
  // was hidden because most rows in the PAPER BOOKS have none -- a fact about
  // 1296 historical rows, not about the flights being flown now. A field you
  // have to remember to expand is a field that ends up empty, and an empty
  // airborne time is what makes an air-time total unusable a year later when
  // it is being billed from.
  it('shows the airborne times without anything to expand', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await screen.findByLabelText('Takeoff')

    expect(document.querySelector('details.airborne')).toBeNull()
    expect(screen.queryByText(/Takeoff and landing \(optional\)/i)).not.toBeInTheDocument()
  })

  it('derives the air time from the optional takeoff and landing', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))

    // Blank by default: most rows in the paper books carry no airborne times.
    expect(await screen.findByLabelText('Air time')).toHaveValue('')

    await user.type(screen.getByLabelText('Takeoff'), '0920')
    await user.type(screen.getByLabelText('Landing'), '1025')
    expect(screen.getByLabelText('Air time')).toHaveValue('1:05')
  })

  // The Z is load-bearing (rule 0.4) and cannot be typed on a number pad, so
  // the zone is the toggle's job and the wire format is unchanged.
  it('sends the times with a Z when the zone is UTC', async () => {
    const create = vi.spyOn(api, 'createFlight').mockResolvedValue({
      flight: flight({ date: '2026-07-30', aircraft_reg: 'OH-CAM', total_minutes: 75 }),
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(await screen.findByLabelText('Off block'), '0915')
    await user.type(screen.getByLabelText('On block'), '1030')
    await user.type(screen.getByLabelText('Takeoff'), '0920')
    await user.type(screen.getByLabelText('Landing'), '1025')
    await user.click(screen.getByRole('button', { name: 'Log this flight' }))

    expect(create).toHaveBeenCalledWith(
      expect.objectContaining({
        off_block: '09:15Z',
        on_block: '10:30Z',
        takeoff: '09:20Z',
        landing: '10:25Z',
        // Derived, and stated on the wire: the server still requires the total
        // rather than inventing it.
        total_time: '1:15',
      }),
    )
  })

  it('sends the times bare when the zone is Helsinki local', async () => {
    const create = vi.spyOn(api, 'createFlight').mockResolvedValue({
      flight: flight({ date: '2026-07-30', aircraft_reg: 'OH-CAM', total_minutes: 75 }),
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(await screen.findByRole('button', { name: 'Helsinki local' }))
    await user.type(screen.getByLabelText('Off block'), '1215')
    await user.type(screen.getByLabelText('On block'), '1330')
    await user.click(screen.getByRole('button', { name: 'Log this flight' }))

    expect(create).toHaveBeenCalledWith(
      expect.objectContaining({ off_block: '12:15', on_block: '13:30' }),
    )
  })

  // An empty airborne pair must not travel as a half-written one.
  it('leaves the airborne pair empty when it is not filled in', async () => {
    const create = vi.spyOn(api, 'createFlight').mockResolvedValue({
      flight: flight({ date: '2026-07-30', aircraft_reg: 'OH-CAM', total_minutes: 75 }),
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(await screen.findByLabelText('Off block'), '0915')
    await user.type(screen.getByLabelText('On block'), '1030')
    await user.click(screen.getByRole('button', { name: 'Log this flight' }))

    expect(create).toHaveBeenCalledWith(
      expect.objectContaining({ takeoff: '', landing: '' }),
    )
  })

  // The durations are four digits too, and the H:MM the server wants is this
  // form's job rather than the pilot's.
  it('sends a four-digit duration as the H:MM the API expects', async () => {
    const create = vi.spyOn(api, 'createFlight').mockResolvedValue({
      flight: flight({ date: '2026-07-30', aircraft_reg: 'OH-CAM', total_minutes: 75 }),
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(await screen.findByLabelText('Off block'), '0915')
    await user.type(screen.getByLabelText('On block'), '1030')
    await user.type(screen.getByLabelText('PIC'), '0115')
    await user.type(screen.getByLabelText('Night'), '0020')
    await user.click(screen.getByRole('button', { name: 'Log this flight' }))

    expect(create).toHaveBeenCalledWith(
      expect.objectContaining({ pic_time: '1:15', night_time: '0:20', dual_time: '' }),
    )
  })

  // Half a time is not a time. Refusing here, before the request, is the one
  // place this form is allowed to decide anything -- and it says which field.
  it('refuses a half-typed time instead of sending it', async () => {
    const create = vi.spyOn(api, 'createFlight')
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(await screen.findByLabelText('Off block'), '091')
    await user.type(screen.getByLabelText('On block'), '1030')
    await user.click(screen.getByRole('button', { name: 'Log this flight' }))

    // The exact sentence, not a substring: the card's own guidance says "four
    // digits" too, and an assertion that cannot tell the guidance from the
    // error would pass on a form that reported nothing at all.
    expect(
      await screen.findByText('Write this time as four digits, for example 0915 for 09:15.'),
    ).toBeInTheDocument()
    expect(screen.getByLabelText('Off block')).toHaveAttribute('aria-invalid', 'true')
    expect(create).not.toHaveBeenCalled()
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

// --- Task 13: the aircraft time page ---------------------------------------
//
// "I pay for the aeroplanes by the hour, and some owners charge block time and
// some charge air time" (owner, 2026-08-02). This is money, so the load-bearing
// property is honesty about coverage: block time is known for every flight, air
// time for 19 of 1296, and a page that added up what it had and printed one
// figure would be claiming a completeness it does not have.

describe('the aircraft time page', () => {
  async function open() {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'Aircraft' }))
    return user
  }

  // Both, because an invoice is checked in one and computed in the other.
  // findAllByText: with a single aeroplane in range the figure legitimately
  // appears twice, once in the range total and once in its only row.
  it('gives block and air time in H:MM and in whole minutes', async () => {
    await open()
    expect((await screen.findAllByText('2:15')).length).toBeGreaterThan(0) // block, 135
    expect(screen.getAllByText('1:05').length).toBeGreaterThan(0)          // air, 65
    expect(screen.getAllByText(/135 min/).length).toBeGreaterThan(0)
    expect(screen.getAllByText(/65 min/).length).toBeGreaterThan(0)
  })

  // The figure never travels without the coverage that makes it readable.
  it('states how many flights the air figure was computed from', async () => {
    await open()
    expect(await screen.findByText(/recorded on 1 of 2 flights/i)).toBeInTheDocument()
    // And says outright that the total is partial, so it is not read as the
    // block figure's equal.
    expect(screen.getByText(/partial/i)).toBeInTheDocument()
  })

  // The asymmetry between the two sentences IS the message: block states a
  // fact, air states a fraction. Flattening them into a matching pair of
  // figures is the mistake this page exists to avoid.
  it('does not present air time as complete when it is not', async () => {
    await open()
    expect(await screen.findByText(/recorded on every flight/i)).toBeInTheDocument()
    expect(screen.getByText(/recorded on 1 of 2 flights/i)).toBeInTheDocument()
  })

  // Zero airborne times is not "0:00 airborne". It is "nobody wrote it down",
  // and the page has to say which.
  it('says plainly when no airborne time was recorded at all', async () => {
    vi.spyOn(api, 'aircraftTime').mockResolvedValue({
      range: { from: '', to: '' }, reg: '',
      aircraft: [acTime({ air_minutes: 0, air_known: 0, air_missing: 2 })],
      total: acTime({ registration: '', types: [], air_minutes: 0, air_known: 0, air_missing: 2 }),
      flights: [],
    })
    await open()
    expect(await screen.findByText(/no airborne times recorded/i)).toBeInTheDocument()
  })

  // "The list of flights behind the figure, so a disputed line can be traced to
  // a flight rather than argued against a single number."
  it('shows the flights behind one aeroplane on request', async () => {
    const load = vi.spyOn(api, 'aircraftTime')
    const user = await open()
    await user.click(await screen.findByRole('button', { name: /OH-CAM/ }))

    await waitFor(() =>
      expect(load).toHaveBeenCalledWith(expect.anything(), 'OH-CAM'),
    )
  })

  it('asks the server again when the range changes', async () => {
    const user = await open()
    await screen.findAllByText('2:15')
    await user.type(screen.getByLabelText('From'), '2024-01-01')

    await waitFor(() =>
      expect(api.aircraftTime).toHaveBeenCalledWith(
        expect.objectContaining({ from: '2024-01-01' }),
        undefined,
      ),
    )
  })

  // One registration written with two types is a discrepancy to show, not to
  // resolve by picking the more popular spelling.
  it('shows every type written for a registration', async () => {
    vi.spyOn(api, 'aircraftTime').mockResolvedValue({
      range: { from: '', to: '' }, reg: '',
      aircraft: [acTime({ registration: 'OH-CMU', types: ['C152', 'C172'] })],
      total: acTime({ registration: '', types: [] }),
      flights: [],
    })
    await open()
    expect(await screen.findByText(/C152 · C172/)).toBeInTheDocument()
  })

  it('reports a failure instead of an empty bill', async () => {
    vi.spyOn(api, 'aircraftTime').mockRejectedValue(
      new ApiError(500, 'could not read the logbook'),
    )
    await open()
    expect(await screen.findByRole('alert')).toHaveTextContent('could not read the logbook')
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

/**
 * The fleet page — Task 19.
 *
 * `POST` and `PUT /aircraft` have been live and deployed since 2026-08-02 and
 * `api.updateAircraft` had ZERO callers: a typo'd registration could be created
 * from the flight form and then never corrected, and by ruling it cannot be
 * deleted either. The no-delete ruling is only humane if editing exists.
 */
/**
 * The nth aeroplane's row. A helper because indexing a query result is
 * `HTMLElement | undefined` under this project's strict tsconfig, and a
 * missing row should fail with which row was missing rather than a cast.
 */
function fleetRow(fleet: HTMLElement, n: number): HTMLElement {
  const rows = within(fleet).getAllByRole('listitem')
  const row = rows[n]
  if (!row) throw new Error(`no fleet row ${n}: there are ${rows.length}`)
  return row
}

describe('the fleet page', () => {
  // It is deliberately not a seventh tab: six already share a 390px phone.
  it('is reachable from the Aircraft tab, and lists every aeroplane', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'Aircraft' }))
    await user.click(await screen.findByRole('link', { name: /manage the fleet/i }))

    const fleet = await screen.findByRole('list', { name: 'Fleet' })
    const rows = within(fleet).getAllByRole('listitem')
    expect(rows).toHaveLength(3)
    // The server's order, not ours: never-flown first, then most recently flown.
    expect(rows[0]).toHaveTextContent('OH-PDP')
    expect(rows[0]).toHaveTextContent('not flown yet')
    expect(rows[1]).toHaveTextContent('OH-CTL')
    expect(rows[1]).toHaveTextContent('286 flights')
  })

  it('is reachable directly by URL, so it can be bookmarked', async () => {
    window.history.pushState({}, '', '/logbook/fleet')
    renderApp()
    expect(await screen.findByRole('list', { name: 'Fleet' })).toBeInTheDocument()
  })

  // The U. This is the whole reason the page exists.
  it('corrects a registration, and sends the whole aeroplane', async () => {
    const put = vi.spyOn(api, 'updateAircraft').mockResolvedValue({
      aircraft: { registration: 'OH-PDQ', type: 'P28A', default_class: 'SEP_LAND',
        ifr_capable: false, notes: '', user_added: true, last_flown: '', flights: 0 },
    })
    const user = userEvent.setup()
    window.history.pushState({}, '', '/logbook/fleet')
    renderApp()

    const fleet = await screen.findByRole('list', { name: 'Fleet' })
    const row = fleetRow(fleet, 0)
    await user.click(within(row).getByRole('button', { name: /edit/i }))

    const reg = screen.getByLabelText('Registration')
    await user.clear(reg)
    await user.type(reg, 'OH-PDQ')
    await user.click(screen.getByRole('button', { name: /save/i }))

    // Keyed by the OLD registration -- that is what the route names -- and
    // carrying every field, because the endpoint is a replacement, not a patch.
    await waitFor(() =>
      expect(put).toHaveBeenCalledWith('OH-PDP', {
        registration: 'OH-PDQ', type: 'P28A', default_class: 'SEP_LAND',
        ifr_capable: false, notes: '',
      }),
    )
    expect(await screen.findByText(/OH-PDQ/)).toBeInTheDocument()
  })

  // The owner asked for creating an aircraft here as well as from the form:
  // the form's inline panel asks only for type and class, on purpose.
  it('adds an aeroplane without going near the flight form', async () => {
    const post = vi.spyOn(api, 'createAircraft').mockResolvedValue({
      aircraft: { registration: 'OH-XYZ', type: 'DA40', default_class: 'SEP_LAND',
        ifr_capable: true, notes: 'club', user_added: true, last_flown: '', flights: 0 },
    })
    const user = userEvent.setup()
    window.history.pushState({}, '', '/logbook/fleet')
    renderApp()

    await user.click(await screen.findByRole('button', { name: /add an aircraft/i }))
    await user.type(screen.getByLabelText('Registration'), 'oh-xyz')
    await user.type(screen.getByLabelText('Type'), 'da40')
    await user.click(screen.getByLabelText('IFR capable'))
    await user.type(screen.getByLabelText('Notes'), 'club')
    await user.click(screen.getByRole('button', { name: /save/i }))

    // Upper-cased before it is sent, exactly like the picker does it: the
    // registration is an identifier and 'oh-xyz' must not become a second row.
    await waitFor(() =>
      expect(post).toHaveBeenCalledWith({
        registration: 'OH-XYZ', type: 'DA40', default_class: 'SEP_LAND',
        ifr_capable: true, notes: 'club',
      }),
    )
    const fleet = await screen.findByRole('list', { name: 'Fleet' })
    expect(within(fleet).getAllByRole('listitem')).toHaveLength(4)
  })

  it('shows the server’s reason when a correction collides', async () => {
    vi.spyOn(api, 'updateAircraft').mockRejectedValue(
      new ApiError(409, 'that registration is already in the aircraft list'))
    const user = userEvent.setup()
    window.history.pushState({}, '', '/logbook/fleet')
    renderApp()

    const fleet = await screen.findByRole('list', { name: 'Fleet' })
    await user.click(within(fleetRow(fleet, 0)).getByRole('button', { name: /edit/i }))
    const reg = screen.getByLabelText('Registration')
    await user.clear(reg)
    await user.type(reg, 'OH-CTL')
    await user.click(screen.getByRole('button', { name: /save/i }))

    expect(await screen.findByRole('alert'))
      .toHaveTextContent('that registration is already in the aircraft list')
  })

  // The no-delete ruling (2026-08-02) is asserted on the backend by a
  // route-table test. This is the same guard on the other side of the wire:
  // "symmetry" is exactly how it would get added back.
  it('offers nothing anywhere that deletes an aeroplane', async () => {
    const user = userEvent.setup()
    window.history.pushState({}, '', '/logbook/fleet')
    renderApp()

    const fleet = await screen.findByRole('list', { name: 'Fleet' })
    await user.click(within(fleetRow(fleet, 0)).getByRole('button', { name: /edit/i }))
    for (const b of screen.getAllByRole('button')) {
      expect(b.textContent ?? '').not.toMatch(/delete|remove/i)
    }
    expect(api).not.toHaveProperty('deleteAircraft')
  })
})

/**
 * The PIC picker — Task 21.
 *
 * The owner's ask, verbatim (2026-08-03): "I could have a typo when I write
 * `self`, it could be `sself` or `SELF` or `seeelf` and I need it to be
 * consistent (like the aircraft regs)." `self` is on 1143 of the 1296
 * transcribed flights, so a second spelling would split the busiest value in
 * the column with nothing to flag it.
 */
describe('the pilot picker', () => {
  it('offers the names the record already uses, in the order the server sent', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(screen.getByLabelText('Name of pilot in command'))

    const list = await screen.findByRole('listbox', { name: 'Pilots' })
    const names = within(list).getAllByRole('option').map((o) => o.textContent)
    expect(names[0]).toContain('Jansson')
    expect(names[0]).toContain('not flown with yet')
    expect(names[1]).toContain('self')
    expect(names[1]).toContain('1143 flights')
  })

  it('filters as the name is typed', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(screen.getByLabelText('Name of pilot in command'), 'mar')

    const list = await screen.findByRole('listbox', { name: 'Pilots' })
    const names = within(list).getAllByRole('option').map((o) => o.textContent)
    expect(names.some((n) => n?.includes('Martevuo'))).toBe(true)
    expect(names.some((n) => n?.includes('self'))).toBe(false)
  })

  it('puts the chosen name in the field', async () => {
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.click(screen.getByLabelText('Name of pilot in command'))
    await user.click(await screen.findByRole('option', { name: /self/ }))

    expect(screen.getByLabelText('Name of pilot in command')).toHaveValue('self')
  })

  // The other half of the ask: a name genuinely new has to be enterable, at an
  // airfield, without leaving the form — exactly like a new aeroplane.
  it('adds a name that is genuinely new, without leaving the form', async () => {
    const post = vi.spyOn(api, 'createPilot').mockResolvedValue({
      pilot: { name: 'Lehtinen', user_added: true, last_flown: '', flights: 0 },
    })
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(screen.getByLabelText('Name of pilot in command'), 'Lehtinen')
    await user.click(await screen.findByRole('option', { name: /add lehtinen/i }))

    await waitFor(() => expect(post).toHaveBeenCalledWith('Lehtinen'))
    expect(screen.getByLabelText('Name of pilot in command')).toHaveValue('Lehtinen')
  })

  // The typo the owner named, and the answer to it: the picker will not even
  // OFFER to add a spelling that only differs by case, so `SELF` cannot become
  // a second person beside the `self` on 1143 flights. It filters to the real
  // one instead, which is the thing to tap.
  it('will not offer to add a name that is only a case variant of a known one', async () => {
    const post = vi.spyOn(api, 'createPilot')
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(screen.getByLabelText('Name of pilot in command'), 'SELF')

    const list = await screen.findByRole('listbox', { name: 'Pilots' })
    const names = within(list).getAllByRole('option').map((o) => o.textContent)
    expect(names.some((n) => /add/i.test(n ?? ''))).toBe(false)
    expect(names.some((n) => n?.includes('self'))).toBe(true)
    expect(post).not.toHaveBeenCalled()
  })

  // And if the server refuses anyway -- the roster in this browser is a
  // response fetched some time ago, and another device may have added the name
  // since -- its own words are shown rather than "could not save".
  it('shows the server’s reason when adding a name is refused', async () => {
    vi.spyOn(api, 'createPilot').mockRejectedValue(
      new ApiError(409, 'that name is already in the pilot list'))
    const user = userEvent.setup()
    renderApp()
    await user.click(await screen.findByRole('link', { name: 'New' }))
    await user.type(screen.getByLabelText('Name of pilot in command'), 'Lehtinen')
    await user.click(await screen.findByRole('option', { name: /add lehtinen/i }))

    expect(await screen.findByRole('alert'))
      .toHaveTextContent('that name is already in the pilot list')
  })

  // A flight already in the books may name somebody the roster no longer
  // offers, and the edit form must not silently blank it. Same rule the
  // aircraft picker follows.
  it('keeps a name the roster does not offer when editing a flight', async () => {
    vi.spyOn(api, 'flight').mockResolvedValue({
      flight: flight({ seq: 1000001, source_book: 0, pic_name: 'Kaariainen' }),
    })
    window.history.pushState({}, '', '/logbook/edit/1000001')
    renderApp()

    expect(await screen.findByLabelText('Name of pilot in command'))
      .toHaveValue('Kaariainen')
  })
})
