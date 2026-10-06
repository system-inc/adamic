// Build through a Go overlay inside cohere so its internal package is the independent oracle.
package main

import (
	"fmt"
	estree "github.com/system-inc/cohere/internal/format/estree"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"
)

func written(text string) string {
	var out strings.Builder
	for _, unit := range utf16.Encode([]rune(text)) {
		if unit >= 32 && unit <= 126 && unit != 92 {
			out.WriteByte(byte(unit))
		} else {
			fmt.Fprintf(&out, `\u%04x`, unit)
		}
	}
	return out.String()
}
func bit(value bool) int {
	if value {
		return 1
	}
	return 0
}
func dump(node *estree.Node, depth int) {
	if node == nil {
		fmt.Printf("%d null\n", depth)
		return
	}
	fmt.Printf("%d %s %d %d %d %d %d %d %d %d\n", depth, node.Type(), node.Start(), node.End(), bit(node.Parenthesized), bit(node.HasContentEnd), node.ContentEnd, estree.LocStart(node), estree.LocEnd(node), bit(estree.ShouldIgnoredNodePrintSemicolon(node)))
	for _, key := range node.Keys() {
		value := node.Get(key)
		fmt.Printf("%d .%s ", depth, key)
		switch v := value.(type) {
		case nil:
			fmt.Println("null")
		case *estree.Node:
			fmt.Println("node")
			dump(v, depth+1)
		case []*estree.Node:
			fmt.Printf("list %d\n", len(v))
			for _, child := range v {
				dump(child, depth+1)
			}
		case string:
			fmt.Printf("string %s\n", written(v))
		case bool:
			fmt.Printf("bool %d\n", bit(v))
		case float64:
			text := strconv.FormatFloat(v, 'f', -1, 64)
			if math.IsNaN(v) {
				text = "NaN"
			}
			if math.IsInf(v, 1) {
				text = "Infinity"
			}
			fmt.Printf("number %s\n", text)
		case *estree.Regex:
			fmt.Printf("regex %s\t%s\n", written(v.Pattern), written(v.Flags))
		case *estree.TemplateValue:
			cooked := "<null>"
			if v.Cooked != nil {
				cooked = written(*v.Cooked)
			}
			fmt.Printf("template %s\t%s\n", written(v.Raw), cooked)
		default:
			panic(fmt.Sprintf("unrepresented Go field %s %T", key, value))
		}
	}
}
func run(path string) {
	text, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	program, comments, stripped, err := estree.ParseTypeScript(path, string(text), nil)
	if err != nil {
		panic(err)
	}
	dump(program, 0)
	for _, comment := range comments {
		fmt.Println("comment")
		dump(comment, 0)
	}
	fmt.Printf("stripped %s\n", written(stripped.Text()))
}
func main() {
	if os.Args[1] == "--manifest" {
		data, err := os.ReadFile(os.Args[2])
		if err != nil {
			panic(err)
		}
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				run(path)
			}
		}
	} else {
		run(os.Args[1])
	}
}
