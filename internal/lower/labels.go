package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) labeled(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsLabeledStatement()
	body, err := l.statement(statement.Statement)
	if err != nil {
		return nil, err
	}
	name := statement.Label.Text()
	target := statement.Statement
	for target.Kind == ast.KindLabeledStatement {
		target = target.AsLabeledStatement().Statement
	}
	switch target.Kind {
	case ast.KindForStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindForOfStatement, ast.KindForInStatement:
		labelIteration(body, name)
	}
	return []ir.Statement{ir.Labeled{Name: name, Body: body}}, nil
}

// A for's initializer is a block before its loop; consecutive labels share the same iteration.
func labelIteration(body []ir.Statement, name string) bool {
	for index := range body {
		switch statement := body[index].(type) {
		case ir.Loop:
			statement.Labels = append(statement.Labels, name)
			body[index] = statement
			return true
		case ir.ForOf:
			statement.Labels = append(statement.Labels, name)
			body[index] = statement
			return true
		case ir.Block:
			if labelIteration(statement.Body, name) {
				return true
			}
		case ir.Labeled:
			if labelIteration(statement.Body, name) {
				return true
			}
		}
	}
	return false
}
