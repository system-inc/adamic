package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The array adapter on compiler/views-v4 uses this interface classification.
func viewInterfaceType(target *checker.Type) bool {
	if target.ObjectFlags()&checker.ObjectFlagsInterface != 0 {
		return true
	}
	return target.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		target.Target() != nil &&
		target.Target().ObjectFlags()&checker.ObjectFlagsInterface != 0
}

// An untagged structural cast installs the same checked reads without a tag test.
// Admission preserves nominal identity and the existing writable-slot restrictions.
func (lowering *lowering) structuralViewCast(
	node *ast.Node,
	value ir.Expression,
	source, target *checker.Type,
) (ir.Expression, error) {
	if value.Type() != ir.Object ||
		source.Flags()&checker.TypeFlagsObject == 0 ||
		target.Flags()&checker.TypeFlagsObject == 0 ||
		isClassInstance(target) ||
		!lowering.checker.IsTypeAssignableTo(target, source) {
		return nil, nil
	}
	if len(lowering.checker.GetIndexInfosOfType(target)) != 0 {
		return nil, lowering.notYet(node, "a dictionary checked view")
	}
	if err := lowering.widened(target, source, map[[2]*checker.Type]bool{}); err != nil {
		return nil, lowering.notYet(node, "a writable-slot checked view requiring source contract certification")
	}
	return lowering.view(node, value, source, target)
}
