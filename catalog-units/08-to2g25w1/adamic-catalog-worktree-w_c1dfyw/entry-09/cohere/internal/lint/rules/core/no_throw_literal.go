package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoThrowLiteralObject = rule.Message{
	Id: "object",
	Description: "This throws something that cannot be an Error, so whatever catches it gets no " +
		"message and no stack trace. The failure then surfaces as a bare string or object with no " +
		"record of where it came from, which is the difference between reading one line and " +
		"reading the whole call path. Throw `new Error(...)` instead.",
}

var messageNoThrowLiteralUndef = rule.Message{
	Id: "undef",
	Description: "This throws `undefined`, which reaches the handler carrying nothing at all: no " +
		"message, no stack, no type. It is almost always an expression that was meant to produce " +
		"an error and did not. Throw `new Error(...)` instead.",
}

// NoThrowLiteral flags a `throw` whose argument cannot be an Error object.
//
//	valid:   throw new Error();
//	valid:   throw a;
//	valid:   throw foo ? new Error() : 'literal';
//	valid:   throw 'literal' && new Error();
//	invalid: throw 'error';
//	invalid: throw undefined;
//	invalid: throw foo ? 'not an Error' : 'literal';
//
// # This is the SYNTACTIC sibling of typescript/only-throw-error, not a duplicate of it
//
// Both rules exist upstream and both are ported here. `only-throw-error` asks the type checker what
// the thrown expression's type is, so it can tell `throw someError` from `throw someString` when
// both are identifiers. This one has no types at all: an identifier is "could be an Error" purely
// because it might hold one, which is why `throw a;` is clean here and can report there.
//
// The consequence is that this rule is strictly the weaker of the two on identifiers and is not
// subsumed by it, because it fires on shapes that need no types to condemn.
//
// # The judgment is `couldBeError`, and it is optimistic by construction
//
// Upstream asks whether the argument has ANY possibility of being an Error and reports only when it
// provably cannot be. Eight node types answer yes outright because their value is not knowable
// syntactically, and everything else is either recursion or a no.
//
// The recursion is where the corpus does its work, and each arm is a separate decision:
//
//	=  and &&=   the result is the right operand           throw foo = new Error()  is CLEAN
//	||= and ??=  the result is either operand              throw foo.bar ||= 'literal'  is CLEAN
//	other  op=   arithmetic or bitwise, so a primitive     throw foo += new Error()  REPORTS
//	comma        the result is the last expression         throw new Error(), 1, 2, 3  REPORTS
//	&&           short-circuits to a falsy left, so only   throw 'literal' && new Error()  is CLEAN
//	             the right operand can be the value        throw foo && 'literal'  REPORTS
//	|| and ??    either operand can be the value           throw new Error() || 'literal'  is CLEAN
//	?:           either branch can be the value            throw foo ? 'x' : 'y'  REPORTS
//
// `throw foo += new Error()` reporting while `throw foo = new Error()` is clean is the pair that
// makes the operator split load-bearing rather than decorative, and both are upstream's own cases.
//
// # Our tree folds four of upstream's node types into one
//
// Measured by probing the parse of every corpus case: typescript-go gives `KindBinaryExpression` to
// assignment, comma, and the logical operators alike, discriminating them by the operator token
// rather than by the node kind. Upstream reads three different node types there. So the arms below
// are keyed on the operator token, and the three families are separated inside one case rather than
// by three cases.
//
// `ChainExpression` has no counterpart at all: `obj?.foo` parses as a plain property access and
// `obj?.foo()` as a plain call, so both are already answered by the accessor arms and upstream's
// dedicated case has nothing to map onto. Measured rather than assumed.
//
// # `undefined` gets its own message, and only as a bare global identifier
//
// An identifier is "could be an Error", so `throw undefined` reaches the second test rather than the
// first. It reports `undef` only when the name resolves to the global; upstream asks
// `isGlobalReference` and the corpus pins it with `function foo(undefined) { throw undefined; }`,
// which is CLEAN because the parameter shadows the global and could hold anything.
var NoThrowLiteral = rule.Rule{
	Name:             "no-throw-literal",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				argument := node.AsThrowStatement().Expression
				if argument == nil {
					// `throw;` is a syntax error the parser recovers from. Upstream never sees it,
					// because its parser refuses the file; ours hands back a throw with no argument
					// and dereferencing it would take the run down.
					return
				}

				if !couldBeError(argument) {
					ctx.ReportNode(node, messageNoThrowLiteralObject)
					return
				}

				// Reached only for an argument that could be an Error, which an identifier always
				// is. Upstream orders these the same way, so `throw undefined` never produces both.
				if argument.Kind == ast.KindIdentifier && argument.Text() == "undefined" &&
					!undefinedIsShadowed(ctx, argument) {
					ctx.ReportNode(node, messageNoThrowLiteralUndef)
				}
			},
		}
	},
}

// couldBeError reports whether an expression has any possibility of evaluating to an Error.
//
// This is upstream's `ast-utils.couldBeError`, arm for arm. It is deliberately optimistic: it
// answers true for anything whose value is not knowable from the syntax, so the rule reports only
// what provably cannot be an Error. Widening any arm turns a clean case into a false positive on
// code that is correct, which is why the eight always-true kinds are enumerated rather than reduced
// to "not a literal".
func couldBeError(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	// The eight kinds upstream answers true for outright, because none of them has a value the
	// syntax can rule out. Our `PropertyAccessExpression` and `ElementAccessExpression` are one
	// `MemberExpression` upstream, and both spellings of an optional chain land here too, which is
	// what makes upstream's separate `ChainExpression` case unnecessary rather than missing.
	case ast.KindIdentifier,
		ast.KindCallExpression,
		ast.KindNewExpression,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindTaggedTemplateExpression,
		ast.KindYieldExpression,
		ast.KindAwaitExpression:
		return true

	case ast.KindBinaryExpression:
		// One kind here carries three of upstream's node types, separated by operator token rather
		// than by kind. Measured on the parse of every corpus case; see the rule's doc comment.
		binaryExpression := node.AsBinaryExpression()
		if binaryExpression == nil || binaryExpression.OperatorToken == nil {
			return false
		}

		switch binaryExpression.OperatorToken.Kind {
		// `=` and `&&=` evaluate to the right operand, so that is the only thing the throw can see.
		case ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
			return couldBeError(binaryExpression.Right)

		// `||=` and `??=` evaluate to whichever side wins, so either can reach the throw.
		case ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return couldBeError(binaryExpression.Left) || couldBeError(binaryExpression.Right)

		// `&&` short-circuits on a falsy left operand, and nothing falsy is an Error, so only the
		// right operand can ever be the thrown value.
		case ast.KindAmpersandAmpersandToken:
			return couldBeError(binaryExpression.Right)

		// `||` and `??` can yield either side.
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return couldBeError(binaryExpression.Left) || couldBeError(binaryExpression.Right)

		// The comma operator evaluates to its last expression. Our parser nests `1, 2, 3` as
		// left-associated pairs rather than as one flat list, so "the last expression" is the right
		// operand of the outermost comma, which this reaches by recursion and upstream reaches by
		// indexing. Upstream's `exprs.length !== 0` guard has no counterpart because a binary
		// expression always has two operands.
		case ast.KindCommaToken:
			return couldBeError(binaryExpression.Right)
		}

		// Every remaining binary operator is arithmetic, bitwise, relational or a comparison, and
		// all of them evaluate to a primitive or throw. That covers upstream's explicit "all other
		// assignment operators" case and its `default: return false` in one place.
		return false

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		if conditional == nil {
			return false
		}
		return couldBeError(conditional.WhenTrue) || couldBeError(conditional.WhenFalse)
	}

	return false
}
