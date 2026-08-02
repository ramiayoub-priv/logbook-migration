import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { api, ApiError, type Flight, type FlightDraft } from '../api'
import { useApi } from '../auth'
import {
  clock,
  clockWire,
  digits,
  hhmm,
  minutesToDigits,
  parseClockDigits,
  parseDurationDigits,
  todayISO,
} from '../format'
import { Link } from '../router'

// The flight form, shared by the new-flight page and the edit page.
//
// Extracted 2026-08-02 when editing arrived. One form rather than two, because
// the two must agree about every rule they enforce: the four-digit fields, the
// zone toggle, the derived total, what is refused before the request. Two
// copies would have started identical and drifted at the first fix applied to
// only one of them -- which is exactly how the duration fields kept their
// untypeable colon after the clock fields lost theirs.

// Short labels on purpose: the select is half a phone wide, and "Single engine
// piston — land" truncates to something unreadable. Every pilot reads SEP/MEP.
const CLASSES = [
  { value: 'SEP_LAND', label: 'SEP (land)' },
  { value: 'SEP_SEA', label: 'SEP (sea)' },
  { value: 'MEP_LAND', label: 'MEP (land)' },
  { value: 'MEP_SEA', label: 'MEP (sea)' },
  { value: 'TMG', label: 'TMG' },
]

// Every time on this form -- clock or duration -- is four digits on a number
// pad. See format.ts for why it is exactly four and never three.
const CLOCK_FIELDS = ['off_block', 'on_block', 'takeoff', 'landing'] as const
const DURATION_FIELDS = [
  'night_time',
  'instrument_time',
  'pic_time',
  'dual_time',
  'instructor_time',
] as const

/**
 * The controls in the order they appear on the page.
 *
 * Used for one thing only: deciding which failing field to jump to. A refusal
 * has to land on the FIRST problem the pilot would meet reading down the form,
 * not on whichever one the server happened to list first -- scrolling past two
 * untouched errors to reach a third is how a form gets abandoned, and an
 * abandoned form means the flight never gets logged at all.
 */
const FIELD_ORDER = [
  'date', 'aircraft_reg', 'aircraft_type', 'class', 'dep_place', 'arr_place',
  'off_block', 'on_block', 'total_time', 'takeoff', 'landing', 'air_time',
  'pic_time', 'dual_time', 'night_time', 'instrument_time', 'instructor_time',
  'landings_day', 'landings_night', 'pic_name', 'remarks',
]

/**
 * Saved is what a successful save hands back to the form.
 *
 * The flight comes with it, not just the sentence, because the confirmation
 * names what was ACTUALLY STORED rather than what was typed: the server is the
 * authority on the record, and a screen that read back the draft would agree
 * with the pilot instead of with the logbook.
 */
export interface Saved {
  message: string
  flight: Flight
}

export function emptyDraft(): FlightDraft {
  return {
    date: todayISO(),
    aircraft_reg: '',
    aircraft_type: '',
    class: 'SEP_LAND',
    dep_place: '',
    arr_place: '',
    off_block: '',
    on_block: '',
    takeoff: '',
    landing: '',
    total_time: '',
    night_time: '',
    instrument_time: '',
    pic_time: '',
    dual_time: '',
    instructor_time: '',
    pic_name: '',
    landings_day: 1,
    landings_night: 0,
    remarks: '',
  }
}

/**
 * draftFromFlight turns a stored flight back into what the form holds.
 *
 * The clocks are read from the CANONICAL UTC instants and the zone toggle sits
 * on UTC, rather than reproducing off_block_raw as it was typed. Two reasons.
 * The raw string can be a local time, and showing "12:15" without saying which
 * zone is how an hour goes missing; and every instant in this application is
 * UTC (rule 0.4), so UTC is the honest thing to show. The cost, which is real
 * and small: saving a flight that was entered in local time restates its raw
 * value as the equivalent UTC. The instant does not move, no total changes,
 * and what the pilot now asserts is what gets stored.
 *
 * The airborne pair is carried through even though it is folded away, because
 * a form that submits a field it cannot show would erase that field.
 */
export function draftFromFlight(f: Flight): FlightDraft {
  const clock = (instant: string | null): string => {
    if (!instant) return ''
    const d = new Date(instant)
    if (Number.isNaN(d.getTime())) return ''
    return d.toISOString().slice(11, 16).replace(':', '')
  }
  const dur = (minutes: number): string => (minutes === 0 ? '' : minutesToDigits(minutes))

  return {
    date: f.date,
    aircraft_reg: f.aircraft_reg,
    aircraft_type: f.aircraft_type,
    class: f.class,
    dep_place: f.dep_place,
    arr_place: f.arr_place,
    off_block: clock(f.off_block_utc),
    on_block: clock(f.on_block_utc),
    takeoff: clock(f.takeoff_utc),
    landing: clock(f.landing_utc),
    total_time: '',
    night_time: dur(f.night_minutes),
    instrument_time: dur(f.instrument_minutes),
    pic_time: dur(f.pic_minutes),
    dual_time: dur(f.dual_minutes),
    instructor_time: dur(f.instructor_minutes),
    pic_name: f.pic_name,
    landings_day: f.landings_day,
    landings_night: f.landings_night,
    remarks: f.remarks,
  }
}

export function FlightForm({
  initial,
  submitLabel,
  busyLabel,
  againLabel,
  onSave,
  resetAfterSave,
  footer,
}: {
  initial: FlightDraft
  submitLabel: string
  busyLabel: string
  /** The confirmation's way back to the form: "Log another flight" / "Keep editing…". */
  againLabel: string
  /** Saves the flight and returns what to confirm. May throw an ApiError. */
  onSave: (payload: FlightDraft) => Promise<Saved>
  /** What the form holds when the pilot goes back for another. Absent means "leave it". */
  resetAfterSave?: (previous: FlightDraft) => FlightDraft
  /** Rendered after the submit button -- the delete panel, on the edit page. */
  footer?: React.ReactNode
}) {
  const loadAircraft = useCallback(() => api.aircraft(), [])
  const { data: fleet } = useApi(loadAircraft, [])

  const [draft, setDraft] = useState<FlightDraft>(initial)
  // Defaults to UTC: the owner's rule is that everything is UTC from now on
  // (rule 0.4), and it is what an edited flight is shown in.
  const [utc, setUTC] = useState(true)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [saved, setSaved] = useState<Saved | null>(null)
  const [busy, setBusy] = useState(false)

  // Bumped on every REFUSED submit, and the only thing the jump-to-the-problem
  // effect watches. A counter rather than the error map itself: two submissions
  // that fail on the same field must both move the pilot there, and comparing
  // the maps would call the second one "no change" and leave them staring at a
  // form that appeared to do nothing.
  const [refusals, setRefusals] = useState(0)
  const errorRef = useRef<HTMLParagraphElement>(null)

  useEffect(() => {
    if (refusals === 0) return
    // The topmost failing control, by the order of the page rather than the
    // order the server listed them in.
    const first = FIELD_ORDER.find((f) => fieldErrors[f])
    const el = first ? document.getElementById(first) : null
    if (el) {
      el.scrollIntoView?.({ block: 'center' })
      el.focus?.({ preventScroll: true })
      return
    }
    // No field to blame -- a 409, a 403, a dead network. The banner is the
    // thing to get in front of, and it carries the whole message.
    errorRef.current?.scrollIntoView?.({ block: 'center' })
    errorRef.current?.focus?.({ preventScroll: true })
    // fieldErrors is deliberately not a dependency: it is set in the same batch
    // as `refusals`, so this closure already sees it, and watching it as well
    // would re-run the jump every time a field is corrected.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refusals])

  const aircraft = useMemo(() => {
    const list = fleet?.aircraft ?? []
    // Active first -- the aeroplanes actually being flown -- then the rest, so
    // the picker is short without hiding history.
    return [...list].sort((a, b) =>
      a.active === b.active ? a.registration.localeCompare(b.registration) : a.active ? -1 : 1,
    )
  }, [fleet])

  function set<K extends keyof FlightDraft>(key: K, value: FlightDraft[K]) {
    setDraft((d) => ({ ...d, [key]: value }))
    setFieldErrors((e) => {
      if (!e[key as string]) return e
      const next = { ...e }
      delete next[key as string]
      return next
    })
    setSaved(null)
  }

  /**
   * Choosing an aircraft fills in its type and its usual class.
   *
   * The class is a default, not a constraint: the same registration can be
   * flown on floats or wheels, and it is the class that decides which rating
   * the flight counts towards -- so the field stays editable underneath.
   */
  function chooseAircraft(reg: string) {
    const a = aircraft.find((x) => x.registration === reg)
    setDraft((d) => ({
      ...d,
      aircraft_reg: reg,
      aircraft_type: a?.type ?? d.aircraft_type,
      class: a?.default_class ?? d.class,
    }))
    setSaved(null)
  }

  // Total time is chocks-to-chocks and is DERIVED, not typed. It recomputes on
  // every keystroke, so the moment the fourth digit of the on-block time lands
  // the pilot can see whether the flight came out the length they flew.
  const blockMinutes = useMemo(
    () => blockFrom(draft.off_block, draft.on_block),
    [draft.off_block, draft.on_block],
  )
  // Airborne time, likewise derived. Blank whenever the optional pair is.
  const airMinutes = useMemo(
    () => blockFrom(draft.takeoff, draft.landing),
    [draft.takeoff, draft.landing],
  )

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setFormError(null)
    setFieldErrors({})

    // The wire format is unchanged -- "HH:MM", or "HH:MMZ" when the zone
    // toggle says UTC -- so the server's single conversion authority (rule
    // 0.4) never learns that the form's fields changed shape. Putting the
    // colon and the Z back on is this function's whole job.
    //
    // This is also the one place the form is allowed to refuse anything. Half
    // a time is not a time, and sending "091" for the server to reject would
    // cost a round trip to say something that is obvious here.
    const errs: Record<string, string> = {}
    const wire: Partial<Record<string, string>> = {}

    for (const key of CLOCK_FIELDS) {
      const raw = draft[key]
      if (raw === '') {
        // Blank is not this function's business: off/on block being required
        // and the airborne pair being all-or-nothing are the server's rules,
        // and they stay stated in exactly one place.
        wire[key] = ''
        continue
      }
      const hhmmText = clockWire(raw)
      if (hhmmText === '') {
        errs[key] = 'Write this time as four digits, for example 0915 for 09:15.'
        continue
      }
      wire[key] = utc ? `${hhmmText}Z` : hhmmText
    }

    for (const key of DURATION_FIELDS) {
      const raw = draft[key]
      if (raw === '') {
        wire[key] = ''
        continue
      }
      const minutes = parseDurationDigits(raw)
      if (minutes === null) {
        errs[key] = 'Write this duration as four digits, for example 0115 for 1:15.'
        continue
      }
      wire[key] = hhmm(minutes)
    }

    if (Object.keys(errs).length > 0) {
      refuse(errs, 'This flight cannot be saved as written. See the fields below.')
      return
    }

    setBusy(true)
    try {
      const payload: FlightDraft = {
        ...draft,
        ...(wire as Partial<FlightDraft>),
        // Derived from the clock, never typed. Sent explicitly because the
        // server still requires the total to be stated rather than inventing
        // it -- this form is what states it.
        total_time: blockMinutes === null ? '' : hhmm(blockMinutes),
      }
      setSaved(await onSave(payload))
    } catch (err) {
      // Note what is NOT here: the draft is untouched on every one of these
      // paths. A phone that empties a twenty-field form because the server said
      // 400 is a phone that does not get the flight logged at all.
      if (err instanceof ApiError && err.fields.length > 0) {
        const map: Record<string, string> = {}
        for (const f of err.fields) map[f.field] = f.message
        refuse(map, 'This flight cannot be saved as written. See the fields below.')
      } else if (err instanceof ApiError && (err.status === 409 || err.status === 403)) {
        refuse({}, err.message)
      } else {
        refuse({}, err instanceof Error ? err.message : 'Could not save the flight.')
      }
    } finally {
      setBusy(false)
    }
  }

  /** refuse records a rejected save and sends the pilot to what caused it. */
  function refuse(fields: Record<string, string>, message: string) {
    setFieldErrors(fields)
    setFormError(message)
    setRefusals((n) => n + 1)
  }

  /**
   * The save takes over the screen.
   *
   * Ruled by the owner on 2026-08-02 after the first real day of use: the old
   * confirmation was a <p role="status"> at the TOP of a form three cards long,
   * so on a phone the submit button and the answer to "did that work?" were
   * never on screen together. A screen-reader user was told; the pilot looking
   * at the screen was not.
   *
   * Replacing the form is what makes it unmissable -- a message that shares the
   * page with the thing it is about can always be scrolled away from.
   */
  if (saved) {
    return (
      <SavedFlight
        saved={saved}
        againLabel={againLabel}
        onAgain={() => {
          if (resetAfterSave) setDraft(resetAfterSave(draft))
          setSaved(null)
        }}
      />
    )
  }

  return (
    <form onSubmit={submit}>
      {formError && (
        <p className="error big" role="alert" tabIndex={-1} ref={errorRef}>
          {formError}
        </p>
      )}

      <div className="card">
        <h2>The flight</h2>

        <Field id="date" label="Date" error={fieldErrors['date']}>
          <input
            id="date"
            type="date"
            max={todayISO()}
            value={draft.date}
            onChange={(e) => set('date', e.target.value)}
            aria-invalid={!!fieldErrors['date']}
          />
        </Field>

        <Field id="aircraft_reg" label="Aircraft" error={fieldErrors['aircraft_reg']}>
          <select
            id="aircraft_reg"
            value={draft.aircraft_reg}
            onChange={(e) => chooseAircraft(e.target.value)}
            aria-invalid={!!fieldErrors['aircraft_reg']}
          >
            <option value="">Choose an aircraft…</option>
            {aircraft.map((a) => (
              <option key={a.registration} value={a.registration}>
                {a.registration} — {a.type}
                {a.active ? '' : ' (retired)'}
              </option>
            ))}
            {/* An edited flight may name an aeroplane that has since left the
                seed list. Without this the select would silently show blank
                and the next save would clear the registration. */}
            {draft.aircraft_reg !== '' &&
              !aircraft.some((a) => a.registration === draft.aircraft_reg) && (
                <option value={draft.aircraft_reg}>{draft.aircraft_reg}</option>
              )}
          </select>
        </Field>

        <div className="row">
          <Field id="aircraft_type" label="Type" error={fieldErrors['aircraft_type']}>
            <input
              id="aircraft_type"
              value={draft.aircraft_type}
              onChange={(e) => set('aircraft_type', e.target.value)}
              autoCapitalize="characters"
              aria-invalid={!!fieldErrors['aircraft_type']}
            />
          </Field>
          <Field id="class" label="Class" error={fieldErrors['class']}>
            <select
              id="class"
              value={draft.class}
              onChange={(e) => set('class', e.target.value)}
              aria-invalid={!!fieldErrors['class']}
            >
              {CLASSES.map((c) => (
                <option key={c.value} value={c.value}>
                  {c.label}
                </option>
              ))}
            </select>
          </Field>
        </div>

        <div className="row">
          <Field id="dep_place" label="From" error={fieldErrors['dep_place']}>
            <input
              id="dep_place"
              value={draft.dep_place}
              onChange={(e) => set('dep_place', e.target.value)}
              autoCapitalize="characters"
            />
          </Field>
          <Field id="arr_place" label="To" error={fieldErrors['arr_place']}>
            <input
              id="arr_place"
              value={draft.arr_place}
              onChange={(e) => set('arr_place', e.target.value)}
              autoCapitalize="characters"
            />
          </Field>
        </div>
      </div>

      <div className="card">
        <h2>Times</h2>

        {/*
          One toggle for the whole card instead of a Z the pilot types. A phone
          number pad has no colon key and no Z, so the old free-text "09:15Z"
          field was literally untypeable on the device this app is used on --
          and the Z is load-bearing, so it cannot simply be dropped. Making the
          zone a control states it explicitly and leaves every time field able
          to be four digits.
        */}
        <div className="zonetoggle" role="group" aria-label="Time zone for the times below">
          <button
            type="button"
            className={utc ? 'zone on' : 'zone'}
            aria-pressed={utc}
            onClick={() => setUTC(true)}
          >
            UTC
          </button>
          <button
            type="button"
            className={utc ? 'zone' : 'zone on'}
            aria-pressed={!utc}
            onClick={() => setUTC(false)}
          >
            Helsinki local
          </button>
        </div>
        <p className="muted small">
          {utc
            ? 'The times below are read as UTC.'
            : 'The times below are read as Helsinki local time and converted to UTC. A time in the hour the clocks change is ambiguous and will be refused — write UTC instead.'}
        </p>

        <p className="muted small">
          Times are four digits, no colon: <strong>0915</strong> is 09:15.
        </p>

        <div className="row">
          <Field id="off_block" label="Off block" error={fieldErrors['off_block']}>
            <TimeDigits
              id="off_block"
              kind="clock"
              value={draft.off_block}
              onChange={(v) => set('off_block', v)}
              invalid={!!fieldErrors['off_block']}
            />
          </Field>
          <Field id="on_block" label="On block" error={fieldErrors['on_block']}>
            <TimeDigits
              id="on_block"
              kind="clock"
              value={draft.on_block}
              onChange={(v) => set('on_block', v)}
              invalid={!!fieldErrors['on_block']}
            />
          </Field>
        </div>

        <Field id="total_time" label="Total time" error={fieldErrors['total_time']}>
          {/*
            Derived, and deliberately not editable: this is the figure the whole
            logbook adds up and a licence application is written from, and the
            book totals on block time on 478 of Book 3's 479 rows. Typing it by
            hand next to the clock that already states it is two places for one
            fact to disagree.
          */}
          <input
            id="total_time"
            className="derived"
            readOnly
            tabIndex={-1}
            value={blockMinutes === null ? '' : hhmm(blockMinutes)}
            placeholder="—"
            aria-invalid={!!fieldErrors['total_time']}
          />
          <div className="muted small">Chocks to chocks, from the two times above.</div>
        </Field>

        {/*
          The airborne pair sits here in the open, next to off/on block.

          It was folded behind a disclosure until 2026-08-02 because most rows
          in the PAPER BOOKS have none -- 19 of 1296 -- but that is a fact about
          history, not about the flights being flown now. The owner ruled it
          out of the fold: the aircraft's own logbook, a separate and legally
          required document, is filled from AIRBORNE times, so this is the pair
          that gets copied out after every flight. A field you have to remember
          to expand is a field that ends up empty, and an empty airborne time is
          what makes an air-time total unusable a year later when it is being
          billed from.

          Still optional, and still all-or-nothing as a pair -- that rule is the
          server's and is stated in internal/entry, not here.
        */}
        <p className="muted small">
          Takeoff and landing are optional, but the aircraft's own logbook is filled from them.
          If you give one, give both.
        </p>

        <div className="row">
          <Field id="takeoff" label="Takeoff" error={fieldErrors['takeoff']}>
            <TimeDigits
              id="takeoff"
              kind="clock"
              value={draft.takeoff}
              onChange={(v) => set('takeoff', v)}
              invalid={!!fieldErrors['takeoff']}
            />
          </Field>
          <Field id="landing" label="Landing" error={fieldErrors['landing']}>
            <TimeDigits
              id="landing"
              kind="clock"
              value={draft.landing}
              onChange={(v) => set('landing', v)}
              invalid={!!fieldErrors['landing']}
            />
          </Field>
        </div>

        <Field id="air_time" label="Air time">
          <input
            id="air_time"
            className="derived"
            readOnly
            tabIndex={-1}
            value={airMinutes === null ? '' : hhmm(airMinutes)}
            placeholder="—"
          />
          <div className="muted small">Wheels up to wheels down, from the two times above.</div>
        </Field>

        {/*
          The durations are four digits on the same number pad. They were the
          half of the form the mobile rework missed: they still wanted "1:15"
          on a keyboard with no colon key, which is the identical defect one
          card further down the page.
        */}
        <p className="muted small">
          Durations are four digits too: <strong>0115</strong> is 1:15, <strong>0020</strong> is
          0:20. Leave a field empty if the flight logged none.
        </p>

        <div className="row">
          <Field id="pic_time" label="PIC" error={fieldErrors['pic_time']}>
            <TimeDigits
              id="pic_time"
              kind="duration"
              value={draft.pic_time}
              onChange={(v) => set('pic_time', v)}
              invalid={!!fieldErrors['pic_time']}
            />
            {blockMinutes !== null && draft.pic_time !== minutesToDigits(blockMinutes) && (
              <button type="button" className="link"
                onClick={() => set('pic_time', minutesToDigits(blockMinutes))}>
                Use the whole {hhmm(blockMinutes)}
              </button>
            )}
          </Field>
          <Field id="dual_time" label="Dual" error={fieldErrors['dual_time']}>
            <TimeDigits
              id="dual_time"
              kind="duration"
              value={draft.dual_time}
              onChange={(v) => set('dual_time', v)}
              invalid={!!fieldErrors['dual_time']}
            />
          </Field>
        </div>

        <div className="row three">
          <Field id="night_time" label="Night" error={fieldErrors['night_time']}>
            <TimeDigits
              id="night_time"
              kind="duration"
              value={draft.night_time}
              onChange={(v) => set('night_time', v)}
              invalid={!!fieldErrors['night_time']}
            />
          </Field>
          <Field id="instrument_time" label="Instrument" error={fieldErrors['instrument_time']}>
            <TimeDigits
              id="instrument_time"
              kind="duration"
              value={draft.instrument_time}
              onChange={(v) => set('instrument_time', v)}
              invalid={!!fieldErrors['instrument_time']}
            />
          </Field>
          <Field id="instructor_time" label="Instructor" error={fieldErrors['instructor_time']}>
            <TimeDigits
              id="instructor_time"
              kind="duration"
              value={draft.instructor_time}
              onChange={(v) => set('instructor_time', v)}
              invalid={!!fieldErrors['instructor_time']}
            />
          </Field>
        </div>
      </div>

      <div className="card">
        <h2>Landings and remarks</h2>
        <div className="row">
          <Field id="landings_day" label="Landings — day" error={fieldErrors['landings_day']}>
            <input
              id="landings_day"
              type="number"
              min={0}
              inputMode="numeric"
              value={draft.landings_day}
              onChange={(e) => set('landings_day', Number(e.target.value))}
              aria-invalid={!!fieldErrors['landings_day']}
            />
          </Field>
          <Field id="landings_night" label="Landings — night" error={fieldErrors['landings_night']}>
            <input
              id="landings_night"
              type="number"
              min={0}
              inputMode="numeric"
              value={draft.landings_night}
              onChange={(e) => set('landings_night', Number(e.target.value))}
              aria-invalid={!!fieldErrors['landings_night']}
            />
          </Field>
        </div>

        <Field id="pic_name" label="Name of pilot in command" error={fieldErrors['pic_name']}>
          <input id="pic_name" value={draft.pic_name}
            onChange={(e) => set('pic_name', e.target.value)} />
        </Field>

        <Field id="remarks" label="Remarks" error={fieldErrors['remarks']}>
          <textarea id="remarks" rows={2} value={draft.remarks}
            onChange={(e) => set('remarks', e.target.value)} />
        </Field>
      </div>

      <button className="primary" type="submit" disabled={busy}>
        {busy ? busyLabel : submitLabel}
      </button>
      <p className="muted small">
        The day/night landing split you type here is recorded as verified, because you read it
        off the aeroplane rather than inferring it.
      </p>

      {footer}
    </form>
  )
}

/**
 * The whole screen, after a flight is saved.
 *
 * It names what the SERVER stored, not what was typed: date, registration, the
 * two clock times, the total and the landings. "Saved" on its own is a claim
 * the pilot cannot check, and the whole reason this panel exists is that the
 * previous confirmation could not be seen at all.
 *
 * One live region, and it is this element. The page has been bitten once
 * already by a second implicit role="status" (an <output>), which made the
 * saved-flight announcement ambiguous. Focus moves here as well, because
 * replacing the form removes the button that had focus -- without this, a
 * screen reader lands on the document body and a keyboard user starts over at
 * the top of the page.
 */
function SavedFlight({
  saved,
  againLabel,
  onAgain,
}: {
  saved: Saved
  againLabel: string
  onAgain: () => void
}) {
  const panel = useRef<HTMLDivElement>(null)
  const f = saved.flight
  const landings = f.landings_day + f.landings_night

  useEffect(() => {
    panel.current?.scrollIntoView?.({ block: 'start' })
    panel.current?.focus?.({ preventScroll: true })
  }, [])

  return (
    <div className="card saved" role="status" tabIndex={-1} ref={panel}>
      <p className="tick" aria-hidden="true">
        ✓
      </p>
      <h2>{saved.message}</h2>

      <dl className="detail">
        <div>
          <dt>Date</dt>
          <dd>{f.date}</dd>
        </div>
        <div>
          <dt>Aircraft</dt>
          <dd>
            {f.aircraft_reg} {f.aircraft_type}
          </dd>
        </div>
        <div>
          <dt>Route</dt>
          <dd>
            {f.dep_place} → {f.arr_place}
          </dd>
        </div>
        <div>
          <dt>Off / on</dt>
          <dd>
            {clock(f.off_block_utc)} – {clock(f.on_block_utc)} UTC
          </dd>
        </div>
        {/* Only when they were given. A dash where an airborne time should be
            reads as "none flown" rather than "none recorded". */}
        {f.takeoff_utc && f.landing_utc && (
          <div>
            <dt>Airborne</dt>
            <dd>
              {clock(f.takeoff_utc)} – {clock(f.landing_utc)} UTC
            </dd>
          </div>
        )}
        <div>
          <dt>Total</dt>
          <dd>{hhmm(f.total_minutes)}</dd>
        </div>
        <div>
          <dt>Landings</dt>
          <dd>
            {landings} ({f.landings_day} day, {f.landings_night} night)
          </dd>
        </div>
      </dl>

      <p className="muted small">
        It is in the logbook now, and every total already includes it.
      </p>

      <div className="row">
        <button type="button" className="primary" onClick={onAgain}>
          {againLabel}
        </button>
        <Link to="table" className="buttonlike">
          See it in the table
        </Link>
      </div>
    </div>
  )
}

export function Field({
  id,
  label,
  error,
  children,
}: {
  id: string
  label: string
  error?: string | undefined
  children: React.ReactNode
}) {
  return (
    <div className="field">
      <label htmlFor={id}>{label}</label>
      {children}
      {error && (
        <div className="fielderror" role="alert">
          {error}
        </div>
      )}
    </div>
  )
}

/**
 * TimeDigits is every time field on this form: four numerals, a number pad,
 * and nothing to punctuate.
 *
 * `inputMode="numeric"` with a digits-only filter is what makes the phone show
 * the pad it should have shown all along. The filter is not cosmetic -- it is
 * what lets a pasted "09:15Z" become 0915 instead of a field the pilot has to
 * clean up by hand -- and `maxLength` stops a fifth digit from silently
 * shifting the hours.
 *
 * The echo underneath reads the digits back as the time they mean. Four
 * unpunctuated numerals are quick to type and easy to transpose, so the form
 * shows its reading of them before the flight is saved rather than after.
 */
function TimeDigits({
  id,
  kind,
  value,
  onChange,
  invalid,
}: {
  id: string
  kind: 'clock' | 'duration'
  value: string
  onChange: (v: string) => void
  invalid: boolean
}) {
  const minutes = kind === 'clock' ? parseClockDigits(value) : parseDurationDigits(value)
  const echo =
    value === ''
      ? ''
      : minutes === null
        ? '—'
        : kind === 'clock'
          ? clockWire(value)
          : hhmm(minutes)
  return (
    <>
      <input
        id={id}
        className="digits"
        inputMode="numeric"
        pattern="[0-9]*"
        maxLength={4}
        autoComplete="off"
        placeholder={kind === 'clock' ? '0915' : '0115'}
        value={value}
        onChange={(e) => onChange(digits(e.target.value))}
        aria-invalid={invalid}
      />
      {echo !== '' && <div className="echo muted small">{echo}</div>}
    </>
  )
}

/**
 * blockFrom computes chocks-to-chocks from the two clock entries, rolling past
 * midnight the way the server does.
 *
 * Both entries are in the same zone -- one toggle governs the whole card -- so
 * the difference is the same figure whichever zone that is. A pair that mixes
 * zones cannot be typed here at all, and the server refuses one outright
 * rather than guessing.
 */
export function blockFrom(off: string, on: string): number | null {
  const a = parseClockDigits(off)
  const b = parseClockDigits(on)
  if (a === null || b === null) return null
  const d = b - a
  return d > 0 ? d : d + 24 * 60
}
