package program_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// TestWalkReportsADirectiveNamingAnUnknownRule is the end-to-end form of ESLint's "Definition for
// rule 'x' was not found", measured on ESLint 10.8.1: rule id the unknown name, the whole comment as
// the range, enables checked too, real names and blankets silent.
//
// Two more properties are pinned because each is a way for the check to go quiet: a blanket
// file-scope disable must not silence a finding about a directive that silences nothing, and with
// no registry supplied the check is off rather than reporting every name.
func TestWalkReportsADirectiveNamingAnUnknownRule(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"/* eslint-disable */",
		"// eslint-disable-next-line no-unused-var -- why",
		"export const a = 1;",
		"// eslint-disable-next-line no-console -- a real rule",
		"export const b = 1;",
		"// eslint-disable-next-line structure/consistency-no-enum -- a stale prefix",
		"export const c = 1;",
		"/* eslint-enable no-such-rule */",
		"",
	}, "\n")
	directory := writeProject(t, map[string]string{"tsconfig.json": minimalConfig, "main.ts": source})

	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	listening := rule.Rule{
		Name: "no-console",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {}}
		},
	}

	walk := func() []string {
		result, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{listening})
		if err != nil {
			t.Fatalf("walking: %v", err)
		}
		described := []string{}
		for _, diagnostic := range result.Diagnostics {
			described = append(described, diagnostic.RuleName+"/"+diagnostic.Message.Id+" "+
				source[diagnostic.Range.Pos():diagnostic.Range.End()])
		}
		return described
	}

	if got := walk(); len(got) != 0 {
		t.Fatalf("with no registry supplied the check must be off, got %q", got)
	}

	graph.RegisteredRuleNames = []string{"no-console", "nexus/consistency-no-enum"}
	got := walk()
	want := []string{
		"no-unused-var/ruleNotFound // eslint-disable-next-line no-unused-var -- why",
		"structure/consistency-no-enum/ruleNotFound // eslint-disable-next-line structure/consistency-no-enum -- a stale prefix",
		"no-such-rule/ruleNotFound /* eslint-enable no-such-rule */",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("unknown-rule findings\n got: %q\nwant: %q", got, want)
	}

	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), []rule.Rule{listening})
	if err != nil {
		t.Fatalf("walking: %v", err)
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.RuleName == "structure/consistency-no-enum" &&
			!strings.Contains(diagnostic.Message.Description, "'nexus/consistency-no-enum'") {
			t.Errorf("a stale prefix should be told the registered name, got: %s", diagnostic.Message.Description)
		}
	}
}
