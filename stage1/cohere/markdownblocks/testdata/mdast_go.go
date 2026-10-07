package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/format/markdown/mdast"
	"os"
	"strings"
)

func main() {
	if os.Args[1] == "--error" {
		fmt.Println(mdast.AdamicMalformedEvents(os.Args[2]))
		return
	}
	if os.Args[1] == "--events" {
		mdast.AdamicMdastFromTransport(os.Args[2])
		return
	}
	data, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	f, e := os.Create(os.Args[2])
	if e != nil {
		panic(e)
	}
	defer f.Close()
	fixture := bufio.NewWriter(f)
	defer fixture.Flush()
	output := bufio.NewWriter(os.Stdout)
	defer output.Flush()
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		var c struct{ Name, Text string }
		if e = json.Unmarshal([]byte(line), &c); e != nil {
			panic(e)
		}
		text := strings.TrimPrefix(strings.ReplaceAll(strings.ReplaceAll(c.Text, "\r\n", "\n"), "\r", "\n"), "\ufeff")
		a, b := mdast.AdamicMdastFixture(text)
		fmt.Fprint(fixture, a)
		fmt.Fprintln(output, b)
	}
}
