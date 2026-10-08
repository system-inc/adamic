package mdast

import (
	"bufio"
	"fmt"
	"github.com/system-inc/cohere/internal/format/markdown/micromark"
	"os"
	"strings"
)

func adamicMdastEscape(s string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(s)
}
func adamicMdastOptional(v *string) string {
	if v == nil {
		return "-"
	}
	return "+" + *v
}
func AdamicMdastCanonical(root *Node) string {
	rows := []string{}
	stack := []*Node{root}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		count := -1
		if n.IsParent {
			count = len(n.Children)
		}
		start := "-"
		if n.Start != nil {
			start = fmt.Sprint("+", *n.Start)
		}
		checked := "-"
		if n.Checked != nil {
			checked = fmt.Sprint(*n.Checked)
		}
		p := n.Position
		position := "0,0,0,0,0,0"
		if p != nil {
			position = fmt.Sprintf("%d,%d,%d,%d,%d,%d", p.Start.Line, p.Start.Column, p.Start.Offset, p.End.Line, p.End.Column, p.End.Offset)
		}
		fields := []string{n.NodeType, fmt.Sprint(count), fmt.Sprint(n.IsLiteral), fmt.Sprint(n.ValueNull), n.Value, fmt.Sprint(n.Depth), fmt.Sprint(n.Ordered), start, fmt.Sprint(n.Spread), checked, adamicMdastOptional(n.Lang), adamicMdastOptional(n.Meta), n.URL, adamicMdastOptional(n.Title), adamicMdastOptional(n.Alt), adamicMdastOptional(n.Label), n.Identifier, n.ReferenceType, strings.Join(n.Align, ","), position}
		if n.FrontMatter != nil {
			f := n.FrontMatter
			fields = append(fields, "+"+f.Language, adamicMdastOptional(f.ExplicitLanguage), f.Value, f.StartDelimiter, f.EndDelimiter, f.Raw)
		} else {
			fields = append(fields, "-", "-", "", "", "", "")
		}

		for i, f := range fields {
			fields[i] = adamicMdastEscape(f)
		}
		rows = append(rows, strings.Join(fields, "\t"))
		for i := len(n.Children) - 1; i >= 0; i-- {
			stack = append(stack, n.Children[i])
		}
	}
	return strings.Join(rows, "\n")
}
func AdamicMdastFixture(source string) (string, string) {
	_, content := ParseFrontMatter(source)
	units := micromark.SourceUnits(content)
	events := micromark.Parse(units, micromark.MarkdownConstructs(), nil)
	fixture := "S\t" + adamicMdastEscape(source) + "\n" + micromark.AdamicEventTransport(events) + "F\n"
	root, err := ParseMarkdown(source, nil)
	if err != nil {
		panic(err)
	}
	unitOffsets := map[int]int{}
	for unit, offset := range byteOffsets(source, nil) {
		if _, present := unitOffsets[offset]; !present {
			unitOffsets[offset] = unit
		}
	}
	stack := []*Node{root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node.Position != nil {
			node.Position.Start.Offset = unitOffsets[node.Position.Start.Offset]
			node.Position.End.Offset = unitOffsets[node.Position.End.Offset]
		}
		stack = append(stack, node.Children...)
	}
	return fixture, adamicMdastEscape("content\t" + adamicMdastEscape(content) + "\n" + AdamicMdastCanonical(root))
}

// The fair component benchmark invokes the actual Go compiler on the same token transport.
func AdamicMdastFromTransport(path string) {
	data, e := os.ReadFile(path)
	if e != nil {
		panic(e)
	}
	output := bufio.NewWriter(os.Stdout)
	defer output.Flush()
	for _, document := range micromark.AdamicReadEventDocuments(string(data)) {
		units := micromark.SourceUnits(document.Source)
		offsets := make([]int, len(units)+1)
		for i := range offsets {
			offsets[i] = i
		}
		root := compile(document.Events, compileConfig(), offsets, nil)
		front, content := ParseFrontMatter(document.Source)
		if front != nil {
			lines := strings.Split(front.Raw, "\n")
			node := &Node{NodeType: "frontMatter", FrontMatter: front, Position: &Position{Start: Point{Line: 1, Column: 1}, End: Point{Line: len(lines), Column: len(micromark.SourceUnits(lines[len(lines)-1])) + 1, Offset: len(micromark.SourceUnits(front.Raw))}}}
			root.Children = append([]*Node{node}, root.Children...)
		}
		fmt.Fprintln(output, adamicMdastEscape("content\t"+adamicMdastEscape(content)+"\n"+AdamicMdastCanonical(root)))
	}
}

func AdamicMalformedEvents(name string) (message string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			message = fmt.Sprint(recovered)
		}
	}()
	paragraph := &micromark.Token{Type: "paragraph", Start: micromark.Point{Line: 1, Column: 1}, End: micromark.Point{Line: 1, Column: 1}}
	strong := &micromark.Token{Type: "strong", Start: paragraph.Start, End: paragraph.End}
	events := []micromark.Event{}
	switch name {
	case "unclosed":
		events = []micromark.Event{{Enter: true, Token: paragraph}}
	case "not-open":
		events = []micromark.Event{{Enter: false, Token: paragraph}}
	case "mismatch":
		events = []micromark.Event{{Enter: true, Token: paragraph}, {Enter: false, Token: strong}}
	default:
		panic("bad error witness")
	}
	compile(events, compileConfig(), []int{0}, nil)
	return "no error"
}
