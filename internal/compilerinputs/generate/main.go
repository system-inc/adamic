// Command generate refreshes the compiler input table and admission input file.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/system-inc/adamic/internal/compilerinputs"
)

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	table, admission, err := compilerinputs.Generate(*root)
	if err == nil {
		err = os.WriteFile(filepath.Join(*root, "internal/compilerinputs/compilerinputs_table.go"), table, 0644)
	}
	if err == nil {
		err = os.WriteFile(filepath.Join(*root, "cmd/adamic-admission-delta/inputs.txt"), admission, 0644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
