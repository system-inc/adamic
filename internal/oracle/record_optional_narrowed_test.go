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
// counted build: library calls, length, an element and a user method (two of Compiler's three roads),
// and a template and an if. Mutant: records.go's optionalReceiver block dropped, and includes.ts prints
// "caught" again.
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
	if programs != 7 {
		t.Fatalf("%d programs in %s, want 7", programs, directory)
	}
}

// What still stops on the same shape, at compile time, never the Defined's catchable TypeError
// (#70kdt66): a scalar record value narrowed present and read through ?., since a maybe isn't shown to
// reach every ?. as one yet, and Compiler's third road, a class field read through ?., which stops at
// ?.'s own number result.
func TestRecordOptionalNarrowedStops(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, at, want string }{
		{"scalar.ts", "scalar.ts:6:", "a record value of a scalar the checker narrowed present, read through ?."},
		{"field.ts", "field.ts:12:", "?. to a number, which would be number | undefined"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/record_optional_narrowed_stop", row.name))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), row.want) || !strings.Contains(err.Error(), row.at) {
				t.Fatalf("got %v, want a stop at %s containing %q", err, row.at, row.want)
			}
		})
	}
}
