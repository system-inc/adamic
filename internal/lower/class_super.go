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
	l         *lowering
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
		if (child.Kind == ast.KindThisKeyword || (child.Kind == ast.KindSuperKeyword && child.Parent != nil && child.Parent.Kind == ast.KindPropertyAccessExpression)) && state&superUninitialized != 0 {
			flow.instance.unreadyThis[child] = true
		}
		if child.Kind == ast.KindCallExpression && ast.SkipParentheses(child.AsCallExpression().Expression).Kind == ast.KindSuperKeyword {
			failed = flow.l.notYet(child, "super used as a value or in a deferred function; call it as a statement in each branch")
			return true
		}
		return child.ForEachChild(visit)
	}
	visit(node)
	return failed
}

func (flow *constructorSuperFlow) list(nodes []*ast.Node, state uint8) (uint8, error) {
	for _, node := range nodes {
		next, err := flow.statement(node, state)
		if err != nil {
			return 0, err
		}
		state = next
	}
	return state, nil
}

func (flow *constructorSuperFlow) expression(node *ast.Node, state uint8) (uint8, error) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression && ast.SkipParentheses(node.AsCallExpression().Expression).Kind == ast.KindSuperKeyword {
		for _, argument := range nodesOf(node.AsCallExpression().Arguments) {
			if err := flow.reads(argument, state); err != nil {
				return 0, err
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
		if err := flow.reads(expression.Condition, state); err != nil {
			return 0, err
		}
		left, err := flow.expression(expression.WhenTrue, state)
		if err != nil {
			return 0, err
		}
		right, err := flow.expression(expression.WhenFalse, state)
		return left | right, err
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
		if err := flow.reads(branch.Expression, state); err != nil {
			return 0, err
		}
		left, err := flow.statement(branch.ThenStatement, state)
		if err != nil {
			return 0, err
		}
		right, err := flow.statement(branch.ElseStatement, state)
		return left | right, err
	case ast.KindThrowStatement:
		return 0, flow.reads(node.AsThrowStatement().Expression, state)
	case ast.KindReturnStatement:
		return 0, flow.l.notYet(node, "an explicit return from a derived constructor")
	case ast.KindTryStatement:
		tried := node.AsTryStatement()
		normal, err := flow.statement(tried.TryBlock, state)
		if err != nil {
			return 0, err
		}
		if tried.CatchClause != nil {
			// A failure may occur on either side of super returning. The runtime flag
			// distinguishes a retry from a second successful initialization.
			caught, err := flow.statement(tried.CatchClause.AsCatchClause().Block, state|superInitialized)
			if err != nil {
				return 0, err
			}
			normal |= caught
		}
		if tried.FinallyBlock != nil {
			// Finally also runs on exceptional exits, before this might be initialized.
			if _, err := flow.statement(tried.FinallyBlock, state|normal); err != nil {
				return 0, err
			}
			normal, err = flow.statement(tried.FinallyBlock, normal)
		}
		return normal, err
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

func (l *lowering) prepareSuper(constructor *ast.Node, initializer int) (uint8, error) {
	instance := l.instance
	instance.unreadyThis = map[*ast.Node]bool{}
	instance.superStates = map[*ast.Node]uint8{}
	flow := constructorSuperFlow{l: l, instance: instance}
	for _, parameter := range constructor.Parameters() {
		if err := flow.reads(parameter, superUninitialized); err != nil {
			return 0, err
		}
	}
	state, err := flow.statement(constructor.Body(), superUninitialized)
	if err != nil {
		return 0, err
	}
	instance.superReady = len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "superReady", Type: ir.Boolean, Function: initializer})
	l.result.Functions[initializer].Body = append(l.result.Functions[initializer].Body, ir.Declare{Local: instance.superReady, Value: ir.BooleanConstant{}})
	return state, nil
}

func (l *lowering) superError(message string) ir.Statement {
	return ir.Throw{Value: l.generatedError("ReferenceError", message)}
}

func (l *lowering) superEnd() ir.Statement {
	return ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: l.instance.superReady, Of: ir.Boolean}}, Then: []ir.Statement{l.superError("Must call super constructor in derived class before accessing 'this' or returning from derived constructor")}}
}

// Conditional super used for effects has the same branch and field-initializer order as if.
func (l *lowering) conditionalSuper(node *ast.Node) ([]ir.Statement, bool, error) {
	if l.instance == nil || l.instance.base == nil || node.Kind != ast.KindConditionalExpression {
		return nil, false, nil
	}
	conditional := node.AsConditionalExpression()
	hasSuper := false
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if ast.IsFunctionLike(child) {
			return false
		}
		if child.Kind == ast.KindCallExpression && ast.SkipParentheses(child.AsCallExpression().Expression).Kind == ast.KindSuperKeyword {
			hasSuper = true
			return true
		}
		return child.ForEachChild(visit)
	}
	visit(node)
	if !hasSuper {
		return nil, false, nil
	}
	condition, err := l.condition(conditional.Condition)
	if err != nil {
		return nil, true, err
	}
	yes, err := l.expressionStatement(conditional.WhenTrue)
	if err != nil {
		return nil, true, err
	}
	no, err := l.expressionStatement(conditional.WhenFalse)
	if err != nil {
		return nil, true, err
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
		if err := flow.reads(loop.Expression, state); err != nil {
			return 0, err
		}
	}
	entry := state | superInitialized
	if state == 0 {
		entry = 0
	}
	if node.Kind != ast.KindDoStatement {
		if err := flow.reads(condition, entry); err != nil {
			return 0, err
		}
	}
	flow.breaks = append(flow.breaks, 0)
	flow.continues = append(flow.continues, 0)
	normal, err := flow.statement(body, entry)
	broken := flow.breaks[len(flow.breaks)-1]
	continued := flow.continues[len(flow.continues)-1]
	flow.breaks = flow.breaks[:len(flow.breaks)-1]
	flow.continues = flow.continues[:len(flow.continues)-1]
	if err != nil {
		return 0, err
	}
	normal |= continued
	if increment != nil {
		normal, err = flow.expression(increment, normal)
		if err != nil {
			return 0, err
		}
	}
	if node.Kind == ast.KindDoStatement {
		if err := flow.reads(condition, normal); err != nil {
			return 0, err
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
	if err := flow.reads(switched.Expression, state); err != nil {
		return 0, err
	}
	flow.breaks = append(flow.breaks, 0)
	fallback, fallthroughState := false, uint8(0)
	clauses := switched.CaseBlock.AsCaseBlock().Clauses.Nodes
	for _, clause := range clauses {
		branch := clause.AsCaseOrDefaultClause()
		if clause.Kind == ast.KindDefaultClause {
			fallback = true
		}
		if err := flow.reads(branch.Expression, state); err != nil {
			flow.breaks = flow.breaks[:len(flow.breaks)-1]
			return 0, err
		}
		normal, err := flow.list(branch.Statements.Nodes, state|fallthroughState)
		if err != nil {
			flow.breaks = flow.breaks[:len(flow.breaks)-1]
			return 0, err
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
