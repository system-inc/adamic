package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// loopBodyScopePrograms are #9grybqc's pins, in testdata/loop_body_scope: a rest packed for a callback
// inside a loop the emitter writes itself, map's (q1 and map), an array's forEach, reduce's, a Map's
// and a Set's forEach, Array.from's, the visits that break out early (search), reuse's map in place
// (mapped), and a throw from inside one (throw). Each program's callback is one with fixed parameters
// whose type admits a rest function of the program's, which is what makes the call's layout collect a
// rest (ir.ClosureArgumentLayout): a rest function handed to a library callback itself is NotYet
// (rest_fixed_view's r19), and Array.from refuses a first parameter that's a rest.
var loopBodyScopePrograms = []string{"q1", "map", "for_each", "reduce", "map_for_each", "array_from", "search", "mapped", "throw"}

// packedSlotBySlot is copyRest's C for a pack whose slots the rest doesn't hold alike, each slot read
// out of the call's compound literal by its index (restSlot): no other C reads a slot that way.
var packedSlotBySlot = regexp.MustCompile(`\}\[[0-9]+\]\.(reference|number|boolean|maybe_boolean)\b`)

// TestLoopBodyScope holds each pin to Node natively under ASan and UBSan, on the release build and on
// the JavaScript backend, and natively it must let go of everything: a pack let go of past the loop's
// closing brace is C clang refuses, and one never let go of is a leak (emit_ownership.go's iteration).
// Each program must pack a rest slot by slot, where the index's number is boxed into a union's rest or
// is left undefined in a rest that can't hold it, never retained as a reference (copyRest): should
// lowering stop packing these rests, the pins no longer reach what they hold, and this says so.
// Mutants are in testdata/loop_body_scope/mutants.md.
func TestLoopBodyScope(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/loop_body_scope"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		name, isProgram := strings.CutSuffix(entry.Name(), ".a")
		if !isProgram {
			continue
		}
		programs++
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if !slices.Contains(loopBodyScopePrograms, name) {
				t.Fatalf("%s isn't one of loopBodyScopePrograms", entry.Name())
			}
			absolute, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			if !packedSlotBySlot.MatchString(native.C(program)) {
				t.Fatalf("no rest is packed slot by slot in %s's C", entry.Name())
			}
			sanitized, binary := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Errorf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
			if leaked := leaks(t, program, binary); leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != len(loopBodyScopePrograms) {
		t.Fatalf("found %d programs, want %d", programs, len(loopBodyScopePrograms))
	}
}
