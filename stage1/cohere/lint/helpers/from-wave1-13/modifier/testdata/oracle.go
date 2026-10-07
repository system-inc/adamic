package main

import (
	"encoding/json"
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"
)

type row = collapse.AdamicModifierCase

func main() {
	if os.Args[1] == "--capture" {
		names := []string{"enforce_canonical_classes", "enforce_consistent_class_order", "enforce_consistent_variant_order", "enforce_shorthand_classes", "no_conflicting_classes", "no_unknown_classes"}
		rows := []row{}
		counts := map[string]int{}
		for _, name := range names {
			files, err := filepath.Glob(filepath.Join(os.Args[2], "internal/lint/rules/tailwind", name+"*test.go"))
			if err != nil || len(files) == 0 {
				panic("consumer test missing: " + name)
			}
			for _, file := range files {
				tree, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
				if err != nil {
					panic(err)
				}
				ast.Inspect(tree, func(node ast.Node) bool {
					literal, ok := node.(*ast.BasicLit)
					if !ok || literal.Kind != token.STRING {
						return true
					}
					value, err := strconv.Unquote(literal.Value)
					if err != nil {
						panic(err)
					}
					rows = append(rows, collapse.AdamicModifierInput(name+":"+filepath.Base(file), value))
					counts[name]++
					return true
				})
			}
			if counts[name] == 0 {
				panic("no consumer literals: " + name)
			}
		}
		for _, control := range []string{"", "50", "none", "[0.5]", "[]", "[_]", "[\u00a0]", "[\ufeff]", "[a;b]", "[x)]", "[calc(100%-1rem)]", "(--a_b)", "(-color)", "()", "(--)", "(--a;bad)", "(----)", "[]x", "\u0000𝄞é"} {
			rows = append(rows, collapse.AdamicModifierInput("control", control))
		}
		data, err := json.Marshal(rows)
		if err != nil {
			panic(err)
		}
		if err = os.WriteFile(os.Args[3], data, 0644); err != nil {
			panic(err)
		}
		fmt.Printf("captured %d Go source literals from all %d consumer rule test families; %v\n", len(rows), len(names), counts)
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows []row
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for _, r := range rows {
		n := collapse.AdamicParseModifier(r.Source)
		if n == nil {
			fmt.Println("false||")
		} else {
			fmt.Printf("true|%s|%s\n", n.Kind, units(n.Value))
		}
	}
}

func units(text string) string {
	values := utf16.Encode([]rune(text))
	parts := []string{}
	for _, v := range values {
		parts = append(parts, strconv.Itoa(int(v)))
	}
	return strings.Join(parts, ",")
}
