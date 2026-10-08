package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"os"
	"strings"
)

type Node struct {
	Kind       string `json:"kind"`
	Text       string `json:"text"`
	Expression int    `json:"expression"`
	Name       int    `json:"name"`
	Heritage   []int  `json:"heritage"`
	Types      []int  `json:"types"`
}
type Source struct{ Rule, File, Source string }

func main() {
	var sources []Source
	f, e := os.Open(os.Args[1])
	must(e)
	must(json.NewDecoder(f).Decode(&sources))
	f.Close()
	sources = append(sources, Source{File: "/witness.ts", Source: `class C extends React.Component {} class P extends PureComponent {} class I implements Component {} class X extends (React.Component) {} const E=class extends (React).PureComponent {}; React.useEffect(); (React).useState(); (React.useState)(); useΩ(); use2(); Other.useState(); React['useState'](); createClass({}); (React).createReactClass({}); (createReactClass)({}); Other.createClass({});`})
	nodes := []Node{}
	pointers := []*ast.Node{}
	ids := map[*ast.Node]int{}
	names := []string{"", "Component", "PureComponent", "React", "createClass", "createReactClass", "Other", "useΩ", "use2", "use", "useÉ", "Émile", "Ω", "𐐀", "\U0001d400"}
	seenNames := map[string]bool{}
	var add func(*ast.Node) int
	add = func(n *ast.Node) int {
		if n == nil {
			return -1
		}
		if i, ok := ids[n]; ok {
			return i
		}
		i := len(nodes)
		ids[n] = i
		pointers = append(pointers, n)
		v := Node{Kind: "Other", Expression: -1, Name: -1, Heritage: []int{}, Types: []int{}}
		nodes = append(nodes, v)
		switch n.Kind {
		case ast.KindIdentifier:
			v.Kind = "Identifier"
			v.Text = n.Text()
			if !seenNames[v.Text] {
				seenNames[v.Text] = true
				names = append(names, v.Text)
			}
		case ast.KindParenthesizedExpression:
			v.Kind = "ParenthesizedExpression"
			v.Expression = add(n.AsParenthesizedExpression().Expression)
		case ast.KindPropertyAccessExpression:
			v.Kind = "PropertyAccessExpression"
			v.Expression = add(n.AsPropertyAccessExpression().Expression)
			v.Name = add(n.AsPropertyAccessExpression().Name())
		case ast.KindCallExpression:
			v.Kind = "CallExpression"
			v.Expression = add(n.AsCallExpression().Expression)
		case ast.KindExpressionWithTypeArguments:
			v.Kind = "ExpressionWithTypeArguments"
			v.Expression = add(n.AsExpressionWithTypeArguments().Expression)
		case ast.KindClassDeclaration:
			v.Kind = "ClassDeclaration"
			if h := n.AsClassDeclaration().HeritageClauses; h != nil {
				for _, x := range h.Nodes {
					v.Heritage = append(v.Heritage, add(x))
				}
			}
		case ast.KindClassExpression:
			v.Kind = "ClassExpression"
			if h := n.AsClassExpression().HeritageClauses; h != nil {
				for _, x := range h.Nodes {
					v.Heritage = append(v.Heritage, add(x))
				}
			}
		case ast.KindHeritageClause:
			v.Kind = "HeritageClause"
			if h := n.AsHeritageClause().Types; h != nil {
				for _, x := range h.Nodes {
					v.Types = append(v.Types, add(x))
				}
			}
		}
		nodes[i] = v
		return i
	}
	counts := map[string]int{}
	for _, s := range sources {
		kind := core.ScriptKindTS
		if strings.HasSuffix(s.File, "tsx") {
			kind = core.ScriptKindTSX
		}
		if strings.HasSuffix(s.File, "jsx") {
			kind = core.ScriptKindJSX
		}
		if strings.HasSuffix(s.File, ".js") {
			kind = core.ScriptKindJS
		}
		sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute("/fixture" + s.File)}, s.Source, kind)
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			add(n)
			n.ForEachChild(walk)
			return false
		}
		walk(sf.AsNode())
		counts[s.Rule]++
	}
	// Every scalar catches Unicode category and stride differences, including supplementary capitals.
	for r := rune(0); r <= 0x10ffff; r++ {
		if r >= 0xd800 && r <= 0xdfff {
			continue
		}
		if react.IsLikelyComponentName(string(r)) || (r > 0 && react.IsLikelyComponentName(string(r-1))) {
			names = append(names, string(r), "use"+string(r))
		}
	}
	out, e := os.Create(os.Args[2])
	must(e)
	must(json.NewEncoder(out).Encode(map[string]any{"nodes": nodes, "names": names}))
	out.Close()
	want, e := os.Create(os.Args[3])
	must(e)
	for _, n := range pointers {
		var call *ast.CallExpression
		if n.Kind == ast.KindCallExpression {
			call = n.AsCallExpression()
		}
		fmt.Fprintf(want, "%t %t %t %t %t %t %t %t\n", react.IsEs6ComponentClass(n), react.IsEs5ComponentCall(n), react.AdamicBase(n), react.AdamicIdentifier(n, "React"), react.IsNamespacedMember(n, react.IsHookName), react.IsNamespacedMember(n, func(string) bool { return true }), react.IsNamespacedMember(n, func(string) bool { return false }), react.IsHookCall(call))
	}
	for _, name := range names {
		fmt.Fprintf(want, "%t %t %t %t\n", react.IsLikelyComponentName(name), react.IsHookName(name), react.AdamicBaseName(name), react.AdamicCreateName(name))
	}
	want.Close()
	b, _ := json.Marshal(counts)
	fmt.Printf("%d captured sources; %d nodes; %d names; consumers %s\n", len(sources), len(nodes), len(names), b)
}
func must(e error) {
	if e != nil {
		panic(e)
	}
}
