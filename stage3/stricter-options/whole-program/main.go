// Command whole-program reports every production option decision by ledger ID.
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
)

//go:embed roots.json
var ledgerRoots []byte

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
	// The same 79 implementation roots recorded by the checker ledger.
	var names []string
	if err := json.Unmarshal(ledgerRoots, &names); err != nil {
		panic(err)
	}
	roots := make([]string, len(names))
	for index, name := range names {
		roots[index] = filepath.Join(os.Args[1], name)
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
