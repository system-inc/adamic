package oracle

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// A call whose function value is known is laid out for that function alone (ClosureArgumentLayout,
// #g4arkr4). The type's whole set holds every closure assignable to the callback's type, and
// (...items: string[]) => string is assignable to map's (item: string) => string, so a map of an
// identity literal in a rest arrow packed the arrow's own tail for a callback that takes none, inside
// map's loop, and clang refused the C: the pack's release is the statement's, outside the loop's block
// (probe 3, J-nxq9d78: q1's FunctionTypeTargets [0 1] against ClosureTargets [1]).
//
// Each program's outcome is pinned: "agrees" (Node, native under ASan with the leak check, the release
// build and the JavaScript backend all print the same), "clang" (lowers, and clang refuses its C), or
// "NotYet:line:col". The arrow (q1) and the function expression (q7) agree. A known comparator with no
// rest beside a rest closure of its type takes the runtime comparator, its sort line naming
// adamic_compare_closure, where the type's set gave it an adapter; rest_sort.a's rest comparator keeps
// its adapter. A bound rest arrow through map stops as on main, at the bind (2:15, "bind of a rest-parameter
// function"), so the Bound branch has no live witness yet and stays as defense. A callback that comes in through a
// parameter isn't known, so it's laid out by the type's set, still packs the rest arrow's tail inside
// map's loop, and clang still refuses it: that's the emitter's loop scope (#9grybqc), and the pin moves
// to agrees when it lands. Mutant: ClosureArgumentLayout reading FunctionTypeTargets first again, and
// q1 and q7 reach clang and the known comparator takes the adapter.
func TestClosureLayoutKnownTargets(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path    string
		outcome string
		// sort is what every sort line of the emitted C says of the runtime comparator: "runtime" where
		// each passes adamic_compare_closure, "adapter" where none does, "" where it isn't asked.
		sort string
	}{
		{"internal/oracle/testdata/closure_layout_known/q1_map_identity.ts", "agrees", ""},
		{"internal/oracle/testdata/closure_layout_known/q7_function_expression.ts", "agrees", ""},
		{"internal/oracle/testdata/closure_layout_known/known_comparator.ts", "agrees", "runtime"},
		{"internal/oracle/testdata/comparator_rest/rest_sort.a", "agrees", "adapter"},
		{"internal/oracle/testdata/closure_layout_known/bound_rest_map.ts", "NotYet:2:15", ""},
		{"internal/oracle/testdata/closure_layout_known/parameter_callback.ts", "clang", ""},
	}
	for _, pinned := range cases {
		name := filepath.Base(pinned.path)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, pinned.path))
			if err != nil {
				t.Fatal(err)
			}
			outcome, code := closureLayoutOutcome(t, absolute, name)
			t.Logf("PIN %s outcome=%s", name, outcome)
			if outcome != pinned.outcome {
				t.Errorf("want %s, got %s", pinned.outcome, outcome)
			}
			if pinned.sort == "" {
				return
			}
			sorts, runtimeSorts := 0, 0
			for _, line := range strings.Split(code, "\n") {
				if strings.Contains(line, "adamic_array_sort(") {
					sorts++
					if strings.Contains(line, ", adamic_compare_closure, ") {
						runtimeSorts++
					}
				}
			}
			t.Logf("PIN %s sorts=%d runtime=%d", name, sorts, runtimeSorts)
			switch {
			case sorts == 0:
				t.Errorf("no sort line in the emitted C")
			case pinned.sort == "runtime" && runtimeSorts != sorts:
				t.Errorf("want every sort to take adamic_compare_closure, %d of %d do", runtimeSorts, sorts)
			case pinned.sort == "adapter" && runtimeSorts != 0:
				t.Errorf("want every sort through its adapter, %d of %d take adamic_compare_closure", runtimeSorts, sorts)
			}
		})
	}
}

// closureLayoutOutcome is a program's outcome as TestClosureLayoutKnownTargets pins it, and its C when
// it lowers.
func closureLayoutOutcome(t *testing.T, path, name string) (string, string) {
	t.Helper()
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
	}
	program, err := lowered(t, path)
	var notYet *lower.NotYet
	if errors.As(err, &notYet) {
		t.Logf("PIN %s not yet: %v", name, notYet)
		at := strings.LastIndex(notYet.Where, name)
		if at < 0 {
			return "NotYet " + notYet.Where, ""
		}
		return "NotYet" + notYet.Where[at+len(name):], ""
	}
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	code := native.C(program)
	if err := native.Build(code, filepath.Join(t.TempDir(), "program"), native.Options{Sanitize: true}); err != nil {
		t.Logf("PIN %s %s", name, closureLayoutCut(strings.ReplaceAll(err.Error(), "\n", " / "), 600))
		return "clang", code
	}
	sanitized, leaked, comparable := nativelyCheckingLeaks(t, program, true)
	if leaked != "" || !comparable {
		return "leaks: " + leaked, code
	}
	for _, backend := range []struct {
		name string
		got  run
	}{{"sanitized", sanitized}, {"release", released(t, program)}, {"JavaScript", onJavaScriptBackend(t, program)}} {
		if difference := disagreement(truth, backend.got); difference != "" {
			return fmt.Sprintf("disagrees: %s %s: Node %q, %q", backend.name, difference, truth.stdout, backend.got.stdout), code
		}
	}
	return "agrees", code
}

// closureLayoutCut keeps a logged line short.
func closureLayoutCut(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length] + "..."
}
