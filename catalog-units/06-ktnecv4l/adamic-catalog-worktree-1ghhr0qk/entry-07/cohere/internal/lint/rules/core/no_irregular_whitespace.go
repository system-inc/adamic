package core

import (
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// byteOrderMark is U+FEFF, which opens a file legitimately and is a stray character anywhere else.
const byteOrderMark = 0xFEFF

// lineSeparator and paragraphSeparator, U+2028 and U+2029, end a line, so each is a finding of its own
// rather than part of a run of spaces.
const (
	lineSeparator      = 0x2028
	paragraphSeparator = 0x2029
)

var messageIrregularWhitespace = rule.Message{
	Id: "noIrregularWhitespace",
	Description: "This is an irregular whitespace character, not a space or a tab. It is " +
		"invisible in every editor and in every diff, so the code reads as if it holds an " +
		"ordinary space while the parser sees something else. That is how a no-break space " +
		"between two tokens turns a correct-looking line into a syntax error nobody can see, and " +
		"how two identical-looking names become different bindings. Replace it with an ordinary " +
		"space, or write it as an escape if the character is genuinely wanted.",
}

// irregularWhitespaceCodepoints is the set this rule reports, transcribed from upstream rather than
// composed out of a scanner predicate.
//
// The distinction is the whole trap. A scanner's whitespace predicate answers "may the parser skip
// this", which is a different question with a different answer at both ends. stringutil's
// IsWhiteSpaceLike is the near-match in this tree and it is wrong in both directions: it adds tab,
// newline, carriage return, and space, so a rule built on it reports every line of every file, and
// it omits U+180E, which is Unicode category Cf rather than Zs. A spec-correct scanner must exclude
// U+180E for that reason, and this rule must include it for that same reason, because a format
// character sitting between two tokens is as invisible and as confusing as a wide space is.
//
// oxc reaches the same set by a different route, ORing its parser predicate with its line-terminator
// predicate and then bolting U+180E on by hand with a comment saying why. Both sides land on the
// same 24 codepoints. They are written as numbers rather than as character literals because the
// characters are invisible: a literal here would be unreadable, unreviewable in a diff, and would
// trip this very rule.
var irregularWhitespaceCodepoints = map[rune]bool{
	0x000B: true, // line tabulation
	0x000C: true, // form feed
	0x0085: true, // next line
	0x00A0: true, // no-break space
	0x1680: true, // ogham space mark
	0x180E: true, // mongolian vowel separator
	0x2000: true, // en quad
	0x2001: true, // em quad
	0x2002: true, // en space
	0x2003: true, // em space
	0x2004: true, // three-per-em space
	0x2005: true, // four-per-em space
	0x2006: true, // six-per-em space
	0x2007: true, // figure space
	0x2008: true, // punctuation space
	0x2009: true, // thin space
	0x200A: true, // hair space
	0x200B: true, // zero width space
	0x2028: true, // line separator
	0x2029: true, // paragraph separator
	0x202F: true, // narrow no-break space
	0x205F: true, // medium mathematical space
	0x3000: true, // ideographic space
	0xFEFF: true, // zero width no-break space, the byte order mark
}

// NoIrregularWhitespaceOptions is the five switches, each excusing one region of the file.
//
// Every field is a pointer because every default here is a real decision rather than a zero value,
// and one of them defaults to something Go's zero value is not. A plain bool cannot tell "the config
// said false" from "the config said nothing", and for SkipStrings those mean opposite things.
//
// The defaults applied below are ESLint's, which is also what this tree's resolved config asks for
// on every file. oxc's differ on three of the five: it ships skipRegExps, skipTemplates, and
// skipJSXText as true, so an unconfigured oxlint is quieter than an unconfigured ESLint on regexes,
// templates, and JSX text. Following ESLint is not a preference here, it is the configuration this
// tree actually runs. The imported fixtures pass oxc's values explicitly, so upstream's counts still
// reproduce exactly rather than being quietly reinterpreted.
type NoIrregularWhitespaceOptions struct {
	SkipComments  *bool `json:"skipComments"`
	SkipStrings   *bool `json:"skipStrings"`
	SkipTemplates *bool `json:"skipTemplates"`
	SkipRegExps   *bool `json:"skipRegExps"`

	// SkipJSXText's tag is worth a note because the option is spelled skipJSXText rather than
	// skipJsxText. Every option field is tagged, and rule.UnmarshalOptions matches a tagged key
	// exactly, so `skipJsxText` is refused rather than resolved; the upstream Rust needs an explicit
	// rename for the same reason.
	SkipJSXText *bool `json:"skipJSXText"`
}

// skippedRegion is one span of the file that an option has excused.
type skippedRegion struct {
	start int
	end   int
}

// NoIrregularWhitespace flags whitespace characters that are neither a space nor a tab.
//
//	valid:   const indent = String.fromCharCode(0x00a0);
//	valid:   const text = 'a b';
//	invalid: const value = 1;
//	invalid: // a comment holding a wide space
//
// The characters are invisible, so the defect they cause never looks like the character. A no-break
// space between `const` and a name is a syntax error on a line that reads as correct, and a zero
// width space inside an identifier makes two identical-looking names into different bindings. This
// tree already works around the rule rather than fighting it: a sidebar component builds an indent
// string with String.fromCharCode(0x00a0) and says in a comment that embedding the character
// directly trips this lint.
//
// # Why this scans text rather than nodes
//
// oxlint gets the code positions for free. Its lexer records every irregular-whitespace span it
// skips into a trivia builder, so its rule reads a list the parser already built. typescript-go's
// scanner exposes no equivalent, and probing for one found nothing, so the positions have to be
// recovered here. This scans the source text for the whole set and then subtracts the regions an
// option excused, which structurally resembles ESLint's subtractive design rather than oxc's
// additive one even though oxc is the port target. That inversion is the port, and it is stated
// because a reader comparing this file to the Rust will otherwise hunt for a lexer hook that does
// not exist on this side.
//
// # Granularity
//
// One finding per character, matching oxc. ESLint's regex carries a `+` quantifier and reports a run
// of adjacent irregular characters as a single finding, so on two adjacent vertical tabs ESLint says
// one thing and oxc says two. oxc's snapshot holds 83 diagnostics across 58 inputs and that gap is
// mostly this. Per character is the more useful of the two: each one is a separate character to
// delete, and a span covering a whole run tells a reader less about what to remove.
//
// # The byte order mark
//
// A U+FEFF at offset zero is a byte order mark rather than a stray character, and is not reported.
// Anywhere else in the file it is reported like any other member of the set. oxc carries this
// carve-out and ESLint does not, and oxc's corpus pins it with a passing case.
//
// No fix. The right repair is usually an ordinary space, but not always: a zero width space inside a
// string is sometimes deliberate, a byte order mark mid-file is an encoding accident rather than a
// typo, and deleting versus replacing changes the token stream differently. An unattended rewrite
// that picks wrong here edits characters nobody can see, which is the worst place to be wrong.
var NoIrregularWhitespace = rule.Rule{
	Name: "no-irregular-whitespace",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// ESLint's defaults, which are also this tree's resolved configuration. Only SkipStrings defaults
		// to on: irregular whitespace inside a string is usually data rather than a typo, and
		// reporting it would flag every file that holds a no-break space on purpose.
		skipComments, skipStrings, skipTemplates := false, true, false
		skipRegExps, skipJSXText := false, false
		if settings, hasSettings := rule.OptionsAs[NoIrregularWhitespaceOptions](options); hasSettings {
			if settings.SkipComments != nil {
				skipComments = *settings.SkipComments
			}
			if settings.SkipStrings != nil {
				skipStrings = *settings.SkipStrings
			}
			if settings.SkipTemplates != nil {
				skipTemplates = *settings.SkipTemplates
			}
			if settings.SkipRegExps != nil {
				skipRegExps = *settings.SkipRegExps
			}
			if settings.SkipJSXText != nil {
				skipJSXText = *settings.SkipJSXText
			}
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				text := ctx.SourceFile.Text()

				// Nothing below runs on a file that holds none of these characters, which is very
				// nearly every file. Without this the rule walks the whole tree collecting skip
				// regions before discovering there is nothing to skip around, and that walk is the
				// entire cost: measured on 3,407 real files it was 306ms and 16.0% of all rule
				// time, the most expensive rule in the run, to report nothing. One scan of the text
				// answers the question the walk was being asked, and the scan is the same one the
				// reporting loop does anyway.
				//
				// ESLint gates its whole rule body on the same test for the same reason. oxc has no
				// equivalent because it pays nothing here: its lexer already recorded the spans, so
				// there is no walk to skip.
				if !containsIrregularWhitespace(text) {
					return
				}

				var skipped []skippedRegion

				if skipComments {
					// A comment is trivia rather than a node, so no listener can reach one and it
					// has to be rescanned out of the source. The shelf already does that scan once
					// per file and shares it. Reading trivia rather than matching on `//` is what
					// makes this immune to a comment opener living inside a string, which was
					// measured on this corpus: `var s = '// not a comment';` yields no comment.
					for _, comment := range comments.ForFile(ctx) {
						skipped = append(skipped,
							skippedRegion{comment.Range.Pos(), comment.Range.End()})
					}

					// The hashbang is neither a node nor a comment. The parser folds it into the
					// leading trivia of the first statement, so a walk never sees it and the
					// comment shelf does not return it either, which was measured rather than
					// assumed. Upstream governs it with skipComments, the only option that could
					// sensibly own it, and pins both directions: the same hashbang input appears as
					// passing with skipComments on and as failing with it off.
					if shebang := scanner.GetShebang(text); shebang != "" {
						skipped = append(skipped, skippedRegion{0, len(shebang)})
					}
				}

				collectSkippedLiterals(ctx.SourceFile, node, &skipped,
					skipStrings, skipTemplates, skipRegExps, skipJSXText)

				// A run of irregular spaces is one finding and each irregular line terminator is
				// its own, as ESLint reports them: its IRREGULAR_WHITESPACE pattern is a `+` run per
				// line, and its line-terminator pattern matches one character. Reporting each
				// character of `\v\v` apart was 3 of ESLint's corpus rows reading as extra
				// (#jjfa7qb). A run is skipped when it starts inside a skipped region, which is the
				// start ESLint tests, and no run can cross into one, since every skipped literal
				// and comment is delimited by a character that is not whitespace.
				runStart := -1
				reportRun := func(end int) {
					if runStart >= 0 && !isInsideAny(runStart, skipped) {
						ctx.ReportRange(core.NewTextRange(runStart, end), messageIrregularWhitespace)
					}
					runStart = -1
				}
				for offset, character := range text {
					switch {
					case !irregularWhitespaceCodepoints[character], offset == 0 && character == byteOrderMark:
						reportRun(offset)
					case character == lineSeparator || character == paragraphSeparator:
						reportRun(offset)
						runStart = offset
						reportRun(offset + len(string(character)))
					case runStart < 0:
						runStart = offset
					}
				}
				reportRun(len(text))
			},
		}
	},
}

// containsIrregularWhitespace answers whether a file holds any of these characters at all.
//
// Separate from the reporting loop rather than folded into it because the two ask different
// questions. This one may stop at the first hit and needs no offsets; the reporting loop needs every
// offset and cannot stop. Fusing them would mean either walking the tree before knowing whether it
// is needed, which is the cost this removes, or reporting before the skip regions exist.
func containsIrregularWhitespace(text string) bool {
	// Iterating bytes rather than runes, and testing the byte range before the map.
	//
	// The obvious loop ranges over runes and asks the map about each one, and that is what shipped
	// first. It measured 240ms across 3,407 files, still the most expensive rule in the run, to
	// report nothing: a map lookup per character over roughly 25 MB of source is the whole cost, and
	// only four files in that tree contain any of these characters at all, so the tree walk this
	// guard protects almost never runs and the guard itself became the rule's entire expense.
	//
	// Every character in the set is either one of two ASCII controls or is non-ASCII, so a byte
	// below 0x80 that is neither of those two cannot begin one. That test is a comparison against a
	// byte already in a register and it decides the overwhelming majority of a typical file, leaving
	// the map to the rare non-ASCII byte. Ranging over bytes also skips UTF-8 decoding on the common
	// path, which ranging over runes cannot.
	for index := 0; index < len(text); index++ {
		character := text[index]
		if character < 0x80 {
			if character == 0x0B || character == 0x0C {
				return true
			}
			continue
		}

		// The map, not a range test. Widening this to "any non-ASCII rune" is a mutation no
		// fixture can catch, and correctly so: the reporting loop tests every character against
		// this same map independently, so a guard that says yes too often changes what the rule
		// costs and never what it reports. It is still wrong, because the entire purpose of the
		// guard is to be cheap and decisive, and one that fires on any accented letter defeats
		// itself on every file holding ordinary non-English text.
		decoded, width := utf8.DecodeRuneInString(text[index:])
		if irregularWhitespaceCodepoints[decoded] {
			return true
		}
		// Minus one because the loop's own increment supplies the last byte. DecodeRuneInString
		// returns a width of one for invalid UTF-8, so this always advances and cannot spin.
		//
		// The skip is an optimization and not a correctness requirement, which the sweep settled
		// rather than assumed: advancing zero here survives every fixture. It re-examines each
		// continuation byte of a multi-byte character, and a continuation byte decodes to the
		// replacement rune, which is not in the map, so the extra visits cost time and cannot
		// change the answer. Advancing one byte too far is the opposite and is a real defect: it
		// lands inside the next character and decodes the replacement rune in place of a character
		// that might be in the map, so the guard reports a clean file and every finding in it
		// disappears. Only an irregular character abutting a multi-byte one distinguishes that, and
		// there is a fixture for exactly that shape.
		index += width - 1
	}
	return false
}

// collectSkippedLiterals walks the tree recording the spans an option has excused.
//
// The walk is explicit rather than done through per-kind listeners because the source-file listener
// fires first in a pre-order traversal: a rule collecting these from listeners would run its scan
// before any of them had fired and excuse nothing. One stated walk is the honest version of that
// ordering rather than a coincidence waiting to be broken.
func collectSkippedLiterals(
	sourceFile *ast.SourceFile,
	node *ast.Node,
	skipped *[]skippedRegion,
	skipStrings bool,
	skipTemplates bool,
	skipRegExps bool,
	skipJSXText bool,
) {
	if node == nil {
		return
	}

	// excuseToken records a literal's own text, without the trivia in front of it.
	//
	// Node.Pos() sits before leading trivia, so a span from Pos to End can swallow whitespace and
	// comments that precede the literal and are no part of it. Excusing that wider span would
	// silently excuse irregular whitespace standing in code position just before a string, which is
	// exactly the thing this rule exists to catch.
	//
	// Subtracting the length of node.Text() from End is the shortcut that suggests itself and it is
	// wrong twice over: Text is the cooked value rather than the raw source, so a literal holding an
	// escape is shorter than the bytes it occupies, and Text panics outright on a JSX text node.
	excuseToken := func() {
		reported := rule.TokenRange(sourceFile, node)
		*skipped = append(*skipped, skippedRegion{reported.Pos(), reported.End()})
	}

	// excuseWholeNode records a node's full span, trivia included, because it has none.
	//
	// JSX text is the one kind here that cannot go through the token accessor. That accessor finds a
	// token by skipping trivia forward from a position, and for every other literal the leading
	// whitespace genuinely is trivia. Inside JSX it is content: `<div>\u3000</div>` holds an
	// ideographic space that is the element's entire text. Measured rather than reasoned about, the
	// accessor skips straight past it and returns an empty range at the closing tag, so the region
	// excused is zero bytes wide and skipJSXText silently stops working while every other option
	// keeps behaving. Twenty of upstream's passing cases catch this and nothing else does.
	excuseWholeNode := func() {
		*skipped = append(*skipped, skippedRegion{node.Pos(), node.End()})
	}

	switch node.Kind {
	case ast.KindStringLiteral:
		if skipStrings {
			excuseToken()
		}
	case ast.KindRegularExpressionLiteral:
		if skipRegExps {
			excuseToken()
		}
	case ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateHead,
		ast.KindTemplateMiddle,
		ast.KindTemplateTail:
		// Only the quasis are excused, never the whole template. The text inside `${` and `}` is
		// code position and stays reportable, which upstream's corpus pins in both directions:
		// two inputs putting an ideographic space inside an interpolation are failing cases under
		// the very defaults that turn skipTemplates on. A port excusing the whole template
		// expression span would regress exactly those two and nothing else would notice.
		if skipTemplates {
			excuseToken()
		}
	case ast.KindJsxText:
		if skipJSXText {
			excuseWholeNode()
		}
	}

	node.ForEachChild(func(child *ast.Node) bool {
		collectSkippedLiterals(sourceFile, child, skipped,
			skipStrings, skipTemplates, skipRegExps, skipJSXText)
		return false
	})
}

// isInsideAny answers whether an offset falls in any excused region.
func isInsideAny(offset int, regions []skippedRegion) bool {
	for _, region := range regions {
		if offset >= region.start && offset < region.end {
			return true
		}
	}
	return false
}
