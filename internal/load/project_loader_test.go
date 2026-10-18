package load

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestProjectLoaderOptionAuditFailsClosed(t *testing.T) {
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
	if _, err := Load([]string{filepath.Join(dir, "main.ts")}); err == nil || !strings.Contains(err.Error(), "TS2322") {
		t.Fatalf("unchecked indexed read admitted: %v", err)
	}
}

func TestProjectLoaderOwnershipAndNoCheck(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, config, source, other, want string }{
		{"noCheck", `{"compilerOptions":{"noCheck":true},"files":["main.ts"]}`, "export const value: number = 1;", "", "noCheck"},
		{"mixed", `{"compilerOptions":{` + fixtureOptions + `},"files":["main.ts"]}`, "export const value: number = 1;", "export const proof: number = 1;", "mixed project"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{"tsconfig.json": strings.TrimSuffix(tc.config, "}") + `,"references":[{"path":"./dependency"}]}`, "main.ts": tc.source, "dependency/value.ts": "export const value = 1;", "dependency/tsconfig.json": referenceConfig(`["value.ts"]`, `[]`, "")}
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
		"main.ts":                  "console.log(7);\nexport const value: number = 1;\n",
		"other.ts":                 "export const other: number = 2;\n",
		"dependency/value.ts":      "export const dependency = 1;",
		"dependency/tsconfig.json": `{"compilerOptions":{"strict":true,"composite":true,"target":"es2024","module":"esnext","moduleResolution":"bundler","types":[],"noUncheckedIndexedAccess":true,"exactOptionalPropertyTypes":true},"files":["value.ts"]}`,
		"tsconfig.json":            `{"compilerOptions":{"strict":true,"composite":true,"target":"es2024","module":"esnext","moduleResolution":"bundler","types":[],"noUncheckedIndexedAccess":true,"exactOptionalPropertyTypes":true},"files":["main.ts","other.ts"],"references":[{"path":"./dependency"}]}`,
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
	for _, file := range program.CompilerProgram().GetSourceFiles() {
		if program.FileName(file) == filepath.Join(dir, "dependency/unused.ts") {
			return
		}
	}
	t.Fatal("node import discarded referenced source roots")
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
