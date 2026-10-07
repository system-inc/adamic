package load

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// An apple/ module written in Adamic is found wherever the resolver looks for it, and every place
// it's found is one real path, so two directories importing it import one module.
func TestAppleImplementationsAreOneModuleWhereverTheyreFound(t *testing.T) {
	t.Parallel()
	fs := &appleFS{FS: osvfs.FS()}
	near := "/work/app/node_modules/apple/swiftui/state.ts"
	far := "/node_modules/apple/swiftui/state.ts"
	for _, path := range []string{near, far} {
		if !fs.FileExists(path) {
			t.Fatalf("%s isn't found", path)
		}
		if source, found := fs.ReadFile(path); !found || source == "" {
			t.Fatalf("%s has no source", path)
		}
	}
	if fs.Realpath(near) != fs.Realpath(far) {
		t.Fatalf("one module, two real paths: %s and %s", fs.Realpath(near), fs.Realpath(far))
	}
	if fs.FileExists("/work/app/node_modules/apple/swiftui/missing.ts") {
		t.Fatal("a module with no embedded source is found")
	}
	if fs.FileExists("/work/app/node_modules/apple/appkit/window.ts") {
		t.Fatal("a binding file (declarations) is served as an implementation")
	}
}

// A program imports State from two directories and gets the one class: it loads, and the checker
// sees one declaration of State.
func TestAProgramImportsAnAppleImplementation(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if err := os.MkdirAll(filepath.Join(directory, "deeper"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"deeper/other.a": "import { State } from 'apple/swiftui/state';\nexport const shared = new State<string>('far');\n",
		"main.a":         "import { State } from 'apple/swiftui/state';\nimport { shared } from './deeper/other.a';\nconst local: State<string> = shared;\nlocal.set('near');\nconsole.log(local.value);\n",
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Load([]string{filepath.Join(directory, "main.a")})
	if err != nil {
		t.Fatal(err)
	}
	states := 0
	for _, sourceFile := range program.compiler.GetSourceFiles() {
		if filepath.Base(sourceFile.FileName()) == "state.ts" {
			states++
		}
	}
	if states != 1 {
		t.Fatalf("State's module was loaded %d times", states)
	}
}
