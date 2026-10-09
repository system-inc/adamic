package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreferExponentiationOperator = rule.Message{
	Id: "useExponentiation",
	Description: "This calls `Math.pow` where the `**` operator says the same thing. The operator " +
		"is the language's own spelling for exponentiation, so it reads as arithmetic rather " +
		"than as a function call, and it composes with the surrounding expression under the " +
		"ordinary precedence rules instead of hiding behind a call.",
}

// preferExponentiationOperatorPrecedence is the precedence of `**`, which every parenthesization
// decision below is measured against.
//
// Read from the shared table rather than written as a literal, so the two cannot drift.
var preferExponentiationOperatorPrecedence = operatorAssignmentBinaryPrecedence(ast.KindAsteriskAsteriskToken)

// PreferExponentiationOperator flags `Math.pow(a, b)` and rewrites it to `a ** b`.
//
//	valid:   a ** b
//	valid:   pow(a, b)                    an undeclared name is not Math.pow
//	valid:   var Math; Math.pow(a, b)     a shadowed Math is a different object
//	valid:   Math[pow](a, b)              a computed key is not a static property name
//	invalid: Math.pow(a, b)               fixed to a**b
//	invalid: Math.pow(a + 1, 2)           fixed to (a + 1)**2
//	invalid: Math.pow(a, b, c)            reported, deliberately NOT fixed
//
// # The judgment is small and the repair is the port
//
// Deciding whether to report is one question: is this a call to the global `Math`'s `pow`. Deciding
// what to write is three independent parenthesization questions plus two lexical ones, and the
// corpus spends 148 of its 170 reporting cases on the answer.
//
// # Which calls count
//
// Upstream drives its scope manager's reference tracker, which follows `Math.pow` through a
// destructuring, a rename, a property alias and an object alias. Measured against the installed
// build at 10.8.1, all of these report:
//
//	Math.pow(a, b)            Math['pow'](a, b)         globalThis.Math.pow(a, b)
//	Math?.pow(a, b)           const M = Math; M.pow()   const {pow} = Math; pow()
//
// This port answers the same question through resolution, which covers the direct spellings and not
// the four alias forms. That is a stated divergence rather than an oversight, and it is bounded: an alias binds
// `pow` to a LOCAL declaration, so the checker sees an ordinary local function and cannot tell it
// from any other. Probed directly with controls, `Math.pow` and `const M = Math; M.pow` both
// resolve to an ambient declaration while a destructured, renamed or property alias resolves to a
// source one, indistinguishable from `function pow() {}`.
//
// The gap is smaller than it reads. Upstream's own corpus contains ZERO alias cases: its only bare
// `pow(a, b)` is a VALID case, clean because the name is undeclared. So the tracker reaches a shape
// upstream never tests, and this tree has 29 `Math.pow` sites and no alias sites at all, measured
// with a control. If an alias ever appears, this rule goes quiet on it rather than reporting
// wrongly, which is the safe direction for a rule that rewrites.
//
// # What the repair has to get right
//
// `**` is right-associative and binds tighter than unary minus, which is what makes the base and
// the exponent behave differently:
//
//	base needs parentheses      when it binds no tighter than `**`, OR when it is a unary or await
//	                            expression, because `-2 ** 2` is a SyntaxError rather than a
//	                            different value
//	exponent needs them         only when it binds strictly LOOSER than `**`; `a ** b ** c` already
//	                            means `a ** (b ** c)`, so a nested exponentiation on the right is
//	                            left alone
//	the whole thing needs them  when the surrounding expression would otherwise capture it
//
// Measured, the asymmetry is visible in one pair: `Math.pow(-2, 2)` becomes `(-2)**2` while
// `Math.pow(2, -2)` becomes `2**-2` with no parentheses at all.
//
// # The five refusals
//
// Upstream reports and declines to fix when the call has anything other than exactly two ordinary
// arguments, or when a comment sits anywhere inside it. The corpus spends 19 cases on this: wrong
// argument counts, spread arguments, and a comment in each of nine positions between `Math` and the
// closing parenthesis. A comment cannot be preserved by a rewrite that replaces the whole call, so
// the finding fires and the repair is withheld.
var PreferExponentiationOperator = rule.Rule{
	Name:             "prefer-exponentiation-operator",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if !preferExponentiationOperatorIsMathPow(ctx, node) {
					return
				}
				if fix, canFix := preferExponentiationOperatorFix(ctx, node); canFix {
					ctx.ReportNodeWithFixes(node, messagePreferExponentiationOperator, fix)
					return
				}
				ctx.ReportNode(node, messagePreferExponentiationOperator)
			},
		}
	},
}

// preferExponentiationOperatorIsMathPow answers whether a call is the global Math's pow.
//
// Three shapes reach here: `Math.pow(...)`, `Math['pow'](...)` and `globalThis.Math.pow(...)`, each
// optionally chained. The receiver is required to resolve to an ambient declaration, which is what
// separates the real global from a local `Math` shadowing it, and the property name is required to
// be STATIC, which is what keeps `Math[pow](a, b)` clean.
func preferExponentiationOperatorIsMathPow(ctx rule.Context, node *ast.Node) bool {
	callee := preferExponentiationOperatorUnwrapParentheses(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}

	name, receiver, ok := preferExponentiationOperatorStaticAccess(callee)
	if !ok || name != "pow" {
		return false
	}

	receiver = preferExponentiationOperatorUnwrapParentheses(receiver)
	if receiver == nil {
		return false
	}

	// The receiver has to be the global `Math`, reached either directly or through the global
	// object. `globalThis.Math.pow(a, b)` reports upstream while `globalThis.Object.pow(a, b)` is
	// clean, so the inner property name is checked rather than assumed.
	if innerName, innerReceiver, innerOk := preferExponentiationOperatorStaticAccess(receiver); innerOk {
		if innerName != "Math" {
			return false
		}
		innerReceiver = preferExponentiationOperatorUnwrapParentheses(innerReceiver)
		return innerReceiver != nil && innerReceiver.Kind == ast.KindIdentifier &&
			preferExponentiationOperatorResolvesToAnAmbientName(ctx, innerReceiver)
	}

	if receiver.Kind != ast.KindIdentifier || receiver.Text() != "Math" {
		return false
	}
	return preferExponentiationOperatorResolvesToAnAmbientName(ctx, receiver)
}

// preferExponentiationOperatorStaticAccess reads a static property name off an access expression.
//
// Both spellings count, since `Math['pow']` is the same property as `Math.pow`. A computed key that
// is not a string literal is NOT a static name, which is upstream's own distinction and why
// `Math[pow](a, b)` and a template key holding a substitution are clean, while a plain
// no-substitution template key is not.
func preferExponentiationOperatorStaticAccess(node *ast.Node) (name string, receiver *ast.Node, ok bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		if access.Name() == nil {
			return "", nil, false
		}
		// A private name is a different namespace entirely, so `Math.#pow(a, b)` is not this.
		if access.Name().Kind == ast.KindPrivateIdentifier {
			return "", nil, false
		}
		return access.Name().Text(), access.Expression, true
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		argument := preferExponentiationOperatorUnwrapParentheses(access.ArgumentExpression)
		if argument == nil {
			return "", nil, false
		}
		if value, known := preferExponentiationOperatorStaticStringValue(argument); known {
			return value, access.Expression, true
		}
		return "", nil, false
	}
	return "", nil, false
}

// preferExponentiationOperatorResolvesToAnAmbientName answers whether a name is the global one.
//
// Every declaration must live in a declaration file. A local `var Math` or a parameter named `Math`
// contributes a source declaration and the answer is false, which is how a shadowed global stops
// this rule. Probed across seven shapes with controls before being built on.
func preferExponentiationOperatorResolvesToAnAmbientName(ctx rule.Context, identifier *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		// Unresolvable. A name the checker cannot place is not something to rewrite.
		return false
	}
	// Written as the COMPLEMENT of "declared in source" rather than as "every declaration is
	// ambient", and the difference is load-bearing. `globalThis` resolves to a symbol with ZERO
	// declarations, so an all-ambient test answers false for it and `globalThis.Math.pow(a, b)`
	// went silent while upstream reports it. This is the trap the brief records against
	// `resolvesToAGlobal`, met here for real: the guard was false on exactly the case it had to
	// accept. A local `var Math` or a parameter contributes a source declaration and is still
	// rejected, which is what stops a shadowed global.
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !file.IsDeclarationFile {
			return false
		}
	}
	return true
}

// preferExponentiationOperatorUnwrapParentheses strips every layer of parentheses.
//
// A loop with its own nil check rather than `ast.SkipParentheses`, which dereferences its argument.
func preferExponentiationOperatorUnwrapParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// preferExponentiationOperatorFix builds the rewrite, or declines.
//
// Upstream's fixer, transcribed. It refuses on anything other than exactly two ordinary arguments
// and on any comment inside the call, then assembles three pieces whose parenthesization is decided
// independently.
func preferExponentiationOperatorFix(ctx rule.Context, node *ast.Node) (rule.Fix, bool) {
	call := node.AsCallExpression()
	arguments := call.Arguments.Nodes

	// Exactly two arguments, neither spread. A spread cannot be split into a base and an exponent
	// at all, and any other count is not the two-argument `Math.pow` this rule knows how to write.
	if len(arguments) != 2 {
		return rule.Fix{}, false
	}
	for _, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement {
			return rule.Fix{}, false
		}
	}

	// A comment anywhere inside the call would be discarded, since the whole call is replaced.
	// Upstream spends nine corpus cases on this, one per position a comment can occupy between
	// `Math` and the closing parenthesis.
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	if preferExponentiationOperatorHasCommentInside(ctx, nodeRange.Pos(), node.End()) {
		return rule.Fix{}, false
	}

	// Unwrap before reading the text AND before asking about precedence. Our parser keeps
	// parentheses that ESTree folds away, so `Math.pow((a), (b))` would otherwise render as
	// `(a)**(b)` where upstream writes `a**b`, and the kind tests below would see a parenthesized
	// expression rather than the operand itself.
	base := preferExponentiationOperatorUnwrapParentheses(arguments[0])
	exponent := preferExponentiationOperatorUnwrapParentheses(arguments[1])
	if base == nil || exponent == nil {
		return rule.Fix{}, false
	}
	baseText := ctx.NodeText(base)
	exponentText := ctx.NodeText(exponent)

	if preferExponentiationOperatorBaseNeedsParentheses(base) {
		baseText = "(" + baseText + ")"
	}
	if preferExponentiationOperatorExponentNeedsParentheses(exponent) {
		exponentText = "(" + exponentText + ")"
	}

	replacement := baseText + "**" + exponentText
	needsOuter := preferExponentiationOperatorNeedsOuterParentheses(node)

	// Upstream's statement-start clause. A `{`, `function` or `class` opening an expression
	// STATEMENT is parsed as a block, a function declaration or a class declaration instead, so the
	// whole expression is wrapped. Only when the base was not already parenthesized, since that
	// would have moved the opening character.
	if !needsOuter && !preferExponentiationOperatorBaseNeedsParentheses(base) &&
		preferExponentiationOperatorStartsExpressionStatement(ctx, node) &&
		preferExponentiationOperatorOpensAStatementAmbiguously(baseText) {
		needsOuter = true
	}
	if needsOuter {
		replacement = "(" + replacement + ")"
	}

	// Two lexical repairs, both upstream's. A replacement running into the token before it needs a
	// space, and one starting a statement with a character that could continue the previous line
	// needs a semicolon.
	if preferExponentiationOperatorNeedsLeadingSpace(ctx, node, replacement) {
		replacement = " " + replacement
	} else if preferExponentiationOperatorNeedsLeadingSemicolon(ctx, node, replacement) {
		replacement = ";" + replacement
	}
	if preferExponentiationOperatorNeedsTrailingSpace(ctx, node, replacement) {
		replacement += " "
	}

	return rule.ReplaceRange(core.NewTextRange(nodeRange.Pos(), node.End()), replacement), true
}

// preferExponentiationOperatorHasCommentInside answers whether a comment sits in a span.
func preferExponentiationOperatorHasCommentInside(ctx rule.Context, from int, to int) bool {
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= from && comment.Range.End() <= to {
			return true
		}
	}
	return false
}

// preferExponentiationOperatorBaseNeedsParentheses is upstream's `doesBaseNeedParens`.
//
// Two reasons, and the second is not about precedence at all. `**` is RIGHT-associative, so a base
// binding no tighter than `**` would re-associate: `Math.pow(a ** b, c)` must become
// `(a ** b) ** c` rather than `a ** b ** c`, which means `a ** (b ** c)`. And a unary or await
// expression cannot precede `**` at all, since `-2 ** 2` is a SyntaxError rather than a different
// value, which is why `Math.pow(-2, 2)` becomes `(-2)**2` rather than being declined.
func preferExponentiationOperatorBaseNeedsParentheses(base *ast.Node) bool {
	switch base.Kind {
	case ast.KindPrefixUnaryExpression:
		// Upstream's clause names `UnaryExpression`, which excludes `++a` and `--a`. Those are an
		// `UpdateExpression` there and are legal directly before `**`.
		switch base.AsPrefixUnaryExpression().Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
		default:
			return true
		}
	case ast.KindAwaitExpression, ast.KindTypeOfExpression,
		ast.KindVoidExpression, ast.KindDeleteExpression:
		return true
	}
	return preferExponentiationOperatorPrecedenceOf(base) <= preferExponentiationOperatorPrecedence
}

// preferExponentiationOperatorPrecedenceOf is upstream's `getPrecedence` for the kinds this rule
// meets, and it is NOT the table `operator-assignment` uses.
//
// That table collapses every tightly-binding kind to one value, which is enough when the only
// question is whether an operand binds looser than a binary operator. Here the answers above `**`
// have to be distinguished from each other, because a base binding NO TIGHTER than `**` is wrapped
// while one binding tighter is not, and upstream ranks update, call, member and new expressions
// each on their own rung. Measured, `Math.pow(++a, ++b)` becomes `++a**++b` with no parentheses
// while `Math.pow(a ? b : c, 2)` gains a pair, so the two kinds cannot share a value.
func preferExponentiationOperatorPrecedenceOf(node *ast.Node) int {
	switch node.Kind {
	case ast.KindBinaryExpression:
		operator := node.AsBinaryExpression().OperatorToken
		if operator == nil {
			return -1
		}
		if operator.Kind == ast.KindCommaToken {
			return 0
		}
		if ast.IsAssignmentOperator(operator.Kind) {
			return 1
		}
		return operatorAssignmentBinaryPrecedence(operator.Kind)
	case ast.KindArrowFunction, ast.KindYieldExpression:
		return 1
	case ast.KindConditionalExpression:
		return 3
	case ast.KindPrefixUnaryExpression:
		// ESTree splits this kind in two and the split is load-bearing here. `++a` and `--a` are an
		// `UpdateExpression` at precedence 17, which binds TIGHTER than `**`; `-a`, `!a` and `~a`
		// are a `UnaryExpression` at 16, which does not. Our parser puts all five under one kind,
		// so the operator decides. Measured: `Math.pow(++a, ++b)` becomes `++a**++b` with no
		// parentheses while `Math.pow(-2, 2)` gains a pair.
		//
		// The 17 is upstream's number and is INERT here, which is worth recording rather than
		// leaving for the next reader to rediscover. A mutation lowering it to 16 survives the
		// whole corpus, and the reason is not a fixture gap: `**` scores 15, so 16 and 17 give the
		// same answer at both sites that read this table, and the third site never reaches it
		// because the base test returns before the comparison for exactly these two operators.
		// `++Math.pow(a,b)` is clean upstream anyway, being an invalid assignment target, so no
		// input can put a `++` in the parent position around a reported call.
		//
		// Kept because it is upstream's own value and the equivalence is a property of where `**`
		// sits in the table rather than of this function. If a rule ever compares against an
		// operator between 16 and 17, the two stop agreeing silently.
		switch node.AsPrefixUnaryExpression().Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return 17
		}
		return 16
	case ast.KindTypeOfExpression, ast.KindVoidExpression,
		ast.KindDeleteExpression, ast.KindAwaitExpression:
		return 16
	case ast.KindPostfixUnaryExpression:
		return 17
	case ast.KindCallExpression, ast.KindTaggedTemplateExpression:
		return 18
	case ast.KindNewExpression:
		return 19
	case ast.KindAsExpression, ast.KindSatisfiesExpression:
		// Binds looser than every arithmetic operator, so a `**` under one has to be wrapped. Not
		// in upstream's table because its corpus is JavaScript; measured by comparing parses.
		return 2
	}
	// Everything else is a primary expression: identifiers, literals, member access, array and
	// object literals, function and class expressions, templates, non-null assertions.
	return 20
}

// preferExponentiationOperatorExponentNeedsParentheses is upstream's `doesExponentNeedParens`.
//
// Strictly LOOSER rather than "no tighter", which is the whole difference from the base and follows
// from right-associativity: `a ** b ** c` already groups to the right, so an exponentiation on the
// right needs nothing. Measured, this is why `Math.pow(2, -2)` becomes `2**-2` with no parentheses
// while `Math.pow(-2, 2)` gains a pair.
func preferExponentiationOperatorExponentNeedsParentheses(exponent *ast.Node) bool {
	return preferExponentiationOperatorPrecedenceOf(exponent) < preferExponentiationOperatorPrecedence
}

// preferExponentiationOperatorNeedsOuterParentheses is upstream's
// `doesExponentiationExpressionNeedParens`.
//
// The replaced call was a primary expression, so nothing around it could capture it; a `**`
// expression can be captured. Upstream wraps when the parent is an expression that binds at least
// as tightly, with four exemptions where the position already delimits the operand: the right side
// of another `**`, an argument list, a computed member key, and an array element.
func preferExponentiationOperatorNeedsOuterParentheses(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	// An ARGUMENT of a call or a new expression is already delimited by the argument list, but the
	// CALLEE is not: `Math.pow(a, b)()` must become `(a**b)()` or the call binds to `b`.
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression == node
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression == node
	case ast.KindArrayLiteralExpression:
		return false
	case ast.KindElementAccessExpression:
		if parent.AsElementAccessExpression().ArgumentExpression == node {
			return false
		}
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken != nil &&
			binary.OperatorToken.Kind == ast.KindAsteriskAsteriskToken &&
			binary.Right == node {
			// The right side of another exponentiation is already the tightly-bound side.
			return false
		}
	// A class declaration is upstream's one non-expression parent that still needs parentheses,
	// because `class C extends Math.pow(a, b) {}` becomes an extends clause holding an operator.
	case ast.KindClassDeclaration, ast.KindClassExpression,
		ast.KindExpressionWithTypeArguments, ast.KindHeritageClause:
		return true
	// TypeScript only, so upstream has no arm for it and its corpus reaches it through a custom
	// parser fixture. `as` and `satisfies` bind LOOSER than `**`, so a precedence comparison alone
	// says no wrap is needed; upstream's own asserted output is `(a**b) as any`, because the
	// assertion has to apply to the power rather than to its right operand.
	case ast.KindAsExpression, ast.KindSatisfiesExpression:
		return true
	}

	if !preferExponentiationOperatorIsExpression(parent) {
		return false
	}
	parentPrecedence := preferExponentiationOperatorPrecedenceOf(parent)
	return parentPrecedence == -1 || parentPrecedence >= preferExponentiationOperatorPrecedence
}

// preferExponentiationOperatorIsExpression answers upstream's `parent.type.endsWith("Expression")`.
//
// A statement parent never captures the operand, so the wrap is only considered inside an
// expression. Written as a list of the expression kinds this rule can actually meet rather than as
// a name test, since our kinds are not spelled the way upstream's are.
func preferExponentiationOperatorIsExpression(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindBinaryExpression, ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindConditionalExpression, ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression, ast.KindTaggedTemplateExpression,
		ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindDeleteExpression,
		ast.KindAwaitExpression, ast.KindYieldExpression, ast.KindSpreadElement,
		ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
		return true
	}
	return false
}

// preferExponentiationOperatorNeedsLeadingSpace answers whether the replacement would lex into the
// token before it.
//
// Upstream asks a tokenizer whether the two can be adjacent. The hazard here is an identifier or
// keyword running straight into the replacement's first character, as in `typeof` followed by a
// name, so the test is whether both sides can continue one identifier.
func preferExponentiationOperatorNeedsLeadingSpace(ctx rule.Context, node *ast.Node, replacement string) bool {
	start := rule.TokenRange(ctx.SourceFile, node).Pos()
	if start == 0 || replacement == "" {
		return false
	}
	text := ctx.SourceFile.Text()
	return preferExponentiationOperatorTokensWouldMerge(text[start-1], replacement[0])
}

// preferExponentiationOperatorTokensWouldMerge answers whether two adjacent characters lex as one
// token.
//
// Two ways that happens here. Two identifier characters run together, as `typeof` before a name.
// And a repeated `+` or `-` becomes an increment or decrement: measured, `a+Math.pow(++b, c)` must
// become `a+ ++b**c` rather than `a+++b**c`, which parses as `a++ +b`. The second case is why an
// identifier-only test is not enough, and the corpus spends five cases on it.
func preferExponentiationOperatorTokensWouldMerge(before byte, after byte) bool {
	if preferExponentiationOperatorContinuesIdentifier(before) &&
		preferExponentiationOperatorContinuesIdentifier(after) {
		return true
	}
	switch before {
	case '+', '-':
		return after == before
	}
	return false
}

// preferExponentiationOperatorNeedsTrailingSpace is the same question on the other side.
//
// `Math.pow(a, b)in c` becomes `a**b in c`, because `b` and `in` would otherwise form one token.
func preferExponentiationOperatorNeedsTrailingSpace(ctx rule.Context, node *ast.Node, replacement string) bool {
	text := ctx.SourceFile.Text()
	end := node.End()
	if end >= len(text) || replacement == "" {
		return false
	}
	return preferExponentiationOperatorTokensWouldMerge(replacement[len(replacement)-1], text[end])
}

// preferExponentiationOperatorNeedsLeadingSemicolon reproduces upstream's semicolon insertion.
//
// A replacement opening a statement with an opening parenthesis, bracket, slash or backtick
// can continue the previous line under
// automatic semicolon insertion, turning two statements into one. Upstream only does this when the
// call starts an expression statement and no space was already added.
func preferExponentiationOperatorNeedsLeadingSemicolon(ctx rule.Context, node *ast.Node, replacement string) bool {
	if replacement == "" {
		return false
	}
	switch replacement[0] {
	case '(', '[', '/', '`':
	default:
		return false
	}
	if !preferExponentiationOperatorStartsExpressionStatement(ctx, node) {
		return false
	}
	// Only when the previous statement could actually continue. Upstream asks
	// `needsPrecedingSemicolon`, and measured, a preceding `;` or `}` already terminates it: after
	// `foo;` or `if (foo) {}` no semicolon is added, while after `foo`, `foo()` or `var x = 1` one
	// is. A statement at the top of a file has nothing to join at all.
	start := rule.TokenRange(ctx.SourceFile, node).Pos()
	preceding := text.TrimWhitespace(ctx.SourceFile.Text()[:start])
	if preceding == "" {
		return false
	}
	switch preceding[len(preceding)-1] {
	case ';', '}', '{':
		return false
	}
	return true
}

// preferExponentiationOperatorStartsExpressionStatement answers whether this call is the first
// thing in an expression statement.
func preferExponentiationOperatorStartsExpressionStatement(ctx rule.Context, node *ast.Node) bool {
	current := node
	for current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindExpressionStatement {
			return rule.TokenRange(ctx.SourceFile, parent).Pos() ==
				rule.TokenRange(ctx.SourceFile, node).Pos()
		}
		if !preferExponentiationOperatorIsExpression(parent) &&
			parent.Kind != ast.KindParenthesizedExpression {
			return false
		}
		current = parent
	}
	return false
}

// preferExponentiationOperatorContinuesIdentifier answers whether a byte can sit inside an
// identifier.
func preferExponentiationOperatorContinuesIdentifier(character byte) bool {
	switch {
	case character >= 'a' && character <= 'z',
		character >= 'A' && character <= 'Z',
		character >= '0' && character <= '9',
		character == '_', character == '$':
		return true
	}
	return false
}

// preferExponentiationOperatorStaticStringValue evaluates a computed key to a string when it can.
//
// Upstream reads the STATIC VALUE of the property name rather than testing the node's kind, so
// three spellings that look different are the same property. Measured against the installed build:
// a plain template, a template whose only substitution is a string literal, and a concatenation of
// string literals all report, while a bare identifier key is clean.
func preferExponentiationOperatorStaticStringValue(node *ast.Node) (string, bool) {
	node = preferExponentiationOperatorUnwrapParentheses(node)
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), true
	case ast.KindTemplateExpression:
		// A template is static only when every span's expression is itself static.
		template := node.AsTemplateExpression()
		if template.Head == nil {
			return "", false
		}
		value := template.Head.Text()
		for _, span := range template.TemplateSpans.Nodes {
			templateSpan := span.AsTemplateSpan()
			inner, known := preferExponentiationOperatorStaticStringValue(templateSpan.Expression)
			if !known || templateSpan.Literal == nil {
				return "", false
			}
			value += inner + templateSpan.Literal.Text()
		}
		return value, true
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindPlusToken {
			return "", false
		}
		left, leftKnown := preferExponentiationOperatorStaticStringValue(binary.Left)
		right, rightKnown := preferExponentiationOperatorStaticStringValue(binary.Right)
		if !leftKnown || !rightKnown {
			return "", false
		}
		return left + right, true
	}
	return "", false
}

// preferExponentiationOperatorOpensAStatementAmbiguously answers whether text starting a statement
// would be read as something other than an expression.
//
// `{` opens a block, `function` a function declaration and `class` a class declaration. Upstream
// checks the first TOKEN rather than the first character, which is why an identifier merely
// beginning with those letters does not count.
func preferExponentiationOperatorOpensAStatementAmbiguously(text string) bool {
	if text == "" {
		return false
	}
	if text[0] == '{' {
		return true
	}
	for _, keyword := range []string{"function", "class"} {
		if strings.HasPrefix(text, keyword) &&
			(len(text) == len(keyword) ||
				!preferExponentiationOperatorContinuesIdentifier(text[len(keyword)])) {
			return true
		}
	}
	return false
}
