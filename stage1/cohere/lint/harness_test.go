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
		data, readError := os.ReadFile(filepath.Join("testdata/serialization", name))
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "rules/no-debugger", name),
			data,
			0644,
		); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if removeError := os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); removeError != nil {
		test.Fatal(removeError)
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
