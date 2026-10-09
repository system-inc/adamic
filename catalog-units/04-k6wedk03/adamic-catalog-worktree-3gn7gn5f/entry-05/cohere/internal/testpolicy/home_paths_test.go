package testpolicy

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestHomePathsFindsAPlantedHomePath is the checker against planted literals, each beside its control.
func TestHomePathsFindsAPlantedHomePath(t *testing.T) {
	t.Parallel()
	// Not a machine path: every case is a planted source, read by the checker and never by the filesystem.
	cases := []struct {
		name   string
		source string
		want   int
	}{
		{"a macOS home path", `const root = "/Users/someone/Projects/ahra"`, 1},
		{"a Linux home path", "const root = `/home/someone/checkouts/ahra`", 1},
		{"inside a longer string", `var line = "error at /Users/someone/a.ts:1:1"`, 1},
		{"a repository path", `const root = "/repo/app/layout.tsx"`, 0},
		{"a word that ends in Users", `const root = "/AllUsers/x/y"`, 0},
		{"a home directory alone", `const root = "/Users/"`, 0},
		{"in a comment only", "// ahra pinned /Users/someone/Projects/ahra\nconst root = \"/repo\"", 0},
		{"excused above", "// Not a machine path: the input whose home prefix is trimmed.\nconst root = \"/Users/someone/x\"", 0},
		{"excused above a block", "// Not a machine path: planted cases.\nvar roots = []string{\n\t\"/Users/a/x\",\n\t\"/home/b/y\",\n}", 0},
		{"a block after the excused one", "// Not a machine path: planted cases.\nvar roots = []string{\"/Users/a/x\"}\nvar other = \"/Users/c/z\"", 1},
		{"excused at the line's end", "const root = \"/Users/someone/x\" // Not a machine path: trimmed by the transform.", 0},
		{"a marker with no reason", "// Not a machine path:\nconst root = \"/Users/someone/x\"", 1},
	}
	for _, testCase := range cases {
		found, err := HomePaths("planted_test.go", []byte("package p\n\n"+testCase.source+"\n"))
		if err != nil {
			t.Fatalf("%s: %v", testCase.name, err)
		}
		if len(found) != testCase.want {
			t.Errorf("%s: found %d, expected %d: %v", testCase.name, len(found), testCase.want, found)
		}
	}
}

// TestNoCohereTestNamesAHomeDirectory holds the module to reading outside code by name (#sycrdr6).
func TestNoCohereTestNamesAHomeDirectory(t *testing.T) {
	t.Parallel()
	found := []HomePath{}
	root, _ := eachTestFile(t, func(path string, source []byte) error {
		inFile, err := HomePaths(path, source)
		found = append(found, inFile...)
		return err
	})
	lines := map[string][]string{}
	for _, homePath := range found {
		relative, _ := filepath.Rel(root, homePath.Position.Filename)
		relative = filepath.ToSlash(relative)
		lines[relative] = append(lines[relative], fmt.Sprintf("  %s:%d %q", relative, homePath.Position.Line, homePath.Value))
	}
	files := make([]string, 0, len(lines))
	for file := range lines {
		files = append(files, file)
	}
	sort.Strings(files)
	for _, file := range files {
		t.Errorf("%s names a home directory. Ask for the code by name through internal/corpus, or, for a literal "+
			"that only looks like one, say why above it with %q:\n%s", file, NotAMachinePathMarker, strings.Join(lines[file], "\n"))
	}
}
