package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Reading a resolved return union back cannot recover T when undefined has
// collapsed a parameter's union, or when T itself contains undefined. Use the
// checker's actual substitution, just as classGenericCall does for class layouts.
// If the shim does not expose it, ordinary signature inference remains available.
func (l *lowering) genericReturnTypes(resolved *checker.Signature, declaration *ast.Node) map[*checker.Type]*checker.Type {
	types := map[*checker.Type]*checker.Type{}
	field := reflect.ValueOf(resolved).Elem().FieldByName("mapper")
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
		return types
	}
	mapper := (*typeMapper)(field.UnsafePointer())
	for _, parameter := range declaration.TypeParameters() {
		source := l.checker.GetTypeAtLocation(parameter.Name())
		target := l.concrete(instantiateType(l.checker, source, mapper))
		if target != source {
			types[source] = target
		}
	}
	return types
}
