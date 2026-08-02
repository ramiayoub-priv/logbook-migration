import { useCallback, useState } from 'react'
import { api, ApiError, type Flight } from '../api'
import { useApi } from '../auth'
import { clock, hhmm } from '../format'
import { navigate } from '../router'
import { draftFromFlight, FlightForm } from './FlightForm'

/**
 * The edit page.
 *
 * It exists because the transcription effort closed on 2026-08-02 and this
 * application became the only way the record grows: a flight mistyped on a
 * phone was, until now, correctable only by opening SQLite on the server.
 *
 * Two rules shape the whole page, and both come from the record being a legal
 * one rather than from taste:
 *
 *   - **Only a flight entered in the app can be changed.** The 1296 rows
 *     transcribed from the paper books are closed data (CLAUDE.md rule 0.8),
 *     and the importer owns them — it replaces every row with source_book <> 0
 *     on each run, so an edit to one would be discarded without warning. Such a
 *     flight still LOADS here, and the page says why it cannot be edited: a
 *     page that refused to open it would read as a broken link.
 *   - **Deleting asks twice, and the second question names the flight.** "Are
 *     you sure?" about an unnamed row is how the wrong flight gets deleted.
 */
export function EditFlightPage({ seq }: { seq: number }) {
  const load = useCallback(() => api.flight(seq), [seq])
  const { data, error, loading } = useApi(load, [seq])

  if (loading && !data) return <p className="center">Loading the flight…</p>
  if (error) {
    return (
      <p className="error" role="alert">
        {error}
      </p>
    )
  }
  if (!data) return null

  const flight = data.flight
  if (flight.source_book !== 0) {
    return <PaperFlight flight={flight} />
  }
  // key on seq: opening a different flight must start from its values, not
  // from whatever the last one was left holding.
  return <EditableFlight key={flight.seq} flight={flight} />
}

function EditableFlight({ flight }: { flight: Flight }) {
  return (
    <>
      <p className="muted small">
        Editing the flight of {flight.date}, {flight.aircraft_reg}. Saving replaces every field
        with what this form holds.
      </p>
      <FlightForm
        initial={draftFromFlight(flight)}
        submitLabel="Save this flight"
        busyLabel="Saving…"
        // No "log another" here: there is nothing to log another of when you
        // came to correct one. The way back is to the same flight.
        againLabel="Keep editing this flight"
        onSave={async (payload) => {
          const { flight: saved } = await api.updateFlight(flight.seq, payload)
          return {
            message: `Saved ${saved.date}, ${saved.aircraft_reg}, ${hhmm(saved.total_minutes)}.`,
            flight: saved,
          }
        }}
        footer={<DeletePanel flight={flight} />}
      />
    </>
  )
}

/**
 * A flight from the paper books: shown, explained, and not editable.
 *
 * It is a full stop rather than a disabled form. A form that looked editable
 * and then refused every save would be worse than one that never offered.
 */
function PaperFlight({ flight }: { flight: Flight }) {
  return (
    <div className="card">
      <h2>
        {flight.date} · {flight.aircraft_reg}
      </h2>
      <p>
        This flight was <strong>transcribed from a paper logbook</strong> (book {flight.source_book},
        row {flight.source_row}) and cannot be changed in the app.
      </p>
      <p className="muted small">
        The transcribed books are a closed record: they were reconciled against the paper page by
        page and are no longer edited. Only flights entered in this app can be corrected here. If
        something in this row is wrong, it is a paper-side question rather than an app one.
      </p>
      <dl className="detail">
        <div>
          <dt>Times</dt>
          <dd>
            {clock(flight.off_block_utc)} – {clock(flight.on_block_utc)} UTC
          </dd>
        </div>
        <div>
          <dt>Total</dt>
          <dd>{hhmm(flight.total_minutes)}</dd>
        </div>
        <div>
          <dt>Route</dt>
          <dd>
            {flight.dep_place} → {flight.arr_place}
          </dd>
        </div>
      </dl>
      <button type="button" onClick={() => navigate('table')}>
        Back to the flights
      </button>
    </div>
  )
}

/**
 * The delete control: two steps, and the second one names the flight.
 *
 * The confirmation is a real region with role="alertdialog" rather than a
 * window.confirm, because the browser dialog cannot say which flight it is
 * about, cannot be styled to a phone, and cannot be tested.
 */
function DeletePanel({ flight }: { flight: Flight }) {
  const [asking, setAsking] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function remove() {
    setBusy(true)
    setError(null)
    try {
      await api.deleteFlight(flight.seq)
      // Back to the book, which no longer contains it. Nothing is left on
      // screen claiming to be a flight that is gone.
      navigate('table')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not delete the flight.')
      setBusy(false)
    }
  }

  if (!asking) {
    return (
      <div className="card danger">
        <h2>Delete</h2>
        <p className="muted small">
          Removes this flight from the logbook. Every total changes with it.
        </p>
        <button type="button" className="danger" onClick={() => setAsking(true)}>
          Delete this flight
        </button>
      </div>
    )
  }

  return (
    <div className="card danger" role="alertdialog" aria-labelledby="delete-heading">
      <h2 id="delete-heading">Are you sure?</h2>
      <p>
        This will permanently delete the flight of <strong>{flight.date}</strong> in{' '}
        <strong>{flight.aircraft_reg}</strong>, {clock(flight.off_block_utc)}–
        {clock(flight.on_block_utc)} UTC, total <strong>{hhmm(flight.total_minutes)}</strong>
        {flight.landings_day + flight.landings_night > 0 && (
          <> and {flight.landings_day + flight.landings_night} landing
            {flight.landings_day + flight.landings_night === 1 ? '' : 's'}</>
        )}
        .
      </p>
      <p className="muted small">
        Your totals will drop by {hhmm(flight.total_minutes)}. This cannot be undone from the app.
      </p>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <div className="row">
        <button type="button" className="danger" disabled={busy} onClick={remove}>
          {busy ? 'Deleting…' : 'Yes, delete this flight'}
        </button>
        <button type="button" onClick={() => setAsking(false)}>
          Keep it
        </button>
      </div>
    </div>
  )
}
