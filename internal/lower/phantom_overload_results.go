package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Overload signatures have no runtime bodies. Only a declaration in the same
// symbol group with a real body can authorize skipping their registration.
func (l *lowering) overloadImplementation(declaration *ast.Node) *ast.Node {
	if declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil {
		return nil
	}
	symbol := l.symbol(declaration.Name())
	if symbol == nil {
		return nil
	}
	for _, member := range symbol.Declarations {
		if member.Kind == ast.KindFunctionDeclaration && member.Body() != nil {
			return member
		}
	}
	return nil
}

// A phantom cast may add only a name here. Bidirectional compatibility after
// stripping the brand preserves literals, union arms, readonly and element types.
func (l *lowering) phantomOverloadResult(where, implementation *ast.Node, from, to *checker.Type) bool {
	if (l.hasPhantom(from) || l.hasPhantom(to)) && l.phantomCast(from, to) && l.phantomAssignable(from, to) && l.phantomAssignable(to, from) {
		return true
	}
	if l.phantomArrayBase(from) == nil && l.phantomArrayBase(to) == nil {
		return false
	}
	baseFrom, baseTo := l.phantomArrayView(from), l.phantomArrayView(to)
	if l.checker.IsArrayType(baseFrom) && l.checker.IsArrayType(baseTo) &&
		l.isLibraryType(baseFrom, "ReadonlyArray") == l.isLibraryType(baseTo, "ReadonlyArray") {
		left, right := l.checker.GetElementTypeOfArrayType(baseFrom), l.checker.GetElementTypeOfArrayType(baseTo)
		// An accepted primitive phantom brand adds only absent void fields.
		// Preserve the element's literal constraints in both directions, just
		// as for the same scalar overload result. The array's ownership and
		// mutable slots still undergo the ordinary relation checks.
		if l.phantomAssignable(left, right) && l.phantomAssignable(right, left) &&
			l.sameKeeping(baseFrom, baseTo, map[[2]*checker.Type]bool{}) &&
			l.widened(baseFrom, baseTo, map[[2]*checker.Type]bool{}) == nil {
			return true
		}
	}
	if !l.checker.IsTypeAssignableTo(baseFrom, baseTo) || !l.checker.IsTypeAssignableTo(baseTo, baseFrom) {
		return false
	}
	handled, err := l.phantomArrayCast(where, implementation.Name(), from, to)
	return handled && err == nil
}

func (l *lowering) overloadResults(implementation *ast.Node) error {
	if implementation.Kind != ast.KindFunctionDeclaration || implementation.Name() == nil || implementation.Body() == nil {
		return nil
	}
	symbol := l.symbol(implementation.Name())
	if symbol == nil {
		return nil
	}
	from := l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation)))
	for _, overload := range symbol.Declarations {
		if overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
			continue
		}
		if len(implementation.TypeParameters()) != 0 || len(overload.TypeParameters()) != 0 {
			return l.notYet(overload, "generic overload result proofs")
		}
		overloadSignature := l.checker.GetSignatureFromDeclaration(overload)
		implementationSignature := l.checker.GetSignatureFromDeclaration(implementation)
		for index, parameter := range overloadSignature.Parameters() {
			if index >= len(implementationSignature.Parameters()) {
				break
			}
			fromParameter := l.checker.GetTypeOfSymbol(parameter)
			toParameter := l.checker.GetTypeOfSymbol(implementationSignature.Parameters()[index])
			if !l.sameKeeping(fromParameter, toParameter, map[[2]*checker.Type]bool{}) || l.provenTypesRelation(overload, implementation.Name(), fromParameter, toParameter) != nil {
				return &Refused{Where: l.program.Where(overload), What: "an overload parameter " + parameter.Name + " not proven compatible with its implementation", Fix: "make the implementation accept every overload parameter with compatible writable slots and ownership"}
			}
		}
		to := l.concrete(l.checker.GetReturnTypeOfSignature(overloadSignature))
		if l.sameKeeping(from, to, map[[2]*checker.Type]bool{}) && l.provenTypesRelation(overload, implementation.Name(), from, to) == nil {
			continue
		}
		if l.phantomOverloadResult(overload, implementation, from, to) {
			continue
		}
		return &Refused{Where: l.program.Where(overload), What: "an overload result " + l.checker.TypeToString(to) + " not proven by its implementation result " + l.checker.TypeToString(from), Fix: "return a type proven by the implementation, or use only an accepted phantom brand difference (adamic/overload-results)"}
	}
	return nil
}

func (l *lowering) overloadedCall(call *ast.CallExpression, value ir.Call) (ir.Expression, error) {
	symbol := l.symbol(ast.SkipParentheses(call.Expression))
	overloaded := false
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil {
				overloaded = true
			}
		}
	}
	if !overloaded {
		return value, nil
	}
	parameters := l.result.Functions[value.Function].Parameters
	for index, argument := range value.Arguments {
		if index >= len(parameters) {
			return nil, l.notYet(call.AsNode(), "an overloaded call with more arguments than its implementation")
		}
		takes := l.result.Locals[parameters[index]].Type
		value.Arguments[index] = fit(argument, takes)
		if value.Arguments[index].Type() != takes {
			return nil, l.notYet(call.AsNode(), "an overload argument with a different implementation representation")
		}
	}
	result := l.checker.GetTypeAtLocation(call.AsNode())
	if result.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0 {
		return value, nil
	}
	of, known := l.representation(result)
	if !known {
		return nil, l.notYet(call.AsNode(), "an overload result without a proven representation")
	}
	converted := fit(value, of)
	if converted.Type() != of {
		return nil, l.notYet(call.AsNode(), "an overload result with a different implementation representation")
	}
	return converted, nil
}

// Forwarding a set of overloads as one closure requires a separate signature
// adapter proof; direct calls above fit against the real implementation instead.
func (l *lowering) overloadedValue(node *ast.Node) error {
	symbol := l.symbol(node)
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil {
			return l.notYet(node, "an overloaded function read as a value")
		}
	}
	return nil
}
