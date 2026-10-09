package micromark

// mathText is micromark-extension-math/lib/math-text.js, `mathText(options)`. single is the resolved
// `singleDollarTextMath`: upstream turns null and undefined into true.
//
// To do: next major: clean spaces in HTML compiler.
// This has to be coordinated together with `mdast-util-math`.
func mathText(single bool) *Construct {
	return &Construct{
		Name:     "mathText",
		Previous: previousMathText,
		Resolve:  mathTextResolver,
		Tokenize: func(self *Self, effects *Effects, ok State, nok State) State {
			return tokenizeMathText(single, effects, ok, nok)
		},
	}
}

// mathTextResolver is upstream's module-level `resolveMathText`, one function shared by every
// construct mathText returns.
var mathTextResolver = &Resolver{Resolve: resolveMathText}

// The token types math-text.js spells as literals. `space` is code_text.go's typeSpace.
const (
	typeMathText         = "mathText"
	typeMathTextSequence = "mathTextSequence"
	typeMathTextData     = "mathTextData"
	typeMathTextPadding  = "mathTextPadding"
)

func tokenizeMathText(single bool, effects *Effects, ok State, nok State) State {
	sizeOpen := 0
	var size int
	var token *Token

	var start, sequenceOpen, between, data, sequenceClose State

	// Start of math (text).
	//
	//	> | $a$
	//	    ^
	//	> | \$a$
	//	     ^
	start = func(code Code) State {
		effects.Enter(typeMathText, nil)
		effects.Enter(typeMathTextSequence, nil)
		return sequenceOpen(code)
	}

	// In opening sequence.
	//
	//	> | $a$
	//	    ^
	sequenceOpen = func(code Code) State {
		if code == CodeDollarSign {
			effects.Consume(code)
			sizeOpen++
			return sequenceOpen
		}

		// Not enough markers in the sequence.
		if sizeOpen < 2 && !single {
			return nok(code)
		}

		effects.Exit(typeMathTextSequence)
		return between(code)
	}

	// Between something and something else.
	//
	//	> | $a$
	//	     ^^
	between = func(code Code) State {
		if code == CodeEof {
			return nok(code)
		}

		if code == CodeDollarSign {
			token = effects.Enter(typeMathTextSequence, nil)
			size = 0
			return sequenceClose(code)
		}

		// Tabs don’t work, and virtual spaces don’t make sense.
		if code == CodeSpace {
			effects.Enter(typeSpace, nil)
			effects.Consume(code)
			effects.Exit(typeSpace)
			return between
		}

		if markdownLineEnding(code) {
			effects.Enter(TypeLineEnding, nil)
			effects.Consume(code)
			effects.Exit(TypeLineEnding)
			return between
		}

		// Data.
		effects.Enter(typeMathTextData, nil)
		return data(code)
	}

	// In data.
	//
	//	> | $a$
	//	     ^
	data = func(code Code) State {
		if code == CodeEof || code == CodeSpace || code == CodeDollarSign || markdownLineEnding(code) {
			effects.Exit(typeMathTextData)
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
		if code == CodeDollarSign {
			effects.Consume(code)
			size++
			return sequenceClose
		}

		// Done!
		if size == sizeOpen {
			effects.Exit(typeMathTextSequence)
			effects.Exit(typeMathText)
			return ok(code)
		}

		// More or less accents: mark as data.
		token.Type = typeMathTextData
		return data(code)
	}

	return start
}

func resolveMathText(events []Event, context *TokenizeContext) []Event {
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
			if events[index].Token.Type == typeMathTextData {
				// Then we have padding.
				events[tailExitIndex].Token.Type = typeMathTextPadding
				events[headEnterIndex].Token.Type = typeMathTextPadding
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
			events[enter].Token.Type = typeMathTextData

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

func previousMathText(self *Self, code Code) bool {
	// If there is a previous code, there will always be a tail.
	return code != CodeDollarSign || self.Events[len(self.Events)-1].Token.Type == TypeCharacterEscape
}
