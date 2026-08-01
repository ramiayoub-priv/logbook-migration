import type { DateRange } from '../api'

/**
 * The From–To control shared by the table, statistics and export pages.
 *
 * An empty bound means "open", which is how the API reads a missing parameter,
 * so clearing a field widens the range rather than erroring. An inverted range
 * is allowed through deliberately: the server answers it with an empty result,
 * which is the honest answer to "flights between December and January", and it
 * is easier to see that and fix it than to be blocked by a validation message.
 */
export function RangePicker({
  range,
  onChange,
  children,
}: {
  range: DateRange
  onChange: (r: DateRange) => void
  children?: React.ReactNode
}) {
  const inverted = range.from && range.to && range.from > range.to

  return (
    <div className="card">
      <div className="row">
        <div className="field" style={{ marginBottom: 0 }}>
          <label htmlFor="from">From</label>
          <input
            id="from"
            type="date"
            value={range.from ?? ''}
            onChange={(e) => onChange({ ...range, from: e.target.value || undefined })}
          />
        </div>
        <div className="field" style={{ marginBottom: 0 }}>
          <label htmlFor="to">To</label>
          <input
            id="to"
            type="date"
            value={range.to ?? ''}
            onChange={(e) => onChange({ ...range, to: e.target.value || undefined })}
          />
        </div>
      </div>

      {inverted && (
        <p className="warn small" style={{ marginTop: 10, marginBottom: 0 }}>
          The start is after the end, so this range contains nothing.
        </p>
      )}

      {(range.from || range.to) && (
        <button
          className="link"
          style={{ marginTop: 8 }}
          onClick={() => onChange({})}
          type="button"
        >
          Clear the range
        </button>
      )}

      {children}
    </div>
  )
}
