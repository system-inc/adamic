package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	parser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf16"
)

type Node struct {
	Kind, Text, Keyword                 string
	Expression, Name                    int
	ArgumentsPresent, PropertiesPresent bool
	Arguments, Properties               []int
}
type Source struct{ Rule, File, Source string }
type Corpus struct {
	Nodes          []Node
	Attributes     []int
	Pairs          [][2]string
	Names          []string
	Want           string
	Sources        int
	UnicodeVersion string
}

func main() {
	var inputs []Source
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, &inputs); err != nil {
		panic(err)
	}
	controls := []string{
		`import('plain');import('two',{with:{type:'json'}});import();import(variable);import(123);import('');import(\` + "`template`" + `);import(\` + "`hole${x}`" + `);require('one');require('one','two');require();require(variable);require(123);require('');require(\` + "`template`" + `);require(\` + "`hole${x}`" + `);`,
		`Require('upper');obj.require('member');(require)('parens');require?.('optional');import.defer('deferred');import.source('source');new require('new');const requireName='require';requireName('other');`,
		`import('\u0061\n\u{1f600}');require('\u0062\x00');import('../../internal/a');const A=<script {...props} src={x} async async='again' ASYNC='upper' defer={false}/>;`,
		`const A=<svg xlink:href='namespaced' href=''/>;const B=<div K='kelvin' k='latin' σ='sigma' Σ='capital' ſ='long-s' s='short-s' İ='dotted' i='latin-i'/>;const C=<div/>;`,
		`const A=<div key key={0} key='duplicate' />; const B=<div href={x} HREF='upper' href='last'/>;`,
	}
	// Controls are ordinary source text parsed by the same external Go parser.
	for _, s := range controls {
		inputs = append(inputs, Source{File: "/controls.tsx", Source: strings.ReplaceAll(s, "\\`", "`")})
	}
	c := Corpus{UnicodeVersion: unicode.Version, Attributes: []int{-1, 0}, Names: []string{"", "href", "HREF", "async", "ASYNC", "defer", "src", "rel", "key", "k", "K", "s", "S", "σ", "Σ", "i", "styled", "download", "jsx", "is", "dangerouslySetInnerHTML"}}
	ptrs := []*ast.Node{}
	ids := map[*ast.Node]int{}
	names := map[string]bool{}
	var add func(*ast.Node) int
	add = func(n *ast.Node) int {
		if n == nil {
			return -1
		}
		if id, ok := ids[n]; ok {
			return id
		}
		id := len(ptrs)
		ids[n] = id
		ptrs = append(ptrs, n)
		c.Nodes = append(c.Nodes, Node{})
		v := Node{Kind: "Other", Expression: -1, Name: -1, Arguments: []int{}, Properties: []int{}}
		switch n.Kind {
		case ast.KindIdentifier:
			v.Kind = "Identifier"
			v.Text = n.Text()
			names[v.Text] = true
		case ast.KindPrivateIdentifier:
			v.Kind = "PrivateIdentifier"
			v.Text = n.Text()
		case ast.KindStringLiteral:
			v.Kind = "StringLiteral"
			v.Text = n.Text()
			names[v.Text] = true
		case ast.KindNoSubstitutionTemplateLiteral:
			v.Kind = "NoSubstitutionTemplateLiteral"
			v.Text = n.Text()
			names[v.Text] = true
		case ast.KindImportKeyword:
			v.Kind = "ImportKeyword"
		case ast.KindMetaProperty:
			v.Kind = "MetaProperty"
			v.Text = n.Text()
			if n.AsMetaProperty().KeywordToken == ast.KindImportKeyword {
				v.Keyword = "ImportKeyword"
			}
		case ast.KindCallExpression:
			v.Kind = "CallExpression"
			call := n.AsCallExpression()
			v.Expression = add(call.Expression)
			v.ArgumentsPresent = call.Arguments != nil
			if call.Arguments != nil {
				for _, a := range call.Arguments.Nodes {
					v.Arguments = append(v.Arguments, add(a))
				}
			}
		case ast.KindJsxAttributes:
			v.Kind = "JsxAttributes"
			p := n.AsJsxAttributes().Properties
			v.PropertiesPresent = p != nil
			if p != nil {
				for _, a := range p.Nodes {
					v.Properties = append(v.Properties, add(a))
				}
			}
			c.Attributes = append(c.Attributes, id)
		case ast.KindJsxAttribute:
			v.Kind = "JsxAttribute"
			v.Name = add(n.AsJsxAttribute().Name())
		case ast.KindJsxSpreadAttribute:
			v.Kind = "JsxSpreadAttribute"
		}
		c.Nodes[id] = v
		return id
	}
	seen := map[string]bool{}
	for _, input := range inputs {
		file := "/fixture" + filepath.Ext(input.File)
		key := file + "\x00" + input.Source
		if seen[key] {
			continue
		}
		seen[key] = true
		c.Sources++
		kind := core.ScriptKindTS
		if strings.HasSuffix(file, ".tsx") {
			kind = core.ScriptKindTSX
		} else if strings.HasSuffix(file, ".jsx") {
			kind = core.ScriptKindJSX
		} else if strings.HasSuffix(file, ".js") {
			kind = core.ScriptKindJS
		}
		source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(file), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(file))}, input.Source, kind)
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			add(n)
			n.ForEachChild(walk)
			return false
		}
		walk(source.AsNode())
	}
	// Factory controls add nil lists, nil attribute names/properties and an empty
	// identifier name. Invalid require ASTs with nil argument lists are excluded:
	// the actual Go predicate panics there, outside the parser-node contract.
	f := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	add(f.NewJsxAttributes(nil))
	add(f.NewJsxAttributes(f.NewNodeList([]*ast.Node{nil, f.NewJsxAttribute(nil, nil), f.NewJsxAttribute(f.NewIdentifier(""), nil)})))
	add(f.NewCallExpression(f.NewToken(ast.KindImportKeyword), nil, nil, nil, 0))
	add(f.NewCallExpression(f.NewMetaProperty(ast.KindImportKeyword, f.NewIdentifier("defer")), nil, nil, f.NewNodeList([]*ast.Node{f.NewStringLiteral("factory", 0)}), 0))
	pairSet := map[[2]string]bool{}
	pair := func(a, b string) {
		p := [2]string{a, b}
		if !pairSet[p] {
			pairSet[p] = true
			c.Pairs = append(c.Pairs, p)
		}
	}
	for r := rune(0); r <= unicode.MaxRune; r++ {
		q := unicode.SimpleFold(r)
		if q == r {
			continue
		}
		pair(string(r), string(q))
		pair(string(q), string(r))
		pair("a"+string(r)+"Z", "A"+string(q)+"z")
		pair(string(r)+"!", string(q))
		pair(string(r), string(r+1))
	}
	for _, p := range [][2]string{{"", ""}, {"", "a"}, {"a", ""}, {"Straße", "STRASSE"}, {"İ", "i"}, {"ı", "I"}, {"σ", "ς"}, {"K", "k"}, {"ſ", "S"}, {"ﬃ", "ffi"}, {"é", "e\u0301"}, {"a\x00b", "A\x00B"}, {"a\nb", "A\nB"}, {"世界", "世界"}} {
		pair(p[0], p[1])
	}
	sorted := []string{}
	for s := range names {
		sorted = append(sorted, s)
	}
	sort.Strings(sorted)
	for _, s := range sorted {
		pair(s, s)
		pair(s, strings.ToUpper(s))
		pair(s, strings.ToLower(s))
		pair(s, s+"!")
	}
	var want strings.Builder
	for _, p := range c.Pairs {
		fmt.Fprintf(&want, "fold:%t\n", jsx.MatchIgnoringCase(p[0], p[1]))
	}
	for id := -1; id < len(ptrs); id++ {
		var n *ast.Node
		if id >= 0 {
			n = ptrs[id]
		}
		s, ok := imports.CallExpressionSource(n)
		encoded := ""
		for _, unit := range utf16.Encode([]rune(s)) {
			encoded += fmt.Sprintf("%d,", unit)
		}
		fmt.Fprintf(&want, "call:%t:%s\n", ok, encoded)
	}
	for _, id := range c.Attributes {
		var n *ast.Node
		if id >= 0 {
			n = ptrs[id]
		}
		for _, wanted := range c.Names {
			for mode := 0; mode < 4; mode++ {
				calls := 0
				trace := ""
				matches := func(candidate, name string) bool {
					calls++
					trace += candidate + ":" + name + "|"
					switch mode {
					case 0:
						return jsx.MatchExactly(candidate, name)
					case 1:
						return jsx.MatchIgnoringCase(candidate, name)
					case 2:
						return false
					default:
						return calls == 2
					}
				}
				answer := jsx.HasAttributeNamed(n, wanted, matches)
				fmt.Fprintf(&want, "attr:%t:%d:%s\n", answer, calls, trace)
			}
		}
	}
	c.Want = want.String()
	if err = json.NewEncoder(os.Stdout).Encode(c); err != nil {
		panic(err)
	}
}
