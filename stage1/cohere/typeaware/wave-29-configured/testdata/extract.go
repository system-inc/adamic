// Extract source and option inputs from unchanged production test tables.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
)

type Fixture struct {
	Rule    string          `json:"rule"`
	Source  string          `json:"source"`
	Names   []string        `json:"names"`
	Pattern string          `json:"pattern"`
	Flags   map[string]bool `json:"flags"`
}

func text(e ast.Expr) (string, bool) {
	l, ok := e.(*ast.BasicLit)
	if !ok || l.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(l.Value)
	return s, err == nil
}
func main() {
	var rows []Fixture
	for _, name := range []string{"id_denylist", "id_match"} {
		f, err := parser.ParseFile(token.NewFileSet(), os.Args[1]+"/cohere/internal/lint/rules/core/"+name+"_test.go", nil, 0)
		if err != nil {
			panic(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			l, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			var source string
			var settings ast.Expr
			for _, e := range l.Elts {
				kv, ok := e.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				k, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}
				if k.Name == "source" {
					source, _ = text(kv.Value)
				}
				if k.Name == "settings" {
					settings = kv.Value
				}
			}
			call, ok := settings.(*ast.CallExpr)
			if !ok || source == "" {
				return true
			}
			fn, ok := call.Fun.(*ast.Ident)
			if !ok {
				return true
			}
			row := Fixture{Source: source, Flags: map[string]bool{}}
			switch fn.Name {
			case "idDenylistOf":
				row.Rule = "id-denylist"
				for _, e := range call.Args {
					s, ok := text(e)
					if !ok {
						return true
					}
					row.Names = append(row.Names, s)
				}
			case "idMatchOf":
				row.Rule = "id-match"
				if len(call.Args) != 2 {
					return true
				}
				row.Pattern, _ = text(call.Args[0])

				if flags, ok := call.Args[1].(*ast.CompositeLit); ok {
					for _, e := range flags.Elts {
						kv := e.(*ast.KeyValueExpr)
						k, _ := text(kv.Key)
						v := kv.Value.(*ast.Ident).Name
						row.Flags[k] = v == "true"
					}
				}
			default:
				return true
			}
			rows = append(rows, row)
			return true
		})
	}
	if len(rows) < 180 {
		panic(fmt.Sprintf("too few extracted cases: %d", len(rows)))
	}
	if err := json.NewEncoder(os.Stdout).Encode(rows); err != nil {
		panic(err)
	}
}
