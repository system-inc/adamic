//go:build ignore

// Independently exercise all Unicode predicate boundaries and Go quote spellings.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

func escaped(text string) string {
	var out strings.Builder
	for _, code := range utf16.Encode([]rune(text)) {
		if code >= 32 && code <= 126 && code != 92 {
			out.WriteRune(rune(code))
		} else {
			fmt.Fprintf(&out, `\u%04x`, code)
		}
	}
	return out.String()
}
func main() {
	values := []string{"", "A", "a", "É", "Ω", "𐐀", "😀", "\"\\\a\b\f\n\r\t\v", "hello\u0085\u00a0\u200d\u2028\u2029🌍"}
	for code := rune(0); code <= unicode.MaxRune; code++ {
		if code >= 0xd800 && code <= 0xdfff {
			continue
		}
		if code < 512 || unicode.IsUpper(code) != unicode.IsUpper(code-1) || unicode.IsUpper(code) != unicode.IsUpper(code+1) || unicode.IsPrint(code) != unicode.IsPrint(code-1) || unicode.IsPrint(code) != unicode.IsPrint(code+1) {
			values = append(values, string(code))
		}
	}
	file, err := os.Create(os.Args[1])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(file).Encode(values); err != nil {
		panic(err)
	}
	if err = file.Close(); err != nil {
		panic(err)
	}
	for _, value := range values {
		runes := []rune(value)
		upper := len(runes) > 0 && unicode.IsUpper(runes[0])
		fmt.Printf("%t\t%s\n", upper, escaped(strconv.Quote(value)))
	}
}
