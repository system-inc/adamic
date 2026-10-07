package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// censusRestParameter admits rest only where every call passes a freshly packed array.
// Function values and methods have a different calling convention and remain NotYet.
func (l *lowering) censusRestParameter(declaration, parameter *ast.Node) error {
	if declaration.Kind != ast.KindFunctionDeclaration || len(declaration.TypeParameters()) != 0 {
		return l.notYet(parameter, "a rest parameter outside a nongeneric named function")
	}
	parameters := declaration.Parameters()
	if parameters[len(parameters)-1] != parameter {
		return l.notYet(parameter, "a rest parameter before another parameter")
	}
	proven := l.checker.GetTypeAtLocation(parameter.Name())
	if !l.checker.IsArrayType(proven) {
		return l.notYet(parameter, "a rest parameter other than an array")
	}
	element, known := l.kept(l.checker.GetElementTypeOfArrayType(proven))
	if !known || slotless(element) {
		return l.notYet(parameter, "a rest array of "+l.checker.TypeToString(proven))
	}
	return nil
}

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
	element, _ := l.kept(l.checker.GetElementTypeOfArrayType(proven))
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
		of := l.result.Locals[l.result.Functions[function].Parameters[len(arguments)]].Type
		if !of.IsMaybe() && !of.IsReference() {
			return nil, l.notYet(parameters[len(arguments)], "a missing fixed argument before rest")
		}
		arguments = append(arguments, fit(ir.Undefined{}, of))
	}
	arguments = append(arguments, rest)
	return ir.Call{Function: function, Arguments: arguments, Returns: l.result.Functions[function].Returns}, nil
}
