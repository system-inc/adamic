package load

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionProjectSitesRetainUnconvertedErrors(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const items: number[] = [];
const first: number = items[0];
const point: { x?: number } = { x: undefined };
try { throw new Error('boom'); } catch (error) { const message = error.message; }
const ordinary: number = 'wrong';
`})
	_, err := Load(paths[1:])
	var rejected *CheckError
	if !errors.As(err, &rejected) || len(rejected.OptionSites) != 3 {
		t.Fatalf("unconverted sites must remain errors and be recorded: %v", err)
	}
	if !strings.Contains(err.Error(), "main.ts:5:") || len(rejected.ScheduledOptionSites) != 3 {
		t.Fatalf("ordinary or unconverted error disappeared: %v", err)
	}
	if strings.Contains(err.Error(), "main.ts:2:") || strings.Contains(err.Error(), "main.ts:3:") || strings.Contains(err.Error(), "main.ts:4:") {
		t.Fatalf("indexed row was not deferred to its lowering check: %v", err)
	}
}

func TestProductionProjectPreservesCatchTypeAndLib(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const value = BigInt(1);
try { throw new Error('boom'); } catch (error) { const held = error; }
`})
	program, err := Load(paths[1:])
	if err != nil {
		t.Fatal(err)
	}
	foundCatch := false
	for _, declaration := range program.Declarations(context.Background()) {
		if declaration.Name == "held" {
			foundCatch = true
			if declaration.Type != "any" {
				t.Fatalf("project catch type replaced: %+v", declaration)
			}
		}
	}
	if !foundCatch {
		t.Fatal("catch witness was not checked")
	}
	oldLib := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2015"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const value = BigInt(1);`})
	if _, err := Load(oldLib[1:]); err == nil || !strings.Contains(err.Error(), "BigInt") {
		t.Fatalf("project's old lib was overwritten: %v", err)
	}
}

func TestProductionAdamicKeepsStrictOptions(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":false},"files":["main.a"]}`},
		[2]string{"main.a", `const values: number[] = []; const value: number = values[0];`})
	if _, err := Load(paths[1:]); err == nil || !strings.Contains(err.Error(), "undefined") {
		t.Fatalf(".a lost its proof obligation: %v", err)
	}
}

func TestProductionMixedOwnershipRefused(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `import './proof.a';`},
		[2]string{"proof.a", `export const values: number[] = []; const value: number = values[0];`})
	if _, err := Load(paths[1:2]); err == nil || !strings.Contains(err.Error(), "mixed .a") {
		t.Fatalf("mixed import silently received relaxed options: %v", err)
	}
}

func TestProductionCompositeKeepsProjectRoots(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2020"],"composite":true,"module":"esnext","noEmit":true},"files":["main.ts","other.ts"]}`},
		[2]string{"main.ts", `import { value } from './other'; export const held = value;`},
		[2]string{"other.ts", `export const value = 1;`})
	if _, err := Load(paths[1:2]); err != nil {
		t.Fatalf("composite project root list truncated: %v", err)
	}
}

func TestProductionProjectOverlaySites(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `export const value = 1;`})
	program, err := LoadOverlay(paths[1:], map[string]string{paths[1]: `const items: number[] = []; const first: number = items[0];`})
	if err != nil {
		t.Fatal(err)
	}
	rows := program.OptionDispositions()
	if len(rows) != 1 || rows[0].Kind != "indexed-presence" || rows[0].State != OptionCheckScheduled || rows[0].Site.Line != 1 {
		t.Fatalf("overlay read needs its own recorded contract: %+v", rows)
	}
	_, err = LoadOverlay(paths[1:], map[string]string{paths[1]: `const wrong: number = 'wrong';`})
	if err == nil || !strings.Contains(err.Error(), "not assignable") {
		t.Fatalf("ordinary overlay error was lost: %v", err)
	}
}

func TestProductionLiteralContractIsPending(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"exactOptionalPropertyTypes":false,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const point: { x?: number } = { x: undefined };`})
	program, err := Load(paths[1:])
	if err != nil {
		t.Fatal(err)
	}
	if len(program.optionalLiterals) != 1 || len(program.ExplainedOptionalChecks()) != 0 {
		t.Fatal("a pending literal contract was lost or counted as emitted")
	}
	for _, sites := range program.optionalLiterals {
		if len(sites) != 1 || sites[0].Code != 2375 {
			t.Fatalf("wrong pending site: %v", sites)
		}
	}
}

func TestProductionNullableCallbackContractStaysError(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"exactOptionalPropertyTypes":false,"lib":["ES2024"],"types":[],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `interface Source { slot?: number | undefined; }
interface Target { slot?: number; }
function choose(): (() => Source) | undefined { return undefined; }
const callback: (() => Target) | undefined = choose();`})
	_, err := Load(paths[1:])
	var rejected *CheckError
	if !errors.As(err, &rejected) || len(rejected.OptionSites) != 1 || len(rejected.ScheduledOptionSites) != 0 {
		t.Fatalf("nullable callback must remain an explicit diagnostic, not a present-callback guard: %v", err)
	}
}

func projectFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, text := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const fixtureOptions = `"strict":true,"module":"esnext","moduleResolution":"bundler","target":"es2020","types":[],"skipLibCheck":true`

func TestProjectLoaderReferenceSources(t *testing.T) {
	t.Parallel()
	dir := projectFixture(t, map[string]string{
		"app/main.ts":              "import { value } from \"../dependency/value.js\";\nconsole.log(value);\n",
		"dependency/value.ts":      "export const value: number = 7;\n",
		"app/tsconfig.json":        `{ "compilerOptions": {` + fixtureOptions + `,"composite":true,"rootDir":".","outDir":"../out/app"},"files":["main.ts"],"references":[{"path":"../dependency"}] }`,
		"dependency/tsconfig.json": `{ "compilerOptions": {` + fixtureOptions + `,"composite":true,"rootDir":".","outDir":"../out/dependency"},"files":["value.ts"] }`,
	})
	program, err := Load([]string{filepath.Join(dir, "app/main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Files()) != 1 {
		t.Fatalf("project checking roots became execution entries: %d", len(program.Files()))
	}
	found := 0
	for _, file := range program.CompilerProgram().GetSourceFiles() {
		if program.FileName(file) == filepath.Join(dir, "dependency/value.ts") {
			if file.IsDeclarationFile {
				t.Fatal("reference loaded as a declaration")
			}
			found++
		}
	}
	if found != 1 {
		t.Fatalf("want dependency implementation once, got %d", found)
	}
	// Built outputs must not replace source implementations on later loads.
	if err := os.MkdirAll(filepath.Join(dir, "out/dependency"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "out/dependency/value.d.ts"), []byte("export declare const value: string;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load([]string{filepath.Join(dir, "app/main.ts")}); err != nil {
		t.Fatal(err)
	}
}

func TestProjectLoaderOptionAuditSchedulesCheck(t *testing.T) {
	t.Parallel()
	dir := projectFixture(t, map[string]string{
		"main.ts":       "const values: number[] = [];\nexport const value: number = values[0];\n",
		"tsconfig.json": `{"compilerOptions":{` + fixtureOptions + `},"files":["main.ts"]}`,
	})
	report, err := AuditProjectOptions(context.Background(), filepath.Join(dir, "tsconfig.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.ProjectErrors) != 0 || len(report.Sites) != 1 || len(report.Sites[0].Options) != 1 || report.Sites[0].Options[0] != "noUncheckedIndexedAccess" {
		t.Fatalf("unexpected audit: %+v", report)
	}
	program, err := Load([]string{filepath.Join(dir, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	if len(program.OptionSites()) != 1 || len(program.OptionDispositions()) != 1 || program.OptionDispositions()[0].State != OptionCheckScheduled {
		t.Fatal("indexed option obligation was not scheduled")
	}
}

func TestProjectLoaderOwnershipAndNoCheck(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, config, source, other, want string }{
		{"noCheck", `{"compilerOptions":{"noCheck":true},"files":["main.ts"]}`, "export const value: number = 1;", "", "noCheck"},
		{"mixed", `{"compilerOptions":{` + fixtureOptions + `},"files":["main.ts"]}`, "export const value: number = 1;", "export const proof: number = 1;", "mixed project"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{"tsconfig.json": tc.config, "main.ts": tc.source}
			if tc.other != "" {
				files["proof.a"] = tc.other
			}
			dir := projectFixture(t, files)
			roots := []string{filepath.Join(dir, "main.ts")}
			if tc.other != "" {
				roots = append(roots, filepath.Join(dir, "proof.a"))
			}
			if _, err := Load(roots); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %s, got %v", tc.want, err)
			}
		})
	}
}

func TestProjectLoaderHostConsoleAndCompositeEntries(t *testing.T) {
	t.Parallel()
	dir := projectFixture(t, map[string]string{
		"main.ts":       "console.log(7);\nexport const value: number = 1;\n",
		"other.ts":      "export const other: number = 2;\n",
		"tsconfig.json": `{"compilerOptions":{"strict":true,"composite":true,"target":"es2024","module":"esnext","moduleResolution":"bundler","types":[],"noUncheckedIndexedAccess":true,"exactOptionalPropertyTypes":true},"files":["main.ts","other.ts"]}`,
	})
	p, err := Load([]string{filepath.Join(dir, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Files()) != 1 {
		t.Fatalf("want one entry, got %d", len(p.Files()))
	}
	if err := os.WriteFile(filepath.Join(dir, "other.ts"), []byte("export const other: number = \"bad\";\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load([]string{filepath.Join(dir, "main.ts")}); err == nil || !strings.Contains(err.Error(), "other.ts:1:14: error TS2322") {
		t.Fatalf("configured composite source silently dropped: %v", err)
	}
}

func TestProjectLoaderNodeImportsKeepReferenceRoots(t *testing.T) {
	t.Parallel()
	dir := projectFixture(t, map[string]string{
		"app/main.ts":              `import { existsSync } from 'node:fs'; export const value: boolean = existsSync('x');`,
		"app/host.d.ts":            `declare module 'node:fs' { export function existsSync(path: string): boolean; }`,
		"dependency/unused.ts":     `export const unused: number = 1;`,
		"app/tsconfig.json":        referenceConfig(`["main.ts","host.d.ts"]`, `[{"path":"../dependency"}]`, ""),
		"dependency/tsconfig.json": referenceConfig(`["unused.ts"]`, `[]`, ""),
	})
	program, err := Load([]string{filepath.Join(dir, "app/main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, file := range program.CompilerProgram().GetSourceFiles() {
		if IsNodeLibrary(file) {
			t.Fatal("standalone Node declarations replaced project-owned declarations")
		}
		if program.FileName(file) == filepath.Join(dir, "dependency/unused.ts") {
			found = true
		}
	}
	if !found {
		t.Fatal("node import discarded referenced source roots")
	}
}

func TestProjectLoaderStandaloneNodeKeepsPrelude(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(file, []byte(`import { existsSync } from 'node:fs'; import { readTextFile } from 'adamic'; const present: boolean = existsSync('x'); const text = readTextFile('x'); console.log('checked');`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load([]string{file}); err != nil {
		t.Fatalf("standalone Node rebuild lost prelude: %v", err)
	}
}
