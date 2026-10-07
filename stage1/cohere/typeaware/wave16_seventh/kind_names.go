//go:build ignore

// The listener metadata oracle comes directly from typescript-go's AST.
package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func main() {
	names := make([]string, 0, ast.KindCount)
	for kind := ast.Kind(0); kind < ast.KindCount; kind++ {
		names = append(names, strings.TrimPrefix(kind.String(), "Kind"))
	}
	if err := json.NewEncoder(os.Stdout).Encode(names); err != nil {
		panic(err)
	}
}
