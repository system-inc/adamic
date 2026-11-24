package oracle

import (
	"testing"
)

// The untagged arm (#axqgdck) composed under keepsNullishCast (#b5w3ycg): a cast from a union of plain objects
// with undefined beside them, to one member with undefined kept. The honest program prints what Node prints on the
// sanitized and release builds and the JavaScript backend. A lie takes the same view, which
// TestUntaggedUnionViewStopsAtTheRead's fixtures witness. The fixture sits in a directory of its own, outside
// fixtures, so it carries no counts row.
// Mutant: untaggedUnionArm declining every cast, and the program doesn't lower.
func TestUntaggedUnionCastKeepsNullish(t *testing.T) {
	t.Parallel()
	t.Run("honest", func(t *testing.T) {
		t.Parallel()
		program, path := castFixture(t, "cast_untagged_nullish/honest.ts")
		node := onNode(t, path)
		if node.exitCode != 0 || string(node.stdout) != "/found 1.0.0\nabsent\nundefined\n" {
			t.Fatalf("Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
		}
		native, _ := natively(t, program)
		for backend, result := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			if difference := disagreement(node, result); difference != "" {
				t.Errorf("%s: %s: exit %d stdout %q stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
			}
		}
	})
}
