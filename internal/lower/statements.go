// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// statements lowers a list of statements.
func (l *lowering) statements(nodes []*ast.Node) ([]ir.Statement, error) {
	lowered := []ir.Statement{}
	for _, node := range nodes {
		statements, err := l.statement(node)
		if err != nil {
			return nil, err
		}
		lowered = append(lowered, statements...)
	}
	return lowered, nil
}

// statement lowers one statement to none, one or several.
func (l *lowering) statement(node *ast.Node) ([]ir.Statement, error) {
	switch node.Kind {
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEmptyStatement:
		// Types erase to nothing, and so does an empty statement.
		return nil, nil
	case ast.KindImportDeclaration, ast.KindExportDeclaration:
		// What an import brings in is resolved through the checker at each use, and the module it
		// names runs first (moduleOrder).
		return nil, nil
	case ast.KindExportAssignment:
		return nil, &Refused{Where: l.program.Where(node), What: describe(node), Fix: "export where you declare: export function, export const (one name for one thing)"}
	case ast.KindFunctionDeclaration:
		if l.function != nil {
			return nil, l.notYet(node, "a function inside a function (a closure)")
		}
		// Lowered already, by declareModule.
		return nil, nil
	case ast.KindEnumDeclaration:
		return l.enumDeclaration(node)
	case ast.KindClassDeclaration:
		if l.function != nil {
			return nil, l.notYet(node, "a class inside a function")
		}
		return l.staticDeclaration(node)
	case ast.KindReturnStatement:
		return l.returnStatement(node)
	case ast.KindExpressionStatement:
		return l.expressionStatement(node.AsExpressionStatement().Expression)
	case ast.KindVariableStatement:
		return l.variables(node.AsVariableStatement().DeclarationList)
	case ast.KindBlock:
		body, err := l.statements(node.AsBlock().Statements.Nodes)
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Block{Body: body}}, nil
	case ast.KindIfStatement:
		return l.ifStatement(node)
	case ast.KindForStatement:
		return l.forStatement(node)
	case ast.KindWhileStatement, ast.KindDoStatement:
		return l.whileStatement(node)
	case ast.KindForInStatement:
		return l.forIn(node)
	case ast.KindForOfStatement:
		return l.forOf(node)
	case ast.KindSwitchStatement:
		return l.switchStatement(node)
	case ast.KindLabeledStatement:
		statement := node.AsLabeledStatement().Statement
		for statement.Kind == ast.KindLabeledStatement {
			statement = statement.AsLabeledStatement().Statement
		}
		if statement.Kind != ast.KindSwitchStatement && !ast.IsIterationStatement(statement, false) {
			return nil, l.notYet(node, "a label on a statement other than a loop or switch")
		}
		return l.statement(node.AsLabeledStatement().Statement)
	case ast.KindBreakStatement, ast.KindContinueStatement:
		if node.Label() != nil {
			if node.Kind == ast.KindContinueStatement {
				return nil, l.notYet(node, "a labeled continue")
			}
			return l.labeledBreak(node)
		}
		if node.Kind == ast.KindBreakStatement {
			return []ir.Statement{ir.Break{}}, nil
		}
		return []ir.Statement{ir.Continue{}}, nil
	case ast.KindThrowStatement:
		return l.throwStatement(node)
	case ast.KindTryStatement:
		return l.tryStatement(node)
	}
	return nil, l.notYet(node, describe(node))
}

// expressionStatement lowers an expression used as a statement: a console call, an assignment, or
// ++ and --. Any other expression's value would be thrown away, and stage 0 doesn't lower that yet.
func (l *lowering) expressionStatement(expression *ast.Node) ([]ir.Statement, error) {
	expression = ast.SkipParentheses(expression)
	if value, handled, err := l.recordExpression(expression); handled {
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Evaluate{Value: value}}, nil
	}
	if value, known, err := l.nodeProcessEnvironmentMutation(expression); known {
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Evaluate{Value: value}}, nil
	}
	if statements, handled, err := l.conditionalSuper(expression); handled {
		return statements, err
	}
	switch expression.Kind {
	case ast.KindCallExpression:
		if ast.SkipParentheses(expression.AsCallExpression().Expression).Kind == ast.KindSuperKeyword {
			return l.superStatement(expression)
		}
		if err := l.optionalCall(expression); err != nil {
			return nil, err
		}
		if l.isConsole(expression.AsCallExpression().Expression) {
			statement, err := l.console(expression)
			if err != nil {
				return nil, err
			}
			return []ir.Statement{statement}, nil
		}
		if l.isPreludeFunction(expression.AsCallExpression().Expression, "panic") {
			arguments := expression.AsCallExpression().Arguments.Nodes
			if len(arguments) != 1 {
				return nil, errors.New("lower: " + l.program.Where(expression) + ": panic takes one argument, and the checker let another count through")
			}
			message, err := l.expression(arguments[0])
			if err != nil {
				return nil, err
			}
			return []ir.Statement{ir.Panic{Message: message}}, nil
		}
		// A builtin's result thrown away, like map.set(key, value) or array.push(value).
		if lowered, isBuiltin, err := l.builtin(expression); isBuiltin {
			if err != nil {
				return nil, err
			}
			return []ir.Statement{ir.Evaluate{Value: lowered}}, nil
		}
		// A call for its effects, void or not.
		call, err := l.callOrMethod(expression)
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Evaluate{Value: call}}, nil
	case ast.KindBinaryExpression:
		return l.assignment(expression)
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
		return l.increment(expression)
	}
	return nil, l.notYet(expression, describe(expression)+" as a statement")
}

// returnStatement lowers return, with or without a value.
func (l *lowering) returnStatement(node *ast.Node) ([]ir.Statement, error) {
	if l.function == nil {
		return nil, errors.New("lower: " + l.program.Where(node) + ": return outside a function, and the checker let it through")
	}
	expression := node.AsReturnStatement().Expression
	if expression == nil {
		returned := ir.Return{}
		if l.function.Returns != 0 {
			returned.Value = fit(ir.Undefined{}, l.function.Returns)
		}
		return []ir.Statement{returned}, nil
	}
	if l.isPanicCall(expression) || l.isProcessExit(expression) {
		// return panic('why'): panic never returns, so there is nothing to return, and it is the panic.
		return l.expressionStatement(expression)
	}
	value, err := l.expression(expression)
	if err != nil {
		return nil, err
	}
	if l.function.Returns == 0 {
		return []ir.Statement{ir.Evaluate{Value: value}, ir.Return{}}, nil
	}
	return []ir.Statement{ir.Return{Value: fit(value, l.function.Returns)}}, nil
}
