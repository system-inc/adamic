package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Finite string keys on ordinary structural objects reuse the dictionary slot
// adapter. Each selected field still has its own declared result contract.
func (l *lowering) finiteDictionaryKeys(receiver, keys *checker.Type) bool {
	receiver = l.checker.GetNonNullableType(l.concrete(receiver))
	if receiver.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(receiver) ||
		l.checker.IsArrayType(receiver) || checker.IsTupleType(receiver) || len(l.checker.GetIndexInfosOfType(receiver)) != 0 {
		return false
	}
	for _, key := range castMembers(l.concrete(keys)) {
		if key.Flags()&checker.TypeFlagsStringLiteral == 0 {
			return false
		}
		name, ok := key.AsLiteralType().Value().(string)
		if !ok {
			return false
		}
		field := l.checker.GetPropertyOfType(receiver, name)
		if field == nil || accessorSymbol(field) {
			return false
		}
		for _, declaration := range field.Declarations {
			if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				return false
			}
		}
	}
	return true
}

// Optional named reads evaluate their receiver once and demand the field only
// on the present branch. The field's declaration, not the optional expression,
// decides whether a missing field itself is allowed.
func (l *lowering) optionalDictionaryField(node *ast.Node, object ir.Expression, name string) (ir.Expression, error) {
	if object.Type() != ir.Object && object.Type() != ir.Record {
		return nil, l.notYet(node, "an optional dictionary receiver without object storage")
	}
	declared := l.checker.GetTypeAtLocation(node)
	if field := l.checker.GetSymbolAtLocation(node.Name()); field != nil {
		declared = l.checker.GetTypeOfSymbol(field)
	}
	b := l.libraryArrayBuilder([]ir.Expression{object})
	receiver := b.read(b.parameters[0])
	selected, err := l.dictionaryReadContract(node, receiver, ir.StringConstant{Index: l.constant(name)}, l.concrete(declared))
	if err != nil {
		return nil, err
	}
	if !selected.Type().IsReference() || len(selected.(ir.Property).ViewAllowed) != 0 {
		return nil, l.notYet(node, "an optional dictionary field without a supported reference result")
	}
	present := ir.Unary{Operator: ir.Not, Operand: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: receiver}, Right: ir.IsNull{Value: receiver}}}
	return b.finish("optional_dictionary_field", ir.Conditional{Condition: present, WhenTrue: selected, WhenNot: fit(ir.Undefined{}, selected.Type()), Of: selected.Type()}), nil
}
