package javascript

import (
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// dynamicKeyedRuntime is adamicDynamicGet(value, key), Reflect.get(Object(value), key, value) for a
// value lowering proved neither null nor undefined (ir.DynamicKeyedProperty). It answers and stops
// exactly where native's adamic_dynamic_property_keyed does, in the same order, with the same words
// from internal/ir's one table, so the two backends' stderr agree byte for byte.
//
// It classifies from Adamic's own class metadata, never from JavaScript reflection, since this
// backend's objects carry what native's don't: a class's private fields are string keys named "#..."
// (adamicClass), a method is on its class's prototype, an object literal's getter is an own accessor,
// and a Node error has own stack, errno, syscall and path. So a class's accessors are found among its
// table's own entries and called through adamicGetAccessor, its methods are judged by the names its
// table holds, an own value is read only by Object.hasOwn, and a record, a plain object here, is read
// as recordRuntime reads one, by its own keys.
func dynamicKeyedRuntime(program *ir.Program) string {
	var builder strings.Builder
	stops := []string{}
	for _, stop := range ir.DynamicKeyedStops {
		stops = append(stops, stop.Name+": "+quote(stop.Message))
	}
	builder.WriteString("const adamicDynamicKeyedStops = {" + strings.Join(stops, ", ") + "};\n")
	for _, list := range []struct {
		name  string
		names []string
	}{{"adamicDynamicKeyedPrototypeNames", ir.DynamicKeyedPrototypeNames}, {"adamicDynamicKeyedErrorNames", ir.DynamicKeyedErrorNames}} {
		quoted := []string{}
		for _, name := range list.names {
			quoted = append(quoted, quote(name))
		}
		builder.WriteString("const " + list.name + " = new Set([" + strings.Join(quoted, ", ") + "]);\n")
	}
	builder.WriteString("const adamicDynamicStop = (stop, key) => panic(`${adamicDynamicKeyedStops[stop]} (reading '${key}')`);\n")
	classes := len(program.Classes) > 0
	if classes {
		tables := []string{}
		for class := range program.Classes {
			quoted := []string{}
			for _, name := range program.DynamicKeyedMethodNames(class + 1) {
				quoted = append(quoted, quote(name))
			}
			tables = append(tables, "["+strings.Join(quoted, ", ")+"]")
		}
		builder.WriteString("const adamicDynamicMethodNames = [" + strings.Join(tables, ", ") + "];\n")
		// A class's accessors are an object literal, so indexing it by an Object.prototype name finds
		// Object's own member (constructor, toString), which adamicFindAccessor takes for an accessor
		// without a getter. Only the table's own entries are accessors here, as native's class->accessors
		// holds only the class's own.
		builder.WriteString("const adamicDynamicAccessor = (value, key) => { for (let id = adamicClassIdentities.get(value); id; id = adamicClasses[id - 1].base) { const accessors = adamicClasses[id - 1].accessors; if (Object.hasOwn(accessors, key)) return accessors[key]; } };\n")
	}
	builder.WriteString("const adamicDynamicGet = (value, key) => {\n")
	builder.WriteString("\tif (value === undefined || value === null) return adamicDynamicStop('Nullish', key);\n")
	builder.WriteString("\tif (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') return adamicDynamicStop('Primitive', key);\n")
	builder.WriteString("\tif (value instanceof AdamicClosure) return adamicDynamicStop('Function', key);\n")
	// A tuple is an array on this backend, and every array literal is marked as a tuple is
	// (adamicTuple), so the two share native's one stop.
	builder.WriteString("\tif (Array.isArray(value)) return adamicDynamicStop('Array', key);\n")
	builder.WriteString("\tif (value instanceof Map || value instanceof Set) return adamicDynamicStop('Collection', key);\n")
	builder.WriteString("\tif (ArrayBuffer.isView(value)) return adamicDynamicStop('TypedArray', key);\n")
	if classes {
		builder.WriteString("\tconst id = adamicClassIdentities.get(value);\n")
		// A class's static side is a function on this backend, a class object natively.
		builder.WriteString("\tif (id !== undefined && adamicClasses[id - 1].static) return adamicDynamicStop('StaticClass', key);\n")
	}
	builder.WriteString("\tif (typeof value !== 'object' || value instanceof Promise) return adamicDynamicStop('RuntimeValue', key);\n")
	builder.WriteString("\tif (key.startsWith('#')) return adamicDynamicStop('Private', key);\n")
	builder.WriteString("\tconst error = value instanceof Error;\n")
	builder.WriteString("\tif (error && adamicDynamicKeyedErrorNames.has(key)) return adamicDynamicStop('ErrorProperty', key);\n")
	if classes {
		builder.WriteString("\tif (id !== undefined) {\n")
		builder.WriteString("\t\tconst accessor = adamicDynamicAccessor(value, key);\n")
		builder.WriteString("\t\tif (accessor !== undefined) {\n")
		builder.WriteString("\t\t\tif (adamicClasses[id - 1].literal) return adamicDynamicStop('ObjectLiteralGetter', key);\n")
		builder.WriteString("\t\t\tif (accessor.get === undefined) return adamicDynamicStop('SetterOnly', key);\n")
		builder.WriteString("\t\t\treturn adamicGetAccessor(value, key);\n")
		builder.WriteString("\t\t}\n")
		builder.WriteString("\t\tif (adamicDynamicMethodNames[id - 1].includes(key)) return adamicDynamicStop('Method', key);\n")
		builder.WriteString("\t}\n")
	}
	// A native Error holds its name and message as its own; Node's inherits its name, and the empty
	// message of one made without any.
	builder.WriteString("\tif (error && (key === 'name' || key === 'message')) return value[key];\n")
	builder.WriteString("\tif (Object.hasOwn(value, key)) return value[key];\n")
	builder.WriteString("\tif (adamicDynamicKeyedPrototypeNames.has(key)) return adamicDynamicStop('ObjectPrototype', key);\n")
	builder.WriteString("\treturn undefined;\n")
	builder.WriteString("};\n")
	return builder.String()
}
