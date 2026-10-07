package main

import (
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	"os"
)

func main() {
	if _, err := os.Stdout.Write(tailwind.AdamicReaderForCorpus()); err != nil {
		panic(err)
	}
}
