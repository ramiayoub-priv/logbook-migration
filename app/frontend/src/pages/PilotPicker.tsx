import { useEffect, useMemo, useState } from 'react'
import { api, ApiError, type Pilot } from '../api'
import { useCombobox } from './combobox'
import { Field } from './FlightForm'

/**
 * The pilot-in-command picker: the same treatment the registrations got.
 *
 * WHY (owner ask, 2026-08-03): *"I could have a typo when I write `self`, it
 * could be `sself` or `SELF` or `seeelf` and I need it to be consistent (like
 * the aircraft regs)."* The field was free text, and `self` is on 1143 of the
 * 1296 transcribed flights — so a stray keystroke would split the busiest value
 * in the column into two people, and nothing would ever notice.
 *
 * THE LIST IS MOSTLY THE RECORD'S OWN. It is every distinct `pic_name` already
 * on a flight, counted, plus any name added here and not yet flown with. So it
 * needs no maintenance and cannot fall behind. The order is the server's —
 * never flown with first, then most recent — and is deliberately not re-sorted
 * here: two authorities for it would disagree.
 *
 * IT DOES NOT TIDY THE PAPER. The books contain `Sinervä` and `Sinerva`, and
 * `Stude`, which reads as a word rather than a surname. All of them are offered
 * exactly as written. Merging them would be this component deciding a question
 * that is the owner's alone (rule 0.8); what it prevents is a NEW spelling.
 *
 * TYPING STILL SETS THE VALUE, as in the aircraft picker. A flight being edited
 * may name somebody the roster no longer offers, and having the field silently
 * blank itself would be a legal record losing a name because a dropdown had an
 * opinion.
 */
export function PilotPicker({
  value,
  pilots,
  error,
  onChoose,
  onAdded,
}: {
  value: string
  pilots: Pilot[]
  error?: string | undefined
  onChoose: (name: string) => void
  /** A newly created name, so the caller can fold it into its own list. */
  onAdded: (p: Pilot) => void
}) {
  // What is typed tracks the value until filtering starts, at which point the
  // two deliberately diverge: the text is a query, the value is the choice.
  const [query, setQuery] = useState(value)
  const [problem, setProblem] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const { open, setOpen, wrapRef, listId } = useCombobox()

  useEffect(() => setQuery(value), [value])

  const q = query.trim().toLowerCase()
  const matches = useMemo(
    () => (q === '' ? pilots : pilots.filter((p) => p.name.toLowerCase().includes(q))),
    [pilots, q],
  )

  // Offering to add is driven by "nothing you typed is already a name", not by
  // a button the pilot has to know exists. Compared case-insensitively, because
  // the server refuses a case variant and offering to add one would be an
  // invitation to a refusal.
  const canAdd = q !== '' && !pilots.some((p) => p.name.toLowerCase() === q)

  function choose(name: string) {
    setQuery(name)
    setOpen(false)
    setProblem(null)
    onChoose(name)
  }

  async function add() {
    const name = query.trim()
    setBusy(true)
    setProblem(null)
    try {
      const { pilot } = await api.createPilot(name)
      onAdded(pilot)
      choose(pilot.name)
    } catch (e) {
      // The server's own words. "That name is already in the pilot list" is the
      // common refusal, it is the whole point of the feature, and it tells the
      // pilot to go and pick it rather than that something broke.
      setProblem(e instanceof ApiError ? e.message : 'Could not add that name.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="pilotpicker" ref={wrapRef}>
      <Field id="pic_name" label="Name of pilot in command" error={error}>
        <input
          id="pic_name"
          role="combobox"
          aria-expanded={open}
          aria-controls={listId}
          aria-autocomplete="list"
          autoComplete="off"
          spellCheck={false}
          placeholder="self"
          value={query}
          aria-invalid={!!error}
          onFocus={() => setOpen(true)}
          onClick={() => setOpen(true)}
          onChange={(e) => {
            setQuery(e.target.value)
            setOpen(true)
            setProblem(null)
            // The typed text IS the name until something is chosen: an edited
            // flight naming somebody no longer on the roster must not have that
            // name silently cleared.
            onChoose(e.target.value.trim())
          }}
          onKeyDown={(e) => {
            if (e.key === 'Escape') setOpen(false)
            const only = matches.length === 1 ? matches[0] : undefined
            if (e.key === 'Enter' && open && only) {
              e.preventDefault()
              choose(only.name)
            }
          }}
        />
      </Field>

      {problem && (
        <p className="error small" role="alert">
          {problem}
        </p>
      )}

      {open && (
        <ul className="options" id={listId} role="listbox" aria-label="Pilots">
          {matches.map((p) => (
            <li key={p.name}>
              <button type="button" role="option" aria-selected={p.name === value}
                onClick={() => choose(p.name)}>
                <span className="name">{p.name}</span>
                <span className="when">
                  {p.flights === 0
                    ? 'not flown with yet'
                    : `${p.flights} flights · ${p.last_flown}`}
                </span>
              </button>
            </li>
          ))}
          {canAdd && (
            <li>
              <button type="button" role="option" aria-selected={false} className="addnew"
                disabled={busy} onClick={() => void add()}>
                {busy ? 'Adding…' : `Add ${query.trim()} as a new name`}
              </button>
            </li>
          )}
          {matches.length === 0 && !canAdd && <li className="empty">No names match.</li>}
        </ul>
      )}
    </div>
  )
}
