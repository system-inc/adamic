package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/system-inc/cohere/internal/format/markdown"
)

func main() {
	if os.Args[1] == "--facts" {
		markdown.AdamicASTFacts(os.Args[2])
		return
	}
	input, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer input.Close()
	fixture, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer fixture.Close()
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 65536), 32*1024*1024)
	escape := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	for scanner.Scan() {
		var value struct {
			Name, Text string
			Mode       string
			TabWidth   int
		}
		if err := json.Unmarshal(scanner.Bytes(), &value); err != nil {
			panic(err)
		}
		if value.TabWidth == 0 {
			value.TabWidth = 4
		}
		stream, want, err := markdown.AdamicASTFixture(value.Text, value.TabWidth, value.Mode)
		if err != nil {
			panic(err)
		}
		if _, err := fixture.WriteString(stream); err != nil {
			panic(err)
		}
		fmt.Fprintln(out, escape.Replace(want))
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
