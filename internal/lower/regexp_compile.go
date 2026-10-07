package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Compile constant syntax through exactly the constructor's bytecode compiler.
// Copying a proven RegExp needs no parser: its current program is already valid.
func (l *lowering) regexCompile(node, receiver *ast.Node, args []*ast.Node) (ir.Expression, error) {
	if node.AsCallExpression().Expression.AsPropertyAccessExpression().QuestionDotToken != nil {
		return nil, l.notYet(node, "an optional RegExp.compile call")
	}
	if len(args) > 2 {
		return nil, l.notYet(node, "RegExp.compile with more than two arguments")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, err
	}
	var program ir.Expression
	if len(args) > 0 && l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(args[0])), "RegExp") {
		if len(args) > 1 && !l.intrinsicUndefined(args[1]) {
			return nil, l.notYet(node, "RegExp.compile with a RegExp pattern and evaluated flags (Annex B requires TypeError unless undefined)")
		}
		program, err = l.expression(args[0])
	} else {
		program, err = l.regexConstant(node)
	}
	if err != nil {
		return nil, err
	}
	return ir.RegExpCall{Value: value, Arguments: []ir.Expression{program}, Method: "compile", Returns: ir.Object}, nil
}

// Source/flag inference must not freeze an object mutated through any alias.
// This deliberately conservative whole-module proof covers calls in closures,
// dead branches and unrelated receivers. Direct fresh literals remain provable.
func (l *lowering) regexCompileMayMutate() bool {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found {
			return true
		}
		if node.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "compile" {
				found = true
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		visit(module.AsNode())
	}
	return found
}
