package native_runtime

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/tracked"
)

// ownFile is this test's own file, which spells every old form it looks for.
const ownFile = "internal/native_runtime/source_paths_test.go"

// records are the tracked directories whose files are a record of a past run, kept as they were written, so a path
// one names is the path as it was then: any path with one of these as a directory, or under one of the prefixes.
var (
	recordDirectories = []string{"review", "performance", "evidence"}
	recordPrefixes    = []string{"stage3/census/latent/data/", "stage3/area-stack/", "verify/catalog/"}
)

// oldSourcePaths are the forms a Go file reached the runtime's C by before it moved here from internal/native/runtime:
// the repository path and package-relative ones from beside it (internal/native/runtime/..., ../native/runtime), and
// its path components (filepath.Join(root, "internal", "native", "runtime")).
var oldSourcePaths = []*regexp.Regexp{
	regexp.MustCompile(`native/runtime(?:[^A-Za-z0-9_]|$)`),
	regexp.MustCompile(`"native",\s*"runtime"`),
}

// oldPackagePaths are the forms package native read the C by from its own directory, which every read of would now
// miss: its embed, and a read, walk or absolute path of runtime/ beside it. filepath.Join("runtime", ...) isn't one,
// since it also names a build directory's link to the archive cache.
var oldPackagePaths = regexp.MustCompile(`//go:embed runtime/|(?:os\.ReadFile|os\.WriteFile|os\.ReadDir|filepath\.Abs|filepath\.WalkDir)\("runtime[/"]`)

// trackedGoFiles are the repository's tracked Go files outside records, repository-relative, with the repository's root.
func trackedGoFiles(t *testing.T) (string, []string) {
	t.Helper()
	// Reads: *.go
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := tracked.Files(root)
	if err != nil {
		t.Fatal(err)
	}
	var files []string
	for _, entry := range entries {
		if entry.Submodule() || !strings.HasSuffix(entry.Path, ".go") || record(entry.Path) {
			continue
		}
		files = append(files, entry.Path)
	}
	if !slices.Contains(files, ownFile) {
		t.Fatalf("the tracked files don't hold %s, so they aren't this repository's", ownFile)
	}
	return root, files
}

func record(name string) bool {
	for _, prefix := range recordPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	for _, directory := range strings.Split(filepath.ToSlash(filepath.Dir(name)), "/") {
		if slices.Contains(recordDirectories, directory) {
			return true
		}
	}
	return false
}

// No Go file reaches the runtime's C where it was, internal/native/runtime: every reader moved with it to
// internal/native_runtime/c, and one left behind reads a directory that's gone, or a test's mutant edits nothing. Mutant:
// internal/oracle/date_iso_string_test.go's read put back to internal/native/runtime/date.c, and this names its line.
func TestNoGoFileReadsTheRuntimeWhereItWas(t *testing.T) {
	t.Parallel()
	root, files := trackedGoFiles(t)
	var found []string
	for _, name := range files {
		if name == ownFile {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		inNative := filepath.ToSlash(filepath.Dir(name)) == "internal/native"
		for number, line := range strings.Split(string(contents), "\n") {
			old := inNative && oldPackagePaths.MatchString(line)
			for _, form := range oldSourcePaths {
				old = old || form.MatchString(line)
			}
			if old {
				found = append(found, fmt.Sprintf("%s:%d: %s", name, number+1, strings.TrimSpace(line)))
			}
		}
	}
	if len(found) > 0 {
		t.Fatalf("these read the runtime's C at internal/native/runtime, which moved to internal/native_runtime/c:\n%s", strings.Join(found, "\n"))
	}
}

// wholeNativeWithoutTheRuntime are the lists naming internal/native whole that need no internal/native_runtime beside
// it, by file and the function or package variable holding the list, each with its reason: none compiles or reads the
// runtime's C.
var wholeNativeWithoutTheRuntime = map[string]string{
	"cmd/adamic-gate/parallelcensus/names_test.go nameCensusPinned":                         "the packages a test's name collides across, not a key's files",
	"internal/oracle/program_region_parse_test.go programParseLower":                        "native.C's text and the JavaScript backend's, written as files; nothing is compiled, so no runtime C is read",
	"stage1/typescript/parser/compiler_expressions_shards_test.go compilerExpressionsLower": "native.C's text, written as main.c; compilerExpressionsNative compiles it, under a key naming internal/native_runtime",
}

// A key that names internal/native whole names internal/native_runtime beside it: until the runtime's C moved out,
// "internal/native" in a product's Files covered it, and a key that still names only internal/native would stop
// changing when the C does, and hand back a product built from the old runtime. It's held on every []string literal of
// the tree's Go that holds "internal/native", the Files of a buildcache.Inputs or a list handed to one, since a key's
// list is as often a variable as the field. A list that compiles nothing and reads no C is named in
// wholeNativeWithoutTheRuntime with why. Mutant: internal/native/case_test.go's case-mapping-probe put back to
// Files: []string{"internal/native"}, and this names its line and caseMappingHarness.
func TestEveryKeyNamingNativeWholeNamesTheRuntime(t *testing.T) {
	t.Parallel()
	root, files := trackedGoFiles(t)
	var found []string
	allowed := map[string]bool{}
	for _, name := range files {
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, filepath.Join(root, filepath.FromSlash(name)), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			holder := declarationName(declaration)
			ast.Inspect(declaration, func(node ast.Node) bool {
				literal, ok := node.(*ast.CompositeLit)
				if !ok || !stringSlice(literal.Type) {
					return true
				}
				elements := stringElements(literal)
				if !slices.Contains(elements, "internal/native") || slices.Contains(elements, "internal/native_runtime") {
					return true
				}
				key := name + " " + holder
				if _, ok := wholeNativeWithoutTheRuntime[key]; ok {
					allowed[key] = true
					return true
				}
				found = append(found, fmt.Sprintf("%s:%d (%s)", name, fileSet.Position(literal.Pos()).Line, holder))
				return true
			})
		}
	}
	for key := range wholeNativeWithoutTheRuntime {
		if !allowed[key] {
			found = append(found, key+": allowed, but holds no list naming internal/native whole; drop it")
		}
	}
	slices.Sort(found)
	if len(found) > 0 {
		t.Fatalf("these name internal/native whole without internal/native_runtime, where the runtime's C now is:\n%s", strings.Join(found, "\n"))
	}
}

// declarationName is the function, method or package variables a declaration holds.
func declarationName(declaration ast.Decl) string {
	switch declaration := declaration.(type) {
	case *ast.FuncDecl:
		return declaration.Name.Name
	case *ast.GenDecl:
		var names []string
		for _, spec := range declaration.Specs {
			if value, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range value.Names {
					names = append(names, name.Name)
				}
			}
		}
		return strings.Join(names, ",")
	}
	return ""
}

// stringSlice is whether a literal's type is []string.
func stringSlice(expression ast.Expr) bool {
	array, ok := expression.(*ast.ArrayType)
	if !ok || array.Len != nil {
		return false
	}
	element, ok := array.Elt.(*ast.Ident)
	return ok && element.Name == "string"
}

// stringElements are a literal's elements that are string literals, unquoted.
func stringElements(literal *ast.CompositeLit) []string {
	var elements []string
	for _, element := range literal.Elts {
		basic, ok := element.(*ast.BasicLit)
		if !ok || basic.Kind != token.STRING {
			continue
		}
		if value, err := strconv.Unquote(basic.Value); err == nil {
			elements = append(elements, value)
		}
	}
	return elements
}
