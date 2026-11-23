package oracle

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A .ts cast that only removes undefined or null is the presence check a non-null assertion makes (presentCast,
// #b5w3ycg). The honest program, tsc's `args[0] as Expression` with the value there, prints what Node prints on the
// sanitized and release builds and the JavaScript backend. The lie, undefined at the cast, stops there on each of
// them alike: exit 70, the panic `value!` makes (non_null_checked_test.go), in every cast's words, while
// Node runs on. The fixtures sit in a directory of their own, outside fixtures, so they carry no counts row.
// Mutants: the drops-nullish road removed (the cast stops as before), and neither program lowers, so both rows go
// red. The check skipped, its Coalesce replaced by the value it tests, held below: the lie runs on natively.
func TestDropsNullishCastIsAPresenceCheck(t *testing.T) {
	t.Parallel()
	t.Run("honest", func(t *testing.T) {
		t.Parallel()
		program, path := castFixture(t, "cast_drops_nullish/honest.ts")
		node := onNode(t, path)
		if node.exitCode != 0 || string(node.stdout) != "identifier 1\nidentifier\n" {
			t.Fatalf("Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
		}
		native, _ := natively(t, program)
		for backend, result := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			if difference := disagreement(node, result); difference != "" {
				t.Errorf("%s: %s: exit %d stdout %q stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
			}
		}
	})
	t.Run("lie", func(t *testing.T) {
		t.Parallel()
		program, path := castFixture(t, "cast_drops_nullish/lie.ts")
		if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != "casting\nundefined\n" {
			t.Fatalf("source on Node must run on: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
		}
		want := run{stdout: []byte("casting\n"), stderr: []byte("adamic: panic: cast failed: this Expression | undefined is not a Expression\n"), exitCode: 70}
		native, _ := natively(t, program)
		for backend, result := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			if difference := disagreement(want, result); difference != "" {
				t.Errorf("%s: %s: exit %d stdout %q stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
			}
		}
		changes := 0
		skip := func(node any) any {
			if value, ok := node.(ir.Coalesce); ok && value.Panic != nil {
				changes++
				return value.Value
			}
			return node
		}
		program.Main = mutateReadiness(program.Main, skip)
		for index := range program.Functions {
			program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, skip)
		}
		if changes != 1 {
			t.Fatalf("the no-check mutant changed %d checks, want the cast's one", changes)
		}
		if mutant, _ := nativelyUncached(t, program); mutant.exitCode == 70 || disagreement(want, mutant) == "" {
			t.Fatalf("the no-check mutant escaped: exit %d stdout %q stderr %q", mutant.exitCode, mutant.stdout, mutant.stderr)
		}
	})
}
