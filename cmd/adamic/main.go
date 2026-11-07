// Command adamic is the Adamic compiler, stage 0.
//
//	adamic types <file.a|file.ts>...        check, and print every declaration's proven type
//	adamic c <file.a|file.ts>               print the program as C
//	adamic js <file.a|file.ts>              print the program as JavaScript
//	adamic worker <entry.a> --out <dir>    generate a Cloudflare Worker
//	adamic build <file.a|file.ts> -o <out>  compile it to a native binary
//	adamic --help | -h                      print the usage
//	adamic --version                        name the commit, the Go and the clang a bug report needs
//
// build --count makes a binary that counts its allocations, frees, retains and releases and writes
// them to stderr as it exits, for measuring what the memory model costs.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

const usage = `usage:
  adamic types <file.a|file.ts>...
  adamic c <file.a|file.ts> [--explain-checks]
  adamic worker <entry.a> --out <directory> [--wasm <name>[,<name>...]]
  adamic js <file.a|file.ts> [--explain-checks]
  adamic build [--target wasm32-wasi] [--reactor] <file.a|file.ts> -o <out> [--count] [--debug] [--sanitize] [--tsgo <archive>] [--export <name>]... [--abi-json <path>] [--explain-checks]
  adamic --help | -h
  adamic --version`

func main() {
	os.Exit(run(os.Args[1:]))
}

// run is the command, returning its exit code: 0 on success, 1 when the program is refused or can't
// be compiled, 2 for a usage error, 4 when the compiler itself panicked. Usage asked for with --help is no error: it
// goes to stdout and exits 0, as --version does.
func run(arguments []string) (code int) {
	defer reportCompilerBug(&code, os.Stderr)
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
	case len(arguments) == 1 && (arguments[0] == "--help" || arguments[0] == "-h"):
		fmt.Println(usage)
		return 0
	case len(arguments) == 1 && arguments[0] == "--version":
		writeVersion(os.Stdout)
		return 0
	case len(arguments) >= 2 && arguments[0] == "worker":
		return workerCommand(arguments[1:])
	case len(arguments) >= 2 && arguments[0] == "types":
		program, code := check(arguments[1:])
		if program == nil {
			return code
		}
		for _, declaration := range program.Declarations(context.Background()) {
			fmt.Printf(
				"%s:%d:%d: %s %s: %s\n",
				relative(declaration.File),
				declaration.Line,
				declaration.Column,
				declaration.Kind,
				declaration.Name,
				declaration.Type,
			)
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
	case len(arguments) >= 7 &&
		arguments[0] == "build" &&
		arguments[1] == "--target" &&
		arguments[3] == "--reactor" &&
		arguments[5] == "-o":
		return build(
			arguments[4],
			arguments[6],
			append([]string{"--target", arguments[2], "--reactor"}, arguments[7:]...),
		)
	case len(arguments) >= 6 && arguments[0] == "build" && arguments[1] == "--target" && arguments[4] == "-o":
		return build(arguments[3], arguments[5], append([]string{"--target", arguments[2]}, arguments[6:]...))
	case len(arguments) >= 4 && arguments[0] == "build" && arguments[2] == "-o":
		return build(arguments[1], arguments[3], arguments[4:])
	}
	fmt.Fprintln(os.Stderr, usage)
	return 2
}

// reportCompilerBug turns a panic on the command's goroutine into exit 4 with its stack. Unrecovered, Go
// exits 2, which a script reads as its own usage mistake rather than the compiler's. A panic on another
// goroutine, such as native's compiles, still ends the process with 2 (#4mwb9et).
func reportCompilerBug(code *int, output io.Writer) {
	if failure := recover(); failure != nil {
		fmt.Fprintf(output, "adamic: compiler bug: %v\n%s", failure, debug.Stack())
		*code = 4
	}
}

// writeVersion names what a bug report needs, a line each: the commit adamic was built from, the Go that built it, and
// the clang a native build would run, resolved by toolchain.Path as internal/native's runtimeInputs resolves it. A
// clang that can't be found or run is its own line, never a failure, since a report is most wanted where the
// toolchain is broken.
func writeVersion(output io.Writer) {
	information, ok := debug.ReadBuildInfo()
	fmt.Fprintf(output, "adamic %s\n", revision(information, ok))
	fmt.Fprintf(output, "go %s\n", runtime.Version())
	fmt.Fprintln(output, clangVersion())
}

// revision is the commit go build stamped, with (modified) when the tree had changes, or unknown when nothing was
// stamped (go run, go test, -buildvcs=false, or a source tree without its .git).
func revision(information *debug.BuildInfo, ok bool) string {
	if !ok {
		return "unknown"
	}
	commit, modified := "", false
	for _, setting := range information.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	if commit == "" {
		return "unknown"
	}
	if modified {
		return commit + " (modified)"
	}
	return commit
}

// clangVersion is `clang <path>: <its --version's first line>`, or `clang: not found (<why>)`.
func clangVersion() string {
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		return fmt.Sprintf("clang: not found (%v)", pathError)
	}
	report, outputError := toolchain.Command(compiler, "--version").Output()
	if outputError != nil {
		return fmt.Sprintf("clang: not found (%v)", outputError)
	}
	first, _, _ := strings.Cut(string(report), "\n")
	return fmt.Sprintf("clang %s: %s", compiler, strings.TrimSpace(first))
}

// check loads a program, printing why when it can't. A nil program comes with the exit code.
func check(paths []string) (*load.Program, int) {
	program, loadError := load.Load(paths)
	if loadError != nil {
		var checkError *load.CheckError
		if errors.As(loadError, &checkError) {
			for _, diagnostic := range checkError.Diagnostics {
				fmt.Fprintln(os.Stderr, relative(diagnostic))
			}
			return nil, 1
		}
		fmt.Fprintf(os.Stderr, "adamic: %v\n", loadError)
		return nil, 1
	}
	return program, 0
}

// compile checks and lowers one program.
func compile(path string) (*ir.Program, int) {
	return compileLibrary(path, false, false)
}

func compileLibrary(path string, tsgo, debug bool) (*ir.Program, int) {
	program, code := check([]string{path})
	if program == nil {
		return nil, code
	}
	if tsgo {
		program.EnableTSGo()
	}
	options := lower.Options{
		ProgramRegion:  programRegionRequested(),
		OwnershipQuery: os.Getenv("ADAMIC_OWNERSHIP_QUERY") == "1",
		Debug:          debug,
	}
	lowered, lowerWithOptionsError := lower.LowerWithOptions(context.Background(), program, options)
	if lowerWithOptionsError != nil {
		fmt.Fprintln(os.Stderr, relative("adamic: "+lowerWithOptionsError.Error()))
		return nil, 1
	}
	return lowered, 0
}

// relative shortens a path under the working directory, so output reads the way the command was
// typed. Anything else passes through unchanged.
func relative(text string) string {
	workingDirectory, workingDirectoryError := os.Getwd()
	if workingDirectoryError != nil {
		return text
	}
	prefix := filepath.ToSlash(workingDirectory) + "/"
	if len(text) > len(prefix) && text[:len(prefix)] == prefix {
		return text[len(prefix):]
	}
	return text
}
