package program_test

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// A disable comment naming a real rule cohere doesn't run is a note, not a finding, as @system_cohere
// ruled for #v1ah2qq. trpc carries 43 of them, excalidraw 6 and TanStack/query 160: plugins cohere has not
// ported (import/, jsx-a11y/, @eslint-react/) and ESLint core rules it hasn't. A name that can only be a
// mistake stays an error: an unprefixed name ESLint does not ship, and an unknown name under one of our own
// plugins, which cohere runs in full.
//
// @eslint-react/no-array-index-key is the trap: cohere runs react/no-array-index-key, and matching on the
// bare name would report a TanStack directive as a stale prefix pointing at a rule of another plugin.
func TestADirectiveNamingARuleCohereDoesNotRunIsANoteAndATypoStaysAnError(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"// eslint-disable-next-line import/no-cycle -- another plugin's rule",
		"export const a = 1;",
		"// eslint-disable-next-line jsx-a11y/alt-text, import/no-cycle -- two at once",
		"export const b = 1;",
		"// eslint-disable-next-line @eslint-react/no-array-index-key -- a bare name cohere runs elsewhere",
		"export const c = 1;",
		"// eslint-disable-next-line no-with -- an ESLint core rule cohere hasn't ported",
		"export const d = 1;",
		"// eslint-disable-next-line no-witth -- a typo of a core rule",
		"export const e = 1;",
		"// eslint-disable-next-line nexus/consistency-no-enumm -- a typo under our own plugin",
		"export const f = 1;",
		"",
	}, "\n")
	directory := writeProject(t, map[string]string{"tsconfig.json": minimalConfig, "main.ts": source})
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	graph.RegisteredRuleNames = []string{"no-console", "react/no-array-index-key", "nexus/consistency-no-enum"}
	listening := rule.Rule{
		Name: "no-console",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {}}
		},
	}

	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{listening})
	if err != nil {
		t.Fatalf("walking: %v", err)
	}

	reported := []string{}
	for _, diagnostic := range result.Diagnostics {
		reported = append(reported, diagnostic.RuleName+"/"+diagnostic.Message.Id)
	}
	if want := []string{"no-witth/ruleNotFound", "nexus/consistency-no-enumm/ruleNotFound"}; !slices.Equal(reported, want) {
		t.Errorf("the findings are %q, want only the two typos %q", reported, want)
	}

	want := map[string]int{"import/no-cycle": 2, "jsx-a11y/alt-text": 1, "@eslint-react/no-array-index-key": 1, "no-with": 1}
	if !maps.Equal(result.Coverage.UnrunRuleReferences, want) {
		t.Errorf("the rules counted as not run are %v, want %v", result.Coverage.UnrunRuleReferences, want)
	}
}
