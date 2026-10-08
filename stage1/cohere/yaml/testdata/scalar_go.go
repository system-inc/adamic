// This command only exposes Go cohere's scalar resolver through a source overlay.
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
		switch token.Type {
		case "scalar", "single-quoted-scalar", "double-quoted-scalar", "block-scalar":
			for _, root := range []bool{true, false} {
				fmt.Fprintln(out, compose.PortScalarOutline(token, root))
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
