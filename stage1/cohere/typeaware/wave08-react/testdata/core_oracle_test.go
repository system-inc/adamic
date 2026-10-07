// Owned overlay in cohere's react package. Calls its unmodified production kernels.
package react

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	hir "github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
	"testing"
)

func TestWave08ReactCoreOracle(t *testing.T) {
	for a := 0; a < 6; a++ {
		for b := 0; b < 6; b++ {
			fmt.Printf("wave08-core %d\n", immutabilityMergeKinds(immutabilityValueKind(a), immutabilityValueKind(b)))
		}
	}
	for a := 0; a < 6; a++ {
		fmt.Printf("wave08-core %d\n", immutabilityMutate(immutabilityValueKind(a)))
	}
	p := func(id int) hir.Place { return hir.Place{Identifier: hir.IdentifierId(id)} }
	probe := rule.Rule{Name: "wave08-core", NeedsTypeChecker: true, Run: func(ctx rule.Context, _ any) rule.Listeners {
		return rule.Listeners{ast.KindSourceFile: func(file *ast.Node) {
			var setter, dep *ast.Node
			var visit func(*ast.Node) bool
			visit = func(n *ast.Node) bool {
				if n.Kind == ast.KindIdentifier {
					if n.Text() == "set" {
						setter = n
					}
					if n.Text() == "dep" {
						dep = n
					}
				}
				n.ForEachChild(visit)
				return false
			}
			file.ForEachChild(visit)
			enclosing := &hir.Function{Identifiers: []*hir.Identifier{{Node: setter}, {Node: dep}, {Node: dep}, {Node: dep}}}
			if !isSetterTyped(ctx, enclosing, p(0)) {
				t.Fatal("live Dispatch setter witness missing")
			}
			for kind := 0; kind < 10; kind++ {
				captures := []hir.Place{p(0), p(1)}
				deps := []hir.IdentifierId{1}
				context := []hir.Place{p(0), p(1)}
				load := &hir.Instruction{LValue: p(2), Value: &hir.LoadContext{Place: p(1)}}
				call := &hir.Instruction{LValue: p(3), Value: &hir.CallExpression{Callee: p(0), Args: []hir.Argument{{Place: p(2)}}}}
				block := &hir.BasicBlock{Id: 0, Instructions: []hir.InstructionId{0, 1}, Terminal: &hir.Unreachable{}}
				instructions := []*hir.Instruction{load, call}
				switch kind {
				case 1:
					captures = append(captures, p(3))
					context = append(context, p(3))
				case 2:
					block.Predecessors = []hir.BlockId{0}
				case 3:
					load.Value = &hir.StoreLocal{LValue: p(2), Value: p(1)}
				case 4:
					block.Terminal = &hir.Return{Value: p(2)}
				case 5:
					call.Value = &hir.CallExpression{Callee: p(0), Args: []hir.Argument{{Place: p(2)}, {Place: p(2)}}}
				case 6:
					captures = []hir.Place{p(0)}
					context = []hir.Place{p(0)}
				case 7:
					instructions = append(instructions, &hir.Instruction{LValue: p(4), Value: &hir.CallExpression{Callee: p(0), Args: []hir.Argument{{Place: p(8)}}}})
					block.Instructions = append(block.Instructions, 2)
				case 8, 9:
					captures = []hir.Place{p(0), p(1), p(2)}
					context = []hir.Place{p(0), p(1), p(2)}
					deps = []hir.IdentifierId{1, 2}
					call.LValue = p(4)
					call.Value = &hir.CallExpression{Callee: p(0), Args: []hir.Argument{{Place: p(3)}}}
					if kind == 8 {
						load.LValue = p(3)
						load.Value = &hir.BinaryExpression{Left: p(1), Right: p(2)}
					} else {
						block.Phis = []*hir.Phi{{Place: p(3), Operands: map[hir.BlockId]hir.Place{0: p(1), 1: p(2)}}}
						instructions = []*hir.Instruction{call}
						block.Instructions = []hir.InstructionId{0}
					}
				}
				effect := &hir.Function{Context: context, Identifiers: enclosing.Identifiers, Instructions: instructions, Blocks: []*hir.BasicBlock{block}}
				count := 0
				copy := ctx
				copy.Report = func(d rule.Diagnostic) {
					if d.Message.Id != messageNoDerivingStateInEffects.Id {
						t.Fatal("unexpected message")
					}
					count++
				}
				validateDerivedEffect(copy, enclosing, effect, &hir.FunctionExpression{Captures: captures}, deps)
				if kind == 0 && count != 1 {
					t.Fatalf("positive derivation witness differs: %d", count)
				}
				fmt.Printf("wave08-core %d\n", count)
			}
		}}
	}}
	rule_testing.RunTyped(t, probe, "core.ts", `type Dispatch<T>=(value:T)=>void; declare const set:Dispatch<number>;declare const dep:number;set;dep;`)
}
