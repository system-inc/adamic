package fresh_test

import (
	"context"
	"errors"
	"path/filepath"
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
func freshPrograms(t *testing.T) []string {
	t.Helper()
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
	return paths
}

func checkFreshProgram(t *testing.T, path string) {
	t.Helper()
	writes, proven := 0, 0
	// Runtime-only process probes are checked by their dedicated oracle tests.
	if filepath.Base(path) == "node_process_directory_mutation.a" || filepath.Base(path) == "node_process_environment_mutation.a" {
		return
	}
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
		t.Logf("Lower declined: %v", err)
		return
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
	t.Logf("%d writes, %d proven not to close a cycle", writes, proven)
}

// Methods use outside parameters, so passing a confined node must let it escape
// even when the method's return is a scalar. The global can then reach its array.
func TestMethodKeepsArgument(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../oracle/testdata/fresh_refused/devirt_fresh_method_keeps_argument.a")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	var refused *lower.Refused
	if !errors.As(err, &refused) {
		t.Fatalf("method-kept argument closes a cycle: want refusal, got %v", err)
	}
	if !strings.Contains(refused.Error(), "(adamic/cycle-capable)") || !strings.Contains(refused.Error(), "the write at "+path+":15:") {
		t.Fatalf("refusal does not name the cycle-closing push: %v", refused)
	}
}
