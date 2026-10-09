package compiler

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/outputpaths"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

// sourceExtensionsProgram builds the program /src/tsconfig.json defines over files.
func sourceExtensionsProgram(t *testing.T, files map[string]any) (*Program, *tsoptions.ParsedCommandLine, []string) {
	t.Helper()
	fs := vfstest.FromMap(files, tspath.CaseSensitive)
	host := NewCompilerHost(fs, "", nil, nil, nil)
	config, configDiagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePathFromNormalized("/src/tsconfig.json"), nil, nil, fs, nil)
	messages := []string{}
	for _, diagnostic := range append(configDiagnostics, config.GetConfigFileParsingDiagnostics()...) {
		messages = append(messages, diagnostic.String())
	}
	program := NewProgram(ProgramOptions{Config: config, Host: host})
	return program, config, messages
}

// fileNames are a program's own files: no lib.
func fileNames(program *Program) []string {
	names := []string{}
	for _, file := range program.GetSourceFiles() {
		if strings.HasPrefix(file.FileName().AsString(), "/src/") {
			names = append(names, file.FileName().AsString())
		}
	}
	slices.Sort(names)
	return names
}

// A config's "sourceExtensions" makes files with that extension TypeScript source in its program, under their
// own names: enumerated by include, parsed as TypeScript, and reached by an import that names the extension,
// which resolves to the file itself, never to an X.a.ts. A config without the key never reads such a file,
// so a static library named libfoo.a beside the source is not parsed as TypeScript. The key is inherited
// through extends, and an invalid entry is reported and left out.
func TestSourceExtensions(t *testing.T) {
	t.Parallel()
	source := map[string]any{
		"/src/geometry.a": "export function area(width: number, height: number): number { return width * height; }\n",
		"/src/main.a":     "import { area } from './geometry.a';\nexport const total: number = area(2, 3);\nexport const wrong: string = area(1, 1);\n",
		"/src/plain.ts":   "export const plain = 1;\n",
		"/src/libfoo.a":   "!<arch>\nnot TypeScript at all {{{\n",
	}
	with := func(config string) map[string]any {
		files := map[string]any{"/src/tsconfig.json": config}
		for name, contents := range source {
			files[name] = contents
		}
		return files
	}

	t.Run("opted in", func(t *testing.T) {
		t.Parallel()
		files := with(`{"sourceExtensions": [".a"], "compilerOptions": {"noLib": true, "strict": true, "noEmit": true, "allowImportingTsExtensions": true, "module": "esnext", "moduleResolution": "bundler"}, "include": ["*.a", "*.ts"], "exclude": ["libfoo.a"]}`)
		program, config, configMessages := sourceExtensionsProgram(t, files)
		assert.Equal(t, len(configMessages), 0, configMessages)
		assert.DeepEqual(t, config.SourceExtensions(), []string{".a"})
		assert.DeepEqual(t, fileNames(program), []string{"/src/geometry.a", "/src/main.a", "/src/plain.ts"})

		main := program.GetSourceFile(tspath.RootedFilePathFromNormalized("/src/main.a"))
		assert.Assert(t, main != nil)
		assert.Equal(t, main.ScriptKind, core.ScriptKindTS)
		// The one error is the planted one, in main.a under its own name: the import resolved, so area's
		// return type is known, and nothing about geometry.a.ts or an unresolved module appears.
		diagnostics := program.GetSemanticDiagnostics(context.Background(), main)
		assert.Equal(t, len(diagnostics), 1)
		assert.Equal(t, diagnostics[0].Code(), int32(2322))
		assert.Equal(t, diagnostics[0].File().FileName().AsString(), "/src/main.a")
		resolvedFileName := ""
		for key, resolved := range program.GetResolvedModules()[main.PathKey()] {
			if key.Name == "./geometry.a" && resolved.IsResolved() {
				resolvedFileName = resolved.ResolvedFileName.AsString()
			}
		}
		assert.Equal(t, resolvedFileName, "/src/geometry.a")
	})

	t.Run("not opted in", func(t *testing.T) {
		t.Parallel()
		files := with(`{"compilerOptions": {"noLib": true, "noEmit": true}, "include": ["**/*"]}`)
		program, _, _ := sourceExtensionsProgram(t, files)
		assert.DeepEqual(t, fileNames(program), []string{"/src/plain.ts"})
	})

	t.Run("inherited through extends", func(t *testing.T) {
		t.Parallel()
		files := with(`{"extends": "./base.json", "compilerOptions": {"noLib": true, "noEmit": true}, "include": ["main.a", "geometry.a"]}`)
		files["/src/base.json"] = `{"sourceExtensions": [".a"]}`
		_, config, configMessages := sourceExtensionsProgram(t, files)
		assert.Equal(t, len(configMessages), 0, configMessages)
		assert.DeepEqual(t, config.SourceExtensions(), []string{".a"})
		assert.DeepEqual(t, config.FileNames(), []tspath.RootedFilePath{"/src/main.a", "/src/geometry.a"})
	})

	t.Run("invalid entries are reported and left out", func(t *testing.T) {
		t.Parallel()
		files := with(`{"sourceExtensions": ["a", ".ts", ".a", ".a"], "compilerOptions": {"noLib": true, "noEmit": true}, "include": ["*.ts"]}`)
		_, config, configMessages := sourceExtensionsProgram(t, files)
		assert.DeepEqual(t, config.SourceExtensions(), []string{".a"})
		assert.Equal(t, len(configMessages), 3, configMessages)
	})
}

// A source extension's JavaScript output is X.a.js and its declaration output X.d.a.ts. ChangeExtension
// doesn't strip an extension TypeScript doesn't know, so the JavaScript name used to be the input's own.
func TestSourceExtensionOutputPaths(t *testing.T) {
	t.Parallel()
	files := map[string]any{
		"/src/tsconfig.json": `{"sourceExtensions": [".a"], "compilerOptions": {"noLib": true, "declaration": true, "outDir": "/out"}, "include": ["*.a"]}`,
		"/src/main.a":        "export const value = 1;\n",
	}
	program, _, _ := sourceExtensionsProgram(t, files)
	options := program.Options()
	main := tspath.RootedFilePathFromNormalized("/src/main.a")
	assert.Equal(t, outputpaths.GetOutputJSFileName(main, options, program).AsString(), "/out/main.a.js")
	assert.Equal(t, outputpaths.GetOutputDeclarationFileNameWorker(main, options, program).AsString(), "/out/main.d.a.ts")
}
