// Generate native tables from the oracle toolchain's Unicode classification.
package main

import (
	"fmt"
	"unicode"
)

func main() {
	fmt.Println("// Generated from Go's Unicode PrintRanges and ToUpper. No lint decisions.\nconst printRanges: readonly number[] = [")
	for _, table := range unicode.PrintRanges {
		for _, r := range table.R16 {
			fmt.Printf("%d,%d,%d,\n", r.Lo, r.Hi, r.Stride)
		}
		for _, r := range table.R32 {
			fmt.Printf("%d,%d,%d,\n", r.Lo, r.Hi, r.Stride)
		}
	}
	fmt.Println("];\nconst upperChanges: readonly number[] = [")
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if unicode.ToUpper(r) != r {
			fmt.Printf("%d,\n", r)
		}
	}
	fmt.Print(`];
export function printable(code: number): boolean {
    for(let at = 0; at < printRanges.length; at += 3) {
        const lo = printRanges[at] ?? 0; const hi = printRanges[at+1] ?? 0; const stride = printRanges[at+2] ?? 1;
        if(code >= lo && code <= hi && (code - lo) % stride === 0) {return true;}
    }
    return false;
}
export function upperChangesCode(code: number): boolean {
    let lo = 0; let hi = upperChanges.length;
    while(lo < hi) {const middle = Math.floor((lo+hi)/2);if((upperChanges[middle] ?? 0) < code) {lo = middle+1;}else {hi = middle;}}
    return upperChanges[lo] === code;
}
`)
}
