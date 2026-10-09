package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnneededTernaryOptions configures the rule.
type NoUnneededTernaryOptions struct {
	// DefaultAssignment allows `foo ? foo : bar`, the pre-`||` way to spell a default.
	//
	// Defaults to TRUE upstream, so the second half of this rule is off unless somebody asks for
	// it. A pointer because of that default: the config layer hands a rule configured as a bare
	// severity a nil, and a plain bool field behind a nil decodes to false, which would silently
	// turn the second judgment ON for everyone.
	DefaultAssignment *bool `json:"defaultAssignment"`
}

// noUnneededTernaryDefaultAssignment is upstream's default for the only option.
const noUnneededTernaryDefaultAssignment = true

// resolve applies the default, leaving an absent key alone.
func (options NoUnneededTernaryOptions) resolve() bool {
	if options.DefaultAssignment == nil {
		return noUnneededTernaryDefaultAssignment
	}
	return *options.DefaultAssignment
}

// DecodeNoUnneededTernaryOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto`, because that helper errors on empty input and
// the config layer turns the error into nil, so a rule reached through it cannot tell "no options
// were written" from "the option was written as false". This option defaults to TRUE, so that
// distinction turns the second judgment on rather than merely narrowing it.
func DecodeNoUnneededTernaryOptions(raw []byte) (any, error) {
	var options NoUnneededTernaryOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

var messageNoUnneededTernaryConditionalExpression = rule.Message{
	Id: "unnecessaryConditionalExpression",
	Description: "This conditional returns a boolean literal on both branches, so the whole " +
		"ternary is a long way of writing the condition itself. The branches carry no " +
		"information the test does not already have, and a reader has to check both to learn " +
		"that. Write the condition, negated if the branches are the other way round.",
}

var messageNoUnneededTernaryConditionalAssignment = rule.Message{
	Id: "unnecessaryConditionalAssignment",
	Description: "This conditional tests a value and then returns that same value, which is the " +
		"pre-`||` way of spelling a default. `a ? a : b` evaluates `a` twice and reads as though " +
		"the two branches might differ. Write `a || b`, which says the same thing once.",
}

// NoUnneededTernary flags a conditional expression a simpler expression already says.
//
//	valid:   var a = x === 2 ? 'Yes' : 'No';
//	valid:   var a = foo ? foo : bar;                 // unless defaultAssignment is off
//	invalid: var a = x === 2 ? true : false;
//	invalid: var a = x === 2 ? false : true;
//	invalid: var a = foo ? foo : bar;                 // with defaultAssignment off
//
// # Two judgments, and only the first is on by default
//
// Boolean literals on BOTH branches means the ternary is a restatement of its own test, and that is
// always reported. The default-assignment shape is reported only when the option asks, because
// `foo ? foo : bar` predates `||` and some codebases still spell it that way deliberately.
//
// # The repair rebuilds text, which is what makes it dangerous, and why it is still safe here
//
// Every arm of this fixer REPLACES THE WHOLE TERNARY with reconstructed text rather than deleting a
// span, which is the shape that destroyed type information in two other rules tonight. It is safe
// here for a specific reason worth stating rather than assuming: every piece of the replacement is
// SLICED FROM THE SOURCE by range, never rebuilt from node fields, so an `as` expression, a
// `satisfies`, a non-null `!`, explicit type arguments and a `<T>` cast all survive as written.
// Measured against the installed rule on eight TypeScript shapes upstream's corpus does not contain.
//
// The one thing reconstructed rather than copied is the OPERATOR of an inverted comparison, and that
// is four characters chosen from a fixed table.
//
// # Where the parentheses come from
//
// Two places, both precedence questions, and both measured rather than reasoned:
//
//	inverting a test          `!x` binds tighter than most things, so anything below unary
//	                         precedence is wrapped: `!(a + b)`, but `!f()` and `!a`
//	the default's alternate   the replacement is `test || alternate`, so an alternate binding
//	                         looser than `||` is wrapped, and `??` is wrapped as well because
//	                         mixing it with `||` unparenthesized is a syntax error
//
// `as` and `satisfies` bind LOOSER than every binary operator, which upstream's table has no row
// for. `a ? a : b as any` becomes `a || (b as any)`, measured; a port inheriting upstream's table
// unchanged would give them the tightest precedence and emit `a || b as any`, which parses as
// `(a || b) as any` and asserts the type of the wrong thing.
var NoUnneededTernary = rule.Rule{
	Name: "no-unneeded-ternary",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		allowDefaultAssignment := noUnneededTernaryDefaultAssignment
		if decoded, configured := rule.OptionsAs[NoUnneededTernaryOptions](options); configured {
			allowDefaultAssignment = decoded.resolve()
		}
		return rule.Listeners{
			ast.KindConditionalExpression: func(node *ast.Node) {
				checkNoUnneededTernary(ctx, node, allowDefaultAssignment)
			},
		}
	},
}

// checkNoUnneededTernary judges one conditional expression.
func checkNoUnneededTernary(ctx rule.Context, node *ast.Node, allowDefaultAssignment bool) {
	conditional := node.AsConditionalExpression()
	if conditional.Condition == nil || conditional.WhenTrue == nil || conditional.WhenFalse == nil {
		return
	}

	consequent, consequentIsBoolean := noUnneededTernaryBooleanLiteralValue(conditional.WhenTrue)
	alternate, alternateIsBoolean := noUnneededTernaryBooleanLiteralValue(conditional.WhenFalse)

	if consequentIsBoolean && alternateIsBoolean {
		reportNoUnneededTernaryBooleanBranches(ctx, node, conditional, consequent, alternate)
		return
	}

	if !allowDefaultAssignment && noUnneededTernaryMatchesDefaultAssignment(conditional) {
		reportNoUnneededTernaryDefaultAssignment(ctx, node, conditional)
	}
}

// reportNoUnneededTernaryBooleanBranches reports a ternary whose branches are both boolean literals.
//
// Three repairs, chosen by what the branches say:
//
//	same value        the ternary answers that value regardless, so it collapses to the literal --
//	                  but ONLY when the test is a bare identifier, because evaluating `f()` may
//	                  have been the point. Upstream returns no fix otherwise and so does this.
//	false then true   the ternary is the negation of its test, so it becomes the inverted test
//	true then false   the ternary is the test coerced to boolean, so it becomes the test itself
//	                  when the test already produces a boolean, and `!!test` when it does not
func reportNoUnneededTernaryBooleanBranches(
	ctx rule.Context,
	node *ast.Node,
	conditional *ast.ConditionalExpression,
	consequent bool,
	alternate bool,
) {
	nodeRange := rule.TokenRange(ctx.SourceFile, node)

	if consequent == alternate {
		// `f() ? true : true` is left unrepaired: collapsing it would drop the call, and whether
		// that call mattered is not something the rule can know. Upstream returns null here, which
		// its corpus records as the single `output: null` case.
		if ast.SkipParentheses(conditional.Condition).Kind != ast.KindIdentifier {
			ctx.ReportRange(nodeRange, messageNoUnneededTernaryConditionalExpression)
			return
		}
		literal := "false"
		if consequent {
			literal = "true"
		}
		ctx.ReportRangeWithFixes(nodeRange, messageNoUnneededTernaryConditionalExpression,
			rule.ReplaceRange(nodeRange, literal))
		return
	}

	if alternate {
		// `test ? false : true` is the negation of the test.
		ctx.ReportRangeWithFixes(nodeRange, messageNoUnneededTernaryConditionalExpression,
			rule.ReplaceRange(nodeRange, noUnneededTernaryInvert(ctx, conditional.Condition)))
		return
	}

	// `test ? true : false` is the test, coerced when it is not already a boolean. The double
	// negation is written as a negation OF THE INVERSION rather than as `!!` plus the test, which
	// is upstream's `!${invertExpression(node.test)}` and is what makes `!(a + b)` come out as
	// `!!(a + b)` with the parentheses in the right place.
	replacement := "!" + noUnneededTernaryInvert(ctx, conditional.Condition)
	if noUnneededTernaryAlwaysBoolean(conditional.Condition) {
		replacement = noUnneededTernaryParenthesizedText(ctx, conditional.Condition)
	}
	ctx.ReportRangeWithFixes(nodeRange, messageNoUnneededTernaryConditionalExpression,
		rule.ReplaceRange(nodeRange, replacement))
}

// reportNoUnneededTernaryDefaultAssignment reports `a ? a : b` and repairs it to `a || b`.
func reportNoUnneededTernaryDefaultAssignment(
	ctx rule.Context,
	node *ast.Node,
	conditional *ast.ConditionalExpression,
) {
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	alternateText := noUnneededTernaryParenthesizedText(ctx, conditional.WhenFalse)

	// The alternate is wrapped when it binds looser than `||`, and separately when it is a `??`,
	// which cannot sit beside `||` without parentheses at all. Already-parenthesized alternates are
	// left alone, since `getParenthesisedText` has kept their parentheses.
	if conditional.WhenFalse.Kind != ast.KindParenthesizedExpression &&
		(noUnneededTernaryPrecedence(conditional.WhenFalse) < noUnneededTernaryOrPrecedence ||
			noUnneededTernaryIsCoalesce(conditional.WhenFalse)) {
		alternateText = "(" + alternateText + ")"
	}

	replacement := noUnneededTernaryParenthesizedText(ctx, conditional.Condition) + " || " +
		alternateText
	ctx.ReportRangeWithFixes(nodeRange, messageNoUnneededTernaryConditionalAssignment,
		rule.ReplaceRange(nodeRange, replacement))
}

// noUnneededTernaryOrPrecedence is the precedence of `||`, which the alternate is compared against.
//
// Named rather than written as 4 at the call site, because the number only means anything relative
// to `operatorAssignmentPrecedence`'s table and a bare literal there reads as a magic constant.
const noUnneededTernaryOrPrecedence = 4

// noUnneededTernaryBooleanLiteralValue reads a `true` or `false` literal branch.
//
// The second return separates "not a boolean literal" from the literal `false`, which a caller
// comparing the two branches has to be able to tell apart.
//
// Parentheses are NOT skipped, and that is upstream's behaviour rather than an oversight: espree
// gives them no node, so `foo ? (true) : false` has a parenthesized consequent here and a bare
// literal there. Measured -- upstream reports it, so the skip is required for fidelity, the same
// correction `no-eq-null` documents for its own operands.
func noUnneededTernaryBooleanLiteralValue(branch *ast.Node) (bool, bool) {
	if branch == nil {
		return false, false
	}
	switch ast.SkipParentheses(branch).Kind {
	case ast.KindTrueKeyword:
		return true, true
	case ast.KindFalseKeyword:
		return false, true
	}
	return false, false
}

// noUnneededTernaryMatchesDefaultAssignment answers whether a ternary is `a ? a : b`.
//
// Upstream compares the two identifiers by NAME, not by binding, so a shadowed `a` still matches.
// That is the right call for a rule about spelling: the repair `a || b` evaluates the same `a` the
// ternary's test evaluated, whatever it resolves to.
func noUnneededTernaryMatchesDefaultAssignment(conditional *ast.ConditionalExpression) bool {
	test := ast.SkipParentheses(conditional.Condition)
	consequent := ast.SkipParentheses(conditional.WhenTrue)
	if test.Kind != ast.KindIdentifier || consequent.Kind != ast.KindIdentifier {
		return false
	}
	return test.Text() == consequent.Text()
}

// noUnneededTernaryAlwaysBoolean answers whether an expression already produces a boolean.
//
// Upstream's `isBooleanExpression`: a comparison or relational operator, or a logical negation. It
// is deliberately syntactic rather than type-driven, so `Boolean(x)` and a `boolean`-typed variable
// both answer false and get the `!!` coercion they do not need. Reproduced rather than improved on:
// tightening it would change the repair text on inputs upstream has an answer for.
func noUnneededTernaryAlwaysBoolean(expression *ast.Node) bool {
	unwrapped := ast.SkipParentheses(expression)
	if unwrapped.Kind == ast.KindPrefixUnaryExpression {
		return unwrapped.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	}
	if unwrapped.Kind != ast.KindBinaryExpression {
		return false
	}
	operator := unwrapped.AsBinaryExpression().OperatorToken
	if operator == nil {
		return false
	}
	switch operator.Kind {
	case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken,
		ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken,
		ast.KindLessThanToken, ast.KindLessThanEqualsToken,
		ast.KindInKeyword, ast.KindInstanceOfKeyword:
		return true
	}
	return false
}

// noUnneededTernaryIsCoalesce answers whether an expression is a `??`.
//
// Its own test rather than a precedence comparison, because `??` and `||` share a precedence and
// still cannot be mixed without parentheses -- that is a grammar rule rather than a binding one, so
// no precedence table can express it. Upstream has the same separate check.
func noUnneededTernaryIsCoalesce(expression *ast.Node) bool {
	if expression.Kind != ast.KindBinaryExpression {
		return false
	}
	operator := expression.AsBinaryExpression().OperatorToken
	return operator != nil && operator.Kind == ast.KindQuestionQuestionToken
}

// noUnneededTernaryInvert renders the boolean inverse of an expression, as source text.
//
// Three shapes, in upstream's order. A comparison with a true inverse has its OPERATOR swapped in
// place, which is why `a === b ? false : true` becomes `a !== b` rather than `!(a === b)`. Note the
// relational operators are deliberately absent from that table: `<` and `>=` are not inverses,
// because both answer false for a NaN operand.
//
// Everything else is negated, with parentheses when it binds looser than a unary operator.
//
// # Every piece of this is sliced from the source
//
// The two halves either side of a swapped operator, and the whole expression in the negation arms,
// are taken by RANGE out of the file rather than rebuilt from node fields. That is what makes the
// repair safe on TypeScript: `foo as any`, `foo!`, `f<number>()` and `<boolean>a` all survive
// verbatim because nothing here ever looks at what they are made of.
func noUnneededTernaryInvert(ctx rule.Context, expression *ast.Node) string {
	text := ctx.SourceFile.Text()

	if expression.Kind == ast.KindBinaryExpression {
		if operator := expression.AsBinaryExpression().OperatorToken; operator != nil {
			if inverse := noUnneededTernaryInverseOperator(operator.Kind); inverse != "" {
				expressionRange := rule.TokenRange(ctx.SourceFile, expression)
				operatorRange := rule.TokenRange(ctx.SourceFile, operator)
				return text[expressionRange.Pos():operatorRange.Pos()] + inverse +
					text[operatorRange.End():expressionRange.End()]
			}
		}
	}

	inner := noUnneededTernaryParenthesizedText(ctx, expression)
	if noUnneededTernaryPrecedence(expression) < noUnneededTernaryUnaryPrecedence {
		return "!(" + inner + ")"
	}
	return "!" + inner
}

// noUnneededTernaryPrecedence is upstream's `getPrecedence` as THIS rule needs to ask it.
//
// It defers to the shared table in `operator_assignment.go` for everything both rules agree on, and
// corrects it in two places where that table answers a different question than upstream's. Both
// corrections were measured against the installed rule driven through the TypeScript parser, and
// both were found by a fixture failing rather than by reading the table.
//
// # A parenthesized expression reports what is INSIDE it
//
// Espree produces no node for parentheses, so upstream's `getPrecedence(node.test)` on `(foo + 1)`
// is asking about the addition and answers 12. The shared table answers 20, because our parser gives
// the parentheses a node and a parenthesized expression genuinely binds tightest.
//
// Both answers are right about their own question, and only one produces upstream's text: measured,
// `(foo + 1) ? true : false` becomes `!!((foo + 1))` with BOTH pairs of parentheses. That reads as
// redundant and it is exactly what upstream writes, because `getParenthesisedText` has already kept
// the source pair and the precedence test then adds its own. Reproduced rather than tidied: emitting
// one pair would be a nicer output that no upstream fixture asserts, and the corpus is the
// specification.
//
// # An unknown node type is the LOWEST precedence, not the highest
//
// Upstream's default arm returns 20 for a node in `eslintVisitorKeys` and -1 for anything else, with
// a comment saying the -1 is deliberate: an unrecognised node is assumed to bind loosely so a fixer
// wraps it rather than risking a meaning change. `TSNonNullExpression` is not in those keys, so
// `foo! ? true : false` becomes `!!(foo!)` upstream while the shared table's 20 produces `!!foo!`.
//
// `!!foo!` happens to parse, which is what makes this the dangerous direction: nothing downstream
// would have complained.
func noUnneededTernaryPrecedence(expression *ast.Node) int {
	switch expression.Kind {
	case ast.KindParenthesizedExpression:
		// Ask about what the parentheses hold, which is the question espree makes upstream ask.
		if inner := expression.AsParenthesizedExpression().Expression; inner != nil {
			return noUnneededTernaryPrecedence(inner)
		}
		return -1
	case ast.KindNonNullExpression:
		// TypeScript only, so upstream's table has no row and its default arm answers -1.
		return -1
	}
	return operatorAssignmentPrecedence(expression)
}

// noUnneededTernaryUnaryPrecedence is the precedence of a unary operator, which decides whether a
// negated expression needs wrapping.
const noUnneededTernaryUnaryPrecedence = 16

// noUnneededTernaryInverseOperator returns the true inverse of a comparison, or "" when there is
// none.
//
// Only the four equality operators have one. `<` against `>=` looks like a pair and is not: both
// answer false when either operand is NaN, so swapping them changes the result. Upstream says so in
// a comment on the table and this reproduces the omission rather than helpfully completing it.
func noUnneededTernaryInverseOperator(kind ast.Kind) string {
	switch kind {
	case ast.KindEqualsEqualsToken:
		return "!="
	case ast.KindExclamationEqualsToken:
		return "=="
	case ast.KindEqualsEqualsEqualsToken:
		return "!=="
	case ast.KindExclamationEqualsEqualsToken:
		return "==="
	}
	return ""
}

// noUnneededTernaryParenthesizedText returns an expression's source text, parentheses included.
//
// Upstream's `getParenthesisedText`, which walks outward through enclosing parentheses so that
// `(a + b)` renders with its own parentheses rather than as `a + b`. Our parser gives those
// parentheses a node, so the walk is downward-free: the node's own range already covers them.
//
// Taken by range rather than reconstructed, which is the whole reason this fixer does not lose type
// syntax.
func noUnneededTernaryParenthesizedText(ctx rule.Context, expression *ast.Node) string {
	expressionRange := rule.TokenRange(ctx.SourceFile, expression)
	return ctx.SourceFile.Text()[expressionRange.Pos():expressionRange.End()]
}
