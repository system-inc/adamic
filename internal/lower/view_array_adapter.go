package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func init() { viewArrayContractHook = internArrayViewContract }

// Adapt lane 2's element callback into the shared logical contract graph. Reserve
// the id first: arrays and interfaces can contain each other recursively.
func internArrayViewContract(l *lowering, node *ast.Node, target *checker.Type, build viewContractBuilder) (ir.ViewContractID, error) {
	if checker.IsTupleType(target) {
		return 0, l.notYet(node, "a checked tuple view with per-position optional and rest contracts")
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	contract := ir.ViewContract{Kind: ir.ViewArray, Of: ir.Array, Name: l.checker.TypeToString(target)}
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	handled, err := l.viewArrayContract(node, target, func(element *checker.Type) error {
		var childError error
		contract.Element, childError = build(element)
		return childError
	})
	if err != nil {
		return 0, err
	}
	if !handled {
		return 0, l.notYet(node, "an array checked view without an element contract")
	}
	l.result.ViewContracts[int(id)-1] = contract
	return id, nil
}
