package main

import (
	"encoding/json"
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
)

func main() {
	p, err := esregexp.Compile("^_", "u")
	if err != nil {
		panic(err)
	}
	bytes, err := json.Marshal([]*esregexp.RegExp{p})
	if err != nil {
		panic(err)
	}
	fmt.Printf("source=%s captured=%s\n", p.Source(), bytes)
}
