package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/tracked"
)

var repositoryCorpusRoots = []string{
	"bench", "bridge", "cmd", "internal", "stage1", "stage3/drivers/tsc/corpus",
}

// repositoryCorpusCommit pins the repository's half of the printer corpus: every .ts file under repositoryCorpusRoots as
// it was at this commit on main, read from the checkout's objects, never from the working tree. The printers' agreement
// needs a large TypeScript corpus, not a current one, so a .ts edit anywhere in the repository moves no tsprinter
// product; walked live, one edit to lint's main.ts moved all 162 (#wy8gcke, Planner on #rybwgtz). The corpus moves only
// when this pin does. The checkout that builds the corpus product must hold the commit, as json's does.
const repositoryCorpusCommit = "6f7d5d0ba4a7ccafb6ed69fd531d169f9effe443"

// tsPrinterPinnedCorpusInputs keys the pinned corpus by its commit and roots: a different pin, or another root, is
// another product.
func tsPrinterPinnedCorpusInputs(pin string) buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "tsprinter pinned corpus",
		Files: []string{"stage1/cohere/tsprinter/corpus_test.go"},
		Flags: []string{"repository TypeScript at " + pin, "roots " + strings.Join(repositoryCorpusRoots, " ")},
	}
}

// inRepositoryCorpus says whether a path of the pinned tree is a corpus file: a .ts file under one of the roots.
func inRepositoryCorpus(name string) bool {
	if !strings.HasSuffix(name, ".ts") {
		return false
	}
	for _, root := range repositoryCorpusRoots {
		if strings.HasPrefix(name, root+"/") {
			return true
		}
	}
	return false
}

// corpusListName is the product's list of its files, one repository path a line, in the pinned tree's order.
const corpusListName = "corpus.list"

// buildTSPrinterPinnedCorpus writes each corpus file at the pin under the product, at its repository path, and the list
// of them. Every byte comes through buildcache.PinnedFiles as key-time work, so the build's read set holds no file of
// the working tree.
func buildTSPrinterPinnedCorpus(inputs buildcache.Inputs, pin string) func(directory string) error {
	return func(directory string) error {
		root, err := filepath.Abs(repository)
		if err != nil {
			return err
		}
		files, err := buildcache.PinnedFiles(root, inputs, pin, "TypeScript", inRepositoryCorpus)
		if err != nil {
			return err
		}
		var list strings.Builder
		for _, file := range files {
			path := filepath.Join(directory, filepath.FromSlash(file.Path))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(path, file.Data, 0o644); err != nil {
				return err
			}
			list.WriteString(file.Path + "\n")
		}
		return os.WriteFile(filepath.Join(directory, corpusListName), []byte(list.String()), 0o644)
	}
}

func TestProduct_TSPrinterPinnedCorpus(t *testing.T) {
	t.Parallel()
	inputs := tsPrinterPinnedCorpusInputs(repositoryCorpusCommit)
	buildcache.Product(t, inputs, buildTSPrinterPinnedCorpus(inputs, repositoryCorpusCommit))
}

// pinnedCorpus is the pinned corpus product, fetched once per process: its directory, which case labels name as repo/
// as they name the checkout (tsPrinterLabelRoots, read by parallel tests that never fetched it, so atomic), and its
// files there, in order.
var pinnedCorpus struct {
	once      sync.Once
	directory atomic.Pointer[string]
	files     []string
	err       error
}

func pinnedCorpusFiles(t *testing.T) []string {
	t.Helper()
	pinnedCorpus.once.Do(func() {
		inputs := tsPrinterPinnedCorpusInputs(repositoryCorpusCommit)
		directory, err := buildcache.Get(inputs, buildTSPrinterPinnedCorpus(inputs, repositoryCorpusCommit))
		if err != nil {
			pinnedCorpus.err = fmt.Errorf("tsprinter pinned corpus: %w", err)
			return
		}
		list, err := os.ReadFile(filepath.Join(directory, corpusListName))
		if err != nil {
			pinnedCorpus.err = fmt.Errorf("tsprinter pinned corpus: %w", err)
			return
		}
		for _, name := range strings.Split(strings.TrimSuffix(string(list), "\n"), "\n") {
			pinnedCorpus.files = append(pinnedCorpus.files, filepath.Join(directory, filepath.FromSlash(name)))
		}
		pinnedCorpus.directory.Store(&directory)
	})
	if pinnedCorpus.err != nil {
		t.Fatal(pinnedCorpus.err)
	}
	return append([]string(nil), pinnedCorpus.files...)
}

// tsPrinterLabelRoots are the directories a case label may name a file under, each with the name a stored product
// spells it by: the checkout and the pinned corpus (the repository at its pin) as repo/, the TypeScript source as
// typescript/. Read back (canonical false), repo/ is the checkout, which comes first.
func tsPrinterLabelRoots(root, external string, canonical bool) [][2]string {
	roots := [][2]string{{root, "repo"}}
	if directory := pinnedCorpus.directory.Load(); canonical && directory != nil {
		roots = append(roots, [2]string{*directory, "repo"})
	}
	return append(roots, [2]string{external, "typescript"})
}

// A different pin is a different corpus product, by name key and by declared key: a corpus at another commit is never
// served for this one. Keying builds nothing.
func TestTSPrinterPinnedCorpusRekeysWithThePin(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	keys := func(pin string) (string, string) {
		inputs := tsPrinterPinnedCorpusInputs(pin)
		declared, err := buildcache.Key(root, inputs)
		if err != nil {
			t.Fatal(err)
		}
		return buildcache.NameKey(root, inputs), declared
	}
	nameKey, declared := keys(repositoryCorpusCommit)
	// The pin with its last hex digit changed: the smallest move to another commit id.
	other := repositoryCorpusCommit[:39] + "1"
	if other == repositoryCorpusCommit {
		other = repositoryCorpusCommit[:39] + "2"
	}
	otherNameKey, otherDeclared := keys(other)
	if otherNameKey == nameKey || otherDeclared == declared {
		t.Fatalf("pin %s kept name key %s or key %s", other, nameKey[:12], declared[:12])
	}
}

func trackedRootFiles(t *testing.T, checkout, root string) ([]string, error) {
	t.Helper()
	info, err := os.Stat(filepath.Join(checkout, root))
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("corpus root %s: missing directory (%v)", root, err)
	}
	// git ls-files where the checkout has .git, else the manifest a Loom runner's unpacked source carries. The listing
	// is key-time work (buildcache.Keying): every product that reads the corpus keys each listed .ts file itself
	// (statementCorpusBuild by an input= flag with its sha256, tsPrinterOracleProduct as Files, tscAgreementPrepare by
	// its declared corpus directory), so a .ts added or changed moves those keys. Walked under the trace, the listing
	// put every directory under bench, bridge, cmd, internal and stage1 into their read sets, and any file added there,
	// a css gap's or a Go file's, moved tsprinter's keys and every product downstream of them (#0rx19ha).
	var paths []string
	buildcache.Keying(func() { paths, err = tracked.PathsUnder(checkout, root) })
	if err != nil {
		return nil, fmt.Errorf("corpus root %s: tracked files: %w", root, err)
	}
	var files []string
	for _, name := range paths {
		if strings.HasSuffix(name, ".ts") {
			files = append(files, filepath.Join(checkout, name))
		}
	}
	// Without .git, a .ts file the manifest leaves out (and no .gitignore ignores) is a stale manifest, refused: a
	// check, never a byte a build reads, so it runs as key-time work too.
	var extra []string
	buildcache.Keying(func() { extra, err = tracked.Unrecorded(checkout, root) })
	if err != nil {
		return nil, fmt.Errorf("corpus root %s: %w", root, err)
	}
	for _, name := range extra {
		if strings.HasSuffix(name, ".ts") {
			return nil, fmt.Errorf("corpus root %s: %s is on disk but not in the tree's manifest", root, name)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("corpus root %s: no git-tracked .ts files", root)
	}
	t.Logf("corpus root %s: %d git-tracked .ts files", root, len(files))
	return files, nil
}

func printerCorpusFiles(t *testing.T) []string {
	t.Helper()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to a TypeScript 6.0.3 source checkout at 050880ce; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	// The pin is a check, never an input: tracked.Head asks git as key-time work, with the home's configuration out of
	// reach (home-6's dry run, Oct 11: this git read ~/.gitconfig inside the build, refusing it and the Go oracle beside it).
	if head, err := tracked.Head(source); err != nil || head != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		t.Fatalf("TypeScript pin: %q %v", head, err)
	}
	// The repository's half is the pinned corpus product's files, never the working tree's (repositoryCorpusCommit).
	files := pinnedCorpusFiles(t)
	listed, err := trackedRootFiles(t, source, "src/compiler")
	if err != nil {
		t.Fatal(err)
	}
	return append(files, listed...)
}

func TestTrackedCorpusRoots(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if result := execute(t, nil, "git", "-C", directory, "init", "-q"); result.exitCode != 0 {
		t.Fatal(string(result.stderr))
	}
	if err := os.Mkdir(filepath.Join(directory, "named"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"named/tracked.ts", "named/untracked.ts", "outside.ts"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("x;\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if result := execute(t, nil, "git", "-C", directory, "add", "named/tracked.ts", "outside.ts"); result.exitCode != 0 {
		t.Fatal(string(result.stderr))
	}
	files, err := trackedRootFiles(t, directory, "named")
	if err != nil || len(files) != 1 || files[0] != filepath.Join(directory, "named/tracked.ts") {
		t.Fatalf("named tracked corpus: %v %v", files, err)
	}
	if _, err := trackedRootFiles(t, directory, "missing"); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing root was not named: %v", err)
	}
	if err := os.Mkdir(filepath.Join(directory, "empty"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := trackedRootFiles(t, directory, "empty"); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty root was not named: %v", err)
	}
	// The same checkout as a Loom runner unpacks it, every file but .git, answers from the tree's manifest alike.
	if result := execute(t, nil, "git", "-C", directory, "-c", "user.name=Corpus fixture", "-c", "user.email=corpus@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "fixture"); result.exitCode != 0 {
		t.Fatal(string(result.stderr))
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
	if _, err := trackedRootFiles(t, source, "named"); err == nil || !strings.Contains(err.Error(), tracked.ManifestDirectory) {
		t.Fatalf("a source with neither .git nor a manifest was listed: %v", err)
	}
	if err := tracked.Write(directory, source); err != nil {
		t.Fatal(err)
	}
	if _, err := trackedRootFiles(t, source, "named"); err == nil || !strings.Contains(err.Error(), "named/untracked.ts") {
		t.Fatalf("a .ts file the manifest leaves out was dropped silently: %v", err)
	}
	if err := os.Remove(filepath.Join(source, "named/untracked.ts")); err != nil {
		t.Fatal(err)
	}
	files, err = trackedRootFiles(t, source, "named")
	if err != nil || len(files) != 1 || files[0] != filepath.Join(source, "named/tracked.ts") {
		t.Fatalf("named tracked corpus without .git: %v %v", files, err)
	}
}
