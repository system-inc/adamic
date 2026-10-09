package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoExtraNonNullAssertion = rule.Message{
	Id: "noExtraNonNullAssertion",
	Description: "This non-null assertion is redundant: the value it asserts on has already been " +
		"asserted non-null, or is about to be checked at runtime by the `?.` that follows it. " +
		"Either way the second `!` narrows nothing the first one did not, so it is noise that " +
		"reads as a claim. Remove it.",
}

// NoExtraNonNullAssertion flags a non-null assertion that repeats one already made.
//
//	valid:   foo!
//	valid:   foo!.bar
//	valid:   foo?.bar!.baz
//	invalid: foo!!
//	invalid: foo!!.bar
//	invalid: foo!?.bar
//	invalid: foo!?.()
//
// Ported from `@typescript-eslint/no-extra-non-null-assertion`.
//
// # The three shapes, and why the third is not the other two
//
// The original registers three selectors, and each names a different way one assertion can be
// redundant against another:
//
//	TSNonNullExpression > TSNonNullExpression            foo!!
//	MemberExpression[optional = true] > ...object        foo!?.bar
//	CallExpression[optional = true] > ...callee          foo!?.()
//
// The first is an assertion on an assertion. The second and third are an assertion immediately
// followed by an optional chain, which is redundant for a different reason: `?.` performs the
// runtime check the `!` claimed was unnecessary, so the two contradict each other and the `?.`
// wins.
//
// # `optional = true` is not `IsOptionalChain`
//
// The ESTree selector reads a per-node `optional` boolean, which is true only for the link that
// carries the `?.` itself. typescript-go instead propagates `NodeFlagsOptionalChain` down the whole
// chain, so `IsOptionalChain` is true for every link including ones with no `?.` of their own.
// Translating `[optional = true]` as `IsOptionalChain` therefore reports `foo?.bar!.baz`, which the
// original marks valid and which is the regression upstream issue #2166 was filed for.
//
// `IsOptionalChainRoot` is the faithful translation: it is the link that owns the `?.` token. The
// clean fixture for `foo?.bar!.baz` is what pins this, and it is the fixture most worth keeping.
var NoExtraNonNullAssertion = rule.Rule{
	Name: "@typescript-eslint/no-extra-non-null-assertion",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				// Parentheses are transparent in ESTree, so the original sees through them for
				// free: `(foo!)!` is a non-null of a non-null there. typescript-go materializes a
				// ParenthesizedExpression node, so the walk up has to skip it explicitly or the
				// parenthesized forms silently stop reporting.
				outermost := outermostParenthesizedExpression(node)
				if outermost == nil {
					return
				}
				parent := outermost.Parent
				if parent == nil {
					return
				}

				// foo!! — an assertion whose parent is another assertion.
				if parent.Kind == ast.KindNonNullExpression {
					reportRedundantAssertion(ctx, node)
					return
				}

				// foo!?.bar and foo!?.() — an assertion sitting directly under the link that owns
				// the `?.`, in the position that link reaches through.
				switch parent.Kind {
				case ast.KindPropertyAccessExpression,
					ast.KindElementAccessExpression,
					ast.KindCallExpression:
					if ast.IsOptionalChainRoot(parent) && parent.Expression() == outermost {
						reportRedundantAssertion(ctx, node)
					}
				}
			},
		}
	},
}

// reportRedundantAssertion reports the assertion and proposes deleting just its `!`.
//
// This is a fix rather than a suggestion, matching the original's `fixable: 'code'`, and the
// distinction is earned rather than inherited: removing a redundant assertion cannot change what
// the expression means, because the assertion it duplicates is still there. That is the definition
// the fix/suggestion split uses.
func reportRedundantAssertion(ctx rule.Context, node *ast.Node) {
	ctx.ReportNodeWithFixes(node, messageNoExtraNonNullAssertion,
		rule.RemoveRange(nonNullAssertionOperatorRange(node)))
}

// nonNullAssertionOperatorRange is the span of the `!` token alone.
//
// A parsed NonNullExpression ends immediately after its operator, so the last byte of the node is
// the `!` and no scan is needed to find it. The original computes it the same way
// (`[node.range[1] - 1, node.range[1]]`).
//
// The byte arithmetic is safe because `!` is ASCII: it occupies exactly one byte regardless of what
// multi-byte text surrounds it, so subtracting one from the end cannot land inside a character.
func nonNullAssertionOperatorRange(node *ast.Node) core.TextRange {
	return node.Loc.WithPos(node.End() - 1)
}

// outermostParenthesizedExpression walks up through any parentheses wrapping a node, returning the
// outermost one, or the node itself when it is not parenthesized.
//
// This is how the walk-up guards see what the original sees. ESTree has no node for parentheses, so
// `(foo!)!` presents there as a non-null directly inside a non-null; here it presents as a non-null
// inside a ParenthesizedExpression inside a non-null, and a guard reading `node.Parent` finds the
// paren and stops.
func outermostParenthesizedExpression(node *ast.Node) *ast.Node {
	for node != nil && node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}
