package corpusfiles

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fixture(t *testing.T, root, extension string) (string, string, string) {
	t.Helper()
	directory := t.TempDir()
	runGit(t, directory, "init", "-q")
	runGit(t, directory, "config", "user.email", "corpus@example.invalid")
	runGit(t, directory, "config", "user.name", "Corpus fixture")
	name := filepath.Join(root, "tracked file"+extension)
	writeFile(t, filepath.Join(directory, name), "original\n")
	writeFile(t, filepath.Join(directory, ".gitignore"), "ignored-*\n")
	runGit(t, directory, "add", ".")
	runGit(t, directory, "commit", "-qm", "Initial fixture")
	pin := strings.TrimSpace(runGit(t, directory, "rev-parse", "HEAD"))
	return directory, pin, name
}

func runGit(t *testing.T, directory string, args ...string) string {
	t.Helper()
	data, err := git(directory, args...)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeFile(t *testing.T, name, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}

func requireFailure(t *testing.T, directory, pin, root, pattern string, fragments ...string) {
	t.Helper()
	_, _, _, err := selectFiles(directory, pin, []string{root}, []string{pattern})
	if err == nil {
		t.Fatal("planted corpus fault survived")
	}
	for _, fragment := range fragments {
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("failure %q does not name %q", err, fragment)
		}
	}
	t.Logf("caught: %v", err)
}

// Every converted package exercises the same selection and named fault contract.
// Only the initial fixture is committed; every planted fault stays in scratch.
func TestConvertedPackageContracts(t *testing.T) {
	for _, test := range []struct{ name, root, extension string }{
		{"scanner", "stage1", ".ts"}, {"scanner/profile-runtime", "internal/native/runtime", ".c"}, {"typeaware", "src/compiler", ".ts"},
		{"markdowninline", ".", ".md"}, {"yaml", ".github/workflows", ".yml"},
		{"cssstrings", "internal/format/css/testdata/prettier", ".css"},
		{"cssnumbers", "internal/format/css/testdata/prettier", ".scss"},
		{"graphql/printer", "internal/format/graphql", "_test.go"},
		{"markdownblocks/default", "stage1/cohere/markdownblocks", ".md"},
		{"markdownblocks/census", ".", ".md"},
		{"selector", "internal/format/css", ".css"},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory, pin, tracked := fixture(t, test.root, test.extension)
			pattern := "*" + test.extension
			roots, patterns := []string{test.root}, []string{pattern}
			before := Repository(t, directory, roots, patterns)
			if len(before) != 1 {
				t.Fatalf("fixture count %d", len(before))
			}
			if got := Upstream(t, directory, pin, roots, patterns); !reflect.DeepEqual(before, got) {
				t.Fatal("clean upstream differs")
			}
			stray := filepath.Join(test.root, "stray"+test.extension)
			writeFile(t, filepath.Join(directory, stray), "stray\n")
			if after := Repository(t, directory, roots, patterns); !reflect.DeepEqual(before, after) {
				t.Fatal("untracked file changed count")
			}
			requireFailure(t, directory, pin, test.root, pattern, "untracked", filepath.ToSlash(stray))
			if err := os.Remove(filepath.Join(directory, stray)); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(directory, tracked), "dirty\n")
			requireFailure(t, directory, "", test.root, pattern, "worktree", filepath.ToSlash(tracked))
			requireFailure(t, directory, pin, test.root, pattern, filepath.ToSlash(tracked))
			runGit(t, directory, "add", "--", tracked)
			// Index-only fault: the worktree still equals the index, so only
			// the HEAD/index check can catch this edit.
			requireFailure(t, directory, "", test.root, pattern, "index", filepath.ToSlash(tracked))
			writeFile(t, filepath.Join(directory, tracked), "original\n")
			requireFailure(t, directory, "", test.root, pattern, "index", filepath.ToSlash(tracked))
			runGit(t, directory, "reset", "-q", "HEAD", "--", tracked)
			requireFailure(t, directory, strings.Repeat("0", 40), test.root, pattern, "HEAD", "pin")
			ignored := filepath.Join(test.root, "ignored-stray"+test.extension)
			writeFile(t, filepath.Join(directory, ignored), "ignored\n")
			requireFailure(t, directory, pin, test.root, pattern, "ignored", filepath.ToSlash(ignored))
		})
	}
}

func TestMissingAndEmptyRoots(t *testing.T) {
	directory, pin, _ := fixture(t, "inputs", ".ts")
	requireFailure(t, directory, "", "missing", "*.ts", "missing named root", "missing")
	writeFile(t, filepath.Join(directory, "empty", "other.txt"), "other\n")
	requireFailure(t, directory, "", "empty", "*.ts", "empty", "no tracked files")
	requireFailure(t, directory, pin, "inputs", "*.md", "inputs", "no tracked files")
}

func TestSparseCheckoutCannotShrinkCorpus(t *testing.T) {
	directory, _, tracked := fixture(t, "tests/format/css", ".css")
	// A second file remains available, so omitting the missing-file guard would
	// silently shrink a nonempty corpus rather than hit the empty-root guard.
	writeFile(t, filepath.Join(directory, "tests/format/css/retained.css"), "retained\n")
	runGit(t, directory, "add", ".")
	runGit(t, directory, "commit", "-qm", "Prepare two-file sparse fixture")
	pin := strings.TrimSpace(runGit(t, directory, "rev-parse", "HEAD"))
	// A skipped index entry is deliberately absent. Git status considers this
	// clean, and ls-files still returns it, as in the sparse Prettier checkout.
	runGit(t, directory, "update-index", "--skip-worktree", "--", tracked)
	if err := os.Remove(filepath.Join(directory, tracked)); err != nil {
		t.Fatal(err)
	}
	// Keep its named directory present, so the per-file check is what catches it.
	if err := os.MkdirAll(filepath.Join(directory, "tests/format/css"), 0755); err != nil {
		t.Fatal(err)
	}
	if status := runGit(t, directory, "status", "--porcelain", "--ignored"); status != "" {
		t.Fatalf("sparse control not clean: %s", status)
	}
	requireFailure(t, directory, pin, "tests/format/css", "*.css", "sparse-checkout", filepath.ToSlash(tracked))
}

func TestDeterministicRootsPatternsAndMissingGit(t *testing.T) {
	directory, _, _ := fixture(t, "inputs", ".ts")
	writeFile(t, filepath.Join(directory, "inputs", "z.TS"), "z\n")
	writeFile(t, filepath.Join(directory, "inputs", "a.ts"), "a\n")
	runGit(t, directory, "add", ".")
	runGit(t, directory, "commit", "-qm", "More fixture inputs")
	a := Repository(t, directory, []string{"inputs", "inputs"}, []string{"*.ts", "*.TS"})
	b := Repository(t, directory, []string{"inputs"}, []string{"*.ts"})
	if !reflect.DeepEqual(a, b) || len(a) != 3 || filepath.Base(a[0]) != "a.ts" {
		t.Fatalf("selection: %v / %v", a, b)
	}
	t.Setenv("PATH", t.TempDir())
	requireFailure(t, directory, "", "inputs", "*.ts", "git", "executable file not found")
}

func TestProvisionedPrettierFixtures(t *testing.T) {
	checkout := os.Getenv("ADAMIC_CSS_FIXTURES")
	if checkout == "" {
		t.Skip("set ADAMIC_CSS_FIXTURES to the pinned Prettier fixture checkout (#xq2ecw6)")
	}
	// Sparse checkout excludes other Prettier suites, not individual files
	// inside these declared suites. ls-files still includes skipped entries.
	files := Upstream(t, checkout, PrettierCommit,
		[]string{"tests/format/css", "tests/format/scss", "tests/format/less"},
		[]string{"*.css", "*.scss", "*.less"})
	counts := map[string]int{}
	for _, file := range files {
		counts[filepath.Ext(file)]++
	}
	t.Logf("Prettier fixture counts from Git: %v", counts)
}
