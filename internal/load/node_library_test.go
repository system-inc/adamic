package load

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
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

// Not parallel: checking from a foreign working directory is the contract.
func TestNodeLibraryIgnoresCallerPackage(t *testing.T) {
	directory := t.TempDir()
	archive := filepath.Join("testdata", "node_types", "node-24.0.0.tgz")
	files := publishedNodeTypeFiles(t, archive, "sha512-yZQa2zm87aRVcqDyH5+4Hv9KYgSdgwX1rFnGvpbzMaC7YAljmhBET93TPiTd3ObwTL+gSpIzPKg5BqVxdCvxKg==")
	for name, data := range files {
		path := filepath.Join(directory, "node_modules", "@types", "node", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(directory)
	program, err := nodeSource(t, `import {existsSync} from 'node:fs'; const exists:boolean=existsSync('missing'); console.log('pinned');`)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, file := range program.compiler.GetSourceFiles() {
		if callerNodeTypes(file.FileName().AsString()) {
			t.Fatalf("loaded caller Node declarations: %s", file.FileName())
		}
		found = found || IsNodeLibrary(file)
	}
	if !found {
		t.Fatal("embedded Node declarations were not loaded")
	}
}

// Not parallel: the binary must load Node imports with no checkout or npm nearby.
func TestNodeLibraryOutsideCheckout(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := nodeSource(t, `import {join} from 'node:path'; const path:string=join('a','b'); console.log(path);`); err != nil {
		t.Fatal(err)
	}
}

func TestNodeLibraryDamagedBundleError(t *testing.T) {
	for _, files := range []fstest.MapFS{
		{},
		{"node_types/node_modules/@types/node/package.json": &fstest.MapFile{Data: []byte(`{"name":"@types/node","version":"24.0.0"}`)}},
		{"node_types/node_modules/@types/node/package.json": &fstest.MapFile{Data: []byte(`{"name":"@types/node","version":"25.3.3"}`)}},
	} {
		_, err := nodeTypesIndexFromFS(files)
		if err == nil || !strings.Contains(err.Error(), "reinstall Adamic") || strings.Contains(err.Error(), "stage3/") || strings.Contains(err.Error(), "node_types/") {
			t.Fatalf("want a user-facing bundled declaration error, got %v", err)
		}
	}
}

func TestNodeLibraryKeepsOfficialConsoleSignatures(t *testing.T) {
	t.Parallel()
	if _, err := nodeSource(t, `import type {Stats} from 'node:fs'; console.log(true); console.log(undefined); console.log('a','b');`); err != nil {
		t.Fatal(err)
	}
}
