package micromark

// blockQuote is micromark-core-commonmark/lib/block-quote.js.
var blockQuote = &Construct{Name: "blockQuote", Tokenize: tokenizeBlockQuoteStart, Exit: exitBlockQuote}

// The continuation attempts blockQuote itself, so it is attached in init to break the initialization
// cycle Go would otherwise report.
func init() { blockQuote.Continuation = &Construct{Tokenize: tokenizeBlockQuoteContinuation} }

func tokenizeBlockQuoteStart(self *Self, effects *Effects, ok State, nok State) State {
	var start, after State

	// Start of block quote.
	//
	//	> | > a
	//	    ^
	start = func(code Code) State {
		if code == CodeGreaterThan {
			state := self.ContainerState

			if !state.open {
				effects.Enter(TypeBlockQuote, &Token{container: true})
				state.open = true
			}

			effects.Enter(TypeBlockQuotePrefix, nil)
			effects.Enter(TypeBlockQuoteMarker, nil)
			effects.Consume(code)
			effects.Exit(TypeBlockQuoteMarker)
			return after
		}

		return nok(code)
	}

	// After `>`, before optional whitespace.
	//
	//	> | > a
	//	     ^
	after = func(code Code) State {
		if markdownSpace(code) {
			effects.Enter(TypeBlockQuotePrefixWhitespace, nil)
			effects.Consume(code)
			effects.Exit(TypeBlockQuotePrefixWhitespace)
			effects.Exit(TypeBlockQuotePrefix)
			// Upstream returns `ok` itself, not `ok(code)`: the whitespace is consumed, so the next code goes
			// to `ok`.
			return ok
		}

		effects.Exit(TypeBlockQuotePrefix)
		return ok(code)
	}

	return start
}

// tokenizeBlockQuoteContinuation is the start of block quote continuation.
//
//	  | > a
//	> | > b
//	    ^
func tokenizeBlockQuoteContinuation(self *Self, effects *Effects, ok State, nok State) State {
	var contStart, contBefore State

	// Start of block quote continuation.
	//
	// Also used to parse the first block quote opening.
	//
	//	  | > a
	//	> | > b
	//	    ^
	contStart = func(code Code) State {
		if markdownSpace(code) {
			// Always populated by defaults.
			// Upstream's `undefined` max when codeIndented is disabled is factorySpace's 0 here.
			limit := tabSize
			if disabled(self, "codeIndented") {
				limit = 0
			}
			return factorySpace(effects, contBefore, TypeLinePrefix, limit)(code)
		}

		return contBefore(code)
	}

	// At `>`, after optional whitespace.
	//
	// Also used to parse the first block quote opening.
	//
	//	  | > a
	//	> | > b
	//	    ^
	contBefore = func(code Code) State {
		return effects.Attempt(blockQuote, ok, nok)(code)
	}

	return contStart
}

func exitBlockQuote(self *Self, effects *Effects) {
	effects.Exit(TypeBlockQuote)
}
