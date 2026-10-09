package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// noNonNullText and suggestOptionalChainText are the rule's messages, whose wording lives in
// `policy/messages/no-non-null-assertion.json`. The ids are upstream's. The wording is not: it gives
// the repairs in the order the house's three migration waves settled on (guard, fix the type,
// restructure, then `assert` or `required`), and upstream's recommends `?.`, which those waves found
// hides a missing value as often as it handles one.
var (
	noNonNullText            = policy.MessageOf("@typescript-eslint/no-non-null-assertion", "noNonNull")
	suggestOptionalChainText = policy.MessageOf("@typescript-eslint/no-non-null-assertion", "suggestOptionalChain")
)

func messageNoNonNull() rule.Message {
	return rule.Message{Id: noNonNullText.Id, Description: noNonNullText.Render(nil)}
}

func messageSuggestOptionalChain() rule.Message {
	return rule.Message{Id: suggestOptionalChainText.Id, Description: suggestOptionalChainText.Render(nil)}
}

// NoNonNullAssertion flags every non-null assertion.
//
//	valid:   foo?.bar
//	valid:   let value!: number
//	valid:   const other: NonNullable<Thing> = thing
//	invalid: foo!
//	invalid: foo!.bar
//	invalid: foo!()
//
// Ported from `@typescript-eslint/no-non-null-assertion`.
//
// # It reports unconditionally, and only the repair is conditional
//
// Every assertion is a finding. The parent shape decides only whether a suggestion can be attached,
// so a rule that looked at the parent to decide whether to report would be a different, quieter
// rule. The bare `foo!` gets a finding with no suggestion, which is correct: there is no `?.` that
// means the same thing when nothing follows the assertion.
//
// # A suggestion rather than a fix, and why the split is real here
//
// `foo!.bar` and `foo?.bar` do not mean the same thing. The first evaluates to `bar` or throws; the
// second evaluates to `bar` or `undefined`, which changes the expression's type and pushes a new
// `undefined` into everything downstream. That is a meaning change, so a human chooses it. The
// original marks the rule `hasSuggestions` and not `fixable` for the same reason.
//
// # Definite assignment is a different syntax and cannot reach this rule
//
// `let value!: number` also spells an exclamation mark, and it means the opposite thing: a promise
// to the compiler that the variable is assigned before use, checked rather than erased. In
// typescript-go it parses as an ExclamationToken on the declaration, never a NonNullExpression, so
// a listener keyed to this node kind cannot see it. Verified against the parser rather than assumed,
// and pinned by a clean fixture, because the two constructs are one keystroke apart in source.
var NoNonNullAssertion = rule.Rule{
	Name: "@typescript-eslint/no-non-null-assertion",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				suggestions := optionalChainSuggestions(ctx, node)
				if len(suggestions) == 0 {
					ctx.ReportNode(node, messageNoNonNull())
					return
				}
				ctx.ReportNodeWithSuggestions(node, messageNoNonNull(), suggestions...)
			},
		}
	},
}

// optionalChainSuggestions builds the `?.` repair when one exists for this assertion's position.
//
// Three positions can be repaired, and each rewrites differently:
//
//	foo!.bar    the `!` goes and the `.` becomes `?.`
//	foo![key]   the `!` becomes `?.`, since `?.[` is how a computed access opts in
//	foo!()      the `!` becomes `?.`, same reason
//
// A fourth position, where the following link already carries `?.`, needs only the `!` removed:
// `foo!?.bar` is already guarded and the assertion is the redundant half.
func optionalChainSuggestions(ctx rule.Context, node *ast.Node) []rule.Suggestion {
	parent := node.Parent
	if parent == nil {
		return nil
	}

	suggest := func(fixes ...rule.Fix) []rule.Suggestion {
		return []rule.Suggestion{{Message: messageSuggestOptionalChain(), Fixes: fixes}}
	}
	operator := nonNullAssertionOperatorRange(node)

	switch parent.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// Only the object position. `foo?.[key!]` puts an assertion in the index rather than on the
		// thing being indexed, and rewriting that one to `?.` would be nonsense.
		if parent.Expression() != node {
			return nil
		}
		// An assertion on the left of an assignment cannot become an optional chain: `foo?.bar = 1`
		// is not valid JavaScript. The suggestion would not just change meaning, it would not parse.
		if isAssignmentTarget(parent) {
			return nil
		}
		if parent.QuestionDotToken() != nil {
			return suggest(rule.RemoveRange(operator))
		}
		if parent.Kind == ast.KindElementAccessExpression {
			return suggest(rule.ReplaceRange(operator, "?."))
		}
		// Dot access. The `.` is a separate token from the `!`, and the two can be separated by a
		// comment or a newline, so this replaces each in place rather than rewriting the span
		// between them. Rewriting the span would delete whatever sits in the middle.
		dot := dotTokenAfter(ctx, node, parent)
		if dot.Pos() == dot.End() {
			return nil
		}
		return suggest(rule.RemoveRange(operator), rule.ReplaceRange(dot, "?."))
	case ast.KindCallExpression:
		// The callee position only, so an assertion on an argument is left alone. The original has
		// no assignment guard on this branch, because a call is never an assignment target.
		if parent.Expression() != node {
			return nil
		}
		if parent.QuestionDotToken() != nil {
			return suggest(rule.RemoveRange(operator))
		}
		return suggest(rule.ReplaceRange(operator, "?."))
	}
	return nil
}

// dotTokenAfter finds the `.` between an assertion and the property it reaches for.
//
// It is located by scanning forward from the end of the assertion rather than by arithmetic,
// because the two tokens need not be adjacent: `foo!\n// why\n.bar` is one property access with a
// comment inside it, and the original preserves that comment by editing the two tokens separately.
func dotTokenAfter(ctx rule.Context, node *ast.Node, parent *ast.Node) core.TextRange {
	text := ctx.SourceFile.Text()
	for position := node.End(); position < parent.End() && position < len(text); position++ {
		if text[position] == '.' {
			return core.NewTextRange(position, position+1)
		}
	}
	return core.NewTextRange(0, 0)
}

// isAssignmentTarget reports whether this expression is being written to rather than read.
//
// Ported from typescript-eslint's `isAssignee`, and ported whole rather than by its common case.
// The reference implementation this project consults recognizes only the plain `x = 1` shape, and
// the shortfall is not cosmetic: for `delete foo!.bar` or `foo!.bar++` it offers a suggestion that
// rewrites to `delete foo?.bar` and `foo?.bar++`, neither of which parses. A suggestion that does
// not compile is worse than no suggestion, because a human reads the rule as having checked.
//
// # One deliberate difference from the original, in the other direction
//
// The recursion here walks up through parentheses and type assertions, and stops at a member access.
// The original stops there too, which means `foo!.bar.baz = 1` is *not* treated as an assignment:
// the assertion's parent is `foo!.bar`, whose own parent is the member access `foo!.bar.baz` rather
// than the assignment. So the original reports it with a suggestion, and so does this. That looks
// like a gap and is not one: rewriting `foo!.bar.baz = 1` to `foo?.bar.baz = 1` is a real error the
// compiler catches immediately, unlike the `delete` and `++` cases where the shape is closer to
// valid. Matching the original is the right call regardless, because divergence here would be
// invisible: both readings produce a suggestion nobody may ever apply.
func isAssignmentTarget(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		// `a = 1`, `a += 1`, and the compound forms.
		binary := parent.AsBinaryExpression()
		return ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.Left == node
	case ast.KindDeleteExpression:
		// `delete a.b`
		return parent.AsDeleteExpression().Expression == node
	case ast.KindPrefixUnaryExpression:
		// `--a.b`
		operator := parent.AsPrefixUnaryExpression()
		return reference.IsUpdateOperator(operator.Operator) && operator.Operand == node
	case ast.KindPostfixUnaryExpression:
		// `a.b++`
		operator := parent.AsPostfixUnaryExpression()
		return reference.IsUpdateOperator(operator.Operator) && operator.Operand == node
	case ast.KindArrayLiteralExpression, ast.KindSpreadElement:
		// `[a.b] = [0]` and `[...a.b] = [0]`. A destructuring assignment target parses as an array
		// literal here rather than as a pattern, so this reaches both the element and the spread.
		return true
	case ast.KindPropertyAssignment:
		// `({ foo: a.b } = other)`. The object literal is only a target when it is itself assigned,
		// so this recurses rather than accepting every property in every object.
		if parent.AsPropertyAssignment().Initializer != node {
			return false
		}
		object := parent.Parent
		return object != nil && object.Kind == ast.KindObjectLiteralExpression && isAssignmentTarget(object)
	case ast.KindParenthesizedExpression,
		ast.KindNonNullExpression,
		ast.KindAsExpression,
		ast.KindTypeAssertionExpression,
		ast.KindSatisfiesExpression:
		// `(a.b as number)++`, `[...a.b!] = [0]`. These wrap the target without being it, so the
		// question passes through them unchanged. Parentheses are on this list where the original
		// has no need for them: ESTree does not build a node for parens, and typescript-go does,
		// so `(a.b as number)++` reaches the increment through one here and through nothing there.
		return isAssignmentTarget(parent)
	}
	return false
}
