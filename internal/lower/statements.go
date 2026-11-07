// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// statements lowers a list of statements.
func (lowering *lowering) statements(nodes []*ast.Node) ([]ir.Statement, error) {
	lowered, nestedDeclarationsError := lowering.nestedDeclarations(nodes)
	if nestedDeclarationsError != nil {
		return nil, nestedDeclarationsError
	}
	for _, node := range nodes {
		statements, statementError := lowering.statement(node)
		if statementError != nil {
			return nil, statementError
		}
		lowered = append(lowered, statements...)
	}
	return lowered, nil
}

// statement lowers one statement to none, one or several.
func (lowering *lowering) statement(node *ast.Node) (result []ir.Statement, statementError error) {
	if lowering.debug {
		defer func() {
			// Erased and hoisted declarations have no marker at their old location.
			// A loop marker precedes its lowered block, preserving counter positions.
			if statementError == nil && len(result) > 0 {
				result = append([]ir.Statement{lowering.debugSourceLine(node)}, result...)
			}
		}()
	}
	switch node.Kind {
	case ast.KindModuleDeclaration:
		return lowering.namespaceBody(node)
	case ast.KindDebuggerStatement:
		// Preserve it for JavaScript; native builds have no attached debugger.
		return []ir.Statement{ir.Debugger{}}, nil
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEmptyStatement:
		// Types erase to nothing, and so does an empty statement.
		return nil, nil
	case ast.KindImportDeclaration:
		// What an import brings in is resolved through the checker at each use, and the module it
		// names runs first (moduleOrder).
		return nil, nil
	case ast.KindExportDeclaration:
		if clause := node.AsExportDeclaration().ExportClause; clause == nil || clause.Kind == ast.KindNamedExports {
			// The checker resolves each live binding, and moduleOrder runs re-export dependencies.
			return nil, nil
		}
		return nil, &Refused{
			Where: lowering.program.Where(node),
			What:  "a namespace export",
			Fix:   "export named bindings",
		}
	case ast.KindExportAssignment:
		return nil, &Refused{Where: lowering.program.Where(node), What: describe(node), Fix: "export named bindings"}
	case ast.KindFunctionDeclaration:
		if generatorDeclaration(node) && !loweredGenerator(node) {
			// A generator in a block at a module's top level is declared by no one, so it stops here,
			// where the statements reach it, rather than at a call that finds no function.
			return nil, lowering.generatorNotYet(node, generatorForm(node))
		}
		if node.Parent != nil && (node.Parent.Kind == ast.KindCaseClause || node.Parent.Kind == ast.KindDefaultClause) {
			// Switch functions are initialized at entry, before dispatch.
			return nil, nil
		}
		if lowering.function != nil {
			if local, ok := lowering.locals[lowering.symbol(
				node.Name(),
			)]; ok && lowering.result.Locals[local].NestedFunction != 0 {
				return nil, nil
			}
			return nil, lowering.notYet(node, "a block-scoped nested function declaration")
		}
		// Lowered already, by declareModule.
		return nil, nil
	case ast.KindEnumDeclaration:
		return lowering.enumDeclaration(node)
	case ast.KindClassDeclaration:
		if lowering.function != nil {
			return nil, lowering.notYet(node, "a class inside a function")
		}
		return lowering.staticDeclaration(node)
	case ast.KindReturnStatement:
		return lowering.returnStatement(node)
	case ast.KindExpressionStatement:
		return lowering.expressionStatement(node.AsExpressionStatement().Expression)
	case ast.KindVariableStatement:
		if lowering.hostBindingDeclared(node) {
			// A host binding is the host's: nothing is declared for it (host_bindings.go).
			return nil, nil
		}
		return lowering.variables(node.AsVariableStatement().DeclarationList)
	case ast.KindBlock:
		body, statementsError := lowering.statements(node.AsBlock().Statements.Nodes)
		if statementsError != nil {
			return nil, statementsError
		}
		return []ir.Statement{ir.Block{Body: body}}, nil
	case ast.KindIfStatement:
		return lowering.ifStatement(node)
	case ast.KindForStatement:
		return lowering.forStatement(node)
	case ast.KindWhileStatement, ast.KindDoStatement:
		return lowering.whileStatement(node)
	case ast.KindForInStatement:
		return lowering.forIn(node)
	case ast.KindForOfStatement:
		if body, handled, typedArrayForOfError := lowering.typedArrayForOf(node); handled {
			return body, typedArrayForOfError
		}
		return lowering.forOf(node)
	case ast.KindSwitchStatement:
		return lowering.switchStatement(node)
	case ast.KindLabeledStatement:
		return lowering.labeled(node)
	case ast.KindBreakStatement, ast.KindContinueStatement:
		label := ""
		if node.Label() != nil {
			label = node.Label().Text()
		}
		if node.Kind == ast.KindBreakStatement {
			return []ir.Statement{ir.Break{Label: label}}, nil
		}
		return []ir.Statement{ir.Continue{Label: label}}, nil
	case ast.KindThrowStatement:
		return lowering.throwStatement(node)
	case ast.KindTryStatement:
		return lowering.tryStatement(node)
	}
	return nil, lowering.notYet(node, describe(node))
}

// expressionStatement lowers an expression used as a statement: a console call, an assignment, or
// ++ and --. Any other expression's value would be thrown away, and stage 0 doesn't lower that yet.
func (lowering *lowering) expressionStatement(expression *ast.Node) ([]ir.Statement, error) {
	if weakArrayCallbackError := lowering.weakArrayCallback(expression); weakArrayCallbackError != nil {
		return nil, weakArrayCallbackError
	}
	if value, known, nodeProcessEnvironmentMutationError := lowering.nodeProcessEnvironmentMutation(expression); known {
		if nodeProcessEnvironmentMutationError != nil {
			return nil, nodeProcessEnvironmentMutationError
		}
		return []ir.Statement{ir.Evaluate{Value: value}}, nil
	}
	expression = ast.SkipParentheses(expression)
	if value, handled, recordExpressionError := lowering.recordExpression(expression); handled {
		if recordExpressionError != nil {
			return nil, recordExpressionError
		}
		return []ir.Statement{ir.Evaluate{Value: value}}, nil
	}
	if statements, handled, recordDefineStatementError := lowering.recordDefineStatement(expression); handled {
		return statements, recordDefineStatementError
	}
	if statements, handled, conditionalSuperError := lowering.conditionalSuper(expression); handled {
		return statements, conditionalSuperError
	}
	switch expression.Kind {
	case ast.KindYieldExpression:
		return lowering.yieldStatement(expression)
	case ast.KindAwaitExpression, ast.KindNonNullExpression, ast.KindVoidExpression:
		value, expressionError := lowering.expression(expression)
		if expressionError != nil {
			return nil, expressionError
		}
		return []ir.Statement{ir.Evaluate{Value: value}}, nil
	case ast.KindCallExpression:
		if ast.SkipParentheses(expression.AsCallExpression().Expression).Kind == ast.KindSuperKeyword {
			return lowering.superStatement(expression)
		}
		if genericFunctionIdentityCallError := lowering.genericFunctionIdentityCall(
			expression,
		); genericFunctionIdentityCallError != nil {
			return nil, genericFunctionIdentityCallError
		}
		if optionalCallError := lowering.optionalCall(expression); optionalCallError != nil {
			return nil, optionalCallError
		}
		if function, isGenerator, generatorCalleeError := lowering.generatorCallee(
			expression,
		); generatorCalleeError != nil {
			return nil, generatorCalleeError
		} else if isGenerator {
			// The generator is made and dropped at once: its frame lets go of the arguments it kept,
			// and none of its body runs, as on Node.
			call, generatorCallError := lowering.generatorCall(expression, function)
			if generatorCallError != nil {
				return nil, generatorCallError
			}
			return []ir.Statement{ir.Evaluate{Value: call}}, nil
		}
		if lowering.isConsole(expression.AsCallExpression().Expression) {
			statement, consoleError := lowering.console(expression)
			if consoleError != nil {
				return nil, consoleError
			}
			return []ir.Statement{statement}, nil
		}
		if lowering.isPreludeFunction(expression.AsCallExpression().Expression, "panic") {
			arguments := expression.AsCallExpression().Arguments.Nodes
			if len(arguments) != 1 {
				return nil, errors.New("lower: " +
					lowering.program.Where(expression) +
					": panic takes one argument, and the checker let another count through")
			}
			message, expressionError := lowering.expression(arguments[0])
			if expressionError != nil {
				return nil, expressionError
			}
			return []ir.Statement{ir.Panic{Message: message}}, nil
		}
		if lowering.isProcessExit(expression) {
			value, _, processValueError := lowering.processValue(expression)
			return []ir.Statement{ir.Evaluate{Value: value}}, processValueError
		}
		if value, handled, typedArrayExpressionError := lowering.typedArrayExpression(expression); handled {
			if typedArrayExpressionError != nil {
				return nil, typedArrayExpressionError
			}
			return []ir.Statement{ir.Evaluate{Value: value}}, nil
		}
		// A builtin's result thrown away, like map.set(key, value) or array.push(value).
		if lowered, isBuiltin, builtinError := lowering.builtin(expression); isBuiltin {
			if builtinError != nil {
				return nil, builtinError
			}
			return []ir.Statement{ir.Evaluate{Value: lowered}}, nil
		}
		// A call for its effects, void or not.
		call, callOrMethodError := lowering.callOrMethod(expression)
		if callOrMethodError != nil {
			return nil, callOrMethodError
		}
		if call.Type() == ir.Promise {
			return nil, &Refused{
				Where: lowering.program.Where(expression),
				What:  "an unawaited async task",
				Fix:   "await the call or keep and return its Promise",
			}
		}
		return []ir.Statement{ir.Evaluate{Value: call}}, nil
	case ast.KindBinaryExpression:
		if statements, handled, hostBindingAndStatementError := lowering.hostBindingAndStatement(expression); handled {
			return statements, hostBindingAndStatementError
		}
		if expression.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
			value, commaError := lowering.comma(expression)
			if commaError != nil {
				return nil, commaError
			}
			return []ir.Statement{ir.Evaluate{Value: value}}, nil
		}
		if logicalAssignment(expression.AsBinaryExpression().OperatorToken.Kind) {
			value, logicalAssignmentError := lowering.logicalAssignment(expression)
			if logicalAssignmentError != nil {
				return nil, logicalAssignmentError
			}
			return []ir.Statement{ir.Evaluate{Value: value}}, nil
		}
		return lowering.assignment(expression)
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
		return lowering.increment(expression)
	}
	return nil, lowering.notYet(expression, describe(expression)+" as a statement")
}

// returnStatement lowers return, with or without a value.
func (lowering *lowering) returnStatement(node *ast.Node) ([]ir.Statement, error) {
	if lowering.function == nil {
		return nil, errors.New(
			"lower: " + lowering.program.Where(node) + ": return outside a function, and the checker let it through",
		)
	}
	expression := node.AsReturnStatement().Expression
	if expression != nil && lowering.function.Generator {
		return nil, lowering.generatorNotYet(node, "returning a value (its last result's value isn't built yet)")
	}
	if expression == nil {
		returned := ir.Return{}
		if lowering.function.BodyReturns() != 0 {
			returned.Value = fit(ir.Undefined{}, lowering.function.BodyReturns())
		}
		return []ir.Statement{returned}, nil
	}
	expression = ast.SkipParentheses(expression)
	// Both assignment returns build their own return, so a checked closure refuses either form
	// before them: neither would run the claim's check (Compiler, on #1w5bve2).
	if expression.Kind == ast.KindBinaryExpression &&
		ast.IsAssignmentOperator(expression.AsBinaryExpression().OperatorToken.Kind) &&
		lowering.checkedPredicateClosure(ast.GetContainingFunction(node)) {
		return nil, lowering.notYet(node, "a checked predicate returning an assignment")
	}
	if statements, handled, namespaceReturnAssignmentError := lowering.namespaceReturnAssignment(expression); handled {
		return statements, namespaceReturnAssignmentError
	}
	if expression.Kind == ast.KindBinaryExpression &&
		expression.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		return lowering.returnAssignment(expression)
	}
	if lowering.isPanicCall(expression) || lowering.isProcessExit(expression) {
		// return panic('why'): panic never returns, so there is nothing to return, and it is the panic.
		return lowering.expressionStatement(expression)
	}
	if lowering.function.BodyReturns() != 0 && lowering.isNever(expression) {
		// return fail(), fail returning never: its effects run, and the placeholder after them is held
		// as the function returns (a number's 0, a string's null pointer), never as the call's own
		// representation, which C would refuse to return from a double or a string function.
		value, neverValueError := lowering.neverValue(expression, lowering.function.BodyReturns())
		if neverValueError != nil {
			return nil, neverValueError
		}
		return []ir.Statement{ir.Return{Value: value}}, nil
	}
	// A generic return specialized to void still evaluates the call before returning.
	if lowering.function.BodyReturns() == 0 &&
		expression.Kind == ast.KindCallExpression &&
		lowering.concrete(lowering.checker.GetTypeAtLocation(expression)).Flags()&checker.TypeFlagsVoid != 0 {
		statements, expressionStatementError := lowering.expressionStatement(expression)
		if expressionStatementError != nil {
			return nil, expressionStatementError
		}
		return append(statements, ir.Return{}), nil
	}
	value, expressionError := lowering.expression(expression)
	if expressionError != nil {
		return nil, expressionError
	}
	if checkAsyncReturnError := lowering.checkAsyncReturn(expression, value); checkAsyncReturnError != nil {
		return nil, checkAsyncReturnError
	}
	if lowering.function.BodyReturns() == 0 {
		return []ir.Statement{ir.Evaluate{Value: value}, ir.Return{}}, nil
	}
	return lowering.checkedPredicateReturn(
		ast.GetContainingFunction(node),
		expression,
		fit(value, lowering.function.BodyReturns()),
	)
}

// returnAssignment preserves the right side's value, assigns once, then returns
// that same value. Reading the target again could observe another write.
func (lowering *lowering) returnAssignment(node *ast.Node) ([]ir.Statement, error) {
	if lowering.uninitializedInitializer(node.AsBinaryExpression().Right) {
		return nil, lowering.notYet(node, "returning a placeholder reset")
	}
	statements, assignmentError := lowering.assignment(node)
	if assignmentError != nil {
		return nil, assignmentError
	}
	if len(statements) != 1 {
		return nil, lowering.notYet(node, "returning a destructuring assignment")
	}
	assignment, ok := statements[0].(ir.Assign)
	if !ok {
		return nil, lowering.notYet(node, "returning a property or element assignment")
	}
	held := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "assigned", Type: assignment.Value.Type(), Function: lowering.functionIndex},
	)
	value := assignment.Value
	if checkAsyncReturnError := lowering.checkAsyncReturn(node, value); checkAsyncReturnError != nil {
		return nil, checkAsyncReturnError
	}
	read := ir.Read{Local: held, Of: value.Type()}
	assignment.Value = read
	var returned ir.Expression = read
	if origin := lowering.placeholderOrigin(
		node.AsBinaryExpression().Left,
	); origin != "" && !lowering.placeholderAllowsUnset(node) {
		returned = ir.PlaceholderUse{
			Value:  read,
			Origin: origin,
			Use:    "return",
			Path:   sourceExpression(node),
			Where:  lowering.program.Where(node),
			Of:     lowering.function.BodyReturns(),
		}
	}
	return []ir.Statement{
		ir.Declare{Local: held, Value: value},
		assignment,
		ir.Return{Value: fit(returned, lowering.function.BodyReturns())},
	}, nil
}
