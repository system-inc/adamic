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
			// Existing census overload proofs compare generic binders rigidly.
			continue
		}
		overloadSignature := l.checker.GetSignatureFromDeclaration(overload)
		declared, served := overload.Parameters(), implementation.Parameters()
		for index := 0; index <= max(len(declared), len(served)); index++ {
			fromParameter, parameter, givenRest, err := l.censusOverloadParameter(declared, index)
			if err != nil {
				return err
			}
			toParameter, actual, takesRest, err := l.censusOverloadParameter(served, index)
			if err != nil {
				return err
			}
			if actual == nil || parameter == nil && takesRest {
				continue
			}
			if parameter == nil {
				fromParameter, parameter = l.checker.GetUndefinedType(), actual
			}
			if givenRest && !takesRest {
				fromParameter = l.checker.GetUnionType([]*checker.Type{fromParameter, l.checker.GetUndefinedType()})
			}
			if !l.sameKeeping(fromParameter, toParameter, map[[2]*checker.Type]bool{}) || l.provenTypesRelation(overload, implementation.Name(), fromParameter, toParameter) != nil {
				name := "parameter"
				if ast.IsIdentifier(parameter.Name()) {
					name = parameter.Name().Text()
				}
				return &Refused{Where: l.program.Where(overload), What: "an overload parameter " + name + " not proven compatible with its implementation", Fix: "make the implementation accept every overload parameter with compatible writable slots and ownership"}
			}
		}
		to := l.concrete(l.checker.GetReturnTypeOfSignature(overloadSignature))
		if l.censusNullableOverloadResult(from, to) {
			continue
		}
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
	targets := l.result.CallTargets(value)
	if len(targets) != 1 {
		return nil, l.notYet(call.AsNode(), "an overloaded call without one proven implementation target")
	}
	parameters := l.result.Functions[targets[0]].Parameters
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
