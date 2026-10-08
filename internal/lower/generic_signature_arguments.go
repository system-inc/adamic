package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Fill binders hidden inside conditional or indexed types from the checker's
// resolved call, including default arguments. This is the same checked shim
// bridge used by classGenericCall, and does not invent an erased representation.
func (l *lowering) signatureTypeArguments(resolved *checker.Signature, declaration *ast.Node, into map[*checker.Type]*checker.Type) {
	field := reflect.ValueOf(resolved).Elem().FieldByName("mapper")
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
		return
	}
	mapper := (*typeMapper)(field.UnsafePointer())
	for _, parameter := range declaration.TypeParameters() {
		source := l.checker.GetTypeAtLocation(parameter.Name())
		if _, known := into[source]; known {
			continue
		}
		target := l.concrete(instantiateType(l.checker, source, mapper))
		if target != source && target.Flags()&checker.TypeFlagsTypeParameter == 0 {
			into[source] = target
		}
	}
}
