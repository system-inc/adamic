package micromark

// characterEscape is micromark-core-commonmark/lib/character-escape.js.
var characterEscape = &Construct{Name: "characterEscape", Tokenize: tokenizeCharacterEscape}

func tokenizeCharacterEscape(self *Self, effects *Effects, ok State, nok State) State {
	var start, inside State

	start = func(code Code) State {
		effects.Enter(TypeCharacterEscape, nil)
		effects.Enter(TypeEscapeMarker, nil)
		effects.Consume(code)
		effects.Exit(TypeEscapeMarker)
		return inside
	}

	inside = func(code Code) State {
		// ASCII punctuation.
		if asciiPunctuation(code) {
			effects.Enter(TypeCharacterEscapeValue, nil)
			effects.Consume(code)
			effects.Exit(TypeCharacterEscapeValue)
			effects.Exit(TypeCharacterEscape)
			return ok
		}

		return nok(code)
	}

	return start
}
