package program_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// The footer's rule count is the rules that ran, as the README says, not the rules the walk was handed:
// every registered rule. TanStack/query's config turned most of them off and the footer still read
// "484 rules" when 118 ran (#v1ah2qq). A rule off everywhere and a rule nobody configured did not run; a
// rule offered a file that it declined did, and so did one offered only some files.
func TestRulesRunCountsOnlyTheRulesOfferedAFile(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json": minimalConfig,
		"CohereSettings.json": `{"rules": {"test-listens": "error", "test-declines": "error", "test-off": "off",
			"test-only-in-special": "off"},
			"overrides": [{"files": ["special.ts"], "rules": {"test-only-in-special": "error"}}]}`,
		"main.ts":    "export const value = 1;\n",
		"special.ts": "export const other = 2;\n",
	})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	graph.LintConfig, err = configuration.Load(filepath.Join(directory, "CohereSettings.json"))
	if err != nil {
		t.Fatalf("loading the config: %v", err)
	}

	listens := func(name string) rule.Rule {
		return rule.Rule{Name: name, Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {}}
		}}
	}
	declines := rule.Rule{Name: "test-declines", Run: func(ctx rule.Context, options any) rule.Listeners { return nil }}
	rules := []rule.Rule{listens("test-listens"), declines, listens("test-off"), listens("test-only-in-special"),
		listens("test-nobody-configured")}

	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), rules)
	if err != nil {
		t.Fatalf("walking: %v", err)
	}
	if result.Coverage.RulesRun != 3 {
		t.Errorf("RulesRun is %d of the %d rules handed in, want 3: test-listens, test-declines and test-only-in-special",
			result.Coverage.RulesRun, len(rules))
	}
}
