package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// primitiveIdentityType proves that an operand cannot be an adapter key. Any,
// unknown, object, callable and unresolved parameter types retain normalization.
func (l *lowering) primitiveIdentityType(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !l.primitiveIdentityType(member) {
				return false
			}
		}
		return len(proven.Types()) != 0
	}
	return proven.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0
}
