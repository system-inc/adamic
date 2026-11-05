package ir

import "strings"

// DynamicKeyedStop is one of the places a run-time key read (DynamicKeyedProperty) stops, naming why,
// where the representation can't answer as Node does. Each backend's stop prints its Message and then
// " (reading 'key')", so the two stop byte for byte alike. This table is the only one: the JavaScript
// backend emits it into its helper, and internal/native/runtime/dynamic_keyed_names.h is written from
// it (TestDynamicKeyedNamesHeaderIsCurrent), so no list is written twice by hand.
type DynamicKeyedStop struct {
	Name    string
	Message string
}

// DynamicKeysStop is the whole message, with no key, with which both backends stop Object.keys of a
// host value (Adamic's ruling on #ey7aysy at 08:28Z): an enumeration reads no one key, so it doesn't
// say "(reading 'x')". union.c writes the same words (host_shapes_test.go holds them alike).
const DynamicKeysStop = "this runtime value's own keys aren't enumerated yet"

// DynamicKeyedStops are in the order both backends test them, a value's kind first and then, for an
// object, its key.
var DynamicKeyedStops = []DynamicKeyedStop{
	// Lowering proves the value is neither, so this one is a compiler bug said out loud.
	{"Nullish", "compiler bug: a run-time key read reached null or undefined"},
	// Natively, a class's facts come from the table the emitter writes for a program with the read,
	// with a row for every class it has, so one without a row is a compiler bug said out loud.
	{"UnknownClass", "compiler bug: a run-time key read reached a class with no row in its table"},
	// Adamic's words (#zhgxzjg): answering a primitive needs each primitive's whole prototype, and one
	// missing name would read a method as undefined.
	{"Primitive", "a primitive receiver's properties aren't read yet"},
	{"Function", "a function's properties aren't read by a run-time key yet"},
	// An array answers its indexes, its length and a match's own fields (#ys4naww), and reads any other
	// name as undefined but the ones it inherits, Array.prototype's and Object.prototype's, which stop:
	// a method read would hand out a function no call can reach yet.
	{"ArrayPrototype", "an array's methods aren't read by a run-time key yet"},
	{"Collection", "a Map's or a Set's properties aren't read by a run-time key yet"},
	{"TypedArray", "a typed array's properties aren't read by a run-time key yet"},
	// Lowering refuses a tuple viewed as unknown, but one still reaches the read as a value an unknown
	// holds: an object's field, an array's element, an Object.entries pair, a match's indices pair.
	// Natively a tuple is an object marked tuple, whose runtime-made pairs carry no slot types, so it
	// isn't read as an array yet. The JavaScript backend marks the same values (adamicTuple), so both
	// stop alike.
	{"Tuple", "a tuple's properties aren't read by a run-time key yet"},
	{"StaticClass", "a class's static properties aren't read by a run-time key yet"},
	{"RuntimeValue", "this runtime value's properties aren't read by a run-time key yet"},
	// A class keeps its private fields under names beginning with #, as string keys on the JavaScript
	// backend, and the JavaScript backend can't tell a record from a plain object, so every such key
	// stops, a record's included.
	{"Private", "a name beginning with # isn't read by a run-time key yet"},
	// Object.prototype's accessor answers with the receiver's prototype, which no object of Adamic's
	// carries as a value, and a record may hold the name as its own where Node's setter took it.
	{"Proto", "__proto__ isn't read by a run-time key yet"},
	{"ErrorProperty", "an Error property Node sets and the native Error lacks isn't read yet"},
	{"ObjectLiteralGetter", "an object literal's getter isn't read by a run-time key yet"},
	{"SetterOnly", "a setter without a getter isn't read by a run-time key yet"},
	{"Method", "a method isn't read by a run-time key yet"},
	// A class instance's constructor is its class, a value the read can't find natively: an instance
	// carries its class's metadata, never the class object the program holds.
	{"ClassConstructor", "a class instance's constructor isn't read by a run-time key yet"},
}

// DynamicKeyedPrototypeMethods are Object.prototype's ten methods, the functions every object of
// Adamic's inherits under these names. A run-time key read hands one out as Node does: on the
// JavaScript backend Node's own, natively one immortal closure for each name (union.c), the same from
// every receiver and every read, whose call stops loudly, since no call can reach one yet.
var DynamicKeyedPrototypeMethods = []string{
	"toString", "toLocaleString", "valueOf", "hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable",
	"__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__",
}

// DynamicKeyedPrototypeNames are Object.getOwnPropertyNames(Object.prototype) on Node: Object as
// constructor, the __proto__ accessor, and the ten methods.
var DynamicKeyedPrototypeNames = append([]string{"constructor", "__proto__"}, DynamicKeyedPrototypeMethods...)

// DynamicKeyedArrayPrototypeNames are Object.getOwnPropertyNames(Array.prototype) on Node, less
// length, which every array holds as its own. An array inherits each of them, and Object.prototype's
// names too (DynamicKeyedPrototypeNames), so a run-time key read of any stops (ArrayPrototype).
// TestDynamicKeyedArrayPrototypeNamesAreNodes holds this list to the pinned Node's both ways, so a
// name Node adds goes red rather than reading undefined.
var DynamicKeyedArrayPrototypeNames = []string{
	"constructor", "at", "concat", "copyWithin", "fill", "find", "findIndex", "findLast", "findLastIndex",
	"lastIndexOf", "pop", "push", "reverse", "shift", "unshift", "slice", "sort", "splice", "includes",
	"indexOf", "join", "keys", "entries", "values", "forEach", "filter", "flat", "flatMap", "map", "every",
	"some", "reduce", "reduceRight", "toReversed", "toSorted", "toSpliced", "with", "toLocaleString",
	"toString",
}

// DynamicKeyedErrorNames are the properties Node gives its own errors and the native Error doesn't
// have, so on an Error they stop on both backends even where one side has the name: every Error's
// stack, an Error's cause, and a system error's errno, syscall, path and dest. Natively most host
// errors carry name, message and code only, while Node's carry all of a system error's own properties.
// An Error's toString and constructor are Error.prototype's, never Object.prototype's, so the closures
// a plain object's read hands out would be the wrong functions there.
var DynamicKeyedErrorNames = []string{"stack", "cause", "errno", "syscall", "path", "dest", "toString", "constructor"}

// DynamicKeyedMethodNames are the names a class's method table holds, the inherited ones included, as
// both backends judge a run-time key against them: Methods by MethodName, less a getter's or a
// setter's own entry, which the class's accessors answer.
func (p *Program) DynamicKeyedMethodNames(class int) []string {
	names := []string{}
	for _, method := range p.Classes[class-1].Methods {
		if method < 0 {
			continue
		}
		name := p.Functions[method].MethodName
		if strings.HasPrefix(name, "$get:") || strings.HasPrefix(name, "$set:") {
			continue
		}
		names = append(names, name)
	}
	return names
}
