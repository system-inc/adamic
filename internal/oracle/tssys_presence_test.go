package oracle

import (
	"path/filepath"
	"testing"
)

// The presence tests tsc reads off ts.sys (#p2n17c4 T1, lower/tssys_presence_test.go), modeled in
// testdata/tssys_presence_twin: a system module whose exported let is made by an arrow called where it's
// written, nesting getNodeSystem's literal, re-exported by a barrel of export *, imported by main as a
// namespace and handed to run, which hands it to inner directly and from an arrow. It's the twin of
// testdata/tssys_presence, tsc's own shape, whose createDirectory reads the literal it's a method of and
// so stops lowering at #dv99xzy, which lower pins and this doesn't run; the twin's createDirectory
// decides without the literal. Each presence test is read off the literal's slot as it runs, a method
// and a field each present and absent, and tsc.ts's two straight off the barrel, and main agrees with
// Node natively, sanitized and on the JavaScript backend, and leaks nothing, as .a and as .ts. Node's own
// output is pinned, so a fixture that stopped testing a member shows here before it shows as an
// agreement. This isn't in the oracle's fixtures list, so no counted row moves with it. Mutants:
// lower/tssys_presence.mutants.md.
func TestTsSysPresenceAgreesWithNode(t *testing.T) {
	t.Parallel()
	const want = "tryEnableSourceMapsForHost\nsetBlocking\n" +
		"direct: writeOutputIsTTY false\ndirect: no getMemoryUsage\ndirect: memory -1, base64encode function, clearScreen undefined\nencoded direct\ncreateDirectory out\ndirect: tsconfig.json {}\n" +
		"arrow: writeOutputIsTTY false\narrow: no getMemoryUsage\narrow: memory -1, base64encode function, clearScreen undefined\nencoded arrow\ncreateDirectory out\narrow: tsconfig.json {}\n" +
		"writeFile out.js emitted\n"
	for _, entry := range []string{"main.a", "main.ts"} {
		t.Run(entry, func(t *testing.T) {
			t.Parallel()
			relative := "internal/oracle/testdata/tssys_presence_twin/" + entry
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); string(node.stdout) != want || node.exitCode != 0 {
				t.Fatalf("Node: exit %d, stdout %q, want exit 0 and %q", node.exitCode, node.stdout, want)
			}
			// generatorAgrees holds any program to Node on every backend and to leaking nothing.
			generatorAgrees(t, relative)
		})
	}
}
