package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// An annotation is a promise, including in code no runtime branch visits.
func (l *lowering) functionAnnotations(module *ast.SourceFile) error {
	if !strings.HasSuffix(l.program.FileName(module), ".a") {
		return nil
	}
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if node.Kind == ast.KindTypeReference && l.isLibraryType(l.checker.GetTypeAtLocation(node), "Function") {
			found = &Refused{Where: l.program.Where(node), What: "the Function annotation in .a", Fix: "write a call signature with its parameters and result, like (value: number) => number"}
			return true
		}
		return node.ForEachChild(visit)
	}
	module.AsNode().ForEachChild(visit)
	return found
}

func (l *lowering) functionCallRefusal(node *ast.Node) error {
	if node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression {
		return nil
	}
	var callee *ast.Node
	if node.Kind == ast.KindCallExpression {
		callee = node.AsCallExpression().Expression
	} else {
		callee = node.AsNewExpression().Expression
	}
	callee = ast.SkipParentheses(callee)
	opaque := l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(callee)), "Function")
	if callee.Kind == ast.KindPropertyAccessExpression {
		access := callee.AsPropertyAccessExpression()
		opaque = opaque || l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)), "Function") && (access.Name().Text() == "call" || access.Name().Text() == "apply" || access.Name().Text() == "bind")
	}
	if callee.Kind == ast.KindElementAccessExpression {
		access := callee.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind == ast.KindStringLiteral {
			opaque = opaque || l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)), "Function") && (key.Text() == "call" || key.Text() == "apply" || key.Text() == "bind")
		}
	}
	if opaque {
		return &Refused{Where: l.program.Where(node), What: "calling through the unchecked Function type", Fix: "write a call signature with the actual parameters and result before calling"}
	}
	return nil
}

func sourceFunctionLength(node *ast.Node) int {
	length := 0
	for _, parameter := range node.Parameters() {
		p := parameter.AsParameterDeclaration()
		if ast.IsIdentifier(parameter.Name()) && parameter.Name().Text() == "this" {
			continue
		}
		if p.Initializer != nil || p.DotDotDotToken != nil {
			break
		}
		length++
	}
	return length
}

// Only observations with a defined backend representation are admitted.
func (l *lowering) functionObservation(node *ast.Node) (ir.Expression, bool, error) {
	var receiver *ast.Node
	name := ""
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		if access.QuestionDotToken != nil {
			return nil, false, nil
		}
		receiver, name = access.Expression, access.Name().Text()
	} else if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if access.QuestionDotToken != nil {
			return nil, false, nil
		}
		if access.ArgumentExpression.Kind != ast.KindStringLiteral {
			return nil, false, nil
		}
		receiver, name = access.Expression, access.ArgumentExpression.Text()
	} else {
		return nil, false, nil
	}
	if !l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)), "Function") {
		return nil, false, nil
	}
	if name != "length" {
		return nil, true, l.notYet(node, "observing Function."+name+"; use a call signature and explicit metadata")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	return ir.Narrow{Value: ir.DynamicProperty{Object: fit(value, ir.Union), Name: name}, To: ir.Number}, true, nil
}
