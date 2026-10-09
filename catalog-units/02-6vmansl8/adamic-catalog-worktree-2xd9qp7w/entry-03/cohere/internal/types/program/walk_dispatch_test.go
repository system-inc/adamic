package program_test

import (
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// One worker keeps each rule's slot, the merged listener table and the file's state across every file it
// walks (#9jpmqm9), so nothing one file left there may reach the next. One worker walks all three files
// here, in order, so each rule meets a file after the one that changed its state:
//
//   - a rule that crashed on a.ts, in Run or in a listener, is heard again on b.ts and c.ts;
//   - a kind only a.ts listened for is not called on b.ts, which has that kind and listens for nothing;
//   - a.ts's suppression does not silence the finding at the same place in b.ts;
//   - each file's notes are its own, not a running total.
func TestOneWorkerHandsEveryFileItsOwnRuleState(t *testing.T) {
	t.Parallel()
	// The same layout in each file, so a finding in b.ts or c.ts sits where a.ts's directive covers.
	declarations := "export const first = 1;\nexport function second(): void {}\n"
	directory := writeProject(t, map[string]string{
		"tsconfig.json":       minimalConfig,
		"CohereSettings.json": `{"rules": {"test-run-crashes-on-a": "error", "test-listener-crashes-on-a": "error", "test-functions-only-in-a": "error", "test-notes": "error"}}`,
		"a.ts":                "// cohere-disable-next-line test-run-crashes-on-a -- covers a.ts alone\n" + declarations,
		"b.ts":                "// a plain comment, the length of a.ts's directive, nothing else at all\n" + declarations,
		"c.ts":                "// a plain comment, the length of a.ts's directive, nothing else at all\n" + declarations,
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory, SingleThreaded: true})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	if graph.Workers() != 1 {
		t.Fatalf("the walk has %d workers, so files need not meet each other's leftovers", graph.Workers())
	}
	graph.LintConfig, err = configuration.Load(filepath.Join(directory, "CohereSettings.json"))
	if err != nil {
		t.Fatalf("loading the config: %v", err)
	}

	isA := func(ctx rule.Context) bool { return filepath.Base(ctx.SourceFile.FileName().AsString()) == "a.ts" }
	reportsEachConstant := func(ctx rule.Context) rule.Listeners {
		return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
			ctx.ReportNode(node, rule.Message{Id: "sawConstant", Description: "saw a constant"})
		}}
	}
	runCrashes := rule.Rule{
		Name: "test-run-crashes-on-a",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			if isA(ctx) {
				panic("planted in Run")
			}
			return reportsEachConstant(ctx)
		},
	}
	listenerCrashes := rule.Rule{
		Name: "test-listener-crashes-on-a",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
				if isA(ctx) {
					panic("planted in a listener")
				}
				ctx.ReportNode(node, rule.Message{Id: "sawConstant", Description: "saw a constant"})
			}}
		},
	}
	functionsOnlyInA := rule.Rule{
		Name: "test-functions-only-in-a",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			if !isA(ctx) {
				return nil
			}
			return rule.Listeners{ast.KindFunctionDeclaration: func(node *ast.Node) {
				ctx.ReportNode(node, rule.Message{Id: "sawFunction", Description: "saw a function in " + filepath.Base(ctx.SourceFile.FileName().AsString())})
			}}
		},
	}
	notes := rule.Rule{
		Name: "test-notes",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) { ctx.Note("constant") }}
		},
	}

	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{runCrashes, listenerCrashes, functionsOnlyInA, notes})
	if err != nil {
		t.Fatalf("walking: %v", err)
	}

	reported := map[string][]string{}
	for _, diagnostic := range result.Diagnostics {
		file := filepath.Base(diagnostic.SourceFile.FileName().AsString())
		reported[file] = append(reported[file], diagnostic.RuleName+": "+diagnostic.Message.Description)
	}
	want := map[string][]string{
		"a.ts": {"test-functions-only-in-a: saw a function in a.ts"},
		"b.ts": {"test-listener-crashes-on-a: saw a constant", "test-run-crashes-on-a: saw a constant"},
		"c.ts": {"test-listener-crashes-on-a: saw a constant", "test-run-crashes-on-a: saw a constant"},
	}
	if !reflect.DeepEqual(reported, want) {
		t.Errorf("the findings by file are %v, want %v", reported, want)
	}

	crashed := []string{}
	for _, crash := range result.Coverage.RulesCrashed {
		crashed = append(crashed, crash.RuleName+" on "+filepath.Base(crash.FileName))
	}
	if want := []string{"test-run-crashes-on-a on a.ts", "test-listener-crashes-on-a on a.ts"}; !sameMembers(crashed, want) {
		t.Errorf("the crashes are %v, want %v", crashed, want)
	}

	wantNotes := map[string]program.RuleNotes{}
	for _, file := range []string{"a.ts", "b.ts", "c.ts"} {
		wantNotes[filepath.Join(directory, file)] = program.RuleNotes{"test-notes": {"constant": 1}}
	}
	if !reflect.DeepEqual(result.Notes, wantNotes) {
		t.Errorf("the notes are %v, want %v", result.Notes, wantNotes)
	}
}

// A Context is good only while its file is dispatched. A rule that kept its Report and called it after the
// walk would file a finding against whatever file the worker held, so the call panics, naming the rule,
// instead of filing one anywhere.
func TestAReportKeptPastItsFilePanicsNamingTheRule(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json": minimalConfig,
		"main.ts":       "export const kept = 1;\n",
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}

	var kept func(rule.Diagnostic)
	keeper := rule.Rule{
		Name: "test-keeps-its-report",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			kept = ctx.Report
			return nil
		},
	}
	if _, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{keeper}); err != nil {
		t.Fatalf("walking: %v", err)
	}
	if kept == nil {
		t.Fatal("the rule never ran, so there is no Report to call late")
	}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("a Report called after its file's dispatch filed a finding instead of panicking")
		}
		if message, _ := recovered.(string); !strings.Contains(message, "test-keeps-its-report") {
			t.Errorf("the panic does not name the rule: %v", recovered)
		}
	}()
	kept(rule.Diagnostic{Message: rule.Message{Id: "late", Description: "too late"}})
}

func sameMembers(got []string, want []string) bool {
	counts := map[string]int{}
	for _, value := range got {
		counts[value]++
	}
	for _, value := range want {
		counts[value]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}
