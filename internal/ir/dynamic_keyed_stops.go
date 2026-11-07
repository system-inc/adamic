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
	{"Collection", "a Map's or a Set's properties aren't read by a run-time key yet"},
	{"TypedArray", "a typed array's properties aren't read by a run-time key yet"},
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
	// A Stats answers what native's shape holds (#zhrjf2t), and a field Node gives it beyond those
	// (DynamicKeyedStatsFieldNames) stops in narrower words than a host object's, until #6xfkdwj.
	{"StatsField", "this Stats field isn't held natively yet"},
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

// DynamicKeyedFunction is a function a run-time key read hands out under its name, and its length on
// Node, which natively its closure carries.
type DynamicKeyedFunction struct {
	Name   string
	Length int
}

// DynamicKeyedFunctions are such functions, in a table's order.
type DynamicKeyedFunctions []DynamicKeyedFunction

// Names are the functions' names, in order.
func (functions DynamicKeyedFunctions) Names() []string {
	names := []string{}
	for _, function := range functions {
		names = append(names, function.Name)
	}
	return names
}

// DynamicKeyedArrayPrototypeNames are Object.getOwnPropertyNames(Array.prototype) on Node, less
// length, which every array holds as its own, each with Node's Array.prototype[name].length. An array
// and a tuple inherit each of them, and Object.prototype's names too (DynamicKeyedPrototypeNames), and
// a run-time key read hands one out as Node does (#g187f1e): on the JavaScript backend Node's own,
// natively one immortal closure for each name with this length (union.c, from
// dynamic_keyed_names.h), whose call stops loudly. An array's constructor, toString and toLocaleString
// are these, never Object.prototype's. TestDynamicKeyedArrayPrototypeNamesAreNodes and
// TestDynamicKeyedArrayPrototypeLengthsAreNodes hold this table to the pinned Node's both ways, so a
// name Node adds goes red rather than reading undefined, and a length that drifts goes red rather than
// reading wrong. A length is read by a static read of the function the key read handed out
// (array_methods.a), never by a run-time key yet: propertyOf(method, 'length') stops in the Function
// form on both backends (stop_function_length.a), and the head that lifts that stop pins it printing
// Node's numbers.
var DynamicKeyedArrayPrototypeNames = DynamicKeyedFunctions{
	{"constructor", 1}, {"at", 1}, {"concat", 1}, {"copyWithin", 2}, {"fill", 1}, {"find", 1},
	{"findIndex", 1}, {"findLast", 1}, {"findLastIndex", 1}, {"lastIndexOf", 1}, {"pop", 0}, {"push", 1},
	{"reverse", 0}, {"shift", 0}, {"unshift", 1}, {"slice", 2}, {"sort", 1}, {"splice", 2},
	{"includes", 1}, {"indexOf", 1}, {"join", 1}, {"keys", 0}, {"entries", 0}, {"values", 0},
	{"forEach", 1}, {"filter", 1}, {"flat", 0}, {"flatMap", 1}, {"map", 1}, {"every", 1}, {"some", 1},
	{"reduce", 1}, {"reduceRight", 1}, {"toReversed", 0}, {"toSorted", 1}, {"toSpliced", 2},
	{"with", 2}, {"toLocaleString", 0}, {"toString", 0},
}

// DynamicKeyedRegExpStringIteratorNames are the string-keyed names a matchAll iterator inherits on
// Node, each with the length of the function a read hands out: Object.getOwnPropertyNames of its
// prototype, %RegExpStringIteratorPrototype% (next), and then of Iterator.prototype, whose constructor
// is an accessor whose getter answers Iterator. The iterator has no own property, and above
// Iterator.prototype is Object.prototype (DynamicKeyedPrototypeNames). A run-time key read of the
// iterator hands one out as Node does (#zhrjf2t, the first host shape it lifts): on the JavaScript
// backend Node's own, natively one immortal closure for each name with this length (union.c, from
// dynamic_keyed_names.h), whose call stops loudly. TestDynamicKeyedRegExpStringIteratorNamesAreNodes
// and TestDynamicKeyedRegExpStringIteratorLengthsAreNodes hold this table to the pinned Node's both
// ways. A length is read by a static read of the function the key read handed out
// (matchall_reads.a), never by a run-time key yet: propertyOf(next, 'length') stops in the Function
// form on both backends (stop_function_length_iterator.a), and the head that lifts that stop pins it
// printing Node's numbers.
var DynamicKeyedRegExpStringIteratorNames = DynamicKeyedFunctions{
	{"next", 0}, {"constructor", 0}, {"reduce", 1}, {"toArray", 0}, {"forEach", 1}, {"some", 1},
	{"every", 1}, {"find", 1}, {"map", 1}, {"filter", 1}, {"take", 1}, {"drop", 1}, {"flatMap", 1},
}

// DynamicKeyedStatsNames are the functions an fs.Stats inherits on Node, each with the length of the
// function a read hands out: Stats.prototype's own constructor (Stats) and _checkModeProperty, then
// StatsBase.prototype's seven is methods, whose own constructor Stats.prototype's shadows. Above
// StatsBase.prototype is Object.prototype (DynamicKeyedPrototypeNames). Stats.prototype's other names
// are its four date accessors, which DynamicKeyedStatsFieldNames and native's shape share between
// them. A run-time key read of a Stats hands one out as Node does (#zhrjf2t, the second host shape
// it lifts): on the JavaScript backend Node's own, natively one immortal closure for each name with
// this length (union.c, from dynamic_keyed_names.h), whose call stops loudly, naming the prototype
// Node's function lives on. TestDynamicKeyedStatsNamesAreNodes and TestDynamicKeyedStatsLengthsAreNodes
// hold this table to the pinned Node's both ways.
var DynamicKeyedStatsNames = DynamicKeyedFunctions{
	{"constructor", 14}, {"_checkModeProperty", 1}, {"isDirectory", 0}, {"isFile", 0}, {"isBlockDevice", 0},
	{"isCharacterDevice", 0}, {"isSymbolicLink", 0}, {"isFIFO", 0}, {"isSocket", 0},
}

// DynamicKeyedStatsFieldNames are the fields Node gives an fs.Stats that native's Stats doesn't hold:
// its own numbers, then Stats.prototype's date accessors, less the four node_fs_file.c's stat shape
// holds (size, mtimeMs, mtime and atime). A run-time key read of one stops on both backends in the
// StatsField form rather than answer a value native would have to make up (#zhrjf2t; #6xfkdwj fills
// them).
// TestDynamicKeyedStatsFieldNamesAreNodes holds this list to the pinned Node's, in order.
var DynamicKeyedStatsFieldNames = []string{
	"dev", "mode", "nlink", "uid", "gid", "rdev", "blksize", "ino", "blocks", "atimeMs", "ctimeMs",
	"birthtimeMs", "ctime", "birthtime",
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
func (program *Program) DynamicKeyedMethodNames(class int) []string {
	names := []string{}
	for _, method := range program.Classes[class-1].Methods {
		if method < 0 {
			continue
		}
		name := program.Functions[method].MethodName
		if strings.HasPrefix(name, "$get:") || strings.HasPrefix(name, "$set:") {
			continue
		}
		names = append(names, name)
	}
	return names
}
