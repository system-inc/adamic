package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	compilerchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	hir "github.com/system-inc/adamic/stage1/cohere/typeaware/wave_26_react_reporting/rawhir"
)

// reactHIR returns raw graph data, including checker type alias/symbol names.
// All three lint validators remain native consumers; this never calls a rule.
func (p *Program) reactHIR(out *fields, c *compilerchecker.Checker, node *ast.Node, question string) (string, error) {
	if question != "react-hir\nplain" && question != "react-hir\nmemo" {
		return "", fmt.Errorf("react-hir requires plain or memo view")
	}
	if node.Kind != ast.KindSourceFile && !ast.IsFunctionLike(node) {
		return "", fmt.Errorf("react-hir requires a source file or function")
	}
	text, err := hir.EncodeFunctions(c, node, question == "react-hir\nmemo")
	if err != nil {
		return "", err
	}
	out.text(text)
	return out.String(), nil
}
