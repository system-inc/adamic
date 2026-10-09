package core

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUselessEscapeOptions relaxes the regex half of the rule for named characters.
//
// Only the regex half. The option is `allowRegexCharacters` and upstream's own corpus pins that
// reading: `"\#/"` configured with `["#"]` still reports, because a hash in a string literal is not
// a regex character however the option is spelled.
type NoUselessEscapeOptions struct {
	// AllowRegexCharacters names characters whose escape inside a regex is tolerated. Each entry is
	// one character; anything longer can never match a single escaped character and is inert.
	AllowRegexCharacters []string `json:"allowRegexCharacters"`
}

// NoUselessEscape flags a backslash that changes nothing about what the code means.
//
//	valid:   "\""
//	valid:   "\u00a9"
//	valid:   `\``
//	valid:   /\w\$\*\^\./
//	valid:   /[a-z-]/
//	invalid: "\'"
//	invalid: `\"`
//	invalid: /\!/
//	invalid: /[\[]/
//
// An escape that the language does not need is not a no-op to a reader: it says "this character is
// special here" about a character that is not, which is exactly the signal somebody follows when
// deciding whether a pattern is doing something subtle. Removing it is the version that means the
// same thing and reads as what it is.
//
// # Three surfaces, three different tables
//
// String literals, template literals, and regex literals each answer "which escapes are needed"
// differently, and a port that shares one table across them is wrong on every case that separates
// them. A string literal needs `\n` and its own quote character; a template needs backtick and the
// `${` sequence instead; a regex needs whichever of two dozen metacharacters is special at that
// position. `\d` is required in none of the three and mandatory in the fourth reading nobody has:
// inside a regex it is a class escape and reports nothing, inside a string it is useless and
// reports.
//
// # Position decides inside a regex, not just the character
//
// Two characters cannot be judged by the table alone.
//
// A caret must be escaped outside a character class, and inside one only at the very start, since
// that is the one position where it means negation. So `/\^bar/` and `/[\^bar]/` are both clean and
// `/[^\^]/` is not.
//
// A hyphen is the mirror image: escaping it is only meaningful inside a class and away from its
// edges, where it would otherwise open a range. `/[a\-b]/` is clean and `/[ab\-]/`, `/[\-ab]/` and
// `/\-/` are not.
//
// Under the `v` flag a third rule appears. That flag reserves nineteen punctuators as doubled
// operators, so `&&`, `!!` and their kin have syntactic meaning and escaping one half of a pair is
// how you write the literal character. `/[\&&]/v` is clean because the escape separates the pair,
// while `/[\&\&]/v` reports once: the second escape is redundant once the first has done the
// separating.
//
// # This is a fix rather than a suggestion
//
// Deleting the backslash from an escape the language did not need cannot change what the code
// means, because "the language did not need it" is precisely the predicate the rule computes. That
// is the fix-versus-suggestion line: a fix preserves meaning and the engine applies it unattended.
// It is also why the predicate has to be exactly right rather than approximately right, and why the
// fixtures assert the rewritten source rather than the finding count. A count-only fixture passes
// for a fix that writes the right character at the wrong offset, which here would eat the character
// before the backslash and leave the backslash standing.
//
// # Why this does not use regexpattern.Walk
//
// The shelf holds a pattern walker and it is the wrong shape for this rule, in three ways that its
// two shipped callers do not feel. It skips every escape that matches a set rather than a single
// character, so `\p`, `\q`, `\k` and `\B` never arrive, and each of those is a case this rule has to
// answer. It reports `ClassDepth` but not which class, and both the caret and hyphen rules need the
// class's own start and end. And it does not recurse into a `v`-flag nested class, so
// `/[[\.&]--[\.&]]/v` walks to silence where upstream reports twice, measured directly rather than
// read off the source.
//
// That package's own doc anticipates this: a third caller wanting more structure should extend it
// rather than find it half-built. Extending it would change the contract for `no-regex-spaces` and
// `no-control-regex`, which is a larger change than this rule justifies, so the escape scan here is
// built on the layer underneath instead. `regexsyntax.SkipPatternEscape` and `regexsyntax.ClassEnd`
// are the primitives that answer where an escape ends and where a class ends, `v`-flag nesting and
// multi-byte identity escapes included, and both are used rather than reimplemented.
//
// # A JSX attribute is not this rule's surface
//
// `<foo attr="\d"/>` is silent and `<foo attr={"\d"}/>` reports. The first is JSX attribute text,
// where a backslash is a literal backslash and the language never read it as an escape, so there is
// nothing useless about it. The second is a real string literal that merely sits inside JSX. The
// difference is visible as the literal's parent and nothing else.
var NoUselessEscape = rule.Rule{
	Name: "no-useless-escape",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		allowed := map[rune]bool{}
		if configured, isConfigured := rule.OptionsAs[NoUselessEscapeOptions](options); isConfigured {
			for _, entry := range configured.AllowRegexCharacters {
				// One character per entry. A longer string can never equal a single escaped
				// character, so admitting it would widen nothing and hide a typo in the configuration.
				if character, size := utf8.DecodeRuneInString(entry); size == len(entry) {
					allowed[character] = true
				}
			}
		}

		report := func(start int, end int, escaped string) {
			ctx.Report(rule.Diagnostic{
				Range:      core.NewTextRange(start, end),
				SourceFile: ctx.SourceFile,
				Message: rule.Message{
					Id: "noUselessEscape",
					Description: fmt.Sprintf(
						"The backslash before %s does nothing here: the language reads this as the "+
							"plain character either way. An escape that is not needed says the "+
							"character is special at this position when it is not, which is the "+
							"signal a reader follows when deciding whether something subtle is "+
							"happening. Write %s.", escaped, escaped),
				},
				Fixes: []rule.Fix{{Range: core.NewTextRange(start, end), Text: escaped}},
			})
		}

		reportQuoted := func(literalStart int, raw string, offsets []int) {
			for _, offset := range offsets {
				escaped, size := utf8.DecodeRuneInString(raw[offset:])
				start := literalStart + offset - 1
				report(start, literalStart+offset+size, string(escaped))
			}
		}

		return rule.Listeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				// The backslash in JSX attribute text was never an escape, so there is nothing
				// useless about it. Checked on the parent because that is the only thing that
				// separates the two spellings; the literal's own text is identical.
				if node.Parent != nil && node.Parent.Kind == ast.KindJsxAttribute {
					return
				}
				literalRange := rule.TokenRange(ctx.SourceFile, node)
				raw := ctx.SourceFile.Text()[literalRange.Pos():literalRange.End()]
				reportQuoted(literalRange.Pos(), raw, uselessStringEscapes(raw))
			},

			ast.KindNoSubstitutionTemplateLiteral: func(node *ast.Node) {
				if isTaggedTemplateBody(node) {
					return
				}
				literalRange := rule.TokenRange(ctx.SourceFile, node)
				raw := ctx.SourceFile.Text()[literalRange.Pos():literalRange.End()]
				reportQuoted(literalRange.Pos(), raw, uselessTemplateEscapes(raw))
			},

			ast.KindTemplateExpression: func(node *ast.Node) {
				if isTaggedTemplateBody(node) {
					return
				}
				for _, part := range templateTextParts(node) {
					partRange := rule.TokenRange(ctx.SourceFile, part)
					raw := ctx.SourceFile.Text()[partRange.Pos():partRange.End()]
					reportQuoted(partRange.Pos(), raw, uselessTemplateEscapes(raw))
				}
			},

			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				text := node.Text()
				pattern, flags := regexsyntax.PatternAndFlags(text)
				if pattern == "" {
					return
				}
				// Pos() sits before leading trivia, so the literal's start is derived from its end.
				// Same reasoning as `no-control-regex`: reading Pos() puts the finding on whatever
				// comment precedes the literal, which no suppressing directive can reach.
				patternStart := node.End() - len(text) + 1

				for _, escape := range uselessPatternEscapes(pattern, regexsyntax.ParseRegexFlags(flags)) {
					escaped, size := utf8.DecodeRuneInString(pattern[escape+1:])
					if allowed[escaped] {
						continue
					}
					report(patternStart+escape, patternStart+escape+1+size, string(escaped))
				}
			},
		}
	},
}

// isTaggedTemplateBody reports whether a template literal is the quasi of a tagged template.
//
// A tag receives the raw strings alongside the cooked ones, so `String.raw`\.“ is code that reads
// the backslash deliberately and every escape in it is load-bearing to somebody. Upstream is silent
// on all of them and so is this, including for tags it cannot resolve: `myFunc`\.“ passes too,
// because whether the tag looks at `raw` is not a question the syntax can answer.
//
// # The quasi, not merely a child
//
// Testing the parent's kind alone is wrong, and a fixture is what said so. A template literal can be
// the tag of a tagged template rather than its quasi: the input "`\\a```" parses as the literal
// "`\\a`" tagging the empty literal "“", and both are children of the same
// TaggedTemplateExpression. A tag position is an ordinary expression whose escapes are cooked
// normally, so the escape in it is as useless as it would be anywhere else, and upstream reports it.
// Suppressing on parent kind made that one input silent while the other 292 stayed green, which is
// the shape of a discrimination only one case in the corpus can see.
//
// Upstream states the same test as a span comparison against the tagged expression's quasi. Pointer
// identity says it here, because the two nodes are the same node when they match.
func isTaggedTemplateBody(node *ast.Node) bool {
	if node.Parent == nil || node.Parent.Kind != ast.KindTaggedTemplateExpression {
		return false
	}
	tagged := node.Parent.AsTaggedTemplateExpression()
	return tagged != nil && tagged.Template == node
}

// templateTextParts returns the literal chunks of a template expression, head through tail.
//
// The substitutions between them are ordinary expressions and hold no escapes of their own. The
// parts arrive with their delimiters attached, so the head carries its opening backtick and its
// trailing `${` and a middle carries `}` and `${`; the scan below relies on that, because a
// backslash immediately before a `${` behaves differently from one anywhere else.
func templateTextParts(node *ast.Node) []*ast.Node {
	expression := node.AsTemplateExpression()
	if expression == nil {
		return nil
	}

	parts := []*ast.Node{expression.Head}
	if expression.TemplateSpans == nil {
		return parts
	}
	for _, span := range expression.TemplateSpans.Nodes {
		templateSpan := span.AsTemplateSpan()
		if templateSpan == nil {
			continue
		}
		parts = append(parts, templateSpan.Literal)
	}
	return parts
}

// validStringEscapes are the characters after a backslash that a string or template needs.
//
// The line terminators are in here because a backslash before one is a line continuation, which is
// the one escape in a string literal that removes a character rather than naming one. The two
// exotic ones are the Unicode line and paragraph separators, which the language treats as line
// terminators too.
const validStringEscapes = "\\nrvtbfux\n\r\u2028\u2029"

// uselessStringEscapes returns the offsets, within a quoted literal's raw text, of each escaped
// character whose backslash does nothing.
//
// The offset points at the escaped character rather than at its backslash, which is what lets the
// caller decode a multi-byte one without re-finding it.
//
// Two things make this harder than scanning for backslashes. The literal's own quote character is a
// required escape and the other quote is not, so the same two bytes are clean in `"\'"` written one
// way and a finding written the other, and the quote has to be read from the text rather than
// assumed. And a doubled backslash is one escaped backslash rather than the start of another
// escape, so parity has to come out right or every `"\\a"` reports the `a` that follows an escape
// that already consumed its backslash.
func uselessStringEscapes(raw string) []int {
	if len(raw) <= 1 {
		return nil
	}

	quote, quoteSize := utf8.DecodeRuneInString(raw)
	// The closing quote is trimmed for readability rather than for correctness, and a sweep is what
	// established the difference: including it changed no fixture. It cannot, because the only extra
	// position the scan would visit is that quote, and a backslash can never sit there. A backslash
	// immediately before the closing quote would have escaped it, in which case the quote was not
	// closing and the literal did not end where it appears to.
	body := raw[quoteSize : len(raw)-1]

	var offsets []int
	for index := 0; index < len(body); index++ {
		if body[index] != '\\' {
			continue
		}
		if index+1 >= len(body) {
			break
		}

		escaped, size := utf8.DecodeRuneInString(body[index+1:])
		// A digit after a backslash is a legacy octal or a null escape. Neither is useless in the
		// sense this rule means, and `no-octal` is the rule that has an opinion about them.
		isNeeded := escaped == quote ||
			(escaped >= '0' && escaped <= '9') ||
			strings.ContainsRune(validStringEscapes, escaped)
		if !isNeeded {
			offsets = append(offsets, quoteSize+index+1)
		}
		// Consuming the escaped character is what makes the parity right: the second backslash of
		// `\\` can never be read as opening an escape of its own.
		index += size
	}
	return offsets
}

// uselessTemplateEscapes is uselessStringEscapes for one literal chunk of a template.
//
// A template answers the question differently in three places and each is a real case upstream
// pins. A backtick always needs its escape, since it would otherwise end the literal. A `$` needs
// one only when a `{` follows, because `${` is what opens a substitution and a lone dollar is
// ordinary text. And a `{` needs one only when a `$` precedes it, which is the same rule read from
// the other end: “ `\${{${foo}` “ and “ `$\{{${foo}` “ are both clean, and each escapes a
// different half of the same pair.
//
// So this walks forward tracking the previous unescaped character rather than testing membership in
// a table, and the offsets it returns are relative to the chunk's raw text including its delimiters.
func uselessTemplateEscapes(raw string) []int {
	if len(raw) <= 1 {
		return nil
	}

	var offsets []int
	// The delimiter is skipped rather than examined: a head opens with a backtick and a middle or
	// tail opens with `}`, and neither is content. Seeding the previous character with a backtick
	// matches upstream and makes a leading `\{` report, since no `$` can precede the first
	// character of a chunk.
	previous := '`'
	// Both bounds are the delimiters rather than content, and neither is load-bearing: a sweep
	// widening either end changed no fixture. A chunk always opens with a backtick or a closing brace
	// and always ends with a backtick or the brace of `${`, so neither position can be a backslash
	// and visiting them would only run the loop's non-escape path. They are stated because a reader
	// asking which bytes are content should find the answer here rather than derive it.
	for index := 1; index < len(raw)-1; {
		current, size := utf8.DecodeRuneInString(raw[index:])
		if current != '\\' {
			previous = current
			index += size
			continue
		}

		if index+size >= len(raw)-1 {
			break
		}
		escaped, escapedSize := utf8.DecodeRuneInString(raw[index+size:])

		useless := false
		switch {
		case escaped >= '0' && escaped <= '9', escaped == '`':
			// A backtick always needs escaping and a digit is somebody else's rule.
		case escaped == '{':
			useless = previous != '$'
		case escaped == '$':
			following, _ := utf8.DecodeRuneInString(raw[index+size+escapedSize:])
			useless = following != '{'
		default:
			useless = !strings.ContainsRune(validStringEscapes, escaped)
		}
		if useless {
			offsets = append(offsets, index+size)
		}

		// The escaped character becomes the previous one, which is what makes `$\{` clean: the
		// dollar was written unescaped and the brace's escape is therefore load-bearing.
		previous = escaped
		index += size + escapedSize
	}
	return offsets
}

// The four tables are the whole of the regex half, and which one applies depends on where the
// escape sits and which flags are on.
//
// Read as "escapes that are needed here". Outside a class every metacharacter is in play, so the
// list is longest. Inside an ordinary class most of them stop being special and drop out, which is
// why `/\(bar\)/` is clean and `/[\(paren]/` is not. Under the `v` flag a class is a set expression
// with its own operators, so the list changes again rather than shrinking further.
const (
	regexGeneralEscapes               = "\\bcdDfnpPrsStvwWxu0123456789]"
	regexNonClassEscapes              = "\\bcdDfnpPrsStvwWxu0123456789]^/.$*+?[{}|()Bk"
	regexClassSetEscapes              = "\\bcdDfnpPrsStvwWxu0123456789]q/[{}|()-"
	regexClassSetReservedDoublePuncts = "!#$%&*+,.:;<=>?@^`~"
)

// classSpan is one character class's extent within a pattern.
type classSpan struct {
	start    int
	end      int
	negative bool
}

// uselessPatternEscapes returns the offset of each backslash in a pattern whose escape does nothing.
//
// The scan is written here rather than on `regexpattern.Walk` for the reasons on the rule, and it is
// built from the layer that walker uses: `SkipPatternEscape` says how far an escape reaches,
// including a `\p{...}` property or a multi-byte identity escape like `\（`, and `ClassEnd` says
// where a class stops, `v`-flag nesting included.
func uselessPatternEscapes(pattern string, flags regexsyntax.RegexFlags) []int {
	var offsets []int
	scanEscapes(pattern, flags, nil, &offsets)
	return offsets
}

// scanEscapes walks one region of a pattern, recursing into each class it meets.
//
// The enclosing class is carried down rather than looked up, because the caret and hyphen rules both
// ask where this escape sits relative to *its own* class's brackets. Under the `v` flag classes
// nest, and the innermost is the one that answers: in `/[[\.&]--[\.&]]/v` each dot sits at the start
// of an inner class, not of the outer one.
func scanEscapes(pattern string, flags regexsyntax.RegexFlags, enclosing *classSpan, offsets *[]int) {
	end := len(pattern)
	start := 0
	if enclosing != nil {
		start = enclosing.start + 1
		end = enclosing.end - 1
		if enclosing.negative {
			start++
		}
	}

	for index := start; index < end; {
		switch pattern[index] {
		case '\\':
			step, ok := regexsyntax.SkipPatternEscape(pattern, index, flags)
			if !ok {
				return
			}
			if isUselessPatternEscape(pattern, index, flags, enclosing) {
				*offsets = append(*offsets, index)
			}
			index += step

		case '[':
			classEnd, ok := regexsyntax.ClassEnd(pattern, index, flags)
			if !ok {
				return
			}
			negative := index+1 < classEnd && pattern[index+1] == '^'
			scanEscapes(pattern, flags, &classSpan{start: index, end: classEnd, negative: negative}, offsets)
			index = classEnd

		default:
			_, width := utf8.DecodeRuneInString(pattern[index:])
			if width == 0 {
				width = 1
			}
			index += width
		}
	}
}

// isUselessPatternEscape decides one escape, given the class it sits in or nil for none.
func isUselessPatternEscape(pattern string, index int, flags regexsyntax.RegexFlags, enclosing *classSpan) bool {
	if index+1 >= len(pattern) {
		return false
	}
	escaped, size := utf8.DecodeRuneInString(pattern[index+1:])

	needed := regexNonClassEscapes
	if enclosing != nil {
		needed = regexGeneralEscapes
		if flags.UnicodeSets {
			needed = regexClassSetEscapes
		}
	}
	// No guard against a non-ASCII escaped character before the lookup, because the lookup already
	// answers it. All three tables are ASCII, and ContainsRune of an ASCII-only string is false for
	// every rune above 127, so a guard here would be a branch that can only agree with the line after
	// it. A sweep found it that way rather than a reading: dropping it changed no fixture, and the
	// tables are what make that true rather than a coincidence in the corpus.
	if strings.ContainsRune(needed, escaped) {
		return false
	}

	if enclosing == nil {
		return true
	}

	if escaped == '^' {
		// A caret means negation only as a class's first character, so escaping it there is the
		// only way to write a literal caret in that one position.
		if enclosing.start+1 == index {
			return false
		}
	}

	if flags.UnicodeSets {
		return !isReservedDoublePunctuatorEscape(pattern, index, index+1+size, escaped, enclosing)
	}

	if escaped == '-' {
		// A hyphen opens a range only away from a class's edges, so escaping it is meaningful
		// exactly there and pointless at either end.
		if enclosing.start+1 != index && index+1+size != enclosing.end-1 {
			return false
		}
	}
	return true
}

// isReservedDoublePunctuatorEscape reports whether a `v`-flag escape is separating a doubled
// operator, which is the one thing that makes escaping one of those punctuators necessary.
//
// The `v` flag reserves nineteen punctuators as doubled operators, so `&&` and `!!` are syntax and
// the way to write the literal character beside its twin is to escape one of them. Which one does
// not matter, so both `/[\&&]/v` and `/[&\&]/v` are clean while `/[\&\&]/v` reports once: the pair
// is already separated by the first escape and the second adds nothing.
//
// The caret carries an extra condition, and it is the reason this cannot be a symmetric test. A
// caret at the very start of a negated class is the negation rather than a member, so a `^` escaped
// immediately after it is not separating a pair at all and stays a finding.
func isReservedDoublePunctuatorEscape(pattern string, escapeStart int, escapeEnd int, escaped rune, enclosing *classSpan) bool {
	if !strings.ContainsRune(regexClassSetReservedDoublePuncts, escaped) {
		return false
	}

	// The twin may follow the escape.
	if escapeEnd < len(pattern) {
		if following, _ := utf8.DecodeRuneInString(pattern[escapeEnd:]); following == escaped {
			return true
		}
	}

	// Or precede the backslash. Decoded backwards rather than by subtracting one, because a
	// multi-byte character before the escape would otherwise be read as its last byte.
	//
	// A sweep replacing this decode with a plain byte read survived, and it survived because it
	// cannot be caught rather than because the fixtures are thin. Distinguishing the two would need a
	// multi-byte character sitting immediately before an escaped reserved punctuator whose final
	// UTF-8 byte equals that punctuator, and no such character exists: every punctuator in the table
	// is at most 0x7E and every UTF-8 continuation byte is at least 0x80, so the two ranges do not
	// meet. The size it returns reaches the caret arithmetic below, and only on a path where the
	// preceding character has already been shown equal to a caret, so it is one there in every case.
	//
	// The decode stays because it states the intent rather than depending on that arithmetic, and a
	// later table gaining a non-ASCII member would silently break the byte read.
	preceding, size := utf8.DecodeLastRuneInString(pattern[:escapeStart])
	if size == 0 || preceding != escaped {
		return false
	}
	if escaped != '^' {
		return true
	}

	// A caret preceded by a caret is only a real pair when the earlier one is a member rather than
	// the class's negation, so `/[^\^]/v` reports and `/[^^\^]/v` does not.
	if !enclosing.negative {
		return true
	}
	return enclosing.start+1 < escapeStart-size
}
