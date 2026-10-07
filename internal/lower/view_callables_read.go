package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// prepareViewCallableRead is the read-dispatch adapter. Call it only after the
// receiver has KnownViewed or Unknown provenance; an ordinary certified receiver
// does not need an asserted signature. A cast never invokes this adapter.
func (l *lowering) prepareViewCallableRead(node *ast.Node, declared *checker.Type) (ir.ViewContractID, error) {
	target := declared
	if l.includesUndefined(target) {
		target = l.checker.GetNonNullableType(target)
	}
	if !l.callableViewContract(target) {
		return 0, l.notYet(node, "checked callable read of "+l.checker.TypeToString(declared))
	}
	id := l.result.ViewContractTypes[int(target.Id())]
	if id == 0 {
		var err error
		id, err = buildViewCallableContract(l, node, target, nil)
		if err != nil {
			return 0, err
		}
	}
	// Shapes compare representations only. Do not recursively admit object members
	// of parameter/result types while preparing a callable read.
	build := func(child *checker.Type) (ir.ViewContractID, error) {
		of, known := l.representation(child)
		if !known || of == 0 {
			return 0, l.notYet(node, "checked callable signature representation "+l.checker.TypeToString(child))
		}
		childID := ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: of, Name: l.checker.TypeToString(child)})
		return childID, nil
	}
	if err := l.completeViewCallableShapeContract(node, target, id, build); err != nil {
		return 0, err
	}
	return id, nil
}
