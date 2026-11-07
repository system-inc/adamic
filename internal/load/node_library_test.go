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

func nodeSource(test *testing.T, source string) (*Program, error) {
	test.Helper()
	file := filepath.Join(test.TempDir(), "main.a")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		test.Fatal(err)
	}
	return Load([]string{file})
}

func TestNodeLibraryUsesPinnedDeclarations(test *testing.T) {
	test.Parallel()
	program, err := nodeSource(
		test,
		`import {readFileSync,statSync} from 'node:fs'; import type {Stats} from 'node:fs'; const bytes:Buffer=readFileSync('x'); const text:string=readFileSync('x','utf8'); const stat:Stats|undefined=statSync('x',{throwIfNoEntry:false}); console.log('checked');`,
	)
	if err != nil {
		test.Fatal(err)
	}
	found := false
	for _, file := range program.compiler.GetSourceFiles() {
		if IsNodeLibrary(file) {
			found = true
		}
	}
	if !found {
		test.Fatal("no pinned Node declaration source was loaded")
	}
	_, err = nodeSource(test, `import {existsSync} from 'node:fs'; const wrong:number=existsSync('x');`)
	if err == nil || !strings.Contains(err.Error(), "TS2322") {
		test.Fatalf("want genuine boolean signature rejection, got %v", err)
	}
}

// Not parallel: checking from a foreign working directory is the contract.
func TestNodeLibraryIgnoresCallerPackage(test *testing.T) {
	directory := test.TempDir()
	archive := filepath.Join("testdata", "node_types", "node-24.0.0.tgz")
	files := publishedNodeTypeFiles(
		test,
		archive,
		"sha512-yZQa2zm87aRVcqDyH5+4Hv9KYgSdgwX1rFnGvpbzMaC7YAljmhBET93TPiTd3ObwTL+gSpIzPKg5BqVxdCvxKg==",
	)
	for name, data := range files {
		path := filepath.Join(directory, "node_modules", "@types", "node", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			test.Fatal(err)
		}
	}
	test.Chdir(directory)
	program, err := nodeSource(
		test,
		`import {existsSync} from 'node:fs'; const exists:boolean=existsSync('missing'); console.log('pinned');`,
	)
	if err != nil {
		test.Fatal(err)
	}
	found := false
	for _, file := range program.compiler.GetSourceFiles() {
		if callerNodeTypes(file.FileName().AsString()) {
			test.Fatalf("loaded caller Node declarations: %s", file.FileName())
		}
		found = found || IsNodeLibrary(file)
	}
	if !found {
		test.Fatal("embedded Node declarations were not loaded")
	}
}

// Not parallel: the binary must load Node imports with no checkout or npm nearby.
func TestNodeLibraryOutsideCheckout(test *testing.T) {
	test.Chdir(test.TempDir())
	if _, err := nodeSource(
		test,
		`import {join} from 'node:path'; const path:string=join('a','b'); console.log(path);`,
	); err != nil {
		test.Fatal(err)
	}
}

func TestNodeLibraryDamagedBundleError(test *testing.T) {
	test.Parallel()
	for _, files := range []fstest.MapFS{
		{},
		{"node_types/node_modules/@types/node/package.json": &fstest.MapFile{
			Data: []byte(`{"name":"@types/node","version":"24.0.0"}`),
		}},
		{"node_types/node_modules/@types/node/package.json": &fstest.MapFile{
			Data: []byte(`{"name":"@types/node","version":"25.3.3"}`),
		}},
	} {
		_, err := nodeTypesIndexFromFS(files)
		if err == nil ||
			!strings.Contains(err.Error(), "reinstall Adamic") ||
			strings.Contains(err.Error(), "stage3/") ||
			strings.Contains(err.Error(), "node_types/") {
			test.Fatalf("want a user-facing bundled declaration error, got %v", err)
		}
	}
}

func TestNodeLibraryKeepsOfficialConsoleSignatures(test *testing.T) {
	test.Parallel()
	if _, err := nodeSource(
		test,
		`import type {Stats} from 'node:fs'; console.log(true); console.log(undefined); console.log('a','b');`,
	); err != nil {
		test.Fatal(err)
	}
}

func TestNodeLibraryLazyRequire(test *testing.T) {
	test.Parallel()
	program, err := Load([]string{"testdata/node_activation/require.a"})
	if err != nil {
		test.Fatal(err)
	}
	if !program.fs.nodeTypes {
		test.Fatal("builtin require did not rebuild with the Node root")
	}
}

func TestNodeLibraryImportTypes(test *testing.T) {
	test.Parallel()
	// Exercise type-only activation independently of require, then the combined program.
	source := `type Inspector = typeof import("inspector"); type Stats = import("fs").Stats; type Prefixed = import("node:fs").Stats;`
	program, err := nodeSource(test, source)
	if err != nil {
		test.Fatal(err)
	}
	if !program.fs.nodeTypes {
		test.Fatal("import types did not rebuild with the Node root")
	}
	if _, err := Load(
		[]string{"testdata/node_activation/require.a", "testdata/node_activation/import_type.a"},
	); err != nil {
		test.Fatal(err)
	}
}

func TestNodeLibraryActivationStaysNarrow(test *testing.T) {
	test.Parallel()
	program, err := Load([]string{"testdata/node_activation/plain.a"})
	if err != nil {
		test.Fatal(err)
	}
	assertNoNode := func(program *Program) {
		test.Helper()
		if program.fs.nodeTypes {
			test.Fatal("rebuilt a non-Node program with the Node root")
		}
		for _, file := range program.compiler.GetSourceFiles() {
			if IsNodeLibrary(file) {
				test.Fatalf("non-Node program loaded %s", file.FileName())
			}
		}
	}
	assertNoNode(program)
	program, err = nodeSource(
		test,
		`declare function require(name: string): unknown; const user = require("some-user-package"); const relative = require("./fs"); const unknownNode = require("node:not-a-builtin"); const thirdParty = require("undici-types"); const dynamic = require("f" + "s"); const host = {require}; host.require("fs");`,
	)
	if err != nil {
		test.Fatal(err)
	}
	assertNoNode(program)
}

func TestNodeLibraryBuiltinSpecifiers(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"fs", "node:fs", "path", "node:path", "inspector", "fs/promises", "node:test"} {
		program, err := nodeSource(test, `function lazy() { return require("`+name+`"); }`)
		if err != nil {
			test.Fatalf("%s: %v", name, err)
		}
		if !program.fs.nodeTypes {
			test.Fatalf("%s did not activate Node", name)
		}
	}
}

func TestProduct_NodeLibraryRequireMutant(test *testing.T) {
	test.Parallel()
	nodeLibraryRequireMutant(test)
}

func nodeLibraryRequireMutant(test testing.TB) string {
	test.Helper()
	return buildcache.GoTestMutated(
		test,
		"",
		"node-library-require.test",
		"./internal/load",
		nil,
		[]buildcache.Mutation{{
			File:   "internal/load/node_library.go",
			Before: "case ast.KindCallExpression:",
			After:  "case ast.KindUnknown:",
		}},
	)
}

func TestNodeLibraryRequireMutant(test *testing.T) {
	test.Parallel()
	// Run the declared product with its require branch disabled. The independent
	// checker fixture must fail with TS2591.
	binary := nodeLibraryRequireMutant(test)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(
		ctx,
		binary,
		"-test.run=^TestNodeLibraryLazyRequire$",
		"-test.count=1",
		"-test.timeout=30s",
		"-test.v",
	)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil ||
		err == nil ||
		!strings.Contains(string(output), "TS2591") ||
		!strings.Contains(string(output), "FAIL: TestNodeLibraryLazyRequire") {
		test.Fatalf("require mutant must fail the real checker with TS2591: %v\n%s", err, output)
	}
}

func TestNodePreludeConsoleSignatures(test *testing.T) {
	test.Parallel()
	for _, nullable := range []bool{false, true} {
		source := prelude
		if nullable {
			source = strings.ReplaceAll(source, "message: string", "message: string | null | undefined")
		}
		adapted, err := nodePrelude(source)
		if err != nil {
			test.Fatal(err)
		}
		if !strings.Contains(adapted, "declare var console: Console;") || strings.Contains(adapted, "log(message:") {
			test.Fatalf("console was not adapted: %s", adapted)
		}
		// Standalone main has no Console/process overlay. Exercise both declarations
		// with nested parameter types and trivia, independent of the method signatures.
		source += "\n// overlay\ninterface Console { log(message: { value: string | null }): void; error(message: undefined): void; }\ndeclare const process: { nested: { code: number }; };\n"
		adapted, err = nodePrelude(source, "Console", "process")
		if err != nil {
			test.Fatal(err)
		}
		if strings.Contains(adapted, "interface Console") ||
			!strings.Contains(adapted, "declare var process: NodeJS.Process;") ||
			!strings.Contains(adapted, "// overlay") {
			test.Fatalf("overlay was not adapted: %s", adapted)
		}
	}
}

func TestNodePreludeMissingDeclarations(test *testing.T) {
	test.Parallel()
	declarations := []string{
		"declare const console: {log(value: unknown): void;};",
		"interface Console {error(value: unknown): void;}",
		"declare const process: {code: number;};",
	}
	names := []string{"console", "Console", "process"}
	for index, name := range names {
		test.Run(name, func(test *testing.T) {
			source := ""
			for other, declaration := range declarations {
				if other != index {
					source += declaration + "\n"
				}
			}
			adapted, err := nodePrelude(source, "Console", "process")
			if err == nil || !strings.Contains(err.Error(), "missing the "+name+" declaration") || adapted != "" {
				test.Fatalf("want clear missing %s error and no prelude, got %q, %v", name, adapted, err)
			}
		})
	}
}

func TestNodePreludeRejectsMalformedAndDuplicateDeclarations(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"declare const console: {",
		"declare const console: {}; declare var console: {};",
		"declare const console: {}, other: {};",
	} {
		if adapted, err := nodePrelude(source); err == nil || adapted != "" {
			test.Fatalf("invalid prelude passed through: %q, %v", adapted, err)
		}
	}
}

// Not parallel: this simulates the independently queued prelude change in the
// embedded source without changing its declarations on this branch.
// Not parallel: replaces the package prelude used by loader calls.
func TestNodeLibraryNullablePrelude(test *testing.T) {
	original := prelude
	test.Cleanup(func() { prelude = original })
	prelude = strings.ReplaceAll(prelude, "message: string", "message: string | null | undefined")
	if _, err := nodeSource(
		test,
		`import 'node:process'; console.log(true); console.log(undefined); console.error('a','b');`,
	); err != nil {
		test.Fatal(err)
	}
}
