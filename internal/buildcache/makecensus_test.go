package buildcache

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	gotoken "go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
)

// makeCensus is every use, in a make_*.go file, of something this package defines in a machinery file: each as
// "make file: identifier (defined in file)", sorted, once. Empty when the make files decide a product's bytes with
// nothing but each other, other packages and the standard library.
func makeCensus(directory string) ([]string, error) {
	fileSet := gotoken.NewFileSet()
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var files []*ast.File
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(directory, name), nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		// The census reads the package this machine builds: a file another platform builds is no file here.
		if !buildsHere(name) {
			continue
		}
		files = append(files, file)
	}
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}}
	checked, err := (&types.Config{Importer: importer.ForCompiler(fileSet, "source", nil)}).Check("github.com/system-inc/adamic/internal/buildcache", fileSet, files, info)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for identifier, object := range info.Uses {
		if object == nil || object.Pkg() != checked || !object.Pos().IsValid() {
			continue
		}
		user := filepath.Base(fileSet.Position(identifier.Pos()).Filename)
		definer := filepath.Base(fileSet.Position(object.Pos()).Filename)
		if censusIsMakeFile(user) && !censusIsMakeFile(definer) {
			seen[fmt.Sprintf("%s: %s (defined in %s)", user, identifier.Name, definer)] = true
		}
	}
	census := []string{}
	for line := range seen {
		census = append(census, line)
	}
	sort.Strings(census)
	return census, nil
}

// censusIsMakeFile is isMakeFile, or the files MAKE_CENSUS_FILES lists (comma-separated) to try a candidate set.
func censusIsMakeFile(name string) bool {
	if explore := os.Getenv("MAKE_CENSUS_FILES"); explore != "" {
		return slices.Contains(strings.Split(explore, ","), name)
	}
	return isMakeFile(name)
}

// buildsHere says whether a file named name builds on this machine's platform by its name alone, as go reads a
// _<goos>.go suffix.
func buildsHere(name string) bool {
	base := strings.TrimSuffix(name, ".go")
	for _, goos := range []string{"linux", "darwin", "windows", "other"} {
		if strings.HasSuffix(base, "_"+goos) {
			return goos == hostOS() || goos == "other" && !slices.Contains([]string{"linux", "darwin"}, hostOS())
		}
	}
	return true
}

func hostOS() string { return runtime.GOOS }

// TestTheMakeFilesUseNothingFromTheMachinery is guard 1 of the buildcache split (#a328f0h, Loom's ruling 45922): the
// files that decide a product's bytes (make_*.go), the only files of this package a product's key reads, use nothing
// defined in the rest of it, the machinery that keys, traces, settles and stores. So a machinery change can't change a
// product's bytes, and leaving it out of every key is sound. Mutant: a make file calling a machinery helper.
func TestTheMakeFilesUseNothingFromTheMachinery(t *testing.T) {
	t.Parallel()
	census, err := makeCensus(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(census) > 0 {
		t.Fatalf("the make files use %d things the machinery defines:\n%s", len(census), strings.Join(census, "\n"))
	}
}

// A binary that links this package, and isn't its test, keys it by its make files and its listing alone (makeFilesOnly,
// #a328f0h): settle.go, reads.go and the rest of the machinery are in no recipe but this package's own tests', where
// every file still is. Mutants, each caught: makeFilesOnly off (settle.go keyed); the make files skipped too (make_go.go
// missing); the listing dropped; the package's own test recipe narrowed.
// Not parallel: go list runs in the repository.
func TestABinaryLinkingBuildcacheKeysItsMakeFilesAlone(t *testing.T) {
	root := repositoryRootForTest(t)
	linking := recipeOf(root, journalEntry{Package: buildcachePackage + "/cmd/find"})
	if linking.reason != "" {
		t.Fatal(linking.reason)
	}
	for _, want := range []readEntry{{Kind: "content", Path: "internal/buildcache/make_go.go"}, {Kind: "content", Path: "internal/buildcache/make_relative.go"},
		{Kind: "listing", Path: "internal/buildcache"}, {Kind: "content", Path: "internal/buildcache/cmd/find/main.go"}} {
		if !slices.Contains(linking.entries, want) {
			t.Errorf("cmd/find's recipe lacks %+v", want)
		}
	}
	for _, machinery := range []string{"settle.go", "reads.go", "store.go", "buildcache.go"} {
		if entry := (readEntry{Kind: "content", Path: "internal/buildcache/" + machinery}); slices.Contains(linking.entries, entry) {
			t.Errorf("cmd/find's recipe keys the machinery: %s", entry.Path)
		}
	}
	own := recipeOf(root, journalEntry{Package: buildcachePackage + ".test"})
	if own.reason != "" || !slices.Contains(own.entries, readEntry{Kind: "content", Path: "internal/buildcache/settle.go"}) {
		t.Errorf("this package's own tests' recipe doesn't key settle.go (%s)", own.reason)
	}
}
