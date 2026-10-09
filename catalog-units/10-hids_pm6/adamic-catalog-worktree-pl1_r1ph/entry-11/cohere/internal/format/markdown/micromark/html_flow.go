package micromark

import (
	"slices"
	"strings"
)

// htmlFlow is micromark-core-commonmark/lib/html-flow.js.
var htmlFlow = &Construct{
	Concrete:  true,
	Name:      "htmlFlow",
	ResolveTo: &Resolver{Resolve: resolveToHtmlFlow},
	Tokenize:  tokenizeHtmlFlow,
}

// htmlFlowBlankLineBefore is upstream's file-local `blankLineBefore`.
var htmlFlowBlankLineBefore = &Construct{Partial: true, Tokenize: tokenizeHtmlFlowBlankLineBefore}

// htmlFlowNonLazyContinuationStart is upstream's file-local `nonLazyContinuationStart`.
var htmlFlowNonLazyContinuationStart = &Construct{Partial: true, Tokenize: tokenizeHtmlFlowNonLazyContinuationStart}

// htmlBlockNames is micromark-util-html-tag-name's htmlBlockNames, copied exactly: the lowercase HTML
// “block” tag names, which get the relaxed rules of condition 6.
var htmlBlockNames = []string{
	"address",
	"article",
	"aside",
	"base",
	"basefont",
	"blockquote",
	"body",
	"caption",
	"center",
	"col",
	"colgroup",
	"dd",
	"details",
	"dialog",
	"dir",
	"div",
	"dl",
	"dt",
	"fieldset",
	"figcaption",
	"figure",
	"footer",
	"form",
	"frame",
	"frameset",
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"head",
	"header",
	"hr",
	"html",
	"iframe",
	"legend",
	"li",
	"link",
	"main",
	"menu",
	"menuitem",
	"nav",
	"noframes",
	"ol",
	"optgroup",
	"option",
	"p",
	"param",
	"search",
	"section",
	"summary",
	"table",
	"tbody",
	"td",
	"tfoot",
	"th",
	"thead",
	"title",
	"tr",
	"track",
	"ul",
}

// htmlRawNames is micromark-util-html-tag-name's htmlRawNames, copied exactly: the lowercase HTML “raw”
// tag names of condition 1, whose HTML runs until a closing tag also in this list.
var htmlRawNames = []string{"pre", "script", "style", "textarea"}

func resolveToHtmlFlow(events []Event, context *TokenizeContext) []Event {
	index := len(events)

	// `while (index--)`: stops at the last enter of htmlFlow, or runs out at -1.
	for {
		index--
		if index < 0 {
			break
		}
		if events[index].Enter && events[index].Token.Type == TypeHtmlFlow {
			break
		}
	}

	if index > 1 && events[index-2].Token.Type == TypeLinePrefix {
		// Upstream shares the prefix's start point object with both tokens; nothing mutates a point in place
		// afterwards, so value copies are equivalent.
		// Add the prefix start to the HTML token.
		events[index].Token.Start = events[index-2].Token.Start
		// Add the prefix start to the HTML line token.
		events[index+1].Token.Start = events[index-2].Token.Start
		// Remove the line prefix.
		events = splice(events, index-2, 2, nil)
	}

	return events
}

func tokenizeHtmlFlow(self *Self, effects *Effects, ok State, nok State) State {
	var marker int
	var closingTag bool
	// buffer only ever holds ASCII letters, digits and `-`, so a Go string's bytes are its code units.
	var buffer string
	var index int
	// markerB is a quote while completeAttributeValueQuoted runs; upstream resets it to null after the
	// closing quote, which no state reads before setting it again.
	var markerB Code

	var start, before, open, declarationOpen, commentOpenInside, cdataOpenInside, tagCloseStart, tagName,
		basicSelfClosing, completeClosingTagAfter, completeAttributeNameBefore, completeAttributeName,
		completeAttributeNameAfter, completeAttributeValueBefore, completeAttributeValueQuoted,
		completeAttributeValueUnquoted, completeAttributeValueQuotedAfter, completeEnd, completeAfter,
		continuation, continuationStart, continuationStartNonLazy, continuationBefore,
		continuationCommentInside, continuationRawTagOpen, continuationRawEndTag, continuationCdataInside,
		continuationDeclarationInside, continuationClose, continuationAfter State

	// Start of HTML (flow).
	start = func(code Code) State {
		// To do: parse indent like `markdown-rs`.
		return before(code)
	}

	// At `<`, after optional whitespace.
	before = func(code Code) State {
		effects.Enter(TypeHtmlFlow, nil)
		effects.Enter(TypeHtmlFlowData, nil)
		effects.Consume(code)
		return open
	}

	// After `<`, at tag name or other stuff.
	open = func(code Code) State {
		if code == CodeExclamationMark {
			effects.Consume(code)
			return declarationOpen
		}

		if code == CodeSlash {
			effects.Consume(code)
			closingTag = true
			return tagCloseStart
		}

		if code == CodeQuestionMark {
			effects.Consume(code)
			marker = htmlInstruction
			// To do:
			// tokenizer.concrete = true
			// To do: use `markdown-rs` style interrupt.
			// While we’re in an instruction instead of a declaration, we’re on a `?`
			// right now, so we do need to search for `>`, similar to declarations.
			if self.IsInterrupt() {
				return ok
			}
			return continuationDeclarationInside
		}

		// ASCII alphabetical.
		if asciiAlpha(code) {
			// Always the case.
			effects.Consume(code)
			buffer = string(rune(code))
			return tagName
		}

		return nok(code)
	}

	// After `<!`, at declaration, comment, or CDATA.
	declarationOpen = func(code Code) State {
		if code == CodeDash {
			effects.Consume(code)
			marker = htmlComment
			return commentOpenInside
		}

		if code == CodeLeftSquareBracket {
			effects.Consume(code)
			marker = htmlCdata
			index = 0
			return cdataOpenInside
		}

		// ASCII alphabetical.
		if asciiAlpha(code) {
			effects.Consume(code)
			marker = htmlDeclaration
			// // Do not form containers.
			// tokenizer.concrete = true
			if self.IsInterrupt() {
				return ok
			}
			return continuationDeclarationInside
		}

		return nok(code)
	}

	// After `<!-`, inside a comment, at another `-`.
	commentOpenInside = func(code Code) State {
		if code == CodeDash {
			effects.Consume(code)
			// // Do not form containers.
			// tokenizer.concrete = true
			if self.IsInterrupt() {
				return ok
			}
			return continuationDeclarationInside
		}

		return nok(code)
	}

	// After `<![`, inside CDATA, expecting `CDATA[`.
	cdataOpenInside = func(code Code) State {
		value := cdataOpeningString
		// `value.charCodeAt(index++)`: the increment happens whether or not the comparison passes. index
		// never passes the string's length, because a match at the last unit returns before reading again.
		expected := Code(value[index])
		index++
		if code == expected {
			effects.Consume(code)

			if index == len(value) {
				// // Do not form containers.
				// tokenizer.concrete = true
				if self.IsInterrupt() {
					return ok
				}
				return continuation
			}

			return cdataOpenInside
		}

		return nok(code)
	}

	// After `</`, in closing tag, at tag name.
	tagCloseStart = func(code Code) State {
		if asciiAlpha(code) {
			// Always the case.
			effects.Consume(code)
			buffer = string(rune(code))
			return tagName
		}

		return nok(code)
	}

	// In tag name.
	tagName = func(code Code) State {
		if code == CodeEof || code == CodeSlash || code == CodeGreaterThan || markdownLineEndingOrSpace(code) {
			slash := code == CodeSlash
			name := strings.ToLower(buffer)

			if !slash && !closingTag && slices.Contains(htmlRawNames, name) {
				marker = htmlRaw
				// // Do not form containers.
				// tokenizer.concrete = true
				if self.IsInterrupt() {
					return ok(code)
				}
				return continuation(code)
			}

			if slices.Contains(htmlBlockNames, strings.ToLower(buffer)) {
				marker = htmlBasic

				if slash {
					effects.Consume(code)
					return basicSelfClosing
				}

				// // Do not form containers.
				// tokenizer.concrete = true
				if self.IsInterrupt() {
					return ok(code)
				}
				return continuation(code)
			}

			marker = htmlComplete
			// Do not support complete HTML when interrupting.
			if self.IsInterrupt() && !self.Parser.Lazy[self.Now().Line] {
				return nok(code)
			}
			if closingTag {
				return completeClosingTagAfter(code)
			}
			return completeAttributeNameBefore(code)
		}

		// ASCII alphanumerical and `-`.
		if code == CodeDash || asciiAlphanumeric(code) {
			effects.Consume(code)
			buffer += string(rune(code))
			return tagName
		}

		return nok(code)
	}

	// After closing slash of a basic tag name.
	basicSelfClosing = func(code Code) State {
		if code == CodeGreaterThan {
			effects.Consume(code)
			// // Do not form containers.
			// tokenizer.concrete = true
			if self.IsInterrupt() {
				return ok
			}
			return continuation
		}

		return nok(code)
	}

	// After closing slash of a complete tag name.
	completeClosingTagAfter = func(code Code) State {
		if markdownSpace(code) {
			effects.Consume(code)
			return completeClosingTagAfter
		}

		return completeEnd(code)
	}

	// At an attribute name.
	//
	// At first, this state is used after a complete tag name, after whitespace, where it expects optional
	// attributes or the end of the tag. It is also reused after attributes, when expecting more optional
	// attributes.
	completeAttributeNameBefore = func(code Code) State {
		if code == CodeSlash {
			effects.Consume(code)
			return completeEnd
		}

		// ASCII alphanumerical and `:` and `_`.
		if code == CodeColon || code == CodeUnderscore || asciiAlpha(code) {
			effects.Consume(code)
			return completeAttributeName
		}

		if markdownSpace(code) {
			effects.Consume(code)
			return completeAttributeNameBefore
		}

		return completeEnd(code)
	}

	// In attribute name.
	completeAttributeName = func(code Code) State {
		// ASCII alphanumerical and `-`, `.`, `:`, and `_`.
		if code == CodeDash || code == CodeDot || code == CodeColon || code == CodeUnderscore || asciiAlphanumeric(code) {
			effects.Consume(code)
			return completeAttributeName
		}

		return completeAttributeNameAfter(code)
	}

	// After attribute name, at an optional initializer, the end of the tag, or whitespace.
	completeAttributeNameAfter = func(code Code) State {
		if code == CodeEqualsTo {
			effects.Consume(code)
			return completeAttributeValueBefore
		}

		if markdownSpace(code) {
			effects.Consume(code)
			return completeAttributeNameAfter
		}

		return completeAttributeNameBefore(code)
	}

	// Before unquoted, double quoted, or single quoted attribute value, allowing whitespace.
	completeAttributeValueBefore = func(code Code) State {
		if code == CodeEof || code == CodeLessThan || code == CodeEqualsTo || code == CodeGreaterThan || code == CodeGraveAccent {
			return nok(code)
		}

		if code == CodeQuotationMark || code == CodeApostrophe {
			effects.Consume(code)
			markerB = code
			return completeAttributeValueQuoted
		}

		if markdownSpace(code) {
			effects.Consume(code)
			return completeAttributeValueBefore
		}

		return completeAttributeValueUnquoted(code)
	}

	// In double or single quoted attribute value.
	completeAttributeValueQuoted = func(code Code) State {
		if code == markerB {
			effects.Consume(code)
			markerB = CodeEof
			return completeAttributeValueQuotedAfter
		}

		if code == CodeEof || markdownLineEnding(code) {
			return nok(code)
		}

		effects.Consume(code)
		return completeAttributeValueQuoted
	}

	// In unquoted attribute value.
	completeAttributeValueUnquoted = func(code Code) State {
		if code == CodeEof ||
			code == CodeQuotationMark ||
			code == CodeApostrophe ||
			code == CodeSlash ||
			code == CodeLessThan ||
			code == CodeEqualsTo ||
			code == CodeGreaterThan ||
			code == CodeGraveAccent ||
			markdownLineEndingOrSpace(code) {
			return completeAttributeNameAfter(code)
		}

		effects.Consume(code)
		return completeAttributeValueUnquoted
	}

	// After double or single quoted attribute value, before whitespace or the end of the tag.
	completeAttributeValueQuotedAfter = func(code Code) State {
		if code == CodeSlash || code == CodeGreaterThan || markdownSpace(code) {
			return completeAttributeNameBefore(code)
		}

		return nok(code)
	}

	// In certain circumstances of a complete tag where only an `>` is allowed.
	completeEnd = func(code Code) State {
		if code == CodeGreaterThan {
			effects.Consume(code)
			return completeAfter
		}

		return nok(code)
	}

	// After `>` in a complete tag.
	completeAfter = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			// // Do not form containers.
			// tokenizer.concrete = true
			return continuation(code)
		}

		if markdownSpace(code) {
			effects.Consume(code)
			return completeAfter
		}

		return nok(code)
	}

	// In continuation of any HTML kind.
	continuation = func(code Code) State {
		if code == CodeDash && marker == htmlComment {
			effects.Consume(code)
			return continuationCommentInside
		}

		if code == CodeLessThan && marker == htmlRaw {
			effects.Consume(code)
			return continuationRawTagOpen
		}

		if code == CodeGreaterThan && marker == htmlDeclaration {
			effects.Consume(code)
			return continuationClose
		}

		if code == CodeQuestionMark && marker == htmlInstruction {
			effects.Consume(code)
			return continuationDeclarationInside
		}

		if code == CodeRightSquareBracket && marker == htmlCdata {
			effects.Consume(code)
			return continuationCdataInside
		}

		if markdownLineEnding(code) && (marker == htmlBasic || marker == htmlComplete) {
			effects.Exit(TypeHtmlFlowData)
			return effects.Check(htmlFlowBlankLineBefore, continuationAfter, continuationStart)(code)
		}

		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeHtmlFlowData)
			return continuationStart(code)
		}

		effects.Consume(code)
		return continuation
	}

	// In continuation, at eol.
	continuationStart = func(code Code) State {
		return effects.Check(htmlFlowNonLazyContinuationStart, continuationStartNonLazy, continuationAfter)(code)
	}

	// In continuation, at eol, before non-lazy content.
	continuationStartNonLazy = func(code Code) State {
		effects.Enter(TypeLineEnding, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEnding)
		return continuationBefore
	}

	// In continuation, before non-lazy content.
	continuationBefore = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			return continuationStart(code)
		}

		effects.Enter(TypeHtmlFlowData, nil)
		return continuation(code)
	}

	// In comment continuation, after one `-`, expecting another.
	continuationCommentInside = func(code Code) State {
		if code == CodeDash {
			effects.Consume(code)
			return continuationDeclarationInside
		}

		return continuation(code)
	}

	// In raw continuation, after `<`, at `/`.
	continuationRawTagOpen = func(code Code) State {
		if code == CodeSlash {
			effects.Consume(code)
			buffer = ""
			return continuationRawEndTag
		}

		return continuation(code)
	}

	// In raw continuation, after `</`, in a raw tag name.
	continuationRawEndTag = func(code Code) State {
		if code == CodeGreaterThan {
			name := strings.ToLower(buffer)

			if slices.Contains(htmlRawNames, name) {
				effects.Consume(code)
				return continuationClose
			}

			return continuation(code)
		}

		if asciiAlpha(code) && len(buffer) < htmlRawSizeMax {
			// Always the case.
			effects.Consume(code)
			buffer += string(rune(code))
			return continuationRawEndTag
		}

		return continuation(code)
	}

	// In cdata continuation, after `]`, expecting `]>`.
	continuationCdataInside = func(code Code) State {
		if code == CodeRightSquareBracket {
			effects.Consume(code)
			return continuationDeclarationInside
		}

		return continuation(code)
	}

	// In declaration or instruction continuation, at `>`.
	continuationDeclarationInside = func(code Code) State {
		if code == CodeGreaterThan {
			effects.Consume(code)
			return continuationClose
		}

		// More dashes.
		if code == CodeDash && marker == htmlComment {
			effects.Consume(code)
			return continuationDeclarationInside
		}

		return continuation(code)
	}

	// In closed continuation: everything we get until the eol/eof is part of it.
	continuationClose = func(code Code) State {
		if code == CodeEof || markdownLineEnding(code) {
			effects.Exit(TypeHtmlFlowData)
			return continuationAfter(code)
		}

		effects.Consume(code)
		return continuationClose
	}

	// Done.
	continuationAfter = func(code Code) State {
		effects.Exit(TypeHtmlFlow)
		// // Feel free to interrupt.
		// tokenizer.interrupt = false
		// // No longer concrete.
		// tokenizer.concrete = false
		return ok(code)
	}

	return start
}

func tokenizeHtmlFlowNonLazyContinuationStart(self *Self, effects *Effects, ok State, nok State) State {
	var start, after State

	// At eol, before continuation.
	start = func(code Code) State {
		if markdownLineEnding(code) {
			effects.Enter(TypeLineEnding, nil)
			effects.Consume(code)
			effects.Exit(TypeLineEnding)
			return after
		}

		return nok(code)
	}

	// A continuation.
	after = func(code Code) State {
		if self.Parser.Lazy[self.Now().Line] {
			return nok(code)
		}
		return ok(code)
	}

	return start
}

func tokenizeHtmlFlowBlankLineBefore(self *Self, effects *Effects, ok State, nok State) State {
	var start State

	// Before eol, expecting blank line.
	start = func(code Code) State {
		effects.Enter(TypeLineEnding, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEnding)
		return effects.Attempt(blankLine, ok, nok)
	}

	return start
}
