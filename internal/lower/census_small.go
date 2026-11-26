package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// censusRestDeclaration recovers the declaration without adding state to lower.go.
func (l *lowering) censusRestDeclaration(function int) *ast.Node {
	for symbol, index := range l.functions {
		if index != function {
			continue
		}
		for _, declaration := range symbol.Declarations {
			if declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() == nil {
				continue
			}
			parameters := declaration.Parameters()
			if len(parameters) > 0 && parameters[len(parameters)-1].AsParameterDeclaration().DotDotDotToken != nil {
				return declaration
			}
		}
	}
	return nil
}

// censusRestCall evaluates fixed arguments before rest items, and copies each spread
// when encountered. The array is new even when the only item is a spread.
func (l *lowering) censusRestCall(call *ast.CallExpression, function int, declaration *ast.Node) (ir.Expression, error) {
	parameters := declaration.Parameters()
	fixed := len(parameters) - 1
	proven := l.checker.GetTypeAtLocation(parameters[fixed].Name())
	element, _ := l.kept(l.elementTypeOfArray(proven))
	rest := ir.ArrayLiteral{Element: element}
	arguments := []ir.Expression{}
	for index, argument := range call.Arguments.Nodes {
		spread := argument.Kind == ast.KindSpreadElement
		valueNode := argument
		if spread {
			if index < fixed {
				return nil, l.notYet(argument, "a spread filling fixed parameters before a rest parameter")
			}
			valueNode = argument.AsSpreadElement().Expression
		}
		value, err := l.expression(valueNode)
		if err != nil {
			return nil, err
		}
		if index < fixed {
			arguments = append(arguments, value)
			continue
		}
		if spread {
			source := l.checker.GetTypeAtLocation(valueNode)
			if !l.isArrayType(source) {
				return nil, l.notYet(argument, "a rest spread other than an array")
			}
			held, known := l.kept(l.elementTypeOfArray(source))
			if !known || held != element {
				return nil, l.notYet(argument, "a rest spread with differently held elements")
			}
		} else {
			value = fit(value, element)
			if value.Type() != element {
				return nil, l.notYet(argument, "a rest argument with another representation")
			}
		}
		rest.Elements = append(rest.Elements, value)
		rest.Spread = append(rest.Spread, spread)
	}
	// Missing fixed arguments must keep their positions before the rest array.
	for len(arguments) < fixed {
		of := l.result.Locals[l.result.Functions[function].Parameters[len(arguments)]].Type
		if !of.IsMaybe() && !of.IsReference() {
			return nil, l.notYet(parameters[len(arguments)], "a missing fixed argument before rest")
		}
		arguments = append(arguments, fit(ir.Undefined{}, of))
	}
	arguments = append(arguments, rest)
	return l.censusOverloadResult(call, ir.Call{Function: function, Arguments: arguments, Returns: l.result.Functions[function].Returns})
}

// A MaybeBoolean field has a tagged byte, and a fitted Union occupies one counted reference
// (boxedSlot). Those are the only slotless representations today, so this refuses none; it refuses
// any slotless representation added later until a field form is arranged for it.
func censusFieldSlotless(of ir.Type) bool {
	return slotless(of) && of != ir.MaybeBoolean && of != ir.Union
}

// Only the implementation of a named overload set has executable code.
func (l *lowering) censusImplementation(declaration *ast.Node) *ast.Node {
	if declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil {
		return nil
	}
	symbol := l.symbol(declaration.Name())
	if symbol == nil {
		return nil
	}
	for _, candidate := range symbol.Declarations {
		if candidate.Kind == ast.KindFunctionDeclaration && candidate.Body() != nil {
			return candidate
		}
	}
	return nil
}

// Use the same nominal, invariant and strictly contravariant relation as class overrides.
func (l *lowering) censusRelated(from, to *checker.Type) bool {
	from, to = l.concrete(from), l.concrete(to)
	return l.classAssignable(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil
}

// censusOverloads judges every overload of implementation's set, once per implementation, and records
// each one's verdict in overloadVerdicts: censusOverload's Refused or NotYet, or nil when it's served. An
// unserved overload no longer stops its set where it's declared. A direct call is judged by the overload
// it resolves to, any other entry and any use of the set as a value by the whole set (overloadCallVerdict,
// overloadSetStop, Adamic's amendment of rulings.md:131 on #4czdn4q). Only a function declaration's set is
// judged so: a method's overloads can dispatch to another class's implementation, which two neighbors of
// this rule keep stopped, "an overloaded override" (class_inheritance.go) and "a function without a body"
// (functions.go), so a method set is never judged per call (Language's must on #4czdn4q).
func (l *lowering) censusOverloads(implementation *ast.Node) {
	if implementation.Kind != ast.KindFunctionDeclaration || implementation.Body() == nil {
		return
	}
	symbol := l.symbol(implementation.Name())
	if symbol == nil {
		return
	}
	if _, judged := l.overloadVerdicts[implementation]; judged {
		return
	}
	if l.overloadVerdicts == nil {
		l.overloadVerdicts = map[*ast.Node]error{}
	}
	ordinal := 0
	for _, overload := range symbol.Declarations {
		if overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
			continue
		}
		ordinal++
		l.overloadVerdicts[overload] = l.censusOverload(implementation, overload, ordinal)
	}
	// The implementation's own key, always nil, says its set is judged.
	l.overloadVerdicts[implementation] = nil
}

// overloadVerdict is overload's own verdict from its set's census, which it runs if it hasn't run: nil
// when overload is served or isn't a bodyless function declaration of a set with an implementation.
func (l *lowering) overloadVerdict(overload *ast.Node) error {
	if overload == nil || overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
		return nil
	}
	implementation := l.censusImplementation(overload)
	if implementation == nil {
		return nil
	}
	l.censusOverloads(implementation)
	return l.overloadVerdicts[overload]
}

// overloadSetVerdict is the whole set's verdict, the first of its overloads' errors in declaration order,
// or nil when every overload is served. It runs implementation's census if it hasn't run.
func (l *lowering) overloadSetVerdict(implementation *ast.Node) error {
	l.censusOverloads(implementation)
	symbol := l.symbol(implementation.Name())
	if symbol == nil {
		return nil
	}
	for _, overload := range symbol.Declarations {
		if overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
			continue
		}
		if err := l.overloadVerdicts[overload]; err != nil {
			return err
		}
	}
	return nil
}

// overloadSetImplementation is the implementation of the overloaded function declaration symbol names,
// or nil when symbol isn't one.
func (l *lowering) overloadSetImplementation(symbol *ast.Symbol) *ast.Node {
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil {
			if implementation := l.censusImplementation(declaration); implementation != nil {
				return implementation
			}
		}
	}
	return nil
}

// overloadSetStop is the whole set's verdict said at node, a use of symbol's overload set other than a
// direct call: a value, an identity marker, a comparator taken by name. nil when symbol isn't an
// overloaded function declaration or every overload is served.
func (l *lowering) overloadSetStop(node *ast.Node, symbol *ast.Symbol) error {
	implementation := l.overloadSetImplementation(symbol)
	if implementation == nil {
		return nil
	}
	return overloadVerdictAt(l.overloadSetVerdict(implementation), l.program.Where(node))
}

// overloadVerdictAt is verdict, censusOverload's Refused or NotYet, said at where: the same kind, What
// and Fix, only Where moved to the call or use it stops.
func overloadVerdictAt(verdict error, where string) error {
	switch stopped := verdict.(type) {
	case *Refused:
		moved := *stopped
		moved.Where = where
		return &moved
	case *NotYet:
		moved := *stopped
		moved.Where = where
		return &moved
	}
	return verdict
}

// Compare generic signatures under the same rigid type parameters, never under one
// lucky runtime instantiation. Mapping the implementation's binders is alpha-renaming.
func (l *lowering) censusOverload(implementation, overload *ast.Node, ordinal int) error {
	label := fmt.Sprintf("overload %d of %s", ordinal, implementation.Name().Text())
	declaredTypes, servedTypes := overload.TypeParameters(), implementation.TypeParameters()
	outerMapper := l.typeMapper
	defer func() { l.typeMapper = outerMapper }()
	// Rigid, whoever asks: lowering an instance of a generic call installs that call's mapper
	// (instantiateFunctionSignature) before its signature reaches here, and a set judged under it was
	// judged at one call's types and kept for every other (#6zqxygf, Correctness's stack on #68x1t2y).
	// An implementation with binders of its own replaces it below; one without would have kept it.
	l.typeMapper = nil
	if len(servedTypes) > 0 {
		sources, targets := []*checker.Type{}, []*checker.Type{}
		inferred := map[*checker.Type]*checker.Type{}
		if len(servedTypes) > len(declaredTypes) {
			// Extra binders need a witness from the entire signature, not positional
			// alpha-renaming: implementations may insert a binder before an old one.
			pairs := [][2]*checker.Type{}
			for index, parameter := range implementation.Parameters() {
				if index < len(overload.Parameters()) {
					pairs = append(pairs, [2]*checker.Type{l.checker.GetTypeAtLocation(parameter), l.checker.GetTypeAtLocation(overload.Parameters()[index])})
				}
			}
			pairs = append(pairs, [2]*checker.Type{l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation)), l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))})
			for _, pair := range pairs {
				l.inferTypes(pair[0], pair[1], inferred)
			}
			served := map[*checker.Type]bool{}
			for _, parameter := range servedTypes {
				served[l.checker.GetTypeAtLocation(parameter.Name())] = true
			}
			// What inferTypes left unwitnessed in a union, once every pair has witnessed what it can, before
			// the bottom fallback below (censusInferRemainder).
			for _, pair := range pairs {
				l.censusInferRemainder(pair[0], pair[1], served, inferred)
			}
		}
		for index, parameter := range servedTypes {
			source := l.checker.GetTypeAtLocation(parameter.Name())
			target := inferred[source]
			if len(servedTypes) <= len(declaredTypes) {
				target = l.checker.GetTypeAtLocation(declaredTypes[index].Name())
			} else if target == nil {
				// A binder with no evidence can use bottom as a candidate witness.
				// This is not permission to erase it: every constraint, incoming
				// parameter and outgoing result below must still hold for that witness.
				target = l.checker.GetNonNullableType(l.checker.GetUndefinedType())
			}
			sources = append(sources, source)
			targets = append(targets, target)
		}
		l.typeMapper = newTypeMapper(sources, targets)
		for index, parameter := range servedTypes {
			if constraint := l.checker.GetBaseConstraintOfType(sources[index]); constraint != nil && !l.censusRelated(targets[index], constraint) {
				where, name := overload, "type "+l.checker.TypeToString(targets[index])
				if len(servedTypes) <= len(declaredTypes) {
					where, name = declaredTypes[index], "type parameter "+declaredTypes[index].Name().Text()
				}
				return &Refused{Where: l.program.Where(where), What: label + " " + name + " cannot satisfy implementation parameter " + parameter.Name().Text(), Fix: "make implementation constraints accept every type admitted by the overload"}
			}
		}
	}
	declared, served := overload.Parameters(), implementation.Parameters()
	for index := 0; index <= max(len(declared), len(served)); index++ {
		given, parameter, givenRest, err := l.censusOverloadParameter(declared, index)
		if err != nil {
			return l.notYet(overload, label+" with a non-array rest parameter")
		}
		takes, actual, takesRest, err := l.censusOverloadParameter(served, index)
		if err != nil {
			return l.notYet(overload, label+" with a non-array implementation rest parameter")
		}
		if actual == nil {
			continue
		} // JavaScript evaluates and ignores extra arguments.
		if parameter == nil && takesRest {
			continue
		} // Missing rest items make an empty array.
		if parameter == nil {
			given = l.checker.GetUndefinedType()
			parameter = actual
		}
		if givenRest && !takesRest {
			given = l.checker.GetUnionType([]*checker.Type{given, l.checker.GetUndefinedType()})
		}
		// A callback the implementation takes wider than promised is still served when the body only tests
		// it or calls it with what the overload's caller passed (overloadCallbackUse, #1w5bve2's block 3).
		if !l.censusRelated(given, takes) && !l.overloadCallbackUse(implementation, overload, actual, given, takes, label, parameter) {
			name := func(node *ast.Node) string {
				if ast.IsIdentifier(node.Name()) {
					return node.Name().Text()
				}
				return fmt.Sprintf("%d", index+1)
			}
			return &Refused{Where: l.program.Where(parameter), What: label + " parameter " + name(parameter) + " cannot be served by implementation parameter " + name(actual), Fix: "make the implementation accept every value admitted by this overload, without mutable widening or bivariance"}
		}
	}
	promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
	produced := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation))
	if !l.censusRelated(produced, promised) {
		from, to := l.concrete(produced), l.concrete(promised)
		if (l.hasPhantom(from) || l.hasPhantom(to)) && l.phantomCast(from, to) && l.phantomAssignable(from, to) && l.phantomAssignable(to, from) {
			return nil
		}
		served, err := l.overloadPhantomResult(implementation, overload, label, produced, promised)
		if err != nil {
			return err // An overload mint in .a, refused at each call that resolves to it (#z1qkw7d).
		}
		if served {
			// Only phantom brands differ: a primitive brand erased, or an extension at the top minted and listed
			// (rulings.md:34 with :32, #z5d0jye step 3, #z1qkw7d).
			return nil
		}
		if l.censusNullableOverloadResult(produced, promised) {
			return nil // Each resolved call proves the result or checks its presence.
		}
		proven, failed := l.predicateOverloadResultProven(implementation, overload)
		if proven {
			// Every call's callback is proven where it's passed: an overload's predicate parameter whose
			// implementation takes a plain boolean callback isn't a checked claim (checkedFunctionTypeClaim), so
			// predicateArguments demands a named function or arrow whose body proves both branches.
			return nil
		}
		if overloadVoidPromise(promised, l.concrete(produced)) {
			return nil // No read of a void value lowers (rulings.md:321), so any result serves it (#z5d0jye step 2).
		}
		if l.censusProveOverloadResult(implementation, overload) {
			// The body proves every reachable return under this overload's parameters (rulings.md:127's
			// erasure, under :128's proof contract), so no call needs a check (#z5d0jye step 2).
			return nil
		}
		if l.narrowOverload(overload) {
			return nil // In .ts each call resolving to it checks its result (overloadResultCheck, rulings.md:127).
		}
		if l.checkedAssertionSource(implementation) && l.checkedAssertionSource(overload) && l.overloadKindExtractIntersection(promised, produced) {
			// tsc's transformNamedEvaluation shape, a gap Adamic named rather than a covariance error, and in
			// .ts only: a .a file never holds a hidden gap, so there it stays the refusal below (#0kkpw2e set 2).
			return l.notYet(overload, "an overload result Extract over intersection members the implementation's result doesn't serve")
		}
		if l.overloadArrayCertificate(implementation, overload) {
			// The body builds a fresh array of its iterable's elements or their mapped values only, which is
			// this overload's result when its arguments are exactly the caller's, so every call keeps that
			// premise at overloadCertificateCall (certificate A, Adamic's ruling on #5fws6q8).
			l.noteOverloadCertificate(overload, true)
			return nil
		}
		if l.overloadMapCertificate(implementation, overload) {
			// The body builds a fresh Map whose every set puts a key and a value this overload's Map holds when
			// its arguments are exactly the caller's, its mapper's default discharged, so every call keeps that
			// premise at overloadCertificateCall (certificate B, Adamic's ruling on #5fws6q8).
			l.noteOverloadCertificate(overload, true)
			return nil
		}
		if l.overloadFilterCertificate(implementation, overload) {
			// The body returns its first parameter as the caller passed it, or a fresh array it built only
			// from that parameter's elements, either of which is this overload's result when its arguments
			// are exactly the caller's, so every call keeps that premise at overloadCertificateCall
			// (certificate D, #4z7y6pq).
			l.noteOverloadCertificate(overload, true)
			return nil
		}
		if l.overloadNodeTestProof(implementation, overload) {
			// Every return is the node argument before anything writes it, a value the node test just said
			// yes to, or undefined: this overload's result at a call that passes its test, so each direct call
			// is judged at overloadNodeTestCall (Adamic's ruling on #0kkpw2e set 3).
			l.noteOverloadNodeTest(overload, true)
			return nil
		}
		if l.overloadGuardedArrayProof(implementation, overload) {
			// Every return is the array parameter narrowed to undefined, or an array whose every element the
			// overload's own predicate just said yes to: this overload's NodeArray at a call that passes its
			// test, so each direct call is judged at overloadGuardedArrayCall (Lower's design (b) on #ytagqs5).
			l.noteOverloadGuardedArray(overload, true)
			return nil
		}
		what := label + " result " + l.checker.TypeToString(promised) + " cannot be served by implementation result " + l.checker.TypeToString(produced)
		if failed != nil {
			what += " (return at " + l.program.Where(failed) + " isn't the value its predicate checked)"
		}
		return &Refused{Where: l.program.Where(overload), What: what, Fix: "make the implementation result covariant with every overload result"}
	}
	return nil
}

// censusInferRemainder reads, for censusOverload alone, a witness inferTypes leaves out: a binder of served
// that is a member of a union in the implementation's declared type, every other member of which is
// witnessed or concrete, stands for the overload's type there less each of its members another member's
// witness already takes, or for the whole of it when that leaves nothing or it isn't a union. tsc's
// arrayToMap overload 4 is the case: makeValue's V1 | V2 against U, V1 witnessed as T, where inferTypes'
// union arm asks T to fit U and witnesses nothing, so V2 was bottom and the parameter census refused
// makeValue; here V2 is U. inferTypes is left as it is, since it also instantiates real generic calls. A
// witness read here is a candidate like any other: censusOverload still holds every constraint, parameter
// and the result to it, and a generic implementation's body is right under every instantiation its
// constraints allow, so a better witness lets an overload through only where its parameters and result
// relate under it (Lower's design on #5fws6q8). The walk is inferTypes': through an optional member, a
// reference's type arguments, and a call signature's parameters and result. It never overwrites a
// witness, and a union with more than one unwitnessed binder gives none.
func (l *lowering) censusInferRemainder(declared, instantiated *checker.Type, served map[*checker.Type]bool, inferred map[*checker.Type]*checker.Type) {
	if declared == nil || instantiated == nil {
		return
	}
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		var missing *checker.Type
		known := []*checker.Type{}
		for _, member := range declared.Types() {
			if witness, found := inferred[member]; found {
				known = append(known, witness)
			} else if member.Flags()&checker.TypeFlagsTypeParameter != 0 {
				if missing != nil {
					return // More than one unwitnessed binder has no unique witness.
				}
				missing = member
			} else {
				known = append(known, l.concrete(member))
			}
		}
		if missing != nil {
			if served[missing] {
				inferred[missing] = l.censusRemainder(l.concrete(instantiated), known)
			}
			return
		}
		// An optional implementation parameter wraps the member the overload exposes, as in inferTypes.
		if present := l.definedMembers(declared); len(present) == 1 {
			given := l.checker.GetUnionType(l.definedMembers(instantiated))
			if given.Flags()&checker.TypeFlagsUndefined == 0 {
				l.censusInferRemainder(present[0], given, served, inferred)
			}
		}
		return
	}
	if declared.Flags()&checker.TypeFlagsObject == 0 || instantiated.Flags()&checker.TypeFlagsObject == 0 {
		return
	}
	if declared.ObjectFlags()&checker.ObjectFlagsReference != 0 && instantiated.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		declaredArguments, instantiatedArguments := l.checker.GetTypeArguments(declared), l.checker.GetTypeArguments(instantiated)
		for index := range declaredArguments {
			if index < len(instantiatedArguments) {
				l.censusInferRemainder(declaredArguments[index], instantiatedArguments[index], served, inferred)
			}
		}
	}
	declaredSignatures, instantiatedSignatures := l.checker.GetSignaturesOfType(declared, checker.SignatureKindCall), l.checker.GetSignaturesOfType(instantiated, checker.SignatureKindCall)
	if len(declaredSignatures) == 0 || len(instantiatedSignatures) == 0 {
		return
	}
	declaredParameters, instantiatedParameters := declaredSignatures[0].Parameters(), instantiatedSignatures[0].Parameters()
	for index := range declaredParameters {
		if index < len(instantiatedParameters) {
			l.censusInferRemainder(l.checker.GetTypeOfSymbol(declaredParameters[index]), l.checker.GetTypeOfSymbol(instantiatedParameters[index]), served, inferred)
		}
	}
	l.censusInferRemainder(l.checker.GetReturnTypeOfSignature(declaredSignatures[0]), l.checker.GetReturnTypeOfSignature(instantiatedSignatures[0]), served, inferred)
}

// censusRemainder is instantiated less each of its members some type of known already takes, or the whole
// of it when that leaves nothing or it isn't a union.
func (l *lowering) censusRemainder(instantiated *checker.Type, known []*checker.Type) *checker.Type {
	if instantiated.Flags()&checker.TypeFlagsUnion == 0 {
		return instantiated
	}
	kept := []*checker.Type{}
	for _, member := range instantiated.Types() {
		taken := false
		for _, other := range known {
			taken = taken || l.checker.IsTypeAssignableTo(member, other)
		}
		if !taken {
			kept = append(kept, member)
		}
	}
	if len(kept) == 0 {
		return instantiated
	}
	return l.checker.GetUnionType(kept)
}

func (l *lowering) censusOverloadParameter(parameters []*ast.Node, index int) (*checker.Type, *ast.Node, bool, error) {
	if len(parameters) == 0 {
		return nil, nil, false, nil
	}
	parameter := parameters[min(index, len(parameters)-1)]
	rest := parameter.AsParameterDeclaration().DotDotDotToken != nil
	if index >= len(parameters) && !rest {
		return nil, nil, false, nil
	}
	proven := l.checker.GetTypeAtLocation(parameter)
	if ast.IsIdentifier(parameter.Name()) {
		proven = l.censusCallableParameterType(l.symbol(parameter.Name()))
	}
	proven = l.concrete(proven)
	if rest {
		if !l.isArrayType(proven) {
			return nil, parameter, true, l.notYet(parameter, "an overload rest other than an array")
		}
		proven = l.elementTypeOfArray(proven)
	}
	return proven, parameter, rest, nil
}

// overloadVoidPromise says whether an overload promises exactly void, not a union holding it, over an
// implementation whose result owns nothing: number, boolean, their literals or undefined. Such a result
// needs no release however the call drops it (Lower's condition on #z5d0jye step 2).
func overloadVoidPromise(promised, produced *checker.Type) bool {
	if promised.Flags()&checker.TypeFlagsVoid == 0 || promised.Flags()&checker.TypeFlagsUnion != 0 {
		return false
	}
	owns := func(member *checker.Type) bool {
		return member.Flags()&^(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0
	}
	if produced.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range produced.Types() {
			if owns(member) {
				return false
			}
		}
		return true
	}
	return !owns(produced)
}

// Calls use the resolved overload's result representation, even when the
// implementation proves a narrower result. A plain false must become a present
// boolean | undefined, and a scalar result must be boxed when the caller sees a union.
func (l *lowering) censusOverloadResult(call *ast.CallExpression, value ir.Expression) (ir.Expression, error) {
	symbol := l.symbol(ast.SkipParentheses(call.Expression))
	overloaded := false
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil && l.censusImplementation(declaration) != nil {
				overloaded = true
				break
			}
		}
	}
	if !overloaded {
		return l.checkedPredicateResult(call, value, nil, nil)
	}
	resolved := l.checker.GetResolvedSignature(call.AsNode())
	if resolved != nil && resolved.Declaration() != nil {
		overload := resolved.Declaration()
		implementation := l.censusImplementation(overload)
		if implementation != nil && overload.Body() == nil {
			if overload.Type() != nil && overload.Type().Kind == ast.KindTypePredicate {
				return l.checkedPredicateResult(call, value, implementation, overload)
			}
			if overloadVoidPromise(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload)), l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation)))) {
				// A void promise's value is never read (rulings.md:321), so it keeps the implementation's
				// representation, the one its release reads it as, as a discarded result does (:146).
				return value, nil
			}
			if l.narrowOverloadCall(call.AsNode()) {
				checked, err := l.overloadResultCheck(call.AsNode(), value)
				if err != nil || l.overloadResultDiscarded(call.AsNode()) {
					// A result no read reaches keeps the implementation's representation, the one its
					// release reads it as (Adamic's condition on #z5d0jye).
					return checked, err
				}
				return l.censusOverloadFit(call, checked)
			}
			ordinal := 0
			for _, declaration := range symbol.Declarations {
				if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil {
					ordinal++
					if declaration == overload {
						break
					}
				}
			}
			if !l.censusProveOverloadResult(implementation, overload) {
				promised := l.checker.GetReturnTypeOfSignature(resolved)
				produced := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation))
				if l.censusHasUndefined(produced) && !l.censusHasUndefined(promised) {
					of, err := l.typeOf(call.AsNode())
					if err != nil {
						return nil, err
					}
					message := fmt.Sprintf("overload %d of %s result: expected %s, got undefined", ordinal, implementation.Name().Text(), l.checker.TypeToString(promised))
					return ir.Coalesce{Value: value, Of: of, Panic: ir.StringConstant{Index: l.constant(message)}}, nil
				}
			}
		}
	}
	return l.censusOverloadFit(call, value)
}

// censusOverloadFit fits an overloaded call's result to the representation the resolved overload's
// result has.
func (l *lowering) censusOverloadFit(call *ast.CallExpression, value ir.Expression) (ir.Expression, error) {
	if value.Type() == 0 {
		return value, nil
	}
	of, err := l.typeOf(call.AsNode())
	if err != nil {
		return nil, err
	}
	value = fit(value, of)
	if value.Type() != of {
		return nil, l.notYet(call.AsNode(), "an overload result requiring another representation")
	}
	return value, nil
}

// Callable slots use the same tagged boolean representation as fields.
func censusCallableSlotless(of ir.Type) bool {
	return slotless(of) && of != ir.MaybeBoolean && of != ir.Union
}

// A default's body-local type excludes undefined, but its incoming argument can be undefined.
func (l *lowering) censusCallableParameter(parameter *ast.Symbol) (ir.Type, bool) {
	of, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
	for _, declaration := range parameter.Declarations {
		if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().Initializer != nil {
			of = ir.Maybe(of)
		}
	}
	return of, known
}

// Signature symbols describe the body-local type of a defaulted parameter. The
// callable contract also accepts undefined, which selects that parameter's default.
func (l *lowering) censusCallableParameterType(parameter *ast.Symbol) *checker.Type {
	of := l.checker.GetTypeOfSymbol(parameter)
	for _, declaration := range parameter.Declarations {
		if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().Initializer != nil {
			return l.checker.GetUnionType([]*checker.Type{of, l.checker.GetUndefinedType()})
		}
	}
	return of
}

// callableParameterTypeAt is a signature's parameter type at position, as a call there sees it: its own
// parameter's (censusCallableParameterType, which adds undefined for one with an initializer), or, at or
// past a rest parameter, the rest's element at that position (GetTypeAtPosition), never the rest's whole
// tuple or array (#6xzs6gj).
func (l *lowering) callableParameterTypeAt(signature *checker.Signature, position int) *checker.Type {
	parameters := signature.Parameters()
	if position < len(parameters) && !(signature.HasRestParameter() && position == len(parameters)-1) {
		return l.censusCallableParameterType(parameters[position])
	}
	return l.checker.GetTypeAtPosition(signature, position)
}

// Only exactly true is truthy in boolean | undefined. Equality evaluates the value once.
func censusBooleanCondition(value ir.Expression) ir.Expression {
	if value.Type() == ir.MaybeBoolean {
		return ir.Binary{Operator: ir.Equal, Left: value, Right: fit(ir.BooleanConstant{Value: true}, ir.MaybeBoolean)}
	}
	return value
}

func (l *lowering) censusBooleanLogical(node *ast.Node, operator ast.Kind, left, right ir.Expression) (ir.Expression, bool) {
	if operator != ast.KindAmpersandAmpersandToken && operator != ast.KindBarBarToken {
		return nil, false
	}
	boolean := func(of ir.Type) bool { return of == ir.Boolean || of == ir.MaybeBoolean }
	if !boolean(left.Type()) || !boolean(right.Type()) {
		return nil, false
	}
	of, err := l.typeOf(node)
	if err != nil || !boolean(of) {
		// Two booleans give a boolean whatever the checker narrowed the whole to: a branch it
		// knows is dead types the expression never, and that branch must still lower.
		if left.Type() != ir.Boolean || right.Type() != ir.Boolean {
			return nil, false
		}
		of = ir.Boolean
	}
	lowered := ir.And
	if operator == ast.KindBarBarToken {
		lowered = ir.Or
	}
	if lowered == ir.And && left.Type() == ir.MaybeBoolean {
		of = ir.MaybeBoolean
	}
	if of == ir.MaybeBoolean {
		left, right = fit(left, of), fit(right, of)
	} else {
		left, right = censusBooleanCondition(left), censusBooleanCondition(right)
	}
	return ir.Binary{Operator: lowered, Left: left, Right: right}, true
}

// ToBoolean never calls user conversion methods. Every represented value can be tested.
// NumberCall already carries a single evaluated argument through both backends and analyses.
func censusCondition(value ir.Expression) ir.Expression {
	if value.Type() == ir.Boolean {
		return value
	}
	return ir.NumberCall{Function: "toBoolean", Arguments: []ir.Expression{value}}
}

// A never-rest signature is an erased callable marker. No ordinary argument can
// inhabit its rest element; storage compares results without pairing parameters.
func (l *lowering) censusNeverRestSignature(signature *checker.Signature) bool {
	parameters := signature.Parameters()
	if len(parameters) != 1 {
		return false
	}
	rest := false
	for _, declaration := range parameters[0].Declarations {
		if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
			rest = true
		}
	}
	if !rest {
		return false
	}
	parameter := l.concrete(l.checker.GetTypeOfSymbol(parameters[0]))
	if l.isArrayType(parameter) {
		parameter = l.elementTypeOfArray(parameter)
	}
	return parameter.Flags()&checker.TypeFlagsNever != 0
}
