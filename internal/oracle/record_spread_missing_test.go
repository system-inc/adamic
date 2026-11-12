package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A record spread from a source that may be undefined or null lowers where the record alone does,
// and the missing arm copies nothing, as JavaScript's { ...undefined } does (#bzrdem0, tracing.ts:183's
// { ...args, results }). Each program agrees with Node natively, sanitized and on JavaScript, in .ts
// and in .a: undefined_args (s1), overwritten_key (s2) and null_args. tracing_write_stack_event (s3,
// .ts, tsc's own shape) lowers past the spread and stops at a later wall, JSON.stringify of object
// references (#d9ep7de), pinned below with its measured text; main stops it at the spread instead. A
// fixed object that may be missing keeps main's stop, pinned in internal/lower's
// TestRecordSpreadFromMissingFixedObjectStops (s4). Mutants: internal/lower/record-spread-missing.mutants.md.
func TestRecordSpreadFromMissingAgreesWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/record_spread_missing"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") && !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			if entry.Name() == "tracing_write_stack_event.ts" {
				recordSpreadFromMissingStopsAtJSON(t, directory+"/"+entry.Name())
				return
			}
			loomViewsDAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 7 {
		t.Fatalf("%d programs in %s, want 7: a program added or removed needs this count changed with it", programs, directory)
	}
}

// recordSpreadFromMissingStopsAtJSON pins tracing_write_stack_event.ts's stop. writeStackEvent, holding
// the spread, is declared before writeEvent and function bodies lower in declaration order, so a stop
// in writeEvent's JSON.stringify(args) at 28:37 is reached only after the spread on line 21 has lowered.
func recordSpreadFromMissingStopsAtJSON(t *testing.T, relative string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, relative))
	if err != nil {
		t.Fatal(err)
	}
	if node := onNode(t, path); node.exitCode != 0 {
		t.Fatalf("Node: %#v", node)
	}
	_, err = lowered(t, path)
	const want = "tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet"
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), want) {
		t.Errorf("got %v, want the spread lowered and JSON.stringify of object references to stop at %s", err, want)
	}
}
