package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// spreadOverwrite fits a field written after a leading spread, { ...source, k: value }, to the slot the
// spread brings for it, or stops when the write can't land in that slot. Natively the literal takes or copies
// the source object and writes k into the source's own slot (reuse.go's take path, emit_objects.go's copy), so
// two writes can't be kept (#xy74m5x, both found by family 65, where Node and the JavaScript backend agree):
//   - a value written as a reference into a slot the source holds as a scalar, or the other way: native's write
//     (emit_objects.go, reuse.go) releases the old slot as a reference exactly when the value is one, so a number
//     slot written as a reference released the number's bits as a pointer, exit 139 ({ ...source, k: undefined }
//     over { readonly k: number });
//   - a field optional in the source's type: an actual source without it has no slot, so the write panics,
//     exit 70 for a string and 139 for a number ({ ...source, k: undefined } over { readonly k?: string } given {}).
//
// Writes that keep the slot's kind are kept, as on main: a reference over a reference releases the old one and
// stores the new (undefined over a string slot agrees with Node), and a scalar over a scalar stores its member
// (boolean | undefined over a number slot, review/agree's fxspptb_oct9_native_p01_maybe_boolean_spread.ts). A
// source that may be undefined is kept too: its fallback object is laid out by the literal (reuse.go). A field the source's type doesn't declare may still
// be in the actual source by width, so it stops too. The fix is Optimizer's and Lower's: a spread that writes its
// own layout.
func (l *lowering) spreadOverwrite(literal *ast.Node, property *ast.Node, name string, value ir.Expression) (ir.Expression, error) {
	source := literal.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression
	field := l.checker.GetPropertyOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(source)), name)
	if field == nil {
		// By width the actual source may hold the field anyway ({ readonly a: number } given { a: 1, k: "x" }),
		// in a slot laid out for its own value. A property or a placeholder stops at its caller in these words
		// before it gets here; a method meets it here.
		return nil, l.notYet(property, "a spread that adds a field the source doesn't have")
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
	if value.Type().IsReference() != held.IsReference() {
		return nil, l.notYet(property, "an overwrite of "+name+" with "+typeName(value.Type())+" after a spread whose source holds "+name+" as "+typeName(held)+", so the write would take a scalar slot for a reference or a reference for a scalar")
	}
	return value, nil
}
