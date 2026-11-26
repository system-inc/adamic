package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// spreadOverwrite fits a field written after a leading spread, { ...source, k: value }, to the slot the
// spread brings for it, and says why the write can't land in that slot when it can't. Natively a spread
// whose fields all fit takes or copies the source object and writes k into the source's own slot (reuse.go's
// take path, emit_objects.go's copy). Three writes can't be kept that way (#xy74m5x, found by family 65,
// where Node and the JavaScript backend agree):
//   - a value written as a reference into a slot the source holds as a scalar, or the other way: the copy's
//     write releases the old slot as a reference exactly when the value is one, so a number slot written as
//     a reference released the number's bits as a pointer, exit 139 ({ ...source, k: undefined } over
//     { readonly k: number });
//   - a field optional in the source's type: an actual source without it has no slot, so the write panicked,
//     exit 70 for a string and 139 for a number ({ ...source, k: undefined } over { readonly k?: string }
//     given {});
//   - a field the source's type doesn't declare, which by width the actual source may still hold, in a slot
//     laid out for its own value ({ readonly a: number } given { a: 1, k: "x" }).
//
// For each the reason is returned, in the words the stop said, and the caller decides: a literal of fields
// and methods lowers to adamic_object_spread_with instead (ir.ObjectLiteral's SpreadWith, #k7g6f9h), which
// builds the result's shape at run time from the actual source's own keys plus the overwrites, and a
// placeholder or a source that may be undefined stops with the reason, as before.
//
// Writes that keep the slot's kind fit it, as on main: a reference over a reference releases the old one and
// stores the new (undefined over a string slot agrees with Node), and a scalar over a scalar stores its member
// (boolean | undefined over a number slot, review/agree's fxspptb_oct9_native_p01_maybe_boolean_spread.ts).
func (l *lowering) spreadOverwrite(literal *ast.Node, name string, value ir.Expression) (ir.Expression, string, error) {
	source := literal.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression
	field := l.checker.GetPropertyOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(source)), name)
	if field == nil {
		return value, spreadAddsField, nil
	}
	if field.Flags&ast.SymbolFlagsOptional != 0 {
		return value, "an overwrite of " + name + " after a spread whose source may not have it (its type makes " + name + " optional), so the copied object may have no slot to write", nil
	}
	held, err := l.typeOfSymbol(source, field)
	if err != nil {
		return nil, "", err
	}
	if held == ir.Union {
		// Spread keeps its source shape. Overrides must retain that slot's counted representation rather
		// than write an unboxed scalar into it.
		value = fit(value, held)
	}
	if value.Type().IsReference() != held.IsReference() {
		return value, "an overwrite of " + name + " with " + typeName(value.Type()) + " after a spread whose source holds " + name + " as " + typeName(held) + ", so the write would take a scalar slot for a reference or a reference for a scalar", nil
	}
	return value, "", nil
}

// spreadAddsField is the reason a field the spread's source type doesn't declare can't be written into the
// copy: by width the actual source may hold it anyway, in a slot laid out for its own value.
const spreadAddsField = "a spread that adds a field the source doesn't have"

// spreadWithAccessorStop is what a SpreadWith literal says when its source may hold an accessor Node's spread
// copies as an own key, an object literal's own, or one named as a class's public field (spreadAccessorClass):
// adamic_object_spread_with passes over an accessor, so lowering stops the spread first wherever one may reach it.
const spreadWithAccessorStop = "a spread with an overwrite its source's slots can't hold, of a source that may hold an accessor as an own key, which adamic_object_spread_with doesn't copy yet (#vjvcnnj)"
