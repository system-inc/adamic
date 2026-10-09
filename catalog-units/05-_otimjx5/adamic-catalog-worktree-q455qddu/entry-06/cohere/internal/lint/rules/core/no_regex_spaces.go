package core

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/literal"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexpattern"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageRegexSpaces = rule.Message{
	Id: "multipleSpaces",
	Description: "This regular expression matches a run of spaces that nobody can count without " +
		"putting a cursor on it. Three spaces and four look the same in every editor, so the " +
		"pattern silently stops matching the moment someone deletes one, and the diff that did it " +
		"shows a change to a line of whitespace. Writing the count as a quantifier makes the " +
		"intent readable and makes an accidental edit visible.",
}

// NoRegexSpaces flags a run of two or more literal spaces in a regular expression.
//
//	valid:   /foo {3}bar/
//	valid:   /  +/           // the run is quantified, so it means something other than a count
//	valid:   /[  ]/          // inside a class, where repetition is not what the spaces do
//	valid:   /bar  baz/
//	valid:   new RegExp('  ', flags)   // flags unknown, so the parse is unknown
//	invalid: /foo   bar/
//	invalid: new RegExp('foo   bar')
//
// Three things about this rule are easy to get backwards, and each is a real defect if you do.
//
// **At most one finding per regex, and it is the FIRST qualifying run, not the last.** Upstream's
// accumulator is named `last_space_span` and does not hold the last run: once a run of two or more
// has latched, every later run is discarded, and the single space bookkeeping exists only to step
// past lone spaces on the way to the first real one. So `/  foo   /` reports the leading pair and
// says nothing about the trailing three. That is why upstream's corpus is 26 inputs and 26
// diagnostics: no input can report twice, by construction rather than by coincidence. Reading the
// field name and tracking the last run instead passes 25 of the 26 imported cases.
//
// **Depth is the whole discrimination and it is checked on the character, not the pattern.** A
// space inside a quantifier or a character class is not a run somebody lost count of: `/  +/` is
// one space repeated and `/[  ]/` is a set that happens to name the space twice. The walk carries
// both depths so this rule can skip either, which is the case `regexpattern` was built for.
//
// **The constructor path parses the COOKED string but reports RAW offsets, and it needs both.** For
// a regex literal the two texts are the same and none of this arises. For the constructor they are
// not, and each half is load bearing in a different direction.
//
// The parse has to be cooked because that is the string the RegExp constructor receives. In
// `new RegExp(' \\[   ')` the `\\[` is a JS string escape, so the pattern the engine sees is ` [   `,
// which opens a character class that never closes. Parsing the raw bytes instead reads `\\[` as a
// regex escape, finds no class, and reports three spaces that upstream and ESLint both call clean.
//
// The span has to be raw because that is what indexes the file. `new RegExp('\\\\d  ')` has a raw body
// one byte longer than its cooked value, so a cooked offset applied to the file points one byte
// short, and the fix would eat a character it did not mean to.
//
// oxc gets both at once by running its reader in string-literal mode, cooking while carrying each
// unit's original span. We have no such reader, so `literal.CookedToRaw` rebuilds the correspondence
// instead. ESLint takes the third route: it computes against the cooked value and then refuses the
// fix whenever cooked and raw differ, so it reports those cases without repairing them. Ours
// repairs them, which is upstream's behavior rather than ESLint's.
//
// # Where this deliberately differs from oxc
//
// oxc hardcodes a nil flags text into its constructor parser, so the `v` flag never reaches the
// parse on that path and nested classes are not honored there. Its own snapshot shows the
// consequence: `/[[    ]    ]    /v` reports the trailing run and
// `new RegExp('[[    ]    ]    ', 'v')` reports a run five bytes earlier, inside the nested class,
// though the two spell one regex. Its fix vector bakes that in, so copying its fixtures would
// import the bug together with the fixture that blesses it.
//
// We pass the real flags on both paths, so the two spellings agree, which is also what ESLint does
// for both. This is rule-local rather than a house pattern upstream follows: of oxc's three rules
// using that parser, only this one withholds the flags.
//
// The repair is a fix rather than a suggestion because it cannot change what the pattern matches: a
// run of n spaces and ` {n}` accept exactly the same input, and there is no second valid answer to
// choose between.
//
// A pattern the scanner cannot make sense of is skipped rather than reported, matching upstream and
// for the same reason as the sibling regex rules: an unterminated construct is a syntax error the
// parser has already refused, so a second complaint here is noise on a file that does not compile.
var NoRegexSpaces = rule.Rule{
	Name: "no-regex-spaces",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// reportLiteral walks a regex literal's pattern and reports its first unquantified,
		// unclassed run of two or more spaces. A literal's source text and the pattern the engine
		// receives are the same bytes, so no cooking or offset mapping is involved here; the
		// constructor path below is where that stops being true.
		reportLiteral := func(pattern string, flags string, patternStart int) {
			// The cheap gate upstream uses, and it is load bearing rather than an optimization: a
			// pattern whose only adjacent spaces are spelled out (`\x20\x20`) or separated by an
			// escape (`/bar \ baz/`) never reaches the parser at all.
			if !strings.Contains(pattern, "  ") {
				return
			}

			start, end, found := firstConsecutiveSpaces(pattern, regexsyntax.ParseRegexFlags(flags))
			if !found {
				return
			}

			runRange := core.NewTextRange(patternStart+start, patternStart+end)
			ctx.ReportRangeWithFixes(runRange, messageRegexSpaces,
				rule.ReplaceRange(runRange, " {"+strconv.Itoa(end-start)+"}"))
		}

		// checkConstructor handles both `RegExp(...)` and `new RegExp(...)`, which upstream treats
		// identically and which differ here only in which node carries the arguments.
		checkConstructor := func(callee *ast.Node, arguments *ast.NodeList) {
			callee = ast.SkipParentheses(callee)
			if callee == nil || callee.Kind != ast.KindIdentifier {
				return
			}
			if callee.AsIdentifier().Text != "RegExp" {
				return
			}
			if arguments == nil || len(arguments.Nodes) == 0 {
				return
			}

			// A second argument that is not a string literal means the flags are unknowable, and
			// the flags decide how the pattern parses: under `v` a nested class swallows spaces
			// that are at depth zero without it. Reporting on a guess would be a finding whose
			// position depends on a value this file does not contain, so the case is skipped.
			// `new RegExp('  ', flags)` is a pass case upstream for exactly this reason.
			flags := ""
			if len(arguments.Nodes) >= 2 {
				flagsNode := arguments.Nodes[1]
				if flagsNode.Kind != ast.KindStringLiteral {
					return
				}
				flags = flagsNode.AsStringLiteral().Text
			}

			patternNode := arguments.Nodes[0]
			if patternNode.Kind != ast.KindStringLiteral {
				return
			}

			// TokenRange rather than Loc, because Loc starts before leading trivia and would put
			// both the finding and its fix into whatever precedes the literal.
			literalRange := rule.TokenRange(ctx.SourceFile, patternNode)
			raw := ctx.SourceFile.Text()[literalRange.Pos():literalRange.End()]
			// The quotes are the literal's delimiters rather than part of the pattern. Anything
			// shorter than two characters has no body to strip them from.
			if len(raw) < 2 {
				return
			}
			rawBody := raw[1 : len(raw)-1]
			cooked := patternNode.AsStringLiteral().Text

			// The gate runs on the RAW body, so a pattern whose adjacent spaces are only spelled
			// out never reaches the parse. See the doc comment.
			if !strings.Contains(rawBody, "  ") {
				return
			}

			offsets := literal.CookedToRaw(rawBody, cooked)
			if offsets == nil {
				return
			}

			start, end, found := firstConsecutiveSpaces(cooked, regexsyntax.ParseRegexFlags(flags))
			if !found {
				return
			}

			runRange := core.NewTextRange(
				literalRange.Pos()+1+offsets[start],
				literalRange.Pos()+1+offsets[end])
			ctx.ReportRangeWithFixes(runRange, messageRegexSpaces,
				rule.ReplaceRange(runRange, " {"+strconv.Itoa(end-start)+"}"))
		}

		return rule.Listeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				text := node.Text()
				pattern, flags := regexsyntax.PatternAndFlags(text)
				if pattern == "" {
					return
				}
				// The node's own text ends at End(), but Pos() sits before leading trivia, so the
				// start is derived from the end rather than read directly. Getting this backwards
				// reports the finding at whatever comment precedes the literal, on a line no
				// suppression directive the author can write is able to reach.
				literalStart := node.End() - len(text)
				reportLiteral(pattern, flags, literalStart+1)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				checkConstructor(call.Expression, call.Arguments)
			},

			ast.KindNewExpression: func(node *ast.Node) {
				newExpression := node.AsNewExpression()
				checkConstructor(newExpression.Expression, newExpression.Arguments)
			},
		}
	},
}

// firstConsecutiveSpaces returns the pattern-relative span of the first run of two or more spaces
// that sits outside every quantifier and character class.
//
// The accumulator holds at most one span and the rules for replacing it are what produce "first run
// of two or more" rather than "last run of any width". A space adjacent to the held span extends
// it; a space that is not adjacent replaces it only while the span is still a single space. So the
// held span walks forward across lone spaces and then stops moving the instant it reaches width
// two, which is the first qualifying run and the one reported.
//
// A walk that cannot make sense of the pattern returns nothing rather than a partial answer, since
// a span measured inside a pattern the scanner lost track of is a position in a file that does not
// parse.
func firstConsecutiveSpaces(pattern string, flags regexsyntax.RegexFlags) (int, int, bool) {
	start, end := 0, 0
	held := false

	walked := regexpattern.Walk(pattern, flags, func(character regexpattern.Character) bool {
		if character.QuantifierDepth > 0 || character.ClassDepth > 0 {
			return true
		}
		// The value rather than the source text, so a space spelled `\x20` or ` ` is not one
		// of these: it has already been counted by whoever wrote it out, which is the repair this
		// rule asks for.
		if character.Value != ' ' {
			return true
		}

		switch {
		case !held:
			start, end, held = character.Start, character.End, true
		case end == character.Start:
			end = character.End
		case end-start == 1:
			start, end = character.Start, character.End
		}
		return true
	})

	if !walked || !held || end-start < 2 {
		return 0, 0, false
	}
	return start, end, true
}

// The cooked-to-raw offset mapping this rule needs already exists: `literal.CookedToRaw` in
// no_misleading_character_class.go, which asks the identical question of the identical inputs (a
// string literal's raw body and its cooked value) and returns the identical shape. Adopted rather
// than duplicated, and probed against this rule's own corpus before adopting rather than taken on
// the strength of its doc comment: 18 raw and cooked pairs from this corpus, and the two
// implementations agreed on 17.
//
// The eighteenth was a trailing line continuation, where the shipped mapper returned a complete
// mapping and should have declined, because its terminal check confirmed only that the cooked text
// had finished and not that the raw one had. That was a live defect in its own rule as well as this
// one, and it is fixed there rather than worked around here.
