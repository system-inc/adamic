package gocache

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newCache makes a directory in Go's cache layout, README included.
func newCache(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "README"), []byte(readmeMarker+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return directory
}

// put writes a cache file of the given size, last used at the given time.
func put(t *testing.T, directory, name string, bytes int, used time.Time) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("x", bytes)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, used, used); err != nil {
		t.Fatal(err)
	}
	return path
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// A directory that does not say it is a Go build cache is never trimmed, however large.
func TestADirectoryWithoutGosReadmeIsRefused(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := put(t, directory, "ab/0000ab-d", 4096, time.Now().Add(-48*time.Hour))
	if _, err := Trim(directory, Limit{Cap: 1, Target: 0, Spare: InUseWindow}, time.Now()); err == nil || !strings.Contains(err.Error(), "refusing to trim") {
		t.Fatalf("a directory with no README was trimmed: %v", err)
	}
	if !exists(file) {
		t.Fatal("a refused trim removed a file")
	}
}

// Over its cap, a cache loses its least recently used entries until it is at its target, and keeps the
// rest, everything that is not an entry, and what is used now.
func TestTheLeastRecentlyUsedEntriesGoFirstDownToTheTarget(t *testing.T) {
	t.Parallel()
	directory := newCache(t)
	now := time.Now()
	oldest := put(t, directory, "00/00aa-d", 400, now.Add(-72*time.Hour))
	older := put(t, directory, "01/01aa-a", 400, now.Add(-48*time.Hour))
	old := put(t, directory, "02/02aa-d", 400, now.Add(-24*time.Hour))
	recent := put(t, directory, "03/03aa-d", 400, now.Add(-time.Minute))
	notAnEntry := put(t, directory, "04/testexpand", 400, now.Add(-96*time.Hour))
	trimFile := put(t, directory, "trim.txt", 400, now.Add(-96*time.Hour))

	// 1,600 bytes of entries against a cap of 1,000 and a target of 900: the two oldest go.
	trimmed, err := Trim(directory, Limit{Cap: 1000, Target: 900, Spare: InUseWindow}, now)
	if err != nil {
		t.Fatal(err)
	}
	if trimmed.Before != 1600 || trimmed.After != 800 || trimmed.Removed != 2 || trimmed.InUse != 400 {
		t.Fatalf("trimmed %+v, want 1600 to 800 bytes in 2 entries, with 400 in use", trimmed)
	}
	for _, gone := range []string{oldest, older} {
		if exists(gone) {
			t.Errorf("%s survived", gone)
		}
	}
	for _, kept := range []string{old, recent, notAnEntry, trimFile, filepath.Join(directory, "README")} {
		if !exists(kept) {
			t.Errorf("%s was removed", kept)
		}
	}

	// Under its cap, nothing is touched.
	trimmed, err = Trim(directory, Limit{Cap: 1000, Target: 0, Spare: InUseWindow}, now)
	if err != nil || trimmed.Removed != 0 || trimmed.Before != 800 {
		t.Fatalf("a cache under its cap was trimmed: %+v, %v", trimmed, err)
	}
}

// Entries used within the in-use window are what a running build may be reading, so no trim removes them,
// even with the cache over its cap and nothing else to take. The #3sgjy0h gate removed young entries to
// hold a burst under the cap, and a live `go test` failed with "could not import ... no such file".
func TestEntriesInUseAreNeverRemoved(t *testing.T) {
	t.Parallel()
	directory := newCache(t)
	now := time.Now()
	stale := put(t, directory, "00/00aa-d", 400, now.Add(-InUseWindow-time.Minute))
	inUse := []string{
		put(t, directory, "01/01aa-d", 400, now.Add(-InUseWindow+time.Minute)),
		put(t, directory, "02/02aa-d", 400, now.Add(-30*time.Minute)),
		put(t, directory, "03/03aa-a", 400, now),
	}

	trimmed, err := Trim(directory, Limit{Cap: 1000, Target: 100, Spare: InUseWindow}, now)
	if err != nil {
		t.Fatal(err)
	}
	if trimmed.Removed != 1 || trimmed.InUse != 1200 || trimmed.After != 1200 {
		t.Fatalf("trimmed %+v, want only the stale entry removed and 1200 bytes in use kept over the cap", trimmed)
	}
	if exists(stale) {
		t.Error("the entry unused for longer than the window survived")
	}
	for _, path := range inUse {
		if !exists(path) {
			t.Errorf("%s, used within the window, was removed", path)
		}
	}
}

// A trim that has kept every build out spares only its short window: the entries a build read half an hour
// ago go when the cache is over its cap, and the last minutes' writes stay (#jc6ca7r).
func TestAShortSpareWindowReachesRecentlyUsedEntries(t *testing.T) {
	t.Parallel()
	directory := newCache(t)
	now := time.Now()
	older := put(t, directory, "00/00aa-d", 400, now.Add(-30*time.Minute))
	recent := put(t, directory, "01/01aa-d", 400, now.Add(-20*time.Minute))
	justWritten := put(t, directory, "02/02aa-a", 400, now.Add(-2*time.Minute))

	trimmed, err := Trim(directory, Limit{Cap: 1000, Target: 500, Spare: 10 * time.Minute}, now)
	if err != nil {
		t.Fatal(err)
	}
	if trimmed.Removed != 2 || trimmed.After != 400 || trimmed.InUse != 400 {
		t.Fatalf("trimmed %+v, want the two entries older than the window removed, oldest first", trimmed)
	}
	if exists(older) || exists(recent) || !exists(justWritten) {
		t.Errorf("after the trim: 30 minutes old %v, 20 minutes old %v, 2 minutes old %v; want only the last kept",
			exists(older), exists(recent), exists(justWritten))
	}
}

// An executable entry is a directory holding the binary. It is measured whole and removed file by file,
// and one holding anything but plain files is left alone.
func TestAnExecutableEntryIsRemovedByName(t *testing.T) {
	t.Parallel()
	directory := newCache(t)
	now := time.Now()
	old := now.Add(-48 * time.Hour)
	binary := put(t, directory, "00/00aa-d/cohere", 2000, old)
	if err := os.Chtimes(filepath.Dir(binary), old, old); err != nil {
		t.Fatal(err)
	}
	nested := put(t, directory, "01/01aa-d/inner/file", 2000, old)
	if err := os.Chtimes(filepath.Dir(filepath.Dir(nested)), old, old); err != nil {
		t.Fatal(err)
	}

	trimmed, err := Trim(directory, Limit{Cap: 1000, Target: 0, Spare: InUseWindow}, now)
	if err != nil {
		t.Fatal(err)
	}
	if trimmed.Removed != 1 || trimmed.Before != 2000 {
		t.Fatalf("trimmed %+v, want the one executable entry of 2000 bytes", trimmed)
	}
	if exists(filepath.Dir(binary)) {
		t.Error("the executable entry survived")
	}
	if !exists(nested) {
		t.Error("an entry holding a directory was removed")
	}
}

// Positive control on a real cache: the go command fills it, a trim to nothing removes every entry, and
// the go command opens the trimmed cache and builds again with no error.
func TestTheGoCommandBuildsFromATrimmedCache(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go command")
	}
	cache := t.TempDir()
	module := t.TempDir()
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module trimmed\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	build := func() {
		t.Helper()
		command := exec.Command("go", "build", "-o", filepath.Join(t.TempDir(), "trimmed"), ".")
		command.Dir = module
		command.Env = append(os.Environ(), "GOCACHE="+cache, "GOFLAGS=", "GOWORK=off")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("go build: %v: %s", err, output)
		}
	}
	build()

	// A day from now nothing is in use, so a target of zero takes every entry.
	trimmed, err := Trim(cache, Limit{Cap: 0, Target: 0, Spare: InUseWindow}, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if trimmed.Removed == 0 || trimmed.After != 0 {
		t.Fatalf("trimmed %+v, want every entry gone", trimmed)
	}
	build()
}
