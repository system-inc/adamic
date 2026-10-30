package lower

import (
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) isArrayPredicateCall(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	return call.QuestionDotToken == nil && len(call.Arguments.Nodes) == 1 && callee.Kind == ast.KindPropertyAccessExpression && callee.AsPropertyAccessExpression().QuestionDotToken == nil && callee.Name().Text() == "isArray" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array") && l.libraryMember(callee)
}

func (l *lowering) arrayIsArray(node *ast.Node) (ir.Expression, bool, error) {
	if !l.isArrayPredicateCall(node) {
		return nil, false, nil
	}
	argument := node.AsCallExpression().Arguments.Nodes[0]
	if !l.arrayPredicateDomain(l.checker.GetTypeAtLocation(argument)) && !l.exactObject(argument, 0) {
		return nil, true, l.notYet(argument, "Array.isArray on a tuple or an erased object/any/unknown view")
	}
	value, err := l.expression(argument)
	if err != nil {
		return nil, true, err
	}
	return ir.ArrayIsArray{Value: fit(value, ir.Union)}, true, nil
}

// The library's any[] predicate describes the runtime test, not an element
// contract. Only array members already promised by the declared input survive.
func (l *lowering) arrayPredicateMembers(proven *checker.Type, array bool) *checker.Type {
	members := []*checker.Type{}
	var collect func(*checker.Type)
	collect = func(member *checker.Type) {
		member = l.concrete(member)
		if member.Flags()&checker.TypeFlagsUnion != 0 {
			for _, part := range member.Types() {
				collect(part)
			}
			return
		}
		if member.Flags()&checker.TypeFlagsUnknown != 0 {
			if array {
				members = append(members, predicateReadonlyArray(l.checker, member, true))
			} else {
				members = append(members, member)
			}
			return
		}
		if member.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
			return
		}
		_, _, nodeArray := l.nodeArrayLayoutOf(member)
		if (l.checker.IsArrayType(member) || nodeArray) == array {
			members = append(members, member)
		}
	}
	collect(proven)
	return l.checker.GetUnionType(members)
}

// Tuples currently have an object representation. Erased object views can hide
// one too, so the native array brand cannot decide their JavaScript arrayness.
func (l *lowering) arrayPredicateDomain(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !l.arrayPredicateDomain(member) {
				return false
			}
		}
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnknown != 0 {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter|checker.TypeFlagsIntersection) != 0 || checker.IsTupleType(proven) {
		return false
	}
	if _, _, known := l.nodeArrayLayoutOf(proven); known {
		return true
	}
	if l.checker.IsArrayType(proven) {
		return true
	}
	return proven.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(proven) || l.isLibraryType(proven, "Uint8Array", "Int32Array", "Float64Array")
}

// TypeScript's built-in signature is value is any[]. Recover the declared
// array member at a direct, dominating positive test instead of adopting any.
func (l *lowering) arrayPredicateType(node *ast.Node) *checker.Type {
	node = ast.SkipParentheses(node)
	actual := l.checker.GetTypeAtLocation(node)
	if node.Kind != ast.KindIdentifier {
		return actual
	}
	symbol := l.symbol(node)
	if symbol == nil || node.FlowNodeData() == nil {
		return actual
	}
	declared := l.checker.GetTypeOfSymbol(symbol)
	if !l.arrayPredicateDomain(declared) && declared.Flags()&checker.TypeFlagsUnknown == 0 {
		return actual
	}
	flow := node.FlowNodeData().FlowNode
	for steps := 0; flow != nil && steps < 256; steps++ {
		if flow.Flags&(ast.FlowFlagsAssignment|ast.FlowFlagsLabel|ast.FlowFlagsCall) != 0 {
			return actual
		}
		if flow.Flags&ast.FlowFlagsCondition != 0 && flow.Node != nil && l.arrayPredicateFlowCall(flow.Node) {
			call := ast.SkipParentheses(flow.Node).AsCallExpression()
			if l.symbol(ast.SkipParentheses(call.Arguments.Nodes[0])) == symbol {
				return l.arrayPredicateMembers(declared, flow.Flags&ast.FlowFlagsTrueCondition != 0)
			}
		}
		flow = flow.Antecedent
	}
	return actual
}

// Indexed reads and their ?? result retain the recovered element contract,
// including the missing element required by noUncheckedIndexedAccess.
func (l *lowering) arrayPredicateObservedType(node *ast.Node) *checker.Type {
	node = ast.SkipParentheses(node)
	actual := l.checker.GetTypeAtLocation(node)
	if node.Kind == ast.KindIdentifier {
		return l.arrayPredicateType(node)
	}
	if actual.Flags()&checker.TypeFlagsAny == 0 {
		return actual
	}
	switch node.Kind {
	case ast.KindElementAccessExpression:
		array := l.arrayPredicateType(node.AsElementAccessExpression().Expression)
		if l.checker.IsArrayType(array) {
			element := l.checker.GetElementTypeOfArrayType(array)
			if element != nil && element.Flags()&checker.TypeFlagsAny == 0 {
				return l.checker.GetUnionType([]*checker.Type{element, l.checker.GetUndefinedType()})
			}
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindQuestionQuestionToken {
			left := l.arrayPredicateObservedType(binary.Left)
			if left.Flags()&checker.TypeFlagsAny == 0 {
				return l.checker.GetUnionType([]*checker.Type{l.checker.GetNonNullableType(left), l.checker.GetTypeAtLocation(binary.Right)})
			}
		}
	}
	return actual
}

// The pinned checker owns construction of the readonly array type, just as it
// owns predicate flow facts. This introduces no any element contract.
//
//go:linkname predicateReadonlyArray github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).createArrayTypeEx
func predicateReadonlyArray(receiver *checker.Checker, element *checker.Type, readonly bool) *checker.Type

// A proven wrapper has the same arrayness test as the builtin. Its annotation
// cannot supply element facts missing from the caller's declared union.
func (l *lowering) arrayPredicateFlowCall(node *ast.Node) bool {
	if l.isArrayPredicateCall(node) {
		return true
	}
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	if call.QuestionDotToken != nil || len(call.Arguments.Nodes) != 1 {
		return false
	}
	symbol := l.symbol(ast.SkipParentheses(call.Expression))
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration || len(declaration.Parameters()) != 1 || declaration.Body() == nil || declaration.Type() == nil || declaration.Type().Kind != ast.KindTypePredicate {
			continue
		}
		target := l.checker.GetTypeFromTypeNode(declaration.Type().AsTypePredicateNode().Type)
		if !checker.Checker_isTypeIdenticalTo(l.checker, target, predicateReadonlyArray(l.checker, l.checker.GetUnknownType(), true)) {
			continue
		}
		body := declaration.Body()
		if body.Kind != ast.KindBlock || len(body.AsBlock().Statements.Nodes) != 1 {
			continue
		}
		statement := body.AsBlock().Statements.Nodes[0]
		if statement.Kind != ast.KindReturnStatement || statement.AsReturnStatement().Expression == nil {
			continue
		}
		returned := ast.SkipParentheses(statement.AsReturnStatement().Expression)
		if l.isArrayPredicateCall(returned) && l.symbol(ast.SkipParentheses(returned.AsCallExpression().Arguments.Nodes[0])) == l.symbol(declaration.Parameters()[0].Name()) {
			return true
		}
	}
	return false
}
