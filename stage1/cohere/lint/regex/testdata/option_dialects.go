package main

import (
	"encoding/json"
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"os"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var cases [][2]string
	if err := json.Unmarshal(data, &cases); err != nil {
		panic(err)
	}
	for _, pair := range cases {
		pattern, err := esregexp.Compile(pair[0], "u")
		if err != nil {
			fmt.Printf("%s\tSyntaxError\n", pair[0])
			continue
		}
		matched, err := pattern.TestOrError(pair[1])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\t%t\n", pair[0], matched)
	}
}
