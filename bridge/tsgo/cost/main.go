// Direct Go baseline for identical bridge fact work, not the findings oracle.
// It deliberately shares Inspect to isolate crossing and native string cost.
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"unicode/utf16"

	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
)

func main() {
	args := os.Args[1:]
	if len(args) != 7 {
		panic("usage: cost config file start end kind question count")
	}
	first, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		panic(err)
	}
	last, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		panic(err)
	}
	count, err := strconv.Atoi(args[6])
	if err != nil || count < 2 {
		panic("invalid count")
	}
	started := time.Now()
	program, err := bridge.Open(args[0], []string{args[1]})
	if err != nil {
		panic(err)
	}
	load := time.Since(started)
	var total, initial time.Duration
	units := 0
	for index := 0; index < count; index++ {
		before := time.Now()
		answer, err := program.Inspect(args[1], first, last, args[4], args[5])
		if err != nil {
			panic(err)
		}
		units += len(utf16.Encode([]rune(answer)))
		elapsed := time.Since(before)
		total += elapsed
		if index == 0 {
			initial = elapsed
		}
	}
	fmt.Println(units)
	fmt.Fprintf(os.Stderr, "factcost: load_ns=%d query_ns=%d queries=%d first_query_ns=%d\n", load.Nanoseconds(), total.Nanoseconds(), count, initial.Nanoseconds())
}
