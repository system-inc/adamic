package buildcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// pinnedRepository is a repository with two commits, each holding a.json, and returns it and both ids.
func pinnedRepository(t *testing.T) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	gitIn(t, root, "init", "-q")
	commit := func(content string) string {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, "a.json"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		gitIn(t, root, "add", "a.json")
		gitIn(t, root, "commit", "-q", "-m", content)
		output, err := toolchain.GitCommandIn(root, "rev-parse", "HEAD").Output()
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(output))
	}
	return root, commit(`{"pinned":true}`), commit(`{"pinned":false}`)
}

// A pinned commit's files are read as key-time work, so a trace never counts .git as a build's read (#jwn3n9q), and
// what git answers is the commit's own: a replace ref standing another commit in under its id is ignored. A commit
// that isn't a full id, or that the key's flags don't name, is refused, so the key moves
// with the pin. Mutants: run outside keying; replace objects read; the flag check dropped; a short id taken.
func TestPinnedGitReadsTheCommitsOwnBytesAsKeyTimeWork(t *testing.T) {
	root, pinned, later := pinnedRepository(t)
	inputs := Inputs{Name: "pinned corpus", Flags: []string{"repository JSON at " + pinned}}
	// Another commit replaces the pin, as a test's replace ref could.
	gitIn(t, root, "replace", pinned, later)
	var keyed int
	previous := pinnedKeying
	pinnedKeying = func(compute func()) {
		keyed++
		previous(compute)
	}
	t.Cleanup(func() { pinnedKeying = previous })
	output, err := PinnedGit(root, inputs, pinned, nil, "show", pinned+":a.json")
	if err != nil || string(output) != `{"pinned":true}` || keyed != 1 {
		t.Fatalf("the pin's a.json read %q (%v), as keying %d times", output, err, keyed)
	}
	output, err = PinnedGit(root, inputs, pinned, []byte(pinned+":a.json\n"), "cat-file", "--batch")
	if err != nil || !strings.HasSuffix(strings.TrimSpace(string(output)), `{"pinned":true}`) {
		t.Fatalf("the pin's a.json through stdin read %q: %v", output, err)
	}
	for name, c := range map[string]struct {
		inputs Inputs
		commit string
		want   string
	}{
		"a short id":             {inputs, pinned[:12], "isn't a full commit id"},
		"a commit no flag names": {inputs, later, "no flag of its key names commit"},
	} {
		if _, err := PinnedGit(root, c.inputs, c.commit, nil, "show", c.commit+":a.json"); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
