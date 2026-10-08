//go:build lintoracle

package high_level_intermediate_representation

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// Go inline_iife.go:379 ranges over remap.Blocks while allocating identifiers
// and instructions. Same input, different byte output: no decoder can supply
// that future map iteration as a legitimate prepass input fact.
func TestUnit3GoInlineOrderGap(t *testing.T) {
	t.Parallel()
	const source = "function C(p: boolean) { return (() => { if (p) return 1; return 2; })(); }"
	variants := map[string]bool{}
	before := ""
	probe := rule.Rule{Name: "unit3-go-order", NeedsTypeChecker: true, Run: func(ctx rule.Context, _ any) rule.Listeners {
		return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
			ForEachFunctionLike(node, func(n *ast.Node) {
				f := Lower(n, ctx.TypeChecker)
				if f == nil {
					t.Fatal("input did not lower")
				}
				Construct(f)
				before = oracleDump(f)
				for i := 0; i < 128; i++ {
					clone := CloneFunction(f)
					if oracleDump(clone) != before {
						t.Fatal("prepass input varies")
					}
					if InlineImmediatelyInvokedFunctionExpressions(clone) != 1 {
						t.Fatal("fixture did not inline")
					}
					variants[oracleDump(clone)] = true
				}
			})
		}}
	}}
	rule_testing.RunTypedFiles(t, probe, map[string]string{"/fixture.ts": source}, "/fixture.ts")
	if len(variants) < 2 {
		t.Fatalf("Go oracle order gap changed: %d distinct outputs from 128 identical inputs", len(variants))
	}
	t.Logf("Go produced %d distinct byte outputs from 128 identical prepass graphs", len(variants))
	if destination := os.Getenv("HIR_UNIT3_ORDER_EXPORT"); destination != "" {
		if err := os.MkdirAll(destination, 0755); err != nil {
			t.Fatal(err)
		}
		outputs := []string{}
		for dump := range variants {
			outputs = append(outputs, dump)
		}
		sort.Strings(outputs)
		for name, body := range map[string]string{"oracle_order_input.ts": source + "\n", "oracle_order_before.dump": before, "oracle_order_after_a.dump": outputs[0], "oracle_order_after_b.dump": outputs[1]} {
			if err := os.WriteFile(filepath.Join(destination, name), []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
}
