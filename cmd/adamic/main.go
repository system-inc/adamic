// Command adamic is the Adamic compiler, stage 0.
//
//	adamic types <file.a|file.ts>...        check, and print every declaration's proven type
//	adamic c <file.a|file.ts>               print the program as C
//	adamic js <file.a|file.ts>              print the program as JavaScript
//	adamic build <file.a|file.ts> -o <out>  compile it to a native binary
//
// build --count makes a binary that counts its allocations, frees, retains and releases and writes
// them to stderr as it exits, for measuring what the memory model costs.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const usage = `usage:
  adamic types <file.a|file.ts>...
  adamic c <file.a|file.ts> [--explain-checks]
  adamic js <file.a|file.ts> [--explain-checks]
  adamic build [--target wasm32-wasi] <file.a|file.ts> -o <out> [--count] [--sanitize] [--tsgo <archive>] [--explain-checks]`

func main() {
	os.Exit(run(os.Args[1:]))
}

// run is the command, returning its exit code: 0 on success, 1 when the program is refused or can't
// be compiled, 2 for a usage error.
func run(arguments []string) int {
	// Explanation goes to stderr so C/JavaScript output remains usable.
	explain := false
	filtered := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		if argument == "--explain-checks" {
			explain = true
		} else {
			filtered = append(filtered, argument)
		}
	}
	if explain {
		if len(filtered) == 2 && (filtered[0] == "c" || filtered[0] == "js") {
			return renderWithChecks(filtered[0], filtered[1])
		}
		if len(filtered) > 0 && filtered[0] == "build" {
			arguments = append(filtered, "--explain-checks")
		} else {
			fmt.Fprintln(os.Stderr, usage)
			return 2
		}
	} else {
		arguments = filtered
	}
	switch {
	case len(arguments) == 3 && (arguments[0] == "c" || arguments[0] == "js") && arguments[2] == "--explain-checks":
		return renderWithChecks(arguments[0], arguments[1])
	case len(arguments) >= 5 && arguments[0] == "build" && arguments[2] == "--explain-checks" && arguments[3] == "-o":
		return build(arguments[1], arguments[4], append([]string{"--explain-checks"}, arguments[5:]...))
	case len(arguments) >= 2 && arguments[0] == "types":
		program, code := check(arguments[1:])
		if program == nil {
			return code
		}
		for _, declaration := range program.Declarations(context.Background()) {
			fmt.Printf("%s:%d:%d: %s %s: %s\n", relative(declaration.File), declaration.Line, declaration.Column, declaration.Kind, declaration.Name, declaration.Type)
		}
		return 0
	case len(arguments) == 2 && arguments[0] == "c":
		lowered, code := compile(arguments[1])
		if lowered == nil {
			return code
		}
		if native.UsesTSGo(lowered) {
			fmt.Fprintln(os.Stderr, "adamic: tsgo requires a native build with --tsgo <archive>")
			return 1
		}
		fmt.Print(native.C(lowered))
		return 0
	case len(arguments) == 2 && arguments[0] == "js":
		lowered, code := compile(arguments[1])
		if lowered == nil {
			return code
		}
		if native.UsesTSGo(lowered) {
			fmt.Fprintln(os.Stderr, "adamic: tsgo is an external native checker library; JavaScript is not supported")
			return 1
		}
		fmt.Print(javascript.JavaScript(lowered))
		return 0
	case len(arguments) >= 6 && arguments[0] == "build" && arguments[1] == "--target" && arguments[4] == "-o":
		return build(arguments[3], arguments[5], append([]string{"--target", arguments[2]}, arguments[6:]...))
	case len(arguments) >= 4 && arguments[0] == "build" && arguments[2] == "-o":
		return build(arguments[1], arguments[3], arguments[4:])
	}
	fmt.Fprintln(os.Stderr, usage)
	return 2
}

// check loads a program, printing why when it can't. A nil program comes with the exit code.
func check(paths []string) (*load.Program, int) {
	program, err := load.Load(paths)
	if err != nil {
		var checkError *load.CheckError
		if errors.As(err, &checkError) {
			for _, diagnostic := range checkError.Diagnostics {
				fmt.Fprintln(os.Stderr, relative(diagnostic))
			}
			return nil, 1
		}
		fmt.Fprintf(os.Stderr, "adamic: %v\n", err)
		return nil, 1
	}
	return program, 0
}

// compile checks and lowers one program.
func compile(path string) (*ir.Program, int) {
	return compileLibrary(path, false)
}

func compileLibrary(path string, tsgo bool) (*ir.Program, int) {
	program, code := check([]string{path})
	if program == nil {
		return nil, code
	}
	if tsgo {
		program.EnableTSGo()
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		fmt.Fprintln(os.Stderr, relative("adamic: "+err.Error()))
		return nil, 1
	}
	return lowered, 0
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
