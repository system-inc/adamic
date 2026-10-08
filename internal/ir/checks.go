package ir

import (
	"reflect"
	"strings"
)

// InsertedCheck names a guard in the lowered program. An option audit row is
// not a guard: only a panic-bearing expression actually present in IR counts.
type InsertedCheck struct {
	Kind  string
	Where string
}

// InsertedChecks reports each emitted guard occurrence, including specialized
// function bodies. These guards are checked at runtime, never counted as trusted.
// The message prefixes are the use-site contracts shared by both backends.
func InsertedChecks(program *Program) []InsertedCheck {
	checks := []InsertedCheck{}
	prefixes := []struct{ prefix, kind string }{
		{"indexed read is absent: ", "indexed-presence"},
		{"catch value is not Error: ", "catch-error"},
		{"JSON.stringify result is undefined: ", "json-stringify-defined"},
		{"optional property write is undefined: ", "optional-write"},
	}
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.Kind() == reflect.Struct && value.CanInterface() {
			if coalesce, ok := value.Interface().(Coalesce); ok && coalesce.Panic != nil {
				if message, ok := coalesce.Panic.(StringConstant); ok {
					text := program.Strings[message.Index]
					for _, contract := range prefixes {
						if where, found := strings.CutPrefix(text, contract.prefix); found {
							checks = append(checks, InsertedCheck{Kind: contract.kind, Where: where})
							break
						}
					}
				}
			}
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				visit(value.Elem())
			}
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				visit(value.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				visit(value.Index(index))
			}
		}
	}
	visit(reflect.ValueOf(program.Main))
	for _, function := range program.Functions {
		visit(reflect.ValueOf(function.Body))
	}
	return checks
}
