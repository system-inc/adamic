package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

func (e *emitter) forInDeclarations() {
	e.declarations = append(e.declarations,
		"extern const adamic_methods adamic_for_in_metadata;",
		"adamic_array *adamic_for_in_keys(const adamic_heap *);",
		"bool adamic_for_in_own(const adamic_heap *, const adamic_string *);",
		"adamic_object *adamic_for_in_copy(const adamic_object *);",
		"void adamic_for_in_initialize(adamic_object *, size_t);",
		"void adamic_for_in_write(adamic_object *, const char *);")
}

func (e *emitter) enumeratesKeys() bool {
	found := false
	walkExpressions(e.program, func(expression ir.Expression) {
		if keys, ok := expression.(ir.ObjectKeys); ok && keys.Enumeration {
			found = true
		}
	})
	return found
}

func (e *emitter) forInEmptyShape(literal ir.ObjectLiteral) string {
	fields := emptyFields(literal)
	fields = append(fields, ir.Field{Name: "\x01", Value: ir.ArrayLiteral{Element: ir.String}})
	base := e.shape(fields)
	name := base + "_enumeration"
	e.forInDeclarations()
	// Repeated identical declarations are avoided independently of ordinary literal shapes.
	key := "for-in:" + base
	if declared, ok := e.shapes[key]; ok {
		return declared
	}
	e.shapes[key] = name
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_shape %s = {%d, %s_names, %s_references, &adamic_for_in_metadata};", name, len(fields), base, base))
	return name
}

// A structural object slot may hold an array only when every use forwards it to
// runtime enumeration. Other object operations still require object storage.
func (e *emitter) forInOnly(local int, visiting map[int]bool) bool {
	if visiting[local] {
		return false
	}
	visiting[local] = true
	defer delete(visiting, local)
	direct := func(value ir.Expression) bool { read, ok := value.(ir.Read); return ok && read.Local == local }
	safe := true
	var scan func(reflect.Value)
	scan = func(value reflect.Value) {
		if !safe || !value.IsValid() {
			return
		}
		if value.Kind() == reflect.Interface {
			if value.IsNil() {
				return
			}
			switch node := value.Interface().(type) {
			case ir.ObjectKeys:
				if node.Enumeration && direct(node.Object) {
					return
				}
			case ir.ForInOwn:
				if direct(node.Object) {
					scan(reflect.ValueOf(node.Key))
					return
				}
			case ir.Declare:
				if direct(node.Value) {
					safe = e.forInOnly(node.Local, visiting)
					return
				}
			case ir.Assign:
				if direct(node.Value) {
					safe = e.forInOnly(node.Local, visiting)
					return
				}
			case ir.Call:
				for i, argument := range node.Arguments {
					if direct(argument) {
						for _, target := range e.program.CallTargets(node) {
							if i >= len(e.program.Functions[target].Parameters) || !e.forInOnly(e.program.Functions[target].Parameters[i], visiting) {
								safe = false
								return
							}
						}
					} else {
						scan(reflect.ValueOf(argument))
					}
				}
				return
			case ir.Read:
				if node.Local == local {
					safe = false
					return
				}
			}
			scan(value.Elem())
			return
		}
		switch value.Kind() {
		case reflect.Struct:
			for i := 0; i < value.NumField(); i++ {
				scan(value.Field(i))
			}
		case reflect.Slice:
			for i := 0; i < value.Len(); i++ {
				scan(value.Index(i))
			}
		}
	}
	scan(reflect.ValueOf(e.program.Main))
	for _, function := range e.program.Functions {
		scan(reflect.ValueOf(function.Body))
	}
	return safe
}

func (e *emitter) enumerationSlot(local int, value string) string {
	if e.program.Locals[local].Type == ir.Object && e.enumeratesKeys() && e.forInOnly(local, map[int]bool{}) {
		return "((adamic_object *)" + value + ")"
	}
	return value
}

func (e *emitter) enumerationArgument(local int, from ir.Type, value string) string {
	if e.program.Locals[local].Type != ir.Object || !e.enumeratesKeys() || !e.forInOnly(local, map[int]bool{}) {
		return value
	}
	if !from.IsReference() {
		boxed, fresh := converted(from, ir.Union, value)
		boxed = "((adamic_object *)" + boxed + ")"
		if fresh {
			return e.own(ir.Object, boxed)
		}
		return boxed
	}
	return e.enumerationSlot(local, value)
}
