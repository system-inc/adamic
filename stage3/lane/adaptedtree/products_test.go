package adaptedtree

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Not parallel: establishes the tests' stage3/lane-relative working directory, where Product's paths resolve as lane's
// tests resolve them.
func TestMain(suite *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	os.Exit(suite.Run())
}

// The adapted tree is settled by this package's own test binary, so its key holds this package's recipe, never lane's,
// which every lane test file's edit would move. It holds no empty directory, which the house's lay would drop
// (#rhj78rk), and git reads its .keep files as no ref: refs/heads and refs/tags list none, as the pin is checked out
// detached, and no ref of apply.py's own refs/stage3 names one. Mutant: keepEmptyDirectories never called, red
// (.git/refs/heads is empty).
func TestProduct_Stage3AdaptedTree(test *testing.T) {
	test.Parallel()
	product := Product(test)
	tree := filepath.Join(product, "tree")
	walkError := filepath.WalkDir(product, func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil || !entry.IsDir() {
			return pathError
		}
		entries, readDirectoryError := os.ReadDir(path)
		if readDirectoryError == nil && len(entries) == 0 {
			return fmt.Errorf("the product holds the empty directory %s, which a lay would drop", path)
		}
		return readDirectoryError
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	if refs, outputError := toolchain.GitCommandIn(
		tree,
		"for-each-ref",
		"refs/heads",
		"refs/tags",
	).Output(); outputError != nil || len(bytes.TrimSpace(refs)) > 0 {
		test.Fatalf(
			"the adapted tree's branches and tags read %q (%v), want none: the pin is checked out detached",
			refs,
			outputError,
		)
	}
	if refs, outputError := toolchain.GitCommandIn(
		tree,
		"for-each-ref",
		"--format=%(refname)",
	).Output(); outputError != nil || bytes.Contains(refs, []byte(".keep")) {
		test.Fatalf("the adapted tree's refs read %q (%v), want none naming a .keep", refs, outputError)
	}
	head, outputError := toolchain.GitCommandIn(tree, "rev-parse", "HEAD").Output()
	if outputError != nil {
		test.Fatal(outputError)
	}
	repository, absolutePathError := filepath.Abs(repositoryFromLane)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if pin := typescriptPin(test, repository); string(bytes.TrimSpace(head)) != pin {
		test.Fatalf("the adapted tree is at %s, not the pinned %s", head, pin)
	}
	// built/local/run.js is the harness hereby tests makes, the last thing the build writes.
	if _, statError := os.Stat(filepath.Join(tree, "built/local/run.js")); statError != nil {
		test.Fatalf("the adapted tree isn't built: %v", statError)
	}
}
