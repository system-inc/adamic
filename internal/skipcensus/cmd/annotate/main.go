// annotate migrates a historical JSON census to source comments using go/ast.
package main

import (
	"flag"
	"fmt"
	"github.com/system-inc/adamic/internal/skipcensus"
	"os"
)

func main() {
	root := flag.String("root", ".", "repository root")
	table := flag.String("table", "", "historical table to convert")
	flag.Parse()
	if *table == "" {
		fmt.Fprintln(os.Stderr, "-table is required")
		os.Exit(1)
	}
	f, err := os.Open(*table)
	if err != nil {
		panic(err)
	}
	rows, err := skipcensus.Load(f)
	f.Close()
	if err != nil {
		panic(err)
	}
	count, err := skipcensus.Annotate(*root, rows)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("converted %d sites; class and provider preserved for every row\n", count)
}
