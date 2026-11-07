package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// arrayOrUndefined proves the only falsy member is undefined. Do not extend this
// to strings or numbers: their empty and zero values differ from ?? semantics.
func (lowering *lowering) arrayOrUndefined(proven *checker.Type) bool {
	proven = lowering.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if member.Flags()&checker.TypeFlagsUndefined == 0 && !lowering.arrayOrUndefined(member) {
				return false
			}
		}
		return true
	}
	return lowering.isArrayType(proven)
}
