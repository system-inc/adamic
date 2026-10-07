package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func (p *Program) syntaxMetadata(out *fields, node *ast.Node, question string) error {
	if question != "syntax-metadata" {
		return fmt.Errorf("unexpected syntax-metadata suffix")
	}
	out.number(uint64(node.Flags))
	out.number(uint64(node.ModifierFlags()))
	out.yes(ast.IsTypeNode(node))
	var initializer *ast.Node
	switch node.Kind {
	case ast.KindForStatement:
		initializer = node.AsForStatement().Initializer
	case ast.KindForInStatement, ast.KindForOfStatement:
		initializer = node.AsForInOrOfStatement().Expression
	}
	out.yes(initializer != nil)
	if initializer != nil {
		out.number(uint64(initializer.Pos()))
		out.number(uint64(initializer.End()))
	}
	generator := false
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		generator = node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		generator = node.AsFunctionExpression().AsteriskToken != nil
	}
	out.yes(generator)
	return nil
}
