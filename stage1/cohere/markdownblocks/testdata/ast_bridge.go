package markdown

import (
	"fmt"
	"strings"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/markdown/mdast"
)

func adamicAstEscape(text string) string {
	return strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`).Replace(text)
}
func adamicAstSerialize(root *Node, source string) string {
	nodes := []*Node{}
	ids := map[*Node]int{}
	positions := map[*mdast.Position]int{}
	var walk func(*Node)
	walk = func(node *Node) {
		ids[node] = len(nodes)
		nodes = append(nodes, node)
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	lines := []string{}
	flag := func(value bool, mask int) int {
		if value {
			return mask
		}
		return 0
	}
	offsets := make([]int, len(source)+1)
	units := 0
	for index, point := range source {
		offsets[index] = units
		if point > 0xffff {
			units += 2
		} else {
			units++
		}
	}
	offsets[len(source)] = units
	offset := func(value int) int { return offsets[value] }
	for _, node := range nodes {
		mask := flag(node.Raw != nil, 1) + flag(node.IsParent, 2) + flag(node.Ordered, 4) + flag(node.IsIndented, 8) + flag(node.IsAligned, 16) + flag(node.OriginalAltText != nil, 32) + flag(node.IsLiteral, 64) + flag(node.IsCJ, 128) + flag(node.HasLeadingPunctuation, 256) + flag(node.HasTrailingPunctuation, 512) + flag(node.Position != nil, 1024)
		raw, alt := "", ""
		if node.Raw != nil {
			raw = *node.Raw
		}
		if node.OriginalAltText != nil {
			alt = *node.OriginalAltText
		}
		pos, start, end, startLine, endLine, startCol, endCol := -1, 0, 0, 0, 0, 0, 0
		if node.Position != nil {
			if _, ok := positions[node.Position]; !ok {
				positions[node.Position] = len(positions)
			}
			pos = positions[node.Position]
			start = offset(node.Position.Start.Offset)
			end = offset(node.Position.End.Offset)
			startLine = node.Position.Start.Line
			endLine = node.Position.End.Line
			startCol = node.Position.Start.Column
			endCol = node.Position.End.Column
		}
		children := []string{}
		for _, child := range node.Children {
			children = append(children, fmt.Sprint(ids[child]))
		}
		lines = append(lines, fmt.Sprintf("N\t%s\t%s\t%d\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%s\t%s", node.NodeType, adamicAstEscape(node.Value), mask, adamicAstEscape(raw), adamicAstEscape(alt), pos, start, end, startLine, endLine, startCol, endCol, node.Kind, strings.Join(children, ",")))
	}
	return strings.Join(lines, "\n")
}
func AdamicASTFixture(input string, tabWidth int, mode string) (fixture, want string, err error) {
	// Core normalization precedes the parser, as in the existing layout fixture.
	source := strings.TrimPrefix(strings.ReplaceAll(strings.ReplaceAll(input, "\r\n", "\n"), "\r", "\n"), "\ufeff")
	nodes := &arena.Arena[Node]{}
	root, err := mdast.ParseMarkdown(source, nodes)
	if err != nil {
		return "", "", err
	}
	if mode == "adjacent_text" {
		paragraph := root.Children[0]
		paragraph.Children = nil
		for i := 0; i < len(source); i++ {
			paragraph.Children = append(paragraph.Children, nodes.New(Node{
				NodeType: "text", IsLiteral: true, Value: source[i : i+1],
				Position: &mdast.Position{Start: mdast.Point{Line: 1, Column: i + 1, Offset: i}, End: mdast.Point{Line: 1, Column: i + 2, Offset: i + 1}},
			}))
		}
	}
	fixture = "A\t" + adamicAstEscape(source) + "\t" + fmt.Sprint(tabWidth) + "\n" + adamicAstSerialize(root, source) + "\nF\n"
	defer func() {
		if value := recover(); value != nil {
			want = "E\t" + fmt.Sprint(value)
		}
	}()
	root = preprocess(root, source, tabWidth, nodes)
	want = adamicAstSerialize(root, source)
	return
}
