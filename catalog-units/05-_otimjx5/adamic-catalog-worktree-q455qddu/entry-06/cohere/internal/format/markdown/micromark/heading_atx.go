package micromark

// headingAtx is micromark-core-commonmark/lib/heading-atx.js.
var headingAtx = &Construct{Name: "headingAtx", Resolve: &Resolver{Resolve: resolveHeadingAtx}, Tokenize: tokenizeHeadingAtx}

func resolveHeadingAtx(events []Event, context *TokenizeContext) []Event {
	contentEnd := len(events) - 2
	contentStart := 3

	// Prefix whitespace, part of the opening.
	if events[contentStart].Token.Type == TypeWhitespace {
		contentStart += 2
	}

	// Suffix whitespace, part of the closing.
	if contentEnd-2 > contentStart && events[contentEnd].Token.Type == TypeWhitespace {
		contentEnd -= 2
	}

	if events[contentEnd].Token.Type == TypeAtxHeadingSequence &&
		(contentStart == contentEnd-1 ||
			(contentEnd-4 > contentStart && events[contentEnd-2].Token.Type == TypeWhitespace)) {
		if contentStart+1 == contentEnd {
			contentEnd -= 2
		} else {
			contentEnd -= 4
		}
	}

	if contentEnd > contentStart {
		// Upstream shares the start and end point objects between these tokens and the events they span;
		// nothing mutates a point in place afterwards, so value copies are equivalent.
		content := &Token{
			Type:  TypeAtxHeadingText,
			Start: events[contentStart].Token.Start,
			End:   events[contentEnd].Token.End,
		}
		text := &Token{
			Type:        TypeChunkText,
			Start:       events[contentStart].Token.Start,
			End:         events[contentEnd].Token.End,
			ContentType: ContentTypeText,
		}

		events = splice(events, contentStart, contentEnd-contentStart+1, []Event{
			{Enter: true, Token: content, Context: context},
			{Enter: true, Token: text, Context: context},
			{Enter: false, Token: text, Context: context},
			{Enter: false, Token: content, Context: context},
		})
	}

	return events
}

func tokenizeHeadingAtx(self *Self, effects *Effects, ok State, nok State) State {
	size := 0

	var start, before, sequenceOpen, atBreak, sequenceFurther, data State

	start = func(code Code) State {
		// To do: parse indent like `markdown-rs`.
		effects.Enter(TypeAtxHeading, nil)
		return before(code)
	}

	before = func(code Code) State {
		effects.Enter(TypeAtxHeadingSequence, nil)
		return sequenceOpen(code)
	}

	sequenceOpen = func(code Code) State {
		if code == CodeNumberSign {
			// `size++ < atxHeadingOpeningFenceSizeMax`: the increment happens whether or not the
			// comparison passes.
			below := size < atxHeadingOpeningFenceSizeMax
			size++
			if below {
				effects.Consume(code)
				return sequenceOpen
			}
		}

		// Always at least one `#`.
		if code == CodeEof || markdownLineEndingOrSpace(code) {
			effects.Exit(TypeAtxHeadingSequence)
			return atBreak(code)
		}

		return nok(code)
	}

	atBreak = func(code Code) State {
		if code == CodeNumberSign {
			effects.Enter(TypeAtxHeadingSequence, nil)
			return sequenceFurther(code)
		}

		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeAtxHeading)
			// To do: interrupt like `markdown-rs`.
			// // Feel free to interrupt.
			// tokenizer.interrupt = false
			return ok(code)
		}

		if markdownSpace(code) {
			return factorySpace(effects, atBreak, TypeWhitespace, 0)(code)
		}

		// To do: generate `data` tokens, add the `text` token later.
		// Needs edit map, see: `markdown.rs`.
		effects.Enter(TypeAtxHeadingText, nil)
		return data(code)
	}

	// In further sequence (after whitespace): could be normal “visible” hashes in the heading or a final
	// sequence.
	sequenceFurther = func(code Code) State {
		if code == CodeNumberSign {
			effects.Consume(code)
			return sequenceFurther
		}

		effects.Exit(TypeAtxHeadingSequence)
		return atBreak(code)
	}

	data = func(code Code) State {
		if code == CodeEof || code == CodeNumberSign || markdownLineEndingOrSpace(code) {
			effects.Exit(TypeAtxHeadingText)
			return atBreak(code)
		}

		effects.Consume(code)
		return data
	}

	return start
}
