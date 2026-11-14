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
// fixed one, so r2 is NotYet at the view. Each stop is pinned by its kind and line:col, never by
// another rule's words. The neighbors keep theirs: an unknown rest (r1, r4) still stops in its body, a
// string rest through string parameters (r3, the string control) and a fixed unknown through a string
// view (r5) lower and agree with Node on release, sanitized native and the JavaScript backend, and an
// optional unknown through a number view (r6) still stops at its view. Mutant: the rest judgment
// dropped from sameKeeping lowers r2 again, and clang refuses its C.
func TestRestThroughFixedView(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/rest_fixed_view"
	// Each program's NotYet as line:col, or "" where it lowers and agrees.
	stops := map[string]string{
		"r1_unknown_rest":                    ":1:50",
		"r2_union_rest":                      ":2:49",
		"r3_same_rest_control":               "",
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
