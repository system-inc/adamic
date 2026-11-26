package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// narrowedArmPropertyReadGap is narrowedArmReadGap's rule for a property read, m.n, this.n, a.m.n or m['n'],
// as narrowedNullishPropertyReadGap is the nullish rule's (#8nj5pn4). The checker narrows a property by a
// discriminant check and keeps the narrowing across a call that writes the property again:
//
//	function reset(m: M): void { m.n = b; }
//	if (m.n.kind === 'a') { reset(m); return String(m.n.a + 1); }
//
// Node prints NaN, since m.n holds b by the read, and native read b's storage through a's layout and printed 1.
// A property has no declaration that keeps its writes in one function, so it's proved by name, as the nullish
// twin proves it: a getter of the name, or a program that could install one, stops the read first, since a
// getter answers each read itself. Then the read lowers when the place it was last seen narrowed (an
// assignment, or a check that mentions it, armObservationOf) is followed by nothing that can run program code
// before the read, or by nothing that can write its name (propertyWriteMayRunBetween). A read with no such
// place stops when a number spells its name, some write could reach any name, or any write of its name exists.
// It's a fail-closed stop: a checked view of the declared union is the read that lifts it.
func (l *lowering) narrowedArmPropertyReadGap(node *ast.Node) error {
	if !l.isExpression(node) || (ast.IsAssignmentTarget(node) && pureWrite(node)) {
		return nil
	}
	var key *ast.Node
	if node.Kind == ast.KindPropertyAccessExpression {
		key = node.Name()
	} else {
		key = ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNoSubstitutionTemplateLiteral && key.Kind != ast.KindNumericLiteral {
			return nil
		}
	}
	property := l.checker.GetSymbolAtLocation(key)
	if property == nil {
		return nil // No declared property, so no declared type the checker narrowed from.
	}
	narrowed, declared := l.checker.GetTypeAtLocation(node), l.checker.GetTypeOfSymbol(property)
	if narrowed.Id() == declared.Id() || narrowed.Flags()&checker.TypeFlagsNever != 0 {
		return nil
	}
	if _, nullish := l.nullishUnit(narrowed); nullish {
		return nil // narrowedNullishPropertyReadGap's.
	}
	present := l.checker.GetNonNullableType(declared)
	if present.Id() == narrowed.Id() {
		return nil // A presence strip, checked or observed at run time, as the binding rule says.
	}
	if storage, known := l.representation(declared); known && storage == ir.Union {
		return nil // The held union's member is checked at run time before it's narrowed.
	}
	presentStorage, presentKnown := l.representation(present)
	if narrowedStorage, narrowedKnown := l.representation(narrowed); presentKnown && narrowedKnown && narrowedStorage == presentStorage && !presentStorage.IsReference() && !l.armReadSpecializes(node) {
		return nil // The same scalar storage either way, as the binding rule says.
	}
	name := key.Text()
	stop := l.notYet(node, "a read of property "+name+" narrowed to "+l.checker.TypeToString(narrowed)+" where a call may have rewritten it")
	writes := l.propertyWritesOf()
	if writes.accessors || l.accessorName(name) {
		return stop
	}
	_, notNumeric := strconv.ParseFloat(name, 64)
	numeric := notNumeric == nil
	observation := l.armObservationOf(node)
	if observation == nil {
		if numeric || writes.everyName || len(writes.named[name]) > 0 {
			return stop
		}
		return nil
	}
	if l.propertyWriteMayRunBetween(node, name, numeric, observation.from, observation.target, l.appendsBetween(observation, node, numeric)) {
		return stop
	}
	return nil
}

// armObservationOf walks the read's flow straight back to the place it was last seen narrowed: a plain
// assignment of it, after which it holds what was assigned, or a condition that mentions it, the check that
// narrowed it. A call, which codeBetween judges, an assignment of something else and a join whose every way
// in reaches the same place are walked through. A loop, a label of any other kind, a write of what holds the
// read or the start of a function ends the walk with none. A condition's target is nil, since it writes
// nothing to exempt.
func (l *lowering) armObservationOf(read *ast.Node) *nullishObservation {
	data := read.FlowNodeData()
	if data == nil || data.FlowNode == nil {
		return nil
	}
	return l.armObservationAlong(read, data.FlowNode, 0, map[*ast.FlowNode]*nullishObservation{})
}

func (l *lowering) armObservationAlong(read *ast.Node, flow *ast.FlowNode, joins int, seen map[*ast.FlowNode]*nullishObservation) *nullishObservation {
	for flow != nil {
		switch {
		case flow.Flags&ast.FlowFlagsAssignment != 0:
			if predicateMatchingReference(l.checker, read, flow.Node) {
				parent := flow.Node.Parent
				if parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && parent.AsBinaryExpression().Left == flow.Node {
					return &nullishObservation{target: flow.Node, from: parent.End(), assigned: true}
				}
				return nil
			}
			if enumTagContainsReference(l.checker, read, flow.Node) || l.mentionsReference(read, flow.Node) {
				return nil
			}
		case flow.Flags&ast.FlowFlagsCondition != 0:
			if l.mentionsReference(read, flow.Node) {
				return &nullishObservation{from: flow.Node.End()}
			}
		case flow.Flags&(ast.FlowFlagsCall|ast.FlowFlagsArrayMutation) != 0:
		case flow.Flags&ast.FlowFlagsBranchLabel != 0:
			if found, walked := seen[flow]; walked {
				return found
			}
			seen[flow] = nil
			if joins >= nullishObservationDepth {
				return nil
			}
			var common *nullishObservation
			for list := flow.Antecedents; list != nil; list = list.Next {
				if list.Flow.Flags&ast.FlowFlagsUnreachable != 0 {
					continue
				}
				found := l.armObservationAlong(read, list.Flow, joins+1, seen)
				if found == nil || common != nil && found.from != common.from {
					return nil
				}
				common = found
			}
			seen[flow] = common
			return common
		default:
			return nil
		}
		flow = flow.Antecedent
	}
	return nil
}
