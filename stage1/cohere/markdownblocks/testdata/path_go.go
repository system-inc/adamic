package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/format/markdown"
	"os"
	"strings"
)

func main() {
	if os.Args[1] == "--key-gap" {
		fmt.Print(markdown.AdamicPathKeyWitness())
		return
	}
	if os.Args[1] == "--facts" {
		markdown.AdamicPathFacts(os.Args[2])
		return
	}
	file, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	output, e := os.Create(os.Args[2])
	if e != nil {
		panic(e)
	}
	defer output.Close()
	stdout := bufio.NewWriter(os.Stdout)
	defer stdout.Flush()
	escape := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	for _, line := range strings.Split(string(file), "\n") {
		if line == "" {
			continue
		}
		var v struct{ Text string }
		if e := json.Unmarshal([]byte(line), &v); e != nil {
			panic(e)
		}
		facts, want, e := markdown.AdamicPathFixture(v.Text)
		if e != nil {
			panic(e)
		}
		if _, e := output.WriteString(facts); e != nil {
			panic(e)
		}
		fmt.Fprintln(stdout, escape.Replace(want))
	}
}
