package load

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestProjectOptionAttribution(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const items: number[] = [];
const first: number = items[0];
const point: { x?: number } = { x: undefined };
try { throw new Error('boom'); } catch (error) { const message = error.message; }
const ordinary: number = 'wrong';
`})
	report, err := AuditProjectOptions(context.Background(), paths[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(report.ProjectErrors) != 1 || !strings.Contains(report.ProjectErrors[0], "TS2322") || !strings.Contains(report.ProjectErrors[0], "main.ts:5:") {
		t.Fatalf("ordinary error lost or reclassified: %+v", report)
	}
	if len(report.Sites) != 3 {
		t.Fatalf("want three stricter sites, got %+v", report.Sites)
	}
	for index, want := range []struct {
		line   int
		option string
	}{{2, "noUncheckedIndexedAccess"}, {3, "exactOptionalPropertyTypes"}, {4, "useUnknownInCatchVariables"}} {
		site := report.Sites[index]
		if site.Line != want.line || !slices.Equal(site.Options, []string{want.option}) || site.File != paths[1] {
			t.Errorf("site %d: got %+v, want line %d, option %s", index, site, want.line, want.option)
		}
	}
}

func TestProjectOptionsPreserveInheritedLibAndStrictness(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"base.json", `{"compilerOptions":{"strict":true,"lib":["es2020","dom"],"noUncheckedIndexedAccess":true,"useUnknownInCatchVariables":false}}`},
		[2]string{"tsconfig.json", `{"extends":"./base.json","files":["main.ts"]}`},
		[2]string{"main.ts", `const element = document.body;
const values: number[] = [];
const value: number = values[0];
`})
	report, err := AuditProjectOptions(context.Background(), paths[1])
	if err != nil {
		t.Fatal(err)
	}
	if len(report.ProjectErrors) != 1 || !strings.Contains(report.ProjectErrors[0], "main.ts:3:") || len(report.Sites) != 0 {
		t.Fatalf("project lib or existing stricter option changed: %+v", report)
	}
}

func TestProjectOptionsReportConfigErrors(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"tsconfig.json", `{"compilerOptions":{"lib":["not-a-library"]},"files":["missing.ts"]}`})
	if _, err := AuditProjectOptions(context.Background(), paths[0]); err == nil {
		t.Fatal("invalid lib must stay an error")
	}
	if _, err := AuditProjectOptions(context.Background(), filepath.Join(filepath.Dir(paths[0]), "absent.json")); err == nil {
		t.Fatal("missing config must stay an error")
	}
}
