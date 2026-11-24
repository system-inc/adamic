package oracle

import (
	"strings"
	"testing"
)

// Union target (a) (#b5w3ycg): a cast from a union of plain objects no literal tag tells apart to a union of its
// members is a checked view of the target union, with no test at the cast, read through the fields every target
// member declares in one representation (view()'s union contract). The honest program prints what Node prints on the
// sanitized and release builds and the JavaScript backend. The lie, a value of the third member, stops at its first
// read of a field it lacks, with the view's own message, on both backends alike, while Node runs on. The fixtures sit
// in a directory of their own, outside fixtures, so they carry no counts row. Mutant: untaggedUnionArm's gate held to
// one target again, and neither program lowers.
func TestUntaggedUnionTargetCast(t *testing.T) {
	t.Parallel()
	t.Run("honest", func(t *testing.T) {
		t.Parallel()
		program, path := castFixture(t, "cast_untagged_union_target/honest.ts")
		node := onNode(t, path)
		if node.exitCode != 0 || string(node.stdout) != "from a\nfrom b\n4\n8\n" {
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
		program, path := castFixture(t, "cast_untagged_union_target/lie.ts")
		const stop = "adamic: panic: field read failed: (x as A | B).shared is not initialized; expected string, found missing\n"
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
