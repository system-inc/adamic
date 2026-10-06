// Extract declarations, not observations: help remains calculated by the typed flag parser.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
)

type definition struct{ name, kind, value, usage string }

func literal(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.BasicLit:
		text, err := strconv.Unquote(value.Value)
		if err != nil {
			panic(err)
		}
		return text
	case *ast.BinaryExpr:
		if value.Op != token.ADD {
			panic("non-add literal")
		}
		return literal(value.X) + literal(value.Y)
	case *ast.Ident:
		if value.Name == "projectMarker" {
			return "tsconfig.json"
		}
		if value.Name == "false" {
			return "false"
		}
	case *ast.SelectorExpr:
		if value.Sel.Name == "DefaultMaxPasses" {
			return "10"
		}
	}
	panic("unrecognized flag declaration")
}
func extract(source string) []definition {
	file, err := parser.ParseFile(token.NewFileSet(), source, nil, 0)
	if err != nil {
		panic(err)
	}
	var definitions []definition
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		qualifier, ok := selector.X.(*ast.Ident)
		if !ok || (qualifier.Name != "flag" && qualifier.Name != "flags") {
			return true
		}
		kind := selector.Sel.Name
		if kind != "String" && kind != "Bool" && kind != "Int" {
			return true
		}
		definitions = append(definitions, definition{literal(call.Args[0]), kind, literal(call.Args[1]), literal(call.Args[2])})
		return true
	})
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].name < definitions[j].name })
	return definitions
}
func main() {
	source := "cohere/command/cohere/main.go"
	if len(os.Args) > 2 {
		source = os.Args[1]
	}
	definitions := extract(source)
	if len(os.Args) > 2 {
		var input struct {
			Program   string   `json:"program"`
			Arguments []string `json:"arguments"`
		}
		if err := json.Unmarshal([]byte(os.Args[2]), &input); err != nil {
			panic(err)
		}
		set := flag.NewFlagSet(input.Program, flag.ContinueOnError)
		var stderr bytes.Buffer
		set.SetOutput(&stderr)
		for _, item := range definitions {
			switch item.kind {
			case "Bool":
				set.Bool(item.name, item.value == "true", item.usage)
			case "String":
				set.String(item.name, item.value, item.usage)
			case "Int":
				value, err := strconv.Atoi(item.value)
				if err != nil {
					panic(err)
				}
				set.Int(item.name, value, item.usage)
			}
		}
		err := set.Parse(input.Arguments)
		code := 0
		if err != nil && !errors.Is(err, flag.ErrHelp) {
			code = 2
		}
		fmt.Printf("%d\t%t\t%q\n", code, err != nil, stderr.String())
		given := map[string]bool{}
		set.Visit(func(value *flag.Flag) { given[value.Name] = true })
		set.VisitAll(func(value *flag.Flag) {
			fmt.Printf("%s\t%q\t%t\n", value.Name, value.Value.String(), given[value.Name])
		})
		for _, argument := range set.Args() {
			fmt.Printf("arg\t%q\n", argument)
		}
		return
	}
	out, err := os.Create("stage1/cohere/command/definitions.ts")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	fmt.Fprintln(out, "// Generated from pinned Go command flag declarations by testdata/definitions.go.")
	fmt.Fprintln(out, "export interface Definition { readonly name: string; readonly kind: string; readonly value: string; readonly usage: string; }")
	fmt.Fprintln(out, "export const definitions: Definition[] = [")
	for _, item := range definitions {
		fmt.Fprintf(out, "\t{ name: %s, kind: %s, value: %s, usage: %s },\n", strconv.Quote(item.name), strconv.Quote(item.kind), strconv.Quote(item.value), strconv.Quote(item.usage))
	}
	fmt.Fprintln(out, "];")
	fmt.Fprintln(out, "export const renameDefinitions: Definition[] = [")
	for _, item := range extract("cohere/command/cohere/rename.go") {
		fmt.Fprintf(out, "\t{ name: %s, kind: %s, value: %s, usage: %s },\n", strconv.Quote(item.name), strconv.Quote(item.kind), strconv.Quote(item.value), strconv.Quote(item.usage))
	}
	fmt.Fprintln(out, "];")
}
