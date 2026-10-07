package main

import (
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"strings"
)

func main() {
	for _, prefix := range []string{"", "a", "^a", "👍", "[", "-", "\x00"} {
		for count := 0; count <= 12; count++ {
			for _, suffix := range []string{"-", "a", "--"} {
				body := prefix + strings.Repeat("\\", count) + suffix
				text := esregexp.Wave09EscapeTrailingDash(body)
				parts := []string{}
				for _, value := range text {
					parts = append(parts, fmt.Sprint(value))
				}
				fmt.Println(strings.Join(parts, ","))
			}
		}
	}
}
