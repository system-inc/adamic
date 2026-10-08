// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

// assignment lowers =, and the compound assignments, to a local.
func (l *lowering) assignment(node *ast.Node) ([]ir.Statement, error) {
	binary := node.AsBinaryExpression()
	operator, isCompound := compoundAssignments[binary.OperatorToken.Kind]
	if binary.OperatorToken.Kind != ast.KindEqualsToken && !isCompound {
		return nil, l.notYet(node, describe(node)+" as a statement")
	}
	target := ast.SkipParentheses(binary.Left)
	if isCompound && l.enumNeverIdentity(target, map[*ast.Node]bool{}) != nil {
		value, err := l.expression(target)
		return []ir.Statement{ir.Evaluate{Value: value}}, err
	}
	if target.Kind == ast.KindPropertyAccessExpression {
		if isCompound {
			return l.updateProperty(node, target, operator, binary.Right)
		}
		return l.setProperty(target, binary.Right)
	}
	// array[index] += value never reaches here: the element may be missing, so the checker refuses it
	// (noUncheckedIndexedAccess).
	if target.Kind == ast.KindElementAccessExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		if body, handled, err := l.typedArrayWrite(target, binary.Right); handled {
			return body, err
		}
		return l.setIndex(target, binary.Right)
	}
	if target.Kind == ast.KindArrayLiteralExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		return l.destructuringAssignment(target, binary.Right)
	}
	if target.Kind == ast.KindElementAccessExpression && isCompound {
		return l.updateIndex(node, target, operator, binary.Right)
	}
	local, isLocal := l.local(target)
	if !ast.IsIdentifier(target) || !isLocal {
		return nil, l.notYet(target, "assigning to "+describe(target))
	}
	if l.caught[l.symbol(target)] {
		return nil, l.notYet(target, "assigning to what a catch caught")
	}
	if l.alwaysUndefined[l.symbol(target)] {
		// Its type is unknown, so anything could be written to it, and it holds only undefined.
		return nil, l.notYet(target, "assigning to a parameter that only ever receives undefined")
	}
	value, err := l.expression(binary.Right)
	if err != nil {
		return nil, err
	}
	if isCompound {
		current := ir.Expression(ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checked(local)})
		if operator == ast.KindPlusToken {
			current, value = l.spelled(target, current), l.spelled(binary.Right, value)
		}
		if value, err = l.combine(node, operator, current, value); err != nil {
			return nil, err
		}
	}
	return []ir.Statement{ir.Assign{Local: local, Value: fit(value, l.result.Locals[local].Type), Checked: l.checked(local)}}, nil
}

// tupleField reads element index of the tuple held in the local held, as the tuple's element type
// holds it, and makes it what the name it goes to holds: a string element going to a string | number
// name is boxed on the way.
func (l *lowering) tupleField(where *ast.Node, held int, elements []*checker.Type, index int, to ir.Type) (ir.Expression, error) {
	if index >= len(elements) {
		return nil, l.notYet(where, "a name past its tuple's elements")
	}
	of, isKnown := l.representation(elements[index])
	if !isKnown || slotless(of) {
		return nil, l.notYet(where, "a tuple element of type "+l.checker.TypeToString(elements[index]))
	}
	value := fit(ir.Expression(ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: strconv.Itoa(index), Of: of}), to)
	if value.Type() != to {
		return nil, l.notYet(where, "a tuple element of type "+l.checker.TypeToString(elements[index])+" given to a "+typeName(to))
	}
	return value, nil
}

// destructuringAssignment lowers [a, b] = tuple: the tuple is evaluated whole and held, then each
// name is assigned its field, in order, as JavaScript assigns them. A name left out ([, b]) is
// skipped. Anything but plain names, or a value that isn't a tuple, is not lowered yet.
func (l *lowering) destructuringAssignment(pattern *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	if !checker.IsTupleType(l.checker.GetTypeAtLocation(valueNode)) {
		return nil, l.notYet(pattern, "destructuring other than a tuple")
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Object {
		return nil, l.notYet(pattern, "destructuring a "+typeName(value.Type()))
	}
	elements := l.checker.GetTypeArguments(l.checker.GetTypeAtLocation(valueNode))
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "tuple", Type: ir.Object, Function: l.functionIndex})
	statements := []ir.Statement{ir.Declare{Local: held, Value: value}}
	for index, element := range pattern.AsArrayLiteralExpression().Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			continue
		}
		local, isLocal := l.local(element)
		if !ast.IsIdentifier(element) || !isLocal || l.alwaysUndefined[l.symbol(element)] {
			return nil, l.notYet(element, "assigning to "+describe(element)+" in a destructuring assignment")
		}
		field, err := l.tupleField(element, held, elements, index, l.result.Locals[local].Type)
		if err != nil {
			return nil, err
		}
		statements = append(statements, ir.Assign{Local: local, Value: field, Checked: l.checked(local)})
	}
	// The held tuple is the block's, released when it ends.
	return []ir.Statement{ir.Block{Body: statements}}, nil
}

var compoundAssignments = map[ast.Kind]ast.Kind{
	ast.KindPlusEqualsToken:             ast.KindPlusToken,
	ast.KindMinusEqualsToken:            ast.KindMinusToken,
	ast.KindAsteriskEqualsToken:         ast.KindAsteriskToken,
	ast.KindSlashEqualsToken:            ast.KindSlashToken,
	ast.KindPercentEqualsToken:          ast.KindPercentToken,
	ast.KindAsteriskAsteriskEqualsToken: ast.KindAsteriskAsteriskToken,

	ast.KindAmpersandEqualsToken:                         ast.KindAmpersandToken,
	ast.KindBarEqualsToken:                               ast.KindBarToken,
	ast.KindCaretEqualsToken:                             ast.KindCaretToken,
	ast.KindLessThanLessThanEqualsToken:                  ast.KindLessThanLessThanToken,
	ast.KindGreaterThanGreaterThanEqualsToken:            ast.KindGreaterThanGreaterThanToken,
	ast.KindGreaterThanGreaterThanGreaterThanEqualsToken: ast.KindGreaterThanGreaterThanGreaterThanToken,
}

// increment lowers ++ and -- on a number local or field, as a statement, where prefix and postfix
// agree.
func (l *lowering) increment(node *ast.Node) ([]ir.Statement, error) {
	var operator ast.Kind
	var operand *ast.Node
	if node.Kind == ast.KindPrefixUnaryExpression {
		operator, operand = node.AsPrefixUnaryExpression().Operator, node.AsPrefixUnaryExpression().Operand
	} else {
		operator, operand = node.AsPostfixUnaryExpression().Operator, node.AsPostfixUnaryExpression().Operand
	}
	if operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken {
		return nil, l.notYet(node, describe(node)+" as a statement")
	}
	operand = ast.SkipParentheses(operand)
	if l.enumNeverIdentity(operand, map[*ast.Node]bool{}) != nil {
		value, err := l.expression(operand)
		return []ir.Statement{ir.Evaluate{Value: value}}, err
	}
	if operand.Kind == ast.KindPropertyAccessExpression {
		step := ast.KindPlusToken
		if operator == ast.KindMinusMinusToken {
			step = ast.KindMinusToken
		}
		return l.updateProperty(node, operand, step, nil)
	}
	local, isLocal := l.local(operand)
	if !ast.IsIdentifier(operand) || !isLocal {
		return nil, l.notYet(operand, "incrementing "+describe(operand))
	}
	step := ir.Add
	if operator == ast.KindMinusMinusToken {
		step = ir.Subtract
	}
	current := ir.Read{Local: local, Of: ir.Number, Checked: l.checked(local)}
	return []ir.Statement{ir.Assign{Local: local, Value: ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: 1}}, Checked: l.checked(local)}}, nil
}
