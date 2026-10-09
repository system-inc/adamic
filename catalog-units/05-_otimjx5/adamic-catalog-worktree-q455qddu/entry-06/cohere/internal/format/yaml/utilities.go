package yaml

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/yaml/unist"
)

// src/language-yaml/utilities.js.
//
// mapNode and defineShortcut are not ported. Upstream's preprocess uses them to copy every parent node
// and give the copies getters (document.head, item.key, value.content); unist.Node.Field answers those
// names itself, and nothing the printer does can tell a copy from the original (see preprocess).

// isNode is upstream's isNode(value, types): a node whose type is one of types. A nil node is none.
func isNode(node *unist.Node, types ...string) bool {
	return node != nil && slices.Contains(types, node.NodeType)
}

// childAt is node.children[index], nil where upstream's is undefined. The shortcuts print-preprocess.js
// defines are this: document.head and .body, item.key and .value, and every .content.
func childAt(node *unist.Node, index int) *unist.Node {
	if node == nil || index < 0 || index >= len(node.Children) {
		return nil
	}
	return node.Children[index]
}

// contentOf is node.content: children[0] of a documentBody, sequenceItem, flowSequenceItem,
// mappingKey or mappingValue.
func contentOf(node *unist.Node) *unist.Node { return childAt(node, 0) }

// isJavaScriptWhitespace is JavaScript's \s, which is also the set String.prototype.trim removes:
// WhiteSpace and LineTerminator in ECMA-262.
func isJavaScriptWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x00A0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return character >= 0x2000 && character <= 0x200A
}

// trim, trimStart and trimEnd are JavaScript's String.prototype methods of those names.
func trim(text string) string      { return strings.TrimFunc(text, isJavaScriptWhitespace) }
func trimStart(text string) string { return strings.TrimLeftFunc(text, isJavaScriptWhitespace) }
func trimEnd(text string) string   { return strings.TrimRightFunc(text, isJavaScriptWhitespace) }

// isNextLineEmpty is upstream's isNextLineEmpty: from the node's last character, is the line after the
// node's line empty?
//
// Upstream walks UTF-16 units from end.offset - 1. Here it walks runes from the start of the node's
// last rune. Only "\n" and \S are tested, and a rune answers both as each of its units would.
func isNextLineEmpty(node *unist.Node, text string) bool {
	newlineCount := 0
	index := node.Position.End.Offset
	if index > 0 {
		_, width := utf8.DecodeLastRuneInString(text[:index])
		index -= width
	}
	// Upstream's text[-1] for a node ending at 0 is undefined: not "\n", and \S of it cannot count while
	// newlineCount is 0. So starting at 0 is the same.
	for index < len(text) {
		character, width := utf8.DecodeRuneInString(text[index:])

		if character == '\n' {
			newlineCount++
		}

		if newlineCount == 1 && !isJavaScriptWhitespace(character) {
			return false
		}

		if newlineCount == 2 {
			return true
		}
		index += width
	}

	return false
}

// isLastDescendantNode is upstream's isLastDescendantNode: is the path, at every list it passes
// through, at the last element?
func isLastDescendantNode(path *astPath) bool {
	node := currentNode(path)

	switch node.NodeType {
	case "tag", "anchor", "comment":
		return false
	}

	stack := path.Stack()
	for index := 1; index < len(stack); index++ {
		item := stack[index]
		parentItem := stack[index-1]

		// Every list a YAML node holds is a []*unist.Node (children and the comment lists), and an
		// index on the stack is an int.
		if list, isList := parentItem.([]*unist.Node); isList {
			if position, isNumber := item.(int); isNumber && position != len(list)-1 {
				return false
			}
		}
	}

	return true
}

// getLastDescendantNode is upstream's getLastDescendantNode.
func getLastDescendantNode(node *unist.Node) *unist.Node {
	if len(node.Children) > 0 {
		return getLastDescendantNode(node.Children[len(node.Children)-1])
	}
	return node
}

// isPrettierIgnore is upstream's isPrettierIgnore.
func isPrettierIgnore(comment *unist.Node) bool {
	return trim(comment.Value) == "prettier-ignore"
}

// hasPrettierIgnore is upstream's hasPrettierIgnore: the last leading comment, or for a document body,
// the document head's last end comment, is prettier-ignore.
func hasPrettierIgnore(path *astPath) bool {
	node := currentNode(path)

	if node.NodeType == "documentBody" {
		documentHead := childAt(parentNode(path), 0)
		return hasEndComments(documentHead) &&
			isPrettierIgnore(documentHead.EndComments[len(documentHead.EndComments)-1])
	}

	return hasLeadingComments(node) && isPrettierIgnore(node.LeadingComments[len(node.LeadingComments)-1])
}

// isEmptyNode is upstream's isEmptyNode.
func isEmptyNode(node *unist.Node) bool {
	return len(node.Children) == 0 && !hasComments(node)
}

// hasComments is upstream's hasComments.
func hasComments(node *unist.Node) bool {
	return hasLeadingComments(node) ||
		hasMiddleComments(node) ||
		hasIndicatorComment(node) ||
		hasTrailingComment(node) ||
		hasEndComments(node)
}

// The has* predicates are upstream's, which read node?.field: a nil node has none.

func hasLeadingComments(node *unist.Node) bool {
	return node != nil && len(node.LeadingComments) > 0
}

func hasMiddleComments(node *unist.Node) bool {
	return node != nil && len(node.MiddleComments) > 0
}

func hasIndicatorComment(node *unist.Node) bool {
	return node != nil && node.IndicatorComment != nil
}

func hasTrailingComment(node *unist.Node) bool {
	return node != nil && node.TrailingComment != nil
}

func hasEndComments(node *unist.Node) bool {
	return node != nil && len(node.EndComments) > 0
}

// splitWithSingleSpace is upstream's splitWithSingleSpace:
//
//	" a   b c   d e   f " -> [" a   b", "c   d", "e   f "]
//
// Upstream splits on /(?<!^| ) (?! |$)/, a lookaround RE2 does not have: a space that is neither at the
// start nor preceded by a space, and neither followed by a space nor at the end. The space is ASCII, so
// the scan is by byte.
func splitWithSingleSpace(text string) []string {
	if text == "" {
		return []string{}
	}
	parts := []string{}
	start := 0
	for index := 0; index < len(text); index++ {
		if text[index] != ' ' {
			continue
		}
		if index == 0 || text[index-1] == ' ' {
			continue
		}
		if index+1 == len(text) || text[index+1] == ' ' {
			continue
		}
		parts = append(parts, text[start:index])
		start = index + 1
	}
	return append(parts, text[start:])
}

// getFlowScalarLineContents is upstream's getFlowScalarLineContents: the words of each line of a plain
// or quoted scalar, lines joined the way YAML folds them unless proseWrap is preserve.
func getFlowScalarLineContents(nodeType string, content string, settings *settings) [][]string {
	split := strings.Split(content, "\n")
	rawLineContents := make([]string, len(split))
	for index, lineContent := range split {
		last := len(split) - 1
		switch {
		case index == 0 && index == last:
			rawLineContents[index] = lineContent
		case index != 0 && index != last:
			rawLineContents[index] = trim(lineContent)
		case index == 0:
			rawLineContents[index] = trimEnd(lineContent)
		default:
			rawLineContents[index] = trimStart(lineContent)
		}
	}

	if settings.proseWrap == "preserve" {
		lines := make([][]string, len(rawLineContents))
		for index, lineContent := range rawLineContents {
			if lineContent != "" {
				lines[index] = []string{lineContent}
			} else {
				lines[index] = []string{}
			}
		}
		return lines
	}

	lines := [][]string{}
	for index, line := range rawLineContents {
		words := splitWithSingleSpace(line)

		if index > 0 &&
			len(rawLineContents[index-1]) > 0 &&
			len(words) > 0 &&
			!(
			// trailing backslash in quoteDouble should be preserved
			nodeType == "quoteDouble" && strings.HasSuffix(lastWord(lines[len(lines)-1]), "\\")) {
			lines[len(lines)-1] = append(append([]string{}, lines[len(lines)-1]...), words...)
		} else {
			lines = append(lines, words)
		}
	}

	if settings.proseWrap == "never" {
		joined := make([][]string, len(lines))
		for index, words := range lines {
			joined[index] = []string{strings.Join(words, " ")}
		}
		return joined
	}
	return lines
}

// lastWord is words.at(-1). Upstream calls .endsWith on it, which would throw on an empty line; that
// cannot happen there, since the previous raw line was not empty and so split to at least one word.
func lastWord(words []string) string {
	if len(words) == 0 {
		return ""
	}
	return words[len(words)-1]
}

// leadingSpacePattern is /^(?<leadingSpace> *)[^\n\r ]/m.
//
// RE2's (?m) ^ matches only after \n, where JavaScript's also matches after \r, U+2028 and U+2029. The
// text has no \r (line endings are normalized before parsing), and a line start after U+2028 or U+2029
// is never the first match: the line holding that character has a non-space, non-newline character (the
// character itself) and so already matched at its own start.
var leadingSpacePattern = regexp.MustCompile(`(?m)^( *)[^\n\r ]`)

// trailingSpacesAndTabsPattern is /[ \t]+$/ without the m flag: at the very end.
var trailingSpacesAndTabsPattern = regexp.MustCompile(`[ \t]+$`)

// getBlockValueLineContents is upstream's getBlockValueLineContents: the words of each line of a block
// scalar, with the indentation removed and the trailing empty lines its chomping keeps.
func getBlockValueLineContents(node *unist.Node, parentIndent int, isLastDescendant bool, options *options, settings *settings) [][]string {
	content := ""
	if node.Position.Start.Line != node.Position.End.Line {
		// exclude open line `>` or `|`
		raw := options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset]
		content = raw[strings.IndexByte(raw, '\n')+1:]
	}

	if content == "" {
		return [][]string{}
	}

	// leadingSpaceCount is a count of UTF-16 units, as String.prototype.slice takes. -1 is upstream's
	// Number.POSITIVE_INFINITY, which slices every line to "".
	leadingSpaceCount := 0
	if node.Indent == nil {
		matches := leadingSpacePattern.FindStringSubmatch(content)
		if matches != nil {
			leadingSpaceCount = len(matches[1])
		} else {
			leadingSpaceCount = -1
		}
	} else {
		leadingSpaceCount = *node.Indent - 1 + parentIndent
	}

	split := strings.Split(content, "\n")
	rawLineContents := make([]string, len(split))
	for index, lineContent := range split {
		rawLineContents[index] = sliceUnits(lineContent, leadingSpaceCount)
	}

	if settings.proseWrap == "preserve" || node.NodeType == "blockLiteral" {
		lines := make([][]string, len(rawLineContents))
		for index, lineContent := range rawLineContents {
			if lineContent != "" {
				lines[index] = []string{lineContent}
			} else {
				lines[index] = []string{}
			}
		}
		return removeUnnecessaryTrailingNewlines(lines, node, content, isLastDescendant)
	}

	lines := [][]string{}
	for index, line := range rawLineContents {
		words := splitWithSingleSpace(line)

		if index > 0 &&
			len(words) > 0 &&
			len(rawLineContents[index-1]) > 0 &&
			!startsWithJavaScriptWhitespace(words[0]) &&
			// This test against a `string[]`, should be a mistake
			// originally introduced in https://github.com/prettier/prettier/pull/4742/files#diff-a4dc2e1922e1d8d5ac20818480f777c9a2d5af739eaa3a0409b08bf29a9d0f74R282
			//
			// The regex tests the array as a string, its words joined by ",".
			!startsOrEndsWithJavaScriptWhitespace(strings.Join(lines[len(lines)-1], ",")) {
			lines[len(lines)-1] = append(append([]string{}, lines[len(lines)-1]...), words...)
		} else {
			lines = append(lines, words)
		}
	}

	for index, originalWords := range lines {
		words := []string{}
		for _, word := range originalWords {
			// disallow trailing spaces
			if len(words) > 0 && endsWithJavaScriptWhitespace(words[len(words)-1]) {
				words[len(words)-1] += " " + word
			} else {
				words = append(words, word)
			}
		}
		lines[index] = words
	}

	if settings.proseWrap == "never" {
		for index, words := range lines {
			lines[index] = []string{strings.Join(words, " ")}
		}
	}

	return removeUnnecessaryTrailingNewlines(lines, node, content, isLastDescendant)
}

// removeUnnecessaryTrailingNewlines is the closure of that name inside getBlockValueLineContents.
func removeUnnecessaryTrailingNewlines(lineContents [][]string, node *unist.Node, content string, isLastDescendant bool) [][]string {
	if node.Chomping == "keep" {
		if strings.HasSuffix(content, "\n") && len(lineContents[len(lineContents)-1]) == 0 {
			return lineContents[:len(lineContents)-1]
		}
		return lineContents
	}

	trailingNewlineCount := 0
	for index := len(lineContents) - 1; index >= 0; index-- {
		if everyLineBlank(lineContents[index]) {
			trailingNewlineCount++
		} else {
			break
		}
	}

	switch {
	case trailingNewlineCount == 0:
		return lineContents
	case trailingNewlineCount >= 2 && !isLastDescendant:
		// next empty line
		return lineContents[:len(lineContents)-(trailingNewlineCount-1)]
	default:
		return lineContents[:len(lineContents)-trailingNewlineCount]
	}
}

// everyLineBlank is lineContents[i].every((line) => line.replace(/[ \t]+$/, "") === ""), true for none.
func everyLineBlank(words []string) bool {
	for _, word := range words {
		if trailingSpacesAndTabsPattern.ReplaceAllString(word, "") != "" {
			return false
		}
	}
	return true
}

// startsWithJavaScriptWhitespace is /^\s/.test(text).
func startsWithJavaScriptWhitespace(text string) bool {
	character, width := utf8.DecodeRuneInString(text)
	return width > 0 && isJavaScriptWhitespace(character)
}

// endsWithJavaScriptWhitespace is /\s$/.test(text), $ without the m flag being the very end.
func endsWithJavaScriptWhitespace(text string) bool {
	character, width := utf8.DecodeLastRuneInString(text)
	return width > 0 && isJavaScriptWhitespace(character)
}

// startsOrEndsWithJavaScriptWhitespace is /^\s|\s$/.test(text).
func startsOrEndsWithJavaScriptWhitespace(text string) bool {
	return startsWithJavaScriptWhitespace(text) || endsWithJavaScriptWhitespace(text)
}

// sliceUnits is text.slice(count) where count is UTF-16 units, and -1 is Infinity.
//
// A count that falls inside an astral character leaves JavaScript a lone low surrogate at the start of
// the result, which no UTF-8 string can hold. It becomes U+FFFD, what that surrogate turns into when
// the JavaScript output is encoded as UTF-8.
func sliceUnits(text string, count int) string {
	if count < 0 {
		return ""
	}
	units := 0
	for index, character := range text {
		if units >= count {
			return text[index:]
		}
		if character > 0xFFFF {
			units += 2
			if units > count {
				return "�" + text[index+utf8.RuneLen(character):]
			}
		} else {
			units++
		}
	}
	return ""
}

// isInlineNode is upstream's isInlineNode.
func isInlineNode(node *unist.Node) bool {
	if node == nil {
		return true
	}

	switch node.NodeType {
	case "plain", "quoteDouble", "quoteSingle", "alias", "flowMapping", "flowSequence":
		return true
	default:
		return false
	}
}
