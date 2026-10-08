package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

type switchGroup struct {
	tests     []ir.Expression
	body      []ir.Statement
	isDefault bool
}

// switchBodyLeaves recognizes bodies that cannot reach a following group. Missing a form only
// selects the general lowering; it never removes a possible fallthrough edge.
func switchBodyLeaves(body []ir.Statement) bool {
	if len(body) == 0 {
		return false
	}
	switch statement := body[len(body)-1].(type) {
	case ir.Break, ir.Continue, ir.Return, ir.Throw, ir.Panic:
		return true
	case ir.Block:
		return switchBodyLeaves(statement.Body)
	case ir.If:
		return switchBodyLeaves(statement.Then) && switchBodyLeaves(statement.Else)
	case ir.Try:
		return (statement.HasFinally && switchBodyLeaves(statement.Finally)) ||
			(switchBodyLeaves(statement.Body) && (!statement.HasCatch || switchBodyLeaves(statement.Catch)))
	}
	return false
}

// fallthroughSwitch first chooses the entry group in source test order, with default only as a
// fallback. The entered body and every later body then run in one breakable switch, until an abrupt
// exit leaves it. Bodies are emitted once, so their locals, ownership and flow points are not cloned.
func (l *lowering) fallthroughSwitch(value ir.Expression, groups []switchGroup, unmatchedCheck ir.Statement) []ir.Statement {
	entry := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "switch_entry", Type: ir.Number, Function: l.functionIndex})
	dispatch := ir.Switch{Value: value}
	if unmatchedCheck != nil {
		// An open enum can reach a checker-never default only when no case matched.
		// Falling through from a matched case must still execute the default body.
		dispatch.Default = []ir.Statement{unmatchedCheck}
	}
	bodies := []ir.Statement{}
	for index, group := range groups {
		position := ir.NumberConstant{Value: float64(index)}
		choose := []ir.Statement{ir.Assign{Local: entry, Value: position}}
		if len(group.tests) > 0 {
			dispatch.Cases = append(dispatch.Cases, ir.Case{Tests: group.tests, Body: choose})
		}
		if group.isDefault {
			// Each branch owns its statement addresses, including instrumentation points.
			dispatch.Default = append(dispatch.Default, ir.Assign{Local: entry, Value: position})
		}
		if len(group.body) > 0 {
			bodies = append(bodies, ir.If{
				Condition: ir.Binary{Operator: ir.LessOrEqual, Left: ir.Read{Local: entry, Of: ir.Number}, Right: position},
				Then:      group.body,
			})
		}
	}
	return []ir.Statement{ir.Block{Body: []ir.Statement{
		ir.Declare{Local: entry, Value: ir.NumberConstant{Value: float64(len(groups))}},
		dispatch,
		ir.Switch{Value: ir.BooleanConstant{Value: true}, Default: bodies},
	}}}
}

// labeledBreak resolves the source label to a breakable depth. Dispatch switches introduced by
// fallthrough lowering finish before any source body, so they never change this depth.
func (l *lowering) labeledBreak(node *ast.Node) ([]ir.Statement, error) {
	depth := 0
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		iteration := ast.IsIterationStatement(parent, false)
		block := parent.Kind == ast.KindBlock && parent.Parent != nil && parent.Parent.Kind == ast.KindLabeledStatement
		if parent.Kind != ast.KindSwitchStatement && !iteration && !block {
			continue
		}
		if node.Label() == nil && !block {
			return []ir.Statement{ir.Break{Depth: depth}}, nil
		}
		for label := parent.Parent; label != nil && label.Kind == ast.KindLabeledStatement; label = label.Parent {
			if node.Label() != nil && label.Label().Text() == node.Label().Text() {
				return []ir.Statement{ir.Break{Depth: depth}}, nil
			}
		}
		depth++
	}
	return nil, l.notYet(node, "a break to a label that is not a loop, switch or block")
}

// Continue counts loops alone: a switch or a labeled block has no update target.
func (l *lowering) labeledContinue(node *ast.Node) ([]ir.Statement, error) {
	depth := 0
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if !ast.IsIterationStatement(parent, false) {
			continue
		}
		for label := parent.Parent; label != nil && label.Kind == ast.KindLabeledStatement; label = label.Parent {
			if label.Label().Text() == node.Label().Text() {
				return []ir.Statement{ir.Continue{Depth: depth}}, nil
			}
		}
		depth++
	}
	return nil, l.notYet(node, "a continue to a label that is not a loop")
}
