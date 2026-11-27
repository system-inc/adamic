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
	return target.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.Target() != nil && target.Target().ObjectFlags()&checker.ObjectFlagsInterface != 0
}

// An untagged structural cast installs the same checked reads without a tag test.
// Admission preserves nominal identity and the existing writable-slot restrictions.
//
// castProof proved the types' half of the guard and answered structuralView, so castRoadValue takes what this
// returns as the cast's value: it never answers nil without an error. A value castProof couldn't see, one held
// otherwise than as a plain object, stops in the interface arm's words. Over all of tsc in first-link trap mode,
// utilities.ts:5163's `node.parent as Declaration` read its operand held as a Union, and the nil this returned
// was dereferenced in expression's noteViewArrayStorage, a compiler panic (#bgm34gh).
func (l *lowering) structuralViewCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if value.Type() != ir.Object {
		return nil, l.notYet(node, "a checked cast from "+l.checker.TypeToString(source)+" to "+l.checker.TypeToString(target)+" whose value isn't a plain object")
	}
	if source.Flags()&checker.TypeFlagsObject == 0 && source.Flags()&checker.TypeFlagsIntersection == 0 || target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(target) || !l.checker.IsTypeAssignableTo(target, source) {
		return nil, l.notYet(node, "a structural checked view of "+l.checker.TypeToString(target)+" over "+l.checker.TypeToString(source)+" that its proof didn't hold")
	}
	if len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return nil, l.notYet(node, "a dictionary checked view")
	}
	if err := l.widened(target, source, map[[2]*checker.Type]bool{}); err != nil {
		return nil, l.notYet(node, "a writable-slot checked view requiring source contract certification")
	}
	return l.view(node, value, source, target)
}
