package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"os"
)

type input struct {
	Pattern, Flags string
	AllowEscape    bool
}

func bit(value bool) int {
	if value {
		return 1
	}
	return 0
}
func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows []input
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for i, row := range rows {
		fmt.Printf("case %d\n", i)
		flags := regexsyntax.ParseRegexFlags(row.Flags)
		var spans [][2]int
		ok := regexsyntax.IterateRegexCharacterClasses(row.Pattern, flags, func(start, end int) { spans = append(spans, [2]int{start, end}) })
		fmt.Printf("scan %d\n", bit(ok))
		for _, span := range spans {
			fmt.Printf("span %d %d\n", span[0], span[1])
			elements, end, ok := regexsyntax.ParseRegexCharacterClassWithEnd(row.Pattern, span[0], span[1], flags)
			fmt.Printf("parse %d %d\n", bit(ok), end)
			for _, element := range elements {
				fmt.Printf("%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n", element.Kind, element.Value, bit(element.IsUBrace), bit(element.IsLoneSurrogate), element.Max, bit(element.MaxIsUBrace), element.Start, element.End)
			}
		}
	}
}
