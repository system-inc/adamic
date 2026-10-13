package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

// CFG instructions do not expose every nested loop header as Part zero.
// Finalize the non-executable array metadata once over the complete IR tree.
func finalizeArrayReadMetadata(program *ir.Program) {
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(v reflect.Value) reflect.Value {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return v
			}
			r := reflect.New(v.Type()).Elem()
			r.Set(rewrite(v.Elem()))
			return r
		case reflect.Struct:
			r := reflect.New(v.Type()).Elem()
			for i := 0; i < v.NumField(); i++ {
				r.Field(i).Set(rewrite(v.Field(i)))
			}
			switch n := r.Interface().(type) {
			case ir.ArrayIndex:
				r.Set(reflect.ValueOf(markProgramViewArrayRead(program, n)))
			case ir.ArrayViewRead:
				r.Set(reflect.ValueOf(markProgramViewArrayUse(program, n)))
			}
			return r
		case reflect.Slice:
			if v.IsNil() {
				return v
			}
			r := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				r.Index(i).Set(rewrite(v.Index(i)))
			}
			return r
		case reflect.Array:
			r := reflect.New(v.Type()).Elem()
			for i := 0; i < v.Len(); i++ {
				r.Index(i).Set(rewrite(v.Index(i)))
			}
			return r
		default:
			return v
		}
	}
	program.Main = rewrite(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = rewrite(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
	}
}
