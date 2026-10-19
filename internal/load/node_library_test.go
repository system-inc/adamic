package load

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
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
	t.Parallel()
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

func TestNodeLibraryLazyRequire(t *testing.T) {
	t.Parallel()
	program, err := Load([]string{"testdata/node_activation/require.a"})
	if err != nil {
		t.Fatal(err)
	}
	if !program.fs.nodeTypes {
		t.Fatal("builtin require did not rebuild with the Node root")
	}
}

func TestNodeLibraryImportTypes(t *testing.T) {
	t.Parallel()
	// Exercise type-only activation independently of require, then the combined program.
	source := `type Inspector = typeof import("inspector"); type Stats = import("fs").Stats; type Prefixed = import("node:fs").Stats;`
	program, err := nodeSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	if !program.fs.nodeTypes {
		t.Fatal("import types did not rebuild with the Node root")
	}
	if _, err := Load([]string{"testdata/node_activation/require.a", "testdata/node_activation/import_type.a"}); err != nil {
		t.Fatal(err)
	}
}

func TestNodeLibraryActivationStaysNarrow(t *testing.T) {
	t.Parallel()
	program, err := Load([]string{"testdata/node_activation/plain.a"})
	if err != nil {
		t.Fatal(err)
	}
	assertNoNode := func(program *Program) {
		t.Helper()
		if program.fs.nodeTypes {
			t.Fatal("rebuilt a non-Node program with the Node root")
		}
		for _, file := range program.compiler.GetSourceFiles() {
			if IsNodeLibrary(file) {
				t.Fatalf("non-Node program loaded %s", file.FileName())
			}
		}
	}
	assertNoNode(program)
	program, err = nodeSource(t, `declare function require(name: string): unknown; const user = require("some-user-package"); const relative = require("./fs"); const unknownNode = require("node:not-a-builtin"); const thirdParty = require("undici-types"); const dynamic = require("f" + "s"); const host = {require}; host.require("fs");`)
	if err != nil {
		t.Fatal(err)
	}
	assertNoNode(program)
}

func TestNodeLibraryBuiltinSpecifiers(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"fs", "node:fs", "path", "node:path", "inspector", "fs/promises", "node:test"} {
		program, err := nodeSource(t, `function lazy() { return require("`+name+`"); }`)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !program.fs.nodeTypes {
			t.Fatalf("%s did not activate Node", name)
		}
	}
}

func TestNodeLibraryRequireMutant(t *testing.T) {
	t.Parallel()
	// Compile the actual loader with its require branch disabled. The independent
	// checker fixture above must fail, rather than merely disagreeing with a predicate.
	original, err := os.ReadFile("node_library.go")
	if err != nil {
		t.Fatal(err)
	}
	mutant := strings.Replace(string(original), "case ast.KindCallExpression:", "case ast.KindUnknown:", 1)
	if mutant == string(original) {
		t.Fatal("require mutation no longer matches the hook")
	}
	directory := t.TempDir()
	replacement := filepath.Join(directory, "node_library.go")
	if err := os.WriteFile(replacement, []byte(mutant), 0600); err != nil {
		t.Fatal(err)
	}
	originalPath, err := filepath.Abs("node_library.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{originalPath: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-overlay", overlayPath, "-run", "^TestNodeLibraryLazyRequire$", "-count=1", "-timeout=30s", ".")
	output, err := command.CombinedOutput()
	if ctx.Err() != nil || err == nil || !strings.Contains(string(output), "TS2591") || !strings.Contains(string(output), "FAIL: TestNodeLibraryLazyRequire") {
		t.Fatalf("require mutant must fail the real checker with TS2591: %v\n%s", err, output)
	}
}
