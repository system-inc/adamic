package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// A mutable write can store a value hiding a field the slot's readers expect (#4ysqm2m). append(held,
// hidden), with held a Target[] and hidden a Source, infers append's T as Source, the common supertype, so
// held goes in as a Source[], and append's push stores a Source whose hidden label held's reads then take as
// a Target's string: on main it lowered and native segfaulted where Node printed 5. The invariant-mutable rule
// judged the write by assignability alone, which a Source passes; it now runs the hiding rule's walk in the
// write direction too (writeHidesField, internal/lower/invariance.go), on a mutable container's element and a
// writable field alike. Each program's row says what it shows:
//
//   - stops: refused at compile time, in .ts and .a alike, where the stop holds the row's words: the lie
//     (append_lie.ts and its .a twin), the honest Source values the type lets through all the same
//     (append_source_honest.ts, the fix's price, never a false refusal) and the writable-field road
//     (field_lie.ts);
//   - agrees: it lowers, and release native, sanitized native and the JavaScript backend print what Node
//     prints: an append whose values are Targets (append_honest.ts) and a readonly view no write goes through
//     (readonly_view.ts);
//   - diagnoses: tsc refuses it with the row's diagnostic, so the inference the fix stands on is the
//     checker's own (inference.ts: T is Source, so label doesn't exist on what append returns).
//
// Mutant (internal/lower/generic-slot-hiding.mutants.md): writeHidesField returning false turns append_lie.ts,
// append_lie.a, append_source_honest.ts and field_lie.ts red.
const genericSlotHidingDirectory = "internal/oracle/testdata/generic_slot_hiding"

type genericSlotHidingRow struct{ kind, detail string }

const invariantMutableFix = "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"

var genericSlotHidingRows = map[string]genericSlotHidingRow{
	"append_lie.ts":           {kind: "stops", detail: genericSlotHidingDirectory + "/append_lie.ts:26:8: Adamic 0.1 refuses a value of type Target[] seen as Source[] | undefined, which can write Source where Target is read; " + invariantMutableFix},
	"append_lie.a":            {kind: "stops", detail: genericSlotHidingDirectory + "/append_lie.a:22:8: Adamic 0.1 refuses a value of type Target[] seen as Source[] | undefined, which can write Source where Target is read; " + invariantMutableFix},
	"append_source_honest.ts": {kind: "stops", detail: genericSlotHidingDirectory + "/append_source_honest.ts:25:8: Adamic 0.1 refuses a value of type Target[] seen as Source[] | undefined, which can write Source where Target is read; " + invariantMutableFix},
	"field_lie.ts":            {kind: "stops", detail: genericSlotHidingDirectory + "/field_lie.ts:14:22: Adamic 0.1 refuses a value of type Holder seen as Loose, which can write Source where Target is read; " + invariantMutableFix},
	"append_honest.ts":        {kind: "agrees"},
	"readonly_view.ts":        {kind: "agrees"},
	"inference.ts":            {kind: "diagnoses", detail: genericSlotHidingDirectory + "/inference.ts:21:21: error TS2339: Property 'label' does not exist on type 'Source'"},
}

func TestGenericSlotHiding(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(filepath.Join(repository, genericSlotHidingDirectory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		row, listed := genericSlotHidingRows[entry.Name()]
		if !listed {
			t.Errorf("%s has no row: say what it shows", entry.Name())
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := genericSlotHidingDirectory + "/" + entry.Name()
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			switch row.kind {
			case "agrees":
				loomViewsDAgrees(t, path)
			case "stops":
				_, err := lowered(t, absolute)
				if err == nil {
					t.Fatalf("lowered, want a stop holding %q", row.detail)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); !strings.Contains(got, row.detail) {
					t.Fatalf("stop: got %q, want one holding %q", got, row.detail)
				}
			case "diagnoses":
				_, err := load.Load([]string{absolute})
				if err == nil {
					t.Fatalf("type-checked, want tsc's %q", row.detail)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); !strings.Contains(got, row.detail) {
					t.Fatalf("diagnostic: got %q, want one holding %q", got, row.detail)
				}
			default:
				t.Fatalf("row kind %q", row.kind)
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != len(genericSlotHidingRows) {
		t.Fatalf("found %d programs, want %d", programs, len(genericSlotHidingRows))
	}
}
