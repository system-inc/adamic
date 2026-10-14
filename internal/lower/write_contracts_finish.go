package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

// A monomorphized allocation can reveal a narrower contract after a broad alias's
// writer was lowered. Protect those allocation fields across every named alias,
// including earlier function bodies. Initializers without a source origin stay trusted.
func finishWriteContracts(program *ir.Program) {
	if program.CheckedElements {
		note := func(node any) bool {
			var origin ir.WriteCheck
			switch write := node.(type) {
			case ir.ArrayPush:
				origin = write.WriteOrigin
			case ir.ArraySplice:
				if len(write.Items) > 0 {
					origin = write.WriteOrigin
				}
			case ir.SetIndex:
				origin = write.WriteOrigin
			case ir.MapSet:
				origin = write.WriteOrigin
			case ir.ArrayFill:
				if write.Array != nil {
					origin = write.WriteOrigin
				}
			}
			if origin.Expression != "" {
				program.WriteChecks = append(program.WriteChecks, origin)
			}
			return true
		}
		walk(program.Main, note)
		for _, function := range program.Functions {
			walk(function.Body, note)
		}
	}
	if len(program.CheckedWrites) == 0 && !program.CheckedElements {
		return
	}
	if program.CheckedWrites == nil {
		program.CheckedWrites = map[string]bool{}
	}
	note := func(node any) bool {
		if literal, ok := node.(ir.ObjectLiteral); ok {
			for _, field := range literal.Fields {
				if c := field.Contract; c != nil && (len(c.Allowed) > 0 || c.Reference) {
					program.CheckedWrites[field.Name] = true
				}
			}
		}
		return true
	}
	walk(program.Main, note)
	for _, function := range program.Functions {
		walk(function.Body, note)
	}
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		if value.Kind() == reflect.Interface {
			if value.IsNil() {
				return value
			}
			mapped := transform(value.Elem())
			if write, ok := mapped.Interface().(ir.SetProperty); ok && write.WriteCheck == "" && write.WriteOrigin.Expression != "" && program.CheckedWrites[write.Name] {
				write.WriteCheck = write.WriteOrigin.Expression
				program.WriteChecks = append(program.WriteChecks, write.WriteOrigin)
				mapped = reflect.ValueOf(write)
			}
			result := reflect.New(value.Type()).Elem()
			result.Set(mapped)
			return result
		}
		switch value.Kind() {
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(transform(value.Field(i)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(transform(value.Index(i)))
			}
			return result
		}
		return value
	}
	program.Main = transform(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for index := range program.Functions {
		program.Functions[index].Body = transform(reflect.ValueOf(program.Functions[index].Body)).Interface().([]ir.Statement)
	}
}
