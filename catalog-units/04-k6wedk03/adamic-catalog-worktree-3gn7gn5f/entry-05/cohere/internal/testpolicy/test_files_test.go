package testpolicy

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// eachTestFile calls visit with every _test.go file of this module and its source, and returns the module
// root and how many files it read. It fails the test when it read fewer than 100, since a walk that read
// nothing would let every policy pass.
func eachTestFile(t *testing.T, visit func(path string, source []byte) error) (string, int) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	files := 0
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if entry.IsDir() {
			switch entry.Name() {
			case "TypeScript", "testdata", "node_modules", ".cache", ".git", ".build":
				return filepath.SkipDir
			}
			// A directory with a go.mod of its own is another module (the TypeScript shims), which this
			// module's go test ./... never runs, so its tests are not this module's policies to hold.
			if path != root {
				if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files++
		return visit(path, source)
	})
	if err != nil {
		t.Fatal(err)
	}
	if files < 100 {
		t.Fatalf("read only %d test files under %s", files, root)
	}
	return root, files
}
