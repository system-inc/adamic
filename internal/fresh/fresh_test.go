package fresh_test

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Every write in every program the oracle runs is one lowering recorded (so its holder's type is
// known) and one the proof knows how to judge: a write it can't place would keep every slot of its
// kind refused, and an IR node it doesn't know would keep every slot refused, both silently stricter
// than they need be. A new IR node shows up here first.
func freshPrograms(test *testing.T) []string {
	test.Helper()
	var paths []string
	for _, pattern := range []string{
		"../../dedication/dedication.a",
		"../load/testdata/0.1/compile/*.ts",
		"../load/testdata/0.1/compile/07_modules/main.ts",
		"../oracle/testdata/*.a",
		"../oracle/testdata/*.ts",
		"../oracle/testdata/non_null_refused/non_null_possible_*.a",
		"../oracle/testdata/non_null_refused/non_null_refuse_*.a",
		"../oracle/testdata/modules/main.a",
		"../flow/testdata/*.a",
	} {
		matches, globError := filepath.Glob(pattern)
		if globError != nil {
			test.Fatal(globError)
		}
		paths = append(paths, matches...)
	}
	// The oracle's .ts twins are its .a programs checked, so they stay in; its non-null .ts programs were never here.
	paths = slices.DeleteFunc(paths, func(path string) bool {
		return filepath.Ext(path) == ".ts" && strings.HasPrefix(filepath.Base(path), "non_null")
	})
	if len(paths) < 60 {
		test.Fatalf("found only %d programs: the globs no longer find the fixtures", len(paths))
	}
	return paths
}

func checkFreshProgram(test *testing.T, path string) {
	test.Helper()
	writes, proven := 0, 0
	absolute, absolutePathError := filepath.Abs(path)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	// This existing corpus fixture contains Map.set and Array.push reference writes.
	// Unlike a refused fixture or a scalar-only program, its proof cannot be empty.
	floorPath, absolutePathError := filepath.Abs("../flow/testdata/mutations.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	needsWrites := absolute == floorPath
	program, loadError := load.Load([]string{absolute})
	if loadError != nil {
		test.Fatalf("%s: Load: %v", path, loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		if needsWrites {
			test.Fatalf("%s: known-write corpus fixture must lower: %v", path, lowerError)
		}
		// The oracle's fixtures that stage 0 refuses on purpose.
		test.Logf("Lower declined: %v", lowerError)
		return
	}
	for _, write := range fresh.ProveWrites(lowered) {
		writes++
		if write.Proven {
			proven++
		}
		if write.Kind == fresh.WriteUnknown {
			test.Errorf("%s: %s", path, write.Why)
		}
		if write.Site == 0 && write.Kind != fresh.WriteUnknown {
			test.Errorf("%s: a write lowering didn't record, in function %d", path, write.Function)
		}
	}
	if needsWrites && writes == 0 {
		test.Errorf("%s: known-write corpus fixture reported zero writes", path)
	}
	test.Logf("%d writes, %d proven not to close a cycle", writes, proven)
}

// Methods use outside parameters, so passing a confined node must let it escape
// even when the method's return is a scalar. The global can then reach its array.
func TestMethodKeepsArgument(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../oracle/testdata/fresh_refused/devirt_fresh_method_keeps_argument.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatalf("graph program must be admitted: %v", lowerError)
	}
	found := false
	for _, write := range fresh.ProveWrites(program) {
		if write.Kind != fresh.WriteElement {
			continue
		}
		found = true
		if write.Proven || write.Why == "" {
			test.Fatalf("method-kept argument must escape before the push: %+v", write)
		}
		test.Logf("method-kept argument: %+v", write)
	}
	if !found {
		test.Fatal("missing cycle-closing element write")
	}
}
