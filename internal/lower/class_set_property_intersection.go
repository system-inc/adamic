package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An intersection refines a view without changing its object's slots. Every
// constituent declaring this reference or presence pair must agree on how that field is held.
func (l *lowering) intersectionKeptSlot(proven *checker.Type, name string, of ir.Type) bool {
	if !of.IsReference() && !of.IsMaybe() {
		return false
	}
	for _, part := range proven.Types() {
		field := l.checker.GetPropertyOfType(part, name)
		if field == nil {
			continue
		}
		held, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || held != of {
			return false
		}
	}
	return true
}
