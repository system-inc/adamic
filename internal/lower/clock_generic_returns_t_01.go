package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// clockGenericReturnsT01 proves only an optional plain-object intersection. The
// checker retains every field (including refinements and readonly declarations);
// no constituent or type argument is erased to choose the pointer representation.
// Undefined uses the existing missing object pointer; null is never admitted.
func (lowering *lowering) clockGenericReturnsT01(result *checker.Type) (ir.Type, bool) {
	result = lowering.concrete(result)
	if result.Flags()&checker.TypeFlagsUnion == 0 || len(result.Types()) != 2 {
		return 0, false
	}
	var object *checker.Type
	missing := false
	for _, member := range result.Types() {
		member = lowering.concrete(member)
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			missing = true
		} else {
			object = member
		}
	}
	if !missing || object == nil || object.Flags()&checker.TypeFlagsIntersection == 0 {
		return 0, false
	}
	// This hand-off owns the generic object plus refinement signature, not the
	// other optional nongeneric intersection briefs.
	if len(object.Types()) != 2 {
		return 0, false
	}
	generic := false
	for _, part := range object.Types() {
		// Only a type reference has type arguments; the checker dereferences nil for anything
		// else, such as the string in a branded string & { __brand: void }.
		part = lowering.concrete(part)
		generic = generic || part.Flags()&checker.TypeFlagsObject != 0 &&
			part.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
			len(lowering.checker.GetTypeArguments(part)) != 0
	}
	if !generic {
		return 0, false
	}
	if !lowering.clockGenericReturnsT01Shape(object, map[*checker.Type]bool{}) {
		return 0, false
	}
	return ir.Object, true
}

func (lowering *lowering) clockGenericReturnsT01Shape(proven *checker.Type, active map[*checker.Type]bool) bool {
	proven = lowering.concrete(proven)
	if proven == nil || active[proven] {
		return false
	}
	flags := proven.Flags()
	if flags&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 {
		return true
	}
	if flags&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return false
	}
	active[proven] = true
	defer delete(active, proven)
	// Arrays are admitted only as nested string fields, using the existing string
	// array representation. Intersections of containers never pass this check.
	if lowering.isArrayType(proven) {
		arguments := lowering.checker.GetTypeArguments(proven)
		return len(arguments) == 1 && lowering.concrete(arguments[0]).Flags()&checker.TypeFlagsStringLike != 0
	}
	if checker.IsTupleType(proven) || isClassInstance(proven) ||
		len(lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 ||
		len(lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 ||
		len(lowering.checker.GetIndexInfosOfType(proven)) != 0 {
		return false
	}
	if symbol := proven.Symbol(); symbol != nil {
		for _, declaration := range symbol.Declarations {
			if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
				return false
			}
		}
	}
	if flags&checker.TypeFlagsIntersection != 0 {
		for _, part := range proven.Types() {
			part = lowering.concrete(part)
			if part.Flags()&checker.TypeFlagsObject == 0 ||
				lowering.isArrayType(part) ||
				!lowering.clockGenericReturnsT01Shape(part, active) {
				return false
			}
		}
	}
	fields := lowering.checker.GetPropertiesOfType(proven)
	if len(fields) == 0 {
		return false
	}
	for _, field := range fields {
		if !lowering.clockGenericReturnsT01Shape(lowering.checker.GetTypeOfSymbol(field), active) {
			return false
		}
	}
	return true
}
