package main

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/registry"
)

// The flag exists so a caller can tell a binary that lacks a rule from one whose rule found
// nothing. Those produce identical finding lists, and only this can separate them.
func TestRulesFlagListsEveryRegisteredRule(t *testing.T) {
	t.Parallel()

	binary := buildCohere(t)

	output, err := exec.Command(binary, "-rules").Output()
	if err != nil {
		t.Fatalf("running -rules: %v", err)
	}
	printed := strings.Fields(string(output))

	expected := make([]string, 0, len(registry.All()))
	for _, registered := range registry.All() {
		expected = append(expected, registered.Name)
	}
	sort.Strings(expected)

	if len(printed) != len(expected) {
		t.Fatalf("printed %d rules, registry has %d", len(printed), len(expected))
	}
	for index := range expected {
		if printed[index] != expected[index] {
			t.Fatalf("at %d: printed %q, registry has %q", index, printed[index], expected[index])
		}
	}
	if len(printed) == 0 {
		t.Fatal("printed nothing, which would pass every comparison vacuously")
	}
}

// TestRulesFlagDisclosesADevelopmentBuildBesideTheList pins the qualification to the number.
//
// A development build implements whatever was on disk when it was compiled, which in a shared
// worktree includes other authors' uncommitted work. On one night three people read a count off
// this surface and reported it as the repository's; `-version` had carried the disclosure the whole
// time and nobody ran it. A disclosure one command to the side is a disclosure that gets skipped,
// so it has to travel with the number a reader carries away.
//
// The split matters as much as the note. The list stays on stdout so two binaries can still be
// diffed directly, and the note goes to stderr so it cannot corrupt that diff.
func TestRulesFlagDisclosesADevelopmentBuildBesideTheList(t *testing.T) {
	t.Parallel()

	binary := buildCohere(t)

	command := exec.Command(binary, "-rules")
	var standardError strings.Builder
	command.Stderr = &standardError
	standardOutput, err := command.Output()
	if err != nil {
		t.Fatalf("running -rules: %v", err)
	}

	// A test binary is built from the working tree, so it is always a development build. If this
	// ever stops holding, the assertion below is measuring nothing and should fail rather than pass
	// quietly.
	note := standardError.String()
	if note == "" {
		t.Fatal("a locally built binary printed no provenance note, so the disclosure is not reaching the surface people quote")
	}
	if !strings.Contains(note, "local build") {
		t.Fatalf("the note does not say the build is local: %q", note)
	}
	// Built from the working tree, the binary is either dirty (the disclaimer) or a clean named
	// commit (which reproduces it). Either way the note must say which, so one of the two must hold.
	if !strings.Contains(note, "not necessarily what is committed") && !strings.Contains(note, "reproduces them") {
		t.Fatalf("the note says neither that the count may differ from the repository nor which commit reproduces it: %q", note)
	}

	// The count has to appear in the note itself. A note that qualifies "the rules" in the abstract
	// is the same disclosure-to-the-side this exists to remove; the reader carries away a number.
	if !strings.Contains(note, strconv.Itoa(len(strings.Fields(string(standardOutput))))) {
		t.Fatalf("the note does not carry the rule count it qualifies: %q", note)
	}

	// stdout must remain exactly the list, or the diff-two-binaries use stops working.
	for _, line := range strings.Split(strings.TrimSpace(string(standardOutput)), "\n") {
		if strings.Contains(line, "note:") || strings.Contains(line, "local build") {
			t.Fatalf("provenance leaked onto stdout and would corrupt a diff: %q", line)
		}
	}
}
