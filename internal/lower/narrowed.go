package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A narrowing outlives a call. The checker narrows undefined out of a variable at a check (if
// (chain !== undefined)) or an assignment, and keeps the narrowing across a call that assigns the
// variable again, since it doesn't look inside the call:
//
//	let chain: Link | undefined = { value: 1 };
//	function drop(): void { chain = undefined; }
//	if (chain !== undefined) { drop(); console.log(`${chain.value}`); }
//
// Node throws a TypeError there, and native read through a null pointer; with a number | undefined,
// Node computed with undefined (NaN) and native printed a number that was never there. So a read
// the checker narrowed undefined out of is checked, the same in both backends, wherever it goes to
// a place typed not to hold undefined. A variable, a field and an array element are all read so.

// declaredUndefined reports whether node, a variable or a field, is declared with undefined in its
// type. An array element always may be undefined (noUncheckedIndexedAccess), and its caller says so.
func (l *lowering) declaredUndefined(node *ast.Node) bool {
	at := node
	if node.Kind == ast.KindPropertyAccessExpression {
		at = node.Name()
	}
	symbol := l.checker.GetSymbolAtLocation(at)
	return symbol != nil && l.includesUndefined(l.checker.GetTypeOfSymbol(symbol))
}

// checkNarrowed checks value, read at node, where the checker has narrowed undefined out of what's
// declared to hold it (declared says it is). A number or a boolean that may be undefined as stage 0
// holds it is unwrapped, checked (ir.Unwrap); a reference is checked as it is (ir.Defined). Read
// through a property next, a reference panics with the TypeError JavaScript throws there, word for
// word; anywhere else, with Adamic's own words, since JavaScript would go on with undefined in a
// place typed not to hold it.
func (l *lowering) checkNarrowed(node *ast.Node, value ir.Expression, declared bool) ir.Expression {
	if !declared || l.mayHoldUndefined(node) {
		return value
	}
	here := l.checker.GetTypeAtLocation(node)
	if l.includesUndefined(here) {
		return value
	}
	if value.Type().IsMaybe() {
		if narrowed, isKnown := l.representation(here); isKnown && narrowed == value.Type().Present() {
			return ir.Unwrap{Value: value}
		}
		return value
	}
	if !value.Type().IsReference() || value.Type() == ir.Union {
		return value
	}
	message := "undefined where the checker narrowed it away: a call since the narrowing put it back"
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node {
		if written := parent.Parent; written != nil && written.Kind == ast.KindBinaryExpression && written.AsBinaryExpression().Left == parent && written.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			// object.name = value: JavaScript evaluates the value first and throws at the write, so
			// the check is the write's own (native checks the object there; JavaScript throws).
			return value
		}
		message = "TypeError: Cannot read properties of undefined (reading '" + parent.Name().Text() + "')"
	}
	return ir.Defined{Value: value, Message: message}
}

// mayHoldUndefined reports whether the place node's value goes can hold undefined, so JavaScript
// reading undefined there goes on without throwing and without a value of the wrong type: typeof's
// operand, either side of a comparison with undefined, the left of ??, the receiver of ?., and a
// value going straight into a place whose type has undefined in it (an argument, an array element,
// an assignment, an initializer or a return of such a type). There the read is left as it is.
func (l *lowering) mayHoldUndefined(node *ast.Node) bool {
	parent := node.Parent
	inner := node
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		inner, parent = parent, parent.Parent
	}
	if parent != nil {
		switch parent.Kind {
		case ast.KindTypeOfExpression:
			return true
		case ast.KindPropertyAccessExpression:
			access := parent.AsPropertyAccessExpression()
			if access.Expression == inner && access.QuestionDotToken != nil {
				return true
			}
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindQuestionQuestionToken:
				if binary.Left == inner {
					return true
				}
			case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
				other := binary.Right
				if binary.Right == inner {
					other = binary.Left
				}
				if other = ast.SkipParentheses(other); other.Kind == ast.KindIdentifier && other.Text() == "undefined" {
					return true
				}
			}
		}
	}
	// Where the value goes straight into a place of a type: an argument, an array element, an
	// initializer, a return, the right of an assignment, an object's field. Not a branch of ?: or
	// anything else the value only passes through, which is lowered as one type on both sides.
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindArrayLiteralExpression, ast.KindVariableDeclaration,
		ast.KindReturnStatement, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
	case ast.KindBinaryExpression:
		if binary := parent.AsBinaryExpression(); binary.OperatorToken.Kind != ast.KindEqualsToken || binary.Right != inner {
			return false
		}
	default:
		return false
	}
	contextual := l.checker.GetContextualType(inner, checker.ContextFlagsNone)
	return contextual != nil && l.includesUndefined(contextual)
}
