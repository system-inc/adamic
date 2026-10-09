package program_test

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
	"github.com/system-inc/cohere/internal/types/program"
)

// A rule's notes are counted per file, per rule and per key, never summed across any of them, because
// the findings cache stores a file's notes beside its findings and a refresh re-walks only some of a
// file's rules. Two rules noting the same key in the same file stay apart, and a file nothing noted in
// does not appear.
func TestNotesAreCountedPerFilePerRulePerKey(t *testing.T) {
	t.Parallel()
	notingEachDeclaration := func(name string, key func(declaration *ast.Node) string) rule.Rule {
		return rule.Rule{
			Name: name,
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
					ctx.Note(key(node))
				}}
			},
		}
	}
	byName := notingEachDeclaration("test-by-name", func(declaration *ast.Node) string { return declaration.Name().Text() })
	constant := notingEachDeclaration("test-constant", func(declaration *ast.Node) string { return "a" })

	directory := writeProject(t, map[string]string{
		"tsconfig.json":       minimalConfig,
		"CohereSettings.json": `{"rules": {"test-by-name": "error", "test-constant": "error"}}`,
		"main.ts":             "export const a = 1;\nexport const b = 2;\nexport const a2 = a + b;\n",
		"other.ts":            "export const a = 3;\n",
		"quiet.ts":            "export function f(): void {}\n",
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	graph.LintConfig, err = configuration.Load(filepath.Join(directory, "CohereSettings.json"))
	if err != nil {
		t.Fatalf("loading the config: %v", err)
	}
	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{byName, constant})
	if err != nil {
		t.Fatalf("walking: %v", err)
	}

	want := map[string]program.RuleNotes{
		filepath.Join(directory, "main.ts"): {
			"test-by-name":  {"a": 1, "b": 1, "a2": 1},
			"test-constant": {"a": 3},
		},
		filepath.Join(directory, "other.ts"): {
			"test-by-name":  {"a": 1},
			"test-constant": {"a": 1},
		},
	}
	if !reflect.DeepEqual(result.Notes, want) {
		t.Errorf("notes are %v, want %v", result.Notes, want)
	}
}

// The caller-data rule notes each write a `@processState` type exempts, under the tagged type and the
// file declaring it, so --coverage can say which tag silenced how much. Two exempt writes in one file are
// two notes; a write through an untagged type is a finding and no note.
func TestTheCallerDataRuleNotesEachProcessStateExemption(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json":       minimalConfig,
		"CohereSettings.json": `{"rules": {"nexus/correctness-no-caller-data-mutation": "error"}}`,
		"hub.ts": "/** @processState the one hub every helper keeps current */\n" +
			"export interface HubInterface { count: number; names: string[] }\n" +
			"export interface PlainInterface { count: number }\n" +
			"export function touch(hub: HubInterface): void { hub.count += 1; hub.names.push('x'); }\n" +
			"export function plain(state: PlainInterface): void { state.count = 1; }\n",
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	graph.LintConfig, err = configuration.Load(filepath.Join(directory, "CohereSettings.json"))
	if err != nil {
		t.Fatalf("loading the config: %v", err)
	}
	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{nexus.CorrectnessNoCallerDataMutation})
	if err != nil {
		t.Fatalf("walking: %v", err)
	}

	hub := filepath.Join(directory, "hub.ts")
	want := map[string]program.RuleNotes{
		hub: {"nexus/correctness-no-caller-data-mutation": {"HubInterface in " + hub: 2}},
	}
	if !reflect.DeepEqual(result.Notes, want) {
		t.Errorf("notes are %v, want %v", result.Notes, want)
	}
	if len(result.Diagnostics) != 1 {
		t.Errorf("%d findings, want the one write through the untagged type: %v", len(result.Diagnostics), result.Diagnostics)
	}
}

// The caller-data rule notes each write a `@mutates` contract excuses, under the tagged method and
// parameter and the file declaring them. The contract sits in another file, as Base's scheduler does
// for api's jobs, and is reached through the override; the neighbouring parameter is a finding and no
// note (#tnn31qs).
func TestTheCallerDataRuleNotesEachOutParameterContract(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json":       minimalConfig,
		"CohereSettings.json": `{"rules": {"nexus/correctness-no-caller-data-mutation": "error"}}`,
		"base.ts": "export interface RecordInterface { status: string; log: string[] }\n" +
			"export abstract class ScheduledExecutable {\n" +
			"    /** @mutates entity the scheduler persists the record after run returns */\n" +
			"    protected abstract run(entity: RecordInterface, context: RecordInterface): void;\n" +
			"}\n",
		"job.ts": "import { RecordInterface, ScheduledExecutable } from './base';\n" +
			"export class Job extends ScheduledExecutable {\n" +
			"    protected run(job: RecordInterface, context: RecordInterface): void { job.status = 'done'; job.log.push('done'); context.status = 'x'; }\n" +
			"}\n",
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	graph.LintConfig, err = configuration.Load(filepath.Join(directory, "CohereSettings.json"))
	if err != nil {
		t.Fatalf("loading the config: %v", err)
	}
	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{nexus.CorrectnessNoCallerDataMutation})
	if err != nil {
		t.Fatalf("walking: %v", err)
	}

	job := filepath.Join(directory, "job.ts")
	want := map[string]program.RuleNotes{
		job: {"nexus/correctness-no-caller-data-mutation": {"ScheduledExecutable.run(entity) in " + filepath.Join(directory, "base.ts"): 2}},
	}
	if !reflect.DeepEqual(result.Notes, want) {
		t.Errorf("notes are %v, want %v", result.Notes, want)
	}
	if len(result.Diagnostics) != 1 {
		t.Errorf("%d findings, want the one write through the untagged parameter: %v", len(result.Diagnostics), result.Diagnostics)
	}
}
