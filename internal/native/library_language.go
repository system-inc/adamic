package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) libraryLanguageValue(value ir.Expression) string {
	switch value := value.(type) {
	case ir.ObjectKeys:
		return nativeEmitter.own(
			ir.Array,
			fmt.Sprintf("adamic_class_object_keys(%s)", nativeEmitter.value(value.Object)),
		)
	case ir.ClosureSelf:
		return nativeEmitter.own(ir.Closure, "adamic_retain(self)")
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
