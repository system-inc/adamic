package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// clockArrayConstraint proves the array header ABI from declared array ancestry.
// Structural length properties and tuples do not prove this ABI. Array metadata
// remains unsupported by the existing array property lowering.
func (l *lowering) clockArrayConstraint(constraint *checker.Type, seen map[*checker.Type]bool) bool {
	if constraint == nil || seen[constraint] || checker.IsTupleType(constraint) {
		return false
	}
	seen[constraint] = true
	if l.checker.IsArrayType(constraint) {
		return true
	}
	if constraint.Flags()&checker.TypeFlagsObject != 0 && constraint.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		constraint = constraint.Target()
	}
	if constraint.Flags()&checker.TypeFlagsObject == 0 || constraint.ObjectFlags()&checker.ObjectFlagsInterface == 0 {
		return false
	}
	for _, base := range l.checker.GetBaseTypes(constraint) {
		if l.clockArrayConstraint(base, seen) {
			return true
		}
	}
	return false
}
