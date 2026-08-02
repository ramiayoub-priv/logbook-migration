package store_test

import (
	"errors"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

// The aircraft list used to be DERIVED: rebuilt from the flights on every
// import, which meant the only aeroplanes that could exist were the ones
// already flown. That made the first flight in a new aeroplane unenterable --
// the form's registration is a dropdown fed by this list.
//
// Owner ruling 2026-08-02: aircraft become real records with a write path,
// there is NO delete and NO active/retired concept (an aeroplane, once added,
// stays; the dropdown is filterable instead), and the list is ordered by what
// was flown most recently.

func aircraftDB(t *testing.T) *store.DB {
	t.Helper()
	db := openTemp(t)
	if _, err := db.Import(sample(t), "test"); err != nil {
		t.Fatal(err)
	}
	return db
}

func newAircraft(reg, typ string, class csvbook.Class) csvbook.Aircraft {
	return csvbook.Aircraft{Registration: reg, Type: typ, DefaultClass: class}
}

func TestAddAircraftMakesAnAeroplaneThatHasNeverBeenFlownAvailable(t *testing.T) {
	db := aircraftDB(t)

	got, err := db.AddAircraft(newAircraft("OH-XYZ", "C152", csvbook.ClassSEPLand))
	if err != nil {
		t.Fatalf("AddAircraft: %v", err)
	}
	if got.Registration != "OH-XYZ" || got.Type != "C152" {
		t.Errorf("AddAircraft returned %+v", got)
	}
	if !got.UserAdded {
		t.Error("an aeroplane added by hand must be marked as such, so the importer " +
			"can tell which rows it owns")
	}
	if got.Flights != 0 || got.LastFlown != "" {
		t.Errorf("a never-flown aeroplane reports %d flights, last flown %q",
			got.Flights, got.LastFlown)
	}

	list, err := db.AircraftList()
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, a := range list {
		if a.Registration == "OH-XYZ" {
			found = true
		}
	}
	if !found {
		t.Error("the new aeroplane is not in the list the form is built from")
	}
}

// Registrations are unique and the second attempt must be refused rather than
// quietly overwriting the first -- an overwrite would change the type or class
// of an aeroplane already in the books.
func TestAddAircraftRefusesARegistrationThatAlreadyExists(t *testing.T) {
	db := aircraftDB(t)

	if _, err := db.AddAircraft(newAircraft("OH-XYZ", "C152", csvbook.ClassSEPLand)); err != nil {
		t.Fatal(err)
	}
	_, err := db.AddAircraft(newAircraft("OH-XYZ", "P28A", csvbook.ClassSEPSea))
	if !errors.Is(err, store.ErrDuplicateAircraft) {
		t.Fatalf("second add returned %v, want ErrDuplicateAircraft", err)
	}

	// And the first one is untouched.
	a, err := db.AircraftByReg("OH-XYZ")
	if err != nil {
		t.Fatal(err)
	}
	if a.Type != "C152" || a.DefaultClass != csvbook.ClassSEPLand {
		t.Errorf("the refused add still changed the existing row: %+v", a)
	}
}

// One that already came from the books is a duplicate too.
func TestAddAircraftRefusesOneAlreadyInTheBooks(t *testing.T) {
	db := aircraftDB(t)
	_, err := db.AddAircraft(newAircraft("OH-CTL", "C172", csvbook.ClassSEPSea))
	if !errors.Is(err, store.ErrDuplicateAircraft) {
		t.Fatalf("adding OH-CTL again returned %v, want ErrDuplicateAircraft", err)
	}
}

func TestUpdateAircraftCorrectsATypo(t *testing.T) {
	db := aircraftDB(t)
	if _, err := db.AddAircraft(newAircraft("OH-XZY", "C152", csvbook.ClassSEPLand)); err != nil {
		t.Fatal(err)
	}

	got, err := db.UpdateAircraft("OH-XZY", csvbook.Aircraft{
		Registration: "OH-XYZ", Type: "C152", DefaultClass: csvbook.ClassSEPLand,
		Notes: "fixed the registration",
	})
	if err != nil {
		t.Fatalf("UpdateAircraft: %v", err)
	}
	if got.Registration != "OH-XYZ" || got.Notes != "fixed the registration" {
		t.Errorf("UpdateAircraft returned %+v", got)
	}
	if _, err := db.AircraftByReg("OH-XZY"); !errors.Is(err, store.ErrAircraftNotFound) {
		t.Error("the old registration is still present after being renamed")
	}
}

func TestUpdateAircraftRefusesToCollideWithAnother(t *testing.T) {
	db := aircraftDB(t)
	if _, err := db.AddAircraft(newAircraft("OH-XYZ", "C152", csvbook.ClassSEPLand)); err != nil {
		t.Fatal(err)
	}
	_, err := db.UpdateAircraft("OH-XYZ", newAircraft("OH-CTL", "C172", csvbook.ClassSEPSea))
	if !errors.Is(err, store.ErrDuplicateAircraft) {
		t.Fatalf("renaming onto an existing registration returned %v, want ErrDuplicateAircraft", err)
	}
}

func TestUpdateAircraftOnSomethingThatIsNotThere(t *testing.T) {
	db := aircraftDB(t)
	_, err := db.UpdateAircraft("OH-NOPE", newAircraft("OH-NOPE", "C152", csvbook.ClassSEPLand))
	if !errors.Is(err, store.ErrAircraftNotFound) {
		t.Fatalf("err = %v, want ErrAircraftNotFound", err)
	}
}

// THE PROPERTY THAT MATTERS MOST. The aircraft table is a seed list for a form.
// Flights carry their own registration, type and class as written on paper, so
// nothing about editing an aeroplane may move a single minute of the legal
// record (CLAUDE.md rules 0.2 and 0.8).
func TestEditingAnAircraftMovesNoFlightAndNoTotal(t *testing.T) {
	db := aircraftDB(t)

	before, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.AddAircraft(newAircraft("OH-XYZ", "C152", csvbook.ClassSEPLand)); err != nil {
		t.Fatal(err)
	}
	// Rewrite an aeroplane that HAS flights against it, including its class --
	// the sea/land split is the most dangerous field on the row.
	if _, err := db.UpdateAircraft("OH-CTL", csvbook.Aircraft{
		Registration: "OH-CTL", Type: "PA18", DefaultClass: csvbook.ClassSEPLand,
	}); err != nil {
		t.Fatal(err)
	}

	after, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Fatalf("flight count moved from %d to %d", len(before), len(after))
	}
	for i := range before {
		if before[i] != after[i] {
			t.Errorf("flight %d changed when its aircraft was edited:\n before %+v\n after  %+v",
				before[i].Seq, before[i], after[i])
		}
	}
}

// Ordering is the replacement for the retired/active idea the owner dropped:
// the aeroplane flown most recently is the one most likely to be flown next,
// and one just added has not been flown at all -- it was added BECAUSE it is
// about to be.
func TestAircraftListPutsTheNewestAndMostRecentlyFlownFirst(t *testing.T) {
	db := aircraftDB(t)
	if _, err := db.AddAircraft(newAircraft("OH-XYZ", "C152", csvbook.ClassSEPLand)); err != nil {
		t.Fatal(err)
	}

	list, err := db.AircraftList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 3 {
		t.Fatalf("expected at least 3 aircraft, got %d", len(list))
	}
	if list[0].Registration != "OH-XYZ" {
		t.Errorf("a never-flown aeroplane is at position of %q, want it first: %v",
			list[0].Registration, regs(list))
	}
	// The rest run newest-flown to oldest.
	flown := list[1:]
	for i := 1; i < len(flown); i++ {
		if flown[i-1].LastFlown < flown[i].LastFlown {
			t.Errorf("out of order at %d: %q (%s) before %q (%s)",
				i, flown[i-1].Registration, flown[i-1].LastFlown,
				flown[i].Registration, flown[i].LastFlown)
		}
	}
}

// last flown and the flight count are DERIVED, never stored (rule 0.5 in
// spirit: the same reason cumulative totals are computed).
func TestAircraftListDerivesWhatWasFlownAndWhen(t *testing.T) {
	db := aircraftDB(t)

	list, err := db.AircraftList()
	if err != nil {
		t.Fatal(err)
	}
	byReg := map[string]store.AircraftRow{}
	for _, a := range list {
		byReg[a.Registration] = a
	}

	ctl, ok := byReg["OH-CTL"]
	if !ok {
		t.Fatal("OH-CTL is missing from the list")
	}
	if ctl.Flights != 1 {
		t.Errorf("OH-CTL reports %d flights, want 1", ctl.Flights)
	}
	if ctl.LastFlown != "2021-06-01" {
		t.Errorf("OH-CTL last flown %q, want 2021-06-01", ctl.LastFlown)
	}
	if ctl.UserAdded {
		t.Error("OH-CTL came from the books and must not be marked user-added")
	}
}

// The importer is being retired from production, but it still exists for
// scratch databases and tests -- and its DELETE was unqualified, which would
// take every hand-added aeroplane with it. Same trap source_book = 0 solves
// for flights.
func TestAHandAddedAircraftSurvivesAReimport(t *testing.T) {
	db := aircraftDB(t)
	if _, err := db.AddAircraft(newAircraft("OH-XYZ", "C152", csvbook.ClassSEPLand)); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Import(sample(t), "re-import"); err != nil {
		t.Fatalf("re-import: %v", err)
	}

	got, err := db.AircraftByReg("OH-XYZ")
	if err != nil {
		t.Fatalf("the hand-added aeroplane did not survive the re-import: %v", err)
	}
	if !got.UserAdded {
		t.Error("it survived but lost its provenance")
	}
	// And the imported ones are still rebuilt, exactly once each.
	list, err := db.AircraftList()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, a := range list {
		seen[a.Registration]++
	}
	for reg, n := range seen {
		if n != 1 {
			t.Errorf("%s appears %d times after a re-import", reg, n)
		}
	}
}

func regs(list []store.AircraftRow) []string {
	out := make([]string, 0, len(list))
	for _, a := range list {
		out = append(out, a.Registration)
	}
	return out
}
