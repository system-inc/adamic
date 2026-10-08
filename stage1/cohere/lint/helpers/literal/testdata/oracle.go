//go:build lintoracle

package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/ecmascript/literal"
	"os"
	"strconv"
)

func main() {
	b, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	var rows []struct{ Kind, A, B, Want string }
	if e = json.Unmarshal(b, &rows); e != nil {
		panic(e)
	}
	for _, r := range rows {
		if r.Kind == "tail" {
			fmt.Println(literal.AdamicTail(r.A))
		} else if r.Kind == "width" {
			i, e := strconv.Atoi(r.B)
			if e != nil {
				panic(e)
			}
			fmt.Println(literal.AdamicWidth(r.A, i))
		} else if r.Kind == "mapping" {
			b, e := json.Marshal(literal.CookedToRaw(r.A, r.B))
			if e != nil {
				panic(e)
			}
			fmt.Println(string(b))
		} else {
			fmt.Println(literal.AdamicProduced(r.A, r.B))
		}
	}
}
