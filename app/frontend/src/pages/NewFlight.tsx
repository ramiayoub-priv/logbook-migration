import { api } from '../api'
import { hhmm } from '../format'
import { emptyDraft, FlightForm } from './FlightForm'

/**
 * The new-flight page: the form, plus what "saved" means here.
 *
 * Everything about the fields themselves lives in FlightForm, shared with the
 * edit page — the two must agree about every rule they enforce, and two copies
 * would drift at the first fix applied to only one of them.
 */
export function NewFlightPage() {
  return (
    <FlightForm
      initial={emptyDraft()}
      submitLabel="Log this flight"
      busyLabel="Saving…"
      againLabel="Log another flight"
      onSave={async (payload) => {
        const { flight } = await api.createFlight(payload)
        return {
          message: `Logged ${flight.date}, ${flight.aircraft_reg}, ${hhmm(flight.total_minutes)}.`,
          flight,
        }
      }}
      // A fresh form, but keep the date and aircraft: the next entry after a
      // day of circuits is usually the same aeroplane on the same day.
      resetAfterSave={(previous) => ({
        ...emptyDraft(),
        date: previous.date,
        aircraft_reg: previous.aircraft_reg,
        aircraft_type: previous.aircraft_type,
        class: previous.class,
        pic_name: previous.pic_name,
      })}
    />
  )
}
