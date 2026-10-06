package markdown

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/markdown/mdast"
	"github.com/system-inc/cohere/internal/format/printing"
)

type adamicBlock struct {
	node     *Node
	document doc.Doc
}
type adamicItem struct {
	node   *Node
	marker string
	blocks []adamicBlock
}
type adamicList struct {
	node            *Node
	sibling         int
	ancestorAligned bool
	next            *Node
	items           []adamicItem
}
type adamicHTML struct {
	value    string
	rootLast bool
	widths   []int
}
type adamicCode struct {
	indented                     bool
	value, language, metadata    string
	languageWidth, metadataWidth int
	widths                       []int
}
type adamicCell struct {
	document doc.Doc
	width    int
}
type adamicTable struct {
	rows      [][]adamicCell
	align     []string
	rowWidths []int
}
type adamicWord struct {
	value, previous, next string
	flags                 int
	width                 int
}
type adamicStructureChild struct {
	document   doc.Doc
	whitespace bool
}
type adamicStructure struct {
	kind, source string
	depth        int
	setext       bool
	children     []adamicStructureChild
}
type adamicRoot struct {
	source   string
	children []adamicBlock
}
type adamicLeaf struct {
	kind, value, source, url, title, label, referenceType, alt, originalAlt, metadata string
	hasTitle                                                                          bool
	flags, sibling                                                                    int
	children                                                                          []adamicBlock
}
type adamicWhitespace struct{ fields []string }
type adamicDocuments struct {
	whitespaces []adamicWhitespace
	leaves      []adamicLeaf
	roots       []adamicRoot
	structures  []adamicStructure
	native      bool
	lines       []string
	documents   int
	items       int
	htmls       []adamicHTML
	codes       []adamicCode
	tables      []adamicTable
	quotes      [][]adamicBlock
	lists       []adamicList
	words       []adamicWord
	groups      map[*doc.Group]int
	ids         map[*doc.GroupID]int
}

func adamicEscape(text string) string {
	return strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`).Replace(text)
}
func adamicFlag(value bool) int {
	if value {
		return 1
	}
	return 0
}
func (out *adamicDocuments) groupID(id *doc.GroupID) int {
	if id == nil {
		return -1
	}
	if value, found := out.ids[id]; found {
		return value
	}
	value := len(out.ids)
	out.ids[id] = value
	return value
}
func (out *adamicDocuments) add(kind, text string, width, flag, group int, children []int) int {
	ids := make([]string, len(children))
	for i, id := range children {
		ids[i] = strconv.Itoa(id)
	}
	out.lines = append(out.lines, fmt.Sprintf("D\t%s\t%s\t%d\t%d\t%d\t%s", kind, adamicEscape(text), width, flag, group, strings.Join(ids, ",")))
	result := out.documents
	out.documents++
	return result
}
func (out *adamicDocuments) serialize(document doc.Doc) int {
	if parts, array := doc.Parts(document); array {
		ids := make([]int, len(parts))
		for i, part := range parts {
			ids[i] = out.serialize(part)
		}
		return out.add("a", "", 0, 0, -1, ids)
	}
	switch node := document.(type) {
	case doc.Text:
		return out.add("t", string(node), doc.StringWidth(string(node)), 0, -1, nil)
	case *doc.Line:
		flags := adamicFlag(node.Soft) + 2*adamicFlag(node.Hard) + 4*adamicFlag(node.Literal)
		return out.add("h", "", 0, flags, -1, nil)
	case *doc.Indent:
		return out.add("i", "", 0, 0, -1, []int{out.serialize(node.Contents)})
	case *doc.Align:
		kind := []string{"w", "s", "r", "d"}[node.Kind]
		return out.add(kind, node.String, node.Width, 0, -1, []int{out.serialize(node.Contents)})
	case *doc.Fill:
		ids := make([]int, len(node.Parts))
		for i, part := range node.Parts {
			ids[i] = out.serialize(part)
		}
		return out.add("f", "", 0, 0, -1, ids)
	case *doc.Group:
		if id, known := out.groups[node]; known {
			return id
		}
		parts := []doc.Doc{node.Contents}
		flag := adamicFlag(node.Break)
		if node.ExpandedStates != nil {
			parts = node.ExpandedStates
			flag += 2
		}
		ids := make([]int, len(parts))
		for i, part := range parts {
			ids[i] = out.serialize(part)
		}
		id := out.add("g", "", 0, flag, out.groupID(node.ID), ids)
		out.groups[node] = id
		return id
	case *doc.IfBreak:
		return out.add("b", "", 0, 0, out.groupID(node.GroupID), []int{out.serialize(node.BreakContents), out.serialize(node.FlatContents)})
	case *doc.Label:
		if out.native && strings.HasPrefix(node.Label, "adamic-list:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-list:"))
			frame := out.lists[index]
			var itemIDs []string
			for _, item := range frame.items {
				var blocks []string
				for _, block := range item.blocks {
					id := out.serialize(block.document)
					n := block.node
					blocks = append(blocks, fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d", n.NodeType, id, n.Position.Start.Line, n.Position.End.Line, n.Position.Start.Column, adamicFlag(n.IsIndented), adamicFlag(isPrettierIgnore(n) == "next")))
				}
				checked := -1
				if item.node.Checked != nil {
					checked = adamicFlag(*item.node.Checked)
				}
				out.lines = append(out.lines, fmt.Sprintf("I\t%s\t%d\t%d\t%d\t%d\t%s", adamicEscape(item.marker), item.node.Position.Start.Line, item.node.Position.End.Line, adamicFlag(item.node.Spread), checked, strings.Join(blocks, ";")))
				itemIDs = append(itemIDs, strconv.Itoa(out.items))
				out.items++
			}
			start := 0
			if frame.node.Start != nil {
				start = *frame.node.Start
			}
			nextIndented := frame.next != nil && frame.next.NodeType == "code" && frame.next.IsIndented
			nextCode := ""
			if nextIndented {
				nextCode = frame.next.Value
			}
			out.lines = append(out.lines, fmt.Sprintf("L\t%d\t%d\t%d\t%d\t%d\t%s\t%s", adamicFlag(frame.node.Ordered), start, frame.sibling, adamicFlag(frame.ancestorAligned), adamicFlag(nextIndented), adamicEscape(nextCode), strings.Join(itemIDs, ",")))
			id := out.documents
			out.documents++
			return id
		}
		if out.native && strings.HasPrefix(node.Label, "adamic-quote:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-quote:"))
			var blocks []string
			for _, block := range out.quotes[index] {
				id := out.serialize(block.document)
				n := block.node
				blocks = append(blocks, fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d", n.NodeType, id, n.Position.Start.Line, n.Position.End.Line, n.Position.Start.Column, adamicFlag(n.IsIndented), adamicFlag(isPrettierIgnore(n) == "next")))
			}
			out.lines = append(out.lines, "Q\t"+strings.Join(blocks, ";"))
			id := out.documents
			out.documents++
			return id
		}
		if out.native && strings.HasPrefix(node.Label, "adamic-table:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-table:"))
			frame := out.tables[index]
			rows := []string{}
			for _, row := range frame.rows {
				cells := []string{}
				for _, cell := range row {
					cells = append(cells, fmt.Sprintf("%d,%d", out.serialize(cell.document), cell.width))
				}
				rows = append(rows, strings.Join(cells, ";"))
			}
			widths := []string{}
			for _, width := range frame.rowWidths {
				widths = append(widths, strconv.Itoa(width))
			}
			out.lines = append(out.lines, fmt.Sprintf("T\t%s\t%s\t%s", strings.Join(frame.align, ","), strings.Join(widths, ","), strings.Join(rows, ":")))
			id := out.documents
			out.documents++
			return id
		}
		if out.native && strings.HasPrefix(node.Label, "adamic-code:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-code:"))
			frame := out.codes[index]
			widths := []string{}
			for _, width := range frame.widths {
				widths = append(widths, strconv.Itoa(width))
			}
			out.lines = append(out.lines, fmt.Sprintf("C\t%d\t%s\t%s\t%d\t%d\t%s\t%s", adamicFlag(frame.indented), adamicEscape(frame.language), adamicEscape(frame.metadata), frame.languageWidth, frame.metadataWidth, strings.Join(widths, ","), adamicEscape(frame.value)))
			id := out.documents
			out.documents++
			return id
		}
		if out.native && strings.HasPrefix(node.Label, "adamic-html:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-html:"))
			frame := out.htmls[index]
			widths := []string{}
			for _, width := range frame.widths {
				widths = append(widths, strconv.Itoa(width))
			}
			out.lines = append(out.lines, fmt.Sprintf("H\t%d\t%s\t%s", adamicFlag(frame.rootLast), strings.Join(widths, ","), adamicEscape(frame.value)))
			id := out.documents
			out.documents++
			return id
		}
		if out.native && strings.HasPrefix(node.Label, "adamic-whitespace:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-whitespace:"))
			out.lines = append(out.lines, "S\t"+strings.Join(out.whitespaces[index].fields, "\t"))
			id := out.documents
			out.documents++
			return id
		}
		if out.native && strings.HasPrefix(node.Label, "adamic-leaf:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-leaf:"))
			frame := out.leaves[index]
			children := []string{}
			for _, child := range frame.children {
				n := child.node
				start, end, column := 0, 0, 0
				if n.Position != nil {
					start, end, column = n.Position.Start.Line, n.Position.End.Line, n.Position.Start.Column
				}
				children = append(children, fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d", n.NodeType, out.serialize(child.document), start, end, column, adamicFlag(n.IsIndented), adamicFlag(isPrettierIgnore(n) == "next")))
			}
			out.lines = append(out.lines, fmt.Sprintf("Y\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%d\t%d\t%s", frame.kind, adamicEscape(frame.value), adamicEscape(frame.source), adamicEscape(frame.url), adamicEscape(frame.title), adamicFlag(frame.hasTitle), adamicEscape(frame.label), frame.referenceType, adamicEscape(frame.alt), adamicEscape(frame.originalAlt), adamicEscape(frame.metadata), frame.flags, frame.sibling, strings.Join(children, ";")))
			id := out.documents
			out.documents++
			return id
		}
		if out.native && strings.HasPrefix(node.Label, "adamic-root:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-root:"))
			frame := out.roots[index]
			for _, child := range frame.children {
				id := out.serialize(child.document)
				n := child.node
				out.lines = append(out.lines, fmt.Sprintf("J\t%d\t%s\t%d\t%d\t%d\t%d\t%s", id, n.NodeType, n.Position.Start.Line, n.Position.End.Line, utf16Length(frame.source[:n.Position.Start.Offset]), utf16Length(frame.source[:n.Position.End.Offset]), adamicEscape(n.Value)))
			}
			out.lines = append(out.lines, "O\t"+adamicEscape(frame.source))
			id := out.documents
			out.documents++
			return id
		}
		if out.native && strings.HasPrefix(node.Label, "adamic-structure:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-structure:"))
			frame := out.structures[index]
			children := []string{}
			for _, child := range frame.children {
				children = append(children, fmt.Sprintf("%d,%d", out.serialize(child.document), adamicFlag(child.whitespace)))
			}
			out.lines = append(out.lines, fmt.Sprintf("N\t%s\t%d\t%d\t%s\t%s", frame.kind, frame.depth, adamicFlag(frame.setext), adamicEscape(frame.source), strings.Join(children, ";")))
			id := out.documents
			out.documents++
			return id
		}
		if out.native && strings.HasPrefix(node.Label, "adamic-word:") {
			index, _ := strconv.Atoi(strings.TrimPrefix(node.Label, "adamic-word:"))
			word := out.words[index]
			out.lines = append(out.lines, fmt.Sprintf("W\t%s\t%d\t%d\t%s\t%s", adamicEscape(word.value), word.flags, word.width, adamicEscape(word.previous), adamicEscape(word.next)))
			id := out.documents
			out.documents++
			return id
		}
		return out.add("l", "", 0, 0, -1, []int{out.serialize(node.Contents)})
	default:
		switch document {
		case doc.BreakParent:
			return out.add("p", "", 0, 0, -1, nil)
		case doc.Trim:
			return out.add("x", "", 0, 0, -1, nil)
		}
		panic(fmt.Sprintf("unexpected Markdown doc %T", document))
	}
}

// AdamicListFixture uses real parsing/preprocessing/printing. Labels are transparent to Go's doc
// printer, and replace only list construction and the preceding native word component in the fixture.
func AdamicListFixture(input string) (native, canonical, formatted string, err error) {
	bom := strings.HasPrefix(input, "\ufeff")
	text := strings.TrimPrefix(input, "\ufeff")
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	nodes := &arena.Arena[Node]{}
	ast, err := mdast.ParseMarkdown(text, nodes)
	if err != nil {
		return "", "", "", err
	}
	lists := []adamicList{}
	quotes := [][]adamicBlock{}
	tables := []adamicTable{}
	codes := []adamicCode{}
	htmls := []adamicHTML{}
	words := []adamicWord{}
	structures := []adamicStructure{}
	roots := []adamicRoot{}
	leaves := []adamicLeaf{}
	whitespaces := []adamicWhitespace{}
	printer := *mdastPrinter
	printer.PrintPrettierIgnored = func(path *astPath, options *options, print printing.PrintFunc, args any) doc.Doc {
		node := currentNode(path)
		frame := adamicLeaf{kind: "ignored", value: node.NodeType, source: options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset], sibling: -1, flags: adamicFlag(path.HasAncestor(func(n *Node) bool { return n.NodeType == "blockquote" }))}
		id := len(leaves)
		leaves = append(leaves, frame)
		return doc.NewLabel(fmt.Sprintf("adamic-leaf:%d", id), printPrettierIgnored(path, options, print, args))
	}
	printer.Print = func(path *astPath, options *options, print printing.PrintFunc, args any) doc.Doc {
		original := printMdast(path, options, print, args)
		node := currentNode(path)
		if shouldRemainTheSameContent(path) {
			id := len(leaves)
			leaves = append(leaves, adamicLeaf{kind: "preservedLabel", source: options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset], sibling: -1})
			return doc.NewLabel(fmt.Sprintf("adamic-leaf:%d", id), original)
		}
		if node.NodeType == "whitespace" {
			frame := adamicWhitespace{fields: []string{adamicEscape(node.Value), settingsOf(options).proseWrap, "0"}}
			appendToken := func(n *Node) {
				if n == nil {
					frame.fields = append(frame.fields, "0", "", "", "", "0")
					return
				}
				frame.fields = append(frame.fields, "1", n.NodeType, adamicEscape(n.Value), n.Kind, strconv.Itoa(adamicFlag(n.IsCJ)+2*adamicFlag(n.HasLeadingPunctuation)+4*adamicFlag(n.HasTrailingPunctuation)))
			}
			appendToken(previousNode(path))
			appendToken(nextNode(path))
			appendToken(siblingAt(path, pathIndex(path)+2))
			kinds, setext, samples := []string{}, []string{}, []string{}
			for _, ancestor := range path.Ancestors() {
				kinds = append(kinds, ancestor.NodeType)
				setext = append(setext, strconv.Itoa(adamicFlag(ancestor.Position != nil && isSetextHeading(ancestor))))
			}
			siblings := parentNode(path).Children
			for i := 1; i < len(siblings)-1; i++ {
				n := siblings[i]
				// Transport only samples that the native/original policy can count.
				// Keep the actual whitespace value and neighboring kinds, not Go's decision.
				previousKind, nextKind := siblings[i-1].Kind, siblings[i+1].Kind
				if n.NodeType != "whitespace" || (n.Value != "" && n.Value != " ") || !((previousKind == kindCJLetter && nextKind == kindNonCJK) || (previousKind == kindNonCJK && nextKind == kindCJLetter)) {
					continue
				}
				value := -1
				if n.Value == "" {
					value = 0
				} else if n.Value == " " {
					value = 1
				}
				samples = append(samples, fmt.Sprintf("%s,%d,%s,%s", n.NodeType, value, siblings[i-1].Kind, siblings[i+1].Kind))
			}
			frame.fields = append(frame.fields, strings.Join(kinds, ","), strings.Join(setext, ","), strings.Join(samples, ";"))
			saved := settingsOf(options).proseWrap
			wanted := []string{}
			classify := func(d doc.Doc) string {
				switch n := d.(type) {
				case doc.Text:
					return "T" + string(n)
				case *doc.Line:
					if n.Soft {
						return "S"
					}
					if n.Hard {
						return "H"
					}
					return "L"
				}
				if _, ok := doc.Parts(d); ok {
					return "H"
				}
				panic("unexpected whitespace doc")
			}
			for _, mode := range []string{"preserve", "always", "never"} {
				settingsOf(options).proseWrap = mode
				wanted = append(wanted, classify(printMdast(path, options, print, args)), classify(printWhitespace(path, node.Value, mode, true, options)))
			}
			settingsOf(options).proseWrap = saved
			frame.fields = append(frame.fields, strings.Join(wanted, ","))
			id := len(whitespaces)
			whitespaces = append(whitespaces, frame)
			return doc.NewLabel(fmt.Sprintf("adamic-whitespace:%d", id), original)
		}
		switch node.NodeType {
		case "frontMatter", "emphasis", "strong", "delete", "inlineCode", "wikiLink", "link", "image", "thematicBreak", "linkReference", "imageReference", "definition", "footnoteReference", "footnoteDefinition", "break", "liquidNode", "math", "inlineMath", "text", "tableCell":
			if !shouldRemainTheSameContent(path) {
				frame := adamicLeaf{kind: node.NodeType, value: node.Value, url: node.URL, referenceType: node.ReferenceType, hasTitle: node.Title != nil, sibling: -1}
				if node.Position != nil {
					frame.source = options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset]
				}
				if node.FrontMatter != nil {
					frame.source = node.FrontMatter.Raw
				}
				if node.Title != nil {
					frame.title = *node.Title
				}
				if node.Label != nil {
					frame.label = *node.Label
				}
				if node.Alt != nil {
					frame.alt = *node.Alt
				}
				if node.OriginalAltText != nil {
					frame.originalAlt = *node.OriginalAltText
				}
				if node.Meta != nil {
					frame.metadata = *node.Meta
				}
				if node.NodeType == "emphasis" {
					frame.flags = adamicFlag(len(node.Children) > 0 && isAutolink(node.Children[0])) + 2*adamicFlag(prevOrNextWord(path)) + 4*adamicFlag(printing.CallParent(path, func(p *astPath) bool { return currentNode(p).NodeType == "strong" && prevOrNextWord(p) }, 0)) + 8*adamicFlag(path.HasAncestor(func(n *Node) bool { return n.NodeType == "emphasis" }))
				}
				if node.NodeType == "inlineCode" {
					frame.flags = adamicFlag(path.HasAncestor(func(n *Node) bool { return n.NodeType == "tableCell" }))
				}
				if node.NodeType == "wikiLink" {
					frame.flags = adamicFlag(node.ValueNull)
				}
				if node.NodeType == "thematicBreak" {
					ancestors := path.Ancestors()
					for i, ancestor := range ancestors {
						if ancestor.NodeType == "list" {
							frame.sibling = getNthListSiblingIndex(ancestor, ancestors[i+1])
							break
						}
					}
				}
				if node.IsParent {
					path.Each(func(childPath *astPath, _ int, _ any) {
						frame.children = append(frame.children, adamicBlock{currentNode(childPath), print(nil, nil)})
					}, "children")
				}
				id := len(leaves)
				leaves = append(leaves, frame)
				return doc.NewLabel(fmt.Sprintf("adamic-leaf:%d", id), original)
			}
		}
		if node.NodeType == "root" {
			frame := adamicRoot{source: options.OriginalText}
			path.Each(func(childPath *astPath, _ int, _ any) {
				frame.children = append(frame.children, adamicBlock{currentNode(childPath), print(nil, nil)})
			}, "children")
			id := len(roots)
			roots = append(roots, frame)
			return doc.NewLabel(fmt.Sprintf("adamic-root:%d", id), original)
		}
		if (node.NodeType == "heading" || node.NodeType == "sentence" || node.NodeType == "paragraph") && !shouldRemainTheSameContent(path) {
			frame := adamicStructure{kind: node.NodeType, depth: node.Depth, setext: isSetextHeading(node)}
			if node.NodeType == "heading" {
				frame.source = options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset]
			}
			path.Each(func(childPath *astPath, _ int, _ any) {
				frame.children = append(frame.children, adamicStructureChild{print(nil, nil), currentNode(childPath).NodeType == "whitespace"})
			}, "children")
			id := len(structures)
			structures = append(structures, frame)
			return doc.NewLabel(fmt.Sprintf("adamic-structure:%d", id), original)
		}
		if node.NodeType == "word" && !shouldRemainTheSameContent(path) {
			emphasis, found := path.FindAncestor(func(n *Node) bool { return n.NodeType == "emphasis" || n.NodeType == "strong" })
			grandparent, _ := path.Grandparent()
			leading := found && path.IsFirst() && printingCallParentIsFirst(path) && grandparent == emphasis
			pseudo := settingsOf(options).proseWrap == "preserve" && parentNode(path).NodeType == "sentence" && isNewLine(previousNode(path)) && (path.IsLast() || isNewLine(nextNode(path)))
			previous, next := "", ""
			if n := previousNode(path); n != nil {
				previous = n.Value
			}
			if n := nextNode(path); n != nil {
				next = n.Value
			}
			printed, ok := original.(doc.Text)
			if !ok {
				panic("word did not produce Text")
			}
			id := len(words)
			words = append(words, adamicWord{node.Value, previous, next, adamicFlag(found) + 2*adamicFlag(leading) + 4*adamicFlag(pseudo), doc.StringWidth(string(printed))})
			return doc.NewLabel(fmt.Sprintf("adamic-word:%d", id), original)
		}
		if node.NodeType == "html" {
			frame := adamicHTML{value: node.Value, rootLast: parentNode(path).NodeType == "root" && path.IsLast()}
			value := frame.value
			if frame.rootLast {
				value = strings.TrimRight(value, javaScriptSpaceText)
			}
			for _, line := range strings.Split(value, "\n") {
				frame.widths = append(frame.widths, doc.StringWidth(line))
			}
			id := len(htmls)
			htmls = append(htmls, frame)
			return doc.NewLabel(fmt.Sprintf("adamic-html:%d", id), original)
		}
		if node.NodeType == "code" {
			frame := adamicCode{indented: node.IsIndented, value: node.Value}
			if node.Lang != nil {
				frame.language = *node.Lang
			}
			if node.Meta != nil {
				frame.metadata = *node.Meta
			}
			frame.languageWidth = doc.StringWidth(frame.language)
			if frame.metadata != "" {
				frame.metadataWidth = doc.StringWidth(" " + frame.metadata)
			}
			for _, line := range strings.Split(frame.value, "\n") {
				frame.widths = append(frame.widths, doc.StringWidth(line))
			}
			id := len(codes)
			codes = append(codes, frame)
			return doc.NewLabel(fmt.Sprintf("adamic-code:%d", id), original)
		}
		if node.NodeType == "table" {
			frame := adamicTable{align: node.Align}
			path.Each(func(rowPath *astPath, _ int, _ any) {
				cells := []adamicCell{}
				rowPath.Each(func(_ *astPath, _ int, _ any) {
					document := print(nil, nil)
					cellText := doc.Print(document, doc.Options{PrintWidth: 120, TabWidth: 4})
					cells = append(cells, adamicCell{document, doc.StringWidth(cellText)})
				}, "children")
				frame.rows = append(frame.rows, cells)
			}, "children")
			for _, line := range strings.Split(doc.Print(original, doc.Options{PrintWidth: 120, TabWidth: 4}), "\n") {
				frame.rowWidths = append(frame.rowWidths, doc.StringWidth(line))
			}
			id := len(tables)
			tables = append(tables, frame)
			return doc.NewLabel(fmt.Sprintf("adamic-table:%d", id), original)
		}
		if node.NodeType == "blockquote" {
			blocks := []adamicBlock{}
			path.Each(func(blockPath *astPath, _ int, _ any) {
				blocks = append(blocks, adamicBlock{currentNode(blockPath), print(nil, nil)})
			}, "children")
			id := len(quotes)
			quotes = append(quotes, blocks)
			return doc.NewLabel(fmt.Sprintf("adamic-quote:%d", id), original)
		}
		if node.NodeType != "list" {
			return original
		}
		frame := adamicList{node: node, sibling: getNthListSiblingIndex(node, parentNode(path)), ancestorAligned: true, next: nextNode(path)}
		for _, ancestor := range path.Ancestors() {
			if ancestor.NodeType == "list" && !ancestor.IsAligned {
				frame.ancestorAligned = false
			}
		}
		path.Each(func(itemPath *astPath, _ int, _ any) {
			itemNode := currentNode(itemPath)
			item := adamicItem{node: itemNode}
			end := itemNode.Position.End.Offset
			if len(itemNode.Children) > 0 {
				end = itemNode.Children[0].Position.Start.Offset
			}
			item.marker = text[itemNode.Position.Start.Offset:end]
			itemPath.Each(func(blockPath *astPath, _ int, _ any) {
				item.blocks = append(item.blocks, adamicBlock{currentNode(blockPath), print(nil, nil)})
			}, "children")
			frame.items = append(frame.items, item)
		}, "children")
		id := len(lists)
		lists = append(lists, frame)
		return doc.NewLabel(fmt.Sprintf("adamic-list:%d", id), original)
	}
	options := &options{Printer: &printer, OriginalText: text, EmbeddedLanguageFormatting: "off", Settings: &settings{proseWrap: "preserve", singleQuote: true, tabWidth: 4, printWidth: 120, nodes: nodes}}
	document, err := printing.PrintAstToDoc(ast, nil, options)
	if err != nil {
		return "", "", "", err
	}
	// Serialize before Print mutates group break flags, so native propagation is exercised.
	for _, side := range []bool{true, false} {
		out := &adamicDocuments{native: side, whitespaces: whitespaces, leaves: leaves, roots: roots, structures: structures, htmls: htmls, codes: codes, tables: tables, quotes: quotes, lists: lists, words: words, groups: map[*doc.Group]int{}, ids: map[*doc.GroupID]int{}}
		root := out.serialize(document)
		out.lines = append(out.lines, fmt.Sprintf("R\t%d\t%d", root, adamicFlag(bom)))
		stream := strings.Join(out.lines, "\n") + "\n"
		if side {
			native = stream
		} else {
			canonical = stream
		}
	}
	formatted = doc.Print(document, doc.Options{TabWidth: 4, PrintWidth: 120})
	if bom {
		formatted = "\ufeff" + formatted
	}
	baseline, err := Format(text, formatoptions.Default(), nil)
	if err != nil {
		return "", "", "", err
	}
	if bom {
		baseline = "\ufeff" + baseline
	}
	if baseline != formatted {
		panic("fixture collector changed the actual Go formatter output")
	}
	return native, canonical, formatted, nil
}
