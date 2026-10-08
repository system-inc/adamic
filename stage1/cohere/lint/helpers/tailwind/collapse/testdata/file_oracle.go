package main

import (
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
)

func main() {
	value, found, err := collapse.AdamicLoadFileValue(os.Args[1])
	if err != nil {
		panic(err)
	}
	if !found {
		panic("missing Go theme value")
	}
	for _, value := range []byte(value) {
		fmt.Println(value)
	}
}
