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
	}{{"adamicDynamicKeyedPrototypeNames", ir.DynamicKeyedPrototypeNames}, {"adamicDynamicKeyedArrayPrototypeNames", ir.DynamicKeyedArrayPrototypeNames.Names()}, {"adamicDynamicKeyedErrorNames", ir.DynamicKeyedErrorNames}, {"adamicDynamicKeyedStatsFieldNames", ir.DynamicKeyedStatsFieldNames}} {
		quoted := []string{}
		for _, name := range list.names {
			quoted = append(quoted, quote(name))
		}
		builder.WriteString("const " + list.name + " = new Set([" + strings.Join(quoted, ", ") + "]);\n")
	}
	builder.WriteString("const adamicDynamicStop = (stop, key) => panic(`${adamicDynamicKeyedStops[stop]} (reading '${key}')`);\n")
	// %RegExpStringIteratorPrototype%, held once: a matchAll iterator on this backend is Node's own.
	builder.WriteString("const adamicRegExpStringIteratorPrototype = Object.getPrototypeOf(''.matchAll(/(?:)/g));\n")
	// Stats.prototype, held once, for an fs.Stats is Node's own on this backend too. Reading fs.Stats'
	// prototype stats nothing and calls nothing, so the deprecated constructor never warns.
	builder.WriteString("const adamicStatsPrototype = adamicNodeFSFile.Stats.prototype;\n")
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
	// An array answers its own properties as Node's does (#ys4naww), its indexes, its length and a
	// match's fields, as native's keyed_array does, and a name it inherits Node's own function
	// (#g187f1e), where native hands out its immortal closure for the name; __proto__ stops, as on an
	// object, and any other key reads undefined. A tuple is an array on this backend, so it answers by
	// the same rule, as native's keyed_tuple does.
	builder.WriteString("\tif (Array.isArray(value)) {\n")
	builder.WriteString("\t\tif (Object.hasOwn(value, key)) {\n")
	// A match's indices pairs are tuples natively, so they're marked as the static read marks them
	// (ir.RegExpProperty).
	builder.WriteString("\t\t\tif (key === 'indices' && Object.hasOwn(value, 'input')) value.indices?.forEach((pair) => pair !== undefined && adamicTuple(pair));\n")
	builder.WriteString("\t\t\treturn value[key];\n")
	builder.WriteString("\t\t}\n")
	builder.WriteString("\t\tif (key === '__proto__') return adamicDynamicStop('Proto', key);\n")
	builder.WriteString("\t\tif (adamicDynamicKeyedArrayPrototypeNames.has(key) || adamicDynamicKeyedPrototypeNames.has(key)) return value[key];\n")
	builder.WriteString("\t\treturn undefined;\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\tif (value instanceof Map || value instanceof Set) return adamicDynamicStop('Collection', key);\n")
	builder.WriteString("\tif (ArrayBuffer.isView(value)) return adamicDynamicStop('TypedArray', key);\n")
	if classes {
		builder.WriteString("\tconst id = adamicClassIdentities.get(value);\n")
		// A class's static side is a function on this backend, a class object natively.
		builder.WriteString("\tif (id !== undefined && adamicClasses[id - 1].static) return adamicDynamicStop('StaticClass', key);\n")
	}
	// A program's function is an AdamicClosure, which stopped above, so a function here is Node's own,
	// one this helper handed out (an array's map, a matchAll iterator's next, Object.prototype's
	// toString), where native hands out an immortal closure: it stops in the Function form, as the
	// closure does natively, never as a runtime value. A program class's static side is a function on
	// this backend too, so its own stop comes first.
	builder.WriteString("\tif (typeof value === 'function') return adamicDynamicStop('Function', key);\n")
	builder.WriteString("\tif (typeof value !== 'object' || value instanceof Promise) return adamicDynamicStop('RuntimeValue', key);\n")
	builder.WriteString("\tif (key.startsWith('#')) return adamicDynamicStop('Private', key);\n")
	builder.WriteString("\tif (key === '__proto__') return adamicDynamicStop('Proto', key);\n")
	builder.WriteString("\tconst error = value instanceof Error;\n")
	builder.WriteString("\tif (error && adamicDynamicKeyedErrorNames.has(key)) return adamicDynamicStop('ErrorProperty', key);\n")
	// A matchAll iterator answers as Node's does (#zhrjf2t), Node's own member or undefined, before the
	// host test, as native's keyed_regexp_string_iterator does: native knows it by its shape
	// (adamic_regex_iterator_is), this backend by its prototype.
	builder.WriteString("\tif (Object.getPrototypeOf(value) === adamicRegExpStringIteratorPrototype) return value[key];\n")
	// So does a Stats (#zhrjf2t), as native's keyed_stats does, but for a field native doesn't hold,
	// which stops here as it does there, though Node's own Stats holds it.
	builder.WriteString("\tif (Object.getPrototypeOf(value) === adamicStatsPrototype) return adamicDynamicKeyedStatsFieldNames.has(key) ? adamicDynamicStop('RuntimeValue', key) : value[key];\n")
	// host is the one predicate for a value whose prototype isn't Object's, a program class's instance
	// aside (#ey7aysy, Adamic's ruling on #0sps03r), native's host_object (union.c) on this backend's
	// objects: a Stats, a Dirent, an AdamicIterator, a match's groups object, process.memoryUsage()'s
	// (library_node_process.go) and an Error. It stops for every key before an own property is read,
	// an Error aside, whose own fields answer as Node's do, and whose prototype's names stop below.
	instance := ""
	if classes {
		instance = " && id === undefined"
	}
	builder.WriteString("\tconst host = Object.getPrototypeOf(value) !== Object.prototype" + instance + ";\n")
	builder.WriteString("\tif (host && !error) return adamicDynamicStop('RuntimeValue', key);\n")
	if classes {
		builder.WriteString("\tif (id !== undefined) {\n")
		builder.WriteString("\t\tconst accessor = adamicDynamicAccessor(value, key);\n")
		builder.WriteString("\t\tif (accessor !== undefined) {\n")
		builder.WriteString("\t\t\tif (adamicClasses[id - 1].literal) return adamicDynamicStop('ObjectLiteralGetter', key);\n")
		builder.WriteString("\t\t\tif (accessor.get === undefined) return adamicDynamicStop('SetterOnly', key);\n")
		builder.WriteString("\t\t\treturn adamicGetAccessor(value, key);\n")
		builder.WriteString("\t\t}\n")
		builder.WriteString("\t\tif (adamicDynamicMethodNames[id - 1].includes(key)) return adamicDynamicStop('Method', key);\n")
		// An object literal is a plain object here, whose constructor is Object, as native's is.
		builder.WriteString("\t\tif (!adamicClasses[id - 1].literal && key === 'constructor') return adamicDynamicStop('ClassConstructor', key);\n")
		builder.WriteString("\t}\n")
	}
	// A native Error holds its name and message as its own; Node's inherits its name, and the empty
	// message of one made without any.
	builder.WriteString("\tif (error && (key === 'name' || key === 'message')) return value[key];\n")
	builder.WriteString("\tif (Object.hasOwn(value, key)) return value[key];\n")
	// Node's own Object.prototype member, where native hands out its immortal closure for the name.
	// Only an Error reaches here as a host, and Error.prototype's members aren't Object's, so it stops.
	builder.WriteString("\tif (adamicDynamicKeyedPrototypeNames.has(key)) {\n")
	builder.WriteString("\t\tif (host) return adamicDynamicStop('RuntimeValue', key);\n")
	builder.WriteString("\t\treturn value[key];\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\treturn undefined;\n")
	builder.WriteString("};\n")
	return builder.String()
}

// dynamicHostRuntime guards an in test and the read it narrows on a value viewed as unknown
// (ir.HasProperty, ir.DynamicProperty), Object.keys (ir.ObjectKeys, ir.ObjectCall), hasOwnProperty
// (ir.HasOwn) and ToPrimitive's own-member test (ir.OwnFunction), as native guards
// adamic_has_property, adamic_dynamic_property, adamic_class_object_keys, adamic_object_keys,
// adamic_object_has and adamic_object_own_function (union.c's adamic_host_stop, Adamic's ruling on
// #ey7aysy at 23:03Z): a host value stops in the runtime value form before its properties are
// consulted. A host value here is one whose prototype isn't Object's and that native holds as a
// plain object's kind, so an array, a collection, a typed array, a closure, a promise, an Error and a
// program class's instance aren't hosts, as native's host_object passes each of them by; an Error
// answers its own names as Node's does. adamicDynamicGet's host test names the same values for a
// run-time key read.
func dynamicHostRuntime(program *ir.Program) string {
	message := ""
	for _, stop := range ir.DynamicKeyedStops {
		if stop.Name == "RuntimeValue" {
			message = stop.Message
		}
	}
	instance := ""
	if len(program.Classes) > 0 {
		instance = " && adamicClassIdentities.get(value) === undefined"
	}
	var builder strings.Builder
	builder.WriteString("const adamicHostValue = value => typeof value === 'object' && value !== null && !Array.isArray(value) && !(value instanceof Map || value instanceof Set) && !ArrayBuffer.isView(value) && !(value instanceof AdamicClosure) && !(value instanceof Promise) && !(value instanceof Error) && Object.getPrototypeOf(value) !== Object.prototype" + instance + ";\n")
	builder.WriteString("const adamicHostStop = key => panic(`" + message + " (reading '${key}')`);\n")
	builder.WriteString("const adamicDynamicHas = (value, key) => adamicHostValue(value) ? adamicHostStop(key) : key in value;\n")
	builder.WriteString("const adamicDynamicRead = (value, key) => adamicHostValue(value) ? adamicHostStop(key) : value[key];\n")
	builder.WriteString("const adamicHostKeys = value => adamicHostValue(value) ? panic(" + quote(ir.DynamicKeysStop) + ") : Object.keys(value);\n")
	builder.WriteString("const adamicHostHasOwn = (value, key) => adamicHostValue(value) ? adamicHostStop(key) : value.hasOwnProperty(key);\n")
	builder.WriteString("const adamicHostCheck = (value, key) => adamicHostValue(value) ? adamicHostStop(key) : true;\n")
	return builder.String()
}
