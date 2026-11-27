package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// bindAbsentStop is the words native stops with for a non-optional bind of a method its receiver lacks:
// the words every other road to a missing method uses, which #8vptg16 turns into Node's catchable
// TypeError. The emitter writes them without the file's name, since ir.BoundMethod carries no location.
const bindAbsentStop = "adamic: panic: method weigh is absent\n"

// A spread copies own data only, so { ...new Scale() } seen as a type declaring weigh has no weigh, and
// value.weigh.bind(value) found nothing: natively the bound closure kept a null code pointer and its call
// was an ASan SEGV, where Node throws reading bind off undefined (#1k4z29h). The bind now stops natively
// where it's made, before any fixed argument runs (fixed_argument: Node prints no note), and both forms of
// weigh, a method signature and a function-typed field, stop alike. The JavaScript backend already throws as
// Node does, but it runs fixed_argument's argument before it reads bind (#w8gg81z), so that row pins today's
// output there. A bind of a method or a field the receiver holds agrees with Node on every backend.
// Mutant (bind_absent.mutants.md): the emitter's check dropped, and the stopping rows end on closure.c's
// invariant instead, "a bound method has no code to call", with no SEGV, which shows the backstop holds.
func TestBindOfAbsentMethodStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, node string
		stops      bool
		// javascript is the JavaScript backend's stdout where it differs from Node's today, naming the task.
		javascript string
	}{
		{"spread_method.ts", "threw\n", true, ""},
		{"spread_field.ts", "threw\n", true, ""},
		// The JavaScript backend runs the fixed argument before reading bind, so it prints note first (#w8gg81z).
		{"fixed_argument.ts", "threw\n", true, "note\nthrew\n"},
		{"present_method.ts", "20\n", false, ""},
		{"present_field.ts", "5\n", false, ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_absent", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.node)}, node); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := node
			if probe.javascript != "" {
				want = run{stdout: []byte(probe.javascript)}
			}
			if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "" {
				t.Fatalf("the JavaScript backend: %s", difference)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program)} {
				which := []string{"sanitized native", "release native"}[index]
				if probe.stops {
					if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), bindAbsentStop) || strings.Contains(string(got.stderr), "AddressSanitizer") {
						t.Fatalf("%s: want exit 70 with %q and no stdout, got exit %d, stdout %q, stderr %q", which, bindAbsentStop, got.exitCode, got.stdout, got.stderr)
					}
					continue
				}
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, which, got.exitCode, got.stdout, got.stderr)
				}
			}
			if !probe.stops {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
