// Build the instrumented emitted C with Adamic's ordinary release options.
package main

import (
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/native"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: build.go source.c output")
		os.Exit(2)
	}
	source, err := os.ReadFile(os.Args[1])
	if err == nil {
		err = native.Build(string(source), os.Args[2], native.Options{Release: true})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
