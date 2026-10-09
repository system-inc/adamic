package markdown

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
)

// src/language-markdown/print/whitespace.js.

var singleLineNodeTypes = map[string]bool{"tableCell": true, "link": true, "wikiLink": true}

// lineBreakBetweenTheseAndCJConvertsToSpace: a line break between a character from this set and CJ can
// be converted to a space. Includes only ASCII punctuation marks for now.
const lineBreakBetweenTheseAndCJConvertsToSpace = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

// convertsToSpaceAroundCJ is the Set's has() of one UTF-16 unit; an empty value's unit is undefined.
func convertsToSpaceAroundCJ(unit rune) bool {
	return unit >= 0 && unit < 0x80 && strings.ContainsRune(lineBreakBetweenTheseAndCJConvertsToSpace, unit)
}

// isInSentenceWithCJSpaces determines the preferred style of spacing between Chinese or Japanese and
// non-CJK characters in the parent sentence node: true if a space tends to be inserted between them.
func isInSentenceWithCJSpaces(path *astPath) bool {
	sentenceNode := parentNode(path)
	if sentenceNode.UsesCJSpaces == nil {
		spaces, empties := 0, 0
		children := sentenceNode.Children

		for index := 1; index < len(children)-1; index++ {
			node := children[index]
			if node.NodeType == "whitespace" && (node.Value == " " || node.Value == "") {
				previousKind := children[index-1].Kind
				nextKind := children[index+1].Kind
				if (previousKind == kindCJLetter && nextKind == kindNonCJK) ||
					(previousKind == kindNonCJK && nextKind == kindCJLetter) {
					if node.Value == " " {
						spaces++
					} else {
						empties++
					}
				}
			}
		}

		// Inject a property to cache the result.
		usesCJSpaces := spaces > empties
		sentenceNode.UsesCJSpaces = &usesCJSpaces
	}

	return *sentenceNode.UsesCJSpaces
}

// lineBreakCanBeConvertedToSpace checks whether the given "\n" node can be converted to a space.
// Chinese and Japanese don't use U+0020 Space to divide words, so line breaks shouldn't be replaced with
// spaces for those languages.
func lineBreakCanBeConvertedToSpace(path *astPath, isLink bool) bool {
	if isLink {
		return true
	}

	previous, next := previousNode(path), nextNode(path)

	// e.g. " \nletter"
	if previous == nil || next == nil {
		return true
	}

	previousKind := previous.Kind
	nextKind := next.Kind

	if
	// "\n" between non-CJK or Korean characters always can be converted to a space. Korean Hangul
	// simulates Latin words. See https://github.com/prettier/prettier/issues/6516
	(isNonCJKOrKoreanLetter(previousKind) && isNonCJKOrKoreanLetter(nextKind)) ||
		// Han & Hangul: same way preferred
		(previousKind == kindKLetter && nextKind == kindCJLetter) ||
		(nextKind == kindKLetter && previousKind == kindCJLetter) {
		return true
	}

	// Do not convert \n to a space:
	if
	// around CJK punctuation
	previousKind == kindCJKPunctuation ||
		nextKind == kindCJKPunctuation ||
		// between CJ
		(previousKind == kindCJLetter && nextKind == kindCJLetter) {
		return false
	}

	// The rest of this function deals only with line breaks between CJ and non-CJK characters.

	// Convert a line break between CJ and certain non-letter characters (e.g. ASCII punctuation) to a
	// space. E.g. :::\n句子句子句子\n::: → ::: 句子句子句子 :::
	//
	// Note: line breaks like "(\n句子句子\n)" or "句子\n." are suppressed in `isBreakable(...)`.
	if convertsToSpaceAroundCJ(firstUnit(next.Value)) || convertsToSpaceAroundCJ(lastUnit(previous.Value)) {
		return true
	}

	// Converting a line break between CJ and non-ASCII punctuation to a space is undesired in many
	// cases, e.g. "〜" (U+301C) in "ア〜\nエの中から1つ選べ。" or "…" (U+2026) in "これはひどい……\nなんと汚いコミットログなんだ……".
	if previous.HasTrailingPunctuation || next.HasLeadingPunctuation {
		return false
	}

	// If the sentence uses the style with spaces between CJ and non-CJK, "\n" can be converted to a
	// space.
	return isInSentenceWithCJSpaces(path)
}

// isNonCJKOrKoreanLetter is true if kind is Korean letter or non-CJK.
func isNonCJKOrKoreanLetter(kind string) bool {
	return kind == kindNonCJK || kind == kindKLetter
}

// isBreakable checks whether whitespace can be printed as a line break.
func isBreakable(path *astPath, value string, proseWrap string, isLink bool) bool {
	if proseWrap != "always" ||
		path.HasAncestor(func(node *Node) bool {
			// options.parser is never "mdx" here.
			return singleLineNodeTypes[node.NodeType] || (node.NodeType == "heading" && !isSetextHeading(node))
		}) {
		return false
	}

	if isLink {
		return value != ""
	}

	previous, next := previousNode(path), nextNode(path)

	// [1]: We will make a breaking change to the rule to convert spaces between a Chinese or Japanese
	//      character and another character in the future. Such a space must have been always
	//      interchangeable with a line break.
	// [2]: we should not break lines even between Chinese/Japanese characters because Chrome & Safari
	//      replaces "\n" between such characters with " " now.
	// [3]: Hangul (Korean) must simulate Latin words; see https://github.com/prettier/prettier/issues/6516

	if previous == nil || next == nil {
		// empty side is Latin ASCII symbol (e.g. *, [, ], or `)
		// value is " " or "\n" (not "")
		return true
	}

	if value == "" {
		// [1] & [2] & [3]
		// At least either of previous or next is non-Latin (=CJK)
		return false
	}

	if (previous.Kind == kindKLetter && next.Kind == kindCJLetter) ||
		(next.Kind == kindKLetter && previous.Kind == kindCJLetter) {
		return true
	}

	// [1] & [2]
	if previous.IsCJ || next.IsCJ {
		return false
	}

	return true
}

// printWhitespace prints a whitespace node's value. isLink is a special mode of (un)wrapping that
// preserves the normalized form of link labels (https://spec.commonmark.org/0.30/#matches).
func printWhitespace(path *astPath, value string, proseWrap string, isLink bool, _ *options) doc.Doc {
	if proseWrap == "preserve" && value == "\n" {
		return doc.Hardline
	}

	canBeSpace := value == " " || (value == "\n" && lineBreakCanBeConvertedToSpace(path, isLink))

	if isBreakable(path, value, proseWrap, isLink) {
		if canBeSpace {
			return doc.LineDoc
		}
		return doc.Softline
	}

	if canBeSpace {
		return doc.Text(" ")
	}
	return doc.Text("")
}
