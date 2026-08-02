import { useCallback, useState } from 'react'
import { api, type AircraftTime, type DateRange, type Flight } from '../api'
import { useApi } from '../auth'
import { airMinutes, clock, hhmm } from '../format'
import { RangePicker } from './RangePicker'

/**
 * What each aeroplane cost, over a date range.
 *
 * Asked for by the owner on 2026-08-02, in these terms: *"I pay for the
 * aeroplanes by the hour, and some owners charge block time and some charge air
 * time."* That makes this page about money, which changes what it owes the
 * reader compared with the statistics page next door.
 *
 * THREE THINGS ARE LOAD-BEARING AND NONE OF THEM ARE COSMETIC.
 *
 * 1. **Block and air are never mixed and never silently completed.** Block time
 *    is recorded on every flight in the logbook; air time on 19 of the 1296
 *    transcribed ones. So the air figure is always shown WITH the count of
 *    flights it was computed from, and a range with no airborne times says
 *    "none recorded" rather than printing 0:00 — a zero is a claim that the
 *    aeroplane never left the ground, and it is the claim that would quietly
 *    deflate an invoice. That is rule 0.2 applied to a figure nobody had
 *    computed before: surface the gap, never paper over it.
 *
 * 2. **Both totals appear in H:MM and in whole minutes**, at the owner's
 *    instruction. An invoice is checked in one and computed in the other, and
 *    doing the conversion in your head next to a bill is where a mistake gets
 *    paid for.
 *
 * 3. **The flights behind the figure are reachable**, so a disputed line can be
 *    traced to a flight rather than argued against a single number.
 *
 * The arithmetic is NOT here. It is in `internal/stats`, pure and at 100%,
 * alongside the licence totals — this is the same class of code, and a browser
 * that sums a bill is a second place for the figure to be wrong.
 */
export function AircraftTimePage() {
  const [range, setRange] = useState<DateRange>({})
  const [reg, setReg] = useState<string | undefined>(undefined)

  const load = useCallback(() => api.aircraftTime(range, reg), [range.from, range.to, reg])
  const { data, error, loading } = useApi(load, [range.from, range.to, reg])

  return (
    <>
      <RangePicker range={range} onChange={setRange} />

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {loading && !data && <p className="center">Adding up the aeroplanes…</p>}

      {data && (
        <>
          <Totals total={data.total} />

          {data.aircraft.length === 0 ? (
            <p className="center">No flights in this range.</p>
          ) : (
            <div className="card" style={{ padding: 0 }}>
              <div className="tablewrap">
                <table>
                  <thead>
                    <tr>
                      <th>Aircraft</th>
                      <th className="num">Flights</th>
                      <th className="num">Block</th>
                      <th className="num">Block min</th>
                      <th className="num">Air</th>
                      <th className="num">Air min</th>
                      <th>Air recorded on</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.aircraft.map((a) => (
                      <Row
                        key={a.registration}
                        a={a}
                        chosen={a.registration === reg}
                        onChoose={() =>
                          setReg(a.registration === reg ? undefined : a.registration)
                        }
                      />
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {reg && <Behind reg={reg} flights={data.flights} onClose={() => setReg(undefined)} />}
        </>
      )}
    </>
  )
}

/**
 * The range total, and the two sentences that make it honest.
 *
 * The block sentence and the air sentence are deliberately different in kind:
 * one states a fact, the other states a fraction. That asymmetry IS the
 * message, and flattening the two into a matching pair of figures is exactly
 * the mistake this page exists to avoid.
 */
function Totals({ total }: { total: AircraftTime }) {
  return (
    <>
      <h3>This range</h3>
      <div className="figures">
        <Figure label="Flights" value={String(total.flights)} />
        <Figure
          label="Block time"
          value={hhmm(total.block_minutes)}
          minutes={total.block_minutes}
        />
        <Figure
          label="Air time"
          value={total.air_known === 0 ? '—' : hhmm(total.air_minutes)}
          {...(total.air_known === 0 ? {} : { minutes: total.air_minutes })}
        />
      </div>

      <p className="muted small">
        Block time is chocks to chocks and is <strong>recorded on every flight</strong>, so
        the figure above covers all {total.flights} of them.
      </p>

      {total.air_known === 0 ? (
        <p className="warn small">
          <strong>No airborne times recorded</strong> on any flight in this range, so there
          is no air time to report — not zero minutes airborne. If an owner bills on air
          time, fill in takeoff and landing when you log the flight.
        </p>
      ) : (
        <p className={total.air_missing > 0 ? 'warn small' : 'muted small'}>
          Air time is wheels up to wheels down, and is{' '}
          <strong>
            recorded on {total.air_known} of {total.flights} flight
            {total.flights === 1 ? '' : 's'}
          </strong>{' '}
          in this range.
          {total.air_missing > 0 && (
            <>
              {' '}
              The other {total.air_missing} carr{total.air_missing === 1 ? 'ies' : 'y'} no
              takeoff or landing time, so the air figure is a <strong>partial</strong> total
              and is not comparable with the block figure beside it.
            </>
          )}
        </p>
      )}

      {/* Almost always absent. When a flight's block time and the total the
          licence runs on disagree, this page and the statistics page will
          legitimately differ, and saying so beats being asked why. */}
      {total.block_differs_from_total > 0 && (
        <p className="warn small">
          {total.block_differs_from_total} flight
          {total.block_differs_from_total === 1 ? ' has a' : 's have'} block time that differs
          from the total time the licence figures run on. This page bills on{' '}
          <strong>block</strong>, so it will not match the statistics page for{' '}
          {total.block_differs_from_total === 1 ? 'that flight' : 'those flights'}.
        </p>
      )}
    </>
  )
}

function Row({
  a,
  chosen,
  onChoose,
}: {
  a: AircraftTime
  chosen: boolean
  onChoose: () => void
}) {
  return (
    <tr className={chosen ? 'chosen' : undefined}>
      <td>
        {/* A button rather than a link: this selects what the page shows below
            it, it does not navigate anywhere. */}
        <button type="button" className="link plain" onClick={onChoose} aria-pressed={chosen}>
          {a.registration}
        </button>
        {/* Every type written for this registration, not the most popular one.
            Two spellings is a discrepancy to show, not to resolve here. */}
        <span className="muted"> {a.types.join(' · ')}</span>
      </td>
      <td className="num">{a.flights}</td>
      <td className="num">{hhmm(a.block_minutes)}</td>
      <td className="num muted">{a.block_minutes} min</td>
      <td className="num">{a.air_known === 0 ? '—' : hhmm(a.air_minutes)}</td>
      <td className="num muted">{a.air_known === 0 ? '' : `${a.air_minutes} min`}</td>
      <td className={a.air_missing > 0 ? 'warntext' : undefined}>
        {a.air_known} of {a.flights}
      </td>
    </tr>
  )
}

/** The flights behind one aeroplane's figure. */
function Behind({
  reg,
  flights,
  onClose,
}: {
  reg: string
  flights: Flight[]
  onClose: () => void
}) {
  return (
    <div className="card">
      <h2>{reg} — the flights behind the figure</h2>
      <p className="muted small">
        In the book's own order. Air time is worked out from the takeoff and landing times and
        is blank where the flight recorded none.
      </p>
      {flights.length === 0 ? (
        <p className="center">No flights in this range.</p>
      ) : (
        <div className="tablewrap">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>From</th>
                <th>To</th>
                <th>Off</th>
                <th>On</th>
                <th className="num">Block</th>
                <th className="num">Air</th>
              </tr>
            </thead>
            <tbody>
              {flights.map((f) => {
                const air = airMinutes(f.takeoff_utc, f.landing_utc)
                return (
                  <tr key={f.seq}>
                    <td>{f.date}</td>
                    <td>{f.dep_place}</td>
                    <td>{f.arr_place}</td>
                    <td>{clock(f.off_block_utc)}</td>
                    <td>{clock(f.on_block_utc)}</td>
                    <td className="num">{hhmm(f.block_minutes)}</td>
                    <td className="num">{air === null ? '' : hhmm(air)}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
      <button type="button" onClick={onClose} style={{ marginTop: 12 }}>
        Close
      </button>
    </div>
  )
}

function Figure({
  label,
  value,
  minutes,
}: {
  label: string
  value: string
  minutes?: number
}) {
  return (
    <div className="figure">
      <div className="label">{label}</div>
      <div className="value">{value}</div>
      {/* Both units, at the owner's instruction: an invoice is checked in H:MM
          and computed in minutes. */}
      {minutes !== undefined && <div className="muted small">{minutes} min</div>}
    </div>
  )
}
