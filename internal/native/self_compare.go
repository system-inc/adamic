package native

import (
	"fmt"
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// Distinct parameters keep clang from diagnosing identical emitted operands.
// Number equality remains IEEE equality, including NaN and signed zero.
func (e *emitter) scalarEquality(operator ir.Operator, operandType ir.Type, left, right string) string {
	name, representation := "reference", "const void *"
	switch operandType {
	case ir.Number:
		name, representation = "number", "double"
	case ir.Boolean:
		name, representation = "boolean", "bool"
	}
	helper := "adamic_same_" + name
	declaration := fmt.Sprintf("static inline bool %s(%s left, %s right) { return left == right; }", helper, representation, representation)
	if !slices.Contains(e.declarations, declaration) {
		e.declarations = append(e.declarations, declaration)
	}
	comparison := fmt.Sprintf("%s(%s, %s)", helper, left, right)
	if operator == ir.NotEqual {
		return "(!" + comparison + ")"
	}
	return comparison
}
