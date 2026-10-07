// Command adamic is the Adamic compiler, stage 0.
//
//	adamic types <file.a|file.ts>...        check, and print every declaration's proven type
//	adamic c <file.a|file.ts>...            print the program as C
//	adamic js <file.a|file.ts>...           print the program as JavaScript
//	adamic build <file.a|file.ts>... -o <out>  compile it to a native binary
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
  adamic c <file.a|file.ts>...
  adamic js <file.a|file.ts>...
  adamic build [--target wasm32-wasi] <file.a|file.ts>... -o <out> [--count] [--sanitize] [--tsgo <archive>]
  adamic build --project <tsconfig.json> --entry <file.a|file.ts> -o <out> [--count] [--sanitize] [--tsgo <archive>]`

func main() {
	os.Exit(run(os.Args[1:]))
}

// run is the command, returning its exit code: 0 on success, 1 when the program is refused or can't
// be compiled, 2 for a usage error.
func run(arguments []string) int {
	switch {
	case len(arguments) >= 2 && arguments[0] == "types":
		program, code := check(arguments[1:])
		if program == nil {
			return code
		}
		for _, declaration := range program.Declarations(context.Background()) {
			fmt.Printf("%s:%d:%d: %s %s: %s\n", relative(declaration.File), declaration.Line, declaration.Column, declaration.Kind, declaration.Name, declaration.Type)
		}
		return 0
	case len(arguments) >= 2 && arguments[0] == "c":
		lowered, code := compileInput(arguments[1:], "", false)
		if lowered == nil {
			return code
		}
		if native.UsesTSGo(lowered) {
			fmt.Fprintln(os.Stderr, "adamic: tsgo requires a native build with --tsgo <archive>")
			return 1
		}
		fmt.Print(native.C(lowered))
		return 0
	case len(arguments) >= 2 && arguments[0] == "js":
		lowered, code := compileInput(arguments[1:], "", false)
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
	case len(arguments) >= 4 && arguments[0] == "build":
		project := ""
		start := 1
		if arguments[1] == "--project" {
			project = arguments[2]
			if len(arguments) < 7 || arguments[3] != "--entry" {
				break
			}
			start = 5
		}
		for index := start; index+1 < len(arguments); index++ {
			if arguments[index] == "-o" {
				if project == "" && index == start || project != "" && index != start {
					break
				}
				return buildInput(projectPaths(arguments, start, index, project), project, arguments[index+1], arguments[index+2:])
			}
		}
	}
	fmt.Fprintln(os.Stderr, usage)
	return 2
}

func projectPaths(arguments []string, start, end int, project string) []string {
	if project != "" {
		return []string{arguments[4]}
	}
	return arguments[start:end]
}

// check loads a program, printing why when it can't. A nil program comes with the exit code.
func check(paths []string) (*load.Program, int) { return checkInput(paths, "") }

func checkInput(paths []string, project string) (*load.Program, int) {
	var program *load.Program
	var err error
	if project != "" {
		entry := ""
		if len(paths) == 1 {
			entry = paths[0]
		}
		program, err = load.LoadProjectEntry(project, entry)
	} else {
		program, err = load.Load(paths)
	}
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
	return compileInput([]string{path}, "", tsgo)
}

func compileInput(paths []string, project string, tsgo bool) (*ir.Program, int) {
	program, code := checkInput(paths, project)
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
