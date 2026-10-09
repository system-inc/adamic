// Package values ports postcss-values-parser 2.0.1 (lib/parser.js, lib/tokenize.js and the node
// classes), the parser Prettier's language-css hands declaration values and some at-rule params to.
//
// Prettier constructs it in src/language-css/parse/parse-value.js as
// `new PostcssValuesParser(value, { loose: true }).parse()`, importing lib/parser.js directly, so the
// library's index.js is not part of the port. Options mirrors that options object.
//
// Nodes are *estree.Node with the JavaScript objects' own fields: raws, value, source, sourceIndex,
// nodes, and each class's extras (unit, quoted, inline, isHex, isColor, parenType, unbalanced).
// sourceIndex is a byte offset into the value; source line and column stay in the library's units,
// columns counting UTF-16 code units. Ranges are [0, 0]; the glue computes Prettier's positions.
package values

import (
	"fmt"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// Options is the parser's options object. Its default is { loose: false }; Prettier passes
// { loose: true }.
type Options struct {
	// Loose relaxes the operator checks, keeps url() arguments as tokens, and reads `//` as an inline
	// comment.
	Loose bool
}

// Error is a throw from the library: a ParserError, a TokenizeError, or the TypeError upstream hits
// reading a property of undefined. Error() is JavaScript's String(error), "Name: message".
type Error struct {
	Name    string
	Message string
}

func (thrown *Error) Error() string { return thrown.Name + ": " + thrown.Message }

// Parse is `new Parser(value, options).parse()`: the Root, holding one Value, holding the nodes.
// Everything the library throws comes back as an *Error marked printing.Syntax, which Prettier turns
// into a value-unknown node.
func Parse(value string, options Options) (result *estree.Node, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if thrown, isThrown := recovered.(*Error); isThrown {
				result, err = nil, printing.Syntax(thrown)
				return
			}
			result, err = nil, fmt.Errorf("css values: parsing %q: %v", value, recovered)
		}
	}()
	return newParser(value, options).parse(), nil
}

// utf16Units is the text as JavaScript holds it, UTF-16 code units, with the byte offset of every unit
// and of the end. The second unit of a surrogate pair maps two bytes into its four-byte character, so
// the mapping stays one to one. A byte that is not UTF-8 becomes U+FFFD, one unit, as decoding the file
// would make it.
func utf16Units(text string) ([]uint16, []int) {
	units := make([]uint16, 0, len(text))
	offsets := make([]int, 0, len(text)+1)
	for position, character := range text {
		if character == utf8.RuneError {
			if _, size := utf8.DecodeRuneInString(text[position:]); size == 1 {
				units = append(units, 0xFFFD)
				offsets = append(offsets, position)
				continue
			}
		}
		if character >= 0x10000 {
			high, low := utf16.EncodeRune(character)
			units = append(units, uint16(high), uint16(low))
			offsets = append(offsets, position, position+2)
			continue
		}
		units = append(units, uint16(character))
		offsets = append(offsets, position)
	}
	offsets = append(offsets, len(text))
	return units, offsets
}

func itoa(number int) string { return strconv.Itoa(number) }
