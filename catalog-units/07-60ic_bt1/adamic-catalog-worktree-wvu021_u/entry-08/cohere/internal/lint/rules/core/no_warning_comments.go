package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoWarningCommentsLocation says where in a comment a term has to appear to count.
type NoWarningCommentsLocation string

const (
	// NoWarningCommentsStart matches only at the head of the comment, past leading whitespace
	// and any configured decoration characters. The default.
	NoWarningCommentsStart NoWarningCommentsLocation = "Start"

	// NoWarningCommentsAnywhere matches the term anywhere in the comment, on word boundaries.
	NoWarningCommentsAnywhere NoWarningCommentsLocation = "Anywhere"
)

// UnmarshalJSON reads upstream's spelling, `"start"` or `"anywhere"`, and refuses anything else.
//
// The config used to want `"Start"` and `"Anywhere"`, which refused upstream's own spelling while
// no ESLint version accepted ours; the PascalCase form is now refused instead (#d21war2). An
// unrecognized value is an error rather than a fallback, because every arm is selected by string
// equality and an unknown value would silently pick a third behaviour of matching nothing.
func (location *NoWarningCommentsLocation) UnmarshalJSON(raw []byte) error {
	var spelling string
	if err := json.Unmarshal(raw, &spelling); err != nil {
		return err
	}
	switch spelling {
	case "start":
		*location = NoWarningCommentsStart
	case "anywhere":
		*location = NoWarningCommentsAnywhere
	default:
		return fmt.Errorf(`location is %q, which is neither "start" nor "anywhere"`, spelling)
	}
	return nil
}

// MarshalJSON writes the spelling UnmarshalJSON reads, so an options value round-trips.
func (location NoWarningCommentsLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToLower(string(location)))
}

// NoWarningCommentsOptions configures the rule.
//
// Upstream's option is a single object, and the config takes it in exactly that form.
type NoWarningCommentsOptions struct {
	// Terms are the words that make a comment a warning. Absent means upstream's default set.
	Terms []string `json:"terms"`

	// Location is where the term has to appear. Absent means Start.
	//
	// Omitted when empty, because the empty location is the absent one and upstream's enum has no
	// spelling for it: written out, it would be refused on the way back in.
	Location NoWarningCommentsLocation `json:"location,omitempty"`

	// Decoration are single characters allowed to precede the term under Start, so a banner
	// comment still matches. Each entry must be exactly one non-whitespace character.
	//
	// Absent is NOT the same as empty: with no decoration configured, `/** TODO */` is clean
	// because the asterisk blocks the anchor, and configuring `["*"]` makes it report. Measured.
	Decoration []string `json:"decoration"`
}

// noWarningCommentsDefaultTerms is upstream's `terms` default.
var noWarningCommentsDefaultTerms = []string{"todo", "fixme", "xxx"}

// noWarningCommentsCharacterLimit is upstream's `CHAR_LIMIT`, the width the quoted comment is
// truncated to in the message.
const noWarningCommentsCharacterLimit = 40

// DecodeNoWarningCommentsOptions reads this rule's configuration from the config layer.
//
// Hand-rolled because a decoration entry has to be exactly one non-whitespace character, which
// upstream expresses as a schema `pattern` and we have no schema layer for. The location is checked
// by its own UnmarshalJSON.
func DecodeNoWarningCommentsOptions(raw []byte) (any, error) {
	options := NoWarningCommentsOptions{}
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	for index, decoration := range options.Decoration {
		if len([]rune(decoration)) != 1 || text.TrimWhitespace(decoration) == "" {
			return options, fmt.Errorf(
				"no-warning-comments: decoration %d is %q, wanted exactly one non-whitespace "+
					"character", index, decoration)
		}
	}
	return options, nil
}

// NoWarningComments reports a comment carrying a term that marks unfinished work.
//
//	valid:   // this is fine
//	valid:   // something todo            (under the default Start location)
//	invalid: // TODO: fix this
//	invalid: /* FIXME */
//	invalid: // something todo            (under Anywhere)
//
// # What it is for, and why the default is Start
//
// A `TODO` is a note to a future reader that the code is knowingly incomplete, and the rule exists
// so those notes are counted rather than accumulating unread. Matching only at the START of a
// comment is deliberate: a comment whose first word is `TODO` is a marker, while one that merely
// mentions the word in prose is usually describing something rather than deferring it.
//
// # The comment VALUE, not its source text
//
// Upstream matches against `node.value`, which is the comment body with the delimiters removed and
// nothing else stripped. That is the whole reason `/** TODO */` is clean by default while
// `/* TODO */` reports: the surviving asterisk sits between the anchor and the term. Our comment
// reader hands back the full source text including delimiters, so they are removed here, and
// getting that wrong in either direction moves a whole class of comment across the line.
//
// # Word boundaries, which are ASCII on both sides
//
// The term is bounded by `\b` on whichever side starts or ends with a word character, so `fixed`
// and `affix` do not match `fix`, while a term written `!FIX` is unbounded on its left. Go's `\b`
// is ASCII-only and JavaScript's is too even under the `u` flag, which was measured rather than
// assumed across six non-ASCII shapes: both agree that `ÜTODO`, `TODOé` and `日TODO` match while
// `ТODO` spelled with a Cyrillic Te does not.
//
// # The message quotes the comment, truncated on a word boundary
//
// Upstream rebuilds the quoted text word by word and stops before exceeding forty characters,
// appending an ellipsis when it stopped early. So the quote never splits a word, and a comment whose
// FIRST word already exceeds the limit quotes as nothing but the ellipsis. Reproduced exactly,
// because the quoted text is the part of the message a reader scans.
var NoWarningComments = rule.Rule{
	Name: "no-warning-comments",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := NoWarningCommentsOptions{}
		if decoded, configured := rule.OptionsAs[NoWarningCommentsOptions](options); configured {
			settings = decoded
		}
		terms := settings.Terms
		if terms == nil {
			terms = noWarningCommentsDefaultTerms
		}
		location := settings.Location
		if location == "" {
			location = NoWarningCommentsStart
		}

		compiled := noWarningCommentsMatchers(terms, location, settings.Decoration)
		if len(compiled.matchers) == 0 {
			// No terms means nothing to look for, and upstream's empty `terms` array behaves the
			// same way. Returning no listener keeps the cost at zero rather than walking every
			// comment in the tree to compare against nothing.
			return rule.Listeners{}
		}

		return rule.Listeners{
			ast.KindSourceFile: func(file *ast.Node) {
				checkNoWarningComments(ctx, compiled)
			},
		}
	},
}

// noWarningCommentsCache holds the compiled matchers for each configuration seen.
//
// `Run` is called ONCE PER FILE, so compiling the patterns there compiled them 3,516 times on this
// tree: measured at 235ms total, against 27ms for the two comparable comment-reading rules already
// shipped. The patterns depend only on the configuration, which is the same for every file in a
// run, so they are built once and reused.
//
// Keyed on the configuration rather than held in a package variable, because a single run can lint
// under more than one configuration and a bare variable would serve the first one's patterns to
// every later file.
var noWarningCommentsCache sync.Map

// noWarningCommentsMatchers compiles the patterns for one configuration, or returns the cached set.
func noWarningCommentsMatchers(
	terms []string,
	location NoWarningCommentsLocation,
	decoration []string,
) noWarningCommentsCompiled {
	key := strings.Join(terms, "\x00") + "\x01" + string(location) + "\x01" +
		strings.Join(decoration, "\x00")
	if cached, found := noWarningCommentsCache.Load(key); found {
		return cached.(noWarningCommentsCompiled)
	}

	skip, leadsAreExact := noWarningCommentsStartSkip(location, decoration)
	matchers := make([]*regexp.Regexp, 0, len(terms))
	matched := make([]string, 0, len(terms))
	leads := make([]rune, 0, len(terms))
	for _, term := range terms {
		matcher, err := noWarningCommentsMatcherFor(term, location, decoration)
		if err != nil {
			// A term that cannot be compiled is a configuration mistake rather than a finding.
			// Skipped rather than reported, matching upstream, whose `escapeRegExp` makes every
			// term compilable and which therefore has no failure mode here at all.
			continue
		}
		matchers = append(matchers, matcher)
		matched = append(matched, term)
		leads = append(leads, noWarningCommentsLeadOf(term, skip, leadsAreExact))
	}

	compiled := noWarningCommentsCompiled{matchers: matchers, terms: matched, leads: leads, skip: skip}
	noWarningCommentsCache.Store(key, compiled)
	return compiled
}

// noWarningCommentsCompiled is one configuration's compiled patterns and what they are checked with.
type noWarningCommentsCompiled struct {
	matchers []*regexp.Regexp

	// terms are the terms that compiled, aligned with the matchers, so the message can name the term
	// that matched rather than the term at the same index in the configured list. Those two lists
	// differ whenever a term fails to compile.
	terms []string

	// leads are the character each matcher's match has to begin with once the skipped characters are
	// passed, aligned with the matchers, or noWarningCommentsNoLead where that cannot be known without
	// the pattern. See noWarningCommentsStartSkip.
	leads []rune

	// skip is the characters the Start prefix passes over, ASCII whitespace and the decoration.
	skip []rune
}

// noWarningCommentsNoLead marks a matcher that is always run.
const noWarningCommentsNoLead rune = -1

// noWarningCommentsStartSkip is the set of characters the Start prefix `^[\s<decoration>]*` passes
// over, and whether a term's first character then says exactly which comments its pattern can match.
//
// Every comment in a file meets every term's pattern, three by default, and nearly every comment
// fails at its first character: the prefix is anchored, so a Start match has to begin, once the
// skipped characters are passed, with the term's own first character, case folded. Checking that
// character first runs the pattern only where it could match (#fcac58b).
//
// The check is exact only where the prefix is plain. Under Anywhere there is no anchor, so it never
// applies. The class is case-insensitive like the rest of the pattern, so a decoration character with
// a case fold of its own would skip its other case too; rather than mirror that, such a configuration
// runs every pattern, and so does one with a `-` in it. `\s` is Go's, four ASCII control characters
// and the space, the same class the pattern compiles (the gap from JavaScript's is #tez6dna's, and
// moving it here would change findings). TestNoWarningCommentsLeadAnswersAsThePatternDoes holds the
// check to the bare pattern across these shapes.
func noWarningCommentsStartSkip(location NoWarningCommentsLocation, decoration []string) ([]rune, bool) {
	if location != NoWarningCommentsStart {
		return nil, false
	}
	skip := []rune{'\t', '\n', '\f', '\r', ' '}
	for _, entry := range decoration {
		for _, character := range entry {
			// QuoteMeta leaves `-` alone, so between two decoration characters it spells a range
			// in the class, and the pattern skips characters this set would not.
			if character == '-' || unicode.SimpleFold(character) != character {
				return nil, false
			}
			skip = append(skip, character)
		}
	}
	return skip, true
}

// noWarningCommentsLeadOf is the character a term's Start match begins with, or noWarningCommentsNoLead.
//
// A term that begins with a skipped character has no single lead, since the prefix can stop short of
// it, and an empty term matches every comment.
func noWarningCommentsLeadOf(term string, skip []rune, leadsAreExact bool) rune {
	if !leadsAreExact || term == "" {
		return noWarningCommentsNoLead
	}
	lead, _ := utf8.DecodeRuneInString(term)
	if slices.Contains(skip, lead) {
		return noWarningCommentsNoLead
	}
	return lead
}

// noWarningCommentsFirstUnskipped is the first character of a comment past the skipped ones, as the
// pattern reads it (an invalid byte is utf8.RuneError, in both), and false when nothing is left.
func noWarningCommentsFirstUnskipped(value string, skip []rune) (rune, bool) {
	for _, character := range value {
		if !slices.Contains(skip, character) {
			return character, true
		}
	}
	return 0, false
}

// noWarningCommentsFoldEqual reports whether two characters match under the pattern's `(?i)`, which
// matches a literal against every member of its simple case-folding orbit.
func noWarningCommentsFoldEqual(first rune, second rune) bool {
	if first == second {
		return true
	}
	for folded := unicode.SimpleFold(first); folded != first; folded = unicode.SimpleFold(folded) {
		if folded == second {
			return true
		}
	}
	return false
}

// checkNoWarningComments walks every comment in the file.
//
// One listener on the source file rather than a per-node one, because a comment is trivia and has
// no node kind to anchor on. `comments.ForFile` caches one scan per file, so this is a read of an
// already-built list.
func checkNoWarningComments(ctx rule.Context, compiled noWarningCommentsCompiled) {
	for _, comment := range comments.ForFile(ctx) {
		value := noWarningCommentsValueOf(comment)
		first, hasFirst := noWarningCommentsFirstUnskipped(value, compiled.skip)

		// Asked only once a term has matched, since the exemption can only spare a finding.
		selfDirectiveChecked := false
		for index, matcher := range compiled.matchers {
			if lead := compiled.leads[index]; lead != noWarningCommentsNoLead &&
				(!hasFirst || !noWarningCommentsFoldEqual(lead, first)) {
				continue
			}
			if !matcher.MatchString(value) {
				continue
			}
			if !selfDirectiveChecked {
				// A directive comment configuring THIS rule is exempt, so `/* eslint
				// no-warning-comments: "error" */` does not report itself. Upstream tests the
				// directive shape and the rule's own name together, and both halves matter: an
				// ordinary comment mentioning the rule name is not exempt.
				if noWarningCommentsIsSelfDirective(value) {
					break
				}
				selfDirectiveChecked = true
			}
			ctx.ReportRange(comment.Range, rule.Message{
				Id: "unexpectedComment",
				Description: fmt.Sprintf(
					"This comment opens with `%s`, which marks the code as knowingly "+
						"unfinished: %q. A warning comment is a note to a future reader that "+
						"something was deferred, so it should be tracked somewhere that gets "+
						"read rather than left where only the next editor of this file will "+
						"find it.",
					compiled.terms[index], noWarningCommentsQuote(value)),
			})
		}
	}
}

// noWarningCommentsValueOf returns the comment body with its delimiters removed.
//
// Upstream's `node.value`. A line comment loses its two slashes; a block comment loses `/*` and
// `*/` and NOTHING else, which is why an inner asterisk survives into the match and blocks the
// Start anchor for `/** TODO */`.
func noWarningCommentsValueOf(comment comments.Comment) string {
	text := comment.Text
	if comment.IsBlock {
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		return text
	}
	return strings.TrimPrefix(text, "//")
}

// noWarningCommentsSelfDirective matches a directive comment configuring this rule.
var noWarningCommentsSelfDirective = regexp.MustCompile(`\bno-warning-comments\b`)

// noWarningCommentsDirectivePrefixes are the words that make a comment a directive to the linter.
//
// Upstream's `isDirectiveComment` also accepts `globals`, `exported` and the `eslint-*` family. The
// exemption only fires when the comment ALSO names this rule, so the list only has to be wide
// enough to cover the shapes that can configure it.
var noWarningCommentsDirectivePrefixes = []string{
	"eslint", "eslint-disable", "eslint-enable", "eslint-disable-line",
	"eslint-disable-next-line", "globals", "exported",
}

// noWarningCommentsIsSelfDirective answers whether a comment configures this rule.
//
// Both halves are required, which is upstream's `isDirectiveComment(node) && selfConfigRegEx`. A
// comment that merely mentions the rule name in prose is not a directive and is not exempt, and a
// directive for a different rule is not exempt either.
func noWarningCommentsIsSelfDirective(value string) bool {
	if !noWarningCommentsSelfDirective.MatchString(value) {
		return false
	}
	trimmed := text.TrimWhitespace(value)
	for _, prefix := range noWarningCommentsDirectivePrefixes {
		if trimmed == prefix {
			return true
		}
		if strings.HasPrefix(trimmed, prefix) {
			// The character after the prefix has to be a separator, or `eslintfoo` would count.
			rest := trimmed[len(prefix):]
			if rest == "" || rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\n' ||
				rest[0] == '\r' {
				return true
			}
		}
	}
	return false
}

// noWarningCommentsMatcherFor builds the pattern for one term, which is upstream's
// `convertToRegExp`.
//
// Three pieces, and each is a decision rather than a detail:
//
//	the prefix   under Start it is `^[\s<decoration>]*`, so leading whitespace and any configured
//	             decoration characters are skipped and the term still counts as first. Under
//	             Anywhere it is a word boundary, but only when the term BEGINS with a word
//	             character -- a term written `!FIX` has no boundary to assert on its left.
//	the term     escaped, so a term containing regex punctuation matches literally.
//	the suffix   a word boundary, again only when the term ENDS with a word character.
//
// The flags are case-insensitive. Upstream also passes `u`, which for this pattern changes only how
// `\b` is computed, and both engines compute it over ASCII word characters -- measured across six
// non-ASCII shapes rather than assumed, since a difference there would move every term adjacent to
// an accented letter.
func noWarningCommentsMatcherFor(
	term string,
	location NoWarningCommentsLocation,
	decoration []string,
) (*regexp.Regexp, error) {
	escaped := regexp.QuoteMeta(term)

	prefix := ""
	switch {
	case location == NoWarningCommentsStart:
		prefix = "^[\\s" + regexp.QuoteMeta(strings.Join(decoration, "")) + "]*"
	case noWarningCommentsBeginsWithWordCharacter(term):
		prefix = "\\b"
	}

	suffix := ""
	if noWarningCommentsEndsWithWordCharacter(term) {
		suffix = "\\b"
	}

	return regexp.Compile("(?i)" + prefix + escaped + suffix)
}

// noWarningCommentsBeginsWithWordCharacter is upstream's `/^\w/u` test on the term.
func noWarningCommentsBeginsWithWordCharacter(term string) bool {
	return term != "" && noWarningCommentsIsWordCharacter(term[0])
}

// noWarningCommentsEndsWithWordCharacter is upstream's `/\w$/u` test on the term.
func noWarningCommentsEndsWithWordCharacter(term string) bool {
	return term != "" && noWarningCommentsIsWordCharacter(term[len(term)-1])
}

// noWarningCommentsIsWordCharacter is the ASCII `\w` class.
//
// Byte-wise rather than rune-wise on purpose. JavaScript's `\w` is ASCII even under the `u` flag,
// so a term beginning with a non-ASCII letter gets NO word boundary asserted, and testing the first
// RUNE against a Unicode letter class would add one. Measured: `ТODO` spelled with a Cyrillic Te is
// clean upstream, which is only true if the boundary is ASCII on both sides.
func noWarningCommentsIsWordCharacter(character byte) bool {
	switch {
	case character >= 'a' && character <= 'z':
		return true
	case character >= 'A' && character <= 'Z':
		return true
	case character >= '0' && character <= '9':
		return true
	case character == '_':
		return true
	}
	return false
}

// noWarningCommentsQuote renders the comment for the message, truncated the way upstream truncates.
//
// Words are appended one at a time and the loop stops before exceeding forty characters, so the
// quote never splits a word. A comment whose first word already exceeds the limit therefore quotes
// as the empty string plus an ellipsis, which looks like a defect and is upstream's behaviour.
//
// The split is on runs of whitespace, so a comment spanning several lines collapses to one line in
// the message.
func noWarningCommentsQuote(value string) string {
	var quoted strings.Builder
	truncated := false
	for _, word := range text.WhitespaceFields(value) {
		candidate := word
		if quoted.Len() > 0 {
			candidate = quoted.String() + " " + word
		}
		if len(candidate) > noWarningCommentsCharacterLimit {
			truncated = true
			break
		}
		quoted.Reset()
		quoted.WriteString(candidate)
	}
	if truncated {
		return quoted.String() + "..."
	}
	return quoted.String()
}
