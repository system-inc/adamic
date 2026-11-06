package ir

import (
	"reflect"
	"strings"
)

// Inert reports whether an expression's own operation runs no program code and reaches nothing
// outside the program's own values: no call, no callback, no generator's body, no getter, no other
// task run while it waits, and no host (a file, the process, the console). What its operands do is
// theirs, and a proof walks them on its own. The list is fixed, and anything not on it, an operation
// added later included, is not inert, so a proof that asks this before passing a node it has no case
// for fails closed on a node it has never seen (Compiler's and Adamic's ruling on #qk8rztp S2c-2b:
// worker/purity.go's and native/reuse.go's defaults).
//
// Off the list, and each such proof's own case to judge: a call (Call, CallClosure, and the runtime
// loops that call a callback, ArrayMap, ArrayVisit, ArrayReduce, ArrayFrom, ArraySort, MapForEach,
// ParallelMap), a generator's step (GeneratorNext), a pattern's replacement callback (RegExpCall), a
// run-time key read, which calls whichever getter its key names (DynamicKeyedProperty), an encoding,
// which internal/flow takes to call back (JSONEncode), a suspension (Await, Yield), and every host
// operation (FileStatus, NodeFSFile, NodeHostCall, ProcessCall, ProgramArguments, ReadDirectory,
// ReadTextFile, WriteTextFile).
//
// The proofs that ask it, worker/purity.go's for a Wasm build and native/reuse.go's, judge the
// native program, so inert means natively: what the runtime runs for the node, given what lowering
// lets reach it. Seven entries are JavaScript operations that can run program code there (Compiler's
// question on #qk8rztp S2c-2b), and each is inert only because a stop keeps that case out, which
// inertExpressions cites beside it.
func Inert(expression Expression) bool {
	if expression == nil {
		return true
	}
	return inertExpressions[reflect.TypeOf(expression).Name()]
}

var inertExpressions = func() map[string]bool {
	result := map[string]bool{}
	for _, name := range strings.Fields(`ArrayConcat ArrayFill ArrayIndex ArrayIsArray ArrayLiteral ArrayOfScalars ArrayPop
		ArrayPush ArrayReverse ArraySearch ArraySlice ArraySplice Binary BooleanConstant BooleanToString Box CharCodeAt
		CheckedCast ClosureSelf Coalesce CodePoints CollectionIterator Comma Conditional Defined Effects ErrorToString
		GeneratorDone HasAccessor HasOwn HasProperty InstanceOf IsCollectionIterator IsGenerator IsNull IsUndefined
		IterableOf JSONDecode JSONNull JoinEnter JoinLeave Length LibraryGlobal Logical LogicalAssignment MakeClosure
		MakeError MapClear MapDelete MapEntries MapGet MapHas MapKeys MapNew MapSet MapSize MapValues MathCall MaybeOf
		Narrow NodeBufferCall Null NumberCall NumberConstant NumberFormat NumberToString ObjectKeys ObjectLiteral
		OwnFunction PhantomMember PlaceholderUse PromiseValue Property Read RecordCall RecordCoalesce RecordLiteral
		RecordStorage RegExpGroup RegExpNew RegExpProperty SetAdd SetNew SetValues StringCall StringConstant
		StringFromCodes StringIndex StringIterator StringLength ToFixed Trim Truthy TypeOf TypedArrayFill TypedArrayNew
		TypedArraySet TypedArraySort TypedArraySubarray Unary Undefined UnknownKind Unwrap Utf8At Utf8Length Void WeakOf
		WeakTarget`) {
		result[name] = true
	}
	// The seven that can run program code in JavaScript, each with the stop that keeps it out.
	for _, name := range []string{
		// Concat: ToPrimitive on an object operand. Its parts are strings: `+` makes one only of two
		// strings, a primitive converted first and an object left as it is (lower/expression.go:990,
		// concatenated at :1126), so an object operand stops, "<operator> with a <type> and a
		// <type>" (:1121), and a template converts each value or stops, "a template interpolating
		// an object, an array, a map, a function or undefined" (:1194). An object's string by
		// ToPrimitive is to_string_hint's, an ordinary Call (lower/to_primitive.go).
		"Concat",
		// ArrayJoin: an element's toString. Every road that makes one takes a number, maybe-number,
		// boolean or string element only: "join on an array of objects, arrays, maps or functions"
		// (lower/object.go:1572), "join on nested arrays of objects or functions"
		// (lower/library_array.go:622), "numeric coercion of an array with non-primitive elements"
		// (lower/census_coercion.go:31).
		"ArrayJoin",
		// UnionToString: an object's toString. It is made only for a union whose members are numbers,
		// booleans, strings and undefined (writable, lower/expression.go:1288, used at :1176 and
		// lower/library_string.go:92), or past to_string_hint's kind test, and an object goes to
		// to_string_hint, a Call, or stops, "a template interpolating a union with an object, an
		// array, a map or a function in it" (lower/expression.go:1181). The runtime panics on an
		// object member and calls nothing (runtime/union.c:48, adamic_union_to_string).
		"UnionToString",
		// MaybeToString: an object's toString. Its value is a MaybeNumber or a MaybeBoolean, a
		// representation that holds no object, and lowering makes it for those two only
		// (lower/expression.go:1126's concatenated, lower/library_string.go's String).
		"MaybeToString",
		// DynamicProperty: a class getter on an unknown. Natively the read takes a shape's slot, a
		// record's entry or an array's length, and calls no accessor (runtime/union.c:325,
		// dynamic_property); a getter by a run-time key is DynamicKeyedProperty's keyed_getter, off
		// this list.
		"DynamicProperty",
		// JSONStringify: toJSON and getters. Lowering stops an object reference, "JSON.stringify
		// object references (structural types can hide fields and toJSON; runtime shapes need complete
		// value metadata)" (lower/library_json_stringify.go:226), a literal with a method or a toJSON
		// key (:104, :134), a record that may hold a toJSON (:210) and a replacer function (:33).
		"JSONStringify",
		// ObjectCall: assign's setters and getters, and entries' getters. Natively every method reads or
		// writes slots and calls no accessor (runtime/library_object.c: adamic_object_values at :123,
		// adamic_object_assign at :156). Lowering refuses keys, values and entries in a program that
		// writes any getter or setter, "Object enumeration with getter or accessor descriptors"
		// (lower/entries_provenance.go:113), and stops every other method, assign among them, "on a
		// shape not proven by a plain literal or its const binding" (lower/library_object.go:97).
		"ObjectCall",
	} {
		result[name] = true
	}
	return result
}()
