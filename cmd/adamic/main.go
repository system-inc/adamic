// Command adamic is the Adamic compiler. Stage 0 can do one thing: check a program and print the
// type it proved for every declaration.
//
//	adamic types <file.a|file.ts>...
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/system-inc/adamic/internal/load"
)

const usage = "usage: adamic types <file.a|file.ts>..."

func main() {
	os.Exit(run(os.Args[1:]))
}

// run is the command, returning its exit code: 0 when the program checks, 1 when it doesn't, 2 for
// a usage error.
func run(arguments []string) int {
	if len(arguments) < 2 || arguments[0] != "types" {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	program, err := load.Load(arguments[1:])
	if err != nil {
		var checkError *load.CheckError
		if errors.As(err, &checkError) {
			for _, diagnostic := range checkError.Diagnostics {
				fmt.Fprintln(os.Stderr, relative(diagnostic))
			}
			return 1
		}
		fmt.Fprintf(os.Stderr, "adamic: %v\n", err)
		return 1
	}
	for _, declaration := range program.Declarations(context.Background()) {
		fmt.Printf("%s:%d:%d: %s %s: %s\n", relative(declaration.File), declaration.Line, declaration.Column, declaration.Kind, declaration.Name, declaration.Type)
	}
	return 0
}

// relative shortens a path under the working directory, so output reads the way the command was
// typed. Anything else passes through unchanged.
func relative(text string) string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return text
	}
	prefix := filepath.ToSlash(workingDirectory) + "/"
	if len(text) > len(prefix) && text[:len(prefix)] == prefix {
		return text[len(prefix):]
	}
	return text
}
