package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// maximumGenericDepth is how many instantiations of generic functions may be lowered inside one
// another. Ordinary programs nest a few; a function that calls itself with an ever larger type
// (polymorphic recursion) would instantiate without end, and 0.1 refuses it (docs/0.1.md).
const maximumGenericDepth = 32

// instantiateFunction lowers a generic module function for the type arguments of one call, once per
// distinct set of what they're held as at runtime, as a generic class is (instantiate). Within it,
// each type parameter is what this call made it.
//
// The checker resolves the call's signature with its type arguments substituted, but doesn't export
// the mapping itself, so it's read back the way it was made: each declared parameter's type, and the
// result's, against the resolved signature's, through arrays and tuples. A type parameter that can't
// be read back that way is left unmapped, and the body says not yet wherever it needs to know it.
func (l *lowering) instantiateFunction(call *ast.Node, declaration *ast.Node) (int, error) {
	resolved := l.checker.GetResolvedSignature(call)
	target := l.checker.GetSignatureFromDeclaration(declaration)
	if resolved == nil || target == nil {
		return 0, l.notYet(call, "a call to a generic function whose signature the checker didn't resolve")
	}
	substitution := map[*checker.Type]ir.Type{}
	declared, given := target.Parameters(), resolved.Parameters()
	for index := range declared {
		if index < len(given) {
			l.inferTypes(l.checker.GetTypeOfSymbol(declared[index]), l.checker.GetTypeOfSymbol(given[index]), substitution)
		}
	}
	l.inferTypes(l.checker.GetReturnTypeOfSignature(target), l.checker.GetReturnTypeOfSignature(resolved), substitution)

	key := l.program.Where(declaration)
	name := declaration.Name().Text()
	for _, parameter := range declaration.TypeParameters() {
		held, isKnown := substitution[l.checker.GetTypeAtLocation(parameter.Name())]
		if !isKnown {
			// Not read back (it stands only inside a union or an object, Result<Value>): left
			// unmapped, so whatever in the body needs to know how it's held says not yet, there.
			key += ",unread"
			name += "_unread"
			continue
		}
		key += "," + typeName(held)
		name += "_" + typeName(held)
	}
	if existing, isLowered := l.genericInstances[key]; isLowered {
		return existing, nil
	}
	if l.genericDepth >= maximumGenericDepth {
		return 0, &Refused{Where: l.program.Where(call), What: "a generic function instantiated without end (polymorphic recursion)", Fix: "call it with the same type arguments it was called with, or write a function per type"}
	}

	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name})
	if l.genericInstances == nil {
		l.genericInstances = map[string]int{}
	}
	l.genericInstances[key] = index

	// Lower with this instantiation's meaning of each type parameter, and locals of its own, outside
	// whatever function or closure called it.
	outerSubstitution, outerLocals, outerClosures := l.substitution, l.locals, l.closures
	l.substitution, l.closures = substitution, nil
	l.locals = map[*ast.Symbol]int{}
	for symbol, local := range outerLocals {
		if l.result.Locals[local].Global {
			l.locals[symbol] = local
		}
	}
	l.genericDepth++
	defer func() {
		l.substitution, l.locals, l.closures = outerSubstitution, outerLocals, outerClosures
		l.genericDepth--
	}()
	if err := l.lowerFunction(index, declaration, -1); err != nil {
		return 0, err
	}
	return index, nil
}

// inferTypes records what each type parameter in declared stands for, by where it stands in
// instantiated: the type itself, or an array's or a tuple's elements. What a type parameter stands
// for is held as the caller's own substitution makes it, so a generic function calling another with
// its own type parameter passes it on.
func (l *lowering) inferTypes(declared *checker.Type, instantiated *checker.Type, into map[*checker.Type]ir.Type) {
	if declared == nil || instantiated == nil {
		return
	}
	if declared.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if _, isSet := into[declared]; !isSet {
			if held, isKnown := l.representation(instantiated); isKnown {
				into[declared] = held
			}
		}
		return
	}
	if (l.checker.IsArrayType(declared) && l.checker.IsArrayType(instantiated)) || (checker.IsTupleType(declared) && checker.IsTupleType(instantiated)) {
		declaredArguments, instantiatedArguments := l.checker.GetTypeArguments(declared), l.checker.GetTypeArguments(instantiated)
		for index := range declaredArguments {
			if index < len(instantiatedArguments) {
				l.inferTypes(declaredArguments[index], instantiatedArguments[index], into)
			}
		}
	}
}
