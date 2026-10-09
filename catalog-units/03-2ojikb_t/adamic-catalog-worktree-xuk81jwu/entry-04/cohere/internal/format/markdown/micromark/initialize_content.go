package micromark

// contentInitial is upstream's content, micromark/lib/initialize/content.js: definitions, then a
// paragraph of text chunks.
var contentInitial *InitialConstruct

func init() { contentInitial = &InitialConstruct{Tokenize: initializeContent} }

func initializeContent(self *Self, effects *Effects) State {
	var contentStart, afterContentStartConstruct, paragraphInitial, lineStart, data State
	var previous *Token

	afterContentStartConstruct = func(code Code) State {
		if code == CodeEof {
			effects.Consume(code)
			return nil
		}

		effects.Enter(TypeLineEnding, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEnding)
		return factorySpace(effects, contentStart, TypeLinePrefix, 0)
	}

	paragraphInitial = func(code Code) State {
		effects.Enter(TypeParagraph, nil)
		return lineStart(code)
	}

	lineStart = func(code Code) State {
		token := effects.Enter(TypeChunkText, &Token{ContentType: ContentTypeText, Previous: previous})

		if previous != nil {
			previous.Next = token
		}

		previous = token

		return data(code)
	}

	data = func(code Code) State {
		if code == CodeEof {
			effects.Exit(TypeChunkText)
			effects.Exit(TypeParagraph)
			effects.Consume(code)
			return nil
		}

		if markdownLineEnding(code) {
			effects.Consume(code)
			effects.Exit(TypeChunkText)
			return lineStart
		}

		// Data.
		effects.Consume(code)
		return data
	}

	contentStart = effects.Attempt(self.Parser.Constructs.ContentInitial, afterContentStartConstruct, paragraphInitial)

	return contentStart
}
