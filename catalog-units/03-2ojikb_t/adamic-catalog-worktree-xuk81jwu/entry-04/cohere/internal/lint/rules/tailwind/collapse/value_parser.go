// The CSS value parser: a small recursive-descent pass that turns a value string into a node tree
// of words, functions, and separators.
//
// Ported from `src/value-parser.ts` at the v4.3.3 tag, read at the tag rather than from the
// minified bundle, so a disagreement with the engine is a finding rather than version skew.
//
// # What this is for
//
// `segment.go` already carries a value splitter, and the two are not redundant. `segment` answers
// one question — where are the top-level separators — and answers it over a flat string, which is
// all `inferDataType` needs. This answers the structural question: what is the tree. The `@utility`
// evaluator needs the tree to read a `--value(--percentage-*, [*])` argument list, where the
// arguments are nodes rather than substrings, and callers downstream depend on the node shapes
// rather than on the top-level split.
//
// # Whitespace is preserved, exactly
//
// A run of separator characters becomes one separator node holding the run verbatim, so
// `foo(bar,  baz)` keeps its two spaces and reprints identically. This matters more than it looks:
// the parser is used on values that must round-trip, and a parser that normalised `, ` to `,` would
// agree with the engine on structure and disagree on every printed result.
//
// # Deliberate fidelity to JavaScript semantics
//
// Three behaviors here are consequences of how the upstream loop is written rather than of stated
// intent. They are ported deliberately, are pinned by fixtures measured against the shipped engine,
// and each is called out at its branch below:
//
//   - A pending buffer in front of an unmatched `)` is discarded, because upstream flushes it with
//     `tail?.nodes.push(...)` and `tail` is undefined when the stack is empty. `a)b` parses to a
//     single word `b`, and `foo\(bar)` parses to nothing at all.
//   - A trailing `\` yields the literal text `\undefined`, because upstream concatenates
//     `input[i + 1]`, and indexing one past the end of a JavaScript string is `undefined`, which
//     stringifies rather than throwing.
//   - The final buffer flush always lands at the top level, never inside a function still left
//     open, so `foo(bar` puts `bar` beside the function rather than in it.
//
// A port that "fixed" any of these would be wrong in the direction of a plausible tree, which is
// the failure mode that is expensive to notice.
package tailwind

import "strings"

// ValueNodeKind distinguishes the three node shapes the parser produces.
type ValueNodeKind string

const (
	// ValueNodeKindWord is a run of ordinary characters, and also a lone `/`, which upstream
	// deliberately emits as its own word rather than as a separator.
	ValueNodeKindWord ValueNodeKind = "word"
	// ValueNodeKindFunction is a call. Value holds the name, which is empty for a bare `(`.
	ValueNodeKindFunction ValueNodeKind = "function"
	// ValueNodeKindSeparator is a maximal run of separator characters, held verbatim.
	ValueNodeKindSeparator ValueNodeKind = "separator"
)

// ValueNode is one node of the parsed value tree.
//
// One struct rather than an interface with three implementations, matching the upstream tagged
// union. Callers switch on Kind, and Nodes is meaningful only for a function. The flat shape also
// keeps the JSON encoding identical to the engine's, which is what lets the fixture be compared
// structurally without a translation layer in the test.
type ValueNode struct {
	Kind ValueNodeKind `json:"kind"`
	// Value is the word text, the separator run, or the function name.
	Value string `json:"value"`
	// Nodes are the arguments of a function, and nil for every other kind. A function with no
	// arguments carries an empty non-nil slice, matching upstream's `[]`, because the difference
	// is visible in the encoded fixture.
	Nodes []ValueNode `json:"nodes,omitempty"`
}

// The separator characters, as the upstream constant set. `/` is not among them: it gets its own
// branch and becomes a word.
const (
	valueBackslash   = '\\'
	valueCloseParen  = ')'
	valueColon       = ':'
	valueComma       = ','
	valueDoubleQuote = '"'
	valueEquals      = '='
	valueGreaterThan = '>'
	valueLessThan    = '<'
	valueNewline     = '\n'
	valueOpenParen   = '('
	valueSingleQuote = '\''
	valueSlash       = '/'
	valueSpace       = ' '
	valueTab         = '\t'
)

// isValueSeparator reports whether the byte joins a separator run.
//
// Byte-wise rather than rune-wise, which is safe because every member is ASCII and UTF-8
// continuation bytes are all >= 0x80, so no multi-byte character can be mistaken for one.
func isValueSeparator(character byte) bool {
	switch character {
	case valueColon, valueComma, valueEquals, valueGreaterThan,
		valueLessThan, valueNewline, valueSpace, valueTab:
		return true
	}
	return false
}

// ParseValue parses a CSS value string into a node tree.
//
// Total: every input parses, including unbalanced and truncated ones, because real class strings
// contain arbitrary text and upstream never signals an error. What malformed input produces is
// documented at each branch and pinned by fixtures rather than left to inference.
func ParseValue(input string) []ValueNode {
	// Upstream normalises CRLF before doing anything else, so a `\r\n` in the input becomes a
	// one-character separator rather than a two-character one. A lone `\r` is untouched and falls
	// through to the default branch as an ordinary character, which is why this is a replace
	// rather than a general newline normalisation.
	input = strings.ReplaceAll(input, "\r\n", "\n")

	ast := []ValueNode{}

	// The stack holds the path of open functions. Nodes are referenced by index rather than by
	// pointer: appending to a parent's Nodes slice can reallocate it, so a pointer captured
	// earlier would write into a stale backing array. Each frame records where its node lives so
	// the append is always applied through the current slice header.
	type openFunction struct {
		// container is the frame's parent index in the stack, or -1 when the node lives in ast.
		container int
		// index is the position of this function's node within its container's slice.
		index int
	}
	var stack []openFunction

	// nodesFor returns the slice a frame's children live in, so a push reads and writes the live
	// header rather than a copy.
	var nodesFor func(frame int) []ValueNode
	nodesFor = func(frame int) []ValueNode {
		if frame < 0 {
			return ast
		}
		parent := nodesFor(stack[frame].container)
		return parent[stack[frame].index].Nodes
	}
	var setNodes func(frame int, nodes []ValueNode)
	setNodes = func(frame int, nodes []ValueNode) {
		if frame < 0 {
			ast = nodes
			return
		}
		parent := nodesFor(stack[frame].container)
		parent[stack[frame].index].Nodes = nodes
	}

	// push appends a node to whichever list is currently open.
	push := func(node ValueNode) int {
		frame := len(stack) - 1
		nodes := append(nodesFor(frame), node)
		setNodes(frame, nodes)
		return len(nodes) - 1
	}

	buffer := strings.Builder{}

	// flushBuffer emits the pending buffer as a word into the currently open list, if non-empty.
	flushBuffer := func() {
		if buffer.Len() == 0 {
			return
		}
		push(ValueNode{Kind: ValueNodeKindWord, Value: buffer.String()})
		buffer.Reset()
	}

	for index := 0; index < len(input); index++ {
		character := input[index]

		switch {
		// A `\` escapes the next character: both are consumed into the buffer, and neither is
		// examined. This is what keeps `a\(b` from opening a function and `a\ b` from splitting.
		case character == valueBackslash:
			if index+1 < len(input) {
				// One byte, not one rune, and the two are provably indistinguishable here.
				// Upstream indexes by UTF-16 code unit, so for a character outside the BMP it
				// consumes only the high surrogate and lets the low surrogate fall to the default
				// branch, which appends it unchanged. The same holds byte-wise: every UTF-8
				// continuation byte is >= 0x80, so it can match no separator, quote, or paren and
				// reaches the default branch, which appends it. Either way the escaped character
				// arrives in the buffer whole. Verified by parsing every string of length 1 to 3
				// over an alphabet of backslash, quotes, parens, separators and multi-byte
				// characters against the engine: 2,379 inputs, no disagreement under either rule.
				buffer.WriteString(input[index : index+2])
				index++
			} else {
				// A trailing backslash reads one past the end. In JavaScript that is `undefined`,
				// and `'\\' + undefined` is the seven-character string `\undefined`. Reproduced
				// literally because the fixture measures it, absurd as it reads.
				buffer.WriteString(`\undefined`)
			}

		// A `/` is its own word rather than a separator. Upstream keeps it that way for
		// `theme(colors.red.500/10)`, where the slash carries meaning and has no spaces around it.
		case character == valueSlash:
			flushBuffer()
			push(ValueNode{Kind: ValueNodeKindWord, Value: string(input[index])})

		// A run of separator characters becomes one separator node holding the run verbatim.
		case isValueSeparator(character):
			flushBuffer()
			start := index
			end := index + 1
			for ; end < len(input); end++ {
				if !isValueSeparator(input[end]) {
					break
				}
			}
			index = end - 1
			push(ValueNode{Kind: ValueNodeKindSeparator, Value: input[start:end]})

		// A quoted string is taken whole into the buffer, closing only on a matching quote, so a
		// `'` inside a `"` string does not end it and a font name keeps its commas.
		case character == valueSingleQuote || character == valueDoubleQuote:
			start := index
			for scan := index + 1; scan < len(input); scan++ {
				if input[scan] == valueBackslash {
					scan++
					continue
				}
				if input[scan] == character {
					index = scan
					break
				}
			}
			// An unterminated string leaves `i` untouched upstream, so only the quote character
			// itself is consumed and parsing resumes at the next character. `"a b` is therefore
			// three nodes, not one: the word `"a`, a separator, and the word `b`. Advancing to the
			// end of input instead would swallow the tail, which agrees with the engine on
			// `"unterminated` and disagrees on every unterminated string that contains a
			// separator.
			buffer.WriteString(input[start : index+1])

		// A `(` opens a function whose name is whatever the buffer held, which is empty for a bare
		// `(` and for a `(` right after a separator.
		case character == valueOpenParen:
			name := buffer.String()
			buffer.Reset()
			container := len(stack) - 1
			position := push(ValueNode{Kind: ValueNodeKindFunction, Value: name, Nodes: []ValueNode{}})
			stack = append(stack, openFunction{container: container, index: position})

		// A `)` closes the innermost function.
		case character == valueCloseParen:
			if len(stack) > 0 {
				// The pending buffer belongs to the function being closed, so it is flushed
				// before the frame is popped.
				flushBuffer()
				stack = stack[:len(stack)-1]
			} else {
				// Nothing is open. Upstream pops undefined and flushes with `tail?.nodes.push`,
				// so the optional chain silently drops the word. `a)b` really does parse to just
				// `b`, and `foo\(bar)` to nothing. Discarding rather than emitting is the whole
				// behavior of this branch.
				buffer.Reset()
			}

		default:
			buffer.WriteByte(character)
		}
	}

	// The remainder is a word at the top level, even when a function is still open, which is why
	// `foo(bar` yields an empty function beside the word `bar` rather than containing it.
	if buffer.Len() > 0 {
		ast = append(ast, ValueNode{Kind: ValueNodeKindWord, Value: buffer.String()})
	}

	return ast
}

// ValueToCss reprints a node tree, the inverse of ParseValue for every input that round-trips.
//
// Not every input does, and the exceptions are the malformed ones described at the top of this
// file: text discarded in front of an unmatched `)` cannot come back, and an unclosed function
// gains the `)` it never had.
func ValueToCss(nodes []ValueNode) string {
	var builder strings.Builder
	writeValueCss(&builder, nodes)
	return builder.String()
}

func writeValueCss(builder *strings.Builder, nodes []ValueNode) {
	for _, node := range nodes {
		switch node.Kind {
		case ValueNodeKindWord, ValueNodeKindSeparator:
			builder.WriteString(node.Value)
		case ValueNodeKindFunction:
			builder.WriteString(node.Value)
			builder.WriteByte('(')
			writeValueCss(builder, node.Nodes)
			builder.WriteByte(')')
		}
	}
}
