package load

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func referenceConfig(files string, references string, extra string) string {
	return `{"compilerOptions":{` + fixtureOptions + `,"composite":true,"rootDir":".","outDir":"../out"` + extra + `},"files":` + files + `,"references":` + references + `}`
}

func TestProjectReferencesTransitive(t *testing.T) {
	t.Parallel()
	dir := projectFixture(t, map[string]string{
		"app/main.ts":     `import { value } from "../middle/value.js"; console.log(value);`,
		"middle/value.ts": `export const value: number = 7;`,
		"leaf/base.ts":    `export const base: number = 5;`,
		// This leaf is unimported. Imports alone must not satisfy reference traversal.
		"leaf/unused.ts":       `export const unused: number = 1;`,
		"app/tsconfig.json":    referenceConfig(`["main.ts"]`, `[{"path":"../middle"}]`, ""),
		"middle/tsconfig.json": referenceConfig(`["value.ts"]`, `[{"path":"../leaf/tsconfig.json"}]`, ""),
		"leaf/tsconfig.json":   referenceConfig(`["base.ts","unused.ts"]`, `[]`, ""),
	})
	path := filepath.Join(dir, "app/main.ts")
	program, err := Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, source := range program.CompilerProgram().GetSourceFiles() {
		if program.FileName(source) == filepath.Join(dir, "leaf/unused.ts") {
			found = true
		}
	}
	if !found {
		t.Fatal("transitive configured source leaf/unused.ts was dropped")
	}
	if err := os.WriteFile(filepath.Join(dir, "leaf/unused.ts"), []byte(`export const unused: number = "bad";`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load([]string{path}); err == nil || !strings.Contains(err.Error(), "leaf/unused.ts:1:14: error TS2322") {
		t.Fatalf("want unimported transitive type error, got %v", err)
	}
}

func TestProjectReferencesDiamond(t *testing.T) {
	t.Parallel()
	dir := projectFixture(t, map[string]string{
		"app/main.ts":         `export const value: number = 7;`,
		"left/value.ts":       `export const left: number = 1;`,
		"right/value.ts":      `export const right: number = 2;`,
		"leaf/value.ts":       `export const leaf: number = 4;`,
		"app/tsconfig.json":   referenceConfig(`["main.ts"]`, `[{"path":"../left"},{"path":"../right"}]`, ""),
		"left/tsconfig.json":  referenceConfig(`["value.ts"]`, `[{"path":"../leaf"}]`, ""),
		"right/tsconfig.json": referenceConfig(`["value.ts"]`, `[{"path":"../leaf"}]`, ""),
		"leaf/tsconfig.json":  referenceConfig(`["value.ts"]`, `[]`, ""),
	})
	program, err := Load([]string{filepath.Join(dir, "app/main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, root := range program.CompilerProgram().CommandLine().FileNames() {
		if seen[root.AsString()] {
			t.Fatalf("duplicate checking root %s", root)
		}
		seen[root.AsString()] = true
	}
	if !seen[filepath.Join(dir, "leaf/value.ts")] {
		t.Fatal("missing diamond leaf root")
	}
	// Every branch must remain compatible with the shared option profile.
	config := referenceConfig(`["value.ts"]`, `[{"path":"../leaf"}]`, `,"noImplicitAny":false`)
	if err := os.WriteFile(filepath.Join(dir, "right/tsconfig.json"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load([]string{filepath.Join(dir, "app/main.ts")}); err == nil || !strings.Contains(err.Error(), "noImplicitAny") {
		t.Fatalf("want conflicting second path, got %v", err)
	}
}

func TestProjectReferencesConflictingOptions(t *testing.T) {
	t.Parallel()
	for _, option := range []string{"strictNullChecks", "strictFunctionTypes", "strictBindCallApply", "strictPropertyInitialization", "strictBuiltinIteratorReturn", "useUnknownInCatchVariables", "noImplicitAny", "noImplicitThis", "noUncheckedIndexedAccess", "exactOptionalPropertyTypes", "target", "module", "types", "noUncheckedSideEffectImports", "forceConsistentCasingInFileNames", "deduplicatePackages"} {
		t.Run(option, func(t *testing.T) {
			extra := fmt.Sprintf(`,"%s":false`, option)
			if option == "noUncheckedIndexedAccess" || option == "exactOptionalPropertyTypes" {
				extra = fmt.Sprintf(`,"%s":true`, option)
			}
			if option == "target" {
				extra = `,"target":"es2024"`
			}
			if option == "module" {
				extra = `,"module":"commonjs"`
			}
			if option == "types" {
				extra = `,"types":["node"]`
			}
			dir := projectFixture(t, map[string]string{
				"app/main.ts":              `export const value: number = 7;`,
				"dependency/value.ts":      `export const value: number = 7;`,
				"app/tsconfig.json":        referenceConfig(`["main.ts"]`, `[{"path":"../dependency"}]`, ""),
				"dependency/tsconfig.json": referenceConfig(`["value.ts"]`, `[]`, extra),
			})
			want := fmt.Sprintf("load: project reference options conflict: %s references %s: compiler option %s differs; separate checker ownership is required", filepath.Join(dir, "app/tsconfig.json"), filepath.Join(dir, "dependency/tsconfig.json"), option)
			if _, err := Load([]string{filepath.Join(dir, "app/main.ts")}); err == nil || err.Error() != want {
				t.Fatalf("want %q, got %v", want, err)
			}
		})
	}
}

func TestProjectReferencesAudit(t *testing.T) {
	t.Parallel()
	dir := projectFixture(t, map[string]string{
		"app/main.ts":              `export const value: number = 7;`,
		"dependency/value.ts":      `const values: number[] = []; export const value: number = values[0];`,
		"app/tsconfig.json":        referenceConfig(`["main.ts"]`, `[{"path":"../dependency"}]`, ""),
		"dependency/tsconfig.json": referenceConfig(`["value.ts"]`, `[]`, ""),
	})
	report, err := AuditProjectOptions(context.Background(), filepath.Join(dir, "app/tsconfig.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.ProjectErrors) != 0 || len(report.Sites) != 1 {
		t.Fatalf("want one referenced option site without TS6305, got %+v", report)
	}
	if _, err := Load([]string{filepath.Join(dir, "app/main.ts")}); err == nil || !strings.Contains(err.Error(), "dependency/value.ts:1:43: error TS2322") {
		t.Fatalf("referenced option site admitted: %v", err)
	}
}

func TestProjectReferencesCompatibleOptions(t *testing.T) {
	t.Parallel()
	dependency := strings.Replace(referenceConfig(`["value.ts"]`, `[]`, `,"strictNullChecks":true,"moduleDetection":"auto"`), `"composite":true`, `"composite":false`, 1)
	dir := projectFixture(t, map[string]string{
		"app/main.ts":              `import { value } from "../dependency/value.js"; export const result: number = value;`,
		"dependency/value.ts":      `export const value: number = 7;`,
		"app/tsconfig.json":        referenceConfig(`["main.ts"]`, `[{"path":"../dependency"}]`, `,"outDir":"../different-output"`),
		"dependency/tsconfig.json": dependency,
	})
	if _, err := Load([]string{filepath.Join(dir, "app/main.ts")}); err != nil {
		t.Fatal(err)
	}
}

func TestProjectReferencesCycle(t *testing.T) {
	t.Parallel()
	dir := projectFixture(t, map[string]string{
		"app/main.ts":              `export const value: number = 7;`,
		"dependency/value.ts":      `export const value: number = 7;`,
		"app/tsconfig.json":        referenceConfig(`["main.ts"]`, `[{"path":"../dependency"}]`, ""),
		"dependency/tsconfig.json": referenceConfig(`["value.ts"]`, `[{"path":"../app"}]`, ""),
	})
	if _, err := Load([]string{filepath.Join(dir, "app/main.ts")}); err == nil || !strings.Contains(err.Error(), "circular project reference") {
		t.Fatalf("want explicit circular reference refusal, got %v", err)
	}
}

func TestProjectReferencesSharedPrelude(t *testing.T) {
	t.Parallel()
	config := referenceConfig(`["main.ts","../prelude.d.ts"]`, `[{"path":"../dependency"}]`, `,"lib":["es2020"]`)
	dir := projectFixture(t, map[string]string{
		"app/main.ts":              "import { value } from \"../dependency/value.js\"; console.log(`${value}`);",
		"dependency/value.ts":      `export const value: number = 7;`,
		"prelude.d.ts":             prelude,
		"app/tsconfig.json":        config,
		"dependency/tsconfig.json": referenceConfig(`["value.ts","../prelude.d.ts"]`, `[]`, `,"lib":["es2020"]`),
	})
	program, err := Load([]string{filepath.Join(dir, "app/main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	if hasHostConsole(program.CompilerProgram().GetSourceFiles()) {
		t.Fatal("physical prelude console replaced the embedded declaration used by lowering")
	}
}
