// Command whole-program reports every production option decision by ledger ID.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: whole-program <adapted-tree> <rows.csv>")
		os.Exit(2)
	}
	input, err := os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer input.Close()
	rows, err := load.ReadOptionLedger(input)
	if err != nil {
		panic(err)
	}
	// The entire configured compiler tree is checked, including unimported files.
	roots := []string{}
	entries, err := filepath.Glob(filepath.Join(os.Args[1], "src/compiler/*.ts"))
	if err != nil {
		panic(err)
	}
	roots = append(roots, entries...)
	for _, directory := range []string{"transformers", "transformers/declarations"} {
		entries, err = filepath.Glob(filepath.Join(os.Args[1], "src/compiler", directory, "*.ts"))
		if err != nil {
			panic(err)
		}
		roots = append(roots, entries...)
	}
	_, report, loadErr := load.LoadOptionLedger(roots, os.Args[1], rows)
	if len(report.Rows) != len(rows) {
		fmt.Fprintln(os.Stderr, loadErr)
		os.Exit(1)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		panic(err)
	}
	if loadErr != nil {
		fmt.Fprintln(os.Stderr, loadErr)
	}
	// A complete census succeeds; its remaining-error rows still reject builds.
	if len(report.OrdinaryErrors) > 0 {
		os.Exit(1)
	}
}
