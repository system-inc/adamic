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
	data, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		units := []uint16{}
		if line != "-" {
			for _, v := range strings.Split(line, ",") {
				n, e := strconv.ParseUint(v, 10, 16)
				if e != nil {
					panic(e)
				}
				units = append(units, uint16(n))
			}
		}
		fmt.Fprintln(out, micromark.AdamicKernelEvents(units))
	}
}
