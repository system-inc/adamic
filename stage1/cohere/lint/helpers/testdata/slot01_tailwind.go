package tailwind

import (
	"fmt"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// Observe the real reader's private state after its real factory ran.
func AdamicFactoryVerdicts(settings ClassLiteralSettings, names []string) []string {
	reader := NewClassLiteralReader(settings)
	result := []string{fmt.Sprintf("%d %d %d %t", len(reader.attributeNames), len(reader.calleeNames), len(reader.variablePatterns), reader.values == nil)}
	for _, name := range names {
		result = append(result, fmt.Sprintf("%t %t", reader.attributeNames[name], reader.calleeNames[name]))
	}
	for _, pattern := range reader.variablePatterns {
		result = append(result, pattern.String())
		for _, name := range names {
			result = append(result, fmt.Sprint(pattern.MatchString(name)))
		}
	}
	return result
}

// Compare the real cache, including shared slice results and a cached nil-node result.
func AdamicMemoVerdicts(source string) []string {
	code := "const a = <div className={" + strconv.Quote(source) + "}/>; const b = <div className={" + strconv.Quote(source) + "}/>;"
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/fixture.tsx", Path: tspath.Path("/fixture.tsx")}, code, core.ScriptKindTSX)
	var nodes []*ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindJsxAttribute {
			nodes = append(nodes, node)
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if len(nodes) != 2 {
		panic("memo fixture nodes")
	}
	nodes = append(nodes, nil)
	result := []string{}
	for _, bound := range []bool{false, true} {
		reader := NewClassLiteralReader(DefaultClassLiteralSettings())
		if bound {
			reader.values = map[*ast.Node]classValues{}
		}
		for index, id := range []int{0, 0, 1, 0, 2, 2} {
			value := reader.classValuesIn(nodes[id])
			texts := []string{}
			for _, literal := range value.literals {
				texts = append(texts, literal.Text)
			}
			result = append(result, fmt.Sprintf("%d %d %d", len(texts), len(value.templates), len(reader.values)))
			result = append(result, texts...)
			if index == 0 && len(value.literals) > 0 {
				value.literals[0].Text += "!"
			}
		}
	}
	return result
}

func AdamicLiteralVerdicts(node *ast.Node, origin string) (int, int, string) {
	token := rule.TokenRange(ast.GetSourceFileOfNode(node), node)
	literal := classLiteralFrom(node, ClassLiteralOrigin(origin))
	if literal.Node != node {
		panic("literal node identity changed")
	}
	return token.Pos(), token.End(), fmt.Sprintf("%d %d %t %t\n%s\n%s", literal.Range.Pos(), literal.Range.End(), literal.Edges.Leading, literal.Edges.Trailing, literal.Text, literal.Origin)
}
