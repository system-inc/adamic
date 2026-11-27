package oracle

import (
	"strings"
	"testing"
)

// #gbr6vs2: a .ts cast from an intersection of plain objects that only reads a field narrower is a structural view
// (item 19 on #bbssdcy, tsc's toBuilderStateFileInfoForMultiEmit). Each read through it takes the slot as stored,
// with its runtime kind, and checks it against the target's field, so the source's string | false | undefined slot is
// read as what it holds. The honest program prints what Node prints on the sanitized and release builds and the
// JavaScript backend: a string through the guarded cast, false through the other arm, and an unguarded undefined. The
// target keeps signature readonly; a writable one keeps its NotYet (the view-site rule, Compiler on #gbr6vs2). The lie,
// a false signature cast with no test, stops at its read in the view's words on every backend while Node runs on.
// The fixtures sit outside fixtures, so they carry no counts row. Mutant: castRoad's gate without its intersection
// arm, and neither program lowers.
func TestIntersectionFieldRefinementView(t *testing.T) {
	t.Parallel()
	t.Run("honest", func(t *testing.T) {
		t.Parallel()
		program, path := castFixture(t, "cast_intersection_view/honest.ts")
		node := onNode(t, path)
		if node.exitCode != 0 || string(node.stdout) != "v1 sig-a undefined\nv2 undefined true\nv3 v3 undefined\nv3 undefined\n" {
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
		program, path := castFixture(t, "cast_intersection_view/lie.ts")
		const stop = "adamic: panic: field read failed: direct(b).signature matches no member of string | undefined; expected string | undefined, found boolean\n"
		native, _ := natively(t, program)
		for backend, result := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			if result.exitCode != 70 || string(result.stdout) != "casting\n" || string(result.stderr) != stop || strings.Contains(string(result.stderr), "cast failed") {
				t.Errorf("%s: exit %d, stdout %q, stderr %q; want 70 at the read, in the view's words, never at the cast", backend, result.exitCode, result.stdout, result.stderr)
			}
		}
		if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != "casting\nfalse\n" {
			t.Fatalf("source on Node must run on: exit %d stdout %q", node.exitCode, node.stdout)
		}
	})
}
