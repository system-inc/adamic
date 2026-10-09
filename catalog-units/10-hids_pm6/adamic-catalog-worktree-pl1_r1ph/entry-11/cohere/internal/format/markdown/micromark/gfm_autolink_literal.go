package micromark

import "strings"

// gfmAutolinkLiteralExtension is micromark-extension-gfm-autolink-literal/lib/syntax.js (2.1.0).
//
// This version registers on codes only: there is no `null` (any code) key and no resolveAll. The start is
// gated by each construct's `previous`, which the text tokenizer consults before trying a construct, and
// by the tokenize functions re-checking it against self.Previous.

// The extension's token types.
const (
	typeLiteralAutolink      = "literalAutolink"
	typeLiteralAutolinkEmail = "literalAutolinkEmail"
	typeLiteralAutolinkHttp  = "literalAutolinkHttp"
	typeLiteralAutolinkWww   = "literalAutolinkWww"
)

var gfmAutolinkLiteralWwwPrefix = &Construct{Tokenize: tokenizeGfmAutolinkLiteralWwwPrefix, Partial: true}

var gfmAutolinkLiteralDomain = &Construct{Tokenize: tokenizeGfmAutolinkLiteralDomain, Partial: true}

var gfmAutolinkLiteralPath = &Construct{Tokenize: tokenizeGfmAutolinkLiteralPath, Partial: true}

var gfmAutolinkLiteralTrail = &Construct{Tokenize: tokenizeGfmAutolinkLiteralTrail, Partial: true}

var gfmAutolinkLiteralEmailDomainDotTrail = &Construct{
	Tokenize: tokenizeGfmAutolinkLiteralEmailDomainDotTrail,
	Partial:  true,
}

var gfmAutolinkLiteralWwwAutolink = &Construct{
	Name:     "wwwAutolink",
	Tokenize: tokenizeGfmAutolinkLiteralWwwAutolink,
	Previous: func(self *Self, code Code) bool { return gfmAutolinkLiteralPreviousWww(code) },
}

var gfmAutolinkLiteralProtocolAutolink = &Construct{
	Name:     "protocolAutolink",
	Tokenize: tokenizeGfmAutolinkLiteralProtocolAutolink,
	Previous: func(self *Self, code Code) bool { return gfmAutolinkLiteralPreviousProtocol(code) },
}

var gfmAutolinkLiteralEmailAutolink = &Construct{
	Name:     "emailAutolink",
	Tokenize: tokenizeGfmAutolinkLiteralEmailAutolink,
	Previous: func(self *Self, code Code) bool { return gfmAutolinkLiteralPreviousEmail(code) },
}

// gfmAutolinkLiteralText is upstream's module-level `text` record. Upstream stores a lone construct for
// most codes and a list for four; combineExtensions wraps a lone one in a list, so lists here are the same.
var gfmAutolinkLiteralText = func() map[Code][]*Construct {
	text := map[Code][]*Construct{}

	code := Code(48)

	// Add alphanumerics.
	for code < 123 {
		text[code] = []*Construct{gfmAutolinkLiteralEmailAutolink}
		code++
		if code == 58 {
			code = 65
		} else if code == 91 {
			code = 97
		}
	}
	text[CodePlusSign] = []*Construct{gfmAutolinkLiteralEmailAutolink}
	text[CodeDash] = []*Construct{gfmAutolinkLiteralEmailAutolink}
	text[CodeDot] = []*Construct{gfmAutolinkLiteralEmailAutolink}
	text[CodeUnderscore] = []*Construct{gfmAutolinkLiteralEmailAutolink}
	text[CodeUppercaseH] = []*Construct{gfmAutolinkLiteralEmailAutolink, gfmAutolinkLiteralProtocolAutolink}
	text[CodeLowercaseH] = []*Construct{gfmAutolinkLiteralEmailAutolink, gfmAutolinkLiteralProtocolAutolink}
	text[CodeUppercaseW] = []*Construct{gfmAutolinkLiteralEmailAutolink, gfmAutolinkLiteralWwwAutolink}
	text[CodeLowercaseW] = []*Construct{gfmAutolinkLiteralEmailAutolink, gfmAutolinkLiteralWwwAutolink}

	return text
}()

// To do: perform email autolink literals on events, afterwards.
// That’s where `markdown-rs` and `cmark-gfm` perform it.
// It should look for `@`, then for atext backwards, and then for a label
// forwards.
// To do: `mailto:`, `xmpp:` protocol as prefix.

func gfmAutolinkLiteralExtension() *Extension {
	return &Extension{Text: &ConstructRecord{ByCode: gfmAutolinkLiteralText}}
}

// tokenizeGfmAutolinkLiteralEmailAutolink is an email autolink literal: `a contact@example.org b`.
//
// It is tried at nearly every position of text, and nearly every try fails its first check, so its state is
// an emailAutolink from the parse's Memory, whose states were bound when the slot was made, and a try
// allocates nothing (#93dpede, #vbjv3d6).
func tokenizeGfmAutolinkLiteralEmailAutolink(self *Self, effects *Effects, ok State, nok State) State {
	run := effects.memory().emailAutolink()
	run.self, run.effects, run.ok, run.nok = self, effects, ok, nok
	return run.startState
}

// emailAutolink is one email autolink literal: the closure state upstream keeps in variables, and its
// states, each made once per slot.
type emailAutolink struct {
	self    *Self
	effects *Effects
	ok, nok State

	dot, data bool

	startState, atextState, emailDomainState, emailDomainDotState, emailDomainAfterState State
}

func newEmailAutolink() *emailAutolink {
	run := &emailAutolink{}
	run.startState, run.atextState = run.start, run.atext
	run.emailDomainState, run.emailDomainDotState, run.emailDomainAfterState = run.emailDomain, run.emailDomainDot, run.emailDomainAfter
	return run
}

// release zeroes the autolink for its next parse, keeping its bound states.
func (run *emailAutolink) release() {
	*run = emailAutolink{
		startState: run.startState, atextState: run.atextState, emailDomainState: run.emailDomainState,
		emailDomainDotState: run.emailDomainDotState, emailDomainAfterState: run.emailDomainAfterState,
	}
}

// Start of email autolink literal.
func (run *emailAutolink) start(code Code) State {
	if !gfmAutolinkLiteralAtext(code) || !gfmAutolinkLiteralPreviousEmail(run.self.Previous) ||
		gfmAutolinkLiteralPreviousUnbalanced(run.self.Events) {
		return run.nok(code)
	}
	run.effects.Enter(typeLiteralAutolink, nil)
	run.effects.Enter(typeLiteralAutolinkEmail, nil)
	return run.atext(code)
}

// In email atext.
func (run *emailAutolink) atext(code Code) State {
	if gfmAutolinkLiteralAtext(code) {
		run.effects.Consume(code)
		return run.atextState
	}
	if code == CodeAtSign {
		run.effects.Consume(code)
		return run.emailDomainState
	}
	return run.nok(code)
}

// In email domain.
//
// The reference code is a bit overly complex as it handles the `@`, of which
// there may be just one.
// Source: <https://github.com/github/cmark-gfm/blob/ef1cfcb/extensions/autolink.c#L318>
func (run *emailAutolink) emailDomain(code Code) State {
	// Dot followed by alphanumerical (not `-` or `_`).
	if code == CodeDot {
		return run.effects.Check(gfmAutolinkLiteralEmailDomainDotTrail, run.emailDomainAfterState, run.emailDomainDotState)(code)
	}

	// Alphanumerical, `-`, and `_`.
	if code == CodeDash || code == CodeUnderscore || asciiAlphanumeric(code) {
		run.data = true
		run.effects.Consume(code)
		return run.emailDomainState
	}

	// To do: `/` if xmpp.

	// Note: normally we’d truncate trailing punctuation from the link.
	// However, email autolink literals cannot contain any of those markers,
	// except for `.`, but that can only occur if it isn’t trailing.
	// So we can ignore truncating!
	return run.emailDomainAfter(code)
}

// In email domain, on dot that is not a trail.
func (run *emailAutolink) emailDomainDot(code Code) State {
	run.effects.Consume(code)
	run.dot = true
	return run.emailDomainState
}

// After email domain.
func (run *emailAutolink) emailDomainAfter(code Code) State {
	// Domain must not be empty, must include a dot, and must end in alphabetical.
	// Source: <https://github.com/github/cmark-gfm/blob/ef1cfcb/extensions/autolink.c#L332>.
	if run.data && run.dot && asciiAlpha(run.self.Previous) {
		run.effects.Exit(typeLiteralAutolinkEmail)
		run.effects.Exit(typeLiteralAutolink)
		return run.ok(code)
	}
	return run.nok(code)
}

// tokenizeGfmAutolinkLiteralWwwAutolink is a `www` autolink literal: `a www.example.org b`.
func tokenizeGfmAutolinkLiteralWwwAutolink(self *Self, effects *Effects, ok State, nok State) State {
	var wwwStart, wwwAfter State

	// Start of www autolink literal.
	wwwStart = func(code Code) State {
		if (code != CodeUppercaseW && code != CodeLowercaseW) || !gfmAutolinkLiteralPreviousWww(self.Previous) ||
			gfmAutolinkLiteralPreviousUnbalanced(self.Events) {
			return nok(code)
		}
		effects.Enter(typeLiteralAutolink, nil)
		effects.Enter(typeLiteralAutolinkWww, nil)
		// Note: we *check*, so we can discard the `www.` we parsed.
		// If it worked, we consider it as a part of the domain.
		return effects.Check(
			gfmAutolinkLiteralWwwPrefix,
			effects.Attempt(gfmAutolinkLiteralDomain, effects.Attempt(gfmAutolinkLiteralPath, wwwAfter, nil), nok),
			nok,
		)(code)
	}

	// After a www autolink literal.
	wwwAfter = func(code Code) State {
		effects.Exit(typeLiteralAutolinkWww)
		effects.Exit(typeLiteralAutolink)
		return ok(code)
	}

	return wwwStart
}

// tokenizeGfmAutolinkLiteralProtocolAutolink is a protocol autolink literal: `a https://example.org b`.
func tokenizeGfmAutolinkLiteralProtocolAutolink(self *Self, effects *Effects, ok State, nok State) State {
	// Only ASCII letters ever reach the buffer, so a byte is one UTF-16 code unit and len is `.length`.
	buffer := ""
	seen := false

	var protocolStart, protocolPrefixInside, protocolSlashesInside, afterProtocol, protocolAfter State

	// Start of protocol autolink literal.
	protocolStart = func(code Code) State {
		if (code == CodeUppercaseH || code == CodeLowercaseH) && gfmAutolinkLiteralPreviousProtocol(self.Previous) &&
			!gfmAutolinkLiteralPreviousUnbalanced(self.Events) {
			effects.Enter(typeLiteralAutolink, nil)
			effects.Enter(typeLiteralAutolinkHttp, nil)
			buffer += string(rune(code))
			effects.Consume(code)
			return protocolPrefixInside
		}
		return nok(code)
	}

	// In protocol.
	protocolPrefixInside = func(code Code) State {
		// `5` is size of `https`
		if asciiAlpha(code) && len(buffer) < 5 {
			buffer += string(rune(code))
			effects.Consume(code)
			return protocolPrefixInside
		}
		if code == CodeColon {
			protocol := strings.ToLower(buffer)
			if protocol == "http" || protocol == "https" {
				effects.Consume(code)
				return protocolSlashesInside
			}
		}
		return nok(code)
	}

	// In slashes.
	protocolSlashesInside = func(code Code) State {
		if code == CodeSlash {
			effects.Consume(code)
			if seen {
				return afterProtocol
			}
			seen = true
			return protocolSlashesInside
		}
		return nok(code)
	}

	// After protocol, before domain.
	afterProtocol = func(code Code) State {
		// To do: this is different from `markdown-rs`:
		// https://github.com/wooorm/markdown-rs/blob/b3a921c761309ae00a51fe348d8a43adbc54b518/src/construct/gfm_autolink_literal.rs#L172-L182
		if code == CodeEof || asciiControl(code) || markdownLineEndingOrSpace(code) || unicodeWhitespace(code) ||
			unicodePunctuation(code) {
			return nok(code)
		}
		return effects.Attempt(
			gfmAutolinkLiteralDomain,
			effects.Attempt(gfmAutolinkLiteralPath, protocolAfter, nil),
			nok,
		)(code)
	}

	// After a protocol autolink literal.
	protocolAfter = func(code Code) State {
		effects.Exit(typeLiteralAutolinkHttp)
		effects.Exit(typeLiteralAutolink)
		return ok(code)
	}

	return protocolStart
}

// tokenizeGfmAutolinkLiteralWwwPrefix is the `www.` prefix.
func tokenizeGfmAutolinkLiteralWwwPrefix(self *Self, effects *Effects, ok State, nok State) State {
	size := 0

	var wwwPrefixInside, wwwPrefixAfter State

	// In www prefix.
	wwwPrefixInside = func(code Code) State {
		if (code == CodeUppercaseW || code == CodeLowercaseW) && size < 3 {
			size++
			effects.Consume(code)
			return wwwPrefixInside
		}
		if code == CodeDot && size == 3 {
			effects.Consume(code)
			return wwwPrefixAfter
		}
		return nok(code)
	}

	// After www prefix.
	wwwPrefixAfter = func(code Code) State {
		// If there is *anything*, we can link.
		if code == CodeEof {
			return nok(code)
		}
		return ok(code)
	}

	return wwwPrefixInside
}

// tokenizeGfmAutolinkLiteralDomain is a domain: `example.org` in `https://example.org`.
func tokenizeGfmAutolinkLiteralDomain(self *Self, effects *Effects, ok State, nok State) State {
	underscoreInLastSegment := false
	underscoreInLastLastSegment := false
	seen := false

	var domainInside, domainAtPunctuation, domainAfter State

	// In domain.
	domainInside = func(code Code) State {
		// Check whether this marker, which is a trailing punctuation
		// marker, optionally followed by more trailing markers, and then
		// followed by an end.
		if code == CodeDot || code == CodeUnderscore {
			return effects.Check(gfmAutolinkLiteralTrail, domainAfter, domainAtPunctuation)(code)
		}

		// GH documents that only alphanumerics (other than `-`, `.`, and `_`) can
		// occur, which sounds like ASCII only, but they also support `www.點看.com`,
		// so that’s Unicode.
		// Instead of some new production for Unicode alphanumerics, markdown
		// already has that for Unicode punctuation and whitespace, so use those.
		// Source: <https://github.com/github/cmark-gfm/blob/ef1cfcb/extensions/autolink.c#L12>.
		if code == CodeEof || markdownLineEndingOrSpace(code) || unicodeWhitespace(code) ||
			(code != CodeDash && unicodePunctuation(code)) {
			return domainAfter(code)
		}
		seen = true
		effects.Consume(code)
		return domainInside
	}

	// In domain, at potential trailing punctuation, that was not trailing.
	domainAtPunctuation = func(code Code) State {
		// There is an underscore in the last segment of the domain
		if code == CodeUnderscore {
			underscoreInLastSegment = true
		} else {
			// Otherwise, it’s a `.`: save the last segment underscore in the
			// penultimate segment slot.
			underscoreInLastLastSegment = underscoreInLastSegment
			underscoreInLastSegment = false
		}
		effects.Consume(code)
		return domainInside
	}

	// After domain.
	domainAfter = func(code Code) State {
		// Note: that’s GH says a dot is needed, but it’s not true:
		// <https://github.com/github/cmark-gfm/issues/279>
		if underscoreInLastLastSegment || underscoreInLastSegment || !seen {
			return nok(code)
		}
		return ok(code)
	}

	return domainInside
}

// tokenizeGfmAutolinkLiteralPath is a path: `/stuff` in `https://example.org/stuff`. Upstream's path
// tokenizer takes no nok: it always succeeds.
func tokenizeGfmAutolinkLiteralPath(self *Self, effects *Effects, ok State, nok State) State {
	sizeOpen := 0
	sizeClose := 0

	var pathInside, pathAtPunctuation State

	// In path.
	pathInside = func(code Code) State {
		if code == CodeLeftParenthesis {
			sizeOpen++
			effects.Consume(code)
			return pathInside
		}

		// To do: `markdown-rs` also needs this.
		// If this is a paren, and there are less closings than openings,
		// we don’t check for a trail.
		if code == CodeRightParenthesis && sizeClose < sizeOpen {
			return pathAtPunctuation(code)
		}

		// Check whether this trailing punctuation marker is optionally
		// followed by more trailing markers, and then followed
		// by an end.
		if code == CodeExclamationMark || code == CodeQuotationMark || code == CodeAmpersand ||
			code == CodeApostrophe || code == CodeRightParenthesis || code == CodeAsterisk || code == CodeComma ||
			code == CodeDot || code == CodeColon || code == CodeSemicolon || code == CodeLessThan ||
			code == CodeQuestionMark || code == CodeRightSquareBracket || code == CodeUnderscore || code == CodeTilde {
			return effects.Check(gfmAutolinkLiteralTrail, ok, pathAtPunctuation)(code)
		}
		if code == CodeEof || markdownLineEndingOrSpace(code) || unicodeWhitespace(code) {
			return ok(code)
		}
		effects.Consume(code)
		return pathInside
	}

	// In path, at potential trailing punctuation, that was not trailing.
	pathAtPunctuation = func(code Code) State {
		// Count closing parens.
		if code == CodeRightParenthesis {
			sizeClose++
		}
		effects.Consume(code)
		return pathInside
	}

	return pathInside
}

// tokenizeGfmAutolinkLiteralTrail is the trail of a domain or path.
//
// This calls `ok` if this *is* the trail, followed by an end, which means
// the entire trail is not part of the link.
// It calls `nok` if this *is* part of the link.
func tokenizeGfmAutolinkLiteralTrail(self *Self, effects *Effects, ok State, nok State) State {
	var trail, trailBracketAfter, trailCharacterReferenceStart, trailCharacterReferenceInside State

	// In trail of domain or path.
	trail = func(code Code) State {
		// Regular trailing punctuation.
		if code == CodeExclamationMark || code == CodeQuotationMark || code == CodeApostrophe ||
			code == CodeRightParenthesis || code == CodeAsterisk || code == CodeComma || code == CodeDot ||
			code == CodeColon || code == CodeSemicolon || code == CodeQuestionMark || code == CodeUnderscore ||
			code == CodeTilde {
			effects.Consume(code)
			return trail
		}

		// `&` followed by one or more alphabeticals and then a `;`, is
		// as a whole considered as trailing punctuation.
		// In all other cases, it is considered as continuation of the URL.
		if code == CodeAmpersand {
			effects.Consume(code)
			return trailCharacterReferenceStart
		}

		// Needed because we allow literals after `[`, as we fix:
		// <https://github.com/github/cmark-gfm/issues/278>.
		// Check that it is not followed by `(` or `[`.
		if code == CodeRightSquareBracket {
			effects.Consume(code)
			return trailBracketAfter
		}
		if
		// `<` is an end.
		code == CodeLessThan ||
			// So is whitespace.
			code == CodeEof || markdownLineEndingOrSpace(code) || unicodeWhitespace(code) {
			return ok(code)
		}
		return nok(code)
	}

	// In trail, after `]`.
	//
	// > 👉 **Note**: this deviates from `cmark-gfm` to fix a bug.
	// > See end of <https://github.com/github/cmark-gfm/issues/278> for more.
	trailBracketAfter = func(code Code) State {
		// Whitespace or something that could start a resource or reference is the end.
		// Switch back to trail otherwise.
		if code == CodeEof || code == CodeLeftParenthesis || code == CodeLeftSquareBracket ||
			markdownLineEndingOrSpace(code) || unicodeWhitespace(code) {
			return ok(code)
		}
		return trail(code)
	}

	// In character-reference like trail, after `&`.
	trailCharacterReferenceStart = func(code Code) State {
		// When non-alpha, it’s not a trail.
		if asciiAlpha(code) {
			return trailCharacterReferenceInside(code)
		}
		return nok(code)
	}

	// In character-reference like trail.
	trailCharacterReferenceInside = func(code Code) State {
		// Switch back to trail if this is well-formed.
		if code == CodeSemicolon {
			effects.Consume(code)
			return trail
		}
		if asciiAlpha(code) {
			effects.Consume(code)
			return trailCharacterReferenceInside
		}

		// It’s not a trail.
		return nok(code)
	}

	return trail
}

// tokenizeGfmAutolinkLiteralEmailDomainDotTrail is a dot in an email domain trail.
//
// This calls `ok` if this *is* the trail, followed by an end, which means
// the trail is not part of the link.
// It calls `nok` if this *is* part of the link.
func tokenizeGfmAutolinkLiteralEmailDomainDotTrail(self *Self, effects *Effects, ok State, nok State) State {
	var start, after State

	// Dot.
	start = func(code Code) State {
		// Must be dot.
		effects.Consume(code)
		return after
	}

	// After dot.
	after = func(code Code) State {
		// Not a trail if alphanumeric.
		if asciiAlphanumeric(code) {
			return nok(code)
		}
		return ok(code)
	}

	return start
}

// gfmAutolinkLiteralPreviousWww is upstream's previousWww. See:
// <https://github.com/github/cmark-gfm/blob/ef1cfcb/extensions/autolink.c#L156>.
func gfmAutolinkLiteralPreviousWww(code Code) bool {
	return code == CodeEof || code == CodeLeftParenthesis || code == CodeAsterisk || code == CodeUnderscore ||
		code == CodeLeftSquareBracket || code == CodeRightSquareBracket || code == CodeTilde ||
		markdownLineEndingOrSpace(code)
}

// gfmAutolinkLiteralPreviousProtocol is upstream's previousProtocol. See:
// <https://github.com/github/cmark-gfm/blob/ef1cfcb/extensions/autolink.c#L214>.
func gfmAutolinkLiteralPreviousProtocol(code Code) bool {
	return !asciiAlpha(code)
}

// gfmAutolinkLiteralPreviousEmail is upstream's previousEmail.
func gfmAutolinkLiteralPreviousEmail(code Code) bool {
	// Do not allow a slash “inside” atext.
	// The reference code is a bit weird, but that’s what it results in.
	// Source: <https://github.com/github/cmark-gfm/blob/ef1cfcb/extensions/autolink.c#L307>.
	// Other than slash, every preceding character is allowed.
	return !(code == CodeSlash || gfmAutolinkLiteralAtext(code))
}

func gfmAutolinkLiteralAtext(code Code) bool {
	return code == CodePlusSign || code == CodeDash || code == CodeDot || code == CodeUnderscore ||
		asciiAlphanumeric(code)
}

// gfmAutolinkLiteralPreviousUnbalanced is upstream's previousUnbalanced: whether an unbalanced label
// start comes before, walking back only as far as the last token a previous walk already cleared.
func gfmAutolinkLiteralPreviousUnbalanced(events []Event) bool {
	index := len(events)
	result := false

	for index > 0 {
		index--
		token := events[index].Token
		if (token.Type == TypeLabelLink || token.Type == TypeLabelImage) && !token.balanced {
			result = true
			break
		}

		// If we’ve seen this token, and it was marked as not having any unbalanced
		// bracket before it, we can exit.
		if token.gfmAutolinkLiteralWalkedInto {
			result = false
			break
		}
	}

	if len(events) > 0 && !result {
		// Mark the last token as “walked into” w/o finding
		// anything.
		events[len(events)-1].Token.gfmAutolinkLiteralWalkedInto = true
	}

	return result
}
