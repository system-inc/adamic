package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// clockGenericReturnsT04 retains all fields of a concrete object intersection.
// The nested union consists solely of plain shapes; readonly remains checker-enforced.
func (l *lowering) clockGenericReturnsT04(proven *checker.Type) (ir.Type, bool) {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsIntersection == 0 {
		return 0, false
	}
	// This hand-off owns the expression/typeArguments result, not other signatures.
	expression := l.checker.GetPropertyOfType(proven, "expression")
	arguments := l.checker.GetPropertyOfType(proven, "typeArguments")
	if expression == nil || arguments == nil || l.concrete(l.checker.GetTypeOfSymbol(expression)).Flags()&checker.TypeFlagsUnion == 0 || !l.checker.IsArrayType(l.concrete(l.checker.GetTypeOfSymbol(arguments))) {
		return 0, false
	}
	for _, part := range proven.Types() {
		if !l.clockT04PlainShape(l.concrete(part)) {
			return 0, false
		}
	}
	if !l.clockT04Fields(proven, 0) {
		return 0, false
	}
	return ir.Object, true
}

func (l *lowering) clockT04PlainShape(t *checker.Type) bool {
	return t.Flags()&checker.TypeFlagsObject != 0 && !l.checker.IsArrayType(t) && !checker.IsTupleType(t) && !isClassInstance(t) && !l.isLibraryType(t, "Map", "ReadonlyMap", "Set", "ReadonlySet") && len(l.checker.GetIndexInfosOfType(t)) == 0 && len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) == 0 && len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) == 0
}

func (l *lowering) clockT04Fields(t *checker.Type, depth int) bool {
	if depth > 16 || len(l.checker.GetIndexInfosOfType(t)) != 0 {
		return false
	}
	fields := l.checker.GetPropertiesOfType(t)
	if len(fields) == 0 {
		return false
	}
	for _, field := range fields {
		if !l.clockT04Field(l.concrete(l.checker.GetTypeOfSymbol(field)), depth+1) {
			return false
		}
	}
	return true
}

func (l *lowering) clockT04Field(t *checker.Type, depth int) bool {
	if depth > 16 {
		return false
	}
	if t.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 {
		return true
	}
	if l.checker.IsArrayType(t) && !checker.IsTupleType(t) {
		args := l.checker.GetTypeArguments(t)
		return len(args) == 1 && l.concrete(args[0]).Flags()&checker.TypeFlagsStringLike != 0
	}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range t.Types() {
			member = l.concrete(member)
			if !l.clockT04ObjectShape(member) || !l.clockT04Fields(member, depth+1) {
				return false
			}
		}
		return len(t.Types()) > 0
	}
	return l.clockT04ObjectShape(t) && l.clockT04Fields(t, depth+1)
}

// The checker distributes object & (A | B) to (object & A) | (object & B).
// The object keyword contributes no fields; each actual shape and every merged
// field still require proof. It is never used to erase an arbitrary constituent.
func (l *lowering) clockT04ObjectShape(t *checker.Type) bool {
	if t.Flags()&checker.TypeFlagsIntersection == 0 {
		return l.clockT04PlainShape(t)
	}
	shaped := false
	for _, part := range t.Types() {
		part = l.concrete(part)
		if part.Flags() == checker.TypeFlagsNonPrimitive {
			continue
		}
		if !l.clockT04PlainShape(part) {
			return false
		}
		shaped = true
	}
	return shaped && len(l.checker.GetIndexInfosOfType(t)) == 0 && len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) == 0 && len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) == 0
}
