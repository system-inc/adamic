// Each source here reaches a walk outside the package with nothing declaring it, and must be flagged where it is.
package flagged

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func visit(path string, entry fs.DirEntry, err error) error { return err }

func straightToWalkDir() {
	_ = filepath.WalkDir("../..", visit) // want `^walking files outside this package that its key doesn't name: a change to them won't rerun this test; declare it with // Reads: <glob> and a matching reads line in cloud/fast-gate/executors\.txt$`
}

func throughAVariable() {
	root := "../.." // want `walking files outside this package`
	_ = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error { return err })
}

// The runtime_owners_test.go shape: a joined constant handed to a helper that a TempDir test also calls.
func censusOfTheRepository(t *testing.T) []string {
	return censusIn(t, filepath.Join("..", "..")) // want `walking files outside this package`
}

func censusOfAFixture(t *testing.T) []string {
	return censusIn(t, t.TempDir())
}

func censusIn(t *testing.T, root string) []string {
	var files []string
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		files = append(files, path)
		return err
	})
	return files
}

func absolute() {
	root, err := filepath.Abs("../..") // want `walking files outside this package`
	if err != nil {
		return
	}
	_, _ = os.ReadDir(filepath.Join(root, "internal"))
}

const repository = "../../.."

func namedConstant() {
	_, _ = os.ReadDir(repository) // want `walking files outside this package`
}

// The dry shape: the test hands "../.." to checkMirrors, which hands it on to goFiles, which walks it.
func TestCohereMirrors(t *testing.T) {
	if problems := checkMirrors("../.."); len(problems) > 0 { // want `walking files outside this package`
		t.Error(problems)
	}
}

// The walk-up shape of cmd/adamic-gate's parallel_test.go, flagged where the root-finding function is called.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		data, err := os.ReadFile(filepath.Join(directory, "go.mod"))
		if err == nil && strings.Contains(string(data), "module github.com/system-inc/adamic\n") {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("no repository root above the working directory")
		}
		directory = parent
	}
}

func TestEveryTest(t *testing.T) {
	root := repositoryRoot(t) // want `walking files outside this package`
	_ = eachTest(root)
}

func eachTest(root string) error {
	return filepath.WalkDir(root, visit)
}

// A walk-up written inline, with a for clause, is flagged where it starts from the working directory.
func inlineWalkUp() {
	working, _ := os.Getwd() // want `walking files outside this package`
	for directory := working; ; directory = filepath.Dir(directory) {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			_, _ = filepath.Glob(filepath.Join(directory, "cmd", "*", "*_test.go"))
			return
		}
	}
}

func glob() {
	_, _ = filepath.Glob("../*/testdata/*.json") // want `walking files outside this package`
}

func globThroughAParameter() {
	_ = goFilesIn("../lower") // want `walking files outside this package`
}

func goFilesIn(directory string) []string {
	files, _ := filepath.Glob(filepath.Join(directory, "*.go"))
	return files
}

// A ReadDir loop that recurses by hand walks as far as WalkDir would.
func readDirLoop() {
	_ = listAll("../../internal") // want `walking files outside this package`
}

func listAll(directory string) []string {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		names = append(names, path)
		if entry.IsDir() {
			names = append(names, listAll(path)...)
		}
	}
	return names
}

func fileSystem() {
	_ = fs.WalkDir(os.DirFS("../.."), ".", visit) // want `walking files outside this package`
}

func parseDirectory() {
	_, _ = parser.ParseDir(token.NewFileSet(), "../load", nil, 0) // want `walking files outside this package`
}

func concatenated() {
	_, _ = os.ReadDir("../" + "stage1") // want `walking files outside this package`
}

func walkedInAClosure() {
	root := filepath.Join("..", "..", "cmd") // want `walking files outside this package`
	walk := func() error { return filepath.WalkDir(root, visit) }
	_ = walk()
}

func emptyGlob() {
	// Reads:
	_ = filepath.WalkDir("../..", visit) // want `walking files outside this package`
}

func twoGlobsInOneComment() {
	// Reads: internal/*.go cmd/*.go
	_ = filepath.WalkDir("../..", visit) // want `walking files outside this package`
}

func noSpaceAfterTheColon() {
	//Reads: internal/**
	_ = filepath.WalkDir("../..", visit) // want `walking files outside this package`
}

func declaredTooFarAbove() {
	// Reads: internal/**
	_ = os.Getenv("HOME")
	_ = filepath.WalkDir("../..", visit) // want `walking files outside this package`
}

// The comment belongs on the source, not on the walk in the helper the source reaches.
func declaredOnTheWalkInstead() {
	_ = declaredWalk("../..") // want `walking files outside this package`
}

func declaredWalk(root string) error {
	// Reads: internal/**
	return filepath.WalkDir(root, visit)
}

// The css_test.go shape: a path out of the package joined with values the rule doesn't read is judged by its start.
func siblingSlices() {
	for _, slice := range []string{"flagged", "selector", "values"} {
		_, _ = os.ReadDir(filepath.Join("..", slice)) // want `walking files outside this package`
	}
}

// A path out of the package and back into another one leaves it, and the helper's join is read at each call.
func intoAnotherPackage() {
	_ = globUnder("../other") // want `walking files outside this package`
	_ = globUnder("../flagged")
}

func globUnder(root string) []string {
	paths, _ := filepath.Glob(filepath.Join(root, "testdata", "*.a"))
	return paths
}

// A helper that calls itself with a path joined onto its parameter still settles.
func recursiveJoin() {
	deeper("../..") // want `walking files outside this package`
}

func deeper(directory string) {
	if _, err := os.ReadDir(directory); err != nil {
		return
	}
	deeper(filepath.Join(directory, "next"))
}
