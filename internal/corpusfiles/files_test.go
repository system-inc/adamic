package corpusfiles

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/tracked"
)

func fixture(test *testing.T, root, extension string) (string, string, string) {
	test.Helper()
	directory := test.TempDir()
	runGit(test, directory, "init", "-q")
	runGit(test, directory, "config", "user.email", "corpus@example.invalid")
	runGit(test, directory, "config", "user.name", "Corpus fixture")
	name := filepath.Join(root, "tracked file"+extension)
	writeFile(test, filepath.Join(directory, name), "original\n")
	writeFile(test, filepath.Join(directory, ".gitignore"), "ignored-*\n")
	runGit(test, directory, "add", ".")
	runGit(test, directory, "commit", "-qm", "Initial fixture")
	pin := strings.TrimSpace(runGit(test, directory, "rev-parse", "HEAD"))
	return directory, pin, name
}

func runGit(test *testing.T, directory string, args ...string) string {
	test.Helper()
	data, err := git(directory, args...)
	if err != nil {
		test.Fatal(err)
	}
	return string(data)
}

func writeFile(test *testing.T, name, text string) {
	test.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(text), 0644); err != nil {
		test.Fatal(err)
	}
}

func requireFailure(test *testing.T, directory, pin, root, pattern string, fragments ...string) {
	test.Helper()
	_, _, _, err := selectFiles(directory, pin, []string{root}, []string{pattern}, nil)
	if err == nil {
		test.Fatal("planted corpus fault survived")
	}
	for _, fragment := range fragments {
		if !strings.Contains(err.Error(), fragment) {
			test.Fatalf("failure %q does not name %q", err, fragment)
		}
	}
	test.Logf("caught: %v", err)
}

// convertedPackages are the converted packages' named roots and the extension each selects.
var convertedPackages = []struct{ name, root, extension string }{
	{"scanner", "stage1", ".ts"},
	{"scanner/profile-runtime", "internal/native/runtime", ".c"},
	{"typeaware", "src/compiler", ".ts"},
	{"markdowninline", ".", ".md"},
	{"yaml", ".github/workflows", ".yml"},
	{"cssstrings", "internal/format/css/testdata/prettier", ".css"},
	{"cssnumbers", "internal/format/css/testdata/prettier", ".scss"},
	{"graphql/printer", "internal/format/graphql", "_test.go"},
	{"markdownblocks/default", "stage1/cohere/markdownblocks", ".md"},
	{"markdownblocks/census", ".", ".md"},
	{"selector", "internal/format/css", ".css"},
}

// Every converted package exercises the same selection and named fault contract.
// Only the initial fixture is committed; every planted fault stays in scratch.
func TestConvertedPackageContracts(test *testing.T) {
	test.Parallel()
	for _, testCase := range convertedPackages {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			directory, pin, tracked := fixture(test, testCase.root, testCase.extension)
			pattern := "*" + testCase.extension
			roots, patterns := []string{testCase.root}, []string{pattern}
			before := Repository(test, directory, roots, patterns)
			if len(before) != 1 {
				test.Fatalf("fixture count %d", len(before))
			}
			if got := Upstream(test, directory, pin, roots, patterns); !reflect.DeepEqual(before, got) {
				test.Fatal("clean upstream differs")
			}
			stray := filepath.Join(testCase.root, "stray"+testCase.extension)
			writeFile(test, filepath.Join(directory, stray), "stray\n")
			if after := Repository(test, directory, roots, patterns); !reflect.DeepEqual(before, after) {
				test.Fatal("untracked file changed count")
			}
			requireFailure(test, directory, pin, testCase.root, pattern, "untracked", filepath.ToSlash(stray))
			if err := os.Remove(filepath.Join(directory, stray)); err != nil {
				test.Fatal(err)
			}
			writeFile(test, filepath.Join(directory, tracked), "dirty\n")
			requireFailure(test, directory, "", testCase.root, pattern, "worktree", filepath.ToSlash(tracked))
			requireFailure(test, directory, pin, testCase.root, pattern, filepath.ToSlash(tracked))
			runGit(test, directory, "add", "--", tracked)
			// Index-only fault: the worktree still equals the index, so only
			// the HEAD/index check can catch this edit.
			requireFailure(test, directory, "", testCase.root, pattern, "index", filepath.ToSlash(tracked))
			writeFile(test, filepath.Join(directory, tracked), "original\n")
			requireFailure(test, directory, "", testCase.root, pattern, "index", filepath.ToSlash(tracked))
			runGit(test, directory, "reset", "-q", "HEAD", "--", tracked)
			requireFailure(test, directory, strings.Repeat("0", 40), testCase.root, pattern, "HEAD", "pin")
			ignored := filepath.Join(testCase.root, "ignored-stray"+testCase.extension)
			writeFile(test, filepath.Join(directory, ignored), "ignored\n")
			requireFailure(test, directory, pin, testCase.root, pattern, "ignored", filepath.ToSlash(ignored))
		})
	}
}

func TestMissingAndEmptyRoots(test *testing.T) {
	test.Parallel()
	directory, pin, _ := fixture(test, "inputs", ".ts")
	requireFailure(test, directory, "", "missing", "*.ts", "missing named root", "missing")
	writeFile(test, filepath.Join(directory, "empty", "other.txt"), "other\n")
	requireFailure(test, directory, "", "empty", "*.ts", "empty", "no tracked files")
	requireFailure(test, directory, pin, "inputs", "*.md", "inputs", "no tracked files")
}

func TestSparseCheckoutCannotShrinkCorpus(test *testing.T) {
	test.Parallel()
	directory, _, tracked := fixture(test, "tests/format/css", ".css")
	// A second file remains available, so omitting the missing-file guard would
	// silently shrink a nonempty corpus rather than hit the empty-root guard.
	writeFile(test, filepath.Join(directory, "tests/format/css/retained.css"), "retained\n")
	runGit(test, directory, "add", ".")
	runGit(test, directory, "commit", "-qm", "Prepare two-file sparse fixture")
	pin := strings.TrimSpace(runGit(test, directory, "rev-parse", "HEAD"))
	// A skipped index entry is deliberately absent. Git status considers this
	// clean, and ls-files still returns it, as in the sparse Prettier checkout.
	runGit(test, directory, "update-index", "--skip-worktree", "--", tracked)
	if err := os.Remove(filepath.Join(directory, tracked)); err != nil {
		test.Fatal(err)
	}
	// Keep its named directory present, so the per-file check is what catches it.
	if err := os.MkdirAll(filepath.Join(directory, "tests/format/css"), 0755); err != nil {
		test.Fatal(err)
	}
	if status := runGit(test, directory, "status", "--porcelain", "--ignored"); status != "" {
		test.Fatalf("sparse control not clean: %s", status)
	}
	requireFailure(test, directory, pin, "tests/format/css", "*.css", "sparse-checkout", filepath.ToSlash(tracked))
}

// Not parallel: test.Setenv changes the process-wide PATH environment variable.
func TestDeterministicRootsPatternsAndMissingGit(test *testing.T) {
	directory, _, _ := fixture(test, "inputs", ".ts")
	writeFile(test, filepath.Join(directory, "inputs", "z.TS"), "z\n")
	writeFile(test, filepath.Join(directory, "inputs", "a.ts"), "a\n")
	runGit(test, directory, "add", ".")
	runGit(test, directory, "commit", "-qm", "More fixture inputs")
	a := Repository(test, directory, []string{"inputs", "inputs"}, []string{"*.ts", "*.TS"})
	b := Repository(test, directory, []string{"inputs"}, []string{"*.ts"})
	if !reflect.DeepEqual(a, b) || len(a) != 3 || filepath.Base(a[0]) != "a.ts" {
		test.Fatalf("selection: %v / %v", a, b)
	}
	test.Setenv("PATH", test.TempDir())
	requireFailure(test, directory, "", "inputs", "*.ts", "git", "executable file not found")
}

func TestProvisionedPrettierFixtures(test *testing.T) {
	test.Parallel()
	checkout := os.Getenv("ADAMIC_CSS_FIXTURES")
	if checkout == "" {
		test.Skip("set ADAMIC_CSS_FIXTURES to the pinned Prettier fixture checkout (#xq2ecw6)")
	}
	// Sparse checkout excludes other Prettier suites, not individual files
	// inside these declared suites. ls-files still includes skipped entries.
	files := Upstream(test, checkout, PrettierCommit,
		[]string{"tests/format/css", "tests/format/scss", "tests/format/less"},
		[]string{"*.css", "*.scss", "*.less"})
	counts := map[string]int{}
	for _, file := range files {
		counts[filepath.Ext(file)]++
	}
	test.Logf("Prettier fixture counts from Git: %v", counts)
}

// unpacked is the fixture as a Loom runner holds a source: its tracked files, with their executable bits, no .git,
// and the manifest build-tree writes (internal/tracked), unless manifest is false.
func unpacked(test *testing.T, directory string, manifest bool) string {
	test.Helper()
	source := test.TempDir()
	for _, name := range names([]byte(runGit(test, directory, "ls-files", "-z"))) {
		info, err := os.Lstat(filepath.Join(directory, name))
		if err != nil {
			test.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			test.Fatal(err)
		}
		writeFile(test, filepath.Join(source, name), string(data))
		if err := os.Chmod(filepath.Join(source, name), info.Mode().Perm()); err != nil {
			test.Fatal(err)
		}
	}
	if manifest {
		if err := tracked.Write(directory, source); err != nil {
			test.Fatal(err)
		}
	}
	return source
}

// The same contract in a source with no .git: the selection is the checkout's, and every fault git would name
// against HEAD is named against the manifest. An index-only fault has no analog: the manifest is HEAD and index.
func TestManifestPackageContracts(test *testing.T) {
	test.Parallel()
	for _, testCase := range convertedPackages {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			checkout, pin, tracked := fixture(test, testCase.root, testCase.extension)
			directory := unpacked(test, checkout, true)
			pattern := "*" + testCase.extension
			roots, patterns := []string{testCase.root}, []string{pattern}
			before := Repository(test, directory, roots, patterns)
			if want := []string{filepath.Join(directory, tracked)}; !reflect.DeepEqual(before, want) {
				test.Fatalf("manifest selection %q, the checkout's is %q", before, want)
			}
			if got := Upstream(test, directory, pin, roots, patterns); !reflect.DeepEqual(before, got) {
				test.Fatal("clean upstream differs")
			}
			// Unlike a checkout, a source can't tell an untracked file from one a stale manifest left out, so a
			// matching file it doesn't record is refused, unless the tree's .gitignore ignores it.
			stray := filepath.Join(testCase.root, "stray"+testCase.extension)
			writeFile(test, filepath.Join(directory, stray), "stray\n")
			requireFailure(test, directory, "", testCase.root, pattern, "doesn't record", filepath.ToSlash(stray))
			requireFailure(test, directory, pin, testCase.root, pattern, "untracked", filepath.ToSlash(stray))
			if err := os.Remove(filepath.Join(directory, stray)); err != nil {
				test.Fatal(err)
			}
			writeFile(test, filepath.Join(directory, tracked), "dirty\n")
			requireFailure(test, directory, "", testCase.root, pattern, "worktree", filepath.ToSlash(tracked))
			requireFailure(test, directory, pin, testCase.root, pattern, "dirty", filepath.ToSlash(tracked))
			writeFile(test, filepath.Join(directory, tracked), "original\n")
			if err := os.Chmod(filepath.Join(directory, tracked), 0755); err != nil {
				test.Fatal(err)
			}
			requireFailure(test, directory, "", testCase.root, pattern, "worktree", filepath.ToSlash(tracked))
			if err := os.Chmod(filepath.Join(directory, tracked), 0644); err != nil {
				test.Fatal(err)
			}
			requireFailure(test, directory, strings.Repeat("0", 40), testCase.root, pattern, "HEAD", "pin")
			ignored := filepath.Join(testCase.root, "ignored-stray"+testCase.extension)
			writeFile(test, filepath.Join(directory, ignored), "ignored\n")
			if after := Repository(test, directory, roots, patterns); !reflect.DeepEqual(before, after) {
				test.Fatal("an ignored file changed the selection")
			}
			requireFailure(test, directory, pin, testCase.root, pattern, "ignored", filepath.ToSlash(ignored))
			if err := os.Remove(filepath.Join(directory, ignored)); err != nil {
				test.Fatal(err)
			}
			if err := os.Remove(filepath.Join(directory, tracked)); err != nil {
				test.Fatal(err)
			}
			requireFailure(
				test,
				directory,
				"",
				testCase.root,
				pattern,
				"missing tracked corpus files",
				filepath.ToSlash(tracked),
			)
			requireFailure(
				test,
				directory,
				pin,
				testCase.root,
				pattern,
				"missing tracked corpus files",
				filepath.ToSlash(tracked),
			)
		})
	}
}

// Upstream status covers every file under a root, not only the selected ones: a deleted tracked file of another kind
// is a dirty root, as git status reports it, while Repository, which checks only selected files, still selects.
func TestManifestUpstreamCoversEveryFileUnderARoot(test *testing.T) {
	test.Parallel()
	checkout, pin, _ := fixture(test, "inputs", ".ts")
	writeFile(test, filepath.Join(checkout, "inputs", "notes.txt"), "notes\n")
	runGit(test, checkout, "add", ".")
	runGit(test, checkout, "commit", "-qm", "Notes beside the corpus")
	pin = strings.TrimSpace(runGit(test, checkout, "rev-parse", "HEAD"))
	directory := unpacked(test, checkout, true)
	Upstream(test, directory, pin, []string{"inputs"}, []string{"*.ts"})
	writeFile(test, filepath.Join(directory, "inputs", "notes.txt"), "edited\n")
	requireFailure(test, directory, pin, "inputs", "*.ts", "dirty", "inputs/notes.txt")
	if err := os.Remove(filepath.Join(directory, "inputs", "notes.txt")); err != nil {
		test.Fatal(err)
	}
	requireFailure(test, directory, pin, "inputs", "*.ts", "dirty", "inputs/notes.txt")
	if files := Repository(test, directory, []string{"inputs"}, []string{"*.ts"}); len(files) != 1 {
		test.Fatalf("repository selection %q", files)
	}
	requireFailure(test, directory, "", "missing", "*.ts", "missing named root", "missing")
	writeFile(test, filepath.Join(directory, "empty", "other.txt"), "other\n")
	requireFailure(test, directory, "", "empty", "*.ts", "empty", "no tracked files")
}

// With neither .git nor a manifest, selection fails by name: the commit and the tracked files are unknown.
func TestNoGitAndNoManifestFails(test *testing.T) {
	test.Parallel()
	checkout, pin, _ := fixture(test, "inputs", ".ts")
	directory := unpacked(test, checkout, false)
	requireFailure(
		test,
		directory,
		pin,
		"inputs",
		"*.ts",
		"no .git",
		tracked.ManifestDirectory,
		"commit and tracked files",
	)
	requireFailure(test, directory, "", "inputs", "*.ts", "no .git", tracked.ManifestDirectory)
}
