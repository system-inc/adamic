package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// checkedIndexedRead guards the lookup already emitted by ArrayIndex/StringIndex/record MapGet.
// Coalesce evaluates that lookup once and stops if its presence result is false;
// it is shared by the native and JavaScript backends. Observations of undefined
// retain their value. Migrated null sentinels are values, distinct from the
// absent reference NULL. Legacy nullable elements still need a separate slot.
func (l *lowering) checkedIndexedRead(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	if !l.program.RequiresIndexedPresenceChecks() || node.Kind != ast.KindElementAccessExpression || l.includesUndefined(l.concrete(l.checker.GetTypeAtLocation(node))) {
		return value, nil
	}
	if l.acceptsUndefined(node) && !l.program.RequiresIndexedSite(node) {
		return value, nil
	}
	// Optional destinations can be absent even after this read succeeds.
	// The area has fixed shapes, so reject rather than reach a missing-slot panic.
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression {
		assignment := parent.AsBinaryExpression()
		target := ast.SkipParentheses(assignment.Left)
		if assignment.Right == node && assignment.OperatorToken.Kind == ast.KindEqualsToken && target.Kind == ast.KindPropertyAccessExpression {
			if field := l.checker.GetSymbolAtLocation(target); field != nil && field.Flags&ast.SymbolFlagsOptional != 0 {
				return nil, l.notYet(target, "indexed result written to an optional property requires own-presence representation 5bb775ca and alias fixes through 7e7464e6")
			}
		}
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
	// Area typed-array reads carry absence through Unwrap. Reuse that lookup.
	if unwrapped, ok := lookup.(ir.Unwrap); ok {
		lookup = unwrapped.Value
	}
	var of ir.Type
	switch index := lookup.(type) {
	case ir.ArrayIndex:
		arguments := l.typeArguments(receiver)
		if !index.Element.UsesNullSentinel() && (l.includesNull(l.concrete(element)) || len(arguments) > 0 && l.includesNull(l.concrete(arguments[0]))) {
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
	return ir.Coalesce{UndefinedOnly: true, Value: lookup, Panic: ir.StringConstant{Index: l.constant(message)}, Of: of}, nil
}

// The contextual tuple of an array binding describes its destinations, not an
// alias of the array as fixed tuple storage. destructureArray checks each slot.
func (l *lowering) indexedArrayBindingInitializer(node *ast.Node) bool {
	if !l.program.RequiresIndexedPresenceChecks() || !l.checker.IsArrayType(l.checker.GetTypeAtLocation(node)) {
		return false
	}
	use := node
	for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
		use = use.Parent
	}
	parent := use.Parent
	return parent != nil && parent.Kind == ast.KindVariableDeclaration && parent.Name().Kind == ast.KindArrayBindingPattern && parent.Initializer() == use
}
