package javascript

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// unknownKindRuntime is adamicUnknownKind(value), ir.UnknownKind: the number internal/ir gives the kind
// of value String's dispatch takes apart (lower/to_primitive.go). It answers as native's
// adamic_union_kind does, from Adamic's own class metadata where a class is concerned, so the two
// backends take the same arm: a tuple is an array here, a record a plain object, and a class's static
// side a function, which its class identity tells from a closure.
func unknownKindRuntime(program *ir.Program) string {
	var builder strings.Builder
	kind := func(test string, number int) {
		fmt.Fprintf(&builder, "\tif (%s) return %d;\n", test, number)
	}
	builder.WriteString("const adamicUnknownKind = (value) => {\n")
	kind("value === undefined", ir.UnknownUndefined)
	kind("value === null", ir.UnknownNull)
	kind("typeof value === 'number'", ir.UnknownNumber)
	kind("typeof value === 'boolean'", ir.UnknownBoolean)
	kind("typeof value === 'string'", ir.UnknownString)
	kind("value instanceof AdamicClosure", ir.UnknownFunction)
	kind("Array.isArray(value)", ir.UnknownArray)
	kind("value instanceof Map || value instanceof Set", ir.UnknownCollection)
	kind("ArrayBuffer.isView(value)", ir.UnknownTypedArray)
	kind("value instanceof Error", ir.UnknownError)
	if len(program.Classes) > 0 {
		builder.WriteString("\tconst id = adamicClassIdentities.get(value);\n")
		fmt.Fprintf(&builder, "\tif (id !== undefined) return adamicClasses[id - 1].static ? %d : %d;\n", ir.UnknownStaticClass, ir.UnknownInstance)
	}
	kind("typeof value !== 'object' || value instanceof Promise", ir.UnknownOther)
	fmt.Fprintf(&builder, "\treturn %d;\n", ir.UnknownObject)
	builder.WriteString("};\n")
	return builder.String()
}

// joinRuntime is ir.ArrayOfScalars, ir.JoinEnter and ir.JoinLeave, as native's adamic_array_of_scalars,
// adamic_join_enter and adamic_join_leave answer them: the scalar test by each element's own kind, so
// both backends stop on the same arrays, and a join stack of the program's own, since its join is the
// helper's loop and not Array.prototype.join, whose stack is V8's.
const joinRuntime = "const adamicArrayOfScalars = (array) => array.length > 0 && array.every((element) => element === undefined || element === null || typeof element === 'number' || typeof element === 'boolean');\n" +
	"const adamicJoining = [];\n" +
	"const adamicJoinEnter = (array) => { if (adamicJoining.includes(array)) return false; adamicJoining.push(array); return true; };\n" +
	"const adamicJoinLeave = () => { adamicJoining.pop(); };\n"
