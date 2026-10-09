package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A read of a structural object checks its own heap kind. Register every descendant
// field too: returning or aliasing that object must not lose the checks on its reads.
// Interface descendants share the same data checks, including inherited fields.

func viewDataType(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if !viewDataType(member) {
				return false
			}
		}
		return true
	}
	return target.Flags()&checker.TypeFlagsUndefined != 0 || interfaceScalar(target) || target.Flags()&checker.TypeFlagsObject != 0
}

func viewInterfaceType(target *checker.Type) bool {
	if target.ObjectFlags()&checker.ObjectFlagsInterface != 0 {
		return true
	}
	return target.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.Target() != nil && target.Target().ObjectFlags()&checker.ObjectFlagsInterface != 0
}

// An untagged structural cast installs the same checked reads without a tag test.
// Admission preserves nominal identity and the existing writable-slot restrictions.
func (l *lowering) structuralViewCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if value.Type() != ir.Object || source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(target) || !l.checker.IsTypeAssignableTo(target, source) {
		return nil, nil
	}
	if len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return nil, l.notYet(node, "a dictionary checked view")
	}
	if err := l.widened(target, source, map[[2]*checker.Type]bool{}); err != nil {
		return nil, l.notYet(node, "a writable-slot checked view requiring source contract certification")
	}
	return l.view(node, value, target)
}
