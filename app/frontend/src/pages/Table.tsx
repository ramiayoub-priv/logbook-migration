import { useCallback, useMemo, useState } from 'react'
import { api, type DateRange } from '../api'
import { useApi } from '../auth'
import { airMinutes, clock, hhmm, hhmmOrBlank } from '../format'
import { Link } from '../router'
import { RangePicker } from './RangePicker'

export function TablePage() {
  const [range, setRange] = useState<DateRange>({})
  const load = useCallback(() => api.flights(range), [range.from, range.to])
  const { data, error, loading } = useApi(load, [range.from, range.to])

  /**
   * Newest first, because the flight you are looking for is almost always the
   * one you just flew.
   *
   * This is a VIEW concern and it lives here rather than in the query. The
   * server returns the book's own seq order -- oldest first -- and that order
   * is load-bearing: every cumulative figure and the whole EASA export are
   * built on it (rule 0.5), and there is exactly one `ORDER BY seq` feeding
   * the table, the statistics and all three PDFs. Reversing it there would
   * have reversed the export too.
   *
   * It reverses BOOK order, not date order. 21 rows across the three books are
   * genuinely out of date order -- including the three 28/08/2025 late entries
   * that sit at the end of Book 3 -- so sorting on the date would quietly move
   * rows out of the order the paper keeps them in.
   */
  const rows = useMemo(() => (data ? [...data.flights].reverse() : []), [data])

  return (
    <>
      <RangePicker range={range} onChange={setRange} />

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}

      {loading && !data && <p className="center">Loading your flights…</p>}

      {data && (
        <>
          <p className="muted small">
            {data.count} flight{data.count === 1 ? '' : 's'}, newest first.
          </p>

          {data.count === 0 ? (
            <p className="center">No flights in this range.</p>
          ) : (
            <div className="card" style={{ padding: 0 }}>
              <div className="tablewrap">
                <table>
                  <thead>
                    <tr>
                      <th>Date</th>
                      <th>Aircraft</th>
                      <th>From</th>
                      <th>To</th>
                      <th>Off</th>
                      <th>On</th>
                      <th className="num">Total</th>
                      {/* The airborne group. It sits after the block group
                          rather than interleaved, so the two pairs cannot be
                          misread for one another at a glance -- off/on block
                          and takeoff/landing are four times in the same
                          format, and the aircraft's logbook is filled from
                          only one of the pairs. */}
                      <th>Takeoff</th>
                      <th>Landing</th>
                      <th className="num">Air</th>
                      <th className="num">PIC</th>
                      <th className="num">Dual</th>
                      <th className="num">Night</th>
                      <th className="num">Instr</th>
                      <th className="num">Instructor</th>
                      <th className="num">Ldg D/N</th>
                      <th>Remarks</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {rows.map((f) => (
                      <tr key={f.seq}>
                        <td>{f.date}</td>
                        <td>
                          {f.aircraft_reg}
                          <span className="muted"> {f.aircraft_type}</span>
                          {isSea(f.class) && <span className="muted"> · sea</span>}
                        </td>
                        <td>{f.dep_place}</td>
                        <td>{f.arr_place}</td>
                        <td>{clock(f.off_block_utc)}</td>
                        <td>{clock(f.on_block_utc)}</td>
                        <td className="num">{hhmm(f.total_minutes)}</td>
                        {/* Blank, not 0:00, when the row has no airborne pair.
                            Only 19 of the 1296 transcribed rows carry one, and
                            a zero would say the aeroplane never left the
                            ground. Air time is computed here, at render, from
                            the two instants -- never stored. */}
                        <td>{clock(f.takeoff_utc)}</td>
                        <td>{clock(f.landing_utc)}</td>
                        <td className="num">{airTime(f.takeoff_utc, f.landing_utc)}</td>
                        <td className="num">{hhmmOrBlank(f.pic_minutes)}</td>
                        <td className="num">{hhmmOrBlank(f.dual_minutes)}</td>
                        <td className="num">{hhmmOrBlank(f.night_minutes)}</td>
                        <td className="num">{hhmmOrBlank(f.instrument_minutes)}</td>
                        <td className="num">{hhmmOrBlank(f.instructor_minutes)}</td>
                        <td className="num">
                          {f.landings_day}/{f.landings_night}
                          {/* The day/night split on 30 rows was inferred from the
                              presence of night time rather than read off the page
                              (Task 8). The table says so rather than showing it
                              like a figure somebody checked. */}
                          {!f.landings_verified && (
                            <span
                              className="flag"
                              title="The day/night split on this row was inferred, not read from the paper page."
                            >
                              *
                            </span>
                          )}
                        </td>
                        <td>{f.remarks}</td>
                        <td>
                          {/*
                            Only flights entered in the app can be corrected.
                            The 1296 transcribed rows are closed data (rule
                            0.8) and the importer owns them, so an Edit link on
                            one would offer something the server refuses. No
                            link is the honest answer; the edit page explains
                            it in full for anyone who reaches it by URL.
                          */}
                          {f.source_book === 0 && (
                            <Link to="edit" seq={f.seq} className="rowedit">
                              Edit<span className="visually-hidden">
                                {' '}
                                the flight of {f.date} in {f.aircraft_reg}
                              </span>
                            </Link>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          <p className="muted small">
            All times UTC. <strong>Total</strong> is chocks to chocks;{' '}
            <strong>Air</strong> is wheels up to wheels down, worked out from the takeoff and
            landing times and blank where the flight has none.{' '}
            <span className="flag">*</span> marks a day/night landing split that was inferred
            rather than read from the paper page.
          </p>
        </>
      )}
    </>
  )
}

function isSea(cls: string): boolean {
  return cls === 'SEP_SEA' || cls === 'MEP_SEA'
}

/** airTime renders the derived airborne time, or nothing when it is unknown. */
function airTime(takeoff: string | null, landing: string | null): string {
  const m = airMinutes(takeoff, landing)
  return m === null ? '' : hhmm(m)
}
