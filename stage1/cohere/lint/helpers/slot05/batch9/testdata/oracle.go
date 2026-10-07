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
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
)

type node struct {
	Kind              string `json:"kind"`
	Text              string `json:"text"`
	Expression        int    `json:"expression"`
	Name              int    `json:"name"`
	PropertyName      int    `json:"propertyName"`
	KeywordToken      string `json:"keywordToken"`
	ArgumentsPresent  bool   `json:"argumentsPresent"`
	Arguments         []int  `json:"arguments"`
	PropertiesPresent bool   `json:"propertiesPresent"`
	Properties        []int  `json:"properties"`
}
type named struct {
	Text  string `json:"text"`
	Named bool   `json:"named"`
}
type fold struct {
	Candidate string `json:"candidate"`
	Wanted    string `json:"wanted"`
	Answer    bool   `json:"answer"`
}
type query struct {
	Node    int    `json:"node"`
	Wanted  string `json:"wanted"`
	Mode    string `json:"mode"`
	Matcher string `json:"matcher"`
}
type sample struct {
	Nodes   []node  `json:"nodes"`
	Names   []named `json:"names"`
	Folds   []fold  `json:"folds"`
	Queries []query `json:"queries"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output, symbol := os.Args[1], os.Args[2], os.Args[3]
	symbol = map[string]string{"source": "github.com/system-inc/cohere/internal/lint/ecmascript/imports.CallExpressionSource", "imported": "github.com/system-inc/cohere/internal/lint/ecmascript/imports.ImportedNameOf", "attribute": "github.com/system-inc/cohere/internal/lint/ecmascript/jsx.HasAttributeNamed"}[symbol]
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

	controls := []string{
		"import('m'); import('m', {with:{type:'json'}}); import(''); import(`m`); import(`m${x}`); import(x); import(); import.defer('m'); import.defer(''); import.defer(x); import.defer(`m`); import.source('m'); import.meta('m');",
		"require('m'); require(''); require(`m`); require(`m${x}`); require(); require('m','n'); require(x); require?.('m'); require<string>('m'); (require)('m'); obj.require('m'); Require('m');",
		"import { forwardRef, Component as Local, 'a-b' as AB, '' as Empty, type T, '\\u0061' as escaped } from 'react'; import X from 'm'; import * as All from 'm';",
		"const x = <script src='s' async defer={false} ASYNC='' {...p} xlink:href='x' />; const y = <style jsx global/>; const z = <a href='/about'/>; const n = <div />;",
		"const x = <div a b a href='one' href='two' K='v' K='v' Σ='v' σ='v' ς='v' ſ='v' S='v' />;",
		"const x = <div {...p} xlink:href='x' async={false} />; const y=<div a='x' b='y'/>;",
		"require('line\\nend'); import('nul\\u0000end'); import('😀'); import { 'line\\nend' as X } from 'm';",
	}
	for _, source := range controls {
		sources[source] = true
	}
	ordered := []string{}
	for source := range sources {
		ordered = append(ordered, source)
	}
	sort.Strings(ordered)
	corpus := []sample{}
	var expected strings.Builder
	verdicts := 0
	observe := func(roots []*ast.Node) {
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
		for _, root := range roots {
			visit(root)
		}
		index := func(n *ast.Node) int {
			if n == nil {
				return -1
			}
			i, ok := indexes[n]
			if !ok {
				panic("missing named field")
			}
			return i
		}
		s := sample{Nodes: []node{}, Names: []named{}, Folds: []fold{}, Queries: []query{}}
		candidates := map[string]bool{"": true}
		wanted := []string{"", "async", "ASYNC", "href", "jsx", "global", "defer", "K", "k", "Σ", "σ", "ς", "S", "s"}
		for _, n := range actual {
			p := node{Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), Expression: -1, Name: -1, PropertyName: -1, Arguments: []int{}, Properties: []int{}}
			if n.Kind == ast.KindIdentifier || ast.IsStringLiteralLike(n) || n.Kind == ast.KindMetaProperty {
				p.Text = n.Text()
			}
			switch n.Kind {
			case ast.KindCallExpression:
				c := n.AsCallExpression()
				p.Expression = index(c.Expression)
				p.ArgumentsPresent = c.Arguments != nil
				if c.Arguments != nil {
					for _, a := range c.Arguments.Nodes {
						p.Arguments = append(p.Arguments, index(a))
					}
				}
			case ast.KindMetaProperty:
				p.KeywordToken = strings.TrimPrefix(n.AsMetaProperty().KeywordToken.String(), "Kind")
			case ast.KindImportSpecifier:
				p.Name = index(n.Name())
				p.PropertyName = index(n.PropertyName())
			case ast.KindJsxAttributes:
				a := n.AsJsxAttributes()
				p.PropertiesPresent = a.Properties != nil
				if a.Properties != nil {
					for _, a := range a.Properties.Nodes {
						p.Properties = append(p.Properties, index(a))
					}
				}
			}
			s.Nodes = append(s.Nodes, p)
			name, ok := jsx.AttributeName(n)
			s.Names = append(s.Names, named{name, ok})
			if ok {
				candidates[name] = true
			}
		}
		extra := []string{}
		for c := range candidates {
			extra = append(extra, c)
		}
		sort.Strings(extra)
		wanted = append(wanted, extra...)
		// EqualFold remains a separately owned callback; these answers come from Go.
		for _, c := range extra {
			for _, w := range wanted {
				s.Folds = append(s.Folds, fold{c, w, jsx.MatchIgnoringCase(c, w)})
			}
		}
		answer := func(i int, n *ast.Node) {
			if strings.HasSuffix(symbol, ".CallExpressionSource") {
				s.Queries = append(s.Queries, query{Node: i, Mode: "source"})
				source, found := imports.CallExpressionSource(n)
				fmt.Fprintln(&expected, found)
				fmt.Fprintln(&expected, source)
				verdicts++
				return
			}
			if strings.HasSuffix(symbol, ".ImportedNameOf") {
				s.Queries = append(s.Queries, query{Node: i, Mode: "imported"})
				fmt.Fprintln(&expected, imports.ImportedNameOf(n))
				verdicts++
				return
			}
			names := wanted
			if n == nil || n.Kind != ast.KindJsxAttributes {
				names = []string{"async"}
			}
			for _, w := range names {
				for _, matcher := range []string{"exact", "fold", "always", "never"} {
					trace := []string{}
					match := func(c, w string) bool {
						trace = append(trace, c)
						switch matcher {
						case "exact":
							return jsx.MatchExactly(c, w)
						case "fold":
							return jsx.MatchIgnoringCase(c, w)
						case "always":
							return true
						case "never":
							return false
						}
						panic("bad matcher")
					}
					s.Queries = append(s.Queries, query{i, w, "attribute", matcher})
					found := jsx.HasAttributeNamed(n, w, match)
					fmt.Fprintln(&expected, found)
					fmt.Fprintln(&expected, len(trace))
					for _, c := range trace {
						fmt.Fprintln(&expected, c)
					}
					verdicts++
				}
			}
		}
		for i, n := range actual {
			answer(i, n)
		}
		answer(-1, nil)
		corpus = append(corpus, s)
	}
	for _, source := range ordered {
		f := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/probe.tsx", Path: tspath.Path("/probe.tsx")}, source, core.ScriptKindTSX)
		observe([]*ast.Node{f.AsNode()})
	}
	// Factory controls discriminate nil lists/names and present empty aliases.
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	empty := factory.NewIdentifier("")
	local := factory.NewIdentifier("local")
	name := factory.NewIdentifier("async")
	observe([]*ast.Node{
		factory.NewImportSpecifier(false, nil, nil), factory.NewImportSpecifier(false, empty, local), factory.NewImportSpecifier(false, nil, local),
		factory.NewJsxAttributes(nil), factory.NewJsxAttributes(factory.NewNodeList([]*ast.Node{})),
		factory.NewJsxAttributes(factory.NewNodeList([]*ast.Node{factory.NewJsxAttribute(nil, nil), factory.NewJsxAttribute(empty, nil), factory.NewJsxAttribute(name, nil), factory.NewJsxAttribute(factory.NewIdentifier("ASYNC"), nil)})),
		factory.NewCallExpression(factory.NewToken(ast.KindImportKeyword), nil, nil, nil, 0),
		factory.NewCallExpression(factory.NewMetaProperty(ast.KindImportKeyword, factory.NewIdentifier("defer")), nil, nil, factory.NewNodeList([]*ast.Node{factory.NewStringLiteral("", 0)}), 0),
		factory.NewCallExpression(factory.NewMetaProperty(ast.KindNewKeyword, factory.NewIdentifier("defer")), nil, nil, factory.NewNodeList([]*ast.Node{factory.NewStringLiteral("m", 0)}), 0),
	})
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct test strings and controls, %d Go verdicts\n", len(consumers), len(corpus), verdicts)
}
