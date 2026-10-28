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
	types     map[string]ast.Expr
	needed    map[string]bool
	sortedMap *ast.TypeSpec
	positions *token.FileSet
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
		if printed(t.X) != "ast" && printed(t.X) != "checker" {
			panic("state copy: foreign state type " + printed(t) + " requires an explicit copier")
		}
		return value
	case *ast.StarExpr:
		// The opaque mapper belongs to the checker; copying its empty facade loses identity.
		if printed(t.X) == "typeMapper" {
			return value
		}
		if name, ok := t.X.(*ast.Ident); ok {
			if _, ok := g.types[name.Name].(*ast.StructType); ok {
				g.needed[name.Name] = true
				return "latentCopyPointer_" + name.Name + "(" + value + ", seen)"
			}
		}
		if name, ok := t.X.(*ast.SelectorExpr); ok && printed(name.X) == "ir" {
			return "latentCopyIR(" + value + ", seen)"
		}
		if name, ok := t.X.(*ast.SelectorExpr); ok && (printed(name.X) == "ast" || printed(name.X) == "checker" || printed(name) == "load.Program") {
			return value
		}
		panic("state copy: foreign or unrecognized pointer " + printed(t) + " requires an explicit copier")
	case *ast.ArrayType:
		if t.Len != nil {
			panic("state copy: fixed array " + printed(t) + " requires an explicit copier")
		}
		ty := printed(t)
		return "func(value " + ty + ") " + ty + " {if value==nil{return nil}; result:=make(" + ty + ",len(value)); for i,item:=range value {result[i]=" + g.copy(t.Elt, "item") + "}; return result}(" + value + ")"
	case *ast.MapType:
		ty := printed(t)
		return g.copyMap(ty, t.Key, t.Value, value)
	case *ast.IndexListExpr:
		selector, ok := t.X.(*ast.SelectorExpr)
		if ok && printed(selector.X) == "ir" && selector.Sel.Name == "SortedMap" && len(t.Indices) == 2 && g.sortedMap != nil {
			definition, ok := g.sortedMap.Type.(*ast.MapType)
			if !ok || printed(definition.Key) != "K" || printed(definition.Value) != "V" {
				panic("state copy: ir.SortedMap definition is no longer map[K]V")
			}
			return g.copyMap(printed(t), t.Indices[0], t.Indices[1], value)
		}
		panic("state copy: unsupported field type " + printed(t))
	case *ast.InterfaceType:
		panic("state copy: anonymous interface requires an explicit copier")
	default:
		panic("state copy: unsupported field type " + printed(t))
	}
}
func (g *generator) copyField(owner string, id *ast.Ident, field *ast.Field) (result string) {
	defer func() {
		if failure := recover(); failure != nil {
			panic(fmt.Sprintf("%s: %s.%s: %v", g.positions.Position(field.Pos()), owner, id.Name, failure))
		}
	}()
	return g.copy(field.Type, "value."+id.Name)
}
func (g *generator) copyMap(ty string, key, item ast.Expr, value string) string {
	return "func(value " + ty + ") " + ty + " {if value==nil{return nil}; result:=make(" + ty + ",len(value)); for key,item:=range value {result[" + g.copy(key, "key") + "]=" + g.copy(item, "item") + "}; return result}(" + value + ")"
}
func loadGenerator(repository, lowerInput string) generator {
	g := generator{types: map[string]ast.Expr{}, needed: map[string]bool{"lowering": true}, positions: token.NewFileSet()}

	irFiles, err := filepath.Glob(filepath.Join(repository, "internal/ir/*.go"))
	if err != nil {
		panic(err)
	}
	for _, path := range irFiles {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			panic(err)
		}
		for _, decl := range file.Decls {
			if group, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range group.Specs {
					if spec, ok := spec.(*ast.TypeSpec); ok && spec.Name.Name == "SortedMap" {
						g.sortedMap = spec
					}
				}
			}
		}
	}
	paths, err := filepath.Glob(filepath.Join(repository, "internal/lower/*.go"))
	if err != nil {
		panic(err)
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		input := path
		if filepath.Base(path) == "lower.go" {
			input = lowerInput
		}
		file, err := parser.ParseFile(g.positions, input, nil, 0)
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
	return g
}
func generate(repository, lowerInput string) []byte {
	g := loadGenerator(repository, lowerInput)
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
					var expr string
					if name == "lowering" && id.Name == "impliedTargetQuery" && printed(field.Type) == "func(*ast.Node) *checker.Type" {
						// Production leaves this test query nil. A closure needs an audited copier.
						fmt.Fprint(&body, "if value.impliedTargetQuery != nil {panic(\"latent state copy: populated impliedTargetQuery is unsupported\")};")
						expr = "value." + id.Name
					} else if name == "lowering" && id.Name == "sourceIdentities" && printed(field.Type) == "map[*ast.Node]ir.SourceIdentity" {
						// Shared on purpose: a cache of identities that are functions of the program alone, one map for
						// every lowering of a latent run (lower.go.txt's latentRun), so a copy reads and fills the same one.
						expr = "value." + id.Name
					} else {
						expr = g.copyField(name, id, field)
					}
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
	return text
}
func main() {
	if len(os.Args) != 4 {
		panic("usage: statecopy repository overlaid-lower.go output.go")
	}
	text := generate(os.Args[1], os.Args[2])
	if err := os.WriteFile(os.Args[3], text, 0600); err != nil {
		panic(err)
	}
}
