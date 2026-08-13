import { useCallback, useEffect, useRef, useState } from 'react'
import { api, ApiError, type Aircraft, type AircraftDraft } from '../api'
import { useApi } from '../auth'
import { Field } from './FlightForm'

/**
 * The fleet page: every aeroplane, and the only place one can be corrected.
 *
 * WHY IT EXISTS (Task 19, owner ask 2026-08-03). `POST` and `PUT /aircraft`
 * shipped on 2026-08-02 and the picker on the flight form used only the first
 * of them, so `api.updateAircraft` had zero callers. A registration typed
 * wrongly at an airfield could be created and then never corrected -- and by
 * the owner's ruling it cannot be deleted either. **The no-delete ruling is
 * only humane if editing exists**, which is what this page is for.
 *
 * WHY IT IS NOT A SEVENTH TAB. Six already share a 390px phone, and the label
 * "Statistics" was cut to "Stats" to fit the sixth. This page is opened rarely
 * -- once per new aeroplane, once per typo -- so it hangs off the Aircraft tab
 * and has its own URL, instead of taking permanent space from pages used on
 * every flight.
 *
 * IT CANNOT MOVE A TOTAL. The aircraft table is a seed list, not the record:
 * every flight carries its own registration, type and class exactly as written
 * on paper. That is asserted on the backend by reading every flight back after
 * an edit, and it is why editing the 38 aeroplanes from the paper books is
 * allowed here without breaching rule 0.8.
 */
export function FleetPage() {
  const load = useCallback(() => api.aircraft(), [])
  const { data, error, loading } = useApi(load, [])

  // The list is held here rather than re-fetched after every write. The
  // server's order -- never-flown first, then most recently flown -- is the one
  // authority for ordering (it lives in its SQL), so a write splices its result
  // into the list in place instead of re-sorting anything locally. A newly
  // added aeroplane goes on the front, which is the same rule: it has not been
  // flown at all.
  const [fleet, setFleet] = useState<Aircraft[] | null>(null)
  useEffect(() => {
    if (data) setFleet(data.aircraft)
  }, [data])

  // Exactly one form is open at a time, by construction: opening either closes
  // the other. Two open forms would mean two controls labelled "Registration".
  const [adding, setAdding] = useState(false)
  const [editing, setEditing] = useState<string | null>(null)

  return (
    <>
      <div className="card">
        <h2>Fleet</h2>
        <p className="muted small">
          Every aeroplane the logbook knows about — the {fleet?.filter((a) => !a.user_added).length ?? 38}{' '}
          from the paper books and any added since. This list only seeds the new-flight form:
          each flight records the registration, type and class it was flown under, so nothing
          changed here can move a total.
        </p>
        <button
          onClick={() => {
            setEditing(null)
            setAdding(true)
          }}
        >
          Add an aircraft
        </button>
      </div>

      {error && <p className="error" role="alert">{error}</p>}
      {loading && !fleet && <p className="center">Loading the fleet…</p>}

      {adding && (
        <AircraftForm
          title="New aircraft"
          onCancel={() => setAdding(false)}
          onSave={(draft) => api.createAircraft(draft)}
          onSaved={(a) => {
            setAdding(false)
            setFleet((list) => [a, ...(list ?? []).filter((x) => x.registration !== a.registration)])
          }}
        />
      )}

      {fleet && (
        <ul className="fleet" aria-label="Fleet">
          {fleet.map((a) => (
            <li key={a.registration}>
              {editing === a.registration ? (
                <AircraftForm
                  title={`Correcting ${a.registration}`}
                  aircraft={a}
                  onCancel={() => setEditing(null)}
                  // Keyed by the registration as it stands, because that is what
                  // the route names. Renaming an aeroplane is the whole point of
                  // the U, so the key and the payload deliberately differ.
                  onSave={(draft) => api.updateAircraft(a.registration, draft)}
                  onSaved={(saved) => {
                    setEditing(null)
                    setFleet((list) =>
                      (list ?? []).map((x) => (x.registration === a.registration ? saved : x)),
                    )
                  }}
                />
              ) : (
                <div className="entry">
                  <div>
                    <strong className="reg">{a.registration}</strong>{' '}
                    <span className="type">{a.type}</span>
                    <br />
                    <span className="muted small">
                      {classLabel(a.default_class)}
                      {a.ifr_capable && ' · IFR'}
                      {' · '}
                      {a.flights === 0
                        ? 'not flown yet'
                        : `${a.flights} flights · last flown ${a.last_flown}`}
                      {a.user_added && ' · added here'}
                      {a.notes && ` · ${a.notes}`}
                    </span>
                  </div>
                  <button
                    className="link"
                    onClick={() => {
                      setAdding(false)
                      setEditing(a.registration)
                    }}
                  >
                    Edit
                  </button>
                </div>
              )}
            </li>
          ))}
        </ul>
      )}
    </>
  )
}

const CLASSES = [
  ['SEP_LAND', 'SEP land'],
  ['SEP_SEA', 'SEP sea'],
  ['MEP_LAND', 'MEP land'],
  ['MEP_SEA', 'MEP sea'],
  ['TMG', 'TMG'],
] as const

function classLabel(c: string): string {
  return CLASSES.find(([v]) => v === c)?.[1] ?? c
}

/**
 * One form, used to add and to correct.
 *
 * Both endpoints take the same complete aeroplane -- a `PUT` here is a
 * replacement, not a patch -- so one form for both is not a shortcut: two would
 * be two chances to send a partial aeroplane to an endpoint that treats a
 * missing field as an empty one.
 *
 * It asks for more than the flight form's inline panel does (which asks only
 * for type and class, because it is used standing at an airfield). Notes and
 * IFR capability belong somewhere, and this is the somewhere.
 */
function AircraftForm({
  title,
  aircraft,
  onSave,
  onSaved,
  onCancel,
}: {
  title: string
  aircraft?: Aircraft
  onSave: (draft: AircraftDraft) => Promise<{ aircraft: Aircraft }>
  onSaved: (a: Aircraft) => void
  onCancel: () => void
}) {
  const [reg, setReg] = useState(aircraft?.registration ?? '')
  const [type, setType] = useState(aircraft?.type ?? '')
  const [cls, setCls] = useState(aircraft?.default_class ?? 'SEP_LAND')
  const [ifr, setIfr] = useState(aircraft?.ifr_capable ?? false)
  const [notes, setNotes] = useState(aircraft?.notes ?? '')
  const [problem, setProblem] = useState<string | null>(null)
  const [fields, setFields] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState(false)
  const regRef = useRef<HTMLInputElement>(null)

  useEffect(() => regRef.current?.focus(), [])

  async function save() {
    // Upper-cased and trimmed before it is sent, the same way the picker does
    // it. A registration is an identifier: "oh-xyz" must not be able to become
    // a second row alongside OH-XYZ. The server normalises too -- this is so
    // what is sent matches what will come back.
    const draft: AircraftDraft = {
      registration: reg.trim().toUpperCase(),
      type: type.trim().toUpperCase(),
      default_class: cls,
      ifr_capable: ifr,
      notes: notes.trim(),
    }
    setBusy(true)
    setProblem(null)
    setFields({})
    try {
      const { aircraft: saved } = await onSave(draft)
      onSaved(saved)
    } catch (e) {
      // The server's own message: "that registration is already in the aircraft
      // list" is the common refusal and says far more than "could not save".
      setProblem(e instanceof ApiError ? e.message : 'Could not save the aircraft.')
      if (e instanceof ApiError) {
        setFields(Object.fromEntries(e.fields.map((f) => [f.field, f.message])))
      }
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="card aircraftform">
      <h3>{title}</h3>
      {problem && (
        <p className="error" role="alert">
          {problem}
        </p>
      )}
      <Field id="fleet_reg" label="Registration" error={fields['registration']}>
        <input
          id="fleet_reg"
          ref={regRef}
          value={reg}
          onChange={(e) => setReg(e.target.value)}
          autoCapitalize="characters"
          spellCheck={false}
          autoComplete="off"
          placeholder="OH-CTL"
        />
      </Field>
      <div className="row">
        <Field id="fleet_type" label="Type" error={fields['type']}>
          <input
            id="fleet_type"
            value={type}
            onChange={(e) => setType(e.target.value)}
            autoCapitalize="characters"
            spellCheck={false}
            autoComplete="off"
            placeholder="C172"
          />
        </Field>
        <Field id="fleet_class" label="Class" error={fields['default_class']}>
          <select id="fleet_class" value={cls} onChange={(e) => setCls(e.target.value)}>
            {CLASSES.map(([v, label]) => (
              <option key={v} value={v}>
                {label}
              </option>
            ))}
          </select>
        </Field>
      </div>
      <div className="field checkbox">
        <input
          id="fleet_ifr"
          type="checkbox"
          checked={ifr}
          onChange={(e) => setIfr(e.target.checked)}
        />
        <label htmlFor="fleet_ifr">IFR capable</label>
      </div>
      <Field id="fleet_notes" label="Notes">
        <input id="fleet_notes" value={notes} onChange={(e) => setNotes(e.target.value)} />
      </Field>
      <p className="muted small">
        The class is a default for the new-flight form and never constrains what a flight may
        record: the same registration can be flown on floats or on wheels.
      </p>
      <div className="actions">
        <button type="button" onClick={() => void save()} disabled={busy}>
          {busy ? 'Saving…' : 'Save aircraft'}
        </button>
        <button type="button" className="secondary" onClick={onCancel} disabled={busy}>
          Cancel
        </button>
      </div>
    </div>
  )
}
