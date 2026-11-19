package lower

import (
	"fmt"
	"strings"
	"testing"
)

// A call that reaches two declarations before they're initialized names the earlier one's read, in its
// own words, the same on every run, though the reach set it's found in is a map (#88389y1).
func TestNamespaceFirstStopIsStable(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, read, words string }{
		{"namespaces", `function f(): number { return B.y + A.x; }
const early = f();
namespace A { export const x = 1; }
namespace B { export const y = 2; }
console.log(String(early));
`, "A.x", "a namespace read before runtime initialization"},
		{"enum_and_namespace", `function g(): number { return N.v + E.a; }
const early = g();
enum E { a = 1 }
namespace N { export const v = 2; }
console.log(String(early));
`, "E.a", "reading an enum before its runtime initialization"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			where := fmt.Sprintf("main.ts:1:%d:", strings.Index(probe.source, probe.read)+1)
			first := ""
			for run := range 50 {
				_, err := lowerSourceAs(t, probe.source, ".ts")
				if err == nil {
					t.Fatal("lowered, want the read before initialization stopped")
				}
				// Each run lowers in its own scratch directory, so the stop is compared from the file name on.
				stop := err.Error()
				if at := strings.Index(stop, "main.ts:"); at >= 0 {
					stop = stop[at:]
				}
				if run == 0 {
					first = stop
					t.Logf("stops at: %s", first)
					if !strings.HasPrefix(first, where) || !strings.Contains(first, probe.words) {
						t.Fatalf("got %q, want %q at the earlier declaration's read, %s", first, probe.words, where)
					}
				} else if stop != first {
					t.Fatalf("run %d stopped at %q, run 0 at %q", run, stop, first)
				}
			}
		})
	}
}
