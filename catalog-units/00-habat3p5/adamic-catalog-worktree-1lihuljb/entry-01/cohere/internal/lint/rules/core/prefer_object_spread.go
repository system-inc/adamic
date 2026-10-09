package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreferObjectSpreadUseSpread = rule.Message{
	Id: "useSpreadMessage",
	Description: "Use an object spread instead of `Object.assign` eg: `{ ...foo }`. The spread " +
		"says the result is a new object at a glance, where the call has to be read to find out " +
		"whether it mutates its first argument.",
}

var messagePreferObjectSpreadUseLiteral = rule.Message{
	Id: "useLiteralMessage",
	Description: "Use an object literal instead of `Object.assign`. eg: `{ foo: bar }`. Copying " +
		"an object literal into a fresh object achieves nothing the literal did not already do.",
}

// PreferObjectSpread flags `Object.assign` called with an object literal first argument.
//
//	valid:   Object.assign(foo, bar)          (first argument is not a literal, so it mutates foo)
//	valid:   Object.assign({}, ...foo)        (an array spread argument, which spread cannot express)
//	valid:   function f(Object) { Object.assign({}, foo) }   (a shadowed Object is not the global)
//	invalid: Object.assign({}, foo)   ->  ({ ...foo})
//	invalid: Object.assign({ a: 1 })  ->  ({a: 1})
//
// # Two messages, decided by argument count alone
//
// One argument means the call copies a literal into a fresh object and achieves nothing, so the
// advice is to write the literal. Two or more means a merge, so the advice is a spread. Nothing
// else distinguishes them.
//
// # Three declines, and each rules out a shape spread cannot express
//
// An ARRAY SPREAD argument, `Object.assign({}, ...foo)`, has no object-spread equivalent: the call
// spreads a list of sources and `{ ...foo }` would spread the array itself.
//
// An ACCESSOR in any object-literal argument, when there are two or more arguments. Object spread
// evaluates a getter and copies its RESULT, while `Object.assign` also invokes it but the
// difference shows when the accessor is on a later source being merged onto an earlier one.
// Upstream declines only for the multi-argument case, so a lone `Object.assign({ get a() {} })`
// still reports. Measured against the installed rule, which reports the single-argument form and
// is silent on both two-argument forms.
//
// A SHADOWED `Object`. The calls come from the shelf's port of eslint-utils' ReferenceTracker, as
// ESLint finds them, so a parameter, a local binding or an import named `Object` is not the global
// and the call is not this rule's business, while an alias such as `const { assign } = Object` is
// (#jjfa7qb). A file that writes `Object` anywhere is not followed at all: the tracker declines a
// global that is reassigned, which is what ESLint's does.
//
// # The fixer rewrites the call into an object literal
//
// It is a token-level rewrite rather than a reconstruction, which is the whole safety argument:
// every argument's own bytes are copied verbatim by being left in place, and the edits only touch
// the call's callee, its parentheses, and the braces of literal arguments. Nothing inside an
// argument is rebuilt, so a type assertion, a template literal or a comment survives.
//
// The parenthesization is upstream's and it defaults to WRAPPING, which is the safe direction. An
// arrow body is the case that makes it load bearing: `() => Object.assign({}, foo)` must become
// `() => ({ ...foo})` and not `() => { ...foo}`, which parses as a block and returns nothing.
// Measured against the installed rule.
var PreferObjectSpread = rule.Rule{
	Name: "prefer-object-spread",

	// Telling the global `Object` from a shadow is name resolution, which the tracker asks the
	// checker for.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// Every trace starts at a reference to Object, so a file that never spells it has
				// nothing to follow, and the tracker's index is never built.
				if ctx.TypeChecker == nil || !strings.Contains(ctx.SourceFile.Text(), "Object") {
					return
				}
				tracker := reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)
				for _, tracked := range tracker.GlobalReferences(objectAssignCallTraceMap) {
					checkPreferObjectSpread(ctx, tracked.Node)
				}
			},
		}
	},
}

// objectAssignCallTraceMap asks the tracker for every call of the global Object's assign.
var objectAssignCallTraceMap = map[string]*reference.TraceMap{
	"Object": {Members: map[string]*reference.TraceMap{"assign": {Call: true}}},
}

// checkPreferObjectSpread judges one call expression.
func checkPreferObjectSpread(ctx rule.Context, node *ast.Node) {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil {
		return
	}

	call := node.AsCallExpression()
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return
	}

	arguments := call.Arguments.Nodes

	// The first argument must be an object literal. `Object.assign(foo, bar)` mutates `foo`, which
	// is a different operation entirely rather than a longer spelling of one.
	//
	// Unwrapped for the JUDGMENT, because our parser keeps a KindParenthesizedExpression where
	// upstream's folds it away: `Object.assign(({ a: 1 }), (foo))` is a plain object literal there
	// and a wrapper here, and without this the rule goes silent on one of upstream's own firing
	// cases. The FIX still reads the wrapped node, since upstream removes those parentheses too.
	if preferObjectSpreadUnwrap(arguments[0]).Kind != ast.KindObjectLiteralExpression {
		return
	}

	for _, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement {
			return
		}
	}

	// An accessor in a literal argument, only when merging. See the rule doc for why the
	// single-argument form still reports.
	if len(arguments) > 1 && preferObjectSpreadHasArgumentWithAccessors(arguments) {
		return
	}

	message := messagePreferObjectSpreadUseSpread
	if len(arguments) == 1 {
		message = messagePreferObjectSpreadUseLiteral
	}

	fixes, fixable := preferObjectSpreadFix(ctx, node, call)
	if !fixable {
		ctx.ReportNode(node, message)
		return
	}
	ctx.ReportNodeWithFixes(node, message, fixes...)
}

// preferObjectSpreadHasArgumentWithAccessors is upstream's `hasArgumentsWithAccessors`.
func preferObjectSpreadHasArgumentWithAccessors(arguments []*ast.Node) bool {
	for _, argument := range arguments {
		inner := preferObjectSpreadUnwrap(argument)
		if inner.Kind != ast.KindObjectLiteralExpression {
			continue
		}
		for _, property := range inner.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind == ast.KindGetAccessor || property.Kind == ast.KindSetAccessor {
				return true
			}
		}
	}
	return false
}

// preferObjectSpreadFix rewrites the call into an object literal.
//
// # The shape, and why it is a token rewrite rather than a reconstruction
//
// Upstream emits several small edits and so does this: remove the callee, turn the argument-list
// parentheses into braces, strip the braces off each object-literal argument, and prefix each
// non-literal argument with `...`. Every argument's own text is left exactly where it is, so
// nothing inside one can be lost. That is the difference between this fixer and the ones in this
// tree that rebuilt a span from node properties and dropped a type annotation.
//
// # Every span is derived from our own tree
//
// Upstream finds the parentheses with token queries against a token stream this tree does not
// expose the same way, and its arithmetic reads `node.range[0]` to `leftParen.range[0]` for the
// callee removal. Reproducing that offset blindly is the failure this brief warns about, so the
// opening parenthesis is found by scanning forward from the CALLEE's end rather than by assuming
// it is adjacent. Measured: it is adjacent for `Object.assign(`, two characters later for
// `Object.assign  (`, and five later for `Object.assign<Foo>(`, and the last of those is the one
// an offset assumption would corrupt.
func preferObjectSpreadFix(ctx rule.Context, node *ast.Node,
	call *ast.CallExpression) ([]rule.Fix, bool) {

	source := ctx.SourceFile.Text()
	nodeSpan := rule.TokenRange(ctx.SourceFile, node)
	calleeSpan := rule.TokenRange(ctx.SourceFile, call.Expression)

	// The opening parenthesis of the argument list, found by scanning rather than assumed. Anything
	// between the callee and it -- whitespace, type arguments -- is removed with the callee, which
	// is upstream's "remove everything before the opening paren".
	leftParen := -1
	for index := calleeSpan.End(); index < nodeSpan.End(); index++ {
		if source[index] == '(' {
			leftParen = index
			break
		}
	}
	if leftParen < 0 {
		return nil, false
	}

	// The closing parenthesis is the call's last character. Verified rather than assumed, because a
	// recovered parse can produce a call whose span ends elsewhere and the repair would then write
	// a brace over something that is not a parenthesis. Measured on the corpus's HTML-comment case,
	// which TypeScript does not parse: the span ends on the `d` of `weird`.
	//
	// This guard is REDUNDANT with the argument-span check further down, and that took a paired
	// mutation to establish. Neutralising either alone SURVIVES every fixture, because the other
	// declines the same input: with a malformed span the second argument extends past this wrong
	// `rightParen` and the argument check refuses. Neutralising BOTH fails 4 lines. Kept because
	// declining at the earliest point that can tell is cheaper than declining after building
	// per-argument edits, and because the two ask different questions of the same bad parse.
	rightParen := nodeSpan.End() - 1
	if rightParen <= leftParen || source[rightParen] != ')' {
		return nil, false
	}

	fixes := []rule.Fix{
		rule.RemoveRange(core.NewTextRange(nodeSpan.Pos(), leftParen)),
	}

	if preferObjectSpreadNeedsParens(ctx, node) {
		prefix := "({"
		if preferObjectSpreadNeedsPrecedingSemicolon(ctx, node, nodeSpan) {
			prefix = ";({"
		}
		fixes = append(fixes,
			rule.ReplaceRange(core.NewTextRange(leftParen, leftParen+1), prefix),
			rule.ReplaceRange(core.NewTextRange(rightParen, rightParen+1), "})"))
	} else {
		fixes = append(fixes,
			rule.ReplaceRange(core.NewTextRange(leftParen, leftParen+1), "{"),
			rule.ReplaceRange(core.NewTextRange(rightParen, rightParen+1), "}"))
	}

	for _, argument := range call.Arguments.Nodes {
		argumentFixes, ok := preferObjectSpreadArgumentFix(ctx, argument, leftParen, rightParen)
		if !ok {
			return nil, false
		}
		fixes = append(fixes, argumentFixes...)
	}

	return fixes, true
}

// preferObjectSpreadArgumentFix rewrites one argument of the call.
//
// An object literal loses its braces and the whitespace just inside them; anything else gains a
// `...` prefix, wrapped in parentheses when the argument binds looser than spread.
func preferObjectSpreadArgumentFix(ctx rule.Context, argument *ast.Node,
	leftParen int, rightParen int) ([]rule.Fix, bool) {

	source := ctx.SourceFile.Text()

	// The outermost parentheses wrapping this argument, if any. Upstream walks tokens outward while
	// they pair up and stops at the argument list's own parenthesis; our parser gives them to us as
	// nested KindParenthesizedExpression nodes instead, so the walk is down rather than out.
	outer := argument
	inner := argument
	var wrappers []*ast.Node
	for inner.Kind == ast.KindParenthesizedExpression {
		wrappers = append(wrappers, inner)
		next := inner.AsParenthesizedExpression().Expression
		if next == nil {
			return nil, false
		}
		inner = next
	}

	outerSpan := rule.TokenRange(ctx.SourceFile, outer)
	if outerSpan.Pos() <= leftParen || outerSpan.End() > rightParen {
		return nil, false
	}

	if inner.Kind == ast.KindObjectLiteralExpression {
		// Upstream collects the braces AND every wrapping parenthesis into one sorted list, takes
		// the OUTERMOST pair as `left`/`right`, removes the inner ones bare, and strips whitespace
		// around the outermost pair only. The order matters: stripping around the braces instead
		// eats the indentation that upstream keeps, which is a one-case difference in output and
		// exactly what the corpus's multi-line parenthesized argument pins.
		var fixes []rule.Fix
		innerSpan := rule.TokenRange(ctx.SourceFile, inner)

		openBrace := innerSpan.Pos()
		closeBrace := innerSpan.End() - 1
		if openBrace >= closeBrace || source[openBrace] != '{' || source[closeBrace] != '}' {
			return nil, false
		}

		// The outermost bracket pair is the first wrapper's parentheses when there is one, and the
		// braces otherwise. Every bracket strictly inside it is removed bare.
		outermostOpen, outermostClose := openBrace, closeBrace
		if len(wrappers) > 0 {
			outerWrapperSpan := rule.TokenRange(ctx.SourceFile, wrappers[0])
			outermostOpen = outerWrapperSpan.Pos()
			outermostClose = outerWrapperSpan.End() - 1
			// The braces are now inner brackets, removed bare.
			fixes = append(fixes,
				rule.RemoveRange(core.NewTextRange(openBrace, openBrace+1)),
				rule.RemoveRange(core.NewTextRange(closeBrace, closeBrace+1)))
			// So is every wrapper except the outermost.
			for _, wrapper := range wrappers[1:] {
				wrapperSpan := rule.TokenRange(ctx.SourceFile, wrapper)
				fixes = append(fixes,
					rule.RemoveRange(core.NewTextRange(wrapperSpan.Pos(), wrapperSpan.Pos()+1)),
					rule.RemoveRange(core.NewTextRange(wrapperSpan.End()-1, wrapperSpan.End())))
			}
		}
		openBrace, closeBrace = outermostOpen, outermostClose

		// Upstream removes the brace AND the whitespace just inside it, which is what turns
		// `{ a: 1 }` into `a: 1` rather than ` a: 1 `. It declines to eat leading whitespace when a
		// line comment precedes the brace, since doing so would pull the brace's line onto the
		// comment's and comment out the code.
		leftEnd := preferObjectSpreadEndWithSpaces(source, openBrace+1)
		rightStart := preferObjectSpreadStartWithSpaces(ctx, source, closeBrace)
		if rightStart < leftEnd {
			rightStart = leftEnd
		}
		fixes = append(fixes,
			rule.RemoveRange(core.NewTextRange(openBrace, leftEnd)),
			rule.RemoveRange(core.NewTextRange(rightStart, closeBrace+1)))

		// Remove the comma that separated this argument from the next, when keeping it would
		// double up. Two shapes need it and upstream names both: an EMPTY literal, which leaves
		// nothing before the comma at all and produces `({, ...foo})`, and a literal that already
		// ends in a trailing comma, which would produce `({a: 1,, ...foo})`.
		//
		// This is the branch the first draft of this fixer omitted, and it failed 43 of the 63
		// vectors with exactly that leading comma.
		properties := inner.AsObjectLiteralExpression().Properties
		emptyLiteral := properties == nil || len(properties.Nodes) == 0
		if emptyLiteral || preferObjectSpreadHasTrailingComma(source, innerSpan, properties) {
			if comma := preferObjectSpreadCommaAfter(source, outerSpan.End(), rightParen); comma >= 0 {
				fixes = append(fixes, rule.RemoveRange(core.NewTextRange(comma, comma+1)))
			}
		}
		return fixes, true
	}

	// A non-literal argument becomes a spread. Parenthesized when the argument binds looser than
	// the spread operator, which upstream limits to an assignment, an arrow and a conditional.
	if preferObjectSpreadArgumentNeedsParens(inner, len(wrappers) > 0) {
		return []rule.Fix{
			rule.ReplaceRange(core.NewTextRange(outerSpan.Pos(), outerSpan.Pos()), "...("),
			rule.ReplaceRange(core.NewTextRange(outerSpan.End(), outerSpan.End()), ")"),
		}, true
	}
	return []rule.Fix{
		rule.ReplaceRange(core.NewTextRange(outerSpan.Pos(), outerSpan.Pos()), "..."),
	}, true
}

// preferObjectSpreadNeedsParens says whether the resulting object literal must be wrapped.
//
// The default is to WRAP, and that direction is the safety property rather than a preference: an
// unwrapped `{ ... }` in expression position parses as a BLOCK, which either fails to parse or,
// worse, parses as a block that returns nothing. `() => Object.assign({}, foo)` becoming
// `() => { ...foo}` is the case that makes it concrete, and it is the same class of meaning-changing
// repair this tree has shipped twice before.
//
// Five parents are exempt because a brace there is unambiguously an object literal already. That
// list is upstream's verbatim; measured against the installed rule, a variable declarator, an array
// element, a return argument, a call argument and a property value all repair without parentheses,
// while an arrow body and a bare expression statement are wrapped.
func preferObjectSpreadNeedsParens(ctx rule.Context, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return true
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindArrayLiteralExpression,
		ast.KindReturnStatement, ast.KindCallExpression, ast.KindPropertyAssignment:
		return false
	case ast.KindBinaryExpression:
		// Upstream's AssignmentExpression arm: only the left side of an assignment needs the
		// wrap, and only when it is not already parenthesized. Any other binary operator falls
		// through to the default and is wrapped.
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil &&
			binary.OperatorToken.Kind == ast.KindEqualsToken {
			// Upstream's AssignmentExpression arm: `parent.left === node && !isParenthesised`.
			// The RIGHT side needs no wrap at all, since a brace there is unambiguously a literal.
			// Measured: `eventData = Object.assign({}, eventData, {a:1})` repairs to
			// `eventData = { ... }` with no parentheses.
			if binary.Left != node {
				return false
			}
			return !preferObjectSpreadIsParenthesised(node)
		}
		return !preferObjectSpreadIsParenthesised(node)
	}
	return !preferObjectSpreadIsParenthesised(node)
}

// preferObjectSpreadIsParenthesised says whether the call already sits inside parentheses.
//
// Our parser gives them as a wrapping node, where upstream asks its token stream, so this reads the
// parent kind rather than scanning text.
func preferObjectSpreadIsParenthesised(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression
}

// preferObjectSpreadNeedsPrecedingSemicolon reproduces upstream's `needsPrecedingSemicolon`,
// narrowed to what this rule can produce.
//
// The rewrite puts a `(` at the start of a statement, which continues the previous line rather than
// starting a new one when that line had no semicolon. Upstream inserts `;` in front for exactly
// that case, and its own corpus writes `const x = 1\nObject.assign({}, foo)` to pin it.
//
// Asked only when the call IS the whole expression statement, since that is the only position where
// a leading `(` can be misread.
func preferObjectSpreadNeedsPrecedingSemicolon(ctx rule.Context, node *ast.Node,
	nodeSpan core.TextRange) bool {
	_ = nodeSpan

	// Upstream asks `isStartOfExpressionStatement`, which is whether the node BEGINS an expression
	// statement rather than whether it is the whole of one. The difference is real and the corpus
	// pins it: `foo\nObject.assign({}, bar).doSomething()` and `foo\nObject.assign({}, bar), 2`
	// both need the semicolon, and in neither is the call the statement's own expression.
	//
	// Walk up while this node is the leftmost part of its parent, which is what "starts" means.
	statement := node
	for statement.Parent != nil && statement.Parent.Kind != ast.KindExpressionStatement {
		if !preferObjectSpreadStartsParent(ctx, statement) {
			return false
		}
		statement = statement.Parent
	}
	parent := statement.Parent
	if parent == nil || parent.Kind != ast.KindExpressionStatement {
		return false
	}
	source := ctx.SourceFile.Text()

	// Walk back over whitespace to the previous non-space character. A semicolon or an opening
	// brace already terminates the statement; anything else, and the automatic semicolon would not
	// have been inserted, so one is written.
	index := rule.TokenRange(ctx.SourceFile, parent).Pos() - 1
	for index >= 0 && preferObjectSpreadIsSpace(source[index]) {
		index--
	}
	if index < 0 {
		return false
	}
	switch source[index] {
	case ';', '{', '}':
		return false
	}
	return true
}

// preferObjectSpreadEndWithSpaces returns the offset past any whitespace starting at `start`.
func preferObjectSpreadEndWithSpaces(source string, start int) int {
	end := start
	for end < len(source) && preferObjectSpreadIsSpace(source[end]) {
		end++
	}
	return end
}

// preferObjectSpreadStartWithSpaces returns the offset of the whitespace run ending at `end`.
//
// Upstream refuses to walk back over whitespace when the token before is a LINE comment, because
// pulling the closing brace up onto that line would comment it out. Reproduced by scanning for a
// line comment in the whitespace being crossed rather than by asking a token stream.
func preferObjectSpreadStartWithSpaces(ctx rule.Context, source string, end int) int {
	start := end
	for start > 0 && preferObjectSpreadIsSpace(source[start-1]) {
		start--
	}
	// If crossing that whitespace would move the brace onto a line holding a `//` comment, do not
	// cross it. The whitespace run contains a newline exactly when such a comment could precede.
	for index := start; index < end; index++ {
		if source[index] == '\n' {
			if preferObjectSpreadLineBeforeHasLineComment(source, index) {
				return end
			}
			break
		}
	}
	return start
}

// preferObjectSpreadLineBeforeHasLineComment says whether the line ending at `newline` holds a
// `//` comment that is not inside a string.
//
// Deliberately simple: it scans the line for `//` outside quotes. A false positive here costs a
// less tidy repair rather than a wrong one, since the only consequence is keeping whitespace.
func preferObjectSpreadLineBeforeHasLineComment(source string, newline int) bool {
	lineStart := newline
	for lineStart > 0 && source[lineStart-1] != '\n' {
		lineStart--
	}
	var quote byte
	for index := lineStart; index+1 < newline; index++ {
		character := source[index]
		if quote != 0 {
			if character == '\\' {
				index++
				continue
			}
			if character == quote {
				quote = 0
			}
			continue
		}
		switch character {
		case '\'', '"', '`':
			quote = character
		case '/':
			if source[index+1] == '/' {
				return true
			}
		}
	}
	return false
}

// preferObjectSpreadArgumentNeedsParens is upstream's `argNeedsParens`.
//
// Three expression kinds bind looser than the spread operator, so spreading them without
// parentheses changes what is spread. Everything else is left bare, which is upstream's default.
func preferObjectSpreadArgumentNeedsParens(argument *ast.Node, alreadyParenthesised bool) bool {
	if alreadyParenthesised {
		return false
	}
	switch argument.Kind {
	case ast.KindArrowFunction, ast.KindConditionalExpression:
		return true
	case ast.KindBinaryExpression:
		binary := argument.AsBinaryExpression()
		return binary != nil && binary.OperatorToken != nil &&
			preferObjectSpreadIsAssignmentOperator(binary.OperatorToken.Kind)
	}
	return false
}

// preferObjectSpreadIsAssignmentOperator says whether an operator writes to its left operand.
func preferObjectSpreadIsAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken, ast.KindPlusEqualsToken, ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken, ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken, ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken, ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken, ast.KindBarEqualsToken, ast.KindCaretEqualsToken,
		ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// preferObjectSpreadIsSpace says whether a byte is whitespace.
func preferObjectSpreadIsSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}

// preferObjectSpreadHasTrailingComma says whether an object literal's last property is followed by
// a comma before the closing brace.
func preferObjectSpreadHasTrailingComma(source string, span core.TextRange,
	properties *ast.NodeList) bool {

	if properties == nil || len(properties.Nodes) == 0 {
		return false
	}
	last := properties.Nodes[len(properties.Nodes)-1]
	for index := last.End(); index < span.End()-1; index++ {
		if source[index] == ',' {
			return true
		}
		if !preferObjectSpreadIsSpace(source[index]) {
			return false
		}
	}
	return false
}

// preferObjectSpreadCommaAfter returns the offset of the comma separating this argument from the
// next, or -1 when this is the last argument.
//
// Bounded by the call's own closing parenthesis so a comma belonging to something else can never
// be reached.
func preferObjectSpreadCommaAfter(source string, from int, rightParen int) int {
	for index := from; index < rightParen; index++ {
		if source[index] == ',' {
			return index
		}
		if !preferObjectSpreadIsSpace(source[index]) {
			return -1
		}
	}
	return -1
}

// preferObjectSpreadStartsParent says whether a node begins its parent's source span.
//
// Compared on trimmed token ranges rather than raw positions, since a node's own Pos includes
// leading trivia and two nodes sharing a start would then disagree over a comment.
func preferObjectSpreadStartsParent(ctx rule.Context, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	return rule.TokenRange(ctx.SourceFile, parent).Pos() ==
		rule.TokenRange(ctx.SourceFile, node).Pos()
}

// preferObjectSpreadUnwrap removes the parenthesis nodes our parser keeps and upstream's folds away.
//
// A loop rather than one step, because `((x))` nests, and written out rather than through
// `ast.SkipParentheses`, which dereferences its argument and would panic on a recovered parse whose
// parenthesized expression has no inner expression.
func preferObjectSpreadUnwrap(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		inner := node.AsParenthesizedExpression().Expression
		if inner == nil {
			return node
		}
		node = inner
	}
	return node
}
