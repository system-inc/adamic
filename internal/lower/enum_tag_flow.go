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
func (lowering *lowering) enumTagHazardOf(union *checker.Type, field string) enumTagHazard {
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
		for _, property := range lowering.checker.GetPropertiesOfType(member) {
			if field != "" && property.Name != field {
				continue
			}
			tag := lowering.checker.GetTypeOfSymbol(property)
			if lowering.openNumericEnumType(tag) && !sharedByEveryMember(lowering.checker, union, property.Name, tag) {
				return enumTagOpenHazard
			}
		}
	}
	if lowering.enumTagsShareValue(union, field) {
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
//
// A reference that is a property access (`h.n`) is proven only under Lower's local rule (#qeanycw,
// Adamic's ruling on #00rd54t, Nov 4): no flow node marks an ordinary call, and a call can store
// another member in h.n while the checker's narrowing of it stands. So every step that narrows it
// sits in the test of a conditional whose branch holds the read, of an `&&` whose right holds it, or
// of an if whose block holds it, with no loop, label or function between, and nothing that can run
// program code completes from where that step starts to the read (accessLocalTests, codeBetween):
// code inside the step runs after it reads h.n too, as `h.n.kind === kindOf()` does. The exact
// predicate's own call is the one exception, since its body reads one plain field and compares it,
// and that field's name may have no accessor anywhere in the program.
func (lowering *lowering) enumTagRefinementProven(reference *ast.Node, declared *checker.Type) bool {
	data := reference.FlowNodeData()
	if data == nil || data.FlowNode == nil {
		return false
	}
	walk := &enumTagWalk{
		lowering:   lowering,
		reference:  reference,
		declared:   declared,
		seen:       map[*ast.FlowNode]bool{},
		firstProof: -1,
		exempt:     map[*ast.Node]bool{},
	}
	if reference.Kind == ast.KindPropertyAccessExpression {
		tests, local := lowering.accessLocalTests(reference)
		if !local {
			return false
		}
		walk.tests = tests
	}
	if !walk.visit(data.FlowNode) || walk.proved == 0 {
		return false
	}
	return walk.tests == nil || !lowering.codeBetween(walk.firstProof, reference, walk.exempt)
}

type enumTagWalk struct {
	lowering  *lowering
	reference *ast.Node
	declared  *checker.Type
	seen      map[*ast.FlowNode]bool
	proved    int
	// tests are the local tests an access reference's narrowing steps must sit in, firstProof is where
	// the earliest such step starts, and exempt holds the exact predicate calls among them; all are
	// unused for a name, which no call can change.
	tests      []*ast.Node
	firstProof int
	exempt     map[*ast.Node]bool
}

// proves counts a step that narrows the reference. For an access reference it must sit in one of
// its local tests (node nil, a switch clause's, never does).
func (walk *enumTagWalk) proves(node *ast.Node) bool {
	if walk.tests == nil {
		walk.proved++
		return true
	}
	if node == nil {
		return false
	}
	for _, testCase := range walk.tests {
		if node.Pos() >= testCase.Pos() && node.End() <= testCase.End() {
			walk.proved++
			if walk.firstProof < 0 || node.Pos() < walk.firstProof {
				walk.firstProof = node.Pos()
			}
			return true
		}
	}
	return false
}

func (walk *enumTagWalk) visit(flow *ast.FlowNode) bool {
	for flow != nil {
		if walk.seen[flow] {
			return true
		}
		walk.seen[flow] = true
		switch {
		case flow.Flags&ast.FlowFlagsUnreachable != 0:
			return true
		case flow.Flags&ast.FlowFlagsStart != 0:
			// As the checker does, only a plain name, or this inside an arrow, carries narrowing in
			// from the function around it.
			if flow.Node != nil && (walk.reference.Kind == ast.KindIdentifier ||
				(walk.reference.Kind == ast.KindThisKeyword && ast.IsArrowFunction(flow.Node))) {
				if outer := flow.Node.FlowNodeData(); outer != nil && outer.FlowNode != nil {
					flow = outer.FlowNode
					continue
				}
			}
			return true
		case flow.Flags&ast.FlowFlagsAssignment != 0:
			if predicateMatchingReference(walk.lowering.checker, walk.reference, flow.Node) {
				// An assignment that leaves the declared type, like const node: Node = make(), narrows
				// nothing, and nothing before it reaches the read. One that narrows to what was assigned
				// isn't a recognized step.
				assigned := predicateFlowType(
					walk.lowering.checker,
					walk.reference,
					walk.declared,
					walk.declared,
					nil,
					flow,
				)
				return checker.Checker_isTypeIdenticalTo(walk.lowering.checker, assigned, walk.declared)
			}
			if enumTagContainsReference(walk.lowering.checker, walk.reference, flow.Node) {
				return true
			}
			flow = flow.Antecedent
		case flow.Flags&ast.FlowFlagsCondition != 0:
			if field := walk.directField(flow.Node); field == "" {
				if call, exact := walk.exactKindCall(flow.Node); exact {
					if !walk.kindPredicateStep(flow, call) {
						return false
					}
					flow = flow.Antecedent
					continue
				}
			}
			if !walk.step(flow, walk.directField(flow.Node)) {
				return false
			}
			flow = flow.Antecedent
		case flow.Flags&ast.FlowFlagsSwitchClause != 0:
			data := flow.Node.AsFlowSwitchClauseData()
			if !walk.stepAt(flow, nil, walk.directField(data.SwitchStatement.AsSwitchStatement().Expression)) {
				return false
			}
			flow = flow.Antecedent
		case flow.Flags&(ast.FlowFlagsCall|ast.FlowFlagsArrayMutation) != 0:
			if !walk.step(flow, "") {
				return false
			}
			flow = flow.Antecedent
		case flow.Flags&ast.FlowFlagsReduceLabel != 0:
			// Its target's antecedents are walked whole where the label is reached.
			flow = flow.Antecedent
		case flow.Flags&ast.FlowFlagsLabel != 0:
			for list := flow.Antecedents; list != nil; list = list.Next {
				if !walk.visit(list.Flow) {
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
func (walk *enumTagWalk) step(flow *ast.FlowNode, field string) bool {
	return walk.stepAt(flow, flow.Node, field)
}

// stepAt is step with the node the step's narrowing is written at, which proves places for an
// access reference.
func (walk *enumTagWalk) stepAt(flow *ast.FlowNode, at *ast.Node, field string) bool {
	before := predicateFlowType(
		walk.lowering.checker,
		walk.reference,
		walk.declared,
		walk.declared,
		nil,
		flow.Antecedent,
	)
	if field != "" {
		if walk.lowering.enumTagHazardOf(before, field) != enumTagNoHazard {
			return false
		}
		return walk.proves(at)
	}
	after := predicateFlowType(walk.lowering.checker, walk.reference, walk.declared, walk.declared, nil, flow)
	if checker.Checker_isTypeIdenticalTo(walk.lowering.checker, after, before) {
		return true
	}
	if checker.Checker_isTypeIdenticalTo(
		walk.lowering.checker,
		after,
		walk.lowering.checker.GetNonNullableType(before),
	) {
		return walk.proves(at)
	}
	return false
}

// predicateProves counts an exact predicate's call as a narrowing step. For an access reference the
// call is exempt from codeBetween, but its body's field read runs inside it, so that field's name may
// have no getter anywhere in the program (accessorName).
func (walk *enumTagWalk) predicateProves(at *ast.Node, call kindCall) bool {
	if walk.tests != nil && walk.lowering.accessorName(call.field) {
		return false
	}
	if !walk.proves(at) {
		return false
	}
	if walk.tests != nil {
		walk.exempt[call.call] = true
	}
	return true
}

// kindCall is a direct call to an exact kind predicate whose argument is the reference: the
// comparison `reference.field === member` its body makes, and whether a `!` turns it around.
type kindCall struct {
	call    *ast.Node
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
func (walk *enumTagWalk) exactKindCall(node *ast.Node) (kindCall, bool) {
	negated := false
	unwrap := func() {
		node = ast.SkipParentheses(node)
		for node.Kind == ast.KindPrefixUnaryExpression &&
			node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
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
	symbol := walk.lowering.symbol(callee)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return kindCall{}, false
	}
	function := symbol.Declarations[0]
	if function.Kind == ast.KindVariableDeclaration {
		variable := function.AsVariableDeclaration()
		if function.Parent == nil ||
			function.Parent.Flags&ast.NodeFlagsConst == 0 ||
			variable.Type != nil ||
			variable.Initializer == nil {
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
	// A default, an optional mark or a rest would let the body test something other than the
	// argument: isA(undefined) tests the default's kind (Language's attack on 5d222aba).
	if len(parameters) == 0 || !ast.IsIdentifier(parameters[0].Name()) || len(call.Arguments.Nodes) == 0 {
		return kindCall{}, false
	}
	if declaration := parameters[0].AsParameterDeclaration(); declaration.DotDotDotToken != nil ||
		declaration.QuestionToken != nil ||
		declaration.Initializer != nil {
		return kindCall{}, false
	}
	argument := call.Arguments.Nodes[0]
	if argument.Kind == ast.KindSpreadElement ||
		!predicateMatchingReference(walk.lowering.checker, walk.reference, argument) {
		return kindCall{}, false
	}
	parameter := parameters[0].Name()
	annotation := function.Type()
	if annotation == nil || annotation.Kind != ast.KindTypePredicate {
		return kindCall{}, false
	}
	predicate := annotation.AsTypePredicateNode()
	if predicate.AssertsModifier != nil ||
		predicate.Type == nil ||
		predicate.ParameterName.Kind != ast.KindIdentifier ||
		predicate.ParameterName.Text() != parameter.Text() {
		return kindCall{}, false
	}
	body := function.Body()
	if body == nil {
		return kindCall{}, false
	}
	compared := body
	if body.Kind == ast.KindBlock {
		statements := body.AsBlock().Statements.Nodes
		if len(statements) != 1 ||
			statements[0].Kind != ast.KindReturnStatement ||
			statements[0].AsReturnStatement().Expression == nil {
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
	parameterSymbol := walk.lowering.checker.GetSymbolAtLocation(parameters[0].Name())
	read := func(side *ast.Node) string {
		side = ast.SkipParentheses(side)
		if side.Kind != ast.KindPropertyAccessExpression || side.Flags&ast.NodeFlagsOptionalChain != 0 {
			return ""
		}
		access := side.AsPropertyAccessExpression()
		object := ast.SkipParentheses(access.Expression)
		if access.QuestionDotToken != nil ||
			object.Kind != ast.KindIdentifier ||
			parameterSymbol == nil ||
			walk.lowering.checker.GetSymbolAtLocation(object) != parameterSymbol {
			return ""
		}
		return access.Name().Text()
	}
	member := func(side *ast.Node) *checker.Type {
		side = ast.SkipParentheses(side)
		constant := walk.lowering.symbol(side)
		if constant == nil || constant.Flags&ast.SymbolFlagsEnumMember == 0 {
			return nil
		}
		return walk.lowering.checker.GetTypeAtLocation(side)
	}
	field, constant := read(binary.Left), member(binary.Right)
	if field == "" || constant == nil {
		field, constant = read(binary.Right), member(binary.Left)
	}
	if field == "" || constant == nil {
		return kindCall{}, false
	}
	return kindCall{call: node, field: field, member: constant, negated: negated}, true
}

// kindPredicateStep judges an exact kind predicate's call as the comparison its body makes. It
// passes only when that comparison would: the field is read directly off every object member
// flowing in (no accessor), and it can't lie there (enumTagHazardOf, as step asks of a comparison).
// The checker narrows a predicate's call by the predicate's declared type, not by its body, so the
// step also passes only when that narrowing is identical to the comparison's on the side this path
// takes: a predicate whose type disagrees with its own body proves nothing here. R-T80's check at the
// call is kept; only the refusal of the read moves.
func (walk *enumTagWalk) kindPredicateStep(flow *ast.FlowNode, call kindCall) bool {
	before := predicateFlowType(
		walk.lowering.checker,
		walk.reference,
		walk.declared,
		walk.declared,
		nil,
		flow.Antecedent,
	)
	members := []*checker.Type{before}
	if before.Flags()&checker.TypeFlagsUnion != 0 {
		members = before.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		property := walk.lowering.checker.GetPropertyOfType(member, call.field)
		if property == nil ||
			property.Flags&(ast.SymbolFlagsGetAccessor|ast.SymbolFlagsSetAccessor|ast.SymbolFlagsMethod) != 0 {
			return false
		}
	}
	if walk.lowering.enumTagHazardOf(before, call.field) != enumTagNoHazard {
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
		tag := walk.lowering.checker.GetTypeOfSymbol(walk.lowering.checker.GetPropertyOfType(member, call.field))
		if taken {
			if checker.Checker_isTypeAssignableTo(walk.lowering.checker, call.member, tag) {
				kept = append(kept, member)
			}
		} else if !checker.Checker_isTypeIdenticalTo(walk.lowering.checker, tag, call.member) {
			kept = append(kept, member)
		}
	}
	if len(kept) == 0 {
		return false
	}
	after := predicateFlowType(walk.lowering.checker, walk.reference, walk.declared, walk.declared, nil, flow)
	if !checker.Checker_isTypeIdenticalTo(walk.lowering.checker, after, walk.lowering.checker.GetUnionType(kept)) {
		return false
	}
	return walk.predicateProves(flow.Node, call)
}

// directField is the field a recognized comparison reads off the reference: `r.f === x`, `!==`,
// `==`, `!=` (either side), `switch (r.f)`, or the truthiness of `r.f`, under parentheses and `!`.
// An optional chain, an element access, and an accessor are not a field read directly.
func (walk *enumTagWalk) directField(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	for node.Kind == ast.KindPrefixUnaryExpression &&
		node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		node = ast.SkipParentheses(node.AsPrefixUnaryExpression().Operand)
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken, ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
			if field := walk.fieldRead(binary.Left); field != "" {
				return field
			}
			return walk.fieldRead(binary.Right)
		}
		return ""
	}
	return walk.fieldRead(node)
}

func (walk *enumTagWalk) fieldRead(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression || node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return ""
	}
	access := node.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil ||
		!predicateMatchingReference(walk.lowering.checker, walk.reference, access.Expression) {
		return ""
	}
	name := access.Name().Text()
	// On a union the property read is the checker's union of each member's own, which doesn't carry
	// an accessor's flags, so each member's own is asked: a getter is a call, never a field read.
	owner := walk.lowering.checker.GetTypeAtLocation(access.Expression)
	members := []*checker.Type{owner}
	if owner.Flags()&checker.TypeFlagsUnion != 0 {
		members = owner.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		property := walk.lowering.checker.GetPropertyOfType(member, name)
		if property == nil ||
			property.Flags&(ast.SymbolFlagsGetAccessor|ast.SymbolFlagsSetAccessor|ast.SymbolFlagsMethod) != 0 {
			return ""
		}
	}
	return name
}

// accessLocalTests is where an access reference's narrowing may be written (#qeanycw): the tests of
// the conditionals, `&&` chains and ifs that hold the read, walking out from it until a function, a
// class, a loop or a label, any of which would let code run between a test and the read that no
// source order shows. The reference must be a chain of plain data fields off a name: no optional
// link, no element access, no `this` (a constructor still writes its readonly fields), and no link
// a getter or setter could serve, judged by name over the program and failing closed (accessorName).
func (lowering *lowering) accessLocalTests(read *ast.Node) ([]*ast.Node, bool) {
	for node := read; ; {
		access := node.AsPropertyAccessExpression()
		if access.QuestionDotToken != nil ||
			node.Flags&ast.NodeFlagsOptionalChain != 0 ||
			lowering.accessorName(access.Name().Text()) ||
			!lowering.plainDataLink(node) {
			return nil, false
		}
		if access.Expression.Kind == ast.KindIdentifier {
			break
		}
		if access.Expression.Kind != ast.KindPropertyAccessExpression {
			return nil, false
		}
		node = access.Expression
	}
	tests := []*ast.Node{}
	child := read
	for parent := read.Parent; parent != nil; child, parent = parent, parent.Parent {
		if ast.IsFunctionLike(parent) || ast.IsClassLike(parent) || parent.Kind == ast.KindSourceFile {
			break
		}
		switch parent.Kind {
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindLabeledStatement:
			return tests, len(tests) != 0
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if child == conditional.WhenTrue || child == conditional.WhenFalse {
				tests = append(tests, conditional.Condition)
			}
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken && child == binary.Right {
				tests = append(tests, binary.Left)
			}
		case ast.KindIfStatement:
			statement := parent.AsIfStatement()
			if child != statement.Expression {
				tests = append(tests, statement.Expression)
			}
		}
	}
	return tests, len(tests) != 0
}

// plainDataLink is whether one link of an access path is a data field on every object member its
// object can be: never an accessor or a method, whose every read can answer differently.
func (lowering *lowering) plainDataLink(link *ast.Node) bool {
	access := link.AsPropertyAccessExpression()
	owner := lowering.checker.GetTypeAtLocation(access.Expression)
	members := []*checker.Type{owner}
	if owner.Flags()&checker.TypeFlagsUnion != 0 {
		members = owner.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		property := lowering.checker.GetPropertyOfType(member, access.Name().Text())
		if property == nil ||
			property.Flags&(ast.SymbolFlagsGetAccessor|ast.SymbolFlagsSetAccessor|ast.SymbolFlagsMethod) != 0 {
			return false
		}
	}
	return true
}

// accessorName is whether some getter or setter the program writes, in a class or an object literal,
// has this name. An interface's readonly field can be served by a class's getter, which the field's
// own symbol never shows, so the name decides, and an accessor with a computed name makes every name
// one (fail closed). Declaration files write no accessor bodies. Collected once.
func (lowering *lowering) accessorName(name string) bool {
	if lowering.writtenAccessorNames == nil {
		lowering.writtenAccessorNames = map[string]bool{}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
				if key := node.Name(); key != nil && (key.Kind == ast.KindIdentifier ||
					key.Kind == ast.KindStringLiteral ||
					key.Kind == ast.KindPrivateIdentifier) {
					lowering.writtenAccessorNames[key.Text()] = true
				} else {
					lowering.writtenAccessorNames[""] = true
				}
			}
			node.ForEachChild(visit)
			return false
		}
		for _, file := range lowering.program.Files() {
			if !file.IsDeclarationFile {
				file.AsNode().ForEachChild(visit)
			}
		}
	}
	return lowering.writtenAccessorNames[name] || lowering.writtenAccessorNames[""]
}

// codeBetween is whether anything that can run program code completes after from, where an access
// reference's earliest narrowing step starts, and before the read starts, anywhere in the function
// that holds the read. A nested function's body runs only when called, and the call is what counts.
// An exempt node, an exact predicate's call, isn't counted, though its arguments are.
func (lowering *lowering) codeBetween(from int, read *ast.Node, exempt map[*ast.Node]bool) bool {
	container := read.Parent
	for container != nil && !ast.IsFunctionLike(container) && container.Kind != ast.KindSourceFile {
		container = container.Parent
	}
	if container == nil {
		return true
	}
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found || node.Pos() >= read.Pos() {
			return found
		}
		if node.End() > from && node.End() <= read.Pos() && !exempt[node] && lowering.mayRunCode(node) {
			found = true
			return true
		}
		if ast.IsFunctionLike(node) {
			return false
		}
		node.ForEachChild(visit)
		return found
	}
	container.ForEachChild(visit)
	return found
}

// mayRunCode is whether evaluating node can run program code beyond its parts', or change what the
// read sees: a call of any kind, new, a tagged template, await, yield, delete, a class (its static
// parts run), a spread or a destructuring pattern (iterators and getters), instanceof
// (Symbol.hasInstance), an element access, a property read a getter of that name could serve, an
// operator or a template substitution that converts an object to a primitive, and any write to a
// property or an element, which may store through another name for the read's object (`g.n = b`
// with g the same object as h makes no flow node for h.n).
func (lowering *lowering) mayRunCode(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression, ast.KindAwaitExpression, ast.KindYieldExpression,
		ast.KindDeleteExpression, ast.KindClassExpression, ast.KindSpreadElement, ast.KindSpreadAssignment, ast.KindObjectBindingPattern,
		ast.KindArrayBindingPattern, ast.KindElementAccessExpression, ast.KindForOfStatement, ast.KindForInStatement:
		return true
	case ast.KindPropertyAccessExpression:
		return lowering.accessorName(node.AsPropertyAccessExpression().Name().Text())
	case ast.KindTemplateSpan:
		return !lowering.primitiveTyped(node.AsTemplateSpan().Expression)
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			return writesStorage(unary.Operand)
		}
		return (unary.Operator == ast.KindPlusToken ||
			unary.Operator == ast.KindMinusToken ||
			unary.Operator == ast.KindTildeToken) && !lowering.primitiveTyped(unary.Operand)
	case ast.KindPostfixUnaryExpression:
		return writesStorage(node.AsPostfixUnaryExpression().Operand)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) && writesStorage(binary.Left) {
			return true
		}
		switch binary.OperatorToken.Kind {
		case ast.KindInstanceOfKeyword:
			return true
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken, ast.KindAmpersandAmpersandToken, ast.KindBarBarToken,
			ast.KindQuestionQuestionToken, ast.KindCommaToken, ast.KindInKeyword, ast.KindEqualsToken, ast.KindBarBarEqualsToken,
			ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken:
			// A write's target is a property access or element access judged on its own, or a name.
			return binary.Left.Kind == ast.KindObjectLiteralExpression ||
				binary.Left.Kind == ast.KindArrayLiteralExpression
		}
		return !lowering.primitiveTyped(binary.Left) || !lowering.primitiveTyped(binary.Right)
	}
	return false
}

// writesStorage is whether a write's target is a property or an element, not a plain name.
func writesStorage(target *ast.Node) bool {
	target = ast.SkipParentheses(target)
	return target.Kind == ast.KindPropertyAccessExpression || target.Kind == ast.KindElementAccessExpression
}

// primitiveTyped is whether every member of node's type is a primitive, whose conversion runs no
// program code.
func (lowering *lowering) primitiveTyped(node *ast.Node) bool {
	of := lowering.checker.GetTypeAtLocation(node)
	members := []*checker.Type{of}
	if of.Flags()&checker.TypeFlagsUnion != 0 {
		members = of.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsPrimitive == 0 {
			return false
		}
	}
	return true
}
