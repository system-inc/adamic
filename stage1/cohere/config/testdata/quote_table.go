// Regenerate the pinned Go strconv.IsPrint data used by canonical config records:
// go run stage1/cohere/config/testdata/quote_table.go > stage1/cohere/config/quote_table.ts
package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("// Go strconv.IsPrint Unicode table. Static character data, not a host parser or resolver.")
	fmt.Println("import { panic } from 'adamic';")
	fmt.Println("const ranges: readonly (readonly number[])[] = [")
	start := -1
	for point := 0; point <= 0x110000; point++ {
		printable := point < 0x110000 && strconv.IsPrint(rune(point))
		if printable && start < 0 {
			start = point
		}
		if !printable && start >= 0 {
			fmt.Printf("\t[%d, %d],\n", start, point-1)
			start = -1
		}
	}
	fmt.Println("];")
	fmt.Println("export function goIsPrint(point: number): boolean {")
	fmt.Println("\tif (point >= 32 && point <= 126) { return true; }")
	fmt.Println("\tlet low = 0; let high = ranges.length;")
	fmt.Println("\twhile (low < high) { const middle = Math.floor((low + high) / 2); const range = ranges[middle] ?? panic('missing Unicode range'); if (point < (range[0] ?? 0)) { high = middle; } else if (point > (range[1] ?? 0)) { low = middle + 1; } else { return true; } }")
	fmt.Println("\treturn false;\n}")
}
