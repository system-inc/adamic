package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// clockGenericReturnsT06 proves the concrete MemberName/literal result without
// dropping intersection fields or changing checker-enforced readonly properties.
func (l *lowering) clockGenericReturnsT06(t *checker.Type) (ir.Type, bool) {
	t = l.concrete(t)
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		intersection := false
		for _, part := range t.Types() {
			part = l.concrete(part)
			if !l.clockT06Shape(part) {
				return 0, false
			}
			intersection = intersection || part.Flags()&checker.TypeFlagsIntersection != 0
		}
		return ir.Object, intersection
	}
	if t.Flags()&checker.TypeFlagsIntersection == 0 || !l.clockT06Shape(t) {
		return 0, false
	}
	return ir.Object, true
}

func (l *lowering) clockT06Shape(t *checker.Type) bool {
	t = l.concrete(t)
	if t.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, part := range t.Types() {
			if !l.clockT06Plain(l.concrete(part)) {
				return false
			}
		}
	} else if !l.clockT06Plain(t) {
		return false
	}
	if len(l.checker.GetIndexInfosOfType(t)) != 0 {
		return false
	}
	kind, text := l.checker.GetPropertyOfType(t, "kind"), l.checker.GetPropertyOfType(t, "text")
	if kind == nil || text == nil {
		return false
	}
	if l.concrete(l.checker.GetTypeOfSymbol(kind)).Flags()&checker.TypeFlagsStringLike == 0 || l.concrete(l.checker.GetTypeOfSymbol(text)).Flags()&checker.TypeFlagsStringLike == 0 {
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(t) {
		ft := l.concrete(l.checker.GetTypeOfSymbol(field))
		if ft.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 {
			continue
		}
		// The only nested reference authorized here is a proven string array.
		if !l.checker.IsArrayType(ft) || checker.IsTupleType(ft) {
			return false
		}
		args := l.checker.GetTypeArguments(ft)
		if len(args) != 1 || l.concrete(args[0]).Flags()&checker.TypeFlagsStringLike == 0 {
			return false
		}
	}
	return true
}

func (l *lowering) clockT06Plain(t *checker.Type) bool {
	return t.Flags()&checker.TypeFlagsObject != 0 && !l.checker.IsArrayType(t) && !checker.IsTupleType(t) && !isClassInstance(t) && !l.isLibraryType(t, "Map", "ReadonlyMap", "Set", "ReadonlySet") && len(l.checker.GetIndexInfosOfType(t)) == 0 && len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) == 0 && len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) == 0
}
