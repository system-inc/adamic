package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) libraryLanguageValue(value ir.Expression) string {
	switch value := value.(type) {
	case ir.ForInOwn:
		e.forInDeclarations()
		return fmt.Sprintf("adamic_for_in_own((const adamic_heap *)%s, %s)", e.value(value.Object), e.value(value.Key))
	case ir.ObjectKeys:
		if value.Enumeration {
			e.forInDeclarations()
			return e.own(ir.Array, fmt.Sprintf("adamic_for_in_keys((const adamic_heap *)%s)", e.value(value.Object)))
		}
		return e.own(ir.Array, fmt.Sprintf("adamic_class_object_keys(%s)", e.value(value.Object)))
	case ir.ClosureSelf:
		return e.own(ir.Closure, "adamic_retain(self)")
	case ir.LibraryGlobal:
		return fmt.Sprintf("(%s)adamic_library_identity(%d)", cType(value.Type()), libraryIdentityIndex(value.Name))
	}
	panic(fmt.Sprintf("native: no library language value for %T", value))
}

func libraryIdentityIndex(name string) int {
	for index, candidate := range []string{"Object", "Array", "String", "Number", "JSON", "Map", "Set"} {
		if candidate == name {
			return index
		}
	}
	panic("native: unknown library identity " + name)
}
