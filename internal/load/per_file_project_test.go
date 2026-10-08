package load

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func projectProgram(t *testing.T, options, source string) []string {
	t.Helper()
	paths := writeProgram(t, [2]string{"main.ts", source}, [2]string{"tsconfig.json", `{"compilerOptions":` + options + `,"files":["main.ts"]}`})
	return paths[:1]
}

func TestProjectLibCustomSet(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/project-lib/custom_set.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, lib := range []string{"es2024", "es2025"} {
		t.Run(lib, func(t *testing.T) {
			t.Parallel()
			paths := projectProgram(t, `{"strict":true,"target":"es2024","lib":["`+lib+`"],"types":[]}`, string(source))
			if lib == "es2024" {
				if _, err := Load(paths); err != nil {
					t.Fatal(err)
				}
			} else {
				diagnostics := checkErrors(t, paths)
				if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "TS2740") || !strings.Contains(diagnostics[0], "union") {
					t.Fatalf("want ES2025's larger Set shape: %v", diagnostics)
				}
			}
		})
	}
}

func TestProjectIteratorWitnesses(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"iterable", "callback", "entries", "set-copy", "nested-entries"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata/overlay-iterators", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Load(projectProgram(t, `{"strict":true,"target":"es2024","lib":["es2024"],"types":[]}`, string(source))); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProjectLibAndOptions(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, options, source, code string }{
		{"es5", `{"strict":true,"lib":["es5"],"types":[]}`, `const fixed = (1).toFixed();`, ""},
		{"es5 excludes Map", `{"strict":true,"lib":["es5"],"types":[]}`, `const map = new Map();`, "TS2583"},
		{"es2024 excludes union", `{"strict":true,"lib":["es2024"],"types":[]}`, `new Set<string>().union(new Set<string>());`, "TS2550"},
		{"es2025 retains union", `{"strict":true,"lib":["es2025"],"types":[]}`, `new Set<string>().union(new Set<string>());`, ""},
		{"DOM retains console", `{"strict":true,"lib":["es2024","dom"],"types":[]}`, `const body = document.body; console.log(42, 'value');`, ""},
		{"default lib follows target", `{"strict":true,"target":"es2024","types":[]}`, `const body = document.body;`, ""},
		{"unchecked reads disabled", `{"strict":true,"noUncheckedIndexedAccess":false,"lib":["es2024"],"types":[]}`, `const values: number[] = [1]; const first: number = values[0];`, ""},
		{"unchecked reads enabled", `{"strict":true,"noUncheckedIndexedAccess":true,"lib":["es2024"],"types":[]}`, `const values: number[] = [1]; const first: number = values[0];`, "TS2322"},
		{"optional exactness disabled retains unconverted error", `{"strict":true,"exactOptionalPropertyTypes":false,"lib":["es2024"],"types":[]}`, `const value: { x?: number } = { x: undefined };`, ""},
		{"implicit any allowed", `{"strict":false,"lib":["es2024"],"types":[]}`, `function identity(value) { return value; }`, ""},
		{"implicit any rejected", `{"strict":true,"lib":["es2024"],"types":[]}`, `function identity(value) { return value; }`, "TS7006"},
		{"isolated declaration needs annotation", `{"strict":true,"declaration":true,"isolatedDeclarations":true,"noEmit":true,"lib":["es2024"],"types":[]}`, `export const value = (() => 1)();`, "TS9010"},
		{"isolated declaration accepts annotation", `{"strict":true,"declaration":true,"isolatedDeclarations":true,"noEmit":true,"lib":["es2024"],"types":[]}`, `export const value: number = (() => 1)();`, ""},
		{"invalid lib", `{"lib":["not-a-lib"],"types":[]}`, `const value = 1;`, "TS6046"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			paths := projectProgram(t, probe.options, probe.source)
			if probe.code == "" {
				if _, err := Load(paths); err != nil {
					t.Fatal(err)
				}
			} else {
				diagnostics := checkErrors(t, paths)
				if !strings.Contains(strings.Join(diagnostics, "\n"), probe.code) {
					t.Fatalf("want %s: %v", probe.code, diagnostics)
				}
			}
		})
	}
}

func TestProjectExtendsAndDeclarationRoots(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"main.ts", `const value: number = [1][0]; const global: string = projectName; new Set<string>().projectMethod();`},
		[2]string{"globals.d.ts", `declare const projectName: string; interface Set<T> { projectMethod(): number; }`},
		[2]string{"base.json", `{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":false,"target":"es2024","lib":["es2024"],"types":[]}}`},
		[2]string{"tsconfig.json", `{"extends":"./base.json","files":["main.ts","globals.d.ts"]}`},
	)
	program, err := Load(paths[:1])
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Files()) != 1 {
		t.Fatal("project declarations were exposed as program implementation roots")
	}
	if declaration := program.Declarations(context.Background()); len(declaration) == 0 {
		t.Fatal("no proven declarations")
	}
	_, err = LoadOverlay(paths[:1], map[string]string{paths[2]: `{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":true,"lib":["es2024"],"types":[]}}`})
	if err == nil || !strings.Contains(err.Error(), "TS2322") {
		t.Fatalf("config overlay was ignored: %v", err)
	}
}

func TestProjectConfigErrorsAndMixedRoots(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.ts", `const value = 1;`}, [2]string{"tsconfig.json", `{"extends":"./missing.json","files":["main.ts"]}`})
	if diagnostics := checkErrors(t, paths[:1]); !strings.Contains(strings.Join(diagnostics, "\n"), "TS5083") {
		t.Fatalf("missing extends was ignored: %v", diagnostics)
	}
	first := projectProgram(t, `{"lib":["es2024"],"types":[]}`, `const first = 1;`)
	second := projectProgram(t, `{"lib":["es2025"],"types":[]}`, `const second = 2;`)
	if _, err := Load(append(first, second...)); err == nil || !strings.Contains(err.Error(), "different projects") {
		t.Fatalf("mixed project roots were accepted: %v", err)
	}
}

func TestAdamicDefaultsRemainIndependentOfProject(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"main.a", `const values: number[] = [1]; const first: number = values[0];`},
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":false,"lib":["es5"]},"files":["main.a"]}`},
	)
	if diagnostics := checkErrors(t, paths[:1]); !strings.Contains(strings.Join(diagnostics, "\n"), "TS2322") {
		t.Fatalf("project weakened Adamic defaults: %v", diagnostics)
	}
	if _, err := LoadOverlay(paths[:1], map[string]string{paths[0]: `new Set<string>().union(new Set<string>());`}); err != nil {
		t.Fatalf("standalone Set extensions lost: %v", err)
	}
}

func TestProjectExplicitAmbientRoots(t *testing.T) {
	t.Parallel()
	paths := projectProgram(t, `{"strict":true,"lib":["es2024"],"types":[]}`, `const name: string = externalName;`)
	declarations := writeProgram(t, [2]string{"external.d.ts", `declare const externalName: string;`})
	if _, err := Load(append(paths, declarations...)); err != nil {
		t.Fatalf("ambient root treated as a second project: %v", err)
	}
}

func TestProjectModuleDetection(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"first.ts", `const shared = 1;`},
		[2]string{"second.ts", `const shared = 2;`},
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2024"],"types":[],"moduleDetection":"auto"},"files":["first.ts","second.ts"]}`},
	)
	if diagnostics := checkErrors(t, paths[:2]); len(diagnostics) != 2 || !strings.Contains(strings.Join(diagnostics, "\n"), "TS2451") {
		t.Fatalf("project scripts were forced into modules: %v", diagnostics)
	}
	config := `{"compilerOptions":{"strict":true,"lib":["es2024"],"types":[],"moduleDetection":"force"},"files":["first.ts","second.ts"]}`
	if _, err := LoadOverlay(paths[:2], map[string]string{paths[2]: config}); err != nil {
		t.Fatalf("project module option ignored: %v", err)
	}
}

func TestProjectInheritedConsoleDeclaration(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"main.ts", `console.log(); console.log(42, 'value'); console.table([1]); console.error(42);`},
		[2]string{"globals.d.ts", `declare namespace NodeConsole {
interface Console { log(message?: any, ...optionalParams: any[]): void; error(message?: any, ...optionalParams: any[]): void; table(data: unknown): void; }
}
interface Console extends NodeConsole.Console {}
declare var console: Console;`},
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2024"],"types":[]},"files":["main.ts","globals.d.ts"]}`},
	)
	if _, err := Load(paths[:1]); err != nil {
		t.Fatalf("prelude removed inherited project console signatures: %v", err)
	}
}
