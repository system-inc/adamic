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
				// An assignment that leaves the declared type, like const node: Node = make(), narrows
				// nothing, and nothing before it reaches the read. One that narrows to what was assigned
				// isn't a recognized step.
				assigned := predicateFlowType(w.l.checker, w.reference, w.declared, w.declared, nil, flow)
				return checker.Checker_isTypeIdenticalTo(w.l.checker, assigned, w.declared)
			}
			if enumTagContainsReference(w.l.checker, w.reference, flow.Node) {
				return true
			}
			flow = flow.Antecedent
		case flow.Flags&ast.FlowFlagsCondition != 0:
			if field := w.directField(flow.Node); field == "" {
				if call, exact := w.exactKindCall(flow.Node); exact {
					if !w.kindPredicateStep(flow, call) {
						return false
					}
					flow = flow.Antecedent
					continue
				}
			}
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

// kindCall is a direct call to an exact kind predicate whose argument is the reference: the
// comparison `reference.field === member` its body makes, and whether a `!` turns it around.
type kindCall struct {
	field   string
	member  *checker.Type
	negated bool
}

// exactKindCall recognizes a condition that calls an exact kind predicate on the reference, under
// parentheses and `!` (#00rd54t, Adamic's ruling, Nov 3, extending rulings.md:213). The predicate is
// exact by its body alone: a function declaration, or a const arrow, that the checker's symbol for
// the callee resolves to (so a shadowing local never qualifies), with one declaration (so not
// overloaded), whose declared return is `p is T` on its first parameter p after any `this`, and
// whose body is only `return p.f === E.M`, either order, E.M an enum member. An asserts predicate,
// an inferred one and a const with an annotation don't count. The callee is a bare name, so never a
// method, a namespace member or a value read off an object, and an import or re-export under another
// name resolves to the one declaration. The reference must be the argument at p, as the checker
// matches it. Whether the call proves anything is judged at the call (kindPredicateStep).
func (w *enumTagWalk) exactKindCall(node *ast.Node) (kindCall, bool) {
	negated := false
	unwrap := func() {
		node = ast.SkipParentheses(node)
		for node.Kind == ast.KindPrefixUnaryExpression && node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
			negated = !negated
			node = ast.SkipParentheses(node.AsPrefixUnaryExpression().Operand)
		}
	}
	unwrap()
	// `isX(n) === true`, `!== false` and their turns: the call compared with a boolean literal is the
	// call, kept or turned around. `==` and `!=` never reach here (rulings.md:172 refuses them).
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if operator != ast.KindEqualsEqualsEqualsToken && operator != ast.KindExclamationEqualsEqualsToken {
			return kindCall{}, false
		}
		left, right := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
		if left.Kind == ast.KindTrueKeyword || left.Kind == ast.KindFalseKeyword {
			left, right = right, left
		}
		if right.Kind != ast.KindTrueKeyword && right.Kind != ast.KindFalseKeyword {
			return kindCall{}, false
		}
		if (operator == ast.KindExclamationEqualsEqualsToken) != (right.Kind == ast.KindFalseKeyword) {
			negated = !negated
		}
		node = left
		unwrap()
	}
	if node.Kind != ast.KindCallExpression {
		return kindCall{}, false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	// A bare name only: `ts.isBlock(n)` through a namespace import is a property access and doesn't count.
	if callee.Kind != ast.KindIdentifier || call.QuestionDotToken != nil || node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return kindCall{}, false
	}
	symbol := w.l.symbol(callee)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return kindCall{}, false
	}
	function := symbol.Declarations[0]
	if function.Kind == ast.KindVariableDeclaration {
		variable := function.AsVariableDeclaration()
		if function.Parent == nil || function.Parent.Flags&ast.NodeFlagsConst == 0 || variable.Type != nil || variable.Initializer == nil {
			return kindCall{}, false
		}
		function = ast.SkipParentheses(variable.Initializer)
		if function.Kind != ast.KindArrowFunction {
			return kindCall{}, false
		}
	} else if function.Kind != ast.KindFunctionDeclaration {
		return kindCall{}, false
	}
	// The first parameter after a `this` declaration, and the argument at the same place.
	parameters := function.Parameters()
	if len(parameters) > 0 && ast.IsIdentifier(parameters[0].Name()) && parameters[0].Name().Text() == "this" {
		parameters = parameters[1:]
	}
	if len(parameters) == 0 || !ast.IsIdentifier(parameters[0].Name()) || parameters[0].AsParameterDeclaration().DotDotDotToken != nil || len(call.Arguments.Nodes) == 0 {
		return kindCall{}, false
	}
	argument := call.Arguments.Nodes[0]
	if argument.Kind == ast.KindSpreadElement || !predicateMatchingReference(w.l.checker, w.reference, argument) {
		return kindCall{}, false
	}
	parameter := parameters[0].Name()
	annotation := function.Type()
	if annotation == nil || annotation.Kind != ast.KindTypePredicate {
		return kindCall{}, false
	}
	predicate := annotation.AsTypePredicateNode()
	if predicate.AssertsModifier != nil || predicate.Type == nil || predicate.ParameterName.Kind != ast.KindIdentifier || predicate.ParameterName.Text() != parameter.Text() {
		return kindCall{}, false
	}
	body := function.Body()
	if body == nil {
		return kindCall{}, false
	}
	compared := body
	if body.Kind == ast.KindBlock {
		statements := body.AsBlock().Statements.Nodes
		if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement || statements[0].AsReturnStatement().Expression == nil {
			return kindCall{}, false
		}
		compared = statements[0].AsReturnStatement().Expression
	}
	compared = ast.SkipParentheses(compared)
	if compared.Kind != ast.KindBinaryExpression {
		return kindCall{}, false
	}
	binary := compared.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {
		return kindCall{}, false
	}
	parameterSymbol := w.l.checker.GetSymbolAtLocation(parameters[0].Name())
	read := func(side *ast.Node) string {
		side = ast.SkipParentheses(side)
		if side.Kind != ast.KindPropertyAccessExpression || side.Flags&ast.NodeFlagsOptionalChain != 0 {
			return ""
		}
		access := side.AsPropertyAccessExpression()
		object := ast.SkipParentheses(access.Expression)
		if access.QuestionDotToken != nil || object.Kind != ast.KindIdentifier || parameterSymbol == nil || w.l.checker.GetSymbolAtLocation(object) != parameterSymbol {
			return ""
		}
		return access.Name().Text()
	}
	member := func(side *ast.Node) *checker.Type {
		side = ast.SkipParentheses(side)
		constant := w.l.symbol(side)
		if constant == nil || constant.Flags&ast.SymbolFlagsEnumMember == 0 {
			return nil
		}
		return w.l.checker.GetTypeAtLocation(side)
	}
	field, constant := read(binary.Left), member(binary.Right)
	if field == "" || constant == nil {
		field, constant = read(binary.Right), member(binary.Left)
	}
	if field == "" || constant == nil {
		return kindCall{}, false
	}
	return kindCall{field: field, member: constant, negated: negated}, true
}

// kindPredicateStep judges an exact kind predicate's call as the comparison its body makes. It
// passes only when that comparison would: the field is read directly off every object member
// flowing in (no accessor), and it can't lie there (enumTagHazardOf, as step asks of a comparison).
// The checker narrows a predicate's call by the predicate's declared type, not by its body, so the
// step also passes only when that narrowing is identical to the comparison's on the side this path
// takes: a predicate whose type disagrees with its own body proves nothing here. R-T80's check at the
// call is kept; only the refusal of the read moves.
func (w *enumTagWalk) kindPredicateStep(flow *ast.FlowNode, call kindCall) bool {
	before := predicateFlowType(w.l.checker, w.reference, w.declared, w.declared, nil, flow.Antecedent)
	members := []*checker.Type{before}
	if before.Flags()&checker.TypeFlagsUnion != 0 {
		members = before.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		property := w.l.checker.GetPropertyOfType(member, call.field)
		if property == nil || property.Flags&(ast.SymbolFlagsGetAccessor|ast.SymbolFlagsSetAccessor|ast.SymbolFlagsMethod) != 0 {
			return false
		}
	}
	if w.l.enumTagHazardOf(before, call.field) != enumTagNoHazard {
		return false
	}
	taken := flow.Flags&ast.FlowFlagsTrueCondition != 0
	if call.negated {
		taken = !taken
	}
	kept := []*checker.Type{}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			// A comparison of a field reads no primitive, so the true side has none, and the false
			// side keeps it as the checker does.
			if !taken {
				kept = append(kept, member)
			}
			continue
		}
		tag := w.l.checker.GetTypeOfSymbol(w.l.checker.GetPropertyOfType(member, call.field))
		if taken {
			if checker.Checker_isTypeAssignableTo(w.l.checker, call.member, tag) {
				kept = append(kept, member)
			}
		} else if !checker.Checker_isTypeIdenticalTo(w.l.checker, tag, call.member) {
			kept = append(kept, member)
		}
	}
	if len(kept) == 0 {
		return false
	}
	after := predicateFlowType(w.l.checker, w.reference, w.declared, w.declared, nil, flow)
	if !checker.Checker_isTypeIdenticalTo(w.l.checker, after, w.l.checker.GetUnionType(kept)) {
		return false
	}
	w.proved++
	return true
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
