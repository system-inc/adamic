package micromark

// content is micromark-core-commonmark/lib/content.js: a run of lines that becomes definitions and a
// paragraph once its content is tokenized.
var content = &Construct{Resolve: &Resolver{Resolve: resolveContent}, Tokenize: tokenizeContent}

var continuationConstruct = &Construct{Partial: true, Tokenize: tokenizeContinuation}

// resolveContent tokenizes the content chunks.
func resolveContent(events []Event, _ *TokenizeContext) []Event {
	events, _ = subtokenize(events)
	return events
}

func tokenizeContent(self *Self, effects *Effects, ok State, nok State) State {
	var previous *Token

	var chunkStart, chunkInside, contentEnd, contentContinue State

	chunkStart = func(code Code) State {
		effects.Enter(TypeContent, nil)
		previous = effects.Enter(TypeChunkContent, &Token{ContentType: ContentTypeContent})
		return chunkInside(code)
	}

	chunkInside = func(code Code) State {
		if code == CodeEof {
			return contentEnd(code)
		}

		// To do: in `markdown-rs`, each line is parsed on its own, and everything is stitched together
		// resolving.
		if markdownLineEnding(code) {
			return effects.Check(continuationConstruct, contentContinue, contentEnd)(code)
		}

		// Data.
		effects.Consume(code)
		return chunkInside
	}

	contentEnd = func(code Code) State {
		effects.Exit(TypeChunkContent)
		effects.Exit(TypeContent)
		return ok(code)
	}

	contentContinue = func(code Code) State {
		effects.Consume(code)
		effects.Exit(TypeChunkContent)
		previous.Next = effects.Enter(TypeChunkContent, &Token{ContentType: ContentTypeContent, Previous: previous})
		previous = previous.Next
		return chunkInside
	}

	return chunkStart
}

func tokenizeContinuation(self *Self, effects *Effects, ok State, nok State) State {
	var startLookahead, prefixed State

	startLookahead = func(code Code) State {
		effects.Exit(TypeChunkContent)
		effects.Enter(TypeLineEnding, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEnding)
		return factorySpace(effects, prefixed, TypeLinePrefix, 0)
	}

	prefixed = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			return nok(code)
		}

		if len(self.Events) > 0 {
			tail := self.Events[len(self.Events)-1]
			if !disabled(self, "codeIndented") &&
				tail.Token.Type == TypeLinePrefix &&
				utf16Length(tail.Context.SliceSerialize(tail.Token, true)) >= tabSize {
				return ok(code)
			}
		}

		return effects.Interrupt(self.Parser.Constructs.Flow, nok, ok)(code)
	}

	return startLookahead
}
