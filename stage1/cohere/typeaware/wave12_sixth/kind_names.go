//go:build ignore

// Independent metadata oracle: enumerate the same pinned AST names the registry validates.
package main

import (
	"encoding/json"
	"os"
	"strings"

	syntax "github.com/microsoft/TypeScript/tsc/shim/ast"
)

func main() {
	var names []string
	for kind := syntax.Kind(0); kind < syntax.KindCount; kind++ {
		names = append(names, strings.TrimPrefix(kind.String(), "Kind"))
	}
	if err := json.NewEncoder(os.Stdout).Encode(names); err != nil {
		panic(err)
	}
}
