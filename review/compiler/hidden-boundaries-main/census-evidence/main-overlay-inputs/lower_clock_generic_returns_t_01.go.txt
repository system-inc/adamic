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
func (l *lowering) clockGenericReturnsT01(result *checker.Type) (ir.Type, bool) {
	result = l.concrete(result)
	if result.Flags()&checker.TypeFlagsUnion == 0 || len(result.Types()) != 2 {
		return 0, false
	}
	var object *checker.Type
	missing := false
	for _, member := range result.Types() {
		member = l.concrete(member)
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
		part = l.concrete(part)
		generic = generic || part.Flags()&checker.TypeFlagsObject != 0 && part.ObjectFlags()&checker.ObjectFlagsReference != 0 && len(l.checker.GetTypeArguments(part)) != 0
	}
	if !generic {
		return 0, false
	}
	if !l.clockGenericReturnsT01Shape(object, map[*checker.Type]bool{}) {
		return 0, false
	}
	return ir.Object, true
}

func (l *lowering) clockGenericReturnsT01Shape(proven *checker.Type, active map[*checker.Type]bool) bool {
	proven = l.concrete(proven)
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
	if l.checker.IsArrayType(proven) {
		arguments := l.checker.GetTypeArguments(proven)
		return len(arguments) == 1 && l.concrete(arguments[0]).Flags()&checker.TypeFlagsStringLike != 0
	}
	if checker.IsTupleType(proven) || isClassInstance(proven) ||
		len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 ||
		len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 ||
		len(l.checker.GetIndexInfosOfType(proven)) != 0 {
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
			part = l.concrete(part)
			if part.Flags()&checker.TypeFlagsObject == 0 || l.checker.IsArrayType(part) || !l.clockGenericReturnsT01Shape(part, active) {
				return false
			}
		}
	}
	fields := l.checker.GetPropertiesOfType(proven)
	if len(fields) == 0 {
		return false
	}
	for _, field := range fields {
		if !l.clockGenericReturnsT01Shape(l.checker.GetTypeOfSymbol(field), active) {
			return false
		}
	}
	return true
}
