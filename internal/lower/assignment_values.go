package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// optionalAssignmentValue judges every handle that holds an assignment's result against
// its receiving slot (#7n03qzg). A chain's innermost handle can write through the result
// even when its outermost handle declares exactly the receiving type.
func (l *lowering) optionalAssignmentValue(node *ast.Node, target *checker.Type) *optionalWidening {
	if target == nil {
		return nil
	}
	value := node
	for value.Kind == ast.KindBinaryExpression && value.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		value = ast.SkipParentheses(value.AsBinaryExpression().Right)
	}
	assigned := l.checker.GetTypeAtLocation(value)
	// Any and unknown do not prove absence, even when non-nullable keeps their identity.
	nonnull := assigned.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) == 0 && l.checker.GetNonNullableType(assigned) == assigned
	judgeValue := false
	for assignment := node; assignment.Kind == ast.KindBinaryExpression && assignment.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken; assignment = ast.SkipParentheses(assignment.AsBinaryExpression().Right) {
		left := ast.SkipParentheses(assignment.AsBinaryExpression().Left)
		if left.Kind == ast.KindObjectLiteralExpression || left.Kind == ast.KindArrayLiteralExpression {
			return &optionalWidening{handle: l.notYet(assignment, "an assignment to a destructuring pattern used as a value")}
		}
		source := l.assignmentWriteType(left)
		if l.assignmentEvolvingLet(left, source) {
			// An evolving let has no fixed declared handle. GetTypeAtLocation(value)
			// gives the post-store value's RHS type; later flow writes are judged separately.
			if assigned.Flags()&checker.TypeFlagsAny != 0 {
				return &optionalWidening{handle: l.notYet(assignment, "an assignment used as a value without a known target write type")}
			}
			judgeValue = true
			continue
		}
		if source == nil || source.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
			if !assignmentPrimitiveMembers(assigned) {
				return &optionalWidening{handle: l.notYet(assignment, "an assignment used as a value without a known target write type")}
			}
			judgeValue = true
			continue
		}
		if nonnull {
			// The value just stored is neither null nor undefined, literal or not.
			source = l.checker.GetNonNullableType(source)
		}
		if assignmentPrimitiveMembers(source) {
			// Primitives are copied values. Reassigning their handle cannot change a field of the result.
			judgeValue = true
			continue
		}
		if found := l.optionalWidened(source, target, nil, map[[2]*checker.Type]bool{}); found != nil {
			return found
		}
		// The handle can declare a conflicting field, not just hide one. optionalWidened
		// skips nonassignable pairs, so keep the ordinary mutable relation here too.
		if found := l.widened(source, target, map[[2]*checker.Type]bool{}); found != nil {
			return &optionalWidening{handle: l.wideningRefusal(node, source, target, found)}
		}
		if !l.checker.IsTypeAssignableTo(source, target) {
			return &optionalWidening{handle: l.notYet(node, "an assignment used as a value whose target write type "+l.checker.TypeToString(source)+" cannot be read as "+l.checker.TypeToString(target))}
		}
	}
	if judgeValue {
		return l.optionalValue(value, target)
	}
	return nil
}

// assignmentPrimitiveMembers proves there is no object alias, including in a union.
// An intersection, object, non-primitive or opaque type is never taken as that proof.
func assignmentPrimitiveMembers(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !assignmentPrimitiveMembers(member) {
				return false
			}
		}
		return true
	}
	return proven.Flags()&(checker.TypeFlagsPrimitive|checker.TypeFlagsNever) != 0
}

// Only a plain, unannotated let without an initializer has an evolving-any
// declared type. An annotation, initializer, property or other binding keeps its write type.
func (l *lowering) assignmentEvolvingLet(left *ast.Node, source *checker.Type) bool {
	if !ast.IsIdentifier(left) || source == nil || source.Flags()&checker.TypeFlagsAny == 0 {
		return false
	}
	symbol := l.symbol(left)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindVariableDeclaration && declaration.Type() == nil && declaration.Initializer() == nil && declaration.Parent != nil && declaration.Parent.Flags&ast.NodeFlagsLet != 0 {
			return true
		}
	}
	return false
}

// assignmentWriteType reads declared storage, never the flow-narrowed value of an
// identifier or a getter's read type. The checker reads an element on the left of =
// with AccessFlagsWriting, including array and record slots, so noUncheckedIndexedAccess
// adds no undefined there. Unresolved targets are not guessed from the right operand.
func (l *lowering) assignmentWriteType(left *ast.Node) *checker.Type {
	switch left.Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression:
		if symbol := l.symbol(left); symbol != nil {
			return checker.Checker_getWriteTypeOfSymbol(l.checker, symbol)
		}
	case ast.KindElementAccessExpression:
		return l.checker.GetTypeAtLocation(left)
	}
	return nil
}

// Destructuring results have several distinct handles. Hold them before the ordinary
// invariant relation can give a different stop for their aggregate pattern type.
func (l *lowering) assignmentPatternValueGap(node *ast.Node) error {
	if node.Kind != ast.KindBinaryExpression || node.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
		return nil
	}
	left := ast.SkipParentheses(node.AsBinaryExpression().Left)
	if left.Kind != ast.KindObjectLiteralExpression && left.Kind != ast.KindArrayLiteralExpression {
		return nil
	}
	if viewSite(node) && l.slotType(node) != nil || l.impliedTarget(node) != nil {
		return l.notYet(node, "an assignment to a destructuring pattern used as a value")
	}
	return nil
}

func init() {
	refusalRules.register(3090, "assignmentPatternValueGap", (*lowering).assignmentPatternValueGap)
}
