package core

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every test in this package runs the three gated rules with their gates checked: a root a gate
// skips is built anyway and must hold nothing the rule could report.
func init() {
	CheckCodePathGates = true
}

// TestCodePathRootsAreIndexedOncePerFile runs the three rules that read codePathRoots over one file
// with one FileCache, the way the walk runs them, on a file each of them has a root to build for.
// The index must be filled once for all three, not once per rule.
func TestCodePathRootsAreIndexedOncePerFile(t *testing.T) {
	t.Parallel()
	source := "export function total(items: number[]) {\n" +
		"  let sum = 0;\n" +
		"  for (const item of items) { sum += item; }\n" +
		"  if (sum > 10) { return; }\n" +
		"  sum = 1;\n" +
		"  console.log(sum);\n" +
		"}\n"
	fills := -1
	probe := rule.Rule{Name: "code-path-roots", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
			for _, gated := range []rule.Rule{NoUselessReturn, NoUnreachableLoop, NoUselessAssignment} {
				if listener := gated.Run(ctx, nil)[ast.KindSourceFile]; listener != nil {
					listener(node)
				}
			}
			fills = ctx.FileCache.Fills()["control_flow_graph.Roots"]
		}}
	}}
	rule_testing.RunTypedFiles(t, probe, map[string]string{"/fixture.ts": source}, "/fixture.ts")
	if fills != 1 {
		t.Fatalf("the root index was filled %d times for one file across three rules, want once", fills)
	}
}
