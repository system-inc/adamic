package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreferSpread = rule.Message{
	Id: "preferSpread",
	Description: "This `.apply()` call passes a `this` that is already the receiver, so the only " +
		"thing it is doing is spreading an array into the argument list. Spread syntax says that " +
		"directly: `foo(...args)` instead of `foo.apply(null, args)`. The `.apply()` form also " +
		"breaks the moment someone adds a method named `apply` to the receiver, and it hides the " +
		"call target from every tool that reads call sites.",
}

// PreferSpread flags a `.apply()` call that spread syntax expresses directly.
//
//	valid:   foo(...args)
//	valid:   foo.apply(obj, args)
//	valid:   obj.foo.apply(null, args)
//	valid:   foo.apply(null, [1, 2, 3])
//	invalid: foo.apply(undefined, args)
//	invalid: foo.apply(null, args)
//	invalid: obj.foo.apply(obj, args)
//
// `.apply()` exists to do two things at once: set `this`, and spread an array into the argument
// list. Where the `this` it sets is the one the call would have had anyway, only the spreading is
// left, and spread syntax does that in fewer characters while keeping the call target visible.
//
// So the rule fires only where the `this` argument is redundant, which is two cases. For a callee
// that is a plain identifier there is no implicit receiver, so `null` and `undefined` change nothing.
// For a callee reached through a member access the receiver is the object it was reached through, so
// passing that same object changes nothing.
//
// "That same object" is decided by comparing how both are written, not by what they evaluate to.
// That is ESLint's `equalTokens` oracle and it is the only sound one available: `a[i++].foo.apply(
// a[i++], args)` is written identically on both sides and is not the same object, which ESLint's own
// documentation names as a known limitation. Ours inherits it deliberately rather than trying to be
// cleverer, because the alternative is evaluating expressions a linter cannot evaluate.
//
// A call whose second argument is an array literal is left alone. Rewriting `foo.apply(null, [1, 2])`
// to `foo(1, 2)` is a real improvement, but it is `no-useless-call`'s improvement: the defect there
// is building an array to immediately take apart, not the `.apply()` at all. Splitting them keeps
// each rule's message true.
//
// Divergence from ESLint, in the stricter direction. ESLint parses to ESTree, which has no node for
// parentheses, so it sees `(a?.b).c` and `a?.b.c` as the same token stream and reports. Ours keeps
// the parenthesized node, and those two are genuinely different: `(a?.b).c` throws when `a` is null
// while `a?.b.c` short-circuits, so treating them as interchangeable is what would be wrong. The
// parentheses are stripped where they wrap the whole callee or the whole receiver, since those are
// pure grouping, and kept where they sit inside the compared expressions.
//
// No fix. The rewrite has to move the spread into the argument list and delete the `this` argument,
// and where the receiver expression has side effects, `a[i++].foo.apply(a[i++], args)` again, the
// rewrite changes how many times they run.
var PreferSpread = rule.Rule{
	Name: "prefer-spread",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpression := node.AsCallExpression()
				if callExpression == nil || callExpression.Arguments == nil {
					return
				}

				// Exactly two arguments is what `.apply()` takes, and anything else is either a
				// different call or one whose meaning the rewrite would not preserve.
				arguments := callExpression.Arguments.Nodes
				if len(arguments) != 2 {
					return
				}

				callee := ast.SkipParentheses(callExpression.Expression)
				if callee == nil || !isApplyMemberAccess(callee) {
					return
				}

				// An array literal is `no-useless-call`'s concern, and a spread already there means
				// the argument count is not what it looks like. Parentheses are stripped first
				// because ESTree has none, so `foo.apply(null, ([1, 2]))` has to read as the array
				// literal it is.
				secondArgument := ast.SkipParentheses(arguments[1])
				if secondArgument == nil ||
					secondArgument.Kind == ast.KindArrayLiteralExpression ||
					secondArgument.Kind == ast.KindSpreadElement {
					return
				}

				// `memberAccessObject` returns nil for a callee that is neither a property nor an
				// element access, and `ast.SkipParentheses` dereferences its argument, so passing
				// that nil straight in would panic. It cannot happen here: the
				// `isApplyMemberAccess(callee)` gate above has already excluded every other kind.
				//
				// The nil check therefore sits after the call rather than before it, guarding the
				// unwrapped result rather than the input. That is safe by an invariant established
				// fifteen lines earlier and by nothing local, so **widening `isApplyMemberAccess`
				// to accept another callee kind reintroduces a panic here**, at a line the change
				// does not touch. A panic takes down the walk for the whole file, so every other
				// rule's verdict on that file is lost with it.
				appliedFunction := ast.SkipParentheses(memberAccessObject(callee))
				if appliedFunction == nil {
					return
				}

				// Parentheses around the `this` argument are pure grouping and ESTree has no node
				// for them, so `apply((obj), args)` has to compare equal to `apply(obj, args)`.
				// Stripping here rather than inside the comparison keeps the two sides symmetric:
				// the receiver was already unwrapped above.
				thisArgument := ast.SkipParentheses(arguments[0])
				if thisArgument == nil {
					return
				}

				// A callee that is not itself a member access has no implicit receiver, so the only
				// `this` values that change nothing are the empty ones.
				expectedReceiver := memberAccessObject(appliedFunction)
				if expectedReceiver == nil {
					if !isNullOrUndefined(thisArgument) {
						return
					}
					ctx.ReportNode(node, messagePreferSpread)
					return
				}

				if !hasSameTokens(ctx.SourceFile, expectedReceiver, thisArgument) {
					return
				}

				ctx.ReportNode(node, messagePreferSpread)
			},
		}
	},
}

// isApplyMemberAccess says whether an expression reads the property `apply` off something.
//
// Both access forms count, because `foo['apply'](null, args)` is the same call as `foo.apply(null,
// args)`. A computed access with a non-literal key does not, since `foo[name]` is only `apply` if
// `name` happens to hold that string, which is exactly what a linter cannot know.
//
// The private-identifier check is belt and braces rather than load-bearing: `#apply` is a different
// property that merely spells the same word, and typescript-go already keeps the sigil in the
// identifier's text, so the name comparison excludes it on its own. Measured rather than assumed,
// after a mutation removing the check changed no test outcome. It stays because the comparison it
// guards is against a bare name, and a reader has no way to see from here that the sigil survives
// into it.
func isApplyMemberAccess(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		propertyAccess := node.AsPropertyAccessExpression()
		if propertyAccess == nil || propertyAccess.Name() == nil {
			return false
		}
		if propertyAccess.Name().Kind == ast.KindPrivateIdentifier {
			return false
		}
		return propertyAccess.Name().Text() == "apply"

	case ast.KindElementAccessExpression:
		elementAccess := node.AsElementAccessExpression()
		if elementAccess == nil || elementAccess.ArgumentExpression == nil {
			return false
		}
		argument := ast.SkipParentheses(elementAccess.ArgumentExpression)
		if argument == nil || !ast.IsStringLiteralLike(argument) {
			return false
		}
		return argument.Text() == "apply"
	}
	return false
}

// memberAccessObject returns the object a member access reads through, or nil when the node is not
// a member access.
//
// Nil is the answer that matters as much as the node: it is how the rule distinguishes a bare
// identifier callee, which has no implicit receiver, from a member access, which has one.
func memberAccessObject(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		propertyAccess := node.AsPropertyAccessExpression()
		if propertyAccess == nil {
			return nil
		}
		return propertyAccess.Expression
	case ast.KindElementAccessExpression:
		elementAccess := node.AsElementAccessExpression()
		if elementAccess == nil {
			return nil
		}
		return elementAccess.Expression
	}
	return nil
}

// isNullOrUndefined says whether an expression is one of the three ways to write "no receiver".
//
// `void 0` is included alongside `null` and `undefined` because it is the same value and appears in
// exactly this position in minified and defensive code, where `undefined` might be shadowed.
func isNullOrUndefined(node *ast.Node) bool {
	unwrapped := ast.SkipParentheses(node)
	if unwrapped == nil {
		return false
	}
	switch unwrapped.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		return unwrapped.Text() == "undefined"
	case ast.KindVoidExpression:
		voidExpression := unwrapped.AsVoidExpression()
		return voidExpression != nil && voidExpression.Expression != nil &&
			ast.SkipParentheses(voidExpression.Expression).Kind == ast.KindNumericLiteral
	}
	return false
}

// hasSameTokens says whether two expressions are written with the same tokens.
//
// This is ESLint's `equalTokens` oracle, and tokenSignature is almost it. The one gap is a node with
// no children, which the signature records as its own source text verbatim so that two string
// literals differing in whitespace stay distinct. That is right for a literal and wrong for an empty
// bracket pair: `[]` and `[ ]` are the same empty array written two ways, and ESLint's token stream
// sees `[` `]` for both. Comparing an empty ArrayLiteral or ObjectLiteral by kind alone closes it
// without touching the literal case, which still needs its text.
//
// Found by a fixture, not by reading: `[].concat.apply([ ], args)` is in ESLint's own suite and this
// rule stayed silent on it until the signatures were printed side by side.
func hasSameTokens(sourceFile *ast.SourceFile, left *ast.Node, right *ast.Node) bool {
	if isEmptyBracketLiteral(left) && isEmptyBracketLiteral(right) {
		return left.Kind == right.Kind
	}
	return tokenSignature(sourceFile, left) == tokenSignature(sourceFile, right)
}

// isEmptyBracketLiteral says whether a node is `[]` or `{}` with nothing inside.
func isEmptyBracketLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrayLiteralExpression:
		arrayLiteral := node.AsArrayLiteralExpression()
		return arrayLiteral != nil &&
			(arrayLiteral.Elements == nil || len(arrayLiteral.Elements.Nodes) == 0)
	case ast.KindObjectLiteralExpression:
		objectLiteral := node.AsObjectLiteralExpression()
		return objectLiteral != nil &&
			(objectLiteral.Properties == nil || len(objectLiteral.Properties.Nodes) == 0)
	}
	return false
}
