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
	{"Tuple", "a tuple's properties aren't read by a run-time key yet"},
	{"Array", "an array's properties aren't read by a run-time key yet"},
	{"Collection", "a Map's or a Set's properties aren't read by a run-time key yet"},
	{"TypedArray", "a typed array's properties aren't read by a run-time key yet"},
	{"StaticClass", "a class's static properties aren't read by a run-time key yet"},
	{"RuntimeValue", "this runtime value's properties aren't read by a run-time key yet"},
	// A class keeps its private fields under names beginning with #, as string keys on the JavaScript
	// backend, and the JavaScript backend can't tell a record from a plain object, so every such key
	// stops, a record's included.
	{"Private", "a name beginning with # isn't read by a run-time key yet"},
	{"ErrorProperty", "an Error property Node sets and the native Error lacks isn't read yet"},
	{"ObjectLiteralGetter", "an object literal's getter isn't read by a run-time key yet"},
	{"SetterOnly", "a setter without a getter isn't read by a run-time key yet"},
	{"Method", "a method isn't read by a run-time key yet"},
	{"ObjectPrototype", "an Object.prototype name isn't read by a run-time key yet"},
}

// DynamicKeyedPrototypeNames are Object.getOwnPropertyNames(Object.prototype) on Node: a name no
// object of Adamic's holds as its own and every JavaScript object inherits, a function or an accessor
// a run-time key read would hand out as a value.
var DynamicKeyedPrototypeNames = []string{
	"constructor", "__proto__", "toString", "toLocaleString", "valueOf", "hasOwnProperty", "isPrototypeOf",
	"propertyIsEnumerable", "__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__",
}

// DynamicKeyedErrorNames are the properties Node gives its own errors and the native Error doesn't
// have, so on an Error they stop on both backends even where one side has the name: every Error's
// stack, an Error's cause, and a system error's errno, syscall, path and dest. Natively most host
// errors carry name, message and code only, while Node's carry all of a system error's own properties.
var DynamicKeyedErrorNames = []string{"stack", "cause", "errno", "syscall", "path", "dest"}

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
