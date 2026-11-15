package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An overload whose declared result is narrower than its implementation's is checked at each call that
// resolves to it, in .ts, on exactly the narrowing it promises, and refused in .a (rulings.md:127, built
// for literal, tag and branded narrowing on #z5d0jye). censusOverload accepts the declaration only
// because every call is checked, so the check follows the promise, not the path: a call resolving to
// the narrow overload is checked even where the body would have produced the right value. The check is
// R-T80's true-side membership (predicateMembership) with the call's result as its argument: a literal
// is compared, a tag is read, and a tagged or branded member is read through its view from the call on,
// so a lying tag stops at its first field read wherever the value goes. A result no read reaches is
// erased and counted, never checked, as :146 and :137 read (Adamic's ruling on #z5d0jye): a lie nothing
// observes changes nothing a program can see. Every road a call takes goes through overloadResultCheck,
// and checkOverloadResults, after lowering, panics on any call to a narrow overload the lowering lowered
// that it never settled, so a road nobody listed can't carry the lie.

// narrowOverload is whether overload, a bodyless function declaration of a named overload set, is
// accepted at its declaration only because each call checks its result: in .ts, a result the
// implementation's doesn't serve by relation, by a phantom brand, by presence alone or by a predicate
// proof, but narrows to a checkable promise. It's a pure function of the declaration, memoized.
func (l *lowering) narrowOverload(overload *ast.Node) bool {
	if overload == nil || overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil || overload.Name() == nil {
		return false
	}
	if narrow, known := l.narrowOverloads[overload]; known {
		return narrow
	}
	narrow := l.narrowOverloadJudged(overload)
	if l.narrowOverloads == nil {
		l.narrowOverloads = map[*ast.Node]bool{}
	}
	l.narrowOverloads[overload] = narrow
	return narrow
}

func (l *lowering) narrowOverloadJudged(overload *ast.Node) bool {
	implementation := l.censusImplementation(overload)
	// A generic implementation's binders map to the overload's only by censusOverload's alpha-renaming,
	// so its narrowing isn't judged here: it stays refused.
	if implementation == nil || len(implementation.TypeParameters()) != 0 || !l.checkedAssertionSource(implementation) || !l.checkedAssertionSource(overload) {
		return false
	}
	// A predicate overload's claim is R-T80's, and an async or generator body's returns aren't the call's
	// result.
	if (overload.Type() != nil && overload.Type().Kind == ast.KindTypePredicate) || ast.HasSyntacticModifier(implementation, ast.ModifierFlagsAsync) || generatorDeclaration(implementation) {
		return false
	}
	outerMapper := l.typeMapper
	defer func() { l.typeMapper = outerMapper }()
	l.typeMapper = nil
	promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
	produced := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation))
	if l.censusRelated(produced, promised) || l.censusNullableOverloadResult(produced, promised) {
		return false
	}
	from, to := l.concrete(produced), l.concrete(promised)
	if (l.hasPhantom(from) || l.hasPhantom(to)) && l.phantomCast(from, to) && l.phantomAssignable(from, to) && l.phantomAssignable(to, from) {
		return false
	}
	if proven, _ := l.predicateOverloadResultProven(implementation, overload); proven {
		return false
	}
	// A narrowing, never a different result: every value the overload promises is one the implementation
	// can produce, and the promise is one the call can check.
	return l.checker.IsTypeAssignableTo(to, from) && l.overloadResultCheckable(to, map[*checker.Type]bool{})
}

// overloadResultCheckable is whether predicateMembership checks promised whole at a call: undefined, a
// string, number or boolean (a literal compared), an object read by its literal tag or through a branded
// supertype's view, a non-generic class instance, or a union of those. A type parameter is judged as its
// constraint at the declaration and as its instance at each call. An array, an intersection, an untagged
// interface and anything else stays refused (:126): a container's elements would be checked once while
// the result may alias a wider mutable view, and an untagged record has no failing field to check.
func (l *lowering) overloadResultCheckable(promised *checker.Type, visited map[*checker.Type]bool) bool {
	if promised == nil || visited[promised] {
		return false
	}
	visited[promised] = true
	defer delete(visited, promised)
	if promised.Flags()&checker.TypeFlagsTypeParameter != 0 {
		constraint := l.checker.GetBaseConstraintOfType(promised)
		return constraint != nil && constraint != promised && l.overloadResultCheckable(l.concrete(constraint), visited)
	}
	if promised.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range promised.Types() {
			if !l.overloadResultCheckable(member, visited) {
				return false
			}
		}
		return len(promised.Types()) > 0
	}
	if promised.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 {
		return true
	}
	if promised.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(promised) {
		return false
	}
	if isClassInstance(promised) {
		symbol := promised.Symbol()
		if symbol == nil {
			return false
		}
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindClassDeclaration && len(declaration.TypeParameters()) == 0 {
				return true
			}
		}
		return false
	}
	return l.fieldLiteral(promised, "kind") != nil || l.brandedSupertype(promised)
}

// narrowOverloadCall is whether node, a call, a new or a tagged template, resolves to a narrow overload.
func (l *lowering) narrowOverloadCall(node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression && node.Kind != ast.KindTaggedTemplateExpression {
		return false
	}
	resolved := l.checker.GetResolvedSignature(node)
	return resolved != nil && l.narrowOverload(resolved.Declaration())
}

// overloadResultCheck is value, the result of call, checked against the narrow overload the call
// resolves to: through a function that takes the result, panics when it isn't what the overload
// promises, and hands it back, so it's one road for a direct call, a closure call and an instance's.
// A call the body proves under the overload's parameters (censusProveOverloadResult) isn't checked, and
// a result no read reaches (overloadResultDiscarded) is erased, each counted. Anything but a call to a
// narrow overload is value itself.
func (l *lowering) overloadResultCheck(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if !l.narrowOverloadCall(node) {
		return value, nil
	}
	resolved := l.checker.GetResolvedSignature(node)
	overload := resolved.Declaration()
	implementation := l.censusImplementation(overload)
	if l.overloadCalls == nil {
		l.overloadCalls = map[*ast.Node]bool{}
	}
	l.overloadCalls[node] = true
	if l.censusProveOverloadResult(implementation, overload) {
		l.result.OverloadResultChecks.Proven++
		return value, nil
	}
	if l.overloadResultDiscarded(node) {
		l.result.OverloadResultChecks.Erased++
		return value, nil
	}
	if value.Type() == 0 {
		return nil, l.notYet(node, "a checked overload result with no value")
	}
	promised := l.concrete(l.checker.GetReturnTypeOfSignature(resolved))
	if !l.overloadResultCheckable(promised, map[*checker.Type]bool{}) {
		return nil, l.notYet(node, "an overload result whose instance "+l.checker.TypeToString(promised)+" can't be checked at the call")
	}
	produced := l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation)))
	ordinal := l.predicateOverloadOrdinal(implementation, overload)
	name := implementation.Name().Text()
	check := len(l.result.Functions)
	// Same-named overloads may share this IR label: native keys include the semantic body and resolved
	// callee, so the label doesn't encode encounter order.
	l.result.Functions = append(l.result.Functions, ir.Function{Name: fmt.Sprintf("%s_overload_check", name), Returns: value.Type()})
	local := func(name string, of ir.Type) int {
		index := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: check})
		return index
	}
	result := local("result", value.Type())
	l.result.Functions[check].Parameters = []int{result}
	read := ir.Read{Local: result, Of: value.Type()}
	setup, membership, err := l.predicateMembership(node, read, produced, promised, local, 0)
	if err != nil {
		return nil, err
	}
	// The check's call is where its value enters a view, for checkViewWidenedReads (view_widened.go): a
	// read through a wider type the result is assigned or passed to is weighed as the predicate road's is
	// (Language's q06, #z5d0jye).
	if l.overloadCheckCalls == nil {
		l.overloadCheckCalls = map[int]bool{}
	}
	l.overloadCheckCalls[check] = true
	message := fmt.Sprintf("overload %d of %s result at %s: expected %s", ordinal, name, l.program.Where(node), l.checker.TypeToString(promised))
	body := append(append([]ir.Statement{}, setup...), ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: membership}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}}, ir.Return{Value: read})
	l.result.Functions[check].Body = body
	l.result.OverloadResultChecks.Checked++
	return ir.Call{Function: check, Arguments: []ir.Expression{value}, Returns: value.Type()}, nil
}

// overloadResultDiscarded is whether no read reaches node's result: a call statement, the left side of
// a comma, a void operand, or an arrow's expression body when the arrow's own result is void. A
// conditional or any other parent reads it, so it's checked.
func (l *lowering) overloadResultDiscarded(node *ast.Node) bool {
	child, parent := node, node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child, parent = parent, parent.Parent
	}
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindExpressionStatement, ast.KindVoidExpression:
		return true
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindCommaToken && binary.Left == child
	case ast.KindArrowFunction:
		if parent.Body() != child {
			return false
		}
		returns := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(parent))
		return returns != nil && returns.Flags()&checker.TypeFlagsVoid != 0
	}
	return false
}

// noteOverloadCall records node, a call, a new or a tagged template the lowering is lowering, when it
// resolves to a narrow overload, so checkOverloadResults can ask after lowering whether it was settled.
// expression, callOrMethod, call, callFunction and callClosure note it, so a road entering at any of them
// is seen, a call statement's included.
func (l *lowering) noteOverloadCall(node *ast.Node) {
	if !l.narrowOverloadCall(node) {
		return
	}
	if l.overloadCalls == nil {
		l.overloadCalls = map[*ast.Node]bool{}
	}
	if _, settled := l.overloadCalls[node]; !settled {
		l.overloadCalls[node] = false
	}
}

// noteClassInitializer records only roots the constructor or static initialization road lowered.
// Generic classes keep the existing call notes because their meanings depend on instantiation.
func (l *lowering) noteClassInitializer(class, root *ast.Node) {
	if root == nil || len(class.TypeParameters()) != 0 {
		return
	}
	if l.classInitializers == nil {
		l.classInitializers = map[*ast.Node]map[*ast.Node]bool{}
	}
	if l.classInitializers[class] == nil {
		l.classInitializers[class] = map[*ast.Node]bool{}
	}
	l.classInitializers[class][root] = true
}

// walkLoweredRegions calls each with every node lowering actually lowered, once lowering is done: it
// walks module statements, bodies recorded in functionNodes, and the instance field expressions, static
// field expressions and static block bodies recorded in classInitializers when they actually lower. It
// never descends into nested functions or classes, each of which is its own region when it lowers.
// Unconstructed instance initializers are absent; generic class initialization stays on the call notes.
// A body lowering never lowers, such as an uncalled generic function's, is no region. checkOverloadResults
// and checkCheckedPredicateParameters each judge these regions, so the two can't walk apart.
func (l *lowering) walkLoweredRegions(modules []*ast.SourceFile, each func(node *ast.Node)) {
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) || ast.IsClassLike(node) {
			return false
		}
		each(node)
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		for _, statement := range module.Statements.Nodes {
			visit(statement)
		}
	}
	for _, declaration := range l.functionNodes {
		declaration.ForEachChild(visit)
	}
	for _, roots := range l.classInitializers {
		for root := range roots {
			visit(root)
		}
	}
}

// checkOverloadResults is the guard behind overloadResultCheck, run once lowering is done. Every narrow
// overload call in the lowered regions (walkLoweredRegions) must be settled (checked, proven or erased),
// even if a lowering road failed to note it. Missing or unsettled calls are compiler bugs because
// censusOverload accepted the promise only on that condition.
func (l *lowering) checkOverloadResults(modules []*ast.SourceFile) {
	l.walkLoweredRegions(modules, func(node *ast.Node) {
		if l.narrowOverloadCall(node) && !l.overloadCalls[node] {
			panic(fmt.Sprintf("lower: compiler bug: a call to an overload accepted narrow at %s lowered without its result check", l.program.Where(node)))
		}
	})
	for node, settled := range l.overloadCalls {
		if !settled {
			panic(fmt.Sprintf("lower: compiler bug: a call to an overload accepted narrow at %s lowered without its result check", l.program.Where(node)))
		}
	}
}
