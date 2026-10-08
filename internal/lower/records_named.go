package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Dictionary entries have one storage kind. Heterogeneous values use the existing
// counted union boxes. Optional booleans use those boxes at the record boundary.
func (l *lowering) recordStorageType(t *checker.Type) (ir.Type, bool) {
	of, known := l.kept(t)
	if of == ir.MaybeBoolean {
		return ir.Union, known
	}
	return of, known && of != 0 && of != ir.Weak
}

// The checker resolves literal keys to their named declaration and dynamic keys
// to the index signature. Never interpret a boxed index value as an unboxed field.
func recordTagObservation(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return comparedWithUndefined(node) || parent != nil && parent.Kind == ast.KindTypeOfExpression
}

func (l *lowering) recordReadType(node *ast.Node, read ir.Expression, here *checker.Type) (ir.Expression, error) {
	if recordTagObservation(node) {
		return read, nil // These operations observe the runtime tag rather than unbox it.
	}
	to, known := l.representation(here)
	if !known || read.Type() != ir.Union || to == ir.Union {
		return read, nil
	}
	name := ""
	switch to.Present() {
	case ir.Number:
		name = "number"
	case ir.Boolean:
		name = "boolean"
	case ir.String:
		name = "string"
	case ir.Closure:
		name = "function"
	default:
		return nil, l.notYet(node, "a dictionary member narrowed to an object kind without a checked runtime tag")
	}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	value := b.read(b.parameters[0])
	matches := ir.Expression(ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: value}, Right: ir.StringConstant{Index: l.constant(name)}})
	if l.includesUndefined(here) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: value}}
	}
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant("dictionary member does not have its declared type")}}}})
	return b.finish("record_declared_member", ir.Narrow{Value: value, To: to}), nil
}

// An unrestricted key can hit a narrower named slot. Preserve that slot's
// contract instead of trusting the checker's permissive index assignment.
func (l *lowering) recordNamedWrite(access, value *ast.Node) error {
	access = ast.SkipParentheses(access)
	if access.Kind != ast.KindElementAccessExpression {
		return nil
	}
	a := access.AsElementAccessExpression()
	key := ast.SkipParentheses(a.ArgumentExpression)
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral || key.Kind == ast.KindNumericLiteral {
		return nil // The checker already applies the named property's write type.
	}
	from := l.checker.GetTypeAtLocation(value)
	for _, property := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(a.Expression)) {
		to := l.checker.GetNonMissingTypeOfSymbol(property)
		if l.checker.IsReadonlySymbol(property) || !l.checker.IsTypeAssignableTo(from, to) || !l.enumAssignable(from, to) || l.widened(from, to, map[[2]*checker.Type]bool{}) != nil || !l.sameRecordStorage(from, to, map[[2]*checker.Type]bool{}) {
			return l.notYet(access, "an unrestricted dictionary write that can violate named property "+property.Name+"; use a literal key or a value accepted by every named property")
		}
	}
	return nil
}

// Erasing or widening a named contract would allow another alias to write through
// the index signature. Storage identity alone is not a proof of that relation.
func (l *lowering) sameRecordNamedContracts(from, to, inside, viewed *checker.Type, visited map[[2]*checker.Type]bool) bool {
	for _, pair := range [][3]*checker.Type{{from, to, viewed}, {to, from, inside}} {
		if len(l.checker.GetIndexInfosOfType(pair[0])) == 0 {
			continue // Finite partial records retain their existing element invariance rule.
		}
		for _, property := range l.checker.GetPropertiesOfType(pair[0]) {
			other := l.checker.GetPropertyOfType(pair[1], property.Name)
			t := l.checker.GetNonMissingTypeOfSymbol(property)
			if other == nil {
				if l.checker.IsReadonlySymbol(property) || !l.checker.IsTypeAssignableTo(pair[2], t) {
					return false
				}
				continue
			}
			u := l.checker.GetNonMissingTypeOfSymbol(other)
			if l.checker.IsReadonlySymbol(property) != l.checker.IsReadonlySymbol(other) || !l.checker.IsTypeAssignableTo(t, u) || !l.checker.IsTypeAssignableTo(u, t) || !l.sameRecordStorage(t, u, visited) || l.widened(t, u, map[[2]*checker.Type]bool{}) != nil || l.widened(u, t, map[[2]*checker.Type]bool{}) != nil {
				return false
			}
		}
	}
	return true
}
