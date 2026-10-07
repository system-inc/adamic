// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) ifStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsIfStatement()
	condition, err := l.condition(statement.Expression)
	if err != nil {
		return nil, err
	}
	then, err := l.statement(statement.ThenStatement)
	if err != nil {
		return nil, err
	}
	lowered := ir.If{Condition: condition, Then: then}
	if statement.ElseStatement != nil {
		if lowered.Else, err = l.statement(statement.ElseStatement); err != nil {
			return nil, err
		}
	}
	return []ir.Statement{lowered}, nil
}

// forStatement lowers for (initializer; condition; incrementor) to a loop inside a block, so the
// loop's own declarations are scoped to it.
func (l *lowering) forStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsForStatement()
	block := ir.Block{}
	if initializer := statement.Initializer; initializer != nil {
		var err error
		var initial []ir.Statement
		if initializer.Kind == ast.KindVariableDeclarationList {
			initial, err = l.variables(initializer)
		} else {
			initial, err = l.expressionStatement(initializer)
		}
		if err != nil {
			return nil, err
		}
		block.Body = append(block.Body, initial...)
	}
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}}
	for _, statement := range block.Body {
		if declared, isDeclare := statement.(ir.Declare); isDeclare && initializerIsLet(statement, node) {
			loop.PerIteration = append(loop.PerIteration, declared.Local)
		}
	}
	if statement.Condition != nil {
		condition, err := l.condition(statement.Condition)
		if err != nil {
			return nil, err
		}
		loop.Condition = condition
	}
	if statement.Incrementor != nil {
		update, err := l.expressionStatement(statement.Incrementor)
		if err != nil {
			return nil, err
		}
		loop.Update = update
	}
	body, err := l.statement(statement.Statement)
	if err != nil {
		return nil, err
	}
	loop.Body = body
	block.Body = append(block.Body, loop)
	return []ir.Statement{block}, nil
}

// whileStatement lowers while and do...while.
func (l *lowering) whileStatement(node *ast.Node) ([]ir.Statement, error) {
	var expression, body *ast.Node
	checkAfter := node.Kind == ast.KindDoStatement
	if checkAfter {
		expression, body = node.AsDoStatement().Expression, node.AsDoStatement().Statement
	} else {
		expression, body = node.AsWhileStatement().Expression, node.AsWhileStatement().Statement
	}
	condition, err := l.condition(expression)
	if err != nil {
		return nil, err
	}
	loweredBody, err := l.statement(body)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Loop{Condition: condition, Body: loweredBody, CheckAfter: checkAfter}}, nil
}

// condition applies JavaScript's ToBoolean, with one evaluation of the operand.
func (l *lowering) condition(node *ast.Node) (ir.Expression, error) {
	value, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	return truthy(value), nil
}

func truthy(value ir.Expression) ir.Expression {
	if value.Type() == ir.Boolean {
		return value
	}
	return ir.Truthy{Value: value}
}

// initializerIsLet reports whether a for loop declares its variables with let, which JavaScript gives
// a fresh binding each iteration (const can't change, so a copy would be the same).
func initializerIsLet(_ ir.Statement, node *ast.Node) bool {
	initializer := node.AsForStatement().Initializer
	return initializer != nil && initializer.Kind == ast.KindVariableDeclarationList && initializer.Flags&ast.NodeFlagsLet != 0
}
