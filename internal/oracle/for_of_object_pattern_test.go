package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A for...of whose variable is an object pattern, over a plain array, holds each element once per step and
// destructures it by the declaration road (#rm22wzv): getters.ts reads each getter once per step in the pattern's
// order, and defaults_rename.ts takes a default and a rename, each beside Node on sanitized native, release native
// and the JavaScript backend. A nested pattern and a rest element keep the declaration road's stops, and an object
// pattern over a Set, no plain array, keeps "a for...of destructuring an object". Mutant: the destructure run after
// the loop's body instead of before it, so each step prints before its getters run, reds getters.ts against Node.
func TestForOfObjectPatternReadsAsADeclaration(t *testing.T) {
	t.Parallel()
	directory, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/for_of_object_pattern"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"getters.ts", "defaults_rename.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, name)
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if d := disagreement(truth, got); d != "" {
					t.Fatalf("%s: %+v", d, got)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
	for _, stop := range []struct{ name, what string }{
		{"nested.ts", "a nested destructuring pattern"},
		{"rest.ts", "a rest element in destructuring"},
		{"set.ts", "a for...of destructuring an object"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowered(t, filepath.Join(directory, stop.name))
			var gap *lower.NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, stop.what) {
				t.Fatalf("%s: got %v, want a NotYet saying %q", stop.name, err, stop.what)
			}
			t.Logf("%s stops: %s", stop.name, gap.What)
		})
	}
}
