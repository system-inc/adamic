// Built through the cohere command overlay, never in the Adamic module.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/system-inc/cohere/internal/format/yaml/cst"
)

func hex(out io.Writer, units []uint16) {
	for _, unit := range units {
		fmt.Fprintf(out, "%04x", unit)
	}
}
func list(out io.Writer, tokens []*cst.Token) {
	fmt.Fprint(out, "[")
	for _, token := range tokens {
		outline(out, token)
		fmt.Fprint(out, ",")
	}
	fmt.Fprint(out, "]")
}
func flag(value bool) int {
	if value {
		return 1
	}
	return 0
}
func outline(out io.Writer, token *cst.Token) {
	if token == nil {
		fmt.Fprint(out, "-")
		return
	}
	fmt.Fprintf(out, "%s|%d|", token.Type, token.Offset)
	if token.HasIndent() {
		fmt.Fprint(out, token.Indent)
	} else {
		fmt.Fprint(out, "-")
	}
	fmt.Fprint(out, "|")
	hex(out, token.Source)
	fmt.Fprint(out, "|")
	hex(out, utf16.Encode([]rune(token.Message)))
	fmt.Fprint(out, "|")
	list(out, token.Start)
	fmt.Fprint(out, "|")
	outline(out, token.FlowStart)
	fmt.Fprint(out, "|")
	outline(out, token.Value)
	fmt.Fprint(out, "|")
	if token.End == nil {
		fmt.Fprint(out, "-")
	} else {
		list(out, token.End)
	}
	fmt.Fprint(out, "|")
	list(out, token.Props)
	fmt.Fprint(out, "|[")
	for _, item := range token.Items {
		list(out, item.Start)
		fmt.Fprintf(out, "|%d|", flag(item.HasKey()))
		outline(out, item.Key)
		fmt.Fprint(out, "|")
		if item.Sep == nil {
			fmt.Fprint(out, "-")
		} else {
			list(out, item.Sep)
		}
		fmt.Fprint(out, "|")
		outline(out, item.Value)
		fmt.Fprintf(out, "|%d,", flag(item.ExplicitKey))
	}
	fmt.Fprint(out, "]")
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
		size, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		text := utf16.Encode([]rune(unescape.Replace(parts[1])))
		lines := cst.NewLineCounter()
		parser := cst.NewParser(lines.AddNewLine)
		fmt.Fprintf(out, "case %d\n", number)
		number++
		emit := func(source []uint16, incomplete bool) {
			for token := range parser.Parse(source, incomplete) {
				outline(out, token)
				fmt.Fprintln(out)
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
		fmt.Fprint(out, "lines ")
		for index, start := range lines.LineStarts {
			if index > 0 {
				fmt.Fprint(out, ",")
			}
			fmt.Fprint(out, start)
		}
		fmt.Fprintln(out)
	}
}
