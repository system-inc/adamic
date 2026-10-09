package postcss

import "strings"

// postcss-scss 4.0.9, lib/scss-parser.js: `class ScssParser extends Parser`, overriding atrule, comment,
// createTokenizer, raw and rule.
//
// Go has no inheritance: scssParser embeds the *parser and is installed as its `this` (parser.go's
// parserMethods), so the base parser's calls to those five methods reach these, and `super.method()` is
// s.parser.method(). Everything else ScssParser calls on `this` (init, spacesAndCommentsFromStart,
// stringFrom, checkMissedSemicolon, ...) is the base parser's, reached through the embedding.

type scssParser struct {
	*parser
}

// newScssParser is new ScssParser(input): Parser's constructor with this ScssParser.
func newScssParser(in *input) *parser {
	s := &scssParser{parser: &parser{}}
	s.this = s
	s.construct(in)
	return s.parser
}

// TypeError is the JavaScript TypeError postcss-scss throws where its rule reads tokens[0] of an empty
// array (a nested property with no word in it, `:"x" {}`). Prettier only turns errors with a line into its
// own and lets this one escape, so the file does not format; ParseSCSS marks it with printing.Syntax.
type TypeError struct {
	// Message is the error's message as V8 words it.
	Message string
}

func (typeError *TypeError) Error() string {
	return "TypeError: " + typeError.Message
}

func (s *scssParser) atrule(atToken token) {
	name := atToken.value
	prev := atToken
	for !s.tokenizer.endOfFile() {
		next, _ := s.tokenizer.nextToken()
		if next.kind == "word" && next.start == prev.end+1 {
			name += next.value
			prev = next
		} else {
			s.tokenizer.back(next)
			break
		}
	}

	s.parser.atrule(token{kind: "at-word", value: name, start: atToken.start, end: prev.end})
}

func (s *scssParser) comment(commentToken token) {
	if commentToken.inline {
		node := newComment()
		s.init(node, commentToken.start)
		node.raws["inline"] = true
		pos := s.input.fromOffset(commentToken.end)
		node.source.end = &position{
			column: pos.column,
			line:   pos.line,
			offset: commentToken.end + 1,
		}

		text := commentToken.value[2:]
		if strings.TrimFunc(text, isJavaScriptWhitespace) == "" {
			node.text = ""
			node.raws["left"] = text
			node.raws["right"] = ""
		} else {
			// text.match(/^(\s*)([^]*\S)(\s*)$/), as Parser.comment reads it.
			trimmedStart := strings.TrimLeftFunc(text, isJavaScriptWhitespace)
			trimmed := strings.TrimRightFunc(trimmedStart, isJavaScriptWhitespace)
			fixed := replaceCommentDelimiters(trimmed)
			node.text = fixed
			node.raws["left"] = text[:len(text)-len(trimmedStart)]
			node.raws["right"] = trimmedStart[len(trimmed):]
			node.raws["text"] = trimmed
		}
		node.mark("text")
	} else {
		s.parser.comment(commentToken)
	}
}

func (s *scssParser) createTokenizer() {
	s.tokenizer = newScssTokenizer(s.input)
}

func (s *scssParser) raw(each *node, prop string, tokens []token, customProperty bool) {
	s.parser.raw(each, prop, tokens, customProperty)
	if raws, isRaw := each.raws[prop].(*rawValue); isRaw {
		scss := raws.raw
		var all strings.Builder
		for _, i := range tokens {
			if i.kind == "comment" && i.inline {
				text := replaceCommentDelimiters(i.value[2:])
				all.WriteString("/*" + text + "*/")
			} else {
				all.WriteString(i.value)
			}
		}
		raws.raw = all.String()
		if scss != raws.raw {
			raws.scss = scss
			raws.hasScss = true
		}
	}
}

func (s *scssParser) rule(tokens []token) {
	withColon := false
	brackets := 0
	value := ""
	for _, i := range tokens {
		if withColon {
			if i.kind != "comment" && i.kind != "{" {
				value += i.value
			}
		} else if i.kind == "space" && strings.Contains(i.value, "\n") {
			break
		} else if i.kind == "(" {
			brackets += 1
		} else if i.kind == ")" {
			brackets -= 1
		} else if brackets == 0 && i.kind == ":" {
			withColon = true
		}
	}

	if !withColon || strings.TrimFunc(value, isJavaScriptWhitespace) == "" || startsLikeAProperty(value) {
		s.parser.rule(tokens)
	} else {
		tokens = tokens[:len(tokens)-1]
		node := newNestedDeclaration()
		s.init(node, tokens[0].start)

		// The last token that is not a space; there is one, the colon.
		var last token
		for i := len(tokens) - 1; i >= 0; i-- {
			if tokens[i].kind != "space" {
				last = tokens[i]
				break
			}
		}
		// `if (last[3])`: an end of 0, or none, falls to the start.
		if last.end > 0 {
			pos := s.input.fromOffset(last.end)
			node.source.end = &position{
				column: pos.column,
				line:   pos.line,
				offset: last.end + 1,
			}
		} else {
			pos := s.input.fromOffset(last.start)
			node.source.end = &position{
				column: pos.column,
				line:   pos.line,
				offset: last.start + 1,
			}
		}

		for {
			if len(tokens) == 0 {
				// tokens[0][0] on an empty array.
				panic(&TypeError{Message: "Cannot read properties of undefined (reading '0')"})
			}
			if tokens[0].kind == "word" {
				break
			}
			node.raws["before"] = node.rawString("before") + tokens[0].value
			tokens = tokens[1:]
		}

		// `if (tokens[0][2])`: a start of 0 keeps the one init set, which is the same offset.
		if tokens[0].start > 0 {
			pos := s.input.fromOffset(tokens[0].start)
			node.source.start = &position{
				column: pos.column,
				line:   pos.line,
				offset: tokens[0].start,
			}
		}

		node.prop = ""
		node.mark("prop")
		for len(tokens) > 0 {
			kind := tokens[0].kind
			if kind == ":" || kind == "space" || kind == "comment" {
				break
			}
			node.prop += tokens[0].value
			tokens = tokens[1:]
		}

		node.raws["between"] = ""

		var each token
		for len(tokens) > 0 {
			each = tokens[0]
			tokens = tokens[1:]

			if each.kind == ":" {
				node.raws["between"] = node.rawString("between") + each.value
				break
			} else {
				node.raws["between"] = node.rawString("between") + each.value
			}
		}

		if node.prop != "" && (node.prop[0] == '_' || node.prop[0] == '*') {
			node.raws["before"] = node.rawString("before") + node.prop[:1]
			node.prop = node.prop[1:]
		}
		node.raws["between"] = node.rawString("between") + s.spacesAndCommentsFromStart(&tokens)
		// precheckMissedSemicolon is a hook for postcss-safe-parser; the base parser's does nothing.

		for i := len(tokens) - 1; i > 0; i-- {
			each = tokens[i]
			if each.value == "!important" {
				node.important = true
				node.mark("important")
				text := s.stringFrom(&tokens, i)
				text = s.spacesFromEnd(&tokens) + text
				if text != " !important" {
					node.raws["important"] = text
				}
				break
			} else if each.value == "important" {
				cache := copyTokens(tokens, 0, len(tokens))
				str := ""
				for j := i; j > 0; j-- {
					kind := cache[j].kind
					if strings.HasPrefix(strings.TrimFunc(str, isJavaScriptWhitespace), "!") && kind != "space" {
						break
					}
					str = cache[len(cache)-1].value + str
					cache = cache[:len(cache)-1]
				}
				if strings.HasPrefix(strings.TrimFunc(str, isJavaScriptWhitespace), "!") {
					node.important = true
					node.mark("important")
					node.raws["important"] = str
					tokens = cache
				}
			}

			if each.kind != "space" && each.kind != "comment" {
				break
			}
		}

		s.raw(node, "value", tokens, false)

		if strings.Contains(node.value, ":") {
			s.checkMissedSemicolon(tokens)
		}

		s.current = node
	}
}

// startsLikeAProperty is /^[#:A-Za-z-]/.test(value).
func startsLikeAProperty(value string) bool {
	if value == "" {
		return false
	}
	first := value[0]
	return first == '#' || first == ':' || first == '-' || (first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z')
}

// replaceCommentDelimiters is text.replace(/(\*\/|\/\*)/g, '*//*'): every "*/" and "/*", found left to
// right without overlap, becomes "*//*".
func replaceCommentDelimiters(text string) string {
	var result strings.Builder
	for index := 0; index < len(text); index++ {
		if index+1 < len(text) && ((text[index] == '*' && text[index+1] == '/') || (text[index] == '/' && text[index+1] == '*')) {
			result.WriteString("*//*")
			index++
			continue
		}
		result.WriteByte(text[index])
	}
	return result.String()
}
