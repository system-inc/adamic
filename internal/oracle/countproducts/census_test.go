package countproducts

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leancensus"
	"github.com/system-inc/adamic/internal/toolchain"
)

// importPath is this package's, as go names it.
const importPath = leancensus.ModulePath + "/internal/oracle/countproducts"

// censusExports is the export data of this package and every package it and its tests import, built ahead
// (TestProduct_CountProductsCensusExports), so the census type-checks with real types.
func censusExports(t testing.TB) string {
	t.Helper()
	return buildcache.GoExports(t, "./internal/oracle/countproducts")
}

func TestProduct_CountProductsCensusExports(t *testing.T) {
	t.Parallel()
	censusExports(t)
}

// repositoryRoot is the tree this package sits in, three directories up.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// dependencies are the packages this package's own files link, its tests apart, itself among them, as go list names
// them: what a count's lean recipe keys (buildcache's leanRecipeOf, the caller's import closure), and what
// TestTheMeasuringImportsNothingThatLowers holds. go list, as internal/native_runtime/compiler_free's guard reads it.
func dependencies(t *testing.T) []string {
	t.Helper()
	listed, err := toolchain.CommandIn(repositoryRoot(t), "go", "list", "-deps", "-f", "{{.ImportPath}}", "./internal/oracle/countproducts").Output()
	if err != nil {
		t.Fatalf("go list -deps ./internal/oracle/countproducts: %v", err)
	}
	found := strings.Fields(string(listed))
	if !slices.Contains(found, importPath) {
		t.Fatalf("go list's dependencies of internal/oracle/countproducts don't hold it, so they aren't its: %q", found)
	}
	return found
}

// recipeSlice are the packages of the tree a count's recipe keys whole: every one this package links (leanRecipeOf
// lists the caller with -deps, every non-test file of each), internal/native, ir and flow among them, which turn C into
// the counted binary. A count may reach them as it likes, since each of their files is in its key; what the census holds
// is this package's own code.
func recipeSlice(t *testing.T) []string {
	t.Helper()
	var slice []string
	for _, dependency := range dependencies(t) {
		if dependency != importPath && strings.HasPrefix(dependency, leancensus.ModulePath+"/") {
			slice = append(slice, dependency)
		}
	}
	return slice
}

// censusOf is the census of this package's code from entries (leancensus): what it reaches past declaredFiles and its
// recipe's slice, the package variables it reads that anything but their initializers write, and what it can't follow.
// This package's tests are read with it only when they build lean, and none does: internal/oracle's test binary settles
// every count, linking this package's files alone.
func censusOf(t *testing.T, entries []string) []string {
	t.Helper()
	root := repositoryRoot(t)
	exports, tree, _, err := leancensus.Imports(censusExports(t))
	if err != nil {
		t.Fatal(err)
	}
	census := leancensus.Census{
		Root: root, Directory: filepath.Join(root, "internal", "oracle", "countproducts"), ImportPath: importPath,
		Entries:  entries,
		Declared: declaredFiles(),
		Slice:    recipeSlice(t),
		Exports:  exports, Tree: tree, TestsBuildLean: false,
	}
	findings, err := census.Findings()
	if err != nil {
		t.Fatal(err)
	}
	return findings
}

// A count's build, BuildCountMeasurements, a LeanProductOf build of what it's handed, reaches nothing of this package
// but the files its key declares (declaredFiles), and reads no package variable anything but its initializer writes:
// what its key doesn't hold can't change a count. What clang and the counted programs read as they run is its window's.
// It's one of the two tests buildcache's leanCallers names as this package's census, the gate that opens the lean
// recipe, this package's import closure, to a count, with TestTheCountProductsPathReadsOnlyWhatItsKeyHolds, the other.
// Mutants that each fail it: run.go left out of declaredFiles; a package variable the build reads, written by Product.
func TestTheCountBuildReachesOnlyItsDeclaredFiles(t *testing.T) {
	t.Parallel()
	if findings := censusOf(t, []string{"BuildCountMeasurements"}); len(findings) > 0 {
		t.Fatalf("a count's build reaches past what its key holds:\n%s", strings.Join(findings, "\n"))
	}
}

// pathEntries are the ways into a count from another package: internal/oracle's rows (Rows), its product (Product),
// its one row measured in process (Measure), and the build its window runs. Every exported function of this package is
// one of them.
var pathEntries = []string{"Rows", "Product", "Measure", "BuildCountMeasurements"}

// A count's path to its build, from every exported function of this package, reads nothing in process its key doesn't
// hold: no package variable anything but its initializer writes, nothing past declaredFiles and its recipe's slice, and
// nothing the census can't follow, so what Product keys (keyingOf, the rows file's bytes among them) is all a count
// stands on. It's the other test buildcache's leanCallers names as this package's census, beside
// TestTheCountBuildReachesOnlyItsDeclaredFiles. Mutants that each fail it: an exported function left out of
// pathEntries; Product reading a package variable another function writes.
func TestTheCountProductsPathReadsOnlyWhatItsKeyHolds(t *testing.T) {
	t.Parallel()
	findings := censusOf(t, pathEntries)
	findings = append(findings, unfollowedEntries(t, repositoryRoot(t), pathEntries)...)
	if len(findings) > 0 {
		t.Fatalf("a count's path to its build reads past what its key holds:\n%s", strings.Join(findings, "\n"))
	}
}

// unfollowedEntries are this package's exported functions entries doesn't name, each a way into a count from another
// package that no census follows, read from its files as the line builds them (linux/amd64).
func unfollowedEntries(t *testing.T, root string, entries []string) []string {
	t.Helper()
	context := build.Default
	context.GOOS, context.GOARCH = "linux", "amd64"
	directory := filepath.Join(root, "internal", "oracle", "countproducts")
	listed, err := context.ImportDir(directory, 0)
	if err != nil {
		t.Fatal(err)
	}
	var findings []string
	for _, name := range listed.GoFiles {
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(directory, name), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Recv == nil && function.Name.IsExported() && !slices.Contains(entries, function.Name.Name) {
				findings = append(findings, fmt.Sprintf("internal/oracle/countproducts %s: exported, a way into a count, and no entry of pathEntries, so no census follows it", function.Name.Name))
			}
		}
	}
	return findings
}
