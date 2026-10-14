package css

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// cssCorpusRoots are the directories, relative to the repository, that css's corpus takes its CSS, SCSS and Less files
// from. The Go oracle and the corpus product's key walk only these, never the whole repository: a walk of the
// repository lists its root and every directory under it, and since a product's read set takes the reads its process
// made before it, a new file anywhere in the repository moved every corpus-reading product, all 388 PostCSS answers
// among them (#t128y5x). Sorted, so walking them in order visits the files in the order a whole-repository walk did.
// TestCSSCorpusRootsHoldEveryCorpusFile keeps them complete.
var cssCorpusRoots = []string{"cohere/internal/format/css/testdata", "cohere/internal/lint/rules/tailwind", "stage1/cohere/css/testdata"}

// cssCorpusFile reports whether a file name has one of the corpus's extensions.
func cssCorpusFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".css", ".scss", ".less":
		return true
	}
	return false
}

// cssCorpusRootWalk walks every corpus root under the repository in order, as the Go oracle's side tests do, skipping
// .git and node_modules, and calls visit with each corpus file's absolute path.
func cssCorpusRootWalk(repo string, visit func(path string) error) error {
	for _, root := range cssCorpusRoots {
		err := filepath.WalkDir(filepath.Join(repo, filepath.FromSlash(root)), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if !cssCorpusFile(path) {
				return nil
			}
			return visit(path)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Every corpus file in the repository lies under a corpus root, so walking the roots finds exactly the files a walk of
// the whole repository did, in the same order. This test is that whole-repository walk, kept out of every product's
// process. A CSS file added outside the roots fails here, naming the root to add.
func TestCSSCorpusRootsHoldEveryCorpusFile(t *testing.T) {
	t.Parallel()
	if !slices.IsSorted(cssCorpusRoots) {
		t.Fatalf("cssCorpusRoots %v aren't sorted, so their walk wouldn't keep a whole-repository walk's order", cssCorpusRoots)
	}
	repo, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	var whole, rooted []string
	err = filepath.WalkDir(repo, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if cssCorpusFile(path) {
			whole = append(whole, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := cssCorpusRootWalk(repo, func(path string) error { rooted = append(rooted, path); return nil }); err != nil {
		t.Fatal(err)
	}
	for _, path := range whole {
		if !slices.Contains(rooted, path) {
			relative, _ := filepath.Rel(repo, path)
			t.Errorf("%s is a corpus file outside cssCorpusRoots: add its directory to them", filepath.ToSlash(relative))
		}
	}
	if !slices.Equal(whole, rooted) {
		t.Fatalf("the roots' walk finds %d corpus files, the repository's %d, or in another order", len(rooted), len(whole))
	}
	for _, root := range cssCorpusRoots {
		held := 0
		for _, path := range rooted {
			if strings.HasPrefix(path, filepath.Join(repo, filepath.FromSlash(root))+string(filepath.Separator)) {
				held++
			}
		}
		if held == 0 {
			t.Errorf("corpus root %s holds no corpus file: remove it", root)
		}
	}
	t.Logf("%d corpus files, all under %d roots, in the whole-repository walk's order", len(rooted), len(cssCorpusRoots))
}
