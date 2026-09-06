import { useState } from 'react'
import { api, type DateRange } from '../api'
import { RangePicker } from './RangePicker'

export function ExportPage() {
  const [range, setRange] = useState<DateRange>({})

  return (
    <>
      <RangePicker range={range} onChange={setRange} />

      <div className="card">
        <h2>EASA logbook</h2>
        <p className="muted small">
          A reproduction of the whole logbook in EASA format — all three paper books as one
          continuous record, fifteen flights to a page with the page totals block. This is the
          document to hand to an authority.
        </p>
        <p className="muted small">
          <strong>It always covers every flight</strong>, whatever range is set above: a partial
          logbook would understate a licence total.
        </p>
        {/* A plain link, not a fetch. The browser sends the session cookie and
            honours Content-Disposition, which is what produces a real saved
            file rather than a blob held in the page. */}
        <a className="primary" href={api.exportURL('easa')} role="button"
           style={{ display: 'block', textAlign: 'center', textDecoration: 'none',
                    padding: '12px 16px', borderRadius: 10 }}>
          Download the EASA logbook
        </a>
      </div>

      <div className="card">
        <h2>Flight table</h2>
        <p className="muted small">
          Every flight in the selected range with everything the application knows about it:
          the block times and the takeoff/landing pair with the airborne time derived from it,
          where each row came from in the paper books, and whether its landing split was read
          or inferred. The same columns as the table on screen.
        </p>
        <a href={api.exportURL('table', range)}>Download the flight table (PDF)</a>
      </div>

      <div className="card">
        <h2>Statistics</h2>
        <p className="muted small">
          The totals for the selected range: time by function, the seaplane and landplane
          split, and landings.
        </p>
        <a href={api.exportURL('statistics', range)}>Download the statistics sheet (PDF)</a>
      </div>
    </>
  )
}
