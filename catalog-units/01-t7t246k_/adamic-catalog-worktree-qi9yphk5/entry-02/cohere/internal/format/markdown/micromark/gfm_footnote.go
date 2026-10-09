package micromark

import "slices"

// gfmFootnote is micromark-extension-gfm-footnote/lib/syntax.js: footnote definitions in the document,
// footnote calls in text, and the potential call that turns a `![^a]` image start into a `!` and a call.

// The extension's own token types, as upstream spells them.
const (
	typeGfmFootnoteCall                  = "gfmFootnoteCall"
	typeGfmFootnoteCallLabelMarker       = "gfmFootnoteCallLabelMarker"
	typeGfmFootnoteCallMarker            = "gfmFootnoteCallMarker"
	typeGfmFootnoteCallString            = "gfmFootnoteCallString"
	typeGfmFootnoteDefinition            = "gfmFootnoteDefinition"
	typeGfmFootnoteDefinitionLabel       = "gfmFootnoteDefinitionLabel"
	typeGfmFootnoteDefinitionLabelMarker = "gfmFootnoteDefinitionLabelMarker"
	typeGfmFootnoteDefinitionMarker      = "gfmFootnoteDefinitionMarker"
	typeGfmFootnoteDefinitionLabelString = "gfmFootnoteDefinitionLabelString"
	typeGfmFootnoteDefinitionWhitespace  = "gfmFootnoteDefinitionWhitespace"
	typeGfmFootnoteDefinitionIndent      = "gfmFootnoteDefinitionIndent"
	// gfmFootnoteLabelSizeMax is upstream's inline 999, the same cap as linkReferenceSizeMax.
	gfmFootnoteLabelSizeMax = 999
)

var gfmFootnoteIndent = &Construct{Tokenize: tokenizeGfmFootnoteIndent, Partial: true}

// To do: micromark should support a `_hiddenGfmFootnoteSupport`, which only
// affects label start (image).
// That will let us drop `tokenizePotentialGfmFootnote*`.
// It currently has a `_hiddenFootnoteSupport`, which affects that and more.
// That can be removed when `micromark-extension-footnote` is archived.

var gfmFootnoteDefinition = &Construct{
	Name:         "gfmFootnoteDefinition",
	Tokenize:     tokenizeGfmFootnoteDefinitionStart,
	Continuation: &Construct{Tokenize: tokenizeGfmFootnoteDefinitionContinuation},
	Exit:         gfmFootnoteDefinitionEnd,
}

var gfmFootnoteCall = &Construct{Name: "gfmFootnoteCall", Tokenize: tokenizeGfmFootnoteCall}

var gfmPotentialFootnoteCall = &Construct{
	Name:      "gfmPotentialFootnoteCall",
	Add:       "after",
	Tokenize:  tokenizePotentialGfmFootnoteCall,
	ResolveTo: &Resolver{Resolve: resolveToPotentialGfmFootnoteCall},
}

// gfmFootnoteExtension is upstream's gfmFootnote(): the extension enabling GFM footnote syntax.
func gfmFootnoteExtension() *Extension {
	return &Extension{
		Document: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeLeftSquareBracket: {gfmFootnoteDefinition},
		}},
		Text: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeLeftSquareBracket:  {gfmFootnoteCall},
			CodeRightSquareBracket: {gfmPotentialFootnoteCall},
		}},
	}
}

// gfmFootnoteDefined is `self.parser.gfmFootnotes || (self.parser.gfmFootnotes = [])`: the normalized
// identifiers of the footnote definitions seen so far, one list per parser, shared by every tokenizer of
// that parse.
func gfmFootnoteDefined(parser *ParseContext) *[]string {
	return &parser.GfmFootnotes
}

// To do: remove after micromark update.
func tokenizePotentialGfmFootnoteCall(self *Self, effects *Effects, ok State, nok State) State {
	index := len(self.Events)
	defined := gfmFootnoteDefined(self.Parser)
	var labelStart *Token

	// Find an opening.
	for {
		// `while (index--)`.
		if index == 0 {
			break
		}
		index--

		token := self.Events[index].Token
		if token.Type == TypeLabelImage {
			labelStart = token
			break
		}

		// Exit if we’ve walked far enough.
		if token.Type == typeGfmFootnoteCall || token.Type == TypeLabelLink || token.Type == TypeLabel ||
			token.Type == TypeImage || token.Type == TypeLink {
			break
		}
	}

	return func(code Code) State {
		if labelStart == nil || !labelStart.balanced {
			return nok(code)
		}

		id := NormalizeIdentifier(self.SliceSerialize(&Token{
			Start: labelStart.End,
			End:   self.Now(),
		}, false))

		// `id.codePointAt(0) !== 94`: an empty id reads undefined, which is not 94. `^` is one byte, so
		// `id.slice(1)` is id[1:].
		if len(id) == 0 || id[0] != '^' || !slices.Contains(*defined, id[1:]) {
			return nok(code)
		}

		effects.Enter(typeGfmFootnoteCallLabelMarker, nil)
		effects.Consume(code)
		effects.Exit(typeGfmFootnoteCallLabelMarker)
		return ok(code)
	}
}

// To do: remove after micromark update.
func resolveToPotentialGfmFootnoteCall(events []Event, context *TokenizeContext) []Event {
	index := len(events)

	// Find an opening.
	for {
		// `while (index--)`.
		if index == 0 {
			break
		}
		index--

		if events[index].Token.Type == TypeLabelImage && events[index].Enter {
			break
		}
	}

	// Change the `labelImageMarker` to a `data`.
	events[index+1].Token.Type = TypeData
	events[index+3].Token.Type = typeGfmFootnoteCallLabelMarker

	// The whole (without `!`):
	// Upstream copies each point with Object.assign; points are values here.
	call := &Token{
		Type:  typeGfmFootnoteCall,
		Start: events[index+3].Token.Start,
		End:   events[len(events)-1].Token.End,
	}
	// The `^` marker
	marker := &Token{
		Type:  typeGfmFootnoteCallMarker,
		Start: events[index+3].Token.End,
		End:   events[index+3].Token.End,
	}
	// Increment the end 1 character.
	marker.End.Column++
	marker.End.Offset++
	marker.End.bufferIndex++
	stringToken := &Token{
		Type:  typeGfmFootnoteCallString,
		Start: marker.End,
		End:   events[len(events)-1].Token.Start,
	}
	chunk := &Token{
		Type:        TypeChunkString,
		ContentType: ContentTypeString,
		Start:       stringToken.Start,
		End:         stringToken.End,
	}

	replacement := []Event{
		// Take the `labelImageMarker` (now `data`, the `!`)
		events[index+1],
		events[index+2],
		{Enter: true, Token: call, Context: context},
		// The `[`
		events[index+3],
		events[index+4],
		// The `^`.
		{Enter: true, Token: marker, Context: context},
		{Enter: false, Token: marker, Context: context},
		// Everything in between.
		{Enter: true, Token: stringToken, Context: context},
		{Enter: true, Token: chunk, Context: context},
		{Enter: false, Token: chunk, Context: context},
		{Enter: false, Token: stringToken, Context: context},
		// The ending (`]`, properly parsed and labelled).
		events[len(events)-2],
		events[len(events)-1],
		{Enter: false, Token: call, Context: context},
	}

	// `events.splice(index, events.length - index + 1, ...replacement)`: the count runs one past the end,
	// which splice clamps, so it removes everything from index on.
	return splice(events, index, len(events)-index, replacement)
}

func tokenizeGfmFootnoteCall(self *Self, effects *Effects, ok State, nok State) State {
	defined := gfmFootnoteDefined(self.Parser)
	size := 0
	data := false

	var start, callStart, callData, callEscape State

	// Note: the implementation of `markdown-rs` is different, because it houses
	// core *and* extensions in one project.
	// Therefore, it can include footnote logic inside `label-end`.
	// We can’t do that, but luckily, we can parse footnotes in a simpler way than
	// needed for labels.

	// Start of footnote label.
	//
	// ```markdown
	// > | a [^b] c
	//       ^
	// ```
	start = func(code Code) State {
		effects.Enter(typeGfmFootnoteCall, nil)
		effects.Enter(typeGfmFootnoteCallLabelMarker, nil)
		effects.Consume(code)
		effects.Exit(typeGfmFootnoteCallLabelMarker)
		return callStart
	}

	// After `[`, at `^`.
	//
	// ```markdown
	// > | a [^b] c
	//        ^
	// ```
	callStart = func(code Code) State {
		if code != CodeCaret {
			return nok(code)
		}
		effects.Enter(typeGfmFootnoteCallMarker, nil)
		effects.Consume(code)
		effects.Exit(typeGfmFootnoteCallMarker)
		effects.Enter(typeGfmFootnoteCallString, nil)
		effects.Enter(TypeChunkString, &Token{ContentType: ContentTypeString})
		return callData
	}

	// In label.
	//
	// ```markdown
	// > | a [^b] c
	//         ^
	// ```
	callData = func(code Code) State {
		if
		// Too long.
		size > gfmFootnoteLabelSizeMax ||
			// Closing brace with nothing.
			(code == CodeRightSquareBracket && !data) ||
			// Space or tab is not supported by GFM for some reason.
			// `\n` and `[` not being supported makes sense.
			code == CodeEof || code == CodeLeftSquareBracket || markdownLineEndingOrSpace(code) {
			return nok(code)
		}

		if code == CodeRightSquareBracket {
			effects.Exit(TypeChunkString)
			token := effects.Exit(typeGfmFootnoteCallString)
			if !slices.Contains(*defined, NormalizeIdentifier(self.SliceSerialize(token, false))) {
				return nok(code)
			}
			effects.Enter(typeGfmFootnoteCallLabelMarker, nil)
			effects.Consume(code)
			effects.Exit(typeGfmFootnoteCallLabelMarker)
			effects.Exit(typeGfmFootnoteCall)
			return ok
		}

		if !markdownLineEndingOrSpace(code) {
			data = true
		}
		size++
		effects.Consume(code)
		if code == CodeBackslash {
			return callEscape
		}
		return callData
	}

	// On character after escape.
	//
	// ```markdown
	// > | a [^b\c] d
	//           ^
	// ```
	callEscape = func(code Code) State {
		if code == CodeLeftSquareBracket || code == CodeBackslash || code == CodeRightSquareBracket {
			effects.Consume(code)
			size++
			return callData
		}
		return callData(code)
	}

	return start
}

func tokenizeGfmFootnoteDefinitionStart(self *Self, effects *Effects, ok State, nok State) State {
	defined := gfmFootnoteDefined(self.Parser)
	var identifier string
	size := 0
	data := false

	var start, labelAtMarker, labelInside, labelEscape, labelAfter, whitespaceAfter State

	// Start of GFM footnote definition.
	//
	// ```markdown
	// > | [^a]: b
	//     ^
	// ```
	start = func(code Code) State {
		// Upstream sets `_container = true` on the entered token; passing it as the fields is the same.
		effects.Enter(typeGfmFootnoteDefinition, &Token{container: true})
		effects.Enter(typeGfmFootnoteDefinitionLabel, nil)
		effects.Enter(typeGfmFootnoteDefinitionLabelMarker, nil)
		effects.Consume(code)
		effects.Exit(typeGfmFootnoteDefinitionLabelMarker)
		return labelAtMarker
	}

	// In label, at caret.
	//
	// ```markdown
	// > | [^a]: b
	//      ^
	// ```
	labelAtMarker = func(code Code) State {
		if code == CodeCaret {
			effects.Enter(typeGfmFootnoteDefinitionMarker, nil)
			effects.Consume(code)
			effects.Exit(typeGfmFootnoteDefinitionMarker)
			effects.Enter(typeGfmFootnoteDefinitionLabelString, nil)
			effects.Enter(TypeChunkString, &Token{ContentType: ContentTypeString})
			return labelInside
		}

		return nok(code)
	}

	// In label.
	//
	// > 👉 **Note**: `cmark-gfm` prevents whitespace from occurring in footnote
	// > definition labels.
	//
	// ```markdown
	// > | [^a]: b
	//       ^
	// ```
	labelInside = func(code Code) State {
		if
		// Too long.
		size > gfmFootnoteLabelSizeMax ||
			// Closing brace with nothing.
			(code == CodeRightSquareBracket && !data) ||
			// Space or tab is not supported by GFM for some reason.
			// `\n` and `[` not being supported makes sense.
			code == CodeEof || code == CodeLeftSquareBracket || markdownLineEndingOrSpace(code) {
			return nok(code)
		}

		if code == CodeRightSquareBracket {
			effects.Exit(TypeChunkString)
			token := effects.Exit(typeGfmFootnoteDefinitionLabelString)
			identifier = NormalizeIdentifier(self.SliceSerialize(token, false))
			effects.Enter(typeGfmFootnoteDefinitionLabelMarker, nil)
			effects.Consume(code)
			effects.Exit(typeGfmFootnoteDefinitionLabelMarker)
			effects.Exit(typeGfmFootnoteDefinitionLabel)
			return labelAfter
		}

		if !markdownLineEndingOrSpace(code) {
			data = true
		}
		size++
		effects.Consume(code)
		if code == CodeBackslash {
			return labelEscape
		}
		return labelInside
	}

	// After `\`, at a special character.
	//
	// > 👉 **Note**: `cmark-gfm` currently does not support escaped brackets:
	// > <https://github.com/github/cmark-gfm/issues/240>
	//
	// ```markdown
	// > | [^a\*b]: c
	//         ^
	// ```
	labelEscape = func(code Code) State {
		if code == CodeLeftSquareBracket || code == CodeBackslash || code == CodeRightSquareBracket {
			effects.Consume(code)
			size++
			return labelInside
		}

		return labelInside(code)
	}

	// After definition label.
	//
	// ```markdown
	// > | [^a]: b
	//         ^
	// ```
	labelAfter = func(code Code) State {
		if code == CodeColon {
			effects.Enter(TypeDefinitionMarker, nil)
			effects.Consume(code)
			effects.Exit(TypeDefinitionMarker)

			// Upstream pushes even from inside a check (the document's container check runs this
			// tokenizer before the attempt does): a check restores events, not the parser's list.
			if !slices.Contains(*defined, identifier) {
				*defined = append(*defined, identifier)
			}

			// Any whitespace after the marker is eaten, forming indented code
			// is not possible.
			// No space is also fine, just like a block quote marker.
			return factorySpace(effects, whitespaceAfter, typeGfmFootnoteDefinitionWhitespace, 0)
		}

		return nok(code)
	}

	// After definition prefix.
	//
	// ```markdown
	// > | [^a]: b
	//           ^
	// ```
	whitespaceAfter = func(code Code) State {
		// `markdown-rs` has a wrapping token for the prefix that is closed here.
		return ok(code)
	}

	return start
}

func tokenizeGfmFootnoteDefinitionContinuation(self *Self, effects *Effects, ok State, nok State) State {
	// Start of footnote definition continuation.
	//
	// ```markdown
	//   | [^a]: b
	// > |     c
	//     ^
	// ```
	//
	// Either a blank line, which is okay, or an indented thing.
	return effects.Check(blankLine, ok, effects.Attempt(gfmFootnoteIndent, ok, nok))
}

func gfmFootnoteDefinitionEnd(self *Self, effects *Effects) {
	effects.Exit(typeGfmFootnoteDefinition)
}

func tokenizeGfmFootnoteIndent(self *Self, effects *Effects, ok State, nok State) State {
	var afterPrefix State

	afterPrefix = func(code Code) State {
		if len(self.Events) > 0 {
			tail := self.Events[len(self.Events)-1]
			if tail.Token.Type == typeGfmFootnoteDefinitionIndent &&
				utf16Length(tail.Context.SliceSerialize(tail.Token, true)) == 4 {
				return ok(code)
			}
		}
		return nok(code)
	}

	return factorySpace(effects, afterPrefix, typeGfmFootnoteDefinitionIndent, 4+1)
}
