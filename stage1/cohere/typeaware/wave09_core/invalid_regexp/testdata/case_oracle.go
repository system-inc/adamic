package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"os"
	"unicode"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--tables" {
		data := struct {
			Version     string
			Upper, Fold []int32
		}{Version: unicode.Version}
		for value := rune(0); value <= unicode.MaxRune; value++ {
			if upper := unicode.ToUpper(value); upper != value {
				data.Upper = append(data.Upper, value, upper)
			}
			if folded := unicode.SimpleFold(value); folded != value {
				data.Fold = append(data.Fold, value, folded)
			}
		}
		if err := json.NewEncoder(os.Stdout).Encode(data); err != nil {
			panic(err)
		}
		return
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for value := rune(-1); value <= unicode.MaxRune+1; value++ {
		fmt.Fprintf(out, "%d\t%d\n", esregexp.Canonicalize(value, false), esregexp.Canonicalize(value, true))
	}
}
