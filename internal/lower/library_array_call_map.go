package lower

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Port the non-array ArraySpeciesCreate failure from V8 ArrayMap in
// src/builtins/array-map.tq. A proven oversized length must throw before the
// first callback; no sparse result or element representation is needed.
func (l *lowering) libraryArrayMapOversized(node *ast.Node, written []*ast.Node) (ir.Expression, error) {
	if len(written) < 2 || !l.libraryArrayOversizedLength(written[0]) {
		return nil, l.notYet(node, "map.call without a proven immutable oversized length (generic mapping requires sparse presence)")
	}
	arguments := []ir.Expression{}
	for _, argument := range written {
		if argument.Kind == ast.KindSpreadElement {
			return nil, l.notYet(node, "map.call with spread arguments")
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, value)
	}
	// V8 gets length before validating the callback. The proven primitive
	// numeric field read has no effects; all supplied argument effects run first.
	if arguments[1].Type() != ir.Closure {
		return nil, l.notYet(node, "map.call without a proven callable callback")
	}
	b := l.libraryArrayBuilder(arguments)
	local := b.local("range_error", ir.Object)
	b.body = append(b.body,
		ir.Declare{Local: local, Value: ir.MakeError{Message: ir.StringConstant{Index: l.constant("Invalid array length")}}},
		ir.SetProperty{Object: b.read(local), Name: "name", Value: ir.StringConstant{Index: l.constant("RangeError")}, Site: l.libraryArraySyntheticWriteSite(node)},
		ir.Throw{Value: b.read(local)},
	)
	return b.finish("array_map_oversized", ir.Undefined{}), nil
}

func (l *lowering) libraryArrayOversizedLength(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.AsVariableDeclaration().Type != nil || !l.libraryArrayMapReceiverUses(declaration) {
			return false
		}
		node = declaration.AsVariableDeclaration().Initializer
		if node == nil {
			return false
		}
		node = ast.SkipParentheses(node)
	}
	if node.Kind != ast.KindObjectLiteralExpression || !l.libraryArrayExactShape(node) {
		return false
	}
	var value *ast.Node
	for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind != ast.KindPropertyAssignment {
			return false
		}
		name, known := l.libraryArrayLikeFieldName(field.Name())
		if !known {
			return false
		}
		if name == "length" {
			if value != nil {
				return false
			}
			value = field.AsPropertyAssignment().Initializer
		}
	}
	if value == nil {
		return false
	}
	length, known := l.libraryArrayConstantNumber(value)
	return known && !math.IsNaN(length) && math.Trunc(length) > 4294967295
}

// The receiver may only be passed to map.call. No alias, property write or
// escape can change its length between initialization and argument evaluation.
func (l *lowering) libraryArrayMapReceiverUses(declaration *ast.Node) bool {
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsBlockScoped == 0 {
		return false
	}
	statement := list.Parent
	if statement == nil || statement.Kind != ast.KindVariableStatement || ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) || statement.Parent == nil || statement.Parent.Kind != ast.KindSourceFile && statement.Parent.Kind != ast.KindBlock {
		return false
	}
	symbol := l.symbol(declaration.Name())
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false
	}
	proven := true
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if !proven {
			return true
		}
		if part.Kind == ast.KindShorthandPropertyAssignment {
			value := l.checker.GetShorthandAssignmentValueSymbol(part)
			if value != nil && l.checker.GetExportSymbolOfSymbol(value) == symbol {
				proven = false // A shorthand stores the value, not its property symbol.
				return true
			}
		}
		if ast.IsIdentifier(part) && l.symbol(part) == symbol && part != declaration.Name() {
			outer := libraryArrayOuter(part)
			parent := outer.Parent
			if parent != nil && parent.Kind == ast.KindTypeQuery {
				return false
			}
			if !libraryArrayInitializedUse(declaration, part) || parent == nil || parent.Kind != ast.KindCallExpression {
				proven = false
				return true
			}
			call := parent.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if len(call.Arguments.Nodes) == 0 || call.Arguments.Nodes[0] != outer || callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "call" || callee.AsPropertyAccessExpression().QuestionDotToken != nil || l.libraryArrayExplicitMethod(callee.AsPropertyAccessExpression().Expression) != "map" {
				proven = false
			}
			return false
		}
		return part.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return proven
}

// Only exact numeric expressions are admitted. In particular, arbitrary Go
// math.Pow folding is not a proof of V8's last-bit result near the length bound.
func (l *lowering) libraryArrayConstantNumber(node *ast.Node) (float64, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNumericLiteral {
		value, err := l.numericLiteral(node)
		if number, known := value.(ir.NumberConstant); known && err == nil {
			return number.Value, true
		}
	}
	if l.isLibraryGlobal(node, "Infinity") {
		return math.Inf(1), true
	}
	if node.Kind == ast.KindPrefixUnaryExpression {
		unary := node.AsPrefixUnaryExpression()
		value, known := l.libraryArrayConstantNumber(unary.Operand)
		if known && unary.Operator == ast.KindMinusToken {
			return -value, true
		}
		if known && unary.Operator == ast.KindPlusToken {
			return value, true
		}
	}
	power := func(base, exponent float64) (float64, bool) {
		if base == 2 && exponent == math.Trunc(exponent) && exponent >= 0 && exponent <= 1023 {
			return math.Ldexp(1, int(exponent)), true
		}
		return 0, false
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		left, a := l.libraryArrayConstantNumber(binary.Left)
		right, b := l.libraryArrayConstantNumber(binary.Right)
		if a && b {
			switch binary.OperatorToken.Kind {
			case ast.KindPlusToken:
				return left + right, true
			case ast.KindMinusToken:
				return left - right, true
			case ast.KindAsteriskToken:
				return left * right, true
			case ast.KindSlashToken:
				return left / right, true
			case ast.KindAsteriskAsteriskToken:
				return power(left, right)
			}
		}
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "pow" && callee.AsPropertyAccessExpression().QuestionDotToken == nil && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Math") && len(call.Arguments.Nodes) == 2 {
			base, a := l.libraryArrayConstantNumber(call.Arguments.Nodes[0])
			exponent, b := l.libraryArrayConstantNumber(call.Arguments.Nodes[1])
			if a && b {
				return power(base, exponent)
			}
		}
	}
	return 0, false
}
