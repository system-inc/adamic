package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// relatedShape is the type a walk relates an allocation's shape by when it asks whether a value seen as
// another type may be that allocation. An object literal's own type is fresh, and the checker's relation
// judges a fresh type for excess properties: { name: "c", back: holder } never relates to { name: string },
// though a value seen as the latter may be it, and holds back (#sbvnh96, #yfmm1vv, as #jm2ba55 found in
// prototypeHazard). So a fresh object literal is related by its regular widened type, freshness taken off by
// the checker's own rule. Every other type is itself.
func (l *lowering) relatedShape(shape *checker.Type) *checker.Type {
	if shape.Flags()&checker.TypeFlagsObject == 0 || shape.ObjectFlags()&checker.ObjectFlagsFreshLiteral == 0 {
		return shape
	}
	return l.checker.GetWidenedType(regularTypeOfObjectLiteral(l.checker, shape))
}
