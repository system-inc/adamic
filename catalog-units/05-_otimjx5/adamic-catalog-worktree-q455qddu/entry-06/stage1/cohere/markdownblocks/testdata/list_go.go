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
	source, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer source.Close()
	native, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer native.Close()
	canonical, err := os.Create(os.Args[3])
	if err != nil {
		panic(err)
	}
	defer canonical.Close()
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 65536), 32*1024*1024)
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	for scanner.Scan() {
		var input struct {
			Name string
			Text string
		}
		if err := json.Unmarshal(scanner.Bytes(), &input); err != nil {
			panic(err)
		}
		actual, original, want, err := markdown.AdamicListFixture(input.Text)
		if err != nil {
			panic(err)
		}
		if _, err := native.WriteString(actual); err != nil {
			panic(err)
		}
		if _, err := canonical.WriteString(original); err != nil {
			panic(err)
		}
		fmt.Println(encode.Replace(want))
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
