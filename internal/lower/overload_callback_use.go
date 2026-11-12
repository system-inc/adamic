package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
	"os"
)

// An overload can promise its caller a narrower callback than its implementation takes, as tsc's debug.ts
// assertNode does: overload 1 takes `test: (node: T) => node is U`, the implementation takes
// `test: ((node: Node) => boolean) | undefined`. Callback parameters are strictly contravariant, so the
// census refuses that pair in general, since an implementation may call test with any Node, a lie to the
// caller's `(node: T) => ...`. But assertNode's body only ever calls test with node, and node holds exactly
// what the overload's caller passed, a T. So the census asks how the body uses the callback before it
// refuses (Adamic's ruling on #1w5bve2, block 3, Nov 3, and on #8n7eet3, Nov 10; its rules 1, 2, 4 and 5
// are built here, rule 3 isn't):
//
//   - Rule 2: a presence test, the callback compared by ===, !==, == or != with undefined or null.
//   - Rule 1: a direct call, whose every argument serves the parameter the overload promises at its
//     position. An argument naming one of the implementation's own parameters carries the overload's type
//     at that position, when the body never assigns that parameter and the checker's flow type there is
//     its declared type, or its declared type made non-nullable. An argument naming a const for-of binding
//     in the implementation's own body, whose loop runs over exactly such a parameter that the body reads
//     nowhere else but in presence tests, carries the element of the overload's type at that position,
//     under the same flow check on the binding.
//   - Rule 4: a use inside a nested function may only be a presence test, since a call there could run
//     after the body has moved on, with values the overload never promised.
//   - Rule 5: anything else, a pass to another function, a store, a return, .call, a spread, refuses.

// overloadCallbackUse is whether implementation's parameter actual serves overload's parameter at the same
// position though censusRelated(given, takes) refused it, because the body uses the callback only by rules 1,
// 2 and 4 above. given and takes arrive concrete from censusOverloadParameter, under the mapper censusOverload
// set for this overload, so each overload is judged on its own. Every precondition fails closed.
func (l *lowering) overloadCallbackUse(implementation, overload, actual *ast.Node, given, takes *checker.Type) bool {
	if os.Getenv("ADAMIC_PROBE_CALLBACK") != "" {
		fmt.Fprintf(os.Stderr, "PROBE use given=%s takes=%s concreteGiven=%s mapper=%v\n", l.checker.TypeToString(given), l.checker.TypeToString(takes), l.checker.TypeToString(l.concrete(given)), l.typeMapper != nil)
	}
	if !overloadCallbackPlainParameter(actual) || implementation.Body() == nil {
		return false
	}
	// Rule 1 reads the overload's parameter at an implementation parameter's index, and a this parameter on
	// either side shifts one list against the other, so neither may have one.
	if overloadThisParameter(implementation) || overloadThisParameter(overload) {
		return false
	}
	promisedSignatures := l.checker.GetSignaturesOfType(given, checker.SignatureKindCall)
	if len(promisedSignatures) != 1 || len(l.checker.GetSignaturesOfType(given, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	// The implementation's own presence tests are what let it take a missing callback, so undefined and
	// null are taken away before its one signature is read.
	acceptedSignatures := l.checker.GetSignaturesOfType(l.checker.GetNonNullableType(takes), checker.SignatureKindCall)
	if len(acceptedSignatures) != 1 {
		return false
	}
	promised, accepted := promisedSignatures[0], acceptedSignatures[0]
	// A this parameter would make the call's receiver part of the promise, and a direct call passes none.
	if promised.ThisParameter() != nil {
		return false
	}
	// The body reads the callback's answer as the implementation types it, so what the caller's callback
	// returns must serve that: a `node is U` predicate returns boolean, which fits boolean.
	if !l.censusRelated(l.checker.GetReturnTypeOfSignature(promised), l.checker.GetReturnTypeOfSignature(accepted)) {
		return false
	}
	// arguments reads every parameter without naming one, so no walk over names could follow it.
	if overloadCallbackReadsArguments(implementation) {
		return false
	}
	symbol := l.symbol(actual.Name())
	if symbol == nil {
		return false
	}
	// A default runs at the call with whatever the earlier parameters hold, so a use there is judged as
	// one inside a nested function: a presence test at most.
	for _, parameter := range implementation.Parameters() {
		if initializer := parameter.AsParameterDeclaration().Initializer; initializer != nil && !l.overloadCallbackUses(implementation, overload, actual, symbol, promised, initializer, true) {
			return false
		}
	}
	return l.overloadCallbackUses(implementation, overload, actual, symbol, promised, implementation.Body(), false)
}

// overloadCallbackPlainParameter is whether parameter is a plain name, with no default and no rest, so the
// value it holds on entry is the very argument the caller passed at its position.
func overloadCallbackPlainParameter(parameter *ast.Node) bool {
	declaration := parameter.AsParameterDeclaration()
	return ast.IsIdentifier(parameter.Name()) && declaration.Initializer == nil && declaration.DotDotDotToken == nil
}

// overloadCallbackReadsArguments is whether anything in implementation names arguments, nested functions
// and defaults included. It fails closed: a nested function's own arguments counts too, and so does a
// property declared with that name, since telling them apart buys nothing tsc's helpers need.
func overloadCallbackReadsArguments(implementation *ast.Node) bool {
	found := false
	var reads ast.Visitor
	reads = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && node.Text() == "arguments" && (node.Parent == nil || node.Parent.Kind != ast.KindPropertyAccessExpression || node.Parent.AsPropertyAccessExpression().Name() != node) {
			found = true
		}
		return found || node.ForEachChild(reads)
	}
	implementation.ForEachChild(reads)
	return found
}

// overloadCallbackUses is whether every reference to symbol, actual's, in root is a presence test (rule 2)
// or, outside a nested function, a direct call whose arguments serve promised (rule 1). nested says root is
// already inside a function, a class or a default, where only a presence test stands (rule 4). Anything
// else, the callback passed on, stored, returned, spread or read for a member, refuses (rule 5).
func (l *lowering) overloadCallbackUses(implementation, overload, actual *ast.Node, symbol *ast.Symbol, promised *checker.Signature, root *ast.Node, nested bool) bool {
	var holds func(node *ast.Node, nested bool) bool
	holds = func(node *ast.Node, nested bool) bool {
		if ast.IsPartOfTypeNode(node) {
			return true // A type query names the callback's type, never its value.
		}
		if ast.IsIdentifier(node) {
			// A shorthand property's name may resolve to the property rather than the binding it reads, so a
			// name matching the callback's is held to be it there: `{ test }` stores the callback.
			if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Parent.Name() == node && node.Text() == actual.Name().Text() {
				return false
			}
			if l.symbol(node) != symbol {
				return true
			}
			if l.overloadCallbackPresenceTest(node) {
				return true
			}
			if nested {
				return false
			}
			return l.overloadCallbackCall(implementation, overload, node, promised)
		}
		inner := nested || ast.IsFunctionLike(node) || ast.IsClassLike(node)
		every := true
		node.ForEachChild(func(child *ast.Node) bool {
			every = holds(child, inner)
			return !every
		})
		return every
	}
	return holds(root, nested)
}

// overloadCallbackPresenceTest is whether reference, through parentheses, is one operand of ===, !==, == or
// != whose other operand is undefined, the global's own symbol, or null. Comparing a function value with
// either never calls it, loose or strict.
func (l *lowering) overloadCallbackPresenceTest(reference *ast.Node) bool {
	operand := reference
	for operand.Parent != nil && operand.Parent.Kind == ast.KindParenthesizedExpression {
		operand = operand.Parent
	}
	if operand.Parent == nil || operand.Parent.Kind != ast.KindBinaryExpression {
		return false
	}
	comparison := operand.Parent.AsBinaryExpression()
	if !ast.IsEqualityOperator(comparison.OperatorToken.Kind) {
		return false
	}
	other := comparison.Left
	if comparison.Left == operand {
		other = comparison.Right
	} else if comparison.Right != operand {
		return false
	}
	other = ast.SkipParentheses(other)
	return other.Kind == ast.KindNullKeyword || (ast.IsIdentifier(other) && l.symbol(other) == l.checker.GetUndefinedSymbol())
}

// overloadCallbackCall is whether reference is the callee of a call, through parentheses or a non-null !,
// with no ?. and no spread, whose argument at each of promised's positions serves promised's parameter
// there (rule 1). A missing argument is undefined, which serves an optional parameter and nothing else. An
// argument past promised's parameters, with no rest there, refuses: the callee is the caller's function,
// and tsc lets one with a further optional parameter, `(dog: Dog, label?: string) => boolean`, stand
// where `(value: T) => boolean` is promised, so the extra lands in a parameter the promise never showed
// (Language's must on #8n7eet3). A rest in promised takes every argument from its position on, each judged
// against the rest's element there, and a missing rest item is just a shorter array.
func (l *lowering) overloadCallbackCall(implementation, overload, reference *ast.Node, promised *checker.Signature) bool {
	callee := reference
	for callee.Parent != nil && (callee.Parent.Kind == ast.KindParenthesizedExpression || callee.Parent.Kind == ast.KindNonNullExpression) {
		callee = callee.Parent
	}
	if callee.Parent == nil || callee.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := callee.Parent.AsCallExpression()
	if call.Expression != callee || call.QuestionDotToken != nil || callee.Parent.Flags&ast.NodeFlagsOptionalChain != 0 {
		return false
	}
	for _, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return false // A spread moves which value lands at which position.
		}
	}
	fixed := len(promised.Parameters())
	if promised.HasRestParameter() {
		fixed--
	}
	for position := 0; position < max(fixed, len(call.Arguments.Nodes)); position++ {
		if position >= fixed && !promised.HasRestParameter() {
			return false
		}
		argument := l.checker.GetUndefinedType()
		if position < len(call.Arguments.Nodes) {
			argument = l.overloadCallbackArgument(implementation, overload, call.Arguments.Nodes[position])
		}
		parameter := l.callableParameterTypeAt(promised, position)
		// A promised parameter that holds any promises nothing: tsc lets `(dog: Dog, label?: string)` stand
		// where `(value: T, ...rest: any[])` or `(value: any)` is promised, and every argument relates to any,
		// so the caller's function would get whatever the implementation passes (Compiler's road on #8n7eet3).
		if l.overloadCallbackHoldsAny(parameter) || !l.censusRelated(argument, parameter) {
			return false
		}
	}
	return true
}

// overloadCallbackArgument is the type argument carries into the callback (rule 1's argument type). One
// naming an implementation parameter p carries the overload's type at p's position, as the overload's caller
// passed it, when all of these hold: p is a plain name, with no default and no rest; overload declares that
// position with a parameter that isn't a rest; the body never assigns p; and the checker's flow type at the
// argument is p's declared type, or that made non-nullable, in which case the overload's type is made
// non-nullable too. Any other narrowing, by a predicate, a typeof or a copy into a local, keeps the flow
// type, since the overload's type says nothing about what that narrowing kept. One naming a const for-of
// binding is overloadCallbackElement's.
func (l *lowering) overloadCallbackArgument(implementation, overload, argument *ast.Node) *checker.Type {
	flow := l.concrete(l.checker.GetTypeAtLocation(argument))
	name := ast.SkipParentheses(argument)
	if !ast.IsIdentifier(name) {
		return flow
	}
	symbol := l.symbol(name)
	if symbol == nil {
		return flow
	}
	for position, parameter := range implementation.Parameters() {
		if !ast.IsIdentifier(parameter.Name()) || l.symbol(parameter.Name()) != symbol {
			continue
		}
		if !overloadCallbackPlainParameter(parameter) || position >= len(overload.Parameters()) {
			return flow
		}
		if overload.Parameters()[position].AsParameterDeclaration().DotDotDotToken != nil || l.overloadCallbackAssigned(implementation, symbol) {
			return flow
		}
		admitted := l.overloadCallbackAdmitted(overload, position)
		// An overload parameter that holds any admits whatever its caller passes, so it promises nothing
		// about what the body hands the callback from it.
		if l.overloadCallbackHoldsAny(admitted) {
			return flow
		}
		own := l.checker.GetTypeAtLocation(parameter.Name())
		switch flow {
		case l.concrete(own):
			return admitted
		case l.concrete(l.checker.GetNonNullableType(own)):
			return l.checker.GetNonNullableType(admitted)
		}
		return flow
	}
	return l.overloadCallbackElement(implementation, overload, symbol, flow)
}

// overloadCallbackAdmitted is overload's type at position, concrete, with its own optionality:
// censusCallableParameterType adds undefined for a default, and an optional `?` adds it here.
func (l *lowering) overloadCallbackAdmitted(overload *ast.Node, position int) *checker.Type {
	declared := overload.Parameters()[position]
	admitted := l.checker.GetTypeAtLocation(declared)
	if ast.IsIdentifier(declared.Name()) {
		admitted = l.censusCallableParameterType(l.symbol(declared.Name()))
	}
	if declared.AsParameterDeclaration().QuestionToken != nil {
		admitted = l.checker.GetUnionType([]*checker.Type{admitted, l.checker.GetUndefinedType()})
	}
	return l.concrete(admitted)
}

// overloadCallbackElement is the type an argument naming symbol, no implementation parameter, carries: its
// flow type, flow, unless symbol is a const for-of binding over an implementation parameter p, as tsc's
// adapted debug.ts assertEachNode loops over nodes (Adamic's ruling on #8n7eet3, Nov 10). That binding
// carries the element of the overload's type at p's position, as the overload's caller passed the array,
// when all of these hold: the binding is one name with no initializer, in a for-of, not a for await,
// whose head is its declaration list, inside the implementation itself and not a nested function; the loop's
// expression, through parentheses, is exactly p; p could stand for the overload's type as an argument
// itself, a plain name at a position overload declares and not as a rest, never assigned; the body reads p
// nowhere but that loop's expression and presence tests (overloadCallbackOnlyIterated); and the checker's
// flow type at the argument is the binding's declared type, or that made non-nullable, in which case the
// element is made non-nullable too. Any other narrowing keeps the flow type, as for p itself.
func (l *lowering) overloadCallbackElement(implementation, overload *ast.Node, symbol *ast.Symbol, flow *checker.Type) *checker.Type {
	declaration := symbol.ValueDeclaration
	// A let could be given another value inside the loop, and a pattern or an initializer hands the name
	// something other than the element.
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(declaration.Name()) || declaration.AsVariableDeclaration().Initializer != nil || !ast.IsVarConst(declaration) {
		return flow
	}
	if declaration.Parent == nil || declaration.Parent.Kind != ast.KindVariableDeclarationList || len(declaration.Parent.AsVariableDeclarationList().Declarations.Nodes) != 1 {
		return flow
	}
	loop := declaration.Parent.Parent
	if loop == nil || loop.Kind != ast.KindForOfStatement {
		return flow
	}
	statement := loop.AsForInOrOfStatement()
	// for await hands the binding each element awaited, another value when an element is thenable, and a
	// loop inside a nested function could run after the body has moved on (rule 4).
	if statement.Initializer != declaration.Parent || statement.AwaitModifier != nil || ast.GetContainingFunction(loop) != implementation {
		return flow
	}
	source := ast.SkipParentheses(statement.Expression)
	if !ast.IsIdentifier(source) {
		return flow
	}
	iterated := l.symbol(source)
	if iterated == nil {
		return flow
	}
	for position, parameter := range implementation.Parameters() {
		if !ast.IsIdentifier(parameter.Name()) || l.symbol(parameter.Name()) != iterated {
			continue
		}
		if !overloadCallbackPlainParameter(parameter) || position >= len(overload.Parameters()) {
			return flow
		}
		if overload.Parameters()[position].AsParameterDeclaration().DotDotDotToken != nil || l.overloadCallbackAssigned(implementation, iterated) || !l.overloadCallbackOnlyIterated(implementation, parameter, iterated, source) {
			return flow
		}
		// An array's number index is the element its iterator yields. A type with no number index, as a
		// Set or a Map has none, fails closed. admitted is concrete already, so its element is too.
		admitted := l.overloadCallbackAdmitted(overload, position)
		element := l.checker.GetIndexTypeOfType(l.checker.GetNonNullableType(admitted), l.checker.GetNumberType())
		// As for p itself, an overload type that holds any promises nothing about its elements.
		if element == nil || l.overloadCallbackHoldsAny(admitted) {
			return flow
		}
		own := l.checker.GetTypeAtLocation(declaration.Name())
		switch flow {
		case l.concrete(own):
			return element
		case l.concrete(l.checker.GetNonNullableType(own)):
			return l.checker.GetNonNullableType(element)
		}
		return flow
	}
	return flow
}

// overloadCallbackAssigned is whether anything in implementation, its defaults and nested functions
// included, assigns symbol by any road. ast.IsAssignmentTarget follows =, compound and logical assignments,
// ++ and --, destructuring targets and for-in and for-of heads through parentheses and non-null, but stops
// at a type assertion, and `(node as N) = other` writes node too, so each wrapper is tried in turn.
func (l *lowering) overloadCallbackAssigned(implementation *ast.Node, symbol *ast.Symbol) bool {
	assigned := false
	var writes ast.Visitor
	writes = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && l.symbol(node) == symbol {
			for target := node; target.Parent != nil; target = target.Parent {
				if ast.IsAssignmentTarget(target) {
					assigned = true
					break
				}
				switch target.Parent.Kind {
				case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
					continue
				}
				break
			}
		}
		return assigned || node.ForEachChild(writes)
	}
	implementation.ForEachChild(writes)
	return assigned
}

// overloadCallbackOnlyIterated is whether every reference to symbol, parameter's, in implementation, its
// defaults and nested functions included, is source, the one for-of expression overloadCallbackElement
// follows, or a presence test (overloadCallbackPresenceTest). Any other reference could change the
// caller's array or let something else change it, an index or member write, a push or a splice, a pass to
// a call, an alias, a spread or a capture, so it refuses wherever it stands: one allowlist, stricter than
// no write inside the loop (Adamic's word on #8n7eet3). arguments, which reads every parameter without
// naming one, overloadCallbackUse has already refused.
func (l *lowering) overloadCallbackOnlyIterated(implementation, parameter *ast.Node, symbol *ast.Symbol, source *ast.Node) bool {
	other := false
	var reads ast.Visitor
	reads = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false // A type query names the array's type, never its value.
		}
		if ast.IsIdentifier(node) {
			// As in overloadCallbackUses, a shorthand property's name that matches the parameter's is held
			// to be it: `{ nodes }` stores the array.
			if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Parent.Name() == node && node.Text() == parameter.Name().Text() {
				other = true
			} else if l.symbol(node) == symbol && node != parameter.Name() && node != source && !l.overloadCallbackPresenceTest(node) {
				other = true
			}
			return other
		}
		return other || node.ForEachChild(reads)
	}
	implementation.ForEachChild(reads)
	return !other
}

// overloadCallbackAnyCap bounds how many distinct types one overloadCallbackHoldsAny walk reads. A program has
// finitely many declared types, but a generic that holds a deeper instance of itself, `next?: Grow<Grow<T>>`,
// makes a new one at every step, and the cap is what ends that walk, failing closed. The walk once stopped
// at 32 positions deep instead, and tsc's Node graph (parent, symbol, declarations, emitNode) runs deeper
// than that with no any in it: Correctness's census on #68x1t2y (typescript/overload-proof-census 2378688b,
// debug-assert-overloads) read 1,896 types from destructuring.ts's array argument, every one any-free, and the
// bound alone refused assertEachNode's and assertNode's five units.
const overloadCallbackAnyCap = 32

// overloadCallbackHoldsAny is whether any appears anywhere in t a value can sit: t itself, a union or
// intersection member, a type argument, a property, an index signature's value, or a signature's
// parameters or result. It walks what t reaches with no depth bound, each type once, and answers yes past
// overloadCallbackAnyCap distinct types, failing closed. l.anyHolders memoizes it across the program. A
// walk that found any records yes for t alone, since the other types it read may reach no any at all. A walk
// that finished without one records no for every type it read, since what each of those reaches it read
// too. A cap is never recorded: how many types a walk reads depends on what earlier walks recorded.
func (l *lowering) overloadCallbackHoldsAny(t *checker.Type) bool {
	if t == nil {
		return false
	}
	if held, known := l.anyHolders[t]; known {
		return held
	}
	if l.anyHolders == nil {
		l.anyHolders = map[*checker.Type]bool{}
	}
	seen := map[*checker.Type]bool{t: true}
	pending := []*checker.Type{t}
	reach := func(next *checker.Type) {
		if next != nil && !seen[next] {
			seen[next] = true
			pending = append(pending, next)
		}
	}
	for len(pending) > 0 {
		if len(seen) > overloadCallbackAnyCap {
			return true
		}
		current := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		held, known := l.anyHolders[current]
		if known && !held {
			continue // Everything current reaches was read already, with no any in it.
		}
		if (known && held) || current.Flags()&checker.TypeFlagsAny != 0 {
			l.anyHolders[t] = true
			return true
		}
		l.overloadCallbackAnyReach(current, reach)
	}
	for read := range seen {
		l.anyHolders[read] = false
	}
	return false
}

// overloadCallbackAnyReach calls reach with each type a value of t can hold one step in, as
// overloadCallbackHoldsAny reads them: a union or intersection's members; a type reference's type
// arguments; and, unless the library declares t, its signatures' parameters and results, its properties
// and its index signatures' values.
func (l *lowering) overloadCallbackAnyReach(t *checker.Type, reach func(*checker.Type)) {
	if t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range t.Types() {
			reach(member)
		}
		return
	}
	if t.Flags()&checker.TypeFlagsObject == 0 {
		return
	}
	if t.ObjectFlags()&checker.ObjectFlagsReference != 0 && t.Target() != nil {
		for _, argument := range l.checker.GetTypeArguments(t) {
			reach(argument)
		}
	}
	// A library type's members are the library's, not values the type holds: ReadonlyArray's every and
	// forEach take `thisArg?: any`, so walking them would find any in every array. Its type arguments,
	// reached above, are what a value of it holds.
	if symbol := t.Symbol(); symbol != nil && len(symbol.Declarations) > 0 {
		library := true
		for _, declaration := range symbol.Declarations {
			file := ast.GetSourceFileOfNode(declaration)
			if file == nil || !(load.IsLibrary(file) || load.IsPrelude(file)) {
				library = false
			}
		}
		if library {
			return
		}
	}
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		for _, signature := range l.checker.GetSignaturesOfType(t, kind) {
			reach(l.checker.GetReturnTypeOfSignature(signature))
			for _, parameter := range signature.Parameters() {
				reach(l.checker.GetTypeOfSymbol(parameter))
			}
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(t) {
		reach(l.checker.GetTypeOfSymbol(property))
	}
	for _, index := range l.checker.GetIndexInfosOfType(t) {
		reach(index.ValueType())
	}
}
