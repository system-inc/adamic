package micromark

// codeIndented is micromark-core-commonmark/lib/code-indented.js.
var codeIndented = &Construct{Name: "codeIndented", Tokenize: tokenizeCodeIndented}

// codeIndentedFurtherStart is upstream's file-local `furtherStart`.
var codeIndentedFurtherStart = &Construct{Partial: true, Tokenize: tokenizeCodeIndentedFurtherStart}

func tokenizeCodeIndented(self *Self, effects *Effects, ok State, nok State) State {
	var start, afterPrefix, atBreak, inside, after State

	// Parsing note: it is not needed to check if this first line is a filled line (that it has a
	// non-whitespace character), because blank lines are parsed already, so we never run into that.
	start = func(code Code) State {
		// To do: manually check if interrupting like `markdown-rs`.
		effects.Enter(TypeCodeIndented, nil)
		// To do: use an improved `space_or_tab` function like `markdown-rs`, so that we can drop the next
		// state.
		return factorySpace(effects, afterPrefix, TypeLinePrefix, tabSize+1)(code)
	}

	// At start, after 1 or 4 spaces.
	afterPrefix = func(code Code) State {
		if len(self.Events) > 0 {
			tail := self.Events[len(self.Events)-1]
			if tail.Token.Type == TypeLinePrefix &&
				utf16Length(tail.Context.SliceSerialize(tail.Token, true)) >= tabSize {
				return atBreak(code)
			}
		}
		return nok(code)
	}

	atBreak = func(code Code) State {
		if code == CodeEof {
			return after(code)
		}

		if markdownLineEnding(code) {
			return effects.Attempt(codeIndentedFurtherStart, atBreak, after)(code)
		}

		effects.Enter(TypeCodeFlowValue, nil)
		return inside(code)
	}

	inside = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeCodeFlowValue)
			return atBreak(code)
		}

		effects.Consume(code)
		return inside
	}

	after = func(code Code) State {
		effects.Exit(TypeCodeIndented)
		// To do: allow interrupting like `markdown-rs`.
		// Feel free to interrupt.
		// tokenizer.interrupt = false
		return ok(code)
	}

	return start
}

func tokenizeCodeIndentedFurtherStart(self *Self, effects *Effects, ok State, nok State) State {
	var furtherStart, afterPrefix State

	// At eol, trying to parse another indent.
	furtherStart = func(code Code) State {
		// To do: improve `lazy` / `pierce` handling.
		// If this is a lazy line, it can’t be code.
		if self.Parser.Lazy[self.Now().Line] {
			return nok(code)
		}

		if markdownLineEnding(code) {
			effects.Enter(TypeLineEnding, nil)
			effects.Consume(code)
			effects.Exit(TypeLineEnding)
			return furtherStart
		}

		// To do: the code here in `micromark-js` is a bit different from `markdown-rs` because there it
		// can attempt spaces. We can’t yet.
		//
		// To do: use an improved `space_or_tab` function like `markdown-rs`, so that we can drop the next
		// state.
		return factorySpace(effects, afterPrefix, TypeLinePrefix, tabSize+1)(code)
	}

	// At start, after 1 or 4 spaces.
	afterPrefix = func(code Code) State {
		if len(self.Events) > 0 {
			tail := self.Events[len(self.Events)-1]
			if tail.Token.Type == TypeLinePrefix &&
				utf16Length(tail.Context.SliceSerialize(tail.Token, true)) >= tabSize {
				return ok(code)
			}
		}
		if markdownLineEnding(code) {
			return furtherStart(code)
		}
		return nok(code)
	}

	return furtherStart
}
