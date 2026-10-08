package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// restParameterIndex describes the source calling convention, including an instantiated signature.
func restParameterIndex(signature *checker.Signature) int {
	if signature == nil {
		return -1
	}
	for index, parameter := range signature.Parameters() {
		for _, declaration := range parameter.Declarations {
			if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				return index
			}
		}
	}
	return -1
}

// sameRestConvention refuses callable views that would require an argument adapter.
// Packing an array is safe only when both ends agree on its position and element storage.
func (l *lowering) sameRestConvention(from, to *checker.Signature) bool {
	if l.censusNeverRestSignature(to) {
		return true
	}
	a, b := restParameterIndex(from), restParameterIndex(to)
	if a != b {
		return false
	}
	if a < 0 {
		return true
	}
	source := l.concrete(l.checker.GetTypeOfSymbol(from.Parameters()[a]))
	target := l.concrete(l.checker.GetTypeOfSymbol(to.Parameters()[b]))
	if !l.checker.IsArrayType(source) || !l.checker.IsArrayType(target) {
		return source == target
	}
	held, known := l.kept(l.checker.GetElementTypeOfArrayType(source))
	viewed, represented := l.kept(l.checker.GetElementTypeOfArrayType(target))
	return known && represented && held == viewed
}

// packRestArguments is shared by direct calls and closure calls. Resolve generics
// from this call's signature rather than from whichever instantiation lowered last.
func (l *lowering) packRestArguments(call *ast.CallExpression, signature *checker.Signature) ([]ir.Expression, error) {
	if signature == nil || restParameterIndex(signature) != len(signature.Parameters())-1 {
		return nil, l.notYet(call.AsNode(), "a rest call without a resolved final rest parameter")
	}
	parameters := signature.Parameters()
	fixed := len(parameters) - 1
	proven := l.concrete(l.checker.GetTypeOfSymbol(parameters[fixed]))
	if !l.checker.IsArrayType(proven) {
		return l.packRestTuple(call, parameters, proven)
	}
	element, known := l.kept(l.checker.GetElementTypeOfArrayType(proven))
	if !known || (slotless(element) && element != ir.Union) {
		return nil, l.notYet(call.AsNode(), "a rest call with unrepresented elements")
	}
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
			source := l.concrete(l.checker.GetTypeAtLocation(valueNode))
			if !l.checker.IsArrayType(source) {
				return nil, l.notYet(argument, "a rest spread other than an array")
			}
			held, known := l.kept(l.checker.GetElementTypeOfArrayType(source))
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
		of, known := l.censusCallableParameter(parameters[len(arguments)])
		if !known {
			return nil, l.notYet(call.AsNode(), "a missing fixed rest-call argument without a representation")
		}
		if !of.IsMaybe() && !of.IsReference() {
			return nil, l.notYet(call.AsNode(), "a missing fixed argument before rest")
		}
		arguments = append(arguments, fit(ir.Undefined{}, of))
	}
	arguments = append(arguments, rest)
	return arguments, nil
}
