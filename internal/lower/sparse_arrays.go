package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"math"
)

func (l *lowering) newSparseArray(node *ast.Node) (ir.Expression, bool, error) {
	made := node.AsNewExpression()
	if !l.isLibraryGlobal(made.Expression, "Array") {
		return nil, false, nil
	}
	if made.Arguments == nil || len(made.Arguments.Nodes) != 1 {
		return nil, true, l.notYet(node, "new Array with other than one proven length")
	}
	length, err := l.expression(made.Arguments.Nodes[0])
	if err != nil {
		return nil, true, err
	}
	constant, known := l.provenArrayLength(made.Arguments.Nodes[0], length)
	if !known || constant.Value < 0 || constant.Value > 4294967295 || math.Trunc(constant.Value) != constant.Value || math.IsNaN(constant.Value) {
		return nil, true, l.notYet(node, "new Array with a length not statically proven valid")
	}
	element, err := l.elementType(node)
	if err != nil {
		return nil, true, err
	}
	return ir.ArrayHoles{Length: length, Element: element}, true, nil
}

func (l *lowering) sparseCall(node *ast.Node) error {
	if !l.program.ContainsSparseArrays() {
		return nil
	}
	if hasSpread(node) {
		return l.notYet(node, "spread arguments in a program with sparse arrays")
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	if (l.isLibraryGlobal(receiver, "Array") && callee.Name().Text() == "from") || ((l.isLibraryGlobal(receiver, "Map") || l.isLibraryGlobal(receiver, "Object")) && callee.Name().Text() == "groupBy") {
		for _, argument := range call.Arguments.Nodes {
			if l.checker.IsArrayType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(argument))) {
				return l.notYet(node, "array iteration by a library function in a program with sparse arrays")
			}
		}
	}
	if l.checker.IsArrayType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))) {
		switch callee.Name().Text() {
		case "fill", "push", "at":
			return nil
		case "indexOf", "lastIndexOf", "includes", "join", "map", "filter", "some", "every", "forEach", "reduce":
			// The host's checkArrayHoles pass owns these established hole contracts.
			return nil
		case "sort":
			if len(call.Arguments.Nodes) == 0 {
				if element, err := l.elementType(receiver); err == nil && element == ir.String {
					return nil
				}
			}
		}
		return l.notYet(node, "array methods other than fill, push and at in a program with sparse arrays")
	}
	return nil
}

func (l *lowering) sparseExpression(node *ast.Node) error {
	if !l.program.ContainsSparseArrays() {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindArrayLiteralExpression {
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			if element.Kind == ast.KindSpreadElement {
				return l.notYet(node, "array spread in a program with sparse arrays")
			}
		}
	}
	if node.Kind == ast.KindNewExpression {
		made := node.AsNewExpression()
		if made.Arguments != nil && (l.isLibraryGlobal(made.Expression, "Map") || l.isLibraryGlobal(made.Expression, "Set")) {
			for _, argument := range made.Arguments.Nodes {
				if l.checker.IsArrayType(l.checker.GetTypeAtLocation(argument)) {
					return l.notYet(node, "collection construction from arrays in a program with sparse arrays")
				}
			}
		}
	}
	return nil
}

// A non-const enum member still emits its runtime field read. Validate its
// immutable declaration value without erasing receiver effects or enum readiness.
func (l *lowering) provenArrayLength(node *ast.Node, value ir.Expression) (ir.NumberConstant, bool) {
	if constant, known := value.(ir.NumberConstant); known {
		return constant, true
	}
	if member := l.enumMember(node); member != nil {
		constant, err := l.enumConstant(member)
		if err == nil {
			number, known := constant.(ir.NumberConstant)
			return number, known
		}
	}
	return ir.NumberConstant{}, false
}
