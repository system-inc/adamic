//go:build ignore

package main

import (
	"encoding/json"
	syntax "github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"strings"
)

func main() {
	values := map[string]int{}
	for kind := syntax.Kind(0); kind < syntax.KindCount; kind++ {
		values[strings.TrimPrefix(kind.String(), "Kind")] = int(kind)
	}
	if err := json.NewEncoder(os.Stdout).Encode(values); err != nil {
		panic(err)
	}
}
