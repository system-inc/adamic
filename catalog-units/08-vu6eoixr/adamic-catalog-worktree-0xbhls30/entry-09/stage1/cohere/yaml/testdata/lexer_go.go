// Built as cohere/command/formatter_comparison/main.go through a Go overlay.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"

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
	for _, line := range strings.Split(string(source), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		chunkSize, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		text := utf16.Encode([]rune(unescape.Replace(parts[1])))
		lexer := cst.NewLexer()
		fmt.Fprintf(out, "case %d\n", number)
		number++
		emit := func(source []uint16, incomplete bool) {
			for token := range lexer.Lex(source, incomplete) {
				fmt.Fprint(out, "=")
				for _, unit := range token {
					fmt.Fprintf(out, "%04x", unit)
				}
				fmt.Fprintln(out)
			}
		}
		if chunkSize == 0 {
			emit(text, false)
		} else {
			for start := 0; start < len(text); start += chunkSize {
				emit(text[start:min(start+chunkSize, len(text))], true)
			}
			emit(nil, false)
		}
	}
}
