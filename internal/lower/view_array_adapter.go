package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func init() { viewArrayContractHook = buildViewArrayContract }
func buildViewArrayContract(l *lowering, node *ast.Node, target *checker.Type, build viewContractBuilder) (ir.ViewContractID, error) {
	if checker.IsTupleType(target) {
		return 0, l.notYet(node, "a checked tuple view with optional or rest positions")
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	contract := ir.ViewContract{Kind: ir.ViewArray, Name: l.checker.TypeToString(target), Of: ir.Array}
	l.result.ViewContractTypes[int(target.Id())] = id
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	handled, err := l.viewArrayContract(node, target, func(element *checker.Type) error {
		var err error
		contract.Element, err = build(l.concrete(element))
		return err
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
