package buildcache

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// quotedInclude is a preprocessor line including a file by a quoted name, the form a C compiler looks for first beside
// the file that includes it: #include "name", with spaces allowed around the #.
var quotedInclude = regexp.MustCompile(`^\s*#\s*include\s*"([^"]+)"`)

// includingExtensions are the tracked files the guard reads for quoted includes: C, C++ and header files, every line,
// and Go files, their cgo preambles alone.
var includingExtensions = []string{".c", ".h", ".hh", ".hpp", ".hxx", ".cc", ".cpp", ".cxx", ".go"}

// No tracked file includes, by a quoted name, a file whose extension a package listing leaves out (packageSourceExtensions,
// Compiler's condition on fix 1b): such a file beside a package's sources could be added or removed and move no key,
// though the include it answers changes what compiles. A C, C++ or header file is read whole, every line, comments
// included, which can only find more. A .go file is read in its cgo preamble alone, the comment right before its import
// "C", as cmd/cgo takes it: Go files hold C sources for other programs in strings (internal/native's probes), which no
// package's cgo compiles. The files are the ones Tracked names, each checked-out submodule's included (cohere's), so a
// runner's source, which has no .git, lists them from its manifest. Mutants, each caught: ".h" dropped from
// packageSourceExtensions, which internal/buildcache/testdata/withc/main.go's #include "greeting.h" then names; a
// tracked .c file that includes "x.inc".
func TestNoTrackedIncludeNamesAFileAPackageListingLeavesOut(t *testing.T) {
	t.Parallel()
	root := repositoryRootForTest(t)
	walked := 0
	err := filepath.WalkDir(root, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if file != root && (entry.Name() == ".git" || entry.Name() == "node_modules" || !Tracked(root, file, true)) {
				return filepath.SkipDir
			}
			return nil
		}
		// A link is read where it points, as the file the tree carries there.
		if !entry.Type().IsRegular() || !slices.Contains(includingExtensions, filepath.Ext(file)) || !Tracked(root, file, false) {
			return nil
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		walked++
		name := filepath.ToSlash(mustRelative(root, file))
		lines := strings.Split(string(content), "\n")
		numbers := make([]int, len(lines))
		for index := range lines {
			numbers[index] = index + 1
		}
		if filepath.Ext(file) == ".go" {
			lines, numbers = cgoPreambles(name, content)
		}
		for index, line := range lines {
			match := quotedInclude.FindStringSubmatch(line)
			if match == nil {
				continue
			}
			if extension := path.Ext(match[1]); !slices.Contains(packageSourceExtensions, extension) {
				t.Errorf("%s:%d includes %q, whose extension %q no package listing names (packageSourceExtensions), so adding or removing it moves no key", name, numbers[index], match[1], extension)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if walked == 0 {
		t.Fatalf("the tree at %s carries no C, header or Go file, so the guard read nothing", root)
	}
	t.Logf("%d C, header and Go files", walked)
}

// cgoPreambles is each line of a Go file's cgo preambles, its comment markers taken off, with the line it's on: the
// comment right before an import of "C", or before the import declaration holding it alone, as cmd/cgo reads it. A file
// that doesn't parse whole (a testdata corpus's) still gives what parsed, and a preamble is read before anything that
// can fail after the imports. A file that never names "C" has none, and isn't parsed.
func cgoPreambles(name string, content []byte) ([]string, []int) {
	if !bytes.Contains(content, []byte(`"C"`)) {
		return nil, nil
	}
	files := token.NewFileSet()
	// The error is a syntax error at most, since content is already read, and the partial file still holds every import
	// declaration ahead of it: cmd/cgo reads nothing else of a file for its preamble.
	parsed, _ := parser.ParseFile(files, name, content, parser.ImportsOnly|parser.ParseComments)
	if parsed == nil {
		return nil, nil
	}
	var lines []string
	var numbers []int
	for _, declaration := range parsed.Decls {
		imports, ok := declaration.(*ast.GenDecl)
		if !ok || imports.Tok != token.IMPORT {
			continue
		}
		for _, spec := range imports.Specs {
			imported, ok := spec.(*ast.ImportSpec)
			if !ok || imported.Path.Value != `"C"` {
				continue
			}
			preamble := imported.Doc
			if preamble == nil && len(imports.Specs) == 1 {
				preamble = imports.Doc
			}
			if preamble == nil {
				continue
			}
			for _, comment := range preamble.List {
				first := files.Position(comment.Slash).Line
				text, line := strings.CutPrefix(comment.Text, "//")
				if !line {
					text = strings.TrimSuffix(strings.TrimPrefix(comment.Text, "/*"), "*/")
				}
				for offset, one := range strings.Split(text, "\n") {
					lines, numbers = append(lines, one), append(numbers, first+offset)
				}
			}
		}
	}
	return lines, numbers
}
