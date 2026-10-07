// Numeric listener contracts from unmodified production Go rule maps.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"sort"
)

type declaration struct {
	Name  string
	Kinds []int
}

func main() {
	names := map[string]bool{"prefer-const": true, "radix": true, "@typescript-eslint/non-nullable-type-assertion-style": true, "no-label-var": true, "no-invalid-regexp": true, "no-misleading-character-class": true}
	var declarations []declaration
	for _, subject := range registry.All() {
		if !names[subject.Name] {
			continue
		}
		entry := declaration{Name: subject.Name}
		for kind := range subject.Run(rule.Context{}, nil) {
			entry.Kinds = append(entry.Kinds, int(kind))
		}
		sort.Ints(entry.Kinds)
		declarations = append(declarations, entry)
	}
	if len(declarations) != len(names) {
		panic("missing claimed rule")
	}
	sort.Slice(declarations, func(i, j int) bool { return declarations[i].Name < declarations[j].Name })
	if len(os.Args) > 1 && os.Args[1] == "--json" {
		if err := json.NewEncoder(os.Stdout).Encode(declarations); err != nil {
			panic(err)
		}
		return
	}
	for _, entry := range declarations {
		fmt.Printf("%s\t", entry.Name)
		for i, kind := range entry.Kinds {
			if i > 0 {
				fmt.Print(",")
			}
			fmt.Print(kind)
		}
		fmt.Println()
	}
}
