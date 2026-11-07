package load

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// writeProgram writes files into a fresh directory and returns their absolute paths, in the order
// given, so a test reads as the program it checks.
func writeProgram(test *testing.T, files ...[2]string) []string {
	test.Helper()
	directory := test.TempDir()
	paths := make([]string, 0, len(files))
	for _, file := range files {
		path := filepath.Join(directory, file[0])
		if writeError := os.WriteFile(path, []byte(file[1]), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		paths = append(paths, path)
	}
	return paths
}

// declarationsByName loads a program and indexes what it proved, failing the test if it doesn't load.
func declarationsByName(test *testing.T, paths []string) map[string]Declaration {
	test.Helper()
	program, loadError := Load(paths)
	if loadError != nil {
		test.Fatalf("Load: %v", loadError)
	}
	byName := map[string]Declaration{}
	for _, declaration := range program.Declarations(context.Background()) {
		byName[declaration.Kind+" "+declaration.Name] = declaration
	}
	return byName
}

// checkErrors loads a program the checker should reject and returns what it said.
func checkErrors(test *testing.T, paths []string) []string {
	test.Helper()
	_, loadError := Load(paths)
	var checkError *CheckError
	if !errors.As(loadError, &checkError) {
		test.Fatalf("Load: want a CheckError, got %v", loadError)
	}
	return checkError.Diagnostics
}

func TestDeclarationsCarryTheirProvenTypes(test *testing.T) {
	test.Parallel()
	paths := writeProgram(test, [2]string{"main.ts", `const answer = 42;
function double(value: number): number {
	return value * 2;
}
const [first, second] = ['a', 7] as const;
type Shape = { readonly kind: 'Circle'; readonly radius: number } | { readonly kind: 'Square'; readonly side: number };
`})
	byName := declarationsByName(test, paths)
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
			test.Errorf("no %s %s among %v", want.Kind, want.Name, byName)
			continue
		}
		if got.Type != want.Type || got.Line != want.Line || got.Column != want.Column {
			test.Errorf(
				"%s %s: got %d:%d %q, want %d:%d %q",
				want.Kind,
				want.Name,
				got.Line,
				got.Column,
				got.Type,
				want.Line,
				want.Column,
				want.Type,
			)
		}
	}
	// A type alias reports what it stands for, not its own name.
	if shape := byName["type Shape"]; !strings.Contains(shape.Type, `kind: "Circle"`) {
		test.Errorf("type Shape: got %q, want the union it stands for", shape.Type)
	}
}

func TestATypeErrorFailsTheLoad(test *testing.T) {
	test.Parallel()
	paths := writeProgram(test, [2]string{"main.ts", "const broken: number = 'x';\n"})
	diagnostics := checkErrors(test, paths)
	// The message text too: a diagnostic carries a key, not text, and an empty message once passed
	// every check that looked only at the code.
	if len(diagnostics) != 1 || !strings.HasSuffix(
		diagnostics[0],
		"main.ts:1:7: error TS2322: Type 'string' is not assignable to type 'number'.",
	) {
		test.Errorf("got %q, want one TS2322 at main.ts:1:7 with its message", diagnostics)
	}
}

// Each of these compiles under plain strict and must not under Adamic's options, so each one proves
// its option is really on.
func TestAdamicOptionsAreOn(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		option string
		source string
		code   string
	}{
		{"noUncheckedIndexedAccess", "const items: number[] = [1];\nconst first: number = items[0];\n", "TS2322"},
		{"exactOptionalPropertyTypes", "const point: { x?: number } = { x: undefined };\n", "TS2375"},
		{"lib is es2024, not the DOM", "const element = document.body;\n", "TS2584"},
	} {
		test.Run(probe.option, func(test *testing.T) {
			test.Parallel()
			diagnostics := checkErrors(test, writeProgram(test, [2]string{"main.ts", probe.source}))
			if !slices.ContainsFunc(
				diagnostics,
				func(diagnostic string) bool { return strings.Contains(diagnostic, probe.code) },
			) {
				test.Errorf("got %q, want %s", diagnostics, probe.code)
			}
		})
	}
}

// Every file is a module, as the oracle runs it, imports or not. As scripts, two files declaring one
// name would share a global scope and collide (TS2451), and a local console would collide with the
// prelude's.
func TestEveryFileIsAModule(test *testing.T) {
	test.Parallel()
	paths := writeProgram(test,
		[2]string{"first.a", "const name = 'first';\n"},
		[2]string{"second.a", "const name = 'second';\nconst console = 'mine';\n"},
	)
	if _, loadError := Load(paths); loadError != nil {
		test.Errorf("Load: %v", loadError)
	}
}

func TestAdamicFilesImportEachOther(test *testing.T) {
	test.Parallel()
	paths := writeProgram(test,
		[2]string{"main.a", "import { area } from './geometry.a';\nconst unit = area(1);\n"},
		[2]string{"geometry.a", "export function area(radius: number): number {\n\treturn Math.PI * radius ** 2;\n}\n"},
	)
	byName := declarationsByName(test, paths[:1])
	unit := byName["variable unit"]
	if unit.Type != "number" || !strings.HasSuffix(unit.File, "/main.a") {
		test.Errorf("variable unit: got %q in %s, want number in main.a", unit.Type, unit.File)
	}
	// The imported file is checked too, and named as written: an error there points at geometry.a.
	broken := writeProgram(test,
		[2]string{"main.a", "import { area } from './geometry.a';\nconst unit: string = area(1);\n"},
		[2]string{"geometry.a", "export function area(radius: number): number {\n\treturn 'wide';\n}\n"},
	)
	diagnostics := checkErrors(test, broken[:1])
	if len(diagnostics) != 2 ||
		!strings.Contains(diagnostics[0], "/geometry.a:2:2: error TS2322") ||
		!strings.Contains(diagnostics[1], "/main.a:2:7: error TS2322") {
		test.Errorf("got %q, want TS2322 in geometry.a:2:2 and main.a:2:7", diagnostics)
	}
}

func TestThePreludeDeclaresConsoleAndPanic(test *testing.T) {
	test.Parallel()
	paths := writeProgram(test, [2]string{"main.ts", `import { panic } from 'adamic';
const ages = new Map<string, number>([['Kirk', 42]]);
const age = ages.get('Kirk') ?? panic('no age');
console.log(` + "`${age}`" + `);
`})
	if age := declarationsByName(test, paths)["variable age"]; age.Type != "number" {
		test.Errorf("variable age: got %q, want number (panic returns never)", age.Type)
	}
	// Control: console.log takes one string in 0.1, so a number is a type error.
	diagnostics := checkErrors(test, writeProgram(test, [2]string{"main.ts", "console.log(42);\n"}))
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "TS2345") {
		test.Errorf("got %q, want TS2345", diagnostics)
	}
}

func TestLoadRefusesWhatIsNotAdamic(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name  string
		files [][2]string
		want  string
	}{
		{"a JavaScript file", [][2]string{{"main.js", "const answer = 42;\n"}}, "is not an Adamic source"},
		{"a missing file", nil, "no file at"},
		{
			"an .a file shadowed by an .a.ts",
			[][2]string{{"main.a", "const a = 1;\n"}, {"main.a.ts", "const b = 2;\n"}},
			"both exist",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			paths := writeProgram(test, probe.files...)
			if len(paths) == 0 {
				paths = []string{filepath.Join(test.TempDir(), "missing.a")}
			}
			_, loadError := Load(paths[:1])
			if loadError == nil || !strings.Contains(loadError.Error(), probe.want) {
				test.Errorf("got %v, want an error containing %q", loadError, probe.want)
			}
		})
	}
	if _, loadError := Load(nil); loadError == nil {
		test.Error("Load(nil): want an error, got none")
	}
}

// Every program in docs/0.1.md typechecks under stage 0, as each did under tsc 6.0.3. The five
// refusals load too: they're valid TypeScript, and refusing them is cohere:adamic's job, not the
// checker's.
func TestTheProgramsInTheSpecLoad(test *testing.T) {
	test.Parallel()
	for _, group := range []string{"compile", "refuse"} {
		entries, readDirectoryError := os.ReadDir(filepath.Join("testdata", "0.1", group))
		if readDirectoryError != nil {
			test.Fatal(readDirectoryError)
		}
		if len(entries) == 0 {
			test.Fatalf("testdata/0.1/%s is empty", group)
		}
		for _, entry := range entries {
			path, absolutePathError := filepath.Abs(filepath.Join("testdata", "0.1", group, entry.Name()))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			if entry.IsDir() {
				path = filepath.Join(path, "main.ts")
			}
			test.Run(group+"/"+entry.Name(), func(test *testing.T) {
				test.Parallel()
				if _, loadError := Load([]string{path}); loadError != nil {
					test.Errorf("Load: %v", loadError)
				}
			})
		}
	}
}

func TestOverlayPrecedesDiskForAdamicAliases(test *testing.T) {
	test.Parallel()
	for _, extension := range []string{".ts", ".a"} {
		test.Run(extension, func(test *testing.T) {
			test.Parallel()
			paths := writeProgram(test, [2]string{"main" + extension, "const value: number = 'disk';"})
			replacement := "const value: number = 42;"
			if _, loadOverlayError := LoadOverlay(
				paths,
				map[string]string{paths[0]: replacement},
			); loadOverlayError != nil {
				test.Fatalf("overlay not checked: %v", loadOverlayError)
			}
			diagnostics := checkErrors(test, paths)
			if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "TS2322") {
				test.Fatalf("disk source changed: %v", diagnostics)
			}
		})
	}
}

// These checks are house style. Strict typing still includes the undefined result at an implicit exit.
func TestHouseStyleOptionsDoNotRejectValidControlFlow(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"function sign(value: number) { if (value > 0) { return 1; } }",
		"function choose(value: number): number { let result=0; switch(value) { case 1: result+=1; default: result+=2; } return result; }",
	} {
		test.Run(source, func(test *testing.T) {
			test.Parallel()
			if _, loadError := Load(writeProgram(test, [2]string{"main.a", source})); loadError != nil {
				test.Fatal(loadError)
			}
		})
	}
	diagnostics := checkErrors(test, writeProgram(
		test,
		[2]string{"main.a", "function sign(value: number): number { if (value > 0) { return 1; } }"},
	))
	if !slices.ContainsFunc(
		diagnostics,
		func(diagnostic string) bool { return strings.Contains(diagnostic, "TS2366") },
	) {
		test.Fatalf("the explicit result still needs undefined: %v", diagnostics)
	}
}

// A program loaded with a display names its files through it, everywhere it names them: a file's name, a node's
// position, a declaration's file, and a diagnostic. Lowered-program products load so to name no build directory
// (#tqrqx60: landable-13b's estree products held their random build directory, so no two builds were the same bytes).
// Mutant: displayName ignoring the display.
func TestALoadedProgramNamesItsFilesAsDisplayed(test *testing.T) {
	test.Parallel()
	paths := writeProgram(test, [2]string{"main.ts", "import { half } from './helper.ts';\nconst answer = half(84);\n"},
		[2]string{"helper.ts", "export function half(value: number): number {\n\treturn value / 2;\n}\n"})
	directory := filepath.Dir(paths[0])
	display := func(name string) string { return strings.Replace(name, directory, "<build cache>/product", 1) }
	program, loadDisplayedError := LoadDisplayed(paths[:1], display)
	if loadDisplayedError != nil {
		test.Fatal(loadDisplayedError)
	}
	if got := program.FileName(program.Files()[0]); got != "<build cache>/product/main.ts" {
		test.Errorf("the file is named %q", got)
	}
	for _, declaration := range program.Declarations(context.Background()) {
		if strings.Contains(declaration.File, directory) ||
			!strings.HasPrefix(declaration.File, "<build cache>/product/") {
			test.Errorf("%s %s is declared in %q", declaration.Kind, declaration.Name, declaration.File)
		}
	}
	broken := writeProgram(test, [2]string{"main.ts", "const broken: number = 'x';\n"})
	_, loadDisplayedError = LoadDisplayed(
		broken,
		func(name string) string { return strings.Replace(name, filepath.Dir(broken[0]), "<repository>", 1) },
	)
	var rejected *CheckError
	if !errors.As(loadDisplayedError, &rejected) ||
		len(rejected.Diagnostics) != 1 ||
		!strings.HasPrefix(rejected.Diagnostics[0], "<repository>/main.ts:1:7: ") {
		test.Errorf("the diagnostic reads %v", loadDisplayedError)
	}
}

// A load writes one phase line to ADAMIC_BUILD_LOG, in internal/native's buildLog format, whatever its time (#9x03wa7).
// Mutant: load's line dropped.
//
// Not parallel: test.Setenv changes ADAMIC_BUILD_LOG for the process.
func TestLoadLogsItsPhaseOnce(test *testing.T) {
	paths := writeProgram(test, [2]string{"main.ts", "console.log('phase');\n"})
	log := filepath.Join(test.TempDir(), "build.log")
	test.Setenv("ADAMIC_BUILD_LOG", log)
	if _, loadError := Load(paths); loadError != nil {
		test.Fatal(loadError)
	}
	written, readError := os.ReadFile(log)
	if readError != nil {
		test.Fatal(readError)
	}
	if !regexp.MustCompile(`^phase load - off [0-9]+\.[0-9]{3}\n$`).Match(written) {
		test.Fatalf("wrote %q, want one load phase line", written)
	}
}
