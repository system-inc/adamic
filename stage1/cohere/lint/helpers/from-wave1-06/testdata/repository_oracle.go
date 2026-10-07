package main

import (
	"fmt"
	tailwind "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
)

func main() {
	if os.Args[1] == "definitions" {
		fmt.Print(string(tailwind.AdamicWave06Definitions()))
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Print(string(tailwind.AdamicWave06RepositoryCases(data)))
}
