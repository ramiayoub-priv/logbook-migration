import { useCallback } from 'react'
import { api } from '../api'
import { useApi } from '../auth'

/**
 * The "needs review" list.
 *
 * This page is the visible half of CLAUDE.md rule 0.2: the importer surfaces
 * every disagreement it finds in the source data instead of correcting it, and
 * this is where they surface to. The table is rewritten on every import, so an
 * item resolved in the CSVs simply disappears -- there is no second place to
 * tick anything off, and nothing here can be dismissed from the app.
 */
export function ReviewPage() {
  const load = useCallback(() => api.discrepancies(), [])
  const { data, error, loading } = useApi(load, [])

  return (
    <>
      <div className="card">
        <h2>Needs review</h2>
        <p className="muted small">
          Everything the importer found in the paper transcription that it would not act on by
          itself. These are reported, never corrected automatically — the paper book is
          authoritative and only you can rule on a disagreement.
        </p>
        <p className="muted small">
          The list is rebuilt from the CSVs on every import. Fix a row at the source and it
          leaves this page.
        </p>
      </div>

      {error && <p className="error" role="alert">{error}</p>}
      {loading && !data && <p className="center">Loading the review list…</p>}

      {data && data.count === 0 && <p className="center">Nothing needs review.</p>}

      {data && data.count > 0 && (
        <>
          <p className="muted small">{data.count} item{data.count === 1 ? '' : 's'}.</p>
          {Object.entries(groupByKind(data.discrepancies)).map(([kind, items]) => (
            <div className="card" key={kind} style={{ padding: 0 }}>
              <div style={{ padding: '12px 14px 0' }}>
                <h2 style={{ marginBottom: 4 }}>{humanKind(kind)}</h2>
                <p className="muted small">{items.length} row{items.length === 1 ? '' : 's'}</p>
              </div>
              <div className="tablewrap">
                <table>
                  <thead>
                    <tr>
                      <th>Source</th>
                      <th>Date</th>
                      <th>Detail</th>
                    </tr>
                  </thead>
                  <tbody>
                    {items.map((d, i) => (
                      <tr key={`${d.book}-${d.row}-${i}`}>
                        <td>book {d.book}, line {d.row}</td>
                        <td>{d.date}</td>
                        <td style={{ whiteSpace: 'normal' }}>{d.detail}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ))}
        </>
      )}
    </>
  )
}

function groupByKind<T extends { kind: string }>(items: T[]): Record<string, T[]> {
  const out: Record<string, T[]> = {}
  for (const item of items) (out[item.kind] ??= []).push(item)
  return out
}

/** humanKind turns the importer's kind into something readable, unchanged if new. */
function humanKind(kind: string): string {
  const known: Record<string, string> = {
    landings_unverified: 'Landing split inferred, not read from the page',
    cumulative_break: 'A cumulative column disagrees with the rows',
    component_exceeds_total: 'A component time exceeds the flight total',
    block_total_mismatch: 'Block time and total time disagree',
    unknown_time_origin: 'A clock time could not be resolved to UTC',
    registration_format: 'Registration is not in the Finnish OH- format',
    unknown_aircraft_type: 'Aircraft type outside the known set',
    type_conflict: 'One registration written with two different types',
    date_format: 'Date written with dots rather than slashes',
  }
  return known[kind] ?? kind
}
