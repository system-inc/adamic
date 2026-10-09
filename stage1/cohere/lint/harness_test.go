package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Prepare products once before the parallel comparison leaf.
// Not parallel: prepares the shared emitted-mismatch products before parallel leaves.
// Not parallel: prepares shared emitted-mismatch products before the parallel comparison leaf.
func TestEmittedJavaScriptMismatch(t *testing.T) {
	started := time.Now()
	deadline := time.AfterFunc(90*time.Second, func() { panic("P0: emitted JavaScript setup exceeded 90s") })
	defer deadline.Stop()
	emittedMismatchSetup(t)
	emittedMismatchUnion(t)
	elapsed := time.Since(started)
	t.Logf("TestEmittedJavaScriptMismatch (setup): %.3fs cooked=%t", elapsed.Seconds(), elapsed >= 90*time.Second)
	if elapsed >= 60*time.Second {
		t.Fatal("setup exceeds 60s budget")
	}
}

func serializationPort(t *testing.T) string {
	directory := mutant(t, "", "")
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
		t.Fatal(err)
	}
	return directory
}

// Not parallel: prepares shared serialization products before parallel leaves.
// Not parallel: initializes shared suggestion-serialization fixtures and products.
func TestCompleteSuggestionSerialization(t *testing.T) {
	completeSuggestionUnion(t)
	completeSuggestionSetup(t)
}

// Not parallel: prepares shared suggestion products before parallel leaves.
// Not parallel: initializes shared suggestion and automatic-fix fixtures and products.
func TestSuggestionAlongsideAutomaticFix(t *testing.T) {
	suggestionAlongsideSetup(t)
	suggestionAlongsideUnion(t)
}

// Not parallel: prepares shared witness-script-kind products before parallel leaves.
// Not parallel: initializes shared witness-script-kind state before its parallel shards.
func TestWitnessScriptKind(t *testing.T) {
	witnessScriptKindSetup(t)
	witnessScriptKindUnion(t)
	t.Logf("TestWitnessScriptKind (setup): %s", strings.Join(witnessScriptKindState.keys, ", "))
}
