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
		if child.Flags()&checker.TypeFlagsVoid != 0 {
			of, known = ir.Type(254), true
		}
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

// Record read metadata now, but leave unsupported shapes as descriptors. The
// shared allocation pass decides whether their refusal is demanded. This also
// handles helper bodies lowered before a cast appears in another function.
func (l *lowering) prepareViewCallableProperty(node *ast.Node, declared *checker.Type, property *ir.Property) {
	if property.Of != ir.Closure && (property.Of != ir.Union || !l.runtimeViewCallableShape(declared)) {
		return
	}
	if l.viewCallableMarkerType(declared) {
		property.ViewContract = l.viewCallableMarkerContract(l.checker.GetNonNullableType(declared))
		return
	}
	if !l.runtimeViewCallableShape(declared) {
		return
	}
	id, err := l.prepareViewCallableRead(node, declared)
	if err != nil {
		return
	}
	l.result.ViewContractTypes[int(declared.Id())] = id
	l.result.ViewContracts[id-1].Unsupported = ""
	property.ViewContract = id
}

func (l *lowering) runtimeViewCallableShape(target *checker.Type) bool {
	if l.includesUndefined(target) {
		target = l.checker.GetNonNullableType(target)
	}
	signatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(signatures) != 1 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	s := signatures[0]
	if len(s.TypeParameters()) != 0 || s.HasRestParameter() || s.MinArgumentCount() != len(s.Parameters()) {
		return false
	}
	for _, p := range s.Parameters() {
		of, known := l.representation(l.checker.GetTypeOfSymbol(p))
		if !known || !viewCallableScalarRepresentation(of) {
			return false
		}
	}
	result := l.checker.GetReturnTypeOfSignature(s)
	if result.Flags()&checker.TypeFlagsVoid != 0 {
		return true
	}
	of, known := l.representation(result)
	return known && viewCallableScalarRepresentation(of)
}

func viewCallableScalarRepresentation(of ir.Type) bool {
	return of == ir.Number || of == ir.Boolean || of == ir.String || of == ir.MaybeNumber || of == ir.MaybeBoolean
}
