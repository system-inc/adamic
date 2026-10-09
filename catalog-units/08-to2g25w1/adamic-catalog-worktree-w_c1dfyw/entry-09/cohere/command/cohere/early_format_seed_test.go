package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
	"github.com/system-inc/cohere/internal/types/program"
)

// seedSources are the files the test below formats: shapes whose ESTree reads the node flags the binder writes
// beside (let, const and using declarations, exports, a class with a body), in .ts and .tsx, and a declaration
// file, whose parse differs by name.
var seedSources = map[string]string{
	"Values.ts":      "export const alpha = 1;\nlet beta = alpha + 1;\nexport { beta };\n",
	"Shapes.ts":      "export class Shape {\n    constructor(private readonly side: number) {}\n    area(): number {\n        return this.side * this.side;\n    }\n}\n",
	"Using.ts":       "export function read(): void {\n    using handle = { [Symbol.dispose]() {} };\n}\n",
	"View.tsx":       "export function View(properties: { label: string }) {\n    return <div className=\"view\">{properties.label}</div>;\n}\n",
	"Unformatted.ts": "export   const   gamma=[1,2,3].map((x)=>x*2)\n",
	"Globals.d.ts":   "declare const version: string;\n",
}

// The early pass formats a TypeScript file from the program's own tree once the program has bound it, rather
// than parsing it twice more (#dk2502g). Two things are proved here, on a real program.
//
// Taken, the tree formats exactly as a parse of the file's own does: every file's result is compared with a
// speculation that has no program, and every TypeScript file must have been formatted from the program's tree.
//
// And it is never read while the binder writes it. The checkers bind every file on their own goroutines, and
// here they are created at the same moment the pass starts reading the same files, many times over: run under
// -race, a tree read before its binding finished is reported. boundTreeOf without its IsBound check is the
// mutant that must be caught there; it was, see #dk2502g.
func TestTheEarlyPassFormatsFromTheProgramsBoundTreeAndNeverDuringBinding(t *testing.T) {
	t.Parallel()
	// Resolved, so the names are the program's own: on macOS the temporary directory is behind a symbolic link.
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var candidates []string
	for name, source := range seedSources {
		for index := range 8 {
			fileName := filepath.Join(directory, fmt.Sprintf("%d%s", index, name))
			if err := os.WriteFile(fileName, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			candidates = append(candidates, fileName)
		}
	}
	configuration := `{"compilerOptions":{"strict":true,"module":"esnext","target":"esnext","moduleResolution":"bundler","jsx":"preserve","lib":["esnext","esnext.disposable"]},"include":["**/*.ts","**/*.tsx"]}`
	if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
		t.Fatal(err)
	}

	formatter := native.Formatter{Options: formatoptions.Default()}
	transformCounting := func(programs *programOffer, fromProgram *atomic.Int32) func(string, string, *ast.SourceFile) (string, error) {
		return func(fileName string, text string, parsed *ast.SourceFile) (string, error) {
			if programs != nil && parsed != nil {
				if program := programs.program.Load(); program != nil && parsed == program.GetSourceFile(tspath.RootedFilePath(fileName)) {
					fromProgram.Add(1)
				}
			}
			return formatter.FormatParsed(fileName, text, parsed)
		}
	}

	var none atomic.Int32
	reference := speculateFormatOn(candidates, transformCounting(nil, &none), 1, 4, nil)
	reference.wait()

	for round := range 6 {
		graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
		if err != nil {
			t.Fatal(err)
		}
		var programs programOffer
		programs.offer(graph)
		var fromProgram atomic.Int32

		// The checkers bind every file while the pass reads them.
		start := make(chan struct{})
		var binding sync.WaitGroup
		binding.Add(1)
		go func() {
			defer binding.Done()
			<-start
			_, release := graph.Program.GetTypeCheckerForFile(context.Background(), graph.ProjectFiles()[0])
			release()
		}()
		close(start)
		during := speculateFormatOn(candidates, transformCounting(&programs, &fromProgram), 1, 8, &programs)
		during.wait()
		binding.Wait()
		sameResults(t, fmt.Sprintf("round %d, during binding", round), reference, during)

		// Bound now: every file must be formatted from the program's tree, once each.
		fromProgram.Store(0)
		after := speculateFormatOn(candidates, transformCounting(&programs, &fromProgram), 1, 8, &programs)
		after.wait()
		sameResults(t, fmt.Sprintf("round %d, after binding", round), reference, after)
		if formatted := int(fromProgram.Load()); formatted != len(candidates) {
			t.Fatalf("round %d: %d of %d files were formatted from the program's tree", round, formatted, len(candidates))
		}
	}
}

// sameResults fails unless two speculations kept the same files with the same results.
func sameResults(t *testing.T, label string, expected *formatSpeculation, actual *formatSpeculation) {
	t.Helper()
	if len(actual.attempts) != len(expected.attempts) {
		t.Errorf("%s: %d files kept, against %d with no program", label, len(actual.attempts), len(expected.attempts))
	}
	for fileName, want := range expected.attempts {
		got, kept := actual.attempts[fileName]
		if !kept {
			t.Errorf("%s: %s was not kept", label, filepath.Base(fileName))
			continue
		}
		if got.result.Text != want.result.Text || got.result.TransformSkipped != want.result.TransformSkipped ||
			errorText(got.err) != errorText(want.err) {
			t.Errorf("%s: %s formatted otherwise from the program's tree", label, filepath.Base(fileName))
		}
	}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
