package program_test

import (
	"context"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

// The lib files TypeScript ships are never type-checked, and the project's own files are (#fay5rd1).
// The project file here redeclares a property of a lib interface with another type: TypeScript reports
// that at the later declaration, the project's, when it checks the project file. So the diagnostic a run
// reports is still found with the lib files unchecked, and the control is that it is found at all, in the
// project file. Without the skip, every lib file would be checked; with it, none is.
func TestTheLibFilesAreNotCheckedAndTheProjectsFilesAre(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json": `{"compilerOptions": {"target": "ES2022", "strict": true, "noEmit": true}, "include": ["./*.ts"]}`,
		"globals.ts":    "interface Array<T> {\n  length: string;\n}\n",
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}

	libraries, checked := 0, 0
	for _, sourceFile := range graph.SourceFiles() {
		if !strings.HasPrefix(sourceFile.FileName().AsString(), "bundled:///") {
			if graph.Program.SkipTypeChecking(sourceFile, false) {
				t.Errorf("the project's own %s is not checked", sourceFile.FileName())
			}
			continue
		}
		libraries++
		if !graph.Program.SkipTypeChecking(sourceFile, false) {
			checked++
		}
	}
	if libraries < 10 || checked != 0 {
		t.Fatalf("%d of %d lib files are still checked", checked, libraries)
	}

	found := false
	for _, diagnostic := range graph.AllDiagnostics(context.Background()) {
		if diagnostic.Code() == 2717 && diagnostic.File() != nil && strings.HasSuffix(diagnostic.File().FileName().AsString(), "globals.ts") {
			found = true
		}
	}
	if !found {
		t.Fatal("the conflicting redeclaration of Array's length was not reported in the project's file (TS2717)")
	}
}
