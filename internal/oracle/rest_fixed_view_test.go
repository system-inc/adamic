package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A closure with a rest parameter seen through a view with fixed parameters is called through
// ClosureArgumentLayout, whose caller packs the rest from its own argument slots with no conversion to
// the rest element's representation (copyRest, #2t3h708). (...items: (string | number)[]) seen as
// (a: string, b: number) => string handed b's raw double where a boxed union goes: Node and the
// JavaScript backend printed string:x,number:2, and native stopped at "clang failed" in .a and .ts (r2).
// The relation that judges a function value's parameters against a view's by representation
// (sameKeeping) now judges the rest's element against each view parameter it collects as it judges a
// fixed one, so r2 is NotYet at the view. And since any fixed view can be widened by arity in the next
// one, a rest function seen through a program view with no rest is NotYet at that view
// (restHiddenByView), so a string rest through string parameters (r3) stops too. Each stop is pinned by
// its kind and line:col, never by another rule's words. The neighbors keep theirs: an unknown rest (r1,
// r4) still stops in its body, a fixed unknown through a string view (r5) lowers and agrees with Node
// on release, sanitized native and the JavaScript backend, and an optional unknown through a number
// view (r6) still stops at its view. Mutants: the rest judgment dropped from sameKeeping lowers r2
// again, and clang refuses its C; restHiddenByView answering false lowers r17 again, into UBSan's
// mismatched function pointer.
func TestRestThroughFixedView(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/rest_fixed_view"
	// Each program's NotYet as line:col, or "" where it lowers and agrees.
	stops := map[string]string{
		"r1_unknown_rest":                    ":1:50",
		"r2_union_rest":                      ":2:49",
		"r3_same_rest_control":               ":2:49",
		"r4_unknown_tail_after_head":         ":1:77",
		"r5_fixed_unknown_through_string":    "",
		"r6_optional_unknown_through_number": ":2:39",

		// The other roads to the same relation, each measured with the fix (J-1zw6v14): an argument, a
		// return, a field, an element, and a view with a rest of its own, alone and after a fixed head (r12
		// and r13 reached clang on main too). A cast stops earlier, at the cast rule.
		"r7_road_argument":         ":3:19",
		"r8_road_return":           ":2:60",
		"r9_road_field":            ":2:62",
		"r10_road_element":         ":2:57",
		"r11_road_cast":            ":2:15",
		"r12_view_rest":            ":2:49",
		"r13_view_rest_after_head": ":2:60",

		// Language's attack on the first cut: a fixed view hides the rest, and the next view widens it
		// by arity, a relation of two fixed types that no longer shows the rest. Each stops at the
		// first view (r15 two declarations, r16 the second hop an argument, r17 a zero-parameter view,
		// which reached the closure through a mismatched function pointer under UBSan). A view with a
		// rest of its own (r14), a fixed function widened (r18) and a rest function handed to a library
		// callback, whose adapter calls the source signature (r19, Language's union element covering
		// map's three parameters), lower and agree. A class method with a rest seen through a method
		// with none (r20) is the same hidden rest, which reached the method through a mismatched
		// function pointer on main. A host function's own rest (setTimeout's, r21) is exempt, since its
		// value is a host lowering's closure, and seen through a narrower view and then widened it agrees.
		"r14_rest_view_control":           "",
		"r15_hidden_two_hops":             ":2:36",
		"r16_hidden_argument":             ":2:36",
		"r17_hidden_zero_hop":             ":2:28",
		"r18_fixed_widened_control":       "",
		"r19_library_callback_union_rest": "",
		"r20_hidden_class_method":         ":2:42",
		"r21_host_rest_widened":           "",
	}
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		extension := filepath.Ext(entry.Name())
		if extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			stop, pinned := stops[strings.TrimSuffix(entry.Name(), extension)]
			if !pinned {
				t.Fatalf("%s has no pinned stop", entry.Name())
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
			if stop != "" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, entry.Name()+stop) {
					t.Fatalf("want NotYet at %s%s, got %v", entry.Name(), stop, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 2*len(stops) {
		t.Fatalf("found %d programs, want %d", programs, 2*len(stops))
	}
}
