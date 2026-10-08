package lower

import "github.com/system-inc/adamic/internal/ir"

// Optional scalar fields now use MaybeOf or Box even for constant values.
// Unpack only constants, preserving the old representation and effect guard.
func nodeFSFileOptionConstant(value ir.Expression) (ir.Expression, bool) {
	if boxed, wrapped := value.(ir.Box); wrapped {
		return nodeFSFileOptionConstant(boxed.Value)
	}
	if maybe, wrapped := value.(ir.MaybeOf); wrapped {
		if maybe.Value == nil {
			return ir.Undefined{}, true
		}
		return nodeFSFileOptionConstant(maybe.Value)
	}
	switch value.(type) {
	case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Undefined, ir.Null:
		return value, true
	}
	return nil, false
}
