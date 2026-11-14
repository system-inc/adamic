package lower

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A read the checker types never holds a value no member of its declared type can be, so it runs only
// when a value lied about its type: rulings.md:74 makes it a loud trap, as enum_never.go already does for
// a numeric enum (#p6vawef). But the checker also keeps a narrowing across a call that writes the
// narrowed place (#r0vzaaa's family), and a never made that way is false on a valid run:
//
//	if (holder.node.kind === K.A) return 'a';   // removes NodeA
//	flip(holder);                               // puts a NodeA back
//	if (holder.node.kind === K.B) return 'b';   // removes NodeB: never
//	return stray(holder.node);                  // Node prints stray
//
// A never is made of several narrowings, each removing members, and a call between any two can undo
// the earlier one, so a never read is trusted only when nothing that can write the place runs anywhere
// between the earliest narrowing it's made of and the read. That one is found by walking the read's
// flow back until the declared type holds (neverMakingAlong). A trusted read traps, naming itself and
// the narrowing that made it never. Any other is rulings.md:151's case, a narrowing a call broke, which
// stops until the read through the stored layout is built (#54vt253): code between that may write the
// place, a walk that meets a loop, a join whose ways in disagree, or a narrowing it can't place. A
// binding only its own function writes is trusted however its walk ends, since nothing between can
// write it unseen.

// neverStep is one narrowing a never read is made of: a check, a switch on the reference, an assertion
// call or an assignment, written at node. from is where what runs after it starts, and observed is an
// assignment's own target, which writes the place but is the narrowing itself.
type neverStep struct {
	node     *ast.Node
	from     int
	observed *ast.Node
	kind     string
}

// neverMaking is what a never read is made of: the earliest narrowing, where what may write the place is
// judged from, and the one that made it never, which the trap names. placeholder is a never assigned
// whole, `let flag: boolean = null!`, which no narrowing made, so the placeholder rule keeps it. One with
// no steps is a binding only its own function writes whose walk found none (neverRead), trusted without.
type neverMaking struct {
	earliest    neverStep
	madeNever   neverStep
	placeholder bool
}

// neverRead judges a read the checker types never: an identifier, this, a property or an element
// access, not in a type, not a declaration's name (a shorthand's is a read), not an assignment's target
// and not a direct callee. It answers the making the caller traps when the read is trusted, a not yet
// when it isn't, and neither when the read isn't one, the enum form traps it already, the declared type
// is never itself, the flow can't reach the read, or a never value was assigned whole.
func (l *lowering) neverRead(node *ast.Node) (*neverMaking, error) {
	node = ast.SkipParentheses(node)
	if !l.neverReadPosition(node) || l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNever == 0 || l.enumNeverIdentity(node, map[*ast.Node]bool{}) != nil {
		return nil, nil
	}
	stop := l.notYet(node, "a read the checker narrowed to never across code that may write it")
	data := node.FlowNodeData()
	if data == nil || data.FlowNode == nil {
		return nil, stop
	}
	if data.FlowNode.Flags&ast.FlowFlagsUnreachable != 0 {
		return nil, nil // Proven unreachable, so nothing runs it.
	}
	if node.Kind == ast.KindThisKeyword {
		return nil, stop // No symbol gives this a declared type to walk back to.
	}
	declared := l.enumStoredType(node)
	if declared == nil {
		return nil, stop
	}
	if declared.Flags()&checker.TypeFlagsNever != 0 {
		return nil, nil // Declared never: no narrowing made it, and what passes it there lied first.
	}
	// A binding only its own function writes can't change behind anything between: each of its writes is
	// in the flow the checker read, so its never is the checker's proof wherever the walk ends, as
	// narrowedNeverReadGap reads it. A const an aliased condition narrows (const flag = big === 'x', then
	// flag's false side) is one the walk can't place. The walk still names the narrowing when it finds one.
	own := false
	if node.Kind == ast.KindIdentifier {
		symbol := l.bindingNamed(node)
		own = symbol != nil && !l.writtenAcrossFunctions(symbol)
	}
	var making *neverMaking
	if predicateFlowType(l.checker, node, declared, declared, nil, data.FlowNode).Flags()&checker.TypeFlagsNever != 0 {
		making = l.neverMakingAlong(node, declared, data.FlowNode, 0, map[*ast.FlowNode]*neverMaking{})
	} // Otherwise the walk's own reading of the flow disagrees with the checker's, so it can't be trusted.
	switch {
	case making == nil && own:
		return &neverMaking{}, nil
	case making == nil:
		return nil, stop
	case making.placeholder:
		return nil, nil
	case l.neverPlaceWrittenBetween(node, making.earliest.from, making.earliest.observed):
		return nil, stop
	}
	return making, nil
}

// neverReadPosition is whether node is a reference read: not a type, a declaration's name but a
// shorthand's, the name after a dot, an assignment's target or a direct callee.
func (l *lowering) neverReadPosition(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
	default:
		return false
	}
	shorthand := node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment
	if !shorthand && (!l.isExpression(node) || ast.IsDeclarationName(node)) || ast.IsRightSideOfPropertyAccess(node) || ast.IsAssignmentTarget(node) {
		return false
	}
	callee := node
	for callee.Parent != nil && callee.Parent.Kind == ast.KindParenthesizedExpression {
		callee = callee.Parent
	}
	return callee.Parent == nil || callee.Parent.Kind != ast.KindCallExpression || callee.Parent.AsCallExpression().Expression != callee
}

// neverMakingAlong walks the read's flow back from flow to where its declared type holds, as the
// checker computes it at each point (predicateFlowType). What the walk passed last before that point
// narrowed it first: the earliest step. The step past which the type stops being never is the one
// that made it never. An assignment of the reference ends the walk as the earliest step itself, since
// the type before it doesn't reach the read. A label with one way in is walked through, as the checker
// does; a join is walked through only when every way in reaches the same earliest step, as
// nullishObservationOf walks one; a loop, a function's start and anything else end the walk with
// nothing, and so does an earliest step that doesn't mention the reference, an aliased condition,
// whose test ran where the alias was made.
func (l *lowering) neverMakingAlong(read *ast.Node, declared *checker.Type, flow *ast.FlowNode, joins int, seen map[*ast.FlowNode]*neverMaking) *neverMaking {
	var previous *ast.FlowNode
	var madeNever *neverStep
	asked := false // Whether the step past which the type stops being never was asked for, placed or not.
	for flow != nil {
		after := predicateFlowType(l.checker, read, declared, declared, nil, flow)
		if l.checker.IsTypeAssignableTo(declared, after) {
			earliest := l.neverStepAt(read, previous)
			if earliest == nil {
				return nil
			}
			if madeNever == nil {
				madeNever = earliest
			}
			return &neverMaking{earliest: *earliest, madeNever: *madeNever}
		}
		if !asked && previous != nil && after.Flags()&checker.TypeFlagsNever == 0 {
			asked = true
			madeNever = l.neverStepAt(read, previous)
		}
		switch {
		case flow.Flags&ast.FlowFlagsAssignment != 0:
			if predicateMatchingReference(l.checker, read, flow.Node) {
				step, value := l.neverAssignmentStep(flow.Node)
				if step == nil {
					return nil
				}
				if l.checker.GetTypeAtLocation(value).Flags()&checker.TypeFlagsNever != 0 {
					return &neverMaking{placeholder: true}
				}
				if madeNever == nil {
					madeNever = step
				}
				return &neverMaking{earliest: *step, madeNever: *madeNever}
			}
			if enumTagContainsReference(l.checker, read, flow.Node) {
				return nil // A write of what holds the reference resets it to the declared type, which didn't hold.
			}
		case flow.Flags&(ast.FlowFlagsCondition|ast.FlowFlagsSwitchClause|ast.FlowFlagsCall|ast.FlowFlagsArrayMutation) != 0:
		case flow.Flags&ast.FlowFlagsLabel != 0 && flow.Antecedents != nil && flow.Antecedents.Next == nil:
			flow = flow.Antecedents.Flow
			continue
		case flow.Flags&ast.FlowFlagsBranchLabel != 0:
			joined := l.neverMakingJoined(read, declared, flow, joins, seen)
			if joined == nil {
				return nil
			}
			making := *joined
			if madeNever != nil {
				making.madeNever = *madeNever
			}
			return &making
		default:
			return nil
		}
		previous = flow
		flow = flow.Antecedent
	}
	return nil
}

// neverMakingJoined is the making every way into a join reaches, the same earliest step on each, or nil.
func (l *lowering) neverMakingJoined(read *ast.Node, declared *checker.Type, join *ast.FlowNode, joins int, seen map[*ast.FlowNode]*neverMaking) *neverMaking {
	if found, walked := seen[join]; walked {
		return found
	}
	seen[join] = nil // A cycle can't arise without a loop label, but if one did it ends here.
	if joins >= nullishObservationDepth {
		return nil
	}
	var common *neverMaking
	for list := join.Antecedents; list != nil; list = list.Next {
		if list.Flow.Flags&ast.FlowFlagsUnreachable != 0 {
			continue
		}
		found := l.neverMakingAlong(read, declared, list.Flow, joins+1, seen)
		if found == nil || common != nil && (found.placeholder != common.placeholder || found.earliest.node != common.earliest.node) {
			return nil
		}
		if common == nil {
			common = found
		}
	}
	seen[join] = common
	return common
}

// neverStepAt is the narrowing a flow node writes, when it mentions the reference: a check, from where
// the condition starts, since what it runs after reading the reference runs after the narrowing too; a
// switch on it, from its expression; or an assertion call. Nil for anything else.
func (l *lowering) neverStepAt(read *ast.Node, flow *ast.FlowNode) *neverStep {
	if flow == nil {
		return nil
	}
	var step *neverStep
	switch {
	case flow.Flags&ast.FlowFlagsCondition != 0:
		step = &neverStep{node: flow.Node, kind: "the check"}
	case flow.Flags&ast.FlowFlagsSwitchClause != 0:
		step = &neverStep{node: flow.Node.AsFlowSwitchClauseData().SwitchStatement.AsSwitchStatement().Expression, kind: "the switch on"}
	case flow.Flags&ast.FlowFlagsCall != 0:
		step = &neverStep{node: flow.Node, kind: "the call"}
	default:
		return nil
	}
	if !l.mentionsReference(read, step.node) {
		return nil
	}
	step.from = step.node.Pos()
	return step
}

// neverAssignmentStep is an assignment of the reference as a step, with the value it assigns: a plain =,
// or a declaration's initializer. Nil for any other form, which the walk can't place.
func (l *lowering) neverAssignmentStep(target *ast.Node) (*neverStep, *ast.Node) {
	if target.Kind == ast.KindVariableDeclaration {
		if initializer := target.AsVariableDeclaration().Initializer; initializer != nil {
			return &neverStep{node: target, from: target.End(), observed: target.Name(), kind: "the assignment"}, initializer
		}
		return nil, nil
	}
	if parent := target.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && parent.AsBinaryExpression().Left == target {
		return &neverStep{node: parent, from: parent.End(), observed: target, kind: "the assignment"}, parent.AsBinaryExpression().Right
	}
	return nil, nil
}

// neverPlaceWrittenBetween is whether something that can write the read's place, or what holds it, may
// run between from and the read, by #r0vzaaa's tests (narrowed_nullish.go). A binding only its own
// function writes can't change behind a call, and a write there is in the flow the walk read; one
// another function writes may change behind any code between (codeBetween). A property's name is judged
// by propertyWriteMayRunBetween, after a getter of it, or one the program could install, which answers
// each read itself. Each link of the path is judged the same way: holder in holder.node may be rebound
// by a closure the code between calls. A key that isn't a literal, or a link that isn't a reference,
// can't be judged.
func (l *lowering) neverPlaceWrittenBetween(read *ast.Node, from int, observed *ast.Node) bool {
	for reference := read; ; {
		reference = ast.SkipParentheses(reference)
		var name string
		numeric := false
		switch reference.Kind {
		case ast.KindIdentifier:
			symbol := l.bindingNamed(reference)
			return symbol == nil || l.writtenAcrossFunctions(symbol) && l.codeBetween(from, read, nil)
		case ast.KindThisKeyword:
			return false
		case ast.KindPropertyAccessExpression:
			name = reference.Name().Text()
		case ast.KindElementAccessExpression:
			key := ast.SkipParentheses(reference.AsElementAccessExpression().ArgumentExpression)
			if key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNoSubstitutionTemplateLiteral && key.Kind != ast.KindNumericLiteral {
				return true
			}
			name = key.Text()
			_, notNumeric := strconv.ParseFloat(name, 64)
			numeric = notNumeric == nil
		default:
			return true
		}
		if l.propertyWritesOf().accessors || l.accessorName(name) || l.propertyWriteMayRunBetween(read, name, numeric, from, observed, nil) {
			return true
		}
		reference = reference.Expression()
	}
}

// neverReadEvaluated is what a trusted never read evaluates before the trap: a binding's read, at its
// declared slot, or a property or an element read's receiver and key, since a field narrowed to never has
// no representation of its own to read at (typeOf's not yet), and what it would read is never used. A
// getter of the name stops the read before it's trusted (neverPlaceWrittenBetween), so leaving the read
// itself out runs nothing less of the program's.
func (l *lowering) neverReadEvaluated(node *ast.Node) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return l.discardedValue(node.Expression())
	case ast.KindElementAccessExpression:
		receiver, err := l.discardedValue(node.Expression())
		if err != nil {
			return nil, err
		}
		key, err := l.discardedValue(node.AsElementAccessExpression().ArgumentExpression)
		if err != nil {
			return nil, err
		}
		return ir.Comma{Left: receiver, Right: key}, nil
	}
	return l.enumNeverValue(node)
}

// neverReadTrap is the trusted never read made loud: what the read evaluates, then a panic naming it and
// the narrowing that made it never, shared by both backends, as enumNeverCheck's is. It gives back what
// the read's slot holds, so no conversion follows a value that's never made: a binding's own slot where
// nothing around it has one, a field's declared one, or a number's.
func (l *lowering) neverReadTrap(node *ast.Node, evaluated ir.Expression, making *neverMaking) ir.Expression {
	node = ast.SkipParentheses(node)
	otherwise := ir.Number
	if node.Kind == ast.KindIdentifier || node.Kind == ast.KindThisKeyword {
		otherwise = evaluated.Type()
	} else if declared := l.enumStoredType(node); declared != nil {
		if representation, known := l.representation(declared); known && !slotless(representation) {
			otherwise = representation
		}
	}
	of := l.neverCheckReturns(node, otherwise)
	message := "unreachable read of " + strings.Join(strings.Fields(sourceExpression(node)), " ") + " at " + l.program.Where(node)
	if making.madeNever.node != nil {
		message += ", never after " + making.madeNever.kind + " " + strings.Join(strings.Fields(sourceExpression(making.madeNever.node)), " ") + " at " + l.program.Where(making.madeNever.node)
	}
	function := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{
		Name: "never_read", Returns: of,
		Body: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}},
	})
	return ir.Comma{Left: evaluated, Right: ir.Call{Function: function, Returns: of}}
}
