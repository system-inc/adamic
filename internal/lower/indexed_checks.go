package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// checkedIndexedRead guards the lookup already emitted by ArrayIndex/StringIndex/record MapGet.
// Coalesce evaluates that lookup once and stops if its presence result is false;
// it is shared by the native and JavaScript backends. Observations of undefined
// retain their value. A null element needs a separate presence slot and refuses
// here until that representation is carried through this check.
func (l *lowering) checkedIndexedRead(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	if !l.program.RequiresIndexedPresenceChecks() || node.Kind != ast.KindElementAccessExpression || l.includesUndefined(l.concrete(l.checker.GetTypeAtLocation(node))) {
		return value, nil
	}
	if l.acceptsUndefined(node) && !l.program.RequiresIndexedSite(node) {
		return value, nil
	}
	return l.indexedPresenceGuard(node, value, l.checker.GetTypeAtLocation(node), l.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression))
}

// indexedPresenceGuard is shared by explicit subscripts and implicit array
// binding reads. Both keep the lookup's absence result until Coalesce checks it.
func (l *lowering) indexedPresenceGuard(node *ast.Node, value ir.Expression, element, receiver *checker.Type) (ir.Expression, error) {
	lookup := value
	if defined, ok := lookup.(ir.Defined); ok {
		lookup = defined.Value
	}
	var of ir.Type
	switch index := lookup.(type) {
	case ir.ArrayIndex:
		arguments := l.typeArguments(receiver)
		if l.includesNull(l.concrete(element)) || len(arguments) > 0 && l.includesNull(l.concrete(arguments[0])) {
			return nil, l.notYet(node, "an indexed presence check on nullable array elements (the lookup needs to retain its presence slot)")
		}
		if index.Element == ir.Union || index.Element == ir.Weak {
			return nil, l.notYet(node, "an indexed presence check on tagged or weak array elements")
		}
		of = index.Element.Present()
	case ir.MapGet:
		if index.Map.Type() != ir.Record {
			return nil, l.notYet(node, "an indexed presence check for this representation")
		}
		of = index.ValueType.Present()
	case ir.StringIndex:
		of = ir.String
	default:
		// Tuple fields have static positions checked by TypeScript itself.
		if checker.IsTupleType(l.checker.GetNonNullableType(receiver)) {
			return value, nil
		}
		return nil, l.notYet(node, "an indexed presence check for this representation")
	}
	message := "indexed read is absent: " + l.program.Where(node)
	return ir.Coalesce{Value: lookup, Panic: ir.StringConstant{Index: l.constant(message)}, Of: of}, nil
}
