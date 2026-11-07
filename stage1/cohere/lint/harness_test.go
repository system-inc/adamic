package lint

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEmittedJavaScriptMismatch(test *testing.T) {
	test.Parallel()
	emittedMismatchUnion(test)
}

// Independently selectable preparation check; shards use the same once themselves.
func TestEmittedJavaScriptMismatch_Setup(test *testing.T) {
	test.Parallel()
	started := time.Now()
	emittedMismatchSetup(test)
	test.Logf("TestEmittedJavaScriptMismatch (setup): %.3fs", time.Since(started).Seconds())
}

func serializationPort(test *testing.T) string {
	directory := mutant(test, "", "")
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
		if err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), data, 0644); err != nil {
			test.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
		test.Fatal(err)
	}
	return directory
}

func TestCompleteSuggestionSerialization(test *testing.T) {
	test.Parallel()
	completeSuggestionUnion(test)
}

func TestSuggestionAlongsideAutomaticFix(test *testing.T) {
	test.Parallel()
	suggestionAlongsideUnion(test)
}
