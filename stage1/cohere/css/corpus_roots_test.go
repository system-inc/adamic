package css

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// Every corpus file in the repository lies under a corpus root, so walking the roots finds exactly the files a walk of
// the whole repository did, in the same order. This test is that whole-repository walk, kept out of every product's
// process. A CSS file added outside the roots fails here, naming the root to add.
func TestCssCorpusRootsHoldEveryCorpusFile(test *testing.T) {
	test.Parallel()
	if !slices.IsSorted(oracles.CorpusRoots()) {
		test.Fatalf(
			"cssCorpusRoots %v aren't sorted, so their walk wouldn't keep a whole-repository walk's order",
			oracles.CorpusRoots(),
		)
	}
	// Reads: *.[cC][sS][sS]
	// Reads: *.[lL][eE][sS][sS]
	// Reads: *.[sS][cC][sS][sS]
	// Reads: cohere
	repo, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	var whole, rooted []string
	walkError := filepath.WalkDir(repo, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if oracles.IsCorpusFile(path) {
			whole = append(whole, path)
		}
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	if walkCorpusRootsError := oracles.WalkCorpusRoots(
		repo,
		func(path string) error { rooted = append(rooted, path); return nil },
	); walkCorpusRootsError != nil {
		test.Fatal(walkCorpusRootsError)
	}
	for _, path := range whole {
		if !slices.Contains(rooted, path) {
			relative, _ := filepath.Rel(repo, path)
			test.Errorf(
				"%s is a corpus file outside cssCorpusRoots: add its directory to them",
				filepath.ToSlash(relative),
			)
		}
	}
	if !slices.Equal(whole, rooted) {
		test.Fatalf(
			"the roots' walk finds %d corpus files, the repository's %d, or in another order",
			len(rooted),
			len(whole),
		)
	}
	for _, root := range oracles.CorpusRoots() {
		held := 0
		for _, path := range rooted {
			if strings.HasPrefix(path, filepath.Join(repo, filepath.FromSlash(root))+string(filepath.Separator)) {
				held++
			}
		}
		if held == 0 {
			test.Errorf("corpus root %s holds no corpus file: remove it", root)
		}
	}
	test.Logf(
		"%d corpus files, all under %d roots, in the whole-repository walk's order",
		len(rooted),
		len(oracles.CorpusRoots()),
	)
}
