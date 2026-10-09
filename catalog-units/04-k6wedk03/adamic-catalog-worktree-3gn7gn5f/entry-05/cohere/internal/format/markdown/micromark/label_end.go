package micromark

import "slices"

// labelEnd is micromark-core-commonmark/lib/label-end.js.
var labelEnd = &Construct{
	Name:       "labelEnd",
	ResolveAll: labelEndResolveAll,
	ResolveTo:  &Resolver{Resolve: resolveToLabelEnd},
	Tokenize:   tokenizeLabelEnd,
}

// labelEndResolveAll is the one resolver labelStartLink, labelStartImage and labelEnd share. Upstream's
// label starts say `resolveAll: labelEnd.resolveAll`; resolveAll calls a function once by identity, so
// the three must hold the same pointer.
var labelEndResolveAll = &Resolver{Resolve: resolveAllLabelEnd}

func init() {
	labelStartLink.ResolveAll = labelEndResolveAll
	labelStartImage.ResolveAll = labelEndResolveAll
}

// resourceConstruct, referenceFullConstruct and referenceCollapsedConstruct are not partial upstream, so
// attempting one sets the tokenizer's currentConstruct, as here.
var resourceConstruct = &Construct{Tokenize: tokenizeResource}

var referenceFullConstruct = &Construct{Tokenize: tokenizeReferenceFull}

var referenceCollapsedConstruct = &Construct{Tokenize: tokenizeReferenceCollapsed}

func resolveAllLabelEnd(events []Event, context *TokenizeContext) []Event {
	index := -1
	newEvents := []Event{}

	for {
		index++
		if index >= len(events) {
			break
		}
		token := events[index].Token
		newEvents = append(newEvents, events[index])

		if token.Type == TypeLabelImage || token.Type == TypeLabelLink || token.Type == TypeLabelEnd {
			// Remove the marker.
			offset := 2
			if token.Type == TypeLabelImage {
				offset = 4
			}
			token.Type = TypeData
			index += offset
		}
	}

	// If the events are equal, we don't have to copy newEvents to events
	if len(events) != len(newEvents) {
		events = splice(events, 0, len(events), newEvents)
	}

	return events
}

func resolveToLabelEnd(events []Event, context *TokenizeContext) []Event {
	index := len(events)
	offset := 0
	var token *Token
	// Upstream's open and close start undefined and are tested for truthiness, so an index of 0 reads as
	// unset. -1 is undefined here and `> 0` is the truthiness test.
	open := -1
	close := -1
	var media []Event

	// Find an opening.
	for {
		// `while (index--)`.
		if index == 0 {
			break
		}
		index--

		token = events[index].Token

		if open > 0 {
			// If we see another link, or inactive link label, we’ve been here before.
			if token.Type == TypeLink || (token.Type == TypeLabelLink && token.inactive) {
				break
			}

			// Mark other link openings as inactive, as we can’t have links in
			// links.
			if events[index].Enter && token.Type == TypeLabelLink {
				token.inactive = true
			}
		} else if close > 0 {
			if events[index].Enter &&
				(token.Type == TypeLabelImage || token.Type == TypeLabelLink) &&
				!token.balanced {
				open = index

				if token.Type != TypeLabelLink {
					offset = 2
					break
				}
			}
		} else if token.Type == TypeLabelEnd {
			close = index
		}
	}

	groupType := TypeImage
	if events[open].Token.Type == TypeLabelLink {
		groupType = TypeLink
	}

	// Upstream spreads each point into a fresh object; points are values here.
	group := &Token{
		Type:  groupType,
		Start: events[open].Token.Start,
		End:   events[len(events)-1].Token.End,
	}

	label := &Token{
		Type:  TypeLabel,
		Start: events[open].Token.Start,
		End:   events[close].Token.End,
	}

	text := &Token{
		Type:  TypeLabelText,
		Start: events[open+offset+2].Token.End,
		End:   events[close-2].Token.Start,
	}

	media = []Event{
		{Enter: true, Token: group, Context: context},
		{Enter: true, Token: label, Context: context},
	}

	// Opening marker.
	media = pushEvents(media, slices.Clone(events[open+1:open+offset+3]))

	// Text open.
	media = pushEvents(media, []Event{{Enter: true, Token: text, Context: context}})

	// Always populated by defaults.

	// Between. The slice is a copy upstream, and the resolvers splice in place, so clone it.
	media = pushEvents(media, resolveAll(
		context.Parser.Constructs.InsideSpan,
		slices.Clone(jsSlice(events, open+offset+4, close-3)),
		context,
	))

	// Text close, marker close, label close.
	media = pushEvents(media, []Event{
		{Enter: false, Token: text, Context: context},
		events[close-2],
		events[close-1],
		{Enter: false, Token: label, Context: context},
	})

	// Reference, resource, or so.
	media = pushEvents(media, slices.Clone(events[close+1:]))

	// Media close.
	media = pushEvents(media, []Event{{Enter: false, Token: group, Context: context}})

	return splice(events, open, len(events), media)
}

// jsSlice is Array#slice for non-negative bounds: an end before the start yields an empty list rather
// than a panic.
func jsSlice(events []Event, start int, end int) []Event {
	if end > len(events) {
		end = len(events)
	}
	if start > end {
		return nil
	}
	return events[start:end]
}

func tokenizeLabelEnd(self *Self, effects *Effects, ok State, nok State) State {
	index := len(self.Events)
	var labelStart *Token
	var defined bool

	var start, after, referenceNotFull, labelEndOk, labelEndNok State

	// Find an opening.
	for {
		// `while (index--)`.
		if index == 0 {
			break
		}
		index--

		if (self.Events[index].Token.Type == TypeLabelImage || self.Events[index].Token.Type == TypeLabelLink) &&
			!self.Events[index].Token.balanced {
			labelStart = self.Events[index].Token
			break
		}
	}

	// Start of label end.
	//
	// ```markdown
	// > | [a](b) c
	//       ^
	// > | [a][b] c
	//       ^
	// > | [a][] b
	//       ^
	// > | [a] b
	// ```
	start = func(code Code) State {
		// If there is not an okay opening.
		if labelStart == nil {
			return nok(code)
		}

		// If the corresponding label (link) start is marked as inactive,
		// it means we’d be wrapping a link, like this:
		//
		// ```markdown
		// > | a [b [c](d) e](f) g.
		//                  ^
		// ```
		//
		// We can’t have that, so it’s just balanced brackets.
		if labelStart.inactive {
			return labelEndNok(code)
		}

		defined = slices.Contains(self.Parser.Defined, NormalizeIdentifier(self.SliceSerialize(&Token{
			Start: labelStart.End,
			End:   self.Now(),
		}, false)))
		effects.Enter(TypeLabelEnd, nil)
		effects.Enter(TypeLabelMarker, nil)
		effects.Consume(code)
		effects.Exit(TypeLabelMarker)
		effects.Exit(TypeLabelEnd)
		return after
	}

	// After `]`.
	//
	// ```markdown
	// > | [a](b) c
	//       ^
	// > | [a][b] c
	//       ^
	// > | [a][] b
	//       ^
	// > | [a] b
	//       ^
	// ```
	after = func(code Code) State {
		// Note: `markdown-rs` also parses GFM footnotes here, which for us is in
		// an extension.

		// Resource (`[asd](fgh)`)?
		if code == CodeLeftParenthesis {
			bogus := labelEndNok
			if defined {
				bogus = labelEndOk
			}
			return effects.Attempt(resourceConstruct, labelEndOk, bogus)(code)
		}

		// Full (`[asd][fgh]`) or collapsed (`[asd][]`) reference?
		if code == CodeLeftSquareBracket {
			bogus := labelEndNok
			if defined {
				bogus = referenceNotFull
			}
			return effects.Attempt(referenceFullConstruct, labelEndOk, bogus)(code)
		}

		// Shortcut (`[asd]`) reference?
		if defined {
			return labelEndOk(code)
		}
		return labelEndNok(code)
	}

	// After `]`, at `[`, but not at a full reference.
	//
	// > 👉 **Note**: we only get here if the label is defined.
	//
	// ```markdown
	// > | [a][] b
	//        ^
	// > | [a] b
	//        ^
	// ```
	referenceNotFull = func(code Code) State {
		return effects.Attempt(referenceCollapsedConstruct, labelEndOk, labelEndNok)(code)
	}

	// Done, we found something.
	//
	// ```markdown
	// > | [a](b) c
	//           ^
	// > | [a][b] c
	//           ^
	// > | [a][] b
	//          ^
	// > | [a] b
	//        ^
	// ```
	labelEndOk = func(code Code) State {
		// Note: `markdown-rs` does a bunch of stuff here.
		return ok(code)
	}

	// Done, it’s nothing.
	//
	// There was an okay opening, but we didn’t match anything.
	//
	// ```markdown
	// > | [a](b c
	//        ^
	// > | [a][b c
	//        ^
	// > | [a] b
	//        ^
	// ```
	labelEndNok = func(code Code) State {
		labelStart.balanced = true
		return nok(code)
	}

	return start
}

func tokenizeResource(self *Self, effects *Effects, ok State, nok State) State {
	var resourceStart, resourceBefore, resourceOpen, resourceDestinationAfter, resourceDestinationMissing,
		resourceBetween, resourceTitleAfter, resourceEnd State

	// At a resource.
	//
	// ```markdown
	// > | [a](b) c
	//        ^
	// ```
	resourceStart = func(code Code) State {
		effects.Enter(TypeResource, nil)
		effects.Enter(TypeResourceMarker, nil)
		effects.Consume(code)
		effects.Exit(TypeResourceMarker)
		return resourceBefore
	}

	// In resource, after `(`, at optional whitespace.
	//
	// ```markdown
	// > | [a](b) c
	//         ^
	// ```
	resourceBefore = func(code Code) State {
		if markdownLineEndingOrSpace(code) {
			return factoryWhitespace(effects, resourceOpen)(code)
		}
		return resourceOpen(code)
	}

	// In resource, after optional whitespace, at `)` or a destination.
	//
	// ```markdown
	// > | [a](b) c
	//         ^
	// ```
	resourceOpen = func(code Code) State {
		if code == CodeRightParenthesis {
			return resourceEnd(code)
		}

		return factoryDestination(
			effects,
			resourceDestinationAfter,
			resourceDestinationMissing,
			TypeResourceDestination,
			TypeResourceDestinationLiteral,
			TypeResourceDestinationLiteralMarker,
			TypeResourceDestinationRaw,
			TypeResourceDestinationString,
			linkResourceDestinationBalanceMax,
		)(code)
	}

	// In resource, after destination, at optional whitespace.
	//
	// ```markdown
	// > | [a](b) c
	//          ^
	// ```
	resourceDestinationAfter = func(code Code) State {
		if markdownLineEndingOrSpace(code) {
			return factoryWhitespace(effects, resourceBetween)(code)
		}
		return resourceEnd(code)
	}

	// At invalid destination.
	//
	// ```markdown
	// > | [a](<<) b
	//         ^
	// ```
	resourceDestinationMissing = func(code Code) State {
		return nok(code)
	}

	// In resource, after destination and whitespace, at `(` or title.
	//
	// ```markdown
	// > | [a](b ) c
	//           ^
	// ```
	resourceBetween = func(code Code) State {
		if code == CodeQuotationMark || code == CodeApostrophe || code == CodeLeftParenthesis {
			return factoryTitle(
				effects,
				resourceTitleAfter,
				nok,
				TypeResourceTitle,
				TypeResourceTitleMarker,
				TypeResourceTitleString,
			)(code)
		}

		return resourceEnd(code)
	}

	// In resource, after title, at optional whitespace.
	//
	// ```markdown
	// > | [a](b "c") d
	//              ^
	// ```
	resourceTitleAfter = func(code Code) State {
		if markdownLineEndingOrSpace(code) {
			return factoryWhitespace(effects, resourceEnd)(code)
		}
		return resourceEnd(code)
	}

	// In resource, at `)`.
	//
	// ```markdown
	// > | [a](b) d
	//          ^
	// ```
	resourceEnd = func(code Code) State {
		if code == CodeRightParenthesis {
			effects.Enter(TypeResourceMarker, nil)
			effects.Consume(code)
			effects.Exit(TypeResourceMarker)
			effects.Exit(TypeResource)
			return ok
		}

		return nok(code)
	}

	return resourceStart
}

func tokenizeReferenceFull(self *Self, effects *Effects, ok State, nok State) State {
	var referenceFull, referenceFullAfter, referenceFullMissing State

	// In a reference (full), at the `[`.
	//
	// ```markdown
	// > | [a][b] d
	//        ^
	// ```
	referenceFull = func(code Code) State {
		return factoryLabel(
			self,
			effects,
			referenceFullAfter,
			referenceFullMissing,
			TypeReference,
			TypeReferenceMarker,
			TypeReferenceString,
		)(code)
	}

	// In a reference (full), after `]`.
	//
	// ```markdown
	// > | [a][b] d
	//          ^
	// ```
	referenceFullAfter = func(code Code) State {
		// `.slice(1, -1)` drops the `[` and `]`, one code unit each, so one byte each here too.
		serialized := self.SliceSerialize(self.Events[len(self.Events)-1].Token, false)
		if slices.Contains(self.Parser.Defined, NormalizeIdentifier(serialized[1:len(serialized)-1])) {
			return ok(code)
		}
		return nok(code)
	}

	// In reference (full) that was missing.
	//
	// ```markdown
	// > | [a][b d
	//        ^
	// ```
	referenceFullMissing = func(code Code) State {
		return nok(code)
	}

	return referenceFull
}

func tokenizeReferenceCollapsed(self *Self, effects *Effects, ok State, nok State) State {
	var referenceCollapsedStart, referenceCollapsedOpen State

	// In reference (collapsed), at `[`.
	//
	// > 👉 **Note**: we only get here if the label is defined.
	//
	// ```markdown
	// > | [a][] d
	//        ^
	// ```
	referenceCollapsedStart = func(code Code) State {
		// We only attempt a collapsed label if there’s a `[`.

		effects.Enter(TypeReference, nil)
		effects.Enter(TypeReferenceMarker, nil)
		effects.Consume(code)
		effects.Exit(TypeReferenceMarker)
		return referenceCollapsedOpen
	}

	// In reference (collapsed), at `]`.
	//
	// > 👉 **Note**: we only get here if the label is defined.
	//
	// ```markdown
	// > | [a][] d
	//         ^
	// ```
	referenceCollapsedOpen = func(code Code) State {
		if code == CodeRightSquareBracket {
			effects.Enter(TypeReferenceMarker, nil)
			effects.Consume(code)
			effects.Exit(TypeReferenceMarker)
			effects.Exit(TypeReference)
			return ok
		}

		return nok(code)
	}

	return referenceCollapsedStart
}
