package micromark

import "slices"

// codeFenced is micromark-core-commonmark/lib/code-fenced.js.
var codeFenced = &Construct{Name: "codeFenced", Concrete: true, Tokenize: tokenizeCodeFenced}

// codeFencedNonLazyContinuation is upstream's file-local `nonLazyContinuation`.
var codeFencedNonLazyContinuation = &Construct{Partial: true, Tokenize: tokenizeCodeFencedNonLazyContinuation}

func tokenizeCodeFenced(self *Self, effects *Effects, ok State, nok State) State {
	initialPrefix := 0
	sizeOpen := 0
	var marker Code

	var start, beforeSequenceOpen, sequenceOpen, infoBefore, info, metaBefore, meta, atNonLazyBreak,
		contentBefore, contentStart, beforeContentChunk, contentChunk, after State

	// closeStart is upstream's `closeStart`, declared inside tokenizeCodeFenced because its tokenizer
	// closes over `marker` and `sizeOpen`, and reads `self.parser` from the outer context.
	closeStart := &Construct{Partial: true, Tokenize: func(_ *Self, effects *Effects, ok State, nok State) State {
		size := 0

		var startBefore, start, beforeSequenceClose, sequenceClose, sequenceCloseAfter State

		startBefore = func(code Code) State {
			effects.Enter(TypeLineEnding, nil)
			effects.Consume(code)
			effects.Exit(TypeLineEnding)
			return start
		}

		// Before closing fence, at optional whitespace.
		start = func(code Code) State {
			// Always populated by defaults.

			// To do: `enter` here or in next state?
			effects.Enter(TypeCodeFencedFence, nil)
			if markdownSpace(code) {
				// Upstream passes `undefined` (no limit) when codeIndented is disabled, else 4.
				max := tabSize
				if slices.Contains(self.Parser.Constructs.Disable, "codeIndented") {
					max = 0
				}
				return factorySpace(effects, beforeSequenceClose, TypeLinePrefix, max)(code)
			}
			return beforeSequenceClose(code)
		}

		// In closing fence, after optional whitespace, at sequence.
		beforeSequenceClose = func(code Code) State {
			if code == marker {
				effects.Enter(TypeCodeFencedFenceSequence, nil)
				return sequenceClose(code)
			}

			return nok(code)
		}

		// In closing fence sequence.
		sequenceClose = func(code Code) State {
			if code == marker {
				size++
				effects.Consume(code)
				return sequenceClose
			}

			if size >= sizeOpen {
				effects.Exit(TypeCodeFencedFenceSequence)
				if markdownSpace(code) {
					return factorySpace(effects, sequenceCloseAfter, TypeWhitespace, 0)(code)
				}
				return sequenceCloseAfter(code)
			}

			return nok(code)
		}

		// After closing fence sequence, after optional whitespace.
		sequenceCloseAfter = func(code Code) State {
			if code == CodeEof || markdownLineEnding(code) {
				effects.Exit(TypeCodeFencedFence)
				return ok(code)
			}

			return nok(code)
		}

		return startBefore
	}}

	// Start of code.
	start = func(code Code) State {
		// To do: parse whitespace like `markdown-rs`.
		return beforeSequenceOpen(code)
	}

	// In opening fence, after prefix, at sequence.
	beforeSequenceOpen = func(code Code) State {
		initialPrefix = 0
		if len(self.Events) > 0 {
			tail := self.Events[len(self.Events)-1]
			if tail.Token.Type == TypeLinePrefix {
				initialPrefix = utf16Length(tail.Context.SliceSerialize(tail.Token, true))
			}
		}

		marker = code
		effects.Enter(TypeCodeFenced, nil)
		effects.Enter(TypeCodeFencedFence, nil)
		effects.Enter(TypeCodeFencedFenceSequence, nil)
		return sequenceOpen(code)
	}

	// In opening fence sequence.
	sequenceOpen = func(code Code) State {
		if code == marker {
			sizeOpen++
			effects.Consume(code)
			return sequenceOpen
		}

		if sizeOpen < codeFencedSequenceSizeMin {
			return nok(code)
		}

		effects.Exit(TypeCodeFencedFenceSequence)
		if markdownSpace(code) {
			return factorySpace(effects, infoBefore, TypeWhitespace, 0)(code)
		}
		return infoBefore(code)
	}

	// In opening fence, after the sequence (and optional whitespace), before info.
	infoBefore = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeCodeFencedFence)
			if self.IsInterrupt() {
				return ok(code)
			}
			return effects.Check(codeFencedNonLazyContinuation, atNonLazyBreak, after)(code)
		}

		effects.Enter(TypeCodeFencedFenceInfo, nil)
		effects.Enter(TypeChunkString, &Token{ContentType: ContentTypeString})
		return info(code)
	}

	// In info.
	info = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeChunkString)
			effects.Exit(TypeCodeFencedFenceInfo)
			return infoBefore(code)
		}

		if markdownSpace(code) {
			effects.Exit(TypeChunkString)
			effects.Exit(TypeCodeFencedFenceInfo)
			return factorySpace(effects, metaBefore, TypeWhitespace, 0)(code)
		}

		if code == CodeGraveAccent && code == marker {
			return nok(code)
		}

		effects.Consume(code)
		return info
	}

	// In opening fence, after info and whitespace, before meta.
	metaBefore = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			return infoBefore(code)
		}

		effects.Enter(TypeCodeFencedFenceMeta, nil)
		effects.Enter(TypeChunkString, &Token{ContentType: ContentTypeString})
		return meta(code)
	}

	// In meta.
	meta = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeChunkString)
			effects.Exit(TypeCodeFencedFenceMeta)
			return infoBefore(code)
		}

		if code == CodeGraveAccent && code == marker {
			return nok(code)
		}

		effects.Consume(code)
		return meta
	}

	// At eol/eof in code, before a non-lazy closing fence or content.
	atNonLazyBreak = func(code Code) State {
		return effects.Attempt(closeStart, after, contentBefore)(code)
	}

	// Before code content, not a closing fence, at eol.
	contentBefore = func(code Code) State {
		effects.Enter(TypeLineEnding, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEnding)
		return contentStart
	}

	// Before code content, not a closing fence.
	contentStart = func(code Code) State {
		if initialPrefix > 0 && markdownSpace(code) {
			return factorySpace(effects, beforeContentChunk, TypeLinePrefix, initialPrefix+1)(code)
		}
		return beforeContentChunk(code)
	}

	// Before code content, after optional prefix.
	beforeContentChunk = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			return effects.Check(codeFencedNonLazyContinuation, atNonLazyBreak, after)(code)
		}

		effects.Enter(TypeCodeFlowValue, nil)
		return contentChunk(code)
	}

	// In code content.
	contentChunk = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeCodeFlowValue)
			return beforeContentChunk(code)
		}

		effects.Consume(code)
		return contentChunk
	}

	// After code.
	after = func(code Code) State {
		effects.Exit(TypeCodeFenced)
		return ok(code)
	}

	return start
}

func tokenizeCodeFencedNonLazyContinuation(self *Self, effects *Effects, ok State, nok State) State {
	var start, lineStart State

	start = func(code Code) State {
		if code == CodeEof {
			return nok(code)
		}

		effects.Enter(TypeLineEnding, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEnding)
		return lineStart
	}

	lineStart = func(code Code) State {
		if self.Parser.Lazy[self.Now().Line] {
			return nok(code)
		}
		return ok(code)
	}

	return start
}
