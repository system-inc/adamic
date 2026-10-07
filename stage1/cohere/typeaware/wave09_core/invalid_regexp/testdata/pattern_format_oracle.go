package main

import (
	"encoding/json"
	"fmt"
	core "github.com/system-inc/cohere/internal/lint/rules/core"
	"os"
	"strings"
	"unicode/utf16"
)

func written(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != 92 {
			out.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			hi, lo := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, hi, lo)
		}
	}
	return out.String()
}
func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows []core.Wave09PatternInput
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	if len(os.Args) > 2 && os.Args[2] == "--inputs" {
		for i, row := range rows {
			rows[i] = core.Wave09PatternCompileInput(row.Pattern, row.Flags)
		}
		data, err = json.Marshal(rows)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
		return
	}
	for _, row := range rows {
		fmt.Println(core.Wave09PatternCompileInput(row.Pattern, row.Flags).Canonical)
		fmt.Println(written(core.Wave09PatternMessage(row.Pattern, row.Flags)))
	}
}
