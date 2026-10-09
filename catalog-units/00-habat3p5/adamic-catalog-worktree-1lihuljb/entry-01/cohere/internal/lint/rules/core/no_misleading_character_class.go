package core

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/literal"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The six findings are separate message ids rather than one, because they are six different
// mistakes that happen to share a shape. A reader who sees "combining class" knows to normalize the
// literal; one who sees "surrogate pair" without the flag knows to add the flag. Collapsing them
// into a single id would make the count assertable and the advice useless.
var messageMisleadingSurrogatePair = rule.Message{
	Id: "surrogatePair",
	Description: "This character class holds a surrogate pair spelled partly as a code point escape, " +
		"so what looks like one astral character is two independent class members. The class matches " +
		"either half on its own and never the character the author wrote, which is a match that " +
		"silently succeeds on the wrong input rather than an error anybody sees.",
}

var messageMisleadingSurrogatePairWithoutUnicodeFlag = rule.Message{
	Id: "surrogatePairWithoutUFlag",
	Description: "This character class holds an astral character while the pattern has neither the " +
		"u nor the v flag, so the engine reads it as its two UTF-16 halves and the class matches " +
		"either half alone. `/[👍]/` matches the two lone surrogates that spell it and never the " +
		"emoji; adding the flag is what makes the class mean the character it looks like.",
}

var messageMisleadingCombiningClass = rule.Message{
	Id: "combiningClass",
	Description: "This character class holds a base character followed by a combining mark, which " +
		"renders as one glyph and is two code points. The class matches the base or the mark " +
		"separately, so `/[Á]/` accepts a bare `A` and accepts a stray accent, and rejects the " +
		"accented letter it appears to name.",
}

var messageMisleadingEmojiModifier = rule.Message{
	Id: "emojiModifier",
	Description: "This character class holds an emoji followed by a skin tone modifier, which is one " +
		"glyph made of two code points. The class matches the unmodified emoji or the bare modifier, " +
		"so it accepts inputs nobody meant to accept and rejects the modified emoji itself.",
}

var messageMisleadingRegionalIndicator = rule.Message{
	Id: "regionalIndicatorSymbol",
	Description: "This character class holds a pair of regional indicator symbols, which render as " +
		"one flag and are two code points. The class matches either letter alone, so a flag class " +
		"built this way accepts every other flag that shares a letter with it.",
}

var messageMisleadingZeroWidthJoiner = rule.Message{
	Id: "zwj",
	Description: "This character class holds characters spliced by a zero width joiner, which renders " +
		"as one glyph and is three or more code points. The class matches each piece on its own, so " +
		"a family emoji written this way matches the individual people in it and never the family.",
}

// NoMisleadingCharacterClass flags a character class holding something that looks like one character
// and is several.
//
//	valid:   /^[abc]$/
//	valid:   /^[👍]$/u          // the flag makes the astral character one class member
//	valid:   /[\uD83D]/         // a lone surrogate is one code unit and means what it says
//	valid:   /[́]/         // a lone combining mark, likewise
//	valid:   /^[\q{👶🏻}]$/v      // the v-flag string disjunction matches the sequence as a unit
//	invalid: /^[Á]$/u
//	invalid: /^[👍]$/
//	invalid: /^[👶🏻]$/u
//	invalid: /^[🇯🇵]$/u
//	invalid: /^[👨‍👩‍👦]$/u
//	invalid: new RegExp("[🎵]")
//
// A character class is a set of single characters, and the language builds that set out of code
// points rather than out of glyphs. Every finding here is the same gap seen five ways: text that a
// human reads as one character is stored as several, so the class quietly becomes a set of the
// pieces. The failure is not a thrown error and not a non-match; it is a match against the wrong
// input, which is why it survives review and testing and shows up as a security or correctness bug
// much later.
//
// # Why six ids and not one
//
// The five sequence shapes are detected independently and all of them run over every sequence, so
// one class can report several times and with several ids. `/[👨‍👩‍👦]/` reports five findings: three
// surrogate pairs from the three astral characters and two joins from the two zero width joiners.
// That is upstream's count on that input, and it is right rather than duplicated, because each
// finding names a different pair of adjacent members. A fixture asserting one finding per input
// would be wrong here, and the snapshot says so: 174 diagnostics from 130 failing inputs.
//
// The surrogate pair case splits in two because the repair differs. When neither half was written as
// a code point escape the pattern is simply missing its flag, and adding `u` fixes it; that is the
// suggestion below. When one half was written as `\u{...}` the author already knew about code
// points, so adding a flag would not be the repair and none is offered.
//
// # What the u and v flags change
//
// Under `u` or `v` an astral character is a single class member, so `/[👍]/u` is clean while `/[👍]/`
// is not. That is the whole of the without-flag finding. The other four detectors fire in both
// modes, because a combining mark or a joiner is two code points whatever the flag says.
//
// `v` adds a second thing: it gives a class internal structure. `\q{...}` matches a string as a
// unit, a nested `[...]` is its own class, and `--` and `&&` are set operators. Each of those breaks
// a run of characters in two, so `/[🇯[A]🇵]/v` is clean where `/[🇯🇵]/u` is not: the two indicators
// are no longer adjacent. That break is what `regexsyntax.RegexCharBreaker` marks, and honoring it
// is the difference between reproducing upstream and reporting four false positives its corpus
// names.
//
// # A range endpoint breaks the run too
//
// `[a-z👍]` holds a range and then a character, and the range's two endpoints are not adjacent to
// each other in the sense this rule cares about, because everything between them is also in the set.
// Upstream ends the current sequence at a range's minimum and starts a new one at its maximum, which
// is why `/[👍-￿]/` reports once rather than twice: the pair before the range is
// adjacent, the pair spanning into it is not.
//
// # The constructor half
//
// `new RegExp("[🎵]")` and `RegExp("[🎵]", "u")` are checked alongside literals, because the mistake
// is identical and the constructor form is the one that survives the parser. The calls are found the
// way ESLint 10.8.1 finds them, through the shelf's port of eslint-utils' ReferenceTracker, so a local
// named RegExp is not the global and an alias or `globalThis.RegExp` is (#jjfa7qb).
//
// The pattern and the flags are read as ESLint reads them, through reference.ConstantStringIn: a
// string, a substitution-free template, a concatenation of those, or a constant binding holding one.
// A pattern that is not constant is not checked, and neither is one whose value is a RegExp object.
// Flags that are present and not constant decline the call, since they could hold the `u` that makes
// the class correct. A regex literal passed with flags is checked under the call's flags instead of its
// own, and only there; passed alone it is left to the literal listener, under its own.
//
// A finding in a string or template literal points at the characters, through the cooked-to-raw
// mapping below. A pattern that reached the call any other way, a name or a concatenation, has no
// characters in the file to point at, so each kind of finding is reported once, at the argument, as
// ESLint does.
//
// A pattern written as a string carries one more layer of escaping than a literal does, and that
// layer is the rule's sharpest edge upstream. `new RegExp("[\\uD83D\\uDC4D]")` passes the regex
// engine the six characters `\uD83D`, so the regex sees an escape; `new RegExp("[👍]")`
// passes it the astral character itself. Both report, and they report for different reasons. This
// port reads the string's cooked value, which is what the engine receives, so both land on the same
// path and the distinction upstream draws between them does not need to be redrawn here.
//
// # allowEscape
//
// `{"allowEscape": true}` lets an author who spelled a member as an escape keep it: `/[Á]/`
// is clean, because nobody writes `́` without knowing it is a separate code point. What counts
// as escaped is upstream's test, applied to each member's source text: it starts with a backslash
// and is not just backslashes in front of the very character it denotes. So `́`, `\x41` and
// `\n` are escapes, and `\\` or `\è` are not, since the character is still sitting there in plain
// sight. For a pattern written as a string the source text is the string's raw text, so a string
// escape counts as well as a regex one: `RegExp("[Á]")` and `RegExp("[A\\u0301]")` are both
// clean. A pattern that reached the call through a name or a concatenation has no source text of
// its own, and nothing in it counts as escaped.
//
// An escaped member drops out of every check it would have taken part in, with one exception: a
// combining mark is excused only by its own spelling. The base it attaches to may be escaped or not,
// so `/[\n̅]/` still reports, the mark being right there in the source.
//
// An astral character split into its halves takes the spelling of the whole: `\u{1F44D}` in a
// string is two escaped halves, and `\👍` is two plain ones.
var NoMisleadingCharacterClass = rule.Rule{
	Name: "no-misleading-character-class",

	// The tracker tells the global RegExp from a local one by where its symbol is declared, and a
	// constant argument is followed to its binding the same way.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[NoMisleadingCharacterClassOptions](options)
		allowEscape := settings.AllowEscape
		// The regex literals a RegExp call took with flags, so already checked under the call's.
		// Filled from the file node, which the walk enters before any literal.
		var checkedByACall map[*ast.Node]bool
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				checkedByACall = map[*ast.Node]bool{}
				// Every trace starts at a reference to RegExp, so a file that never spells it has
				// nothing to follow, and the tracker's index is never built.
				if ctx.TypeChecker == nil || !strings.Contains(ctx.SourceFile.Text(), "RegExp") {
					return
				}
				tracker := reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)
				for _, tracked := range tracker.GlobalReferences(regExpCallTraceMap) {
					checkRegExpConstructorCall(ctx, tracked.Node, checkedByACall, allowEscape)
				}
			},
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				if checkedByACall[node] {
					return
				}
				text := node.Text()
				pattern, flags := regexsyntax.PatternAndFlags(text)
				if pattern == "" {
					return
				}
				literalStart := node.End() - len(text)
				checkRegexPattern(ctx, pattern, literalStart+1, flags, patternSource(pattern, allowEscape),
					func() []rule.Suggestion {
						return unicodeFlagSuggestionForLiteral(node, text, flags)
					})
			},
		}
	},
}

// NoMisleadingCharacterClassOptions is upstream's one option object.
//
// AllowEscape defaults to false, as upstream's, so a bare severity reports every misleading class
// however its members were spelled.
type NoMisleadingCharacterClassOptions struct {
	AllowEscape bool `json:"allowEscape"`
}

// memberSource returns the source text of the pattern bytes [start, end), for the allowEscape test.
// Nil means no member counts as escaped: the option is off, or the pattern has no source of its own.
type memberSource func(start int, end int) string

// patternSource is the memberSource for a pattern whose bytes are its own source text.
func patternSource(pattern string, allowEscape bool) memberSource {
	if !allowEscape {
		return nil
	}
	return func(start int, end int) string {
		return pattern[start:end]
	}
}

// checkRegExpConstructorCall checks a call or construction of RegExp, as ESLint's tracker loop does.
func checkRegExpConstructorCall(
	ctx rule.Context,
	call *ast.Node,
	checkedByACall map[*ast.Node]bool,
	allowEscape bool,
) {
	arguments := call.Arguments()
	if len(arguments) == 0 {
		return
	}
	patternNode := ast.SkipParentheses(arguments[0])
	hasFlags := len(arguments) > 1
	flags := ""
	flagsKnown := true
	if hasFlags {
		flags, flagsKnown = reference.ConstantStringIn(ctx, arguments[1])
	}

	// A regex literal handed over with flags is checked under them, and the literal listener steps
	// aside for it, whether or not the flags can be read. Handed over alone, it is the literal
	// listener's, under its own flags.
	if patternNode.Kind == ast.KindRegularExpressionLiteral {
		if !hasFlags {
			return
		}
		checkedByACall[patternNode] = true
		if !flagsKnown {
			return
		}
		text := patternNode.Text()
		pattern, _ := regexsyntax.PatternAndFlags(text)
		if pattern == "" {
			return
		}
		checkRegexPattern(ctx, pattern, patternNode.End()-len(text)+1, flags,
			patternSource(pattern, allowEscape), nil)
		return
	}

	if !flagsKnown || reference.IsConstantRegExpIn(ctx, patternNode) {
		return
	}
	pattern, isConstant := reference.ConstantStringIn(ctx, patternNode)
	if !isConstant {
		return
	}
	switch patternNode.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		checkRegexPatternInStringLiteral(ctx, pattern, patternNode, flags, allowEscape)
	default:
		reported := map[string]bool{}
		checkRegexPatternWithReporter(pattern, flags, nil, func(finding pendingMisleadingFinding) {
			if reported[finding.message.Id] {
				return
			}
			reported[finding.message.Id] = true
			ctx.ReportNode(patternNode, finding.message)
		})
	}
}

// checkRegexPatternInStringLiteral checks a pattern that reached the constructor as a string, where
// the text the engine sees and the text on disk are two different strings.
//
// The pattern checked is the literal's cooked value, because that is what the RegExp constructor
// receives: `"[\\uD83D\\uDC4D]"` hands the engine a backslash and the letters, and `"[👍]"` hands it
// the character. Both are correct inputs to the class scanner.
//
// But a finding has to point at the file, and cooked offsets are not file offsets. `"[á]"` is
// six bytes on disk and four cooked, and `"\\u200D"` is eight on disk and seven cooked, so an
// offset taken from the cooked string lands mid-character in the raw one and the reported span is
// garbage. That is not a hypothetical: it is the defect this function exists to fix, and it was
// invisible to every message-id fixture in the corpus because those assert what fired and never
// where.
//
// The mapping is built by decoding the raw text and recording, for each cooked byte, which raw byte
// produced it. When the two are the same length the mapping is the identity and is skipped, which is
// the common case.
func checkRegexPatternInStringLiteral(
	ctx rule.Context,
	pattern string,
	node *ast.Node,
	flags string,
	allowEscape bool,
) {
	textRange := rule.TokenRange(ctx.SourceFile, node)
	raw := ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	// The raw text carries its quotes or backticks; the cooked value does not.
	if len(raw) < 2 {
		return
	}
	rawBody := raw[1 : len(raw)-1]
	bodyStart := textRange.Pos() + 1

	// A pure shortcut, and a mutation sweep confirmed it as one: forcing every literal through the
	// mapper below leaves every fixture green, because when the two texts are the same length the
	// mapper computes exactly the identity mapping. It stays because the common case is a pattern
	// with no escapes at all, and walking it rune by rune to rediscover that offsets are offsets is
	// work this rule does on every string argument in the tree.
	//
	// Every string escape is wider than what it produces, so a raw text as long as the cooked one
	// holds no escape and is the cooked text, which is why its members' source is the pattern itself.
	if len(rawBody) == len(pattern) {
		checkRegexPattern(ctx, pattern, bodyStart, flags, patternSource(pattern, allowEscape), nil)
		return
	}

	offsets := literal.CookedToRaw(rawBody, pattern)
	// Unreachable through the parser we have, and kept for the same reason as the flag guard above:
	// a sweep replacing this test with `false` survives, because the mapper only refuses when the
	// raw and cooked texts genuinely disagree, and a cooked value the parser produced from a raw
	// text it read cannot disagree with it. What the line protects against is this file's own
	// escape-width table falling behind the parser's, which is a change nobody would notice without
	// it: the alternative to declining is reporting at an offset computed from a desynced walk,
	// which points at real source and is wrong.
	if offsets == nil {
		// The raw text could not be reconciled with the cooked one, so no offset can be trusted.
		// Declining is the only honest move: a finding at a guessed span points somewhere real and
		// wrong, which is worse than no finding at all.
		return
	}
	// A member's source is the raw text its cooked bytes came from, which is where a string escape
	// shows: `"[Á]"` cooks to the mark itself, and the raw `́` is what excuses it.
	var source memberSource
	if allowEscape {
		source = func(start int, end int) string {
			return rawBody[offsets[start]:offsets[end]]
		}
	}
	checkRegexPatternMapped(ctx, pattern, bodyStart, flags, offsets, source)
}

// checkRegexPatternMapped is checkRegexPattern for a pattern whose offsets need translating back
// into the raw source before anything is reported.
func checkRegexPatternMapped(
	ctx rule.Context,
	pattern string,
	bodyStart int,
	flagsText string,
	offsets []int,
	source memberSource,
) {
	collector := &mappedFindingCollector{ctx: ctx, bodyStart: bodyStart, offsets: offsets}
	checkRegexPatternWithReporter(pattern, flagsText, source, collector.report)
}

// mappedFindingCollector reports a finding at the raw offsets its cooked span maps to.
type mappedFindingCollector struct {
	ctx       rule.Context
	bodyStart int
	offsets   []int
}

func (collector *mappedFindingCollector) report(finding pendingMisleadingFinding) {
	if finding.start >= len(collector.offsets) || finding.end >= len(collector.offsets) {
		return
	}
	collector.ctx.ReportRange(
		core.NewTextRange(
			collector.bodyStart+collector.offsets[finding.start],
			collector.bodyStart+collector.offsets[finding.end]),
		finding.message)
}

// checkRegexPattern reports every misleading sequence in every character class of a pattern.
//
// suggestions is called only for the without-flag surrogate finding, which is the one finding whose
// repair is known, and it is a function rather than a value so a literal's rewrite is not computed
// for the patterns that never report.
func checkRegexPattern(
	ctx rule.Context,
	pattern string,
	patternStart int,
	flagsText string,
	source memberSource,
	suggestions func() []rule.Suggestion,
) {
	checkRegexPatternWithReporter(pattern, flagsText, source, func(finding pendingMisleadingFinding) {
		textRange := core.NewTextRange(patternStart+finding.start, patternStart+finding.end)
		if finding.suggest && suggestions != nil {
			ctx.ReportRangeWithSuggestions(textRange, finding.message, suggestions()...)
			return
		}
		ctx.ReportRange(textRange, finding.message)
	})
}

// checkRegexPatternWithReporter finds every misleading sequence in a pattern and hands each to a
// reporter, which decides where in the file it lands.
//
// The two callers differ only in that translation: a regex literal's pattern offsets are file
// offsets plus a constant, and a string literal's are not, because the text the engine sees was
// decoded from the text on disk. Keeping the search in one place is what stops the two from drifting
// into two slightly different rules.
func checkRegexPatternWithReporter(
	pattern string,
	flagsText string,
	source memberSource,
	report func(pendingMisleadingFinding),
) {
	// The cheap gate: no bracket, no class. Most patterns in a real tree have none, and the class
	// walk below is the expensive part.
	if !strings.ContainsRune(pattern, '[') {
		return
	}

	flags := regexsyntax.ParseRegexFlags(flagsText)

	// Under `u` or `v` a brace that opens no quantifier is a SyntaxError, so `RegExp('{ [Á]', 'u')`
	// never runs and the class inside it cannot mislead anybody. Upstream declines it because its
	// parser refuses the pattern outright, and two of its clean cases are this exact shape.
	//
	// Checked here rather than by handing the pattern to the repository's regex compiler, which was
	// measured against these inputs and is the wrong gate: it accepts this brace, rejects the `v`
	// flag entirely, and rejects `/[👍-￿]/` which upstream reports on. Trading two false
	// positives for four false negatives is not a trade, so the one construct that matters is
	// tested directly.
	if flags.UV() && patternHasLoneQuantifierBrace(pattern, flags) {
		return
	}

	// Findings are collected before any is reported, because a pattern that turns out to be
	// unparseable must produce none at all. `new RegExp('[Á] [ ')` opens a well-formed class and
	// then an unterminated one; the walk reports the first and only then discovers the second, so
	// reporting as it goes would emit a finding about a pattern that throws a SyntaxError before it
	// can match anything. Upstream declines the whole call because its parser refuses the pattern
	// outright, and five of its clean cases are exactly this shape.
	var found []pendingMisleadingFinding
	classesParsed := true
	walked := regexsyntax.IterateRegexCharacterClasses(pattern, flags, func(start, end int) {
		elements, _, ok := regexsyntax.ParseRegexCharacterClassWithEnd(pattern, start, end, flags)
		if !ok {
			classesParsed = false
			return
		}
		for _, sequence := range characterSequences(pattern, elements, flags, source) {
			found = append(found, misleadingSequenceFindings(sequence)...)
		}
	})
	if !walked || !classesParsed {
		return
	}

	for _, finding := range found {
		report(finding)
	}
}

// pendingMisleadingFinding is one finding held back until the whole pattern is known to parse.
type pendingMisleadingFinding struct {
	start   int
	end     int
	message rule.Message
	// suggest marks the one finding whose repair is known: an astral character in a pattern that is
	// simply missing its flag.
	suggest bool
}

// misleadingCharacter is one class member, reduced to what the detectors need: the code point it
// denotes, where its source sits, and whether it was written as a code point escape.
type misleadingCharacter struct {
	value uint32
	start int
	end   int
	// isCodePointEscape records `\u{...}` spelling, which decides only which of the two surrogate
	// pair messages fires. An author who wrote `\u{d83d}` chose the code point deliberately, so
	// telling them to add a flag would be the wrong advice.
	isCodePointEscape bool
	// escaped is allowEscape's verdict on this member: on, and written as an escape. Off, it is never
	// set, so the detectors read it unconditionally.
	escaped bool
}

// characterSequences splits a class's elements into the runs of adjacent characters the detectors
// scan.
//
// The splitting is the part that is easy to get wrong and that upstream's clean cases are mostly
// about. Three things end a run:
//
//   - a breaker, which is `\d`, `\p{...}`, a v-flag `\q{...}`, a nested class, or a set operator.
//     Characters on opposite sides of one are not adjacent, which is why `/[🇯\q{abc}🇵]/v` is clean.
//   - a range's minimum, which ends the run, with the range's maximum starting a new one. The
//     characters a range covers sit between its endpoints, so the endpoints are not neighbours.
//   - the class itself ending. Two classes side by side are two runs, which is why
//     `/[‌][‍][a]/` is clean while `/[‌‍]/` is not.
//
// Under no unicode flag an astral character occupies two UTF-16 code units and the engine treats
// each as its own class member, so it is expanded here into its surrogate halves. That expansion is
// the entire without-flag surrogate finding, and it is done in the rule rather than in the shared
// scanner because it is a question about what this rule means by adjacency rather than about how the
// pattern parses.
//
// source is the allowEscape test's view of the file, nil when the option is off or the pattern has
// no source of its own.
func characterSequences(
	pattern string,
	elements []regexsyntax.RegexCharElement,
	flags regexsyntax.RegexFlags,
	source memberSource,
) [][]misleadingCharacter {
	var sequences [][]misleadingCharacter
	var current []misleadingCharacter

	flush := func() {
		if len(current) > 0 {
			sequences = append(sequences, current)
			current = nil
		}
	}

	appendCharacter := func(value uint32, start int, end int, isCodePointEscape bool) {
		escaped := source != nil && isAcceptableEscape(source(start, end), value)
		// Under no unicode flag the engine sees UTF-16 code units, so an astral character is two
		// members rather than one. Neither half has source text of its own, and where each points is
		// ESLint's UTF-16 span moved onto bytes: the high half is the empty place where the character
		// starts and the low half is the whole character. So a pair of halves points at the whole
		// character, and a join ending at a high half, as in `👨‍👩` read without the flag, stops
		// where 👩 starts, which is where ESLint's span stops too.
		if !flags.UV() && value > 0xFFFF && !isCodePointEscape {
			high := 0xD800 + ((value - 0x10000) >> 10)
			low := 0xDC00 + ((value - 0x10000) & 0x3FF)
			current = append(current,
				misleadingCharacter{value: high, start: start, end: start, escaped: escaped},
				misleadingCharacter{value: low, start: start, end: end, escaped: escaped})
			return
		}
		current = append(current, misleadingCharacter{
			value: value, start: start, end: end, isCodePointEscape: isCodePointEscape, escaped: escaped,
		})
	}

	for _, element := range elements {
		switch element.Kind {
		case regexsyntax.RegexCharBreaker:
			flush()
		case regexsyntax.RegexCharRange:
			// The minimum joins the run that was building and closes it; the maximum opens the next
			// one. Reproducing that shape rather than dropping ranges is what makes
			// `/[👍-￿]/` report exactly once.
			minimumEnd, maximumStart := rangeEndpointBounds(pattern, element, flags)
			appendCharacter(element.Value, element.Start, minimumEnd, element.IsUBrace)
			flush()
			appendCharacter(element.Max, maximumStart, element.End, element.MaxIsUBrace)
		default:
			appendCharacter(element.Value, element.Start, element.End, element.IsUBrace)
		}
	}
	flush()
	return sequences
}

// rangeEndpointBounds splits a range element at its hyphen, returning where its minimum ends and its
// maximum starts.
//
// The shelf's range element spans `min-max` whole, and each endpoint needs its own text twice over:
// a finding that ends at the minimum, as in `/[👍-￿]/`, points at the pair and not at
// the range, and allowEscape asks how each endpoint was spelled. The minimum's width is read the way
// the class parser read it, and if the hyphen is not where that width says, the whole range is
// returned for both rather than a guess.
func rangeEndpointBounds(
	pattern string,
	element regexsyntax.RegexCharElement,
	flags regexsyntax.RegexFlags,
) (int, int) {
	start := element.Start
	width := 0
	switch {
	case pattern[start] != '\\':
		_, width = utf8.DecodeRuneInString(pattern[start:])
	case flags.UV() && element.Value > 0xFFFF && start+2 < len(pattern) && pattern[start+1] == 'u' &&
		pattern[start+2] != '{':
		// `👍` under a unicode flag is two escapes the parser folded into one member.
		width = 12
	default:
		width, _ = regexsyntax.SkipPatternEscape(pattern, start, flags)
	}
	minimumEnd := start + width
	if width == 0 || minimumEnd >= element.End || pattern[minimumEnd] != '-' {
		return element.End, element.Start
	}
	return minimumEnd, minimumEnd + 1
}

// isAcceptableEscape is upstream's test for a member allowEscape excuses: its source starts with a
// backslash and is not merely backslashes in front of the character it denotes.
//
// `́`, `\x41` and `\n` pass. `\\` and `\è` do not, because the character is written out after
// the backslash, and neither does `\👍` for either of its halves, which is why the value compared is
// the whole character's.
func isAcceptableEscape(source string, value uint32) bool {
	if !strings.HasPrefix(source, `\`) {
		return false
	}
	last, width := utf8.DecodeLastRuneInString(source)
	prefix := source[:len(source)-width]
	if prefix != "" && strings.Trim(prefix, `\`) == "" && uint32(last) == value {
		return false
	}
	return true
}

// misleadingSequenceFindings runs every detector over one run of adjacent characters.
//
// All five run, and each reports independently, so one run can produce findings of several kinds.
// The order matches upstream's so that a reader comparing this port against it, or comparing two
// diagnostic streams, sees the same sequence rather than the same set.
func misleadingSequenceFindings(sequence []misleadingCharacter) []pendingMisleadingFinding {
	var found []pendingMisleadingFinding
	report := func(from misleadingCharacter, to misleadingCharacter, message rule.Message) {
		found = append(found, pendingMisleadingFinding{
			start: from.start, end: to.end, message: message})
	}

	// An escaped member takes part in nothing, except as the base a combining mark attaches to: the
	// mark's own spelling is what allowEscape asks about, as upstream's combiningClass reads its
	// previous member from the unfiltered run.
	for index := 1; index < len(sequence); index++ {
		previous, current := sequence[index-1], sequence[index]

		if !current.escaped && isCombiningCharacter(current.value) && !isCombiningCharacter(previous.value) {
			report(previous, current, messageMisleadingCombiningClass)
		}
		if previous.escaped || current.escaped {
			continue
		}
		if isRegionalIndicatorSymbol(previous.value) && isRegionalIndicatorSymbol(current.value) {
			report(previous, current, messageMisleadingRegionalIndicator)
		}
		if isEmojiModifier(current.value) && !isEmojiModifier(previous.value) {
			report(previous, current, messageMisleadingEmojiModifier)
		}
	}

	// The joiner detector reads three characters rather than two, because a joiner is only
	// misleading when it actually joins: it needs something on each side, and neither side may be
	// another joiner. `/[‍]/` is a class holding one joiner and means what it says.
	//
	// A chain of joins is one finding, as ESLint's zwj generator yields it: a join whose left
	// character is the previous join's right one extends that sequence, so the family emoji
	// 👨‍👩‍👦 reports once over all five characters rather than as two overlapping pairs. Reporting
	// each pair was 11 of ESLint's corpus rows reading as extra (#jjfa7qb).
	joinStart, joinEnd := -1, -1
	for index := 1; index+1 < len(sequence); index++ {
		if sequence[index].value != zeroWidthJoiner {
			continue
		}
		if sequence[index-1].escaped || sequence[index].escaped || sequence[index+1].escaped {
			continue
		}
		if sequence[index-1].value == zeroWidthJoiner || sequence[index+1].value == zeroWidthJoiner {
			continue
		}
		if joinStart >= 0 && joinEnd == index-1 {
			joinEnd = index + 1
			continue
		}
		if joinStart >= 0 {
			report(sequence[joinStart], sequence[joinEnd], messageMisleadingZeroWidthJoiner)
		}
		joinStart, joinEnd = index-1, index+1
	}
	if joinStart >= 0 {
		report(sequence[joinStart], sequence[joinEnd], messageMisleadingZeroWidthJoiner)
	}

	for index := 1; index < len(sequence); index++ {
		previous, current := sequence[index-1], sequence[index]
		if previous.escaped || current.escaped || !isSurrogatePair(previous.value, current.value) {
			continue
		}
		// Which of the two messages fires is decided by spelling rather than by value. When either
		// half was written `\u{...}` the author picked the code point on purpose and adding a flag
		// is not the repair, so the finding carries no suggestion. When neither was, the pattern is
		// missing its flag and that is exactly what to propose.
		if previous.isCodePointEscape || current.isCodePointEscape {
			report(previous, current, messageMisleadingSurrogatePair)
			continue
		}
		found = append(found, pendingMisleadingFinding{
			start:   previous.start,
			end:     current.end,
			message: messageMisleadingSurrogatePairWithoutUnicodeFlag,
			suggest: true,
		})
	}
	return found
}

// unicodeFlagSuggestionForLiteral proposes appending `u` to a regex literal's flags.
//
// This is a suggestion and never a fix, and the reason is the whole point of the rewrite: adding `u`
// changes what the rest of the pattern means. Under `u` an identity escape like `\a` is a syntax
// error, a lone `{` is a syntax error, and `\p{L}` stops being the letters `p{L}` and becomes a
// property escape. So the edit can turn a working pattern into one that throws at parse time, and a
// rewrite that can break the file is a thing a human chooses rather than a thing the engine applies
// while nobody is looking.
//
// Upstream reaches the same conclusion and declares the rule `suggestion` rather than `fix`. It also
// re-parses the pattern under the new flag and withholds the rewrite when the result would not
// parse; this port does not, and instead withholds it whenever the pattern holds anything whose
// meaning `u` is known to change. That is a coarser test in the same direction: it declines some
// rewrites upstream would offer, and never offers one upstream would decline.
func unicodeFlagSuggestionForLiteral(node *ast.Node, text string, flags string) []rule.Suggestion {
	// No fixture can kill this line, and it stays anyway. A mutation sweep dropped each half of the
	// test and both mutants survived, which sent me looking for the input that would tell the
	// versions apart. There is none, and the reason is one step upstream: reaching here at all takes
	// a `suggest` finding, which takes two adjacent surrogate halves, and under `u` or `v` the class
	// parser has already folded `👍` into a single astral element before this rule sees
	// it. The only way two halves survive that fold is for one to be written `\u{...}`, which routes
	// to the no-suggestion branch before this function is ever called.
	//
	// So the guard is unreachable through the parser we have today, not through the language. If the
	// shelf ever stops folding, this is the line that keeps the rule from proposing a flag a pattern
	// already carries, and its absence would be silent.
	if strings.ContainsRune(flags, 'u') || strings.ContainsRune(flags, 'v') {
		return nil
	}
	pattern, _ := regexsyntax.PatternAndFlags(text)
	if patternMeaningChangesUnderUnicodeFlag(pattern) {
		return nil
	}
	return []rule.Suggestion{{
		Message: rule.Message{
			Id: "suggestUnicodeFlag",
			Description: "Add the u flag so the class matches the character rather than its parts. " +
				"The flag changes how the whole pattern parses, so this is offered rather than applied.",
		},
		Fixes: []rule.Fix{{
			Range: core.NewTextRange(node.End(), node.End()),
			Text:  "u",
		}},
	}}
}

// patternMeaningChangesUnderUnicodeFlag reports whether adding `u` could change how a pattern parses.
//
// Three constructs are legal without the flag and are errors or mean something else with it: an
// identity escape of a character that needs no escaping, an unquantified brace, and a lone `]`.
// Rather than enumerate the escapes the flag forbids, this refuses any escape outside the set the
// flag leaves alone, which keeps a new escape from silently becoming safe when it is not.
func patternMeaningChangesUnderUnicodeFlag(pattern string) bool {
	for index := 0; index < len(pattern); index++ {
		character := pattern[index]
		if character == '{' || character == '}' {
			// A brace that does not open a quantifier is a literal brace without the flag and a
			// syntax error with it. Deciding which it is means parsing the quantifier grammar, so
			// the coarse answer is taken: braces are only safe when they close a `\u{...}`, which
			// the escape branch below has already consumed.
			return true
		}
		if character != '\\' || index+1 >= len(pattern) {
			continue
		}
		next := pattern[index+1]
		if next == 'u' && index+2 < len(pattern) && pattern[index+2] == '{' {
			closing := strings.IndexByte(pattern[index+3:], '}')
			if closing < 0 {
				return true
			}
			index += 3 + closing
			continue
		}
		if !strings.ContainsRune(unicodeFlagSafeEscapes, rune(next)) {
			return true
		}
		index++
	}
	return false
}

// patternHasLoneQuantifierBrace reports whether a pattern holds a `{` outside a character class
// that opens no counted quantifier.
//
// Without a unicode flag such a brace is an ordinary literal character and the pattern is fine. With
// one it is a SyntaxError, which is the whole reason this is asked: a pattern that cannot compile
// cannot mislead a reader about what it matches, so the rule has nothing to say about it.
//
// Only `{` is examined and a stray `}` is ignored, because the two are not symmetric in the grammar:
// a lone closing brace is legal under the unicode flag while a lone opening brace is not.
func patternHasLoneQuantifierBrace(pattern string, flags regexsyntax.RegexFlags) bool {
	for index := 0; index < len(pattern); {
		switch pattern[index] {
		case '\\':
			step, ok := regexsyntax.SkipPatternEscape(pattern, index, flags)
			if !ok {
				return false
			}
			index += step
		case '[':
			// A brace inside a class is a literal brace in every mode, so classes are stepped over
			// whole rather than scanned.
			end, ok := regexsyntax.ClassEnd(pattern, index, flags)
			if !ok {
				return false
			}
			index = end
		case '{':
			if !quantifierBraceEnds(pattern, index) {
				return true
			}
			index++
		default:
			index++
		}
	}
	return false
}

// quantifierBraceEnds reports whether the `{` at index opens a counted quantifier.
//
// The accepted shapes are the language's three: `{n}`, `{n,}` and `{n,m}`. Anything else is a
// literal brace, which is the case the caller is looking for.
func quantifierBraceEnds(pattern string, index int) bool {
	scan := index + 1
	digitsStart := scan
	for scan < len(pattern) && pattern[scan] >= '0' && pattern[scan] <= '9' {
		scan++
	}
	if scan == digitsStart {
		return false
	}
	if scan < len(pattern) && pattern[scan] == ',' {
		scan++
		for scan < len(pattern) && pattern[scan] >= '0' && pattern[scan] <= '9' {
			scan++
		}
	}
	return scan < len(pattern) && pattern[scan] == '}'
}

// unicodeFlagSafeEscapes are the escape characters whose meaning the `u` flag leaves alone.
//
// Everything here is either a syntax character that must be escaped in both modes, a named escape,
// or a class shorthand. An identity escape of anything else is legal without the flag and rejected
// with it, which is the case this list exists to exclude.
const unicodeFlagSafeEscapes = `^$\.*+?()[]{}|/dDsSwWbBnrtvf0123456789xuck<>=!:,-`

// zeroWidthJoiner is U+200D, the code point that splices emoji into one glyph.
const zeroWidthJoiner = 0x200D

// isRegionalIndicatorSymbol reports whether a code point is one of the 26 letters that pair into a
// flag. Each stands for one Latin letter, and two of them make an ISO 3166 country code.
func isRegionalIndicatorSymbol(value uint32) bool {
	return value >= 0x1F1E6 && value <= 0x1F1FF
}

// isEmojiModifier reports whether a code point is one of the five skin tone modifiers.
func isEmojiModifier(value uint32) bool {
	return value >= 0x1F3FB && value <= 0x1F3FF
}

// isSurrogatePair reports whether two code units are a lead and a trail that together denote one
// astral character.
func isSurrogatePair(high uint32, low uint32) bool {
	return high >= 0xD800 && high <= 0xDBFF && low >= 0xDC00 && low <= 0xDFFF
}

// isCombiningCharacter reports whether a code point is a mark that attaches to the character before
// it, covering the Mn, Mc and Me general categories and the variation selectors.
//
// The ranges are the blocks dedicated to combining marks rather than the full Unicode property,
// which is what upstream matches and is a deliberate narrowing on both sides: script-specific marks
// in the Devanagari and Hiragana blocks share the same categories and are not listed, so a Devanagari
// sign following its consonant is not reported. Reproduced rather than extended, because widening it
// here would make this port report on inputs upstream's corpus asserts are clean, and the fix for a
// stated gap is a Unicode table rather than a longer list of guesses.
func isCombiningCharacter(value uint32) bool {
	switch {
	case value >= 0x0300 && value <= 0x036F: // Combining Diacritical Marks
		return true
	case value >= 0x1AB0 && value <= 0x1AFF: // Combining Diacritical Marks Extended
		return true
	case value >= 0x1DC0 && value <= 0x1DFF: // Combining Diacritical Marks Supplement
		return true
	case value >= 0x20D0 && value <= 0x20FF: // Combining Diacritical Marks for Symbols
		return true
	case value >= 0xFE00 && value <= 0xFE0F: // Variation Selectors
		return true
	case value >= 0xFE20 && value <= 0xFE2F: // Combining Half Marks
		return true
	case value >= 0xE0100 && value <= 0xE01EF: // Variation Selectors Supplement
		return true
	}
	return false
}
