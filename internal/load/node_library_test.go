package load

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// givenNode reports whether a program was built with Node's declarations: whether it holds the Node prelude, which Load
// roots in the standalone console's place exactly when it rebuilds with the Node root.
func givenNode(program *Program) bool {
	for _, file := range program.compiler.GetSourceFiles() {
		if file.FileName() == preludeNodePath {
			return true
		}
	}
	return false
}

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
	if !givenNode(program) {
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
	if !givenNode(program) {
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
		if givenNode(program) {
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
		if !givenNode(program) {
			t.Fatalf("%s did not activate Node", name)
		}
	}
}

func TestProduct_NodeLibraryRequireMutant(t *testing.T) {
	t.Parallel()
	nodeLibraryRequireMutant(t)
}

func nodeLibraryRequireMutant(t testing.TB) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "node-library-require.test", "./internal/load", nil, []buildcache.Mutation{{
		File:   "internal/load/node_library.go",
		Before: "case ast.KindCallExpression:",
		After:  "case ast.KindUnknown:",
	}})
}

func TestNodeLibraryRequireMutant(t *testing.T) {
	t.Parallel()
	// Run the declared product with its require branch disabled. The independent
	// checker fixture must fail with TS2591.
	binary := nodeLibraryRequireMutant(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "-test.run=^TestNodeLibraryLazyRequire$", "-test.count=1", "-test.timeout=30s", "-test.v")
	output, err := command.CombinedOutput()
	if ctx.Err() != nil || err == nil || !strings.Contains(string(output), "TS2591") || !strings.Contains(string(output), "FAIL: TestNodeLibraryLazyRequire") {
		t.Fatalf("require mutant must fail the real checker with TS2591: %v\n%s", err, output)
	}
}

// recordreplay's declaration (#b7phwab) rides only the Node prelude: a program with no Node types still can't name
// NodeJS (TS2503, the promise of stage3/drivers/parser/native-nodejs-process-type.a), and a Node program reads
// process.recordreplay typed unknown. Mutants: prelude_node.d.ts's namespace moved into prelude.d.ts, and the plain
// row loads; the namespace dropped from prelude_node.d.ts, and the Node row fails with TS2339.
func TestNodeProcessAugmentationRidesOnlyTheNodePrelude(t *testing.T) {
	t.Parallel()
	for name, text := range map[string]string{"prelude.d.ts": prelude, "prelude_console.d.ts": preludeConsole} {
		if strings.Contains(text, "namespace NodeJS") {
			t.Fatalf("%s declares NodeJS, so a program without Node types can name NodeJS.Process", name)
		}
	}
	if !strings.Contains(preludeNode, "namespace NodeJS") || !strings.Contains(preludeNode, "recordreplay?: unknown;") {
		t.Fatalf("the Node prelude doesn't carry recordreplay's declaration: %q", preludeNode)
	}
	_, err := nodeSource(t, "declare const process: NodeJS.Process | undefined;\nconsole.log(\"loaded\");\n")
	if err == nil || !strings.Contains(err.Error(), "TS2503") {
		t.Fatalf("a program without Node types named NodeJS.Process: want TS2503, got %v", err)
	}
	if _, err := nodeSource(t, "import 'node:process';\nconst value: unknown = process.recordreplay;\nconsole.log(typeof value);\n"); err != nil {
		t.Fatalf("a Node program can't read process.recordreplay: %v", err)
	}
}

// A Node program has Node's console and never the standalone one, whatever the standalone one's signatures: here they
// take string | null | undefined, and Node's console.log(true) and console.error('a', 'b') still check. Mutant: Load
// rooting the standalone console beside the Node one, and the two consoles collide (TS2451).
// Not parallel: replaces the package's standalone console, which loader calls read.
func TestNodeLibraryNullablePrelude(t *testing.T) {
	original := preludeConsole
	t.Cleanup(func() { preludeConsole = original })
	preludeConsole = strings.ReplaceAll(preludeConsole, "message: string", "message: string | null | undefined")
	if preludeConsole == original {
		t.Fatal("the standalone console has no message: string signature to widen")
	}
	if _, err := nodeSource(t, `import 'node:process'; console.log(true); console.log(undefined); console.error('a','b');`); err != nil {
		t.Fatal(err)
	}
}
