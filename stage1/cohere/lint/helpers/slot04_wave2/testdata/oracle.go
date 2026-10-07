package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	"os"
	"strings"
)

type Row struct{ Name, Source string }
type Edges struct {
	Leading  bool `json:"leading"`
	Trailing bool `json:"trailing"`
}
type Span struct {
	Expression int     `json:"expression"`
	Valid      bool    `json:"valid"`
	Edges      []Edges `json:"edges"`
}
type Node struct {
	Kind        string `json:"kind"`
	Text        string `json:"text"`
	Start       int    `json:"start"`
	End         int    `json:"end"`
	Name        int    `json:"name"`
	Initializer int    `json:"initializer"`
	Expression  int    `json:"expression"`
	Left        int    `json:"left"`
	Right       int    `json:"right"`
	WhenTrue    int    `json:"whenTrue"`
	WhenFalse   int    `json:"whenFalse"`
	Operator    string `json:"operator"`
	Elements    []int  `json:"elements"`
	Spans       []Span `json:"spans"`
}

func main() {
	path := os.Args[1]
	adapt := path == "--ast"
	if adapt {
		path = os.Args[2]
	}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var rows []Row
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	output := []map[string]any{}
	for _, row := range rows {
		kind := core.ScriptKindTS
		switch {
		case strings.HasSuffix(row.Name, ".tsx"):
			kind = core.ScriptKindTSX
		case strings.HasSuffix(row.Name, ".jsx"):
			kind = core.ScriptKindJSX
		case strings.HasSuffix(row.Name, ".js"):
			kind = core.ScriptKindJS
		}
		name := tspath.NormalizePath("/corpus/" + row.Name)
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: name, Path: tspath.Path(name)}, row.Source, kind)
		nodes := []*ast.Node{}
		ids := map[*ast.Node]int{nil: -1}
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			if _, ok := ids[n]; ok {
				return
			}
			ids[n] = len(nodes)
			nodes = append(nodes, n)
			n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
		index := func(n *ast.Node) int {
			id, ok := ids[n]
			if !ok {
				panic("raw AST field absent from arena")
			}
			return id
		}
		projected := []Node{}
		for _, n := range nodes {
			item := Node{Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), Name: -1, Initializer: -1, Expression: -1, Left: -1, Right: -1, WhenTrue: -1, WhenFalse: -1, Start: -1, End: -1, Elements: []int{}, Spans: []Span{}}
			switch n.Kind {
			case ast.KindIdentifier, ast.KindJsxNamespacedName:
				item.Text = n.Text()
			case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
				literal := tailwind.AdamicLiteral(n)
				item.Text = literal.Text
				item.Start = literal.Range.Pos()
				item.End = literal.Range.End()
			case ast.KindJsxAttribute:
				item.Name = index(n.AsJsxAttribute().Name())
				item.Initializer = index(n.AsJsxAttribute().Initializer)
			case ast.KindJsxExpression:
				item.Expression = index(n.AsJsxExpression().Expression)
			case ast.KindParenthesizedExpression:
				item.Expression = index(n.AsParenthesizedExpression().Expression)
			case ast.KindAsExpression:
				item.Expression = index(n.AsAsExpression().Expression)
			case ast.KindSatisfiesExpression:
				item.Expression = index(n.AsSatisfiesExpression().Expression)
			case ast.KindConditionalExpression:
				c := n.AsConditionalExpression()
				item.WhenTrue = index(c.WhenTrue)
				item.WhenFalse = index(c.WhenFalse)
			case ast.KindBinaryExpression:
				b := n.AsBinaryExpression()
				item.Left = index(b.Left)
				item.Right = index(b.Right)
				if b.OperatorToken != nil {
					item.Operator = strings.TrimPrefix(b.OperatorToken.Kind.String(), "Kind")
				}
			case ast.KindArrayLiteralExpression:
				if e := n.AsArrayLiteralExpression().Elements; e != nil {
					for _, v := range e.Nodes {
						item.Elements = append(item.Elements, index(v))
					}
				}
			case ast.KindTemplateExpression:
				template := n.AsTemplateExpression()
				if template.Head != nil && template.TemplateSpans != nil {
					before := template.Head.Text()
					spans := template.TemplateSpans.Nodes
					for i, s := range spans {
						span := s.AsTemplateSpan()
						entry := Span{Expression: -1, Edges: []Edges{}}
						if span != nil && span.Literal != nil {
							entry.Valid = true
							entry.Expression = index(span.Expression)
							after := span.Literal.Text()
							for mask := 0; mask < 4; mask++ {
								l, t := tailwind.AdamicHole(before, after, i == 0, i == len(spans)-1, mask >= 2, mask%2 == 1)
								entry.Edges = append(entry.Edges, Edges{l, t})
							}
							before = after
						}
						item.Spans = append(item.Spans, entry)
					}
				}
			}
			projected = append(projected, item)
		}
		if adapt {
			output = append(output, map[string]any{"name": row.Name, "ast": projected})
		} else {
			origins := []string{"Attribute", "Callee", "Variable"}
			for i := -1; i < len(nodes); i++ {
				var n *ast.Node
				if i >= 0 {
					n = nodes[i]
				}
				for _, line := range tailwind.AdamicOutputs(n, origins[(i+1)%3], ids) {
					fmt.Println(line)
				}
			}
		}
	}
	if adapt {
		if err = json.NewEncoder(os.Stdout).Encode(output); err != nil {
			panic(err)
		}
	}
}
