package program_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/types/program"
)

// A file saved while a run is reading is never recorded against that run's verdict. The run may have read
// it before the save, and the stat taken when it records would sign the new bytes, so the next run would
// replay a verdict about bytes that are gone. The same recording made after the save, by a run that began
// reading after it, is fine, which is what makes the refusal mean something.
func TestARunIsNotRecordedOverAFileChangedWhileItRan(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "a.ts")
	if err := os.WriteFile(file, []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	readSince := time.Now()
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(file, []byte("export const a = 2;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := program.RecordRunCache("key", []string{file}, nil, nil, nil, nil, 0, readSince); err == nil ||
		!strings.Contains(err.Error(), "a.ts changed after the run began reading") {
		t.Fatalf("a run that began reading before a.ts was saved was recorded (%v)", err)
	}
	if _, err := program.RecordRunCache("key", []string{file}, nil, nil, nil, nil, 0, time.Now()); err != nil {
		t.Fatalf("a run that began reading after the save was refused: %v", err)
	}
}

// Different bytes of the same size under the old modification time, as cp -p, rsync -t and touch -r leave
// a file, are a miss. Only the change time moved.
func TestASameSizeEditUnderARestoredModificationTimeIsAMiss(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "a.ts")
	if err := os.WriteFile(file, []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cache, err := program.RecordRunCache("key", []string{file}, nil, nil, nil, nil, 0, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Check("key"); err != nil {
		t.Fatalf("an untouched file missed, so the miss below proves nothing: %v", err)
	}
	information, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(file, []byte("export const a = 9;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(file, information.ModTime(), information.ModTime()); err != nil {
		t.Fatal(err)
	}
	if err := cache.Check("key"); !errors.Is(err, program.ErrRunCacheMiss) {
		t.Fatalf("a same-size edit under the restored modification time was a hit (%v)", err)
	}
}

// A recorded directory is an input by its own signature, and its parent is not: the directory's change time
// and inode already catch it being replaced. Recording the parent made every replay depend on whatever else
// lives beside the project, ~/Projects on a developer's machine and the benchmark's work directory, where
// something is created all day (#r9jevk9). A file's directory is still recorded, which is what catches a file
// added beside one the run read.
func TestADirectoryInputDoesNotMakeItsParentOne(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	project := filepath.Join(parent, "project")
	if err := os.MkdirAll(filepath.Join(project, "source"), 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(project, "source", "a.ts")
	if err := os.WriteFile(file, []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cache, err := program.RecordRunCache("key", []string{project, file}, nil, nil, nil, nil, 0, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	recorded := map[string]bool{}
	for _, input := range cache.Inputs {
		recorded[input.Path] = true
	}
	for _, want := range []string{project, file, filepath.Join(project, "source")} {
		if !recorded[want] {
			t.Errorf("%s is not recorded", want)
		}
	}
	if recorded[parent] {
		t.Errorf("the directory holding the project, %s, is recorded because the project directory is", parent)
	}
	if err := os.WriteFile(filepath.Join(parent, "beside.txt"), []byte("beside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cache.Check("key"); err != nil {
		t.Fatalf("a file created beside the project made the run miss: %v", err)
	}
}

// --cache-dump says, under each recorded run, whether it would replay now, and when it would not, which input
// changed: a miss is never silent (#r9jevk9).
func TestTheCacheDumpNamesTheInputThatKeepsARunFromReplaying(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "a.ts")
	if err := os.WriteFile(file, []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cache, err := program.RecordRunCache("key", []string{file}, nil, nil, nil, nil, 0, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	table := program.NewCacheTable()
	table.Runs["--no-fix"] = cache
	dump := func() string {
		var out bytes.Buffer
		program.DumpCacheTable(&out, directory, table, program.CacheTableIdentity{})
		return out.String()
	}
	if got := dump(); !strings.Contains(got, "now: every input as recorded") {
		t.Fatalf("an unchanged run is not reported as replayable:\n%s", got)
	}
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(file, []byte("export const a = 22;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := dump(); !strings.Contains(got, "now: would not replay") || !strings.Contains(got, "a.ts changed size") {
		t.Fatalf("the dump does not name the input that changed:\n%s", got)
	}
}
