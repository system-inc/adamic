//go:build ignore

// Snapshot Go's quoting character classification for exact config diagnostics.
package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

func main() {
	var text strings.Builder
	text.WriteString("// Generated from the pinned Go unicode.IsPrint; run testdata/settings_tables.go.\nexport const printable: readonly (readonly number[])[] = [\n")
	start := -1
	for code := 0; code <= 0x110000; code++ {
		present := code < 0x110000 && unicode.IsPrint(rune(code))
		if present && start < 0 {
			start = code
		}
		if !present && start >= 0 {
			fmt.Fprintf(&text, "    [%d, %d],\n", start, code-1)
			start = -1
		}
	}
	text.WriteString("];\n")
	if e := os.WriteFile("stage1/cohere/command/settings/printable.a", []byte(text.String()), 0644); e != nil {
		panic(e)
	}
}
