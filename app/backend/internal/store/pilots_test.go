package store_test

import (
	"errors"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/store"
	"github.com/ramiayoub/logbook/backend/internal/timeutil"
)

// The pilot roster -- Task 21.
//
// The owner's ask, verbatim (2026-08-03): "I could have a typo when I write
// `self`, it could be `sself` or `SELF` or `seeelf` and I need it to be
// consistent (like the aircraft regs)." So the names that may be written into
// the PIC field become a list to pick from, derived from the names the record
// already contains, plus any added by hand.
//
// WHAT THIS MUST NOT DO is tidy the paper. The books contain `Sinervä` six
// times and `Sinerva` once, which is almost certainly one person -- and rule
// 0.8 puts that out of reach of everything except an owner ruling. The roster
// therefore reports the spellings that EXIST; it only stops new ones being
// invented.

func namesOf(t *testing.T, db *store.DB) []string {
	t.Helper()
	list, err := db.PilotList()
	if err != nil {
		t.Fatalf("PilotList: %v", err)
	}
	out := make([]string, 0, len(list))
	for _, p := range list {
		out = append(out, p.Name)
	}
	return out
}

// TestPilotListIsDerivedFromTheFlightsAndOrderedLikeTheAircraft: the roster is
// not a table anybody maintains. Every name the record uses is in it, counted,
// and the ordering is the aircraft ordering for the same reason -- a name never
// used yet was added because it is about to be, and after that the person you
// flew with last week is the likeliest one to be logging now.
func TestPilotListIsDerivedFromTheFlightsAndOrderedLikeTheAircraft(t *testing.T) {
	db := openTemp(t)
	seedFlights(t, db, []seedFlight{
		{seq: 1, date: "2020-01-01", pic: "self"},
		{seq: 2, date: "2024-05-05", pic: "Martevuo"},
		{seq: 3, date: "2026-07-30", pic: "self"},
		{seq: 4, date: "2019-03-03", pic: ""}, // an empty PIC is not a person
	})
	if _, err := db.AddPilot("Autere"); err != nil {
		t.Fatalf("AddPilot: %v", err)
	}

	// Never used first, then most recently flown with, then by name.
	if got, want := namesOf(t, db), []string{"Autere", "self", "Martevuo"}; !equalStrings(got, want) {
		t.Errorf("PilotList = %v, want %v", got, want)
	}

	list, err := db.PilotList()
	if err != nil {
		t.Fatalf("PilotList: %v", err)
	}
	byName := map[string]store.PilotRow{}
	for _, p := range list {
		byName[p.Name] = p
	}
	if p := byName["self"]; p.Flights != 2 || p.LastFlown != "2026-07-30" {
		t.Errorf("self: %d flights, last %q; want 2 and 2026-07-30", p.Flights, p.LastFlown)
	}
	if p := byName["Autere"]; p.Flights != 0 || !p.UserAdded {
		t.Errorf("Autere: %+v; want 0 flights and UserAdded", p)
	}
	if p := byName["Martevuo"]; p.UserAdded {
		t.Error("a name that came from a flight is not user-added")
	}
}

// TestAddPilotRefusesAVariantOfANameAlreadyKnown is the owner's ask, asserted.
// `SELF` must not be able to become a second entry beside `self`, whether the
// original came from the books or was added here.
func TestAddPilotRefusesAVariantOfANameAlreadyKnown(t *testing.T) {
	db := openTemp(t)
	seedFlights(t, db, []seedFlight{{seq: 1, date: "2020-01-01", pic: "self"}})
	if _, err := db.AddPilot("Autere"); err != nil {
		t.Fatalf("AddPilot: %v", err)
	}

	for _, name := range []string{"SELF", "Self", " self ", "AUTERE", "autere"} {
		if _, err := db.AddPilot(name); !errors.Is(err, store.ErrDuplicatePilot) {
			t.Errorf("AddPilot(%q) = %v, want ErrDuplicatePilot", name, err)
		}
	}
	if got := namesOf(t, db); len(got) != 2 {
		t.Errorf("the roster grew to %v", got)
	}

	// A genuinely different name is still allowed -- including one that differs
	// from a known name by more than case. Two spellings of one surname is a
	// data question for the owner, not something this refuses on a guess.
	if _, err := db.AddPilot("Sinerva"); err != nil {
		t.Errorf("AddPilot(Sinerva): %v", err)
	}
}

func TestAddPilotRefusesAnEmptyName(t *testing.T) {
	db := openTemp(t)
	if _, err := db.AddPilot("   "); err == nil {
		t.Error("AddPilot accepted a blank name")
	}
}

// TestAddPilotKeepsTheSpellingItWasGiven: a name is not a registration. `self`
// is lower case in all 1143 rows that carry it and `Martevuo` is capitalised;
// upper-casing the way registrations are upper-cased would invent a spelling
// the books do not use.
func TestAddPilotKeepsTheSpellingItWasGiven(t *testing.T) {
	db := openTemp(t)
	p, err := db.AddPilot("  von Hertzen  ")
	if err != nil {
		t.Fatalf("AddPilot: %v", err)
	}
	if p.Name != "von Hertzen" {
		t.Errorf("stored %q, want %q -- trimmed, never re-cased", p.Name, "von Hertzen")
	}
}

// TestAddingAPilotWritesNothingToTheFlights is the rule 0.8 guard. The roster
// is a seed list for a form, exactly like the aircraft table: it must be
// incapable of touching the legal record.
func TestAddingAPilotWritesNothingToTheFlights(t *testing.T) {
	db := openTemp(t)
	seedFlights(t, db, []seedFlight{
		{seq: 1, date: "2020-01-01", pic: "self"},
		{seq: 2, date: "2024-05-05", pic: "Martevuo"},
	})
	before := readAllFlights(t, db)

	if _, err := db.AddPilot("Jansson"); err != nil {
		t.Fatalf("AddPilot: %v", err)
	}

	after := readAllFlights(t, db)
	if len(before) != len(after) {
		t.Fatalf("the flight count changed: %d -> %d", len(before), len(after))
	}
	for i := range before {
		if before[i] != after[i] {
			t.Errorf("flight %d changed:\n before %+v\n after  %+v", i, before[i], after[i])
		}
	}
}

// TestTheRealBooksNameEighteenPeople runs the roster over the frozen record.
//
// The figures are the census taken on 2026-08-03 and they are now a fixed
// point: the CSVs are closed (rule 0.8), so a change here is a defect and the
// fix is never to update the constant.
//
// `self` IS 1143 HERE AND 1145 IN THE CSV FILES, and the difference is not a
// bug in either. Books 2 and 3 each open with the previous book's final row as
// a cumulative seed, and the importer skips both -- 1298 CSV rows, 1296
// flights. Both seed rows carry `self`. Counting the files instead of the
// record is how you would get this wrong, and this test is what caught it.
//
// It also states the two things that are WRONG in the frozen data and are
// deliberately left alone: `Sinervä` and `Sinerva` are both here, and so is
// `Stude`, which reads as a word rather than a surname.
func TestTheRealBooksNameEighteenPeople(t *testing.T) {
	lb := loadRealBooks(t)
	db := openTemp(t)
	if _, err := db.Import(lb, "real books"); err != nil {
		t.Fatalf("Import: %v", err)
	}

	list, err := db.PilotList()
	if err != nil {
		t.Fatalf("PilotList: %v", err)
	}
	if len(list) != 18 {
		t.Errorf("the books name %d people, want 18", len(list))
	}

	byName := map[string]store.PilotRow{}
	for _, p := range list {
		byName[p.Name] = p
		if p.UserAdded {
			t.Errorf("%q is marked user-added, but nothing was added by hand", p.Name)
		}
	}
	for name, want := range map[string]int{
		"self": 1143, "Martevuo": 54, "Autere": 30, "Stude": 18, "Jansson": 16,
		"Sinervä": 6, "Sinerva": 1,
	} {
		if got := byName[name].Flights; got != want {
			t.Errorf("%q flies %d flights, want %d", name, got, want)
		}
	}
	// The empty PIC cell on one row is not a person and must not be offered.
	if _, ok := byName[""]; ok {
		t.Error("the roster offers an empty name")
	}
}

// --- helpers ----------------------------------------------------------------

type seedFlight struct {
	seq  int
	date string
	pic  string
}

// seedFlights writes flights through the importer, which is the only way rows
// get into a test database without reaching past the store's own contract.
func seedFlights(t *testing.T, db *store.DB, rows []seedFlight) {
	t.Helper()
	lb := &csvbook.Logbook{}
	for _, r := range rows {
		lb.Flights = append(lb.Flights, csvbook.Flight{
			Seq: r.seq, Date: r.date, SourceBook: 1, SourceRow: r.seq,
			AircraftReg: "OH-CTL", AircraftType: "C172", Class: csvbook.ClassSEPLand,
			PICName: r.pic, TimeOrigin: timeutil.OriginNone,
		})
	}
	lb.Aircraft = []csvbook.Aircraft{
		{Registration: "OH-CTL", Type: "C172", DefaultClass: csvbook.ClassSEPLand},
	}
	// The importer verifies what it wrote against these before committing, so a
	// seed with no totals is refused outright -- which is the check doing its
	// job, not a nuisance (rule 0.2).
	lb.Totals.Flights = len(lb.Flights)
	if _, err := db.Import(lb, "test"); err != nil {
		t.Fatalf("seeding flights: %v", err)
	}
}

func readAllFlights(t *testing.T, db *store.DB) []csvbook.Flight {
	t.Helper()
	f, err := db.Flights()
	if err != nil {
		t.Fatalf("Flights: %v", err)
	}
	return f
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
