package oracle

import (
	"encoding/json"
	"os"
	"testing"
)

// TestStage3FixtureHook exposes the existing helpers through the oracle test
// binary. It is dormant in ordinary gates and never caches a fixture observation.
func TestStage3FixtureHook(t *testing.T) {
	t.Parallel()
	path := os.Getenv("ADAMIC_STAGE3_FIXTURE")
	if path == "" {
		t.Skip("called only by the Stage 3 fixture runner")
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	result, binary := nativelyUncached(t, program)
	if result.exitCode == 0 {
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatalf("leaks: %s", report)
		}
	}
	// Byte slices are base64 in JSON, preserving even non-UTF-8 output.
	data, err := json.Marshal([]any{result.stdout, result.stderr, result.exitCode})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("ADAMIC_STAGE3_RESULT"), data, 0600); err != nil {
		t.Fatal(err)
	}
}
