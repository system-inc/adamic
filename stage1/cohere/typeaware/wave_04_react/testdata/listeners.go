// Metadata oracle parses production Go listeners and resolves real parser enums.
package main

import (
	"encoding/json"
	"fmt"
	tsast "github.com/microsoft/TypeScript/tsc/shim/ast"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	kinds := map[string]tsast.Kind{}
	for kind := tsast.Kind(0); kind < tsast.KindCount; kind++ {
		kinds[kind.String()] = kind
	}

	var subjects [][2]string
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, &subjects); err != nil {
		panic(err)
	}
	type row struct {
		File, Name string
		Kinds      []int
		KindNames  []string
	}
	rows := []row{}
	for _, subject := range subjects {
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(os.Args[2], subject[1]), nil, 0)
		if err != nil {
			panic(err)
		}
		r := row{File: subject[0]}
		tables := 0
		goast.Inspect(file, func(node goast.Node) bool {
			composite, ok := node.(*goast.CompositeLit)
			if !ok {
				return true
			}
			selector, ok := composite.Type.(*goast.SelectorExpr)
			if !ok {
				return true
			}
			base, ok := selector.X.(*goast.Ident)
			if !ok || base.Name != "rule" {
				return true
			}
			if selector.Sel.Name == "Rule" {
				for _, element := range composite.Elts {
					kv, ok := element.(*goast.KeyValueExpr)
					if !ok {
						continue
					}
					key, ok := kv.Key.(*goast.Ident)
					if ok && key.Name == "Name" {
						literal, ok := kv.Value.(*goast.BasicLit)
						if !ok {
							panic("computed rule name")
						}
						r.Name, err = strconv.Unquote(literal.Value)
						if err != nil {
							panic(err)
						}
					}
				}
			}
			if selector.Sel.Name == "Listeners" {
				tables++
				for _, element := range composite.Elts {
					kv, ok := element.(*goast.KeyValueExpr)
					if !ok {
						panic("listener shape")
					}
					key, ok := kv.Key.(*goast.SelectorExpr)
					if !ok {
						panic("computed listener")
					}
					kind, ok := kinds[key.Sel.Name]
					if !ok {
						panic(key.Sel.Name)
					}
					r.Kinds = append(r.Kinds, int(kind))
				}
			}
			return true
		})
		if tables != 1 || r.Name == "" || len(r.Kinds) == 0 {
			panic("missing listeners " + subject[1])
		}
		sort.Ints(r.Kinds)
		for _, kind := range r.Kinds {
			r.KindNames = append(r.KindNames, strings.TrimPrefix(tsast.Kind(kind).String(), "Kind"))
		}
		rows = append(rows, r)
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(os.Args[3], encoded, 0600); err != nil {
		panic(err)
	}
	for _, r := range rows {
		values := []string{}
		for _, kind := range r.Kinds {
			values = append(values, strings.TrimPrefix(tsast.Kind(kind).String(), "Kind"))
		}
		fmt.Printf("%s\t%s\n", r.Name, strings.Join(values, ","))
	}
}
