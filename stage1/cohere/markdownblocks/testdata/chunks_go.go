package main

import (
	"bufio"
	"fmt"
	"github.com/system-inc/cohere/internal/format/markdown/micromark"
	"os"
	"strconv"
	"strings"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		units := []uint16{}
		if line != "-" {
			for _, field := range strings.Split(line, ",") {
				unit, e := strconv.ParseUint(field, 10, 16)
				if e != nil {
					panic(e)
				}
				units = append(units, uint16(unit))
			}
		}
		fmt.Fprintln(out, micromark.AdamicInputChunks(units))
	}
}
