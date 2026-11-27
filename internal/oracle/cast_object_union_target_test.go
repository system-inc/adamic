package oracle

import (
	"strings"
	"testing"
)

// Cause U of #eg3krrp: one plain object cast to a union target no literal tag tells apart is union target (a)'s
// checked view, the source taken as the union's one member, with no test at the cast. The honest program prints what
// Node prints on the sanitized and release builds and the JavaScript backend, for tsc's objectFlags shape and its
// build info shape. The lie, a Type that is neither member, stops at its first read of a field it lacks, with the
// view's own message, on both backends alike, while Node runs on. The fixtures sit in a directory of their own,
// outside fixtures, so they carry no counts row. Mutant: castRoad without its union target arm, and neither program
// lowers.
func TestObjectUnionTargetCast(t *testing.T) {
	t.Parallel()
	t.Run("honest", func(t *testing.T) {
		t.Parallel()
		program, path := castFixture(t, "cast_object_union_target/honest.ts")
		node := onNode(t, path)
		if node.exitCode != 0 || string(node.stdout) != "4\n8\n1\n2\n" {
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
		program, path := castFixture(t, "cast_object_union_target/lie.ts")
		const stop = "adamic: panic: field read failed: (t as ObjectType | UnionType).objectFlags is not initialized; expected number, found missing\n"
		native, _ := natively(t, program)
		for backend, result := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			if result.exitCode != 70 || string(result.stdout) != "casting\n" || string(result.stderr) != stop || strings.Contains(string(result.stderr), "cast failed") {
				t.Errorf("%s: exit %d, stdout %q, stderr %q; want 70 at the read, in the view's words, never at the cast", backend, result.exitCode, result.stdout, result.stderr)
			}
		}
		if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != "casting\nundefined\n" {
			t.Fatalf("source on Node must run on: exit %d stdout %q", node.exitCode, node.stdout)
		}
	})
}
