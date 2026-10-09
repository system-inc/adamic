package markdown

import (
	"strconv"
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-markdown/print/list.js, printList and printListItem (the legacy pair is MDX's).

// maximumOrderedListMarker is the maximum number that can be parsed as a list item number as intended in
// CommonMark: https://spec.commonmark.org/0.31.2/#ordered-list-marker
const maximumOrderedListMarker = 999_999_999

func printList(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	nthSiblingIndex := getNthListSiblingIndex(node, parentNode(path))

	isGitDiffFriendlyOrderedList := hasGitDiffFriendlyOrderedList(node, options.OriginalText)

	minIndent := requiredIndent(path)

	return printChildren(path, options, print, func(path *astPath) (doc.Doc, bool) {
		getPrefix := func() string {
			var rawPrefix string
			if node.Ordered {
				var number int
				switch {
				case path.IsFirst():
					number = *node.Start
				case isGitDiffFriendlyOrderedList:
					number = 1
				default:
					number = min(*node.Start+pathIndex(path), maximumOrderedListMarker)
				}
				rawPrefix = strconv.Itoa(number)
				if nthSiblingIndex%2 == 0 {
					rawPrefix += ". "
				} else {
					rawPrefix += ") "
				}
			} else if nthSiblingIndex%2 == 0 {
				rawPrefix = "- "
			} else {
				rawPrefix = "* "
			}

			prefix := rawPrefix
			if node.IsAligned && node.Ordered {
				prefix = alignListPrefix(rawPrefix, settingsOf(options))
			}

			if len(prefix) >= minIndent {
				return prefix
			}

			prefix = strings.TrimRight(prefix, javaScriptSpaceText)
			trailingSpaces := min(minIndent-len(prefix), 4) // 5+ will cause indented code block
			if trailingSpaces > 0 {
				prefix += strings.Repeat(" ", trailingSpaces)
			}
			leadingSpaces := min(minIndent-len(prefix), 3) // 4+ will cause indented code block
			if leadingSpaces > 0 {
				prefix = strings.Repeat(" ", leadingSpaces) + prefix
			}

			return prefix
		}

		prefix := getPrefix()
		childNode := currentNode(path)

		if len(childNode.Children) == 2 &&
			childNode.Children[1].NodeType == "html" &&
			childNode.Children[0].Position.Start.Column != childNode.Children[1].Position.Start.Column {
			return doc.Concat{doc.Text(prefix), printListItem(path, options, print, prefix)}, true
		}

		return doc.Concat{
			doc.Text(prefix),
			doc.AlignWithString(strings.Repeat(" ", len(prefix)), printListItem(path, options, print, prefix)),
		}, true
	})
}

func printListItem(path *astPath, options *options, print printing.PrintFunc, listPrefix string) doc.Doc {
	node := currentNode(path)
	prefix := ""
	if node.Checked != nil {
		if *node.Checked {
			prefix = "[x] "
		} else {
			prefix = "[ ] "
		}
	}
	return doc.Concat{
		doc.Text(prefix),
		printChildren(path, options, print, func(path *astPath) (doc.Doc, bool) {
			child := currentNode(path)
			if (path.IsFirst() && child.NodeType != "list") || (child.NodeType == "code" && child.IsIndented) {
				return doc.AlignWithString(strings.Repeat(" ", len(prefix)), print(nil, nil)), true
			}

			alignment := strings.Repeat(" ", clamp(settingsOf(options).tabWidth-len(listPrefix), 0, 3)) // 4+ will cause indented code block
			return doc.Concat{doc.Text(alignment), doc.AlignWithString(alignment, print(nil, nil))}, true
		}),
	}
}

// requiredIndent: if the list is followed by an indented code block, its content has to be indented
// deeper than the code block.
//
// Upstream starts with `if (node.checked === null) return 0`, but node is the list, which has no
// `checked`, and undefined is not null, so the test never passes and is not ported.
func requiredIndent(path *astPath) int {
	next := nextNode(path)
	if !(next != nil && next.NodeType == "code" && next.IsIndented) {
		return 0
	}

	leadingSpaces := 0
	for _, character := range next.Value {
		if character == '\t' {
			leadingSpaces += 4
		} else if character == ' ' {
			leadingSpaces++
		} else {
			break
		}
	}
	return 4 + // base indent of the code block
		leadingSpaces +
		1 // at least one space more than the code block
}

func alignListPrefix(prefix string, settings *settings) string {
	additionalSpaces := 0
	restSpaces := len(prefix) % settings.tabWidth
	if restSpaces != 0 {
		additionalSpaces = settings.tabWidth - restSpaces
	}
	if additionalSpaces >= 4 {
		additionalSpaces = 0 // 4+ will cause indented code block
	}
	return prefix + strings.Repeat(" ", additionalSpaces)
}

func clamp(value int, minimum int, maximum int) int {
	return max(minimum, min(value, maximum))
}
