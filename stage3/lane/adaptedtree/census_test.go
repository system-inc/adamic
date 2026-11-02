package adaptedtree

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leancensus"
)

// buildcacheSlice are the packages of the tree buildcache links, which a lean build's recipe keys whole with buildcache
// (buildcache's leanRecipeOf): the build may reach them as it likes, and nothing else of the tree.
var buildcacheSlice = []string{
	leancensus.ModulePath + "/internal/buildcache", leancensus.ModulePath + "/internal/buildcache/goplan",
	leancensus.ModulePath + "/internal/buildcache/location", leancensus.ModulePath + "/internal/childguard",
	leancensus.ModulePath + "/internal/toolchain",
}

// censusExports is the export data of this package and every package it and its tests import, built ahead
// (TestProduct_AdaptedTreeCensusExports), so the build's census type-checks with real types and runs no go.
func censusExports(t testing.TB) string {
	t.Helper()
	return buildcache.GoExports(t, "./stage3/lane/adaptedtree")
}

func TestProduct_AdaptedTreeCensusExports(t *testing.T) {
	t.Parallel()
	censusExports(t)
}

// The adapted tree's build, a LeanProductOf build of what it's handed, reaches nothing but this package's file, which
// its recipe keys, and buildcache's slice, and reads no package variable written outside its initializer: what its key
// doesn't hold can't change what it builds. The census sees Go alone; what apply.py and node read as it runs is the
// window's, keyed under the trace (buildcache's TestAHandedBuildsUndeclaredReadIsKeyedByItsWindow). Mutant that fails
// it: a package variable the build reads, written by Product.
func TestTheAdaptedTreeBuildReachesOnlyItsDeclaredFile(t *testing.T) {
	t.Parallel()
	// TestMain runs the package's tests from stage3/lane, where repositoryFromLane is the checkout.
	root, err := filepath.Abs(repositoryFromLane)
	if err != nil {
		t.Fatal(err)
	}
	exports, tree, _, err := leancensus.Imports(censusExports(t))
	if err != nil {
		t.Fatal(err)
	}
	// A slice package that isn't listed is one buildcache no longer links, or renamed: the slice is out of date.
	for _, pkg := range buildcacheSlice {
		if !slices.Contains(tree, pkg) {
			t.Fatalf("buildcache's slice names %s, which this package's listing doesn't hold: %v", pkg, tree)
		}
	}
	census := leancensus.Census{
		Root: root, Directory: filepath.Join(root, "stage3", "lane", "adaptedtree"), ImportPath: leancensus.ModulePath + "/stage3/lane/adaptedtree",
		Entries:  []string{"build"},
		Declared: []string{"stage3/lane/adaptedtree/adaptedtree.go"},
		Slice:    buildcacheSlice,
		Exports:  exports, Tree: tree, TestsBuildLean: true,
	}
	findings, err := census.Findings()
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		t.Fatalf("the adapted tree's build reaches past what its key holds:\n%s", strings.Join(findings, "\n"))
	}
}
