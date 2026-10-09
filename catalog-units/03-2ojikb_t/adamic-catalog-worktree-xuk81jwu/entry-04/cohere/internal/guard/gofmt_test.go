package guard

import (
	"bytes"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
)

// Every Go file cohere owns is gofmt-clean, so `go test ./...` refuses an unformatted tree before a
// release does. The release's first step is `gofmt -l ./command ./internal`, and on 2026-10-03 head
// failed it (62ac5d74) with nothing in the suite to say so.
//
// Wider than the release's set: every `.go` file in the module, testdata, TypeScript-shim/ and the shim
// generators included, since the generator emitting unformatted shims is a regression of the same kind.
// Excluded are only what cohere does not own:
//   - the TypeScript submodule, which is upstream's;
//   - directories whose name begins with a dot, `.cache` among them, which holds snapshots and worktrees
//     of other trees;
//   - node_modules, which is npm's.
//
// Files whose name begins with a dot are skipped too, as gofmt skips them. go/format is gofmt's own printer,
// run in process and on every core rather than as a command. A file that does not parse fails too, since
// gofmt stops the release on it as surely.
func TestEveryGoFileCohereOwnsIsGofmtClean(t *testing.T) {
	t.Parallel()

	// This package sits two directories below the module root.
	module, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	var files []string
	err = filepath.WalkDir(module, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != module && (path == filepath.Join(module, "TypeScript") ||
				strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".go") && !strings.HasPrefix(entry.Name(), ".") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// The same guard the release keeps: a walk that found nothing would pass over a tree nobody examined.
	// The release's own set is required in particular, so a walk that lost it cannot pass on the rest.
	var underReleaseSet int
	for _, path := range files {
		relative, _ := filepath.Rel(module, path)
		if strings.HasPrefix(filepath.ToSlash(relative), "command/") || strings.HasPrefix(filepath.ToSlash(relative), "internal/") {
			underReleaseSet++
		}
	}
	if underReleaseSet == 0 {
		t.Fatal("found no Go files under command/ and internal/, so a clean answer would mean nothing")
	}

	var mutex sync.Mutex
	var unformatted []string
	work := make(chan string)
	var workers sync.WaitGroup
	for range runtime.GOMAXPROCS(0) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for path := range work {
				problem := gofmtProblem(path)
				if problem == "" {
					continue
				}
				mutex.Lock()
				relative, _ := filepath.Rel(module, path)
				unformatted = append(unformatted, filepath.ToSlash(relative)+problem)
				mutex.Unlock()
			}
		}()
	}
	for _, path := range files {
		work <- path
	}
	close(work)
	workers.Wait()

	if len(unformatted) > 0 {
		sort.Strings(unformatted)
		t.Fatalf("%d of %d Go files are not gofmt-clean, and any under command/ or internal/ would fail the release's first step; run `gofmt -w` on them:\n  %s",
			len(unformatted), len(files), strings.Join(unformatted, "\n  "))
	}
}

// gofmtProblem says what is wrong with one file, or nothing when gofmt would leave it as it is.
func gofmtProblem(path string) string {
	source, err := os.ReadFile(path)
	if err != nil {
		return " (unreadable: " + err.Error() + ")"
	}
	formatted, err := format.Source(source)
	if err != nil {
		return " (does not parse: " + err.Error() + ")"
	}
	if !bytes.Equal(source, formatted) {
		return " (not formatted)"
	}
	return ""
}
