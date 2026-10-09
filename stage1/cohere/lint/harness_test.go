package lint

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEmittedJavaScriptMismatch(t *testing.T) {
	t.Parallel()
	emittedMismatchUnion(t)
}

// Independently selectable preparation check; shards use the same once themselves.
func TestEmittedJavaScriptMismatch_Setup(t *testing.T) {
	t.Parallel()
	started := time.Now()
	emittedMismatchSetup(t)
	t.Logf("TestEmittedJavaScriptMismatch (setup): %.3fs", time.Since(started).Seconds())
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

func TestCompleteSuggestionSerialization(t *testing.T) {
	t.Parallel()
	completeSuggestionUnion(t)
}

func TestSuggestionAlongsideAutomaticFix(t *testing.T) {
	t.Parallel()
	suggestionAlongsideUnion(t)
}
