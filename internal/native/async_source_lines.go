package native

import (
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// Instructions retain the addresses of their original statement slots. Keep
// source metadata beside the graph: its last instruction must remain the one
// that throws or awaits. A child's markers never escape its own slice.
//
// Named slice types over ir.Statement and statements in maps have no recorded
// slots. Missing #line metadata is incomplete, never wrong; keep statement-slot
// addresses as the only keys rather than substituting another kind of key.
func asyncSourceLines(body []ir.Statement) map[*ir.Statement]ir.SourceLine {
	return asyncSourceLineValues(reflect.ValueOf(body))
}

func asyncSourceLineValues(body reflect.Value) map[*ir.Statement]ir.SourceLine {
	locations := map[*ir.Statement]ir.SourceLine{}
	var walk func(reflect.Value, ir.SourceLine)
	walk = func(value reflect.Value, source ir.SourceLine) {
		if !value.IsValid() || !value.CanInterface() {
			return
		}
		if value.Type() == reflect.TypeFor[[]ir.Statement]() {
			statements := value.Interface().([]ir.Statement)
			for index := range statements {
				if marker, ok := statements[index].(ir.SourceLine); ok {
					source = marker
					continue
				}
				if source.File != "" {
					locations[&statements[index]] = source
				}
				walk(reflect.ValueOf(statements[index]), source)
			}
			return
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				walk(value.Elem(), source)
			}
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				if !value.Type().Field(index).IsExported() {
					continue
				}
				walk(value.Field(index), source)
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				walk(value.Index(index), source)
			}
		}
	}
	walk(body, ir.SourceLine{})
	return locations
}
