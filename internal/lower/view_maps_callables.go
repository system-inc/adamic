package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Private Map descriptors do not remove the shared array/field callable refusal.
// A producer's immutable code identity supplies the physical signature at stores.
func (l *lowering) mapCallableEntrySlot(node *ast.Node, target *checker.Type) ir.ViewContractID {
	present := l.checker.GetNonNullableType(target)
	if !l.runtimeViewCallableShape(present) {
		return 0
	}
	signatures := l.checker.GetSignaturesOfType(present, checker.SignatureKindCall)
	signature := signatures[0]
	for _, parameter := range signature.Parameters() {
		if !l.mapEntryTypeProven(l.checker.GetTypeOfSymbol(parameter), map[*checker.Type]bool{}) {
			return 0
		}
	}
	if !l.mapEntryTypeProven(l.checker.GetReturnTypeOfSignature(signature), map[*checker.Type]bool{}) {
		return 0
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewCallable, Of: ir.Closure, Name: l.checker.TypeToString(present)})
	build := func(child *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, l.concrete(child)) }
	if err := l.completeViewCallableShapeContract(node, present, id, build); err != nil {
		return 0
	}
	if !l.includesNull(target) && !l.includesUndefined(target) {
		return id
	}
	of, known := l.representation(target)
	if !known {
		return 0
	}
	nullable := ir.ViewContract{Kind: ir.ViewNullable, Of: of, Element: id, Null: l.includesNull(target), Undefined: l.includesUndefined(target), Name: l.checker.TypeToString(target)}
	l.result.ViewContracts = append(l.result.ViewContracts, nullable)
	return ir.ViewContractID(len(l.result.ViewContracts))
}
