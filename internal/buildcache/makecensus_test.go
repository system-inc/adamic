package buildcache

import (
	"fmt"
	"go/ast"
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
	// Hermetic: a runner has no GOROOT, and a source importer reads the standard library from one (chg_7fdbgdq4 on Cloud:
	// "could not import bytes ... ($goroot not set)"). The census needs only which of this package's own objects each use
	// names, which resolves without any import's contents, so every import is an empty package and the errors that leaves
	// (a selector into one) are left unsaid. A file that doesn't parse still fails it.
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}}
	checked, _ := (&types.Config{Importer: emptyImports{}, Error: func(error) {}}).Check("github.com/system-inc/adamic/internal/buildcache", fileSet, files, info)
	if checked == nil {
		return nil, fmt.Errorf("the package in %s didn't type-check at all", directory)
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

// emptyImports answers every import with an empty, complete package named for its path's last element (the one
// before a major version suffix, as math/rand/v2 is rand), so the census type-checks with no GOROOT and no module cache.
type emptyImports struct{}

func (emptyImports) Import(path string) (*types.Package, error) {
	elements := strings.Split(path, "/")
	name := elements[len(elements)-1]
	if len(elements) > 1 && len(name) > 1 && name[0] == 'v' && strings.Trim(name[1:], "0123456789") == "" {
		name = elements[len(elements)-2]
	}
	imported := types.NewPackage(path, name)
	imported.MarkComplete()
	return imported, nil
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

// The census holds with no GOROOT, as on a runner, and still sees a make file using a machinery helper beside a standard
// library call it can't see into: a fixture package whose make file calls helper (defined in helper.go) and uses bytes.
// Mutants: the census back on a source importer (with no GOROOT it stops at the first import and sees nothing, which the
// whole package's census alone passes); the empty imports' errors read as a failure.
// Not parallel: t.Setenv("GOROOT"), which is process-wide.
func TestTheMakeCensusHoldsWithNoGoRootAndStillSeesAMachineryUse(t *testing.T) {
	t.Setenv("GOROOT", "")
	directory := t.TempDir()
	for name, content := range map[string]string{
		"helper.go":  "package buildcache\n\nfunc helper() int { return 1 }\n",
		"make_x.go":  "package buildcache\n\nimport \"bytes\"\n\nfunc made() int { var buffer bytes.Buffer; buffer.WriteString(\"x\"); return helper() + buffer.Len() }\n",
		"make_y.go":  "package buildcache\n\nfunc alsoMade() int { return made() }\n",
		"other_x.go": "package buildcache\n\nfunc unused() {}\n",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	census, err := makeCensus(directory)
	if err != nil || !slices.Equal(census, []string{"make_x.go: helper (defined in helper.go)"}) {
		t.Fatalf("census %q, err %v, want make_x.go's use of helper alone", census, err)
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

// IsMachinery is exactly what a binary linking this package leaves out of its key, of this package's own files, plus
// everything under cmd/, which no product links (#kjgmzt8, Node's ruling): a file of the package is machinery when, and
// only when, cmd/find's recipe doesn't key it. So the seed rule Loom asks it of grows with the split's exclusions, never
// a second list. Mutants, each caught: a make file called machinery; settle.go not; cmd/traced not; a file outside the
// package (go.mod, another package's) called machinery.
// Not parallel: go list runs in the repository.
func TestMachineryIsWhatAProductsKeyLeavesOut(t *testing.T) {
	root := repositoryRootForTest(t)
	linking := recipeOf(root, journalEntry{Package: buildcachePackage + "/cmd/find"})
	if linking.reason != "" {
		t.Fatal(linking.reason)
	}
	entries, err := os.ReadDir(filepath.Join(root, "internal", "buildcache"))
	if err != nil {
		t.Fatal(err)
	}
	machinery := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		file := "internal/buildcache/" + entry.Name()
		keyed := slices.Contains(linking.entries, readEntry{Kind: "content", Path: file})
		if IsMachinery(file) == keyed {
			t.Errorf("%s: keyed %v, machinery %v", file, keyed, IsMachinery(file))
		}
		if !keyed {
			machinery++
		}
	}
	if machinery == 0 {
		t.Fatal("no file of the package is machinery, so the comparison saw nothing left out")
	}
	err = filepath.WalkDir(filepath.Join(root, "internal", "buildcache", "cmd"), func(file string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(root, file)
		if err == nil && !IsMachinery(filepath.ToSlash(relative)) {
			t.Errorf("%s isn't machinery", relative)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, outside := range []string{"go.mod", "go.sum", "go.work", "internal/toolchain/toolchain.go", "internal/buildcache/make_go.go",
		"internal/buildcachex/settle.go", "internal/buildcache/testdata/settle.go", "cmd/traced/main.go"} {
		if IsMachinery(outside) {
			t.Errorf("%s is called machinery", outside)
		}
	}
}
