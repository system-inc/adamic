package program_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/types/program"
)

// libraryParsesConfig is a one-file project at a given target and lib, the shape a rule test builds.
func libraryParsesConfig(target string, libs string) string {
	lib := ""
	if libs != "" {
		lib = fmt.Sprintf(`, "lib": [%s]`, libs)
	}
	return fmt.Sprintf(`{"compilerOptions": {"target": %q, "module": "esnext", "moduleResolution": "bundler", "strict": true, "noEmit": true%s}, "include": ["./*.ts"]}`, target, lib)
}

func buildWithLibraryParses(t testing.TB, directory string, parses *program.LibraryParses) *program.Graph {
	t.Helper()
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory, LibraryParses: parses})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	return graph
}

func isBundledLibrary(sourceFile *ast.SourceFile) bool {
	return strings.HasPrefix(sourceFile.FileName().AsString(), "bundled:///")
}

// Two programs built with one cache hold the same parse of every bundled lib file, and their own parses of
// everything else. The control is the project file: a cache that shared it would serve one test's fixture to
// another.
func TestLibraryParsesShareEachLibFileAcrossPrograms(t *testing.T) {
	t.Parallel()
	parses := program.NewLibraryParses()
	first := buildWithLibraryParses(t, writeProject(t, map[string]string{
		"tsconfig.json": libraryParsesConfig("ES2022", ""),
		"main.ts":       "export const first: number = 1;\n",
	}), parses)
	second := buildWithLibraryParses(t, writeProject(t, map[string]string{
		"tsconfig.json": libraryParsesConfig("ES2022", ""),
		"main.ts":       "export const second: string = 'two';\n",
	}), parses)

	secondByName := map[string]*ast.SourceFile{}
	for _, sourceFile := range second.SourceFiles() {
		secondByName[sourceFile.FileName().AsString()] = sourceFile
	}
	shared, libraries := 0, 0
	for _, sourceFile := range first.SourceFiles() {
		other, inBoth := secondByName[sourceFile.FileName().AsString()]
		if !isBundledLibrary(sourceFile) {
			if inBoth && other == sourceFile {
				t.Errorf("%s, a project file, is one parse in two programs", sourceFile.FileName())
			}
			continue
		}
		libraries++
		if !inBoth || other != sourceFile {
			t.Errorf("%s was parsed twice with one cache", sourceFile.FileName())
			continue
		}
		shared++
	}
	if libraries < 10 || shared != libraries {
		t.Fatalf("shared %d of %d lib files, and an ES2022 program loads dozens", shared, libraries)
	}
	if parses.Len() != libraries {
		t.Errorf("the cache holds %d parses for %d lib files", parses.Len(), libraries)
	}
}

// A program built from a cache another target or lib filled loads exactly the files, and reports exactly the
// diagnostics, an unshared build of the same config does. Sharing must only ever save a parse, never change
// which files a program holds or what it says about them.
func TestLibraryParsesNeverChangeWhatAProgramLoadsOrReports(t *testing.T) {
	t.Parallel()
	parses := program.NewLibraryParses()
	// One program fills the cache first with a wide lib set.
	buildWithLibraryParses(t, writeProject(t, map[string]string{
		"tsconfig.json": libraryParsesConfig("ESNext", `"esnext", "dom", "dom.iterable"`),
		"main.ts":       "export const filler = 1;\n",
	}), parses)

	// Each case uses something only a narrower or different lib has, so a leaked file changes the verdict.
	for _, testCase := range []struct {
		name   string
		target string
		libs   string
		source string
		// errors is whether the unshared build reports one: the control that a leaked lib file, which
		// would supply the missing declaration, changes the verdict rather than passing unseen.
		errors bool
	}{
		{"lib es5 has no Symbol", "ES2015", `"es5"`, "export const unique = Symbol('unique');\n", true},
		{"ES2015 has no Array.prototype.includes", "ES2015", `"es2015"`, "export const found = [1].includes(1);\n", true},
		{"no dom means no document", "ES2022", `"es2022"`, "export const body = document.body;\n", true},
		{"ES2022 with dom has both", "ES2022", `"es2022", "dom"`, "export const both = [1].at(0) ?? document.title;\n", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			directory := writeProject(t, map[string]string{
				"tsconfig.json": libraryParsesConfig(testCase.target, testCase.libs),
				"main.ts":       testCase.source,
			})
			unshared := buildWithLibraryParses(t, directory, nil)
			shared := buildWithLibraryParses(t, directory, parses)

			names := func(graph *program.Graph) []string {
				fileNames := []string{}
				for _, sourceFile := range graph.SourceFiles() {
					fileNames = append(fileNames, sourceFile.FileName().AsString())
				}
				slices.Sort(fileNames)
				return fileNames
			}
			if !slices.Equal(names(unshared), names(shared)) {
				t.Fatalf("the shared build loads different files:\nunshared %v\nshared   %v", names(unshared), names(shared))
			}

			messages := func(graph *program.Graph) []string {
				texts := []string{}
				for _, diagnostic := range graph.AllDiagnostics(context.Background()) {
					texts = append(texts, fmt.Sprintf("TS%d %s", diagnostic.Code(), diagnostic.MessageText()))
				}
				slices.Sort(texts)
				return texts
			}
			if reported := len(messages(unshared)) > 0; reported != testCase.errors {
				t.Fatalf("control: the unshared build reports %v, so this case cannot tell a leaked lib file apart", messages(unshared))
			}
			if !slices.Equal(messages(unshared), messages(shared)) {
				t.Fatalf("the shared build reports differently:\nunshared %v\nshared   %v", messages(unshared), messages(shared))
			}
		})
	}
}

// Programs built and checked at once share parses without a race. Run under -race, this is the guard on the
// cache's locking and on binding a shared file from several programs at once.
func TestLibraryParsesAreSafeAcrossConcurrentPrograms(t *testing.T) {
	t.Parallel()
	parses := program.NewLibraryParses()
	var group sync.WaitGroup
	for index := range 8 {
		directory := writeProject(t, map[string]string{
			"tsconfig.json": libraryParsesConfig("ES2022", ""),
			"main.ts":       fmt.Sprintf("export const value%d: number = '%d';\n", index, index),
		})
		group.Add(1)
		go func() {
			defer group.Done()
			graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory, LibraryParses: parses})
			if err != nil {
				t.Errorf("building: %v", err)
				return
			}
			if diagnostics := graph.AllDiagnostics(context.Background()); len(diagnostics) != 1 {
				t.Errorf("want the one planted type error, got %d", len(diagnostics))
			}
		}()
	}
	group.Wait()
}

// BenchmarkOneFileProgram is the rule-test fixture's build, with and without shared lib parses.
func BenchmarkOneFileProgram(b *testing.B) {
	directory := writeProject(b, map[string]string{
		"tsconfig.json": libraryParsesConfig("ES2022", ""),
		"main.ts":       "export const value: number = 1;\n",
	})
	for _, mode := range []struct {
		name   string
		parses *program.LibraryParses
	}{{"unshared", nil}, {"shared", program.NewLibraryParses()}} {
		b.Run(mode.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				buildWithLibraryParses(b, directory, mode.parses)
			}
		})
	}
}
