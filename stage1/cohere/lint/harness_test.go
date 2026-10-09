package lint

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Prepare products once before the parallel comparison leaf.
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

func TestDotARename(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{ownedWitnesses(t, directory, "no-var")[0] + "\tno-var"})
	oracle := goOracle(t)
	want := compare(t, oracle, buildPort(t, directory, true), directory, path)
	copied := mutant(t, "", "")
	entry := filepath.Join(copied, "rules/no-var/rule.a")
	before, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	renamed := filepath.Join(copied, "rules/no-var/rule.ts")
	if err := os.Rename(entry, renamed); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(renamed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("rename changed module bytes")
	}
	got := compare(t, oracle, buildPort(t, copied, true), copied, path)
	if !bytes.Equal(got, want) {
		t.Fatal("rename changed results")
	}
	t.Logf("rename only: .ts and .a identical on all three runtimes against Go (%d bytes)", len(want))
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
	completeSuggestionUnion(t)
	completeSuggestionSetup(t)
}

func TestSuggestionAlongsideAutomaticFix(t *testing.T) {
	suggestionAlongsideSetup(t)
	suggestionAlongsideUnion(t)
}

func TestWitnessScriptKind(t *testing.T) {
	witnessScriptKindSetup(t)
	witnessScriptKindUnion(t)
	t.Logf("TestWitnessScriptKind (setup): %s", strings.Join(witnessScriptKindState.keys, ", "))
}
