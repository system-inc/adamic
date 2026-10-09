package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

const messageOperatorAssignmentReplacedId = "replaced"
const messageOperatorAssignmentUnexpectedId = "unexpected"

// operatorAssignmentReplacedMessage is the `always` finding: a long form that has a shorthand.
func operatorAssignmentReplacedMessage(operator string) rule.Message {
	return rule.Message{
		Id: messageOperatorAssignmentReplacedId,
		Description: fmt.Sprintf("Assignment (=) can be replaced with operator assignment (%s). "+
			"Writing the target twice makes a reader check that the two spellings match, and a "+
			"later rename of one and not the other is a bug the shorthand cannot express.",
			operator),
	}
}

// operatorAssignmentUnexpectedMessage is the `never` finding: a shorthand where the long form is
// the house style.
func operatorAssignmentUnexpectedMessage(operator string) rule.Message {
	return rule.Message{
		Id: messageOperatorAssignmentUnexpectedId,
		Description: fmt.Sprintf("Unexpected operator assignment (%s) shorthand. This project "+
			"writes the operation out, so the target and the operator are read separately rather "+
			"than fused into one token.", operator),
	}
}

// OperatorAssignmentSetting is which of the two things this rule enforces.
type OperatorAssignmentSetting string

const (
	// OperatorAssignmentAlways requires the shorthand where one exists. Upstream's default.
	OperatorAssignmentAlways OperatorAssignmentSetting = "always"

	// OperatorAssignmentNever forbids it.
	OperatorAssignmentNever OperatorAssignmentSetting = "never"
)

// OperatorAssignmentOptions carries upstream's single positional option.
//
// A pointer, because upstream's `defaultOptions: ["always"]` means an ABSENT option requires the
// shorthand. The zero value of a plain setting is the empty string, which matches neither arm and
// makes the rule silent on every file it is meant to catch, and a rule configured as a bare
// severity in the live config is exactly that case.
type OperatorAssignmentOptions struct {
	Require *OperatorAssignmentSetting
}

// DefaultOperatorAssignmentSettings is upstream's `always`.
func DefaultOperatorAssignmentSettings() OperatorAssignmentOptions {
	always := OperatorAssignmentAlways
	return OperatorAssignmentOptions{Require: &always}
}

// DecodeOperatorAssignmentOptions turns the configured value into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` for the same two reasons the other string-enum
// rules in this package are: the wire value is a bare STRING rather than an object, since cohere's
// config layer strips the severity tuple before dispatch, and the default is not the zero value.
// An unrecognised string is an error rather than a quiet fallback, because falling back would
// enforce a mode nobody asked for and say nothing.
func DecodeOperatorAssignmentOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultOperatorAssignmentSettings(), nil
	}

	var configured string
	if err := json.Unmarshal(raw, &configured); err != nil {
		return DefaultOperatorAssignmentSettings(), err
	}

	setting := OperatorAssignmentSetting(configured)
	switch setting {
	case OperatorAssignmentAlways, OperatorAssignmentNever:
		return OperatorAssignmentOptions{Require: &setting}, nil
	}
	return DefaultOperatorAssignmentSettings(),
		fmt.Errorf("operator-assignment takes \"always\" or \"never\", got %s", configured)
}

// operatorAssignmentShorthand maps a binary operator to its compound assignment form, and is the
// whole set of operators this rule knows.
//
// Upstream splits the same set in two, by whether the operator commutes, because the reversed form
// `x = 1 * x` is reportable only for a commutative operator. That split is kept below rather than
// here, so this table stays a single answer to "does a shorthand exist".
var operatorAssignmentShorthand = map[ast.Kind]struct {
	compound ast.Kind
	text     string
	// commutative marks the four operators for which `x = y OP x` is also the same computation, and
	// therefore reportable. Upstream's list is exactly `* & ^ |`. Notably `+` is NOT here: string
	// concatenation does not commute, so `x = y + x` is a different program.
	commutative bool
}{
	ast.KindAsteriskToken:                          {ast.KindAsteriskEqualsToken, "*=", true},
	ast.KindAmpersandToken:                         {ast.KindAmpersandEqualsToken, "&=", true},
	ast.KindCaretToken:                             {ast.KindCaretEqualsToken, "^=", true},
	ast.KindBarToken:                               {ast.KindBarEqualsToken, "|=", true},
	ast.KindPlusToken:                              {ast.KindPlusEqualsToken, "+=", false},
	ast.KindMinusToken:                             {ast.KindMinusEqualsToken, "-=", false},
	ast.KindSlashToken:                             {ast.KindSlashEqualsToken, "/=", false},
	ast.KindPercentToken:                           {ast.KindPercentEqualsToken, "%=", false},
	ast.KindLessThanLessThanToken:                  {ast.KindLessThanLessThanEqualsToken, "<<=", false},
	ast.KindGreaterThanGreaterThanToken:            {ast.KindGreaterThanGreaterThanEqualsToken, ">>=", false},
	ast.KindGreaterThanGreaterThanGreaterThanToken: {ast.KindGreaterThanGreaterThanGreaterThanEqualsToken, ">>>=", false},
	ast.KindAsteriskAsteriskToken:                  {ast.KindAsteriskAsteriskEqualsToken, "**=", false},
}

// operatorAssignmentLongForm is the reverse table, used by the `never` arm.
//
// Built from the table above rather than written twice, so the two cannot drift. The logical
// assignments `&&= ||= ??=` are deliberately absent from both: upstream's `prohibit` skips them by
// name, because they short-circuit and `x = x && y` is NOT the same program as `x &&= y` when `x`
// has a setter.
var operatorAssignmentLongForm = func() map[ast.Kind]struct {
	binary ast.Kind
	text   string
} {
	table := map[ast.Kind]struct {
		binary ast.Kind
		text   string
	}{}
	for binary, shorthand := range operatorAssignmentShorthand {
		table[shorthand.compound] = struct {
			binary ast.Kind
			text   string
		}{binary, strings.TrimSuffix(shorthand.text, "=")}
	}
	return table
}()

// OperatorAssignment requires or forbids compound assignment where a shorthand exists.
//
//	valid (always):   x = y + x           the target is on the right of a non-commutative operator
//	valid (always):   x = (x + y) - z     the adjacent operand is a group, not the target
//	valid (always):   x = x && y          logical operators have no compound form here
//	valid (never):    x = x + y
//	invalid (always): x = x + 1           reports, fixes to `x += 1`
//	invalid (always): x = 1 * x           reports, and is deliberately NOT fixed
//	invalid (never):  x *= y + 1          reports, fixes to `x = x * (y + 1)`
//
// # Two arms that share nothing but a listener
//
// `always` looks at a plain `=` whose right side is a binary expression, and asks whether the
// target reappears as one of that expression's operands. `never` looks at a compound assignment and
// expands it. They are written as separate functions because they share no logic, which is
// upstream's own shape: it installs one or the other as the listener rather than branching inside.
//
// # Which side the target may appear on
//
//	x = x OP y     always reportable, for every operator with a shorthand
//	x = y OP x     reportable only when OP commutes, which upstream limits to `* & ^ |`
//
// `+` is not commutative here even though addition is, because `+` is also string concatenation:
// `x = y + x` and `x += y` differ whenever either side is a string. Upstream's table says so and
// the corpus pins it with `x = y + x` as a clean case.
//
// The reversed form is reported and NEVER fixed, which upstream explains at the line: `x = b * a`
// evaluates `b` before `a`, and `a *= b` evaluates `a` first, so a custom `valueOf` on either side
// would observe the change. That decline is reproduced.
//
// # Comparing the two references
//
// This is upstream's `astUtils.isSameReference(left, right, true)`, and the third argument matters.
// With it set, `x.y` and `x["y"]` are NOT the same reference, which is why `x[y] = x['y'] + z` and
// `x.y = x['y'] / z` are both clean upstream. A shelf helper of the same name already exists in this
// package, in `no_self_assign.go`, and it is wrong for this rule in BOTH directions, measured
// against the installed build at 10.8.1:
//
//	x.y = x["y"] + 1    the shelf says same, upstream is CLEAN      would over-report
//	x[y] = x[y] + 1     the shelf says different, upstream REPORTS  would lose the finding
//
// So this rule carries its own comparison rather than calling that one. The second row is the
// interesting half: a non-static computed key compares by structure here, so `x[y] = x[y] + 1`
// reports, while `x[fn()] = x[fn()] + y` stays clean because a call is not a reference at all.
//
// # Parentheses, which this parser keeps and upstream's folds away
//
// Measured against the installed build: `x = (x + 1)` reports and fixes to `x += 1`,
// `x = ((x + 1))` likewise, `(x) = x + 1` reports and fixes to `(x) += 1`, and `x = (x) + 1` fixes
// to `x += 1`. So parentheses are transparent to every question this rule asks, and the unwrap runs
// on the right-hand side, on the assignment target, and on each operand of the binary expression.
//
// The unwrap is a loop rather than one step, because `((x))` nests, and it is written here rather
// than reaching for `ast.SkipParentheses`, which dereferences its argument.
var OperatorAssignment = rule.Rule{
	Name: "operator-assignment",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[OperatorAssignmentOptions](options)
		if settings.Require == nil {
			settings = DefaultOperatorAssignmentSettings()
		}
		never := *settings.Require == OperatorAssignmentNever

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if never {
					operatorAssignmentProhibit(ctx, node)
					return
				}
				operatorAssignmentVerify(ctx, node)
			},
		}
	},
}

// operatorAssignmentUnwrapParentheses strips every layer of parentheses from an expression.
//
// Written as a loop with its own nil check rather than calling `ast.SkipParentheses`, which
// dereferences its argument. Every caller here can be handed a nil operand from a recovered parse.
func operatorAssignmentUnwrapParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// operatorAssignmentVerify is upstream's `cohere`, the `always` arm.
func operatorAssignmentVerify(ctx rule.Context, node *ast.Node) {
	assignment := node.AsBinaryExpression()
	if assignment.OperatorToken == nil || assignment.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}

	right := operatorAssignmentUnwrapParentheses(assignment.Right)
	if right == nil || right.Kind != ast.KindBinaryExpression {
		return
	}
	operation := right.AsBinaryExpression()
	if operation.OperatorToken == nil {
		return
	}
	shorthand, hasShorthand := operatorAssignmentShorthand[operation.OperatorToken.Kind]
	if !hasShorthand {
		return
	}

	target := operatorAssignmentUnwrapParentheses(assignment.Left)
	operationLeft := operatorAssignmentUnwrapParentheses(operation.Left)
	operationRight := operatorAssignmentUnwrapParentheses(operation.Right)

	if operatorAssignmentIsSameReference(target, operationLeft) {
		if fix, canFix := operatorAssignmentContractFix(ctx, node, assignment, right, shorthand.text); canFix {
			ctx.ReportNodeWithFixes(node, operatorAssignmentReplacedMessage(shorthand.text), fix)
			return
		}
		ctx.ReportNode(node, operatorAssignmentReplacedMessage(shorthand.text))
		return
	}

	// The reversed form, reportable only for a commutative operator and never fixed. Upstream's
	// comment gives the reason and it is a real one rather than caution: `x = b * a` reads `b`
	// first and `x *= b` reads `x` first, so a custom valueOf on either operand observes the swap.
	if shorthand.commutative && operatorAssignmentIsSameReference(target, operationRight) {
		ctx.ReportNode(node, operatorAssignmentReplacedMessage(shorthand.text))
	}
}

// operatorAssignmentProhibit is upstream's `prohibit`, the `never` arm.
func operatorAssignmentProhibit(ctx rule.Context, node *ast.Node) {
	assignment := node.AsBinaryExpression()
	if assignment.OperatorToken == nil {
		return
	}
	longForm, isCompound := operatorAssignmentLongForm[assignment.OperatorToken.Kind]
	if !isCompound {
		return
	}

	if fix, canFix := operatorAssignmentExpandFix(ctx, node, assignment, longForm.binary, longForm.text); canFix {
		ctx.ReportNodeWithFixes(node, operatorAssignmentUnexpectedMessage(longForm.text+"="), fix)
		return
	}
	ctx.ReportNode(node, operatorAssignmentUnexpectedMessage(longForm.text+"="))
}

// operatorAssignmentIsSameReference asks whether two expressions name the same place.
//
// This is upstream's `isSameReference` with `disableStaticComputedKey` set to TRUE, which is how
// this rule calls it, and the flag is the reason this is not the shelf's helper of the same name.
// With the flag set there is no static-key equivalence at all: a dot access and a bracket access
// are compared structurally like any other pair, so `x.y` and `x["y"]` differ because one is a
// property access and the other an element access.
//
// Deliberately narrow, matching upstream's switch. Anything that is not an identifier, `this`,
// `super`, a literal, or an access built from those is false, which is what keeps
// `x[fn()] = x[fn()] + y` clean: a call expression falls off the end of the switch.
func operatorAssignmentIsSameReference(left *ast.Node, right *ast.Node) bool {
	if left == nil || right == nil {
		return false
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
		// Upstream compares cooked literal values. `Text()` is the cooked value for a string and
		// the canonical rendering for a number, so `0x10` and `16` compare equal here exactly as
		// they do upstream.
		return left.Text() == right.Text()

	case ast.KindPropertyAccessExpression:
		leftAccess := left.AsPropertyAccessExpression()
		rightAccess := right.AsPropertyAccessExpression()
		return operatorAssignmentIsSameReference(leftAccess.Name(), rightAccess.Name()) &&
			operatorAssignmentIsSameReference(
				operatorAssignmentUnwrapParentheses(leftAccess.Expression),
				operatorAssignmentUnwrapParentheses(rightAccess.Expression))

	case ast.KindElementAccessExpression:
		leftAccess := left.AsElementAccessExpression()
		rightAccess := right.AsElementAccessExpression()
		return operatorAssignmentIsSameReference(
			operatorAssignmentUnwrapParentheses(leftAccess.ArgumentExpression),
			operatorAssignmentUnwrapParentheses(rightAccess.ArgumentExpression)) &&
			operatorAssignmentIsSameReference(
				operatorAssignmentUnwrapParentheses(leftAccess.Expression),
				operatorAssignmentUnwrapParentheses(rightAccess.Expression))
	}
	return false
}

// operatorAssignmentCanBeFixed is upstream's `canBeFixed`.
//
// The repair is only safe when writing the target once instead of twice cannot change which getters,
// setters or `toString` calls run. An identifier is safe, and so is a member access whose OBJECT is
// an identifier or `this` and whose key is static. Anything deeper, such as `a.b.c`, is refused
// because evaluating `a.b` twice may not be the same as evaluating it once.
func operatorAssignmentCanBeFixed(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return true
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		// An optional access is a `ChainExpression` upstream, and `canBeFixed` names
		// `MemberExpression` only, so upstream reports and declines. Measured against the installed
		// build at 10.8.1: `obj.a = obj?.a + b` reports `replaced` and carries a null fix. The two
		// spellings still compare as the SAME reference, which is why the finding fires at all, so
		// this decline lives here rather than in the comparison.
		if access.QuestionDotToken != nil {
			return false
		}
		object := operatorAssignmentUnwrapParentheses(access.Expression)
		return object != nil &&
			(object.Kind == ast.KindIdentifier || object.Kind == ast.KindThisKeyword)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		if access.QuestionDotToken != nil {
			return false
		}
		object := operatorAssignmentUnwrapParentheses(access.Expression)
		if object == nil || (object.Kind != ast.KindIdentifier && object.Kind != ast.KindThisKeyword) {
			return false
		}
		// Upstream requires a computed key to be a Literal. A variable subscript such as `x[y]`
		// therefore reports and is not fixed, which the corpus pins.
		argument := operatorAssignmentUnwrapParentheses(access.ArgumentExpression)
		return argument != nil && operatorAssignmentIsLiteral(argument)
	}
	return false
}

// operatorAssignmentIsLiteral answers upstream's `type === "Literal"`.
func operatorAssignmentIsLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return true
	}
	return false
}

// operatorAssignmentCommentsExistBetween reports whether a comment sits in a span.
//
// Upstream calls `sourceCode.commentsExistBetween` and declines the repair when one does, because
// both fixers rewrite a span that would swallow it. The shelf's per-file comment scan answers the
// same question and is cached across every rule that asks.
func operatorAssignmentCommentsExistBetween(ctx rule.Context, from int, to int) bool {
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= from && comment.Range.End() <= to {
			return true
		}
	}
	return false
}

// operatorAssignmentContractFix builds the `always` repair: `x = x + 1` becomes `x += 1`.
//
// Upstream writes the text before the `=` unchanged, then the compound operator, then the text
// after the binary operator. Copying those two slices verbatim is what preserves everything living
// inside the replaced span, which is the hazard this project has now been bitten by twice: a fixer
// that RE-RENDERS a construct loses whatever it did not think to re-render. Nothing is re-rendered
// here, so a type annotation, a non-null assertion or a generic argument inside either slice
// survives because it is copied as bytes.
func operatorAssignmentContractFix(
	ctx rule.Context,
	node *ast.Node,
	assignment *ast.BinaryExpression,
	operation *ast.Node,
	compound string,
) (rule.Fix, bool) {
	if !operatorAssignmentCanBeFixed(operatorAssignmentUnwrapParentheses(assignment.Left)) {
		return rule.Fix{}, false
	}
	if !operatorAssignmentCanBeFixed(
		operatorAssignmentUnwrapParentheses(operation.AsBinaryExpression().Left)) {
		return rule.Fix{}, false
	}

	text := ctx.SourceFile.Text()
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	equalsRange := rule.TokenRange(ctx.SourceFile, assignment.OperatorToken)
	operatorRange := rule.TokenRange(ctx.SourceFile, operation.AsBinaryExpression().OperatorToken)

	// A comment anywhere between the `=` and the binary operator would be discarded by the rewrite,
	// since that whole stretch is replaced by the compound operator.
	if operatorAssignmentCommentsExistBetween(ctx, equalsRange.End(), operatorRange.Pos()) {
		return rule.Fix{}, false
	}

	leftText := text[nodeRange.Pos():equalsRange.Pos()]
	// The slice ends at the UNWRAPPED binary expression rather than at the assignment node. Those
	// are the same offset upstream, whose parser folds parentheses away, and they differ here: for
	// `x = (x + y)` the node ends after the closing paren while the binary ends before it, and
	// slicing to the node carried a stray `)` into the replacement. Upstream's own fix for that
	// input spans the whole node and writes `x += y`, which is what this reproduces.
	rightText := text[operatorRange.End():operation.End()]

	return rule.ReplaceRange(core.NewTextRange(nodeRange.Pos(), node.End()),
		leftText+compound+rightText), true
}

// operatorAssignmentExpandFix builds the `never` repair: `x += 1` becomes `x = x + 1`.
//
// Two things make this harder than its twin, and both are upstream's.
//
// The right operand may need parentheses. `x *= y + 1` becomes `x = x * (y + 1)` rather than
// `x = x * y + 1`, which is a different computation. Upstream compares the right operand's
// precedence against the NEW binary operator's and wraps when it does not bind tighter.
// `x <<= y + 1` is the case that shows the comparison is against the new operator rather than a
// constant: `+` binds tighter than `<<`, so no parentheses are added.
//
// The operator may need a space after it. `x+=+y` becomes `x= x+ +y`, because `+` followed by `+`
// would lex as `++`. Upstream asks a tokenizer whether the two can sit adjacent; the set of pairs
// that cannot is small and closed, and it was enumerated rather than reasoned about.
func operatorAssignmentExpandFix(
	ctx rule.Context,
	node *ast.Node,
	assignment *ast.BinaryExpression,
	binaryOperator ast.Kind,
	operatorText string,
) (rule.Fix, bool) {
	target := operatorAssignmentUnwrapParentheses(assignment.Left)
	if !operatorAssignmentCanBeFixed(target) {
		return rule.Fix{}, false
	}

	text := ctx.SourceFile.Text()
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	operatorRange := rule.TokenRange(ctx.SourceFile, assignment.OperatorToken)

	// A comment before the operator would be duplicated, since the left text is written twice.
	if operatorAssignmentCommentsExistBetween(ctx, nodeRange.Pos(), operatorRange.Pos()) {
		return rule.Fix{}, false
	}

	leftText := text[nodeRange.Pos():operatorRange.Pos()]
	right := assignment.Right

	var rightText string
	switch {
	case operatorAssignmentNeedsParentheses(right, binaryOperator):
		// The gap between the operator and the operand is preserved, then the operand is wrapped.
		rightRange := rule.TokenRange(ctx.SourceFile, right)
		gap := text[operatorRange.End():rightRange.Pos()]
		rightText = gap + "(" + text[rightRange.Pos():right.End()] + ")"
	default:
		prefix := ""
		remainder := text[operatorRange.End():node.End()]
		if operatorAssignmentNeedsSpaceBefore(operatorText, remainder) {
			prefix = " "
		}
		rightText = prefix + remainder
	}

	return rule.ReplaceRange(core.NewTextRange(nodeRange.Pos(), node.End()),
		leftText+"= "+leftText+operatorText+rightText), true
}

// operatorAssignmentNeedsParentheses answers whether expanding would change precedence.
//
// Upstream's test is `getPrecedence(right) <= getPrecedence({BinaryExpression, newOperator})` and
// `!isParenthesised(right)`. The second half is why an already-parenthesized operand is not wrapped
// twice: measured, `x *= (y + 1)` becomes `x = x * (y + 1)` with one pair of parentheses.
func operatorAssignmentNeedsParentheses(right *ast.Node, binaryOperator ast.Kind) bool {
	if right == nil {
		return false
	}
	// Upstream's `!isParenthesised(right)` half, kept because it is upstream's own test and
	// INERT here, which is worth saying rather than leaving for the next reader to rediscover.
	// A mutation deleting it survived the whole corpus, and the reason is not a fixture gap:
	// `operatorAssignmentPrecedence` answers 20 for a parenthesized expression while the loosest
	// binary operator scores 15, so the comparison below already declines every input this line
	// could decline. No distinguishing input exists as long as those two numbers stay apart.
	//
	// It stays because the equivalence is a property of the precedence table rather than of this
	// function, so a later edit there could separate them silently.
	if right.Kind == ast.KindParenthesizedExpression {
		return false
	}
	return operatorAssignmentPrecedence(right) <= operatorAssignmentBinaryPrecedence(binaryOperator)
}

// operatorAssignmentPrecedence is upstream's `getPrecedence` over the kinds this rule can meet.
//
// Only the ordering against a binary operator matters, so anything binding at least as tightly as a
// unary expression collapses to one high value rather than being enumerated exactly. The unknown
// case returns -1, matching upstream's own reasoning: an unrecognised node is assumed to bind
// loosely so that the fixer parenthesizes rather than risks changing meaning.
func operatorAssignmentPrecedence(node *ast.Node) int {
	switch node.Kind {
	case ast.KindConditionalExpression:
		return 3
	case ast.KindBinaryExpression:
		operator := node.AsBinaryExpression().OperatorToken
		if operator == nil {
			return -1
		}
		if operator.Kind == ast.KindCommaToken {
			// A sequence expression is a comma-operator BinaryExpression here rather than its own
			// node kind, which is upstream's `SequenceExpression` at precedence 0.
			return 0
		}
		if ast.IsAssignmentOperator(operator.Kind) {
			return 1
		}
		return operatorAssignmentBinaryPrecedence(operator.Kind)
	case ast.KindArrowFunction, ast.KindYieldExpression:
		return 1
	case ast.KindPrefixUnaryExpression, ast.KindTypeOfExpression, ast.KindVoidExpression,
		ast.KindDeleteExpression, ast.KindAwaitExpression:
		return 16
	case ast.KindPostfixUnaryExpression:
		return 17
	case ast.KindCallExpression:
		return 18
	case ast.KindNewExpression:
		return 19
	case ast.KindAsExpression, ast.KindSatisfiesExpression:
		// TypeScript only, so upstream's table has no row for either and a port inheriting that
		// table gives them the tightest precedence by default. Measured, that is backwards: both
		// bind LOOSER than every arithmetic operator, so `x + 1 as number` parses as
		// `(x + 1) as number` rather than `x + (1 as number)`.
		//
		// That made the `never` expansion silently re-associate: `x += 1 as number` was rewritten
		// to `x = x + 1 as number`, which asserts the type of the SUM rather than of the addend.
		// The two parses were compared directly and they differ, which is what turned this from a
		// suspicion into a defect. Ranking them below the loosest binary operator makes the fixer
		// parenthesize, and no upstream fixture could have caught it because the corpus is
		// JavaScript.
		return 2
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindSuperKeyword,
		ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression, ast.KindArrayLiteralExpression,
		ast.KindObjectLiteralExpression, ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression, ast.KindTaggedTemplateExpression,
		ast.KindFunctionExpression, ast.KindClassExpression,
		ast.KindParenthesizedExpression, ast.KindNonNullExpression:
		return 20
	}
	return -1
}

// operatorAssignmentBinaryPrecedence is upstream's binary-operator table.
func operatorAssignmentBinaryPrecedence(operator ast.Kind) int {
	switch operator {
	case ast.KindQuestionQuestionToken, ast.KindBarBarToken:
		return 4
	case ast.KindAmpersandAmpersandToken:
		return 5
	case ast.KindBarToken:
		return 6
	case ast.KindCaretToken:
		return 7
	case ast.KindAmpersandToken:
		return 8
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
		ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		return 9
	case ast.KindLessThanToken, ast.KindLessThanEqualsToken,
		ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken,
		ast.KindInKeyword, ast.KindInstanceOfKeyword:
		return 10
	case ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanToken:
		return 11
	case ast.KindPlusToken, ast.KindMinusToken:
		return 12
	case ast.KindAsteriskToken, ast.KindSlashToken, ast.KindPercentToken:
		return 13
	case ast.KindAsteriskAsteriskToken:
		return 15
	}
	return -1
}

// operatorAssignmentNeedsSpaceBefore answers whether the new operator would lex into the text after
// it.
//
// Upstream asks espree whether the two tokens can be adjacent. Rather than port a tokenizer, the
// surface was enumerated against the installed build at 10.8.1: every one of the twelve compound
// operators was expanded against fourteen right-operand shapes written flush against it, and
// exactly FIVE pairs need a space.
//
//   - before  +y      would lex as `++`
//   - before  ++y     would lex as `++` then `+`
//   - before  -y      would lex as `--`
//   - before  --y     would lex as `--` then `-`
//     /  before  /re/    would lex as a line comment
//
// So the test is: does the operator's last character equal the operand's first, for the three
// characters where doubling means something else. That covers all five rows, and the enumeration is
// what makes that a measurement rather than a guess.
func operatorAssignmentNeedsSpaceBefore(operatorText string, remainder string) bool {
	trimmed := strings.TrimLeft(remainder, " \t")
	if len(trimmed) != len(remainder) || trimmed == "" {
		// Something already separates them, or there is nothing to separate.
		return false
	}
	last := operatorText[len(operatorText)-1]
	if last != trimmed[0] {
		return false
	}
	switch last {
	case '+', '-', '/':
		return true
	}
	return false
}
