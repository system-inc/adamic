package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// harnessCalls are the assert.js property calls, longest first so notSameValue is not read as
// sameValue. assert.throws is handled apart: its first argument, a constructor, becomes the name
// assertThrows takes.
var harnessCalls = []struct {
	from string
	to   string
}{
	{"assert.notSameValue", "assertNotSameValue"},
	{"assert.compareArray", "assertCompareArray"},
	{"assert._isSameValue", "assertIsSameValue"},
	{"assert.sameValue", "assertSameValue"},
}

// rewriteHarnessCalls renames harness property calls in code, leaving strings and comments alone.
// assert.throws(TypeError, fn) becomes assertThrows("TypeError", fn). A first argument that isn't a
// plain name (or a dotted name) is left as written, and the checker then refuses the call.
func rewriteHarnessCalls(source string) string {
	return rewriteHarnessCallsWithSentinels(source, globalUndefinedUnshadowed(source), globalIntrinsicUnshadowed(source, "Date"))
}

func rewriteHarnessCallsWithSentinels(source string, undefinedSafe, dateSafe bool) string {
	var builder strings.Builder
	builder.Grow(len(source))
	for index := 0; index < len(source); {
		if source[index] == '/' && index+1 < len(source) && source[index+1] == '/' {
			end := strings.IndexByte(source[index:], '\n')
			if end < 0 {
				builder.WriteString(source[index:])
				break
			}
			builder.WriteString(source[index : index+end+1])
			index += end + 1
			continue
		}
		if source[index] == '/' && index+1 < len(source) && source[index+1] == '*' {
			end := strings.Index(source[index+2:], "*/")
			if end < 0 {
				builder.WriteString(source[index:])
				break
			}
			builder.WriteString(source[index : index+2+end+2])
			index += 2 + end + 2
			continue
		}
		if source[index] == '\'' || source[index] == '"' {
			end := endOfString(source, index)
			builder.WriteString(source[index:end])
			index = end
			continue
		}
		if source[index] == '`' {
			index = writeTemplate(&builder, source, index, undefinedSafe, dateSafe)
			continue
		}
		if !identifierBoundary(source, index) {
			builder.WriteByte(source[index])
			index++
			continue
		}
		if strings.HasPrefix(source[index:], "assert.throws") && identifierEnded(source, index+len("assert.throws")) {
			index = writeAssertThrows(&builder, source, index)
			continue
		}
		replaced := false
		for _, call := range harnessCalls {
			if strings.HasPrefix(source[index:], call.from) && identifierEnded(source, index+len(call.from)) {
				if call.from == "assert.sameValue" || call.from == "assert.notSameValue" {
					if rewritten, next, safe := rewriteSentinelAssertion(source, index, call.from, undefinedSafe, dateSafe); safe {
						builder.WriteString(rewritten)
						index = next
						replaced = true
						break
					}
				}
				builder.WriteString(call.to)
				index += len(call.from)
				replaced = true
				break
			}
		}
		if replaced {
			continue
		}
		builder.WriteByte(source[index])
		index++
	}
	return builder.String()
}

// writeAssertThrows copies assert.throws(...) with the callee renamed and a name-like first
// argument quoted. It returns the index just after the first argument, or after the callee if the
// argument isn't a name.
func writeAssertThrows(builder *strings.Builder, source string, index int) int {
	builder.WriteString("assertThrows")
	index += len("assert.throws")
	for index < len(source) && isSpace(source[index]) {
		builder.WriteByte(source[index])
		index++
	}
	if index >= len(source) || source[index] != '(' {
		return index
	}
	builder.WriteByte('(')
	index++
	argumentStart := index
	for index < len(source) && isSpace(source[index]) {
		index++
	}
	name, next := constructorName(source, index)
	if name == "" {
		return argumentStart
	}
	for cursor := argumentStart; cursor < index; cursor++ {
		builder.WriteByte(source[cursor])
	}
	builder.WriteByte('"')
	builder.WriteString(name)
	builder.WriteByte('"')
	return next
}

// constructorName is TypeError or foo.Bar at index, and the index after it. Empty if the text
// there isn't a dotted name.
func constructorName(source string, index int) (string, int) {
	start := index
	if index >= len(source) || !isIdentifierStart(rune(source[index])) {
		return "", index
	}
	index++
	for index < len(source) {
		character := rune(source[index])
		if isIdentifierPart(character) {
			index++
			continue
		}
		if source[index] == '.' && index+1 < len(source) && isIdentifierStart(rune(source[index+1])) {
			index++
			continue
		}
		break
	}
	return source[start:index], index
}

func identifierBoundary(source string, index int) bool {
	if index == 0 {
		return true
	}
	previous, _ := utf8.DecodeLastRuneInString(source[:index])
	return !isIdentifierPart(previous)
}

func identifierEnded(source string, index int) bool {
	if index >= len(source) {
		return true
	}
	next, _ := utf8.DecodeRuneInString(source[index:])
	return !isIdentifierPart(next)
}

func isIdentifierStart(character rune) bool {
	return character == '$' || character == '_' || unicode.IsLetter(character)
}

func isIdentifierPart(character rune) bool {
	return isIdentifierStart(character) || unicode.IsDigit(character)
}

func isSpace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\n' || character == '\r'
}

func endOfString(source string, index int) int {
	quote := source[index]
	index++
	for index < len(source) {
		if source[index] == '\\' && index+1 < len(source) {
			index += 2
			continue
		}
		if source[index] == quote {
			return index + 1
		}
		if quote != '`' && source[index] == '\n' {
			return index
		}
		index++
	}
	return index
}

// writeTemplate copies a template literal, rewriting harness calls inside each ${ } expression.
func writeTemplate(builder *strings.Builder, source string, index int, undefinedSafe, dateSafe bool) int {
	builder.WriteByte('`')
	index++
	for index < len(source) {
		if source[index] == '\\' && index+1 < len(source) {
			builder.WriteString(source[index : index+2])
			index += 2
			continue
		}
		if source[index] == '`' {
			builder.WriteByte('`')
			return index + 1
		}
		if source[index] == '$' && index+1 < len(source) && source[index+1] == '{' {
			expression, next, ok := expressionInside(source, index+2, 1)
			if !ok {
				builder.WriteString(source[index:])
				return len(source)
			}
			builder.WriteString("${")
			builder.WriteString(rewriteHarnessCallsWithSentinels(expression, undefinedSafe, dateSafe))
			builder.WriteByte('}')
			index = next
			continue
		}
		builder.WriteByte(source[index])
		index++
	}
	return index
}

// expressionInside is the source of a ${ } expression starting at index, already inside one brace,
// and the index just after its closing brace.
func expressionInside(source string, index int, depth int) (string, int, bool) {
	start := index
	for index < len(source) && depth > 0 {
		if source[index] == '\\' && index+1 < len(source) {
			index += 2
			continue
		}
		if source[index] == '\'' || source[index] == '"' || source[index] == '`' {
			index = endOfQuoted(source, index)
			continue
		}
		if source[index] == '{' {
			depth++
		} else if source[index] == '}' {
			depth--
			if depth == 0 {
				return source[start:index], index + 1, true
			}
		}
		index++
	}
	return "", start, false
}

func endOfQuoted(source string, index int) int {
	if source[index] == '`' {
		index++
		for index < len(source) {
			if source[index] == '\\' && index+1 < len(source) {
				index += 2
				continue
			}
			if source[index] == '`' {
				return index + 1
			}
			if source[index] == '$' && index+1 < len(source) && source[index+1] == '{' {
				_, next, ok := expressionInside(source, index+2, 1)
				if !ok {
					return len(source)
				}
				index = next
				continue
			}
			index++
		}
		return index
	}
	return endOfString(source, index)
}

// stripUseStrict drops directive lines. A module is already strict, and stage 0 can't lower a
// string used as a statement, so leaving the directive in would refuse every test that carries one.
func stripUseStrict(source string) string {
	var builder strings.Builder
	for _, line := range strings.Split(source, "\n") {
		trim := strings.TrimSpace(line)
		if trim == `"use strict";` || trim == `'use strict';` || trim == `"use strict"` || trim == `'use strict'` {
			continue
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	return builder.String()
}

// codeOnly is the source with comments and strings blanked, so a feature scan doesn't fire on a
// comment that mentions eval or Symbol. Newlines stay, so positions aren't required.
func codeOnly(source string) string {
	var builder strings.Builder
	builder.Grow(len(source))
	for index := 0; index < len(source); {
		if source[index] == '/' && index+1 < len(source) && (source[index+1] == '/' || source[index+1] == '*') {
			if source[index+1] == '/' {
				end := strings.IndexByte(source[index:], '\n')
				if end < 0 {
					break
				}
				builder.WriteByte('\n')
				index += end + 1
				continue
			}
			end := strings.Index(source[index+2:], "*/")
			if end < 0 {
				break
			}
			chunk := source[index : index+2+end+2]
			builder.WriteString(strings.Repeat("\n", strings.Count(chunk, "\n")))
			index += 2 + end + 2
			continue
		}
		if source[index] == '\'' || source[index] == '"' || source[index] == '`' {
			end := endOfQuoted(source, index)
			chunk := source[index:end]
			builder.WriteString(strings.Repeat("\n", strings.Count(chunk, "\n")))
			index = end
			continue
		}
		builder.WriteByte(source[index])
		index++
	}
	return builder.String()
}
