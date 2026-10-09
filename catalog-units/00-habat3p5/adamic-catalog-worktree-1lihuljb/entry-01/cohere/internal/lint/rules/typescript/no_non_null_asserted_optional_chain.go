package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoNonNullOptionalChain = rule.Message{
	Id: "noNonNullOptionalChain",
	Description: "An optional chain returns `undefined` by design when a link is nullish, so " +
		"asserting the result non-null contradicts the reason the `?.` was written. The assertion " +
		"is erased at compile time and the `undefined` still arrives at runtime, which turns a " +
		"guard the author asked for into a crash the types promised could not happen. Handle the " +
		"`undefined` case instead, or drop the `?.` if the value cannot actually be nullish.",
}

var messageSuggestRemovingNonNull = rule.Message{
	Id:          "suggestRemovingNonNull",
	Description: "Remove the non-null assertion.",
}

// NoNonNullAssertedOptionalChain flags a non-null assertion applied to an optional chain.
//
//	valid:   foo?.bar
//	valid:   foo?.bar!.baz
//	valid:   foo?.bar!()
//	invalid: foo?.bar!
//	invalid: foo?.bar()!
//	invalid: (foo?.bar)!
//
// Ported from `@typescript-eslint/no-non-null-asserted-optional-chain`.
//
// # The distinction the rule turns on, which is not "is there a `!` after a `?.`"
//
// `foo?.bar!` is wrong and `foo?.bar!.baz` is fine, and the difference is what the assertion covers.
// In the first the `!` applies to the whole chain, scrubbing the `undefined` the `?.` exists to
// produce. In the second the `!` applies to one link and the chain continues past it, so the chain's
// own short-circuit still governs the result: the `undefined` survives.
//
// A rule that scanned for an assertion following an optional access would report both, and that is
// the port failure this comment exists to prevent. The two shapes are one character apart in source
// and opposite in meaning.
//
// # Translating a selector for a node kind that does not exist here
//
// The original is written against ESTree, where an optional chain is wrapped in a `ChainExpression`
// node, and it registers two selectors naming the two nestings that wrapper can take:
//
//	ChainExpression > TSNonNullExpression      foo?.bar!     the `!` is inside the chain, at its end
//	TSNonNullExpression > ChainExpression      (foo?.bar)!   the `!` is outside, wrapping the chain
//
// Parentheses invert the nesting, which is why the original needs two selectors rather than one.
// typescript-go has no `ChainExpression`: chain membership is a flag on each link. So the two
// selectors become two conditions on the assertion node:
//
//	it carries the chain flag        -> report only when it is the outermost link
//	it does not carry the flag       -> report when what it wraps is a chain, parens skipped
//
// # Which branch each shape actually takes, which is not the obvious mapping
//
// The tempting reading is that `foo?.bar!` takes the first branch and `(foo?.bar)!` takes the
// second. Measured against the parser, that is wrong: for `foo?.bar!` the assertion does **not**
// carry the chain flag, because the flag marks links that a further chain reaches through and
// nothing follows this one. So `foo?.bar!` and `(foo?.bar)!` both arrive at the second condition,
// differing only in whether a paren has to be skipped to find the chain underneath.
//
// The first condition is not redundant. It is what `foo?.bar!.baz` takes: there the assertion does
// carry the flag, because `.baz` reaches through it, and `IsOutermostOptionalChain` is false, so the
// rule correctly declines. That case is the whole point of the rule, and it is handled by the branch
// that never fires on any invalid shape.
//
// This was found by mutation rather than by reading: deleting the second condition broke every
// unparenthesized fixture, which is the opposite of what the selector names predict. The comment is
// written from what the parser does rather than from what the original's selectors suggest.
//
// # Why parentheses need an explicit skip
//
// `(foo?.bar)!` puts a ParenthesizedExpression between the assertion and the chain, and the compiler
// treats parens as ending a chain, so neither the assertion nor the paren carries the chain flag.
// Without `SkipParentheses` the condition finds a paren where it expected a chain and reports
// nothing, which is a silent miss on a shape upstream reports.
//
// # Validation
//
// This translation was checked against fourteen shapes taken from the original's own test file and
// from a run of the real plugin, including both of the mid-chain cases that are valid, and it agreed
// on all fourteen. That answer key is the only real evidence available for this rule: the tree it
// gates contains zero non-null assertions of any kind, so a clean run over it proves nothing.
var NoNonNullAssertedOptionalChain = rule.Rule{
	Name: "@typescript-eslint/no-non-null-asserted-optional-chain",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				if !assertsOverAnOptionalChain(node) {
					return
				}
				// A suggestion rather than a fix, and the original says why at its own site:
				// removing the assertion can obviously break type checking, because the assertion
				// is what was making the surrounding code compile. The value may genuinely need a
				// null check the author now has to write, and only they can decide what it does.
				ctx.ReportNodeWithSuggestions(node, messageNoNonNullOptionalChain, rule.Suggestion{
					Message: messageSuggestRemovingNonNull,
					Fixes:   []rule.Fix{rule.RemoveRange(nonNullAssertionOperatorRange(node))},
				})
			},
		}
	},
}

// assertsOverAnOptionalChain reports whether this assertion covers a whole optional chain, rather
// than one link of a chain that continues past it.
func assertsOverAnOptionalChain(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindNonNullExpression {
		return false
	}

	// The assertion is a link the chain reaches through, as in `foo?.bar!.baz`. Only report when
	// nothing reaches through it, which for this branch means never in practice: a trailing
	// assertion does not carry the flag at all. Kept because it is what declines the mid-chain
	// shapes, and declining them is the rule's entire subtlety.
	if ast.IsOptionalChain(node) {
		return ast.IsOutermostOptionalChain(node)
	}

	// The assertion covers a whole chain: `foo?.bar!`, and with a paren to skip, `(foo?.bar)!`.
	nonNull := node.AsNonNullExpression()
	if nonNull == nil {
		return false
	}
	// SkipParentheses dereferences its argument, so the nil check has to come before the call
	// rather than after it.
	//
	// This check is defensive rather than reachable, and the distinction is worth stating because it
	// is invisible otherwise: reordering it after the call leaves the package's crash guard green,
	// since a parsed NonNullExpression always carries an expression and no shape in that guard
	// produces one without. So the guard proves those shapes are safe; it does not prove this line
	// is load-bearing. It is kept for a synthesized or reparsed node, where that invariant is the
	// parser's promise rather than ours.
	if nonNull.Expression == nil {
		return false
	}
	inner := ast.SkipParentheses(nonNull.Expression)
	return inner != nil && ast.IsOptionalChain(inner)
}
