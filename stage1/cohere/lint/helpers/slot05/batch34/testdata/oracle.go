package main

import (
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"os"
)

func main() {
	r, err := esregexp.Compile(os.Args[1], "u")
	if err != nil {
		panic(err)
	}
	fmt.Println(r.Test("a"))
}
