package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// clockGenericReturnsT03 proves an object intersection result without widening its
// fields. Readonly is enforced by the checker; the object retains every field.
func (l *lowering) clockGenericReturnsT03(proven *checker.Type) (ir.Type, bool) {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsIntersection == 0 {
		return 0, false
	}
	for _, part := range proven.Types() {
		part = l.concrete(part)
		if part.Flags()&checker.TypeFlagsObject == 0 || l.checker.IsArrayType(part) || checker.IsTupleType(part) || isClassInstance(part) || l.isLibraryType(part, "Map", "ReadonlyMap", "Set", "ReadonlySet") || len(l.checker.GetIndexInfosOfType(part)) != 0 || len(l.checker.GetSignaturesOfType(part, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(part, checker.SignatureKindConstruct)) != 0 {
			return 0, false
		}
	}
	fields := l.checker.GetPropertiesOfType(proven)
	if len(fields) == 0 || len(l.checker.GetIndexInfosOfType(proven)) != 0 {
		return 0, false
	}
	for _, field := range fields {
		of := l.concrete(l.checker.GetTypeOfSymbol(field))
		if of.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 {
			continue
		}
		if !l.checker.IsArrayType(of) || checker.IsTupleType(of) {
			return 0, false
		}
		arguments := l.checker.GetTypeArguments(of)
		if len(arguments) != 1 || l.concrete(arguments[0]).Flags()&checker.TypeFlagsStringLike == 0 {
			return 0, false
		}
	}
	return ir.Object, true
}
