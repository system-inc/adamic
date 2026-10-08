package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Keys depend on actual own-property presence, never on element conformance.
// The runtime selects the existing record table or fixed-object key adapter.
func (l *lowering) dictionaryKeysCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	args := node.AsCallExpression().Arguments.Nodes
	if name != "keys" || len(args) != 1 || !l.stringDictionary(l.checker.GetTypeAtLocation(args[0])) {
		return nil, false, nil
	}
	receiver, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	if receiver.Type() != ir.Object && receiver.Type() != ir.Record {
		return nil, true, l.notYet(node, "dictionary keys without object storage")
	}
	return ir.RecordCall{Method: "keys", Arguments: []ir.Expression{receiver}, Returns: ir.Array, DictionaryKeys: true, ViewWhere: sourceExpression(node)}, true, nil
}

// Values and entries snapshot actual keys, then use the same checked element
// probe as an explicit lookup. Nested values keep their ordinary view obligations.
func (l *lowering) dictionaryValuesCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	args := node.AsCallExpression().Arguments.Nodes
	if (name != "values" && name != "entries") || len(args) != 1 {
		return nil, false, nil
	}
	target := l.concrete(l.checker.GetTypeAtLocation(args[0]))
	if !l.stringDictionary(target) && l.finitePartialRecordElement(target) == nil {
		return nil, false, nil
	}
	element := l.finitePartialRecordElement(target)
	if element == nil {
		element = l.checker.GetIndexInfosOfType(target)[0].ValueType()
	}
	of, known := l.representation(element)
	if !known {
		return nil, true, l.lazyReadRefusal(node, sourceExpression(node), "dictionary enumeration element")
	}
	receiver, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	b := l.libraryArrayBuilder([]ir.Expression{receiver})
	object := b.read(b.parameters[0])
	keys := b.declare("keys", ir.RecordCall{Method: "keys", Arguments: []ir.Expression{object}, Returns: ir.Array, DictionaryKeys: true})
	key := b.local("key", ir.String)
	selected, err := l.dictionaryReadContract(node, object, b.read(key), l.concrete(element))
	if err != nil {
		return nil, true, err
	}
	read := selected.(ir.Property)
	read.Of = of
	read.View = sourceExpression(node) + "[key]"
	var value ir.Expression = read
	output := of
	if name == "entries" {
		if of == ir.Array {
			arguments := l.checker.GetTypeArguments(l.checker.GetTypeAtLocation(node))
			if len(arguments) != 1 {
				return nil, true, l.notYet(node, "dictionary entries without a tuple result")
			}
			if _, err := l.viewContract(node, arguments[0]); err != nil {
				return nil, true, err
			}
		}
		value = ir.ObjectLiteral{Tuple: true, Fields: []ir.Field{{Name: "0", Value: b.read(key)}, {Name: "1", Value: value}}}
		output = ir.Object
	}
	result := b.declare("result", ir.ArrayLiteral{Element: output})
	if name == "entries" && of == ir.Array {
		// A private result contains checked child views, not certified plain tuples.
		l.result.DictionaryEntryOrigins = append(l.result.DictionaryEntryOrigins, b.read(result))
	}
	b.body = append(b.body, ir.ForOf{Iterable: b.read(keys), Local: key, Element: ir.String, Body: []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{DictionaryProduction: true, Array: b.read(result), Value: value, Element: output, Site: l.writeSite(node)}}}})
	return b.finish("dictionary_"+name, b.read(result)), true, nil
}

// Reuse the common reachability graph for derived view containers, while keeping
// the actual cast-origin ledger unchanged. Plain programs still have no views.
func dictionaryEnumerationOrigins(program *ir.Program) []ir.Expression {
	origins := append([]ir.Expression(nil), program.ViewOrigins...)
	return append(origins, program.DictionaryEntryOrigins...)
}
