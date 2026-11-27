package load

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// The prelude is exactly three files, each a real one beside this, each served at its path in the prelude
// directory, and IsPrelude, which answers by that directory, says yes to the two a program holds and to nothing
// of its own. A fourth prelude file nobody names here fails it, whether it's put beside these or served.
func TestThePreludeIsThreeFiles(t *testing.T) {
	t.Parallel()
	named := map[string]tspath.RootedFilePath{
		"prelude.d.ts":         preludePath,
		"prelude_console.d.ts": preludeConsolePath,
		"prelude_node.d.ts":    preludeNodePath,
	}

	// On disk: every prelude file beside load's source is one of the three.
	onDisk, err := filepath.Glob("prelude*.d.ts")
	if err != nil {
		t.Fatal(err)
	}
	want := make([]string, 0, len(named))
	for file := range named {
		want = append(want, file)
	}
	slices.Sort(want)
	slices.Sort(onDisk)
	if !slices.Equal(onDisk, want) {
		t.Fatalf("the prelude files beside load are %q, want exactly %q", onDisk, want)
	}

	// Served: each file's bytes at its path, and nothing at a path in the directory nobody names.
	view := &sourceFS{FS: osvfs.FS()}
	for file, path := range named {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source, served := view.ReadFile(path)
		if !served || source != string(data) || !view.FileExists(path) {
			t.Errorf("%s isn't served as %s, byte for byte", path, file)
		}
	}
	unnamed := tspath.RootedFilePath(preludeDirectory + "/other.d.ts")
	if _, served := view.ReadFile(unnamed); served || view.FileExists(unnamed) {
		t.Errorf("the prelude directory serves %s, which nobody names", unnamed)
	}

	// In a program: the prelude and the standalone console, or the prelude and Node's, and never a program's own
	// file, even one in a directory of the prelude directory's name.
	directory := filepath.Join(t.TempDir(), "adamic-prelude")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		source string
		want   []string
	}{
		{"console.log('standalone');\n", []string{string(preludePath), string(preludeConsolePath)}},
		{"import 'node:process';\nconsole.log('node');\n", []string{string(preludePath), string(preludeNodePath)}},
	} {
		entry := filepath.Join(directory, "main.a")
		if err := os.WriteFile(entry, []byte(row.source), 0o644); err != nil {
			t.Fatal(err)
		}
		program, err := Load([]string{entry})
		if err != nil {
			t.Fatal(err)
		}
		held := []string{}
		for _, file := range program.compiler.GetSourceFiles() {
			if IsPrelude(file) {
				held = append(held, file.FileName().AsString())
			}
		}
		slices.Sort(held)
		if !slices.Equal(held, row.want) {
			t.Errorf("%q holds the prelude files %q, want %q", row.source, held, row.want)
		}
		for _, own := range program.Files() {
			if IsPrelude(own) {
				t.Errorf("IsPrelude says yes to the program's own %s", own.FileName())
			}
		}
	}
}
