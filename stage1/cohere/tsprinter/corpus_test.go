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
func TestTsPrinterPinnedCorpusRekeysWithThePin(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(oracles.Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	keys := func(pin string) (string, string) {
		inputs := oracles.TSPrinterPinnedCorpusInputs(pin)
		declared, keyError := buildcache.Key(root, inputs)
		if keyError != nil {
			test.Fatal(keyError)
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
		test.Fatalf("pin %s kept name key %s or key %s", other, nameKey[:12], declared[:12])
	}
}

func TestTrackedCorpusRoots(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	if result := execute(test, nil, "git", "-C", directory, "init", "-q"); result.ExitCode != 0 {
		test.Fatal(string(result.Stderr))
	}
	if createDirectoryError := os.Mkdir(filepath.Join(directory, "named"), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	for _, name := range []string{"named/tracked.ts", "named/untracked.ts", "outside.ts"} {
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte("x;\n"), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if result := execute(
		test,
		nil,
		"git",
		"-C",
		directory,
		"add",
		"named/tracked.ts",
		"outside.ts",
	); result.ExitCode != 0 {
		test.Fatal(string(result.Stderr))
	}
	files, trackedRootFilesError := oracles.TrackedRootFiles(test, directory, "named")
	if trackedRootFilesError != nil || len(files) != 1 || files[0] != filepath.Join(directory, "named/tracked.ts") {
		test.Fatalf("named tracked corpus: %v %v", files, trackedRootFilesError)
	}
	if _, trackedRootFilesError := oracles.TrackedRootFiles(
		test,
		directory,
		"missing",
	); trackedRootFilesError == nil || !strings.Contains(trackedRootFilesError.Error(), "missing") {
		test.Fatalf("missing root was not named: %v", trackedRootFilesError)
	}
	if createDirectoryError := os.Mkdir(filepath.Join(directory, "empty"), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if _, trackedRootFilesError := oracles.TrackedRootFiles(
		test,
		directory,
		"empty",
	); trackedRootFilesError == nil || !strings.Contains(trackedRootFilesError.Error(), "empty") {
		test.Fatalf("empty root was not named: %v", trackedRootFilesError)
	}
	// The same checkout as a Loom runner unpacks it, every file but .git, answers from the tree's manifest alike.
	if result := execute(
		test,
		nil,
		"git",
		"-C",
		directory,
		"-c",
		"user.name=Corpus fixture",
		"-c",
		"user.email=corpus@example.invalid",
		"-c",
		"commit.gpgsign=false",
		"commit",
		"-qm",
		"fixture",
	); result.ExitCode != 0 {
		test.Fatal(string(result.Stderr))
	}
	source := test.TempDir()
	for _, name := range []string{"named/tracked.ts", "named/untracked.ts", "outside.ts"} {
		if createDirectoryError := os.MkdirAll(
			filepath.Join(source, filepath.Dir(name)),
			0755,
		); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(filepath.Join(source, name), []byte("x;\n"), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if _, trackedRootFilesError := oracles.TrackedRootFiles(
		test,
		source,
		"named",
	); trackedRootFilesError == nil || !strings.Contains(trackedRootFilesError.Error(), tracked.ManifestDirectory) {
		test.Fatalf("a source with neither .git nor a manifest was listed: %v", trackedRootFilesError)
	}
	if writeError := tracked.Write(directory, source); writeError != nil {
		test.Fatal(writeError)
	}
	if _, trackedRootFilesError := oracles.TrackedRootFiles(
		test,
		source,
		"named",
	); trackedRootFilesError == nil || !strings.Contains(trackedRootFilesError.Error(), "named/untracked.ts") {
		test.Fatalf("a .ts file the manifest leaves out was dropped silently: %v", trackedRootFilesError)
	}
	if removeError := os.Remove(filepath.Join(source, "named/untracked.ts")); removeError != nil {
		test.Fatal(removeError)
	}
	files, trackedRootFilesError = oracles.TrackedRootFiles(test, source, "named")
	if trackedRootFilesError != nil || len(files) != 1 || files[0] != filepath.Join(source, "named/tracked.ts") {
		test.Fatalf("named tracked corpus without .git: %v %v", files, trackedRootFilesError)
	}
}
