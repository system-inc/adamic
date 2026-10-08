// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// statements lowers a list of statements.
func (l *lowering) statements(nodes []*ast.Node) ([]ir.Statement, error) {
	lowered, err := l.nestedDeclarations(nodes)
	if err != nil {
		return nil, err
	}
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
		// The checker resolves bindings and moduleOrder preserves module evaluation edges.
		return nil, nil
	case ast.KindExportAssignment:
		return nil, &Refused{Where: l.program.Where(node), What: describe(node), Fix: "export where you declare: export function, export const (one name for one thing)"}
	case ast.KindFunctionDeclaration:
		if node.Parent != nil && (node.Parent.Kind == ast.KindCaseClause || node.Parent.Kind == ast.KindDefaultClause) {
			// Switch functions are initialized at entry, before dispatch.
			return nil, nil
		}
		if l.function != nil {
			if local, ok := l.locals[l.symbol(node.Name())]; ok && l.result.Locals[local].NestedFunction != 0 {
				return nil, nil
			}
			if len(node.TypeParameters()) > 0 {
				return l.nestedGenericFunction(node)
			}
			return nil, l.notYet(node, "a block-scoped nested function declaration")
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
		if body, handled, err := l.typedArrayForOf(node); handled {
			return body, err
		}
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
	if statements, handled, err := l.conditionalSuper(expression); handled {
		return statements, err
	}
	// Asserted logical targets need the checked read and held receiver/index path.
	// Dispatch before generic logical-assignment expression lowering can claim them.
	if expression.Kind == ast.KindBinaryExpression {
		binary := expression.AsBinaryExpression()
		if logicalAssignment(binary.OperatorToken.Kind) && nonNullAssignmentTarget(binary.Left) != nil {
			return l.assignment(expression)
		}
	}
	switch expression.Kind {
	case ast.KindNonNullExpression:
		value, err := l.expression(expression)
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Evaluate{Value: value}}, nil
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
		if value, handled, err := l.typedArrayExpression(expression); handled {
			if err != nil {
				return nil, err
			}
			return []ir.Statement{ir.Evaluate{Value: value}}, nil
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
		l.recordOrdinaryPredicateChecks(expression.AsCallExpression())
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
	expression = ast.SkipParentheses(expression)
	if expression.Kind == ast.KindBinaryExpression && expression.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		return l.returnAssignment(expression)
	}
	if l.isPanicCall(expression) {
		// return panic('why'): panic never returns, so there is nothing to return, and it is the panic.
		return l.expressionStatement(expression)
	}
	// A generic return specialized to void still evaluates the call before returning.
	if l.function.Returns == 0 && expression.Kind == ast.KindCallExpression && l.concrete(l.checker.GetTypeAtLocation(expression)).Flags()&checker.TypeFlagsVoid != 0 {
		statements, err := l.expressionStatement(expression)
		if err != nil {
			return nil, err
		}
		return append(statements, ir.Return{}), nil
	}
	value, err := l.expression(expression)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Return{Value: fit(value, l.function.Returns)}}, nil
}

// returnAssignment preserves the right side's value, assigns once, then returns
// that same value. Reading the target again could observe another write.
func (l *lowering) returnAssignment(node *ast.Node) ([]ir.Statement, error) {
	statements, err := l.assignment(node)
	if err != nil {
		return nil, err
	}
	if len(statements) != 1 {
		return nil, l.notYet(node, "returning a destructuring assignment")
	}
	assignment, ok := statements[0].(ir.Assign)
	if !ok {
		return nil, l.notYet(node, "returning a property or element assignment")
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "assigned", Type: assignment.Value.Type(), Function: l.functionIndex})
	value := assignment.Value
	read := ir.Read{Local: held, Of: value.Type()}
	assignment.Value = read
	return []ir.Statement{ir.Declare{Local: held, Value: value}, assignment, ir.Return{Value: fit(read, l.function.Returns)}}, nil
}
