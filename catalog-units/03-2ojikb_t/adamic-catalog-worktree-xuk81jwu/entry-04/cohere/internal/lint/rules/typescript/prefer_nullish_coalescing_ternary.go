package typescript

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	type_checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The ternary and if-statement arms of prefer-nullish-coalescing.
//
// Both arms ask one question of a test: does it check its subject for null and undefined, and
// nothing else, so that `??` (or `??=`) would say the same thing? Upstream answers it with three
// functions, `getOperatorAndNodesInsideTestExpression`, `getBranchNodes` and
// `getNullishCoalescingParams`, and they port here under the same names so the two can be read side
// by side.
//
// # Every shape test reads through parentheses, because upstream's parser folds them away
//
// ESTree has no ParenthesizedExpression. `(a !== null) && (a !== undefined) ? a : b` presents its
// test to upstream as a LogicalExpression of two BinaryExpressions, and `(a) ? a : b` presents an
// Identifier. Ours keeps every pair as a node, so each node this file inspects is unwrapped first,
// and a test written with defensive parentheses reads exactly as upstream reads it.
//
// # The fixes reproduce upstream's text, including where that text is odd
//
// Replayed against the installed 8.67.0 build, every suggestion below matched upstream's own
// expected output on all 221 ternary and if-statement cases. Two behaviours a reader might "fix"
// are upstream's, measured, and kept:
//
//   - `getTextWithParentheses` keeps ONE pair of parentheses around the subject, the innermost, so
//     `((a)) ? ((a)) : b` suggests `(a) ?? b`.
//   - The if-statement arm accepts any assignment operator in its body, not only `=`, because
//     ESTree's AssignmentExpression covers `+=` too. `if (!a) a += b` reports and suggests
//     `a ??= b`. It is a suggestion, never applied unattended, and ESLint makes the same one.

// preferNullishCoalescingTest is what a ternary's or an if statement's test was read as.
//
// `operator` is upstream's spelling: `""` for a bare subject (`a ? a : b`), `"!"` for its negation,
// and the comparison operator otherwise. `comparedNodes` holds the operands of every comparison in
// the test, in source order, and is empty for the two truthiness shapes.
type preferNullishCoalescingTest struct {
	operator      string
	comparedNodes []*ast.Node
}

// messagePreferNullishOverTernary and messagePreferNullishOverAssignment are the two arms' ids.
var messagePreferNullishOverTernary = rule.Message{
	Id: "preferNullishOverTernary",
	Description: "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, " +
		"as it is simpler to read. The test checks the value for null and undefined and then names " +
		"it again in a branch; `??` says the same thing once, so the tested expression and the " +
		"returned one cannot drift apart.",
}

var messagePreferNullishOverAssignment = rule.Message{
	Id: "preferNullishOverAssignment",
	Description: "Prefer using nullish coalescing operator (`??=`) instead of an assignment " +
		"expression, as it is simpler to read. The `if` assigns only when the target is null or " +
		"undefined, which is exactly what `??=` does, in one statement that names the target once.",
}

// preferNullishCoalescingJudgeTernary is upstream's `ConditionalExpression` listener.
func preferNullishCoalescingJudgeTernary(ctx rule.Context, node *ast.Node,
	settings PreferNullishCoalescingOptions) (preferNullishCoalescingFinding, bool) {

	if settings.IgnoreTernaryTests {
		return preferNullishCoalescingFinding{}, false
	}
	conditional := node.AsConditionalExpression()
	test := preferNullishCoalescingUnwrap(conditional.Condition)
	reading, ok := preferNullishCoalescingReadTest(test)
	if !ok {
		return preferNullishCoalescingFinding{}, false
	}

	// upstream's `getBranchNodes`: which branch runs when the subject is NOT nullish.
	nonNullishBranch, nullishBranch := conditional.WhenTrue, conditional.WhenFalse
	switch reading.operator {
	case "", "!=", "!==":
	default:
		nonNullishBranch, nullishBranch = conditional.WhenFalse, conditional.WhenTrue
	}

	subject, fixable := preferNullishCoalescingParams(ctx, node, test,
		preferNullishCoalescingUnwrap(nonNullishBranch), reading, settings)
	if !fixable {
		return preferNullishCoalescingFinding{}, false
	}

	// The right operand keeps its parentheses when it had them, and gains a pair when its own
	// precedence would bind looser than `??`: a nested ternary, an assignment, an arrow, a sequence.
	// That is upstream's `getWrappedCode` against `OperatorPrecedence.Coalesce`, which is strict, so a
	// `||` operand (LogicalOR, one above Coalesce) is written bare, exactly as upstream writes it.
	nullishInner := preferNullishCoalescingUnwrap(nullishBranch)
	rightOperand := preferNullishCoalescingTextWithParentheses(ctx, nullishInner)
	if nullishInner == nullishBranch &&
		ast.GetExpressionPrecedence(nullishInner) <= ast.OperatorPrecedenceCoalesce {
		rightOperand = "(" + rightOperand + ")"
	}

	reportRange := rule.TokenRange(ctx.SourceFile, node)
	return preferNullishCoalescingFinding{
		reportRange: reportRange,
		message:     messagePreferNullishOverTernary,
		suggestion: rule.Message{
			Id:          messagePreferNullishSuggest.Id,
			Description: preferNullishCoalescingSuggestion("??"),
		},
		fixes: []rule.Fix{rule.ReplaceRange(reportRange,
			preferNullishCoalescingTextWithParentheses(ctx, subject)+" ?? "+rightOperand)},
	}, true
}

// preferNullishCoalescingJudgeIf is upstream's `IfStatement` listener: an `if` with no `else` whose
// body is one assignment to the subject it tested.
//
// # Silent inside a function React Compiler compiles, unless the option says the compiler is off
//
// The repair is `??=`, and babel-plugin-react-compiler 1.0.0 refuses all three logical assignment
// shorthands (BuildHIR, `Handle ??= operators in AssignmentExpression`), so taking the suggestion in a
// component or a hook costs that whole function its compilation. Measured on the waves that took it
// (#cn8sthd): SecretRow in ahra, WebSocketViaSharedWorkerProviderInternal in Structure and
// ChatReasoningAndTools in www each went from CompileSuccess to CompileError on the one `??=`, and
// compiled again in the long form. `logical-assignment-operators` already stays silent there for the
// same reason, through the same `IsInsideComponentOrHook`, so the two rules no longer pull one site in
// opposite directions. The ternary and `||` checks are not gated: `??` is an ordinary expression the
// compiler lowers, and a `||=` it would rewrite was already refused before the rewrite.
func preferNullishCoalescingJudgeIf(ctx rule.Context, node *ast.Node,
	settings PreferNullishCoalescingOptions) (preferNullishCoalescingFinding, bool) {

	ifStatement := node.AsIfStatement()
	if settings.IgnoreIfStatements || ifStatement.ElseStatement != nil {
		return preferNullishCoalescingFinding{}, false
	}

	// Exactly one statement, braced or not. A block holding the assignment and anything else, even
	// an empty statement, is left alone, because `??=` could only replace part of it.
	var statement *ast.Node
	consequent := ifStatement.ThenStatement
	isBlock := consequent.Kind == ast.KindBlock
	switch {
	case isBlock:
		statements := consequent.AsBlock().Statements
		if statements != nil && len(statements.Nodes) == 1 &&
			statements.Nodes[0].Kind == ast.KindExpressionStatement {
			statement = statements.Nodes[0]
		}
	case consequent.Kind == ast.KindExpressionStatement:
		statement = consequent
	}
	if statement == nil {
		return preferNullishCoalescingFinding{}, false
	}

	assignment := preferNullishCoalescingUnwrap(statement.AsExpressionStatement().Expression)
	if assignment == nil || assignment.Kind != ast.KindBinaryExpression ||
		!ast.IsAssignmentOperator(assignment.AsBinaryExpression().OperatorToken.Kind) {
		return preferNullishCoalescingFinding{}, false
	}
	target := preferNullishCoalescingUnwrap(assignment.AsBinaryExpression().Left)
	if !preferNullishCoalescingIsMemberAccessLike(target) {
		return preferNullishCoalescingFinding{}, false
	}

	test := preferNullishCoalescingUnwrap(ifStatement.Expression)
	reading, ok := preferNullishCoalescingReadTest(test)
	if !ok {
		return preferNullishCoalescingFinding{}, false
	}
	// Only the tests that are true when the subject IS nullish guard an assignment that `??=` can
	// replace. `if (a != null) a = b` assigns when `a` is present, the opposite of `??=`.
	switch reading.operator {
	case "!", "==", "===":
	default:
		return preferNullishCoalescingFinding{}, false
	}

	// The suggestion spells the assignment's target, not the subject the test mentions. Upstream
	// shadows its `nullishCoalescingLeftNode` with the target here, so `if (foo?.a == null) foo.a = b`
	// suggests `foo.a ??= b`: the left side of an assignment cannot carry `?.`, and the test's
	// spelling would be a syntax error. Measured: two corpus cases differ only in this.
	if _, fixable := preferNullishCoalescingParams(ctx, node, test, target, reading, settings); !fixable {
		return preferNullishCoalescingFinding{}, false
	}
	// Asked last, at a finding: the ancestor walk is paid only where the rule has something to say.
	if settings.ReactCompiler && react.IsInsideComponentOrHook(node) {
		return preferNullishCoalescingFinding{}, false
	}

	// Comments inside the `if` would vanish with it, so upstream carries them out: the ones before
	// the assignment go in front of the new statement, and in a block the ones after it go behind.
	// A block puts each on its own line; a bare body keeps them on the line, space separated.
	separator := " "
	if isBlock {
		separator = "\n"
	}
	before := preferNullishCoalescingCommentsIn(ctx,
		assignment.Pos(), rule.TokenRange(ctx.SourceFile, assignment).Pos(), separator)
	after := ""
	if isBlock {
		after = preferNullishCoalescingCommentsIn(ctx, statement.End(), consequent.End()-1, "\n")
	}

	replacement := before + preferNullishCoalescingTextWithParentheses(ctx, target) + " ??= " +
		preferNullishCoalescingTextWithParentheses(ctx,
			preferNullishCoalescingUnwrap(assignment.AsBinaryExpression().Right)) + ";"
	if after != "" {
		replacement += " " + after[:len(after)-1]
	}

	reportRange := rule.TokenRange(ctx.SourceFile, node)
	return preferNullishCoalescingFinding{
		reportRange: reportRange,
		message:     messagePreferNullishOverAssignment,
		suggestion: rule.Message{
			Id:          messagePreferNullishSuggest.Id,
			Description: preferNullishCoalescingSuggestion("??="),
		},
		fixes: []rule.Fix{rule.ReplaceRange(reportRange, replacement)},
	}, true
}

// preferNullishCoalescingReadTest is upstream's `getOperatorAndNodesInsideTestExpression`, reporting
// false where upstream's operator is null.
//
// Three shapes read: a bare or negated subject, one comparison, and two comparisons joined by `||`
// or `&&`. The joined form names its operator only when both comparisons point the same way:
// `===` and `===` under `||` is `===`, any `==` among them makes it `==`, and under `&&` the same
// holds for `!==` and `!=`. A join whose either half compares two nullish literals
// (`a !== undefined && null !== null`) is refused outright, since it tests nothing about `a`.
func preferNullishCoalescingReadTest(test *ast.Node) (preferNullishCoalescingTest, bool) {
	if test == nil {
		return preferNullishCoalescingTest{}, false
	}
	if preferNullishCoalescingIsMemberAccessLike(test) {
		return preferNullishCoalescingTest{operator: ""}, true
	}

	switch test.Kind {
	case ast.KindPrefixUnaryExpression:
		unary := test.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindExclamationToken &&
			preferNullishCoalescingIsMemberAccessLike(preferNullishCoalescingUnwrap(unary.Operand)) {
			return preferNullishCoalescingTest{operator: "!"}, true
		}
		return preferNullishCoalescingTest{}, false
	case ast.KindBinaryExpression:
	default:
		return preferNullishCoalescingTest{}, false
	}

	binary := test.AsBinaryExpression()
	if preferNullishCoalescingIsComparisonShaped(test) {
		operator := preferNullishCoalescingEqualityOperator(binary.OperatorToken.Kind)
		if operator == "" {
			return preferNullishCoalescingTest{}, false
		}
		return preferNullishCoalescingTest{operator: operator, comparedNodes: []*ast.Node{
			preferNullishCoalescingUnwrap(binary.Left), preferNullishCoalescingUnwrap(binary.Right),
		}}, true
	}

	join := binary.OperatorToken.Kind
	if join != ast.KindBarBarToken && join != ast.KindAmpersandAmpersandToken {
		return preferNullishCoalescingTest{}, false
	}
	left := preferNullishCoalescingUnwrap(binary.Left)
	right := preferNullishCoalescingUnwrap(binary.Right)
	if !preferNullishCoalescingIsComparisonShaped(left) || !preferNullishCoalescingIsComparisonShaped(right) {
		return preferNullishCoalescingTest{}, false
	}
	if preferNullishCoalescingComparesTwoNullishLiterals(left) ||
		preferNullishCoalescingComparesTwoNullishLiterals(right) {
		return preferNullishCoalescingTest{}, false
	}

	leftOperator := preferNullishCoalescingEqualityOperator(left.AsBinaryExpression().OperatorToken.Kind)
	rightOperator := preferNullishCoalescingEqualityOperator(right.AsBinaryExpression().OperatorToken.Kind)
	strict, loose := "===", "=="
	if join == ast.KindAmpersandAmpersandToken {
		strict, loose = "!==", "!="
	}
	operator := ""
	switch {
	case leftOperator == strict && rightOperator == strict:
		operator = strict
	case (leftOperator == strict || leftOperator == loose) &&
		(rightOperator == strict || rightOperator == loose):
		operator = loose
	default:
		return preferNullishCoalescingTest{}, false
	}

	return preferNullishCoalescingTest{operator: operator, comparedNodes: []*ast.Node{
		preferNullishCoalescingUnwrap(left.AsBinaryExpression().Left),
		preferNullishCoalescingUnwrap(left.AsBinaryExpression().Right),
		preferNullishCoalescingUnwrap(right.AsBinaryExpression().Left),
		preferNullishCoalescingUnwrap(right.AsBinaryExpression().Right),
	}}, true
}

// preferNullishCoalescingParams is upstream's `getNullishCoalescingParams`: given the test and the
// branch that runs when the subject is present, find the subject and decide whether `??` would
// mean the same thing.
//
// A truthiness test (`a ? a : b`) is a `||` in disguise, so it takes the `||` arm's whole
// eligibility, conditional tests and primitive exemptions included. A comparison test is judged by
// which of null and undefined it rules out:
//
//	both, or a loose `==` / `!=` on either     fixable
//	neither                                    not fixable
//	strictly undefined only                    fixable when the type cannot be null
//	strictly null only                         fixable when the type cannot be undefined
//
// and an `any` or `unknown` subject is never fixable on a one-sided strict check, since either
// could be in it.
func preferNullishCoalescingParams(ctx rule.Context, node *ast.Node, test *ast.Node,
	nonNullish *ast.Node, reading preferNullishCoalescingTest,
	settings PreferNullishCoalescingOptions) (*ast.Node, bool) {

	if len(reading.comparedNodes) == 0 {
		subject := test
		if test.Kind == ast.KindPrefixUnaryExpression {
			subject = preferNullishCoalescingUnwrap(test.AsPrefixUnaryExpression().Operand)
		}
		if !preferNullishCoalescingSimilarMemberAccess(subject, nonNullish) {
			return nil, false
		}
		return subject, preferNullishCoalescingEligible(ctx, node, subject, settings)
	}

	var subject *ast.Node
	checksNull, checksUndefined := false, false
	for _, compared := range reading.comparedNodes {
		switch {
		case compared.Kind == ast.KindNullKeyword:
			checksNull = true
		case preferNullishCoalescingIsUndefinedIdentifier(compared):
			checksUndefined = true
		case preferNullishCoalescingSimilarMemberAccess(compared, nonNullish):
			// The first mention wins. In `a?.b?.c !== undefined && a.b.c !== null ? a.b.c : d` it
			// is the one carrying every optional link, which the suggestion needs.
			if subject == nil {
				subject = compared
			}
		default:
			return nil, false
		}
	}
	if subject == nil {
		return nil, false
	}

	if checksUndefined == checksNull {
		return subject, checksUndefined
	}
	if reading.operator == "==" || reading.operator == "!=" {
		return subject, true
	}
	if ctx.TypeChecker == nil {
		return nil, false
	}
	flags := preferNullishCoalescingUnionFlags(ctx.TypeChecker.GetTypeAtLocation(subject))
	if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return nil, false
	}
	if checksUndefined {
		return subject, flags&checker.TypeFlagsNull == 0
	}
	return subject, flags&checker.TypeFlagsUndefined == 0
}

// preferNullishCoalescingIsMemberAccessLike is upstream's `isMemberAccessLike`: an identifier or a
// property access, optional or not.
//
// Upstream also admits a ChainExpression, which in ESTree wraps an optional call as well as an
// optional member. An optional call can never be similar to anything (`isNodeEqual` has no arm for
// calls), so admitting it changes no verdict, and leaving it out keeps the predicate honest about
// what it accepts.
func preferNullishCoalescingIsMemberAccessLike(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return true
	}
	return false
}

// preferNullishCoalescingIsComparisonShaped is ESTree's BinaryExpression: a binary operator that is
// not logical, not a comma and not an assignment, all of which ESTree gives their own node types.
func preferNullishCoalescingIsComparisonShaped(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	operator := node.AsBinaryExpression().OperatorToken.Kind
	switch operator {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
		ast.KindCommaToken:
		return false
	}
	return !ast.IsAssignmentOperator(operator)
}

// preferNullishCoalescingEqualityOperator spells one of the four equality operators, or "".
func preferNullishCoalescingEqualityOperator(kind ast.Kind) string {
	switch kind {
	case ast.KindEqualsEqualsToken:
		return "=="
	case ast.KindExclamationEqualsToken:
		return "!="
	case ast.KindEqualsEqualsEqualsToken:
		return "==="
	case ast.KindExclamationEqualsEqualsToken:
		return "!=="
	}
	return ""
}

// preferNullishCoalescingComparesTwoNullishLiterals is upstream's `isNodeNullishComparison`.
func preferNullishCoalescingComparesTwoNullishLiterals(comparison *ast.Node) bool {
	binary := comparison.AsBinaryExpression()
	return preferNullishCoalescingIsNullishLiteral(preferNullishCoalescingUnwrap(binary.Left)) &&
		preferNullishCoalescingIsNullishLiteral(preferNullishCoalescingUnwrap(binary.Right))
}

func preferNullishCoalescingIsNullishLiteral(node *ast.Node) bool {
	return node != nil &&
		(node.Kind == ast.KindNullKeyword || preferNullishCoalescingIsUndefinedIdentifier(node))
}

// preferNullishCoalescingIsUndefinedIdentifier is upstream's `isUndefinedIdentifier`: the name, not
// the value, so a shadowed `undefined` reads as one, as it does upstream.
func preferNullishCoalescingIsUndefinedIdentifier(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindIdentifier && node.Text() == "undefined"
}

// preferNullishCoalescingSimilarMemberAccess is upstream's `areNodesSimilarMemberAccess`: the same
// chain of names, whatever optional links it carries. `a.b.c`, `a?.b.c` and `(a?.b).c` are similar,
// and so are `a.b` and `a['b']`, a computed string key matching the name it spells.
//
// Optional chaining needs no step of its own here. ESTree wraps an optional chain in a
// ChainExpression that upstream has to skip; ours marks the access itself, so two chains differing
// only in their `?.` links are already two property accesses with the same names.
func preferNullishCoalescingSimilarMemberAccess(left *ast.Node, right *ast.Node) bool {
	left, right = preferNullishCoalescingUnwrap(left), preferNullishCoalescingUnwrap(right)
	if preferNullishCoalescingIsMemberAccess(left) && preferNullishCoalescingIsMemberAccess(right) {
		leftObject, leftProperty, leftComputed := preferNullishCoalescingMemberParts(left)
		rightObject, rightProperty, rightComputed := preferNullishCoalescingMemberParts(right)
		if !preferNullishCoalescingSimilarMemberAccess(leftObject, rightObject) {
			return false
		}
		if leftComputed == rightComputed {
			return preferNullishCoalescingNodesEqual(leftProperty, rightProperty)
		}
		// One side computed, the other not: `a['b']` against `a.b`. Only a string key can spell a
		// name, so a numeric key never matches, as upstream's `value === name` never does.
		computed, named := leftProperty, rightProperty
		if rightComputed {
			computed, named = rightProperty, leftProperty
		}
		return computed.Kind == ast.KindStringLiteral && named.Kind == ast.KindIdentifier &&
			computed.Text() == named.Text()
	}
	return preferNullishCoalescingNodesEqual(left, right)
}

// preferNullishCoalescingNodesEqual is typescript-eslint's `isNodeEqual`, and it is deliberately
// narrow: `this`, a literal, an identifier, or a member access built from those. Anything else, a
// call, a non-null assertion, a template, a private name, is never equal, even to itself.
//
// A member access compares its property and its object and IGNORES whether the access is computed,
// as upstream's does. It is reached only for a property that is itself an access (`a[b.c]` against
// `a[b.c]`), where that quirk cannot change a verdict on real code, and is kept rather than tidied
// so the two implementations answer alike.
func preferNullishCoalescingNodesEqual(left *ast.Node, right *ast.Node) bool {
	left, right = preferNullishCoalescingUnwrap(left), preferNullishCoalescingUnwrap(right)
	if left == nil || right == nil {
		return false
	}
	if preferNullishCoalescingIsMemberAccess(left) && preferNullishCoalescingIsMemberAccess(right) {
		leftObject, leftProperty, _ := preferNullishCoalescingMemberParts(left)
		rightObject, rightProperty, _ := preferNullishCoalescingMemberParts(right)
		return preferNullishCoalescingNodesEqual(leftProperty, rightProperty) &&
			preferNullishCoalescingNodesEqual(leftObject, rightObject)
	}
	if left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case ast.KindThisKeyword, ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindBigIntLiteral:
		return left.Text() == right.Text()
	case ast.KindNumericLiteral:
		// ESTree compares the literal's VALUE, so `1` and `1.0` are one key. Parsed rather than
		// trusted to arrive normalized; an unparsable spelling falls back to its text.
		leftValue, leftError := strconv.ParseFloat(left.Text(), 64)
		rightValue, rightError := strconv.ParseFloat(right.Text(), 64)
		if leftError != nil || rightError != nil {
			return left.Text() == right.Text()
		}
		return leftValue == rightValue
	}
	return false
}

func preferNullishCoalescingIsMemberAccess(node *ast.Node) bool {
	return node != nil &&
		(node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression)
}

// preferNullishCoalescingMemberParts splits an access into ESTree's object, property and computed.
func preferNullishCoalescingMemberParts(node *ast.Node) (*ast.Node, *ast.Node, bool) {
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		return access.Expression, access.Name(), false
	}
	access := node.AsElementAccessExpression()
	return access.Expression, preferNullishCoalescingUnwrap(access.ArgumentExpression), true
}

// preferNullishCoalescingUnionFlags is typescript-eslint's `getTypeFlags`: every union constituent's
// flags, ORed.
func preferNullishCoalescingUnionFlags(subject *checker.Type) checker.TypeFlags {
	var combined checker.TypeFlags
	if subject == nil {
		return combined
	}
	for part := range type_checking.UnionTypePartsSeq(subject) {
		if part != nil {
			combined |= checker.Type_flags(part)
		}
	}
	return combined
}

// preferNullishCoalescingTextWithParentheses is typescript-eslint's `getTextWithParentheses`: the
// node's text with the ONE pair of parentheses directly around it, if it has one. Not every pair:
// `((a))` gives `(a)`, measured against the installed build.
func preferNullishCoalescingTextWithParentheses(ctx rule.Context, node *ast.Node) string {
	spanned := node
	if node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		spanned = node.Parent
	}
	textRange := rule.TokenRange(ctx.SourceFile, spanned)
	return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
}

// preferNullishCoalescingCommentsIn renders the comments wholly inside [start, end) as upstream's
// `formatComments` does: each comment's own text followed by the separator.
func preferNullishCoalescingCommentsIn(ctx rule.Context, start int, end int, separator string) string {
	rendered := ""
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= start && comment.Range.End() <= end {
			rendered += comment.Text + separator
		}
	}
	return rendered
}
