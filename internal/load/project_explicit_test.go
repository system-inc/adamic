package load

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestProjectRefusesWeakenedOptions(t *testing.T) {
	t.Parallel()
	options := map[string]string{
		"strict": "false", "noImplicitAny": "false", "noImplicitThis": "false", "strictNullChecks": "false",
		"strictFunctionTypes": "false", "strictBindCallApply": "false", "strictPropertyInitialization": "false",
		"strictBuiltinIteratorReturn": "false", "useUnknownInCatchVariables": "false", "alwaysStrict": "false",
		"noUncheckedIndexedAccess": "false", "exactOptionalPropertyTypes": "false", "noImplicitReturns": "false",
		"noFallthroughCasesInSwitch": "false", "erasableSyntaxOnly": "false", "verbatimModuleSyntax": "false",
		"allowImportingTsExtensions": "false", "noEmit": "false", "module": `"commonjs"`, "moduleDetection": `"legacy"`,
		"moduleResolution": `"node16"`, "target": `"es2020"`, "useDefineForClassFields": "false",
		"noCheck": "true", "skipLibCheck": "true", "skipDefaultLibCheck": "true", "noResolve": "true", "noLib": "true", "libReplacement": "true",
	}
	for option, value := range options {
		t.Run(option, func(t *testing.T) {
			t.Parallel()
			paths := writeProgram(t, [2]string{"main.a", "export const value = 1;"},
				[2]string{"base.json", `{"compilerOptions":{"` + option + `":` + value + `}}`},
				[2]string{"tsconfig.json", `{"extends":"./base.json","files":["main.a"]}`})
			_, err := LoadProject(paths[2])
			if err == nil || !strings.Contains(err.Error(), option) || !strings.Contains(err.Error(), "Adamic requires") {
				t.Fatalf("weakened %s: got %v", option, err)
			}
		})
	}
}

func TestProjectGlobOrderAndAdditionalStrictness(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"z.a", "export const z = 1;"}, [2]string{"b.a", "export const b = 2;"},
		[2]string{"a.a", "export const a = 3;"}, [2]string{"skip.a", "export const skip: number = 'bad';"},
		[2]string{"base.json", `{"compilerOptions":{"noUnusedLocals":true},"include":["*.a"],"exclude":["skip.a"]}`},
		[2]string{"tsconfig.json", `{"extends":"./base.json","files":["z.a"]}`})
	program, err := LoadProject(paths[5])
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, file := range program.Files() {
		names = append(names, filepath.Base(program.FileName(file)))
	}
	if !reflect.DeepEqual(names, []string{"z.a", "a.a", "b.a"}) {
		t.Fatalf("root order: %v", names)
	}
	// Project-only extra strictness is effective, not overwritten by the required baseline.
	if err := os.WriteFile(paths[0], []byte("const unused = 1; export const z = 1;"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = LoadProject(paths[5])
	if err == nil || !strings.Contains(err.Error(), "TS6133") {
		t.Fatalf("noUnusedLocals lost: %v", err)
	}
}

func TestProjectRetainsSoundPreludeWithHostTypes(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.a", `import { panic } from 'adamic'; const text = JSON.stringify(undefined) ?? panic('missing');`},
		[2]string{"host.d.ts", `declare var console: { log(...values: unknown[]): void };`},
		[2]string{"tsconfig.json", `{"compilerOptions":{"types":[],"lib":["es2024"]},"files":["main.a","host.d.ts"]}`})
	program, err := LoadProject(paths[2])
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Files()) != 1 {
		t.Fatalf("declaration root became executable: %d", len(program.Files()))
	}
	for _, declaration := range program.Declarations(context.Background()) {
		if declaration.Name == "text" && declaration.Type != "string" {
			t.Errorf("panic signature: %s", declaration.Type)
		}
	}
	if err := os.WriteFile(paths[0], []byte("const text: string = JSON.stringify(undefined);"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = LoadProject(paths[2])
	if err == nil || !strings.Contains(err.Error(), "TS2322") {
		t.Fatalf("sound JSON signature lost: %v", err)
	}
}

func TestProjectConfigErrorsAreReported(t *testing.T) {
	t.Parallel()
	for _, config := range []string{`{"extends":"./absent.json","files":["main.a"]}`, `{"compilerOptions":{"strict":"yes"},"files":["main.a"]}`} {
		paths := writeProgram(t, [2]string{"main.a", "const answer = 1;"}, [2]string{"tsconfig.json", config})
		_, err := LoadProject(paths[1])
		if err == nil || !strings.Contains(err.Error(), "error TS") {
			t.Fatalf("invalid config accepted: %v", err)
		}
	}
}

func TestProjectRefusesAmbiguousAdamicAliases(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.a", "const value = 1;"}, [2]string{"main.a.ts", "const value = 2;"},
		[2]string{"tsconfig.json", `{"files":["main.a"]}`})
	_, err := LoadProject(paths[2])
	if err == nil || !strings.Contains(err.Error(), "both exist") {
		t.Fatalf("ambiguous .a root: %v", err)
	}
}

func TestProjectAmbientDiscoveryRetainsPrelude(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.a", "console.log('ok');"},
		[2]string{"host.d.ts", `export {}; declare global { var console: { log(...values: unknown[]): void }; }`},
		[2]string{"tsconfig.json", `{"compilerOptions":{"types":[]},"files":["main.a","host.d.ts"]}`})
	if _, err := LoadProject(paths[2]); err != nil {
		t.Fatalf("global augmentation: %v", err)
	}
	if err := os.WriteFile(paths[1], []byte("export {}; declare namespace Local { const console: string; }"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProject(paths[2]); err != nil {
		t.Fatalf("namespace stole prelude console: %v", err)
	}
}

func TestProjectCannotDisableCheckingJavaScriptDependencies(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.a", "const value = 1;"},
		[2]string{"tsconfig.json", `{"compilerOptions":{"allowJs":true,"checkJs":false},"files":["main.a"]}`})
	_, err := LoadProject(paths[1])
	if err == nil || !strings.Contains(err.Error(), "checkJs") || !strings.Contains(err.Error(), "Adamic requires") {
		t.Fatalf("unchecked JavaScript dependencies accepted: %v", err)
	}
}

func TestDirectDeclarationInputStillReportsTypes(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"host.d.ts", "declare const answer: number;"})
	program, err := Load(paths)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, declaration := range program.Declarations(context.Background()) {
		if declaration.Name == "answer" && declaration.Type == "number" {
			found = true
		}
	}
	if !found {
		t.Fatal("direct declaration root disappeared from type queries")
	}
}
