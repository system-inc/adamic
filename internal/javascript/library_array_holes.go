package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

func hasArrayHoles(program *ir.Program) bool {
	found := false
	var walk func(reflect.Value)
	walk = func(value reflect.Value) {
		if found || !value.CanInterface() {
			return
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				walk(value.Elem())
			}
		case reflect.Struct:
			if _, ok := value.Interface().(ir.ArraySetLength); ok {
				found = true
				return
			}
			if _, ok := value.Interface().(ir.ArrayHoles); ok {
				found = true
				return
			}
			for index := 0; index < value.NumField(); index++ {
				walk(value.Field(index))
			}
		case reflect.Array, reflect.Slice:
			for index := 0; index < value.Len(); index++ {
				walk(value.Index(index))
			}
		}
	}
	walk(reflect.ValueOf(program))
	return found
}
