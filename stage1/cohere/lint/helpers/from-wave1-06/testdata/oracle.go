package main

import (
	"fmt"
	tailwind "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	for _, key := range strings.Split(string(data), "\n") {
		if key != "" {
			fmt.Print(tailwind.AdamicWave06NewTheme(key))
		}
	}
}
