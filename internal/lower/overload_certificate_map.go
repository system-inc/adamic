package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Certificate B (Correctness's rule B, under Adamic's ruling on #5fws6q8): an overload whose result is
// narrower than its implementation's is served when the implementation returns a fresh Map it built only
// by setting, in one index loop over an array parameter, each element's key, a key parameter's answer
// when it isn't undefined, to the element mapped by a mapper parameter. tsc's core.ts arrayToMap is the
// shape:
//
//	const result = new Map<K, V1 | V2>();
//	for (let i = 0; i < array.length; i++) {
//	    const value = array[i]!;
//	    const key = makeKey(value);
//	    if (key !== undefined) result.set(key, makeValue(value));
//	}
//	return result;
//
// Each set is judged against the Map the overload promises, by the overload's own parameter types, which
// are what the caller passed at a direct call the checker resolved to it: the key against its K, the
// value against its V. A mapper the overload requires and types without undefined is the caller's
// function, so the mapper's default can't run and the value is that function's result. A mapper the
// overload omits, or lets the caller pass as undefined, runs the default, which must be proven the
// identity, so the value is the element (map-default-undefined). The premise holds only at a direct call,
// so B rides certificate A's per-call gate (overloadCertificateCall), never a second one.

// overloadMapShape is what certificate B reads from an implementation's body.
type overloadMapShape struct {
	// local is the fresh Map the body returns.
	local *ast.Symbol
	// array, key and mapper are the indexes of the parameters the loop reads, the key is made by, and the
	// value is made by.
	array, key, mapper int
}

// overloadMapShapeOf is the shape of implementation's body when it's certificate B's: a const local
// initialized to the library's Map constructed with no entries first, its bare return last, and between
// them exactly one for statement, the canonical index loop over a parameter, whose body is exactly the
// element's checked read, its key, and the set the key's undefined test guards. Other statements between
// may run, as under certificate A, since nothing they do can reach the local (overloadLocalConfined). No
// parameter has a default but the mapper, whose default overloadMapCertificate discharges.
func (l *lowering) overloadMapShapeOf(implementation *ast.Node) (overloadMapShape, bool) {
	shape := overloadMapShape{array: -1, key: -1, mapper: -1}
	if !overloadCertificateFunction(implementation) {
		return shape, false
	}
	parameters := implementation.Parameters()
	for _, parameter := range parameters {
		if !ast.IsIdentifier(parameter.Name()) || parameter.AsParameterDeclaration().DotDotDotToken != nil {
			return shape, false
		}
	}
	statements := implementation.Body().AsBlock().Statements.Nodes

	// The fresh Map: a const, so nothing rebinds it, and the library's Map constructed with no entries, so
	// it starts holding nothing. A program's own Map, shadowing the global, isn't the library's.
	name, last := l.overloadFreshLocal(statements, func(initializer *ast.Node) bool {
		return initializer.Kind == ast.KindNewExpression && l.isLibraryGlobal(initializer.AsNewExpression().Expression, "Map") && len(nodesOf(initializer.AsNewExpression().Arguments)) == 0
	})
	if name == nil {
		return shape, false
	}
	shape.local = l.symbol(name)

	// The one loop, for (let i = 0; i < array.length; i++), over a parameter as the caller passed it. Its
	// bound and step are its grammar, not evidence: whatever index it reads at, the element is one of the
	// parameter's, through the checked read below.
	loop := overloadOnlyLoop(statements, ast.KindForStatement)
	if loop == nil {
		return shape, false
	}
	iteration := loop.AsForStatement()
	if iteration.Initializer == nil || iteration.Initializer.Kind != ast.KindVariableDeclarationList || iteration.Initializer.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsLet {
		return shape, false
	}
	counters := iteration.Initializer.AsVariableDeclarationList().Declarations.Nodes
	if len(counters) != 1 || !ast.IsIdentifier(counters[0].Name()) || counters[0].Initializer() == nil || counters[0].Initializer().Kind != ast.KindNumericLiteral || counters[0].Initializer().Text() != "0" {
		return shape, false
	}
	index := l.symbol(counters[0].Name())
	counter := func(node *ast.Node) bool {
		node = ast.SkipParentheses(node)
		return ast.IsIdentifier(node) && l.symbol(node) == index
	}
	condition := iteration.Condition
	if condition == nil || condition.Kind != ast.KindBinaryExpression || condition.AsBinaryExpression().OperatorToken.Kind != ast.KindLessThanToken || !counter(condition.AsBinaryExpression().Left) {
		return shape, false
	}
	bound := ast.SkipParentheses(condition.AsBinaryExpression().Right)
	if bound.Kind != ast.KindPropertyAccessExpression || bound.AsPropertyAccessExpression().QuestionDotToken != nil || bound.Name().Text() != "length" {
		return shape, false
	}
	shape.array = l.overloadParameterIndex(parameters, bound.Expression())
	stepped := false
	switch step := iteration.Incrementor; {
	case step == nil:
	case step.Kind == ast.KindPostfixUnaryExpression:
		stepped = step.AsPostfixUnaryExpression().Operator == ast.KindPlusPlusToken && counter(step.AsPostfixUnaryExpression().Operand)
	case step.Kind == ast.KindPrefixUnaryExpression:
		stepped = step.AsPrefixUnaryExpression().Operator == ast.KindPlusPlusToken && counter(step.AsPrefixUnaryExpression().Operand)
	}
	if shape.array < 0 || !stepped || iteration.Statement.Kind != ast.KindBlock {
		return shape, false
	}
	body := iteration.Statement.AsBlock().Statements.Nodes
	if len(body) != 3 {
		return shape, false
	}

	// const value = array[i]!: the element through lowering's checked read (nonNull), which panics on a
	// missing one, so value is always one of the array's elements.
	read := overloadConstDeclaration(body[0])
	if read == nil || read.Initializer() == nil || read.Initializer().Kind != ast.KindNonNullExpression {
		return shape, false
	}
	access := ast.SkipParentheses(read.Initializer().AsNonNullExpression().Expression)
	if access.Kind != ast.KindElementAccessExpression || access.AsElementAccessExpression().QuestionDotToken != nil || l.overloadParameterIndex(parameters, access.Expression()) != shape.array || !counter(access.AsElementAccessExpression().ArgumentExpression) {
		return shape, false
	}
	element := l.symbol(read.Name())
	isElement := func(node *ast.Node) bool {
		node = ast.SkipParentheses(node)
		return ast.IsIdentifier(node) && l.symbol(node) == element
	}

	// const key = makeKey(value): the key parameter's answer for the element, a const nothing rebinds.
	made := overloadConstDeclaration(body[1])
	if made == nil || made.Initializer() == nil {
		return shape, false
	}
	shape.key = l.overloadParameterCall(parameters, made.Initializer(), isElement)
	if shape.key < 0 || shape.key == shape.array {
		return shape, false
	}
	key := l.symbol(made.Name())
	isKey := func(node *ast.Node) bool {
		node = ast.SkipParentheses(node)
		return ast.IsIdentifier(node) && l.symbol(node) == key
	}

	// if (key !== undefined) result.set(key, makeValue(value)): the global undefined, by its symbol, so the
	// set's key is never undefined, and the set the Map prototype's own, of exactly that key and the
	// mapper's answer for the element. Nothing else is read, since a property read can run a getter that
	// answers differently each time (map-getter).
	if body[2].Kind != ast.KindIfStatement || body[2].AsIfStatement().ElseStatement != nil {
		return shape, false
	}
	test := ast.SkipParentheses(body[2].AsIfStatement().Expression)
	if test.Kind != ast.KindBinaryExpression || test.AsBinaryExpression().OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken || !isKey(test.AsBinaryExpression().Left) {
		return shape, false
	}
	absent := ast.SkipParentheses(test.AsBinaryExpression().Right)
	if !ast.IsIdentifier(absent) || l.symbol(absent) != l.checker.GetUndefinedSymbol() {
		return shape, false
	}
	set := body[2].AsIfStatement().ThenStatement
	if set.Kind == ast.KindBlock && len(set.AsBlock().Statements.Nodes) == 1 {
		set = set.AsBlock().Statements.Nodes[0]
	}
	receiver, arguments := l.overloadMutatorCall(set, shape.local, "Map", "set", 2)
	if receiver == nil || !isKey(arguments[0]) {
		return shape, false
	}
	shape.mapper = l.overloadParameterCall(parameters, arguments[1], isElement)
	if shape.mapper < 0 || shape.mapper == shape.array || shape.mapper == shape.key {
		return shape, false
	}
	for position, parameter := range parameters {
		if parameter.AsParameterDeclaration().Initializer != nil && position != shape.mapper {
			return shape, false // A default other than the mapper's is one no proof here discharges.
		}
	}

	// Every reference to the local is the set's receiver or the return, and no parameter is written.
	return shape, l.overloadLocalConfined(implementation, name, last, shape.local, map[*ast.Node]bool{receiver: true})
}

// overloadMapCertificate is whether certificate B serves overload's result: implementation's body has its
// shape, and each set the loop can make under this overload's arguments puts a key and a value the Map the
// overload promises holds. The types read are the overload's own, its parameters' and its result's, the
// caller's arguments at a direct call, so no binder of the implementation's is read and an overload with
// fewer binders than the implementation, tsc's arrayToMap overload 4 among them, is judged as any other.
// censusOverload asks it under its validated witness for the implementation's binders, and that the
// overload's parameters serve the implementation's under it, the key and mapper taking the element among
// them, is the parameter census's, which runs first (Language's q02 on #5fws6q8).
func (l *lowering) overloadMapCertificate(implementation, overload *ast.Node) bool {
	shape, shaped := l.overloadMapShapeOf(implementation)
	if !shaped {
		return false
	}
	declared := overload.Parameters()
	if len(declared) > len(implementation.Parameters()) || shape.array >= len(declared) || shape.key >= len(declared) {
		return false
	}
	if overloadThisParameter(implementation) || overloadThisParameter(overload) {
		return false // A this parameter shifts one list's positions against the other's.
	}
	for _, parameter := range declared {
		if parameter.AsParameterDeclaration().DotDotDotToken != nil || !ast.IsIdentifier(parameter.Name()) {
			return false // A rest parameter takes any count of arguments, so a position says nothing, and a pattern has no symbol to type.
		}
	}
	promised := l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload)))
	held := l.typeArguments(promised)
	if !l.isLibraryType(promised, "Map") || len(held) != 2 {
		return false
	}
	declaredType := func(position int) *checker.Type {
		return l.concrete(l.censusCallableParameterType(l.symbol(declared[position].Name())))
	}
	array := declaredType(shape.array)
	if !l.isArrayType(array) {
		return false
	}
	// The key is the caller's makeKey's answer with undefined taken out by the guard, and nothing else.
	keys, keyed := l.overloadCallableResult(declaredType(shape.key))
	if !keyed || !l.censusRelated(l.checker.GetUnionType(l.definedMembers(keys)), held[0]) {
		return false
	}
	defaults := shape.mapper >= len(declared)
	if !defaults {
		mapper := declaredType(shape.mapper)
		// The caller can pass undefined where the overload's type admits it, and undefined runs the default
		// as an omitted mapper does, so a present position alone never rules the default out
		// (map-default-undefined). The mapped overloads require a function: their default can't run.
		defaults = ast.HasQuestionToken(declared[shape.mapper]) || l.censusHasUndefined(mapper)
		values, mapped := l.overloadCallableResult(l.checker.GetUnionType(l.definedMembers(mapper)))
		if !mapped || !l.censusRelated(values, held[1]) {
			return false
		}
	}
	if defaults {
		initializer := implementation.Parameters()[shape.mapper].AsParameterDeclaration().Initializer
		if initializer == nil || !l.overloadIdentityDefault(initializer) {
			return false // The default runs, and only the identity's answer is the element.
		}
		if !l.censusRelated(l.elementTypeOfArray(array), held[1]) {
			return false
		}
	}
	return true
}

// overloadIdentityDefault is whether initializer, the mapper's default, answers what it's called with: a
// name that resolves, through an import, to a function declaration that's its symbol's only declaration,
// not async or a generator, taking one plain parameter, whose body is only the return of that parameter.
// A function declaration's binding is never written in a program that lowers: tsc refuses the write
// (2630) and lowering refuses it where it's written (assignments.go), so at every call the default is
// that body. Any other default is refused, however it reads: a variable holding a function can be rebound
// after the set is judged, and the name identity alone proves nothing.
func (l *lowering) overloadIdentityDefault(initializer *ast.Node) bool {
	initializer = ast.SkipParentheses(initializer)
	if !ast.IsIdentifier(initializer) {
		return false
	}
	symbol := l.symbol(initializer)
	if symbol == nil || len(symbol.Declarations) != 1 || !overloadCertificateFunction(symbol.Declarations[0]) {
		return false
	}
	identity := symbol.Declarations[0]
	parameters := identity.Parameters()
	if len(parameters) != 1 || !ast.IsIdentifier(parameters[0].Name()) || parameters[0].AsParameterDeclaration().Initializer != nil || parameters[0].AsParameterDeclaration().DotDotDotToken != nil {
		return false
	}
	statements := identity.Body().AsBlock().Statements.Nodes
	if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement || statements[0].AsReturnStatement().Expression == nil {
		return false
	}
	returned := ast.SkipParentheses(statements[0].AsReturnStatement().Expression)
	return ast.IsIdentifier(returned) && l.symbol(returned) == l.symbol(parameters[0].Name())
}
