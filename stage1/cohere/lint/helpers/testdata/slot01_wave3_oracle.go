// Oracle-only capture for slot 01's third helper batch.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"strconv"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	var paths []string
	data, err := os.ReadFile(os.Args[1])
	must(err)
	must(json.Unmarshal(data, &paths))
	cases, err := os.Create(os.Args[2])
	must(err)
	defer cases.Close()
	expected, err := os.Create(os.Args[3])
	must(err)
	defer expected.Close()
	encode := json.NewEncoder(cases)
	sources := []string{"", `const a = <div className="flex" class="block" title="hidden" CLASSNAME="grid" data-class="gap-2" className />;`, `const a=<div className={(('p-2'))} />;`, `const a=<div className={flag ? 'flex' : 'grid'} />;`, `const a=<div className={['a', 'b']} />;`, "const a=<div className={`flex${flag ? ' hidden' : ''}`} />;"}
	fixtureCount := 0
	for _, path := range paths {
		file, err := goparser.ParseFile(token.NewFileSet(), path, nil, 0)
		must(err)
		constants := map[string]goast.Expr{}
		goast.Inspect(file, func(node goast.Node) bool {
			if declaration, ok := node.(*goast.ValueSpec); ok && len(declaration.Names) == len(declaration.Values) {
				for i, name := range declaration.Names {
					constants[name.Name] = declaration.Values[i]
				}
			}
			return true
		})
		var value func(goast.Expr, int) (string, bool)
		value = func(expression goast.Expr, depth int) (string, bool) {
			if depth > 32 {
				return "", false
			}
			switch e := expression.(type) {
			case *goast.BasicLit:
				if e.Kind == token.STRING {
					s, err := strconv.Unquote(e.Value)
					return s, err == nil
				}
			case *goast.BinaryExpr:
				if e.Op == token.ADD {
					a, ok := value(e.X, depth+1)
					b, ok2 := value(e.Y, depth+1)
					return a + b, ok && ok2
				}
			case *goast.Ident:
				if e, ok := constants[e.Name]; ok {
					return value(e, depth+1)
				}
			case *goast.ParenExpr:
				return value(e.X, depth+1)
			}
			return "", false
		}
		found := 0
		seen := map[string]bool{}
		goast.Inspect(file, func(node goast.Node) bool {
			expression, ok := node.(goast.Expr)
			if !ok {
				return true
			}
			source, ok := value(expression, 0)
			if ok {
				if source != "" && !seen[source] {
					sources = append(sources, source)
					seen[source] = true
					found++
				}
				return false // Preserve complete concatenations rather than parsing their fragments.
			}
			return true
		})
		if found == 0 {
			panic("no string inputs in " + path)
		}
		fmt.Fprintf(os.Stderr, "%s: %d string inputs\n", path, found)
		fixtureCount += found
	}

	queries := 0
	for _, source := range sources {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/fixture.tsx", Path: tspath.Path("/fixture.tsx")}, source, core.ScriptKindTSX)
		nodes := []*ast.Node{}
		ids := map[*ast.Node]int{nil: -1}
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			if _, ok := ids[node]; ok {
				return
			}
			ids[node] = len(nodes)
			nodes = append(nodes, node)
			node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
		for _, node := range nodes {
			if node.Kind != ast.KindJsxAttribute {
				continue
			}
			attribute := node.AsJsxAttribute()
			name := ""
			present := attribute.Name() != nil
			if present {
				name = attribute.Name().Text()
			}
			for _, settings := range []map[string]bool{{"class": true, "className": true}, {"className": false, "data-class": true, "CLASSNAME": true}, {}} {
				pairs := [][]any{}
				// Stable order, including explicitly false membership.
				for _, key := range []string{"class", "className", "data-class", "CLASSNAME"} {
					if value, ok := settings[key]; ok {
						pairs = append(pairs, []any{key, value})
					}
				}
				init := attribute.Initializer
				delegated := tailwind.AdamicWave3Under(init, "Attribute", ids)
				must(encode.Encode([]any{present, name, ids[init], pairs, delegated}))
				tailwind.AdamicWave3Print(expected, tailwind.AdamicWave3Attribute(node, settings, ids))
				queries++
			}
		}
	}
	fmt.Fprintf(os.Stderr, "%d fixture strings; %d source batches; %d helper queries\n", fixtureCount, len(sources), queries)
}
