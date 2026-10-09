// The CSS parser: the character loop that turns a stylesheet into the AST in ast.go.
//
// Ported from `src/css-parser.ts` at Tailwind 4.3.3, read at the pinned tag rather than from the
// minified bundle, so a disagreement is a finding rather than version skew.
//
// # Why this component exists at all
//
// Every other piece of the Tailwind port could in principle have been replaced by a generated
// table. This one could not, and it is the reason the port is necessary rather than an
// optimization. The tables cohere shipped before this were generated per-repo: running the
// generator against two repositories on the same Tailwind 4.3.3 produced tables that differed,
// because each project's own `@utility` blocks and `@theme` entries were baked in as though they
// were framework facts. This parser is what lets cohere read the repository in front of it.
//
// # Why not tdewolff/parse/v2/css
//
// It was evaluated by running it, not by reading it. `tdewolff/parse/v2/css` matches at-rule names
// against a whitelist of known at-rules, and every at-rule Tailwind v4 cares about — `@theme`,
// `@utility`, `@custom-variant`, `@variant` — misses that whitelist and falls through to
// `parseAtRuleUnknown`, which yields a flat token run and no block structure:
//
//	@theme (unknown)   Declaration=0  Token=6  BeginRuleset=0
//	@media (known)     Declaration=1  Token=0  BeginRuleset=1
//
// The block-structure state machine would have to be written by hand on top of it anyway, while
// additionally fighting three behaviours it has by design: it lowercases at-rule names, it elides
// whitespace inside at-rule params (so `(&:where(.dark, .dark *))` comes back with the space after
// the comma deleted, which changes a `@custom-variant` selector), and its `Values()` hands back a
// recycled internal buffer that the next call overwrites.
//
// The upstream file is 717 lines but simpler than that suggests: one loop over bytes, about
// fourteen branches, five pieces of state. Porting it directly costs no dependency and keeps the
// result diffable against upstream when Tailwind changes.
//
// # Bytes, not runes
//
// Upstream indexes UTF-16 code units and every character it compares against is ASCII. This port
// indexes bytes. The two agree because every byte of a multi-byte UTF-8 sequence has its high bit
// set and therefore equals none of the ASCII constants below, so multi-byte text is copied through
// buffers untouched exactly as it is upstream. Indexing runes instead would be slower and would
// change nothing.
package tailwind

import (
	"fmt"
	"strings"

	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// The byte constants upstream names, kept at their upstream names so the port diffs against the
// source it came from.
const (
	byteBackslash       = '\\'
	byteSlash           = '/'
	byteAsterisk        = '*'
	byteDoubleQuote     = '"'
	byteSingleQuote     = '\''
	byteColon           = ':'
	byteSemicolon       = ';'
	byteLineBreak       = '\n'
	byteCarriageReturn  = '\r'
	byteSpace           = ' '
	byteTab             = '\t'
	byteOpenCurly       = '{'
	byteCloseCurly      = '}'
	byteOpenParen       = '('
	byteCloseParen      = ')'
	byteOpenBracket     = '['
	byteCloseBracket    = ']'
	byteDash            = '-'
	byteAtSign          = '@'
	byteExclamationMark = '!'
)

// CSSSyntaxError is a malformed-stylesheet error.
//
// Upstream carries a SourceLocation so it can render `file:line:column`. This port carries the byte
// offset instead: cohere reports on the class literal in a user's source file, never on CSS it
// parsed, so an offset is enough to identify the input that failed and a line table would be
// machinery with no reader. The message text matches upstream so a failure here is greppable
// against the upstream source that produced it.
type CSSSyntaxError struct {
	// Message is the upstream message text, without any location prefix.
	Message string
	// Offset is the byte offset into the parsed input where the error was detected.
	Offset int
}

func (err *CSSSyntaxError) Error() string {
	return fmt.Sprintf("%d: %s", err.Offset, err.Message)
}

// ParseCSS parses a stylesheet into the AST in ast.go.
//
// The returned nodes are the top level of the stylesheet, with license comments (`/*! ... */`)
// hoisted to the front, matching upstream. Any other comment is discarded during parsing rather
// than becoming a node; that is upstream's behaviour, not a simplification, and it is why the
// declaration splitter below can assume no comment contains the `:` it searches for.
//
// A malformed stylesheet returns a *CSSSyntaxError and a nil tree.
func ParseCSS(input string) ([]*Node, error) {
	// A leading byte-order mark is replaced by a space rather than stripped. Upstream notes that
	// any transformation before the loop must not change the length of the input, because source
	// offsets are taken against it; the same reasoning keeps the offsets in CSSSyntaxError honest.
	const byteOrderMark = "\ufeff"
	if strings.HasPrefix(input, byteOrderMark) {
		input = " " + input[len(byteOrderMark):]
	}

	var ast []*Node
	var licenseComments []*Node

	// stack holds the ancestors of parent. A nil entry is the root, which is why this is a slice of
	// pointers rather than a slice of values: upstream pushes `null` onto it for the top level and
	// the close-brace branch distinguishes "popped the root" from "popped a real parent".
	var stack []*Node
	var parent *Node
	var node *Node

	var buffer strings.Builder
	// closingBracketStack is the expected closer for every block currently open, innermost last.
	// Upstream is a string built by concatenation; a byte slice is the same structure without the
	// per-push allocation.
	var closingBracketStack []byte

	// bufferStart is the offset of the first non-whitespace byte in buffer, used for error offsets.
	bufferStart := 0

	// bufferLast reports the final byte of buffer, or 0 when it is empty. strings.Builder cannot be
	// read back, and the loop needs the last byte in one branch and the first in three.
	var bufferFirst, bufferLastByte byte
	appendToBuffer := func(text string) {
		if len(text) == 0 {
			return
		}
		if buffer.Len() == 0 {
			bufferFirst = text[0]
		}
		buffer.WriteString(text)
		bufferLastByte = text[len(text)-1]
	}
	resetBuffer := func() {
		buffer.Reset()
		bufferFirst = 0
		bufferLastByte = 0
	}

	// peekAt returns the byte at index, or 0 past the end. Upstream reads `charCodeAt` past the end
	// and gets NaN, which compares false against every constant; 0 has the same property here
	// because no constant above is 0.
	peekAt := func(index int) byte {
		if index < 0 || index >= len(input) {
			return 0
		}
		return input[index]
	}

	for i := 0; i < len(input); i++ {
		currentChar := input[i]

		// Skip the CR of a CRLF so that every branch below only has to recognize a bare line break.
		if currentChar == byteCarriageReturn && peekAt(i+1) == byteLineBreak {
			continue
		}

		switch {
		// A backslash escapes the byte after it. Both are consumed together, so an escaped `:` or
		// `;` inside a selector such as `.hover\:foo:hover` cannot be mistaken for structure.
		case currentChar == byteBackslash:
			if buffer.Len() == 0 {
				bufferStart = i
			}
			end := i + 2
			if end > len(input) {
				end = len(input)
			}
			appendToBuffer(input[i:end])
			i++

		// Start of a comment. The comment is consumed and dropped, except for a license comment
		// (`/*!`), which is hoisted to the top of the output.
		case currentChar == byteSlash && peekAt(i+1) == byteAsterisk:
			start := i
			for j := i + 2; j < len(input); j++ {
				peekChar := input[j]
				if peekChar == byteBackslash {
					j++
				} else if peekChar == byteAsterisk && peekAt(j+1) == byteSlash {
					i = j + 1
					break
				}
			}

			commentString := input[start:min(i+1, len(input))]

			if len(commentString) > 2 && commentString[2] == byteExclamationMark {
				// The `/*` and `*/` are stripped, leaving `! ...`, matching upstream's slice.
				licenseComments = append(licenseComments, Comment(commentString[2:len(commentString)-2]))
			}

		// Start of a string. The whole string, quotes included, joins the buffer verbatim; a `;`,
		// `}` or `:` inside it is text rather than structure.
		case currentChar == byteSingleQuote || currentChar == byteDoubleQuote:
			end, err := parseCSSString(input, i, currentChar)
			if err != nil {
				return nil, err
			}
			appendToBuffer(input[i : end+1])
			i = end

		// Collapse a run of whitespace: skip this one if the next is also whitespace. This is what
		// keeps the AST from carrying the stylesheet's indentation.
		case (currentChar == byteSpace || currentChar == byteLineBreak || currentChar == byteTab) &&
			isFollowedByWhitespace(input, i, peekAt):
			continue

		// A line break that survived the collapse above becomes a single space, and only when the
		// buffer does not already end in whitespace.
		case currentChar == byteLineBreak:
			if buffer.Len() == 0 {
				continue
			}
			if bufferLastByte != byteSpace && bufferLastByte != byteLineBreak && bufferLastByte != byteTab {
				appendToBuffer(" ")
			}

		// Start of a custom property, and the branch this port exists to get right.
		//
		// A custom property's value is almost unrestricted: it may legally contain `;` and `}`, so
		// neither can be used to find its end. This scans forward tracking bracket depth and ends
		// only at a `;` or `}` seen at depth zero.
		//
		// Two consequences are easy to miss and both were measured against the engine rather than
		// reasoned about. The value is taken by slicing the input directly, so it keeps its
		// original interior whitespace including newlines — the whitespace-collapsing branches
		// above never run over it. `--font-sans` in a real Structure theme comes back from the
		// engine holding a literal `\n        `, and a parser that routed these bytes through the
		// normal buffer path returns the same value with a single space instead: still plausible,
		// still wrong, and wrong identically on all 420 theme entries. And the colon index is
		// recorded during this scan rather than searched for afterwards, because a `:` may appear
		// inside the value (in a `url()`, or a nested `var()` fallback) and `indexOf` would find
		// the wrong one.
		case currentChar == byteDash && peekAt(i+1) == byteDash && buffer.Len() == 0:
			// Shadows the outer stack deliberately, exactly as upstream does: brackets opened
			// inside a custom property are balanced within it and must not disturb the block
			// structure the outer loop is tracking.
			var closingBracketStack []byte

			start := i
			colonIndex := -1
			// terminated records whether the scan found an end. When the input runs out without
			// one, upstream leaves `i` where it was and the buffer unfilled; the declaration
			// splitter then fails, which is the correct outcome for a truncated stylesheet.
			terminated := false

			for j := i + 2; j < len(input); j++ {
				peekChar := input[j]

				switch {
				case peekChar == byteBackslash:
					j++

				case peekChar == byteSingleQuote || peekChar == byteDoubleQuote:
					end, err := parseCSSString(input, j, peekChar)
					if err != nil {
						return nil, err
					}
					j = end

				case peekChar == byteSlash && peekAt(j+1) == byteAsterisk:
					for k := j + 2; k < len(input); k++ {
						innerPeek := input[k]
						if innerPeek == byteBackslash {
							k++
						} else if innerPeek == byteAsterisk && peekAt(k+1) == byteSlash {
							j = k + 1
							break
						}
					}

				// The first colon separates the property from the value. It is recorded as an index
				// into the buffer-relative string that parseCSSDeclaration will be handed, which is
				// why the current buffer length is part of the arithmetic.
				case colonIndex == -1 && peekChar == byteColon:
					colonIndex = buffer.Len() + j - start

				case peekChar == byteSemicolon && len(closingBracketStack) == 0:
					appendToBuffer(input[start:j])
					i = j
					terminated = true

				case peekChar == byteOpenParen:
					closingBracketStack = append(closingBracketStack, byteCloseParen)

				case peekChar == byteOpenBracket:
					closingBracketStack = append(closingBracketStack, byteCloseBracket)

				case peekChar == byteOpenCurly:
					closingBracketStack = append(closingBracketStack, byteCloseCurly)

				// A custom property that ends at the enclosing block's `}` rather than at a `;`,
				// or that ends at the end of the input. `i` is set to `j - 1` so the outer loop's
				// increment lands on the `}` itself and the close-block branch runs for it.
				case (peekChar == byteCloseCurly || j == len(input)-1) && len(closingBracketStack) == 0:
					i = j - 1
					appendToBuffer(input[start:j])
					terminated = true

				case peekChar == byteCloseParen || peekChar == byteCloseBracket || peekChar == byteCloseCurly:
					if len(closingBracketStack) > 0 && peekChar == closingBracketStack[len(closingBracketStack)-1] {
						closingBracketStack = closingBracketStack[:len(closingBracketStack)-1]
					}
				}

				if terminated {
					break
				}
			}

			declaration := parseCSSDeclaration(buffer.String(), colonIndex)
			if declaration == nil {
				return nil, &CSSSyntaxError{Message: "Invalid custom property, expected a value", Offset: start}
			}

			if parent != nil {
				parent.Nodes = append(parent.Nodes, declaration)
			} else {
				ast = append(ast, declaration)
			}

			resetBuffer()

		// End of a body-less at-rule, such as `@charset "UTF-8";` or `@import "tailwindcss";`.
		case currentChar == byteSemicolon && bufferFirst == byteAtSign:
			node = ParseAtRule(buffer.String())

			if parent != nil {
				parent.Nodes = append(parent.Nodes, node)
			} else {
				ast = append(ast, node)
			}

			resetBuffer()
			node = nil

		// End of an ordinary declaration. The `)` guard keeps a `;` inside parentheses — which is
		// legal inside, for example, a `url()` — from ending the declaration early.
		case currentChar == byteSemicolon && topOfStack(closingBracketStack) != byteCloseParen:
			declaration := parseCSSDeclaration(buffer.String(), strings.Index(buffer.String(), ":"))
			if declaration == nil {
				// A stray `;` with nothing before it is not an error, it is empty input.
				if buffer.Len() == 0 {
					continue
				}
				return nil, &CSSSyntaxError{
					Message: fmt.Sprintf("Invalid declaration: `%s`", text.TrimWhitespace(buffer.String())),
					Offset:  bufferStart,
				}
			}

			if parent != nil {
				parent.Nodes = append(parent.Nodes, declaration)
			} else {
				ast = append(ast, declaration)
			}

			resetBuffer()

		// Start of a block. Whatever is in the buffer is the selector or the at-rule that owns it.
		//
		// Upstream builds a `rule` node here unconditionally and lets `parseAtRule` reshape it
		// later; the node is retagged below when the buffer starts with `@`, which is the same
		// result reached one step earlier.
		case currentChar == byteOpenCurly && topOfStack(closingBracketStack) != byteCloseParen:
			closingBracketStack = append(closingBracketStack, byteCloseCurly)

			trimmed := text.TrimWhitespace(buffer.String())
			if strings.HasPrefix(trimmed, "@") {
				node = ParseAtRule(trimmed)
			} else {
				node = StyleRule(trimmed)
			}

			if parent != nil {
				parent.Nodes = append(parent.Nodes, node)
			}

			// The current parent becomes the grandparent while this block is open.
			stack = append(stack, parent)
			parent = node

			resetBuffer()
			node = nil

		// End of a block.
		case currentChar == byteCloseCurly && topOfStack(closingBracketStack) != byteCloseParen:
			if len(closingBracketStack) == 0 {
				return nil, &CSSSyntaxError{Message: "Missing opening {", Offset: i}
			}
			closingBracketStack = closingBracketStack[:len(closingBracketStack)-1]

			// A non-empty buffer at a `}` is a final declaration or at-rule that was not terminated
			// with a `;`.
			if buffer.Len() > 0 {
				if bufferFirst == byteAtSign {
					// `@layer foo { @tailwind utilities }` — a body-less at-rule closing out a block.
					node = ParseAtRule(buffer.String())

					if parent != nil {
						parent.Nodes = append(parent.Nodes, node)
					} else {
						ast = append(ast, node)
					}

					resetBuffer()
					node = nil
				} else {
					// `.foo { color: red }` — a declaration with no trailing semicolon. Comments are
					// already stripped, so the first `:` is the real separator.
					colonIndex := strings.Index(buffer.String(), ":")

					// Upstream only attaches this when there is a parent; a trailing declaration at
					// the top level is dropped rather than being an error.
					if parent != nil {
						declaration := parseCSSDeclaration(buffer.String(), colonIndex)
						if declaration == nil {
							return nil, &CSSSyntaxError{
								Message: fmt.Sprintf("Invalid declaration: `%s`", text.TrimWhitespace(buffer.String())),
								Offset:  bufferStart,
							}
						}
						parent.Nodes = append(parent.Nodes, declaration)
					}
				}
			}

			// Pop one level. A nil grandparent means the block just closed was top-level, so it is
			// the point at which a completed tree joins the output.
			var grandParent *Node
			if len(stack) > 0 {
				grandParent = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}

			if grandParent == nil && parent != nil {
				ast = append(ast, parent)
			}

			parent = grandParent

			resetBuffer()
			node = nil

		case currentChar == byteOpenParen:
			closingBracketStack = append(closingBracketStack, byteCloseParen)
			if buffer.Len() == 0 {
				bufferStart = i
			}
			appendToBuffer("(")

		case currentChar == byteCloseParen:
			if topOfStack(closingBracketStack) != byteCloseParen {
				return nil, &CSSSyntaxError{Message: "Missing opening (", Offset: i}
			}
			closingBracketStack = closingBracketStack[:len(closingBracketStack)-1]
			appendToBuffer(")")

		// Any other byte is content.
		default:
			// Leading whitespace never starts a node, which is what makes bufferStart the offset of
			// the first meaningful byte.
			if buffer.Len() == 0 && (currentChar == byteSpace || currentChar == byteLineBreak || currentChar == byteTab) {
				continue
			}
			if buffer.Len() == 0 {
				bufferStart = i
			}
			appendToBuffer(input[i : i+1])
		}
	}

	// An at-rule that ended the input without a `;`.
	if bufferFirst == byteAtSign {
		ast = append(ast, ParseAtRule(buffer.String()))
	}

	// Everything should be balanced by now. A surviving parent is an unterminated block.
	if len(closingBracketStack) > 0 && parent != nil {
		switch parent.Kind {
		case KindRule:
			return nil, &CSSSyntaxError{
				Message: fmt.Sprintf("Missing closing } at %s", parent.Selector),
				Offset:  bufferStart,
			}
		case KindAtRule:
			return nil, &CSSSyntaxError{
				Message: fmt.Sprintf("Missing closing } at %s %s", parent.Name, parent.Params),
				Offset:  bufferStart,
			}
		}
	}

	if len(licenseComments) > 0 {
		return append(licenseComments, ast...), nil
	}

	return ast, nil
}

// isFollowedByWhitespace reports whether the byte after index is also whitespace, so that a run of
// whitespace collapses to its last byte.
//
// A CRLF counts as whitespace only as the complete pair, matching upstream's peek through the CR to
// the LF behind it. A lone CR does not, so it stays in the buffer as content.
func isFollowedByWhitespace(input string, index int, peekAt func(int) byte) bool {
	peekChar := peekAt(index + 1)
	if peekChar == byteSpace || peekChar == byteLineBreak || peekChar == byteTab {
		return true
	}
	return peekChar == byteCarriageReturn && peekAt(index+2) == byteLineBreak
}

// topOfStack returns the innermost expected closing bracket, or 0 when no block is open.
func topOfStack(stack []byte) byte {
	if len(stack) == 0 {
		return 0
	}
	return stack[len(stack)-1]
}

// ParseAtRule splits an at-rule buffer into its name and params.
//
// The scan starts at index 5 rather than at 1, which is upstream's documented assumption: the
// shortest at-rule in CSS is `@page`, so no name boundary can appear before then, and starting
// later skips five comparisons per at-rule. An at-rule shorter than `@page` therefore parses as
// all-name with empty params. That is upstream's behaviour and it is reproduced rather than
// improved, because a port that handled `@x` correctly would disagree with the engine it is
// checked against.
func ParseAtRule(buffer string, nodes ...*Node) *Node {
	name := buffer
	params := ""

	for i := 5; i < len(buffer); i++ {
		currentChar := buffer[i]
		if currentChar == byteSpace || currentChar == byteTab || currentChar == byteOpenParen {
			name = buffer[:i]
			params = buffer[i:]
			break
		}
	}

	return AtRule(text.TrimWhitespace(name), text.TrimWhitespace(params), nodes...)
}

// parseCSSDeclaration splits a buffer into a property and a value at colonIndex.
//
// A colonIndex of -1 means there is no separator and therefore no declaration; the caller turns
// that into the appropriate syntax error, since the two call sites word it differently.
//
// `!important` is searched for after the colon only, so a property named `--!important` is not
// mistaken for a flag, and the value is cut at it rather than around it — upstream slices to the
// marker and discards everything from there to the end, so trailing text after `!important` is
// dropped rather than kept.
func parseCSSDeclaration(buffer string, colonIndex int) *Node {
	if colonIndex == -1 || colonIndex > len(buffer) {
		return nil
	}

	importantIndex := strings.Index(buffer[colonIndex+1:], "!important")
	valueEnd := len(buffer)
	important := false
	if importantIndex != -1 {
		valueEnd = colonIndex + 1 + importantIndex
		important = true
	}

	declaration := Declaration(
		text.TrimWhitespace(buffer[:colonIndex]),
		text.TrimWhitespace(buffer[colonIndex+1:valueEnd]),
	)
	declaration.Important = important
	return declaration
}

// parseCSSString returns the index of the closing quote of the string starting at startIndex.
//
// A string must close on the same quote it opened with, so an apostrophe inside a double-quoted
// value is content. An unterminated string is an error rather than a value running to end of input,
// which is what stops a missing quote from swallowing the rest of the stylesheet.
func parseCSSString(input string, startIndex int, quoteChar byte) (int, error) {
	for i := startIndex + 1; i < len(input); i++ {
		peekChar := input[i]

		switch {
		case peekChar == byteBackslash:
			i++

		case peekChar == quoteChar:
			return i, nil

		// A `;` at the end of the line inside an unterminated string. Reported separately from the
		// bare line break below because upstream includes the `;` in the message.
		case peekChar == byteSemicolon &&
			(peekByte(input, i+1) == byteLineBreak ||
				(peekByte(input, i+1) == byteCarriageReturn && peekByte(input, i+2) == byteLineBreak)):
			return 0, &CSSSyntaxError{
				Message: fmt.Sprintf("Unterminated string: %s%c", input[startIndex:i+1], quoteChar),
				Offset:  startIndex,
			}

		case peekChar == byteLineBreak ||
			(peekChar == byteCarriageReturn && peekByte(input, i+1) == byteLineBreak):
			return 0, &CSSSyntaxError{
				Message: fmt.Sprintf("Unterminated string: %s%c", input[startIndex:i], quoteChar),
				Offset:  startIndex,
			}
		}
	}

	// Upstream returns the start index when the input runs out, which leaves the caller treating
	// the opening quote as a one-byte token rather than failing.
	return startIndex, nil
}

// peekByte returns the byte at index, or 0 past either end.
func peekByte(input string, index int) byte {
	if index < 0 || index >= len(input) {
		return 0
	}
	return input[index]
}
