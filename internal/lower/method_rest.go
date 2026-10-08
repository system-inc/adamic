package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Keep source arguments separate until the callable packs its own rest array.
func (l *lowering) methodCallableArguments(node *ast.Node, signatures []*checker.Signature) ([]ir.Expression, []bool, error) {
	var parameters []*ast.Symbol
	if len(signatures) == 1 {
		parameters = signatures[0].Parameters()
	}
	rest := -1
	for index, parameter := range parameters {
		for _, declaration := range parameter.Declarations {
			if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				rest = index
			}
		}
	}
	arguments, spreads := []ir.Expression{}, []bool{}
	for index, argument := range node.AsCallExpression().Arguments.Nodes {
		spread := argument.Kind == ast.KindSpreadElement
		written := argument
		if spread {
			if rest < 0 || index < rest {
				return nil, nil, l.notYet(argument, "a callable spread filling fixed parameters")
			}
			written = argument.AsSpreadElement().Expression
		}
		value, err := l.expression(written)
		if err != nil {
			return nil, nil, err
		}
		if rest >= 0 && index >= rest {
			proven := l.checker.GetTypeOfSymbol(parameters[rest])
			if !l.checker.IsArrayType(proven) {
				return nil, nil, l.notYet(argument, "a callable rest other than an array")
			}
			element, known := l.kept(l.checker.GetElementTypeOfArrayType(proven))
			if !known || slotless(element) {
				return nil, nil, l.notYet(argument, "a callable rest without represented elements")
			}
			if spread {
				actual, err := l.elementType(written)
				if err != nil || actual != element {
					return nil, nil, l.notYet(argument, "a callable rest spread with differently held elements")
				}
			} else {
				value = fit(value, element)
			}
		} else if index < len(parameters) {
			if takes, known := l.censusCallableParameter(parameters[index]); known {
				value = fit(value, takes)
			}
		}
		arguments = append(arguments, value)
		spreads = append(spreads, spread)
	}
	return arguments, spreads, nil
}
