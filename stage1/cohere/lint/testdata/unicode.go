package main

import (
	"fmt"
	"strconv"
	"unicode"
)

func main() {
	fmt.Println("// Generated from Go unicode.SimpleFold and strconv.IsPrint; Unicode version " + unicode.Version + ".\n// A data table, not a runtime oracle. Regenerate with testdata/unicode.go.\nconst folds = new Map<number, number>([")
	for r := rune(0); r <= unicode.MaxRune; r++ {
		minimum := r
		for n := unicode.SimpleFold(r); n != r; n = unicode.SimpleFold(n) {
			if n < minimum {
				minimum = n
			}
		}
		if minimum != r {
			fmt.Printf("[%d,%d],\n", r, minimum)
		}
	}
	fmt.Println("]);\n// ASCII case folding needs no table lookup; other scalars retain Go's cycles.\nexport function foldPoint(point: number): number {\n if(point < 128) { return point >= 97 && point <= 122 ? point - 32 : point; }\n return folds.get(point) ?? point;\n}\nconst unprintable: number[] = [")
	start := -1
	for r := 0; r <= unicode.MaxRune+1; r++ {
		bad := r <= unicode.MaxRune && !strconv.IsPrint(rune(r))
		if bad && start < 0 {
			start = r
		}
		if !bad && start >= 0 {
			fmt.Printf("%d,%d,\n", start, r-1)
			start = -1
		}
	}
	fmt.Println("];\nexport function printable(point: number): boolean {\n let left = 0; let right = unprintable.length / 2;\n while(left < right) {const middle = Math.floor((left + right) / 2);const start=unprintable[middle*2] ?? 0;const end=unprintable[middle*2+1] ?? 0;if(point < start){right=middle;}else if(point>end){left=middle+1;}else{return false;}}\n return true;\n}")
	fmt.Println(`
export function foldedRange(character: string, first: number, last: number): boolean {
    const point = character.codePointAt(0) ?? 0;
    if(point >= first && point <= last) { return true; }
    const minimum = folds.get(point) ?? point;
    if(minimum >= first && minimum <= last) { return true; }
    for(const key of folds.keys()) {
        if(key >= first && key <= last && folds.get(key) === minimum) { return true; }
    }
    return false;
}
`)
}
