package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

const (
	superUninitialized uint8 = 1
	superInitialized   uint8 = 2
)

// constructorSuperFlow follows normal paths. A this use must have an initialized binding
// on every path reaching it; a deferred capture before initialization stays conservative.
type constructorSuperFlow struct {
	lowering  *lowering
	instance  *instance
	breaks    []uint8
	continues []uint8
}

func (flow *constructorSuperFlow) reads(node *ast.Node, state uint8) error {
	if node == nil {
		return nil
	}
	var failed error
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if (child.Kind == ast.KindThisKeyword || (child.Kind == ast.KindSuperKeyword &&
			child.Parent != nil &&
			child.Parent.Kind == ast.KindPropertyAccessExpression)) && state&superUninitialized != 0 {
			flow.instance.unreadyThis[child] = true
		}
		if child.Kind == ast.KindCallExpression &&
			ast.SkipParentheses(child.AsCallExpression().Expression).Kind == ast.KindSuperKeyword {
			failed = flow.lowering.notYet(
				child,
				"super used as a value or in a deferred function; call it as a statement in each branch",
			)
			return true
		}
		return child.ForEachChild(visit)
	}
	visit(node)
	return failed
}

func (flow *constructorSuperFlow) list(nodes []*ast.Node, state uint8) (uint8, error) {
	for _, node := range nodes {
		next, statementError := flow.statement(node, state)
		if statementError != nil {
			return 0, statementError
		}
		state = next
	}
	return state, nil
}

func (flow *constructorSuperFlow) expression(node *ast.Node, state uint8) (uint8, error) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression &&
		ast.SkipParentheses(node.AsCallExpression().Expression).Kind == ast.KindSuperKeyword {
		for _, argument := range nodesOf(node.AsCallExpression().Arguments) {
			if readsError := flow.reads(argument, state); readsError != nil {
				return 0, readsError
			}
		}
		flow.instance.superStates[node] |= state
		if state == 0 {
			return 0, nil
		}
		// A second call throws after the base runs; it has no normal continuation.
		if state&superUninitialized == 0 {
			return 0, nil
		}
		return superInitialized, nil
	}
	if node.Kind == ast.KindConditionalExpression {
		expression := node.AsConditionalExpression()
		if readsError := flow.reads(expression.Condition, state); readsError != nil {
			return 0, readsError
		}
		left, expressionError := flow.expression(expression.WhenTrue, state)
		if expressionError != nil {
			return 0, expressionError
		}
		right, expressionError := flow.expression(expression.WhenFalse, state)
		return left | right, expressionError
	}
	return state, flow.reads(node, state)
}

func (flow *constructorSuperFlow) statement(node *ast.Node, state uint8) (uint8, error) {
	if node == nil {
		return state, nil
	}
	switch node.Kind {
	case ast.KindBlock:
		return flow.list(node.AsBlock().Statements.Nodes, state)
	case ast.KindExpressionStatement:
		return flow.expression(node.AsExpressionStatement().Expression, state)
	case ast.KindIfStatement:
		branch := node.AsIfStatement()
		if readsError := flow.reads(branch.Expression, state); readsError != nil {
			return 0, readsError
		}
		left, statementError := flow.statement(branch.ThenStatement, state)
		if statementError != nil {
			return 0, statementError
		}
		right, statementError := flow.statement(branch.ElseStatement, state)
		return left | right, statementError
	case ast.KindThrowStatement:
		return 0, flow.reads(node.AsThrowStatement().Expression, state)
	case ast.KindReturnStatement:
		return 0, flow.lowering.notYet(node, "an explicit return from a derived constructor")
	case ast.KindTryStatement:
		tried := node.AsTryStatement()
		normal, statementError := flow.statement(tried.TryBlock, state)
		if statementError != nil {
			return 0, statementError
		}
		if tried.CatchClause != nil {
			// A failure may occur on either side of super returning. The runtime flag
			// distinguishes a retry from a second successful initialization.
			caught, statementError := flow.statement(tried.CatchClause.AsCatchClause().Block, state|superInitialized)
			if statementError != nil {
				return 0, statementError
			}
			normal |= caught
		}
		if tried.FinallyBlock != nil {
			// Finally also runs on exceptional exits, before this might be initialized.
			if _, statementError := flow.statement(tried.FinallyBlock, state|normal); statementError != nil {
				return 0, statementError
			}
			normal, statementError = flow.statement(tried.FinallyBlock, normal)
		}
		return normal, statementError
	case ast.KindBreakStatement:
		if len(flow.breaks) > 0 {
			flow.breaks[len(flow.breaks)-1] |= state
		}
		return 0, nil
	case ast.KindContinueStatement:
		if len(flow.continues) > 0 {
			flow.continues[len(flow.continues)-1] |= state
		}
		return 0, nil
	case ast.KindForStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindForOfStatement:
		return flow.loop(node, state)
	case ast.KindSwitchStatement:
		return flow.switchCases(node, state)
	}
	return state, flow.reads(node, state)
}

func (lowering *lowering) prepareSuper(constructor *ast.Node, initializer int) (uint8, error) {
	instance := lowering.instance
	instance.unreadyThis = map[*ast.Node]bool{}
	instance.superStates = map[*ast.Node]uint8{}
	flow := constructorSuperFlow{lowering: lowering, instance: instance}
	for _, parameter := range constructor.Parameters() {
		if readsError := flow.reads(parameter, superUninitialized); readsError != nil {
			return 0, readsError
		}
	}
	state, statementError := flow.statement(constructor.Body(), superUninitialized)
	if statementError != nil {
		return 0, statementError
	}
	instance.superReady = len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "superReady", Type: ir.Boolean, Function: initializer},
	)
	lowering.result.Functions[initializer].Body = append(
		lowering.result.Functions[initializer].Body,
		ir.Declare{Local: instance.superReady, Value: ir.BooleanConstant{}},
	)
	return state, nil
}

func (lowering *lowering) superError(message string) ir.Statement {
	return ir.Throw{Value: ir.MakeError{
		Message: ir.StringConstant{Index: lowering.constant(message)},
		Name:    ir.StringConstant{Index: lowering.constant("ReferenceError")},
	}}
}

func (lowering *lowering) superEnd() ir.Statement {
	return ir.If{
		Condition: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: lowering.instance.superReady, Of: ir.Boolean}},
		Then: []ir.Statement{
			lowering.superError(
				"Must call super constructor in derived class before accessing 'this' or returning from derived constructor",
			),
		},
	}
}

// Conditional super used for effects has the same branch and field-initializer order as if.
func (lowering *lowering) conditionalSuper(node *ast.Node) ([]ir.Statement, bool, error) {
	if lowering.instance == nil || lowering.instance.base == nil || node.Kind != ast.KindConditionalExpression {
		return nil, false, nil
	}
	conditional := node.AsConditionalExpression()
	hasSuper := false
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if ast.IsFunctionLike(child) {
			return false
		}
		if child.Kind == ast.KindCallExpression &&
			ast.SkipParentheses(child.AsCallExpression().Expression).Kind == ast.KindSuperKeyword {
			hasSuper = true
			return true
		}
		return child.ForEachChild(visit)
	}
	visit(node)
	if !hasSuper {
		return nil, false, nil
	}
	condition, conditionError := lowering.condition(conditional.Condition)
	if conditionError != nil {
		return nil, true, conditionError
	}
	yes, expressionStatementError := lowering.expressionStatement(conditional.WhenTrue)
	if expressionStatementError != nil {
		return nil, true, expressionStatementError
	}
	no, expressionStatementError := lowering.expressionStatement(conditional.WhenFalse)
	if expressionStatementError != nil {
		return nil, true, expressionStatementError
	}
	return []ir.Statement{ir.If{Condition: condition, Then: yes, Else: no}}, true, nil
}

// Backedges may arrive with an initialized binding. Preserve zero-iteration and
// break paths too; the flag checks missing or repeated initialization at runtime.
func (flow *constructorSuperFlow) loop(node *ast.Node, state uint8) (uint8, error) {
	var body, condition, increment *ast.Node
	switch node.Kind {
	case ast.KindForStatement:
		loop := node.AsForStatement()
		body, condition, increment = loop.Statement, loop.Condition, loop.Incrementor
		if loop.Initializer != nil {
			var err error
			if loop.Initializer.Kind == ast.KindVariableDeclarationList {
				err = flow.reads(loop.Initializer, state)
			} else {
				state, err = flow.expression(loop.Initializer, state)
			}
			if err != nil {
				return 0, err
			}
		}
	case ast.KindWhileStatement:
		body, condition = node.AsWhileStatement().Statement, node.AsWhileStatement().Expression
	case ast.KindDoStatement:
		body, condition = node.AsDoStatement().Statement, node.AsDoStatement().Expression
	case ast.KindForOfStatement:
		loop := node.AsForInOrOfStatement()
		body = loop.Statement
		if readsError := flow.reads(loop.Expression, state); readsError != nil {
			return 0, readsError
		}
	}
	entry := state | superInitialized
	if state == 0 {
		entry = 0
	}
	if node.Kind != ast.KindDoStatement {
		if readsError := flow.reads(condition, entry); readsError != nil {
			return 0, readsError
		}
	}
	flow.breaks = append(flow.breaks, 0)
	flow.continues = append(flow.continues, 0)
	normal, statementError := flow.statement(body, entry)
	broken := flow.breaks[len(flow.breaks)-1]
	continued := flow.continues[len(flow.continues)-1]
	flow.breaks = flow.breaks[:len(flow.breaks)-1]
	flow.continues = flow.continues[:len(flow.continues)-1]
	if statementError != nil {
		return 0, statementError
	}
	normal |= continued
	if increment != nil {
		var expressionError error
		normal, expressionError = flow.expression(increment, normal)
		if expressionError != nil {
			return 0, expressionError
		}
	}
	if node.Kind == ast.KindDoStatement {
		if readsError := flow.reads(condition, normal); readsError != nil {
			return 0, readsError
		}
	}
	result := normal | broken
	if node.Kind != ast.KindDoStatement {
		result |= state
	}
	return result, nil
}

func (flow *constructorSuperFlow) switchCases(node *ast.Node, state uint8) (uint8, error) {
	switched := node.AsSwitchStatement()
	if readsError := flow.reads(switched.Expression, state); readsError != nil {
		return 0, readsError
	}
	flow.breaks = append(flow.breaks, 0)
	fallback, fallthroughState := false, uint8(0)
	clauses := switched.CaseBlock.AsCaseBlock().Clauses.Nodes
	for _, clause := range clauses {
		branch := clause.AsCaseOrDefaultClause()
		if clause.Kind == ast.KindDefaultClause {
			fallback = true
		}
		if readsError := flow.reads(branch.Expression, state); readsError != nil {
			flow.breaks = flow.breaks[:len(flow.breaks)-1]
			return 0, readsError
		}
		normal, listError := flow.list(branch.Statements.Nodes, state|fallthroughState)
		if listError != nil {
			flow.breaks = flow.breaks[:len(flow.breaks)-1]
			return 0, listError
		}
		fallthroughState = normal
	}
	result := fallthroughState | flow.breaks[len(flow.breaks)-1]
	flow.breaks = flow.breaks[:len(flow.breaks)-1]
	if !fallback {
		result |= state
	}
	return result, nil
}
