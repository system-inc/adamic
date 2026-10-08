// statecopy generates typed copies of lowering-owned mutable state for the overlay.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type generator struct {
	types  map[string]ast.Expr
	needed map[string]bool
}

func printed(n ast.Node) string {
	var b bytes.Buffer
	if err := format.Node(&b, token.NewFileSet(), n); err != nil {
		panic(err)
	}
	return b.String()
}
func (g *generator) copy(t ast.Expr, value string) string {
	switch t := t.(type) {
	case *ast.Ident:
		if _, ok := g.types[t.Name].(*ast.StructType); ok {
			g.needed[t.Name] = true
			return "latentCopy_" + t.Name + "(" + value + ", seen)"
		}
		if underlying, ok := g.types[t.Name]; ok {
			switch underlying.(type) {
			case *ast.ArrayType, *ast.MapType:
				panic("state copy: mutable named container " + t.Name + " requires an explicit copier")
			}
		}
		return value
	case *ast.SelectorExpr:
		if printed(t.X) == "ir" {
			return "latentCopyIR(" + value + ", seen)"
		}
		return value
	case *ast.StarExpr:
		if name, ok := t.X.(*ast.Ident); ok {
			if _, ok := g.types[name.Name].(*ast.StructType); ok {
				g.needed[name.Name] = true
				return "latentCopyPointer_" + name.Name + "(" + value + ", seen)"
			}
		}
		if name, ok := t.X.(*ast.SelectorExpr); ok && printed(name.X) == "ir" {
			return "latentCopyIR(" + value + ", seen)"
		}
		return value // Checker, source AST, and loaded-program identity stay shared.
	case *ast.ArrayType:
		if t.Len != nil {
			panic("state copy: fixed array " + printed(t) + " requires an explicit copier")
		}
		ty := printed(t)
		return "func(value " + ty + ") " + ty + " {if value==nil{return nil}; result:=make(" + ty + ",len(value)); for i,item:=range value {result[i]=" + g.copy(t.Elt, "item") + "}; return result}(" + value + ")"
	case *ast.MapType:
		ty := printed(t)
		return "func(value " + ty + ") " + ty + " {if value==nil{return nil}; result:=make(" + ty + ",len(value)); for key,item:=range value {result[key]=" + g.copy(t.Value, "item") + "}; return result}(" + value + ")"
	case *ast.InterfaceType:
		return value
	default:
		panic("state copy: unsupported field type " + printed(t))
	}
}
func main() {
	if len(os.Args) != 4 {
		panic("usage: statecopy repository overlaid-lower.go output.go")
	}
	g := generator{types: map[string]ast.Expr{}, needed: map[string]bool{"lowering": true}}
	paths, err := filepath.Glob(filepath.Join(os.Args[1], "internal/lower/*.go"))
	if err != nil {
		panic(err)
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		input := path
		if filepath.Base(path) == "lower.go" {
			input = os.Args[2]
		}
		file, err := parser.ParseFile(token.NewFileSet(), input, nil, 0)
		if err != nil {
			panic(err)
		}
		for _, decl := range file.Decls {
			if group, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range group.Specs {
					if spec, ok := spec.(*ast.TypeSpec); ok {
						g.types[spec.Name.Name] = spec.Type
					}
				}
			}
		}
	}
	if _, ok := g.types["lowering"].(*ast.StructType); !ok {
		panic("state copy: lowering: expected a struct")
	}
	var body strings.Builder
	done := map[string]bool{}
	for {
		names := []string{}
		for name := range g.needed {
			if !done[name] {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		if len(names) == 0 {
			break
		}
		for _, name := range names {
			done[name] = true
			fields := g.types[name].(*ast.StructType).Fields.List
			fmt.Fprintf(&body, "func latentCopy_%s(value %s, seen map[any]any) %s {result:=value;", name, name, name)
			for _, field := range fields {
				if len(field.Names) == 0 {
					panic("state copy: " + name + ": embedded field unsupported")
				}
				for _, id := range field.Names {
					expr := g.copy(field.Type, "value."+id.Name)
					if expr != "value."+id.Name {
						fmt.Fprintf(&body, "result.%s=%s;", id.Name, expr)
					}
				}
			}
			body.WriteString("return result}\n")
			fmt.Fprintf(&body, "func latentCopyPointer_%s(value *%s, seen map[any]any) *%s {if value==nil{return nil}; if prior,ok:=seen[value];ok{return prior.(*%s)};result:=new(%s);seen[value]=result; *result=latentCopy_%s(*value,seen);return result}\n", name, name, name, name, name, name)
		}
	}
	// Only imports actually mentioned in generated field types are needed.
	var header strings.Builder
	header.WriteString("package lower\nimport(\n")
	for _, pair := range [][2]string{{"ast", "github.com/microsoft/TypeScript/tsc/shim/ast"}, {"checker", "github.com/microsoft/TypeScript/tsc/shim/checker"}, {"ir", "github.com/system-inc/adamic/internal/ir"}} {
		if strings.Contains(body.String(), pair[0]+".") {
			fmt.Fprintf(&header, "%q\n", pair[1])
		}
	}
	header.WriteString(")\n")
	text, err := format.Source([]byte(header.String() + body.String()))
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[3], text, 0600); err != nil {
		panic(err)
	}
}
