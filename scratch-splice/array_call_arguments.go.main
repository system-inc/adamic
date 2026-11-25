package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// The receiver is saved first. Packing expands each spread before the next
// argument runs, including when that argument mutates the spread's source.
func (l *lowering) arrayInsertArguments(node, receiver *ast.Node, name string, written []*ast.Node) (ir.Expression, bool, error) {
	element, err := l.elementType(receiver)
	if err != nil {
		return nil, true, err
	}
	array, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments, spread, err := l.callArguments(written)
	if err != nil {
		return nil, true, err
	}
	for index, value := range arguments {
		if len(spread) > index && spread[index] {
			source := written[index].AsSpreadElement().Expression
			var from ir.Type
			if literal, ok := value.(ir.ArrayLiteral); ok {
				if len(literal.Elements) == 0 {
					literal.Element = element
					arguments[index] = literal
					continue
				}
				from = literal.Element
			} else {
				from, err = l.elementType(source)
				if err != nil {
					return nil, true, err
				}
			}
			if from != element {
				b := l.libraryArrayBuilder([]ir.Expression{value})
				packed := b.declare("spread_values", ir.ArrayLiteral{Element: element})
				item := b.local("spread_item", from)
				fitted := fit(b.read(item), element)
				if fitted.Type() != element {
					return nil, true, l.notYet(source, "spreading elements with an incompatible representation into "+name+" (convert the elements explicitly first)")
				}
				b.body = append(b.body, ir.ForOf{Iterable: b.read(b.parameters[0]), Element: from, Local: item, Body: []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{Array: b.read(packed), Value: fitted, Element: element, Site: l.libraryArraySyntheticWriteSite(node)}}}})
				arguments[index] = b.finish("array_spread_convert", b.read(packed))
			}
		} else {
			arguments[index] = fit(value, element)
			if arguments[index].Type() != element {
				return nil, true, l.notYet(written[index], name+" with an incompatible element representation (convert the value explicitly first)")
			}
		}
	}
	packed := ir.ArrayLiteral{Element: element, Elements: arguments, Spread: spread}
	b := l.libraryArrayBuilder([]ir.Expression{array, packed})
	target, items := b.read(b.parameters[0]), ir.Expression(b.read(b.parameters[1]))
	item := b.local("insert_item", element)
	var insertion ir.Expression = ir.ArrayPush{Array: target, Value: b.read(item), Element: element, Site: l.writeSite(receiver)}
	if name == "unshift" {
		items = ir.ArrayReverse{Array: items}
		insertion = ir.ArraySplice{Array: target, Start: ir.NumberConstant{Value: 0}, Count: ir.NumberConstant{Value: 0}, Items: []ir.Expression{b.read(item)}, Element: element, Site: l.writeSite(receiver)}
	}
	b.body = append(b.body, ir.ForOf{Iterable: items, Element: element, Local: item, Body: []ir.Statement{ir.Evaluate{Value: insertion}}})
	return b.finish("array_"+name+"_arguments", ir.Length{Array: target}), true, nil
}
