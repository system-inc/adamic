package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) hasAsync(modules []*ast.SourceFile) bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindAwaitExpression || (ast.IsFunctionLike(node) && ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)) {
			found = true
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found
}

// lowerAsync deliberately has a closed grammar. No unknown statement or expression is erased.
// It reuses checked primitive expression lowering but not synchronous lifetime optimization.
func (l *lowering) lowerAsync(modules []*ast.SourceFile) (*ir.Program, error) {
	if len(modules) != 1 {
		return nil, l.notYet(modules[0].AsNode(), "async dependency-module evaluation and imported runtime effects")
	}
	plan := &ir.AsyncProgram{Generated: ir.AsyncGeneratedTypes()}
	l.result.Async = plan
	l.functions = map[*ast.Symbol]int{}
	declarations := []*ast.Node{}
	for _, module := range modules {
		for _, node := range module.Statements.Nodes {
			if node.Kind != ast.KindFunctionDeclaration {
				continue
			}
			if !ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
				return nil, l.notYet(node, "synchronous functions in an async module")
			}
			if node.Name() == nil {
				return nil, l.notYet(node, "anonymous/default async declarations")
			}
			if len(node.TypeParameters()) != 0 {
				return nil, l.notYet(node, "generic async functions")
			}
			index := len(plan.Functions)
			l.functions[l.symbol(node.Name())] = index
			returns := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(node))
			if !l.isLibraryType(returns, "Promise") {
				return nil, l.notYet(node, "async return types other than the library Promise")
			}
			arguments := l.checker.GetTypeArguments(returns)
			if len(arguments) != 1 {
				return nil, l.notYet(node, "full Promise compatibility")
			}
			result, err := l.asyncType(arguments[0], node)
			if err != nil {
				return nil, err
			}
			plan.Functions = append(plan.Functions, ir.AsyncFunction{Name: node.Name().Text(), Returns: result})
			declarations = append(declarations, node)
		}
	}
	for index, node := range declarations {
		l.functionIndex = index
		fn := &plan.Functions[index]
		for _, parameter := range node.Parameters() {
			if !ast.IsIdentifier(parameter.Name()) || parameter.AsParameterDeclaration().Initializer != nil || parameter.AsParameterDeclaration().QuestionToken != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil {
				return nil, l.notYet(parameter, "async optional, default or destructured parameters")
			}
			if _, err := l.asyncType(l.checker.GetTypeAtLocation(parameter.Name()), parameter); err != nil {
				return nil, err
			}
			local, err := l.declareLocal(parameter.Name())
			if err != nil {
				return nil, err
			}
			fn.Parameters = append(fn.Parameters, local)
			fn.Locals = append(fn.Locals, local)
		}
	}
	for index, node := range declarations {
		l.functionIndex = index
		if node.Body() == nil || node.Body().Kind != ast.KindBlock {
			return nil, l.notYet(node, "async expression bodies")
		}
		if err := l.asyncBody(&plan.Functions[index], node.Body().AsBlock().Statements.Nodes); err != nil {
			return nil, err
		}
	}
	plan.Entry = len(plan.Functions)
	plan.Functions = append(plan.Functions, ir.AsyncFunction{Name: "module"})
	l.functionIndex = plan.Entry
	var main []*ast.Node
	for _, module := range modules {
		for _, node := range module.Statements.Nodes {
			switch node.Kind {
			case ast.KindFunctionDeclaration, ast.KindImportDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
				continue
			}
			main = append(main, node)
		}
	}
	if err := l.asyncBody(&plan.Functions[plan.Entry], main); err != nil {
		return nil, err
	}
	// No recursion is admitted until eager-call exception handling is suspension-aware.
	marks := make([]uint8, len(plan.Functions))
	var recursive func(int) bool
	recursive = func(index int) bool {
		if marks[index] == 1 {
			return true
		}
		if marks[index] == 2 {
			return false
		}
		marks[index] = 1
		for _, state := range plan.Functions[index].States {
			if state.Await != nil && state.Await.Function >= 0 && recursive(state.Await.Function) {
				return true
			}
		}
		marks[index] = 2
		return false
	}
	for index, node := range declarations {
		if recursive(index) {
			return nil, l.notYet(node, "recursive async call graphs")
		}
	}
	if err := l.findCycles(modules); err != nil {
		return nil, err
	}
	return l.result, nil
}
func (l *lowering) asyncType(proven *checker.Type, where *ast.Node) (ir.Type, error) {
	if proven.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined|checker.TypeFlagsNever) != 0 {
		return 0, nil
	}
	result, known := l.representation(proven)
	if known && (result == ir.Number || result == ir.Boolean || result == ir.String) {
		return result, nil
	}
	return 0, l.notYet(where, "async object, union, Promise-handle or thenable payloads (full Promise compatibility)")
}

// asyncTypeOf observes a named async declaration without making a synchronous wrapper.
// The async grammar refuses assignments to these declarations, and reading a hoisted function
// identifier has no effects. Resolve the symbol so a same-named primitive parameter is not folded.
func (l *lowering) asyncTypeOf(node *ast.Node) (ir.Expression, bool) {
	if l.result.Async == nil {
		return nil, false
	}
	operand := ast.SkipParentheses(node.AsTypeOfExpression().Expression)
	if !ast.IsIdentifier(operand) {
		return nil, false
	}
	if _, declared := l.functions[l.symbol(operand)]; !declared {
		return nil, false
	}
	return ir.StringConstant{Index: l.constant("function")}, true
}

func (l *lowering) asyncValue(node *ast.Node) (ir.Expression, error) {
	if node == nil {
		return nil, nil
	}
	// Limit ordinary expressions to literals, reads, primitive operators, templates and string
	// methods. Calls through source functions/closures and object mutation need suspension effects.
	var failure error
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if failure != nil {
			return true
		}
		switch n.Kind {
		case ast.KindAwaitExpression:
			failure = l.notYet(n, "await nested inside an expression; write a separate declaration")
		case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression, ast.KindNewExpression, ast.KindRegularExpressionLiteral:
			failure = l.notYet(n, "async closures or mutable object payloads")
		case ast.KindCallExpression:
			callee := n.AsCallExpression().Expression
			if kind, err := l.asyncType(l.checker.GetTypeAtLocation(n), n); err != nil || kind == 0 {
				failure = l.notYet(n, "async calls producing object, array, optional or void intermediates")
			}
			if callee.Kind != ast.KindPropertyAccessExpression || l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression).Flags()&checker.TypeFlagsStringLike == 0 {
				failure = l.notYet(n, "calls inside async expressions (networking, pending I/O and full Promise surface)")
			}
		}
		n.ForEachChild(visit)
		return false
	}
	visit(node)
	if failure != nil {
		return nil, failure
	}
	kind, err := l.asyncType(l.checker.GetTypeAtLocation(node), node)
	if err != nil {
		return nil, err
	}
	if kind == 0 {
		return nil, l.notYet(node, "async undefined/void-valued expressions; use Promise.resolve() for a void await")
	}
	return l.expression(node)
}
func (l *lowering) asyncWait(node *ast.Node) (*ir.AsyncWait, error) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if ast.IsIdentifier(callee) {
			if index, found := l.functions[l.symbol(callee)]; found {
				wait := &ir.AsyncWait{Function: index, Of: l.result.Async.Functions[index].Returns}
				for _, arg := range call.Arguments.Nodes {
					value, err := l.asyncValue(arg)
					if err != nil {
						return nil, err
					}
					wait.Arguments = append(wait.Arguments, value)
				}
				return wait, nil
			}
		}
		if callee.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Promise") && callee.Name().Text() == "resolve" && len(call.Arguments.Nodes) <= 1 {
			wait := &ir.AsyncWait{Function: -1}
			if len(call.Arguments.Nodes) == 1 {
				value, err := l.asyncValue(call.Arguments.Nodes[0])
				if err != nil {
					return nil, err
				}
				wait.Value = value
				wait.Of = value.Type()
			}
			return wait, nil
		}
		return nil, l.notYet(node, "await of networking, timer/file I/O, pending-I/O cancellation or the full Promise surface; only primitive Promise.resolve and named async calls are lowered")
	}
	if node.Kind == ast.KindNewExpression {
		return nil, l.notYet(node, "Promise executors and pending-I/O cancellation (full Promise compatibility)")
	}
	value, err := l.asyncValue(node)
	if err != nil {
		return nil, err
	}
	return &ir.AsyncWait{Function: -1, Value: value, Of: value.Type()}, nil
}
func (l *lowering) asyncBody(fn *ir.AsyncFunction, nodes []*ast.Node) error {
	state := ir.AsyncState{Target: -1}
	for _, node := range nodes {
		var awaiting *ast.Node
		target := -1
		switch node.Kind {
		case ast.KindEmptyStatement, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
			continue
		case ast.KindVariableStatement:
			list := node.AsVariableStatement().DeclarationList
			if list.Flags&ast.NodeFlagsBlockScoped == 0 {
				return &Refused{Where: l.program.Where(node), What: "var", Fix: "use const or let"}
			}
			for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
				if !ast.IsIdentifier(declaration.Name()) {
					return l.notYet(declaration, "async destructuring")
				}
				initial := declaration.AsVariableDeclaration().Initializer
				if initial == nil {
					return l.notYet(declaration, "uninitialized async locals")
				}
				kind, err := l.asyncType(l.checker.GetTypeAtLocation(declaration.Name()), declaration)
				if err != nil {
					return err
				}
				if kind == 0 {
					return l.notYet(declaration, "async void-valued locals")
				}
				local, err := l.declareLocal(declaration.Name())
				if err != nil {
					return err
				}
				fn.Locals = append(fn.Locals, local)
				if initial.Kind == ast.KindAwaitExpression {
					wait, err := l.asyncWait(initial.AsAwaitExpression().Expression)
					if err != nil {
						return err
					}
					state.Await = wait
					state.Target = local
					fn.States = append(fn.States, state)
					state = ir.AsyncState{Target: -1}
				} else {
					value, err := l.asyncValue(initial)
					if err != nil {
						return err
					}
					state.Body = append(state.Body, ir.Declare{Local: local, Value: value})
				}
			}
			continue
		case ast.KindExpressionStatement:
			expression := ast.SkipParentheses(node.AsExpressionStatement().Expression)
			if expression.Kind == ast.KindAwaitExpression {
				awaiting = expression.AsAwaitExpression().Expression
			} else if expression.Kind == ast.KindCallExpression && l.isConsole(expression.AsCallExpression().Expression) {
				call := expression.AsCallExpression()
				if len(call.Arguments.Nodes) != 1 {
					return l.notYet(node, "async console arity")
				}
				value, err := l.asyncValue(call.Arguments.Nodes[0])
				if err != nil {
					return err
				}
				stream, err := l.consoleStream(call.Expression)
				if err != nil {
					return err
				}
				if value.Type() != ir.String {
					return l.notYet(node, "async console values other than strings; use a template literal")
				}
				state.Body = append(state.Body, ir.WriteLine{Stream: stream, Value: value})
				continue
			} else if expression.Kind == ast.KindCallExpression && ast.IsIdentifier(expression.AsCallExpression().Expression) {
				if _, exists := l.functions[l.symbol(expression.AsCallExpression().Expression)]; exists {
					return &Refused{Where: l.program.Where(node), What: "an unawaited async task", Fix: "await the call or return its Promise; detached tasks have no owner"}
				}
				return l.notYet(node, "async networking, cancelling pending I/O or the full Promise surface")
			} else {
				return l.notYet(node, "async assignments, callbacks or Promise methods")
			}
		case ast.KindReturnStatement:
			result := node.AsReturnStatement().Expression
			if result != nil && result.Kind == ast.KindAwaitExpression {
				wait, err := l.asyncWait(result.AsAwaitExpression().Expression)
				if err != nil {
					return err
				}
				state.Await = wait
				local := -1
				if wait.Of != 0 {
					local = len(l.result.Locals)
					l.result.Locals = append(l.result.Locals, ir.Local{Name: "return", Type: wait.Of, Function: l.functionIndex})
					fn.Locals = append(fn.Locals, local)
				}
				state.Target = local
				fn.States = append(fn.States, state)
				state = ir.AsyncState{Target: -1}
				if local >= 0 {
					state.Result = ir.Read{Local: local, Of: wait.Of}
				}
			} else {
				value, err := l.asyncValue(result)
				if err != nil {
					return err
				}
				state.Result = value
			}
			fn.States = append(fn.States, state)
			return nil
		case ast.KindThrowStatement:
			expression := node.AsThrowStatement().Expression
			if expression.Kind != ast.KindNewExpression || !l.isLibraryGlobal(expression.AsNewExpression().Expression, "Error") {
				return l.notYet(node, "async non-Error throws")
			}
			if expression.AsNewExpression().Arguments == nil || len(expression.AsNewExpression().Arguments.Nodes) != 1 {
				return l.notYet(node, "async Error construction without one explicit message, or with cause/options")
			}
			value, err := l.asyncValue(expression.AsNewExpression().Arguments.Nodes[0])
			if err != nil {
				return err
			}
			state.Throw = value
			fn.States = append(fn.States, state)
			return nil
		default:
			return l.notYet(node, "async control flow, try/catch/finally, closures or classes; only straight-line bodies are lowered")
		}
		if awaiting != nil {
			wait, err := l.asyncWait(awaiting)
			if err != nil {
				return err
			}
			state.Await = wait
			state.Target = target
			fn.States = append(fn.States, state)
			state = ir.AsyncState{Target: -1}
		}
	}
	fn.States = append(fn.States, state)
	return nil
}
