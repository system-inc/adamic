package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// overlapViewCast lowers a cast from a represented union to one object type that only overlaps it (overlapTarget):
// a checked view of the target over the value's own runtime shape, whatever its tag and its layout, with no test at
// the cast. A value that holds the target's fields reads through it; one that doesn't stops at its first read, with
// the view's own located message. A write through the view is judged member by member: where a member could be
// written something its own slot can't hold, the cast is NotYet, naming the member.
func (l *lowering) overlapViewCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if value.Type() != ir.Object {
		return nil, l.notYet(node, "a checked view of "+l.checker.TypeToString(target)+" over a union held as other than an object")
	}
	if len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return nil, l.notYet(node, "a dictionary checked view")
	}
	for _, member := range castMembers(source) {
		if widening := l.widened(member, target, map[[2]*checker.Type]bool{}); widening != nil {
			return nil, l.notYet(node, "a checked view of "+l.checker.TypeToString(target)+" that could write "+l.checker.TypeToString(member)+" a field it can't hold")
		}
	}
	return l.view(node, value, source, target)
}
