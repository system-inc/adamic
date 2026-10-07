// Command build is an owned validation adapter for helper branches whose CLI
// predates --sanitize. It uses the existing native builder unchanged.
package main

import (
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "expected C source and output path")
		os.Exit(2)
	}
	source, err := os.ReadFile(os.Args[1])
	if err == nil {
		err = native.Build(string(source), os.Args[2], native.Options{Sanitize: true})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
