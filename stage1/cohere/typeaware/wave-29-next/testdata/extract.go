// Parse unchanged Go fixtures; never import Adamic or reinterpret its findings.
package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

type Row struct {
	Rule        string          `json:"rule"`
	Source      string          `json:"source"`
	Options     json.RawMessage `json:"options"`
	Environment string          `json:"environment"`
	Function    string          `json:"function"`
}

func text(e ast.Expr) string {
	if b, ok := e.(*ast.BasicLit); ok && b.Kind == token.STRING {
		s, err := strconv.Unquote(b.Value)
		if err == nil {
			return s
		}
	}
	return ""
}
func main() {
	var rows []Row
	for _, name := range []string{"no_restricted_globals", "no_setter_return", "no_shadow_restricted_names"} {
		f, err := parser.ParseFile(token.NewFileSet(), os.Args[1]+"/cohere/internal/lint/rules/core/"+name+"_test.go", nil, 0)
		if err != nil {
			panic(err)
		}
		visit := func(root ast.Node, function string) {
			ast.Inspect(root, func(n ast.Node) bool {
				l, ok := n.(*ast.CompositeLit)
				if !ok {
					return true
				}
				array, ok := l.Type.(*ast.ArrayType)
				if !ok {
					return true
				}
				fields := []string{}
				if shape, ok := array.Elt.(*ast.StructType); ok {
					for _, field := range shape.Fields.List {
						for _, n := range field.Names {
							fields = append(fields, n.Name)
						}
					}
				}
				if ident, ok := array.Elt.(*ast.Ident); ok && ident.Name == "string" && name == "no_setter_return" && function == "TestNoSetterReturnStaysSilent" {
					for _, e := range l.Elts {
						if s := text(e); s != "" {
							rows = append(rows, Row{Rule: "no-setter-return", Source: s, Function: function})
						}
					}
					return false
				}
				for _, e := range l.Elts {
					rowlit, ok := e.(*ast.CompositeLit)
					if !ok {
						continue
					}
					row := Row{Rule: strings.ReplaceAll(name, "_", "-"), Function: function}
					for i, value := range rowlit.Elts {
						key := ""
						if kv, ok := value.(*ast.KeyValueExpr); ok {
							if ident, ok := kv.Key.(*ast.Ident); ok {
								key = ident.Name
							}
							value = kv.Value
						} else if i < len(fields) {
							key = fields[i]
						}
						switch key {
						case "source", "sourceText":
							row.Source = text(value)
						case "optionsJson":
							row.Options = json.RawMessage(text(value))
						case "environmentGlobals":
							if call, ok := value.(*ast.CallExpr); ok && len(call.Args) > 0 {
								row.Environment = text(call.Args[0])
							}
						}
					}
					if row.Source == "" {
						continue
					}
					if name == "no_restricted_globals" && len(row.Options) == 0 {
						continue
					}
					if function == "TestNoShadowRestrictedNamesRespectsAllowGlobalThis" {
						row.Options = json.RawMessage(`{"reportGlobalThis":false}`)
					}
					rows = append(rows, row)
				}
				return false
			})
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok {
				visit(fn, fn.Name.Name)
			} else {
				visit(d, "global")
			}
		}
	}
	if len(rows) < 200 {
		panic("fixture extraction unexpectedly small")
	}
	if err := json.NewEncoder(os.Stdout).Encode(rows); err != nil {
		panic(err)
	}
}
