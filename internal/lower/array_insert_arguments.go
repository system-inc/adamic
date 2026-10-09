package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Packing arguments snapshots each spread before evaluating the next argument.
// The receiver and packed arguments are held once at the ordinary call boundary.
func (l *lowering) arrayInsertArguments(node, receiver *ast.Node, array ir.Expression, element ir.Type, name string, written []*ast.Node) (ir.Expression, bool, error) {
	values, spreads, err := l.callArguments(written)
	if err != nil {
		return nil, true, err
	}
	for i := range values {
		if len(spreads) > i && spreads[i] {
			source := written[i].AsSpreadElement().Expression
			if tuple := ast.SkipParentheses(source); tuple.Kind != ast.KindArrayLiteralExpression {
				other, err := l.elementType(source)
				if err != nil || other != element {
					return nil, true, l.notYet(source, "spreading elements with a different storage type into "+name)
				}
			} else if literal, ok := values[i].(ir.ArrayLiteral); ok && literal.Element != element {
				return nil, true, l.notYet(source, "spreading elements with a different storage type into "+name)
			}
		} else {
			values[i] = fit(values[i], element)
			if values[i].Type() != element {
				return nil, true, l.notYet(written[i], name+" with a value of another type than the elements")
			}
		}
	}
	b := l.libraryArrayBuilder([]ir.Expression{array, ir.ArrayLiteral{Element: element, Elements: values, Spread: spreads}})
	target, items := b.read(b.parameters[0]), b.read(b.parameters[1])
	if name == "unshift" {
		items = ir.ArrayReverse{Array: items}
	}
	item := b.local("inserted", element)
	var insert ir.Expression = ir.ArrayPush{Array: target, Value: b.read(item), Element: element, Site: l.writeSite(receiver)}
	if name == "unshift" {
		insert = ir.ArraySplice{Array: target, Start: ir.NumberConstant{Value: 0}, Count: ir.NumberConstant{Value: 0}, Items: []ir.Expression{b.read(item)}, Element: element, Site: l.writeSite(receiver)}
	}
	b.body = append(b.body, ir.ForOf{Iterable: items, Local: item, Element: element, Body: []ir.Statement{ir.Evaluate{Value: insert}}})
	return b.finish("array_"+name+"_arguments", ir.Length{Array: target}), true, nil
}
