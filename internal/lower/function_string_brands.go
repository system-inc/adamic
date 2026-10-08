package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A proven primitive string with only void phantom markers has the same result
// lane as a string. Reuse the Map-key proof, including primitive member collisions,
// callable/indexed brands and never fields. Undefined is the string's NULL sentinel.
// This hook belongs only to signatures; ordinary branded slots stay unsupported.
func (l *lowering) stringBrandSignature(proven *checker.Type) (ir.Type, bool) {
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	present := false
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		held, known := l.mapKeyRepresentation(member)
		if !known || held != ir.String {
			return 0, false
		}
		present = true
	}
	return ir.String, present
}
