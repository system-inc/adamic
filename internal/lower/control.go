// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) ifStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsIfStatement()
	condition, conditionError := l.condition(statement.Expression)
	if conditionError != nil {
		return nil, conditionError
	}
	var then []ir.Statement
	var statementError error
	if symbol := l.hostBindingGuard(statement.Expression); symbol != nil {
		// A guard proves an Absent host binding missing, so the then-branch is dead (host_bindings.go).
		then = []ir.Statement{l.hostBindingDeadRead(symbol)}
	} else if then, statementError = l.statement(statement.ThenStatement); statementError != nil {
		return nil, statementError
	}
	lowered := ir.If{Condition: condition, Then: then}
	if statement.ElseStatement != nil {
		if lowered.Else, statementError = l.statement(statement.ElseStatement); statementError != nil {
			return nil, statementError
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
		condition, conditionError := l.condition(statement.Condition)
		if conditionError != nil {
			return nil, conditionError
		}
		loop.Condition = condition
	}
	if statement.Incrementor != nil {
		update, expressionStatementError := l.expressionStatement(statement.Incrementor)
		if expressionStatementError != nil {
			return nil, expressionStatementError
		}
		loop.Update = update
	}
	body, statementError := l.statement(statement.Statement)
	if statementError != nil {
		return nil, statementError
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
	condition, conditionError := l.condition(expression)
	if conditionError != nil {
		return nil, conditionError
	}
	loweredBody, statementError := l.statement(body)
	if statementError != nil {
		return nil, statementError
	}
	return []ir.Statement{ir.Loop{Condition: condition, Body: loweredBody, CheckAfter: checkAfter}}, nil
}

// condition applies JavaScript ToBoolean to one evaluation of its operand.
func (l *lowering) condition(node *ast.Node) (ir.Expression, error) {
	if test, handled, methodPresenceConditionError := l.methodPresenceCondition(node); handled {
		return test, methodPresenceConditionError
	}
	value, expressionError := l.expression(node)
	if expressionError != nil {
		return nil, expressionError
	}
	return censusCondition(value), nil
}

// initializerIsLet reports whether a for loop declares its variables with let, which JavaScript gives
// a fresh binding each iteration (const can't change, so a copy would be the same).
func initializerIsLet(_ ir.Statement, node *ast.Node) bool {
	initializer := node.AsForStatement().Initializer
	return initializer != nil &&
		initializer.Kind == ast.KindVariableDeclarationList &&
		initializer.Flags&ast.NodeFlagsLet != 0
}
