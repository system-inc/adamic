package load

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// countingFileSystem is the disk, counting each path FileExists is asked.
type countingFileSystem struct {
	vfs.FS
	mutex  sync.Mutex
	counts map[string]int
}

func (fileSystem *countingFileSystem) FileExists(path tspath.RootedFilePath) bool {
	fileSystem.mutex.Lock()
	fileSystem.counts[path.AsString()]++
	fileSystem.mutex.Unlock()
	return fileSystem.FS.FileExists(path)
}

// A load stats an .a file and its .a.ts twin once each, however often it names the file: Program.Where asks
// displayName for every unit and finding lowering records, on the disk itself, and one internal/lower test binary
// stat'd one fixture's pair 238,336 times (#npd568j). Mutant: adamicFile's kept answer never read, and the twin is
// asked once a name.
func TestALoadStatsAnAdamicFileOnce(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	if writeError := os.WriteFile(
		filepath.Join(directory, "shapes.a"),
		[]byte("export const shape = 1;\n"),
		0o644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	disk := &countingFileSystem{FS: osvfs.FS(), counts: map[string]int{}}
	fs := &sourceFileSystem{FS: disk}
	twin := tspath.RootedDirectoryPathFromAbsolute(directory).ResolveFile("shapes.a.ts")
	for range 1000 {
		if name := fs.displayName(twin); !strings.HasSuffix(name, "/shapes.a") {
			test.Fatalf("displayName gave %s, want the .a file", name)
		}
	}
	for path, count := range disk.counts {
		if count != 1 {
			test.Fatalf("one load asked %s %d times, want once", path, count)
		}
	}
	if len(disk.counts) != 2 {
		test.Fatalf("one load asked %d paths, want the twin and the .a file: %v", len(disk.counts), disk.counts)
	}
}

// A file made between two loads is found by the second: the kept answers are one load's, so a test that misses an
// import, writes it, and loads again sees it (Node's rule on #npd568j). Within a load the sources hold still, as
// cachedvfs already holds them. Mutant: the answers kept across loads, so the second load reads the first's "missing"
// and the import stays unresolved.
func TestAFileMadeBetweenTwoLoadsIsFound(test *testing.T) {
	test.Parallel()
	paths := writeProgram(
		test,
		[2]string{"main.a", "import { value } from './late.a';\nconsole.log(value.toString());\n"},
	)
	if _, loadError := Load(paths); loadError == nil || !strings.Contains(loadError.Error(), "late.a") {
		test.Fatalf("the first load, before late.a exists, gave %v; want the unresolved import", loadError)
	}
	late := filepath.Join(filepath.Dir(paths[0]), "late.a")
	if writeError := os.WriteFile(late, []byte("export const value = 7;\n"), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	if _, loadError := Load(paths); loadError != nil {
		test.Fatalf("the second load, after late.a was written, gave %v; want it found", loadError)
	}
}
