package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// Named AST relationships only. The native consumer builds and analyses its graph.
func (p *Program) processNodeFields(out *fields, node *ast.Node, question string) (string, error) {
	if question != "process-node-fields" {
		return "", fmt.Errorf("unexpected process-node-fields suffix")
	}
	body := node.Body()
	var initializer, condition, incrementor, statement, expression, thenNode, elseNode, tryBlock, catchNode, finallyBlock *ast.Node
	switch node.Kind {
	case ast.KindClassStaticBlockDeclaration:
		body = node.AsClassStaticBlockDeclaration().Body
	case ast.KindVariableDeclaration:
		initializer = node.AsVariableDeclaration().Initializer
	case ast.KindParameter:
		initializer = node.AsParameterDeclaration().Initializer
	case ast.KindPropertyDeclaration:
		initializer = node.AsPropertyDeclaration().Initializer
	case ast.KindBindingElement:
		initializer = node.AsBindingElement().Initializer
	case ast.KindForStatement:
		s := node.AsForStatement()
		initializer = s.Initializer
		condition = s.Condition
		incrementor = s.Incrementor
		statement = s.Statement
	case ast.KindForInStatement, ast.KindForOfStatement:
		s := node.AsForInOrOfStatement()
		initializer = s.Initializer
		expression = s.Expression
		statement = s.Statement
	case ast.KindIfStatement:
		s := node.AsIfStatement()
		expression = s.Expression
		thenNode = s.ThenStatement
		elseNode = s.ElseStatement
	case ast.KindWhileStatement:
		s := node.AsWhileStatement()
		expression = s.Expression
		statement = s.Statement
	case ast.KindDoStatement:
		s := node.AsDoStatement()
		expression = s.Expression
		statement = s.Statement
	case ast.KindTryStatement:
		s := node.AsTryStatement()
		tryBlock = s.TryBlock
		catchNode = s.CatchClause
		finallyBlock = s.FinallyBlock
	}
	refs := []*ast.Node{node.Name(), node.Type(), body, initializer, condition, incrementor, statement, expression, thenNode, elseNode, tryBlock, catchNode, finallyBlock}
	write := func(n *ast.Node) {
		out.yes(n != nil)
		if n != nil {
			out.text(strings.TrimPrefix(n.Kind.String(), "Kind"))
			out.number(uint64(n.Pos()))
			out.number(uint64(n.End()))
		}
	}
	for _, ref := range refs {
		write(ref)
	}
	var parameters []*ast.Node
	if node.FunctionLikeData() != nil && node.FunctionLikeData().Parameters != nil {
		parameters = node.Parameters()
	}
	out.number(uint64(len(parameters)))
	for _, parameter := range parameters {
		write(parameter)
	}
	return out.String(), nil
}
