package micromark

// hardBreakEscape is micromark-core-commonmark/lib/hard-break-escape.js.
var hardBreakEscape = &Construct{Name: "hardBreakEscape", Tokenize: tokenizeHardBreakEscape}

func tokenizeHardBreakEscape(self *Self, effects *Effects, ok State, nok State) State {
	var start, after State

	start = func(code Code) State {
		effects.Enter(TypeHardBreakEscape, nil)
		effects.Consume(code)
		return after
	}

	after = func(code Code) State {
		if markdownLineEnding(code) {
			effects.Exit(TypeHardBreakEscape)
			return ok(code)
		}

		return nok(code)
	}

	return start
}
