package micromark

import "slices"

// mathFlow is micromark-extension-math/lib/math-flow.js.
var mathFlow = &Construct{Name: "mathFlow", Concrete: true, Tokenize: tokenizeMathFenced}

// mathFlowNonLazyContinuation is upstream's file-local `nonLazyContinuation`. Unlike code-fenced's, it
// succeeds at the end of input.
var mathFlowNonLazyContinuation = &Construct{Partial: true, Tokenize: tokenizeMathFlowNonLazyContinuation}

// The token types math-flow.js spells as literals.
const (
	typeMathFlow              = "mathFlow"
	typeMathFlowFence         = "mathFlowFence"
	typeMathFlowFenceSequence = "mathFlowFenceSequence"
	typeMathFlowFenceMeta     = "mathFlowFenceMeta"
	typeMathFlowValue         = "mathFlowValue"
)

func tokenizeMathFenced(self *Self, effects *Effects, ok State, nok State) State {
	// Read when tokenize runs, before start, as upstream's `const initialSize` is.
	initialSize := 0
	if len(self.Events) > 0 {
		tail := self.Events[len(self.Events)-1]
		if tail.Token.Type == TypeLinePrefix {
			initialSize = utf16Length(tail.Context.SliceSerialize(tail.Token, true))
		}
	}
	sizeOpen := 0

	var start, sequenceOpen, metaBefore, meta, metaAfter, beforeNonLazyContinuation, contentStart,
		beforeContentChunk, contentChunk, after State

	// closingFence is the construct upstream builds inline, `{tokenize: tokenizeClosingFence, partial:
	// true}`, declared inside tokenizeMathFenced because its tokenizer closes over `sizeOpen` and reads
	// `self.parser` from the outer context.
	closingFence := &Construct{Partial: true, Tokenize: func(_ *Self, effects *Effects, ok State, nok State) State {
		size := 0

		var beforeSequenceClose, sequenceClose, afterSequenceClose State

		// In closing fence, after optional whitespace, at sequence.
		//
		//	  | $$
		//	  | \frac{1}{2}
		//	> | $$
		//	    ^
		beforeSequenceClose = func(code Code) State {
			effects.Enter(typeMathFlowFence, nil)
			effects.Enter(typeMathFlowFenceSequence, nil)
			return sequenceClose(code)
		}

		// In closing fence sequence.
		//
		//	  | $$
		//	  | \frac{1}{2}
		//	> | $$
		//	     ^
		sequenceClose = func(code Code) State {
			if code == CodeDollarSign {
				size++
				effects.Consume(code)
				return sequenceClose
			}

			if size < sizeOpen {
				return nok(code)
			}

			effects.Exit(typeMathFlowFenceSequence)
			return factorySpace(effects, afterSequenceClose, TypeWhitespace, 0)(code)
		}

		// After closing fence sequence, after optional whitespace.
		//
		//	  | $$
		//	  | \frac{1}{2}
		//	> | $$
		//	      ^
		afterSequenceClose = func(code Code) State {
			if code == CodeEof || markdownLineEnding(code) {
				effects.Exit(typeMathFlowFence)
				return ok(code)
			}

			return nok(code)
		}

		// Before closing fence, at optional whitespace.
		//
		//	  | $$
		//	  | \frac{1}{2}
		//	> | $$
		//	    ^
		//
		// Upstream passes `undefined` (no limit) when codeIndented is disabled, else 4.
		max := tabSize
		if slices.Contains(self.Parser.Constructs.Disable, "codeIndented") {
			max = 0
		}
		return factorySpace(effects, beforeSequenceClose, TypeLinePrefix, max)
	}}

	// Start of math.
	//
	//	> | $$
	//	    ^
	//	  | \frac{1}{2}
	//	  | $$
	start = func(code Code) State {
		effects.Enter(typeMathFlow, nil)
		effects.Enter(typeMathFlowFence, nil)
		effects.Enter(typeMathFlowFenceSequence, nil)
		return sequenceOpen(code)
	}

	// In opening fence sequence.
	//
	//	> | $$
	//	     ^
	//	  | \frac{1}{2}
	//	  | $$
	sequenceOpen = func(code Code) State {
		if code == CodeDollarSign {
			effects.Consume(code)
			sizeOpen++
			return sequenceOpen
		}

		if sizeOpen < 2 {
			return nok(code)
		}

		effects.Exit(typeMathFlowFenceSequence)
		return factorySpace(effects, metaBefore, TypeWhitespace, 0)(code)
	}

	// In opening fence, before meta.
	//
	//	> | $$asciimath
	//	      ^
	//	  | x < y
	//	  | $$
	metaBefore = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			return metaAfter(code)
		}

		effects.Enter(typeMathFlowFenceMeta, nil)
		effects.Enter(TypeChunkString, &Token{ContentType: ContentTypeString})
		return meta(code)
	}

	// In meta.
	//
	//	> | $$asciimath
	//	       ^
	//	  | x < y
	//	  | $$
	meta = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeChunkString)
			effects.Exit(typeMathFlowFenceMeta)
			return metaAfter(code)
		}

		if code == CodeDollarSign {
			return nok(code)
		}

		effects.Consume(code)
		return meta
	}

	// After meta.
	//
	//	> | $$
	//	      ^
	//	  | \frac{1}{2}
	//	  | $$
	metaAfter = func(code Code) State {
		// Guaranteed to be eol/eof.
		effects.Exit(typeMathFlowFence)

		if self.IsInterrupt() {
			return ok(code)
		}

		return effects.Attempt(mathFlowNonLazyContinuation, beforeNonLazyContinuation, after)(code)
	}

	// After eol/eof in math, at a non-lazy closing fence or content.
	//
	//	  | $$
	//	> | \frac{1}{2}
	//	    ^
	//	> | $$
	//	    ^
	beforeNonLazyContinuation = func(code Code) State {
		return effects.Attempt(closingFence, after, contentStart)(code)
	}

	// Before math content, definitely not before a closing fence.
	//
	//	  | $$
	//	> | \frac{1}{2}
	//	    ^
	//	  | $$
	contentStart = func(code Code) State {
		if initialSize != 0 {
			return factorySpace(effects, beforeContentChunk, TypeLinePrefix, initialSize+1)(code)
		}
		return beforeContentChunk(code)
	}

	// Before math content, after optional prefix.
	//
	//	  | $$
	//	> | \frac{1}{2}
	//	    ^
	//	  | $$
	beforeContentChunk = func(code Code) State {
		if code == CodeEof {
			return after(code)
		}

		if markdownLineEnding(code) {
			return effects.Attempt(mathFlowNonLazyContinuation, beforeNonLazyContinuation, after)(code)
		}

		effects.Enter(typeMathFlowValue, nil)
		return contentChunk(code)
	}

	// In math content.
	//
	//	  | $$
	//	> | \frac{1}{2}
	//	     ^
	//	  | $$
	contentChunk = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(typeMathFlowValue)
			return beforeContentChunk(code)
		}

		effects.Consume(code)
		return contentChunk
	}

	// After math (ha!).
	//
	//	  | $$
	//	  | \frac{1}{2}
	//	> | $$
	//	      ^
	after = func(code Code) State {
		effects.Exit(typeMathFlow)
		return ok(code)
	}

	return start
}

func tokenizeMathFlowNonLazyContinuation(self *Self, effects *Effects, ok State, nok State) State {
	var start, lineStart State

	start = func(code Code) State {
		if code == CodeEof {
			return ok(code)
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
