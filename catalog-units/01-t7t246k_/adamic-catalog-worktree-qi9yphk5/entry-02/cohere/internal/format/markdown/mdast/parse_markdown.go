package mdast

import (
	"strings"
	"unicode/utf16"

	"github.com/system-inc/cohere/internal/format/arena"
)

// ParseMarkdown is the fork's parseMarkdown, src/language-markdown/parse/parse-markdown.js: front matter
// is found and blanked, the rest goes through micromark and mdast-util-from-markdown, and the front matter
// comes back as the root's first child.
//
// text is what Prettier's core hands the parser: line endings normalized to \n and a byte order mark
// removed.
func ParseMarkdown(text string, nodes *arena.Arena[Node]) (*Node, error) {
	frontMatter, content := ParseFrontMatter(text)
	root, err := FromMarkdown(content, text, nodes)
	if err != nil {
		return nil, err
	}

	if frontMatter != nil {
		lines := strings.Split(frontMatter.Raw, "\n")
		node := nodes.New(Node{
			NodeType:    "frontMatter",
			FrontMatter: frontMatter,
			Position: &Position{
				Start: Point{Line: 1, Column: 1, Offset: 0},
				// end: {line: lines.length, column: lines.at(-1).length + 1, offset: raw.length}, where
				// the length is UTF-16 and the offset converts to bytes like every other.
				End: Point{
					Line:   len(lines),
					Column: len(utf16.Encode([]rune(lines[len(lines)-1]))) + 1,
					Offset: len(frontMatter.Raw),
				},
			},
		})
		root.Children = append([]*Node{node}, root.Children...)
	}

	return root, nil
}

// ParseFrontMatter is the fork's src/main/front-matter/parse.js: the front matter, if any, and the text
// with it replaced by spaces so that every later offset is unchanged.
func ParseFrontMatter(text string) (*FrontMatter, string) {
	frontMatter := getFrontMatter(text)
	if frontMatter == nil {
		return nil, text
	}
	return frontMatter, replaceNonLineBreaksWithSpace(frontMatter.Raw) + text[len(frontMatter.Raw):]
}

const delimiterLength = 3

func getFrontMatter(text string) *FrontMatter {
	if len(text) < delimiterLength {
		return nil
	}
	startDelimiter := text[:delimiterLength]

	if startDelimiter != "---" && startDelimiter != "+++" {
		return nil
	}

	firstLineBreakIndex := strings.Index(text[delimiterLength:], "\n")
	if firstLineBreakIndex == -1 {
		return nil
	}
	firstLineBreakIndex += delimiterLength

	explicitLanguage := trimJavaScriptWhitespace(text[delimiterLength:firstLineBreakIndex])

	endDelimiterIndex := indexFrom(text, "\n"+startDelimiter, firstLineBreakIndex)

	language := explicitLanguage
	if language == "" {
		if startDelimiter == "+++" {
			language = "toml"
		} else {
			language = "yaml"
		}
	}

	if endDelimiterIndex == -1 && startDelimiter == "---" && language == "yaml" {
		// In some markdown processors such as pandoc, "..." can be used as the end delimiter for YAML
		// front-matter.
		endDelimiterIndex = indexFrom(text, "\n...", firstLineBreakIndex)
	}

	if endDelimiterIndex == -1 {
		return nil
	}

	frontMatterEndIndex := endDelimiterIndex + 1 + delimiterLength

	// Upstream tests the next character against /\s?/, which every string satisfies, so nothing is
	// refused here.

	raw := text[:frontMatterEndIndex]
	// text.slice(firstLineBreakIndex + 1, endDelimiterIndex) is empty, not an error, when the end
	// delimiter starts on the opening's line break (`---\n---`).
	value := ""
	if firstLineBreakIndex+1 < endDelimiterIndex {
		value = text[firstLineBreakIndex+1 : endDelimiterIndex]
	}
	frontMatter := &FrontMatter{
		Language:       language,
		Value:          value,
		StartDelimiter: startDelimiter,
		EndDelimiter:   raw[len(raw)-delimiterLength:],
		Raw:            raw,
	}
	if explicitLanguage != "" {
		frontMatter.ExplicitLanguage = &explicitLanguage
	}
	return frontMatter
}

func indexFrom(text string, search string, from int) int {
	if from > len(text) {
		return -1
	}
	index := strings.Index(text[from:], search)
	if index < 0 {
		return -1
	}
	return index + from
}

// replaceNonLineBreaksWithSpace replaces every UTF-16 unit but \n with a space, as
// `string.replaceAll(/[^\n]/g, " ")` does without the u flag: an astral character becomes two spaces, so
// the UTF-16 length, and every offset after it, is unchanged.
func replaceNonLineBreaksWithSpace(text string) string {
	var builder strings.Builder
	for _, character := range text {
		switch {
		case character == '\n':
			builder.WriteByte('\n')
		case character >= 0x10000:
			builder.WriteString("  ")
		default:
			builder.WriteByte(' ')
		}
	}
	return builder.String()
}

// trimJavaScriptWhitespace is String#trim, whose whitespace is ECMAScript's WhiteSpace and
// LineTerminator: Go's TrimSpace differs on U+FEFF (JavaScript trims it) and U+0085 (Go trims it).
func trimJavaScriptWhitespace(text string) string {
	return strings.TrimFunc(text, isJavaScriptWhitespace)
}

func isJavaScriptWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x00A0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return character >= 0x2000 && character <= 0x200A
}
