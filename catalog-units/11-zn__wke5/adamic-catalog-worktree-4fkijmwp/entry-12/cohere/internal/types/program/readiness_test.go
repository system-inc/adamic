package program_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// Readiness measures every file against the set whatever the chain enables (#drbrp8c). Three set rules
// and one other, over two files:
//
//   - test-set-enabled is on in the chain: its findings report as ever, and count, suppressed ones too,
//     since Adamic reads no disable comment;
//   - test-set-off is off in the chain: it runs measure-only, so it counts and reports nothing, and its
//     fix never reaches anything that could apply it;
//   - test-set-skips skips b.ts for a missing precondition, which leaves b.ts unmeasured, not ready;
//   - test-other is no set rule: it reports and is never counted.
func TestReadinessCountsEverySetRuleAndReportsOnlyTheChainsOwn(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json":       minimalConfig,
		"CohereSettings.json": `{"rules": {"test-set-enabled": "error", "test-set-off": "off", "test-set-skips": "error", "test-other": "error"}}`,
		"a.ts":                "// cohere-disable-next-line test-set-enabled -- suppressed, still counted\nexport const first = 1;\nexport const second = 2;\n",
		"b.ts":                "export const third = 3;\n",
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory, SingleThreaded: true})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	graph.LintConfig, err = configuration.Load(filepath.Join(directory, "CohereSettings.json"))
	if err != nil {
		t.Fatalf("loading the config: %v", err)
	}
	graph.Readiness = &program.Readiness{Options: map[string][]json.RawMessage{
		"test-set-enabled": nil, "test-set-off": nil, "test-set-skips": nil,
	}}

	eachConstant := func(id string, fix bool) func(ctx rule.Context, options any) rule.Listeners {
		return func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
				message := rule.Message{Id: id, Description: id}
				if fix {
					ctx.ReportNodeWithFixes(node, message, ctx.RemoveNode(node))
					return
				}
				ctx.ReportNode(node, message)
			}}
		}
	}
	rules := []rule.Rule{
		{Name: "test-other", Run: eachConstant("other", false)},
		{Name: "test-set-enabled", Run: eachConstant("enabled", false)},
		{Name: "test-set-off", Run: eachConstant("off", true)},
		{Name: "test-set-skips", Run: func(ctx rule.Context, options any) rule.Listeners {
			if filepath.Base(ctx.SourceFile.FileName().AsString()) == "b.ts" {
				ctx.Skip("b.ts lacks what this rule needs")
				return nil
			}
			return nil
		}, NoListener: rule.NoListenerAnswersInRun},
	}
	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), rules)
	if err != nil {
		t.Fatalf("walking: %v", err)
	}

	reported := map[string]int{}
	for _, diagnostic := range result.Diagnostics {
		reported[diagnostic.RuleName]++
		if diagnostic.RuleName == "test-set-off" {
			t.Errorf("a measure-only rule reported %q, which a reader would see and a fix phase could apply", diagnostic.Message.Id)
		}
	}
	// a.ts's first constant is suppressed for test-set-enabled, so the chain sees one of a.ts's two and b.ts's one.
	if want := map[string]int{"test-other": 3, "test-set-enabled": 2}; !reflect.DeepEqual(reported, want) {
		t.Errorf("reported %v, want %v", reported, want)
	}

	records := map[string]*program.AdamicRecord{}
	for fileName, record := range result.Adamic {
		records[filepath.Base(fileName)] = record
	}
	want := map[string]*program.AdamicRecord{
		"a.ts": {Counts: []program.AdamicCount{{Rule: "test-set-enabled", Findings: 2}, {Rule: "test-set-off", Findings: 2}, {Rule: "test-set-skips", Findings: 0}}},
		"b.ts": {Counts: []program.AdamicCount{{Rule: "test-set-enabled", Findings: 1}, {Rule: "test-set-off", Findings: 1}, {Rule: "test-set-skips", Findings: 0}},
			Skipped: []string{"test-set-skips"}},
	}
	if !reflect.DeepEqual(records, want) {
		got, _ := json.Marshal(records)
		expected, _ := json.Marshal(want)
		t.Errorf("readiness records\n got %s\nwant %s", got, expected)
	}
	if records["b.ts"].Measured() || !records["a.ts"].Measured() {
		t.Error("a file a set rule skipped must read unmeasured, and one every set rule ran on measured")
	}
	if records["a.ts"].Findings() != 4 {
		t.Errorf("a.ts totals %d findings, want 4", records["a.ts"].Findings())
	}
	// The measure-only rule's notes stay out of the run's, and its skip-free note set is empty.
	for fileName, notes := range result.Notes {
		if _, present := notes["test-set-off"]; present {
			t.Errorf("%s carries a measure-only rule's notes into the run's", fileName)
		}
	}
}

// TestReadinessOffMeasuresNothing: without Readiness the walk records no measurement and runs no rule
// the chain leaves off, which is every run before readiness existed.
func TestReadinessOffMeasuresNothing(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json":       minimalConfig,
		"CohereSettings.json": `{"rules": {"test-set-off": "off", "test-other": "error"}}`,
		"a.ts":                "export const first = 1;\n",
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory, SingleThreaded: true})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	graph.LintConfig, err = configuration.Load(filepath.Join(directory, "CohereSettings.json"))
	if err != nil {
		t.Fatalf("loading the config: %v", err)
	}
	ran := false
	rules := []rule.Rule{
		{Name: "test-other", Run: func(ctx rule.Context, options any) rule.Listeners { return nil }, NoListener: rule.NoListenerAnswersInRun},
		{Name: "test-set-off", Run: func(ctx rule.Context, options any) rule.Listeners { ran = true; return nil }, NoListener: rule.NoListenerAnswersInRun},
	}
	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), rules)
	if err != nil {
		t.Fatalf("walking: %v", err)
	}
	if result.Adamic != nil || ran {
		t.Errorf("with readiness off the walk recorded %v and ran the off rule: %t", result.Adamic, ran)
	}
}

// TestReadinessRunsATierEnabledSetRuleOncePerFile: a set rule the chain already runs is counted as it
// reports, never run a second time to measure, so readiness adds nothing to a rule the tier has on (#drbrp8c).
func TestReadinessRunsATierEnabledSetRuleOncePerFile(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json":       minimalConfig,
		"CohereSettings.json": `{"rules": {"test-set-enabled": "error"}}`,
		"a.ts":                "export const first = 1;\n",
		"b.ts":                "export const second = 2;\n",
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory, SingleThreaded: true})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	graph.LintConfig, err = configuration.Load(filepath.Join(directory, "CohereSettings.json"))
	if err != nil {
		t.Fatalf("loading the config: %v", err)
	}
	graph.Readiness = &program.Readiness{Options: map[string][]json.RawMessage{"test-set-enabled": nil}}

	var runs atomic.Int32
	rules := []rule.Rule{{Name: "test-set-enabled", Run: func(ctx rule.Context, options any) rule.Listeners {
		runs.Add(1)
		return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
			ctx.ReportNode(node, rule.Message{Id: "enabled", Description: "enabled"})
		}}
	}}}
	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), rules)
	if err != nil {
		t.Fatalf("walking: %v", err)
	}
	if got := runs.Load(); got != 2 {
		t.Errorf("the tier-enabled set rule ran %d times over 2 files, want once per file", got)
	}
	if len(result.Diagnostics) != 2 {
		t.Errorf("reported %d findings, want each file's one", len(result.Diagnostics))
	}
	for fileName, record := range result.Adamic {
		if want := []program.AdamicCount{{Rule: "test-set-enabled", Findings: 1}}; !reflect.DeepEqual(record.Counts, want) {
			t.Errorf("%s counts %v, want %v: the one run's finding, counted", filepath.Base(fileName), record.Counts, want)
		}
	}
}

// TestAWarmReplayRunsNoRuleForReadiness: readiness is a file's property and rides the findings cache, so a
// walk that replays every file runs no rule to measure, the measure-only and type-aware ones included, and
// reads the records the walk that recorded them took (#drbrp8c).
//
// c.ts carries a directive that silences one of its findings. Replayed whole, it is dispatched with no rule to
// walk so its directive is tallied, and its record must still be the recorded one: an empty walked part read as
// unmeasured would leave every directive file out of readiness (#kdee854).
func TestAWarmReplayRunsNoRuleForReadiness(t *testing.T) {
	t.Parallel()
	typeAware := typeAwareRule(t)
	root := writeProject(t, map[string]string{
		"tsconfig.json":       minimalConfig,
		"CohereSettings.json": `{"rules": {"test-set-enabled": "error", "test-set-off": "off", "` + typeAware.Name + `": "off"}}`,
		"a.ts":                "export const first = 1;\nexport const negated = -first;\n",
		"b.ts":                "export const second = 2;\n",
		"c.ts":                "// cohere-disable-next-line test-set-enabled -- the directive file\nexport const third = 3;\n",
	})

	var runs atomic.Int32
	counted := func(subject rule.Rule) rule.Rule {
		run := subject.Run
		subject.Run = func(ctx rule.Context, options any) rule.Listeners {
			runs.Add(1)
			return run(ctx, options)
		}
		return subject
	}
	eachConstant := func(id string) func(ctx rule.Context, options any) rule.Listeners {
		return func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
				ctx.ReportNode(node, rule.Message{Id: id, Description: id})
			}}
		}
	}
	rules := []rule.Rule{
		counted(rule.Rule{Name: "test-set-enabled", Run: eachConstant("enabled")}),
		counted(rule.Rule{Name: "test-set-off", Run: eachConstant("off")}),
		counted(typeAware),
	}

	walk := func(previous *program.LintCache) (program.Result, *program.LintCache) {
		graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json")})
		if err != nil {
			t.Fatalf("building: %v", err)
		}
		graph.LintConfig, err = configuration.Load(filepath.Join(root, "CohereSettings.json"))
		if err != nil {
			t.Fatalf("loading the config: %v", err)
		}
		graph.Readiness = &program.Readiness{Options: map[string][]json.RawMessage{
			"test-set-enabled": nil, "test-set-off": nil, typeAware.Name: nil,
		}}
		reuse := program.NewFindingsReuse(program.HashRuleSet([]string{"fixture"}), previous, program.PathAnchor{})
		reuse.MeasureReadiness()
		graph.FindingsReuse = reuse
		result, err := graph.Walk(context.Background(), graph.ProjectFiles(), rules)
		if err != nil {
			t.Fatalf("walking: %v", err)
		}
		return result, reuse.Recorded()
	}
	recordsOf := func(result program.Result) map[string]*program.AdamicRecord {
		records := map[string]*program.AdamicRecord{}
		for fileName, record := range result.Adamic {
			records[filepath.Base(fileName)] = record
		}
		return records
	}

	cold, recorded := walk(nil)
	if runs.Load() == 0 {
		t.Fatal("the recording walk ran no rule, so a warm walk running none proves nothing")
	}
	measureOnly := false
	for _, record := range cold.Adamic {
		for _, count := range record.Counts {
			measureOnly = measureOnly || count.Rule == "test-set-off" && count.Findings > 0
		}
	}
	if !measureOnly {
		t.Fatal("the measure-only rule counted nothing on the recording walk, so its replay is untested here")
	}

	runs.Store(0)
	warm, _ := walk(recorded)
	if got := runs.Load(); got != 0 {
		t.Errorf("the warm walk ran rules %d times; readiness rides the findings cache and should run none", got)
	}
	if warm.FilesReplayed != 3 {
		t.Errorf("the warm walk replayed %d of 3 files; every file's record should serve a run that measures", warm.FilesReplayed)
	}
	if warm.Coverage.Suppressed != cold.Coverage.Suppressed || cold.Coverage.Suppressed == 0 {
		t.Errorf("c.ts's directive withheld %d findings warm and %d cold; it should withhold one either way",
			warm.Coverage.Suppressed, cold.Coverage.Suppressed)
	}
	if !reflect.DeepEqual(recordsOf(cold), recordsOf(warm)) {
		got, _ := json.Marshal(recordsOf(warm))
		want, _ := json.Marshal(recordsOf(cold))
		t.Errorf("the warm walk's records\n got %s\nwant %s", got, want)
	}
}
