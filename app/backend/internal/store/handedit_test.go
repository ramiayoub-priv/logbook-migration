package store_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ramiayoub/logbook/backend/internal/csvbook"
	"github.com/ramiayoub/logbook/backend/internal/store"
)

// The edit and delete paths, added 2026-08-02 when the app became the only way
// the record grows.
//
// The load-bearing rule throughout: these touch HAND-ENTERED rows and nothing
// else. The 1296 imported flights are closed data (CLAUDE.md rule 0.8), and a
// change to one of them would also be silently discarded by the next import,
// which owns every row with source_book <> 0.

func TestUpdateFlightReplacesEveryEditableField(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	added := mustAdd(t, db, handEntered())

	edited := handEntered()
	edited.Date = "2026-07-31"
	edited.AircraftReg = "OH-CTL"
	edited.DepPlace = "EFNU"
	edited.ArrPlace = "EFPR"
	edited.OffBlockRaw, edited.OnBlockRaw = "11:00Z", "12:30Z"
	edited.BlockMinutes, edited.TotalMinutes, edited.PICMinutes = 90, 90, 90
	edited.NightMinutes = 20
	edited.LandingsDay = 5
	edited.Remarks = "corrected"

	updated, err := db.UpdateFlight(added.Seq, edited, 1)
	if err != nil {
		t.Fatalf("UpdateFlight: %v", err)
	}

	// The seq is the book's own order and is NOT the pilot's to change by
	// editing a flight: it is what every cumulative computation walks.
	if updated.Seq != added.Seq {
		t.Errorf("seq moved on an edit: %d -> %d", added.Seq, updated.Seq)
	}
	got := findBySeq(t, db, added.Seq)
	if got.Date != "2026-07-31" || got.AircraftReg != "OH-CTL" || got.TotalMinutes != 90 ||
		got.NightMinutes != 20 || got.LandingsDay != 5 || got.Remarks != "corrected" {
		t.Errorf("edit did not land: %+v", got)
	}
	// Still hand-entered, so the importer still knows it may not delete it.
	if got.SourceBook != 0 {
		t.Errorf("source_book changed to %d; the row would be deleted by the next import", got.SourceBook)
	}
}

// The aircraft link is looked up again, exactly as AddFlight does it. Without
// this an edited registration keeps pointing at the old aeroplane.
func TestUpdateFlightRelinksTheAircraft(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	added := mustAdd(t, db, handEntered())

	edited := handEntered()
	edited.AircraftReg = "OH-CTL"
	if _, err := db.UpdateFlight(added.Seq, edited, 1); err != nil {
		t.Fatalf("UpdateFlight: %v", err)
	}

	linked, err := db.FlightAircraftLinked(added.Seq)
	if err != nil {
		t.Fatalf("FlightAircraftLinked: %v", err)
	}
	if !linked {
		t.Error("the edited flight lost its aircraft link")
	}
}

// The whole point of the source_book restriction. An imported row is closed
// data; refusing here means no HTTP handler, present or future, can reach one.
func TestUpdateFlightRefusesAnImportedRow(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	imported, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	paper := imported[0]
	if paper.SourceBook == 0 {
		t.Fatalf("test setup: seq %d is not an imported row", paper.Seq)
	}

	edit := paper
	edit.Remarks = "should never be written"
	if _, err := db.UpdateFlight(paper.Seq, edit, 1); !errors.Is(err, store.ErrNotHandEntered) {
		t.Fatalf("UpdateFlight on an imported row: got %v, want ErrNotHandEntered", err)
	}
	if got := findBySeq(t, db, paper.Seq); got.Remarks == "should never be written" {
		t.Error("an imported row was modified")
	}
}

func TestDeleteFlightRefusesAnImportedRow(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	all, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	before := len(all)
	paper := all[0]

	if _, err := db.DeleteFlight(paper.Seq, 1); !errors.Is(err, store.ErrNotHandEntered) {
		t.Fatalf("DeleteFlight on an imported row: got %v, want ErrNotHandEntered", err)
	}
	after, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != before {
		t.Errorf("the flight count moved: %d -> %d", before, len(after))
	}
}

func TestUpdateAndDeleteReportAMissingFlight(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.UpdateFlight(9_999_999, handEntered(), 1); !errors.Is(err, store.ErrFlightNotFound) {
		t.Errorf("UpdateFlight on a missing seq: got %v, want ErrFlightNotFound", err)
	}
	if _, err := db.DeleteFlight(9_999_999, 1); !errors.Is(err, store.ErrFlightNotFound) {
		t.Errorf("DeleteFlight on a missing seq: got %v, want ErrFlightNotFound", err)
	}
}

// An edit must not be able to create the duplicate that AddFlight refuses --
// two identical rows inflate a licence total either way.
func TestUpdateFlightRefusesADuplicateOfAnotherRow(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	first := mustAdd(t, db, handEntered())

	second := handEntered()
	second.OffBlockRaw, second.OnBlockRaw = "14:00Z", "15:00Z"
	added := mustAdd(t, db, second)

	// Editing the second flight onto the first one's date/aircraft/off-block.
	clash := handEntered()
	if _, err := db.UpdateFlight(added.Seq, clash, 1); !errors.Is(err, store.ErrDuplicateFlight) {
		t.Fatalf("UpdateFlight onto another row's key: got %v, want ErrDuplicateFlight", err)
	}
	_ = first
}

// Saving a flight unchanged is an edit that must succeed. The duplicate check
// has to exclude the row being edited, or every no-op save is a 409.
func TestUpdateFlightAllowsSavingARowOntoItself(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	added := mustAdd(t, db, handEntered())

	same := handEntered()
	same.Remarks = "only the remark changed"
	if _, err := db.UpdateFlight(added.Seq, same, 1); err != nil {
		t.Fatalf("saving a row onto itself: %v", err)
	}
	if got := findBySeq(t, db, added.Seq); got.Remarks != "only the remark changed" {
		t.Errorf("the remark did not change: %q", got.Remarks)
	}
}

func TestDeleteFlightRemovesItAndTheTotalsFollow(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	before, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	added := mustAdd(t, db, handEntered())

	deleted, err := db.DeleteFlight(added.Seq, 1)
	if err != nil {
		t.Fatalf("DeleteFlight: %v", err)
	}
	// What was removed is returned, so the caller can log and report it.
	if deleted.Seq != added.Seq || deleted.AircraftReg != "OH-CAM" {
		t.Errorf("DeleteFlight returned %+v", deleted)
	}

	after, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Errorf("flights after delete = %d, want %d (back to the imported set)", len(after), len(before))
	}
	for _, f := range after {
		if f.Seq == added.Seq {
			t.Error("the deleted flight is still in the book")
		}
	}
}

// --- The audit trail -------------------------------------------------------
//
// The owner asked for a standard in-place edit and a delete. On a legal record
// that is only safe if what was there before survives somewhere: an edit that
// changes a licence total with no trace of the previous value is precisely the
// drift this project exists to have stopped. The audit table is append-only
// and nothing in the app reads it -- it is there for the day someone asks
// "what did this row say before?".

func TestAnEditIsAudited(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	added := mustAdd(t, db, handEntered())

	edited := handEntered()
	edited.TotalMinutes, edited.PICMinutes = 90, 90
	if _, err := db.UpdateFlight(added.Seq, edited, 7); err != nil {
		t.Fatal(err)
	}

	entries, err := db.FlightAudit(added.Seq)
	if err != nil {
		t.Fatalf("FlightAudit: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(entries))
	}
	a := entries[0]
	if a.Action != "update" || a.UserID != 7 || a.Seq != added.Seq {
		t.Errorf("audit entry = %+v", a)
	}
	// The BEFORE state, not the after: 75 minutes is what the row said.
	if !strings.Contains(a.Before, `"total_minutes":75`) {
		t.Errorf("the audit entry does not carry the previous total: %s", a.Before)
	}
	if a.At == "" {
		t.Error("the audit entry has no timestamp")
	}
}

func TestADeleteIsAuditedWithTheWholeRow(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	added := mustAdd(t, db, handEntered())

	if _, err := db.DeleteFlight(added.Seq, 7); err != nil {
		t.Fatal(err)
	}

	entries, err := db.FlightAudit(added.Seq)
	if err != nil {
		t.Fatalf("FlightAudit: %v", err)
	}
	if len(entries) != 1 || entries[0].Action != "delete" {
		t.Fatalf("audit entries = %+v, want one delete", entries)
	}
	// Everything needed to put the flight back, because there is nowhere else
	// left to read it from.
	for _, want := range []string{`"date":"2026-07-30"`, `"aircraft_reg":"OH-CAM"`,
		`"off_block_raw":"09:15Z"`, `"total_minutes":75`, `"landings_day":3`} {
		if !strings.Contains(entries[0].Before, want) {
			t.Errorf("the deleted row's audit copy is missing %s: %s", want, entries[0].Before)
		}
	}
}

// A refused edit must not leave a trace suggesting something happened.
func TestARefusedEditIsNotAudited(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	all, err := db.Flights()
	if err != nil {
		t.Fatal(err)
	}
	paper := all[0]

	edit := paper
	edit.Remarks = "no"
	if _, err := db.UpdateFlight(paper.Seq, edit, 1); !errors.Is(err, store.ErrNotHandEntered) {
		t.Fatal(err)
	}
	entries, err := db.FlightAudit(paper.Seq)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("a refused edit wrote %d audit entries", len(entries))
	}
}

// The importer owns every row with source_book <> 0 and replaces them on every
// run. An edited hand-entered flight must survive that, exactly as a newly
// added one does -- otherwise a correction has a lifespan of one re-import.
func TestAnEditedFlightSurvivesAReimport(t *testing.T) {
	db := openTemp(t)
	if _, err := db.Import(sample(t), "seed"); err != nil {
		t.Fatal(err)
	}
	added := mustAdd(t, db, handEntered())

	edited := handEntered()
	edited.Remarks = "edited before the re-import"
	if _, err := db.UpdateFlight(added.Seq, edited, 1); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Import(sample(t), "re-import"); err != nil {
		t.Fatal(err)
	}
	got := findBySeq(t, db, added.Seq)
	if got.Remarks != "edited before the re-import" {
		t.Errorf("the edit did not survive the re-import: %q", got.Remarks)
	}
}

var _ = csvbook.Flight{}
