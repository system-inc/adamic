package main

import (
	"bufio"
	"fmt"
	"github.com/system-inc/cohere/internal/format/markdown/micromark"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	if len(os.Args) > 1 && os.Args[1] == "--generate" {
		keys, values := []rune{}, []rune{}
		for code := rune(0); code <= unicode.MaxRune; code++ {
			v := unicode.ToLower(unicode.ToUpper(unicode.ToLower(code)))
			if v != code {
				keys = append(keys, code)
				values = append(values, v)
			}
		}
		for i, table := range [][]rune{keys, values} {
			name := "Keys"
			if i == 1 {
				name = "Values"
			}
			fmt.Fprintf(out, "// Go simple Unicode case folding used by micromark.NormalizeIdentifier.\nexport const identifierCase%s: readonly number[] = [\n", name)
			for j, v := range table {
				if j%12 == 0 {
					fmt.Fprint(out, "    ")
				}
				fmt.Fprintf(out, "%d, ", v)
				if j%12 == 11 {
					fmt.Fprintln(out)
				}
			}
			fmt.Fprint(out, "\n];\n")
		}
		return
	}
	for code := rune(0); code <= unicode.MaxRune; code++ {
		if code >= 0xD800 && code <= 0xDFFF {
			continue
		}
		value := strings.ToLower(micromark.NormalizeIdentifier(string(code)))
		if value == "" {
			fmt.Fprintln(out, -1)
		} else {
			r, _ := utf8.DecodeRuneInString(value)
			fmt.Fprintln(out, r)
		}
	}
}
