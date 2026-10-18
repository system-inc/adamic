package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
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

// A zero-argument erased marker is checked from actual producer identity before
// invocation. The call retains its contract across read optimization and
// evaluates the operand once, preserving closure identity.
func (l *lowering) viewCallableStoredMarkerCall(node *ast.Node, signature *checker.Signature) (ir.Expression, error) {
	call := node.AsCallExpression()
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if !l.viewCallableVoidMarker(signature) || len(call.Arguments.Nodes) != 0 || outer.Parent == nil || outer.Parent.Kind != ast.KindExpressionStatement {
		return nil, l.notYet(node, "a call through an erased never-rest callable marker whose result is observed")
	}
	value, err := l.expression(call.Expression)
	if err != nil {
		return nil, err
	}
	declared := l.checker.GetTypeAtLocation(call.Expression)
	id := l.viewCallableMarkerContract(declared)
	source := ast.GetSourceFileOfNode(call.Expression)
	label := source.Text()[scanner.GetTokenPosOfNode(call.Expression, source, false):call.Expression.End()]
	return ir.CallClosure{Closure: value, CheckedDiscard: true, DiscardContract: id, DiscardView: label}, nil
}

func (l *lowering) viewCallableMarkerContract(target *checker.Type) ir.ViewContractID {
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	id := l.result.ViewContractTypes[int(target.Id())]
	if id == 0 {
		id = ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{})
		l.result.ViewContractTypes[int(target.Id())] = id
	}
	l.result.ViewContracts[id-1] = ir.ViewContract{Kind: ir.ViewCallable, Of: ir.Closure, Name: l.checker.TypeToString(target), DiscardResult: true}
	return id
}

func (l *lowering) viewCallableMarkerType(target *checker.Type) bool {
	if l.includesUndefined(target) {
		target = l.checker.GetNonNullableType(target)
	}
	signatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	return len(signatures) == 1 && len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) == 0 && l.viewCallableVoidMarker(signatures[0])
}
