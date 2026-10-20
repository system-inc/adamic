// Nothing here may be flagged: each walk stays in the package, can't be traced, is declared, or isn't a walk.
package allowed

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	fakefilepath "lookalike/filepath"
)

func visit(path string, entry fs.DirEntry, err error) error { return err }

func temporaryDirectory(t *testing.T) {
	_ = filepath.WalkDir(t.TempDir(), visit)
	_ = walkUnder(t.TempDir())
	_ = walkUnder(os.TempDir())
}

func walkUnder(root string) error {
	return filepath.WalkDir(root, visit)
}

func thisDirectory() {
	_ = filepath.WalkDir(".", visit)
	_, _ = filepath.Glob("*.go")
	_, _ = os.ReadDir("subdirectory")
	_ = walkUnder(filepath.Join(".", "subdirectory"))
	// A path that climbs out and back in names a directory of this package.
	_, _ = os.ReadDir("subdirectory/../other")
}

func ownTestdata() {
	_ = filepath.WalkDir("testdata", visit)
	_, _ = filepath.Glob(filepath.Join("testdata", "*.golden"))
	_ = walkUnder("testdata/corpus")
}

// The workingDirectory is the package directory itself, so walking it stays in the package.
func workingDirectory() {
	working, err := os.Getwd()
	if err != nil {
		return
	}
	_ = filepath.WalkDir(working, visit)
}

// The css_test.go shape: the root leaves the package only to name one fixed file, which no walk lists.
const repository = "../../.."

func fixedFile() {
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		return
	}
	_ = exec.Command("node", runner).Run()
	_, _ = os.ReadFile("../../cloud/fast-gate/product-order.json")
	_ = readOracle(filepath.Join("..", ".."))
}

func readOracle(root string) error {
	_, err := os.ReadFile(filepath.Join(root, "oracle", "adamic.mjs"))
	return err
}

func declaredOnTheLine() {
	_ = filepath.WalkDir("../..", visit) // Reads: internal/**/*.go
}

func declaredAbove() {
	// Reads: internal/**/*.go
	_ = filepath.WalkDir("../..", visit)
}

func declaredBelowAReason() {
	// The census reads every package's Go files, so it names them for the planner.
	// Reads: cmd/**/*.go
	// Reads: internal/**/*.go
	_ = walkUnder(filepath.Join("..", ".."))
}

func declaredWalkUp(t *testing.T) {
	// Reads: **/*_test.go
	_ = walkUnder(repositoryRoot(t))
}

func repositoryRoot(t *testing.T) string {
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		directory = filepath.Dir(directory)
	}
}

// A loop up through filepath.Dir that looks for something other than go.mod is not the walk to the repository root.
func nearestTestdata() string {
	directory, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(directory, "testdata")); err == nil {
			return directory
		}
		directory = filepath.Dir(directory)
	}
}

func walksNearestTestdata() {
	_ = walkUnder(nearestTestdata())
}

// A WalkDir of the same name at another package path is not the one the rule knows.
func lookalike() {
	_ = fakefilepath.WalkDir("../..", visit)
	_, _ = fakefilepath.Glob("../*")
}

// A root ranged out of a slice can't be traced, and the rule never guesses.
func untraced(roots []string) {
	for _, root := range roots {
		_ = filepath.WalkDir(root, visit)
	}
}

// The internal/oracle shape: a path out to the repository and back into this package's own directory stays in it.
const reviewRoot = "../allowed" + "/testdata/review"

func backIntoThisPackage() {
	_, _ = filepath.Glob(filepath.Join("..", "allowed", "testdata", "*.a"))
	_, _ = filepath.Glob(filepath.Join(reviewRoot, "agree", "*.a"))
	directory := "allowed/testdata/refused"
	_, _ = os.ReadDir(filepath.Join("..", directory))
	_ = globUnder("../allowed")
	_, _ = os.ReadDir(filepath.Dir("../allowed/testdata"))
}

func globUnder(root string) []string {
	paths, _ := filepath.Glob(filepath.Join(root, "testdata", "*.a"))
	return paths
}
