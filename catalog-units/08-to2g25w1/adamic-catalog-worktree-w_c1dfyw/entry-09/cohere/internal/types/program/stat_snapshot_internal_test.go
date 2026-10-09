package program

import (
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// The check's answers are the disk's answers (#kdee854). Every kind of path a build asks about, a file, a
// directory, a symbolic link to either, a path with nothing there, each asked again with a trailing slash, is asked
// of the answering layer and of the real disk, and the two must agree on whether a file exists, whether a directory
// does, and what a stat says. The disk stopped dropping a trailing separator with typed file paths (#64159), so a
// slashed spelling finds nothing there, and the check leaves it to the disk, as it does a spelling with a doubled
// separator.
func TestTheCheckedStatsAnswerAsTheDiskDoes(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need privileges here")
	}
	root := t.TempDir()
	file := filepath.Join(root, "file.ts")
	directory := filepath.Join(root, "directory")
	fileLink := filepath.Join(root, "file-link.ts")
	directoryLink := filepath.Join(root, "directory-link")
	absent := filepath.Join(root, "absent.ts")
	if err := os.WriteFile(file, []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(file, fileLink); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(directory, directoryLink); err != nil {
		t.Fatal(err)
	}

	recorded, err := RecordRunCache("key", []string{file, directory, fileLink, directoryLink}, nil, []string{absent}, nil, []byte("verdict"), 0, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	snapshot := NewStatSnapshot()
	if err := recorded.CheckNoting("key", snapshot); err != nil {
		t.Fatalf("an untouched tree missed: %v", err)
	}

	disk := osvfs.FS()
	var answered atomic.Int64
	checked := &checkedStatsFS{FS: disk, snapshot: snapshot, answered: &answered}
	paths := []string{file, directory, fileLink, directoryLink, absent}
	for _, path := range paths {
		for _, asked := range []string{path, path + "/", filepath.Dir(path) + "//" + filepath.Base(path)} {
			if got, want := checked.FileExists(tspath.RootedFilePath(asked)), disk.FileExists(tspath.RootedFilePath(asked)); got != want {
				t.Errorf("FileExists(%s): the check says %t, the disk %t", asked, got, want)
			}
			if got, want := checked.DirectoryExists(tspath.RootedDirectoryPath(asked)), disk.DirectoryExists(tspath.RootedDirectoryPath(asked)); got != want {
				t.Errorf("DirectoryExists(%s): the check says %t, the disk %t", asked, got, want)
			}
			got, want := checked.Stat(tspath.RootedPath(asked)), disk.Stat(tspath.RootedPath(asked))
			switch {
			case (got == nil) != (want == nil):
				t.Errorf("Stat(%s): the check found something %t, the disk %t", asked, got != nil, want != nil)
			case got != nil && (got.IsDir() != want.IsDir() || got.Size() != want.Size() || !got.ModTime().Equal(want.ModTime()) || got.Mode() != want.Mode()):
				t.Errorf("Stat(%s): the check says %v %d %v %v, the disk %v %d %v %v", asked,
					got.IsDir(), got.Size(), got.ModTime(), got.Mode(), want.IsDir(), want.Size(), want.ModTime(), want.Mode())
			}
		}
	}
	// Three questions of each path as the check statted it, and none in the other spellings, or the agreement
	// above was the disk agreeing with itself.
	if got, want := answered.Load(), int64(3*len(paths)); got != want {
		t.Errorf("the check answered %d questions, want %d: one of each kind for each path, as the check statted it", got, want)
	}
}
