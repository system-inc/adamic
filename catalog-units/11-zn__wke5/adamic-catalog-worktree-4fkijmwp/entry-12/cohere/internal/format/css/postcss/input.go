// Package postcss is postcss 8.5.16's CSS parser ported to Go: lib/tokenize.js, lib/parser.js and
// lib/parse.js, with the parts of lib/input.js, lib/css-syntax-error.js and the node classes they need.
//
// Prettier's css parser (src/language-css/parser-postcss.js, parseCss) calls
// postcssParse.default(text, { map: false }) on the text left after front matter is taken off, and
// everything Prettier then does to the tree (type prefixes, parseNestedCSS, calculateLoc) is the glue's,
// in internal/format/css. Parse here is only postcss.
//
// ParseSCSS is postcss-scss 4.0.9's parser over the same machinery (scss_parse.go, scss_parser.go,
// scss_tokenize.go, nested_declaration.go), as Prettier's parseScss calls it.
//
// # Units
//
// postcss walks the text in UTF-16 code units, and the port does too: the tokenizer and parser run over
// the text's UTF-16 units so every charCodeAt, indexOf and slice decides exactly as upstream's does.
// Offsets cross to bytes only where the tree is handed out (parse.go): a node's source.start.offset and
// source.end.offset, and a CssSyntaxError's offsets, are bytes into the text postcss parsed (the text
// after a leading byte order mark, which Input drops). Line and column stay postcss's: line counts "\n"
// only, and column is in UTF-16 units, one-based. The glue converts where Prettier computes an offset
// from a line and column.
package postcss

import (
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

// lib/input.js: the Input class, the part of it the parser reaches (css, hasBOM, fromOffset,
// fromLineAndColumn, error).
//
// What is not here: from, file, id and the source map. Prettier passes { map: false } and no `from`, so
// PreviousMap returns before reading anything, Input.origin always answers false, and the error's file
// is unset. The id is a random nanoid nothing reads.

// input is the text as postcss holds it: css with its byte order mark dropped, as UTF-16 units, and the
// byte offset of every unit so slices and offsets can cross back to the Go string.
type input struct {
	// css is this.css, the text after the byte order mark.
	css string
	// hasBOM is this.hasBOM.
	hasBOM bool

	// units are css's UTF-16 code units, what charCodeAt reads. A byte that is not UTF-8 is one unit,
	// U+FFFD, as Node decodes it.
	units []uint16
	// byteOffsets maps a unit index (and the end, len(units)) to its byte offset in css. The second
	// unit of a surrogate pair maps to the byte after the character: a JavaScript slice that cuts a pair
	// there keeps the whole character on its left in Go, and the pieces still join back to the same text.
	byteOffsets []int

	// lineToIndex is getLineToIndex's cache: the unit offset where each "\n"-separated line starts.
	lineToIndex []int
}

func newInput(css string) *input {
	in := &input{css: css}

	// if (this.css[0] === '\uFEFF' || this.css[0] === '\uFFFE'): U+FEFF is EF BB BF in UTF-8, U+FFFE EF BF BE.
	if len(in.css) >= 3 && in.css[0] == 0xEF && ((in.css[1] == 0xBB && in.css[2] == 0xBF) || (in.css[1] == 0xBF && in.css[2] == 0xBE)) {
		in.hasBOM = true
		in.css = in.css[3:]
	} else {
		in.hasBOM = false
	}

	in.units = make([]uint16, 0, len(in.css))
	in.byteOffsets = make([]int, 0, len(in.css)+1)
	for position := 0; position < len(in.css); {
		character, size := utf8.DecodeRuneInString(in.css[position:])
		if character > 0xFFFF {
			high, low := utf16.EncodeRune(character)
			in.units = append(in.units, uint16(high), uint16(low))
			in.byteOffsets = append(in.byteOffsets, position, position+size)
		} else {
			// A byte that is not UTF-8 decodes as RuneError with size 1, which is U+FFFD.
			in.units = append(in.units, uint16(character))
			in.byteOffsets = append(in.byteOffsets, position)
		}
		position += size
	}
	in.byteOffsets = append(in.byteOffsets, len(in.css))

	// getLineToIndex: lines split on "\n", each start one past the previous line's "\n".
	in.lineToIndex = []int{0}
	for index, unit := range in.units {
		if unit == '\n' {
			in.lineToIndex = append(in.lineToIndex, index+1)
		}
	}

	return in
}

// slice is css.slice(start, end) on unit offsets, clamped and emptied the way String.prototype.slice
// treats a reversed or out-of-range pair.
func (in *input) slice(start int, end int) string {
	length := len(in.units)
	start = min(max(start, 0), length)
	end = min(max(end, 0), length)
	if end <= start {
		return ""
	}
	return in.css[in.byteOffsets[start]:in.byteOffsets[end]]
}

// byteOffset converts a unit offset to bytes. postcss can report an offset past the end (freeSemicolon
// adds the length of the spaces before a semicolon to the semicolon's own offset), so an offset beyond
// the text keeps its distance past the end, the spaces it counts being one byte each.
func (in *input) byteOffset(offset int) int {
	if offset < 0 {
		return offset
	}
	if offset > len(in.units) {
		return len(in.css) + offset - len(in.units)
	}
	return in.byteOffsets[offset]
}

// lineAndColumn is the result of fromOffset: { col, line }.
type lineAndColumn struct {
	column int
	line   int
}

// fromOffset finds the line holding a unit offset by binary search over lineToIndex, as upstream does,
// with an offset past the last line's start landing on the last line.
func (in *input) fromOffset(offset int) lineAndColumn {
	lineToIndex := in.lineToIndex
	lastLine := lineToIndex[len(lineToIndex)-1]

	low := 0
	if offset >= lastLine {
		low = len(lineToIndex) - 1
	} else {
		high := len(lineToIndex) - 2
		var middle int
		for low < high {
			middle = low + ((high - low) >> 1)
			if offset < lineToIndex[middle] {
				high = middle - 1
			} else if offset >= lineToIndex[middle+1] {
				low = middle + 1
			} else {
				low = middle
				break
			}
		}
	}
	return lineAndColumn{
		column: offset - lineToIndex[low] + 1,
		line:   low + 1,
	}
}

// fromLineAndColumn is the unit offset of a one-based line and column.
func (in *input) fromLineAndColumn(line int, column int) int {
	index := 0
	if line-1 >= 0 && line-1 < len(in.lineToIndex) {
		index = in.lineToIndex[line-1]
	}
	return index + column - 1
}

// errorAt is input.error(message, offset): a position given as one unit offset.
func (in *input) errorAt(message string, offset int) *CssSyntaxError {
	position := in.fromOffset(offset)
	return &CssSyntaxError{
		Reason: message,
		Line:   position.line,
		Column: position.column,
		Offset: in.byteOffset(offset),
	}
}

// errorBetween is input.error(message, { offset: start }, { offset: end }): a range of unit offsets.
func (in *input) errorBetween(message string, start int, end int) *CssSyntaxError {
	startPosition := in.fromOffset(start)
	endPosition := in.fromOffset(end)
	return &CssSyntaxError{
		Reason:    message,
		Line:      startPosition.line,
		Column:    startPosition.column,
		EndLine:   endPosition.line,
		EndColumn: endPosition.column,
		Offset:    in.byteOffset(start),
		EndOffset: in.byteOffset(end),
		HasEnd:    true,
	}
}

// errorAtLineAndColumn is input.error(message, line, column).
func (in *input) errorAtLineAndColumn(message string, line int, column int) *CssSyntaxError {
	return &CssSyntaxError{
		Reason: message,
		Line:   line,
		Column: column,
		Offset: in.byteOffset(in.fromLineAndColumn(line, column)),
	}
}

// CssSyntaxError is lib/css-syntax-error.js's error as the parser throws it, with the input fields
// Input.error attaches (offset and endOffset, here in bytes).
//
// Prettier reads name, reason, line and column from it to build its own error
// (`${name}: ${reason}` at { line, column }), so those are kept exactly; Column and EndColumn are in
// UTF-16 units, as postcss counts them.
type CssSyntaxError struct {
	// Reason is error.reason, the message without the position.
	Reason string
	// Line and Column are where the problem starts, one-based.
	Line   int
	Column int
	// EndLine and EndColumn are where it ends, when HasEnd.
	EndLine   int
	EndColumn int
	// Offset and EndOffset are error.input.offset and error.input.endOffset, in bytes.
	Offset    int
	EndOffset int
	HasEnd    bool
}

// Name is error.name.
func (*CssSyntaxError) Name() string { return "CssSyntaxError" }

// Error is setMessage's message: no plugin, no file, so "<css input>:line:column: reason".
func (syntaxError *CssSyntaxError) Error() string {
	return "<css input>:" + strconv.Itoa(syntaxError.Line) + ":" + strconv.Itoa(syntaxError.Column) + ": " + syntaxError.Reason
}
