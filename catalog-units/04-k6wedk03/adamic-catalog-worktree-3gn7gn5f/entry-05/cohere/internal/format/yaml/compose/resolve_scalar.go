package compose

// Ported from eemeli/yaml 2.9.0, dist/compose/resolve-flow-scalar.js and resolve-block-scalar.js. Values
// are built as JavaScript strings, []uint16, so escapes that make lone surrogates and every length
// upstream measures come out the same; composeScalar converts the result once.

import "github.com/system-inc/cohere/internal/format/yaml/cst"

func resolveFlowScalar(scalar *cst.Token, strict bool, onError onErrorFunc) scalarResolution {
	offset, tokenType, source, end := scalar.Offset, scalar.Type, scalar.Source, scalar.End
	var scalarType string
	var value []uint16
	relativeError := func(rel int, code string, msg string) { onError(offset+rel, code, msg, false) }
	switch tokenType {
	case "scalar":
		scalarType = Plain
		value = plainValue(source, relativeError)
	case "single-quoted-scalar":
		scalarType = QuoteSingle
		value = singleQuotedValue(source, relativeError)
	case "double-quoted-scalar":
		scalarType = QuoteDouble
		value = doubleQuotedValue(source, relativeError)
	default:
		onError(scalar, "UNEXPECTED_TOKEN", "Expected a flow scalar value, but found: "+tokenType, false)
		return scalarResolution{
			value:      []uint16{},
			valueRange: []int{offset, offset + len(source), offset + len(source)},
		}
	}
	valueEnd := offset + len(source)
	re := resolveEnd(end, valueEnd, strict, onError)
	return scalarResolution{
		value:      value,
		scalarType: scalarType,
		comment:    re.comment,
		valueRange: []int{offset, valueEnd, re.offset},
	}
}

func plainValue(source []uint16, onError func(rel int, code string, msg string)) []uint16 {
	badChar := ""
	switch characterAt(source, 0) {
	case '\t':
		badChar = "a tab character"
	case ',':
		badChar = "flow indicator character ,"
	case '%':
		badChar = "directive indicator character %"
	case '|', '>':
		badChar = "block scalar indicator " + string(rune(source[0]))
	case '@', '`':
		badChar = "reserved character " + string(rune(source[0]))
	}
	if badChar != "" {
		onError(0, "BAD_SCALAR_START", "Plain value cannot start with "+badChar)
	}
	return foldLines(source)
}

func singleQuotedValue(source []uint16, onError func(rel int, code string, msg string)) []uint16 {
	if characterAt(source, len(source)-1) != '\'' || len(source) == 1 {
		onError(len(source), "MISSING_CHAR", "Missing closing 'quote")
	}
	// `.replace(/''/g, "'")`
	folded := foldLines(slice(source, 1, -1))
	result := make([]uint16, 0, len(folded))
	for i := 0; i < len(folded); i++ {
		result = append(result, folded[i])
		if folded[i] == '\'' && i+1 < len(folded) && folded[i+1] == '\'' {
			i++
		}
	}
	return result
}

func isSpaceOrTab(character int) bool { return character == ' ' || character == '\t' }

// foldLinesEnd is where a line's content ends before the newline at n: before an \r right at the
// newline, and before the spaces and tabs before that, but not before from.
func foldLinesEnd(source []uint16, from int, n int) int {
	end := n
	if end-1 >= from && source[end-1] == '\r' {
		end--
	}
	for end-1 >= from && isSpaceOrTab(int(source[end-1])) {
		end--
	}
	return end
}

// foldLines folds a flow scalar's lines: each line break between two lines of content becomes a space,
// and each empty line a newline, with the whitespace around each break trimmed. Upstream does this with
// three sticky regular expressions, two with lookbehind, which Go's regexp cannot run:
//
//	first = /(.*?)(?<![ \t])[ \t]*\r?\n/sy
//	line  = /[ \t]*(.*?)(?:(?<![ \t])[ \t]*)?\r?\n/sy
//	last  = /[ \t]*(.*)/sy
//
// The lazy group of first and line can only end at the first newline the match reaches, so each
// capture is the text up to that newline less an \r right before it and the spaces and tabs before
// that (foldLinesEnd), and line first skips the line's leading spaces and tabs.
func foldLines(source []uint16) []uint16 {
	n := indexOfUnit(source, '\n')
	if n == -1 {
		return source
	}
	res := append([]uint16{}, source[:foldLinesEnd(source, 0, n)]...)
	sep := []uint16{' '}
	pos := n + 1
	for {
		start := pos
		for start < len(source) && isSpaceOrTab(int(source[start])) {
			start++
		}
		newline := -1
		for index := start; index < len(source); index++ {
			if source[index] == '\n' {
				newline = index
				break
			}
		}
		if newline == -1 {
			break
		}
		match := source[start:foldLinesEnd(source, start, newline)]
		if len(match) == 0 {
			if equalsASCII(sep, "\n") {
				res = append(res, sep...)
			} else {
				sep = []uint16{'\n'}
			}
		} else {
			res = append(append(res, sep...), match...)
			sep = []uint16{' '}
		}
		pos = newline + 1
	}
	start := pos
	for start < len(source) && isSpaceOrTab(int(source[start])) {
		start++
	}
	return concatUnits(res, sep, source[start:])
}

func doubleQuotedValue(source []uint16, onError func(rel int, code string, msg string)) []uint16 {
	res := []uint16{}
	for i := 1; i < len(source)-1; i++ {
		ch := source[i]
		if ch == '\r' && characterAt(source, i+1) == '\n' {
			continue
		}
		if ch == '\n' {
			fold, offset := foldNewline(source, i)
			res = append(res, fold...)
			i = offset
		} else if ch == '\\' {
			i++
			next := characterAt(source, i)
			if cc, found := escapeCodes[next]; found {
				res = append(res, cc)
			} else if next == '\n' {
				// skip escaped newlines, but still trim the following line
				next = characterAt(source, i+1)
				for next == ' ' || next == '\t' {
					i++
					next = characterAt(source, i+1)
				}
			} else if next == '\r' && characterAt(source, i+1) == '\n' {
				// skip escaped CRLF newlines, but still trim the following line
				i++
				next = characterAt(source, i+1)
				for next == ' ' || next == '\t' {
					i++
					next = characterAt(source, i+1)
				}
			} else if next == 'x' || next == 'u' || next == 'U' {
				length := 8
				if next == 'x' {
					length = 2
				} else if next == 'u' {
					length = 4
				}
				res = append(res, parseCharCode(source, i+1, length, onError)...)
				i += length
			} else {
				raw := substr(source, i-1, 2)
				onError(i-1, "BAD_DQ_ESCAPE", "Invalid escape sequence "+unitsToString(raw))
				res = append(res, raw...)
			}
		} else if ch == ' ' || ch == '\t' {
			// trim trailing whitespace
			wsStart := i
			next := characterAt(source, i+1)
			for next == ' ' || next == '\t' {
				i++
				next = characterAt(source, i+1)
			}
			if next != '\n' && !(next == '\r' && characterAt(source, i+2) == '\n') {
				if i > wsStart {
					res = append(res, slice(source, wsStart, i+1)...)
				} else {
					res = append(res, ch)
				}
			}
		} else {
			res = append(res, ch)
		}
	}
	if characterAt(source, len(source)-1) != '"' || len(source) == 1 {
		onError(len(source), "MISSING_CHAR", `Missing closing "quote`)
	}
	return res
}

// foldNewline folds a single newline into a space, multiple newlines to N - 1 newlines. Presumes
// `source[offset] === '\n'`.
func foldNewline(source []uint16, offset int) ([]uint16, int) {
	fold := []uint16{}
	ch := characterAt(source, offset+1)
	for ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
		if ch == '\r' && characterAt(source, offset+2) != '\n' {
			break
		}
		if ch == '\n' {
			fold = append(fold, '\n')
		}
		offset += 1
		ch = characterAt(source, offset+1)
	}
	if len(fold) == 0 {
		fold = []uint16{' '}
	}
	return fold, offset
}

var escapeCodes = map[int]uint16{
	'0':  0x00,   // null character
	'a':  0x07,   // bell character
	'b':  '\b',   // backspace
	'e':  0x1b,   // escape character
	'f':  '\f',   // form feed
	'n':  '\n',   // line feed
	'r':  '\r',   // carriage return
	't':  '\t',   // horizontal tab
	'v':  '\v',   // vertical tab
	'N':  0x85,   // Unicode next line
	'_':  0xA0,   // Unicode non-breaking space
	'L':  0x2028, // Unicode line separator
	'P':  0x2029, // Unicode paragraph separator
	' ':  ' ',
	'"':  '"',
	'/':  '/',
	'\\': '\\',
	'\t': '\t',
}

func parseCharCode(source []uint16, offset int, length int, onError func(rel int, code string, msg string)) []uint16 {
	cc := substr(source, offset, length)
	ok := len(cc) == length
	for _, unit := range cc {
		if _, isHex := hexDigit(unit); !isHex {
			ok = false
		}
	}
	// `String.fromCodePoint(code)` throws a RangeError for NaN and anything past U+10FFFF.
	if ok {
		code := int(parseInt(cc, 16))
		if code <= 0x10FFFF {
			if code < 0x10000 {
				return []uint16{uint16(code)}
			}
			code -= 0x10000
			return []uint16{uint16(0xD800 + code>>10), uint16(0xDC00 + code&0x3FF)}
		}
	}
	raw := substr(source, offset-2, length+2)
	onError(offset-2, "BAD_DQ_ESCAPE", "Invalid escape sequence "+unitsToString(raw))
	return raw
}

type blockScalarHeader struct {
	mode    uint16
	indent  int
	chomp   uint16
	comment string
	length  int
}

func resolveBlockScalar(ctx *composeContext, scalar *cst.Token, onError onErrorFunc) scalarResolution {
	start := scalar.Offset
	header := parseBlockScalarHeader(scalar, ctx.options.Strict, onError)
	if header == nil {
		return scalarResolution{value: []uint16{}, valueRange: []int{start, start, start}}
	}
	scalarType := BlockLiteral
	if header.mode == '>' {
		scalarType = BlockFolded
	}
	var lines [][2][]uint16
	if len(scalar.Source) > 0 {
		lines = splitLines(scalar.Source)
	}
	// `content === '' || content === '\r'`
	isEmptyContent := func(content []uint16) bool {
		return len(content) == 0 || len(content) == 1 && content[0] == '\r'
	}
	// determine the end of content & start of chomping
	chompStart := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		content := lines[i][1]
		if isEmptyContent(content) {
			chompStart = i
		} else {
			break
		}
	}
	// shortcut for empty contents
	if chompStart == 0 {
		value := []uint16{}
		if header.chomp == '+' && len(lines) > 0 {
			for range max(1, len(lines)-1) {
				value = append(value, '\n')
			}
		}
		end := start + header.length
		if len(scalar.Source) > 0 {
			end += len(scalar.Source)
		}
		return scalarResolution{value: value, scalarType: scalarType, comment: header.comment, valueRange: []int{start, end, end}}
	}
	// find the indentation level to trim from start
	trimIndent := scalar.Indent + header.indent
	offset := scalar.Offset + header.length
	contentStart := 0
	for i := 0; i < chompStart; i++ {
		indent, content := lines[i][0], lines[i][1]
		if isEmptyContent(content) {
			if header.indent == 0 && len(indent) > trimIndent {
				trimIndent = len(indent)
			}
		} else {
			if len(indent) < trimIndent {
				message := "Block scalars with more-indented leading empty lines must use an explicit indentation indicator"
				onError(offset+len(indent), "MISSING_CHAR", message, false)
			}
			if header.indent == 0 {
				trimIndent = len(indent)
			}
			contentStart = i
			if trimIndent == 0 && !ctx.atRoot {
				message := "Block scalar values in collections must be indented"
				onError(offset, "BAD_INDENT", message, false)
			}
			break
		}
		offset += len(indent) + len(content) + 1
	}
	// include trailing more-indented empty lines in content
	for i := len(lines) - 1; i >= chompStart; i-- {
		if len(lines[i][0]) > trimIndent {
			chompStart = i + 1
		}
	}
	value := []uint16{}
	sep := []uint16{}
	prevMoreIndented := false
	// leading whitespace is kept intact
	for i := 0; i < contentStart; i++ {
		value = append(append(value, slice(lines[i][0], trimIndent, len(lines[i][0]))...), '\n')
	}
	for i := contentStart; i < chompStart; i++ {
		indent, content := lines[i][0], lines[i][1]
		offset += len(indent) + len(content) + 1
		crlf := characterAt(content, len(content)-1) == '\r'
		if crlf {
			content = slice(content, 0, -1)
		}
		/* istanbul ignore if already caught in lexer */
		if len(content) > 0 && len(indent) < trimIndent {
			src := "first line"
			if header.indent != 0 {
				src = "explicit indentation indicator"
			}
			message := "Block scalar lines must not be less indented than their " + src
			rel := 1
			if crlf {
				rel = 2
			}
			onError(offset-len(content)-rel, "BAD_INDENT", message, false)
			indent = []uint16{}
		}
		trimmedIndent := slice(indent, trimIndent, len(indent))
		if scalarType == BlockLiteral {
			value = append(append(append(value, sep...), trimmedIndent...), content...)
			sep = []uint16{'\n'}
		} else if len(indent) > trimIndent || characterAt(content, 0) == '\t' {
			// more-indented content within a folded block
			if equalsASCII(sep, " ") {
				sep = []uint16{'\n'}
			} else if !prevMoreIndented && equalsASCII(sep, "\n") {
				sep = []uint16{'\n', '\n'}
			}
			value = append(append(append(value, sep...), trimmedIndent...), content...)
			sep = []uint16{'\n'}
			prevMoreIndented = true
		} else if len(content) == 0 {
			// empty line
			if equalsASCII(sep, "\n") {
				value = append(value, '\n')
			} else {
				sep = []uint16{'\n'}
			}
		} else {
			value = append(append(value, sep...), content...)
			sep = []uint16{' '}
			prevMoreIndented = false
		}
	}
	switch header.chomp {
	case '-':
	case '+':
		for i := chompStart; i < len(lines); i++ {
			value = append(append(value, '\n'), slice(lines[i][0], trimIndent, len(lines[i][0]))...)
		}
		if characterAt(value, len(value)-1) != '\n' {
			value = append(value, '\n')
		}
	default:
		value = append(value, '\n')
	}
	end := start + header.length + len(scalar.Source)
	return scalarResolution{value: value, scalarType: scalarType, comment: header.comment, valueRange: []int{start, end, end}}
}

func parseBlockScalarHeader(scalar *cst.Token, strict bool, onError onErrorFunc) *blockScalarHeader {
	offset, props := scalar.Offset, scalar.Props
	/* istanbul ignore if should not happen */
	if props[0].Type != "block-scalar-header" {
		onError(props[0], "IMPOSSIBLE", "Block scalar header not found", false)
		return nil
	}
	source := props[0].Source
	mode := source[0]
	indent := 0
	var chomp uint16
	errorOffset := -1
	for i := 1; i < len(source); i++ {
		ch := source[i]
		if chomp == 0 && (ch == '-' || ch == '+') {
			chomp = ch
		} else {
			// `Number(ch)` is truthy only for the digits 1 to 9.
			n := 0
			if ch >= '1' && ch <= '9' {
				n = int(ch - '0')
			}
			if indent == 0 && n != 0 {
				indent = n
			} else if errorOffset == -1 {
				errorOffset = offset + i
			}
		}
	}
	if errorOffset != -1 {
		onError(errorOffset, "UNEXPECTED_TOKEN", "Block scalar header includes extra characters: "+unitsToString(source), false)
	}
	hasSpace := false
	comment := ""
	length := len(source)
	for i := 1; i < len(props); i++ {
		token := props[i]
		switch token.Type {
		case "space":
			hasSpace = true
			length += len(token.Source)
		case "newline":
			length += len(token.Source)
		case "comment":
			if strict && !hasSpace {
				message := "Comments must be separated from other tokens by white space characters"
				onError(token, "MISSING_CHAR", message, false)
			}
			length += len(token.Source)
			comment = unitsToString(substring(token.Source, 1, len(token.Source)))
		case "error":
			onError(token, "UNEXPECTED_TOKEN", token.Message, false)
			length += len(token.Source)
		/* istanbul ignore next should not happen */
		default:
			message := "Unexpected token in block scalar header: " + token.Type
			onError(token, "UNEXPECTED_TOKEN", message, false)
			if token.HasSource() && len(token.Source) > 0 {
				length += len(token.Source)
			}
		}
	}
	return &blockScalarHeader{mode: mode, indent: indent, chomp: chomp, comment: comment, length: length}
}

// splitLines returns the lines of a block scalar's source split up as [indent, content]:
// `source.split(/\n( *)/)`, with the first line's leading spaces split off too.
func splitLines(source []uint16) [][2][]uint16 {
	var split [][]uint16
	start := 0
	for index := 0; index < len(source); index++ {
		if source[index] != '\n' {
			continue
		}
		split = append(split, source[start:index:index])
		spaces := index + 1
		for spaces < len(source) && source[spaces] == ' ' {
			spaces++
		}
		split = append(split, source[index+1:spaces:spaces])
		start = spaces
		index = spaces - 1
	}
	split = append(split, source[start:len(source):len(source)])
	first := split[0]
	leading := 0
	for leading < len(first) && first[leading] == ' ' {
		leading++
	}
	line0 := [2][]uint16{{}, first}
	if leading > 0 {
		line0 = [2][]uint16{first[:leading:leading], first[leading:]}
	}
	lines := [][2][]uint16{line0}
	for i := 1; i < len(split); i += 2 {
		lines = append(lines, [2][]uint16{split[i], split[i+1]})
	}
	return lines
}
