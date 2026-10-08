// Public Go cohere parsing only; this adapter serializes its tree without using port code.
package main

import (
	"bufio"
	"fmt"
	"github.com/system-inc/cohere/internal/format/yaml/unist"
	"os"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func hex(text string) string {
	var out strings.Builder
	for len(text) > 0 {
		character, size := utf8.DecodeRuneInString(text)
		// Scalar sources may carry a lone surrogate as WTF-8.
		if size == 1 && len(text) >= 3 && text[0] == 0xed && text[1] >= 0xa0 && text[1] <= 0xbf && text[2] >= 0x80 && text[2] <= 0xbf {
			character = rune(text[0]&15)<<12 | rune(text[1]&63)<<6 | rune(text[2]&63)
			size = 3
		}
		if character > 0xffff {
			high, low := utf16.EncodeRune(character)
			fmt.Fprintf(&out, "%04x%04x", high, low)
		} else {
			fmt.Fprintf(&out, "%04x", character)
		}
		text = text[size:]
	}
	return out.String()
}
func point(point unist.Point, text string) string {
	offset := point.Offset
	if offset < 0 {
		return fmt.Sprintf("%d,%d,%d", point.Line, point.Column, offset)
	}
	if offset > len(text) {
		offset = len(utf16.Encode([]rune(text))) + offset - len(text)
	} else {
		offset = len(utf16.Encode([]rune(text[:offset])))
	}
	return fmt.Sprintf("%d,%d,%d", point.Line, point.Column, offset)
}
func outline(out *strings.Builder, node *unist.Node, text string) {
	if node == nil {
		out.WriteString("-")
		return
	}
	parent := "-"
	if node.Parent != nil {
		parent = node.Parent.NodeType + "," + point(node.Parent.Position.Start, text) + "," + point(node.Parent.Position.End, text)
	}
	indent := -1
	if node.Indent != nil {
		indent = *node.Indent
	}
	marker := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	params := []string{}
	for _, p := range node.Parameters {
		params = append(params, hex(p))
	}
	fmt.Fprintf(out, "%s|%s|%s|%s|%s|%s|%d|%d|%d|%s|%s|", node.NodeType, point(node.Position.Start, text), point(node.Position.End, text), parent, hex(node.Value), node.Chomping, indent, marker(node.DirectivesEndMarker), marker(node.DocumentEndMarker), hex(node.Name), strings.Join(params, ","))
	list := func(items []*unist.Node) {
		out.WriteString("[")
		for _, item := range items {
			outline(out, item, text)
			out.WriteString(",")
		}
		out.WriteString("]")
	}
	list(node.Children)
	out.WriteString("|")
	outline(out, node.Tag, text)
	out.WriteString("|")
	outline(out, node.Anchor, text)
	out.WriteString("|")
	list(node.MiddleComments)
	out.WriteString("|")
	list(node.LeadingComments)
	out.WriteString("|")
	outline(out, node.TrailingComment, text)
	out.WriteString("|")
	list(node.EndComments)
	out.WriteString("|")
	outline(out, node.IndicatorComment, text)
	out.WriteString("|")
	list(node.Comments)
}
func main() {
	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	unescape := strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t")
	number := 0
	for _, line := range strings.Split(string(source), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		text := unescape.Replace(parts[1])
		fmt.Fprintf(out, "case %d\n", number)
		number++
		node, err := unist.Parse(text, nil)
		if err != nil {
			if e, ok := err.(*unist.SyntaxError); ok {
				fmt.Fprintf(out, "syntax|%s|%s|%s|%s\n", e.Code, hex(e.Message), point(e.Position.Start, text), point(e.Position.End, text))
			} else if e, ok := err.(*unist.ThrownError); ok {
				fmt.Fprintf(out, "throw|%s|%s\n", e.Name, hex(e.Message))
			} else {
				panic(err)
			}
		} else {
			var b strings.Builder
			outline(&b, node, text)
			fmt.Fprintln(out, b.String())
		}
	}
}
