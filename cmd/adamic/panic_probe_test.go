package main

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// A probe, never submitted: which review and gaps programs panic the compiler on this tree. It always fails, so its
// list reaches the log.
func TestProbeCompilerPanics(t *testing.T) {
	var paths []string
	for _, root := range []string{"../../internal/oracle/testdata/review", "../../review", "../../stage1"} {
		_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err == nil && !entry.IsDir() && (strings.HasSuffix(path, ".a") || strings.HasSuffix(path, ".ts")) && (strings.Contains(path, "review") || strings.Contains(path, "/gaps/")) && !strings.HasSuffix(path, ".d.ts") {
				paths = append(paths, path)
			}
			return nil
		})
	}
	var lock sync.Mutex
	var found []string
	t.Run("all", func(t *testing.T) {
		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				code, _, diagnostic := runGuardCommand(t, []string{"c", path})
				if code == 4 || strings.Contains(diagnostic, "panic:") {
					first, _, _ := strings.Cut(diagnostic, "\n")
					lock.Lock()
					found = append(found, path+" exit "+string(rune('0'+code))+": "+first)
					lock.Unlock()
				}
			})
		}
	})
	sort.Strings(found)
	t.Fatalf("%d of %d programs panic the compiler:\n%s", len(found), len(paths), strings.Join(found, "\n"))
}
