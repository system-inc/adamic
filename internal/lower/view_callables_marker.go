package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A function value's discarded result is not a writable location. Container
// aliases still pass the reverse slot relation in widenedProperties first.
func (l *lowering) viewCallableVoidMarker(signature *checker.Signature) bool {
	return l.censusNeverRestSignature(signature) && l.checker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsVoid != 0
}

// Preserve the source signature for an immediate discarded marker call. Stored
// markers cannot borrow this proof, and required arguments cannot disappear.
// Reading the original operand still runs its checked callable field contract.
func (l *lowering) viewCallableDiscardedMarker(node *ast.Node) bool {
	if node.Kind != ast.KindAsExpression {
		return false
	}
	source := l.checker.GetTypeAtLocation(node.AsAsExpression().Expression)
	target := l.checker.GetTypeAtLocation(node)
	from := l.checker.GetSignaturesOfType(source, checker.SignatureKindCall)
	to := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(from) != 1 || len(to) != 1 || !l.viewCallableVoidMarker(to[0]) || !l.runtimeViewCallableShape(source) || len(from[0].Parameters()) != 0 {
		return false
	}
	expression := node
	for expression.Parent != nil && expression.Parent.Kind == ast.KindParenthesizedExpression {
		expression = expression.Parent
	}
	call := expression.Parent
	if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().Expression != expression || len(call.AsCallExpression().Arguments.Nodes) != 0 {
		return false
	}
	for call.Parent != nil && call.Parent.Kind == ast.KindParenthesizedExpression {
		call = call.Parent
	}
	return call.Parent != nil && call.Parent.Kind == ast.KindExpressionStatement
}
