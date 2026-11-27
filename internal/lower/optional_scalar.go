package lower

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A ?. read whose result the checker calls a number or a boolean is held as the maybe it is (#0d56xpd,
// Adamic's rule (d), Nov 26 2026). tsc types parent?.kind as a number where parent is declared present, yet
// ?. tests its receiver as it runs, whatever the checker says, and a hatch can leave that receiver absent:
// tsc's own Node constructors write this.parent = undefined!, and a SourceFile's parent is never set. So the
// read is the receiver's test and the field's value, a number | undefined, and holding it as one trusts
// nothing. A use that observes the value keeps the maybe. A bitwise operator converts it as JavaScript does,
// undefined to NaN, which its ToInt32 makes 0. Every other use needs the scalar and takes it through a
// check, which stops loudly where the receiver was absent, in words that name the hatch as the cause: a
// valid program gets Node's answer, and a receiver the checker calls present can be absent only because a
// non-null assertion wrote it so.

// optionalScalarAbsent is the check's words, Adamic's on #0d56xpd.
const optionalScalarAbsent = "a ?. read whose receiver the checker calls present was absent, through a non-null or definite-assignment assertion"

// optionalScalarRead is node's use of read, its ?. read of a field stored as of, a number or a boolean.
func (l *lowering) optionalScalarRead(node *ast.Node, read ir.Expression, of ir.Type) ir.Expression {
	if l.acceptsUndefined(node) || optionalScalarObserved(node) {
		return read
	}
	if of == ir.Number && bitwiseOperand(node) {
		return ir.Coalesce{Value: read, Fallback: ir.NumberConstant{Value: math.NaN()}, Of: ir.Number}
	}
	message := optionalScalarAbsent + ": " + sourceExpression(node)
	return ir.Coalesce{Value: read, Panic: ir.StringConstant{Index: l.constant(message)}, Of: of}
}

// optionalScalarObserved reports the uses that observe a maybe as it is, which acceptsUndefined doesn't
// list since a narrowed variable or field read there is unwrapped today: a strict comparison, the left of
// ||, a condition, a !, and a switch's discriminant, each of which the lowering already takes a maybe
// into (combine's pair comparison, ir.Logical, censusCondition, nullishNumericSwitch). They're asked only
// for a ?. read here, so a narrowed read keeps its unwrap. Parentheses don't change the use.
func optionalScalarObserved(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return true
		case ast.KindBarBarToken:
			return binary.Left == node
		}
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == node
	case ast.KindSwitchStatement:
		return parent.AsSwitchStatement().Expression == node
	}
	return false
}

// bitwiseOperand reports whether node is an operand of a binary bitwise operator, which converts it with
// ToNumber and then ToInt32, as written through parentheses.
func bitwiseOperand(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	if node.Parent == nil || node.Parent.Kind != ast.KindBinaryExpression {
		return false
	}
	_, isBitwise := bitwise[node.Parent.AsBinaryExpression().OperatorToken.Kind]
	return isBitwise
}
