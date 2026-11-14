package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// These stops name language restrictions, not missing stage-0 representations.
// Keep them on the lowering paths: a dormant generic body must not gain a new refusal.
func (l *lowering) refuseAnyValue(node *ast.Node) error {
	return &Refused{Where: l.program.Where(node), What: "a value of type any: any has no proven type", Fix: "type it unknown and narrow it where it's read, or give it its real type"}
}

func (l *lowering) neverExpression(node *ast.Node) error {
	if ast.IsIdentifier(node) && l.isLibraryGlobal(node, "eval") {
		return &Refused{Where: l.program.Where(node), What: "reading the global eval: a native binary has no compiler at runtime", Fix: "parse the input with decodeJson<T>, or call the function directly"}
	}
	var constructor *ast.Node
	switch node.Kind {
	case ast.KindNewExpression:
		constructor = node.AsNewExpression().Expression
	case ast.KindCallExpression:
		constructor = node.AsCallExpression().Expression
	}
	if constructor != nil && l.isLibraryGlobal(constructor, "Function") {
		return &Refused{Where: l.program.Where(node), What: "constructing a function with the global Function: a native binary has no compiler at runtime", Fix: "write a function or closure directly"}
	}
	var target *ast.Node
	switch node.Kind {
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			target = binary.Left
		}
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			target = unary.Operand
		}
	case ast.KindPostfixUnaryExpression:
		unary := node.AsPostfixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			target = unary.Operand
		}
	}
	if target == nil {
		return nil
	}
	target = ast.SkipParentheses(target)
	if target.Kind != ast.KindPropertyAccessExpression || target.Name().Text() != "length" {
		return nil
	}
	if !l.nonTupleArray(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)) {
		return nil
	}
	return &Refused{Where: l.program.Where(target), What: "assigning a non-tuple array's length: array writes must preserve its existing indices", Fix: "build a new array with slice, or use push, pop or splice"}
}

func (l *lowering) nonTupleArray(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !l.nonTupleArray(member) {
				return false
			}
		}
		return true
	}
	return !checker.IsTupleType(proven) && l.isArrayType(proven)
}
