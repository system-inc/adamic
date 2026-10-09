package micromark

import (
	"slices"
	"unicode/utf16"
)

// attention is micromark-core-commonmark/lib/attention.js: sequences of `*` or `_` that resolveAll pairs
// into emphasis and strong.
var attention = &Construct{Name: "attention", ResolveAll: &Resolver{Resolve: resolveAllAttention}, Tokenize: tokenizeAttention}

// typeAttentionSequence is upstream's `attentionSequence`, missing from the generated symbols. It never
// survives resolveAll: an unpaired sequence becomes `data`.
const typeAttentionSequence = "attentionSequence"

// resolveAllAttention takes all events and resolves attention to emphasis or strong.
func resolveAllAttention(events []Event, context *TokenizeContext) []Event {
	index := -1
	var open int
	var group, text, openingSequence, closingSequence *Token
	var use int
	var nextEvents []Event
	var offset int

	// Walk through all events.
	//
	// Note: performance of this is fine on an mb of normal markdown, but it’s
	// a bottleneck for malicious stuff.
	for index+1 < len(events) {
		index++

		// Find a token that can close.
		if events[index].Enter && events[index].Token.Type == typeAttentionSequence && events[index].Token.close {
			open = index

			// Now walk back to find an opener. Upstream's `while (open--)` tests, then decrements.
			for open > 0 {
				open--

				// Find a token that can open the closer.
				if !events[open].Enter && events[open].Token.Type == typeAttentionSequence && events[open].Token.open &&
					// If the markers are the same:
					firstCodeUnit(context, events[open].Token) == firstCodeUnit(context, events[index].Token) {
					openSize := events[open].Token.End.Offset - events[open].Token.Start.Offset
					closeSize := events[index].Token.End.Offset - events[index].Token.Start.Offset

					// If the opening can close or the closing can open,
					// and the close size *is not* a multiple of three,
					// but the sum of the opening and closing size *is* multiple of three,
					// then don’t match.
					if (events[open].Token.close || events[index].Token.open) && closeSize%3 != 0 && (openSize+closeSize)%3 == 0 {
						continue
					}

					// Number of markers to use from the sequence.
					if openSize > 1 && closeSize > 1 {
						use = 2
					} else {
						use = 1
					}

					// Points are values here, so every `{...point}` copy upstream makes is an assignment.
					start := events[open].Token.End
					end := events[index].Token.Start
					movePoint(&start, -use)
					movePoint(&end, use)

					sequenceType, textType, groupType := TypeEmphasisSequence, TypeEmphasisText, TypeEmphasis
					if use > 1 {
						sequenceType, textType, groupType = TypeStrongSequence, TypeStrongText, TypeStrong
					}

					openingSequence = &Token{Type: sequenceType, Start: start, End: events[open].Token.End}
					closingSequence = &Token{Type: sequenceType, Start: events[index].Token.Start, End: end}
					text = &Token{Type: textType, Start: events[open].Token.End, End: events[index].Token.Start}
					group = &Token{Type: groupType, Start: openingSequence.Start, End: closingSequence.End}

					events[open].Token.End = openingSequence.Start
					events[index].Token.Start = closingSequence.End

					nextEvents = nil

					// If there are more markers in the opening, add them before.
					if events[open].Token.End.Offset-events[open].Token.Start.Offset != 0 {
						nextEvents = pushEvents(nextEvents, []Event{
							{Enter: true, Token: events[open].Token, Context: context},
							{Enter: false, Token: events[open].Token, Context: context},
						})
					}

					// Opening.
					nextEvents = pushEvents(nextEvents, []Event{
						{Enter: true, Token: group, Context: context},
						{Enter: true, Token: openingSequence, Context: context},
						{Enter: false, Token: openingSequence, Context: context},
						{Enter: true, Token: text, Context: context},
					})

					// Always populated by defaults.

					// Between. Upstream's `events.slice` is a copy, and the inner resolvers splice it in place.
					nextEvents = pushEvents(nextEvents, resolveAll(context.Parser.Constructs.InsideSpan, slices.Clone(events[open+1:index]), context))

					// Closing.
					nextEvents = pushEvents(nextEvents, []Event{
						{Enter: false, Token: text, Context: context},
						{Enter: true, Token: closingSequence, Context: context},
						{Enter: false, Token: closingSequence, Context: context},
						{Enter: false, Token: group, Context: context},
					})

					// If there are more markers in the closing, add them after.
					if events[index].Token.End.Offset-events[index].Token.Start.Offset != 0 {
						offset = 2
						nextEvents = pushEvents(nextEvents, []Event{
							{Enter: true, Token: events[index].Token, Context: context},
							{Enter: false, Token: events[index].Token, Context: context},
						})
					} else {
						offset = 0
					}

					events = spliceEvents(events, open-1, index-open+3, nextEvents)

					index = open + len(nextEvents) - offset - 2
					break
				}
			}
		}
	}

	// Remove remaining sequences.
	for _, event := range events {
		if event.Token.Type == typeAttentionSequence {
			event.Token.Type = TypeData
		}
	}

	return events
}

// firstCodeUnit is `context.sliceSerialize(token).charCodeAt(0)`: the first UTF-16 code unit of the
// token's text, or -1 for upstream's NaN when the text is empty, which equals nothing.
func firstCodeUnit(context *TokenizeContext, token *Token) int {
	serialized := context.SliceSerialize(token, false)
	for _, character := range serialized {
		return int(utf16.Encode([]rune{character})[0])
	}
	return -1
}

func tokenizeAttention(self *Self, effects *Effects, ok State, nok State) State {
	attentionMarkers := self.Parser.Constructs.AttentionMarkers
	previous := self.Previous
	before := classifyCharacter(previous)

	var marker Code

	var start, inside State

	// Before a sequence.
	//
	// ```markdown
	// > | **
	//     ^
	// ```
	start = func(code Code) State {
		marker = code
		effects.Enter(typeAttentionSequence, nil)
		return inside(code)
	}

	// In a sequence.
	//
	// ```markdown
	// > | **
	//     ^^
	// ```
	inside = func(code Code) State {
		if code == marker {
			effects.Consume(code)
			return inside
		}

		token := effects.Exit(typeAttentionSequence)

		// To do: next major: move this to resolver, just like `markdown-rs`.
		after := classifyCharacter(code)

		// Always populated by defaults.

		// Upstream's `||` and `&&` over the 0, 1 or 2 of classifyCharacter, read as booleans. Upstream
		// names these open and close; opens and closes keep Go's builtin close unshadowed.
		opens := after == 0 || (after == attentionSideAfter && before != 0) || slices.Contains(attentionMarkers, code)
		closes := before == 0 || (before == attentionSideAfter && after != 0) || slices.Contains(attentionMarkers, previous)

		if marker == CodeAsterisk {
			token.open = opens
			token.close = closes
		} else {
			token.open = opens && (before != 0 || !closes)
			token.close = closes && (after != 0 || !opens)
		}

		return ok(code)
	}

	return start
}

// movePoint moves a point a bit.
//
// Note: `move` only works inside lines! It’s not possible to move past other
// chunks (replacement characters, tabs, or line endings).
func movePoint(point *Point, offset int) {
	point.Column += offset
	point.Offset += offset
	point.bufferIndex += offset
}
