package core

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type wave01AtomicReport struct {
	Start, End   int
	Name, Target string
	Property     bool
	ID, Message  string
}

func Wave01AtomicReportCapture() []byte {
	texts := []string{"counter += await next();", "object.field += await next();", "object[key].deep = await next();", "class C { async f(){object.#field += await next();}}", "// 😀\n计数 += await next();", "// 😀\n对象[键].值 = await next();", "object[\"a\\tb\"] += await next();", "object [key] /* comment */ . field = await next();"}
	names := []string{"counter", "object", "object", "object", "计数", "对象", "object", "object"}
	var results []wave01AtomicReport
	for i, text := range texts {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/workspace/report.ts"}, text, core.ScriptKindTS)
		var assignment *ast.Node
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			if assignment == nil && n.Kind == ast.KindBinaryExpression && isAssignmentOperatorToken(n.AsBinaryExpression().OperatorToken) {
				assignment = n
			}
			n.ForEachChild(walk)
			return false
		}
		walk(file.AsNode())
		property := i != 0 && i != 4
		ctx := rule.Context{SourceFile: file, Report: func(d rule.Diagnostic) {
			results = append(results, wave01AtomicReport{Start: d.Range.Pos(), End: d.Range.End(), Name: names[i], Target: assignmentTargetText(rule.Context{SourceFile: file}, assignment), Property: property, ID: d.Message.Id, Message: d.Message.Description})
		}}
		reportNonAtomicUpdate(ctx, atomicEvent{symbol: &ast.Symbol{Name: names[i]}, assignment: assignment, isProperty: property})
	}
	data, err := json.Marshal(results)
	if err != nil {
		panic(err)
	}
	return data
}
