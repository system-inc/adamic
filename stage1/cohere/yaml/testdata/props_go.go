// This command only exposes Go cohere's property resolver through a source overlay.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/system-inc/cohere/internal/format/yaml/compose"
	"github.com/system-inc/cohere/internal/format/yaml/cst"
)

func main() {
	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	unescape := strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t")
	number := 0
	var visit func(*cst.Token)
	visit = func(token *cst.Token) {
		if token == nil {
			return
		}
		first := func(tokens []*cst.Token) *cst.Token {
			if len(tokens) > 0 {
				return tokens[0]
			}
			return nil
		}
		if token.Type == "document" {
			next := token.Value
			if next == nil {
				next = first(token.End)
			}
			fmt.Fprintln(out, compose.PortPropsOutline(token.Start, "", "doc-start", next, token.Offset, 0, true))
		} else {
			flow := ""
			if token.Type == "flow-collection" {
				flow = "flow sequence"
				if string(utf16.Decode(token.FlowStart.Source)) == "{" {
					flow = "flow map"
				}
			}
			for _, item := range token.Items {
				indicator := "explicit-key-ind"
				next := item.Key
				if next == nil {
					next = first(item.Sep)
				}
				if token.Type == "block-seq" {
					indicator = "seq-item-ind"
					next = item.Value
				}
				fmt.Fprintln(out, compose.PortPropsOutline(item.Start, flow, indicator, next, token.Offset, token.Indent, flow == ""))
				if item.Sep != nil {
					fmt.Fprintln(out, compose.PortPropsOutline(item.Sep, flow, "map-value-ind", item.Value, token.Offset, token.Indent, false))
				}
			}
		}
		visit(token.Value)
		for _, item := range token.Items {
			visit(item.Key)
			visit(item.Value)
		}
	}
	for _, line := range strings.Split(string(source), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		size, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		text := utf16.Encode([]rune(unescape.Replace(parts[1])))
		parser := cst.NewParser(nil)
		fmt.Fprintf(out, "case %d\n", number)
		number++
		emit := func(source []uint16, incomplete bool) {
			for token := range parser.Parse(source, incomplete) {
				visit(token)
			}
		}
		if size == 0 {
			emit(text, false)
		} else {
			for start := 0; start < len(text); start += size {
				emit(text[start:min(start+size, len(text))], true)
			}
			emit(nil, false)
		}
	}
}
