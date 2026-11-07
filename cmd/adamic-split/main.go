// Command adamic-split prints the IR proof for each Wasm boundary decision.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/split"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
func run(args []string, output, errors io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(errors, "usage: adamic-split <file.a>")
		return 2
	}
	source, loadError := load.Load(args)
	if loadError != nil {
		fmt.Fprintln(errors, loadError)
		return 1
	}
	program, lowerError := lower.Lower(context.Background(), source)
	if lowerError != nil {
		fmt.Fprintln(errors, lowerError)
		return 1
	}
	if printError := split.Print(output, split.Analyze(program)); printError != nil {
		fmt.Fprintln(errors, printError)
		return 1
	}
	return 0
}
