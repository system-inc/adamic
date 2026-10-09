// Package dotnotation is the judgment behind `dot-notation`: which bracket accesses with a literal key
// read better as dot accesses, which dot accesses must be bracketed when keywords are disallowed, and
// the repairs for both.
//
// Lifted out of `internal/lint/rules/core/dot_notation.go` because `@typescript-eslint/dot-notation`
// needs the same judgment and a rule package may not import another rule package
// (`TestRulePackagesStayLeaves`). The extension delegates rather than re-deriving, which is what
// upstream does (`baseRule.create(context)`), so the two rules cannot disagree about anything the core
// decides. What the extension adds is one question asked before a computed access is reported; see
// Settings.DeclinesComputed.
package dotnotation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// Settings is what the judgment reads, decoded by each rule from its own option surface.
type Settings struct {
	// AllowKeywords permits `a.class` and friends. Upstream's default is true.
	AllowKeywords bool

	// AllowPattern leaves alone a computed key that matches it, which is how a project keeps
	// `a['snake_case']` while still converting `a['b']`. Nil allows nothing.
	AllowPattern *regexp.Regexp

	// DeclinesComputed, when set, is asked about each computed access the core would otherwise
	// report, and true leaves it alone. It is where typescript-eslint's extension puts its private,
	// protected and index-signature exemptions. Upstream asks before the core looks at the key at all;
	// asking only once every structural test has passed gives the same verdicts, since the question
	// can only remove a finding, and consults the checker on the few accesses that would report rather
	// than on every bracket in the file.
	DeclinesComputed func(access *ast.Node, key *ast.Node) bool
}

// CompileAllowPattern compiles an allowPattern option, and an empty one allows nothing.
//
// A pattern Go's RE2 cannot compile is dropped rather than crashing the run, which reports MORE rather
// than less. Upstream's `new RegExp` throws and takes every other rule's verdict on that file with it.
func CompileAllowPattern(source string) *regexp.Regexp {
	if source == "" {
		return nil
	}
	compiled, err := regexp.Compile(source)
	if err != nil {
		return nil
	}
	return compiled
}

// Listeners is the core rule's listeners under settings: a computed access with a convertible literal
// key reports useDot, and under AllowKeywords false a keyword after a dot reports useBrackets.
func Listeners(ctx rule.Context, settings Settings) rule.Listeners {
	return rule.Listeners{
		ast.KindElementAccessExpression: func(node *ast.Node) {
			access := node.AsElementAccessExpression()
			if access == nil || access.ArgumentExpression == nil {
				return
			}
			// Our parser keeps `KindParenthesizedExpression`; espree folds it away before the
			// rule ever sees it. So `foo[('bar')]` reaches this listener wrapping a node
			// upstream never sees, and without the unwrap it costs two of upstream's own
			// reporting cases. Measured: cases invalid[17] and invalid[18] report nothing
			// until the key is unwrapped, while invalid[19] `(foo)['bar']` passes either way
			// because the parentheses are on the RECEIVER rather than the key.
			//
			// A loop rather than one step, because `((x))` nests. Not `ast.SkipParentheses`:
			// it dereferences its argument, and this project lost 167 files to that on an
			// optional node.
			argument := access.ArgumentExpression
			for argument != nil && argument.Kind == ast.KindParenthesizedExpression {
				argument = argument.AsParenthesizedExpression().Expression
			}
			if argument == nil {
				return
			}

			key, formatted, readable := literalKey(argument)
			if !readable {
				return
			}
			checkComputed(ctx, node, access, argument, key, formatted, settings)
		},

		ast.KindPropertyAccessExpression: func(node *ast.Node) {
			if settings.AllowKeywords {
				return
			}
			access := node.AsPropertyAccessExpression()
			if access == nil || access.Name() == nil {
				return
			}
			name := access.Name()
			if name.Kind != ast.KindIdentifier {
				return
			}
			if !keywords[name.Text()] {
				return
			}
			if fixes, repairable := keywordFix(ctx, node, access, name); repairable {
				ctx.ReportNodeWithFixes(name, useBracketsMessage(name.Text()), fixes...)
				return
			}
			ctx.ReportNode(name, useBracketsMessage(name.Text()))
		},
	}
}

// validIdentifier is upstream's `validIdentifier`, `/^[a-zA-Z_$][\w$]*$/u`.
//
// Deliberately ASCII-only, matching upstream. A key spelled with a non-ASCII letter is a legal
// JavaScript identifier, and this pattern rejects it, so the rule stays silent rather than
// proposing a dot access for it. That is upstream's conservatism rather than an oversight here:
// `\w` in a non-Unicode-aware sense is `[A-Za-z0-9_]`, and Go's `\w` agrees exactly.
var validIdentifier = regexp.MustCompile(`^[a-zA-Z_$][\w$]*$`)

// decimalInteger is upstream's `DECIMAL_INTEGER_PATTERN`, used by `isDecimalInteger`.
//
// It decides whether the fixer must insert a SPACE before the dot. `5['prop']` becomes `5 .prop`
// rather than `5.prop`, because the latter re-lexes as the number `5.` followed by `prop`, which is
// a syntax error -- a fixer writing broken source into a file it was asked to repair.
//
// The pattern is copied from `ast-utils.js:64` rather than reconstructed, and that distinction cost
// four corpus cases. An earlier version of this line read `^(0|[1-9]\d*)$`, which is what the rule
// "looks like" it should need and is not what upstream has. The real one accepts two more shapes,
// and both are in the corpus:
//
//	0[0-7]*[89]\d*    a leading zero followed by an 8 or 9, so `08` and `090` and `018` are
//	                  DECIMAL integers rather than legacy octals, and do need the space
//	[1-9](?:_?\d)*    numeric separators, so `5_000` and `5_000_00` need it too
//
// What stays exempt: `01` and `01234567` are genuine legacy octals whose grammar forbids a
// following fraction, `5.000_000` already contains a dot, and `0b1010_1010` is not decimal. The
// corpus pins the split exactly -- 25 and 26 expect no space, 27 through 31 expect one.
var decimalInteger = regexp.MustCompile(`^(?:0|0[0-7]*[89][0-9]*|[1-9](?:_?[0-9])*)$`)

// keywords is upstream's shared ES3 keyword list, copied verbatim and in order.
//
// ES3 rather than modern: it includes `abstract`, `boolean`, `byte`, `goto` and friends, which are
// not reserved in any current JavaScript. That is deliberate upstream, and the list is what
// `allowKeywords: false` bans, so trimming it to today's reserved words would silently narrow the
// rule.
var keywords = map[string]bool{
	"abstract": true, "boolean": true, "break": true, "byte": true, "case": true, "catch": true,
	"char": true, "class": true, "const": true, "continue": true, "debugger": true,
	"default": true, "delete": true, "do": true, "double": true, "else": true, "enum": true,
	"export": true, "extends": true, "false": true, "final": true, "finally": true, "float": true,
	"for": true, "function": true, "goto": true, "if": true, "implements": true, "import": true,
	"in": true, "instanceof": true, "int": true, "interface": true, "long": true, "native": true,
	"new": true, "null": true, "package": true, "private": true, "protected": true, "public": true,
	"return": true, "short": true, "static": true, "super": true, "switch": true,
	"synchronized": true, "this": true, "throw": true, "throws": true, "transient": true,
	"true": true, "try": true, "typeof": true, "var": true, "void": true, "volatile": true,
	"while": true, "with": true,
}

// useDotMessage builds the finding for a computed access that should be a dot access.
//
// The Id is fixed and only the Description moves, so `ExpectFindings` can still count these while
// the rendered text names the key the way upstream's `{{key}}` interpolation does. Upstream renders
// a string or a null/boolean literal key through `JSON.stringify` and a template key in backticks,
// which is why the caller passes the formatted form rather than the raw value.
func useDotMessage(formattedKey string) rule.Message {
	return rule.Message{
		Id: "useDot",
		Description: fmt.Sprintf(
			"[%s] is better written in dot notation. Bracket notation is for keys computed at "+
				"runtime; with a literal inside, it is the same access written the long way, and it "+
				"reads as though the key were dynamic when it is not. Editors also stop offering "+
				"completion and rename through a bracketed key, so the property quietly drops out of "+
				"every refactor that touches it.",
			formattedKey,
		),
	}
}

// useBracketsMessage builds the finding for a keyword accessed with a dot under
// `allowKeywords: false`.
func useBracketsMessage(key string) rule.Message {
	return rule.Message{
		Id: "useBrackets",
		Description: fmt.Sprintf(
			".%s is a syntax error. This project has turned off `allowKeywords`, which targets "+
				"engines old enough that a reserved word after a dot fails to parse at all. The "+
				"access itself is fine; only the spelling has to change, so write it in brackets.",
			key,
		),
	}
}

// checkComputed is upstream's `checkComputedProperty`.
func checkComputed(
	ctx rule.Context,
	node *ast.Node,
	access *ast.ElementAccessExpression,
	keyNode *ast.Node,
	key string,
	formattedKey string,
	settings Settings,
) {
	if !validIdentifier.MatchString(key) {
		return
	}
	if !settings.AllowKeywords && keywords[key] {
		return
	}
	if settings.AllowPattern != nil && settings.AllowPattern.MatchString(key) {
		return
	}
	if settings.DeclinesComputed != nil && settings.DeclinesComputed(node, keyNode) {
		return
	}

	if fixes, repairable := computedFix(ctx, node, access, keyNode, key); repairable {
		ctx.ReportNodeWithFixes(keyNode, useDotMessage(formattedKey), fixes...)
		return
	}
	ctx.ReportNode(keyNode, useDotMessage(formattedKey))
}

// computedFix builds the repair that turns `a['b']` into `a.b`.
//
// Returns false where upstream returns without yielding, which is a DECLINE rather than an absent
// repair: the finding is still reported, only the fix is withheld. Two corpus cases are `output:
// null` for the comment reason below.
//
// The repair is a single span replacement from the opening bracket to the closing one, which is how
// upstream's two yields combine once the optional-chaining branch is accounted for. Assembling it as
// one edit rather than two avoids the overlapping-edit refusal the harness makes when two fixes
// touch one expression.
func computedFix(
	ctx rule.Context,
	node *ast.Node,
	access *ast.ElementAccessExpression,
	keyNode *ast.Node,
	key string,
) ([]rule.Fix, bool) {
	if ctx.SourceFile == nil || access.Expression == nil {
		return nil, false
	}
	source := ctx.SourceFile.Text()

	// The span runs from the `[` that follows the object to the `]` that closes the access.
	openBracket := strings.IndexByte(source[access.Expression.End():node.End()], '[')
	if openBracket < 0 {
		return nil, false
	}
	openBracket += access.Expression.End()
	closeBracket := node.End() - 1
	if closeBracket <= openBracket || closeBracket > len(source) || source[closeBracket] != ']' {
		return nil, false
	}

	// Upstream: "Don't perform any fixes if there are comments inside the brackets." Applying the
	// repair would delete them, which is why upstream reports without fixing rather than reformatting.
	if hasCommentBetween(ctx, openBracket, closeBracket) {
		return nil, false
	}

	// An optional access already carries its own `?.`, so the dot must not be added again:
	// `obj?.['prop']` becomes `obj?.prop`, not `obj?..prop`. Upstream expresses this as
	// `if (!node.optional)` around the dot insertion.
	prefix := "."
	if isOptionalAccess(source, access.Expression.End(), openBracket) {
		prefix = ""
	} else if decimalInteger.MatchString(strings.TrimSpace(
		source[access.Expression.Pos():access.Expression.End()])) {
		// `5['prop']` must become `5 .prop`. Without the space the result re-lexes as the numeric
		// literal `5.` followed by `prop` and stops parsing, which is a fixer writing a syntax
		// error into a file it was asked to repair. See decimalInteger.
		prefix = " ."
	}

	replacement := prefix + key

	// Upstream's third yield: "Insert a space after the property if it will be connected to the
	// next token." Without it `foo['bar']instanceof baz` repairs to `foo.barinstanceof baz`, which
	// welds two tokens into one identifier and silently changes what the statement means. That is
	// the fixer-corrupts-a-file class, and it was caught by upstream's own `output` fixture rather
	// than by reading -- this arm was simply missing from the first draft.
	//
	// Upstream decides with `canTokensBeAdjacent`, which tokenizes both sides. The question here is
	// narrower, because the left side is always an identifier this rule just wrote: the pair is
	// unsafe exactly when the next character could continue that identifier.
	if closeBracket+1 < len(source) && continuesAnIdentifier(source[closeBracket+1]) {
		replacement += " "
	}

	return []rule.Fix{
		rule.ReplaceRange(core.NewTextRange(openBracket, closeBracket+1), replacement),
	}, true
}

// continuesAnIdentifier answers whether a byte could extend the identifier before it.
//
// ASCII-only, matching the identifier pattern this rule uses to decide a key is convertible: a key
// that reached the fixer is `[a-zA-Z_$][\w$]*`, so the only way the next character can join it is by
// being another such character. A non-ASCII byte cannot start an identifier this rule would have
// accepted, and treating it as a non-continuation only ever inserts one space too few in a case the
// rule never reaches.
func continuesAnIdentifier(character byte) bool {
	switch {
	case character >= 'a' && character <= 'z',
		character >= 'A' && character <= 'Z',
		character >= '0' && character <= '9',
		character == '_',
		character == '$':
		return true
	}
	return false
}

// isOptionalAccess answers whether the access between the object and the bracket is `?.`.
//
// Read off the source text rather than from a node flag, because the question is about the tokens
// the fixer is about to replace and reading them is what keeps the repair and the test of it
// looking at the same bytes.
func isOptionalAccess(source string, from int, to int) bool {
	return strings.Contains(source[from:to], "?.")
}

// literalKey reads a computed key that upstream considers convertible.
//
// Three accepted shapes, matching upstream branch for branch: a string or boolean literal, a `null`
// literal (which upstream handles separately because `typeof null` is `"object"`), and a static
// template literal with no substitutions.
//
// The third return separates "not a convertible key" from "an empty key", which matters because an
// empty string is a real literal and simply fails the identifier test rather than being skipped.
//
// A NUMERIC literal is deliberately absent, and it is the shape a reader adds back. `a[0]` must stay
// bracketed: `a.0` is a syntax error, and upstream's `literalTypesToCheck` is `{string, boolean}`
// precisely to exclude it. Corpus case `a['1']` is CLEAN for the same reason once it reaches the
// identifier pattern, which rejects a leading digit.
func literalKey(argument *ast.Node) (key string, formatted string, readable bool) {
	switch argument.Kind {
	case ast.KindStringLiteral:
		text := argument.Text()
		// Upstream renders a Literal key through `JSON.stringify`, so the message shows the key
		// with double quotes regardless of how it was written in source.
		return text, jsonQuote(text), true

	case ast.KindTrueKeyword:
		return "true", "true", true
	case ast.KindFalseKeyword:
		return "false", "false", true
	case ast.KindNullKeyword:
		return "null", "null", true

	case ast.KindNoSubstitutionTemplateLiteral:
		// Upstream's `isStaticTemplateLiteral` accepts only a template with no substitutions, which
		// is exactly this node kind: a template carrying `${...}` parses as a different kind here and
		// never reaches this arm.
		text := argument.Text()
		return text, "`" + text + "`", true
	}
	return "", "", false
}

// jsonQuote renders a string the way `JSON.stringify` does for the message text.
//
// Only the escapes `JSON.stringify` emits for the characters that can appear in a key this rule
// accepts. The identifier test upstream applies is ASCII-only, so a key reaching the message is
// alphanumerics, underscore and dollar; the quoting below is nonetheless complete for the shapes a
// reader might add later rather than being minimal for today's callers.
func jsonQuote(text string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, character := range text {
		switch character {
		case '"':
			builder.WriteString(`\"`)
		case '\\':
			builder.WriteString(`\\`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			builder.WriteRune(character)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

// hasCommentBetween answers upstream's `commentsExistBetween`.
//
// Upstream declines every repair that would span a comment, because the replacement text does not
// carry it and applying the fix would delete it. Four corpus cases are `output: null` for this
// reason, and the decline is reproduced rather than improved on.
//
// Reads the shared per-file comment list rather than rescanning; see `comments.ForFile`.
func hasCommentBetween(ctx rule.Context, start int, end int) bool {
	if end <= start {
		return false
	}
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= start && comment.Range.End() <= end {
			return true
		}
	}
	return false
}

// keywordFix builds the repair that turns `a.true` into `a["true"]` under
// `allowKeywords: false`.
//
// Returns false where upstream returns without yielding. Two corpus cases are `output: null` here
// and they are not the same decline:
//
//	foo. /* comment */ while    a comment between the dot and the name would be deleted
//	let.if()                    `let[` at the start of a statement parses as a DESTRUCTURING
//	                            declaration rather than a member access, so bracketing changes
//	                            what the statement IS
//
// The second is the sharper one, and it is narrower than it looks: `let?.true` IS repaired, to
// `let?.["true"]`, because the optional-chaining token keeps the form unambiguous. Upstream guards
// with `node.object.name === 'let' && !node.optional`, and both halves are load-bearing -- corpus
// cases 22 and 37 sit on either side of that `&&`.
func keywordFix(
	ctx rule.Context,
	node *ast.Node,
	access *ast.PropertyAccessExpression,
	name *ast.Node,
) ([]rule.Fix, bool) {
	if ctx.SourceFile == nil || access.Expression == nil {
		return nil, false
	}
	source := ctx.SourceFile.Text()

	optional := isOptionalAccess(source, access.Expression.End(), name.Pos())

	// The `let[` hazard. Only for a non-optional access; see the doc comment.
	if !optional &&
		access.Expression.Kind == ast.KindIdentifier &&
		access.Expression.Text() == "let" {
		return nil, false
	}

	// The span to replace runs from the dot to the end of the property name. For an optional
	// access the `?.` stays and only the name is replaced, which is what produces `obj?.["true"]`.
	start := name.Pos()
	if !optional {
		dot := strings.LastIndexByte(source[access.Expression.End():name.Pos()], '.')
		if dot < 0 {
			return nil, false
		}
		start = access.Expression.End() + dot
	}

	// Upstream: "Don't perform any fixes if there are comments between the dot and the property
	// name." Applying the repair would delete them.
	if hasCommentBetween(ctx, access.Expression.End(), name.End()) {
		return nil, false
	}

	return []rule.Fix{
		rule.ReplaceRange(core.NewTextRange(start, name.End()), `["`+name.Text()+`"]`),
	}, true
}
