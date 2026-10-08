// Executed through a Go overlay rooted in cohere. Production cohere is unchanged.
package main

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
)

type node struct {
	Kind          string `json:"kind"`
	Text          string `json:"text"`
	Expression    int    `json:"expression"`
	Name          int    `json:"name"`
	ImportClause  int    `json:"importClause"`
	NamedBindings int    `json:"namedBindings"`
	Elements      []int  `json:"elements"`
}
type query struct {
	Node   int    `json:"node"`
	Wanted string `json:"wanted"`
	Helper string `json:"helper"`
}
type sample struct {
	Nodes     []node   `json:"nodes"`
	Queries   []query  `json:"queries"`
	BaseNames []string `json:"baseNames"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output, symbol := os.Args[1], os.Args[2], os.Args[3]
	data, err := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/helpers/readiness.json"))
	must(err)
	var readiness struct {
		Remaining []struct {
			Rule    string   `json:"rule"`
			Helpers []string `json:"remaining_helpers"`
		} `json:"remaining"`
	}
	must(json.Unmarshal(data, &readiness))
	data, err = os.ReadFile(filepath.Join(root, "stage1/cohere/lint/inventory/inventory.json"))
	must(err)
	var inventory struct {
		Rules []struct {
			Name  string `json:"name"`
			Tests struct {
				Files []string `json:"files"`
			} `json:"tests"`
		} `json:"rules"`
	}
	must(json.Unmarshal(data, &inventory))
	consumers := map[string]bool{}
	for _, r := range readiness.Remaining {
		for _, h := range r.Helpers {
			if h == symbol {
				consumers[r.Rule] = true
			}
		}
	}
	if len(consumers) == 0 {
		panic("no consumers")
	}
	sources := map[string]bool{}
	counts := map[string]int{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		for _, path := range r.Tests.Files {
			file, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(root, path), nil, 0)
			must(err)
			goast.Inspect(file, func(n goast.Node) bool {
				lit, ok := n.(*goast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				source, err := strconv.Unquote(lit.Value)
				must(err)
				if len(source) > 0 {
					sources[source] = true
					counts[r.Name]++
				}
				return true
			})
		}
		if counts[r.Name] == 0 {
			panic("no source strings for " + r.Name)
		}
	}
	for _, source := range []string{"import Thing from 'm';", "import * as All from 'm';", "import { a, b as c, type T } from 'm';", "import Thing, * as All from 'm';", "import Thing, { a as b } from 'm';", "import 'm';", "import type { X } from 'm';", "import { } from 'm';", "import X = require('m');"} {
		sources[source] = true
	}
	for _, source := range []string{"Component; PureComponent; React.Component; React.PureComponent; Other.Component; React['Component']; (React).Component; ((React)).PureComponent; (Component); (React.Component); React.component; React?.Component; React.Component!; React.Component as unknown;", "class C extends React.Component {} class D extends (React).PureComponent {} class E extends (Component) {}"} {
		sources[source] = true
	}
	// Discriminating controls include deep parentheses and wrappers Go does not skip.
	for _, source := range []string{"React; document; ((React)); (((document))); (Other); 'React'; React!; React as unknown; (React as unknown); this; React.Component; React['Component']; R\\u0065act; réact;", "const x = <svg xlink:href='x' {...props} disabled data-id='x'/>; const y = <Foo.Bar></Foo.Bar>; const z = <ns:tag />;", "const x = <Box<string> name='x'></Box>;"} {
		sources[source] = true
	}
	ordered := []string{}
	for source := range sources {
		ordered = append(ordered, source)
	}
	sort.Strings(ordered)
	corpus := []sample{}
	var expected strings.Builder
	for _, source := range ordered {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute("/probe.tsx"), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute("/probe.tsx"))}, source, core.ScriptKindTSX)
		actual := []*ast.Node{}
		indexes := map[*ast.Node]int{}
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			if _, ok := indexes[n]; ok {
				return
			}
			indexes[n] = len(actual)
			actual = append(actual, n)
			n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
		index := func(n *ast.Node) int {
			if n == nil {
				return -1
			}
			i, ok := indexes[n]
			if !ok {
				panic("named field missing from child walk")
			}
			return i
		}
		s := sample{Nodes: []node{}, Queries: []query{}, BaseNames: []string{}}
		for _, n := range actual {
			p := node{Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), Expression: -1, Name: -1, ImportClause: -1, NamedBindings: -1, Elements: []int{}}
			if n.Kind == ast.KindIdentifier {
				p.Text = n.Text()
				if react.AdamicComponentBaseName(p.Text) {
					s.BaseNames = append(s.BaseNames, p.Text)
				}
			}
			if n.Kind == ast.KindParenthesizedExpression {
				p.Expression = index(n.AsParenthesizedExpression().Expression)
			}
			if n.Kind == ast.KindJsxAttribute {
				p.Name = index(n.AsJsxAttribute().Name())
			}
			if n.Kind == ast.KindImportDeclaration {
				p.ImportClause = index(n.AsImportDeclaration().ImportClause)
			}
			if n.Kind == ast.KindImportClause {
				p.Name = index(n.Name())
				p.NamedBindings = index(n.AsImportClause().NamedBindings)
			}
			if n.Kind == ast.KindNamespaceImport || n.Kind == ast.KindImportSpecifier {
				p.Name = index(n.Name())
			}
			if n.Kind == ast.KindNamedImports && n.AsNamedImports().Elements != nil {
				for _, child := range n.AsNamedImports().Elements.Nodes {
					p.Elements = append(p.Elements, index(child))
				}
			}
			if n.Kind == ast.KindPropertyAccessExpression {
				p.Expression = index(n.AsPropertyAccessExpression().Expression)
				p.Name = index(n.Name())
			}
			s.Nodes = append(s.Nodes, p)
		}
		for i, n := range actual {
			if strings.HasSuffix(symbol, ".isComponentBase") {
				s.Queries = append(s.Queries, query{i, "", "component"})
				fmt.Fprintln(&expected, react.AdamicComponentBase(n))
				continue
			}
			if strings.HasSuffix(symbol, ".BindingsOf") {
				result := imports.BindingsOf(n)
				named := []string{}
				for _, item := range result.Named {
					named = append(named, strconv.Itoa(index(item)))
				}
				s.Queries = append(s.Queries, query{i, "", "imports"})
				fmt.Fprintf(&expected, "%d:%d:%s\n", index(result.Default), index(result.Namespace), strings.Join(named, ","))
				continue
			}
			names := []string{"React", "document", "Component", "", "réact"}
			if n.Kind == ast.KindIdentifier {
				names = append(names, n.Text())
			}
			for _, name := range names {
				s.Queries = append(s.Queries, query{i, name, "identifier"})
				fmt.Fprintln(&expected, react.AdamicIdentifierNamed(n, name))
			}
		}
		if strings.HasSuffix(symbol, ".BindingsOf") {
			s.Queries = append(s.Queries, query{-1, "", "imports"})
			result := imports.BindingsOf(nil)
			fmt.Fprintf(&expected, "%d:%d:\n", index(result.Default), index(result.Namespace))
		}
		if strings.HasSuffix(symbol, ".isComponentBase") {
			s.Queries = append(s.Queries, query{-1, "", "component"})
			fmt.Fprintln(&expected, react.AdamicComponentBase(nil))
		}
		corpus = append(corpus, s)
	}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct test-file strings and controls, %d Go verdicts\n", len(consumers), len(corpus), strings.Count(expected.String(), "\n"))
}
