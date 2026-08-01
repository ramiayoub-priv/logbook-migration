import { useCallback, useState } from 'react'
import { api, type DateRange, type Summary } from '../api'
import { useApi } from '../auth'
import { hhmm } from '../format'
import { RangePicker } from './RangePicker'

export function StatisticsPage() {
  const [range, setRange] = useState<DateRange>({})
  const load = useCallback(() => api.stats(range), [range.from, range.to])
  const { data, error, loading } = useApi(load, [range.from, range.to])

  return (
    <>
      <RangePicker range={range} onChange={setRange} />

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {loading && !data && <p className="center">Adding it all up…</p>}

      {data && <Figures s={data.summary} />}
    </>
  )
}

function Figures({ s }: { s: Summary }) {
  return (
    <>
      <h3>Totals</h3>
      <div className="figures">
        <Figure label="Flights" value={String(s.flights)} />
        <Figure label="Total time" value={hhmm(s.total)} />
        <Figure label="PIC" value={hhmm(s.pic)} />
        <Figure label="Dual" value={hhmm(s.dual)} />
        <Figure label="Night" value={hhmm(s.night)} />
        <Figure label="Instrument" value={hhmm(s.instrument)} />
        <Figure label="Instructor" value={hhmm(s.instructor)} />
      </div>

      <h3>Seaplane</h3>
      <div className="figures">
        <Figure label="Seaplane total" value={hhmm(s.sea_total)} />
        <Figure label="Seaplane PIC" value={hhmm(s.sea_pic)} />
        <Figure label="Seaplane instructor" value={hhmm(s.sea_instructor)} />
      </div>

      <h3>Landplane</h3>
      <div className="figures">
        <Figure label="Landplane total" value={hhmm(s.land_total)} />
        <Figure label="Landplane PIC" value={hhmm(s.land_pic)} />
        <Figure label="Landplane instructor" value={hhmm(s.land_instructor)} />
      </div>

      <h3>Landings</h3>
      <div className="figures">
        <Figure label="Day" value={String(s.landings_day)} />
        <Figure label="Night" value={String(s.landings_night)} />
        <Figure label="Seaplane" value={String(s.landings_sea)} />
        <Figure label="Landplane" value={String(s.landings_land)} />
      </div>

      {/* Surfaced, never hidden. The night-landing figure above is partly
          inferred, and a statistics page that stated it without saying so
          would be claiming a verification nobody has done (rule 0.2, Task 8). */}
      {s.landings_unverified > 0 && (
        <p className="warn small">
          {s.landings_unverified} flight{s.landings_unverified === 1 ? '' : 's'} in this range
          carry a day/night landing split that was inferred from the presence of night time
          rather than read from the paper page. The landing total is unaffected; the split
          between day and night is provisional.
        </p>
      )}
    </>
  )
}

function Figure({ label, value }: { label: string; value: string }) {
  return (
    <div className="figure">
      <div className="label">{label}</div>
      <div className="value">{value}</div>
    </div>
  )
}
