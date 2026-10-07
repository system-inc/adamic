// Raw JavaScript regexp syntax substrate, algorithm copied unchanged from pinned cohere.
// See stage1/cohere/typeaware/NOTICE for source provenance and licenses.
package regexp

import (
	"errors"
	"fmt"
	"github.com/dlclark/regexp2/v2"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// MatchTimeout bounds how long one subject may spend in one pattern. A pattern
// written by a user can backtrack exponentially , `(a|aa)+$` against a long
// run of `a` is the shape of it , and a linter that hangs on a config typo is
// worse than one that answers no. A pattern that matches settles in
// microseconds, so a match still running after this has found nothing.
const MatchTimeout = time.Second

// ErrUnsupportedFlag is returned for a flag this package does not implement.
var ErrUnsupportedFlag = errors.New("unsupported regexp flag")

// ErrUnsupportedSyntax is returned for syntax JavaScript itself rejects.
var ErrUnsupportedSyntax = errors.New("unsupported regexp syntax")

// RegExp is a compiled JavaScript regexp.
type RegExp struct {
	source string
	flags  string
	re     *regexp2.Regexp
	// exact is false when the pattern fell back to regexp2's own
	// case-insensitivity, which is close to JavaScript's but not identical.
	exact bool
}

type flagSet struct {
	ignoreCase bool
	multiline  bool
	dotAll     bool
	unicode    bool
	sticky     bool
}

// parseFlags reads a JavaScript flag string. `g` and `d` say how a match is
// driven rather than what matches, so they are accepted and ignored. `y` does
// change what matches and is kept.
func parseFlags(flags string) (flagSet, error) {
	var set flagSet
	seen := map[rune]bool{}
	for _, flag := range flags {
		if seen[flag] {
			return set, fmt.Errorf("%w: %q given twice", ErrUnsupportedFlag, flag)
		}
		seen[flag] = true
		switch flag {
		case 'i':
			set.ignoreCase = true
		case 'm':
			set.multiline = true
		case 's':
			set.dotAll = true
		case 'u':
			set.unicode = true
		case 'y':
			set.sticky = true
		case 'g', 'd':
		case 'v':
			return set, fmt.Errorf("%w: %q", ErrUnsupportedFlag, flag)
		default:
			return set, fmt.Errorf("%w: %q", ErrUnsupportedFlag, flag)
		}
	}
	return set, nil
}

// Compile compiles source under the given JavaScript flags, as
// `new RegExp(source, flags)` would. Pass "" for no flags.
func Compile(source string, flags string) (*RegExp, error) {
	set, err := parseFlags(flags)
	if err != nil {
		return nil, err
	}

	rewritten, exact, err := rewrite(source, rewriteOptions{
		multiline:  set.multiline,
		dotAll:     set.dotAll,
		ignoreCase: set.ignoreCase,
		unicode:    set.unicode,
	})
	if err != nil {
		return nil, err
	}
	if set.sticky {
		// A sticky regexp matches only at `lastIndex`, which is 0 for one just
		// compiled and never moves here, since nothing this package exposes
		// advances it. So it anchors at the start , around the whole pattern,
		// or a top-level `|` would anchor only its first branch.
		rewritten = inputStart + `(?:` + rewritten + `)`
	}

	// Multiline, Singleline and IgnoreCase are all handled by the rewrite, so
	// regexp2 is never told about them , being told twice would undo the work.
	options := regexp2.RegexOptions(regexp2.ECMAScript)
	if set.unicode {
		options |= regexp2.Unicode
	}
	if set.ignoreCase && !exact {
		options |= regexp2.IgnoreCase
	}

	re, err := regexp2.Compile(rewritten, options)
	if err != nil {
		// regexp2 quotes back the pattern it was handed, which is the
		// rewritten one. A message naming `[Oo][Kk]` where its author wrote
		// `ok` is no help, so the source goes back in as written.
		return nil, errors.New(strings.ReplaceAll(err.Error(), rewritten, source))
	}
	re.MatchTimeout = MatchTimeout

	return &RegExp{source: source, flags: flags, re: re, exact: exact}, nil
}

// MustCompile is Compile for a pattern written in this repository rather than
// read from a user, and panics rather than returning an error.
func MustCompile(source string, flags string) *RegExp {
	re, err := Compile(source, flags)
	if err != nil {
		panic(fmt.Sprintf("esregexp: cannot compile /%s/%s: %v", source, flags, err))
	}
	return re
}

// Test reports whether the regexp matches anywhere in s, as
// `RegExp.prototype.test` does.
//
// A match that overruns MatchTimeout counts as no match: the alternative is
// making every caller handle an error that only ever means "this pattern is
// pathological", and a lint rule has nothing better to do with that than skip.
func (r *RegExp) Test(s string) bool {
	if r == nil || r.re == nil {
		return false
	}
	matched, err := r.re.MatchString(s)
	return err == nil && matched
}

// TestOrTimeout is Test for a caller whose no answer is what produces a
// report: an allow-pattern, an ignore-pattern, a pattern a name is required to
// match. Reading a bound match as no match would turn a skipped report into a
// false positive there, so a match that overruns MatchTimeout counts as a
// match and the report is skipped.
func (r *RegExp) TestOrTimeout(s string) bool {
	matched, err := r.TestOrError(s)
	return matched || err != nil
}

// TestOrError is Test for a caller that has to tell "no match" apart from "the
// match ran into MatchTimeout", and that has somewhere other than a match to
// send the second answer.
func (r *RegExp) TestOrError(s string) (bool, error) {
	if r == nil || r.re == nil {
		return false, nil
	}
	return r.re.MatchString(s)
}

// Source returns the pattern as it was written, the way `RegExp.prototype
// .source` does , not the rewritten form handed to regexp2.
func (r *RegExp) Source() string { return r.source }

// Flags returns the flags as they were written.
func (r *RegExp) Flags() string { return r.flags }

// String renders the regexp the way JavaScript writes a regexp literal.
func (r *RegExp) String() string { return "/" + r.source + "/" + r.flags }

// Unwrap returns the compiled regexp2 pattern, for a caller that needs
// capture groups or replacement. The pattern it holds is the rewritten one, so
// its own String does not match Source.
func (r *RegExp) Unwrap() *regexp2.Regexp {
	if r == nil {
		return nil
	}
	return r.re
}

// lineTerminators spells the characters ECMAScript reads as ending a line, in
// the form regexp2 parses inside a character class. It has to be the `\uXXXX`
// form: regexp2 reads neither `\x{2028}` nor a bare `\u{2028}`.
const lineTerminators = `\n\r\u2028\u2029`

const (
	// nonTerminator stands in for a `.`, which in JavaScript matches anything
	// but a line terminator. regexp2 follows .NET, where only `\n` is left out.
	nonTerminator = `[^` + lineTerminators + `]`
	// anyCharacter stands in for a `.` under the `s` flag, and for `[^]`.
	anyCharacter = `[\s\S]`
	// neverMatches stands in for `[]`, the class JavaScript reads as matching
	// nothing at all.
	neverMatches = `(?!)`
	// inputStart and inputEnd are the anchors a pattern without `m` wants:
	// the very start and the very end, where regexp2's own `^` and `$` would
	// also stop at a newline.
	inputStart = `\A`
	inputEnd   = `\z`
	// lineStart and lineEnd are the anchors a pattern with `m` wants.
	// regexp2's Multiline breaks on `\n` alone, where JavaScript breaks on all
	// four terminators, so the break is spelled out as a lookaround.
	lineStart = `(?:\A|(?<=[` + lineTerminators + `]))`
	lineEnd   = `(?:\z|(?=[` + lineTerminators + `]))`
)

// wordCharacters is the set ECMAScript builds `\w` and a word boundary out of.
// It is ASCII whatever the flags say, where .NET , and so regexp2 , reaches
// for a Unicode word set, which is why both have to be spelled out.
func wordCharacters(options rewriteOptions) string {
	return writeClass(wordClassAtoms(options), false, rewriteOptions{})
}

// wordBoundary writes out `\b`, or `\B` when negated, as the pair of
// lookarounds ECMAScript defines it as, so the boundary lands where JavaScript
// puts it rather than where .NET's wider word set would.
func wordBoundary(negated bool, options rewriteOptions) string {
	word := wordCharacters(options)
	if negated {
		return `(?:(?<=` + word + `)(?=` + word + `)|(?<!` + word + `)(?!` + word + `))`
	}
	return `(?:(?<!` + word + `)(?=` + word + `)|(?<=` + word + `)(?!` + word + `))`
}

// literalRune writes r so that regexp2 reads it as the character itself,
// whatever meaning that character carries on its own. A letter or a digit is
// written as it stands; everything else takes the `\uXXXX` form, which is what
// regexp2 parses both inside a character class and out.
func literalRune(r rune) string {
	switch {
	case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		return string(r)
	case r > 0xFFFF:
		// Past what `\uXXXX` names, and nothing up there carries a meaning of
		// its own to escape.
		return string(r)
	}
	return fmt.Sprintf(`\u%04x`, r)
}

// rewriteOptions says which of a JavaScript regexp's flags the rewrite has to
// make up for, because regexp2 reads them differently or not at all.
type rewriteOptions struct {
	multiline  bool
	dotAll     bool
	ignoreCase bool
	unicode    bool
}

// groupKind tells the groups whose closing parenthesis a quantifier may follow
// apart from the ones it may not.
type groupKind uint8

const (
	groupPlain groupKind = iota
	groupLookahead
	groupLookbehind
)

// openGroup is one `(` the walk has not yet seen the `)` of. options is what to
// restore as it closes; only a modifier group makes an entry differ from the
// one below.
type openGroup struct {
	options rewriteOptions
	kind    groupKind
}

// rewrite translates a JavaScript regexp source into one regexp2 reads the
// same way. It walks the source rather than running a regexp over it, because
// what a character means depends on whether it is escaped and whether it sits
// inside a character class.
//
// The second result is false when the source uses a backreference or a
// property escape under `i`. JavaScript compares a backreference by the same
// canonicalization it compares a literal by, and it draws `\p{…}` from
// Unicode's own tables; a widened pattern can name neither, so the caller falls
// back to regexp2's own case-insensitivity there, which is close but not exact.
func rewrite(source string, options rewriteOptions) (string, bool, error) {
	var out strings.Builder
	out.Grow(len(source))

	exact := true
	// current is what the flags say here, which a modifier group changes for
	// the span of one group.
	current := options
	groups := []openGroup{}
	// How a `\1` and a `\k` read is settled by the pattern as a whole, so the
	// groups are counted before the walk rather than as it goes.
	groupCount, named := countGroups(source)
	context := func() escapeContext {
		return escapeContext{unicode: current.unicode, groups: groupCount, named: named}
	}

	for i := 0; i < len(source); {
		r, size := utf8.DecodeRuneInString(source[i:])

		switch r {
		case '\\':
			// `\k<name>` names a group. The name is not text to be matched,
			// so widening the letters in it would rewrite the reference.
			if name, nameSize, ok := namedBackreference(source[i:]); ok {
				if current.ignoreCase {
					exact = false
				}
				out.WriteString(name)
				i += nameSize
				continue
			}
			escape, err := decodeEscape(source, i+size, context())
			if err != nil {
				return "", false, err
			}
			end := i + size + escape.width

			switch escape.kind {
			case escapeRune:
				// An escape that resolves to a character is written as that
				// character. Passed through as written, one .NET reads
				// differently , `\A`, `\a`, `\e` , would keep its .NET meaning,
				// and one .NET does not know at all would refuse to compile.
				if current.ignoreCase {
					if class, widened := CaseClass(escape.r, current.unicode); widened {
						out.WriteString(class)
						i = end
						continue
					}
				}
				out.WriteString(literalRune(escape.r))

			case escapeBackreference:
				if current.ignoreCase {
					exact = false
				}
				out.WriteString(source[i:end])

			case escapeAssertion:
				// A pattern cannot repeat a position, and lowering the
				// assertion first would leave regexp2 reading syntax
				// JavaScript rejects as something it accepts.
				if quantifierWidth(source[end:]) > 0 {
					return "", false, errNothingToRepeat(source[end:])
				}
				out.WriteString(wordBoundary(escape.negated, current))

			case escapeSet:
				switch {
				// `\d` and `\s` regexp2 already reads as ECMAScript does, and
				// `\w` too until `u` and `i` together widen the set past ASCII.
				case escape.set == setWord && current.ignoreCase && current.unicode:
					out.WriteString(writeClass(wordClassAtoms(current), false, rewriteOptions{}))
				case escape.set == setNonWord && current.ignoreCase && current.unicode:
					out.WriteString(writeClass(wordClassAtoms(current), true, rewriteOptions{}))
				default:
					// A property escape names a set out of Unicode's tables,
					// which a widened pattern has no way to name back.
					if escape.set == setProperty && current.ignoreCase {
						exact = false
					}
					out.WriteString(source[i:end])
				}
			}
			i = end
			continue

		case '[':
			body, negated, width, err := readClass(source[i:])
			if err != nil {
				return "", false, err
			}
			atoms, classExact, err := classAtoms(body, current, context())
			if err != nil {
				return "", false, err
			}
			if !classExact {
				exact = false
			}
			out.WriteString(writeClass(atoms, negated, current))
			i += width
			continue

		case '(':
			// `(?<name>` opens a named group. Same as `\k<name>`: the name is
			// not text, so it goes through untouched.
			if opener, openerSize, ok := namedGroupOpener(source[i:]); ok {
				groups = append(groups, openGroup{options: current})
				out.WriteString(opener)
				i += openerSize
				continue
			}
			// `(?i-m:…)` turns a flag on or off over one group. The rewrite
			// answers to the flags as they stand here, so the group opens as a
			// plain one and the walk goes on under what it says.
			if inside, openerSize, ok := modifierGroup(source[i:], current); ok {
				groups = append(groups, openGroup{options: current})
				current = inside
				if options.ignoreCase && !inside.ignoreCase {
					// A pattern that falls back hands ignoring case to regexp2
					// for the whole of itself, and this group is the one place
					// it must not reach.
					out.WriteString(`(?-i:`)
				} else {
					out.WriteString(`(?:`)
				}
				i += openerSize
				continue
			}
			if err := checkGroupConstruct(source[i:]); err != nil {
				return "", false, err
			}
			groups = append(groups, openGroup{options: current, kind: groupKindOf(source[i:])})
			out.WriteByte('(')
			i += size
			continue

		case ')':
			if last := len(groups) - 1; last >= 0 {
				group := groups[last]
				groups = groups[:last]
				current = group.options
				if err := checkGroupQuantifier(group.kind, source[i+size:], options); err != nil {
					return "", false, err
				}
			}
			out.WriteByte(')')
			i += size
			continue

		case '.':
			if current.dotAll {
				out.WriteString(anyCharacter)
			} else {
				out.WriteString(nonTerminator)
			}
			i += size
			continue

		case '^':
			if quantifierWidth(source[i+size:]) > 0 {
				return "", false, errNothingToRepeat(source[i+size:])
			}
			if current.multiline {
				out.WriteString(lineStart)
			} else {
				out.WriteString(inputStart)
			}
			i += size
			continue

		case '$':
			if quantifierWidth(source[i+size:]) > 0 {
				return "", false, errNothingToRepeat(source[i+size:])
			}
			if current.multiline {
				out.WriteString(lineEnd)
			} else {
				out.WriteString(inputEnd)
			}
			i += size
			continue

		default:
			if current.ignoreCase {
				if class, widened := CaseClass(r, current.unicode); widened {
					out.WriteString(class)
					i += size
					continue
				}
			}
			out.WriteRune(r)
			i += size
		}
	}

	return out.String(), exact, nil
}

func errNothingToRepeat(quantifier string) error {
	return fmt.Errorf("%w: %s repeats an assertion", ErrUnsupportedSyntax, quantifier[:quantifierWidth(quantifier)])
}

// quantifierWidth reports how many bytes of source a quantifier spans, or none
// where no quantifier opens it. A `{` that no bound closes is a `{` standing
// for itself, which Annex B allows and which repeats nothing.
func quantifierWidth(source string) int {
	width := 0
	switch {
	case source == "":
		return 0
	case source[0] == '*', source[0] == '+', source[0] == '?':
		width = 1
	case source[0] == '{':
		if width = boundedQuantifierWidth(source); width == 0 {
			return 0
		}
	default:
		return 0
	}
	// A lazy quantifier is still a quantifier.
	if width < len(source) && source[width] == '?' {
		width++
	}
	return width
}

// boundedQuantifierWidth reads `{m}`, `{m,}` or `{m,n}`.
func boundedQuantifierWidth(source string) int {
	i := len("{")
	digits := i
	for isDecimalDigit(source, i) {
		i++
	}
	if i == digits {
		return 0
	}
	if i < len(source) && source[i] == ',' {
		i++
		for isDecimalDigit(source, i) {
			i++
		}
	}
	if i < len(source) && source[i] == '}' {
		return i + 1
	}
	return 0
}

// groupKindOf reads whether a group is a lookaround, which is what decides
// whether a quantifier may follow its close.
func groupKindOf(source string) groupKind {
	switch {
	case strings.HasPrefix(source, "(?="), strings.HasPrefix(source, "(?!"):
		return groupLookahead
	case strings.HasPrefix(source, "(?<="), strings.HasPrefix(source, "(?<!"):
		return groupLookbehind
	}
	return groupPlain
}

// checkGroupQuantifier rejects a quantifier repeating a group that matches no
// characters. Annex B keeps one exception: without `u`, a lookahead may be
// quantified, and matches its body once or not at all. A lookbehind never may.
func checkGroupQuantifier(kind groupKind, rest string, options rewriteOptions) error {
	if kind == groupPlain || quantifierWidth(rest) == 0 {
		return nil
	}
	if kind == groupLookahead && !options.unicode {
		return nil
	}
	return errNothingToRepeat(rest)
}

// namedGroupOpener matches `(?<name>` at the start of source, returning it
// whole. A `(?<=` or `(?<!` is a lookbehind rather than a name, and carries
// text that does get widened, so it is left to the ordinary walk.
func namedGroupOpener(source string) (string, int, bool) {
	if !strings.HasPrefix(source, "(?<") {
		return "", 0, false
	}
	rest := source[len("(?<"):]
	if strings.HasPrefix(rest, "=") || strings.HasPrefix(rest, "!") {
		return "", 0, false
	}
	end := strings.IndexByte(rest, '>')
	if end < 0 {
		return "", 0, false
	}
	width := len("(?<") + end + 1
	return source[:width], width, true
}

// namedBackreference matches `\k<name>` at the start of source.
func namedBackreference(source string) (string, int, bool) {
	if !strings.HasPrefix(source, `\k<`) {
		return "", 0, false
	}
	end := strings.IndexByte(source[len(`\k<`):], '>')
	if end < 0 {
		return "", 0, false
	}
	width := len(`\k<`) + end + 1
	return source[:width], width, true
}

// modifierGroup reads a `(?flags:` or `(?flags-flags:` opener , the syntax
// JavaScript spells turning `i`, `m` or `s` on or off over one group with , and
// returns the flags that hold inside it. A modifier may be named once across
// both sides, and naming none at all is `(?:`, an ordinary group.
func modifierGroup(source string, current rewriteOptions) (rewriteOptions, int, bool) {
	if !strings.HasPrefix(source, "(?") {
		return current, 0, false
	}
	rest := source[len("(?"):]
	end := strings.IndexByte(rest, ':')
	if end <= 0 {
		return current, 0, false
	}

	inside := current
	named := map[byte]bool{}
	enable := true
	for i := range end {
		flag := rest[i]
		if flag == '-' {
			if !enable {
				return current, 0, false
			}
			enable = false
			continue
		}
		if named[flag] {
			return current, 0, false
		}
		named[flag] = true
		switch flag {
		case 'i':
			inside.ignoreCase = enable
		case 'm':
			inside.multiline = enable
		case 's':
			inside.dotAll = enable
		default:
			return current, 0, false
		}
	}
	if len(named) == 0 {
		return current, 0, false
	}
	return inside, len("(?") + end + 1, true
}

// checkGroupConstruct rejects a `(?…` opener JavaScript has no syntax for.
// regexp2 follows .NET, which reads several more , `(?i)` sets a flag there
// and is a syntax error in JavaScript , so accepting them would quietly give
// a pattern a meaning its author never wrote.
func checkGroupConstruct(source string) error {
	if !strings.HasPrefix(source, "(?") {
		return nil
	}
	rest := source[len("(?"):]
	for _, valid := range []string{":", "=", "!", "<="} {
		if strings.HasPrefix(rest, valid) {
			return nil
		}
	}
	if strings.HasPrefix(rest, "<!") {
		return nil
	}
	return fmt.Errorf("%w: (?%.1s", ErrUnsupportedSyntax, rest)
}

// escapeKind says what a backslash escape names, which is what decides how the
// rewrite may use it: a character may be widened under `i` and may bound a
// character range, a set may be neither, and an assertion is not a character at
// all.
type escapeKind uint8

const (
	// escapeRune names one character.
	escapeRune escapeKind = iota
	// escapeSet names a set of characters: `\d`, `\w`, `\s` and their negations,
	// and a `\p{…}` property escape.
	escapeSet
	// escapeAssertion names a position rather than a character: `\b` and `\B`,
	// and only outside a character class.
	escapeAssertion
	// escapeBackreference names an earlier group: `\1`, `\k<name>`.
	escapeBackreference
)

// setKind tells apart the sets whose ECMAScript membership the rewrite has to
// spell out from the ones regexp2 already reads the same way.
type setKind uint8

const (
	setOther setKind = iota
	setWord
	setNonWord
	setProperty
)

// decodedEscape is one backslash escape read the way JavaScript reads it.
//
// width counts the bytes the escape spans after the backslash, which is not
// always the whole of what follows: a `\c` that no control letter completes is
// no escape at all under Annex B, and comes back as a backslash of width none,
// leaving the `c` to be read as the character it is.
type decodedEscape struct {
	kind    escapeKind
	set     setKind
	r       rune
	width   int
	negated bool
}

// escapeContext is what an escape's meaning depends on besides its own text.
// A `\b` is a boundary out in the pattern and a backspace inside a character
// class; a `\1` is a backreference where a first group exists and an octal
// escape where none does; and `u` decides both which syntax is legal and, for
// several escapes, what the legal syntax means.
type escapeContext struct {
	unicode bool
	inClass bool
	groups  int
	named   bool
}

// syntaxCharacters is what `u` narrows an identity escape down to, alongside
// `/`. Without `u`, Annex B lets a backslash stand before nearly anything.
const syntaxCharacters = `^$\.*+?()[]{}|`

// decodeEscape reads the escape whose backslash ends at i, returning what it
// names. An escape JavaScript itself rejects comes back as an error rather than
// as text for regexp2 to judge, which reads a different syntax.
func decodeEscape(source string, i int, ctx escapeContext) (decodedEscape, error) {
	if i >= len(source) {
		return decodedEscape{}, fmt.Errorf(`%w: \ at the end of the pattern`, ErrUnsupportedSyntax)
	}
	r, size := utf8.DecodeRuneInString(source[i:])

	switch r {
	case 'u':
		return decodeUnicodeEscape(source, i, size, ctx)
	case 'x':
		return decodeFixedHex(source, i, size, 2, ctx)
	case 'n':
		return decodedEscape{r: '\n', width: size}, nil
	case 'r':
		return decodedEscape{r: '\r', width: size}, nil
	case 't':
		return decodedEscape{r: '\t', width: size}, nil
	case 'f':
		return decodedEscape{r: '\f', width: size}, nil
	case 'v':
		return decodedEscape{r: '\v', width: size}, nil
	case 'c':
		return decodeControlEscape(source, i, size, ctx)
	case 'd', 'D', 's', 'S':
		return decodedEscape{kind: escapeSet, width: size}, nil
	case 'w':
		return decodedEscape{kind: escapeSet, set: setWord, width: size}, nil
	case 'W':
		return decodedEscape{kind: escapeSet, set: setNonWord, width: size}, nil
	case 'p', 'P':
		return decodePropertyEscape(source, i, size, ctx)
	case 'b':
		if ctx.inClass {
			// Inside a class a `\b` is a backspace, which is the one place the
			// two readings of the escape part ways.
			return decodedEscape{r: '\b', width: size}, nil
		}
		return decodedEscape{kind: escapeAssertion, width: size}, nil
	case 'B':
		if ctx.inClass {
			return identityEscape(r, size, ctx)
		}
		return decodedEscape{kind: escapeAssertion, width: size, negated: true}, nil
	case 'k':
		// `\k<name>` is read whole before this. What is left is a `\k` naming
		// nothing, which Annex B reads as the letter , but only in a pattern
		// that opens no named group, since one that does makes the reference
		// mandatory.
		if ctx.named {
			return decodedEscape{}, fmt.Errorf(`%w: \k naming no group`, ErrUnsupportedSyntax)
		}
		return identityEscape(r, size, ctx)
	}

	if r >= '0' && r <= '9' {
		return decodeNumericEscape(source, i, ctx)
	}
	return identityEscape(r, size, ctx)
}

// decodeUnicodeEscape reads `\uXXXX`, or `\u{X…}` where the `u` flag spells it.
func decodeUnicodeEscape(source string, i int, size int, ctx escapeContext) (decodedEscape, error) {
	if ctx.unicode && strings.HasPrefix(source[i+size:], "{") {
		end := strings.IndexByte(source[i+size:], '}')
		if end < 0 {
			return decodedEscape{}, fmt.Errorf(`%w: \u{ that no } closes`, ErrUnsupportedSyntax)
		}
		value, err := strconv.ParseUint(source[i+size+1:i+size+end], 16, 32)
		if err != nil || value > unicode.MaxRune {
			return decodedEscape{}, fmt.Errorf(`%w: \u{…} naming no character`, ErrUnsupportedSyntax)
		}
		return decodedEscape{r: rune(value), width: size + end + 1}, nil
	}
	return decodeFixedHex(source, i, size, 4, ctx)
}

// decodeFixedHex reads the fixed-width hexadecimal escapes, `\xXX` and
// `\uXXXX`. Without `u`, one that no hex digits complete is not an error but an
// identity escape, so `/\x/` matches an `x` , and under `i` an `X` as well.
func decodeFixedHex(source string, i int, size int, digits int, ctx escapeContext) (decodedEscape, error) {
	start := i + size
	if start+digits <= len(source) {
		if value, err := strconv.ParseUint(source[start:start+digits], 16, 32); err == nil {
			return decodedEscape{r: rune(value), width: size + digits}, nil
		}
	}
	if ctx.unicode {
		return decodedEscape{}, fmt.Errorf(`%w: \%.1s naming no character`, ErrUnsupportedSyntax, source[i:])
	}
	r, _ := utf8.DecodeRuneInString(source[i:])
	return decodedEscape{r: r, width: size}, nil
}

// decodeControlEscape reads `\cX`, the control character an ASCII letter names.
func decodeControlEscape(source string, i int, size int, ctx escapeContext) (decodedEscape, error) {
	if letter := i + size; letter < len(source) {
		c := source[letter]
		// Annex B widens the letter a character class takes to a digit or an
		// underscore, which is the one place the two productions differ.
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' ||
			!ctx.unicode && ctx.inClass && (c >= '0' && c <= '9' || c == '_') {
			return decodedEscape{r: rune(c % 32), width: size + 1}, nil
		}
	}
	if ctx.unicode {
		return decodedEscape{}, fmt.Errorf(`%w: \c naming no control character`, ErrUnsupportedSyntax)
	}
	// Annex B: with no control letter after it the backslash is not an escape
	// at all and stands for itself, leaving the `c` to be read on its own.
	return decodedEscape{r: '\\', width: 0}, nil
}

// decodePropertyEscape reads `\p{…}` and `\P{…}`, which name a set out of
// Unicode's tables. Only the `u` flag spells them; without it .NET would still
// read a property where JavaScript reads the letter.
func decodePropertyEscape(source string, i int, size int, ctx escapeContext) (decodedEscape, error) {
	if !ctx.unicode {
		return identityEscape(rune(source[i]), size, ctx)
	}
	if !strings.HasPrefix(source[i+size:], "{") {
		return decodedEscape{}, fmt.Errorf(`%w: \%.1s naming no property`, ErrUnsupportedSyntax, source[i:])
	}
	end := strings.IndexByte(source[i+size:], '}')
	if end < 0 {
		return decodedEscape{}, fmt.Errorf(`%w: \%.1s{ that no } closes`, ErrUnsupportedSyntax, source[i:])
	}
	// The name belongs to the escape: `\p{Script=Greek}`. Taking it along keeps
	// the walk from reading the letters in it as text and widening them.
	return decodedEscape{kind: escapeSet, set: setProperty, width: size + end + 1}, nil
}

// decodeNumericEscape reads an escape a digit opens, which is a backreference,
// a NUL, or one of Annex B's legacy octal escapes depending on the digits, on
// where it sits, and on how many groups the pattern opens.
func decodeNumericEscape(source string, i int, ctx escapeContext) (decodedEscape, error) {
	if source[i] == '0' && !isDecimalDigit(source, i+1) {
		return decodedEscape{r: 0, width: 1}, nil
	}
	// A number no group answers to is not a backreference at all, and Annex B
	// reads the same text as an escaped character instead. Inside a class there
	// is no backreference to read in the first place.
	if source[i] != '0' && !ctx.inClass {
		if value, width := decimalEscape(source, i); value <= ctx.groups {
			return decodedEscape{kind: escapeBackreference, width: width}, nil
		}
	}
	if ctx.unicode {
		return decodedEscape{}, fmt.Errorf(`%w: \%.1s`, ErrUnsupportedSyntax, source[i:])
	}
	if r, width := decodeLegacyOctal(source, i); width > 0 {
		return decodedEscape{r: r, width: width}, nil
	}
	// `\8` and `\9`, which no octal escape spells.
	return decodedEscape{r: rune(source[i]), width: 1}, nil
}

// decimalEscape reads the whole run of digits a backreference is numbered with,
// counting past what any pattern could hold rather than overflowing.
func decimalEscape(source string, i int) (int, int) {
	const tooMany = 1 << 20
	value, width := 0, 0
	for isDecimalDigit(source, i+width) {
		if value < tooMany {
			value = value*10 + int(source[i+width]-'0')
		}
		width++
	}
	return value, width
}

// decodeLegacyOctal reads the octal escape Annex B keeps for a pattern without
// `u`, which spans up to three digits and never overruns one byte.
func decodeLegacyOctal(source string, i int) (rune, int) {
	if source[i] < '0' || source[i] > '7' {
		return 0, 0
	}
	digits := 3
	if source[i] > '3' {
		digits = 2
	}
	value, width := rune(source[i]-'0'), 1
	for width < digits && i+width < len(source) && source[i+width] >= '0' && source[i+width] <= '7' {
		value = value*8 + rune(source[i+width]-'0')
		width++
	}
	return value, width
}

// identityEscape reads a backslash standing before a character that carries no
// escape of its own. Annex B lets it stand before nearly anything; `u` narrows
// it to the characters a pattern would otherwise read as syntax.
func identityEscape(r rune, size int, ctx escapeContext) (decodedEscape, error) {
	if ctx.unicode && !strings.ContainsRune(syntaxCharacters, r) && r != '/' && (!ctx.inClass || r != '-') {
		return decodedEscape{}, fmt.Errorf(`%w: \%c`, ErrUnsupportedSyntax, r)
	}
	return decodedEscape{r: r, width: size}, nil
}

func isDecimalDigit(source string, i int) bool {
	return i < len(source) && source[i] >= '0' && source[i] <= '9'
}

// countGroups reads how many capturing groups a pattern opens and whether any
// of them is named, which is what a `\1` and a `\k` are read against.
func countGroups(source string) (int, bool) {
	count, named := 0, false
	inClass := false
	for i := 0; i < len(source); {
		switch source[i] {
		case '\\':
			_, size := utf8.DecodeRuneInString(source[i+1:])
			i += 1 + size
			continue
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '(':
			if inClass {
				break
			}
			rest := source[i+1:]
			switch {
			case !strings.HasPrefix(rest, "?"):
				count++
			case strings.HasPrefix(rest, "?<") && !strings.HasPrefix(rest, "?<=") && !strings.HasPrefix(rest, "?<!"):
				count++
				named = true
			}
		}
		_, size := utf8.DecodeRuneInString(source[i:])
		i += size
	}
	return count, named
}

// classAtomKind says what one member of a character class is. Keeping the four
// apart is what lets a range be read as a range only where one was written: in
// `[\d-A]` the `-` separates no two characters, so JavaScript reads three
// members rather than a range, and the `A` goes on being a character that a
// case-insensitive class has to widen.
type classAtomKind uint8

const (
	// classRune is one character.
	classRune classAtomKind = iota
	// classRange is every character between two, inclusive.
	classRange
	// classSet is a set escape, which is written out as it stands and can
	// bound no range.
	classSet
	// classDash is a `-` that separates nothing and so stands for itself.
	classDash
)

type classAtom struct {
	kind   classAtomKind
	lo, hi rune
	text   string
}

// covers reports whether the atom already names r, which is what decides
// whether widening the class has anything to add for it.
func (a classAtom) covers(r rune) bool {
	switch a.kind {
	case classRune, classDash:
		return r == a.lo
	case classRange:
		return r >= a.lo && r <= a.hi
	}
	return false
}

func (a classAtom) write() string {
	switch a.kind {
	case classSet:
		return a.text
	case classRange:
		return literalRune(a.lo) + "-" + literalRune(a.hi)
	default:
		return literalRune(a.lo)
	}
}

// wordClassAtoms spells `\w` as class members, in ascending order. ECMAScript
// builds the set out of ASCII whatever the flags say, and under `u` and `i`
// together it gains the two characters that fold into it: U+017F LATIN SMALL
// LETTER LONG S onto `s`, U+212A KELVIN SIGN onto `k`.
func wordClassAtoms(options rewriteOptions) []classAtom {
	atoms := []classAtom{
		{kind: classRange, lo: '0', hi: '9'},
		{kind: classRange, lo: 'A', hi: 'Z'},
		{kind: classRune, lo: '_'},
		{kind: classRange, lo: 'a', hi: 'z'},
	}
	if options.ignoreCase && options.unicode {
		atoms = append(atoms,
			classAtom{kind: classRune, lo: 0x017F},
			classAtom{kind: classRune, lo: 0x212A},
		)
	}
	return atoms
}

// nonWordClassAtoms spells `\W` as class members. Out in a pattern a negated
// class says it in one word, but inside a class there is nothing to negate
// against, so the complement is written as the ranges between the word
// characters.
func nonWordClassAtoms(options rewriteOptions) []classAtom {
	span := func(lo, hi rune) classAtom {
		if lo == hi {
			return classAtom{kind: classRune, lo: lo}
		}
		return classAtom{kind: classRange, lo: lo, hi: hi}
	}

	atoms := []classAtom{}
	next := rune(0)
	for _, word := range wordClassAtoms(options) {
		hi := word.lo
		if word.kind == classRange {
			hi = word.hi
		}
		if next < word.lo {
			atoms = append(atoms, span(next, word.lo-1))
		}
		next = hi + 1
	}
	return append(atoms, span(next, unicode.MaxRune))
}

// readClass reads a character class off the front of source, which opens at
// `[`, returning its body, whether it is negated and how many bytes it spans.
func readClass(source string) (body string, negated bool, width int, err error) {
	i := len("[")
	if i < len(source) && source[i] == '^' {
		negated = true
		i++
	}
	start := i
	for i < len(source) {
		switch source[i] {
		case '\\':
			_, size := utf8.DecodeRuneInString(source[i+1:])
			i += 1 + size
		case ']':
			return source[start:i], negated, i + 1, nil
		default:
			_, size := utf8.DecodeRuneInString(source[i:])
			i += size
		}
	}
	return "", false, 0, fmt.Errorf("%w: [ that no ] closes", ErrUnsupportedSyntax)
}

// classAtoms reads a class body into the members it names. The second result is
// false where the class holds a property escape under `i`, which names a set
// out of Unicode's tables that a widened class has no way to name back.
func classAtoms(body string, options rewriteOptions, ctx escapeContext) ([]classAtom, bool, error) {
	ctx.inClass = true
	atoms := []classAtom{}
	exact := true

	for i := 0; i < len(body); {
		r, size := utf8.DecodeRuneInString(body[i:])
		if r != '\\' {
			if r == '-' {
				atoms = append(atoms, classAtom{kind: classDash, lo: '-'})
			} else {
				atoms = append(atoms, classAtom{kind: classRune, lo: r})
			}
			i += size
			continue
		}

		escape, err := decodeEscape(body, i+size, ctx)
		if err != nil {
			return nil, false, err
		}
		end := i + size + escape.width
		switch escape.kind {
		case escapeSet:
			// regexp2 reads `\d` and `\s` the way ECMAScript does, and `\w`
			// only until `u` and `i` together widen the set past ASCII.
			switch {
			case escape.set == setWord && options.ignoreCase && options.unicode:
				atoms = append(atoms, wordClassAtoms(options)...)
			case escape.set == setNonWord && options.ignoreCase && options.unicode:
				atoms = append(atoms, nonWordClassAtoms(options)...)
			default:
				if escape.set == setProperty && options.ignoreCase {
					exact = false
				}
				atoms = append(atoms, classAtom{kind: classSet, text: body[i:end]})
			}
		default:
			atoms = append(atoms, classAtom{kind: classRune, lo: escape.r})
		}
		i = end
	}

	joined, err := joinRanges(atoms, options)
	if err != nil {
		return nil, false, err
	}
	return joined, exact, nil
}

// joinRanges reads the `x-y` ranges out of a class's members. A `-` with a set
// on either side of it separates nothing, which Annex B reads as three members
// and `u` rejects outright.
func joinRanges(atoms []classAtom, options rewriteOptions) ([]classAtom, error) {
	joined := make([]classAtom, 0, len(atoms))
	for i := 0; i < len(atoms); i++ {
		if i+2 < len(atoms) && atoms[i+1].kind == classDash {
			lo, hi := atoms[i], atoms[i+2]
			if lo.kind == classRune && hi.kind == classRune {
				if lo.lo > hi.lo {
					return nil, fmt.Errorf("%w: a character range running backwards", ErrUnsupportedSyntax)
				}
				joined = append(joined, classAtom{kind: classRange, lo: lo.lo, hi: hi.lo})
				i += 2
				continue
			}
			if options.unicode {
				return nil, fmt.Errorf("%w: a set at the end of a character range", ErrUnsupportedSyntax)
			}
		}
		joined = append(joined, atoms[i])
	}
	return joined, nil
}

// writeClass writes the members out as a class regexp2 reads the same way,
// widening it first where the flags ask for it.
func writeClass(atoms []classAtom, negated bool, options rewriteOptions) string {
	if len(atoms) == 0 {
		// `[]` never matches and `[^]` matches anything: syntax only JavaScript
		// reads, and regexp2 would take the brackets literally.
		if negated {
			return anyCharacter
		}
		return neverMatches
	}

	var body strings.Builder
	for _, atom := range atoms {
		body.WriteString(atom.write())
	}
	if options.ignoreCase {
		body.WriteString(caseExtras(atoms, options.unicode))
	}
	if negated {
		return "[^" + body.String() + "]"
	}
	return "[" + body.String() + "]"
}

// caseExtras is what a class has to gain to cover every character a `/i`
// comparison accepts for one it already covers.
//
// What the class already names is left alone and the rest is appended, so a
// negated class goes on negating whatever the widened class covers. That is how
// JavaScript reads one too: `[^a]` asks whether any member compares equal to
// the character at hand, and answers no to `A`.
func caseExtras(atoms []classAtom, unicodeMode bool) string {
	covers := func(r rune) bool {
		return slices.ContainsFunc(atoms, func(atom classAtom) bool { return atom.covers(r) })
	}

	var extras strings.Builder
	for _, members := range CaseEquivalenceGroups(unicodeMode) {
		if !slices.ContainsFunc(members, covers) {
			continue
		}
		for _, member := range members {
			if !covers(member) {
				extras.WriteString(literalRune(member))
			}
		}
	}
	return extras.String()
}

// A JavaScript regexp carrying the `i` flag compares two characters by
// Canonicalize, which is neither Go's case folding nor the Unicode
// folding regexp2 reaches for when told to ignore case. It is also not one
// reading but two: `u` and `v` select simple case folding, and their absence
// selects an uppercase mapping that never crosses into ASCII. U+212A KELVIN
// SIGN against `k` comes out differently under each, so every entry point
// here carries the flag through rather than picking one.
//
// So rather than ask regexp2 to ignore case, a pattern is widened: every
// literal and every character class is rewritten to name each character it
// should compare equal to, and the widened pattern is matched exactly.

// CaseClass widens one literal character to the class of characters a `/i`
// comparison accepts for it, reporting false when the character stands alone
// and needs no widening.
func CaseClass(r rune, unicode bool) (string, bool) {
	members := CaseEquivalents(r, unicode)
	if len(members) == 0 {
		return "", false
	}
	var class strings.Builder
	class.WriteByte('[')
	for _, member := range members {
		class.WriteString(EscapeClassRune(member))
	}
	class.WriteByte(']')
	return class.String(), true
}

// CaseCloseClass widens a character class to cover every character a `/i`
// comparison accepts for one it already covers. It takes the body of a class ,
// what sits between the brackets, leading `^` included , and returns the body
// to write instead.
//
// What the class already names is left alone and the rest is appended, so a
// negated class goes on negating whatever the widened class covers. That is
// how JavaScript reads one too: `[^a]` asks whether any member compares equal
// to the character at hand, and answers no to `A`.
func CaseCloseClass(body string, unicode bool) string {
	options := rewriteOptions{ignoreCase: true, unicode: unicode}
	atoms, _, err := classAtoms(body, options, escapeContext{unicode: unicode})
	if err != nil {
		return body
	}
	extras := caseExtras(atoms, unicode)
	if extras == "" {
		return body
	}
	return escapeTrailingDash(body) + extras
}

// EscapeClassRune escapes a character that would carry a meaning of its own
// inside a character class.
func EscapeClassRune(r rune) string {
	if strings.ContainsRune(`\]^-[`, r) {
		return `\` + string(r)
	}
	return string(r)
}

// escapeTrailingDash escapes the `-` that stands for itself at the end of a
// character class, so that appending to the class does not read it as opening
// a range instead.
func escapeTrailingDash(body string) string {
	if !strings.HasSuffix(body, "-") {
		return body
	}
	backslashes := 0
	for i := len(body) - 2; i >= 0 && body[i] == '\\'; i-- {
		backslashes++
	}
	if backslashes%2 == 1 {
		return body
	}
	return body[:len(body)-1] + `\-`
}

// Canonicalize maps a character to the one JavaScript compares it as when case
// does not matter. Which reading applies turns on whether the regexp carries
// `u` or `v`, and the two disagree:
//
//   - With one of those flags, ECMAScript folds by Unicode's simple case
//     folding, so `k`, `K` and U+212A KELVIN SIGN are one character, as are
//     `s`, `S` and U+017F LATIN SMALL LETTER LONG S.
//   - Without, ECMAScript uppercases the character, with the one rule that a
//     character outside ASCII never canonicalizes into ASCII. That splits the
//     two groups above back apart, while leaving U+03A3 Σ, U+03C3 σ and
//     U+03C2 ς as one character under either reading.
//
// Go's simple uppercase stands in for String.prototype.toUpperCase, once the
// characters whose full uppercase spans more than one UTF-16 code unit are
// held back , `ß` is the familiar one , because JavaScript keeps the original
// character there.
//
// https://tc39.es/ecma262/2024/multipage/text-processing.html#sec-runtime-semantics-canonicalize-ch
func Canonicalize(r rune, unicodeMode bool) rune {
	if unicodeMode {
		return simpleFold(r)
	}
	// Canonicalizing one UTF-16 code unit at a time is what makes this the
	// answer for a regexp without the `u` flag, and it leaves a
	// supplementary-plane character to stand for itself.
	if r > 0xFFFF || expandsOnUppercase(r) {
		return r
	}
	upper := unicode.ToUpper(r)
	if r >= utf8.RuneSelf && upper < utf8.RuneSelf {
		return r
	}
	return upper
}

// simpleFold maps a character to the least of the characters it folds together
// with, which serves as the canonical form under `u` and `v`. Simple case
// folding carries no rule about ASCII and no rule about the supplementary
// plane, so neither guard above applies here.
func simpleFold(r rune) rune {
	least := r
	for folded := unicode.SimpleFold(r); folded != r; folded = unicode.SimpleFold(folded) {
		if folded < least {
			least = folded
		}
	}
	return least
}

// expandsOnUppercase reports the characters whose simple uppercase is one
// character but whose full uppercase is several.
func expandsOnUppercase(r rune) bool {
	return r >= 0x1F80 && r <= 0x1F87 ||
		r >= 0x1F90 && r <= 0x1F97 ||
		r >= 0x1FA0 && r <= 0x1FA7 ||
		r == 0x1FB3 || r == 0x1FC3 || r == 0x1FF3
}

// CaseEquivalents returns every character that Canonicalize maps onto the same
// character as r, r included, or nil when r stands alone. The result is shared
// and must not be modified.
//
// A caller comparing one character against another can just canonicalize both.
// This is for a caller that has to widen a set , the members of a regexp
// character class, say , to cover everything a case-insensitive comparison
// would accept.
func CaseEquivalents(r rune, unicodeMode bool) []rune {
	byMember, _ := caseTables(unicodeMode)
	return byMember[r]
}

// CaseEquivalenceGroups returns every group of two or more characters that
// canonicalize alike. The groups are shared and must not be modified.
//
// Widening a range rather than a single character means asking which groups
// reach into it, which needs the groups enumerated rather than looked up.
func CaseEquivalenceGroups(unicodeMode bool) [][]rune {
	_, groups := caseTables(unicodeMode)
	return groups
}

// caseTables groups the characters that canonicalize alike, under whichever
// reading the caller's flags select. Each table is built on first use, so a
// caller that never compares without regard to case pays for neither, and one
// that only ever sees a `u` pattern pays for one.
func caseTables(unicodeMode bool) (map[rune][]rune, [][]rune) {
	if unicodeMode {
		return foldTables()
	}
	return uppercaseTables()
}

var (
	uppercaseTables = sync.OnceValues(func() (map[rune][]rune, [][]rune) { return buildCaseTables(false) })
	foldTables      = sync.OnceValues(func() (map[rune][]rune, [][]rune) { return buildCaseTables(true) })
)

// buildCaseTables collects the characters that canonicalize alike. A group of
// one has nothing to widen, so only the rest are kept: as a lookup by member,
// and as a list to walk.
func buildCaseTables(unicodeMode bool) (map[rune][]rune, [][]rune) {
	grouped := map[rune][]rune{}
	record := func(r rune) {
		canonical := Canonicalize(r, unicodeMode)
		if !slices.Contains(grouped[canonical], r) {
			grouped[canonical] = append(grouped[canonical], r)
		}
	}
	// Every character that canonicalizes onto another one has a case mapping,
	// so the case ranges name them all. Folding reaches further than
	// the uppercase mapping does, though , U+212A KELVIN SIGN joins `k` and
	// `K` without any of the three having the other as its uppercase , so each
	// orbit is walked out rather than assumed to be a pair.
	for _, caseRange := range unicode.CaseRanges {
		for r := rune(caseRange.Lo); r <= rune(caseRange.Hi); r++ {
			record(r)
			record(Canonicalize(r, unicodeMode))
			if unicodeMode {
				for folded := unicode.SimpleFold(r); folded != r; folded = unicode.SimpleFold(folded) {
					record(folded)
				}
			}
		}
	}
	// Three pairs fold together with no case mapping in either direction, so unicode.CaseRanges does
	// not reach them and the walk above never records them. unicode.SimpleFold knows them; the range
	// table does not, and this function is driven by the range table. Measured on Go 1.27: all six
	// runes report inCaseRanges=false while folding correctly.
	//
	// They were previously supplied by an internal Unicode-delta package, and removing that package
	// on a toolchain where its per-rune answers had become identical to the standard library's still
	// dropped these, because seeding a table is a different question from answering a lookup.
	for _, pair := range [...][2]rune{{0x0390, 0x1FD3}, {0x03B0, 0x1FE3}, {0xFB05, 0xFB06}} {
		record(pair[0])
		record(pair[1])
	}

	byMember := map[rune][]rune{}
	groups := [][]rune{}
	for _, members := range grouped {
		if len(members) < 2 {
			continue
		}
		slices.Sort(members)
		groups = append(groups, members)
		for _, member := range members {
			byMember[member] = members
		}
	}
	return byMember, groups
}
