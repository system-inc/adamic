package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// A library call through ?. whose result is a number or a boolean holds that result as a maybe,
// since optionalIntrinsic tests the receiver whatever its type says: a record's value read can be
// missing where the type holds none (#qgthc6g, Language's find on #t5q4prw). Held as the scalar the
// checker types it, the absent arm was undefined in a double, C clang refused, and in a boolean it
// read false where Node reads undefined. record_cycle.ts is Language's program, closing
// a -> r -> array -> a, record_acyclic.ts its twin with no cycle, and record_missing.ts reads keys
// the record lacks, each call there undefined on Node. The rest are receivers the checker proves
// present: a local array, a string and a Map's asserted value. Each runs against Node natively,
// sanitized with the leak check, on the JavaScript backend and by the counted build.
func TestOptionalScalarResults(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/optional_scalar_result"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			graphOracleWitness(t, filepath.Join(directory, entry.Name()))
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 6 {
		t.Fatalf("%d programs in %s, want 6", programs, directory)
	}
}
