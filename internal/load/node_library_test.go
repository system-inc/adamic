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
	if writeError := os.WriteFile(file, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	return Load([]string{file})
}

func TestNodeLibraryUsesPinnedDeclarations(test *testing.T) {
	test.Parallel()
	program, nodeSourceError := nodeSource(
		test,
		`import {readFileSync,statSync} from 'node:fs'; import type {Stats} from 'node:fs'; const bytes:Buffer=readFileSync('x'); const text:string=readFileSync('x','utf8'); const stat:Stats|undefined=statSync('x',{throwIfNoEntry:false}); console.log('checked');`,
	)
	if nodeSourceError != nil {
		test.Fatal(nodeSourceError)
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
	_, nodeSourceError = nodeSource(test, `import {existsSync} from 'node:fs'; const wrong:number=existsSync('x');`)
	if nodeSourceError == nil || !strings.Contains(nodeSourceError.Error(), "TS2322") {
		test.Fatalf("want genuine boolean signature rejection, got %v", nodeSourceError)
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
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0700); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(path, data, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	test.Chdir(directory)
	program, nodeSourceError := nodeSource(
		test,
		`import {existsSync} from 'node:fs'; const exists:boolean=existsSync('missing'); console.log('pinned');`,
	)
	if nodeSourceError != nil {
		test.Fatal(nodeSourceError)
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
	if _, nodeSourceError := nodeSource(
		test,
		`import {join} from 'node:path'; const path:string=join('a','b'); console.log(path);`,
	); nodeSourceError != nil {
		test.Fatal(nodeSourceError)
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
		_, nodeTypesIndexFromFileSystemError := nodeTypesIndexFromFileSystem(files)
		if nodeTypesIndexFromFileSystemError == nil ||
			!strings.Contains(nodeTypesIndexFromFileSystemError.Error(), "reinstall Adamic") ||
			strings.Contains(nodeTypesIndexFromFileSystemError.Error(), "stage3/") ||
			strings.Contains(nodeTypesIndexFromFileSystemError.Error(), "node_types/") {
			test.Fatalf("want a user-facing bundled declaration error, got %v", nodeTypesIndexFromFileSystemError)
		}
	}
}

func TestNodeLibraryKeepsOfficialConsoleSignatures(test *testing.T) {
	test.Parallel()
	if _, nodeSourceError := nodeSource(
		test,
		`import type {Stats} from 'node:fs'; console.log(true); console.log(undefined); console.log('a','b');`,
	); nodeSourceError != nil {
		test.Fatal(nodeSourceError)
	}
}

func TestNodeLibraryLazyRequire(test *testing.T) {
	test.Parallel()
	program, loadError := Load([]string{"testdata/node_activation/require.a"})
	if loadError != nil {
		test.Fatal(loadError)
	}
	if !program.fs.nodeTypes {
		test.Fatal("builtin require did not rebuild with the Node root")
	}
}

func TestNodeLibraryImportTypes(test *testing.T) {
	test.Parallel()
	// Exercise type-only activation independently of require, then the combined program.
	source := `type Inspector = typeof import("inspector"); type Stats = import("fs").Stats; type Prefixed = import("node:fs").Stats;`
	program, nodeSourceError := nodeSource(test, source)
	if nodeSourceError != nil {
		test.Fatal(nodeSourceError)
	}
	if !program.fs.nodeTypes {
		test.Fatal("import types did not rebuild with the Node root")
	}
	if _, loadError := Load(
		[]string{"testdata/node_activation/require.a", "testdata/node_activation/import_type.a"},
	); loadError != nil {
		test.Fatal(loadError)
	}
}

func TestNodeLibraryActivationStaysNarrow(test *testing.T) {
	test.Parallel()
	program, loadError := Load([]string{"testdata/node_activation/plain.a"})
	if loadError != nil {
		test.Fatal(loadError)
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
	program, nodeSourceError := nodeSource(
		test,
		`declare function require(name: string): unknown; const user = require("some-user-package"); const relative = require("./fs"); const unknownNode = require("node:not-a-builtin"); const thirdParty = require("undici-types"); const dynamic = require("f" + "s"); const host = {require}; host.require("fs");`,
	)
	if nodeSourceError != nil {
		test.Fatal(nodeSourceError)
	}
	assertNoNode(program)
}

func TestNodeLibraryBuiltinSpecifiers(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"fs", "node:fs", "path", "node:path", "inspector", "fs/promises", "node:test"} {
		program, nodeSourceError := nodeSource(test, `function lazy() { return require("`+name+`"); }`)
		if nodeSourceError != nil {
			test.Fatalf("%s: %v", name, nodeSourceError)
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
	timeoutContext, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(
		timeoutContext,
		binary,
		"-test.run=^TestNodeLibraryLazyRequire$",
		"-test.count=1",
		"-test.timeout=30s",
		"-test.v",
	)
	output, combinedOutputError := command.CombinedOutput()
	if timeoutContext.Err() != nil ||
		combinedOutputError == nil ||
		!strings.Contains(string(output), "TS2591") ||
		!strings.Contains(string(output), "FAIL: TestNodeLibraryLazyRequire") {
		test.Fatalf("require mutant must fail the real checker with TS2591: %v\n%s", combinedOutputError, output)
	}
}

func TestNodePreludeConsoleSignatures(test *testing.T) {
	test.Parallel()
	for _, nullable := range []bool{false, true} {
		source := prelude
		if nullable {
			source = strings.ReplaceAll(source, "message: string", "message: string | null | undefined")
		}
		adapted, nodePreludeError := nodePrelude(source)
		if nodePreludeError != nil {
			test.Fatal(nodePreludeError)
		}
		if !strings.Contains(adapted, "declare var console: Console;") || strings.Contains(adapted, "log(message:") {
			test.Fatalf("console was not adapted: %s", adapted)
		}
		// Standalone main has no Console/process overlay. Exercise both declarations
		// with nested parameter types and trivia, independent of the method signatures.
		source += "\n// overlay\ninterface Console { log(message: { value: string | null }): void; error(message: undefined): void; }\ndeclare const process: { nested: { code: number }; };\n"
		adapted, nodePreludeError = nodePrelude(source, "Console", "process")
		if nodePreludeError != nil {
			test.Fatal(nodePreludeError)
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
			adapted, nodePreludeError := nodePrelude(source, "Console", "process")
			if nodePreludeError == nil ||
				!strings.Contains(nodePreludeError.Error(), "missing the "+name+" declaration") ||
				adapted != "" {
				test.Fatalf("want clear missing %s error and no prelude, got %q, %v", name, adapted, nodePreludeError)
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
		if adapted, nodePreludeError := nodePrelude(source); nodePreludeError == nil || adapted != "" {
			test.Fatalf("invalid prelude passed through: %q, %v", adapted, nodePreludeError)
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
	if _, nodeSourceError := nodeSource(
		test,
		`import 'node:process'; console.log(true); console.log(undefined); console.error('a','b');`,
	); nodeSourceError != nil {
		test.Fatal(nodeSourceError)
	}
}
