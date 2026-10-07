package main

import (
	"bufio"
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"os"
	"strings"
)

func spelling(text string) string {
	parts := []string{}
	for _, value := range text {
		parts = append(parts, fmt.Sprint(value))
	}
	return strings.Join(parts, ",")
}

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for _, mode := range []bool{false, true} {
		fmt.Fprintln(out, mode)
		for value := rune(-1); value <= 1114112; value++ {
			text, changed := esregexp.CaseClass(value, mode)
			if changed || text != "" {
				fmt.Fprintf(out, "%d\t%t\t%s\n", value, changed, spelling(text))
			}
		}
	}
	for _, value := range []rune{-1, 0, 8, 9, 10, 13, 32, 45, 91, 92, 93, 94, 127, 55296, 56319, 57343, 65533, 65535, 65536, 1114111, 1114112} {
		fmt.Fprintf(out, "escape\t%d\t%s\n", value, spelling(esregexp.EscapeClassRune(value)))
	}
	for index := 0; index < 4000; index++ {
		value := rune((index * 7919) % 1114112)
		fmt.Fprintf(out, "escape\t%d\t%s\n", value, spelling(esregexp.EscapeClassRune(value)))
	}
}
