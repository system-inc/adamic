package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A record spread from a source that may be undefined or null lowers where the record alone does,
// and the missing arm copies nothing, as JavaScript's { ...undefined } does (#bzrdem0, tracing.ts:183's
// { ...args, results }). Each program agrees with Node natively, sanitized and on JavaScript, in .ts
// and in .a: undefined_args (s1), overwritten_key (s2), null_args, and tracing_write_stack_event (s3,
// .ts, tsc's own shape). A fixed object that may be missing keeps main's stop, pinned in internal/lower's
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
			loomViewsDAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 7 {
		t.Fatalf("%d programs in %s, want 7: a program added or removed needs this count changed with it", programs, directory)
	}
}
