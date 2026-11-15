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
// Not parallel: it swaps the package-level pinnedKeying hook.
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

// The caller's trace2 event target reaches PinnedGit's git, so a test can count the processes git ran (json's
// never-fetches tests). Mutant: the variable not passed, and the trace never written.
// Not parallel: it sets GIT_TRACE2_EVENT for the process.
func TestPinnedGitWritesTheCallersTraceTwoEvents(t *testing.T) {
	root, pinned, _ := pinnedRepository(t)
	events := filepath.Join(t.TempDir(), "events.json")
	t.Setenv("GIT_TRACE2_EVENT", events)
	if _, err := PinnedGit(root, Inputs{Name: "pinned corpus", Flags: []string{"at " + pinned}}, pinned, nil, "cat-file", "-e", pinned+"^{commit}"); err != nil {
		t.Fatal(err)
	}
	if content, err := os.ReadFile(events); err != nil || !strings.Contains(string(content), `"cat-file"`) {
		t.Fatalf("the trace2 events read %q: %v", content, err)
	}
}

// A checkout another user owns, as a sealed job's dynamic uid reads the image's root-owned TypeScript checkout, is read
// through PinnedGit's one safe.directory, which names exactly the root it was handed: the same checkout asked through a
// directory below its top level is still refused as dubious, where a wildcard would read it (#1ndpe51). git takes every
// checkout as another user's here (GIT_TEST_ASSUME_DIFFERENT_OWNER, as git's own t0033 does), and the control shows it
// refuses this one with no exemption, so the read is the exemption's. Mutants: the exemption dropped; its value `*`.
// Not parallel: it sets the package-level pinnedEnvironment.
func TestPinnedGitExemptsExactlyTheRootItWasHanded(t *testing.T) {
	root, pinned, _ := pinnedRepository(t)
	inputs := Inputs{Name: "pinned corpus", Flags: []string{"repository JSON at " + pinned}}
	below := filepath.Join(root, "below")
	if err := os.Mkdir(below, 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := "GIT_TEST_ASSUME_DIFFERENT_OWNER=1"
	control := toolchain.GitCommandIn(root, "cat-file", "-e", pinned+"^{commit}")
	// The control runs as PinnedGit's git does, with no system or global configuration, so a box's own safe.directory
	// can't read the checkout for it.
	control.Env = append(control.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", foreign)
	if output, err := control.CombinedOutput(); err == nil || !strings.Contains(string(output), "dubious ownership") {
		t.Fatalf("git with no exemption read the checkout taken as another user's (%v), so this test proves nothing: %s", err, output)
	}
	previous := pinnedEnvironment
	pinnedEnvironment = []string{foreign}
	t.Cleanup(func() { pinnedEnvironment = previous })
	files, err := PinnedFiles(root, inputs, pinned, "JSON", func(path string) bool { return strings.HasSuffix(path, ".json") })
	if err != nil || len(files) != 1 || files[0].Path != "a.json" || string(files[0].Data) != `{"pinned":true}` {
		t.Fatalf("the exempted root, another user's checkout, read %q: %v", files, err)
	}
	if _, err := PinnedGit(below, inputs, pinned, nil, "cat-file", "-e", pinned+"^{commit}"); err == nil || !strings.Contains(err.Error(), "dubious ownership") {
		t.Fatalf("a directory below another user's checkout, which the exemption doesn't name, read: %v", err)
	}
}
