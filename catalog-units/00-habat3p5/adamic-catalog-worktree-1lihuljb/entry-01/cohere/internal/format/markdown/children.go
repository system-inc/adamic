package markdown

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-markdown/print/children.js, for the markdown parser.

// processor is upstream's events.processor: the doc for the child at path, or false to skip it.
type processor func(path *astPath) (doc.Doc, bool)

func printChildren(path *astPath, options *options, print printing.PrintFunc, process processor) doc.Doc {
	if process == nil {
		process = func(*astPath) (doc.Doc, bool) { return print(nil, nil), true }
	}

	parts := doc.Concat{}

	path.Each(func(path *astPath, _ int, _ any) {
		result, keep := process(path)
		if keep {
			if len(parts) > 0 && shouldPrePrintHardline(path) {
				parts = append(parts, doc.Hardline)

				if shouldPrePrintDoubleHardline(path, options) {
					parts = append(parts, doc.Hardline)
				}
			}

			parts = append(parts, result)
		}
	}, "children")

	return parts
}

func shouldPrePrintHardline(path *astPath) bool {
	node, parent := currentNode(path), parentNode(path)
	isInlineNode := inlineNodeTypes[node.NodeType]

	isInlineHTML := node.NodeType == "html" && inlineNodeWrapperTypes[parent.NodeType]

	return !isInlineNode && !isInlineHTML
}

var siblingNodeTypes = map[string]bool{"listItem": true, "definition": true}

func shouldPrePrintDoubleHardline(path *astPath, options *options) bool {
	node, previous, parent := currentNode(path), previousNode(path), parentNode(path)

	if isSetextHeading(node) && node.Position.Start.Line < previous.Position.End.Line {
		return false
	}

	if isPreviousNodeLooseListItem(path) ||
		(node.NodeType == "list" &&
			parent.NodeType == "listItem" &&
			(previous.NodeType == "code" ||
				// Preserve blank line before nested list within listItem (issue #17746)
				previous.NodeType == "paragraph") &&
			previous.Position.End.Line+1 < node.Position.Start.Line) {
		return true
	}

	isSequence := previous.NodeType == node.NodeType
	isSiblingNode := isSequence && siblingNodeTypes[node.NodeType]
	isInTightListItem := parent.NodeType == "listItem" &&
		(node.NodeType == "list" || !printing.CallParent(path, isLooseListItemAt, 0))
	isPrevNodePrettierIgnore := isPrettierIgnore(previous) == "next"
	isBlockHTMLWithoutBlankLineBetweenPrevHTML := node.NodeType == "html" &&
		previous.NodeType == "html" &&
		previous.Position.End.Line+1 == node.Position.Start.Line
	isBlockHTMLWithoutBlankLineBetweenPrevParagraph := node.NodeType == "html" &&
		previous.NodeType == "paragraph" &&
		previous.Position.End.Line+1 == node.Position.Start.Line
	isHTMLDirectAfterListItem := node.NodeType == "html" &&
		parent.NodeType == "listItem" &&
		previous.NodeType == "paragraph" &&
		previous.Position.End.Line+1 == node.Position.Start.Line

	return !(isSiblingNode ||
		isInTightListItem ||
		isPrevNodePrettierIgnore ||
		isBlockHTMLWithoutBlankLineBetweenPrevHTML ||
		isBlockHTMLWithoutBlankLineBetweenPrevParagraph ||
		isHTMLDirectAfterListItem)
}

func isLooseListItem(node *Node, parent *Node, next *Node) bool {
	return node.NodeType == "listItem" &&
		(node.Spread ||
			(parent != nil && parent.NodeType == "list" &&
				next != nil && next.NodeType == "listItem" &&
				node.Position.End.Line+1 < next.Position.Start.Line))
}

// isLooseListItemAt is isLooseListItem called with a path, as upstream's callParent passes one.
func isLooseListItemAt(path *astPath) bool {
	return isLooseListItem(currentNode(path), parentNode(path), nextNode(path))
}

func isPreviousNodeLooseListItem(path *astPath) bool {
	if pathIndex(path) == 0 {
		return false
	}
	return isLooseListItem(previousNode(path), parentNode(path), currentNode(path))
}
