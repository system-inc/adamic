// Run from the repository root with the same Go toolchain as cohere.
package main

import (
	"fmt"
	"os"
	"unicode"
)

func main() {
	f, err := os.Create("stage1/cohere/json/identifierTables.ts")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Fprintln(f, "// Generated from Go Unicode tables, as cohere's JSON parser classifies identifiers.")
	for _, item := range []struct {
		name   string
		tables []*unicode.RangeTable
	}{{"letters", []*unicode.RangeTable{unicode.L, unicode.Nl}}, {"continuations", []*unicode.RangeTable{unicode.Nd, unicode.Mn, unicode.Mc, unicode.Pc}}} {
		fmt.Fprintf(f, "export const %s: readonly (readonly number[])[] = [\n", item.name)
		for _, table := range item.tables {
			for _, r := range table.R16 {
				fmt.Fprintf(f, "[%d, %d, %d],\n", r.Lo, r.Hi, r.Stride)
			}
			for _, r := range table.R32 {
				fmt.Fprintf(f, "[%d, %d, %d],\n", r.Lo, r.Hi, r.Stride)
			}
		}
		fmt.Fprintln(f, "];")
	}
}
