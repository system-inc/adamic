package fresh_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Every write in every program the oracle runs is one lowering recorded (so its holder's type is
// known) and one the proof knows how to judge: a write it can't place would keep every slot of its
// kind refused, and an IR node it doesn't know would keep every slot refused, both silently stricter
// than they need be. A new IR node shows up here first.
func TestEveryWriteIsRecordedAndKnown(t *testing.T) {
	t.Parallel()
	var paths []string
	for _, pattern := range []string{
		"../../dedication/dedication.a",
		"../load/testdata/0.1/compile/*.ts",
		"../load/testdata/0.1/compile/07_modules/main.ts",
		"../oracle/testdata/*.a",
		"../oracle/testdata/modules/main.a",
		"../flow/testdata/*.a",
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	if len(paths) < 60 {
		t.Fatalf("found only %d programs: the globs no longer find the fixtures", len(paths))
	}
	writes, proven := 0, 0
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{absolute})
		if err != nil {
			t.Fatalf("%s: Load: %v", path, err)
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			// The oracle's fixtures that stage 0 refuses on purpose.
			continue
		}
		for _, write := range fresh.ProveWrites(lowered) {
			writes++
			if write.Proven {
				proven++
			}
			if write.Kind == fresh.WriteUnknown {
				t.Errorf("%s: %s", path, write.Why)
			}
			if write.Site == 0 && write.Kind != fresh.WriteUnknown {
				t.Errorf("%s: a write lowering didn't record, in function %d", path, write.Function)
			}
		}
	}
	t.Logf("%d writes in %d programs, %d proven not to close a cycle", writes, len(paths), proven)
}
