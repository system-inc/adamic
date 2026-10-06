package load

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// writeProgram writes files into a fresh directory and returns their absolute paths, in the order
// given, so a test reads as the program it checks.
func writeProgram(t *testing.T, files ...[2]string) []string {
	t.Helper()
	directory := t.TempDir()
	paths := make([]string, 0, len(files))
	for _, file := range files {
		path := filepath.Join(directory, file[0])
		if err := os.WriteFile(path, []byte(file[1]), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	return paths
}

// declarationsByName loads a program and indexes what it proved, failing the test if it doesn't load.
func declarationsByName(t *testing.T, paths []string) map[string]Declaration {
	t.Helper()
	program, err := Load(paths)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	byName := map[string]Declaration{}
	for _, declaration := range program.Declarations(context.Background()) {
		byName[declaration.Kind+" "+declaration.Name] = declaration
	}
	return byName
}

// checkErrors loads a program the checker should reject and returns what it said.
func checkErrors(t *testing.T, paths []string) []string {
	t.Helper()
	_, err := Load(paths)
	var checkError *CheckError
	if !errors.As(err, &checkError) {
		t.Fatalf("Load: want a CheckError, got %v", err)
	}
	return checkError.Diagnostics
}

func TestDeclarationsCarryTheirProvenTypes(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.ts", `const answer = 42;
function double(value: number): number {
	return value * 2;
}
const [first, second] = ['a', 7] as const;
type Shape = { readonly kind: 'Circle'; readonly radius: number } | { readonly kind: 'Square'; readonly side: number };
`})
	byName := declarationsByName(t, paths)
	for _, want := range []Declaration{
		{Kind: "variable", Name: "answer", Type: "42", Line: 1, Column: 7},
		{Kind: "function", Name: "double", Type: "(value: number) => number", Line: 2, Column: 10},
		{Kind: "parameter", Name: "value", Type: "number", Line: 2, Column: 17},
		{Kind: "binding", Name: "first", Type: `"a"`, Line: 5, Column: 8},
		{Kind: "binding", Name: "second", Type: "7", Line: 5, Column: 15},
		{Kind: "property", Name: "radius", Type: "number", Line: 6, Column: 50},
	} {
		got, isFound := byName[want.Kind+" "+want.Name]
		if !isFound {
			t.Errorf("no %s %s among %v", want.Kind, want.Name, byName)
			continue
		}
		if got.Type != want.Type || got.Line != want.Line || got.Column != want.Column {
			t.Errorf("%s %s: got %d:%d %q, want %d:%d %q", want.Kind, want.Name, got.Line, got.Column, got.Type, want.Line, want.Column, want.Type)
		}
	}
	// A type alias reports what it stands for, not its own name.
	if shape := byName["type Shape"]; !strings.Contains(shape.Type, `kind: "Circle"`) {
		t.Errorf("type Shape: got %q, want the union it stands for", shape.Type)
	}
}

func TestATypeErrorFailsTheLoad(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.ts", "const broken: number = 'x';\n"})
	diagnostics := checkErrors(t, paths)
	// The message text too: a diagnostic carries a key, not text, and an empty message once passed
	// every check that looked only at the code.
	if len(diagnostics) != 1 || !strings.HasSuffix(diagnostics[0], "main.ts:1:7: error TS2322: Type 'string' is not assignable to type 'number'.") {
		t.Errorf("got %q, want one TS2322 at main.ts:1:7 with its message", diagnostics)
	}
}

// Each of these compiles under plain strict and must not under Adamic's options, so each one proves
// its option is really on.
func TestAdamicOptionsAreOn(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		option string
		source string
		code   string
	}{
		{"noUncheckedIndexedAccess", "const items: number[] = [1];\nconst first: number = items[0];\n", "TS2322"},
		{"exactOptionalPropertyTypes", "const point: { x?: number } = { x: undefined };\n", "TS2375"},
		// Unannotated, so strict's own TS2366 (a declared return type with a missing return) can't stand in.
		{"noImplicitReturns", "function sign(value: number) {\n\tif (value > 0) {\n\t\treturn 1;\n\t}\n}\n", "TS7030"},
		{"erasableSyntaxOnly", "enum Color {\n\tRed,\n}\n", "TS1294"},
		{"lib is es2024, not the DOM", "const element = document.body;\n", "TS2584"},
	} {
		t.Run(probe.option, func(t *testing.T) {
			t.Parallel()
			diagnostics := checkErrors(t, writeProgram(t, [2]string{"main.ts", probe.source}))
			if !slices.ContainsFunc(diagnostics, func(diagnostic string) bool { return strings.Contains(diagnostic, probe.code) }) {
				t.Errorf("got %q, want %s", diagnostics, probe.code)
			}
		})
	}
}

// Every file is a module, as the oracle runs it, imports or not. As scripts, two files declaring one
// name would share a global scope and collide (TS2451), and a local console would collide with the
// prelude's.
func TestEveryFileIsAModule(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"first.a", "const name = 'first';\n"},
		[2]string{"second.a", "const name = 'second';\nconst console = 'mine';\n"},
	)
	if _, err := Load(paths); err != nil {
		t.Errorf("Load: %v", err)
	}
}

func TestAdamicFilesImportEachOther(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"main.a", "import { area } from './geometry.a';\nconst unit = area(1);\n"},
		[2]string{"geometry.a", "export function area(radius: number): number {\n\treturn Math.PI * radius ** 2;\n}\n"},
	)
	byName := declarationsByName(t, paths[:1])
	unit := byName["variable unit"]
	if unit.Type != "number" || !strings.HasSuffix(unit.File, "/main.a") {
		t.Errorf("variable unit: got %q in %s, want number in main.a", unit.Type, unit.File)
	}
	// The imported file is checked too, and named as written: an error there points at geometry.a.
	broken := writeProgram(t,
		[2]string{"main.a", "import { area } from './geometry.a';\nconst unit: string = area(1);\n"},
		[2]string{"geometry.a", "export function area(radius: number): number {\n\treturn 'wide';\n}\n"},
	)
	diagnostics := checkErrors(t, broken[:1])
	if len(diagnostics) != 2 || !strings.Contains(diagnostics[0], "/geometry.a:2:2: error TS2322") || !strings.Contains(diagnostics[1], "/main.a:2:7: error TS2322") {
		t.Errorf("got %q, want TS2322 in geometry.a:2:2 and main.a:2:7", diagnostics)
	}
}

func TestThePreludeDeclaresConsoleAndPanic(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.ts", `import { panic } from 'adamic';
const ages = new Map<string, number>([['Kirk', 42]]);
const age = ages.get('Kirk') ?? panic('no age');
console.log(` + "`${age}`" + `);
`})
	if age := declarationsByName(t, paths)["variable age"]; age.Type != "number" {
		t.Errorf("variable age: got %q, want number (panic returns never)", age.Type)
	}
	// Control: console.log takes one string in 0.1, so a number is a type error.
	diagnostics := checkErrors(t, writeProgram(t, [2]string{"main.ts", "console.log(42);\n"}))
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "TS2345") {
		t.Errorf("got %q, want TS2345", diagnostics)
	}
}

func TestLoadRefusesWhatIsNotAdamic(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name  string
		files [][2]string
		want  string
	}{
		{"a JavaScript file", [][2]string{{"main.js", "const answer = 42;\n"}}, "is not an Adamic source"},
		{"a missing file", nil, "no file at"},
		{"an .a file shadowed by an .a.ts", [][2]string{{"main.a", "const a = 1;\n"}, {"main.a.ts", "const b = 2;\n"}}, "both exist"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			paths := writeProgram(t, probe.files...)
			if len(paths) == 0 {
				paths = []string{filepath.Join(t.TempDir(), "missing.a")}
			}
			_, err := Load(paths[:1])
			if err == nil || !strings.Contains(err.Error(), probe.want) {
				t.Errorf("got %v, want an error containing %q", err, probe.want)
			}
		})
	}
	if _, err := Load(nil); err == nil {
		t.Error("Load(nil): want an error, got none")
	}
}

// Every program in docs/0.1.md typechecks under stage 0, as each did under tsc 6.0.3. The five
// refusals load too: they're valid TypeScript, and refusing them is cohere:adamic's job, not the
// checker's.
func TestTheProgramsInTheSpecLoad(t *testing.T) {
	t.Parallel()
	for _, group := range []string{"compile", "refuse"} {
		entries, err := os.ReadDir(filepath.Join("testdata", "0.1", group))
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 0 {
			t.Fatalf("testdata/0.1/%s is empty", group)
		}
		for _, entry := range entries {
			path, err := filepath.Abs(filepath.Join("testdata", "0.1", group, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if entry.IsDir() {
				path = filepath.Join(path, "main.ts")
			}
			t.Run(group+"/"+entry.Name(), func(t *testing.T) {
				t.Parallel()
				if _, err := Load([]string{path}); err != nil {
					t.Errorf("Load: %v", err)
				}
			})
		}
	}
}

func TestOverlayPrecedesDiskForAdamicAliases(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".ts", ".a"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			paths := writeProgram(t, [2]string{"main" + extension, "const value: number = 'disk';"})
			replacement := "const value: number = 42;"
			if _, err := LoadOverlay(paths, map[string]string{paths[0]: replacement}); err != nil {
				t.Fatalf("overlay not checked: %v", err)
			}
			diagnostics := checkErrors(t, paths)
			if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "TS2322") {
				t.Fatalf("disk source changed: %v", diagnostics)
			}
		})
	}
}
