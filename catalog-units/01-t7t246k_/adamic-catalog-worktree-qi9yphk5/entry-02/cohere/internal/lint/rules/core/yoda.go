package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// YodaSettings is the decoded option surface.
//
// Upstream's schema is a two element list: an enum, then an object. The rule registers with
// `DecodeOptionList`, so both arrive as written; see the decoder.
type YodaSettings struct {
	// Always requires the literal on the LEFT. Upstream's `"always"`; the default is `"never"`.
	Always bool
	// ExceptRange exempts a comparison that forms a range test, `0 <= x && x < 1`.
	ExceptRange bool
	// OnlyEquality narrows the rule to `==`, `===`, `!=` and `!==`.
	OnlyEquality bool
}

// DefaultYodaSettings is upstream's `defaultOptions: ["never", {exceptRange: false, onlyEquality: false}]`.
func DefaultYodaSettings() YodaSettings {
	return YodaSettings{}
}

// yodaObjectWire is upstream's second option element, whose keys are exactly these two.
type yodaObjectWire struct {
	ExceptRange  *bool `json:"exceptRange"`
	OnlyEquality *bool `json:"onlyEquality"`
}

// DecodeYodaOptions reads upstream's option list off the config: `["always", {"exceptRange":
// true}]`, `["never"]`, or nothing.
//
// The rule registers with `DecodeOptionList`, so both elements arrive. The config layer used to keep
// only the first, so this decoder accepted the flags inside one object beside a `when` key, and a
// nested array, to reach them; both workarounds are refused now, since neither was ever upstream's.
// A key the second element's schema does not declare is refused, which is its
// `additionalProperties: false`, and so is a third element.
func DecodeYodaOptions(list []byte) (any, error) {
	settings := DefaultYodaSettings()
	elements, err := rule.OptionElements(list, 2)
	if err != nil || len(elements) == 0 {
		return settings, err
	}

	var when string
	if err := json.Unmarshal(elements[0], &when); err != nil {
		return settings, fmt.Errorf("yoda takes \"always\" or \"never\" first, got %s", elements[0])
	}
	decoded, err := yodaSettingsFromWhen(when)
	if err != nil {
		return settings, err
	}
	settings = decoded.(YodaSettings)
	if len(elements) < 2 {
		return settings, nil
	}

	var object yodaObjectWire
	if err := rule.UnmarshalOptions(elements[1], &object); err != nil {
		return DefaultYodaSettings(), fmt.Errorf("yoda element 2: %w", err)
	}
	applyYodaFlags(&settings, object)
	return settings, nil
}

// yodaSettingsFromWhen turns the enum into settings, refusing anything upstream's schema refuses.
func yodaSettingsFromWhen(when string) (any, error) {
	switch when {
	case "never":
		return YodaSettings{}, nil
	case "always":
		return YodaSettings{Always: true}, nil
	}
	return DefaultYodaSettings(), fmt.Errorf(
		"yoda takes \"always\" or \"never\", got %q", when)
}

// applyYodaFlags copies whichever flags were written, leaving the rest at their defaults.
func applyYodaFlags(settings *YodaSettings, object yodaObjectWire) {
	if object.ExceptRange != nil {
		settings.ExceptRange = *object.ExceptRange
	}
	if object.OnlyEquality != nil {
		settings.OnlyEquality = *object.OnlyEquality
	}
}

var messageYodaExpected = rule.Message{
	Id: "expected",
	Description: "The literal is on the side a reader does not expect. A condition reads as a " +
		"question about the variable, so putting the constant first inverts the sentence and " +
		"costs a beat on every read.",
}

// Yoda requires a comparison's literal to sit on the expected side.
//
//	valid:   if (value === "red") {}
//	valid:   if (0 <= x && x < 1) {}         (with exceptRange)
//	invalid: if ("red" === value) {}         ->  if (value === "red") {}
//	invalid: if (-1 < str.indexOf(s)) {}     ->  if (str.indexOf(s) > -1) {}
//
// # What counts as a literal is wider than the literal node
//
// Upstream's `looksLikeLiteral` accepts two shapes beyond a plain literal: a NEGATED numeric
// literal, so `-1` is one atom rather than a unary expression, and a template with no
// substitutions. Both matter in the corpus: `if (-1 < str.indexOf(substr))` reports, and
// a template with no substitution reports while one carrying a substitution does not, because a
// template carrying a substitution is not a literal.
//
// The negation arm is prefix-only. Upstream reads `node.prefix`, so a postfix operator is not a
// literal, and our parser gives those different kinds anyway.
//
// # The comparison must be a comparison
//
// `if (5 & foo)` is clean in both directions because `&` is not a comparison operator. Upstream
// tests the operator against a fixed set rather than accepting every binary expression, and its two
// bitwise cases are what state it.
//
// # exceptRange is a structural test on the PARENT, and it is the bulk of this rule
//
// A range test is `0 <= x && x < 1` or `x < 0 || 1 <= x`: a logical expression whose two sides are
// comparisons using `<` or `<=`, whose shared operand is the same reference on both sides, and
// which is WRAPPED IN PARENTHESES. Every one of those four conditions is load bearing and the
// corpus states each.
//
// The parenthesis requirement is the one that reads as a bug and is not. Upstream asks
// `isParenthesised`, so `if (0 <= x && x < 1)` qualifies because the `if` supplies the parens, and
// a bare `0 < a && a < 1` as an expression statement does NOT, which is upstream's own failing case
// `i22`. Our parser keeps a `KindParenthesizedExpression` where upstream folds it away, so this is
// asked of the source text rather than of the tree; see `yodaIsParenthesised`.
//
// The literal ordering matters too: `0 <= x && x < 1` is a range and `1 < a && a < 0` is not,
// because the low bound must not exceed the high one. Upstream compares the two literal VALUES,
// and where either side is not a literal it accepts the shape anyway.
//
// # Same reference, with static keys enabled
//
// `isSameReference` is called here with no third argument, so `disableStaticComputedKey` is FALSE
// and `x.y` compares equal to `x["y"]`. That is the opposite of how `operator-assignment` calls the
// same helper, which is why this rule carries its own copy rather than sharing that one: the flag
// changes the answer on `if (0 <= a.b && a["b"] <= 100)`, a clean case here.
//
// # The fixer reorders verbatim slices rather than rebuilding the expression
//
// Upstream's `getFlippedString` cuts the source at the operator token and reassembles
// `right + betweenText + flippedOperator + afterText + left`. Both operands are copied BYTE FOR
// BYTE out of the original, so nothing inside them can be lost, which is what separates this fixer
// from the two in this tree that dropped type information: those built a replacement from node
// properties, and this one moves existing text.
//
// Measured against the installed rule on nine TypeScript operands, all of which survive the flip
// intact: `as`, `satisfies`, a non-null `!`, optional chaining, an angle-bracket assertion, a
// generic call, and comments in all three positions. The comment behaviour is worth stating because
// it is not what a reader would guess: a comment between an operand and the operator stays with the
// OPERATOR, so `'red' /* mid */ === color` becomes `color /* mid */ === 'red'`.
var Yoda = rule.Rule{
	Name: "yoda",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[YodaSettings](options)
		if !ok {
			settings = DefaultYodaSettings()
		}

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				checkYodaComparison(ctx, node, settings)
			},
		}
	},
}

// checkYodaComparison judges one binary expression.
func checkYodaComparison(ctx rule.Context, node *ast.Node, settings YodaSettings) {
	if ctx.SourceFile == nil {
		return
	}
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return
	}

	operator := binary.OperatorToken.Kind
	if !isYodaComparisonOperator(operator) {
		return
	}
	if settings.OnlyEquality && !isYodaEqualityOperator(operator) {
		return
	}

	// Unwrapped for the JUDGMENT only. Our parser keeps a `KindParenthesizedExpression` where
	// upstream's folds it away, so `(5)` is a Literal there and a wrapper here, and without this
	// the rule goes silent on eleven of upstream's own firing cases including
	// `while ((a) === 0)` and `if (((((((((((foo)))))))))) === ((((((5))))))) `.
	//
	// The FIX still slices the wrapped operands, because upstream keeps the parentheses in its
	// output: `while ((a) === 0)` repairs to `while (0 === (a))`. Unwrapping for the fix as well
	// would delete them, which is a change of text upstream does not make.
	expectedLiteral, expectedNonLiteral := yodaUnwrapParentheses(binary.Right),
		yodaUnwrapParentheses(binary.Left)
	expectedSide := "right"
	if settings.Always {
		expectedLiteral, expectedNonLiteral = yodaUnwrapParentheses(binary.Left),
			yodaUnwrapParentheses(binary.Right)
		expectedSide = "left"
	}

	// The finding is "the side that should NOT hold a literal does, and the side that should does
	// not". Both halves are required: `if (5 === 4)` is clean because both sides are literals.
	if !yodaLooksLikeLiteral(expectedNonLiteral) || yodaLooksLikeLiteral(expectedLiteral) {
		return
	}

	if settings.ExceptRange && yodaIsRangeTest(ctx, node.Parent) {
		return
	}

	// `TokenRange` rather than the token's own Pos, which includes leading trivia: a comment
	// written between the left operand and the operator would otherwise land inside the message,
	// which upstream's `{{operator}}` never carries. Two corpus cases write exactly that shape.
	operatorSpan := rule.TokenRange(ctx.SourceFile, binary.OperatorToken)
	operatorText := ctx.SourceFile.Text()[operatorSpan.Pos():operatorSpan.End()]
	finding := rule.Message{
		Id: messageYodaExpected.Id,
		Description: fmt.Sprintf("Expected literal to be on the %s side of %s. %s",
			expectedSide, operatorText, messageYodaExpected.Description),
	}

	flipped, fixable := yodaFlippedString(ctx, node, binary)
	if !fixable {
		ctx.ReportNode(node, finding)
		return
	}
	ctx.ReportNodeWithFixes(node, finding,
		rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), flipped))
}

// isYodaComparisonOperator is upstream's `/^(?:==|===|!=|!==|<|>|<=|>=)$/u`.
//
// Deliberately not every binary operator. `if (5 & foo)` is one of upstream's clean cases in both
// directions, and accepting the whole kind would report it.
func isYodaComparisonOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken,
		ast.KindLessThanToken, ast.KindGreaterThanToken,
		ast.KindLessThanEqualsToken, ast.KindGreaterThanEqualsToken:
		return true
	}
	return false
}

// isYodaEqualityOperator is upstream's `/^(?:==|===)$/u`, and it is NARROWER than it looks.
//
// Only the two positive forms count. `!=` and `!==` are NOT equality for this option's purposes, so
// `onlyEquality` exempts them along with the relational operators. That reads as an oversight and
// it is upstream's behaviour: measured against the installed rule, `'foo' !== x` under
// `onlyEquality` reports nothing while the same input without the option reports.
//
// Written down because the first port here listed all four, reasoning that a negated equality is
// still an equality. Four of upstream's clean cases say otherwise and the measurement confirms
// them; the intuitive reading is the one that is wrong.
func isYodaEqualityOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken:
		return true
	}
	return false
}

// yodaLooksLikeLiteral is upstream's `node.type === "Literal" || looksLikeLiteral(node)`.
//
// Three shapes count. A plain literal; a NEGATED numeric literal, which upstream folds into one
// atom so `-1` compares as a literal rather than as a unary expression; and a template with no
// substitutions. A template carrying a substitution is NOT a literal, which is what keeps
// a comparison of two templates clean when either carries a substitution.
func yodaLooksLikeLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return true
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		return unary != nil && unary.Operator == ast.KindMinusToken &&
			unary.Operand != nil && unary.Operand.Kind == ast.KindNumericLiteral
	}
	return false
}

// yodaNormalizedLiteralValue is upstream's `getNormalizedLiteral`, reduced to what the caller needs.
//
// The only consumer compares two of these with `<=`, so this answers a comparable value and whether
// one exists. A negative numeric literal answers its negated value, and a static template answers
// its cooked text, which is why a backtick-quoted zero and a single-quoted zero compare
// equal in a range test.
func yodaNormalizedLiteralValue(node *ast.Node) (string, bool, bool) {
	if node == nil {
		return "", false, false
	}
	switch node.Kind {
	case ast.KindNumericLiteral:
		return node.Text(), true, true
	case ast.KindBigIntLiteral:
		return node.Text(), true, true
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), false, true
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary != nil && unary.Operator == ast.KindMinusToken &&
			unary.Operand != nil && unary.Operand.Kind == ast.KindNumericLiteral {
			return "-" + unary.Operand.Text(), true, true
		}
	}
	return "", false, false
}

// yodaLiteralNotGreater says whether the low bound of a range does not exceed the high one.
//
// Upstream writes `leftLiteral.value <= rightLiteral.value` against JavaScript values, so numbers
// compare numerically and strings lexically. Both are reproduced here rather than compared as text,
// because `10` and `9` order the other way as strings and the corpus has numeric bounds.
//
// Where the two sides are of different types, upstream's `<=` coerces. Not reproduced, and the
// reason is that no corpus case mixes them: every range test compares two numbers, two strings or
// two bigints. A mixed pair answers false here, which declines the exemption and reports, and that
// is the conservative direction.
func yodaLiteralNotGreater(left *ast.Node, right *ast.Node) bool {
	leftText, leftNumeric, leftIsLiteral := yodaNormalizedLiteralValue(left)
	rightText, rightNumeric, rightIsLiteral := yodaNormalizedLiteralValue(right)

	// Upstream: neither a literal means not a range; exactly one means accept the shape anyway.
	if !leftIsLiteral && !rightIsLiteral {
		return false
	}
	if !leftIsLiteral || !rightIsLiteral {
		return true
	}
	if leftNumeric != rightNumeric {
		return false
	}
	if !leftNumeric {
		return leftText <= rightText
	}
	leftValue, leftOk := yodaParseNumber(leftText)
	rightValue, rightOk := yodaParseNumber(rightText)
	if !leftOk || !rightOk {
		return leftText <= rightText
	}
	return leftValue <= rightValue
}

// yodaParseNumber renders a numeric literal's canonical text as a float for ordering.
//
// The parser has already normalised the text, so `0x10` arrives as "16" and `1e1` as "10", which
// means this only has to read a decimal rendering. A bigint's `n` suffix is stripped, since
// upstream compares bigints with the same `<=`.
func yodaParseNumber(text string) (float64, bool) {
	if text == "" {
		return 0, false
	}
	if text[len(text)-1] == 'n' {
		text = text[:len(text)-1]
	}
	var value float64
	if _, err := fmt.Sscanf(text, "%g", &value); err != nil {
		return 0, false
	}
	return value, true
}

// yodaIsRangeTest is upstream's `isRangeTest`, asked of the comparison's PARENT.
//
// Four conditions, all load bearing and each stated by the corpus: the parent is a logical `&&` or
// `||`, both its sides are comparisons using `<` or `<=`, the shared operand is the same reference
// on both sides, and the whole thing is wrapped in parentheses.
func yodaIsRangeTest(ctx rule.Context, parent *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	logical := parent.AsBinaryExpression()
	if logical == nil || logical.OperatorToken == nil {
		return false
	}
	isAnd := logical.OperatorToken.Kind == ast.KindAmpersandAmpersandToken
	isOr := logical.OperatorToken.Kind == ast.KindBarBarToken
	if !isAnd && !isOr {
		return false
	}

	left, right := logical.Left, logical.Right
	if left == nil || right == nil ||
		left.Kind != ast.KindBinaryExpression || right.Kind != ast.KindBinaryExpression {
		return false
	}
	leftComparison, rightComparison := left.AsBinaryExpression(), right.AsBinaryExpression()
	if leftComparison == nil || rightComparison == nil ||
		leftComparison.OperatorToken == nil || rightComparison.OperatorToken == nil {
		return false
	}
	if !isYodaRangeOperator(leftComparison.OperatorToken.Kind) ||
		!isYodaRangeOperator(rightComparison.OperatorToken.Kind) {
		return false
	}

	// A "between" test shares the INNER operands, `0 <= x && x < 1`; an "outside" test shares the
	// OUTER ones, `x < 0 || 1 <= x`. Which pair is compared is what distinguishes them.
	between := isAnd &&
		yodaIsSameReference(leftComparison.Right, rightComparison.Left) &&
		yodaLiteralNotGreater(leftComparison.Left, rightComparison.Right)
	outside := isOr &&
		yodaIsSameReference(leftComparison.Left, rightComparison.Right) &&
		yodaLiteralNotGreater(leftComparison.Right, rightComparison.Left)
	if !between && !outside {
		return false
	}

	return yodaIsParenthesised(ctx, parent)
}

// isYodaRangeOperator is upstream's `["<", "<="].includes(operator)`.
func isYodaRangeOperator(operator ast.Kind) bool {
	return operator == ast.KindLessThanToken || operator == ast.KindLessThanEqualsToken
}

// yodaIsParenthesised asks whether the token before the node is `(` and the one after is `)`.
//
// Upstream's `astUtils.isParenthesised` asks exactly that of its token stream. It cannot be asked
// of our tree instead, and the difference is the whole point: our parser keeps a
// `KindParenthesizedExpression` node only where the source actually wrote parentheses, while the
// parentheses of an `if` belong to the `if` statement and produce no node at all. Upstream counts
// BOTH, which is why `if (0 <= x && x < 1)` is a range test and the bare expression statement
// `0 < a && a < 1` is not -- upstream's own failing case.
//
// Read off the source text for that reason, scanning past trivia in each direction.
func yodaIsParenthesised(ctx rule.Context, node *ast.Node) bool {
	source := ctx.SourceFile.Text()
	span := rule.TokenRange(ctx.SourceFile, node)

	before := span.Pos() - 1
	for before >= 0 && yodaIsSpace(source[before]) {
		before--
	}
	if before < 0 || source[before] != '(' {
		return false
	}

	after := span.End()
	for after < len(source) && yodaIsSpace(source[after]) {
		after++
	}
	return after < len(source) && source[after] == ')'
}

// yodaIsSpace says whether a byte is whitespace.
//
// Comments are deliberately NOT skipped. Upstream walks a token stream where a comment is not a
// token, so `if (/* c */ 0 <= x && x < 1)` is parenthesised there and is not here. No corpus case
// writes that shape, so the divergence is recorded rather than measured, and it errs toward
// reporting rather than toward silence.
func yodaIsSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}

// yodaIsSameReference is upstream's `isSameReference` with `disableStaticComputedKey` FALSE.
//
// A separate copy from `operatorAssignmentIsSameReference`, which passes that flag as TRUE, and the
// flag changes real answers rather than being a detail. With it false a dot access and a bracket
// access with the same static key are the SAME reference, so `a.b` equals `a["b"]` equals
// a template subscript naming the same key. Upstream's clean case
// `if (0 <= a.b && a["b"] <= 100)` and its firing case
// `if (0 <= a[b] && a.b < 1) {}` are the pair that separates the two settings: sharing the other
// copy would make the first report and the second silent.
//
// The static-key branch is `property.AccessedName` on the shelf, which answers for the dotted form
// and a string or template subscript alike and answers nothing for a computed access through a
// variable. That last part is what keeps `if (0 <= a[b()] && a[b()] < 1)` reporting: two calls are
// not knowably the same reference even though the text matches.
func yodaIsSameReference(left *ast.Node, right *ast.Node) bool {
	left = yodaUnwrapParentheses(left)
	right = yodaUnwrapParentheses(right)
	if left == nil || right == nil {
		return false
	}

	// The static-key equivalence, which is checked BEFORE the kind comparison because it is what
	// lets a property access equal an element access.
	if yodaIsAccess(left) && yodaIsAccess(right) {
		if name, known := yodaStaticPropertyName(left); known {
			otherName, otherKnown := yodaStaticPropertyName(right)
			return otherKnown && name == otherName &&
				yodaIsSameReference(yodaAccessObject(left), yodaAccessObject(right))
		}
	}

	if left.Kind != right.Kind {
		return false
	}

	switch left.Kind {
	case ast.KindThisKeyword, ast.KindSuperKeyword:
		return true

	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return left.Text() == right.Text()

	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindRegularExpressionLiteral:
		// Upstream compares COOKED literal values, and `Text()` is the cooked value for a string
		// and the canonical rendering for a number, so `0x10` and `16` compare equal here exactly
		// as they do upstream.
		return left.Text() == right.Text()

	case ast.KindPropertyAccessExpression:
		return yodaIsSameReference(left.AsPropertyAccessExpression().Name(),
			right.AsPropertyAccessExpression().Name()) &&
			yodaIsSameReference(left.AsPropertyAccessExpression().Expression,
				right.AsPropertyAccessExpression().Expression)

	case ast.KindElementAccessExpression:
		return yodaIsSameReference(left.AsElementAccessExpression().ArgumentExpression,
			right.AsElementAccessExpression().ArgumentExpression) &&
			yodaIsSameReference(left.AsElementAccessExpression().Expression,
				right.AsElementAccessExpression().Expression)
	}
	return false
}

// yodaStaticPropertyName is upstream's `getStaticPropertyName`, which is WIDER than the shelf's.
//
// `property.AccessedName` answers for an identifier, a string, a template and a number, and
// declines a REGULAR EXPRESSION literal. Upstream's `getStaticStringValue` renders one as
// `/pattern/flags`, so `a[/(?<zero>0)/]` and `a['/(?<zero>0)/']` are the same reference there. That
// exact pair is one of upstream's clean cases and it is the only thing in this corpus that
// separates the two spellings.
//
// Handled here rather than by widening the shelf. `property.Name` has 80 call sites across this
// tree and accepting a new kind would change every one of them silently; whether each wants a
// regex is a measurement nobody has taken. Recorded as a divergence between the shelf and upstream
// so the next reader finds it, rather than fixed in passing.
func yodaStaticPropertyName(access *ast.Node) (string, bool) {
	if name, known := property.AccessedName(access, property.Static); known {
		return name, true
	}
	if access.Kind != ast.KindElementAccessExpression {
		return "", false
	}
	argument := yodaUnwrapParentheses(access.AsElementAccessExpression().ArgumentExpression)
	if argument == nil || argument.Kind != ast.KindRegularExpressionLiteral {
		return "", false
	}
	// A regex literal's own text is already `/pattern/flags`, which is what upstream builds.
	return argument.Text(), true
}

// yodaIsAccess says whether a node is a member access in either spelling.
func yodaIsAccess(node *ast.Node) bool {
	return node.Kind == ast.KindPropertyAccessExpression ||
		node.Kind == ast.KindElementAccessExpression
}

// yodaAccessObject returns the object half of a member access.
func yodaAccessObject(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}

// yodaUnwrapParentheses removes parenthesis nodes our parser keeps and upstream's folds away.
//
// A loop rather than one step, because `((x))` nests, and written out rather than through
// `ast.SkipParentheses`, which dereferences its argument.
func yodaUnwrapParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// yodaFlippedString is upstream's `getFlippedString`, and the shape of it is the safety argument.
//
// The two operands are copied BYTE FOR BYTE out of the source, cut at the operator token, and
// reassembled in the other order with the operator flipped. Nothing is rebuilt from node
// properties, so nothing a node carries can be dropped: a type assertion, a `satisfies`, a non-null
// `!`, optional chaining, type arguments and comments all move with the text they belong to. That
// is the difference between this fixer and the two in this tree that silently deleted type
// information, both of which CONSTRUCTED a replacement.
//
// Measured against the installed rule on nine TypeScript operands, all intact after the flip.
//
// The spacing halves are upstream's too. A flip can put two tokens against each other that could
// not be adjacent, `a=1 === x` becoming `a=x === 1` is fine but a bare `-1<x` would not be, so a
// space is inserted only when the neighbouring token is flush AND the pair would fuse.
func yodaFlippedString(ctx rule.Context, node *ast.Node,
	binary *ast.BinaryExpression) (string, bool) {

	source := ctx.SourceFile.Text()
	span := rule.TokenRange(ctx.SourceFile, node)
	operatorStart := binary.OperatorToken.Pos()
	operatorEnd := binary.OperatorToken.End()

	// The operator token's Pos includes leading trivia, so trim it to the token itself; otherwise
	// the text before the operator is swallowed into the operator's own slice.
	trimmedOperator := rule.TokenRange(ctx.SourceFile, binary.OperatorToken)
	operatorStart = trimmedOperator.Pos()
	operatorEnd = trimmedOperator.End()

	leftSpan := rule.TokenRange(ctx.SourceFile, binary.Left)
	rightSpan := rule.TokenRange(ctx.SourceFile, binary.Right)
	if span.Pos() < 0 || rightSpan.End() > len(source) ||
		leftSpan.End() > operatorStart || operatorEnd > rightSpan.Pos() {
		return "", false
	}

	leftText := source[span.Pos():leftSpan.End()]
	textBeforeOperator := source[leftSpan.End():operatorStart]
	textAfterOperator := source[operatorEnd:rightSpan.Pos()]
	rightText := source[rightSpan.Pos():span.End()]

	flipped, known := yodaFlippedOperator(binary.OperatorToken.Kind)
	if !known {
		return "", false
	}

	prefix, suffix := "", ""
	if span.Pos() > 0 && !yodaIsSpace(source[span.Pos()-1]) &&
		!yodaTokensCanBeAdjacent(source[span.Pos()-1], rightText) {
		prefix = " "
	}
	if span.End() < len(source) && !yodaIsSpace(source[span.End()]) &&
		!yodaTokensCanBeAdjacentReversed(leftText, source[span.End()]) {
		suffix = " "
	}

	return prefix + rightText + textBeforeOperator + flipped + textAfterOperator +
		leftText + suffix, true
}

// yodaFlippedOperator is upstream's OPERATOR_FLIP_MAP.
//
// The equality operators map to themselves and the relational ones reverse, which is what makes the
// repair preserve meaning: `-1 < x` and `x > -1` are the same test.
func yodaFlippedOperator(operator ast.Kind) (string, bool) {
	switch operator {
	case ast.KindEqualsEqualsEqualsToken:
		return "===", true
	case ast.KindExclamationEqualsEqualsToken:
		return "!==", true
	case ast.KindEqualsEqualsToken:
		return "==", true
	case ast.KindExclamationEqualsToken:
		return "!=", true
	case ast.KindLessThanToken:
		return ">", true
	case ast.KindGreaterThanToken:
		return "<", true
	case ast.KindLessThanEqualsToken:
		return ">=", true
	case ast.KindGreaterThanEqualsToken:
		return "<=", true
	}
	return "", false
}

// yodaTokensCanBeAdjacent says whether a byte and the text following it can sit flush.
//
// Upstream re-tokenizes the pair. Reproduced narrowly: two tokens fuse only when the last character
// of the first and the first character of the second are both identifier or number characters, or
// when both are operator characters that could form a longer operator.
func yodaTokensCanBeAdjacent(previous byte, following string) bool {
	if following == "" {
		return true
	}
	return !yodaCharactersFuse(previous, following[0])
}

// yodaTokensCanBeAdjacentReversed is the same question at the other end of the expression.
func yodaTokensCanBeAdjacentReversed(preceding string, next byte) bool {
	if preceding == "" {
		return true
	}
	return !yodaCharactersFuse(preceding[len(preceding)-1], next)
}

// yodaCharactersFuse says whether two adjacent characters would read as one token.
func yodaCharactersFuse(left byte, right byte) bool {
	if yodaIsWordCharacter(left) && yodaIsWordCharacter(right) {
		return true
	}
	// `+ +` must not become `++`, and the same for `-`, which is the pair a flipped comparison can
	// actually produce: a negated numeric literal moving next to an operator.
	return (left == '+' || left == '-') && left == right
}

// yodaIsWordCharacter says whether a byte can appear inside an identifier or a number.
func yodaIsWordCharacter(b byte) bool {
	return b == '_' || b == '$' ||
		(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
