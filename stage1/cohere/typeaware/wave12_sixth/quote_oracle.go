//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"unicode"
	"unicode/utf16"
)

func written(s string) string {
	result := ""
	for _, unit := range utf16.Encode([]rune(s)) {
		if unit >= 32 && unit <= 126 && unit != 92 {
			result += string(rune(unit))
		} else {
			result += fmt.Sprintf("\\u%04x", unit)
		}
	}
	return result
}
func main() {
	points := map[rune]bool{}
	add := func(n uint32) {
		if n <= unicode.MaxRune && !(n >= 0xd800 && n <= 0xdfff) {
			points[rune(n)] = true
		}
	}
	for n := uint32(0); n < 256; n++ {
		add(n)
	}
	for _, table := range unicode.PrintRanges {
		for _, r := range table.R16 {
			for _, n := range []uint32{uint32(r.Lo) - 1, uint32(r.Lo), uint32(r.Lo) + uint32(r.Stride), uint32(r.Hi) - 1, uint32(r.Hi), uint32(r.Hi) + 1} {
				add(n)
			}
		}
		for _, r := range table.R32 {
			for _, n := range []uint32{r.Lo - 1, r.Lo, r.Lo + r.Stride, r.Hi - 1, r.Hi, r.Hi + 1} {
				add(n)
			}
		}
	}
	order := []int{}
	for n := range points {
		order = append(order, int(n))
	}
	sort.Ints(order)
	names := []string{"", "data.value", "a\\b\"c", "테스트", "a\u200db", "a\n\tb"}
	for _, n := range order {
		names = append(names, string(rune(n)))
	}
	expected := ""
	for _, name := range names {
		expected += written(strconv.Quote(name)) + "\n"
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Version  string
		Names    []string
		Expected string
	}{unicode.Version, names, expected}); err != nil {
		panic(err)
	}
}
