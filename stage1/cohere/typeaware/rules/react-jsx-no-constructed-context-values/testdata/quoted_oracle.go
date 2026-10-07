package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func main() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for point := 0; point <= 0x10ffff; point++ {
		if point >= 0xd800 && point <= 0xdfff {
			continue
		}
		fmt.Fprintln(w, strconv.Quote(string(rune(point))))
	}
	for _, name := range []string{"", "élève", "a\u200cb", "x\u200dy", "𐐀a", "quote\"slash\\", "\n\t"} {
		fmt.Fprintln(w, strconv.Quote(name))
	}
}
