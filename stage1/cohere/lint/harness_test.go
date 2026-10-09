package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEmittedJavaScriptMismatch(t *testing.T) {
	t.Parallel()
	emittedMismatchUnion(t)
}

// Not parallel: prepares shared emittedMismatchProducts before parallel shards resume.
func TestEmittedJavaScriptMismatch_Setup(t *testing.T) {
	started := time.Now()
	deadline := time.AfterFunc(90*time.Second, func() { panic("P0: emitted JavaScript setup exceeded 90s") })
	defer deadline.Stop()
	emittedMismatchSetup(t)
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

func TestCompleteSuggestionSerialization(t *testing.T) {
	t.Parallel()
	completeSuggestionUnion(t)
}

func TestSuggestionAlongsideAutomaticFix(t *testing.T) {
	t.Parallel()
	suggestionAlongsideUnion(t)
}

func TestWitnessScriptKind(t *testing.T) {
	t.Parallel()
	witnessScriptKindRequireReady(t)
	witnessScriptKindUnion(t)
	t.Logf("TestWitnessScriptKind (union): %s", strings.Join(witnessScriptKindState.Keys, ", "))
}
