package main

import (
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	for _, action := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if action == "build" {
			fmt.Println(collapse.AdamicNextBuildCount())
		} else {
			fmt.Println(collapse.BuildsSoFar())
		}
	}
}
