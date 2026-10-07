package main

import (
	"encoding/json"
	"fmt"
	"unicode"
)

func main() {
	var ranges []int
	start := -1
	for code := 0; code <= 0x110000; code++ {
		printable := code < 0x110000 && unicode.IsPrint(rune(code))
		if printable && start < 0 {
			start = code
		}
		if !printable && start >= 0 {
			ranges = append(ranges, start, code-1)
			start = -1
		}
	}
	var uppercaseInvariant []int
	start = -1
	for code := 0; code <= 0x110000; code++ {
		invariant := code < 0x110000 && unicode.ToUpper(rune(code)) == rune(code)
		if invariant && start < 0 {
			start = code
		}
		if !invariant && start >= 0 {
			uppercaseInvariant = append(uppercaseInvariant, start, code-1)
			start = -1
		}
	}
	b, _ := json.Marshal(struct {
		Version            string
		Ranges             []int
		UppercaseInvariant []int
	}{unicode.Version, ranges, uppercaseInvariant})
	fmt.Println(string(b))
}
