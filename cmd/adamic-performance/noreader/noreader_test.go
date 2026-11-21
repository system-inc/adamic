package noreader

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// performanceFile is the file nothing but its own command may name.
const performanceFile = "Performance.jsonc"

// codeExtensions are the files that build or test something, Go and the line's scripts, each declared by its own Reads
// line above the walk. Docs, data and the file itself are left unread.
var codeExtensions = []string{".go", ".sh", ".py", ".mjs", ".cjs", ".js"}

// No code outside cmd/adamic-performance names Performance.jsonc, so nothing that builds or tests reads it: the line
// writes it after the gate. Only code is read (codeExtensions), so the walk's key never holds the file itself, which
// moves in every landing, nor the tree's data and baselines; .git, node_modules and every submodule .gitmodules lists
// are never entered. Mutant: a _test.go elsewhere reading Performance.jsonc, named here.
func TestNothingOutsideItsCommandReadsPerformanceJsonc(t *testing.T) {
	t.Parallel()
	// Reads: *.go
	// Reads: *.sh
	// Reads: *.py
	// Reads: *.mjs
	// Reads: *.cjs
	// Reads: *.js
	root := repositoryRoot(t)
	submodules := submodulePaths(t, root)
	read := 0
	var readers []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.Name() == ".git" || entry.IsDir() && (entry.Name() == "node_modules" || slices.Contains(submodules, relative)) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || strings.HasPrefix(relative, "cmd/adamic-performance/") || !slices.Contains(codeExtensions, filepath.Ext(relative)) {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		read++
		if bytes.Contains(contents, []byte(performanceFile)) {
			readers = append(readers, relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// A walk that read nothing would pass on any tree, so it fails here instead.
	if read == 0 {
		t.Fatalf("read no files under %s; the walk looks in the wrong place", root)
	}
	for _, reader := range readers {
		t.Errorf("%s names %s, which the line writes after the gate: nothing that builds or tests may read it, and only cmd/adamic-performance writes it", reader, performanceFile)
	}
	t.Logf("read %d code files; skipped .git, node_modules, cmd/adamic-performance and submodules %v", read, submodules)
}

// repositoryRoot is the directory above the test's that holds Adamic's go.mod.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		module, err := os.ReadFile(filepath.Join(directory, "go.mod"))
		if err == nil && strings.Contains(string(module), "module github.com/system-inc/adamic\n") {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("found no directory above the test's holding Adamic's go.mod")
		}
		directory = parent
	}
}

// submodulePaths are the paths .gitmodules lists, which are other repositories and never read.
func submodulePaths(t *testing.T, root string) []string {
	t.Helper()
	modules, err := os.ReadFile(filepath.Join(root, ".gitmodules"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	var paths []string
	for _, line := range strings.Split(string(modules), "\n") {
		key, value, found := strings.Cut(line, "=")
		if found && strings.TrimSpace(key) == "path" {
			paths = append(paths, strings.TrimSpace(value))
		}
	}
	return paths
}
