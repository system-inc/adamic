package main

import (
	"fmt"
	"unicode"
)

func main() {
	for _, part := range []bool{false, true} {
		name := "keyStart"
		if part {
			name = "keyPart"
		}
		fmt.Printf("const %s: readonly number[] = [\n", name)
		start := -1
		for c := 0; c <= unicode.MaxRune+1; c++ {
			r := rune(c)
			valid := c <= unicode.MaxRune && (unicode.IsLetter(r) || unicode.Is(unicode.Nl, r) || r == '$' || r == '_' || (part && (unicode.IsDigit(r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Pc, r))))
			if valid && start < 0 {
				start = c
			}
			if !valid && start >= 0 {
				fmt.Printf("    %d, %d,\n", start, c-1)
				start = -1
			}
		}
		fmt.Println("];\n")
	}
}
