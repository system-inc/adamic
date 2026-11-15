package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A record value the checker narrowed present, then removed by a call it doesn't see, read through ?.
// inside a try: Node's ?. reads undefined and nothing throws. The narrowed read used to throw its
// TypeError before ?.'s own test, so the try printed "caught" where Node goes on (#70kdt66, J-xgt44wh on
// main 89614cc0). A reference read now reaches ?. unchecked, as narrowed.go's defined leaves one, and each
// program agrees with Node natively, sanitized with the leak check, on the JavaScript backend and by the
// counted build. Mutant: records.go's reference return under optionalReceiver dropped, and includes.ts
// prints "caught" again.
func TestRecordOptionalNarrowedAgrees(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/record_optional_narrowed"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			graphOracleWitness(t, filepath.Join(directory, entry.Name()))
		})
	}
	if programs != 4 {
		t.Fatalf("%d programs in %s, want 4", programs, directory)
	}
}

// A scalar record value narrowed present and read through ?. stops, since a maybe isn't shown to reach
// every ?. as one yet (#70kdt66): never the Defined's catchable TypeError.
func TestRecordOptionalNarrowedScalarStops(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/record_optional_narrowed_stop/scalar.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	want := "a record value of a scalar the checker narrowed present, read through ?."
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("got %v, want a stop containing %q", err, want)
	}
	if !strings.Contains(err.Error(), "scalar.ts:6:") {
		t.Fatalf("got %v, want the stop at scalar.ts:6", err)
	}
}
