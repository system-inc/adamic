package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A for...of whose variable is an object pattern, over a plain array, holds each element once per step and
// destructures it by the declaration road (#rm22wzv): getters.ts reads each getter once per step in the pattern's
// order, and defaults_rename.ts takes a default and a rename, each beside Node on sanitized native, release native
// and the JavaScript backend. A nested pattern and a rest element keep the declaration road's stops, and an object
// pattern over a Set, no plain array, keeps "a for...of destructuring an object". Mutant: destructureFrom reading a
// pattern's bindings in reverse, so each step runs b's getter before a's, reds getters.ts against Node, with C that
// compiles. A mutant of the loop's own code, the prefix dropped or moved or run twice, fails to compile on native.
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
	// The stops are written to the test's own directory, not testdata, since nothing is meant to lift them: each is the
	// declaration road's own stop or the plain array road's edge.
	for _, stop := range []struct{ name, program, what string }{
		{"nested.ts", nestedPattern, "a nested destructuring pattern"},
		{"rest.ts", restPattern, "a rest element in destructuring"},
		{"set.ts", setPattern, "a for...of destructuring an object"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), stop.name)
			if err := os.WriteFile(path, []byte(stop.program), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var gap *lower.NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, stop.what) {
				t.Fatalf("%s: got %v, want a NotYet saying %q", stop.name, err, stop.what)
			}
			t.Logf("%s stops: %s", stop.name, gap.What)
		})
	}
}

// nestedPattern keeps the declaration road's stop for a nested pattern.
const nestedPattern = `// A nested pattern keeps the declaration road's stop (#rm22wzv).
const rows = [{ inner: { value: 1 } }];
for (const { inner: { value } } of rows) {
    console.log(String(value));
}
`

// restPattern keeps the declaration road's stop for a rest element.
const restPattern = `// A rest element keeps the declaration road's stop (#rm22wzv).
const rows = [{ a: 1, b: 2 }];
for (const { a, ...others } of rows) {
    console.log(String(a) + String(others.b));
}
`

// setPattern is an object pattern over a Set, no plain array road, so it keeps its stop by name.
const setPattern = `// An object pattern over a Set is no plain array road, so it keeps its stop by name (#rm22wzv).
const seen = new Set<{ a: number }>([{ a: 1 }]);
for (const { a } of seen) {
    console.log(String(a));
}
`
