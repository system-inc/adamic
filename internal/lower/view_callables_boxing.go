package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Bit zero is undefined. A zero mask is an unknown signature, never all members.
func (l *lowering) viewCallableRepresentationMask(proven *checker.Type) uint16 {
	if isClassInstance(proven) || proven.Flags()&checker.TypeFlagsIntersection != 0 || proven.Flags()&checker.TypeFlagsObject != 0 && l.unsupportedViewFamily(proven) != "" {
		return 0
	}
	if proven.Flags()&checker.TypeFlagsUndefined != 0 {
		return 1
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		var mask uint16
		for _, member := range proven.Types() {
			child := l.viewCallableRepresentationMask(member)
			if child == 0 {
				return 0
			}
			mask |= child
		}
		return mask
	}
	of, known := l.representation(proven)
	if !known || of == 0 || of == ir.Union || proven.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
		return 0
	}
	if of == ir.MaybeNumber {
		return 1 | 1<<ir.Number
	}
	if of == ir.MaybeBoolean {
		return 1 | 1<<ir.Boolean
	}
	return 1 << of
}

func (l *lowering) recordViewCallableRepresentations(index int, function *ir.Function, declaration *ast.Node) {
	signature := l.checker.GetSignatureFromDeclaration(declaration)
	if signature == nil {
		return
	}
	masks := make([]uint16, len(function.Parameters)+1)
	offset := len(function.Parameters) - len(signature.Parameters())
	if offset < 0 {
		return
	}
	for i := 0; i < offset; i++ {
		masks[i] = 1 << ir.Object
	}
	for i, parameter := range signature.Parameters() {
		masks[i+offset] = l.viewCallableRepresentationMask(l.censusCallableParameterType(parameter))
	}
	masks[len(masks)-1] = l.viewCallableRepresentationMask(l.checker.GetReturnTypeOfSignature(signature))
	function.CallableMasks = masks
	l.recordViewCallableProducerPayloads(index, declaration, signature)
}

func (l *lowering) viewCallableBoxedRepresentation(proven *checker.Type) bool {
	of, known := l.representation(proven)
	if !known {
		return false
	}
	if of == ir.Object && !isClassInstance(proven) && l.unsupportedViewFamily(proven) == "" {
		return true
	}
	if viewCallableScalarRepresentation(of) {
		return true
	}
	// Union signatures need known members. Aggregate transitivity has its own hook.
	if of != ir.Union || l.viewCallableRepresentationMask(proven) == 0 {
		return false
	}
	for _, member := range proven.Types() {
		child, known := l.representation(member)
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		if !known || !(viewCallableScalarRepresentation(child) || child == ir.Object && !isClassInstance(member) && l.unsupportedViewFamily(member) == "") {
			return false
		}
	}
	return true
}
