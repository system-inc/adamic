package core

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageRedundantBooleanCall = rule.Message{
	Id: "redundantBooleanCall",
	Description: "This wraps a value in `Boolean(...)` somewhere the value is already coerced to a " +
		"boolean: an `if` test, a `while` test, the operand of `!`, the test of a ternary, or the " +
		"argument of another `Boolean` call. The call runs, allocates nothing useful, and produces " +
		"exactly what the surrounding position would have produced without it. Pass the value itself.",
}

var messageRedundantDoubleNegation = rule.Message{
	Id: "redundantDoubleNegation",
	Description: "This applies `!!` somewhere the value is already coerced to a boolean, so the " +
		"second negation undoes the first and the pair leaves the condition exactly as it found " +
		"it. Two operators that cancel read as if they were doing something. Drop them.",
}

// NoExtraBooleanCastOptions configures whether inner expressions are checked.
type NoExtraBooleanCastOptions struct {
	// EnforceForInnerExpressions extends the rule through expressions whose *result* lands in a
	// boolean context rather than the cast itself: the operands of `&&` and `||`, the right side of
	// `??`, both branches of a ternary, and the last element of a sequence.
	//
	// Off by default in both upstreams, because the extension is not a pure superset in intent.
	// `foo || Boolean(bar)` coerces only when the whole expression is tested, and an author may be
	// building a value rather than a condition.
	EnforceForInnerExpressions bool `json:"enforceForInnerExpressions"`
}

// UnmarshalJSON accepts the deprecated `enforceForLogicalOperands` spelling as a synonym.
//
// This is not cosmetic and it is not quite parity either. ESLint declares the two names as
// *different options*: the older one recurses only through logical operators, while the newer one
// also recurses through `??`, ternary branches, and sequence tails. oxc collapses them with a serde
// alias, so under oxc the deprecated name silently buys the newer, wider behaviour.
//
// This port follows oxc, because oxc is the port target. The divergence is real and measurable:
// upstream's own corpus runs 104 cases under the old spelling and 118 under the new, and the 14
// extra are exactly the `??`, ternary-branch and sequence-tail shapes ESLint's legacy mode would
// decline. Under this rule all 118 report regardless of which name was written. Latent for us today
// because our config passes no options at all, and worth stating rather than discovering.
func (options *NoExtraBooleanCastOptions) UnmarshalJSON(raw []byte) error {
	var decoded struct {
		EnforceForInnerExpressions *bool `json:"enforceForInnerExpressions"`
		EnforceForLogicalOperands  *bool `json:"enforceForLogicalOperands"`
	}
	if err := rule.UnmarshalOptions(raw, &decoded); err != nil {
		return err
	}
	// The newer name wins when both are written, matching serde's alias resolution, where the
	// canonical field name takes the value and the alias only fills an absent one.
	switch {
	case decoded.EnforceForInnerExpressions != nil:
		options.EnforceForInnerExpressions = *decoded.EnforceForInnerExpressions
	case decoded.EnforceForLogicalOperands != nil:
		options.EnforceForInnerExpressions = *decoded.EnforceForLogicalOperands
	}
	return nil
}

// NoExtraBooleanCast flags a boolean cast in a position that already coerces to boolean.
//
//	valid:   var foo = !!bar;
//	valid:   var foo = Boolean(bar);
//	valid:   if (new Boolean(foo)) {}
//	valid:   if ((Boolean(1), 2)) {}
//	invalid: if (!!foo) {}
//	invalid: if (Boolean(foo)) {}
//	invalid: var foo = !!!bar;
//	invalid: var foo = Boolean(!!bar);
//
// # The two findings are not the same kind of finding
//
// Unwrapping `Boolean(x)` is a Fix and the engine applies it unattended, because in a position that
// already coerces, `Boolean(x)` and `x` are indistinguishable to everything downstream.
//
// Removing `!!` is a Suggestion instead. That looks inconsistent and it is deliberate: the two
// rewrites have different blast radii when the surrounding context turns out not to coerce after
// all. Upstream encodes the same split, and it is the only rule in oxc's 946 that declares
// `conditional_fix_or_conditional_suggestion`. Collapsing them into one kind is the mistake this
// comment exists to prevent, because a rule that auto-applies the `!!` removal is auto-applying a
// change of meaning wherever the context analysis is wrong.
//
// # Which positions coerce
//
// Directly: the test of `if`, `while`, `do...while`, `for`, and a ternary; the operand of `!`; and
// the first argument of `Boolean(...)` or `new Boolean(...)`. `new Boolean` counts as a *context*
// even though it is never itself reported, which is why `if (new Boolean(foo)) {}` is a valid case.
//
// Under the option, also transitively: both operands of `&&` and `||`, the right operand of `??`,
// both branches of a ternary, and the last element of a sequence, each only when the enclosing
// expression is itself in a coercing position. The recursion is what makes `if (!!foo || bar) {}`
// report while `var foo = bar || !!baz` stays silent.
//
// # Parentheses are part of the repair, not decoration
//
// The unwrapped expression lands where a call or a `!!` used to sit, and those bind tighter than
// most operators, so `!Boolean(a && b)` becomes `!(a && b)` rather than `!a && b`, which would mean
// something else entirely. The precedence comparison below is the whole of that decision.
var NoExtraBooleanCast = rule.Rule{
	Name: "no-extra-boolean-cast",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		enforceForInnerExpressions := false
		if parsed, ok := rule.OptionsAs[NoExtraBooleanCastOptions](options); ok {
			enforceForInnerExpressions = parsed.EnforceForInnerExpressions
		}

		reporter := booleanCastReporter{
			context:                    ctx,
			sourceText:                 ctx.SourceFile.Text(),
			enforceForInnerExpressions: enforceForInnerExpressions,
		}

		return rule.Listeners{
			ast.KindCallExpression:        reporter.reportBooleanCall,
			ast.KindPrefixUnaryExpression: reporter.reportDoubleNegation,
		}
	},
}

// booleanCastReporter carries the per-file state the two listeners share.
type booleanCastReporter struct {
	context                    rule.Context
	sourceText                 string
	enforceForInnerExpressions bool
}

// reportBooleanCall handles `Boolean(x)` sitting in a coercing position.
func (reporter booleanCastReporter) reportBooleanCall(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil || !isBooleanIdentifier(call.Expression) {
		return
	}
	if !reporter.isFlaggedContext(node) {
		return
	}

	fix, hasFix := reporter.booleanCallFix(node)
	if !hasFix {
		reporter.context.ReportNode(node, messageRedundantBooleanCall)
		return
	}
	reporter.context.ReportNodeWithFixes(node, messageRedundantBooleanCall, fix)
}

// reportDoubleNegation handles the outer `!` of a `!!` pair in a coercing position.
//
// The listener fires on the *inner* negation and walks up, matching upstream, so the node reported
// is the outer `!` and its whole `!!x` span. Listening on the outer one instead would report the
// same finding but would have to look down through parentheses to recognise the pair.
func (reporter booleanCastReporter) reportDoubleNegation(node *ast.Node) {
	unary := node.AsPrefixUnaryExpression()
	if unary == nil || unary.Operator != ast.KindExclamationToken {
		return
	}

	parent := reporter.realParent(node)
	if parent == nil || parent.Kind != ast.KindPrefixUnaryExpression {
		return
	}
	outer := parent.AsPrefixUnaryExpression()
	if outer == nil || outer.Operator != ast.KindExclamationToken {
		return
	}
	if !reporter.isFlaggedContext(parent) {
		return
	}

	argument := ast.SkipParentheses(unary.Operand)
	replacement := reporter.replacementText(argument, parent, reporter.tokenText(argument))
	reporter.context.ReportNodeWithSuggestions(parent, messageRedundantDoubleNegation,
		rule.Suggestion{
			Message: messageRedundantDoubleNegation,
			Fixes:   []rule.Fix{reporter.context.ReplaceNode(parent, replacement)},
		})
}

// booleanCallFix builds the repair for a `Boolean(...)` call, or declines to.
//
// Three shapes, and the argument-less one is the reason this returns a flag rather than a string.
// `Boolean()` is `false`, so in a coercing position it collapses to the literal, and under a `!` the
// whole `!Boolean()` collapses to `true` instead. More than one argument, or a spread, is declined
// outright: `Boolean(...foo)` cannot be reduced to any single expression without knowing what `foo`
// holds, so upstream emits the finding with no repair and so does this.
func (reporter booleanCastReporter) booleanCallFix(node *ast.Node) (rule.Fix, bool) {
	arguments := argumentNodes(node)

	if len(arguments) == 0 {
		if parent := reporter.realParent(node); parent != nil && isLogicalNot(parent) {
			replacement := reporter.padForTokenBoundary(parent, "true")
			return reporter.context.ReplaceNode(parent, replacement), true
		}
		return reporter.context.ReplaceNode(node, "false"), true
	}

	if len(arguments) != 1 || arguments[0].Kind == ast.KindSpreadElement {
		return rule.Fix{}, false
	}

	argument := ast.SkipParentheses(removeDoubleNegation(arguments[0]))
	replacement := reporter.replacementText(argument, node, reporter.tokenText(argument))
	return reporter.context.ReplaceNode(node, replacement), true
}

// isFlaggedContext answers whether a cast at this node sits somewhere that coerces.
//
// Directly, or transitively through the inner-expression forms when the option is on. The recursion
// asks the same question of the parent, so `if (!!a || b) {}` reaches the `if` test through the
// `||` while `var x = !!a || b` runs out of coercing parents and stays silent.
func (reporter booleanCastReporter) isFlaggedContext(node *ast.Node) bool {
	parent := reporter.realParent(node)
	if reporter.isBooleanContext(node, parent) {
		return true
	}
	if parent == nil {
		return false
	}
	return reporter.isInnerBooleanContext(node, parent) && reporter.isFlaggedContext(parent)
}

// isBooleanContext answers whether the parent coerces this node directly.
func (reporter booleanCastReporter) isBooleanContext(node *ast.Node, parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	if isBooleanCallOrConstruction(parent) && reporter.isFirstArgument(node, parent) {
		return true
	}
	if reporter.isTestPosition(node, parent) {
		return true
	}
	return isLogicalNot(parent)
}

// isInnerBooleanContext answers whether the parent *passes through* coercion to this node.
//
// Gated on the option, and each arm is the position whose value becomes the enclosing expression's
// value. `&&` and `||` pass through from either side. `??` passes through only on the right, because
// the left is tested for nullishness rather than truthiness. A ternary passes through its branches
// but not its test, which is already a direct boolean context above. A sequence passes through only
// its last element, which is what makes `if ((Boolean(1), 2)) {}` a valid case.
func (reporter booleanCastReporter) isInnerBooleanContext(node *ast.Node, parent *ast.Node) bool {
	if !reporter.enforceForInnerExpressions {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			return true
		case ast.KindQuestionQuestionToken:
			return sameExpression(binary.Right, node)
		case ast.KindCommaToken:
			// A sequence in this tree is a left-nested chain of comma operators, so the "last
			// element" is the right side of the outermost comma and nothing else.
			return sameExpression(binary.Right, node)
		}
		return false

	case ast.KindConditionalExpression:
		conditional := parent.AsConditionalExpression()
		if conditional == nil {
			return false
		}
		return sameExpression(conditional.WhenTrue, node) || sameExpression(conditional.WhenFalse, node)
	}
	return false
}

// isTestPosition answers whether the node is the tested expression of a statement or ternary.
func (reporter booleanCastReporter) isTestPosition(node *ast.Node, parent *ast.Node) bool {
	var test *ast.Node
	switch parent.Kind {
	case ast.KindIfStatement:
		test = parent.AsIfStatement().Expression
	case ast.KindWhileStatement:
		test = parent.AsWhileStatement().Expression
	case ast.KindDoStatement:
		test = parent.AsDoStatement().Expression
	case ast.KindForStatement:
		test = parent.AsForStatement().Condition
	case ast.KindConditionalExpression:
		test = parent.AsConditionalExpression().Condition
	default:
		return false
	}
	return sameExpression(test, node)
}

// isFirstArgument answers whether the node is the first argument of a call or construction.
func (reporter booleanCastReporter) isFirstArgument(node *ast.Node, parent *ast.Node) bool {
	arguments := argumentNodes(callOrNew(parent))
	if len(arguments) == 0 {
		return false
	}
	return sameExpression(arguments[0], node)
}

// realParent walks up past parentheses, which upstream also skips.
//
// Parentheses are a node in this tree exactly as they are in oxc's, and every question this rule
// asks is about the expression structure rather than the written grouping: `!(Boolean())` has to see
// the `!` as the parent of the call, or its repair produces `!false` instead of `true`.
func (reporter booleanCastReporter) realParent(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent
}

// replacementText renders the unwrapped expression, parenthesised when precedence demands it.
func (reporter booleanCastReporter) replacementText(replacement *ast.Node, target *ast.Node, expressionText string) string {
	result := expressionText
	if reporter.needsParentheses(replacement, target) {
		trimmed := text.TrimWhitespace(expressionText)
		if !(strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")")) {
			result = "(" + expressionText + ")"
		}
	}
	return reporter.padForTokenBoundary(target, result)
}

// needsParentheses answers whether dropping the cast would re-associate the expression.
//
// The comparison is against the *parent of the thing being replaced*, because that is what the
// unwrapped expression will now bind to. Already-parenthesised targets need nothing: the source's
// own parentheses survive the replacement and do the job.
func (reporter booleanCastReporter) needsParentheses(replacement *ast.Node, target *ast.Node) bool {
	if target.Parent != nil && target.Parent.Kind == ast.KindParenthesizedExpression {
		return false
	}
	parent := reporter.realParent(target)
	if parent == nil {
		return false
	}
	replacementPrecedence := ast.GetExpressionPrecedence(replacement)

	switch parent.Kind {
	case ast.KindPrefixUnaryExpression:
		return isLogicalNot(parent) && replacementPrecedence < ast.OperatorPrecedenceUnary

	case ast.KindCallExpression, ast.KindNewExpression:
		// Only a comma expression, because an argument list is itself comma-separated and an
		// unparenthesised sequence would read as two arguments.
		return reporter.isFirstArgument(target, parent) &&
			replacementPrecedence == ast.OperatorPrecedenceComma

	case ast.KindConditionalExpression:
		conditional := parent.AsConditionalExpression()
		if conditional == nil {
			return false
		}
		if sameExpression(conditional.Condition, target) {
			return replacementPrecedence <= ast.OperatorPrecedenceConditional
		}
		if sameExpression(conditional.WhenTrue, target) || sameExpression(conditional.WhenFalse, target) {
			return replacementPrecedence == ast.OperatorPrecedenceComma
		}
		return false

	case ast.KindBinaryExpression:
		return reporter.binaryNeedsParentheses(replacement, target, parent)
	}
	return false
}

// binaryNeedsParentheses is the `&&`, `||`, `??` arm of needsParentheses.
//
// Split out because it is the only arm that has to reason about the operator token rather than the
// node kind, and about a mixing rule that is a syntax error rather than a precedence question.
func (reporter booleanCastReporter) binaryNeedsParentheses(replacement *ast.Node, target *ast.Node, parent *ast.Node) bool {
	binary := parent.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return false
	}
	// Only the logical operators. Upstream reaches this arm solely through `LogicalExpression`,
	// which in this tree is folded into `BinaryExpression` alongside arithmetic, comparison, comma
	// and assignment, so the narrowing that oxc gets from its node kind has to be written out.
	operator := binary.OperatorToken.Kind
	if operator != ast.KindAmpersandAmpersandToken && operator != ast.KindBarBarToken &&
		operator != ast.KindQuestionQuestionToken {
		return false
	}

	replacementPrecedence := ast.GetExpressionPrecedence(replacement)
	operatorPrecedence := ast.GetBinaryOperatorPrecedence(operator)

	if replacementPrecedence < operatorPrecedence {
		return true
	}
	// Equal precedence is fine on the left and needs parentheses on the right, because these
	// operators are left-associative: `a || (b || c)` and `a || b || c` group differently.
	if replacementPrecedence == operatorPrecedence && sameExpression(binary.Right, target) {
		return true
	}
	// `??` cannot be mixed with `&&` or `||` without parentheses in either direction. That is a
	// grammar restriction rather than a precedence one, so it is checked by operator identity and
	// not by comparing numbers, which in this tree cannot even express it: `??` and `||` share the
	// precedence value 5, so the comparisons above see them as equal and let both through.
	if isNullishCoalescing(replacement) {
		return true
	}
	if operator == ast.KindQuestionQuestionToken && isLogicalAndOr(replacement) {
		return true
	}
	return false
}

// padForTokenBoundary inserts a space when the replacement would weld onto a neighbouring token.
//
// `void!Boolean()` becomes `void true` and not `voidtrue`, and `yield!!a ? b : c` becomes
// `yield a ? b : c` and not `yielda ? b : c`. The cast being removed was itself the separator, so
// deleting it is what creates the collision.
func (reporter booleanCastReporter) padForTokenBoundary(target *ast.Node, replacement string) string {
	if replacement == "" {
		return replacement
	}
	span := rule.TokenRange(reporter.context.SourceFile, target)
	start, end := span.Pos(), span.End()

	// Decoded as runes rather than indexed as bytes. Upstream indexes bytes and casts, which reads
	// a continuation byte as a character whenever the neighbour is non-ASCII, and every identifier
	// character above U+007F is multi-byte in UTF-8. Nothing in the imported corpus is non-ASCII so
	// no fixture here can see the difference, which is exactly why it is worth writing correctly:
	// the tree has identifiers with accented letters in it and the byte version would answer them
	// wrong in whichever direction the continuation byte happened to fall.
	previous, previousSize := utf8.DecodeLastRuneInString(reporter.sourceText[:start])
	if previousSize > 0 && isIdentifierPart(previous) {
		if first, _ := utf8.DecodeRuneInString(replacement); isIdentifierPart(first) {
			replacement = " " + replacement
		}
	}
	next, nextSize := utf8.DecodeRuneInString(reporter.sourceText[end:])
	if nextSize > 0 && isIdentifierStart(next) &&
		!isASCIISpace(replacement[len(replacement)-1]) {
		replacement += " "
	}
	return replacement
}

// tokenText returns a node's own source text, without the trivia in front of it.
func (reporter booleanCastReporter) tokenText(node *ast.Node) string {
	span := rule.TokenRange(reporter.context.SourceFile, node)
	return reporter.sourceText[span.Pos():span.End()]
}

// removeDoubleNegation strips one `!!` pair from an expression, if it wears one.
//
// This is what turns `Boolean(!!foo)` into `Boolean(foo)` in a single report rather than needing the
// nested finding to fire first and the file to be linted twice.
func removeDoubleNegation(node *ast.Node) *ast.Node {
	inner := withoutNegation(node)
	if inner == nil {
		return node
	}
	innermost := withoutNegation(inner)
	if innermost == nil {
		return node
	}
	return innermost
}

// withoutNegation returns the operand of a `!`, or nil when the node is not one.
func withoutNegation(node *ast.Node) *ast.Node {
	skipped := ast.SkipParentheses(node)
	if skipped == nil || skipped.Kind != ast.KindPrefixUnaryExpression {
		return nil
	}
	unary := skipped.AsPrefixUnaryExpression()
	if unary == nil || unary.Operator != ast.KindExclamationToken {
		return nil
	}
	return unary.Operand
}

// isBooleanIdentifier answers whether an expression is the bare name `Boolean`.
//
// A name comparison and not a resolution, matching oxc. Under a locally-declared `function
// Boolean(x)` this reports and ESLint does not, because ESLint additionally asks whether the name
// resolves to the global. Reproduced rather than improved on: the divergence is upstream's, it is
// untested on either side, and closing it here would make this rule the only one in the package
// whose answer depends on scope analysis.
func isBooleanIdentifier(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	return node.Text() == "Boolean"
}

// isBooleanCallOrConstruction answers whether a node is `Boolean(...)` or `new Boolean(...)`.
//
// `new Boolean(...)` counts as a coercing *context* while never being reported itself, which is why
// `if (new Boolean(foo)) {}` is valid and `new Boolean(!!foo)` is not.
func isBooleanCallOrConstruction(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindCallExpression:
		return isBooleanIdentifier(node.AsCallExpression().Expression)
	case ast.KindNewExpression:
		return isBooleanIdentifier(node.AsNewExpression().Expression)
	}
	return false
}

// callOrNew narrows a node to a call or construction, or nil.
func callOrNew(node *ast.Node) *ast.Node {
	if node != nil && (node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression) {
		return node
	}
	return nil
}

// argumentNodes returns a call's or construction's arguments, tolerating a missing list.
func argumentNodes(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindCallExpression:
		if list := node.AsCallExpression().Arguments; list != nil {
			return list.Nodes
		}
	case ast.KindNewExpression:
		if list := node.AsNewExpression().Arguments; list != nil {
			return list.Nodes
		}
	}
	return nil
}

// isLogicalNot answers whether a node is a `!` expression.
func isLogicalNot(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	unary := node.AsPrefixUnaryExpression()
	return unary != nil && unary.Operator == ast.KindExclamationToken
}

// isNullishCoalescing answers whether an expression is a `??` at its top level.
func isNullishCoalescing(node *ast.Node) bool {
	return binaryOperatorIs(node, ast.KindQuestionQuestionToken)
}

// isLogicalAndOr answers whether an expression is a `&&` or `||` at its top level.
func isLogicalAndOr(node *ast.Node) bool {
	return binaryOperatorIs(node, ast.KindAmpersandAmpersandToken) ||
		binaryOperatorIs(node, ast.KindBarBarToken)
}

// binaryOperatorIs answers whether a node is a binary expression with the given operator.
func binaryOperatorIs(node *ast.Node, operator ast.Kind) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	return binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == operator
}

// sameExpression answers whether a slot holds this node, looking through parentheses.
//
// Compared by identity after skipping parentheses, which is what upstream's node-id comparison does.
// Comparing spans instead would confuse `(a)` with `a` in a file where both appear at the same
// offsets, and comparing text would confuse any two identical subexpressions.
func sameExpression(slot *ast.Node, node *ast.Node) bool {
	if slot == nil || node == nil {
		return false
	}
	return ast.SkipParentheses(slot) == ast.SkipParentheses(node)
}

// isIdentifierPart answers whether a rune may continue an identifier.
func isIdentifierPart(character rune) bool {
	return character == '$' || character == '_' || unicode.IsLetter(character) ||
		unicode.IsDigit(character)
}

// isIdentifierStart answers whether a rune may begin an identifier.
func isIdentifierStart(character rune) bool {
	return character == '$' || character == '_' || unicode.IsLetter(character)
}

// isASCIISpace answers whether a byte is ASCII whitespace.
func isASCIISpace(character byte) bool {
	switch character {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}
