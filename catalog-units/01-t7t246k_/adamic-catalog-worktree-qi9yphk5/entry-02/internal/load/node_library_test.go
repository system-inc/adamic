package load

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func nodeSource(t *testing.T, source string) (*Program, error) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return Load([]string{file})
}

func TestNodeLibraryUsesPinnedDeclarations(t *testing.T) {
	t.Parallel()
	program, err := nodeSource(t, `import {readFileSync,statSync} from 'node:fs'; import type {Stats} from 'node:fs'; const bytes:Buffer=readFileSync('x'); const text:string=readFileSync('x','utf8'); const stat:Stats|undefined=statSync('x',{throwIfNoEntry:false}); console.log('checked');`)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, file := range program.compiler.GetSourceFiles() {
		if IsNodeLibrary(file) {
			found = true
		}
	}
	if !found {
		t.Fatal("no pinned Node declaration source was loaded")
	}
	_, err = nodeSource(t, `import {existsSync} from 'node:fs'; const wrong:number=existsSync('x');`)
	if err == nil || !strings.Contains(err.Error(), "TS2322") {
		t.Fatalf("want genuine boolean signature rejection, got %v", err)
	}
}

func TestNodeLibraryRejectsDifferentVersion(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	root := filepath.Join(dir, filepath.FromSlash(nodeTypesRelative))
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name":"@types/node","version":"24.0.0"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.d.ts"), []byte("// wrong-version package fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := nodeTypesIndex(dir)
	if err == nil || !strings.Contains(err.Error(), "25.3.3") {
		t.Fatalf("want pinned version refusal, got %v", err)
	}
}

func TestNodeLibraryKeepsOfficialConsoleSignatures(t *testing.T) {
	t.Parallel()
	if _, err := nodeSource(t, `import type {Stats} from 'node:fs'; console.log(true); console.log(undefined); console.log('a','b');`); err != nil {
		t.Fatal(err)
	}
}
