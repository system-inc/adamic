package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// One structural member has an unambiguous reference discriminator. Its fields
// remain lazy obligations in the shared contract graph, including recursive ones.
// Array members reuse the shared element adapter. Callables and collections
// still require their owning member adapters.
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
		if of == ir.Array && l.checker.IsArrayType(member) && !checker.IsTupleType(member) {
			objects++
		} else if of == ir.Object && member.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(member) && len(l.checker.GetIndexInfosOfType(member)) == 0 {
			objects++
		} else if interfaceScalar(member) && (of == ir.String || of == ir.Boolean || of == ir.Number) {
			primitives++
		} else {
			return false
		}
	}
	return objects == 1 && primitives > 0
}

// Boxed unions already have a native reference slot and real heap tags. Permit
// their production only for the members this adapter can read; no widening or
// assertion establishes a member's contract here.
func (l *lowering) objectPrimitiveBoxedField(property *ast.Node) bool {
	if property.Kind != ast.KindPropertyAssignment {
		return false
	}
	target := l.checker.GetTypeAtLocation(property.AsPropertyAssignment().Initializer)
	if l.objectPrimitiveViewType(target) {
		return true
	}
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	for _, member := range target.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		of, known := l.representation(member)
		if !known || !interfaceScalar(member) || (of != ir.Number && of != ir.String && of != ir.Boolean) {
			return false
		}
	}
	return true
}

// A structural conjunction still has an object reference slot even when its
// fields contain aggregates. This proves storage only, never its conjunction.
func (l *lowering) objectPrimitiveIntersectionStorage(target *checker.Type) bool {
	if target.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range target.Types() {
			if !l.objectPrimitiveIntersectionStorage(member) {
				return false
			}
		}
		return len(target.Types()) != 0
	}
	return target.Flags()&checker.TypeFlagsObject != 0 && !l.checker.IsArrayType(target) && !checker.IsTupleType(target) && !isClassInstance(target) && !l.isLibraryType(target, "Map", "ReadonlyMap", "Set", "ReadonlySet") && len(l.checker.GetIndexInfosOfType(target)) == 0 && len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) == 0
}
