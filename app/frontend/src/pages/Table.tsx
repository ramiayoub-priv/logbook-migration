import { useCallback, useState } from 'react'
import { api, type DateRange } from '../api'
import { useApi } from '../auth'
import { clock, hhmm, hhmmOrBlank } from '../format'
import { RangePicker } from './RangePicker'

export function TablePage() {
  const [range, setRange] = useState<DateRange>({})
  const load = useCallback(() => api.flights(range), [range.from, range.to])
  const { data, error, loading } = useApi(load, [range.from, range.to])

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
            {data.count} flight{data.count === 1 ? '' : 's'}, in book order.
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
                      <th className="num">PIC</th>
                      <th className="num">Dual</th>
                      <th className="num">Night</th>
                      <th className="num">Instr</th>
                      <th className="num">Instructor</th>
                      <th className="num">Ldg D/N</th>
                      <th>Remarks</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.flights.map((f) => (
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
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          <p className="muted small">
            All times UTC. <span className="flag">*</span> marks a day/night landing split
            that was inferred rather than read from the paper page.
          </p>
        </>
      )}
    </>
  )
}

function isSea(cls: string): boolean {
  return cls === 'SEP_SEA' || cls === 'MEP_SEA'
}
