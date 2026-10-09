package micromark

// The micromark-factory-* packages: space, whitespace, label, title and destination.

// factorySpace is micromark-factory-space. max 0 is upstream's undefined: no limit.
//
// Its state is a spaceRun from the parse's Memory, whose states were bound when the slot was made, so
// a call allocates nothing (#vbjv3d6).
func factorySpace(effects *Effects, ok State, tokenType string, max int) State {
	run := effects.memory().spaceRun()
	run.effects, run.ok, run.tokenType = effects, ok, tokenType
	run.limit = -1 // Infinity.
	if max != 0 {
		run.limit = max - 1
	}
	return run.startState
}

// spaceRun is one factorySpace: upstream's closure state, with its states made once per slot.
type spaceRun struct {
	effects     *Effects
	ok          State
	tokenType   string
	limit, size int

	startState, prefixState State
}

func newSpaceRun() *spaceRun {
	run := &spaceRun{}
	run.startState, run.prefixState = run.start, run.prefix
	return run
}

// release zeroes the run for its next parse, keeping its bound states.
func (run *spaceRun) release() {
	*run = spaceRun{startState: run.startState, prefixState: run.prefixState}
}

// start is upstream's start. Upstream's size is per call of factorySpace, but a state may be entered
// again once the run it began is done (initializeFlow returns the same one for every line), and there
// size must start over, as it did when each entry made its own run.
func (run *spaceRun) start(code Code) State {
	if !markdownSpace(code) {
		return run.ok(code)
	}
	run.effects.Enter(run.tokenType, nil)
	run.size = 0
	return run.prefix(code)
}

func (run *spaceRun) prefix(code Code) State {
	if markdownSpace(code) {
		// `size++ < limit`: the increment happens whether or not the comparison passes.
		below := run.limit < 0 || run.size < run.limit
		run.size++
		if below {
			run.effects.Consume(code)
			return run.prefixState
		}
	}

	run.effects.Exit(run.tokenType)
	return run.ok(code)
}

// factoryWhitespace is micromark-factory-whitespace.
func factoryWhitespace(effects *Effects, ok State) State {
	seen := false

	var start State
	start = func(code Code) State {
		if markdownLineEnding(code) {
			effects.Enter(TypeLineEnding, nil)
			effects.Consume(code)
			effects.Exit(TypeLineEnding)
			seen = true
			return start
		}

		if markdownSpace(code) {
			tokenType := TypeLineSuffix
			if seen {
				tokenType = TypeLinePrefix
			}
			return factorySpace(effects, start, tokenType, 0)(code)
		}

		return ok(code)
	}

	return start
}

// factoryLabel is micromark-factory-label.
func factoryLabel(self *Self, effects *Effects, ok State, nok State, tokenType string, markerType string, stringType string) State {
	size := 0
	seen := false

	var start, atBreak, labelInside, labelEscape State

	start = func(code Code) State {
		effects.Enter(tokenType, nil)
		effects.Enter(markerType, nil)
		effects.Consume(code)
		effects.Exit(markerType)
		effects.Enter(stringType, nil)
		return atBreak
	}

	atBreak = func(code Code) State {
		if size > linkReferenceSizeMax ||
			code == CodeEof ||
			code == CodeLeftSquareBracket ||
			(code == CodeRightSquareBracket && !seen) ||
			// `_hiddenFootnoteSupport` is never in our constructs: micromark-extension-gfm-footnote
			// tokenizes footnote calls itself and does not set it.
			false {
			return nok(code)
		}

		if code == CodeRightSquareBracket {
			effects.Exit(stringType)
			effects.Enter(markerType, nil)
			effects.Consume(code)
			effects.Exit(markerType)
			effects.Exit(tokenType)
			return ok
		}

		if markdownLineEnding(code) {
			effects.Enter(TypeLineEnding, nil)
			effects.Consume(code)
			effects.Exit(TypeLineEnding)
			return atBreak
		}

		effects.Enter(TypeChunkString, &Token{ContentType: ContentTypeString})
		return labelInside(code)
	}

	labelInside = func(code Code) State {
		if code == CodeEof ||
			code == CodeLeftSquareBracket ||
			code == CodeRightSquareBracket ||
			markdownLineEnding(code) {
			effects.Exit(TypeChunkString)
			return atBreak(code)
		}
		// `size++ > linkReferenceSizeMax`, evaluated last, so it only counts when nothing before matched.
		over := size > linkReferenceSizeMax
		size++
		if over {
			effects.Exit(TypeChunkString)
			return atBreak(code)
		}

		effects.Consume(code)
		if !seen {
			seen = !markdownSpace(code)
		}
		if code == CodeBackslash {
			return labelEscape
		}
		return labelInside
	}

	labelEscape = func(code Code) State {
		if code == CodeLeftSquareBracket || code == CodeBackslash || code == CodeRightSquareBracket {
			effects.Consume(code)
			size++
			return labelInside
		}

		return labelInside(code)
	}

	return start
}

// factoryTitle is micromark-factory-title.
func factoryTitle(effects *Effects, ok State, nok State, tokenType string, markerType string, stringType string) State {
	var marker Code

	var start, begin, atBreak, inside, escape State

	start = func(code Code) State {
		if code == CodeQuotationMark || code == CodeApostrophe || code == CodeLeftParenthesis {
			effects.Enter(tokenType, nil)
			effects.Enter(markerType, nil)
			effects.Consume(code)
			effects.Exit(markerType)
			marker = code
			if code == CodeLeftParenthesis {
				marker = CodeRightParenthesis
			}
			return begin
		}

		return nok(code)
	}

	begin = func(code Code) State {
		if code == marker {
			effects.Enter(markerType, nil)
			effects.Consume(code)
			effects.Exit(markerType)
			effects.Exit(tokenType)
			return ok
		}

		effects.Enter(stringType, nil)
		return atBreak(code)
	}

	atBreak = func(code Code) State {
		if code == marker {
			effects.Exit(stringType)
			return begin(marker)
		}

		if code == CodeEof {
			return nok(code)
		}

		// Note: blank lines can’t exist in content.
		if markdownLineEnding(code) {
			effects.Enter(TypeLineEnding, nil)
			effects.Consume(code)
			effects.Exit(TypeLineEnding)
			return factorySpace(effects, atBreak, TypeLinePrefix, 0)
		}

		effects.Enter(TypeChunkString, &Token{ContentType: ContentTypeString})
		return inside(code)
	}

	inside = func(code Code) State {
		if code == marker || code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeChunkString)
			return atBreak(code)
		}

		effects.Consume(code)
		if code == CodeBackslash {
			return escape
		}
		return inside
	}

	escape = func(code Code) State {
		if code == marker || code == CodeBackslash {
			effects.Consume(code)
			return inside
		}

		return inside(code)
	}

	return start
}

// factoryDestination is micromark-factory-destination. max 0 is upstream's undefined: no limit.
func factoryDestination(effects *Effects, ok State, nok State, tokenType string, literalType string, literalMarkerType string, rawType string, stringType string, max int) State {
	limit := max
	unlimited := max == 0
	balance := 0

	var start, enclosedBefore, enclosed, enclosedEscape, raw, rawEscape State

	start = func(code Code) State {
		if code == CodeLessThan {
			effects.Enter(tokenType, nil)
			effects.Enter(literalType, nil)
			effects.Enter(literalMarkerType, nil)
			effects.Consume(code)
			effects.Exit(literalMarkerType)
			return enclosedBefore
		}

		// ASCII control, space, closing paren.
		if code == CodeEof || code == CodeSpace || code == CodeRightParenthesis || asciiControl(code) {
			return nok(code)
		}

		effects.Enter(tokenType, nil)
		effects.Enter(rawType, nil)
		effects.Enter(stringType, nil)
		effects.Enter(TypeChunkString, &Token{ContentType: ContentTypeString})
		return raw(code)
	}

	enclosedBefore = func(code Code) State {
		if code == CodeGreaterThan {
			effects.Enter(literalMarkerType, nil)
			effects.Consume(code)
			effects.Exit(literalMarkerType)
			effects.Exit(literalType)
			effects.Exit(tokenType)
			return ok
		}

		effects.Enter(stringType, nil)
		effects.Enter(TypeChunkString, &Token{ContentType: ContentTypeString})
		return enclosed(code)
	}

	enclosed = func(code Code) State {
		if code == CodeGreaterThan {
			effects.Exit(TypeChunkString)
			effects.Exit(stringType)
			return enclosedBefore(code)
		}

		if code == CodeEof || code == CodeLessThan || markdownLineEnding(code) {
			return nok(code)
		}

		effects.Consume(code)
		if code == CodeBackslash {
			return enclosedEscape
		}
		return enclosed
	}

	enclosedEscape = func(code Code) State {
		if code == CodeLessThan || code == CodeGreaterThan || code == CodeBackslash {
			effects.Consume(code)
			return enclosed
		}

		return enclosed(code)
	}

	raw = func(code Code) State {
		if balance == 0 && (code == CodeEof || code == CodeRightParenthesis || markdownLineEndingOrSpace(code)) {
			effects.Exit(TypeChunkString)
			effects.Exit(stringType)
			effects.Exit(rawType)
			effects.Exit(tokenType)
			return ok(code)
		}

		if (unlimited || balance < limit) && code == CodeLeftParenthesis {
			effects.Consume(code)
			balance++
			return raw
		}

		if code == CodeRightParenthesis {
			effects.Consume(code)
			balance--
			return raw
		}

		// ASCII control (but *not* `\0`) and space and `(`.
		if code == CodeEof || code == CodeSpace || code == CodeLeftParenthesis || asciiControl(code) {
			return nok(code)
		}

		effects.Consume(code)
		if code == CodeBackslash {
			return rawEscape
		}
		return raw
	}

	rawEscape = func(code Code) State {
		if code == CodeLeftParenthesis || code == CodeRightParenthesis || code == CodeBackslash {
			effects.Consume(code)
			return raw
		}

		return raw(code)
	}

	return start
}
