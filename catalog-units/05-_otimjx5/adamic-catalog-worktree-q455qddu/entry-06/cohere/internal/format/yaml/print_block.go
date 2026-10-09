package yaml

import (
	"strconv"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-yaml/print/block.js.

// printBlock is upstream's printBlock: a block literal (|) or folded (>) scalar.
func printBlock(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	settings := settingsOf(options)
	parentIndent := 0
	for _, ancestor := range path.Ancestors() {
		if ancestor.NodeType == "sequence" || ancestor.NodeType == "mapping" {
			parentIndent++
		}
	}
	isLastDescendant := isLastDescendantNode(path)
	parts := doc.Concat{}
	if node.NodeType == "blockFolded" {
		parts = append(parts, doc.Text(">"))
	} else {
		parts = append(parts, doc.Text("|"))
	}
	if node.Indent != nil {
		parts = append(parts, doc.Text(strconv.Itoa(*node.Indent)))
	}

	if node.Chomping != "clip" {
		if node.Chomping == "keep" {
			parts = append(parts, doc.Text("+"))
		} else {
			parts = append(parts, doc.Text("-"))
		}
	}

	if hasIndicatorComment(node) {
		parts = append(parts, doc.Text(" "), print("indicatorComment", nil))
	}

	lineContents := getBlockValueLineContents(node, parentIndent, isLastDescendant, options, settings)
	contentsParts := doc.Concat{}
	for index, lineWords := range lineContents {
		if index == 0 {
			contentsParts = append(contentsParts, doc.Hardline)
		}
		contentsParts = append(contentsParts, fillWords(lineWords))
		if index != len(lineContents)-1 {
			if len(lineWords) == 0 {
				contentsParts = append(contentsParts, doc.Hardline)
			} else {
				contentsParts = append(contentsParts, doc.MarkAsRoot(doc.Literalline))
			}
		} else if node.Chomping == "keep" && isLastDescendant {
			if len(lineWords) == 0 {
				contentsParts = append(contentsParts, doc.DedentToRoot(doc.Hardline))
			} else {
				contentsParts = append(contentsParts, doc.DedentToRoot(doc.Literalline))
			}
		}
	}
	if node.Indent == nil {
		parts = append(parts, doc.Dedent(alignWithSpaces(settings.tabWidth, contentsParts)))
	} else {
		parts = append(parts, doc.DedentToRoot(alignWithSpaces(*node.Indent-1+parentIndent, contentsParts)))
	}

	return parts
}
