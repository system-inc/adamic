// Run under cohere's pinned Go module. Go regexp is the independent oracle.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"unicode/utf16"
)

type Row struct {
	Pattern *string `json:"go_pattern"`
}

func units(s string) []uint16 { return utf16.Encode([]rune(s)) }
func main() {
	table, _ := os.ReadFile(os.Args[1])
	data, _ := os.ReadFile(os.Args[2])
	var rows []Row
	var corpus []string
	if e := json.Unmarshal(table, &rows); e != nil {
		panic(e)
	}
	if e := json.Unmarshal(data, &corpus); e != nil {
		panic(e)
	}
	row := 0
	for _, r := range rows {
		if r.Pattern == nil {
			continue
		}
		pattern := regexp.MustCompile(*r.Pattern)
		for input, text := range corpus {
			for _, span := range pattern.FindAllStringIndex(text, -1) {
				value := text[span[0]:span[1]]
				fmt.Printf("%d:%d:%d:%d:", row, input, len(units(text[:span[0]])), len(units(text[:span[1]])))
				for _, u := range units(value) {
					fmt.Printf("%d,", u)
				}
				fmt.Println()
			}
		}
		row++
	}
}
