package micromark

// codeText is micromark-core-commonmark/lib/code-text.js.
var codeText = &Construct{
	Name:     "codeText",
	Previous: previousCodeText,
	Resolve:  &Resolver{Resolve: resolveCodeText},
	Tokenize: tokenizeCodeText,
}

// typeSpace is the `space` token code (text) enters between data. Upstream spells it as a literal, not
// a micromark-util-symbol type.
const typeSpace = "space"

// To do: next major: don’t resolve, like `markdown-rs`.
func resolveCodeText(events []Event, context *TokenizeContext) []Event {
	tailExitIndex := len(events) - 4
	headEnterIndex := 3
	var index int
	// enter is upstream's `number | undefined`; -1 is undefined.
	enter := -1

	// If we start and end with an EOL or a space.
	if (events[headEnterIndex].Token.Type == TypeLineEnding || events[headEnterIndex].Token.Type == typeSpace) &&
		(events[tailExitIndex].Token.Type == TypeLineEnding || events[tailExitIndex].Token.Type == typeSpace) {
		index = headEnterIndex

		// And we have data.
		for index++; index < tailExitIndex; index++ {
			if events[index].Token.Type == TypeCodeTextData {
				// Then we have padding.
				events[headEnterIndex].Token.Type = TypeCodeTextPadding
				events[tailExitIndex].Token.Type = TypeCodeTextPadding
				headEnterIndex += 2
				tailExitIndex -= 2
				break
			}
		}
	}

	// Merge adjacent spaces and data.
	index = headEnterIndex - 1
	tailExitIndex++

	for index++; index <= tailExitIndex; index++ {
		if enter == -1 {
			if index != tailExitIndex && events[index].Token.Type != TypeLineEnding {
				enter = index
			}
		} else if index == tailExitIndex || events[index].Token.Type == TypeLineEnding {
			events[enter].Token.Type = TypeCodeTextData

			if index != enter+2 {
				// The enter and exit events of the merged token share one Token, so moving its end moves
				// both, as upstream's shared token object does.
				events[enter].Token.End = events[index-1].Token.End
				events = splice(events, enter+2, index-enter-2, nil)
				tailExitIndex -= index - enter - 2
				index = enter + 2
			}

			enter = -1
		}
	}

	return events
}

func previousCodeText(self *Self, code Code) bool {
	// If there is a previous code, there will always be a tail.
	return code != CodeGraveAccent || self.Events[len(self.Events)-1].Token.Type == TypeCharacterEscape
}

func tokenizeCodeText(self *Self, effects *Effects, ok State, nok State) State {
	sizeOpen := 0
	var size int
	var token *Token

	var start, sequenceOpen, between, data, sequenceClose State

	// Start of code (text).
	//
	//	> | `a`
	//	    ^
	//	> | \`a`
	//	     ^
	start = func(code Code) State {
		effects.Enter(TypeCodeText, nil)
		effects.Enter(TypeCodeTextSequence, nil)
		return sequenceOpen(code)
	}

	// In opening sequence.
	//
	//	> | `a`
	//	    ^
	sequenceOpen = func(code Code) State {
		if code == CodeGraveAccent {
			effects.Consume(code)
			sizeOpen++
			return sequenceOpen
		}

		effects.Exit(TypeCodeTextSequence)
		return between(code)
	}

	// Between something and something else.
	//
	//	> | `a`
	//	     ^^
	between = func(code Code) State {
		// EOF.
		if code == CodeEof {
			return nok(code)
		}

		// To do: next major: don’t do spaces in resolve, but when compiling,
		// like `markdown-rs`.
		// Tabs don’t work, and virtual spaces don’t make sense.
		if code == CodeSpace {
			effects.Enter(typeSpace, nil)
			effects.Consume(code)
			effects.Exit(typeSpace)
			return between
		}

		// Closing fence? Could also be data.
		if code == CodeGraveAccent {
			token = effects.Enter(TypeCodeTextSequence, nil)
			size = 0
			return sequenceClose(code)
		}

		if markdownLineEnding(code) {
			effects.Enter(TypeLineEnding, nil)
			effects.Consume(code)
			effects.Exit(TypeLineEnding)
			return between
		}

		// Data.
		effects.Enter(TypeCodeTextData, nil)
		return data(code)
	}

	// In data.
	//
	//	> | `a`
	//	     ^
	data = func(code Code) State {
		if code == CodeEof || code == CodeSpace || code == CodeGraveAccent || markdownLineEnding(code) {
			effects.Exit(TypeCodeTextData)
			return between(code)
		}

		effects.Consume(code)
		return data
	}

	// In closing sequence.
	//
	//	> | `a`
	//	      ^
	sequenceClose = func(code Code) State {
		// More.
		if code == CodeGraveAccent {
			effects.Consume(code)
			size++
			return sequenceClose
		}

		// Done!
		if size == sizeOpen {
			effects.Exit(TypeCodeTextSequence)
			effects.Exit(TypeCodeText)
			return ok(code)
		}

		// More or less accents: mark as data.
		token.Type = TypeCodeTextData
		return data(code)
	}

	return start
}
