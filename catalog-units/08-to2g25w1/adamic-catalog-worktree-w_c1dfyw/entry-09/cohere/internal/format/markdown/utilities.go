package markdown

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/markdown/mdast"
)

// src/language-markdown/utilities.js.

type Node = mdast.Node

// inlineNodeTypes is INLINE_NODE_TYPES.
var inlineNodeTypes = map[string]bool{
	"liquidNode": true, "inlineCode": true, "emphasis": true, "esComment": true, "strong": true,
	"delete": true, "wikiLink": true, "link": true, "linkReference": true, "image": true,
	"imageReference": true, "footnote": true, "footnoteReference": true, "sentence": true,
	"whitespace": true, "word": true, "break": true, "inlineMath": true,
}

// inlineNodeWrapperTypes is INLINE_NODE_WRAPPER_TYPES.
var inlineNodeWrapperTypes = func() map[string]bool {
	types := map[string]bool{"tableCell": true, "paragraph": true, "heading": true}
	for nodeType := range inlineNodeTypes {
		types[nodeType] = true
	}
	return types
}()

const (
	kindNonCJK          = "non-cjk"
	kindCJLetter        = "cj-letter"
	kindKLetter         = "k-letter"
	kindCJKPunctuation  = "cjk-punctuation"
	javaScriptSpace     = `\t\n\v\f\r \x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}`
	javaScriptSpaceText = "\t\n\v\f\r \u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u2028\u2029\u202f\u205f\u3000\ufeff"
)

func inClass(ranges [][2]rune, point rune) bool {
	index := sort.Search(len(ranges), func(index int) bool { return ranges[index][1] >= point })
	return index < len(ranges) && ranges[index][0] <= point
}

// firstUnit and lastUnit are JavaScript's text[0] and text.at(-1): one UTF-16 unit, which for an astral
// character is a lone surrogate. The class tables answer for surrogates too.
//
// Both decode one character at the end they read rather than encoding the text (#vbjv3d6). Invalid UTF-8
// decodes to U+FFFD a byte at a time either way, and a backward decode stops on the same last character a
// forward one does: a valid character's first byte is never another's continuation.
func firstUnit(text string) rune {
	if text == "" {
		return -1
	}
	character, _ := utf8.DecodeRuneInString(text)
	if high, _ := utf16.EncodeRune(character); high != unicode.ReplacementChar {
		return high
	}
	return character
}

func lastUnit(text string) rune {
	if text == "" {
		return -1
	}
	character, _ := utf8.DecodeLastRuneInString(text)
	if _, low := utf16.EncodeRune(character); low != unicode.ReplacementChar {
		return low
	}
	return character
}

// isPunctuationUnit is PUNCTUATION_REGEXP.test of a one-unit string.
func isPunctuationUnit(unit rune) bool {
	return unit >= 0 && inClass(punctuationRanges, unit)
}

// containsClass is an unanchored regex test: does any code point of text belong to the class.
func containsClass(ranges [][2]rune, text string) bool {
	for _, point := range text {
		if inClass(ranges, point) {
			return true
		}
	}
	return false
}

// splitText splits text into whitespaces and words. A word is a single CJK character or a sequence of
// non-CJK characters. The nodes come from the format's arena (nil to allocate each).
func splitText(text string, nodes *arena.Arena[Node]) []*Node {
	var split []*Node

	appendNode := func(node *Node) {
		var lastNode *Node
		if len(split) > 0 {
			lastNode = split[len(split)-1]
		}
		isBetween := func(kind1 string, kind2 string) bool {
			return (lastNode.Kind == kind1 && node.Kind == kind2) || (lastNode.Kind == kind2 && node.Kind == kind1)
		}
		if lastNode != nil && lastNode.NodeType == "word" &&
			!isBetween(kindNonCJK, kindCJKPunctuation) &&
			// disallow leading/trailing full-width whitespace
			!strings.ContainsRune(lastNode.Value, '\u3000') && !strings.ContainsRune(node.Value, '\u3000') {
			split = append(split, nodes.New(Node{NodeType: "whitespace", Value: "", IsLiteral: true}))
		}
		split = append(split, node)
	}

	tokens := splitKeepingSeparators(text, func(character rune) bool {
		return character == '\t' || character == '\n' || character == ' '
	})
	// innerTokens is reused for every token: its pieces are read only within the token's iteration.
	var innerTokens []string
	for index, token := range tokens {
		// whitespace
		if index%2 == 1 {
			value := " "
			if strings.Contains(token, "\n") {
				value = "\n"
			}
			split = append(split, nodes.New(Node{NodeType: "whitespace", Value: value, IsLiteral: true}))
			continue
		}

		// word separated by whitespace

		if (index == 0 || index == len(tokens)-1) && token == "" {
			continue
		}

		innerTokens = appendSplitOnCJK(innerTokens[:0], token)
		for innerIndex, innerToken := range innerTokens {
			if (innerIndex == 0 || innerIndex == len(innerTokens)-1) && innerToken == "" {
				continue
			}

			// non-CJK word
			if innerIndex%2 == 0 {
				if innerToken != "" {
					appendNode(nodes.New(Node{
						NodeType:               "word",
						Value:                  innerToken,
						IsLiteral:              true,
						Kind:                   kindNonCJK,
						IsCJ:                   false,
						HasLeadingPunctuation:  isPunctuationUnit(firstUnit(innerToken)),
						HasTrailingPunctuation: isPunctuationUnit(lastUnit(innerToken)),
					}))
				}
				continue
			}

			// CJK character

			// punctuation for CJ(K). Korean doesn't use them in horizontal writing usually.
			if containsClass(punctuationRanges, innerToken) {
				appendNode(nodes.New(Node{
					NodeType: "word", Value: innerToken, IsLiteral: true, Kind: kindCJKPunctuation, IsCJ: true,
					HasLeadingPunctuation: true, HasTrailingPunctuation: true,
				}))
				continue
			}

			// Korean uses space to divide words, but Chinese & Japanese do not. This is why Korean should
			// be treated like non-CJK.
			if containsClass(hangulRanges, innerToken) {
				appendNode(nodes.New(Node{NodeType: "word", Value: innerToken, IsLiteral: true, Kind: kindKLetter, IsCJ: false}))
				continue
			}

			appendNode(nodes.New(Node{NodeType: "word", Value: innerToken, IsLiteral: true, Kind: kindCJLetter, IsCJ: true}))
		}
	}

	return split
}

// splitKeepingSeparators is text.split(/(<separator>+)/): pieces alternate with the separator runs that
// divided them, starting and ending with a piece (empty when the text starts or ends with a run).
func splitKeepingSeparators(text string, isSeparator func(rune) bool) []string {
	var parts []string
	start := 0
	index := 0
	for index < len(text) {
		character, size := utf8.DecodeRuneInString(text[index:])
		if !isSeparator(character) {
			index += size
			continue
		}
		runStart := index
		for index < len(text) {
			character, size = utf8.DecodeRuneInString(text[index:])
			if !isSeparator(character) {
				break
			}
			index += size
		}
		parts = append(parts, text[start:runStart], text[runStart:index])
		start = index
	}
	return append(parts, text[start:])
}

// appendSplitOnCJK appends token.split(new RegExp(`(${CJK_REGEXP.source})`, "u")) to parts: pieces
// alternate with the CJK matches, each one CJK character and an optional variation selector after it.
func appendSplitOnCJK(parts []string, token string) []string {
	start := 0
	index := 0
	for index < len(token) {
		character, size := utf8.DecodeRuneInString(token[index:])
		if !inClass(cjkRanges, character) {
			index += size
			continue
		}
		matchStart := index
		index += size
		if index < len(token) {
			next, nextSize := utf8.DecodeRuneInString(token[index:])
			if inClass(variationSelectorRanges, next) {
				index += nextSize
			}
		}
		parts = append(parts, token[start:matchStart], token[matchStart:index])
		start = index
	}
	return append(parts, token[start:])
}

var orderedListItemPattern = regexp.MustCompile(`^[` + javaScriptSpace + `]*(\d+)(\.|\))([` + javaScriptSpace + `]*)`)

// orderedListItemInfo is getOrderedListItemInfo: the item's number and the spaces after its marker.
type orderedListItemInfo struct {
	number        int
	leadingSpaces string
}

func getOrderedListItemInfo(orderListItem *Node, originalText string) orderedListItemInfo {
	text := originalText[orderListItem.Position.Start.Offset:orderListItem.Position.End.Offset]
	if len(orderListItem.Children) > 0 {
		firstChild := orderListItem.Children[0]
		text = originalText[orderListItem.Position.Start.Offset:firstChild.Position.Start.Offset]
	}

	match := orderedListItemPattern.FindStringSubmatch(text)
	if match == nil {
		panic("markdown: an ordered list item without a number")
	}
	number, _ := strconv.ParseFloat(match[1], 64)
	return orderedListItemInfo{number: int(number), leadingSpaces: match[3]}
}

func hasGitDiffFriendlyOrderedList(node *Node, originalText string) bool {
	if !node.Ordered || len(node.Children) < 2 {
		return false
	}

	secondNumber := getOrderedListItemInfo(node.Children[1], originalText).number

	if secondNumber != 1 {
		return false
	}

	firstNumber := getOrderedListItemInfo(node.Children[0], originalText).number

	if firstNumber != 0 {
		return true
	}

	return len(node.Children) > 2 && getOrderedListItemInfo(node.Children[2], originalText).number == 1
}

// mapAst is upstream's mapAst: a preorder walk that replaces each node with what the handler returns and
// each parent's children with their replacements, the parent stack holding the replacements.
//
// Upstream spreads each handled node into a new object and each parent's mapped children into a new
// array. Here neither is copied: no caller holds the tree a pass was given (preprocess replaces it with
// each pass's result), so the copies would only be garbage, the largest share of a markdown format's
// allocation (#93dpede). A parent's children are replaced in place, left to right, so a handler that read
// an earlier sibling through its parent would see the replacement where upstream sees the original. None
// of preprocess's handlers does: they read the next sibling, which is not yet mapped, and the count.
//
// The parent stacks are not copied per parent either (#vbjv3d6). The stack for a node at one depth is
// built in that depth's buffer, which its siblings reuse once its subtree is done, and a handler reads its
// stack only while it runs.
func mapAst(ast *Node, handler func(node *Node, index int, parentStack []*Node) *Node) *Node {
	var stacks [][]*Node
	var preorder func(node *Node, index int, parentStack []*Node) *Node
	preorder = func(node *Node, index int, parentStack []*Node) *Node {
		newNode := handler(node, index, parentStack)
		if newNode.Children != nil {
			depth := len(parentStack)
			if depth == len(stacks) {
				stacks = append(stacks, make([]*Node, 0, depth+1))
			}
			stack := append(append(stacks[depth][:0], newNode), parentStack...)
			stacks[depth] = stack
			for childIndex, child := range newNode.Children {
				newNode.Children[childIndex] = preorder(child, childIndex, stack)
			}
		}
		return newNode
	}
	return preorder(ast, -1, nil)
}

func isAutolink(node *Node) bool {
	if node == nil || node.NodeType != "link" || len(node.Children) != 1 {
		return false
	}

	child := node.Children[0]
	return locStart(node) == locStart(child) && locEnd(node) == locEnd(child)
}

var prettierIgnorePattern = regexp.MustCompile(`(?s)^<!--[` + javaScriptSpace + `]*prettier-ignore(?:-(start|end))?[` + javaScriptSpace + `]*-->$`)

// isPrettierIgnore returns "", "next", "start" or "end". esComment is MDX's and never occurs here.
func isPrettierIgnore(node *Node) string {
	if node == nil || node.NodeType != "html" {
		return ""
	}
	match := prettierIgnorePattern.FindStringSubmatch(node.Value)
	if match == nil {
		return ""
	}
	if match[1] != "" {
		return match[1]
	}
	return "next"
}

func getNthListSiblingIndex(node *Node, parentNode *Node) int {
	index := -1

	for _, childNode := range parentNode.Children {
		if childNode.NodeType == node.NodeType && childNode.Ordered == node.Ordered {
			index++
		} else {
			index = -1
		}

		if childNode == node {
			return index
		}
	}

	return index
}

func isSetextHeading(node *Node) bool {
	return node.Position.Start.Line != node.Position.End.Line
}

func isNewLine(node *Node) bool {
	return node != nil && node.NodeType == "whitespace" && node.Value == "\n"
}

func locStart(node *Node) int { return node.Position.Start.Offset }
func locEnd(node *Node) int   { return node.Position.End.Offset }

// utf16Length is JavaScript's String.length, counted without encoding the text (#vbjv3d6).
func utf16Length(text string) int {
	length := 0
	for _, character := range text {
		length += utf16.RuneLen(character)
	}
	return length
}
