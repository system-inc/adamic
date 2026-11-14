package oracle

import (
	"path/filepath"
	"testing"
)

// process.recordreplay, which tsc's sys.ts reads (#b7phwab, Adamic on #b5w3ycg, Nov 7 2026), is a
// member Node's process doesn't have: testdata/process_recordreplay reads it as a boolean and through
// typeof, and main agrees with Node natively, sanitized and on the JavaScript backend, and leaks
// nothing, as .a and as .ts. Node's own output is pinned, so a fixture that stopped reading the member
// shows here before it shows as an agreement. Not in the oracle's fixtures list, so no counted row
// moves with it. Mutants: lower/process_recordreplay.mutants.md.
func TestProcessRecordReplayAgreesWithNode(t *testing.T) {
	t.Parallel()
	const want = "false\nundefined\n"
	for _, entry := range []string{"main.a", "main.ts"} {
		t.Run(entry, func(t *testing.T) {
			t.Parallel()
			relative := "internal/oracle/testdata/process_recordreplay/" + entry
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
