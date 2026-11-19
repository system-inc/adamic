//go:build !census_overlay

package latent_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The ordinary gate discovers this test without special Go flags. Build the
// same scratch overlay used by the measurement tooling and run its two pins.
func TestLatentCensusOverlayGateHcxsh1k(t *testing.T) {
	t.Parallel()
	// Reads: stage3/census/latent/importclean/*.go
	// Reads: stage3/census/latent/refusalrewrite/*.go
	// Reads: stage3/census/latent/replay/main.go.txt
	// Reads: stage3/census/latent/statecopy/*.go
	// Reads: stage3/census/latent/statementrewrite/*.go
	// Reads: stage3/census/latent/timing/timing.go
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	generate := exec.CommandContext(t.Context(), "python3", "make_overlay.py", root, scratch)
	if output, err := generate.CombinedOutput(); err != nil {
		t.Fatalf("generate census overlay: %v\n%s", err, output)
	}

	tests := []string{
		"TestLatentWrittenAnyDeferredCensusStopsZb0q0zd",
		"TestLatentScopedTailEntryFindingLoom03dg",
	}
	run := exec.CommandContext(t.Context(), "go", "test", "-count=1", "-v", "-tags=census_overlay",
		"-overlay="+filepath.Join(scratch, "overlay.json"), "-run=^("+strings.Join(tests, "|")+")$", ".")
	output, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("census overlay tests: %v\n%s", err, output)
	}
	for _, test := range tests {
		if !strings.Contains(string(output), "--- PASS: "+test+" (") {
			t.Fatalf("missing pass for %s\n%s", test, output)
		}
	}
}
