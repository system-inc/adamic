package lower

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

// The row models call forms, never the declaration's whole any. Pin both its
// documented boundary and the production predicate without registering require.
func TestHostAnyModelledRequireForms(t *testing.T) {
	t.Parallel()
	const literal = "One literal source-map-support argument (string or no-substitution template), after parentheses, assertions and non-nulls"
	const computed = "One module path computed at run time, with no spread"
	forms := map[string]bool{}
	rows := 0
	for _, row := range hostAnyModelledRows(t) {
		if row["Declaration"] != "`require(id): any`" || row["Where"] != "module.d.ts:707" {
			continue
		}
		rows++
		if row["Modelled type"] != "`never`" {
			t.Fatalf("require's Modelled type must be never, got %q", row["Modelled type"])
		}
		for _, form := range strings.Split(row["Forms"], "<br>") {
			forms[strings.TrimSpace(form)] = true
		}
	}
	if rows != 1 || len(forms) != 2 || !forms[literal] || !forms[computed] {
		t.Fatalf("host-any.md must name exactly the two Modelled require forms at module.d.ts:707; got %d rows, forms %v", rows, forms)
	}
	if hostAnyModelled(t)["module.d.ts:707"] {
		t.Fatal("form-scoped require must not enter the declaration-wide allow-list")
	}
	for _, probe := range []struct {
		source string
		form   string
	}{
		{`require("source-map-support")`, literal},
		{`require("source-map-support" as string)`, literal},
		{"require(`source-map-support`)", literal},
		{`require(path)`, computed},
		{`require(a + b)`, computed},
		{`require("fs")`, ""},
		{`require("typescript")`, ""},
		{"require(`fs`)", ""},
		{`require(...args)`, ""},
		{`require()`, ""},
		{`require("a", "b")`, ""},
	} {
		t.Run(probe.source, func(t *testing.T) {
			// Parse without a checker: deliberately invalid arities and spreads
			// still have to be rejected by the form predicate itself.
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{}, probe.source+";", core.ScriptKindTS)
			if len(file.Diagnostics()) != 0 {
				t.Fatalf("invalid probe syntax: %v", file.Diagnostics())
			}
			var calls []*ast.Node
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindCallExpression {
					calls = append(calls, node)
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			if len(calls) != 1 || calls[0].Expression().Kind != ast.KindIdentifier || calls[0].Expression().Text() != "require" {
				t.Fatalf("probe must contain one require call, got %d calls", len(calls))
			}
			if got, want := nodeCommonJSRequireThrows(calls[0]), forms[probe.form]; got != want {
				t.Errorf("nodeCommonJSRequireThrows(%s) = %t; Modelled Forms require %t", probe.source, got, want)
			}
		})
	}
}
