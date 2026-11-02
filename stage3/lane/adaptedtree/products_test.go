package adaptedtree

import (
	"bytes"
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
// which every lane test file's edit would move.
func TestProduct_Stage3AdaptedTree(t *testing.T) {
	t.Parallel()
	tree := filepath.Join(Product(t), "tree")
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
