package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// slotNullish preserves the value's static nullish contract after its checker
// type is erased. Boxed references already distinguish null from undefined.
func (l *lowering) slotNullish(proven *checker.Type, of ir.Type) ir.Type {
	switch of {
	case ir.Object, ir.Array, ir.Map, ir.Closure, ir.Promise, ir.Uint8Array, ir.Int32Array, ir.Float64Array, ir.Uint16Array, ir.Generator:
	default:
		return 0
	}
	if proven == nil {
		return 0
	}
	proven = l.concrete(proven)
	null, undefined := l.includesNull(proven), l.includesUndefined(proven)
	if null && undefined {
		panic("compiler bug: a bare-pointer write admits both null and undefined, which representation boxes as a Union since #7yt238q")
	}
	if null {
		return ir.NullRepresentation
	}
	if undefined {
		return ir.UndefinedRepresentation
	}
	return 0
}

// fieldNullish reads the written expression, or a shorthand property's inferred
// value type, so a narrowed value keeps its non-nullable write. Synthetic literal
// values need no node.
func (l *lowering) fieldNullish(node *ast.Node, value ir.Expression) ir.Type {
	if node == nil || value == nil {
		return 0
	}
	if node.Kind == ast.KindPropertyAssignment {
		node = node.AsPropertyAssignment().Initializer
	}

	return l.slotNullish(l.checker.GetTypeAtLocation(node), value.Type())
}
