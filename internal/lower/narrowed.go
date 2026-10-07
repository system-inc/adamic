package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
// the checker narrowed undefined out of is checked, the same in both backends.

// narrowedAway reports whether node, a variable or a field, reads a value whose declared type has
// undefined in it while its type here doesn't.
func (l *lowering) narrowedAway(node *ast.Node) bool {
	at := node
	if node.Kind == ast.KindPropertyAccessExpression {
		at = node.Name()
	}
	symbol := l.checker.GetSymbolAtLocation(at)
	if symbol == nil {
		return false
	}
	declared, here := l.checker.GetTypeOfSymbol(symbol), l.checker.GetTypeAtLocation(node)
	return (l.includesUndefined(declared) && !l.includesUndefined(here)) || (l.includesNull(declared) && !l.includesNull(here))
}

// defined checks a reference the checker narrowed undefined out of. Read through a property next,
// it throws the TypeError JavaScript throws there, word for word; anywhere else, panics with Adamic's
// own words, since JavaScript would go on with undefined in a place typed not to hold it.
func (l *lowering) defined(node *ast.Node, value ir.Expression) ir.Expression {
	if !value.Type().IsReference() || value.Type() == ir.Union || !l.narrowedAway(node) || comparedWithUndefined(node) || optionalReceiver(node) {
		return value
	}
	null := false
	at := node
	if node.Kind == ast.KindPropertyAccessExpression {
		at = node.Name()
	}
	if symbol := l.checker.GetSymbolAtLocation(at); symbol != nil {
		null = l.includesNull(l.checker.GetTypeOfSymbol(symbol))
	}
	absent := "undefined"
	if null {
		absent = "null"
	}
	message := absent + " where the checker narrowed it away: a call since the narrowing put it back"
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node {
		if written := parent.Parent; written != nil && written.Kind == ast.KindBinaryExpression && written.AsBinaryExpression().Left == parent && written.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			// object.name = value: JavaScript evaluates the value first and throws at the write, so
			// the check is the write's own (native checks the object there; JavaScript throws).
			return value
		}
		message = "TypeError: Cannot read properties of " + absent + " (reading '" + parent.Name().Text() + "')"
	}
	return ir.Defined{Value: value, Message: message, Null: null}
}

// comparedWithUndefined reports whether node is one side of === or !== with undefined on the other:
// asking whether it's there is the one read that must see undefined as it is, with no check and no
// unwrapping.
func comparedWithUndefined(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		node, parent = parent, parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	if operator := binary.OperatorToken.Kind; operator != ast.KindEqualsEqualsEqualsToken && operator != ast.KindExclamationEqualsEqualsToken {
		return false
	}
	other := binary.Right
	if binary.Right == node {
		other = binary.Left
	}
	other = ast.SkipParentheses(other)
	return other.Kind == ast.KindNullKeyword || (other.Kind == ast.KindIdentifier && other.Text() == "undefined")
}

// optionalReceiver recognizes the operand tested by ?. before a property, element or call.
// Parentheses around that operand do not change which value the optional operation tests.
// The optional operation must see an absent value itself, rather than a narrowing check stopping it.
func optionalReceiver(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		node, parent = parent, parent.Parent
	}
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		access := parent.AsPropertyAccessExpression()
		return access.Expression == node && access.QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		access := parent.AsElementAccessExpression()
		return access.Expression == node && access.QuestionDotToken != nil
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		return call.Expression == node && call.QuestionDotToken != nil
	}
	return false
}
