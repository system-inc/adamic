package main

import (
	"bufio"
	"fmt"
	"github.com/system-inc/cohere/internal/format/doc"
	"os"
	"strings"
)

func decode(text string) string {
	var result strings.Builder
	for i := 0; i < len(text); i++ {
		if text[i] != '\\' {
			result.WriteByte(text[i])
			continue
		}
		i++
		if i == len(text) {
			panic("escape")
		}
		switch text[i] {
		case 'n':
			result.WriteByte('\n')
		case 'r':
			result.WriteByte('\r')
		case 't':
			result.WriteByte('\t')
		default:
			result.WriteByte(text[i])
		}
	}
	return result.String()
}
func main() {
	input, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer input.Close()
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 65536), 32*1024*1024)
	for scanner.Scan() {
		fmt.Println(doc.StringWidth(decode(strings.TrimPrefix(scanner.Text(), "x"))))
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
