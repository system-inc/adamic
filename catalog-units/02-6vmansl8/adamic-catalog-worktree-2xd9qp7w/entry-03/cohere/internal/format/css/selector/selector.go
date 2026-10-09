// Package selector is postcss-selector-parser 2.2.3, the selector parser Prettier's CSS printer reads:
// dist/processor.js, dist/parser.js, dist/tokenize.js and dist/selectors/*.js, ported to Go.
//
// Prettier calls it as `new PostcssSelectorParser((selectors) => { result = selectors }).process(selector)`
// (src/language-css/parse/parse-selector.js); Parse returns that `selectors` root.
package selector

import (
	"errors"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// ErrLoopsForever is returned, marked printing.Syntax, for a selector upstream never returns from: a
// namespace bar followed by something other than a word or `*` (`a| b`, `:is(a|)`),
// where parser.js's namespace() returns without consuming the bar and every caller comes straight
// back to it. Prettier would hang on such a selector; the port refuses it instead.
var ErrLoopsForever = errors.New("postcss-selector-parser 2.2.3 never returns on this selector: a namespace bar it does not consume")

// Parse is processor.js's process(selector) with Prettier's callback: the root the parser builds, or
// the error it throws, marked printing.Syntax. Upstream's own errors keep their messages ("Unclosed
// quote", "Expected a closing square bracket.", ...), and the TypeErrors it throws by accident keep
// V8's ("Cannot read properties of undefined (reading '0')").
//
// Nodes keep every own field upstream sets, with the same names and values. sourceIndex is a byte
// offset into selector, converted from upstream's UTF-16 index; source's lines and columns stay as
// upstream computes them, columns in UTF-16 units. Every node's Range is [0, 0]: upstream records a
// start (sourceIndex) and no end, and Prettier's loc.js computes positions from those fields.
func Parse(selector string) (root *estree.Node, err error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		root = nil
		switch thrown := recovered.(type) {
		case thrownError:
			err = printing.Syntax(errors.New(thrown.message))
		case loopsForever:
			err = printing.Syntax(ErrLoopsForever)
		default:
			err = fmt.Errorf("css selector parser: %v", recovered)
		}
	}()

	// new Parser({css: selectors, error, options: {}}), whose constructor returns this.loop().
	css := utf16.Encode([]rune(selector))
	root = newParser(css).loop()
	wellFormedTree(root)
	convertSourceIndexes(root, utf16ToByteOffsets(selector))
	return root, nil
}

// utf16ToByteOffsets maps each UTF-16 index of text (and its end) to a byte offset. An index inside a
// surrogate pair maps to the start of its character. Text that is not valid UTF-8 decodes one U+FFFD,
// one unit, per bad byte, as []rune does for the parser.
func utf16ToByteOffsets(text string) []int {
	offsets := make([]int, 0, len(text)+1)
	for position := 0; position < len(text); {
		character, size := utf8.DecodeRuneInString(text[position:])
		for range utf16.RuneLen(character) {
			offsets = append(offsets, position)
		}
		position += size
	}
	return append(offsets, len(text))
}

// byteOffset converts one UTF-16 index. splitWord can compute a sourceIndex past the end of the text
// (it adds an index into a word merged across escapes to the last token's position), and such an
// index stays past the end by the same distance.
func byteOffset(offsets []int, index int) int {
	if index < 0 {
		return index
	}
	if index < len(offsets) {
		return offsets[index]
	}
	return offsets[len(offsets)-1] + index - (len(offsets) - 1)
}

func convertSourceIndexes(node *estree.Node, offsets []int) {
	if index, isInt := node.Get("sourceIndex").(int); isInt {
		node.Set("sourceIndex", byteOffset(offsets, index))
	}
	for _, child := range node.List("nodes") {
		convertSourceIndexes(child, offsets)
	}
}
