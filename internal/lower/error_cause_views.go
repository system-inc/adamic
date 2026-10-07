package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An opaque value can be held and classified without reflecting its properties.
// After all causes are known, reject reflection that could encounter a value
// whose native descriptors are incomplete, including through an unknown alias.
func (l *lowering) checkErrorCauseReflection() error {
	errors := l.instances["builtin-error:Error"]
	if errors == nil {
		return nil
	}
	seen := map[*checker.Type]bool{}
	var opaque func(*checker.Type) bool
	opaque = func(proven *checker.Type) bool {
		if seen[proven] {
			return false
		}
		seen[proven] = true
		if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				if opaque(member) {
					return true
				}
			}
			return false
		}
		if held, known := l.representation(l.checker.GetNonNullableType(proven)); known && held == ir.Closure {
			return true
		}
		if checker.IsTupleType(proven) || l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "RegExp", "Date", "Stats", "Hash", "Buffer") {
			return true
		}
		if l.checker.IsArrayType(proven) {
			for _, argument := range l.checker.GetTypeArguments(proven) {
				if opaque(argument) {
					return true
				}
			}
			return false
		}
		if proven.Flags()&checker.TypeFlagsObject == 0 {
			return false
		}
		for _, field := range l.checker.GetPropertiesOfType(proven) {
			if l.includesNull(l.checker.GetTypeOfSymbol(field)) || opaque(l.checker.GetTypeOfSymbol(field)) {
				return true
			}
		}
		return false
	}
	incomplete := false
	for _, cause := range errors.errorCauses {
		if opaque(cause) {
			incomplete = true
			break
		}
	}
	if !incomplete {
		return nil
	}
	reflects := false
	inspect := func(node any) bool {
		switch node.(type) {
		case ir.DynamicProperty, ir.HasProperty:
			reflects = true
		}
		return true
	}
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	if reflects {
		return &NotYet{Where: l.result.Source, What: "dynamic property reflection with an opaque Error cause (function, collection or nullable field descriptors are incomplete; narrow the concrete value before constructing the error, or observe only typeof, identity and nominal instanceof)"}
	}
	return nil
}
