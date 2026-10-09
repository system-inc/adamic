//go:build unix

package gocache

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A rebuild after a trim recomputes only entries the trim removed, and only the ones it needs: every entry
// the trim kept is the same file afterward, and the rebuild writes nothing the first build had not
// (#jc6ca7r). Not every removed entry comes back, since an action whose own result is still cached never
// asks for its inputs. Kept entries are told by their inode, since Go refreshes the time of an entry it uses.
func TestARebuildAfterATrimRecomputesOnlyWhatWasTrimmed(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go command")
	}
	cache := t.TempDir()
	module := t.TempDir()
	for name, contents := range map[string]string{
		"go.mod":         "module trimmed\n\ngo 1.21\n",
		"main.go":        "package main\n\nimport \"trimmed/lib\"\n\nfunc main() { println(lib.Value()) }\n",
		"lib/lib.go":     "package lib\n\nfunc Value() int { return 1 }\n",
		"lib/helper.go":  "package lib\n\nfunc helper() int { return 2 }\n",
		"other/other.go": "package other\n\nfunc Other() int { return 3 }\n",
	} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(module, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(module, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
		// An hour old, so both builds see the same sources. Go caches a package directory's module index
		// only once every file in it is two seconds old (modTimeCutoff in cmd/go/internal/modindex), so a
		// first build within two seconds of the writes cached no index for them, the rebuild did, and the
		// test read those three entries as recomputed work it had not trimmed.
		hourAgo := time.Now().Add(-time.Hour)
		if err := os.Chtimes(filepath.Join(module, name), hourAgo, hourAgo); err != nil {
			t.Fatal(err)
		}
	}
	build := func() {
		t.Helper()
		command := exec.Command("go", "build", "-o", t.TempDir()+string(filepath.Separator), "./...")
		command.Dir = module
		command.Env = append(os.Environ(), "GOCACHE="+cache, "GOFLAGS=", "GOWORK=off")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("go build: %v: %s", err, output)
		}
	}
	// Every entry, by path, with its inode.
	entries := func() map[string]uint64 {
		t.Helper()
		found := map[string]uint64{}
		filepath.WalkDir(cache, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !(strings.HasSuffix(path, "-a") || strings.HasSuffix(path, "-d")) {
				return err
			}
			information, err := os.Stat(path)
			if err != nil {
				return err
			}
			found[path] = inode(information)
			return nil
		})
		return found
	}

	build()
	before := entries()
	// Every entry three hours old, so the trim may take any of them, oldest first, down to half the cache.
	old := time.Now().Add(-3 * time.Hour)
	var size int64
	for path := range before {
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
		information, _ := os.Stat(path)
		size += information.Size()
	}
	trimmed, err := Trim(cache, Limit{Cap: size / 2, Target: size / 2, Spare: 10 * time.Minute}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	afterTrim := entries()
	if trimmed.Removed == 0 || len(afterTrim) == len(before) {
		t.Fatalf("the trim removed nothing (%+v), so it shows nothing", trimmed)
	}

	build()
	after := entries()
	recomputed := 0
	for path, number := range after {
		original, wasThere := before[path]
		kept, wasKept := afterTrim[path]
		switch {
		case !wasThere:
			t.Errorf("the rebuild wrote %s, which the first build had not", filepath.Base(path))
		case wasKept && number != kept:
			t.Errorf("%s, which the trim kept, was written again", filepath.Base(path))
		case !wasKept:
			if number == original {
				t.Errorf("%s, which the trim removed, has its old inode", filepath.Base(path))
			}
			recomputed++
		}
	}
	if recomputed > trimmed.Removed {
		t.Errorf("the rebuild recomputed %d entries after a trim that removed %d", recomputed, trimmed.Removed)
	}
	t.Logf("%d entries, %d trimmed, %d recomputed by the rebuild", len(before), trimmed.Removed, recomputed)
}

// inode is a file's inode number, which a rewrite changes and a time update does not.
func inode(information os.FileInfo) uint64 {
	if stat, ok := information.Sys().(*syscall.Stat_t); ok {
		return uint64(stat.Ino)
	}
	return 0
}
