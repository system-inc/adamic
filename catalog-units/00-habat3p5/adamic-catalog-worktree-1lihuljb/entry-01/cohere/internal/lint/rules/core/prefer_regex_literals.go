package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexpattern"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferRegexLiteralsOptions is the rule's option surface.
type PreferRegexLiteralsOptions struct {
	// DisallowRedundantWrapping additionally reports a regex literal handed straight to the
	// constructor, `new RegExp(/a/)`, which upstream treats as a separate judgment because the
	// author already wrote a literal and wrapping it back up is a different mistake from never
	// having written one.
	//
	// Defaults to false, which is upstream's default, so the zero value is the configured default
	// and `rule.DecodeOptionsInto` is safe here. A default-true option would not be: the brief's
	// decoder trap is that an absent key and an explicit false are indistinguishable in a plain
	// struct, and it does not bite when the default already is the zero value.
	DisallowRedundantWrapping bool `json:"disallowRedundantWrapping"`
}

var messageUnexpectedRegExp = rule.Message{
	Id: "unexpectedRegExp",
	Description: "This builds a regular expression by handing constant strings to the `RegExp` " +
		"constructor, which does at runtime what a literal does at parse time. The constructor " +
		"form costs a parse on every evaluation, hides syntax errors until the line is reached " +
		"rather than surfacing them when the file is read, and doubles every backslash so the " +
		"pattern stops looking like the pattern. Write the regular expression literal.",
}

var messageUnexpectedRedundantRegExp = rule.Message{
	Id: "unexpectedRedundantRegExp",
	Description: "This wraps a regular expression literal in the `RegExp` constructor, which " +
		"builds a second identical expression from the first and then discards the original. The " +
		"literal already is the value being asked for, so the constructor adds a copy and a " +
		"runtime parse and changes nothing else.",
}

var messageUnexpectedRedundantRegExpWithFlags = rule.Message{
	Id: "unexpectedRedundantRegExpWithFlags",
	Description: "This hands a regular expression literal to the `RegExp` constructor along with " +
		"a flags argument, and the flags argument wins outright: the flags written on the literal " +
		"are discarded rather than combined. That is the trap, because the literal's own flags sit " +
		"right there in the source looking as though they still apply. Write one literal carrying " +
		"the flags it should have.",
}

var messageReplaceWithLiteral = rule.Message{
	Id:          "replaceWithLiteral",
	Description: "Replace with an equivalent regular expression literal.",
}

// replaceWithLiteralAndFlags and replaceWithIntendedLiteralAndFlags are the two readings of a
// wrapped literal carrying its own flags, and offering both is upstream's point rather than
// indecision. `new RegExp(/a/i, 'g')` evaluates to `/a/g` today, so the equivalent literal drops the
// `i`; but an author who wrote `i` on the literal probably meant to keep it, so the second
// suggestion offers `/a/ig`. Only a human knows which was intended, which is exactly why these are
// suggestions and not fixes.
func messageReplaceWithLiteralAndFlags(flags string) rule.Message {
	return rule.Message{
		Id: "replaceWithLiteralAndFlags",
		Description: fmt.Sprintf(
			"Replace with an equivalent regular expression literal with flags '%s'.", flags),
	}
}

func messageReplaceWithIntendedLiteralAndFlags(flags string) rule.Message {
	return rule.Message{
		Id: "replaceWithIntendedLiteralAndFlags",
		Description: fmt.Sprintf(
			"Replace with a regular expression literal with flags '%s'.", flags),
	}
}

// PreferRegexLiterals flags building a regular expression from constant strings.
//
//	valid:   /abc/
//	valid:   new RegExp(pattern)
//	valid:   new RegExp('a' + suffix)
//	valid:   function f(RegExp) { return new RegExp('a'); }
//	invalid: new RegExp('abc');
//	invalid: RegExp('abc', 'g');
//	invalid: new RegExp(String.raw`\d`, 'g');
//	invalid: new RegExp(/a/, 'g');            (only with disallowRedundantWrapping)
//
// # What counts as constant, and why the list is exactly three shapes
//
// Upstream accepts a string literal, a template literal with no substitutions, and a
// String.raw tagged template with no substitutions. Nothing else, and in particular not
// `'a' + 'b'`: the rule declines to constant-fold, so a concatenation of two literals is
// treated as dynamic even though a reader can see the value. Eleven of upstream's seventy
// clean cases are concatenations, which is the corpus stating that decision loudly enough
// that reproducing it is not a judgment call.
//
// The String.raw arm is the one that carries real weight, because it is the only shape whose
// value differs from what `Text()` answers. `String.raw` + backtick + `\d` reads as a
// two-character pattern, while the same characters in a plain template literal cook to a
// single `\d` control character. The rule takes the RAW text there and the COOKED text
// everywhere else, and getting that backwards silently changes every backslash in the output.
// Reached by slicing the source with the token range and stripping the backticks, since
// `Text()` on a template literal is always cooked.
//
// # The tag has to be the real String.raw
//
// `String.raw` and `String['raw']` both qualify, and both are in the corpus. A local `String`
// does not: upstream tests the receiver against the global, so `function f(String) { new
// RegExp(String.raw` + "`a`" + `) }` is clean. That is the same global-resolution question the
// callee asks, answered the same way.
//
// # Why the checker
//
// The whole discrimination between a finding and a false positive on
// `function f(RegExp) { return new RegExp('a'); }` is whether `RegExp` names the global or a
// local shadowing it, and nothing structural answers that. Upstream uses `ReferenceTracker`
// over its scope analysis; `resolvesToAGlobal` asks the checker which file declares the name,
// which is the same question. See its doc comment in `no_new_native_nonconstructor.go`.
//
// # The suggestions are suggestions, and one of them is deliberately withheld
//
// Upstream declares `hasSuggestions` and ships no fixer, so nothing here is applied unattended.
// That is the right line: rewriting `new RegExp(x)` to `/x/` changes a runtime-parsed value
// into a parse-time one, and where the string is not a valid pattern the rewrite turns working
// code that throws at a known point into code that does not parse at all.
//
// Upstream withholds the suggestion in four situations and each is reproduced:
//
//   - A comment inside the call, since the replacement text would delete it.
//   - A preceding token that cannot sit against a `/`, since `x = y` on the line above turns
//     the new literal into a division. `validPrecedingTokens` is upstream's allowlist and it is
//     copied rather than re-derived, because it is a list of judgments and not a rule.
//   - A pattern the regex grammar rejects, such as a bare `+`.
//   - A pattern holding a character outside upstream's printable allowlist, which is a
//     deliberately conservative gate on what the rewriter is willing to re-spell.
//
// # One divergence, measured, and it is about ecmaVersion
//
// Upstream withholds a suggestion when the pattern or flags are invalid AT THE CONFIGURED
// ecmaVersion, so `new RegExp('abc', 'd')` reports with no suggestion under ecmaVersion 2021
// and with one under 2022. cohere has no per-file ecmaVersion: the program is compiled once
// against one set of compiler options and a rule cannot ask what language edition a file
// claims. So validity is judged at the latest edition, which means this rule offers a
// suggestion in five of upstream's own cases where upstream withholds one.
//
// Measured against eslint 10.8.1 over the whole imported corpus: exactly five invalid cases
// and three valid ones turn on ecmaVersion, out of 251. All eight are marked in the fixture
// tables with the version that decides them. The direction is worth stating: offering a
// suggestion is the safe side of this divergence, because a suggestion is never applied
// without a human, and the code it proposes is valid in every edition this codebase targets.
var PreferRegexLiterals = rule.Rule{
	Name: "prefer-regex-literals",

	// The calls come from the shelf's ReferenceTracker, which tells the global RegExp from a local one
	// by where its symbol is declared: eleven of upstream's clean cases are exactly that shadow.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Declining once per file rather than once per node, following `await_thenable`: the
		// registration path will not hand a typed rule a file with no checker, but a Context built
		// by hand can, and `TestNoRegisteredRuleCrashesOnAbsentOptionalNodes` builds one.
		if ctx.TypeChecker == nil {
			return nil
		}

		settings, _ := rule.OptionsAs[PreferRegexLiteralsOptions](options)

		check := func(node *ast.Node) {
			argumentNodes := node.Arguments()

			if settings.DisallowRedundantWrapping && wrapsARegexLiteral(ctx, argumentNodes) {
				reportRedundantWrapping(ctx, node, argumentNodes)
				return
			}
			if !hasOnlyStaticStringArguments(ctx, argumentNodes) {
				return
			}
			reportConstructedFromStrings(ctx, node, argumentNodes)
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// Every trace starts at a reference to RegExp, so a file that never spells it has
				// nothing to follow, and the tracker's index is never built.
				if !strings.Contains(ctx.SourceFile.Text(), "RegExp") {
					return
				}
				tracker := reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)
				for _, tracked := range tracker.GlobalReferences(regExpCallTraceMap) {
					check(tracked.Node)
				}
			},
		}
	},
}

// staticStringValue answers the pattern or flags a node contributes, and whether it contributes one
// at all.
//
// The three accepted shapes are upstream's, and the second return distinguishes "this is a constant
// empty string" from "this is not constant", which matters because `new RegExp(”)` reports and
// suggests `/(?:)/` while `new RegExp(x)` is clean.
func staticStringValue(ctx rule.Context, node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		// Cooked, which is what both of these evaluate to. A template literal WITH substitutions is
		// a different kind and never arrives here, which is upstream's `isStaticTemplateLiteral`
		// answered by the parser instead of by a quasi count.
		return node.Text(), true
	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		if tagged.Template == nil || tagged.Template.Kind != ast.KindNoSubstitutionTemplateLiteral {
			return "", false
		}
		if !isStringRawTag(ctx, tagged.Tag) {
			return "", false
		}
		return rawTemplateText(ctx, tagged.Template), true
	}
	return "", false
}

// isStringRawTag reports whether a tag is `String.raw` or `String['raw']` on the global `String`.
//
// The receiver must resolve to the global, which is upstream's `isGlobalReference` check and is
// load-bearing rather than decorative: the corpus carries a clean case where a local `String`
// shadows it, and without this test that case reports.
func isStringRawTag(ctx rule.Context, tag *ast.Node) bool {
	if tag == nil {
		return false
	}
	// Upstream runs the tag through `skipChainExpression` before testing it, which in ESTree strips
	// the wrapper an optional chain puts around the whole expression. Our parser has no such
	// wrapper, but it DOES keep parentheses as real nodes where upstream's has already folded them
	// away, so `(String?.raw)` arrives here as a parenthesized expression and would otherwise be
	// declined. Unwrapped in a loop because `((String.raw))` nests.
	//
	// Not `ast.SkipParentheses`: it dereferences its argument, and this one is optional.
	for tag != nil && tag.Kind == ast.KindParenthesizedExpression {
		tag = tag.AsParenthesizedExpression().Expression
	}
	if tag == nil {
		return false
	}
	var receiver *ast.Node
	switch tag.Kind {
	case ast.KindPropertyAccessExpression:
		access := tag.AsPropertyAccessExpression()
		if access.Name() == nil || access.Name().Text() != "raw" {
			return false
		}
		receiver = access.Expression
	case ast.KindElementAccessExpression:
		access := tag.AsElementAccessExpression()
		argument := access.ArgumentExpression
		if argument == nil {
			return false
		}
		// A computed member counts only when the key is itself a constant string, matching
		// upstream's `isSpecificMemberAccess`, which reads a literal key and declines anything else.
		switch argument.Kind {
		case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
			if argument.Text() != "raw" {
				return false
			}
		default:
			return false
		}
		receiver = access.Expression
	default:
		return false
	}
	if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != "String" {
		return false
	}
	return resolvesToAGlobal(ctx, receiver)
}

// rawTemplateText returns a no-substitution template literal's text exactly as written.
//
// `Text()` is cooked, so it cannot serve here: `String.raw` + backtick + `\d` cooks to one control
// character and the rule needs the two characters the author typed. The raw form is the source
// between the backticks, which the token range locates.
func rawTemplateText(ctx rule.Context, template *ast.Node) string {
	if ctx.SourceFile == nil {
		return template.Text()
	}
	span := rule.TokenRange(ctx.SourceFile, template)
	text := ctx.SourceFile.Text()
	if span.Pos() < 0 || span.End() > len(text) || span.Pos() >= span.End() {
		return template.Text()
	}
	written := text[span.Pos():span.End()]
	written = strings.TrimPrefix(written, "`")
	written = strings.TrimSuffix(written, "`")
	return written
}

// hasOnlyStaticStringArguments reports whether every argument is a constant string, at an arity the
// constructor accepts.
//
// The arity test is upstream's and it is why `new RegExp()` with no arguments is clean: zero
// arguments builds `/(?:)/` and the rule says nothing about it, presumably because nobody writes
// that by accident.
func hasOnlyStaticStringArguments(ctx rule.Context, arguments []*ast.Node) bool {
	if len(arguments) != 1 && len(arguments) != 2 {
		return false
	}
	for _, argument := range arguments {
		if _, ok := staticStringValue(ctx, argument); !ok {
			return false
		}
	}
	return true
}

// wrapsARegexLiteral reports the `disallowRedundantWrapping` shape: a regex literal as the first
// argument, optionally followed by a constant string of flags.
func wrapsARegexLiteral(ctx rule.Context, arguments []*ast.Node) bool {
	if len(arguments) == 0 || arguments[0].Kind != ast.KindRegularExpressionLiteral {
		return false
	}
	if len(arguments) == 1 {
		return true
	}
	if len(arguments) != 2 {
		return false
	}
	_, ok := staticStringValue(ctx, arguments[1])
	return ok
}

func reportRedundantWrapping(ctx rule.Context, node *ast.Node, arguments []*ast.Node) {
	literal := arguments[0]
	literalPattern, literalFlags := regexsyntax.PatternAndFlags(literal.Text())

	if len(arguments) == 1 {
		var suggestions []rule.Suggestion
		// The replacement is the literal's own source text rather than a rebuilt
		// pattern-and-flags, matching upstream, which reaches for `getText` here and constructs
		// the string only on the two-argument arm. The difference is visible: a literal written
		// with a redundant escape keeps it.
		if canSuggestLiteral(ctx, node, literalPattern, literalFlags) {
			suggestions = append(suggestions, rule.Suggestion{
				Message: messageReplaceWithLiteral,
				Fixes:   []rule.Fix{replacementFix(ctx, node, literal.Text())},
			})
		}
		ctx.ReportNodeWithSuggestions(node, messageUnexpectedRedundantRegExp, suggestions...)
		return
	}

	argumentFlags, _ := staticStringValue(ctx, arguments[1])

	var suggestions []rule.Suggestion
	if canSuggestLiteral(ctx, node, literalPattern, argumentFlags) {
		suggestions = append(suggestions, rule.Suggestion{
			Message: messageReplaceWithLiteralAndFlags(argumentFlags),
			Fixes: []rule.Fix{replacementFix(ctx, node,
				"/"+literalPattern+"/"+argumentFlags)},
		})
	}
	merged := mergeRegexFlags(literalFlags, argumentFlags)
	if !flagsAreEqual(merged, argumentFlags) && canSuggestLiteral(ctx, node, literalPattern, merged) {
		suggestions = append(suggestions, rule.Suggestion{
			Message: messageReplaceWithIntendedLiteralAndFlags(merged),
			Fixes:   []rule.Fix{replacementFix(ctx, node, "/"+literalPattern+"/"+merged)},
		})
	}
	ctx.ReportNodeWithSuggestions(node, messageUnexpectedRedundantRegExpWithFlags, suggestions...)
}

func reportConstructedFromStrings(ctx rule.Context, node *ast.Node, arguments []*ast.Node) {
	pattern, _ := staticStringValue(ctx, arguments[0])
	flags := ""
	if len(arguments) == 2 {
		flags, _ = staticStringValue(ctx, arguments[1])
	}

	withhold := !canSuggestLiteral(ctx, node, pattern, flags) || !isRewritablePattern(pattern)

	var suggestions []rule.Suggestion
	if !withhold {
		rewritten := pattern
		if rewritten != "" {
			rewritten = escapeForLiteral(rewritten, flags)
		}
		if rewritten == "" {
			rewritten = "(?:)"
		}
		suggestions = append(suggestions, rule.Suggestion{
			Message: messageReplaceWithLiteral,
			Fixes:   []rule.Fix{replacementFix(ctx, node, "/"+rewritten+"/"+flags)},
		})
	}
	ctx.ReportNodeWithSuggestions(node, messageUnexpectedRegExp, suggestions...)
}

// replacementFix builds the edit replacing the whole call with a literal.
//
// `rule.TokenRange` rather than `node.Pos()`, because a node's Pos includes its leading trivia and a
// fix built from it eats the whitespace and comments before the call.
func replacementFix(ctx rule.Context, node *ast.Node, text string) rule.Fix {
	span := rule.TokenRange(ctx.SourceFile, node)
	return rule.ReplaceRange(span, padSoTokensStayApart(ctx, span, text))
}

// padSoTokensStayApart adds a space on either side of the replacement where removing one would
// weld the new literal onto its neighbour.
//
// This is upstream's `getSafeOutput`, which asks `canTokensBeAdjacent` about the token before and
// the token after and inserts a space where the answer is no. Its general form tokenizes both
// sides; the rule only ever writes a regex literal, so exactly two collisions are reachable and the
// corpus names both:
//
//	a/RegExp(/foo/);      ->  a/ /foo/;    a `/` directly before would start a comment
//	RegExp(/foo/)in a;    ->  /foo/ in a;  a word directly after would run into the flags
//
// Upstream also gates on the neighbour actually TOUCHING the node, `tokenBefore.range[1] ===
// node.range[0]`, so a space already in the source is not doubled. That gate is reproduced by
// looking at the byte immediately outside the span rather than at the previous token.
func padSoTokensStayApart(ctx rule.Context, span core.TextRange, text string) string {
	if ctx.SourceFile == nil || text == "" {
		return text
	}
	source := ctx.SourceFile.Text()

	// No hashbang arm here, deliberately. Upstream pads only when the neighbour actually TOUCHES
	// the node, `tokenBefore.range[1] === node.range[0]`, and a hashbang is a whole line above the
	// call, so the newline already separates them. Only the ALLOWLIST gate in
	// `precedingTokenAllowsALiteral` needs to know a hashbang is one token; padding does not, and
	// adding it there produced a leading space upstream's expected output does not carry.
	if span.Pos() > 0 && span.Pos() <= len(source) && text[0] == '/' {
		// A `/` before a `/` opens a line comment, and a `*` before one closes a block comment
		// that was never opened. Both are upstream's `canTokensBeAdjacent` answering no for a
		// left `/` against a regular expression.
		if previous := source[span.Pos()-1]; previous == '/' {
			text = " " + text
		}
	}
	if span.End() >= 0 && span.End() < len(source) {
		// A word character straight after the closing `/` reads as another flag, so `/foo/in a`
		// tokenizes as one literal with flags `in`. Upstream reaches the same answer by asking
		// whether two identifier-ish tokens can sit together.
		if next := source[span.End()]; isWordByteForRegexAdjacency(next) {
			text = text + " "
		}
	}
	return text
}

// sourceStartsWithHashbangBefore reports whether everything before offset is a single hashbang line.
//
// Only the FIRST line can be a hashbang, and only when the file opens with `#!`, so this cannot
// match a `#!` appearing anywhere else.
func sourceStartsWithHashbangBefore(source string, offset int) bool {
	if offset <= 0 || offset > len(source) || !strings.HasPrefix(source, "#!") {
		return false
	}
	firstNewline := strings.IndexByte(source, '\n')
	if firstNewline < 0 || offset <= firstNewline {
		return false
	}
	return text.TrimWhitespace(source[firstNewline+1:offset]) == ""
}

func isWordByteForRegexAdjacency(b byte) bool {
	return b == '_' || b == '$' || (b >= '0' && b <= '9') ||
		(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func mergeRegexFlags(first string, second string) string {
	var merged []rune
	seen := map[rune]bool{}
	for _, flag := range first + second {
		if seen[flag] {
			continue
		}
		seen[flag] = true
		merged = append(merged, flag)
	}
	return string(merged)
}

func flagsAreEqual(first string, second string) bool {
	return sortedFlags(first) == sortedFlags(second)
}

func sortedFlags(flags string) string {
	runes := []rune(flags)
	sort.Slice(runes, func(a int, b int) bool { return runes[a] < runes[b] })
	return string(runes)
}

// isRewritablePattern is upstream's printable allowlist, and it is a gate on the REWRITER rather
// than on the pattern's validity.
//
// Upstream spells it as a regular expression over the pattern text, and its effect is that anything
// outside a hand-picked printable set withholds the suggestion. That withholds a suggestion for
// a pattern holding a non-ASCII letter, whose pattern is perfectly valid: the rule declines to decide how a
// non-ASCII or control character should be spelled inside a literal, which is a conservative call
// about output rather than a judgment about input. Reproduced as a character test rather than as a
// pattern, because the set is fixed.
func isRewritablePattern(pattern string) bool {
	for _, character := range pattern {
		if !isRewritableCharacter(character) {
			return false
		}
	}
	return true
}

func isRewritableCharacter(character rune) bool {
	switch character {
	case ' ', '\t', '\r', '\n', '\v', '\f':
		return true
	case '-', '\\', '[', ']', '(', ')', '{', '}',
		'!', '@', '#', '$', '%', '^', '&', '*', '+', '=', '/', '~', '`',
		'.', '>', '<', '?', ',', '\'', '"', '|', ':', ';', '_':
		return true
	}
	if character >= 'a' && character <= 'z' {
		return true
	}
	if character >= 'A' && character <= 'Z' {
		return true
	}
	if character >= '0' && character <= '9' {
		return true
	}
	return false
}

// escapeForLiteral re-spells the characters that mean something different inside a literal than
// inside a string.
//
// A real newline in a string argument is a newline; the same byte inside a regex literal ends the
// literal, so it has to become the two characters backslash-n. Same for the other line and space
// controls, and for `/`, which would otherwise close the literal early.
//
// Upstream walks the parsed regex and rewrites each CHARACTER node, which matters for a reason that
// is not obvious: it means a control character reached through an identity escape, a backslash
// directly before a real newline, collapses to the same two characters rather than becoming a
// backslash followed by an escape. The walk here is `regexpattern.Walk`, which reports the same
// character positions with the same structure, so the rewrite is the same rewrite over the same
// units.
func escapeForLiteral(pattern string, flags string) string {
	type rewrite struct {
		start int
		end   int
		text  string
	}
	var rewrites []rewrite

	regexpattern.Walk(pattern, regexsyntax.ParseRegexFlags(flags), func(character regexpattern.Character) bool {
		if character.Start < 0 || character.End > len(pattern) || character.Start >= character.End {
			return true
		}
		written := pattern[character.Start:character.End]
		escaped, ok := escapeSpelling(written)
		if !ok {
			return true
		}
		rewrites = append(rewrites, rewrite{start: character.Start, end: character.End, text: escaped})
		return true
	})

	if len(rewrites) == 0 {
		return pattern
	}
	var rebuilt strings.Builder
	previous := 0
	for _, entry := range rewrites {
		if entry.start < previous {
			continue
		}
		rebuilt.WriteString(pattern[previous:entry.start])
		rebuilt.WriteString(entry.text)
		previous = entry.end
	}
	rebuilt.WriteString(pattern[previous:])
	return rebuilt.String()
}

// escapeSpelling maps a character as written to its literal-safe spelling, or declines.
//
// The two-character forms are upstream's, and they are the identity-escape shapes: a backslash
// directly before a real control character. Both spellings collapse to the same escape, which is
// why they share an arm.
func escapeSpelling(written string) (string, bool) {
	switch written {
	case "\n", "\\\n":
		return "\\n", true
	case "\r", "\\\r":
		return "\\r", true
	case "\t", "\\\t":
		return "\\t", true
	case "\v", "\\\v":
		return "\\v", true
	case "\f", "\\\f":
		return "\\f", true
	case "/":
		return "\\/", true
	}
	return "", false
}

// validPrecedingTokens is upstream's allowlist of tokens a regex literal may sit directly after.
//
// Copied rather than re-derived, and the reason is that it is a list of JUDGMENTS rather than a
// grammar. A `/` after an identifier or a closing paren is division, so the rewrite would silently
// change meaning; a `/` after `(` or `=` or `return` can only open a literal. Upstream enumerates
// the safe side and treats everything else as unsafe, which is the conservative direction, and it
// includes some entries that look surprising until you try them: `/` itself is on the list, because
// `a / /b/` is a division by a regex.
//
// The set is spelled as written upstream, including the compound assignment operators, and the
// longest match wins when reading backwards so `>>>=` is not read as `=`.
var validPrecedingTokens = map[string]bool{
	"(": true, ";": true, "[": true, ",": true, "=": true, "+": true, "*": true,
	"-": true, "?": true, "~": true, "%": true, "**": true, "!": true,
	"typeof": true, "instanceof": true, "&&": true, "||": true, "??": true,
	"return": true, "...": true, "delete": true, "void": true, "in": true,
	"<": true, ">": true, "<=": true, ">=": true, "==": true, "===": true,
	"!=": true, "!==": true, "<<": true, ">>": true, ">>>": true,
	"&": true, "|": true, "^": true, ":": true, "{": true, "=>": true,
	"*=": true, "<<=": true, ">>=": true, ">>>=": true, "^=": true, "|=": true,
	"&=": true, "??=": true, "||=": true, "&&=": true, "**=": true,
	"+=": true, "-=": true, "/=": true, "%=": true, "/": true,
	"do": true, "break": true, "continue": true, "debugger": true,
	"case": true, "throw": true,
}

// canSuggestLiteral reports whether the repair may be offered at all.
//
// Three gates, all upstream's, and each withholds the suggestion rather than the finding. The
// finding is about the constructor call and stands either way; only the proposed rewrite is at
// stake, which is why a doubtful case loses its suggestion instead of its report.
func canSuggestLiteral(ctx rule.Context, node *ast.Node, pattern string, flags string) bool {
	if commentsInside(ctx, node) {
		return false
	}
	if !precedingTokenAllowsALiteral(ctx, node) {
		return false
	}
	return patternIsValid(pattern, flags)
}

// commentsInside reports whether any comment sits inside the call being replaced.
//
// The replacement text is built from the pattern and the flags, so a comment anywhere in the call
// would be deleted by the repair. Upstream withholds for exactly this reason and the corpus carries
// two cases, one of which spells out in the comment text that it is explaining the pattern's safety.
func commentsInside(ctx rule.Context, node *ast.Node) bool {
	span := rule.TokenRange(ctx.SourceFile, node)
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= span.Pos() && comment.Range.End() <= span.End() {
			return true
		}
	}
	return false
}

// precedingTokenAllowsALiteral reads the token before the call and asks upstream's allowlist about
// it.
//
// Upstream reaches for `sourceCode.getTokenBefore`, which we have no equivalent of, so the token is
// read out of the source text backwards from the call's start. Comments are skipped the way
// whitespace is, matching `getTokenBefore`, which returns a real token rather than trivia.
//
// A call at the very start of a file has no preceding token and upstream permits the suggestion
// there, which is the `!tokenBefore` arm of its condition.
func precedingTokenAllowsALiteral(ctx rule.Context, node *ast.Node) bool {
	if ctx.SourceFile == nil {
		return false
	}
	text := ctx.SourceFile.Text()
	start := rule.TokenRange(ctx.SourceFile, node).Pos()
	if start <= 0 || start > len(text) {
		return true
	}

	index := lastMeaningfulIndexSkippingComments(text, start)
	if index < 0 {
		return true
	}

	// A hashbang is ONE token to upstream's tokenizer, so the token before a call on the next line
	// is the whole `#!/usr/bin/sh` line rather than the `sh` this backwards scan would otherwise
	// read. Reading `sh` puts an identifier in front of the call and the allowlist declines it,
	// which withholds a suggestion upstream offers. Measured against eslint 10.8.1, which reports
	// the case WITH a suggestion and pads it with a leading space.
	if sourceStartsWithHashbangBefore(text, start) {
		return true
	}

	if word, _ := wordBefore(text, index); word != "" {
		return validPrecedingTokens[word]
	}

	// Punctuation, read longest-first so `>>>=` is not mistaken for `=`. Four bytes is the longest
	// entry in the table.
	for length := 4; length >= 1; length-- {
		if index-length+1 < 0 {
			continue
		}
		candidate := text[index-length+1 : index+1]
		if validPrecedingTokens[candidate] {
			return true
		}
	}
	return false
}

// lastMeaningfulIndexSkippingComments walks backwards past whitespace and comments to the last byte
// of the previous real token.
//
// The package's `lastMeaningfulIndex` skips whitespace only, which is right for its caller and wrong
// here: `getTokenBefore` skips comments too, so `new RegExp(/* note */ 'a')` written after a `=`
// must see the `=`.
func lastMeaningfulIndexSkippingComments(text string, start int) int {
	index := start - 1
	for index >= 0 {
		switch text[index] {
		case ' ', '\t', '\r', '\n', '\v', '\f':
			index--
			continue
		}
		// A block comment ends in `*/`. Walk back to its opening `/*`.
		if index >= 1 && text[index] == '/' && text[index-1] == '*' {
			closing := index - 1
			opening := -1
			for scan := closing - 1; scan >= 1; scan-- {
				if text[scan] == '*' && text[scan-1] == '/' {
					opening = scan - 1
					break
				}
			}
			if opening < 0 {
				return -1
			}
			index = opening - 1
			continue
		}
		// A line comment ends at a newline, which the whitespace arm already consumed, so reaching
		// here on a non-newline byte means this is a real token.
		if lineCommentStart := lineCommentContaining(text, index); lineCommentStart >= 0 {
			index = lineCommentStart - 1
			continue
		}
		return index
	}
	return -1
}

// lineCommentContaining returns the index of the `//` opening a line comment that covers index, or
// -1 when index is not inside one.
//
// Scanning forward from the start of the line is what makes this correct rather than looking
// backwards for `//`: a `//` inside a string earlier on the line is not a comment, and only reading
// the line in order can tell the difference.
func lineCommentContaining(text string, index int) int {
	lineStart := 0
	for scan := index; scan >= 0; scan-- {
		if text[scan] == '\n' {
			lineStart = scan + 1
			break
		}
	}
	inSingle, inDouble, inTemplate := false, false, false
	for scan := lineStart; scan <= index; scan++ {
		character := text[scan]
		if character == '\\' {
			scan++
			continue
		}
		switch {
		case inSingle:
			if character == '\'' {
				inSingle = false
			}
		case inDouble:
			if character == '"' {
				inDouble = false
			}
		case inTemplate:
			if character == '`' {
				inTemplate = false
			}
		case character == '\'':
			inSingle = true
		case character == '"':
			inDouble = true
		case character == '`':
			inTemplate = true
		case character == '/' && scan+1 <= index && text[scan+1] == '/':
			return scan
		}
	}
	return -1
}

// patternIsValid reports whether a pattern and flags form a regular expression at all.
//
// Upstream asks regexpp to validate at the configured ecmaVersion. There is no ecmaVersion here, so
// this asks the same question at the latest edition, which is the divergence recorded on the rule.
//
// `esregexp.Compile` is deliberately NOT used, measured: it refuses the `v` flag outright and
// refuses `\p{Script=Greek}`, both of which are valid JavaScript that upstream accepts, and it
// accepts a bare `]` under `u` which JavaScript rejects. It answers a different question, whether
// the pattern can be EXECUTED by the vendored engine, and a rule deciding whether to offer a
// rewrite must not withhold one because our engine cannot run the pattern.
//
// So validity is answered structurally by the pattern walk, which returns false on a construct the
// scanner cannot make sense of. That covers the shapes the corpus turns on: a bare quantifier with
// nothing to repeat, an unterminated class or brace, and an unterminated group.
func patternIsValid(pattern string, flags string) bool {
	if !flagsAreValid(flags) {
		return false
	}
	if pattern == "" {
		return true
	}
	if !regexpattern.Walk(pattern, regexsyntax.ParseRegexFlags(flags), func(regexpattern.Character) bool {
		return true
	}) {
		return false
	}
	return quantifiersAndGroupsBalance(pattern, flags)
}

// flagsAreValid reports whether a flags string is one JavaScript accepts.
//
// Each flag must be known and no flag may repeat, and `u` and `v` are mutually exclusive, which is
// the corpus's `new RegExp('a', 'uv')` case.
func flagsAreValid(flags string) bool {
	seen := map[rune]bool{}
	for _, flag := range flags {
		if !strings.ContainsRune("dgimsuvy", flag) {
			return false
		}
		if seen[flag] {
			return false
		}
		seen[flag] = true
	}
	// `u` and `v` select different grammars for the same syntax, so JavaScript refuses a pattern
	// asking for both. This is the corpus's `new RegExp('a', 'uv')` case.
	return !(seen['u'] && seen['v'])
}

// quantifiersAndGroupsBalance reports whether every quantifier has something to repeat and every
// group and class is closed.
//
// The pattern walk reports characters and stops on a construct it cannot scan, which catches an
// unterminated class and an unterminated brace, but a leading `+` is a scannable character in its
// own right, so the "nothing to repeat" case needs asking separately. Four of upstream's cases turn
// on exactly that, `RegExp('+')` and its siblings.
func quantifiersAndGroupsBalance(pattern string, flags string) bool {
	parsedFlags := regexsyntax.ParseRegexFlags(flags)
	depth := 0
	previousWasRepeatable := false
	for index := 0; index < len(pattern); {
		character := pattern[index]
		switch character {
		case '\\':
			// `SkipPatternEscape` returns the escape's LENGTH in bytes, while `ClassEnd` below
			// returns an absolute index. Two helpers on the same shelf answering the same shape of
			// question in different units, and conflating them here spun this loop forever on
			// `RegExp('\\b')`: the length 2 was assigned as an index and the scan restarted at
			// byte 2 on every pass. Caught by a fixture hanging rather than failing, which is the
			// one failure mode a green suite cannot show you.
			width, ok := regexsyntax.SkipPatternEscape(pattern, index, parsedFlags)
			if !ok || width <= 0 {
				return false
			}
			index += width
			previousWasRepeatable = true
			continue
		case '[':
			// `ClassEnd` returns an absolute index just past the closing bracket, unlike
			// `SkipPatternEscape` above. The non-advance guard is what keeps a helper returning its
			// input from becoming another spin.
			end, ok := regexsyntax.ClassEnd(pattern, index, parsedFlags)
			if !ok || end <= index {
				return false
			}
			index = end
			previousWasRepeatable = true
			continue
		case '(':
			depth++
			previousWasRepeatable = false
		case ')':
			depth--
			if depth < 0 {
				return false
			}
			previousWasRepeatable = true
		case '*', '+', '?':
			if !previousWasRepeatable {
				return false
			}
			previousWasRepeatable = false
		case '{':
			end, ok := quantifierBraceEnd(pattern, index)
			if ok {
				if !previousWasRepeatable {
					return false
				}
				index = end
				previousWasRepeatable = false
				continue
			}
			// A `{` that opens no quantifier is a LITERAL brace under annex B and a syntax error
			// under `u` or `v`, which drop annex B entirely. Upstream reaches the same split
			// through regexpp's own unicode mode, and the corpus states it: `String.raw` + `\w{1, 2`
			// under `u` withholds its suggestion because the pattern does not parse.
			if parsedFlags.UV() {
				return false
			}
			previousWasRepeatable = true
		case '|':
			previousWasRepeatable = false
		default:
			previousWasRepeatable = true
		}
		index++
	}
	return depth == 0
}

// quantifierBraceEnd returns the index just past a `{n}`, `{n,}` or `{n,m}` quantifier, and whether
// the brace opened one at all.
//
// A `{` that does not open a quantifier is a literal brace in ECMAScript's annex-B grammar, which is
// why declining here is not an error: `RegExp('{')` is a valid pattern matching a brace.
func quantifierBraceEnd(pattern string, start int) (int, bool) {
	index := start + 1
	digits := 0
	for index < len(pattern) && pattern[index] >= '0' && pattern[index] <= '9' {
		index++
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	if index < len(pattern) && pattern[index] == ',' {
		index++
		for index < len(pattern) && pattern[index] >= '0' && pattern[index] <= '9' {
			index++
		}
	}
	if index < len(pattern) && pattern[index] == '}' {
		return index + 1, true
	}
	return 0, false
}
