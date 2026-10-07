package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A nullable callable has one absent value, null, represented by a null pointer.
// The caller refuses unions also containing undefined: distinguishing the two needs a tag.
func (l *lowering) nullableCallable(proven *checker.Type) bool {
	found := false
	for _, member := range proven.Types() {
		if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
			continue
		}
		held, known := l.representation(member)
		if !known || held != ir.Closure {
			return false
		}
		found = true
	}
	return found
}
