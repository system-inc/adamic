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
	packed, err := l.arrayPackedValues(node, name, element, written, arguments, spread)
	if err != nil {
		return nil, true, err
	}
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

// arrayPackedValues packs the values a call writes into an array, callArguments' lowering of written,
// into one array literal of the array's elements, each spread in its place: a spread of the same
// elements as it is, one of another representation converted element by element where each fits, and
// a plain value fitted. push, unshift and splice share it, so a spread reads the same in each.
func (l *lowering) arrayPackedValues(node *ast.Node, name string, element ir.Type, written []*ast.Node, arguments []ir.Expression, spread []bool) (ir.ArrayLiteral, error) {
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
				var err error
				from, err = l.elementType(source)
				if err != nil {
					return ir.ArrayLiteral{}, err
				}
			}
			if from != element {
				converted, err := l.spreadConverted(node, source, value, from, element, "spreading elements with an incompatible representation into "+name+" (convert the elements explicitly first)")
				if err != nil {
					return ir.ArrayLiteral{}, err
				}
				arguments[index] = converted
			}
		} else {
			arguments[index] = fit(value, element)
			if arguments[index].Type() != element {
				return ir.ArrayLiteral{}, l.notYet(written[index], name+" with an incompatible element representation (convert the value explicitly first)")
			}
		}
	}
	return ir.ArrayLiteral{Element: element, Elements: arguments, Spread: spread}, nil
}

// spreadConverted is a spread's elements, held as from, made a new array of element, each fitted on the
// way: an array_spread_convert helper whose for-of pushes fit(item, element), its push's write site at
// site. Where fit can't make an element of one, it stops as NotYet at stop, in what's words.
// arrayPackedValues converts a spread into an array's elements with it, and spreadPackFit a spread into
// the slots its callee reads, so a spread converts the same on both roads.
func (l *lowering) spreadConverted(site, stop *ast.Node, value ir.Expression, from, element ir.Type, what string) (ir.Expression, error) {
	b := l.libraryArrayBuilder([]ir.Expression{value})
	packed := b.declare("spread_values", ir.ArrayLiteral{Element: element})
	item := b.local("spread_item", from)
	fitted := fit(b.read(item), element)
	if fitted.Type() != element {
		return nil, l.notYet(stop, what)
	}
	b.body = append(b.body, ir.ForOf{Iterable: b.read(b.parameters[0]), Element: from, Local: item, Body: []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{Array: b.read(packed), Value: fitted, Element: element, Site: l.libraryArraySyntheticWriteSite(site)}}}})
	return b.finish("array_spread_convert", b.read(packed)), nil
}

// arraySpliceSpreads says whether a splice spreads among its items, the third argument on. tsc's own
// insertStatementsAfterPrologue writes to.splice(statementIndex, 0, ...from) (#p1qw4kf); a splice
// without one keeps arrayMethodArguments' case, as it lowered before.
func arraySpliceSpreads(written []*ast.Node) bool {
	for _, item := range written[min(2, len(written)):] {
		if item.Kind == ast.KindSpreadElement {
			return true
		}
	}
	return false
}

// arraySpliceArguments is splice with a spread among its items (#p1qw4kf). As JavaScript does, the
// receiver, the start, the count and the items, each spread expanded by push's packing, are all
// evaluated before the array changes, so a.splice(1, 0, ...a) inserts a as it was. The helper resolves
// the insertion index from the length before anything is removed, as splice does, removes with one
// plain splice, whose removed elements are the call's result, then inserts each packed item at the
// index plus its position, as unshift inserts. The removal's start is the resolved index, which the
// runtime's clamping leaves as it is.
func (l *lowering) arraySpliceArguments(node, receiver *ast.Node, written []*ast.Node) (ir.Expression, bool, error) {
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
	if spread[0] || spread[1] || arguments[0].Type() != ir.Number || arguments[1].Type() != ir.Number {
		return nil, true, l.notYet(node, "splice without a start and a count that are numbers")
	}
	packed, err := l.arrayPackedValues(node, "splice", element, written[2:], arguments[2:], spread[2:])
	if err != nil {
		return nil, true, err
	}
	b := l.libraryArrayBuilder([]ir.Expression{array, arguments[0], arguments[1], packed})
	target, start, items := b.read(b.parameters[0]), b.read(b.parameters[1]), b.read(b.parameters[3])
	zero, one := ir.NumberConstant{Value: 0}, ir.NumberConstant{Value: 1}
	length := b.declare("length", ir.Length{Array: target})
	integer := b.declare("integer", ir.Conditional{Condition: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{start}}, WhenTrue: zero, WhenNot: ir.MathCall{Function: "trunc", Arguments: []ir.Expression{start}}})
	index := b.declare("index", ir.Conditional{Condition: ir.Binary{Operator: ir.Less, Left: b.read(integer), Right: zero}, WhenTrue: ir.MathCall{Function: "max", Arguments: []ir.Expression{ir.Binary{Operator: ir.Add, Left: b.read(length), Right: b.read(integer)}, zero}}, WhenNot: ir.MathCall{Function: "min", Arguments: []ir.Expression{b.read(integer), b.read(length)}}})
	count := b.declare("item_count", ir.Length{Array: items})
	position := b.declare("position", zero)
	removed := b.local("removed", ir.Array)
	removal := ir.Declare{Local: removed, Value: ir.ArraySplice{Array: target, Start: b.read(index), Count: b.read(b.parameters[2]), Element: element, Site: l.writeSite(receiver)}}
	item := ir.Expression(ir.ArrayIndex{Array: items, Index: b.read(position), Element: element})
	if item.Type().IsMaybe() && item.Type() != element {
		item = ir.Unwrap{Value: item}
	}
	at := ir.Binary{Operator: ir.Add, Left: b.read(index), Right: b.read(position)}
	insert := ir.Evaluate{Value: ir.ArraySplice{Array: target, Start: at, Count: zero, Items: []ir.Expression{item}, Element: element, Site: l.writeSite(receiver)}}
	insertion := ir.Loop{Condition: ir.Binary{Operator: ir.Less, Left: b.read(position), Right: b.read(count)}, Body: []ir.Statement{insert}, Update: []ir.Statement{ir.Assign{Local: position, Value: ir.Binary{Operator: ir.Add, Left: b.read(position), Right: one}}}}
	b.body = append(b.body, removal, insertion)
	return b.finish("array_splice_arguments", b.read(removed)), true, nil
}
