package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// optionalVoidResult recognizes undefined added by an optional receiver to a void
// signature. Its result may be discarded as for the same nonoptional call. It does
// not prove that a void callable returns undefined when its value is observed.
func optionalVoidResult(result *checker.Type) bool {
	if result.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	hasVoid := false
	for _, member := range result.Types() {
		hasVoid = hasVoid || member.Flags()&checker.TypeFlagsVoid != 0
		if member.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined) == 0 {
			return false
		}
	}
	return hasVoid
}
