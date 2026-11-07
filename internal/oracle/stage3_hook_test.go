package oracle

import (
	"encoding/json"
	"os"
	"testing"
)

// TestStage3FixtureHook exposes the existing helpers through the oracle test
// binary. It is dormant in ordinary gates and never caches a fixture observation.
func TestStage3FixtureHook(test *testing.T) {
	test.Parallel()
	path := os.Getenv("ADAMIC_STAGE3_FIXTURE")
	if path == "" {
		test.Skip("called only by the Stage 3 fixture runner")
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	result, binary := nativelyUncached(test, program)
	if result.exitCode == 0 {
		if report := leaksUncached(test, program, binary); report != "" {
			test.Fatalf("leaks: %s", report)
		}
	}
	// Byte slices are base64 in JSON, preserving even non-UTF-8 output.
	data, encodeError := json.Marshal([]any{result.stdout, result.stderr, result.exitCode})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(os.Getenv("ADAMIC_STAGE3_RESULT"), data, 0600); writeError != nil {
		test.Fatal(writeError)
	}
}
