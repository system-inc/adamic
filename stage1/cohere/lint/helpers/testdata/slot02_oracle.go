package main

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	tsparser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
)

type Node struct {
	Kind, Text                            string
	Expression, TagName, Attributes, Name int
}
type Query struct {
	Node   int
	Wanted string
}
type Corpus struct {
	Nodes          []Node
	Queries        []Query
	Want           string
	Files, Sources int
}

func main() {
	var paths []string
	input, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer input.Close()
	if err := json.NewDecoder(input).Decode(&paths); err != nil {
		panic(err)
	}
	sources := []string{"React", "(React)", "(((React)))", "document", "(document)", "react", "React.useState", "React['useState']", "React as unknown", "React!", "'React'", "null", "", "<div children />", "<Thing></Thing>", "<svg xlink:href='x' />", "const \\u0052eact = 1;", "const 世界 = 1;"}
	for _, path := range paths {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			panic(err)
		}
		goast.Inspect(file, func(n goast.Node) bool {
			if lit, ok := n.(*goast.BasicLit); ok && lit.Kind == token.STRING {
				value, err := strconv.Unquote(lit.Value)
				if err != nil {
					panic(err)
				}
				sources = append(sources, value)
			}
			return true
		})
	}
	corpus := Corpus{Files: len(paths)}
	pointers := []*ast.Node{}
	ids := map[*ast.Node]int{}
	var add func(*ast.Node) int
	add = func(n *ast.Node) int {
		if n == nil {
			return -1
		}
		if id, ok := ids[n]; ok {
			return id
		}
		id := len(corpus.Nodes)
		ids[n] = id
		pointers = append(pointers, n)
		corpus.Nodes = append(corpus.Nodes, Node{Kind: "Other", Expression: -1, TagName: -1, Attributes: -1, Name: -1})
		v := Node{Kind: "Other", Expression: -1, TagName: -1, Attributes: -1, Name: -1}
		switch n.Kind {
		case ast.KindIdentifier:
			v.Kind = "Identifier"
			v.Text = n.Text()
		case ast.KindJsxAttribute:
			v.Kind = "JsxAttribute"
			v.Name = add(n.AsJsxAttribute().Name())
		case ast.KindParenthesizedExpression:
			v.Kind = "ParenthesizedExpression"
			v.Expression = add(n.AsParenthesizedExpression().Expression)
		}
		corpus.Nodes[id] = v
		return id
	}
	seen := map[string]bool{}
	for _, source := range sources {
		if seen[source] {
			continue
		}
		seen[source] = true
		corpus.Sources++
		file := tsparser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/source.tsx", Path: tspath.Path("/source.tsx")}, source, core.ScriptKindTSX)
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			add(n)
			n.ForEachChild(walk)
			return false
		}
		walk(file.AsNode())
	}
	var want strings.Builder
	for id := -1; id < len(pointers); id++ {
		var n *ast.Node
		if id >= 0 {
			n = pointers[id]
		}
		corpus.Queries = append(corpus.Queries, Query{Node: id})
		name, named := jsx.AttributeName(n)
		fmt.Fprintf(&want, "%t:%s\n", named, name)
	}
	corpus.Want = want.String()
	if err := json.NewEncoder(os.Stdout).Encode(corpus); err != nil {
		panic(err)
	}
}
