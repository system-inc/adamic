package load

import (
	"os"
	"strings"
	"testing"
)

func TestCompositeProjectCensus(t *testing.T) {
	t.Parallel()
	main, err := os.ReadFile("testdata/project-census/main.a")
	if err != nil {
		t.Fatal(err)
	}
	values, err := os.ReadFile("testdata/project-census/values.a")
	if err != nil {
		t.Fatal(err)
	}
	paths := writeProgram(t,
		[2]string{"main.ts", string(main)},
		[2]string{"values.ts", string(values)},
		[2]string{"tsconfig.json", `{"compilerOptions":{"composite":true,"strict":true,"target":"es2024","lib":["es2024"],"types":[],"module":"nodenext"},"files":["main.ts","values.ts"]}`},
	)
	for _, roots := range [][]string{paths[:1], paths[1:2], paths[:2]} {
		program, err := Load(roots)
		if err != nil {
			t.Errorf("%v: %v", roots, err)
			continue
		}
		if len(program.Files()) != len(roots) {
			t.Errorf("exposed unrequested roots: %d", len(program.Files()))
		}
	}
}

func TestProjectCensusDiagnosticScope(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"main.ts", `import * as values from './values.js'; export const answer: number = values.answer;`},
		[2]string{"values.ts", `export const answer: number = 42; const broken: number = 'wrong';`},
		[2]string{"tsconfig.json", `{"compilerOptions":{"composite":true,"strict":true,"target":"es2024","lib":["es2024"],"types":[],"module":"nodenext"},"files":["main.ts","values.ts"]}`},
	)
	if _, err := Load(paths[:1]); err != nil {
		t.Fatalf("unrequested diagnostics: %v", err)
	}
	for _, roots := range [][]string{paths[1:2], paths[:2]} {
		diagnostics := checkErrors(t, roots)
		if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "values.ts:1:") || !strings.Contains(diagnostics[0], "TS2322") {
			t.Fatalf("requested diagnostics: %v", diagnostics)
		}
	}
}
