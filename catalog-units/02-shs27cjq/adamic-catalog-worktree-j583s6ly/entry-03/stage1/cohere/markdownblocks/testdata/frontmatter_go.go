package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/system-inc/cohere/internal/format/markdown/mdast"
)

func main() {
	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	decode := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`)
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	for _, line := range strings.Split(string(content), "\n") {
		if line == "" {
			continue
		}
		frontMatter, rest := mdast.ParseFrontMatter(decode.Replace(line[1:]))
		if frontMatter == nil {
			fmt.Println("0\t" + encode.Replace(rest))
			continue
		}
		explicit := "0"
		if frontMatter.ExplicitLanguage != nil {
			explicit = "1" + encode.Replace(*frontMatter.ExplicitLanguage)
		}
		fmt.Println(strings.Join([]string{"1", encode.Replace(frontMatter.Language), explicit,
			encode.Replace(frontMatter.Value), encode.Replace(frontMatter.StartDelimiter),
			encode.Replace(frontMatter.EndDelimiter), encode.Replace(frontMatter.Raw), encode.Replace(rest)}, "\t"))
	}
}
