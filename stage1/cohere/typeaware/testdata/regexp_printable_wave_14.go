package main

import (
	"encoding/json"
	"os"
	"strconv"
)

func main() {
	var ranges []int
	start := -1
	for code := 0; code <= 0x110000; code++ {
		printed := code < 0x110000 && strconv.IsPrint(rune(code))
		if printed && start < 0 {
			start = code
		}
		if !printed && start >= 0 {
			ranges = append(ranges, start, code-1)
			start = -1
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(ranges); err != nil {
		panic(err)
	}
}
