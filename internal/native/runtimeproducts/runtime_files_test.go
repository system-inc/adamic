package runtimeproducts

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runtimeFiles is what every runtime archive is keyed by, and buildcache values a directory by every name and file
// under it, so a test that sat under one of them would rebuild all of the archives each time it changed, the waste
// #tb722h2 ends. This holds every runtimeFiles entry to holding no Go test, and to being there. Its mutant: runtimeFiles
// naming internal/native_runtime whole again, where the source paths guard's test sits (Performance's should on H1).
func TestRuntimeFilesHoldNoTest(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	for _, entry := range runtimeFiles {
		// Reads: internal/native_runtime/c, internal/native
		err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(entry)), func(path string, file fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !file.IsDir() && strings.HasSuffix(file.Name(), "_test.go") {
				relative, _ := filepath.Rel(root, path)
				t.Errorf("runtimeFiles' %s holds the test %s, which every runtime archive's key would then hold", entry, filepath.ToSlash(relative))
			}
			return nil
		})
		if os.IsNotExist(err) {
			t.Errorf("runtimeFiles names %s, which isn't there", entry)
		} else if err != nil {
			t.Fatal(err)
		}
	}
}
