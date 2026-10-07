package main

import (
	"fmt"
	tailwind "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
)

func main() {
	if os.Args[1] == "groups" {
		fmt.Print(string(tailwind.AdamicWave06AttachGroups()))
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Print(string(tailwind.AdamicWave06Attach(data, len(os.Args) > 2 && os.Args[2] == "prepare")))
}
