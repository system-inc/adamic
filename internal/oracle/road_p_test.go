package oracle

import (
	"path/filepath"
	"testing"
)

// Road P (#p2n17c4, rulings.md:117 applied to the right of an &&): testdata/road_p tests a class's
// required method and an optional member, given and missing, on the right of an && tested for truthiness
// alone, in each position lower's TestRoadPRightOfAndLowers admits, with every flag pair, and main agrees
// with Node natively, sanitized and on the JavaScript backend, and leaks nothing, as .a and as .ts. Node's
// own output is pinned, so a fixture that stopped testing a position shows here before it shows as an
// agreement. Not in the oracle's fixtures list, so no counted row moves with it. Mutants:
// lower/road_p.mutants.md.
func TestRoadPAgreesWithNode(t *testing.T) {
	t.Parallel()
	const want = "abcd falsef | ---- true- | abcd falsef\n" +
		"a--- falsef | ---- true- | a--- falsef\n" +
		"---- true- | ---- true- | ---- true-\n" +
		"---- true- | ---- true- | ---- true-\n"
	for _, entry := range []string{"main.a", "main.ts"} {
		t.Run(entry, func(t *testing.T) {
			t.Parallel()
			relative := "internal/oracle/testdata/road_p/" + entry
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); string(node.stdout) != want || node.exitCode != 0 {
				t.Fatalf("Node: exit %d, stdout %q, want exit 0 and %q", node.exitCode, node.stdout, want)
			}
			generatorAgrees(t, relative)
		})
	}
}
