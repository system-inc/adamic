package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// nullReference is the shared emission path for a reference kind's null value.
// Kinds not migrated to sentinels retain their existing NULL representation.
func nullReference(of ir.Type) string {
	if !of.UsesNullSentinel() {
		return "NULL"
	}
	return fmt.Sprintf("((%s)adamic_reference_null(%s))", cType(of), referenceKind(of))
}

func referenceKind(of ir.Type) string {
	switch of {
	case ir.String:
		return "adamic_kind_string"
	case ir.Object:
		return "adamic_kind_object"
	case ir.Array:
		return "adamic_kind_array"
	case ir.Map:
		return "adamic_kind_map"
	case ir.Closure:
		return "adamic_kind_closure"
	}
	panic(fmt.Sprintf("native: no nullable reference kind for %d", of))
}

func nullTest(of ir.Type, value string, undefined bool) string {
	if of.UsesNullSentinel() {
		return fmt.Sprintf("adamic_reference_is_null(%s, %s, %t)", value, referenceKind(of), undefined)
	}
	return fmt.Sprintf("(%s == NULL)", value)
}
