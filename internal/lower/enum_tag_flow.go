package lower

import (
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A value comparison proves a layout only when no member the incoming type keeps could hold that
// value with a different layout (#drwm5pj, ruled Oct 20). So a refinement of a union whose tags can
// lie is judged where it happens: at each step that narrowed the read, on every path into it,
// against the type flowing into that step, never against the variable's declared union.

type enumTagHazard int

const (
	enumTagNoHazard enumTagHazard = iota
	enumTagOpenHazard
	enumTagSharedValueHazard
)

// enumTagHazardOf names how a refinement of union could rest on a tag that lies: a field an open
// numeric enum types that isn't one type in every member, or same-named tags from two enums that
// share a value. With field set, only that field is asked about.
func (l *lowering) enumTagHazardOf(union *checker.Type, field string) enumTagHazard {
	if union.Flags()&checker.TypeFlagsUnion == 0 {
		return enumTagNoHazard
	}
	objects := 0
	for _, member := range union.Types() {
		if member.Flags()&checker.TypeFlagsPrimitive == 0 {
			objects++
		}
	}
	if objects < 2 {
		return enumTagNoHazard
	}
	for _, member := range union.Types() {
		for _, property := range l.checker.GetPropertiesOfType(member) {
			if field != "" && property.Name != field {
				continue
			}
			tag := l.checker.GetTypeOfSymbol(property)
			if l.openNumericEnumType(tag) && !sharedByEveryMember(l.checker, union, property.Name, tag) {
				return enumTagOpenHazard
			}
		}
	}
	if l.enumTagsShareValue(union, field) {
		return enumTagSharedValueHazard
	}
	return enumTagNoHazard
}

// The checker's own reference prefix test: an assignment to v resets every narrowing of v.inner.
//
//go:linkname enumTagContainsReference github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).containsMatchingReference
func enumTagContainsReference(receiver *checker.Checker, source, target *ast.Node) bool

// enumTagRefinementProven reports whether every step that narrowed reference, on every path into
// it, is either a recognized comparison on one of its fields read directly that passes against
// the type flowing into it, or a step that changes nothing but null and undefined there, and at
// least one such step explains the narrowing. Anything else, a flow node the walk doesn't know
// among them, leaves the refinement unproven.
func (l *lowering) enumTagRefinementProven(reference *ast.Node, declared *checker.Type) bool {
	data := reference.FlowNodeData()
	if data == nil || data.FlowNode == nil {
		return false
	}
	walk := &enumTagWalk{l: l, reference: reference, declared: declared, seen: map[*ast.FlowNode]bool{}}
	return walk.visit(data.FlowNode) && walk.proved > 0
}

type enumTagWalk struct {
	l         *lowering
	reference *ast.Node
	declared  *checker.Type
	seen      map[*ast.FlowNode]bool
	proved    int
}

func (w *enumTagWalk) visit(flow *ast.FlowNode) bool {
	for flow != nil {
		if w.seen[flow] {
			return true
		}
		w.seen[flow] = true
		switch {
		case flow.Flags&ast.FlowFlagsUnreachable != 0:
			return true
		case flow.Flags&ast.FlowFlagsStart != 0:
			// As the checker does, only a plain name, or this inside an arrow, carries narrowing in
			// from the function around it.
			if flow.Node != nil && (w.reference.Kind == ast.KindIdentifier || (w.reference.Kind == ast.KindThisKeyword && ast.IsArrowFunction(flow.Node))) {
				if outer := flow.Node.FlowNodeData(); outer != nil && outer.FlowNode != nil {
					flow = outer.FlowNode
					continue
				}
			}
			return true
		case flow.Flags&ast.FlowFlagsAssignment != 0:
			if predicateMatchingReference(w.l.checker, w.reference, flow.Node) {
				// An assignment narrows to what was assigned, a step the rule doesn't recognize.
				return false
			}
			if enumTagContainsReference(w.l.checker, w.reference, flow.Node) {
				return true
			}
			flow = flow.Antecedent
		case flow.Flags&ast.FlowFlagsCondition != 0:
			if !w.step(flow, w.directField(flow.Node)) {
				return false
			}
			flow = flow.Antecedent
		case flow.Flags&ast.FlowFlagsSwitchClause != 0:
			data := flow.Node.AsFlowSwitchClauseData()
			if !w.step(flow, w.directField(data.SwitchStatement.AsSwitchStatement().Expression)) {
				return false
			}
			flow = flow.Antecedent
		case flow.Flags&(ast.FlowFlagsCall|ast.FlowFlagsArrayMutation) != 0:
			if !w.step(flow, "") {
				return false
			}
			flow = flow.Antecedent
		case flow.Flags&ast.FlowFlagsReduceLabel != 0:
			// Its target's antecedents are walked whole where the label is reached.
			flow = flow.Antecedent
		case flow.Flags&ast.FlowFlagsLabel != 0:
			for list := flow.Antecedents; list != nil; list = list.Next {
				if !w.visit(list.Flow) {
					return false
				}
			}
			return true
		default:
			return false
		}
	}
	return true
}

// step judges one narrowing step. A recognized comparison on field passes when that field can't
// lie in the type flowing into it. Any other step passes only when it changes nothing at that
// point but null and undefined, which a pointer test proves; a step judged at its own point, not by
// what later steps would leave, can't be hidden behind a redundant comparison after it.
func (w *enumTagWalk) step(flow *ast.FlowNode, field string) bool {
	before := predicateFlowType(w.l.checker, w.reference, w.declared, w.declared, nil, flow.Antecedent)
	if field != "" {
		if w.l.enumTagHazardOf(before, field) != enumTagNoHazard {
			return false
		}
		w.proved++
		return true
	}
	after := predicateFlowType(w.l.checker, w.reference, w.declared, w.declared, nil, flow)
	if checker.Checker_isTypeIdenticalTo(w.l.checker, after, before) {
		return true
	}
	if checker.Checker_isTypeIdenticalTo(w.l.checker, after, w.l.checker.GetNonNullableType(before)) {
		w.proved++
		return true
	}
	return false
}

// directField is the field a recognized comparison reads off the reference: `r.f === x`, `!==`,
// `==`, `!=` (either side), `switch (r.f)`, or the truthiness of `r.f`, under parentheses and `!`.
// An optional chain, an element access, and an accessor are not a field read directly.
func (w *enumTagWalk) directField(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	for node.Kind == ast.KindPrefixUnaryExpression && node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		node = ast.SkipParentheses(node.AsPrefixUnaryExpression().Operand)
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken, ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
			if field := w.fieldRead(binary.Left); field != "" {
				return field
			}
			return w.fieldRead(binary.Right)
		}
		return ""
	}
	return w.fieldRead(node)
}

func (w *enumTagWalk) fieldRead(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression || node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return ""
	}
	access := node.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil || !predicateMatchingReference(w.l.checker, w.reference, access.Expression) {
		return ""
	}
	name := access.Name().Text()
	// On a union the property read is the checker's union of each member's own, which doesn't carry
	// an accessor's flags, so each member's own is asked: a getter is a call, never a field read.
	owner := w.l.checker.GetTypeAtLocation(access.Expression)
	members := []*checker.Type{owner}
	if owner.Flags()&checker.TypeFlagsUnion != 0 {
		members = owner.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		property := w.l.checker.GetPropertyOfType(member, name)
		if property == nil || property.Flags&(ast.SymbolFlagsGetAccessor|ast.SymbolFlagsSetAccessor|ast.SymbolFlagsMethod) != 0 {
			return ""
		}
	}
	return name
}
