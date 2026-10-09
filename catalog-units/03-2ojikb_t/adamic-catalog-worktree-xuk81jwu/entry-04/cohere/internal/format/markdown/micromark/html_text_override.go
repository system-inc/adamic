package micromark

// htmlTextOverride is the fork's src/language-markdown/parse/micromark/micromark-extension-html-text.js: a
// copy of micromark-core-commonmark's dev/lib/html-text.js (commit 774a70c) modified "to preserve
// whitespace inside quoted attribute values". Diffed against html-text.js state by state, the copy differs
// in two places, both deliberate:
//
//   - The construct carries `add: "before"`. combineExtensions puts every construct that does not say
//     "after" ahead of the existing ones (mergeConstructs in parse.go), so the explicit "before" changes
//     nothing; either way text on `<` tries this, then autolink, then core htmlText.
//   - lineEndingAfter does not eat the next line's leading whitespace as a linePrefix (see that state).
//
// Every other state is core's, unchanged, and so is the name, "htmlText". The states are kept here in
// full, rather than shared with html_text.go, so each file follows its own upstream file line for line.
var htmlTextOverride = &Construct{Name: "htmlText", Tokenize: tokenizeHtmlTextOverride, Add: "before"}

// htmlTextOverrideExtension is the fork's overrideHtmlTextSyntax(): the copy, on `<` in text.
func htmlTextOverrideExtension() *Extension {
	return &Extension{
		Text: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeLessThan: {htmlTextOverride},
		}},
	}
}

func tokenizeHtmlTextOverride(self *Self, effects *Effects, ok State, nok State) State {
	marker := htmlTextMarkerUndefined
	var index int
	var returnState State

	var start, open, declarationOpen, commentOpenInside, comment, commentClose, commentEnd, cdataOpenInside,
		cdata, cdataClose, cdataEnd, declaration, instruction, instructionClose, tagCloseStart, tagClose,
		tagCloseBetween, tagOpen, tagOpenBetween, tagOpenAttributeName, tagOpenAttributeNameAfter,
		tagOpenAttributeValueBefore, tagOpenAttributeValueQuoted, tagOpenAttributeValueUnquoted,
		tagOpenAttributeValueQuotedAfter, end, lineEndingBefore, lineEndingAfter, lineEndingAfterPrefix State

	// Start of HTML (text).
	//
	//	> | a <b> c
	//	      ^
	start = func(code Code) State {
		effects.Enter(TypeHtmlText, nil)
		effects.Enter(TypeHtmlTextData, nil)
		effects.Consume(code)
		return open
	}

	// After `<`, at tag name or other stuff.
	//
	//	> | a <b> c
	//	       ^
	//	> | a <!doctype> c
	//	       ^
	//	> | a <!--b--> c
	//	       ^
	open = func(code Code) State {
		if code == CodeExclamationMark {
			effects.Consume(code)
			return declarationOpen
		}
		if code == CodeSlash {
			effects.Consume(code)
			return tagCloseStart
		}
		if code == CodeQuestionMark {
			effects.Consume(code)
			return instruction
		}

		// ASCII alphabetical.
		if asciiAlpha(code) {
			effects.Consume(code)
			return tagOpen
		}
		return nok(code)
	}

	// After `<!`, at declaration, comment, or CDATA.
	//
	//	> | a <!doctype> c
	//	        ^
	//	> | a <!--b--> c
	//	        ^
	//	> | a <![CDATA[>&<]]> c
	//	        ^
	declarationOpen = func(code Code) State {
		if code == CodeDash {
			effects.Consume(code)
			return commentOpenInside
		}
		if code == CodeLeftSquareBracket {
			effects.Consume(code)
			index = 0
			return cdataOpenInside
		}
		if asciiAlpha(code) {
			effects.Consume(code)
			return declaration
		}
		return nok(code)
	}

	// In a comment, after `<!-`, at another `-`.
	//
	//	> | a <!--b--> c
	//	         ^
	//
	// Upstream returns commentEnd, not comment, after `<!--`: so `<!-->` and `<!--->` are comments, as
	// CommonMark 0.31 says.
	commentOpenInside = func(code Code) State {
		if code == CodeDash {
			effects.Consume(code)
			return commentEnd
		}
		return nok(code)
	}

	// In comment.
	//
	//	> | a <!--b--> c
	//	          ^
	comment = func(code Code) State {
		if code == CodeEof {
			return nok(code)
		}
		if code == CodeDash {
			effects.Consume(code)
			return commentClose
		}
		if markdownLineEnding(code) {
			returnState = comment
			return lineEndingBefore(code)
		}
		effects.Consume(code)
		return comment
	}

	// In comment, after `-`.
	//
	//	> | a <!--b--> c
	//	            ^
	commentClose = func(code Code) State {
		if code == CodeDash {
			effects.Consume(code)
			return commentEnd
		}
		return comment(code)
	}

	// In comment, after `--`.
	//
	//	> | a <!--b--> c
	//	             ^
	commentEnd = func(code Code) State {
		if code == CodeGreaterThan {
			return end(code)
		}
		if code == CodeDash {
			return commentClose(code)
		}
		return comment(code)
	}

	// After `<![`, in CDATA, expecting `CDATA[`.
	//
	//	> | a <![CDATA[>&<]]> b
	//	         ^^^^^^
	cdataOpenInside = func(code Code) State {
		value := cdataOpeningString
		// `value.charCodeAt(index++)`: index increments whether or not the code matches. It never reads
		// past the end, since reaching the length moves on to cdata.
		expected := Code(value[index])
		index++
		if code == expected {
			effects.Consume(code)
			if index == len(value) {
				return cdata
			}
			return cdataOpenInside
		}
		return nok(code)
	}

	// In CDATA.
	//
	//	> | a <![CDATA[>&<]]> b
	//	               ^^^
	cdata = func(code Code) State {
		if code == CodeEof {
			return nok(code)
		}
		if code == CodeRightSquareBracket {
			effects.Consume(code)
			return cdataClose
		}
		if markdownLineEnding(code) {
			returnState = cdata
			return lineEndingBefore(code)
		}
		effects.Consume(code)
		return cdata
	}

	// In CDATA, after `]`, at another `]`.
	//
	//	> | a <![CDATA[>&<]]> b
	//	                   ^
	cdataClose = func(code Code) State {
		if code == CodeRightSquareBracket {
			effects.Consume(code)
			return cdataEnd
		}
		return cdata(code)
	}

	// In CDATA, after `]]`, at `>`.
	//
	//	> | a <![CDATA[>&<]]> b
	//	                    ^
	cdataEnd = func(code Code) State {
		if code == CodeGreaterThan {
			return end(code)
		}
		if code == CodeRightSquareBracket {
			effects.Consume(code)
			return cdataEnd
		}
		return cdata(code)
	}

	// In declaration.
	//
	//	> | a <!b> c
	//	         ^
	//
	// The end of input goes to end, which refuses it: the same nok as the other states, by another road.
	declaration = func(code Code) State {
		if code == CodeEof || code == CodeGreaterThan {
			return end(code)
		}
		if markdownLineEnding(code) {
			returnState = declaration
			return lineEndingBefore(code)
		}
		effects.Consume(code)
		return declaration
	}

	// In instruction.
	//
	//	> | a <?b?> c
	//	        ^
	instruction = func(code Code) State {
		if code == CodeEof {
			return nok(code)
		}
		if code == CodeQuestionMark {
			effects.Consume(code)
			return instructionClose
		}
		if markdownLineEnding(code) {
			returnState = instruction
			return lineEndingBefore(code)
		}
		effects.Consume(code)
		return instruction
	}

	// In instruction, after `?`, at `>`.
	//
	//	> | a <?b?> c
	//	          ^
	instructionClose = func(code Code) State {
		if code == CodeGreaterThan {
			return end(code)
		}
		return instruction(code)
	}

	// After `</`, in closing tag, at tag name.
	//
	//	> | a </b> c
	//	        ^
	tagCloseStart = func(code Code) State {
		// ASCII alphabetical.
		if asciiAlpha(code) {
			effects.Consume(code)
			return tagClose
		}
		return nok(code)
	}

	// After `</x`, in a tag name.
	//
	//	> | a </b> c
	//	         ^
	tagClose = func(code Code) State {
		// ASCII alphanumerical and `-`.
		if code == CodeDash || asciiAlphanumeric(code) {
			effects.Consume(code)
			return tagClose
		}
		return tagCloseBetween(code)
	}

	// In closing tag, after tag name.
	//
	//	> | a </b> c
	//	         ^
	tagCloseBetween = func(code Code) State {
		if markdownLineEnding(code) {
			returnState = tagCloseBetween
			return lineEndingBefore(code)
		}
		if markdownSpace(code) {
			effects.Consume(code)
			return tagCloseBetween
		}
		return end(code)
	}

	// After `<x`, in opening tag name.
	//
	//	> | a <b> c
	//	        ^
	tagOpen = func(code Code) State {
		// ASCII alphanumerical and `-`.
		if code == CodeDash || asciiAlphanumeric(code) {
			effects.Consume(code)
			return tagOpen
		}
		if code == CodeSlash || code == CodeGreaterThan || markdownLineEndingOrSpace(code) {
			return tagOpenBetween(code)
		}
		return nok(code)
	}

	// In opening tag, after tag name.
	//
	//	> | a <b> c
	//	        ^
	tagOpenBetween = func(code Code) State {
		if code == CodeSlash {
			effects.Consume(code)
			return end
		}

		// ASCII alphabetical and `:` and `_`.
		if code == CodeColon || code == CodeUnderscore || asciiAlpha(code) {
			effects.Consume(code)
			return tagOpenAttributeName
		}
		if markdownLineEnding(code) {
			returnState = tagOpenBetween
			return lineEndingBefore(code)
		}
		if markdownSpace(code) {
			effects.Consume(code)
			return tagOpenBetween
		}
		return end(code)
	}

	// In attribute name.
	//
	//	> | a <b c> d
	//	         ^
	tagOpenAttributeName = func(code Code) State {
		// ASCII alphabetical and `-`, `.`, `:`, and `_`.
		if code == CodeDash || code == CodeDot || code == CodeColon || code == CodeUnderscore ||
			asciiAlphanumeric(code) {
			effects.Consume(code)
			return tagOpenAttributeName
		}
		return tagOpenAttributeNameAfter(code)
	}

	// After attribute name, before initializer, the end of the tag, or whitespace.
	//
	//	> | a <b c> d
	//	          ^
	tagOpenAttributeNameAfter = func(code Code) State {
		if code == CodeEqualsTo {
			effects.Consume(code)
			return tagOpenAttributeValueBefore
		}
		if markdownLineEnding(code) {
			returnState = tagOpenAttributeNameAfter
			return lineEndingBefore(code)
		}
		if markdownSpace(code) {
			effects.Consume(code)
			return tagOpenAttributeNameAfter
		}
		return tagOpenBetween(code)
	}

	// Before unquoted, double quoted, or single quoted attribute value, allowing whitespace.
	//
	//	> | a <b c=d> e
	//	           ^
	tagOpenAttributeValueBefore = func(code Code) State {
		if code == CodeEof || code == CodeLessThan || code == CodeEqualsTo || code == CodeGreaterThan ||
			code == CodeGraveAccent {
			return nok(code)
		}
		if code == CodeQuotationMark || code == CodeApostrophe {
			effects.Consume(code)
			marker = code
			return tagOpenAttributeValueQuoted
		}
		if markdownLineEnding(code) {
			returnState = tagOpenAttributeValueBefore
			return lineEndingBefore(code)
		}
		if markdownSpace(code) {
			effects.Consume(code)
			return tagOpenAttributeValueBefore
		}
		effects.Consume(code)
		return tagOpenAttributeValueUnquoted
	}

	// In double or single quoted attribute value.
	//
	//	> | a <b c="d"> e
	//	            ^
	tagOpenAttributeValueQuoted = func(code Code) State {
		if code == marker {
			effects.Consume(code)
			marker = htmlTextMarkerUndefined
			return tagOpenAttributeValueQuotedAfter
		}
		if code == CodeEof {
			return nok(code)
		}
		if markdownLineEnding(code) {
			returnState = tagOpenAttributeValueQuoted
			return lineEndingBefore(code)
		}
		effects.Consume(code)
		return tagOpenAttributeValueQuoted
	}

	// In unquoted attribute value.
	//
	//	> | a <b c=d> e
	//	           ^
	tagOpenAttributeValueUnquoted = func(code Code) State {
		if code == CodeEof || code == CodeQuotationMark || code == CodeApostrophe || code == CodeLessThan ||
			code == CodeEqualsTo || code == CodeGraveAccent {
			return nok(code)
		}
		if code == CodeSlash || code == CodeGreaterThan || markdownLineEndingOrSpace(code) {
			return tagOpenBetween(code)
		}
		effects.Consume(code)
		return tagOpenAttributeValueUnquoted
	}

	// After double or single quoted attribute value, before whitespace or the end of the tag.
	//
	//	> | a <b c="d"> e
	//	              ^
	tagOpenAttributeValueQuotedAfter = func(code Code) State {
		if code == CodeSlash || code == CodeGreaterThan || markdownLineEndingOrSpace(code) {
			return tagOpenBetween(code)
		}
		return nok(code)
	}

	// In certain circumstances of a tag where only an `>` is allowed.
	//
	//	> | a <b c="d"> e
	//	              ^
	end = func(code Code) State {
		if code == CodeGreaterThan {
			effects.Consume(code)
			effects.Exit(TypeHtmlTextData)
			effects.Exit(TypeHtmlText)
			return ok
		}
		return nok(code)
	}

	// At eol.
	//
	// > 👉 **Note**: we can’t have blank lines in text, so no need to worry about empty tokens.
	//
	//	> | a <!--a
	//	           ^
	//	  | b-->
	lineEndingBefore = func(code Code) State {
		effects.Exit(TypeHtmlTextData)
		effects.Enter(TypeLineEnding, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEnding)
		return lineEndingAfter
	}

	// After eol, at optional whitespace.
	//
	// > 👉 **Note**: we can’t have blank lines in text, so no need to worry about empty tokens.
	//
	//	  | a <!--a
	//	> | b-->
	//	    ^
	//
	// THE FORK'S CHANGE, the only one in the state machine. Core html-text.js eats the whitespace at the
	// start of the line here with factorySpace, as a `linePrefix` token outside htmlTextData (at most 3
	// columns, factorySpace's max of 4 less one, unless codeIndented is disabled; the rest falls to
	// returnState, which consumes it as data).
	// The fork goes straight to lineEndingAfterPrefix, so that whitespace opens the next htmlTextData and
	// stays in the node's value. Its header says "inside quoted attribute values", but the change applies
	// after every line ending in the construct: comments, CDATA, declarations, instructions, closing tags,
	// between attributes. Every returnState consumes spaces and tabs itself, so the change moves token
	// boundaries without changing what is accepted.
	lineEndingAfter = func(code Code) State {
		// Always populated by defaults.

		return lineEndingAfterPrefix(code)
	}

	// After eol, after optional whitespace.
	//
	// > 👉 **Note**: we can’t have blank lines in text, so no need to worry about empty tokens.
	//
	//	  | a <!--a
	//	> | b-->
	//	    ^
	lineEndingAfterPrefix = func(code Code) State {
		effects.Enter(TypeHtmlTextData, nil)
		return returnState(code)
	}

	return start
}
