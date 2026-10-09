//go:build darwin || linux

package program

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// packFixture is a directory of files and a cache directory beside it, read through a fresh content pack
// each "run", the way successive cohere runs read a project.
type packFixture struct {
	t     *testing.T
	files string
	cache string
}

func newPackFixture(t *testing.T) *packFixture {
	root := t.TempDir()
	fixture := &packFixture{t: t, files: filepath.Join(root, "files"), cache: filepath.Join(root, "cache")}
	if err := os.MkdirAll(fixture.files, 0o755); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func (f *packFixture) path(name string) string { return filepath.Join(f.files, name) }

func (f *packFixture) write(name string, contents string) {
	f.t.Helper()
	if err := os.WriteFile(f.path(name), []byte(contents), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

// run opens the pack, reads every name through it, saves, and returns what it read and the pack's counts.
func (f *packFixture) run(names ...string) (map[string]string, int, int, int) {
	f.t.Helper()
	pack, err := OpenContentPack(f.cache)
	if err != nil {
		f.t.Logf("open: %v", err)
	}
	fileSystem := pack.wrap(osvfs.FS())
	read := map[string]string{}
	for _, name := range names {
		contents, ok := fileSystem.ReadFile(tspath.RootedFilePath(f.path(name)))
		if !ok {
			f.t.Fatalf("%s could not be read", name)
		}
		read[name] = contents
	}
	if err := pack.Save(); err != nil {
		f.t.Fatalf("saving the pack: %v", err)
	}
	hits, misses, damaged := pack.Counts()
	return read, hits, misses, damaged
}

func (f *packFixture) dataFile() string {
	f.t.Helper()
	matches, _ := filepath.Glob(filepath.Join(f.cache, "contents-*.pack"))
	if len(matches) != 1 {
		f.t.Fatalf("want one data file, found %v", matches)
	}
	return matches[0]
}

// A file read once is served from the pack the next time, unopened, and with the same bytes.
func TestTheContentPackServesAnUnchangedFile(t *testing.T) {
	t.Parallel()
	fixture := newPackFixture(t)
	fixture.write("a.ts", "export const a = 1;\n")
	fixture.write("b.ts", "export const b = 2;\n")
	if _, hits, misses, _ := fixture.run("a.ts", "b.ts"); hits != 0 || misses != 2 {
		t.Fatalf("an empty pack served %d and read %d, want 0 and 2", hits, misses)
	}
	read, hits, misses, damaged := fixture.run("a.ts", "b.ts")
	if hits != 2 || misses != 0 || damaged != 0 {
		t.Fatalf("the second run served %d, read %d and found %d damaged, want 2, 0 and 0", hits, misses, damaged)
	}
	if read["a.ts"] != "export const a = 1;\n" || read["b.ts"] != "export const b = 2;\n" {
		t.Fatalf("the pack served %q", read)
	}
}

// Every way a file's bytes can change is a read from disk, the change time catching the one the
// modification time cannot: different bytes of the same size with the modification time put back, as
// cp -p, rsync -t and touch -r all do. A replace by rename is a new inode.
func TestTheContentPackNeverServesChangedBytes(t *testing.T) {
	t.Parallel()
	for _, change := range []struct {
		name   string
		change func(f *packFixture)
	}{
		{"a different size", func(f *packFixture) { f.write("a.ts", "export const a = 100;\n") }},
		{"the same size", func(f *packFixture) { f.write("a.ts", "export const a = 9;\n") }},
		{"the same size with the modification time restored", func(f *packFixture) {
			information, err := os.Stat(f.path("a.ts"))
			if err != nil {
				f.t.Fatal(err)
			}
			f.write("a.ts", "export const a = 9;\n")
			if err := os.Chtimes(f.path("a.ts"), information.ModTime(), information.ModTime()); err != nil {
				f.t.Fatal(err)
			}
			if after, _ := os.Stat(f.path("a.ts")); !after.ModTime().Equal(information.ModTime()) || after.Size() != information.Size() {
				f.t.Fatal("the fixture did not restore the size and modification time, so this proves nothing")
			}
		}},
		{"a replace by rename, size and modification time kept", func(f *packFixture) {
			information, err := os.Stat(f.path("a.ts"))
			if err != nil {
				f.t.Fatal(err)
			}
			f.write("replacement.ts", "export const a = 9;\n")
			if err := os.Chtimes(f.path("replacement.ts"), information.ModTime(), information.ModTime()); err != nil {
				f.t.Fatal(err)
			}
			if err := os.Rename(f.path("replacement.ts"), f.path("a.ts")); err != nil {
				f.t.Fatal(err)
			}
		}},
	} {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			fixture := newPackFixture(t)
			fixture.t = t
			fixture.write("a.ts", "export const a = 1;\n")
			fixture.run("a.ts")
			// Change times move in nanoseconds, but a filesystem may store coarser ones.
			time.Sleep(10 * time.Millisecond)
			change.change(fixture)
			read, hits, _, _ := fixture.run("a.ts")
			want, _ := os.ReadFile(fixture.path("a.ts"))
			if read["a.ts"] != string(want) || hits != 0 {
				t.Fatalf("after %s the pack served %q (%d hits), and the disk holds %q", change.name, read["a.ts"], hits, want)
			}
			// And the new bytes are what the next run is served.
			if read, hits, _, _ := fixture.run("a.ts"); read["a.ts"] != string(want) || hits != 1 {
				t.Fatalf("the run after the change was served %q (%d hits), want %q from the pack", read["a.ts"], hits, want)
			}
		})
	}
}

// A damaged pack costs reads from disk and never a wrong byte, and the next write leaves a sound pack.
func TestADamagedContentPackIsReadAround(t *testing.T) {
	t.Parallel()
	for _, damage := range []struct {
		name        string
		damage      func(f *packFixture)
		damagedRead bool
	}{
		{"a flipped byte in the data", func(f *packFixture) {
			data, _ := os.ReadFile(f.dataFile())
			data[3] ^= 0xff
			os.WriteFile(f.dataFile(), data, 0o644)
		}, true},
		{"a data file truncated to half", func(f *packFixture) {
			information, _ := os.Stat(f.dataFile())
			os.Truncate(f.dataFile(), information.Size()/2)
		}, true},
		{"a flipped byte in the index", func(f *packFixture) {
			index := filepath.Join(f.cache, contentPackIndexName)
			data, _ := os.ReadFile(index)
			data[len(data)/2] ^= 0xff
			os.WriteFile(index, data, 0o644)
		}, false},
		{"a data file deleted", func(f *packFixture) { os.Remove(f.dataFile()) }, false},
	} {
		t.Run(damage.name, func(t *testing.T) {
			t.Parallel()
			fixture := newPackFixture(t)
			fixture.t = t
			fixture.write("a.ts", "export const a = 1;\n")
			fixture.write("b.ts", "export const b = 2;\n")
			fixture.run("a.ts", "b.ts")
			damage.damage(fixture)
			read, hits, _, damaged := fixture.run("a.ts", "b.ts")
			if read["a.ts"] != "export const a = 1;\n" || read["b.ts"] != "export const b = 2;\n" {
				t.Fatalf("a damaged pack served %q", read)
			}
			if damage.damagedRead && damaged == 0 {
				t.Errorf("the damage was read around without being counted, so the pack would not be rewritten")
			}
			if read, hits, _, damaged := fixture.run("a.ts", "b.ts"); hits+damaged == 0 || damaged != 0 || read["b.ts"] != "export const b = 2;\n" {
				t.Fatalf("the run after the damage was served %d, found %d damaged: the pack was not rewritten sound", hits, damaged)
			}
			_ = hits
		})
	}
}

// A pack truncated in place while a run has it mapped faults on the copy. The fault is a read from disk,
// not the end of the process.
func TestAContentPackTruncatedUnderItsMappingIsReadAround(t *testing.T) {
	t.Parallel()
	fixture := newPackFixture(t)
	fixture.write("a.ts", strings.Repeat("export const a = 1;\n", 4096))
	fixture.run("a.ts")
	pack, err := OpenContentPack(fixture.cache)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(fixture.dataFile(), 0); err != nil {
		t.Fatal(err)
	}
	contents, ok := pack.wrap(osvfs.FS()).ReadFile(tspath.RootedFilePath(fixture.path("a.ts")))
	if !ok || contents != strings.Repeat("export const a = 1;\n", 4096) {
		t.Fatalf("a pack truncated under its mapping served %d bytes", len(contents))
	}
	if _, _, damaged := pack.Counts(); damaged != 1 {
		t.Errorf("the fault was read around without being counted (%d damaged)", damaged)
	}
}

// A run that reads one new file appends it rather than rewriting the pack, and a pack mostly dead is
// compacted into a new data file, the old one removed.
func TestTheContentPackAppendsAndCompacts(t *testing.T) {
	t.Parallel()
	fixture := newPackFixture(t)
	fixture.write("a.ts", strings.Repeat("a", 1000))
	fixture.write("b.ts", strings.Repeat("b", 1000))
	fixture.run("a.ts", "b.ts")
	first := fixture.dataFile()

	fixture.write("c.ts", "export const c = 3;\n")
	if _, hits, misses, _ := fixture.run("a.ts", "b.ts", "c.ts"); hits != 2 || misses != 1 {
		t.Fatalf("served %d and read %d, want 2 and 1", hits, misses)
	}
	if fixture.dataFile() != first {
		t.Fatal("one new file rewrote the data file instead of appending to it")
	}
	if read, hits, _, _ := fixture.run("a.ts", "b.ts", "c.ts"); hits != 3 || read["c.ts"] != "export const c = 3;\n" {
		t.Fatalf("the appended file was not served: %d hits, %q", hits, read["c.ts"])
	}

	// Rewriting a and b three times over leaves their first copies dead, more than what is live.
	for round := range 3 {
		fixture.write("a.ts", strings.Repeat(string(rune('d'+round)), 1000))
		fixture.write("b.ts", strings.Repeat(string(rune('d'+round)), 1001))
		fixture.run("a.ts", "b.ts", "c.ts")
	}
	if fixture.dataFile() == first {
		t.Fatal("a pack mostly dead was never compacted")
	}
	if read, hits, _, _ := fixture.run("a.ts", "b.ts", "c.ts"); hits != 3 || read["a.ts"] != strings.Repeat("f", 1000) {
		t.Fatalf("after compacting, served %d with a.ts %q...", hits, read["a.ts"][:5])
	}
}

// A served file is a string over the pack's mapping, not a copy, and it stays whole after the pack that served
// it is dropped and collected (#kdee854, 1b). The compiler keeps source text and substrings of it for the whole
// run, so a view that died with its pack would hand the checker freed memory.
func TestAServedFileIsAViewOfTheMappingThatOutlivesThePack(t *testing.T) {
	t.Parallel()
	fixture := newPackFixture(t)
	want := strings.Repeat("export const a = 1;\n", 4096)
	fixture.write("a.ts", want)
	fixture.run("a.ts")

	pack, err := OpenContentPack(fixture.cache)
	if err != nil {
		t.Fatal(err)
	}
	served, ok := pack.wrap(osvfs.FS()).ReadFile(tspath.RootedFilePath(fixture.path("a.ts")))
	if !ok {
		t.Fatal("the pack served nothing, so nothing below is about a served file")
	}
	start := uintptr(unsafe.Pointer(unsafe.StringData(served)))
	mapped := uintptr(unsafe.Pointer(&pack.mapped[0]))
	if start < mapped || start+uintptr(len(served)) > mapped+uintptr(len(pack.mapped)) {
		t.Errorf("the served file lies outside the pack's mapping, so it was copied")
	}

	pack = nil
	runtime.GC()
	runtime.GC()
	if served != want {
		t.Fatalf("a served file read after its pack was collected holds %d bytes that differ from the file's", len(served))
	}
}
