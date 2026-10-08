//go:build lintoracle

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
)

type Source struct{ Rule, File, Source string }
type Node struct {
	Kind, Text                string
	Name, TagName, Attributes int
	Properties                []int
	PropertiesPresent         bool
}
type Pair struct{ Left, Right string }
type Query struct {
	Node   int
	Wanted string
	Fold   bool
}
type Corpus struct {
	Nodes   []Node
	Pairs   []Pair
	Queries []Query
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	data, e := os.ReadFile(os.Args[1])
	must(e)
	var sources []Source
	must(json.Unmarshal(data, &sources))
	sources = append(sources, Source{Source: `const a = <script SRC="x" src async={false} {...p} xlink:href="x" />; const b = <Foo.script src />; const c = <script:tag />; const d = <script src></script>; const z = <>hello</>;`})
	c := Corpus{}
	var pointers []*ast.Node
	ids := map[*ast.Node]int{nil: -1}
	var add func(*ast.Node) int
	add = func(n *ast.Node) int {
		if id, ok := ids[n]; ok {
			return id
		}
		id := len(c.Nodes)
		ids[n] = id
		pointers = append(pointers, n)
		c.Nodes = append(c.Nodes, Node{Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), Name: -1, TagName: -1, Attributes: -1, Properties: []int{}})
		n.ForEachChild(func(child *ast.Node) bool { add(child); return false })
		return id
	}
	for _, row := range sources {
		file := tspath.RootedFilePathFromAbsolute("/fixture.tsx")
		f := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: file, PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute("/fixture.tsx"))}, row.Source, core.ScriptKindTSX)
		add(f.AsNode())
	}
	texts := map[string]bool{"": true, "script": true, "src": true, "SRC": true, "async": true, "defer": true, "K": true, "k": true, "ſ": true, "S": true, "𐐀": true, "𐐨": true, "ß": true, "SS": true}
	for i, n := range pointers {
		v := &c.Nodes[i]
		switch n.Kind {
		case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
			v.Text = n.Text()
			texts[v.Text] = true
		}
		switch n.Kind {
		case ast.KindJsxAttribute:
			v.Name = ids[n.AsJsxAttribute().Name()]
		case ast.KindJsxOpeningElement:
			v.TagName = ids[n.AsJsxOpeningElement().TagName]
			v.Attributes = ids[n.AsJsxOpeningElement().Attributes]
		case ast.KindJsxSelfClosingElement:
			v.TagName = ids[n.AsJsxSelfClosingElement().TagName]
			v.Attributes = ids[n.AsJsxSelfClosingElement().Attributes]
		case ast.KindJsxAttributes:
			p := n.AsJsxAttributes().Properties
			v.PropertiesPresent = p != nil
			if p != nil {
				for _, p := range p.Nodes {
					v.Properties = append(v.Properties, ids[p])
				}
			}
		}
	}
	ordered := []string{}
	for s := range texts {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	targets := []string{"", "script", "link", "head", "style", "a", "form", "img", "input", "button", "iframe", "title", "hr", "src", "SRC", "async", "defer", "href", "rel", "target", "download", "jsx", "is", "dangerouslySetInnerHTML", "key", "type", "sandbox", "children", "ref"}
	var want strings.Builder
	// AttributeName and ElementParts on every projected node, including declining kinds.
	for _, n := range pointers {
		name, named := jsx.AttributeName(n)
		tag, attrs := jsx.ElementParts(n)
		encoded := ""
		for _, unit := range utf16.Encode([]rune(name)) {
			encoded += fmt.Sprintf("%d,", unit)
		}
		fmt.Fprintf(&want, "%s|%t|%d|%d\n", encoded, named, ids[tag], ids[attrs])
	}
	for _, left := range ordered {
		rights := append([]string{left, strings.ToUpper(left), strings.ToLower(left), left + "x", "", "K", "k", "ſ", "S", "𐐀", "𐐨"}, targets...)
		for _, right := range rights {
			c.Pairs = append(c.Pairs, Pair{left, right})
			fmt.Fprintf(&want, "%t|%t\n", jsx.MatchExactly(left, right), jsx.MatchIgnoringCase(left, right))
		}
	}
	for i, n := range pointers {
		if n.Kind != ast.KindIdentifier && n.Kind != ast.KindJsxAttributes && n.Kind != ast.KindPropertyAccessExpression && n.Kind != ast.KindJsxNamespacedName {
			continue
		}
		for _, name := range targets {
			for _, fold := range []bool{false, true} {
				c.Queries = append(c.Queries, Query{i, name, fold})
				match := jsx.MatchExactly
				if fold {
					match = jsx.MatchIgnoringCase
				}
				fmt.Fprintf(&want, "%t|%t\n", jsx.IsIntrinsicElementNamed(n, name), jsx.HasAttributeNamed(n, name, match))
			}
		}
	}
	// nil inputs exercise the explicit optional contracts.
	c.Queries = append(c.Queries, Query{-1, "src", false})
	fmt.Fprintf(&want, "%t|%t\n", jsx.IsIntrinsicElementNamed(nil, "src"), jsx.HasAttributeNamed(nil, "src", jsx.MatchExactly))
	data, e = json.Marshal(c)
	must(e)
	must(os.WriteFile(os.Args[2], data, 0644))
	must(os.WriteFile(os.Args[3], []byte(want.String()), 0644))
	fmt.Printf("%d sources, %d nodes (AttributeName/ElementParts), %d matcher pairs, %d intrinsic/attribute queries\n", len(sources), len(c.Nodes), len(c.Pairs), len(c.Queries))
}
