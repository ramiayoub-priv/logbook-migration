import { useEffect, useId, useMemo, useRef, useState } from 'react'
import { api, ApiError, type Aircraft } from '../api'
import { Field } from './FlightForm'

/**
 * The aircraft picker: a filterable combobox that can also add an aeroplane.
 *
 * WHY IT IS NOT A <select>. It used to be one, fed by a list derived from the
 * flights — so the only aeroplanes that could be chosen were the ones already
 * flown, and the first flight in a new aeroplane was unenterable. The owner
 * asked for two things on 2026-08-02 and this control is both of them:
 *
 *   1. Type to filter. "If I write P it should filter OH-PDP, OH-PIF and so
 *      on." The type matches too, because typing C172 is as natural as typing
 *      a registration.
 *   2. Add a new aeroplane without leaving the form. Adding it on another
 *      screen first would mean two trips on exactly the day you are standing
 *      at an airfield with a phone.
 *
 * NOTHING IS EVER HIDDEN. The owner dropped the retired/active idea in the same
 * ruling — an aeroplane you flew once in 2009 is not retired — so the full list
 * is always reachable and filtering, not hiding, is what keeps it usable. The
 * order comes from the server (never-flown first, then most recently flown) and
 * is deliberately NOT re-sorted here: one authority for it.
 */
export function AircraftPicker({
  value,
  fleet,
  error,
  onChoose,
  onAdded,
}: {
  value: string
  fleet: Aircraft[]
  error?: string | undefined
  /** Called with the registration and, when known, the aeroplane behind it. */
  onChoose: (reg: string, a?: Aircraft) => void
  /** A newly created aeroplane, so the caller can fold it into its own list. */
  onAdded: (a: Aircraft) => void
}) {
  // What is typed in the box. It tracks `value` until the pilot starts
  // filtering, at which point the two deliberately diverge: the text is a
  // query, the value is the choice.
  const [query, setQuery] = useState(value)
  const [open, setOpen] = useState(false)
  const [adding, setAdding] = useState<string | null>(null)

  const wrapRef = useRef<HTMLDivElement>(null)
  const listId = useId()

  useEffect(() => setQuery(value), [value])

  // Closing on an outside click rather than on blur: blur fires before the
  // click on an option lands, which would close the list out from under the
  // finger that was choosing.
  useEffect(() => {
    if (!open) return
    function onDown(e: MouseEvent | TouchEvent) {
      if (!wrapRef.current?.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDown)
    document.addEventListener('touchstart', onDown)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('touchstart', onDown)
    }
  }, [open])

  const q = query.trim().toUpperCase()
  const matches = useMemo(() => {
    if (q === '') return fleet
    return fleet.filter(
      (a) => a.registration.toUpperCase().includes(q) || a.type.toUpperCase().includes(q),
    )
  }, [fleet, q])

  // Offering to add is driven by "nothing matched what you typed", not by a
  // button the pilot has to know exists.
  const canAdd = q !== '' && !fleet.some((a) => a.registration.toUpperCase() === q)

  function choose(a: Aircraft) {
    setQuery(a.registration)
    setOpen(false)
    onChoose(a.registration, a)
  }

  return (
    <div className="aircraftpicker" ref={wrapRef}>
      <Field id="aircraft_reg" label="Aircraft" error={error}>
        <input
          id="aircraft_reg"
          role="combobox"
          aria-expanded={open}
          aria-controls={listId}
          aria-autocomplete="list"
          autoComplete="off"
          autoCapitalize="characters"
          spellCheck={false}
          placeholder="Registration or type"
          value={query}
          aria-invalid={!!error}
          onFocus={() => setOpen(true)}
          onClick={() => setOpen(true)}
          onChange={(e) => {
            setQuery(e.target.value)
            setOpen(true)
            setAdding(null)
            // The typed text IS the registration until something is chosen.
            // An edited flight naming an aeroplane no longer in the list must
            // not have its registration silently cleared.
            onChoose(e.target.value.toUpperCase().trim())
          }}
          onKeyDown={(e) => {
            if (e.key === 'Escape') setOpen(false)
            // Enter picks the only remaining match, which is what a filter is
            // for: type three characters, press go.
            const only = matches.length === 1 ? matches[0] : undefined
            if (e.key === 'Enter' && open && only) {
              e.preventDefault()
              choose(only)
            }
          }}
        />
      </Field>

      {open && (
        <ul className="options" id={listId} role="listbox" aria-label="Aircraft">
          {matches.map((a) => (
            <li key={a.registration}>
              <button type="button" role="option" aria-selected={a.registration === value}
                onClick={() => choose(a)}>
                <span className="reg">{a.registration}</span>
                <span className="type">{a.type}</span>
                {/* What each aeroplane says about itself, so the list is
                    scannable without hiding anything. */}
                <span className="when">
                  {a.flights === 0 ? 'not flown yet' : `${a.flights} flights · ${a.last_flown}`}
                </span>
              </button>
            </li>
          ))}
          {canAdd && (
            <li>
              <button type="button" role="option" aria-selected={false} className="addnew"
                onClick={() => {
                  setAdding(q)
                  setOpen(false)
                }}>
                Add {q} as a new aircraft
              </button>
            </li>
          )}
          {matches.length === 0 && !canAdd && <li className="empty">No aircraft match.</li>}
        </ul>
      )}

      {adding !== null && (
        <NewAircraft
          registration={adding}
          onCancel={() => setAdding(null)}
          onCreated={(a) => {
            setAdding(null)
            setQuery(a.registration)
            onAdded(a)
            onChoose(a.registration, a)
          }}
        />
      )}
    </div>
  )
}

const CLASSES = [
  ['SEP_LAND', 'SEP land'],
  ['SEP_SEA', 'SEP sea'],
  ['MEP_LAND', 'MEP land'],
  ['MEP_SEA', 'MEP sea'],
  ['TMG', 'TMG'],
] as const

/**
 * The inline "add an aeroplane" panel.
 *
 * It asks for the two things a flight cannot do without — the type and the
 * class — and nothing else. Notes and IFR capability are not worth a field on a
 * phone at an airfield; they can be filled in later on the fleet page.
 *
 * THE CLASS MATTERS MORE THAN IT LOOKS. It seeds the flight's own class, which
 * is what decides whether the flight counts towards the seaplane rating. It
 * stays a default: the flight's class field is editable underneath, because the
 * same registration can be flown on floats or wheels.
 */
function NewAircraft({
  registration,
  onCancel,
  onCreated,
}: {
  registration: string
  onCancel: () => void
  onCreated: (a: Aircraft) => void
}) {
  const [type, setType] = useState('')
  const [cls, setCls] = useState('SEP_LAND')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const typeRef = useRef<HTMLInputElement>(null)

  useEffect(() => typeRef.current?.focus(), [])

  async function save() {
    const t = type.trim().toUpperCase()
    if (t === '') {
      setError('An aircraft type is required.')
      typeRef.current?.focus()
      return
    }
    setBusy(true)
    setError(null)
    try {
      const { aircraft } = await api.createAircraft({
        registration,
        type: t,
        default_class: cls,
        ifr_capable: false,
        notes: '',
      })
      onCreated(aircraft)
    } catch (e) {
      // The message is the server's, because it is the one that knows why --
      // "that registration is already in the aircraft list" is the common case
      // and is far more use than "could not save".
      setError(e instanceof ApiError ? e.message : 'Could not add the aircraft.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="newaircraft">
      <p className="lead">
        New aircraft <strong>{registration}</strong>
      </p>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <div className="row">
        <Field id="new_aircraft_type" label="New aircraft type">
          <input
            id="new_aircraft_type"
            ref={typeRef}
            value={type}
            onChange={(e) => setType(e.target.value)}
            autoCapitalize="characters"
            spellCheck={false}
            placeholder="C152"
          />
        </Field>
        <Field id="new_aircraft_class" label="New aircraft class">
          <select id="new_aircraft_class" value={cls} onChange={(e) => setCls(e.target.value)}>
            {CLASSES.map(([v, label]) => (
              <option key={v} value={v}>
                {label}
              </option>
            ))}
          </select>
        </Field>
      </div>
      <div className="actions">
        <button type="button" onClick={save} disabled={busy}>
          {busy ? 'Saving…' : 'Save aircraft'}
        </button>
        <button type="button" className="secondary" onClick={onCancel} disabled={busy}>
          Cancel
        </button>
      </div>
    </div>
  )
}
