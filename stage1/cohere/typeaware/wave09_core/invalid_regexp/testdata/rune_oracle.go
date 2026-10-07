// Unicode printability and %q are external to Adamic. The table is data, not a lint verdict.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"unicode"
)

func main() {
	if os.Args[1] == "--ranges" {
		var ranges [][2]int
		start := -1
		for value := 0; value <= unicode.MaxRune+1; value++ {
			printable := value <= unicode.MaxRune && unicode.IsPrint(rune(value))
			if printable && start < 0 {
				start = value
			}
			if !printable && start >= 0 {
				ranges = append(ranges, [2]int{start, value - 1})
				start = -1
			}
		}
		data, err := json.Marshal(struct {
			Version string
			Ranges  [][2]int
		}{unicode.Version, ranges})
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var values []int
	if err = json.Unmarshal(data, &values); err != nil {
		panic(err)
	}
	for _, value := range values {
		fmt.Printf("%q\n", rune(value))
	}
}
