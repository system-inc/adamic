package rule

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// directoryProgram is a Program that answers only what FingerprintPath reads.
type directoryProgram struct {
	Program
	directory     string
	caseSensitive bool
}

func (program directoryProgram) GetCurrentDirectory() string     { return program.directory }
func (program directoryProgram) UseCaseSensitiveFileNames() bool { return program.caseSensitive }

// tspathName is FingerprintPath's answer before it learned to slice: tspath's general relative path.
func tspathName(program directoryProgram, path string) string {
	return tspath.GetRelativePathFromDirectory(program.directory, path, caseSensitivity(program.caseSensitive))
}

// FingerprintPath slices a descendant's name rather than having tspath split and join it (#9prvp67), and a
// fingerprint is only sound if the name is the same string either way: a different spelling of one file
// would move every fingerprint and replay nothing, or worse, collide. So the slice is held to tspath's
// answer over every file and directory of this repository, and over the shapes a prefix test gets wrong.
func TestFingerprintPathSlicesWhatTspathWouldJoin(t *testing.T) {
	t.Parallel()

	shapes := []struct {
		name      string
		directory string
		path      string
	}{
		{"a file under the directory", "/repository", "/repository/source/Paths.ts"},
		{"a directory under it", "/repository", "/repository/source/translations"},
		{"a sibling whose name extends the directory's", "/repository", "/repository-two/source/Paths.ts"},
		{"the directory itself", "/repository", "/repository"},
		{"a parent of the directory", "/repository/app", "/repository/Shared.ts"},
		{"a cousin", "/repository/app", "/repository/libraries/structure/Index.ts"},
		{"the filesystem root as the directory", "/", "/repository/Paths.ts"},
		{"a directory spelled with a trailing separator", "/repository/", "/repository/source/Paths.ts"},
		{"a mixed-case directory over a canonical, lowercased path", "/Work/Repository", "/work/repository/ahra/Paths.ts"},
		{"a lowercased sibling that extends the directory's name", "/Work/Repo", "/work/repository/Paths.ts"},
		{"a non-ASCII directory", "/Work/Zo\u00eb", "/work/zo\u00eb/Paths.ts"},
	}
	for _, shape := range shapes {
		for _, caseSensitive := range []bool{true, false} {
			program := directoryProgram{directory: shape.directory, caseSensitive: caseSensitive}
			if got, want := FingerprintPath(program, shape.path), tspathName(program, shape.path); got != want {
				t.Errorf("%s (case-sensitive %v): FingerprintPath gave %q, tspath %q", shape.name, caseSensitive, got, want)
			}
		}
	}

	// Every file and directory of this module, named from its root, the way a fingerprint names a program's
	// files from the run's directory: as written on a case-sensitive host, and canonical (lowercased) under the
	// root's own spelling on a case-insensitive one, which is macOS's and where the slice first never matched.
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	directory := tspath.NormalizePath(filepath.ToSlash(root))
	sensitive := directoryProgram{directory: directory, caseSensitive: true}
	insensitive := directoryProgram{directory: directory, caseSensitive: false}
	compared := 0
	sliced := 0
	walkError := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "TypeScript" || entry.Name() == "node_modules") {
			return filepath.SkipDir
		}
		name := tspath.NormalizePath(filepath.ToSlash(path))
		canonical := string(tspath.CaseInsensitive.PathKey(tspath.RootedPath(name)))
		for _, probe := range []struct {
			program directoryProgram
			path    string
		}{{sensitive, name}, {insensitive, canonical}} {
			if got, want := FingerprintPath(probe.program, probe.path), tspathName(probe.program, probe.path); got != want {
				t.Errorf("%s (case-sensitive %v): FingerprintPath gave %q, tspath %q", probe.path, probe.program.caseSensitive, got, want)
			}
			if _, under := fingerprintPathUnder(probe.program.directory, probe.path, probe.program.caseSensitive); under {
				sliced++
			}
			compared++
		}
		return nil
	})
	if walkError != nil {
		t.Fatal(walkError)
	}
	// A walk that compared nothing would pass everything.
	if compared < 1000 {
		t.Fatalf("compared %d paths, which is not this repository", compared)
	}
	// And the slice has to be what answered, both ways, or this proves tspath equal to itself.
	if sliced < compared-4 {
		t.Fatalf("the slice answered %d of %d paths, so the fast path is not the one being held", sliced, compared)
	}
}
