package main

import (
	"fmt"
	"github.com/system-inc/cohere/internal/format/css"
	"os"
	"strings"
)

func main() {
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	decode := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`)
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	for _, line := range strings.Split(string(b), "\n") {
		if line != "" {
			fmt.Println(encode.Replace(css.AdamicAdjustStrings(decode.Replace(line[1:]), line[0] == 's')))
		}
	}
}
