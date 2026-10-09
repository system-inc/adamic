package micromark

// wikiLinkExtension is @braindb/micromark-extension-wiki-link/dist/syntax.js (0.1.0), as the fork's
// parse-markdown.js configures it: `wikiLinkSyntax({ aliasDivider: { charCodeAt: () => Number.NaN } })`.
//
// That divider is not a string but an object whose charCodeAt always answers NaN, and NaN equals nothing,
// so `code === aliasMarker.charCodeAt(aliasCursor)` is false for every code: the alias branch of
// consumeTarget never fires, and consumeAliasMarker and consumeAlias are unreachable. A `|` is then an
// ordinary target character. This port reproduces the configured behavior, not the alias syntax: the
// dead branch is kept as a comment in consumeTarget and the two unreachable states are left out.
func wikiLinkExtension() *Extension {
	return &Extension{
		// Left square bracket.
		Text: &ConstructRecord{ByCode: map[Code][]*Construct{CodeLeftSquareBracket: {wikiLink}}},
	}
}

// wikiLink carries no name upstream (the object is `{ tokenize }`), so `disable` can never turn it off.
var wikiLink = &Construct{Tokenize: tokenizeWikiLink}

// The extension's own token types, upstream's string literals.
const (
	typeWikiLink       = "wikiLink"
	typeWikiLinkMarker = "wikiLinkMarker"
	typeWikiLinkData   = "wikiLinkData"
	typeWikiLinkTarget = "wikiLinkTarget"
)

// wikiLinkStartMarker and wikiLinkEndMarker are upstream's startMarker "[[" and endMarker "]]", as the
// codes `charCodeAt` reads from them.
var (
	wikiLinkStartMarker = []Code{CodeLeftSquareBracket, CodeLeftSquareBracket}
	wikiLinkEndMarker   = []Code{CodeRightSquareBracket, CodeRightSquareBracket}
)

func tokenizeWikiLink(self *Self, effects *Effects, ok State, nok State) State {
	// Upstream's `data` starts undefined and only ever becomes true: falsy until a target character that
	// is not a line ending or space is consumed.
	data := false
	startMarkerCursor := 0
	endMarkerCursor := 0

	var start, consumeStart, consumeData, consumeTarget, consumeEnd State

	start = func(code Code) State {
		if code != wikiLinkStartMarker[startMarkerCursor] {
			return nok(code)
		}
		effects.Enter(typeWikiLink, nil)
		effects.Enter(typeWikiLinkMarker, nil)
		return consumeStart(code)
	}

	consumeStart = func(code Code) State {
		if startMarkerCursor == len(wikiLinkStartMarker) {
			effects.Exit(typeWikiLinkMarker)
			return consumeData(code)
		}
		if code != wikiLinkStartMarker[startMarkerCursor] {
			return nok(code)
		}
		effects.Consume(code)
		startMarkerCursor++
		return consumeStart
	}

	consumeData = func(code Code) State {
		// Upstream's `markdownLineEnding(code) || code === codes.eof`; its own markdownLineEnding is the
		// core one (`code < codes.horizontalTab`, never null).
		if markdownLineEnding(code) || code == CodeEof {
			return nok(code)
		}
		effects.Enter(typeWikiLinkData, nil)
		effects.Enter(typeWikiLinkTarget, nil)
		return consumeTarget(code)
	}

	consumeTarget = func(code Code) State {
		// Upstream first tests `code === aliasMarker.charCodeAt(aliasCursor)` and, on a match, exits the
		// target and enters wikiLinkAliasMarker. The fork's divider answers NaN, which no code equals, so
		// that branch is dead and left out (see wikiLinkExtension).
		if code == wikiLinkEndMarker[endMarkerCursor] {
			if !data {
				return nok(code)
			}
			effects.Exit(typeWikiLinkTarget)
			effects.Exit(typeWikiLinkData)
			effects.Enter(typeWikiLinkMarker, nil)
			return consumeEnd(code)
		}
		if markdownLineEnding(code) || code == CodeEof {
			return nok(code)
		}
		if !markdownLineEndingOrSpace(code) {
			data = true
		}
		effects.Consume(code)
		return consumeTarget
	}

	consumeEnd = func(code Code) State {
		if endMarkerCursor == len(wikiLinkEndMarker) {
			effects.Exit(typeWikiLinkMarker)
			effects.Exit(typeWikiLink)
			return ok(code)
		}
		if code != wikiLinkEndMarker[endMarkerCursor] {
			return nok(code)
		}
		effects.Consume(code)
		endMarkerCursor++
		return consumeEnd
	}

	return start
}
