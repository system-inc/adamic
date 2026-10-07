package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	parser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Node struct {
	Kind, Text       string
	Expression, Name int
	ReactIdentifier  bool
}
type Source struct{ Rule, File, Source string }
type Corpus struct {
	Nodes            []Node
	Names, MathNames []string
	Want             string
	Sources          int
}

func main() {
	var inputs []Source
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err = json.NewDecoder(file).Decode(&inputs); err != nil {
		panic(err)
	}
	controls := []string{
		"createClass({}); createReactClass({}); React.createClass({}); React.createReactClass({}); (React).createClass({}); (((React.createReactClass)))({}); ((createClass))({}); Preact.createClass({}); react.createClass({}); React['createClass']({}); (React as unknown).createClass({}); React.createClass; createClass; React?.createReactClass({}); React.#createClass({}); new React.createClass({});",
		"const C=<link {...props} href='&#47;about' rel='style&amp;sheet'/>; const D=<div href={x} href='later'/>; const E=<div href href='later'/>; const F=<div href='' href='later'/>;",
		"const A=<div href={'expression'} rel={`template`} src='&amp;lt;' target='&#x1f600;'/>; const B=<svg xlink:href='declined' href='actual'/>; const C=<div HREF='upper' REL='case' href='lower'/>;",
		"const A=<div K='kelvin' k='latin' σ='sigma' Σ='capital' ſ='long-s' s='short-s' İ='dotted' i='latin-i'/>; const B=<div href='&unknown; &#0; &#xD800; &#x110000; &amp; &#39; &quot;'/>;",
		"\\u0063reateClass({}); React.\\u0063reateReactClass({}); const createclass=1; const createReactCLASS=1; const 世界=1;",
	}
	for _, source := range controls {
		inputs = append(inputs, Source{File: "/control.tsx", Source: source})
	}
	corpus := Corpus{}
	pointers := []*ast.Node{}
	ids := map[*ast.Node]int{}
	names := map[string]bool{"": true, "createClass": true, "createReactClass": true, "CreateClass": true, "createclass": true, "createClassExtra": true, "createReactClassExtra": true, "createClass\x00": true, " createClass": true, "createClass ": true}
	mathTexts := []string{}
	var add func(*ast.Node) int
	add = func(n *ast.Node) int {
		if n == nil {
			return -1
		}
		if id, ok := ids[n]; ok {
			return id
		}
		id := len(pointers)
		ids[n] = id
		pointers = append(pointers, n)
		v := Node{Kind: "Other", Expression: -1, Name: -1}
		corpus.Nodes = append(corpus.Nodes, v)
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
			mathTexts = append(mathTexts, v.Text)
			names[v.Text] = true
		case ast.KindNoSubstitutionTemplateLiteral:
			v.Kind = "NoSubstitutionTemplateLiteral"
			v.Text = n.Text()
			mathTexts = append(mathTexts, v.Text)
			names[v.Text] = true
		case ast.KindCallExpression:
			v.Kind = "CallExpression"
			v.Expression = add(n.AsCallExpression().Expression)
		case ast.KindPropertyAccessExpression:
			v.Kind = "PropertyAccessExpression"
			v.Expression = add(n.AsPropertyAccessExpression().Expression)
			v.Name = add(n.AsPropertyAccessExpression().Name())
		case ast.KindParenthesizedExpression:
			v.Kind = "ParenthesizedExpression"
			v.Expression = add(n.AsParenthesizedExpression().Expression)
		}
		corpus.Nodes[id] = v
		return id
	}
	seen := map[string]bool{}
	for _, input := range inputs {
		mathTexts = append(mathTexts, input.Source)
		input.File = "/fixture" + filepath.Ext(input.File)
		key := input.File + "\x00" + input.Source
		if seen[key] {
			continue
		}
		seen[key] = true
		corpus.Sources++
		kind := core.ScriptKindTS
		if strings.HasSuffix(input.File, ".tsx") {
			kind = core.ScriptKindTSX
		} else if strings.HasSuffix(input.File, ".jsx") {
			kind = core.ScriptKindJSX
		} else if strings.HasSuffix(input.File, ".js") {
			kind = core.ScriptKindJS
		}
		source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: input.File, Path: tspath.Path(input.File)}, input.Source, kind)
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
	for id, n := range pointers {
		corpus.Nodes[id].ReactIdentifier = react.AdamicSlot02IsReactIdentifier(n)
	}
	corpus.MathNames = collapse.AdamicSlot02MathNames(mathTexts)

	for name := range names {
		corpus.Names = append(corpus.Names, name)
	}
	sort.Strings(corpus.Names)
	var want strings.Builder
	for _, name := range corpus.Names {
		fmt.Fprintf(&want, "name:%t\n", react.AdamicSlot02CreateClassName(name))
	}
	for id := -1; id < len(pointers); id++ {
		var n *ast.Node
		if id >= 0 {
			n = pointers[id]
		}
		react.AdamicSlot02ResetIdentifierProbe()
		answer := react.IsEs5ComponentCall(n)
		last := -1
		if react.AdamicSlot02IdentifierNode != nil {
			last = ids[react.AdamicSlot02IdentifierNode]
		}
		fmt.Fprintf(&want, "call:%t:%d:%d:%s\n", answer, react.AdamicSlot02IdentifierCalls, last, react.AdamicSlot02IdentifierWanted)
	}
	for _, name := range corpus.MathNames {
		fmt.Fprintf(&want, "math:%t\n", collapse.AdamicSlot02MathFunctionName(name))
	}

	corpus.Want = want.String()
	if err = json.NewEncoder(os.Stdout).Encode(corpus); err != nil {
		panic(err)
	}
}
