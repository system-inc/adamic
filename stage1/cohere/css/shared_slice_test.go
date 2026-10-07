package css

import (
	"path/filepath"
	"testing"
)

// A shared slice has no writable capacity. Appending must copy before writing into its owner's bytes.
func TestSharedSliceAppendAgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/shared_slice_append.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	nativeRun, binary := natively(t, program)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"Node", onNode(t, path)}, {"native", nativeRun}, {"JavaScript backend", onJavaScriptBackend(t, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "1152\n" {
			t.Fatalf("%s: exit %d, stdout %q, stderr %q", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
