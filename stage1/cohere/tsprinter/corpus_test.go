package tsprinter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/tracked"
)

// A different pin is a different corpus product, by name key and by declared key: a corpus at another commit is never
// served for this one. Keying builds nothing.
func TestTSPrinterPinnedCorpusRekeysWithThePin(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(oracles.Repository)
	if err != nil {
		t.Fatal(err)
	}
	keys := func(pin string) (string, string) {
		inputs := oracles.TSPrinterPinnedCorpusInputs(pin)
		declared, err := buildcache.Key(root, inputs)
		if err != nil {
			t.Fatal(err)
		}
		return buildcache.NameKey(root, inputs), declared
	}
	nameKey, declared := keys(oracles.RepositoryCorpusCommit)
	// The pin with its last hex digit changed: the smallest move to another commit id.
	other := oracles.RepositoryCorpusCommit[:39] + "1"
	if other == oracles.RepositoryCorpusCommit {
		other = oracles.RepositoryCorpusCommit[:39] + "2"
	}
	otherNameKey, otherDeclared := keys(other)
	if otherNameKey == nameKey || otherDeclared == declared {
		t.Fatalf("pin %s kept name key %s or key %s", other, nameKey[:12], declared[:12])
	}
}

func TestTrackedCorpusRoots(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if result := execute(t, nil, "git", "-C", directory, "init", "-q"); result.ExitCode != 0 {
		t.Fatal(string(result.Stderr))
	}
	if err := os.Mkdir(filepath.Join(directory, "named"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"named/tracked.ts", "named/untracked.ts", "outside.ts"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("x;\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if result := execute(t, nil, "git", "-C", directory, "add", "named/tracked.ts", "outside.ts"); result.ExitCode != 0 {
		t.Fatal(string(result.Stderr))
	}
	files, err := oracles.TrackedRootFiles(t, directory, "named")
	if err != nil || len(files) != 1 || files[0] != filepath.Join(directory, "named/tracked.ts") {
		t.Fatalf("named tracked corpus: %v %v", files, err)
	}
	if _, err := oracles.TrackedRootFiles(t, directory, "missing"); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing root was not named: %v", err)
	}
	if err := os.Mkdir(filepath.Join(directory, "empty"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := oracles.TrackedRootFiles(t, directory, "empty"); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty root was not named: %v", err)
	}
	// The same checkout as a Loom runner unpacks it, every file but .git, answers from the tree's manifest alike.
	if result := execute(t, nil, "git", "-C", directory, "-c", "user.name=Corpus fixture", "-c", "user.email=corpus@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "fixture"); result.ExitCode != 0 {
		t.Fatal(string(result.Stderr))
	}
	source := t.TempDir()
	for _, name := range []string{"named/tracked.ts", "named/untracked.ts", "outside.ts"} {
		if err := os.MkdirAll(filepath.Join(source, filepath.Dir(name)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(source, name), []byte("x;\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := oracles.TrackedRootFiles(t, source, "named"); err == nil || !strings.Contains(err.Error(), tracked.ManifestDirectory) {
		t.Fatalf("a source with neither .git nor a manifest was listed: %v", err)
	}
	if err := tracked.Write(directory, source); err != nil {
		t.Fatal(err)
	}
	if _, err := oracles.TrackedRootFiles(t, source, "named"); err == nil || !strings.Contains(err.Error(), "named/untracked.ts") {
		t.Fatalf("a .ts file the manifest leaves out was dropped silently: %v", err)
	}
	if err := os.Remove(filepath.Join(source, "named/untracked.ts")); err != nil {
		t.Fatal(err)
	}
	files, err = oracles.TrackedRootFiles(t, source, "named")
	if err != nil || len(files) != 1 || files[0] != filepath.Join(source, "named/tracked.ts") {
		t.Fatalf("named tracked corpus without .git: %v %v", files, err)
	}
}
