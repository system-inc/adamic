// Port of postcss-values-parser 2.0.1 lib/tokenize.js.
//
// The tokenizer walks UTF-16 code units, as upstream's charCodeAt and slice do, so every column and
// every index it reports is in upstream's unit. The parser converts indexes to bytes when it builds
// nodes; columns stay in UTF-16 units.

package values

import "math"

const (
	openBracket  = '{'
	closeBracket = '}'
	openParen    = '('
	closeParen   = ')'
	singleQuote  = '\''
	doubleQuote  = '"'
	backslash    = '\\'
	slash        = '/'
	period       = '.'
	comma        = ','
	colon        = ':'
	asterisk     = '*'
	minus        = '-'
	plus         = '+'
	pound        = '#'
	newline      = '\n'
	space        = ' '
	feed         = '\f'
	tab          = '\t'
	cr           = '\r'
	at           = '@'
	lowerE       = 'e'
	upperE       = 'E'
	digit0       = '0'
	digit9       = '9'
	lowerU       = 'u'
	upperU       = 'U'
)

// undefinedNext is upstream's `next` before anything assigns it. The bracket and paren tokens report
// `next - offset` as their end column without assigning `next` first, so the first token of a value
// that opens with one of them ends at column NaN, and later ones end where the previous token did.
const undefinedNext = math.MinInt

// token is upstream's token array: [type, value, startLine, startColumn, endLine, endColumn, index].
type token struct {
	kind        string
	value       []uint16
	startLine   int
	startColumn int
	endLine     int
	endColumn   int
	// endColumnIsNaN is an end column computed from an undefined `next`.
	endColumnIsNaN bool
	index          int
}

// endColumnValue is token[5] as JavaScript holds it: a number, or NaN.
func (token *token) endColumnValue() any {
	if token.endColumnIsNaN {
		return math.NaN()
	}
	return token.endColumn
}

// atEnd is upstream's /[ \n\t\r\{\(\)'"\\;,/]/g.
func atEnd(code uint16) bool {
	switch code {
	case ' ', '\n', '\t', '\r', '{', '(', ')', '\'', '"', '\\', ';', ',', '/':
		return true
	}
	return false
}

// wordEnd is upstream's /[ \n\t\r\(\)\{\}\*:;@!&'"\+\|~>,\[\]\\]|\/(?=\*)/g at one index.
func wordEnd(css []uint16, index int) bool {
	switch css[index] {
	case ' ', '\n', '\t', '\r', '(', ')', '{', '}', '*', ':', ';', '@', '!', '&', '\'', '"', '+', '|', '~', '>', ',', '[', ']', '\\':
		return true
	case '/':
		return index+1 < len(css) && css[index+1] == '*'
	}
	return false
}

// wordEndNum is upstream's /[ \n\t\r\(\)\{\}\*:;@!&'"\-\+\|~>,\[\]\\]|\//g at one index.
func wordEndNum(css []uint16, index int) bool {
	switch css[index] {
	case ' ', '\n', '\t', '\r', '(', ')', '{', '}', '*', ':', ';', '@', '!', '&', '\'', '"', '-', '+', '|', '~', '>', ',', '[', ']', '\\', '/':
		return true
	}
	return false
}

// alphaNum is upstream's /^[a-z0-9]/i. JavaScript's case-insensitive class without the u flag folds
// only within ASCII here, so this is a plain ASCII test.
func alphaNum(code uint16) bool {
	return (code >= 'a' && code <= 'z') || (code >= 'A' && code <= 'Z') || (code >= '0' && code <= '9')
}

// unicodeRange is upstream's /^[a-f0-9?\-]/i.
func unicodeRange(code uint16) bool {
	return (code >= 'a' && code <= 'f') || (code >= 'A' && code <= 'F') || (code >= '0' && code <= '9') || code == '?' || code == '-'
}

// regexLastIndex is `regex.lastIndex = from; regex.test(css); regex.lastIndex` for the one-character
// global regexes above: one past the match, or 0 when there is none.
func regexLastIndex(css []uint16, from int, matches func(css []uint16, index int) bool) int {
	for index := from; index < len(css); index++ {
		if matches(css, index) {
			return index + 1
		}
	}
	return 0
}

// indexOfUnits is css.indexOf(search, from).
func indexOfUnits(css []uint16, search []uint16, from int) int {
	if from < 0 {
		from = 0
	}
	for index := from; index+len(search) <= len(css); index++ {
		found := true
		for offset := range search {
			if css[index+offset] != search[offset] {
				found = false
				break
			}
		}
		if found {
			return index
		}
	}
	return -1
}

func tokenize(input []uint16, options Options) []token {
	tokens := []token{}
	css := input
	length := len(css)
	offset := -1
	line := 1
	pos := 0
	parentCount := 0
	isURLArg := false

	var code int
	next := undefinedNext
	var quote uint16
	var escaped bool
	var escapePos int

	// charCodeAt is css.charCodeAt: -1 stands in for NaN past either end, equal to no code.
	charCodeAt := func(index int) int {
		if index < 0 || index >= length {
			return -1
		}
		return int(css[index])
	}

	// slice is css.slice for the non-negative bounds this file passes, clamped the way slice clamps.
	slice := func(start int, end int) []uint16 {
		start = min(max(start, 0), length)
		end = min(max(end, 0), length)
		if end < start {
			end = start
		}
		return css[start:end]
	}

	push := func(kind string, value []uint16, startLine int, startColumn int, endLine int, endColumn int, index int) {
		tokens = append(tokens, token{kind: kind, value: value, startLine: startLine, startColumn: startColumn, endLine: endLine, endColumn: endColumn, index: index})
	}

	// pushStale pushes a bracket or paren token, whose end column is `next - offset` with a stale `next`.
	pushStale := func(kind string) {
		each := token{kind: kind, value: slice(pos, pos+1), startLine: line, startColumn: pos - offset, endLine: line, index: pos}
		if next == undefinedNext {
			each.endColumnIsNaN = true
		} else {
			each.endColumn = next - offset
		}
		tokens = append(tokens, each)
	}

	unclosed := func(what string) {
		panic(&Error{Name: "TokenizeError", Message: "Unclosed " + what + " at line: " + itoa(line) + ", column: " + itoa(pos-offset) + ", token: " + itoa(pos)})
	}

	for pos < length {
		code = charCodeAt(pos)

		if code == newline {
			offset = pos
			line += 1
		}

		switch code {
		case newline, space, tab, cr, feed:
			next = pos
			for {
				next += 1
				code = charCodeAt(next)
				if code == newline {
					offset = next
					line += 1
				}
				if !(code == space || code == newline || code == tab || code == cr || code == feed) {
					break
				}
			}

			push("space", slice(pos, next), line, pos-offset, line, next-offset, pos)

			pos = next - 1

		case colon:
			next = pos + 1
			push("colon", slice(pos, next), line, pos-offset, line, next-offset, pos)

			pos = next - 1

		case comma:
			next = pos + 1
			push("comma", slice(pos, next), line, pos-offset, line, next-offset, pos)

			pos = next - 1

		case openBracket:
			pushStale("{")

		case closeBracket:
			pushStale("}")

		case openParen:
			parentCount++
			isURLArg = !isURLArg && parentCount == 1 &&
				len(tokens) > 0 &&
				tokens[len(tokens)-1].kind == "word" &&
				string(utf16Decode(tokens[len(tokens)-1].value)) == "url"
			pushStale("(")

		case closeParen:
			parentCount--
			isURLArg = isURLArg && parentCount > 0
			pushStale(")")

		case singleQuote, doubleQuote:
			if code == singleQuote {
				quote = '\''
			} else {
				quote = '"'
			}
			next = pos
			for {
				escaped = false
				next = indexOfUnits(css, []uint16{quote}, next+1)
				if next == -1 {
					unclosed("quote")
				}
				escapePos = next
				for charCodeAt(escapePos-1) == backslash {
					escapePos -= 1
					escaped = !escaped
				}
				if !escaped {
					break
				}
			}

			push("string", slice(pos, next+1), line, pos-offset, line, next-offset, pos)
			pos = next

		case at:
			lastIndex := regexLastIndex(css, pos+1, func(css []uint16, index int) bool { return atEnd(css[index]) })

			if lastIndex == 0 {
				next = length - 1
			} else {
				next = lastIndex - 2
			}

			push("atword", slice(pos, next+1), line, pos-offset, line, next-offset, pos)
			pos = next

		case backslash:
			next = pos
			code = charCodeAt(next + 1)

			// Upstream tests an `escape` variable here that nothing assigns, so the backslash is always
			// a one-character word and the branch that would take the escaped character is dead.

			push("word", slice(pos, next+1), line, pos-offset, line, next-offset, pos)

			pos = next

		case plus, minus, asterisk:
			next = pos + 1
			nextChar := charCodeAt(pos + 1)

			// Upstream also computes a prevChar here that it never reads.

			// if the operator is immediately followed by a word character, then we
			// have a prefix of some kind, and should fall-through. eg. -webkit

			// look for --* for custom variables
			if code == minus && nextChar == minus {
				next++

				push("word", slice(pos, next), line, pos-offset, line, next-offset, pos)

				pos = next - 1
				break
			}

			push("operator", slice(pos, next), line, pos-offset, line, next-offset, pos)

			pos = next - 1

		default:
			if code == slash && (charCodeAt(pos+1) == asterisk || (options.Loose && !isURLArg && charCodeAt(pos+1) == slash)) {
				isStandardComment := charCodeAt(pos+1) == asterisk

				if isStandardComment {
					next = indexOfUnits(css, []uint16{'*', '/'}, pos+2) + 1
					if next == 0 {
						unclosed("comment")
					}
				} else {
					newlinePos := indexOfUnits(css, []uint16{'\n'}, pos+2)

					if newlinePos != -1 {
						next = newlinePos - 1
					} else {
						next = length
					}
				}

				content := slice(pos, next+1)
				// lines = content.split('\n'), of which only the count and the last line's length are read.
				last := 0
				lastLineLength := len(content)
				for index, unit := range content {
					if unit == '\n' {
						last++
						lastLineLength = len(content) - index - 1
					}
				}

				var nextLine, nextOffset int
				if last > 0 {
					nextLine = line + last
					nextOffset = next - lastLineLength
				} else {
					nextLine = line
					nextOffset = offset
				}

				push("comment", content, line, pos-offset, nextLine, next-nextOffset, pos)

				offset = nextOffset
				line = nextLine
				pos = next

			} else if code == pound && !(pos+1 < length && alphaNum(css[pos+1])) {
				next = pos + 1

				push("#", slice(pos, next), line, pos-offset, line, next-offset, pos)

				pos = next - 1
			} else if (code == lowerU || code == upperU) && charCodeAt(pos+1) == plus {
				next = pos + 2

				for {
					next += 1
					code = charCodeAt(next)
					if !(next < length && unicodeRange(css[next])) {
						break
					}
				}

				push("unicoderange", slice(pos, next), line, pos-offset, line, next-offset, pos)
				pos = next - 1
			} else if code == slash {
				// catch a regular slash, that isn't a comment
				next = pos + 1

				push("operator", slice(pos, next), line, pos-offset, line, next-offset, pos)

				pos = next - 1
			} else {
				regex := wordEnd
				isWordEndNum := false

				// we're dealing with a word that starts with a number
				// those get treated differently
				if code >= digit0 && code <= digit9 {
					regex = wordEndNum
					isWordEndNum = true
				}

				lastIndex := regexLastIndex(css, pos+1, regex)

				if lastIndex == 0 {
					next = length - 1
				} else {
					next = lastIndex - 2
				}

				// Exponential number notation with minus or plus: 1e-10, 1e+10
				if isWordEndNum || code == period {
					ncode := charCodeAt(next)
					ncode1 := charCodeAt(next + 1)
					ncode2 := charCodeAt(next + 2)

					if (ncode == lowerE || ncode == upperE) &&
						(ncode1 == minus || ncode1 == plus) &&
						(ncode2 >= digit0 && ncode2 <= digit9) {
						lastIndex := regexLastIndex(css, next+2, wordEndNum)

						if lastIndex == 0 {
							next = length - 1
						} else {
							next = lastIndex - 2
						}
					}
				}

				push("word", slice(pos, next+1), line, pos-offset, line, next-offset, pos)
				pos = next
			}
		}

		pos++
	}

	return tokens
}
