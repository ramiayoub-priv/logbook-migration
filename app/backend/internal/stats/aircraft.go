package stats

import (
	"slices"
	"strings"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
)

// Aircraft time: what each aeroplane cost, in the two currencies it gets
// charged in.
//
// The owner rents aeroplanes and SOME OWNERS CHARGE BLOCK TIME AND SOME CHARGE
// AIR TIME (owner ruling, 2026-08-02). That makes this money rather than
// presentation, which is why it lives in this package -- pure, no database, and
// held at 100% alongside the licence totals -- rather than being summed in a
// handler or in the browser.
//
// ONE PROPERTY CARRIES THE WHOLE FILE: block time and air time are never mixed
// and neither is ever silently completed.
//
//	block time is known for every flight     -- 1296 of 1296 transcribed rows
//	air  time is known for a small minority  --   19 of 1296 transcribed rows
//
// So an air-time figure is meaningless without the count of flights it was
// computed from, and every AircraftTime carries that count next to it. A page
// that added up the airborne times it happened to have and printed "Air time:
// 3:20" would be claiming a completeness it does not have, on a number an
// invoice gets checked against. That is rule 0.2 -- surface the gap, never
// paper over it -- applied to a figure nobody had computed before.

// AircraftTime is one aeroplane's time over whatever set of flights it was
// computed from. All durations are minutes, as everywhere inside this app.
type AircraftTime struct {
	Registration string

	// Every distinct type written for this registration, sorted.
	//
	// A slice rather than a string because one registration written with two
	// types is a DISCREPANCY, and choosing the more popular one would be
	// resolving it silently. The owner ruled the five historical cases on
	// 2026-08-02 (the "C192" typo and OH-CMU) and the CSVs are guarded against
	// them by TestEveryRegistrationNamesOneRealAircraftType -- but that guard
	// reads the books, and a flight typed into the app can still introduce one.
	// In practice this holds exactly one element.
	Types []string

	Flights int

	// BlockMinutes is chocks-to-chocks, summed. This is the billing figure:
	// an owner who charges "by the hour" on block time charges on this.
	BlockMinutes int

	// AirMinutes is wheels-up-to-wheels-down, summed over ONLY the flights that
	// recorded both instants. Read it with AirKnown or not at all.
	AirMinutes int
	AirKnown   int // flights carrying both airborne instants
	AirMissing int // Flights - AirKnown

	// How many of these flights have BlockMinutes != TotalMinutes.
	//
	// Almost always zero: the books total on block time on 478 of Book 3's 479
	// rows, and the single exception (08/09/2025, block 0:45 vs total 0:38) is
	// a flagged discrepancy. It is counted rather than reconciled because this
	// page bills on BLOCK while the licence totals run on TotalMinutes, and a
	// row where they disagree is a row where the two pages will legitimately
	// differ. Saying so beats being asked why.
	BlockDiffersFromTotal int
}

// AirMinutes is one flight's airborne time, and whether it is known at all.
//
// The second return is not decoration. Absent airborne times are the common
// case, and an unknown reported as 0 is a claim that the aeroplane never left
// the ground -- which, summed over a month, is an air-time bill that is quietly
// too low. Callers must not collapse the two.
//
// The stored instants carry their dates, so a flight that lands after midnight
// comes out right by subtraction; nothing here rolls a clock. (The new-flight
// FORM does roll one, because there the pilot types four digits with no date
// attached. Two different problems -- do not merge them.)
func AirMinutes(f csvbook.Flight) (int, bool) {
	if f.TakeoffUTC.IsZero() || f.LandingUTC.IsZero() {
		return 0, false
	}
	minutes := int(f.LandingUTC.Sub(f.TakeoffUTC).Minutes())
	if minutes < 0 {
		// Unreachable from the store: both write paths build the pair through
		// timeutil.BlockPair, which rolls forward. If it ever happens, the
		// honest answer is "unknown" rather than a negative number silently
		// subtracted from a total.
		return 0, false
	}
	return minutes, true
}

// ByAircraft groups flights by registration.
//
// The caller filters by date first (stats.Filter), exactly as the statistics
// page does -- this function takes whatever set it is given and says nothing
// about which flights belong in it.
//
// Ordered by block time descending, then by registration. Most time first
// because this page answers "what did I fly and what will it cost", and a range
// holding three aeroplanes must not bury them alphabetically among the
// thirty-five with nothing in them. The tie-break exists so the order is
// deterministic; two runs over the same flights must produce the same page.
//
// The caller's slice is never reordered -- Paginate carries the same guarantee
// and the same test. Sorting a legal record as a side effect of rendering it is
// how row order drifts.
func ByAircraft(flights []csvbook.Flight) []AircraftTime {
	type acc struct {
		AircraftTime
		types map[string]bool
	}
	byReg := map[string]*acc{}
	var order []string

	for _, f := range flights {
		a := byReg[f.AircraftReg]
		if a == nil {
			a = &acc{
				AircraftTime: AircraftTime{Registration: f.AircraftReg},
				types:        map[string]bool{},
			}
			byReg[f.AircraftReg] = a
			order = append(order, f.AircraftReg)
		}

		a.Flights++
		a.BlockMinutes += f.BlockMinutes
		a.types[f.AircraftType] = true
		if f.BlockMinutes != f.TotalMinutes {
			a.BlockDiffersFromTotal++
		}

		if air, ok := AirMinutes(f); ok {
			a.AirMinutes += air
			a.AirKnown++
			continue
		}
		a.AirMissing++
	}

	out := make([]AircraftTime, 0, len(order))
	for _, reg := range order {
		a := byReg[reg]
		a.Types = make([]string, 0, len(a.types))
		for typ := range a.types {
			a.Types = append(a.Types, typ)
		}
		slices.Sort(a.Types)
		out = append(out, a.AircraftTime)
	}

	slices.SortFunc(out, func(x, y AircraftTime) int {
		// y - x: block time descending. The registration tie-break is
		// strings.Compare rather than a pair of ifs so that no branch of this
		// comparator can be unreachable -- registrations are unique keys here,
		// so an "equal" arm would be dead code sitting inside the 100% gate.
		if d := y.BlockMinutes - x.BlockMinutes; d != 0 {
			return d
		}
		return strings.Compare(x.Registration, y.Registration)
	})
	return out
}

// TotalAircraftTime sums the rows into the page's total line.
//
// It sums the ROWS rather than re-walking the flights, so the total cannot
// disagree with the lines above it -- a page whose total contradicts its own
// rows is exactly the drift this package exists to prevent, and it is the
// figure an invoice gets checked against. Asserted both ways in the tests.
//
// Registration and Types are left empty: a total is not an aeroplane.
func TotalAircraftTime(rows []AircraftTime) AircraftTime {
	var t AircraftTime
	for _, r := range rows {
		t.Flights += r.Flights
		t.BlockMinutes += r.BlockMinutes
		t.AirMinutes += r.AirMinutes
		t.AirKnown += r.AirKnown
		t.AirMissing += r.AirMissing
		t.BlockDiffersFromTotal += r.BlockDiffersFromTotal
	}
	return t
}
