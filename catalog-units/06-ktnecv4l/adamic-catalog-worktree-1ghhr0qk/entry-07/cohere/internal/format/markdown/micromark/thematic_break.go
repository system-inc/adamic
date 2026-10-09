package micromark

// thematicBreak is micromark-core-commonmark/lib/thematic-break.js.
var thematicBreak = &Construct{Name: "thematicBreak", Tokenize: tokenizeThematicBreak}

func tokenizeThematicBreak(self *Self, effects *Effects, ok State, nok State) State {
	size := 0
	var marker Code

	var start, before, atBreak, sequence State

	start = func(code Code) State {
		effects.Enter(TypeThematicBreak, nil)
		// To do: parse indent like `markdown-rs`.
		return before(code)
	}

	before = func(code Code) State {
		marker = code
		return atBreak(code)
	}

	atBreak = func(code Code) State {
		if code == marker {
			effects.Enter(TypeThematicBreakSequence, nil)
			return sequence(code)
		}

		if size >= thematicBreakMarkerCountMin && (code == CodeEof || markdownLineEnding(code)) {
			effects.Exit(TypeThematicBreak)
			return ok(code)
		}

		return nok(code)
	}

	sequence = func(code Code) State {
		if code == marker {
			effects.Consume(code)
			size++
			return sequence
		}

		effects.Exit(TypeThematicBreakSequence)
		if markdownSpace(code) {
			return factorySpace(effects, atBreak, TypeWhitespace, 0)(code)
		}
		return atBreak(code)
	}

	return start
}
