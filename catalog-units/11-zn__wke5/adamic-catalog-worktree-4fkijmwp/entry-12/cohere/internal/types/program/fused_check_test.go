package program_test

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// A walk that type-checks each file it visits (program.FusedCheck) reports the diagnostics the compiler's
// whole-program check reports, part for part: the files the walk checked, the declaration it never visited,
// an unused @ts-expect-error, and an error whose related information sits in another file.
func TestAFusedCheckReportsWhatTheWholeProgramCheckReports(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"tsconfig.json": minimalConfig,
		"shared.ts":     "export function take(value: string): string {\n    return value;\n}\nexport interface Shape { a: number }\n",
		"global.d.ts":   "declare const declared: { broken: UnknownName };\n",
	}
	// Enough files that every checker has some, each with its own error reaching another file's types.
	for index := range 40 {
		files[fmt.Sprintf("f%02d.ts", index)] = fmt.Sprintf("import { take, type Shape } from './shared';\n"+
			"export const call%d = take(%d);\n"+
			"export const shape%d: Shape = { a: 1, b: %d };\n", index, index, index, index)
	}
	files["unused.ts"] = "// @ts-expect-error nothing here is an error\nexport const fine = 1;\n"
	// The compiler lists an imported file before its importer, so these two sit in the program against the
	// order of their names, and only sorting the assembled diagnostics puts them back as the whole check has them.
	files["a-importer.ts"] = "import { late } from './z-imported';\nexport const early: number = late;\n"
	files["z-imported.ts"] = "export const late: string = 1;\n"

	whole := buildProgram(t, files)
	expected := whole.AllDiagnosticParts(context.Background())

	fused := buildProgram(t, files)
	fused.FusedCheck = program.NewFusedCheck()
	if _, err := fused.Walk(context.Background(), fused.ProjectFiles(), []rule.Rule{typeAwareRule(t)}); err != nil {
		t.Fatalf("walk: %v", err)
	}
	if checked := fused.FusedCheck.Checked(); checked != len(fused.ProjectFiles()) {
		t.Fatalf("the walk checked %d of %d project files, so the parts below came from the remainder, not the walk",
			checked, len(fused.ProjectFiles()))
	}
	actual := fused.FusedCheck.DiagnosticParts(context.Background(), fused)

	if len(expected.Semantic) < 80 {
		t.Fatalf("the whole check found %d semantic diagnostics, too few for a match to prove anything", len(expected.Semantic))
	}
	for _, part := range []struct {
		name             string
		expected, actual []*ast.Diagnostic
	}{
		{"syntactic", expected.Syntactic, actual.Syntactic},
		{"bind", expected.Bind, actual.Bind},
		{"semantic", expected.Semantic, actual.Semantic},
	} {
		if !slices.Equal(renderedFully(part.expected), renderedFully(part.actual)) {
			t.Errorf("%s diagnostics differ\n--- whole program\n%v\n--- fused\n%v", part.name,
				renderedFully(part.expected), renderedFully(part.actual))
		}
	}
}

// A file that does not parse stops a fused check at its syntactic diagnostics, as it stops the whole check.
func TestAFusedCheckStopsAtASyntaxErrorAsTheWholeCheckDoes(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"tsconfig.json": minimalConfig,
		"broken.ts":     "export const value = ;\n",
		"typed.ts":      "export const wrong: number = 'text';\n",
	}
	expected := buildProgram(t, files).AllDiagnosticParts(context.Background())
	fused := buildProgram(t, files)
	fused.FusedCheck = program.NewFusedCheck()
	if _, err := fused.Walk(context.Background(), fused.ProjectFiles(), []rule.Rule{typeAwareRule(t)}); err != nil {
		t.Fatalf("walk: %v", err)
	}
	actual := fused.FusedCheck.DiagnosticParts(context.Background(), fused)
	if len(expected.Syntactic) == 0 {
		t.Fatal("the fixture has no syntax error, so this test proves nothing")
	}
	if !slices.Equal(renderedFully(expected.All()), renderedFully(actual.All())) {
		t.Fatalf("got %v, want %v", renderedFully(actual.All()), renderedFully(expected.All()))
	}
}

func buildProgram(t *testing.T, files map[string]string) *program.Graph {
	t.Helper()
	root := writeProject(t, files)
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json")})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return graph
}

// renderedFully is messagesOf with the message and the related information too, so a diagnostic rebuilt
// from another checker's state reads as different even where its code and position agree.
func renderedFully(diagnostics []*ast.Diagnostic) []string {
	rendered := make([]string, 0, len(diagnostics))
	for index, diagnostic := range diagnostics {
		related := ""
		for _, information := range diagnostic.RelatedInformation() {
			name := ""
			if file := information.File(); file != nil {
				name = filepath.Base(file.FileName().AsString())
			}
			related += fmt.Sprintf(" [%s:%d:TS%d]", name, information.Pos(), information.Code())
		}
		rendered = append(rendered, fmt.Sprintf("%s %s %v%s", messagesOf(diagnostics[index : index+1])[0],
			diagnostic.MessageKey(), diagnostic.MessageArgs(), related))
	}
	return rendered
}
