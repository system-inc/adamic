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
func predicateNarrowableType(receiver *checker.Checker, t *checker.Type, reference *ast.Node, checkMode checker.CheckMode) *checker.Type

// predicateFalseFlow is the program's walk, made with its first module: whether the rule is live,
// and what each flow node, condition and reference has been found to reach, so no read walks
// further than the first read that reached the same node.
type predicateFalseFlow struct {
	l    *lowering
	live bool
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
func (l *lowering) predicateFalseReads(module *ast.SourceFile) error {
	if l.predicateFalse == nil {
		l.predicateFalse = &predicateFalseFlow{l: l, steps: map[predicateFalseKey]*predicateFalseStep{}, calls: map[*ast.FlowNode]*ast.Node{}, checked: map[*ast.Node]bool{}}
		l.predicateFalse.live = l.predicateFalse.escapes()
	}
	if !l.predicateFalse.live {
		return nil
	}
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		found = l.predicateFalse.read(node)
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
// a value by construction. A program with neither keeps its admission exactly.
func (f *predicateFalseFlow) escapes() bool {
	escaped := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch {
		case escaped:
		case ast.IsFunctionExpressionOrArrowFunction(node):
			escaped = node.Type() != nil && node.Type().Kind == ast.KindTypePredicate && f.l.checkedAssertionSource(node) && !f.l.checkedPredicateBodyProven(node.Type())
		case node.Kind == ast.KindIdentifier:
			escaped = f.escapingReference(node)
		}
		if !escaped {
			node.ForEachChild(visit)
		}
		return escaped
	}
	for _, file := range f.l.program.Files() {
		if file.AsNode().ForEachChild(visit) {
			return true
		}
	}
	return false
}

// escapingReference is whether identifier names a checked predicate declaration as a value. A
// shorthand property `{ isS }` stores it; a name read off a namespace, `ns.isS`, is a reference
// whose callee isn't an identifier, so it's no direct call either.
func (f *predicateFalseFlow) escapingReference(identifier *ast.Node) bool {
	parent := identifier.Parent
	var symbol *ast.Symbol
	switch {
	case parent.Kind == ast.KindShorthandPropertyAssignment:
		symbol = f.l.checker.GetShorthandAssignmentValueSymbol(parent)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = f.l.checker.GetAliasedSymbol(symbol)
		}
	case ast.IsPartOfTypeNode(identifier) || ast.IsDeclarationName(identifier):
		return false
	case parent.Kind == ast.KindImportSpecifier || parent.Kind == ast.KindExportSpecifier || parent.Kind == ast.KindImportClause || parent.Kind == ast.KindNamespaceImport:
		return false
	default:
		symbol = f.l.symbol(identifier)
	}
	if symbol == nil || symbol.Flags&ast.SymbolFlagsFunction == 0 {
		return false
	}
	checked := false
	for _, declaration := range symbol.Declarations {
		checked = checked || f.checkedDeclaration(declaration)
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
	direct := reference.Kind == ast.KindIdentifier && callee.Parent != nil && callee.Parent.Kind == ast.KindCallExpression && callee.Parent.AsCallExpression().Expression == callee
	return parent.Kind == ast.KindShorthandPropertyAssignment || !direct
}

// checkedDeclaration is predicates.go's own test for a predicate R-T80 checks at a direct call.
func (f *predicateFalseFlow) checkedDeclaration(declaration *ast.Node) bool {
	checked, seen := f.checked[declaration]
	if !seen {
		checked = declaration.Kind == ast.KindFunctionDeclaration && declaration.Type() != nil && declaration.Type().Kind == ast.KindTypePredicate && (declaration.Body() != nil || f.l.censusImplementation(declaration) != nil) && f.l.checkedAssertionSource(declaration) && !f.l.checkedPredicateBodyProven(declaration.Type())
		f.checked[declaration] = checked
	}
	return checked
}

// read judges one node. Only a reference the checker can narrow is asked about: a variable or
// parameter, this, or a property or element access off one.
func (f *predicateFalseFlow) read(node *ast.Node) error {
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
	default:
		return nil
	}
	data := node.FlowNodeData()
	if data == nil || data.FlowNode == nil || f.search(data.FlowNode, nil, nil, "") == nil {
		return nil
	}
	var declared *checker.Type
	if node.Kind != ast.KindThisKeyword {
		declared = f.declared(node)
		// A read whose type is the one it was declared with believes no narrowing, whatever steps
		// lie behind it, as after an if whose branches join.
		if declared == nil || checker.Checker_isTypeIdenticalTo(f.l.checker, f.l.checker.GetTypeAtLocation(node), declared) {
			return nil
		}
	}
	key, keyed := predicateFalseReferenceKey(f.l, node)
	if !keyed {
		key = fmt.Sprintf("%p", node)
	}
	step := f.search(data.FlowNode, node, declared, key)
	if step == nil {
		return nil
	}
	if step.call == nil {
		return f.l.notYet(node, fmt.Sprintf("a read of %s behind a flow node the false-direction walk doesn't know (flags %d)", predicateFalseText(node), step.unknown.Flags))
	}
	return f.l.notYet(node, fmt.Sprintf("a read of %s narrowed on the false side of %s at %s, a call through a predicate value whose false answer isn't checked", predicateFalseText(node), predicateFalseText(step.call), f.l.program.Where(step.call)))
}

// declared is the type the checker narrows reference from, in its own position, or nil when the
// reference isn't one flow narrows.
func (f *predicateFalseFlow) declared(reference *ast.Node) *checker.Type {
	var symbol *ast.Symbol
	switch reference.Kind {
	case ast.KindIdentifier:
		parent := reference.Parent
		if ast.IsPartOfTypeNode(reference) || ast.IsDeclarationName(reference) || (parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Name() == reference) {
			return nil
		}
		symbol = f.l.symbol(reference)
		if symbol == nil || symbol.Flags&ast.SymbolFlagsVariable == 0 {
			return nil
		}
	case ast.KindPropertyAccessExpression:
		symbol = f.l.checker.GetSymbolAtLocation(reference.AsPropertyAccessExpression().Name())
	case ast.KindElementAccessExpression:
		symbol = f.l.checker.GetSymbolAtLocation(reference.AsElementAccessExpression().ArgumentExpression)
	}
	if symbol == nil || symbol.Flags&(ast.SymbolFlagsVariable|ast.SymbolFlagsProperty|ast.SymbolFlagsAccessor) == 0 {
		return nil
	}
	return predicateNarrowableType(f.l.checker, f.l.checker.GetTypeOfSymbol(symbol), reference, checker.CheckModeNormal)
}

// search answers flow for reference, finding every node a path back from it reaches. A loop makes
// the flow graph cyclic, so it's Tarjan's walk: a cycle's nodes finish together, with one answer,
// and every node is memoized once it's finished, so a later read stops at the first finished node.
// With reference nil it asks about any reference: every step through a predicate value counts, and
// every path is followed.
func (f *predicateFalseFlow) search(flow *ast.FlowNode, reference *ast.Node, declared *checker.Type, key string) *predicateFalseStep {
	if step, done := f.steps[predicateFalseKey{flow: flow, reference: key, declared: declared}]; done {
		return step
	}
	walk := &predicateFalseSearch{flow: f, reference: reference, declared: declared, key: key, index: map[*ast.FlowNode]int{}, low: map[*ast.FlowNode]int{}, partial: map[*ast.FlowNode]*predicateFalseStep{}, stacked: map[*ast.FlowNode]bool{}}
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

func (s *predicateFalseSearch) memo(flow *ast.FlowNode) predicateFalseKey {
	return predicateFalseKey{flow: flow, reference: s.key, declared: s.declared}
}

func (s *predicateFalseSearch) visit(flow *ast.FlowNode) *predicateFalseStep {
	s.index[flow] = len(s.index)
	s.low[flow] = s.index[flow]
	s.stack = append(s.stack, flow)
	s.stacked[flow] = true
	step, next := s.step(flow)
	for _, antecedent := range next {
		if antecedent == nil {
			continue
		}
		if done, finished := s.flow.steps[s.memo(antecedent)]; finished {
			step = predicateFalseFirst(step, done)
			continue
		}
		if _, visited := s.index[antecedent]; !visited {
			step = predicateFalseFirst(step, s.visit(antecedent))
			s.low[flow] = min(s.low[flow], s.low[antecedent])
		} else if s.stacked[antecedent] {
			s.low[flow] = min(s.low[flow], s.index[antecedent])
		}
	}
	s.partial[flow] = step
	if s.low[flow] != s.index[flow] {
		return step
	}
	// flow roots a cycle, or stands alone: each member's answer is all of theirs together.
	start := len(s.stack) - 1
	for s.stack[start] != flow {
		start--
	}
	members := s.stack[start:]
	for _, member := range members {
		step = predicateFalseFirst(step, s.partial[member])
	}
	for _, member := range members {
		s.flow.steps[s.memo(member)] = step
		s.stacked[member] = false
	}
	s.stack = s.stack[:start]
	return step
}

// step is what flow itself says, and the antecedents a path back goes on to.
func (s *predicateFalseSearch) step(flow *ast.FlowNode) (*predicateFalseStep, []*ast.FlowNode) {
	switch {
	case flow.Flags&ast.FlowFlagsUnreachable != 0:
		return nil, nil
	case flow.Flags&ast.FlowFlagsStart != 0:
		// As the checker does, a name, or this inside an arrow, carries narrowing in from the
		// function around it, so a closure made after `if (!test(n)) return` reads the outer step.
		// A hoisted declaration's start has no node, and nothing reaches it from outside.
		reference := s.reference
		if flow.Node != nil && (reference == nil || reference.Kind == ast.KindIdentifier || (reference.Kind == ast.KindThisKeyword && ast.IsArrowFunction(flow.Node))) {
			if outer := flow.Node.FlowNodeData(); outer != nil {
				return nil, []*ast.FlowNode{outer.FlowNode}
			}
		}
		return nil, nil
	case flow.Flags&ast.FlowFlagsAssignment != 0:
		// An assignment to the reference, or to what it's read off, ends what narrowed it before.
		// A compound one reads what came before it.
		if s.reference != nil && (predicateMatchingReference(s.flow.l.checker, s.reference, flow.Node) && !predicateCompoundAssignment(flow.Node) || enumTagContainsReference(s.flow.l.checker, s.reference, flow.Node)) {
			return nil, nil
		}
		return nil, []*ast.FlowNode{flow.Antecedent}
	case flow.Flags&(ast.FlowFlagsCondition|ast.FlowFlagsSwitchClause) != 0:
		call := s.flow.condition(flow)
		if call != nil && s.narrows(flow, call) {
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
func (s *predicateFalseSearch) narrows(flow *ast.FlowNode, call *ast.Node) bool {
	switch {
	case s.reference == nil:
		return true
	case s.declared == nil:
		return predicateFalseReadsThis(call)
	}
	before := predicateFlowType(s.flow.l.checker, s.reference, s.declared, s.declared, nil, flow.Antecedent)
	after := predicateFlowType(s.flow.l.checker, s.reference, s.declared, s.declared, nil, flow)
	return !checker.Checker_isTypeIdenticalTo(s.flow.l.checker, before, after)
}

// condition is the call through a predicate value that flow's false side narrows by, or nil. A
// switch (true) passes by each clause before the ones reached on its false side, and each after
// them too when a default is among them, as the checker's narrowTypeBySwitchOnTrue does; the
// clauses reached narrow by their true side. Any other switch narrows by no call.
func (f *predicateFalseFlow) condition(flow *ast.FlowNode) *ast.Node {
	if call, done := f.calls[flow]; done {
		return call
	}
	var call *ast.Node
	if flow.Flags&ast.FlowFlagsCondition != 0 {
		truth := predicateFalse
		if flow.Flags&ast.FlowFlagsTrueCondition != 0 {
			truth = predicateTrue
		}
		call = f.conditionCall(flow.Node, truth, 0)
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
				if call == nil && clause.Kind == ast.KindCaseClause && (index < start || (fallsFromDefault && index >= end)) {
					call = f.conditionCall(clause.Expression(), predicateFalse, 0)
				}
			}
		}
	}
	f.calls[flow] = call
	return call
}

// conditionCall follows expr as the checker's narrowType does when it's assumed truth, to a call
// through a predicate value narrowed on its false side. It reads through more than the checker
// does, never less: both sides of every comparison, and a let alias as well as a const one, so a
// form the checker narrows through is never missed; the step's own type change then says whether
// the checker narrowed.
func (f *predicateFalseFlow) conditionCall(expr *ast.Node, truth predicateTruth, depth int) *ast.Node {
	// For a?.b and the left of ??, the checker narrows by presence only.
	if ast.IsExpressionOfOptionalChainRoot(expr) || (expr.Parent != nil && expr.Parent.Kind == ast.KindBinaryExpression && expr.Parent.AsBinaryExpression().Left == expr && (expr.Parent.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionToken || expr.Parent.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken)) {
		return nil
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		// The checker reads a condition alias through its initializer, up to five levels:
		// const ok = test(n); if (!ok).
		symbol := f.l.symbol(expr)
		if depth < 5 && symbol != nil && symbol.ValueDeclaration != nil && ast.IsVariableDeclaration(symbol.ValueDeclaration) && symbol.ValueDeclaration.Type() == nil && symbol.ValueDeclaration.Initializer() != nil {
			return f.conditionCall(symbol.ValueDeclaration.Initializer(), truth, depth+1)
		}
	case ast.KindCallExpression:
		return f.valueCall(expr, truth)
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
		return f.conditionCall(expr.Expression(), truth, depth)
	case ast.KindPrefixUnaryExpression:
		if expr.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
			flipped := truth
			if truth != predicateEither {
				flipped = predicateEither &^ truth
			}
			return f.conditionCall(expr.AsPrefixUnaryExpression().Operand, flipped, depth)
		}
	case ast.KindBinaryExpression:
		binary := expr.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindCommaToken:
			return f.conditionCall(binary.Right, truth, depth)
		case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			// A comparison with true or false narrows its other side either way round.
			if call := f.conditionCall(binary.Left, predicateEither, depth); call != nil {
				return call
			}
			return f.conditionCall(binary.Right, predicateEither, depth)
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			if call := f.conditionCall(binary.Left, truth, depth); call != nil {
				return call
			}
			return f.conditionCall(binary.Right, truth, depth)
		case ast.KindInstanceOfKeyword:
			// A [Symbol.hasInstance] predicate narrows as a call does.
			return f.valueCall(expr, truth)
		}
	}
	return nil
}

// valueCall is call when its false side narrows by a type predicate and its callee isn't a direct
// reference to the predicate's declaration. The checker reads an optional call's false side as
// no narrowing.
func (f *predicateFalseFlow) valueCall(call *ast.Node, truth predicateTruth) *ast.Node {
	if truth&predicateFalse == 0 || (call.Kind == ast.KindCallExpression && call.Flags&ast.NodeFlagsOptionalChain != 0) {
		return nil
	}
	signature := predicateEffectsSignature(f.l.checker, call)
	if signature == nil {
		return nil
	}
	predicate := predicateOfSignature(f.l.checker, signature)
	if predicate == nil || (predicate.Kind() != checker.TypePredicateKindIdentifier && predicate.Kind() != checker.TypePredicateKindThis) {
		return nil
	}
	if call.Kind == ast.KindCallExpression {
		declaration := signature.Declaration()
		callee := ast.SkipParentheses(call.AsCallExpression().Expression)
		if declaration != nil && declaration.Kind == ast.KindFunctionDeclaration && declaration.Name() != nil && ast.IsIdentifier(callee) && f.l.symbol(callee) == f.l.symbol(declaration.Name()) {
			return nil
		}
		if f.libraryCall(call, callee, declaration) {
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
func (f *predicateFalseFlow) libraryCall(call *ast.Node, callee *ast.Node, declaration *ast.Node) bool {
	if declaration == nil || !(load.IsLibrary(ast.GetSourceFileOfNode(declaration)) || load.IsPrelude(ast.GetSourceFileOfNode(declaration))) {
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
	symbol := f.l.checker.GetSymbolAtLocation(name)
	if symbol == nil || f.l.checker.GetMergedSymbol(f.l.checker.GetTargetSymbol(symbol)) != f.l.checker.GetSymbolOfDeclaration(declaration) {
		return false
	}
	for _, argument := range call.Arguments() {
		if f.predicateValued(f.l.checker.GetTypeAtLocation(argument)) {
			return false
		}
	}
	return true
}

// predicateValued is whether a value of type t can be called as a type predicate: some call
// signature of it, or of a union or intersection member, carries one.
func (f *predicateFalseFlow) predicateValued(t *checker.Type) bool {
	if t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range t.Types() {
			if f.predicateValued(member) {
				return true
			}
		}
	}
	for _, signature := range f.l.checker.GetSignaturesOfType(t, checker.SignatureKindCall) {
		if predicateOfSignature(f.l.checker, signature) != nil {
			return true
		}
	}
	return false
}

// predicateFalseFirst is the earlier of two steps in the source, so a read's answer doesn't depend
// on which read reached a node first.
func predicateFalseFirst(a, b *predicateFalseStep) *predicateFalseStep {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	position := func(step *predicateFalseStep) (string, int) {
		if step.call == nil {
			return "", -1
		}
		return string(ast.GetSourceFileOfNode(step.call).FileName()), step.call.Pos()
	}
	aFile, aPosition := position(a)
	bFile, bPosition := position(b)
	if bFile < aFile || (bFile == aFile && bPosition < aPosition) {
		return b
	}
	return a
}

// predicateFalseReferenceKey names reference so that two references the checker matches share
// their answers: the symbol of a name, and each property read off it. Any other shape is its own.
func predicateFalseReferenceKey(l *lowering, reference *ast.Node) (string, bool) {
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
			symbol := l.symbol(node)
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
			return object + "[" + argument.Text() + "]", ok && (argument.Kind == ast.KindStringLiteral || argument.Kind == ast.KindNumericLiteral)
		}
		return "", false
	}
	name, ok := key(reference)
	return name + suffix, ok
}

// predicateCompoundAssignment is whether target is written by a compound operator, which reads it.
func predicateCompoundAssignment(target *ast.Node) bool {
	parent := target.Parent
	return parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == target && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind) && parent.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken
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
