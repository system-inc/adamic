package markdown

import (
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/markdown/mdast"
)

// preprocess is src/language-markdown/print/preprocess.js for the markdown parser (the MDX branches are
// not ported: no corpus holds .mdx).
func preprocess(ast *Node, originalText string, tabWidth int, nodes *arena.Arena[Node]) *Node {
	ast = addRawToText(ast, originalText)
	ast = mergeContinuousTexts(ast, nodes)
	ast = transformIndentedCodeblock(ast, originalText)
	ast = markOriginalImageAndLinkAlt(ast, originalText)
	ast = markAlignedList(ast, originalText, tabWidth)
	ast = splitTextIntoSentences(ast, nodes)
	return ast
}

func addRawToText(ast *Node, originalText string) *Node {
	return mapAst(ast, func(node *Node, _ int, _ []*Node) *Node {
		if node.NodeType == "text" {
			// https://github.com/remarkjs/remark-gfm/issues/16
			raw := originalText[node.Position.Start.Offset:node.Position.End.Offset]
			node.Raw = &raw
		}
		return node
	})
}

func mergeChildren(ast *Node, shouldMerge func(previous *Node, node *Node) bool, mergeNode func(previous *Node, node *Node) *Node) *Node {
	return mapAst(ast, func(node *Node, _ int, _ []*Node) *Node {
		if node.Children == nil {
			return node
		}

		var children []*Node
		var lastChild *Node
		changed := false
		for _, child := range node.Children {
			if lastChild != nil && shouldMerge(lastChild, child) {
				child = mergeNode(lastChild, child)
				// Replace the previous node
				children[len(children)-1] = child
				changed = true
			} else {
				children = append(children, child)
			}

			lastChild = child
		}

		if !changed {
			return node
		}
		copied := *node
		copied.Children = children
		return &copied
	})
}

func mergeContinuousTexts(ast *Node, nodes *arena.Arena[Node]) *Node {
	return mergeChildren(ast,
		func(previous *Node, node *Node) bool { return previous.NodeType == "text" && node.NodeType == "text" },
		func(previous *Node, node *Node) *Node {
			// The merged node has no `raw`, as upstream's has none.
			return nodes.New(Node{
				NodeType:  "text",
				IsLiteral: true,
				Value:     previous.Value + node.Value,
				Position:  &mdast.Position{Start: previous.Position.Start, End: node.Position.End},
			})
		})
}

// htmlWhitespace is Prettier's html-whitespace utility: https://infra.spec.whatwg.org/#ascii-whitespace.
const htmlWhitespace = "\t\n\f\r "

func splitTextIntoSentences(ast *Node, nodes *arena.Arena[Node]) *Node {
	canOpenAccidentalWikiLink := map[*Node]bool{}
	// We can't use nodes themselves because they will be cloned.
	riskyParagraphPositions := map[*mdast.Position]bool{}

	markAncestors := func(parentStack []*Node) {
		for _, ancestor := range parentStack {
			if ancestor.NodeType == "paragraph" && canOpenAccidentalWikiLink[ancestor] {
				riskyParagraphPositions[ancestor.Position] = true
			}
		}
	}

	var walkAst func(node *Node, parentStack []*Node)
	walkAst = func(node *Node, parentStack []*Node) {
		if node.NodeType == "wikiLink" {
			// word wrapping can accidentally merge nodes like `[[foo\n[[wiki link]]`
			markAncestors(parentStack)
		} else if node.NodeType == "text" {
			raw := rawOf(node)
			if strings.Contains(raw, "[[") {
				for _, ancestor := range parentStack {
					if ancestor.NodeType == "paragraph" {
						canOpenAccidentalWikiLink[ancestor] = true
					}
				}
			}

			if strings.Contains(raw, "]]") {
				markAncestors(parentStack)
			}
		}
		if node.Children != nil {
			stack := append([]*Node{node}, parentStack...)
			for _, child := range node.Children {
				walkAst(child, stack)
			}
		}
	}
	walkAst(ast, nil)

	return mapAst(ast, func(node *Node, index int, parentStack []*Node) *Node {
		if node.NodeType != "text" {
			return node
		}

		// NOTE: there's trade-off between using `node.value` or `node.raw` here. Using `node.raw`, we can
		// better preserve the original text, especially escaped characters. Using `node.value`, we don't
		// need to care about markers like `\n > ` in `blockquote`s.
		text := rawOf(node)

		paragraphIndex := -1
		for stackIndex, ancestor := range parentStack {
			if ancestor != nil && ancestor.NodeType == "paragraph" {
				paragraphIndex = stackIndex
				break
			}
		}

		var paragraphNode *Node
		if paragraphIndex != -1 {
			paragraphNode = parentStack[paragraphIndex]
		}

		if paragraphNode != nil {
			for _, ancestor := range parentStack[paragraphIndex+1:] {
				if ancestor != nil && ancestor.NodeType == "blockquote" {
					text = getBlockquoteRawText(text, node)
					break
				}
			}

			parentNode := parentStack[0]

			if parentNode != nil && parentNode.NodeType == "paragraph" {
				if index == 0 {
					text = strings.TrimLeft(text, htmlWhitespace)
				}
				if index == len(parentNode.Children)-1 {
					text = strings.TrimRight(text, htmlWhitespace)
				}
			}
		}

		if paragraphNode != nil && riskyParagraphPositions[paragraphNode.Position] {
			return nodes.New(Node{NodeType: "text", IsLiteral: true, Position: node.Position, Value: text})
		}

		return nodes.New(Node{NodeType: "sentence", Position: node.Position, IsParent: true, Children: nonNil(splitText(text, nodes))})
	})
}

// rawOf reads node.raw, which upstream's merged text nodes lack: reading it there throws a TypeError, so
// a file that reaches it is a file Prettier cannot format.
func rawOf(node *Node) string {
	if node.Raw == nil {
		panic("markdown: Cannot read properties of undefined (reading 'includes')")
	}
	return *node.Raw
}

func nonNil(nodes []*Node) []*Node {
	if nodes == nil {
		return []*Node{}
	}
	return nodes
}

var angleBracketsPattern = regexp.MustCompile(`^([ \t]*>[ \t]*)*`)

func getBlockquoteRawText(text string, node *Node) string {
	rawLines := strings.Split(text, "\n")
	valueLines := strings.Split(node.Value, "\n")
	resultLines := make([]string, len(rawLines))
	for index, rawLine := range rawLines {
		valueLine := ""
		if index < len(valueLines) {
			valueLine = valueLines[index]
		}
		leadingTextAngleBrackets := angleBracketsPattern.FindString(valueLine)
		// rawLine.replace(regex, string): the replacement string's `$` patterns would be expanded, and
		// brackets and spaces contain none.
		resultLines[index] = leadingTextAngleBrackets + rawLine[len(angleBracketsPattern.FindString(rawLine)):]
	}
	return strings.Join(resultLines, "\n")
}

var indentedCodePattern = regexp.MustCompile(`^\n?(?: {4,}|\t)`)

func transformIndentedCodeblock(ast *Node, originalText string) *Node {
	return mapAst(ast, func(node *Node, _ int, _ []*Node) *Node {
		if node.NodeType != "code" {
			return node
		}
		// the first char may point to `\n`, e.g. `\n\t\tbar`, just ignore it
		node.IsIndented = indentedCodePattern.MatchString(originalText[node.Position.Start.Offset:node.Position.End.Offset])
		return node
	})
}

// markOriginalImageAndLinkAlt: remark 11 removes nested links so we need to recover the original alt
// text. Upstream also stashes `originalLabelText` on links, which nothing reads, so it is not kept.
func markOriginalImageAndLinkAlt(ast *Node, originalText string) *Node {
	return mapAst(ast, func(node *Node, _ int, _ []*Node) *Node {
		if node.NodeType == "image" || node.NodeType == "imageReference" {
			node.OriginalAltText = getBracketContent(originalText, node.Position.Start.Offset, node.Position.End.Offset)
		}
		return node
	})
}

// getBracketContent walks bytes where upstream walks UTF-16 units. Every character it compares is ASCII,
// and a skipped escape lands inside a multi-byte character only on continuation bytes, which are never
// ASCII, so the two walks agree.
func getBracketContent(text string, startOffset int, endOffset int) *string {
	firstBracket := strings.Index(text[startOffset:], "[")
	if firstBracket != -1 {
		firstBracket += startOffset
	}

	if firstBracket == -1 || firstBracket >= endOffset {
		return nil
	}

	depth := 1
	index := firstBracket + 1

	for index < endOffset {
		character := text[index]

		if character == '\\' {
			index += 2
			continue
		}

		if character == '[' {
			depth++
		} else if character == ']' {
			depth--
			if depth == 0 {
				content := text[firstBracket+1 : index]
				return &content
			}
		}

		index++
	}

	return nil
}

func markAlignedList(ast *Node, originalText string, tabWidth int) *Node {
	getListItemStart := func(listItem *Node) int {
		if len(listItem.Children) == 0 {
			return -1
		}
		return listItem.Children[0].Position.Start.Column - 1
	}

	isAligned := func(list *Node) bool {
		if !list.Ordered {
			// - 123
			// - 123
			return true
		}

		firstItem := list.Children[0]
		var secondItem *Node
		if len(list.Children) > 1 {
			secondItem = list.Children[1]
		}

		firstInfo := getOrderedListItemInfo(firstItem, originalText)

		if utf16Length(firstInfo.leadingSpaces) > 1 {
			// 1.   123
			//
			// 1.   123
			// 1. 123
			return true
		}

		firstStart := getListItemStart(firstItem)

		if firstStart == -1 {
			// 1.
			//
			// 1.
			// 1.
			return false
		}

		if len(list.Children) == 1 {
			// aligned:
			//
			// 11. 123
			//
			// not aligned:
			//
			// 1. 123
			return firstStart%tabWidth == 0
		}

		secondStart := getListItemStart(secondItem)

		if firstStart != secondStart {
			// 11. 123
			// 1. 123
			//
			// 1. 123
			// 11. 123
			return false
		}

		if firstStart%tabWidth == 0 {
			// 11. 123
			// 12. 123
			return true
		}

		// aligned:
		//
		// 11. 123
		// 1.  123
		//
		// not aligned:
		//
		// 1. 123
		// 2. 123
		secondInfo := getOrderedListItemInfo(secondItem, originalText)
		return utf16Length(secondInfo.leadingSpaces) > 1
	}

	return mapAst(ast, func(node *Node, index int, parentStack []*Node) *Node {
		if node.NodeType == "list" && len(node.Children) > 0 {
			// if one of its parents is not aligned, it's not possible to be aligned in sub-lists
			for _, parent := range parentStack {
				if parent.NodeType == "list" && !parent.IsAligned {
					node.IsAligned = false
					return node
				}
			}

			if len(parentStack) > 0 && index+1 < len(parentStack[0].Children) {
				next := parentStack[0].Children[index+1]
				if next.NodeType == "code" && next.IsIndented {
					// hard to align in this case
					node.IsAligned = false
					return node
				}
			}

			node.IsAligned = isAligned(node)
		}

		return node
	})
}
