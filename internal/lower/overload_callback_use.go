package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
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
//   - Rule 1 also follows #d4ja659's numeric element reads and single body forwards, including
//     Compiler's 16:51:03Z const-local widening. Adamic's 18:04:19Z ruling widens presence and
//     the confined array uses below. Under docs/rulings.md:133, every call passes an element of
//     the array, and a use that calls nothing and writes nothing can't break that.
//   - Rule 4: a use inside a nested function may only be a presence test, since a call there could run
//     after the body has moved on, with values the overload never promised.
//   - Rule 5: an unproved pass, a store, a return of the callback, .call, or a spread refuses.

// overloadCallbackUse is whether implementation's parameter actual serves overload's parameter at the same
// position though censusRelated(given, takes) refused it, because the body uses the callback only by rules 1,
// 2 and 4 above. given and takes arrive concrete from censusOverloadParameter, under the mapper censusOverload
// set for this overload, so each overload is judged on its own. Every precondition fails closed.
func (l *lowering) overloadCallbackUse(implementation, overload, actual *ast.Node, given, takes *checker.Type, label string, parameter *ast.Node) bool {
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
	legacySignature := len(acceptedSignatures) == 1
	if !legacySignature && takes.Flags()&checker.TypeFlagsUnion != 0 {
		acceptedSignatures = nil
		for _, member := range takes.Types() {
			acceptedSignatures = append(acceptedSignatures, l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)...)
		}
	}
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
	legacy := legacySignature
	for _, parameter := range implementation.Parameters() {
		if initializer := parameter.Initializer(); initializer != nil &&
			!l.overloadCallbackUses(implementation, overload, actual, symbol, promised, initializer, true) {
			legacy = false
			break
		}
	}
	if legacy && l.overloadCallbackUses(implementation, overload, actual, symbol, promised, implementation.Body(), false) {
		return true
	}
	frame := overloadCallbackFrame{function: implementation, callback: symbol, arrays: map[*ast.Symbol]*checker.Type{}}
	for position, own := range implementation.Parameters() {
		if !overloadCallbackPlainParameter(own) || position >= len(overload.Parameters()) ||
			overload.Parameters()[position].AsParameterDeclaration().DotDotDotToken != nil {
			continue
		}
		admitted := l.overloadCallbackAdmitted(overload, position)
		if l.overloadCallbackArray(admitted) && l.overloadCallbackArray(l.checker.GetTypeAtLocation(own.Name())) &&
			l.checker.GetIndexTypeOfType(l.checker.GetNonNullableType(admitted), l.checker.GetNumberType()) != nil &&
			!l.overloadCallbackHoldsAny(admitted, 0, map[*checker.Type]bool{}) {
			frame.arrays[l.valueSymbol(own.Name())] = admitted
		}
	}
	proof := overloadCallbackProof{l: l, root: overload, promised: promised, visiting: map[*ast.Node]bool{}}
	if !proof.body(frame) || !proof.moved {
		return false
	}
	if len(proof.reads) != 0 {
		if !ast.IsIdentifier(parameter.Name()) {
			return false
		}
		if l.callbackReadProofs == nil {
			l.callbackReadProofs = map[*ast.Node][]overloadCallbackReadProof{}
		}
		refusal := &Refused{Where: l.program.Where(parameter),
			What: label + " parameter " + parameter.Name().Text() + " cannot be served by implementation parameter " + actual.Name().Text(),
			Fix:  "make the implementation accept every value admitted by this overload, without mutable widening or bivariance"}
		l.callbackReadProofs[overload] = append(l.callbackReadProofs[overload], overloadCallbackReadProof{proof.reads, refusal})
	}
	return true
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
			return l.overloadCallbackCall(implementation, overload, node, promised, nil)
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
func (l *lowering) overloadCallbackCall(implementation, overload, reference *ast.Node, promised *checker.Signature, carried func(*ast.Node, *checker.Type) *checker.Type) bool {
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
		parameter := l.callableParameterTypeAt(promised, position)
		argument := l.checker.GetUndefinedType()
		if position < len(call.Arguments.Nodes) {
			node := call.Arguments.Nodes[position]
			if carried == nil {
				argument = l.overloadCallbackArgument(implementation, overload, node)
			} else {
				argument = carried(node, parameter)
			}
		}
		// A promised parameter that holds any promises nothing: tsc lets `(dog: Dog, label?: string)` stand
		// where `(value: T, ...rest: any[])` or `(value: any)` is promised, and every argument relates to any,
		// so the caller's function would get whatever the implementation passes (Compiler's road on #8n7eet3).
		if l.overloadCallbackHoldsAny(parameter, 0, map[*checker.Type]bool{}) || !l.censusRelated(argument, parameter) {
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
		if l.overloadCallbackHoldsAny(admitted, 0, map[*checker.Type]bool{}) {
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
		if element == nil || l.overloadCallbackHoldsAny(admitted, 0, map[*checker.Type]bool{}) {
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
		if ast.IsIdentifier(node) && l.valueSymbol(node) == symbol {
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

// overloadCallbackHoldsAny is whether any appears anywhere in t a value can sit: t itself, a union or
// intersection member, a type argument, a property, an index signature's value, or a signature's
// parameters or result. It walks as voidAnywhere does, and past its depth bound it fails closed.
func (l *lowering) overloadCallbackHoldsAny(t *checker.Type, depth int, seen map[*checker.Type]bool) bool {
	if t == nil || seen[t] {
		return false
	}
	if t.Flags()&checker.TypeFlagsAny != 0 || depth >= 32 {
		return true
	}
	seen[t] = true
	if t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range t.Types() {
			if l.overloadCallbackHoldsAny(member, depth+1, seen) {
				return true
			}
		}
		return false
	}
	if t.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if t.ObjectFlags()&checker.ObjectFlagsReference != 0 && t.Target() != nil {
		for _, argument := range l.checker.GetTypeArguments(t) {
			if l.overloadCallbackHoldsAny(argument, depth+1, seen) {
				return true
			}
		}
	}
	// A library type's members are the library's, not values the type holds: ReadonlyArray's every and
	// forEach take `thisArg?: any`, so walking them would find any in every array. Its type arguments,
	// walked above, are what a value of it holds.
	if symbol := t.Symbol(); symbol != nil && len(symbol.Declarations) > 0 {
		library := true
		for _, declaration := range symbol.Declarations {
			file := ast.GetSourceFileOfNode(declaration)
			if file == nil || !(load.IsLibrary(file) || load.IsPrelude(file)) {
				library = false
			}
		}
		if library {
			return false
		}
	}
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		for _, signature := range l.checker.GetSignaturesOfType(t, kind) {
			if l.overloadCallbackHoldsAny(l.checker.GetReturnTypeOfSignature(signature), depth+1, seen) {
				return true
			}
			for _, parameter := range signature.Parameters() {
				if l.overloadCallbackHoldsAny(l.checker.GetTypeOfSymbol(parameter), depth+1, seen) {
					return true
				}
			}
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(t) {
		if l.overloadCallbackHoldsAny(l.checker.GetTypeOfSymbol(property), depth+1, seen) {
			return true
		}
	}
	for _, index := range l.checker.GetIndexInfosOfType(t) {
		if l.overloadCallbackHoldsAny(index.ValueType(), depth+1, seen) {
			return true
		}
	}
	return false
}

// overloadCallbackProof carries the overload's rigid array types across direct
// forwards. Array confinement is checked independently of callback uses so an
// array-only hand-off cannot establish a new way to invoke the callback.
type overloadCallbackProof struct {
	l        *lowering
	root     *ast.Node
	promised *checker.Signature
	visiting map[*ast.Node]bool
	reads    []overloadCallbackRead
	moved    bool
}

type overloadCallbackRead struct {
	function *ast.Node
	position int
	name     string
}

type overloadCallbackFrame struct {
	function *ast.Node
	callback *ast.Symbol
	arrays   map[*ast.Symbol]*checker.Type
}

func (p *overloadCallbackProof) body(frame overloadCallbackFrame) bool {
	l := p.l
	function := frame.function
	if p.visiting[function] || !overloadCertificateFunction(function) ||
		overloadThisParameter(function) || overloadCallbackReadsArguments(function) {
		return false
	}
	p.visiting[function] = true
	defer delete(p.visiting, function)
	var visit func(*ast.Node, bool) bool
	visit = func(node *ast.Node, nested bool) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		if ast.IsIdentifier(node) && l.valueSymbol(node) == frame.callback {
			if l.overloadCallbackPresenceTest(node) {
				return false
			}
			if l.overloadCallbackTruthiness(frame.callback, node) {
				p.moved = true
				return false
			}
			if nested {
				return true
			}
			if l.overloadCallbackCall(function, p.root, node, p.promised, func(argument *ast.Node, expected *checker.Type) *checker.Type {
				if element := p.element(frame, argument, expected); element != nil {
					p.moved = true
					return element
				}
				if function == l.censusImplementation(p.root) {
					return l.overloadCallbackArgument(function, p.root, argument)
				}
				return l.concrete(l.checker.GetTypeAtLocation(argument))
			}) {
				return false
			}
			return !p.forward(frame, node)
		}
		inner := nested || ast.IsFunctionLike(node) || ast.IsClassLike(node)
		return node.ForEachChild(func(child *ast.Node) bool { return visit(child, inner) })
	}
	for _, parameter := range function.Parameters() {
		if initializer := parameter.Initializer(); initializer != nil && visit(initializer, true) {
			return false
		}
	}
	return !visit(function.Body(), false)
}

func (l *lowering) overloadCallbackTruthiness(symbol *ast.Symbol, reference *ast.Node) bool {
	operand := reference
	for operand.Parent != nil && operand.Parent.Kind == ast.KindParenthesizedExpression {
		operand = operand.Parent
	}
	if operand.Parent == nil {
		return false
	}
	switch operand.Parent.Kind {
	case ast.KindConditionalExpression:
		if operand.Parent.AsConditionalExpression().Condition != operand {
			return false
		}
	case ast.KindIfStatement:
		if operand.Parent.AsIfStatement().Expression != operand {
			return false
		}
	case ast.KindPrefixUnaryExpression:
		if operand.Parent.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken {
			return false
		}
	default:
		return false
	}
	declared := l.concrete(l.checker.GetTypeOfSymbol(symbol))
	parts := []*checker.Type{declared}
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		parts = declared.Types()
	}
	for _, part := range parts {
		if part.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) == 0 &&
			(part.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 ||
				len(l.checker.GetSignaturesOfType(part, checker.SignatureKindCall)) == 0) {
			return false
		}
	}
	return true
}

func overloadCallbackUnwrap(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
			node = node.Expression()
		default:
			return node
		}
	}
	return nil
}

func (p *overloadCallbackProof) element(frame overloadCallbackFrame, argument *ast.Node, expected *checker.Type) *checker.Type {
	l := p.l
	read := overloadCallbackUnwrap(argument)
	declared := (*checker.Type)(nil)
	if ast.IsIdentifier(read) {
		symbol := l.valueSymbol(read)
		if symbol == nil {
			return nil
		}
		binding := symbol.ValueDeclaration
		if binding == nil || binding.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(binding.Name()) ||
			!ast.IsVarConst(binding) || ast.GetContainingFunction(binding) != frame.function {
			return nil
		}
		if frame.function.Body().ForEachChild(func(node *ast.Node) bool {
			return p.captured(node, frame.function, symbol)
		}) {
			return nil
		}
		declared = l.concrete(l.checker.GetTypeAtLocation(binding.Name()))
		read = overloadCallbackUnwrap(binding.Initializer())
		if read == nil {
			// Keep the existing own-body for-of road and its allowlist intact.
			// A forwarded body instead uses the carried array at the new position.
			if frame.function == l.censusImplementation(p.root) || binding.Parent == nil ||
				binding.Parent.Kind != ast.KindVariableDeclarationList || binding.Parent.Parent == nil ||
				binding.Parent.Parent.Kind != ast.KindForOfStatement ||
				len(binding.Parent.AsVariableDeclarationList().Declarations.Nodes) != 1 {
				return nil
			}
			loop := binding.Parent.Parent.AsForInOrOfStatement()
			if loop.AwaitModifier != nil || loop.Initializer != binding.Parent {
				return nil
			}
			read = loop.AsNode()
		}
	}
	if read == nil || (read.Kind != ast.KindElementAccessExpression && read.Kind != ast.KindForOfStatement) {
		return nil
	}
	receiver := ast.SkipParentheses(read.Expression())
	if !ast.IsIdentifier(receiver) {
		return nil
	}
	symbol := l.valueSymbol(receiver)
	admitted := frame.arrays[symbol]
	if admitted == nil || (read.Kind == ast.KindElementAccessExpression && !l.overloadFilterElementAccess(read, symbol)) {
		return nil
	}
	element := l.checker.GetIndexTypeOfType(l.checker.GetNonNullableType(admitted), l.checker.GetNumberType())
	if element == nil {
		return nil
	}
	if declared == nil {
		own := l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(symbol))
		declared = l.concrete(l.checker.GetIndexTypeOfType(own, l.checker.GetNumberType()))
		if declared == nil {
			return nil
		}
		// A plain unchecked indexed read may yield undefined. Retain that arm.
		if l.censusHasUndefined(l.checker.GetTypeAtLocation(read)) {
			declared = l.checker.GetUnionType([]*checker.Type{declared, l.checker.GetUndefinedType()})
			element = l.checker.GetUnionType([]*checker.Type{element, l.checker.GetUndefinedType()})
		}
	} else if l.censusHasUndefined(declared) {
		element = l.checker.GetUnionType([]*checker.Type{element, l.checker.GetUndefinedType()})
	}
	flow := l.concrete(l.checker.GetTypeAtLocation(argument))
	switch flow {
	case declared:
	case l.checker.GetNonNullableType(declared):
		element = l.checker.GetNonNullableType(element)
	default:
		return nil
	}
	checked := len(p.reads)
	if !l.censusRelated(element, expected) ||
		!p.array(frame.function, symbol, frame.callback, true, map[*ast.Node]bool{}) {
		p.reads = p.reads[:checked]
		return nil
	}
	return element
}

func (p *overloadCallbackProof) captured(node, function *ast.Node, symbol *ast.Symbol) bool {
	if ast.IsPartOfTypeNode(node) {
		return false
	}
	return (ast.IsIdentifier(node) && p.l.valueSymbol(node) == symbol &&
		ast.GetContainingFunction(node) != function) ||
		node.ForEachChild(func(child *ast.Node) bool { return p.captured(child, function, symbol) })
}

func (p *overloadCallbackProof) target(reference *ast.Node) (*ast.CallExpression, *ast.Node, int) {
	l := p.l
	child, parent := reference, reference.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child, parent = parent, parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return nil, nil, -1
	}
	call := parent.AsCallExpression()
	if call.QuestionDotToken != nil || parent.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, nil, -1
	}
	position := -1
	for index, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, nil, -1
		}
		if argument == child {
			position = index
			break
		}
	}
	if position < 0 {
		return nil, nil, -1
	}
	callee := ast.SkipParentheses(call.Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
	case ast.KindPropertyAccessExpression:
		// A structural object with the same type is not that binding.
		if callee.Flags&ast.NodeFlagsOptionalChain != 0 ||
			!ast.IsIdentifier(callee.Expression()) || !l.namespaceMember(callee) {
			return nil, nil, -1
		}
	default:
		return nil, nil, -1
	}
	symbol := l.symbol(callee)
	if symbol == nil || l.assignedSymbol(symbol) {
		return nil, nil, -1
	}
	var target *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration {
			return nil, nil, -1
		}
		if declaration.Body() == nil {
			continue
		}
		file := ast.GetSourceFileOfNode(declaration)
		if target != nil || file == nil || load.IsLibrary(file) || load.IsPrelude(file) ||
			!overloadCertificateFunction(declaration) || overloadThisParameter(declaration) ||
			overloadCallbackReadsArguments(declaration) {
			return nil, nil, -1
		}
		target = declaration
	}
	if target == nil || position >= len(target.Parameters()) ||
		!overloadCallbackPlainParameter(target.Parameters()[position]) {
		return nil, nil, -1
	}
	return call, target, position
}

func (p *overloadCallbackProof) forward(frame overloadCallbackFrame, reference *ast.Node) bool {
	l := p.l
	call, target, position := p.target(reference)
	if target == nil {
		return false
	}
	destination := overloadCallbackFrame{function: target,
		callback: l.symbol(target.Parameters()[position].Name()), arrays: map[*ast.Symbol]*checker.Type{}}
	if destination.callback == nil {
		return false
	}
	for index, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			break
		}
		value := ast.SkipParentheses(argument)
		if !ast.IsIdentifier(value) {
			continue
		}
		symbol := l.valueSymbol(value)
		if frame.arrays[symbol] == nil {
			continue
		}
		if index >= len(target.Parameters()) || !overloadCallbackPlainParameter(target.Parameters()[index]) ||
			!p.array(frame.function, symbol, frame.callback, true, map[*ast.Node]bool{}) {
			return false
		}
		destination.arrays[l.valueSymbol(target.Parameters()[index].Name())] = frame.arrays[symbol]
	}
	if len(destination.arrays) == 0 {
		return false
	}
	p.moved = true
	return p.body(destination)
}

func (p *overloadCallbackProof) array(function *ast.Node, symbol, callback *ast.Symbol, mayReturn bool, visiting map[*ast.Node]bool) bool {
	l := p.l
	if symbol == nil || visiting[function] || overloadCallbackReadsArguments(function) ||
		l.overloadCallbackAssigned(function, symbol) {
		return false
	}
	position := -1
	for index, parameter := range function.Parameters() {
		if l.valueSymbol(parameter.Name()) == symbol {
			position = index
			break
		}
	}
	if position < 0 || !overloadCallbackPlainParameter(function.Parameters()[position]) {
		return false
	}
	visiting[function] = true
	defer delete(visiting, function)
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		if ast.IsIdentifier(node) && l.valueSymbol(node) == symbol && node != function.Parameters()[position].Name() {
			if ast.GetContainingFunction(node) != function {
				return true
			}
			return !p.arrayUse(function, symbol, callback, position, node, mayReturn, visiting)
		}
		return node.ForEachChild(visit)
	}
	return !function.ForEachChild(visit)
}

func (p *overloadCallbackProof) arrayUse(function *ast.Node, symbol, callback *ast.Symbol, position int, reference *ast.Node, mayReturn bool, visiting map[*ast.Node]bool) bool {
	l := p.l
	if l.overloadCallbackPresenceTest(reference) {
		return true
	}
	child, parent := reference, reference.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child, parent = parent, parent.Parent
	}
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindElementAccessExpression, ast.KindPropertyAccessExpression:
		if parent.Expression() != child || parent.Flags&ast.NodeFlagsOptionalChain != 0 {
			return false
		}
		for access := parent; access.Parent != nil; access = access.Parent {
			if ast.IsAssignmentTarget(access) || access.Parent.Kind == ast.KindDeleteExpression {
				return false
			}
			switch access.Parent.Kind {
			case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
				continue
			}
			break
		}
		if parent.Kind == ast.KindElementAccessExpression {
			return l.overloadFilterElementAccess(parent, symbol)
		}
		name := parent.Name().Text()
		use := parent
		for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
			use = use.Parent
		}
		if use.Parent != nil && use.Parent.Kind == ast.KindCallExpression && use.Parent.Expression() == use {
			return name == "slice" && !l.overloadCallbackOwnSlice() &&
				l.overloadFilterSlice(use.Parent, symbol)
		}
		if use.Parent != nil && use.Parent.Kind == ast.KindTaggedTemplateExpression {
			return false
		}
		p.reads = append(p.reads, overloadCallbackRead{function, position, name})
		return true
	case ast.KindForOfStatement:
		loop := parent.AsForInOrOfStatement()
		return loop.Expression == child && loop.AwaitModifier == nil
	}
	// These wrappers preserve the array's identity. They never authorize using
	// a cast array as the destination of a forward.
	for parent != nil && (parent.Kind == ast.KindParenthesizedExpression ||
		parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindTypeAssertionExpression) {
		child, parent = parent, parent.Parent
	}
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		return (binary.Left == child || binary.Right == child) &&
			(binary.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken ||
				binary.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken)
	}
	if parent.Kind == ast.KindReturnStatement {
		// An array-only callee would give the caller a new name for the array.
		return parent.Expression() == child && mayReturn
	}
	call, target, index := p.target(reference)
	if target == nil {
		return false
	}
	var forwarded *ast.Symbol
	for destination, argument := range call.Arguments.Nodes {
		value := ast.SkipParentheses(argument)
		if callback != nil && ast.IsIdentifier(value) && l.valueSymbol(value) == callback {
			if destination >= len(target.Parameters()) || !overloadCallbackPlainParameter(target.Parameters()[destination]) {
				return false
			}
			forwarded = l.symbol(target.Parameters()[destination].Name())
		}
	}
	// The entire callback body walk checks every mention of the callback.
	// Stores and unproved passes fail there, as do captures other than
	// presence tests. The hand-off cannot obtain the callback through them.
	return p.array(target, l.symbol(target.Parameters()[index].Name()), forwarded,
		forwarded != nil && p.terminalForward(function, call.AsNode(), callback), visiting)
}

// Returning the array from a forward can create an alias in the caller.
// Follow that return only when the forward is a direct body statement or
// initializer and the caller never uses the callback again.
func (p *overloadCallbackProof) terminalForward(function, call *ast.Node, callback *ast.Symbol) bool {
	parent := call.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindVariableDeclaration && parent.Initializer() == call && parent.Parent != nil {
		parent = parent.Parent.Parent
	}
	if parent == nil || parent.Parent != function.Body() {
		return false
	}
	var later ast.Visitor
	later = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		return ast.IsIdentifier(node) && node.Pos() >= call.End() && p.l.valueSymbol(node) == callback ||
			node.ForEachChild(later)
	}
	return !function.Body().ForEachChild(later)
}

// overloadCallbackOwnSlice searches the program's own declarations for a slice
// member. SourceFiles already omits the prelude and standard library. Node's
// bundled declarations and undici-types are compiler libraries too.
func (l *lowering) overloadCallbackOwnSlice() bool {
	if l.callbackSliceKnown {
		return l.callbackOwnSlice
	}
	l.callbackSliceKnown = true
	for _, file := range l.program.SourceFiles() {
		if load.IsNodeLibrary(file) || strings.HasPrefix(file.FileName().AsString(), "bundled:///node/node_modules/undici-types/") {
			continue
		}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if parent := node.Parent; parent != nil &&
				(ast.IsClassLike(parent) || parent.Kind == ast.KindInterfaceDeclaration || parent.Kind == ast.KindTypeLiteral) &&
				node.Name() != nil {
				name, known := memberKeyText(node)
				if name == "slice" {
					return true
				}
				if !known && node.Name().Kind == ast.KindComputedPropertyName {
					key := l.checker.GetTypeAtLocation(node.Name().Expression())
					if key == nil || key.Flags()&(checker.TypeFlagsESSymbolLike|checker.TypeFlagsNumberLike) == 0 {
						return true
					}
				}
			}
			return node.ForEachChild(visit)
		}
		if file.AsNode().ForEachChild(visit) {
			l.callbackOwnSlice = true
			break
		}
	}
	return l.callbackOwnSlice
}

// Only declarations with Array or ReadonlyArray ancestry carry an array's
// element. A numeric index signature by itself can describe another object.
func (l *lowering) overloadCallbackArray(t *checker.Type) bool {
	if t.Flags()&checker.TypeFlagsTypeParameter != 0 {
		bound := l.checker.GetBaseConstraintOfType(t)
		if bound == nil || bound == t {
			return false
		}
		t = bound
	}
	t = l.checker.GetNonNullableType(t)
	return l.isArrayType(t) || l.nodeArrayBase(t, map[*checker.Type]bool{}) != nil
}

type overloadCallbackReadProof struct {
	reads   []overloadCallbackRead
	refusal *Refused
}

func (l *lowering) activateCallbackReads(overload *ast.Node) {
	proofs := l.callbackReadProofs[overload]
	if len(proofs) == 0 || l.callbackActivated[overload] || l.overloadVerdicts[overload] != nil {
		return
	}
	if l.callbackActivated == nil {
		l.callbackActivated = map[*ast.Node]bool{}
	}
	l.callbackActivated[overload] = true
	l.callbackReadUses = append(l.callbackReadUses, proofs...)
}

func (l *lowering) activateCallbackSetReads(implementation *ast.Node) {
	if len(l.callbackReadProofs) == 0 || implementation == nil {
		return
	}
	if symbol := l.symbol(implementation.Name()); symbol != nil {
		for _, declaration := range symbol.Declarations {
			l.activateCallbackReads(declaration)
		}
	}
}

func (l *lowering) callbackOverloadSetStop(node *ast.Node, symbol *ast.Symbol) error {
	err := l.overloadSetStop(node, symbol)
	if err == nil && len(l.callbackReadProofs) != 0 {
		l.activateCallbackSetReads(l.overloadSetImplementation(symbol))
	}
	return err
}

// refuseReachedCallbackReads asks the same whole-program graph as property
// dispatch, on each lowered instance's unassigned parameter slot. It never
// changes the graph or the dispatcher's judgment. Only overload uses that
// relied on the new proof activate its obligations.
func (l *lowering) refuseReachedCallbackReads(reach *accessorReach, names map[string]bool) error {
	for _, proof := range l.callbackReadUses {
		for _, read := range proof.reads {
			if !names[read.name] {
				continue
			}
			found := false
			for index, function := range l.result.Functions {
				if l.functionNodes[index] != read.function {
					continue
				}
				if read.position >= len(function.Parameters) {
					return proof.refusal
				}
				local := function.Parameters[read.position]
				found = true
				if reach.owner(reach.graph.ReachingAllocations(ir.Read{Local: local, Of: l.result.Locals[local].Type}), read.name) {
					return proof.refusal
				}
			}
			if !found {
				return proof.refusal
			}
		}
	}
	return nil
}
