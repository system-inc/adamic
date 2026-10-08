package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// One structural member has an unambiguous reference discriminator. Its fields
// remain lazy obligations in the shared contract graph, including recursive ones.
// Arrays, callables and collections still require their owning member adapters.
func (l *lowering) objectPrimitiveViewType(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	objects, primitives := 0, 0
	for _, member := range target.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		of, known := l.representation(member)
		if !known {
			return false
		}
		if of == ir.Object && member.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(member) && len(l.checker.GetIndexInfosOfType(member)) == 0 {
			objects++
		} else if interfaceScalar(member) && (of == ir.String || of == ir.Boolean || of == ir.Number) {
			primitives++
		} else {
			return false
		}
	}
	return objects == 1 && primitives > 0
}
