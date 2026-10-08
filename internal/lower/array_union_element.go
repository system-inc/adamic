package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The checker supplies the union's numeric index type. Reading it is safe only
// when every alternative already keeps its slots in that representation. A
// reference union can read any strong reference pointer without boxing it;
// numbers and booleans cannot be reinterpreted as those pointers.
func (l *lowering) arrayUnionElement(node *ast.Node, arrayType *checker.Type) (ir.Type, error) {
	element := l.checker.GetIndexTypeOfType(arrayType, l.checker.GetNumberType())
	if element == nil {
		return 0, l.notYet(node, "a value of type "+l.checker.TypeToString(arrayType)+" where an array goes")
	}
	held, known := l.kept(l.concrete(element))
	if !known || slotless(held) && held != ir.Union {
		return 0, l.notYet(node, "an array union with unsupported element storage")
	}
	for _, member := range arrayType.Types() {
		member = l.phantomArrayView(l.concrete(member))
		if !l.checker.IsArrayType(member) {
			return 0, l.notYet(node, "a union with a non-array alternative where an array goes")
		}
		memberElement := l.concrete(l.checker.GetElementTypeOfArrayType(member))
		memberHeld, memberKnown := l.kept(memberElement)
		compatible := memberHeld == held || held == ir.Union && memberHeld.IsReference() && memberHeld != ir.Weak
		if !memberKnown || memberElement.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNever) != 0 || !compatible {
			return 0, l.notYet(node, "an array union with incompatible element storage; narrow the array before reading its elements")
		}
	}
	return held, nil
}
