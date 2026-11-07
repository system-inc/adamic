package lower

import (
	"fmt"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/load"
)

// A checked predicate used as a value carries only its true check (#vdt9f1q, ruled Oct 25): where
// it's passed, stored or returned, nothing sees how a call through it is used, so a read the checker
// narrowed on the false side of such a call would believe an answer nobody checked. In a program
// where a checked predicate escapes as a value, each read the checker narrowed is walked back through
// tsc's flow graph, every path, and a false step that narrows it by a call whose callee isn't a direct
// reference to a predicate declaration makes the read NotYet. The walk reads the step the checker
// took, keyed on the call, so it doesn't matter where the callee came from, and a flow node it doesn't
// know stops the read too. A direct call is R-T80's, checked both ways at the call.

// The checker's own choice of the signature a condition's call narrows by.
//
//go:linkname predicateEffectsSignature github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getEffectsSignature
func predicateEffectsSignature(receiver *checker.Checker, node *ast.Node) *checker.Signature

// The checker's type for a reference before flow narrows it, in that reference's position: a T
// extends Node read as n.e narrows from Node.
//
//go:linkname predicateNarrowableType github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getNarrowableTypeForReference
func predicateNarrowableType(
	receiver *checker.Checker,
	subjectType *checker.Type,
	reference *ast.Node,
	checkMode checker.CheckMode,
) *checker.Type

// predicateFalseFlow is the program's walk, made with its first module: whether the rule is live,
// and what each flow node, condition and reference has been found to reach, so no read walks
// further than the first read that reached the same node.
type predicateFalseFlow struct {
	lowering *lowering
	live     bool
	// steps holds each finished node's answer: the first false step through a predicate value on a
	// path back from it, or nil when there is none. The reference is "" for the reading that asks
	// whether any reference could meet one, which most reads stop at.
	steps map[predicateFalseKey]*predicateFalseStep
	// calls is each condition's call through a predicate value that its false side narrows by.
	calls map[*ast.FlowNode]*ast.Node
	// checked is whether a predicate declaration is one R-T80 checks.
	checked map[*ast.Node]bool
}

type predicateFalseKey struct {
	flow      *ast.FlowNode
	reference string
	declared  *checker.Type
}

// predicateFalseStep is why a read stops: the call whose false side narrowed it, or a flow node the
// walk doesn't know.
type predicateFalseStep struct {
	call    *ast.Node
	unknown *ast.FlowNode
}

// predicateFalseReads stops at the first read in module that the checker narrowed on the false side
// of a call through a predicate value.
func (lowering *lowering) predicateFalseReads(module *ast.SourceFile) error {
	if lowering.predicateFalse == nil {
		lowering.predicateFalse = &predicateFalseFlow{
			lowering: lowering,
			steps:    map[predicateFalseKey]*predicateFalseStep{},
			calls:    map[*ast.FlowNode]*ast.Node{},
			checked:  map[*ast.Node]bool{},
		}
		lowering.predicateFalse.live = lowering.predicateFalse.escapes()
	}
	if !lowering.predicateFalse.live {
		return nil
	}
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		found = lowering.predicateFalse.read(node)
		if found == nil {
			node.ForEachChild(visit)
		}
		return found != nil
	}
	module.AsNode().ForEachChild(visit)
	return found
}

// escapes is the switch (#vdt9f1q, Adamic's correction): some checked predicate is used as a value
// anywhere in the program. That's a reference to a predicate declaration R-T80 checks that isn't
// the callee of a direct call, wherever it goes (an alias, a field, an argument, a return, an
// element, bind), or a function expression or arrow claiming a predicate it doesn't prove, which is
// a value by construction. A program with neither keeps its admission exactly. The value is found by
// its type as well as by its name (Compiler's must on #vdt9f1q): any expression whose type carries a
// checked predicate's call signature, other than a direct call's callee, is one, so a road that
// doesn't spell the name, ns["isS"], is found too.
func (falseFlow *predicateFalseFlow) escapes() bool {
	if !falseFlow.anyChecked() {
		return false
	}
	escaped := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch {
		case escaped:
		case ast.IsFunctionExpressionOrArrowFunction(node):
			escaped = node.Type() != nil &&
				node.Type().Kind == ast.KindTypePredicate &&
				falseFlow.lowering.checkedAssertionSource(node) &&
				!falseFlow.lowering.checkedPredicateBodyProven(node.Type())
		case node.Kind == ast.KindIdentifier && falseFlow.escapingReference(node):
			escaped = true
		case falseFlow.carrier(node):
			escaped = falseFlow.carries(falseFlow.lowering.checker.GetTypeAtLocation(node), map[*checker.Type]bool{})
		}
		if !escaped {
			node.ForEachChild(visit)
		}
		return escaped
	}
	for _, file := range falseFlow.lowering.program.Files() {
		if file.AsNode().ForEachChild(visit) {
			return true
		}
	}
	return false
}

// escapingReference is whether identifier names a checked predicate declaration as a value. A
// shorthand property `{ isS }` stores it; a qualified namespace/module call names the
// declaration directly, but a reference stored or passed onward is still a value.
func (falseFlow *predicateFalseFlow) escapingReference(identifier *ast.Node) bool {
	parent := identifier.Parent
	var symbol *ast.Symbol
	switch {
	case parent.Kind == ast.KindShorthandPropertyAssignment:
		symbol = falseFlow.lowering.checker.GetShorthandAssignmentValueSymbol(parent)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = falseFlow.lowering.checker.GetAliasedSymbol(symbol)
		}
	case ast.IsPartOfTypeNode(identifier) || ast.IsDeclarationName(identifier):
		return false
	case parent.Kind == ast.KindImportSpecifier ||
		parent.Kind == ast.KindExportSpecifier ||
		parent.Kind == ast.KindImportClause ||
		parent.Kind == ast.KindNamespaceImport:
		return false
	default:
		symbol = falseFlow.lowering.symbol(identifier)
	}
	if symbol == nil || symbol.Flags&ast.SymbolFlagsFunction == 0 {
		return false
	}
	checked := false
	for _, declaration := range symbol.Declarations {
		checked = checked || falseFlow.checkedDeclaration(declaration)
	}
	if !checked {
		return false
	}
	reference := identifier
	if parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Name() == identifier {
		reference = parent
	}
	callee := reference
	for callee.Parent != nil && callee.Parent.Kind == ast.KindParenthesizedExpression {
		callee = callee.Parent
	}
	direct := false
	if callee.Parent != nil &&
		callee.Parent.Kind == ast.KindCallExpression &&
		callee.Parent.AsCallExpression().Expression == callee {
		for _, declaration := range symbol.Declarations {
			direct = direct || falseFlow.lowering.directPredicateCallee(reference, declaration)
		}
	}
	return parent.Kind == ast.KindShorthandPropertyAssignment || !direct
}

// anyChecked is whether the program declares a predicate R-T80 checks, so a program with none never
// pays for reading its expressions' types. A checked closure counts (#1w5bve2): it checks only its
// true side, so a program whose one checked predicate is an arrow needs the walk for its false side.
// So does a checked claim on a function type, whose calls check only their true side.
func (falseFlow *predicateFalseFlow) anyChecked() bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		found = found ||
			(node.Kind == ast.KindFunctionDeclaration && falseFlow.checkedDeclaration(node)) ||
			falseFlow.lowering.checkedPredicateClosure(node) ||
			falseFlow.lowering.checkedFunctionTypeClaim(node)
		if !found {
			node.ForEachChild(visit)
		}
		return found
	}
	for _, file := range falseFlow.lowering.program.Files() {
		if file.AsNode().ForEachChild(visit) {
			return true
		}
	}
	return false
}

// carrier is whether node is an expression that can carry a function value onward, so its type is
// worth reading: a name, a property or element access, or a call's result (Lower's bound, for cost:
// the walk is live in every tsc file). A name a call calls directly is R-T80's, checked both ways at
// that call, so it's no value.
func (falseFlow *predicateFalseFlow) carrier(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindIdentifier:
		if ast.IsPartOfTypeNode(node) || ast.IsDeclarationName(node) || !ast.IsExpressionNode(node) {
			return false
		}
		// The whole property expression owns this reference's type. Visiting its
		// name again would count even a direct namespace call as a value escape.
		if parent := node.Parent; parent != nil &&
			parent.Kind == ast.KindPropertyAccessExpression &&
			parent.Name() == node {
			return false
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression:
		if ast.IsPartOfTypeNode(node) {
			return false
		}
	default:
		return false
	}
	callee := node
	for callee.Parent != nil && callee.Parent.Kind == ast.KindParenthesizedExpression {
		callee = callee.Parent
	}
	if call := callee.Parent; call != nil && call.Kind == ast.KindCallExpression && call.Expression() == callee {
		if signature := falseFlow.lowering.checker.GetResolvedSignature(
			call,
		); signature != nil && falseFlow.lowering.directPredicateCallee(node, signature.Declaration()) {
			return false
		}
	}
	return true
}

// carries is whether a value of type `subjectType` can be a checked predicate: one of its call signatures, or a
// union or intersection member's, is declared by one.
func (falseFlow *predicateFalseFlow) carries(subjectType *checker.Type, seen map[*checker.Type]bool) bool {
	if subjectType == nil || seen[subjectType] {
		return false
	}
	seen[subjectType] = true
	if subjectType.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range subjectType.Types() {
			if falseFlow.carries(member, seen) {
				return true
			}
		}
	}
	for _, signature := range falseFlow.lowering.checker.GetSignaturesOfType(subjectType, checker.SignatureKindCall) {
		if falseFlow.checkedDeclaration(signature.Declaration()) {
			return true
		}
	}
	return false
}

// checkedDeclaration is checkedPredicateDeclaration, remembered per declaration.
func (falseFlow *predicateFalseFlow) checkedDeclaration(declaration *ast.Node) bool {
	checked, seen := falseFlow.checked[declaration]
	if !seen {
		checked = falseFlow.lowering.checkedPredicateDeclaration(declaration)
		falseFlow.checked[declaration] = checked
	}
	return checked
}

// read judges one node. Only a reference the checker can narrow is asked about: a variable or
// parameter, this, or a property or element access off one.
func (falseFlow *predicateFalseFlow) read(node *ast.Node) error {
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
	default:
		return nil
	}
	data := node.FlowNodeData()
	if data == nil || data.FlowNode == nil || falseFlow.search(data.FlowNode, nil, nil, "") == nil {
		return nil
	}
	var declared *checker.Type
	if node.Kind != ast.KindThisKeyword {
		declared = falseFlow.declared(node)
		// A read whose type is the one it was declared with believes no narrowing, whatever steps
		// lie behind it, as after an if whose branches join.
		if declared == nil || checker.Checker_isTypeIdenticalTo(
			falseFlow.lowering.checker,
			falseFlow.lowering.checker.GetTypeAtLocation(node),
			declared,
		) {
			return nil
		}
	}
	key, keyed := predicateFalseReferenceKey(falseFlow.lowering, node)
	if !keyed {
		key = fmt.Sprintf("%p", node)
	}
	step := falseFlow.search(data.FlowNode, node, declared, key)
	if step == nil {
		return nil
	}
	if step.call == nil {
		return falseFlow.lowering.notYet(node, fmt.Sprintf(
			"a read of %s behind a flow node the false-direction walk doesn't know (flags %d)",
			predicateFalseText(node),
			step.unknown.Flags,
		))
	}
	return falseFlow.lowering.notYet(node, fmt.Sprintf(
		"a read of %s narrowed on the false side of %s at %s, a call through a predicate value whose false answer isn't checked",
		predicateFalseText(node),
		predicateFalseText(step.call),
		falseFlow.lowering.program.Where(step.call),
	))
}

// declared is the type the checker narrows reference from, in its own position, or nil when the
// reference isn't one flow narrows.
func (falseFlow *predicateFalseFlow) declared(reference *ast.Node) *checker.Type {
	var symbol *ast.Symbol
	switch reference.Kind {
	case ast.KindIdentifier:
		parent := reference.Parent
		if ast.IsPartOfTypeNode(reference) ||
			ast.IsDeclarationName(reference) ||
			(parent.Kind == ast.KindPropertyAccessExpression &&
				parent.AsPropertyAccessExpression().Name() == reference) {
			return nil
		}
		symbol = falseFlow.lowering.symbol(reference)
		if symbol == nil || symbol.Flags&ast.SymbolFlagsVariable == 0 {
			return nil
		}
	case ast.KindPropertyAccessExpression:
		symbol = falseFlow.lowering.checker.GetSymbolAtLocation(reference.AsPropertyAccessExpression().Name())
	case ast.KindElementAccessExpression:
		symbol = falseFlow.lowering.checker.GetSymbolAtLocation(
			reference.AsElementAccessExpression().ArgumentExpression,
		)
	}
	if symbol == nil || symbol.Flags&(ast.SymbolFlagsVariable|ast.SymbolFlagsProperty|ast.SymbolFlagsAccessor) == 0 {
		return nil
	}
	return predicateNarrowableType(
		falseFlow.lowering.checker,
		falseFlow.lowering.checker.GetTypeOfSymbol(symbol),
		reference,
		checker.CheckModeNormal,
	)
}

// search answers flow for reference, finding every node a path back from it reaches. A loop makes
// the flow graph cyclic, so it's Tarjan's walk: a cycle's nodes finish together, with one answer,
// and every node is memoized once it's finished, so a later read stops at the first finished node.
// With reference nil it asks about any reference: every step through a predicate value counts, and
// every path is followed.
func (falseFlow *predicateFalseFlow) search(
	flow *ast.FlowNode,
	reference *ast.Node,
	declared *checker.Type,
	key string,
) *predicateFalseStep {
	if step, done := falseFlow.steps[predicateFalseKey{flow: flow, reference: key, declared: declared}]; done {
		return step
	}
	walk := &predicateFalseSearch{
		flow:      falseFlow,
		reference: reference,
		declared:  declared,
		key:       key,
		index:     map[*ast.FlowNode]int{},
		low:       map[*ast.FlowNode]int{},
		partial:   map[*ast.FlowNode]*predicateFalseStep{},
		stacked:   map[*ast.FlowNode]bool{},
	}
	return walk.visit(flow)
}

type predicateFalseSearch struct {
	flow      *predicateFalseFlow
	reference *ast.Node
	declared  *checker.Type
	key       string
	index     map[*ast.FlowNode]int
	low       map[*ast.FlowNode]int
	partial   map[*ast.FlowNode]*predicateFalseStep
	stack     []*ast.FlowNode
	stacked   map[*ast.FlowNode]bool
}

func (search *predicateFalseSearch) memo(flow *ast.FlowNode) predicateFalseKey {
	return predicateFalseKey{flow: flow, reference: search.key, declared: search.declared}
}

func (search *predicateFalseSearch) visit(flow *ast.FlowNode) *predicateFalseStep {
	search.index[flow] = len(search.index)
	search.low[flow] = search.index[flow]
	search.stack = append(search.stack, flow)
	search.stacked[flow] = true
	step, next := search.step(flow)
	for _, antecedent := range next {
		if antecedent == nil {
			continue
		}
		if done, finished := search.flow.steps[search.memo(antecedent)]; finished {
			step = predicateFalseFirst(step, done)
			continue
		}
		if _, visited := search.index[antecedent]; !visited {
			step = predicateFalseFirst(step, search.visit(antecedent))
			search.low[flow] = min(search.low[flow], search.low[antecedent])
		} else if search.stacked[antecedent] {
			search.low[flow] = min(search.low[flow], search.index[antecedent])
		}
	}
	search.partial[flow] = step
	if search.low[flow] != search.index[flow] {
		return step
	}
	// flow roots a cycle, or stands alone: each member's answer is all of theirs together.
	start := len(search.stack) - 1
	for search.stack[start] != flow {
		start--
	}
	members := search.stack[start:]
	for _, member := range members {
		step = predicateFalseFirst(step, search.partial[member])
	}
	for _, member := range members {
		search.flow.steps[search.memo(member)] = step
		search.stacked[member] = false
	}
	search.stack = search.stack[:start]
	return step
}

// step is what flow itself says, and the antecedents a path back goes on to.
func (search *predicateFalseSearch) step(flow *ast.FlowNode) (*predicateFalseStep, []*ast.FlowNode) {
	switch {
	case flow.Flags&ast.FlowFlagsUnreachable != 0:
		return nil, nil
	case flow.Flags&ast.FlowFlagsStart != 0:
		// As the checker does, a name, or this inside an arrow, carries narrowing in from the
		// function around it, so a closure made after `if (!test(n)) return` reads the outer step.
		// A hoisted declaration's start has no node, and nothing reaches it from outside.
		reference := search.reference
		if flow.Node != nil && (reference == nil ||
			reference.Kind == ast.KindIdentifier ||
			(reference.Kind == ast.KindThisKeyword && ast.IsArrowFunction(flow.Node))) {
			if outer := flow.Node.FlowNodeData(); outer != nil {
				return nil, []*ast.FlowNode{outer.FlowNode}
			}
		}
		return nil, nil
	case flow.Flags&ast.FlowFlagsAssignment != 0:
		// An assignment to the reference, or to what it's read off, ends what narrowed it before.
		// A compound one reads what came before it.
		if search.reference != nil &&
			(predicateMatchingReference(search.flow.lowering.checker, search.reference, flow.Node) &&
				!predicateCompoundAssignment(flow.Node) ||
				enumTagContainsReference(search.flow.lowering.checker, search.reference, flow.Node)) {
			return nil, nil
		}
		return nil, []*ast.FlowNode{flow.Antecedent}
	case flow.Flags&(ast.FlowFlagsCondition|ast.FlowFlagsSwitchClause) != 0:
		call := search.flow.condition(flow)
		if call != nil && search.narrows(flow, call) {
			return &predicateFalseStep{call: call}, []*ast.FlowNode{flow.Antecedent}
		}
		return nil, []*ast.FlowNode{flow.Antecedent}
	case flow.Flags&(ast.FlowFlagsCall|ast.FlowFlagsArrayMutation|ast.FlowFlagsReduceLabel) != 0:
		// A call narrows only by an assertion, its true direction; a reduce label's target is
		// walked whole where the label is reached.
		return nil, []*ast.FlowNode{flow.Antecedent}
	case flow.Flags&ast.FlowFlagsLabel != 0:
		next := []*ast.FlowNode{}
		for list := flow.Antecedents; list != nil; list = list.Next {
			next = append(next, list.Flow)
		}
		return nil, next
	}
	return &predicateFalseStep{unknown: flow}, nil
}

// narrows is whether the step at flow changes the reference's type, judged against the type
// flowing into it: a false side equal to what came in believes nothing (Compiler's refinement). The
// reading about any reference counts every step. this has no declared type to read here, so it
// counts a step whose call reads this.
func (search *predicateFalseSearch) narrows(flow *ast.FlowNode, call *ast.Node) bool {
	switch {
	case search.reference == nil:
		return true
	case search.declared == nil:
		return predicateFalseReadsThis(call)
	}
	before := predicateFlowType(
		search.flow.lowering.checker,
		search.reference,
		search.declared,
		search.declared,
		nil,
		flow.Antecedent,
	)
	after := predicateFlowType(
		search.flow.lowering.checker,
		search.reference,
		search.declared,
		search.declared,
		nil,
		flow,
	)
	return !checker.Checker_isTypeIdenticalTo(search.flow.lowering.checker, before, after)
}

// condition is the call through a predicate value that flow's false side narrows by, or nil. A
// switch (true) passes by each clause before the ones reached on its false side, and each after
// them too when a default is among them, as the checker's narrowTypeBySwitchOnTrue does; the
// clauses reached narrow by their true side. Any other switch narrows by no call.
func (falseFlow *predicateFalseFlow) condition(flow *ast.FlowNode) *ast.Node {
	if call, done := falseFlow.calls[flow]; done {
		return call
	}
	var call *ast.Node
	if flow.Flags&ast.FlowFlagsCondition != 0 {
		truth := predicateFalse
		if flow.Flags&ast.FlowFlagsTrueCondition != 0 {
			truth = predicateTrue
		}
		call = falseFlow.conditionCall(flow.Node, truth, 0)
	} else {
		data := flow.Node.AsFlowSwitchClauseData()
		if ast.SkipParentheses(data.SwitchStatement.AsSwitchStatement().Expression).Kind == ast.KindTrueKeyword {
			clauses := data.SwitchStatement.AsSwitchStatement().CaseBlock.AsCaseBlock().Clauses.Nodes
			start, end := int(data.ClauseStart), int(data.ClauseEnd)
			fallsFromDefault := start == end
			for index := start; index < end; index++ {
				fallsFromDefault = fallsFromDefault || clauses[index].Kind == ast.KindDefaultClause
			}
			for index, clause := range clauses {
				if call == nil &&
					clause.Kind == ast.KindCaseClause &&
					(index < start || (fallsFromDefault && index >= end)) {
					call = falseFlow.conditionCall(clause.Expression(), predicateFalse, 0)
				}
			}
		}
	}
	falseFlow.calls[flow] = call
	return call
}

// conditionCall follows expr as the checker's narrowType does when it's assumed truth, to a call
// through a predicate value narrowed on its false side. It reads through more than the checker
// does, never less: both sides of every comparison, and a let alias as well as a const one, so a
// form the checker narrows through is never missed; the step's own type change then says whether
// the checker narrowed.
func (falseFlow *predicateFalseFlow) conditionCall(expr *ast.Node, truth predicateTruth, depth int) *ast.Node {
	// For a?.b and the left of ??, the checker narrows by presence only.
	if ast.IsExpressionOfOptionalChainRoot(expr) || (expr.Parent != nil &&
		expr.Parent.Kind == ast.KindBinaryExpression &&
		expr.Parent.AsBinaryExpression().Left == expr &&
		(expr.Parent.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionToken ||
			expr.Parent.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken)) {
		return nil
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		// The checker reads a condition alias through its initializer, up to five levels:
		// const ok = test(n); if (!ok).
		symbol := falseFlow.lowering.symbol(expr)
		if depth < 5 &&
			symbol != nil &&
			symbol.ValueDeclaration != nil &&
			ast.IsVariableDeclaration(symbol.ValueDeclaration) &&
			symbol.ValueDeclaration.Type() == nil &&
			symbol.ValueDeclaration.Initializer() != nil {
			return falseFlow.conditionCall(symbol.ValueDeclaration.Initializer(), truth, depth+1)
		}
	case ast.KindCallExpression:
		return falseFlow.valueCall(expr, truth)
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
		return falseFlow.conditionCall(expr.Expression(), truth, depth)
	case ast.KindPrefixUnaryExpression:
		if expr.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
			flipped := truth
			if truth != predicateEither {
				flipped = predicateEither &^ truth
			}
			return falseFlow.conditionCall(expr.AsPrefixUnaryExpression().Operand, flipped, depth)
		}
	case ast.KindBinaryExpression:
		binary := expr.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindCommaToken:
			return falseFlow.conditionCall(binary.Right, truth, depth)
		case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			// A comparison with true or false narrows its other side either way round.
			if call := falseFlow.conditionCall(binary.Left, predicateEither, depth); call != nil {
				return call
			}
			return falseFlow.conditionCall(binary.Right, predicateEither, depth)
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			if call := falseFlow.conditionCall(binary.Left, truth, depth); call != nil {
				return call
			}
			return falseFlow.conditionCall(binary.Right, truth, depth)
		case ast.KindInstanceOfKeyword:
			// A [Symbol.hasInstance] predicate narrows as a call does.
			return falseFlow.valueCall(expr, truth)
		}
	}
	return nil
}

// valueCall is call when its false side narrows by a type predicate and its callee isn't a direct
// reference to the predicate's declaration. The checker reads an optional call's false side as
// no narrowing.
func (falseFlow *predicateFalseFlow) valueCall(call *ast.Node, truth predicateTruth) *ast.Node {
	if truth&predicateFalse == 0 ||
		(call.Kind == ast.KindCallExpression && call.Flags&ast.NodeFlagsOptionalChain != 0) {
		return nil
	}
	signature := predicateEffectsSignature(falseFlow.lowering.checker, call)
	if signature == nil {
		return nil
	}
	predicate := predicateOfSignature(falseFlow.lowering.checker, signature)
	if predicate == nil ||
		(predicate.Kind() != checker.TypePredicateKindIdentifier && predicate.Kind() != checker.TypePredicateKindThis) {
		return nil
	}
	if call.Kind == ast.KindCallExpression {
		declaration := signature.Declaration()
		callee := ast.SkipParentheses(call.AsCallExpression().Expression)
		if falseFlow.lowering.directPredicateCallee(callee, declaration) {
			return nil
		}
		if falseFlow.libraryCall(call, callee, declaration) {
			return nil
		}
	}
	return call
}

// libraryCall is whether call asks a library predicate by its own name (Lower's ruling, #1761dah):
// what TypeScript's library or Adamic's prelude declares, the engine answers both ways, so its false
// side is as true as its true one. The callee must be the member itself, `Array.isArray(x)`, its
// symbol the declaration's own or an instantiation of it, never a value of its type: `const f:
// typeof Array.isArray` is f, and a mapped `Pick<ArrayConstructor, 'isArray'>` field is a new
// symbol. And no argument may be a predicate, so `!arr.every(isS)`, whose `this is S[]` answer is
// the callback's, still counts.
func (falseFlow *predicateFalseFlow) libraryCall(call *ast.Node, callee *ast.Node, declaration *ast.Node) bool {
	if declaration == nil || !(load.IsLibrary(ast.GetSourceFileOfNode(declaration)) ||
		load.IsPrelude(ast.GetSourceFileOfNode(declaration))) {
		return false
	}
	name := callee
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		name = callee.AsPropertyAccessExpression().Name()
	case ast.KindIdentifier:
	default:
		return false
	}
	symbol := falseFlow.lowering.checker.GetSymbolAtLocation(name)
	if symbol == nil || falseFlow.lowering.checker.GetMergedSymbol(
		falseFlow.lowering.checker.GetTargetSymbol(symbol),
	) != falseFlow.lowering.checker.GetSymbolOfDeclaration(declaration) {
		return false
	}
	for _, argument := range call.Arguments() {
		if falseFlow.lowering.predicateValued(falseFlow.lowering.checker.GetTypeAtLocation(argument)) {
			return false
		}
	}
	return true
}

// predicateValued is whether a value of type `subjectType` can be called as a type predicate: some call
// signature of it, or of a union or intersection member, carries one.
func (lowering *lowering) predicateValued(subjectType *checker.Type) bool {
	if subjectType.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range subjectType.Types() {
			if lowering.predicateValued(member) {
				return true
			}
		}
	}
	for _, signature := range lowering.checker.GetSignaturesOfType(subjectType, checker.SignatureKindCall) {
		if predicateOfSignature(lowering.checker, signature) != nil {
			return true
		}
	}
	return false
}

// predicateFalseFirst is the earlier of two steps in the source, so a read's answer doesn't depend
// on which read reached a node first.
func predicateFalseFirst(left, right *predicateFalseStep) *predicateFalseStep {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	position := func(step *predicateFalseStep) (string, int) {
		if step.call == nil {
			return "", -1
		}
		return string(ast.GetSourceFileOfNode(step.call).FileName()), step.call.Pos()
	}
	aFile, aPosition := position(left)
	bFile, bPosition := position(right)
	if bFile < aFile || (bFile == aFile && bPosition < aPosition) {
		return right
	}
	return left
}

// predicateFalseReferenceKey names reference so that two references the checker matches share
// their answers: the symbol of a name, and each property read off it. Any other shape is its own.
func predicateFalseReferenceKey(lowering *lowering, reference *ast.Node) (string, bool) {
	suffix := ""
	if reference.Parent != nil && reference.Parent.Kind == ast.KindNonNullExpression {
		// The checker's last step treats a read under ! apart.
		suffix = "!"
	}
	var key func(node *ast.Node) (string, bool)
	key = func(node *ast.Node) (string, bool) {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindIdentifier:
			symbol := lowering.symbol(node)
			return fmt.Sprintf("%p", symbol), symbol != nil
		case ast.KindThisKeyword:
			return "this", true
		case ast.KindPropertyAccessExpression:
			access := node.AsPropertyAccessExpression()
			object, ok := key(access.Expression)
			return object + "." + access.Name().Text(), ok && ast.IsIdentifier(access.Name())
		case ast.KindElementAccessExpression:
			access := node.AsElementAccessExpression()
			object, ok := key(access.Expression)
			argument := access.ArgumentExpression
			return object +
				"[" +
				argument.Text() +
				"]", ok && (argument.Kind == ast.KindStringLiteral || argument.Kind == ast.KindNumericLiteral)
		}
		return "", false
	}
	name, ok := key(reference)
	return name + suffix, ok
}

// predicateCompoundAssignment is whether target is written by a compound operator, which reads it.
func predicateCompoundAssignment(target *ast.Node) bool {
	parent := target.Parent
	return parent != nil &&
		parent.Kind == ast.KindBinaryExpression &&
		parent.AsBinaryExpression().Left == target &&
		ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind) &&
		parent.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken
}

// predicateFalseReadsThis is whether call's predicate reads this: as its argument, or as the object
// a this predicate is called on.
func predicateFalseReadsThis(call *ast.Node) bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		found = found || node.Kind == ast.KindThisKeyword
		if !found && !ast.IsFunctionLike(node) {
			node.ForEachChild(visit)
		}
		return found
	}
	call.ForEachChild(visit)
	return found
}

func predicateFalseText(node *ast.Node) string {
	file := ast.GetSourceFileOfNode(node)
	return file.Text()[scanner.GetTokenPosOfNode(node, file, false):node.End()]
}
