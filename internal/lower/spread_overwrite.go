package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// spreadOverwrite fits a field written after a leading spread, { ...source, k: value }, to the slot the
// spread brings for it, or stops when the write can't land in that slot. Natively the literal takes or copies
// the source object and writes k into the source's own slot (reuse.go's take path, emit_objects.go's copy), so
// two writes can't be kept (#xy74m5x, both found by family 65, where Node and the JavaScript backend agree):
//   - a slot the source holds otherwise than the value: a number slot written as a reference released the
//     number's bits as a pointer, exit 139 ({ ...source, k: undefined } over { readonly k: number });
//   - a field optional in the source's type: an actual source without it has no slot, so the write panics,
//     exit 70 for a string and 139 for a number ({ ...source, k: undefined } over { readonly k?: string } given {}).
//
// A source that may be undefined is kept: its fallback object is laid out by the literal (reuse.go), and with
// the representations equal its slot and the source's agree. A field the source's type doesn't declare is
// already a NotYet at its caller. The fix is Optimizer's and Lower's: a spread that writes its own layout.
func (l *lowering) spreadOverwrite(literal *ast.Node, property *ast.Node, name string, value ir.Expression) (ir.Expression, error) {
	source := literal.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression
	field := l.checker.GetPropertyOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(source)), name)
	if field == nil {
		return value, nil
	}
	if field.Flags&ast.SymbolFlagsOptional != 0 {
		return nil, l.notYet(property, "an overwrite of "+name+" after a spread whose source may not have it (its type makes "+name+" optional), so the copied object may have no slot to write")
	}
	held, err := l.typeOfSymbol(source, field)
	if err != nil {
		return nil, err
	}
	if held == ir.Union {
		// Spread keeps its source shape. Overrides must retain that slot's counted representation rather
		// than write an unboxed scalar into it.
		value = fit(value, held)
	}
	if value.Type() != held {
		return nil, l.notYet(property, "an overwrite of "+name+" with "+typeName(value.Type())+" after a spread whose source holds "+name+" as "+typeName(held)+", so the write would land in a slot laid out for the other")
	}
	return value, nil
}
