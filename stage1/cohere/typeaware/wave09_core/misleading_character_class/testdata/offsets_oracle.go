package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/ecmascript/literal"
	"os"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows [][2]string
	if err := json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for _, row := range rows {
		mapping := literal.CookedToRaw(row[0], row[1])
		if mapping == nil {
			fmt.Println("nil")
			continue
		}
		for i, offset := range mapping {
			if i > 0 {
				fmt.Print(",")
			}
			fmt.Print(offset)
		}
		fmt.Println()
	}
}
