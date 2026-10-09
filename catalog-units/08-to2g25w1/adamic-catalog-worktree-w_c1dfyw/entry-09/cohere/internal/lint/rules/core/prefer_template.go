package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreferTemplateUnexpectedStringConcatenation = rule.Message{
	Id: "unexpectedStringConcatenation",
	Description: "This builds a string by concatenating a literal with something that is not one. " +
		"The `+` operator does double duty in JavaScript, so a reader has to work out from the " +
		"operands whether a line adds or joins, and a single non-string operand silently changes " +
		"the answer for the whole expression. A template literal says which one is meant in its " +
		"first character, keeps the literal text contiguous instead of split across quotes and " +
		"plus signs, and does not change meaning when an operand's type does.",
}

// PreferTemplate flags string concatenation that mixes a string literal with a non-literal.
//
//	valid:   var foo = 'bar';
//	valid:   var foo = 'bar' + 'baz';
//	valid:   var foo = `bar${baz}`;
//	valid:   var foo = 1 + 2;
//	invalid: var foo = 'bar' + baz;
//	invalid: var foo = bar + 'baz';
//	invalid: var foo = `bar` + baz;
//
// Ported from `prefer-template` in ESLint, read from the clone at `lib/rules/prefer-template.js`. No
// options (`schema: []`), one message, and `meta.fixable` is `"code"`.
//
// The whole 88-case corpus was extracted from upstream's own tester and replayed against the rule
// through the ESLint Linter API before any code was written. All 79 invalid cases carry an `output`.
//
// # The corpus needs espree, for the same reason `dot-notation`'s did
//
// Five of the reporting cases carry octal or non-octal decimal escapes -- `'\033'`, `'\8'`, `'\0\1'`
// -- and under `@typescript-eslint/parser` all five are FATAL PARSE ERRORS, which read exactly like
// clean verdicts. The oracle said "linted 83 of 88" and those five silences looked like five clean
// cases. Espree in script mode gives 88 of 88 with zero disagreements against the corpus.
//
// Our own parser reads all five, measured directly with a control, so the rule's verdict on them is
// real here rather than inherited. That matters because those five are exactly the cases upstream
// reports on and refuses to FIX; see below.
//
// # What the rule decides, which is narrower than "a concatenation exists"
//
// Three conditions, all required. There must be a string literal somewhere in the concatenation, a
// NON-string somewhere in it, and the report anchors on the TOP of the concatenation chain rather
// than on the operand that triggered it. So `'a' + 'b'` is clean (no non-string), `a + b` is clean
// (no string literal), and `'a' + b + 'c'` reports exactly once rather than twice.
//
// The once-per-chain behaviour is upstream's `done` map keyed on the top expression's start offset.
// Reproduced here as a set of already-reported nodes, because both a `Literal` and a
// `TemplateLiteral` operand can trigger the same chain and upstream visits both.
//
// # A TEMPLATE literal counts as a string literal for triggering, and that is easy to miss
//
// Upstream registers `Literal` and `TemplateLiteral` on the same handler, so a backtick template
// concatenated with an identifier reports. But `astUtils.isStringLiteral` -- which decides whether
// an operand counts as the string
// half -- accepts a template literal too. So a chain of nothing but templates and strings is clean,
// while one template plus an identifier reports. 14 corpus cases exercise this.
var PreferTemplate = rule.Rule{
	Name: "prefer-template",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Upstream's `done` map, keyed on the top expression. Both a string literal and a template
		// literal operand reach the same chain, so without this a two-literal concatenation reports
		// twice.
		reported := map[*ast.Node]bool{}

		check := func(node *ast.Node) {
			if !preferTemplateIsStringLiteral(node) {
				return
			}
			if node.Parent == nil || !preferTemplateIsConcatenation(node.Parent) {
				return
			}
			top := preferTemplateTopConcatenation(node.Parent)
			if reported[top] {
				return
			}
			reported[top] = true

			if !preferTemplateHasNonStringLiteral(top) {
				return
			}
			if fixes, repairable := preferTemplateFix(ctx, top); repairable {
				ctx.ReportNodeWithFixes(top, messagePreferTemplateUnexpectedStringConcatenation, fixes...)
				return
			}
			ctx.ReportNode(top, messagePreferTemplateUnexpectedStringConcatenation)
		}

		return rule.Listeners{
			ast.KindStringLiteral:                 check,
			ast.KindNoSubstitutionTemplateLiteral: check,
			ast.KindTemplateExpression:            check,
			ast.KindNumericLiteral:                check,
			ast.KindBigIntLiteral:                 check,
			ast.KindTrueKeyword:                   check,
			ast.KindFalseKeyword:                  check,
			ast.KindNullKeyword:                   check,
			ast.KindRegularExpressionLiteral:      check,
		}
	},
}

// preferTemplateIsConcatenation answers upstream's `isConcatenation`.
//
// A binary `+` and nothing else. Note this does NOT skip parentheses: `('a') + b` has a
// parenthesized operand rather than a parenthesized concatenation, and upstream's parser folds the
// parens away so its `node.parent` is the binary expression directly. Ours keeps the node, so a
// parenthesized concatenation is a different shape and is handled where the chain is walked.
func preferTemplateIsConcatenation(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	expression := node.AsBinaryExpression()
	return expression != nil && expression.OperatorToken != nil &&
		expression.OperatorToken.Kind == ast.KindPlusToken
}

// preferTemplateTopConcatenation answers upstream's `getTopConcatBinaryExpression`.
//
// Walks up while the parent is still a `+` concatenation, so the whole chain reports once at its
// root. Parentheses are skipped on the way up, because our parser keeps a node upstream's does not:
// `('a' + b) + c` is one chain to espree and two nested ones plus a paren here, and reporting the
// inner one would both point at the wrong span and report twice.
func preferTemplateTopConcatenation(node *ast.Node) *ast.Node {
	current := node
	for current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			// Only continue through a paren when what encloses it is itself a concatenation;
			// otherwise the paren is the boundary of the expression and the chain ends here.
			if parent.Parent != nil && preferTemplateIsConcatenation(parent.Parent) {
				current = parent.Parent
				continue
			}
			return current
		}
		if !preferTemplateIsConcatenation(parent) {
			return current
		}
		current = parent
	}
	return current
}

// preferTemplateIsStringLiteral answers upstream's `astUtils.isStringLiteral`.
//
// A string literal OR a template literal of either shape. The template half is what makes a
// backtick template concatenated with an identifier report, and it is the same predicate upstream
// uses on both sides of the question: it decides what counts as the string half AND, negated, what
// counts as the non-string half.
func preferTemplateIsStringLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression:
		return true
	}
	return false
}

// preferTemplateHasNonStringLiteral answers upstream's `hasNonStringLiteral`.
//
// Recurses through the concatenation and answers true if ANY operand is not a string or template
// literal. Upstream notes that `left` is normally deeper than `right`, so it tests `right` first;
// the order is an optimization rather than a behaviour and is kept for the same reason.
//
// A parenthesized operand is unwrapped, because our parser keeps a node upstream's folds away.
// Without it `'a' + (b)` would see a parenthesized expression, correctly conclude it is not a string
// literal, and report -- which happens to be the right answer here, but for the wrong reason, and
// the same omission makes `'a' + ('b')` report when it should be clean.
func preferTemplateHasNonStringLiteral(node *ast.Node) bool {
	unwrapped := preferTemplateSkipParentheses(node)
	if unwrapped == nil {
		return false
	}
	if preferTemplateIsConcatenation(unwrapped) {
		expression := unwrapped.AsBinaryExpression()
		return preferTemplateHasNonStringLiteral(expression.Right) ||
			preferTemplateHasNonStringLiteral(expression.Left)
	}
	return !preferTemplateIsStringLiteral(unwrapped)
}

// preferTemplateSkipParentheses unwraps parenthesized expressions.
//
// A loop rather than one step, because `((x))` nests. Not `ast.SkipParentheses`, which dereferences
// its argument: this is called on operands that a malformed parse can leave nil, and that helper is
// how this project lost 167 files to a nil panic.
func preferTemplateSkipParentheses(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Kind == ast.KindParenthesizedExpression {
		current = current.AsParenthesizedExpression().Expression
	}
	return current
}

// preferTemplateHasOctalOrNonOctalDecimalEscape answers upstream's
// `hasOctalOrNonOctalDecimalEscapeSequence`, which is what makes the fixer decline.
//
// Recurses the concatenation and tests each string literal's RAW text. Five corpus cases carry
// `output: null` for this reason, and the decline is the interesting half of the rule's repair
// story: a legacy octal escape means different things inside a template literal than inside a
// quoted string, so converting one would change what the program produces rather than how it is
// spelled.
//
// Reads the raw source rather than `Text()`, which returns the COOKED value with escapes already
// resolved -- by which point `\033` and the character it denotes are indistinguishable.
func preferTemplateHasOctalOrNonOctalDecimalEscape(ctx rule.Context, node *ast.Node) bool {
	unwrapped := preferTemplateSkipParentheses(node)
	if unwrapped == nil {
		return false
	}
	if preferTemplateIsConcatenation(unwrapped) {
		expression := unwrapped.AsBinaryExpression()
		return preferTemplateHasOctalOrNonOctalDecimalEscape(ctx, expression.Left) ||
			preferTemplateHasOctalOrNonOctalDecimalEscape(ctx, expression.Right)
	}
	if unwrapped.Kind != ast.KindStringLiteral || ctx.SourceFile == nil {
		return false
	}
	raw := preferTemplateRawText(ctx, unwrapped)
	return preferTemplateRawHasLegacyEscape(raw)
}

// preferTemplateRawHasLegacyEscape scans raw source for an octal or non-octal decimal escape.
//
// Upstream's `astUtils.hasOctalOrNonOctalDecimalEscapeSequence` is a regex over the raw text:
// a backslash followed by a digit, where an even number of preceding backslashes means the
// backslash is itself escaped and the digit is literal.
//
// `\0` alone is NOT a legacy escape -- it is the null character and is legal in a template -- but
// `\0` followed by a digit is. Upstream's pattern encodes that, and the corpus pins it: `'\0\1'`
// declines while a bare `'\0'` does not appear as a decline.
func preferTemplateRawHasLegacyEscape(raw string) bool {
	for index := 0; index < len(raw); index++ {
		if raw[index] != '\\' {
			continue
		}
		// Count the run of backslashes. An even-length run is escaped backslashes and the
		// character after it is literal text.
		start := index
		for index < len(raw) && raw[index] == '\\' {
			index++
		}
		if (index-start)%2 == 0 {
			index--
			continue
		}
		if index >= len(raw) {
			return false
		}
		digit := raw[index]
		if digit < '0' || digit > '9' {
			index--
			continue
		}
		if digit == '0' {
			// `\0` is the null escape and is fine on its own; `\0` followed by a digit is a
			// legacy octal.
			if index+1 < len(raw) && raw[index+1] >= '0' && raw[index+1] <= '9' {
				return true
			}
			index--
			continue
		}
		return true
	}
	return false
}

// preferTemplateRawText reads a node's source text, delimiters included.
//
// `Text()` on a string literal returns the cooked value, which is the wrong thing for every
// question about escapes: by the time it is cooked, `\033` and the character it denotes are the
// same string.
func preferTemplateRawText(ctx rule.Context, node *ast.Node) string {
	if ctx.SourceFile == nil {
		return ""
	}
	source := ctx.SourceFile.Text()
	start, end := node.Pos(), node.End()
	if start < 0 || end > len(source) || start >= end {
		return ""
	}
	return strings.TrimLeft(source[start:end], " \t\r\n")
}

// preferTemplateFix builds the repair that turns a concatenation into a template literal.
//
// Returns false where upstream returns null, which is a DECLINE rather than an absent repair: the
// finding is still reported and only the fix is withheld. Five corpus cases are `output: null`, all
// for the same reason -- a legacy octal or non-octal decimal escape means something different inside
// a template literal than inside a quoted string, so converting one would change what the program
// produces rather than how it is spelled.
//
// Upstream also emits a leading `;` when the expression starts a statement that needs one, guarding
// against automatic semicolon insertion joining it to the line above. No corpus case exercises that
// arm, and reproducing it needs `isStartOfExpressionStatement` plus `needsPrecedingSemicolon`, two
// helpers this tree does not have. It is a stated gap rather than a silent one, and it is in the
// safe direction: the repair is withheld, not written wrong. See preferTemplateNeedsSemicolonGuard.
func preferTemplateFix(ctx rule.Context, top *ast.Node) ([]rule.Fix, bool) {
	if ctx.SourceFile == nil {
		return nil, false
	}
	if preferTemplateHasOctalOrNonOctalDecimalEscape(ctx, top) {
		return nil, false
	}
	built, ok := preferTemplateLiteralFor(ctx, top, "", "")
	if !ok {
		return nil, false
	}

	// Upstream's `prefix`: a leading `;` when the repaired expression starts a statement whose
	// previous token could join with the new backtick under automatic semicolon insertion. Five
	// corpus cases pin it -- `foo\n'bar' + baz` becomes `foo\n;\`bar${  baz}\`` -- and each has a
	// different preceding shape: an identifier, a call, an index, a string, a template.
	if preferTemplateNeedsSemicolonGuard(ctx, top) {
		built = ";" + built
	}
	// `Pos()` is where a node's LEADING TRIVIA begins, not where its first token does, so replacing
	// from it swallows the space before the expression: `var foo = 'a' + b` repaired to
	// `var foo =` with the space eaten. Every one of the 74 repair fixtures caught this at once,
	// which is what a systematic off-by-trivia looks like.
	return []rule.Fix{
		rule.ReplaceRange(core.NewTextRange(preferTemplateTokenStart(ctx, top), top.End()), built),
	}, true
}

// preferTemplateNeedsSemicolonGuard answers whether the repair would need a leading `;`.
//
// Upstream emits one when the top expression starts an expression statement AND the previous token
// could join with a backtick under automatic semicolon insertion. This port declines the repair in
// that narrow position rather than emitting the semicolon, because getting it wrong either writes a
// syntax error or silently joins two statements, and reproducing it needs
// `isStartOfExpressionStatement` plus `needsPrecedingSemicolon`, neither of which this tree has.
//
// The first draft of this guard was "any concatenation that begins an expression statement", which
// DECLINED the repair rather than emitting the semicolon, on the reasoning that withholding is the
// safe direction. It is the safe direction and it was still wrong: it withheld 50 of the 74 repairs
// the corpus specifies, turning a working fixer into one that almost never fires.
//
// Worth recording how that draft was justified, because the justification was a bad measurement. It
// rested on "zero of the 74 outputs begin with a semicolon", counted with
// `output.lstrip().startswith(";")`. Five of them DO carry one -- the semicolon appears after a
// newline, in the middle of the string, where a check anchored at the start cannot see it. So the
// guard was defended by a measurement that asked a question shaped like the answer it got.
//
// Narrowed to the shape that actually needs it: an expression statement whose PREVIOUS
// non-whitespace character can begin an ASI hazard when a backtick follows. That is a closing
// bracket, paren, or an identifier or literal character. Those are the cases where a line ending in
// a value, followed by a line starting with a backtick, parses as a TAGGED TEMPLATE rather than as
// two statements. A statement already ending in a semicolon or a brace cannot be joined, which is
// the common case and the reason this guard almost never fires.
//
// Five corpus cases pin it, each with a different preceding shape: an identifier, a call, an index,
// a string literal, and a template literal.
func preferTemplateNeedsSemicolonGuard(ctx rule.Context, top *ast.Node) bool {
	parent := top.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindExpressionStatement {
		return false
	}
	if ctx.SourceFile == nil {
		return false
	}
	text := ctx.SourceFile.Text()
	position := preferTemplateTokenStart(ctx, top) - 1
	for position >= 0 {
		switch text[position] {
		case ' ', '\t', '\r', '\n':
			position--
			continue
		}
		break
	}
	if position < 0 {
		return false
	}
	// A statement that ends in `;` or `}` cannot be joined to what follows, which is the common
	// case and the reason this guard almost never fires.
	switch text[position] {
	case ';', '{', '}':
		return false
	}
	return true
}

// preferTemplateLiteralFor answers upstream's `getTemplateLiteral`.
//
// Recursive, and the recursion is the whole fixer. Four arms, in upstream's order:
//
//	a string literal      becomes template text, with `${` and backticks escaped
//	a template literal    is copied through verbatim
//	a concatenation       is joined, with the text around the `+` carried into a curly
//	anything else         becomes `${ ... }`
//
// `textBefore` and `textAfter` carry the source text that sat around a `+` sign -- comments, in
// practice -- so it can be placed inside a curly where it remains legal. Upstream's two corpus
// cases with comments are what pin this.
func preferTemplateLiteralFor(ctx rule.Context, node *ast.Node, textBefore string, textAfter string) (string, bool) {
	if node == nil {
		return "", false
	}
	// Every arm below asks what KIND the node is, and a parenthesized expression answers with the
	// paren rather than with what it wraps. Upstream never sees one, so unwrapping here is what
	// keeps `('baz')` taking the string-literal arm instead of the catch-all.
	if unwrapped := preferTemplateSkipParentheses(node); unwrapped != nil && unwrapped != node {
		return preferTemplateLiteralFor(ctx, unwrapped, textBefore, textAfter)
	}

	if node.Kind == ast.KindStringLiteral {
		// From the TOKEN start, not `Pos()`. A literal's `Pos()` includes its leading trivia, so a
		// `/* a */ 'bar'` would carry the comment into the template text -- and since the
		// replacement span also starts after that comment, the result duplicates it. Upstream reads
		// `currentNode.raw`, which is the literal alone.
		raw := preferTemplateRawTextFromToken(ctx, node)
		if len(raw) < 2 {
			return "", false
		}
		quote := raw[0]
		inner := raw[1 : len(raw)-1]
		return "`" + preferTemplateUnescapeQuote(preferTemplateEscapeForTemplate(inner), quote) + "`", true
	}

	if node.Kind == ast.KindNoSubstitutionTemplateLiteral || node.Kind == ast.KindTemplateExpression {
		return preferTemplateRawText(ctx, node), true
	}

	if preferTemplateIsConcatenation(node) && preferTemplateHasStringLiteral(node) {
		expression := node.AsBinaryExpression()
		// Our `OperatorToken` range is NOT upstream's `+` token. Measured: for `'a' + b` the token
		// spans `" +"` -- its `Pos()` includes the leading whitespace and its `End()` runs to the
		// right operand -- so the gaps either side come back empty and every `${  x  }` in the
		// corpus renders as `${x}`. Upstream's token is the bare `+` with the whitespace outside it.
		//
		// So the operator CHARACTER is located rather than the token trusted, and the two gaps are
		// measured against that. All 74 repair fixtures depend on it, which is what made the defect
		// systematic rather than incidental.
		plusOffset, located := preferTemplateOperatorOffset(ctx, expression)
		if !located {
			return "", false
		}
		textBeforePlus := preferTemplateTextBetween(ctx, expression.Left.End(), plusOffset)
		// The right operand's `Pos()` includes its own leading trivia, so measuring to it makes the
		// second gap empty and the rendered curly comes out `${ x }` where upstream writes
		// `${  x  }`. Upstream's `right.range[0]` is the TOKEN start, which is what this helper
		// answers. Same trivia distinction as the replacement span above, one level down.
		textAfterPlus := preferTemplateTextBetween(ctx, plusOffset+1,
			preferTemplateTokenStart(ctx, expression.Right))

		leftEndsWithCurly := preferTemplateEndsWithCurly(ctx, expression.Left)
		rightStartsWithCurly := preferTemplateStartsWithCurly(ctx, expression.Right)

		// `textBeforePlus` and `textAfterPlus` are the raw gaps around the operator, which for
		// ordinary code is the surrounding WHITESPACE rather than a comment. Upstream carries it
		// into the curly, which is why its outputs read `${  baz  }` with two spaces rather than
		// `${baz}`: one space came from each side of the `+`. Dropping it is not cosmetic, because
		// the corpus asserts the rewritten file byte for byte.
		if leftEndsWithCurly {
			// The left side already ends in a curly, so the text around the `+` goes inside it.
			left, leftOk := preferTemplateLiteralFor(ctx, expression.Left, textBefore, textBeforePlus+textAfterPlus)
			right, rightOk := preferTemplateLiteralFor(ctx, expression.Right, "", textAfter)
			if !leftOk || !rightOk || len(left) < 1 || len(right) < 1 {
				return "", false
			}
			return left[:len(left)-1] + right[1:], true
		}
		if rightStartsWithCurly {
			// Otherwise the right side opens with a curly and the text goes there.
			left, leftOk := preferTemplateLiteralFor(ctx, expression.Left, textBefore, "")
			right, rightOk := preferTemplateLiteralFor(ctx, expression.Right, textBeforePlus+textAfterPlus, textAfter)
			if !leftOk || !rightOk || len(left) < 1 || len(right) < 1 {
				return "", false
			}
			return left[:len(left)-1] + right[1:], true
		}

		// Neither side has a curly to hold the text, so the `+` survives and the two halves stay
		// separate template literals. Upstream does the same rather than dropping the text.
		left, leftOk := preferTemplateLiteralFor(ctx, expression.Left, textBefore, "")
		right, rightOk := preferTemplateLiteralFor(ctx, expression.Right, textAfter, "")
		if !leftOk || !rightOk {
			return "", false
		}
		return left + textBeforePlus + "+" + textAfterPlus + right, true
	}

	// The catch-all arm: anything that is not a literal becomes a curly. Upstream's parser folds
	// parentheses away before the rule runs, so its `getText` on `(n * 1000)` returns `n * 1000`.
	// Ours keeps the node, so the parens have to come off here or every parenthesized operand
	// renders one layer too deep.
	inner := preferTemplateSkipParentheses(node)
	if inner == nil {
		inner = node
	}
	// From the TOKEN start for the same reason the literal arm is: the operand's `Pos()` includes
	// the comment that sat after the `+`, and `textBefore` already carries that comment. Reading
	// from `Pos()` here emits it twice.
	return "`${" + textBefore + preferTemplateRawTextFromToken(ctx, inner) + textAfter + "}`", true
}

// preferTemplateEscapeForTemplate escapes what a template literal treats specially.
//
// Upstream's regex is `/(?<!\\)\\*(\$\{|`)/gu` with a parity test on the matched backslashes: an
// ODD-length run means the delimiter is already escaped and is left alone, an even-length run means
// it is live and a backslash is added. Go's RE2 has no lookbehind, so this is a scan, and the
// parity rule is reproduced directly rather than approximated.
//
// The four backslash-count corpus cases are what pin it, and they are the reason this cannot be a
// naive replace: `'0 backslashes: ${bar}'` gains one, `'1 backslash: \${bar}'` gains none, and the
// two- and three-backslash forms both end at three.
func preferTemplateEscapeForTemplate(inner string) string {
	var builder strings.Builder
	for index := 0; index < len(inner); {
		if inner[index] != '\\' {
			if inner[index] == '`' || (inner[index] == '$' && index+1 < len(inner) && inner[index+1] == '{') {
				builder.WriteByte('\\')
			}
			builder.WriteByte(inner[index])
			index++
			continue
		}
		// A run of backslashes, then whatever follows it.
		start := index
		for index < len(inner) && inner[index] == '\\' {
			index++
		}
		run := inner[start:index]
		builder.WriteString(run)
		if index >= len(inner) {
			break
		}
		isDelimiter := inner[index] == '`' ||
			(inner[index] == '$' && index+1 < len(inner) && inner[index+1] == '{')
		// An odd run escapes the delimiter already; an even one leaves it live.
		if isDelimiter && len(run)%2 == 0 {
			builder.WriteByte('\\')
		}
		builder.WriteByte(inner[index])
		index++
	}
	return builder.String()
}

// preferTemplateUnescapeQuote drops the escaping on the ORIGINAL quote character.
//
// Inside a template literal a `\'` or `\"` no longer needs its backslash, and upstream removes it so
// the repair does not leave pointless escapes behind. Upstream builds the pattern from
// `currentNode.raw[0]`, the quote the literal was actually written with, which is why the caller
// passes it rather than trying both.
func preferTemplateUnescapeQuote(inner string, quote byte) string {
	return strings.ReplaceAll(inner, "\\"+string(quote), string(quote))
}

// preferTemplateTextBetween returns the source text between two offsets.
//
// Upstream's `getTextBetween` walks the tokens and concatenates the gaps, which is how it keeps
// comments and drops the tokens themselves. Between an operand and its `+` there are no intervening
// tokens, so the gap IS the text and a slice answers the same question.
func preferTemplateTextBetween(ctx rule.Context, from int, to int) string {
	if ctx.SourceFile == nil || from < 0 || to > len(ctx.SourceFile.Text()) || from >= to {
		return ""
	}
	return ctx.SourceFile.Text()[from:to]
}

// preferTemplateHasStringLiteral answers upstream's `hasStringLiteral`.
func preferTemplateHasStringLiteral(node *ast.Node) bool {
	unwrapped := preferTemplateSkipParentheses(node)
	if unwrapped == nil {
		return false
	}
	if preferTemplateIsConcatenation(unwrapped) {
		expression := unwrapped.AsBinaryExpression()
		return preferTemplateHasStringLiteral(expression.Right) ||
			preferTemplateHasStringLiteral(expression.Left)
	}
	return preferTemplateIsStringLiteral(unwrapped)
}

// preferTemplateStartsWithCurly answers upstream's `startsWithTemplateCurly`.
//
// Whether the node, once converted, begins with a `${`. A concatenation defers to its left side; a
// template literal starts with a curly when its first quasi is EMPTY, which is upstream's
// `quasis[0].range[0] === quasis[0].range[1]`; anything that is not a string literal becomes a curly
// outright.
func preferTemplateStartsWithCurly(ctx rule.Context, node *ast.Node) bool {
	unwrapped := preferTemplateSkipParentheses(node)
	if unwrapped == nil {
		return false
	}
	if preferTemplateIsConcatenation(unwrapped) {
		return preferTemplateStartsWithCurly(ctx, unwrapped.AsBinaryExpression().Left)
	}
	if unwrapped.Kind == ast.KindTemplateExpression {
		raw := preferTemplateRawText(ctx, unwrapped)
		// The head runs from the opening backtick to the first `${`; an empty head means the
		// literal opens with a curly.
		return strings.HasPrefix(raw, "`${")
	}
	if unwrapped.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return false
	}
	return unwrapped.Kind != ast.KindStringLiteral
}

// preferTemplateEndsWithCurly answers upstream's `endsWithTemplateCurly`.
//
// Note upstream's own asymmetry, reproduced here: for a concatenation it calls
// `startsWithTemplateCurly(node.right)` rather than `endsWith`. That reads like a slip and is what
// the corpus asserts, so it is copied rather than corrected.
func preferTemplateEndsWithCurly(ctx rule.Context, node *ast.Node) bool {
	unwrapped := preferTemplateSkipParentheses(node)
	if unwrapped == nil {
		return false
	}
	if preferTemplateIsConcatenation(unwrapped) {
		return preferTemplateStartsWithCurly(ctx, unwrapped.AsBinaryExpression().Right)
	}
	if unwrapped.Kind == ast.KindTemplateExpression {
		raw := preferTemplateRawText(ctx, unwrapped)
		return strings.HasSuffix(raw, "}`")
	}
	if unwrapped.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return false
	}
	return unwrapped.Kind != ast.KindStringLiteral
}

// preferTemplateTokenStart returns where a node's first TOKEN begins, skipping leading trivia.
//
// `Pos()` includes the whitespace and comments before a node, which is the right position for a
// comment scanner and the wrong one for a replacement span: replacing from it deletes the space
// that separated the expression from whatever preceded it.
//
// Walks forward from `Pos()` over whitespace AND comments, because both are leading trivia and the
// replacement must begin at the first real token.
//
// Skipping comments is not optional, and getting it wrong is subtle rather than loud. A scan that
// stops at the first non-whitespace byte stops on the `/` of a leading `/* a */`, so the replacement
// starts one byte into the comment and the repaired source reads `\`* a */ 'bar...` -- a template
// literal that swallowed most of a comment and is still valid syntax. Upstream's own corpus case
// with a leading comment is what caught it, and it was the last of the 74 to pass.
func preferTemplateTokenStart(ctx rule.Context, node *ast.Node) int {
	if ctx.SourceFile == nil {
		return node.Pos()
	}
	text := ctx.SourceFile.Text()
	position := node.Pos()
	for position < len(text) && position < node.End() {
		switch text[position] {
		case ' ', '\t', '\r', '\n':
			position++
			continue
		}
		if text[position] == '/' && position+1 < node.End() {
			if text[position+1] == '/' {
				for position < node.End() && text[position] != '\n' {
					position++
				}
				continue
			}
			if text[position+1] == '*' {
				position += 2
				for position+1 < node.End() && !(text[position] == '*' && text[position+1] == '/') {
					position++
				}
				position += 2
				continue
			}
		}
		break
	}
	return position
}

// preferTemplateOperatorOffset finds the byte offset of the `+` character joining two operands.
//
// Needed because `OperatorToken.Pos()` here is not where the operator is written: it includes the
// leading trivia, so the token spans `" +"` rather than `"+"`, and both gaps upstream measures come
// back empty. Scanning between the operands for the character is exact, because between a left
// operand's end and a right operand's start the only non-trivia byte is the operator itself.
//
// Comments in that gap are skipped rather than searched, since a `+` inside a comment would
// otherwise be mistaken for the operator.
func preferTemplateOperatorOffset(ctx rule.Context, expression *ast.BinaryExpression) (int, bool) {
	if ctx.SourceFile == nil || expression.Left == nil || expression.Right == nil {
		return 0, false
	}
	text := ctx.SourceFile.Text()
	position := expression.Left.End()
	limit := expression.Right.Pos()
	if position < 0 || limit > len(text) {
		return 0, false
	}
	for position < limit {
		if text[position] == '/' && position+1 < limit {
			if text[position+1] == '/' {
				for position < limit && text[position] != '\n' {
					position++
				}
				continue
			}
			if text[position+1] == '*' {
				position += 2
				for position+1 < limit && !(text[position] == '*' && text[position+1] == '/') {
					position++
				}
				position += 2
				continue
			}
		}
		if text[position] == '+' {
			return position, true
		}
		position++
	}
	return 0, false
}

// preferTemplateRawTextFromToken reads a node's source text starting at its first token.
//
// `preferTemplateRawText` trims leading whitespace, which is enough for a node with no comment
// before it and wrong for one with a comment: trimming whitespace leaves the comment in place. This
// uses the same trivia skip the replacement span uses, so the two agree about where the node begins
// and neither drops nor duplicates what sits before it.
func preferTemplateRawTextFromToken(ctx rule.Context, node *ast.Node) string {
	if ctx.SourceFile == nil {
		return ""
	}
	text := ctx.SourceFile.Text()
	start := preferTemplateTokenStart(ctx, node)
	end := node.End()
	if start < 0 || end > len(text) || start >= end {
		return ""
	}
	return text[start:end]
}
