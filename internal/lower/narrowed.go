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
// type and may have it put back by a call. Any call can write a field. A variable only one written
// inside a function other than its own can: a global a function assigns, a let a closure assigns.
// A parameter or a local nothing else writes keeps its narrowing, and is read plainly, which is
// what lets reuse in place take it over. An array element always may be undefined
// (noUncheckedIndexedAccess), and its caller says so.
func (l *lowering) declaredUndefined(node *ast.Node) bool {
	at := node
	if node.Kind == ast.KindPropertyAccessExpression {
		at = node.Name()
	}
	symbol := l.checker.GetSymbolAtLocation(at)
	if symbol == nil || !l.includesUndefined(l.checker.GetTypeOfSymbol(symbol)) {
		return false
	}
	return node.Kind == ast.KindPropertyAccessExpression || l.writtenElsewhere[l.symbol(at)]
}

// findWritesElsewhere notes every variable a module writes (=, a compound assignment, ++ or --) from
// inside a function other than the one that declares it: those are the narrowings a call can undo.
// It runs over every module before anything is lowered, since a closure lowered after a read can
// still run before it.
func (l *lowering) findWritesElsewhere(module *ast.SourceFile) {
	if l.writtenElsewhere == nil {
		l.writtenElsewhere = map[*ast.Symbol]bool{}
	}
	note := func(target *ast.Node, at *ast.Node) {
		target = ast.SkipParentheses(target)
		if target.Kind != ast.KindIdentifier {
			// A destructuring assignment writes every name in it.
			if target.Kind == ast.KindArrayLiteralExpression || target.Kind == ast.KindObjectLiteralExpression {
				var names ast.Visitor
				names = func(child *ast.Node) bool {
					if child.Kind == ast.KindIdentifier {
						l.noteWrite(child, at)
					}
					return child.ForEachChild(names)
				}
				target.ForEachChild(names)
			}
			return
		}
		l.noteWrite(target, at)
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindBinaryExpression:
			if binary := node.AsBinaryExpression(); ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				note(binary.Left, node)
			}
		case ast.KindPrefixUnaryExpression:
			if prefix := node.AsPrefixUnaryExpression(); prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken {
				note(prefix.Operand, node)
			}
		case ast.KindPostfixUnaryExpression:
			note(node.AsPostfixUnaryExpression().Operand, node)
		}
		return node.ForEachChild(visit)
	}
	module.AsNode().ForEachChild(visit)
}

// noteWrite notes identifier's variable when at, the write, is in another function than its
// declaration.
func (l *lowering) noteWrite(identifier *ast.Node, at *ast.Node) {
	symbol := l.symbol(identifier)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return
	}
	if enclosingFunction(at) != enclosingFunction(symbol.Declarations[0]) {
		l.writtenElsewhere[symbol] = true
	}
}

// enclosingFunction is the function-like node around node, or nil at a module's top level.
func enclosingFunction(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLike(current) {
			return current
		}
	}
	return nil
}

// checkNarrowed checks value, read at node, where the checker has narrowed undefined out of what's
// declared to hold it (declared says it is). A number or a boolean that may be undefined as stage 0
// holds it is unwrapped, checked (ir.Unwrap); a reference is checked as it is (ir.Defined). Read
// through a property next, a reference panics with the TypeError JavaScript throws there, word for
// word; anywhere else, with Adamic's own words, since JavaScript would go on with undefined in a
// place typed not to hold it.
func (l *lowering) checkNarrowed(node *ast.Node, value ir.Expression, declared bool) ir.Expression {
	here := l.checker.GetTypeAtLocation(node)
	if l.includesUndefined(here) || l.mayHoldUndefined(node) {
		return value
	}
	if value.Type().IsMaybe() {
		// Read as what it holds wherever the checker narrowed it, and checked where a call may have
		// put undefined back.
		if narrowed, isKnown := l.representation(here); isKnown && narrowed == value.Type().Present() {
			return ir.Unwrap{Value: value, Checked: declared}
		}
		return value
	}
	if !declared || !value.Type().IsReference() || value.Type() == ir.Union {
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
