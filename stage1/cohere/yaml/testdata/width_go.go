package main

import (
	"bufio"
	"fmt"
	"github.com/system-inc/cohere/internal/format/doc"
	"os"
	"strconv"
	"strings"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		var text strings.Builder
		for _, value := range strings.Split(scan.Text(), ",") {
			point, err := strconv.Atoi(value)
			if err != nil {
				panic(err)
			}
			text.WriteRune(rune(point))
		}
		fmt.Fprintln(out, doc.StringWidth(text.String()))
	}
	if err := scan.Err(); err != nil {
		panic(err)
	}
}
