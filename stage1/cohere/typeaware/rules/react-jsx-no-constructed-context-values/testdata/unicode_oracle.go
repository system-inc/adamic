package main

import (
	"fmt"
	"unicode"
)

func main() {
	for point := rune(0); point <= 0x10ffff; point++ {
		if unicode.IsUpper(point) {
			fmt.Println(point)
		}
	}
}
