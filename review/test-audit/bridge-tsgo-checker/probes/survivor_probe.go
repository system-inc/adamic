package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/system-inc/adamic/bridge/tsgo/checker"
)

func decode(wire string) ([]string, int) {
	units := utf16.Encode([]rune(wire))
	var result []string
	for cursor := 0; cursor < len(units); {
		end := cursor
		for end < len(units) && units[end] != '\n' {
			end++
		}
		if end == len(units) {
			panic("missing field header")
		}
		count, err := strconv.Atoi(string(utf16.Decode(units[cursor:end])))
		if err != nil {
			panic(err)
		}
		begin := end + 1
		if begin+count > len(units) {
			return result, begin + count - len(units)
		}
		result = append(result, string(utf16.Decode(units[begin:begin+count])))
		cursor = begin + count
	}
	return result, 0
}

func main() {
	directory, err := os.MkdirTemp("", "u005-survivor-")
	if err != nil {
		panic(err)
	}
	// This scratch directory is retained for evidence; no removal is needed.
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "witness\uffff.ts")
	source := "declare const value: \"x\";\nvalue;\n"
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`), 0600); err != nil {
		panic(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		panic(err)
	}
	program, err := checker.Open(config, []string{file})
	if err != nil {
		panic(err)
	}
	start := uint64(strings.LastIndex(source, "\nvalue"))
	end := start + uint64(len("\nvalue"))
	wire, err := program.Inspect(file, start, end, "Identifier", "symbol-origin")
	if err != nil {
		panic(err)
	}
	_, delta := decode(wire)
	fmt.Printf("symbol-origin frame overrun=%d\n", delta)
	wire, err = program.Inspect(file, start, end, "Identifier", "raw-shape")
	if err != nil {
		panic(err)
	}
	values, delta := decode(wire)
	if delta != 0 || len(values) != 19 {
		panic("unexpected raw-shape framing")
	}
	fmt.Printf("strict-null-checks=%s\nroot-count=%s\n", values[2], values[4])
}
