// Direct Go baseline for identical bridge fact work, not the findings oracle.
// It deliberately shares Inspect to isolate crossing and native string cost.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
)

func main() {
	commandArguments := os.Args[1:]
	if len(commandArguments) != 7 {
		panic("usage: cost config file start end kind question count")
	}
	first, parseError := strconv.ParseUint(commandArguments[2], 10, 64)
	if parseError != nil {
		panic(parseError)
	}
	last, parseError := strconv.ParseUint(commandArguments[3], 10, 64)
	if parseError != nil {
		panic(parseError)
	}
	count, parseError := strconv.Atoi(commandArguments[6])
	if parseError != nil || count < 2 {
		panic("invalid count")
	}
	started := time.Now()
	program, openError := bridge.Open(commandArguments[0], []string{commandArguments[1]})
	if openError != nil {
		panic(openError)
	}
	load := time.Since(started)
	question := commandArguments[5]
	if strings.HasPrefix(question, "property-info:") {
		answer, inspectError := program.Inspect(commandArguments[1], first, last, commandArguments[4], "raw-shape")
		if inspectError != nil {
			panic(inspectError)
		}
		// The first six fields (header, flags, count, first root) are ASCII.
		var root string
		for field := 0; field < 6; field++ {
			newline := strings.IndexByte(answer, '\n')
			length, parseError := strconv.Atoi(answer[:newline])
			if parseError != nil {
				panic(parseError)
			}
			root = answer[newline+1 : newline+1+length]
			answer = answer[newline+1+length:]
		}
		question = "property-info\n" + root + "\n" + strings.TrimPrefix(question, "property-info:")
	}
	var total, initial time.Duration
	units := 0
	for index := 0; index < count; index++ {
		before := time.Now()
		answer, inspectError := program.Inspect(commandArguments[1], first, last, commandArguments[4], question)
		if inspectError != nil {
			panic(inspectError)
		}
		units += len(utf16.Encode([]rune(answer)))
		elapsed := time.Since(before)
		total += elapsed
		if index == 0 {
			initial = elapsed
		}
	}
	fmt.Println(units)
	fmt.Fprintf(
		os.Stderr,
		"factcost: load_ns=%d query_ns=%d queries=%d first_query_ns=%d\n",
		load.Nanoseconds(),
		total.Nanoseconds(),
		count,
		initial.Nanoseconds(),
	)
}
