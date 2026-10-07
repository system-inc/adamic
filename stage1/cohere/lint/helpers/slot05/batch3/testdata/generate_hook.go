// Regenerate with the unit's pinned Go toolchain. One public helper, including its private table.
package main

import (
	"fmt"
	"unicode"
)

func main() {
	fmt.Println("import { panic } from 'adamic';")
	fmt.Printf("// Generated from Go unicode.Upper, Unicode %s. Regeneration is checked by the oracle test.\n", unicode.Version)
	fmt.Println("const upperRanges: readonly { readonly lo: number; readonly hi: number; readonly stride: number }[] = [")
	for _, r := range unicode.Upper.R16 {
		fmt.Printf("    {lo: %d, hi: %d, stride: %d},\n", r.Lo, r.Hi, r.Stride)
	}
	for _, r := range unicode.Upper.R32 {
		fmt.Printf("    {lo: %d, hi: %d, stride: %d},\n", r.Lo, r.Hi, r.Stride)
	}
	fmt.Println("];")
	fmt.Print(`
// Go accepts "use" followed by a Unicode uppercase rune; digits are not uppercase.
export function isHookName(name: string): boolean {
    if(!name.startsWith('use') || name.length === 3) { return false; }
    const scalar = name.codePointAt(3);
    if(scalar === undefined) { return false; }
    let low = 0;
    let high = upperRanges.length;
    while(low < high) {
        const middle = Math.floor((low + high) / 2);
        const range = upperRanges[middle];
        if(range === undefined) { panic('uppercase table index out of bounds'); }
        if(scalar > range.hi) { low = middle + 1; } else { high = middle; }
    }
    const range = upperRanges[low];
    if(range === undefined) { return false; }
    return scalar >= range.lo && (scalar - range.lo) % range.stride === 0;
}
`)
}
