package stats_test

import (
	"testing"
	"time"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/stats"
)

// The aircraft-time aggregation exists because the owner rents aeroplanes and
// SOME OWNERS CHARGE BLOCK TIME AND SOME CHARGE AIR TIME (owner, 2026-08-02).
// It is money, so it lives in this package with the licence totals and is held
// to the same 100% bar.
//
// The load-bearing property under every test here is that block time and air
// time are NEVER MIXED and never silently completed. Block time is known for
// all 1296 transcribed flights; air time is known for 19 of them. A page that
// added up what it had and printed one figure would be claiming a completeness
// it does not have, on a number an invoice gets checked against.

func aircraft(reg, typ string, block int) func(*csvbook.Flight) {
	return func(f *csvbook.Flight) {
		f.AircraftReg = reg
		f.AircraftType = typ
		f.BlockMinutes = block
		f.TotalMinutes = block
	}
}

// airborne sets the optional pair as instants, the way the store holds them.
func airborne(date string, offH, offM, onH, onM int) func(*csvbook.Flight) {
	return func(f *csvbook.Flight) {
		d, err := time.Parse("2006-01-02", date)
		if err != nil {
			panic(err)
		}
		f.TakeoffUTC = d.Add(time.Duration(offH)*time.Hour + time.Duration(offM)*time.Minute)
		f.LandingUTC = d.Add(time.Duration(onH)*time.Hour + time.Duration(onM)*time.Minute)
	}
}

func find(t *testing.T, list []stats.AircraftTime, reg string) stats.AircraftTime {
	t.Helper()
	for _, a := range list {
		if a.Registration == reg {
			return a
		}
	}
	t.Fatalf("no entry for %s in %+v", reg, list)
	return stats.AircraftTime{}
}

// --- AirMinutes -------------------------------------------------------------

func TestAirMinutes(t *testing.T) {
	f := flight(1, "2026-07-30", csvbook.ClassSEPLand,
		aircraft("OH-CAM", "C172", 75), airborne("2026-07-30", 9, 20, 10, 25))
	got, ok := stats.AirMinutes(f)
	if !ok || got != 65 {
		t.Errorf("AirMinutes = %d, %v; want 65, true", got, ok)
	}
}

// Most rows have no airborne pair at all. "Not recorded" must be
// distinguishable from "zero minutes airborne" -- the second is a claim that
// the aeroplane never left the ground, and it is the claim that would silently
// deflate an air-time invoice.
func TestAirMinutesIsUnknownWithoutBothInstants(t *testing.T) {
	base := flight(1, "2026-07-30", csvbook.ClassSEPLand, aircraft("OH-CAM", "C172", 75))

	onlyTakeoff := base
	onlyTakeoff.TakeoffUTC = time.Date(2026, 7, 30, 9, 20, 0, 0, time.UTC)
	onlyLanding := base
	onlyLanding.LandingUTC = time.Date(2026, 7, 30, 10, 25, 0, 0, time.UTC)

	for name, f := range map[string]csvbook.Flight{
		"neither": base, "takeoff only": onlyTakeoff, "landing only": onlyLanding,
	} {
		if got, ok := stats.AirMinutes(f); ok {
			t.Errorf("%s: AirMinutes = %d, true; want unknown", name, got)
		}
	}
}

// The instants carry their dates, so a flight that lands after midnight comes
// out right by subtraction. The FORM has to roll a bare clock by hand; this
// does not, and conflating the two is how one of them gets midnight wrong.
func TestAirMinutesCrossesMidnight(t *testing.T) {
	f := flight(1, "2026-07-30", csvbook.ClassSEPLand, aircraft("OH-CAM", "C172", 80))
	f.TakeoffUTC = time.Date(2026, 7, 30, 23, 30, 0, 0, time.UTC)
	f.LandingUTC = time.Date(2026, 7, 31, 0, 40, 0, 0, time.UTC)

	if got, ok := stats.AirMinutes(f); !ok || got != 70 {
		t.Errorf("AirMinutes = %d, %v; want 70, true", got, ok)
	}
}

// A landing before its takeoff cannot come from the store -- both write paths
// go through timeutil.BlockPair, which rolls forward. If it ever did, the
// answer is "unknown", not a negative number quietly subtracted from a bill.
func TestAirMinutesRefusesANegativePair(t *testing.T) {
	f := flight(1, "2026-07-30", csvbook.ClassSEPLand, aircraft("OH-CAM", "C172", 75))
	f.TakeoffUTC = time.Date(2026, 7, 30, 10, 25, 0, 0, time.UTC)
	f.LandingUTC = time.Date(2026, 7, 30, 9, 20, 0, 0, time.UTC)

	if got, ok := stats.AirMinutes(f); ok {
		t.Errorf("AirMinutes = %d, true; want unknown for a negative pair", got)
	}
}

// --- ByAircraft -------------------------------------------------------------

func TestByAircraftSplitsBlockAndAirWithTheirCoverage(t *testing.T) {
	got := stats.ByAircraft([]csvbook.Flight{
		// Two flights, only one of them carrying airborne times.
		flight(1, "2026-07-01", csvbook.ClassSEPLand,
			aircraft("OH-CAM", "C172", 75), airborne("2026-07-01", 9, 20, 10, 25)),
		flight(2, "2026-07-02", csvbook.ClassSEPLand, aircraft("OH-CAM", "C172", 60)),
		flight(3, "2026-07-03", csvbook.ClassSEPSea, aircraft("OH-CTL", "C172", 45)),
	})

	if len(got) != 2 {
		t.Fatalf("got %d aircraft, want 2: %+v", len(got), got)
	}

	cam := find(t, got, "OH-CAM")
	if cam.Flights != 2 || cam.BlockMinutes != 135 {
		t.Errorf("OH-CAM: %d flights, %d block; want 2, 135", cam.Flights, cam.BlockMinutes)
	}
	// The air total counts ONLY the flight that has one. It is not padded with
	// block time for the other, and the block total is not reduced to match.
	if cam.AirMinutes != 65 {
		t.Errorf("OH-CAM air = %d, want 65 -- only the flight that recorded it", cam.AirMinutes)
	}
	// And the figure travels with the coverage that makes it readable.
	if cam.AirKnown != 1 || cam.AirMissing != 1 {
		t.Errorf("OH-CAM coverage = %d known / %d missing; want 1 / 1",
			cam.AirKnown, cam.AirMissing)
	}
}

// An aeroplane with no airborne times anywhere must report zero air minutes
// AND zero coverage, so the page can say "not recorded on any of these flights"
// rather than printing 0:00 as though it were a measurement.
func TestByAircraftReportsNoAirCoverageAtAll(t *testing.T) {
	got := stats.ByAircraft([]csvbook.Flight{
		flight(1, "2026-07-01", csvbook.ClassSEPLand, aircraft("OH-CAM", "C172", 75)),
		flight(2, "2026-07-02", csvbook.ClassSEPLand, aircraft("OH-CAM", "C172", 60)),
	})
	cam := find(t, got, "OH-CAM")
	if cam.AirMinutes != 0 || cam.AirKnown != 0 || cam.AirMissing != 2 {
		t.Errorf("got air=%d known=%d missing=%d; want 0 / 0 / 2",
			cam.AirMinutes, cam.AirKnown, cam.AirMissing)
	}
}

// Most time first. This page answers "what did I fly and what will it cost", so
// a range holding three aeroplanes must not bury them alphabetically among the
// thirty-five with nothing in it. The tie-break keeps the order deterministic.
func TestByAircraftOrdersByTimeThenRegistration(t *testing.T) {
	got := stats.ByAircraft([]csvbook.Flight{
		flight(1, "2026-07-01", csvbook.ClassSEPLand, aircraft("OH-AAA", "C172", 30)),
		flight(2, "2026-07-02", csvbook.ClassSEPLand, aircraft("OH-ZZZ", "C172", 90)),
		flight(3, "2026-07-03", csvbook.ClassSEPLand, aircraft("OH-MMM", "C172", 30)),
	})
	var order []string
	for _, a := range got {
		order = append(order, a.Registration)
	}
	want := []string{"OH-ZZZ", "OH-AAA", "OH-MMM"}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

// One registration written with two types is a discrepancy, not something to
// resolve by picking the more popular one. The owner ruled the five historical
// cases on 2026-08-02 and the CSVs are guarded against them, but a flight typed
// into the app can still introduce one -- and this page must show it rather
// than choose.
func TestByAircraftKeepsEveryTypeWrittenForARegistration(t *testing.T) {
	got := stats.ByAircraft([]csvbook.Flight{
		flight(1, "2026-07-01", csvbook.ClassSEPLand, aircraft("OH-CMU", "C152", 30)),
		flight(2, "2026-07-02", csvbook.ClassSEPLand, aircraft("OH-CMU", "C172", 30)),
		flight(3, "2026-07-03", csvbook.ClassSEPLand, aircraft("OH-CMU", "C152", 30)),
	})
	cmu := find(t, got, "OH-CMU")
	if len(cmu.Types) != 2 || cmu.Types[0] != "C152" || cmu.Types[1] != "C172" {
		t.Errorf("Types = %v, want both, sorted: [C152 C172]", cmu.Types)
	}
}

// The one row in the whole corpus where block and total disagree (08/09/2025,
// block 0:45 vs total 0:38) is a flagged discrepancy. This page bills on BLOCK,
// so it counts the rows where the two disagree instead of quietly answering a
// question the pilot did not ask.
func TestByAircraftCountsBlockTotalDisagreements(t *testing.T) {
	odd := flight(2, "2026-07-02", csvbook.ClassSEPLand, aircraft("OH-CAM", "C172", 45))
	odd.TotalMinutes = 38

	got := stats.ByAircraft([]csvbook.Flight{
		flight(1, "2026-07-01", csvbook.ClassSEPLand, aircraft("OH-CAM", "C172", 75)),
		odd,
	})
	cam := find(t, got, "OH-CAM")
	if cam.BlockDiffersFromTotal != 1 {
		t.Errorf("BlockDiffersFromTotal = %d, want 1", cam.BlockDiffersFromTotal)
	}
	// Block is what it bills on, so block is what it sums: 75 + 45.
	if cam.BlockMinutes != 120 {
		t.Errorf("BlockMinutes = %d, want 120 -- block, not total", cam.BlockMinutes)
	}
}

// An empty range is a real case, not a degenerate one: a Finnish winter month
// has no flights in it.
func TestByAircraftOnNothing(t *testing.T) {
	if got := stats.ByAircraft(nil); len(got) != 0 {
		t.Errorf("ByAircraft(nil) = %+v, want empty", got)
	}
}

// The caller's slice is never reordered. stats.Paginate has the same guarantee
// and the same test: sorting a legal record as a side effect of rendering it is
// how row order drifts.
func TestByAircraftDoesNotReorderItsInput(t *testing.T) {
	in := []csvbook.Flight{
		flight(1, "2026-07-01", csvbook.ClassSEPLand, aircraft("OH-AAA", "C172", 30)),
		flight(2, "2026-07-02", csvbook.ClassSEPLand, aircraft("OH-ZZZ", "C172", 90)),
	}
	stats.ByAircraft(in)
	if in[0].AircraftReg != "OH-AAA" || in[1].AircraftReg != "OH-ZZZ" {
		t.Errorf("the caller's slice was reordered: %s, %s",
			in[0].AircraftReg, in[1].AircraftReg)
	}
}

// The totals row must be the sum of the rows above it. A page whose total
// disagrees with its own lines is precisely the drift this package exists to
// prevent, and it is the figure an invoice gets checked against.
func TestAircraftTotalAgreesWithItsRows(t *testing.T) {
	flights := []csvbook.Flight{
		flight(1, "2026-07-01", csvbook.ClassSEPLand,
			aircraft("OH-CAM", "C172", 75), airborne("2026-07-01", 9, 20, 10, 25)),
		flight(2, "2026-07-02", csvbook.ClassSEPLand, aircraft("OH-CAM", "C172", 60)),
		flight(3, "2026-07-03", csvbook.ClassSEPSea,
			aircraft("OH-CTL", "C172", 45), airborne("2026-07-03", 12, 0, 12, 40)),
	}
	rows := stats.ByAircraft(flights)
	total := stats.TotalAircraftTime(rows)

	if total.Flights != 3 || total.BlockMinutes != 180 {
		t.Errorf("total = %d flights / %d block; want 3 / 180",
			total.Flights, total.BlockMinutes)
	}
	if total.AirMinutes != 105 || total.AirKnown != 2 || total.AirMissing != 1 {
		t.Errorf("total air = %d over %d known / %d missing; want 105 / 2 / 1",
			total.AirMinutes, total.AirKnown, total.AirMissing)
	}

	var block, air, known, missing, n int
	for _, r := range rows {
		n += r.Flights
		block += r.BlockMinutes
		air += r.AirMinutes
		known += r.AirKnown
		missing += r.AirMissing
	}
	if n != total.Flights || block != total.BlockMinutes || air != total.AirMinutes ||
		known != total.AirKnown || missing != total.AirMissing {
		t.Error("the total disagrees with the sum of its own rows")
	}
}
