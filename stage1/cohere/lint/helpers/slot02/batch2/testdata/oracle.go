package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	parser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	"os"
	"path/filepath"
	"strings"
)

type Node struct {
	Kind, Text       string
	Expression, Name int
}
type Source struct{ Rule, File, Source string }
type Corpus struct {
	Nodes    []Node
	Texts    []string
	Readings []tailwind.AdamicSlot02Reading
	Want     string
	Sources  int
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
	for _, source := range []string{"React.useEffect; (React).useState; (((React))).anything; Preact.useEffect; react.useEffect; (React as unknown).useEffect; React['useEffect']; React?.useEffect; React.#useEffect; React.世界; React.\\u0075seState;", "const buttonClassName=' p-2  flex '; mergeClassNames('p-4', x?'flex':'block', `flex${x?' hidden ':''}`); const C=<div className={'p-2'}/>;", "const o={x:'a b',y:'a\u00a0b',z:'a${hole}b'};"} {
		inputs = append(inputs, Source{File: "/control.tsx", Source: source})
	}
	corpus := Corpus{Texts: []string{"", " \t\n", "a${x} b", "${x} suffix", "x${ y} z", "a\ufeffb", "a\u0085b", "a\u00a0b", "a\u200bb", "😀 é 世界", "$ { x }", "${x}${y}", "a\x00b", "a\r\nb"}}
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
		id := len(pointers)
		ids[n] = id
		pointers = append(pointers, n)
		corpus.Nodes = append(corpus.Nodes, Node{Kind: "Other", Expression: -1, Name: -1})
		v := Node{Kind: "Other", Expression: -1, Name: -1}
		switch n.Kind {
		case ast.KindIdentifier:
			v.Kind = "Identifier"
			v.Text = n.Text()
		case ast.KindParenthesizedExpression:
			v.Kind = "ParenthesizedExpression"
			v.Expression = add(n.AsParenthesizedExpression().Expression)
		case ast.KindPropertyAccessExpression:
			v.Kind = "PropertyAccessExpression"
			v.Expression = add(n.AsPropertyAccessExpression().Expression)
			v.Name = add(n.AsPropertyAccessExpression().Name())
		}
		corpus.Nodes[id] = v
		return id
	}
	seen := map[string]bool{}
	texts := map[string]bool{}
	for _, input := range inputs {
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
			if n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNoSubstitutionTemplateLiteral {
				if text := n.Text(); !texts[text] {
					texts[text] = true
					corpus.Texts = append(corpus.Texts, text)
				}
			}
			n.ForEachChild(walk)
			return false
		}
		walk(source.AsNode())
	}
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	add(factory.NewPropertyAccessExpression(factory.NewIdentifier("React"), nil, nil, 0))
	add(factory.NewPropertyAccessExpression(factory.NewIdentifier("React"), nil, factory.NewIdentifier(""), 0))
	var want strings.Builder
	for id := -1; id < len(pointers); id++ {
		var n *ast.Node
		if id >= 0 {
			n = pointers[id]
		}
		for mode := 0; mode < 3; mode++ {
			calls := 0
			name := ""
			answer := react.IsNamespacedMember(n, func(s string) bool { calls++; name = s; return mode == 0 || (mode == 1 && strings.HasPrefix(s, "use")) })
			fmt.Fprintf(&want, "%t:%d:%s\n", answer, calls, name)
		}
	}
	for _, text := range corpus.Texts {
		fields := tailwind.SplitClasses(text)
		fmt.Fprintf(&want, "split:%d\n", len(fields))
		for _, field := range fields {
			fmt.Fprintln(&want, field)
		}
	}
	count := 0
	for code := 0; code <= 0x10ffff; code++ {
		if len(tailwind.SplitClasses("a"+string(rune(code))+"b")) == 2 {
			fmt.Fprintf(&want, "space:%d\n", code)
			count++
		}
	}
	fmt.Fprintf(&want, "spaces:%d\n", count)
	// Query every actual parser node, not just the subscribed surface kinds.
	corpus.Readings = tailwind.AdamicSlot02Readings(append([]*ast.Node{nil}, pointers...))
	for _, reading := range corpus.Readings {
		fmt.Fprintf(&want, "literals:1:%t:%d\n", reading.Shared, len(reading.Literals))
		for _, literal := range reading.Literals {
			fmt.Fprintln(&want, literal)
		}
	}
	corpus.Want = want.String()
	if err = json.NewEncoder(os.Stdout).Encode(corpus); err != nil {
		panic(err)
	}
}
