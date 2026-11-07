package buildcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// pinnedRepository is a repository with two commits, each holding a.json, and returns it and both ids.
func pinnedRepository(test *testing.T) (string, string, string) {
	test.Helper()
	root := test.TempDir()
	gitIn(test, root, "init", "-q")
	commit := func(content string) string {
		test.Helper()
		if writeError := os.WriteFile(filepath.Join(root, "a.json"), []byte(content), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		gitIn(test, root, "add", "a.json")
		gitIn(test, root, "commit", "-q", "-m", content)
		output, outputError := toolchain.GitCommandIn(root, "rev-parse", "HEAD").Output()
		if outputError != nil {
			test.Fatal(outputError)
		}
		return strings.TrimSpace(string(output))
	}
	return root, commit(`{"pinned":true}`), commit(`{"pinned":false}`)
}

// A pinned commit's files are read as key-time work, so a trace never counts .git as a build's read (#jwn3n9q), and
// what git answers is the commit's own: a replace ref standing another commit in under its id is ignored. A commit
// that isn't a full id, or that the key's flags don't name, is refused, so the key moves
// with the pin. Mutants: run outside keying; replace objects read; the flag check dropped; a short id taken.
// Not parallel: it swaps the package-level pinnedKeying hook.
func TestPinnedGitReadsTheCommitsOwnBytesAsKeyTimeWork(test *testing.T) {
	root, pinned, later := pinnedRepository(test)
	inputs := Inputs{Name: "pinned corpus", Flags: []string{"repository JSON at " + pinned}}
	// Another commit replaces the pin, as a test's replace ref could.
	gitIn(test, root, "replace", pinned, later)
	var keyed int
	previous := pinnedKeying
	pinnedKeying = func(compute func()) {
		keyed++
		previous(compute)
	}
	test.Cleanup(func() { pinnedKeying = previous })
	output, pinnedGitError := PinnedGit(root, inputs, pinned, nil, "show", pinned+":a.json")
	if pinnedGitError != nil || string(output) != `{"pinned":true}` || keyed != 1 {
		test.Fatalf("the pin's a.json read %q (%v), as keying %d times", output, pinnedGitError, keyed)
	}
	output, pinnedGitError = PinnedGit(root, inputs, pinned, []byte(pinned+":a.json\n"), "cat-file", "--batch")
	if pinnedGitError != nil || !strings.HasSuffix(strings.TrimSpace(string(output)), `{"pinned":true}`) {
		test.Fatalf("the pin's a.json through stdin read %q: %v", output, pinnedGitError)
	}
	for name, c := range map[string]struct {
		inputs Inputs
		commit string
		want   string
	}{
		"a short id":             {inputs, pinned[:12], "isn't a full commit id"},
		"a commit no flag names": {inputs, later, "no flag of its key names commit"},
	} {
		if _, pinnedGitError := PinnedGit(
			root,
			c.inputs,
			c.commit,
			nil,
			"show",
			c.commit+":a.json",
		); pinnedGitError == nil || !strings.Contains(pinnedGitError.Error(), c.want) {
			test.Errorf("%s: %v", name, pinnedGitError)
		}
	}
}

// The caller's trace2 event target reaches PinnedGit's git, so a test can count the processes git ran (json's
// never-fetches tests). Mutant: the variable not passed, and the trace never written.
// Not parallel: it sets GIT_TRACE2_EVENT for the process.
func TestPinnedGitWritesTheCallersTraceTwoEvents(test *testing.T) {
	root, pinned, _ := pinnedRepository(test)
	events := filepath.Join(test.TempDir(), "events.json")
	test.Setenv("GIT_TRACE2_EVENT", events)
	if _, pinnedGitError := PinnedGit(
		root,
		Inputs{Name: "pinned corpus", Flags: []string{"at " + pinned}},
		pinned,
		nil,
		"cat-file",
		"-e",
		pinned+"^{commit}",
	); pinnedGitError != nil {
		test.Fatal(pinnedGitError)
	}
	if content, readError := os.ReadFile(events); readError != nil || !strings.Contains(string(content), `"cat-file"`) {
		test.Fatalf("the trace2 events read %q: %v", content, readError)
	}
}
