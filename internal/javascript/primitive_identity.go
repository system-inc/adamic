package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) identityOperand(value ir.Expression, primitive bool) string {
	// These representations contain only primitives, including compiler-generated
	// comparisons without a source checker node. Union and Object are not proofs.
	switch value.Type() {
	case ir.Number, ir.String, ir.Boolean, ir.MaybeNumber, ir.MaybeBoolean:
		primitive = true
	}
	if primitive {
		return e.value(value)
	}
	return "adamicViewAdapterUnderlying(" + e.value(value) + ")"
}
