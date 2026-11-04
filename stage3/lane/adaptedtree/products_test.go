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
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// The adapted tree is settled by this package's own test binary, so its key holds this package's recipe, never lane's,
// which every lane test file's edit would move. It holds no empty directory, which the house's lay would drop
// (#rhj78rk), and git reads its .keep files as no ref. Mutant: keepEmptyDirectories never called, red (.git/refs/heads
// is empty).
func TestProduct_Stage3AdaptedTree(t *testing.T) {
	t.Parallel()
	product := Product(t)
	tree := filepath.Join(product, "tree")
	err := filepath.WalkDir(product, func(path string, entry os.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return err
		}
		entries, err := os.ReadDir(path)
		if err == nil && len(entries) == 0 {
			return fmt.Errorf("the product holds the empty directory %s, which a lay would drop", path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if refs, err := toolchain.GitCommandIn(tree, "for-each-ref").Output(); err != nil || len(bytes.TrimSpace(refs)) > 0 {
		t.Fatalf("the adapted tree's refs read %q (%v), want none: the pin is checked out detached", refs, err)
	}
	head, err := toolchain.GitCommandIn(tree, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	repository, err := filepath.Abs(repositoryFromLane)
	if err != nil {
		t.Fatal(err)
	}
	if pin := typescriptPin(t, repository); string(bytes.TrimSpace(head)) != pin {
		t.Fatalf("the adapted tree is at %s, not the pinned %s", head, pin)
	}
	// built/local/run.js is the harness hereby tests makes, the last thing the build writes.
	if _, err := os.Stat(filepath.Join(tree, "built/local/run.js")); err != nil {
		t.Fatalf("the adapted tree isn't built: %v", err)
	}
}
