package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// An optional call of an optional method the program declares (#tvq1eqm, lower/optional_method_call.go):
// each program in testdata/optional_method_call agrees with Node natively, sanitized and on the
// JavaScript backend, and leaks nothing. tsc's own performance.ts shape on a system that's missing,
// one without the method and ones with it; a receiver evaluated once, undefined, null and present, with
// an argument that counts itself and is never evaluated where the chain stops; and what the lookup
// finds: nothing, undefined in a field, an arrow, a literal's method reading this and a class's method,
// for number, string, boolean and void results; and a method that throws inside the optional call,
// caught by its caller, so the aside's temporaries and the preset result are each released once on the
// throw's way out (Compiler's pin). Node's own output is pinned too, so a fixture that
// stopped exercising its rows shows here before it shows as an agreement.
func TestOptionalMethodCallsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, program := range []struct{ name, want string }{
		{"tsc_shape.a", "no system: none native undefined 0\nbare: none native undefined 0\ndebugging: native native undefined 0\nprofiling: native native true 2\nquiet: none native false 2\n"},
		{"receivers.a", "hello!, ada lookups 1 evaluated 1\nnobody lookups 2 evaluated 1\nnobody lookups 3 evaluated 1\nnull lookups 4 evaluated 1\nhi!, ada lookups 5 evaluated 2\nnull lookups 6 evaluated 2\nas a statement: lookups 7 evaluated 3\nas a statement: lookups 8 evaluated 3\n"},
		{"throws.a", "held:a1+2!|held:a1\ncaught bad held:b2+2\nnone|held:c3\ncaught bad held:d4+2\n"},
		{"methods.a", "empty: -1 undescribed undefined 0\ncleared: -1 undescribed undefined 0\narrow: 5 as#1 arrow true 1\nliteral logs line#3\nliteral: 70 as#2 literal undefined 3\nhammer logs line#5\nhammer: 6 as#4 hammer undefined 5\n5 picked#6 arrow true 6\n-1 none undefined 6\n6 picked#7 hammer undefined 7\n-1 none undefined 7\n"},
	} {
		t.Run(program.name, func(t *testing.T) {
			t.Parallel()
			relative := "internal/oracle/testdata/optional_method_call/" + program.name
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); string(node.stdout) != program.want || node.exitCode != 0 {
				t.Fatalf("Node: exit %d, stdout %q, want exit 0 and %q", node.exitCode, node.stdout, program.want)
			}
			// generatorAgrees holds any program to Node on every backend and to leaking nothing.
			generatorAgrees(t, relative)
		})
	}
}

// Check 49's road: a checked cast lets in an object whose enabled is a number where the type says a
// function or undefined, and Node stops at the call with "viewed.enabled is not a function" (the
// oracle's runner reports an uncaught throw as a panic, exit 70). The view's checked read has no
// contract for an optional callable yet, so lowering stops there and check 49 has nothing to fire on.
// Should it ever lower, it is held to Node on every backend, so a call through a non-function shows as
// the disagreement it is. Mutant: the optional call's stop in unemittableCheckedViewRead dropped.
func TestOptionalMethodCallThroughAViewIsNotYet(t *testing.T) {
	t.Parallel()
	relative := "internal/oracle/testdata/optional_method_call/unproven/view_hatch.ts"
	path, err := filepath.Abs(filepath.Join(repository, relative))
	if err != nil {
		t.Fatal(err)
	}
	if want := onNode(t, path); string(want.stdout) != "true\n" || string(want.stderr) != "adamic: panic: TypeError: viewed.enabled is not a function\n" || want.exitCode != 70 {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q, want the hatch's number called and refused", want.exitCode, want.stdout, want.stderr)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	var refused *lower.Refused
	if errors.As(err, &notYet) || errors.As(err, &refused) {
		t.Logf("stopped: %v", err)
		return
	}
	if err != nil {
		t.Fatalf("got %v, want a not yet or a refusal at the view's read", err)
	}
	generatorAgrees(t, relative)
	t.Error("lowered, want a not yet or a refusal at the view's read")
}
