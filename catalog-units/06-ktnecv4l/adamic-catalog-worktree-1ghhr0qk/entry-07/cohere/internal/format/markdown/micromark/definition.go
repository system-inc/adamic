package micromark

// definition is micromark-core-commonmark/lib/definition.js.
var definition = &Construct{Name: "definition", Tokenize: tokenizeDefinition}

// definitionTitleBefore is upstream's file-local `titleBefore`.
var definitionTitleBefore = &Construct{Partial: true, Tokenize: tokenizeDefinitionTitleBefore}

func tokenizeDefinition(self *Self, effects *Effects, ok State, nok State) State {
	var identifier string

	var start, before, labelAfter, markerAfter, destinationBefore, destinationAfter, after, afterWhitespace State

	// At start of a definition.
	start = func(code Code) State {
		// Do not interrupt paragraphs (but do follow definitions).
		// To do: do `interrupt` the way `markdown-rs` does.
		// To do: parse whitespace the way `markdown-rs` does.
		effects.Enter(TypeDefinition, nil)
		return before(code)
	}

	// After optional whitespace, at `[`.
	before = func(code Code) State {
		// To do: parse whitespace the way `markdown-rs` does.

		return factoryLabel(self, effects, labelAfter,
			// Note: we don’t need to reset the way `markdown-rs` does.
			nok, TypeDefinitionLabel, TypeDefinitionLabelMarker, TypeDefinitionLabelString)(code)
	}

	// After label.
	labelAfter = func(code Code) State {
		// `.slice(1, -1)` drops the brackets, which are one code unit and one byte each, so byte slicing
		// of the serialized label is the same cut.
		label := self.SliceSerialize(self.Events[len(self.Events)-1].Token, false)
		identifier = NormalizeIdentifier(label[1 : len(label)-1])

		if code == CodeColon {
			effects.Enter(TypeDefinitionMarker, nil)
			effects.Consume(code)
			effects.Exit(TypeDefinitionMarker)
			return markerAfter
		}

		return nok(code)
	}

	// After marker.
	markerAfter = func(code Code) State {
		// Note: whitespace is optional.
		if markdownLineEndingOrSpace(code) {
			return factoryWhitespace(effects, destinationBefore)(code)
		}
		return destinationBefore(code)
	}

	// Before destination.
	destinationBefore = func(code Code) State {
		return factoryDestination(effects, destinationAfter,
			// Note: we don’t need to reset the way `markdown-rs` does.
			nok, TypeDefinitionDestination, TypeDefinitionDestinationLiteral, TypeDefinitionDestinationLiteralMarker, TypeDefinitionDestinationRaw, TypeDefinitionDestinationString,
			// Upstream passes no max: no limit.
			0)(code)
	}

	// After destination.
	destinationAfter = func(code Code) State {
		return effects.Attempt(definitionTitleBefore, after, after)(code)
	}

	// After definition.
	after = func(code Code) State {
		if markdownSpace(code) {
			return factorySpace(effects, afterWhitespace, TypeWhitespace, 0)(code)
		}
		return afterWhitespace(code)
	}

	// After definition, after optional whitespace.
	afterWhitespace = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeDefinition)

			// Note: we don’t care about uniqueness.
			// It’s likely that that doesn’t happen very frequently.
			// It is more likely that it wastes precious time.
			self.Parser.Defined = append(self.Parser.Defined, identifier)

			// To do: `markdown-rs` interrupt.
			// // You’d be interrupting.
			// tokenizer.interrupt = true
			return ok(code)
		}

		return nok(code)
	}

	return start
}

func tokenizeDefinitionTitleBefore(self *Self, effects *Effects, ok State, nok State) State {
	var titleBefore, beforeMarker, titleAfter, titleAfterOptionalWhitespace State

	// After destination, at whitespace.
	titleBefore = func(code Code) State {
		if markdownLineEndingOrSpace(code) {
			return factoryWhitespace(effects, beforeMarker)(code)
		}
		return nok(code)
	}

	// At title.
	beforeMarker = func(code Code) State {
		return factoryTitle(effects, titleAfter, nok, TypeDefinitionTitle, TypeDefinitionTitleMarker, TypeDefinitionTitleString)(code)
	}

	// After title.
	titleAfter = func(code Code) State {
		if markdownSpace(code) {
			return factorySpace(effects, titleAfterOptionalWhitespace, TypeWhitespace, 0)(code)
		}
		return titleAfterOptionalWhitespace(code)
	}

	// After title, after optional whitespace.
	titleAfterOptionalWhitespace = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			return ok(code)
		}
		return nok(code)
	}

	return titleBefore
}
