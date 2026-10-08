package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A homogeneous tuple with a trailing rest has variable length and the same
// element storage as an array. Fixed and heterogeneous tuples retain their layout.
func (l *lowering) arrayTupleElement(tuple *checker.Type) (ir.Type, bool) {
	if !checker.IsTupleType(tuple) {
		return 0, false
	}
	flags := tuple.TargetTupleType().ElementFlags()
	elements := l.checker.GetTypeArguments(tuple)
	if len(flags) == 0 || flags[len(flags)-1] != checker.ElementFlagsRest {
		return 0, false
	}
	var shared ir.Type
	for index, element := range elements {
		if index < len(flags)-1 && flags[index] != checker.ElementFlagsRequired {
			return 0, false
		}
		of, known := l.representation(element)
		if !known || slotless(of) || of == ir.Weak || shared != 0 && shared != of {
			return 0, false
		}
		if of.IsReference() && element != elements[0] {
			return 0, false
		}
		shared = of
	}
	return shared, shared != 0
}

func (l *lowering) arrayTupleLiteral(node *ast.Node, tuple *checker.Type, element ir.Type) (ir.Expression, error) {
	literal := ir.ArrayLiteral{Element: element}
	items := node.AsArrayLiteralExpression().Elements.Nodes
	if len(items) < tuple.TargetTupleType().FixedLength() {
		return nil, l.notYet(node, "a rest tuple literal leaving out a required prefix element")
	}
	for _, item := range items {
		if item.Kind == ast.KindOmittedExpression || item.Kind == ast.KindSpreadElement {
			return nil, l.notYet(item, "an omitted or spread element in a rest tuple literal")
		}
		value, err := l.expression(item)
		if err != nil {
			return nil, err
		}
		value = fit(value, element)
		if value.Type() != element {
			return nil, l.notYet(item, "a rest tuple element with a different representation")
		}
		literal.Elements = append(literal.Elements, value)
	}
	return literal, nil
}
