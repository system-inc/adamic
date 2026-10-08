// Package ir is Adamic's intermediate representation: what the backends compile, with every type
// already proven by the checker.
//
// It holds exactly what the programs stage 0 can compile need, and grows one program at a time
// (docs/0.1.md). Nothing in it is speculative: an instruction exists because a fixture lowers to it
// and the oracle has checked what it prints.
package ir

import "fmt"

// Program is one compiled Adamic program.
type Program struct {
	// Source is the entry file's base name, as written, for the header of what the backends emit.
	Source string

	// Strings are the program's string constants, as UTF-8, in first-use order.
	Strings []string

	// Regexps contain immutable C bytecode compiled during lowering.
	Regexps []RegExpProgram

	// Locals are every variable the program declares, each its own, so shadowing is already resolved.
	Locals []Local

	// Functions are the module's function declarations, callable from anywhere in it.
	Functions []Function

	// Classes carry nominal identity, prefix field layouts and method slots. IDs are one-based.
	Classes       []Class
	MethodTargets map[int][]int

	// Main is what the program does, in order.
	Main []Statement

	// ClosuresMayThrow says a function value somewhere in the program can throw, or a class's method,
	// which a call through an interface reaches where it would a function value. Which one a call
	// through a function value reaches isn't known, so every such call can then throw: one written
	// out, and the ones the runtime's loops make (map, the visits, reduce, Array.from, sort), whose
	// callers test for it after each.
	ClosuresMayThrow bool
}

// Class is a class instantiation. Base is zero for a root; Methods has the base slots as a prefix.
type Class struct {
	// Definition is the erased source identity, shared by distinct native layouts.
	Definition   int
	Name         string
	Base         int
	Constructor  int
	Fields       []Field
	OwnStart     int
	Methods      []int
	Accessors    []Accessor
	Literal      bool
	PublicFields []Field
	Static       bool
	StaticParent int   // one-based hidden parent slot, zero for a root
	StaticFlags  []int // one-based presence slot for each data slot, zero for internal storage
}

type Accessor struct {
	Name           string
	Getter, Setter int
}

// Function is a function declaration.
type Function struct {
	Name string

	// Parameters are locals, in order.
	Parameters []int

	// Returns is the result's type, or 0 for void.
	Returns Type

	Body []Statement

	// Closure is a function value made where it's written (an arrow function), and Environment the
	// captured variables it reaches through its cells, in order.
	Closure     bool
	Environment []int

	// Receiver marks a literal method closure whose first parameter receives the calling object.
	Receiver bool

	// MayThrow is a function a throw can leave (docs/memory.md, "Exceptions"): its callers test for
	// one after each call. Lowering works it out over the call graph once every function is lowered.
	MayThrow bool
}

// Type is a value's representation. The checker proved the TypeScript type; this is what's left of
// it at runtime.
type Type int

const (
	Number Type = iota + 1
	Boolean
	String

	// Object is a plain object, of any shape: which fields it has is the checker's business, and the
	// runtime finds each by name.
	Object

	// Array is an array; its element type travels on the expressions and statements that read it.
	Array

	// Map is a Map; its key and value types travel on the expressions that read it. A Set is one too,
	// its values unused: the same order, keys and iteration (lower/set.go).
	Map

	// MaybeNumber is number | undefined: a number that may be missing, as Map.get gives it.
	MaybeNumber

	// Closure is a function value: code, and the variables it captured.
	Closure

	// MaybeBoolean is boolean | undefined: a boolean that may be missing, as an array of booleans'
	// element is.
	MaybeBoolean

	// Union is a value of a union whose members are held differently (string | number, number |
	// Tree): one counted reference, which says at runtime which member it is. A number is boxed to
	// be one, a boolean is one of two constant boxes, and undefined is a null reference.
	Union

	// Weak is where a Weak<Target> is kept (docs/memory.md): a reference that doesn't count, held as
	// a counted handle the target is found through, which says undefined once the target is freed.
	// A value of the type exists only where it's kept (a variable, a parameter, a field, an element,
	// a map's value); reading one is WeakTarget, and keeping one is WeakOf.
	Weak

	// Typed arrays hold numbers in one flat buffer of the element width.
	Uint8Array
	Int32Array
	Float64Array
)

// Maybe is the type of a value of type t that may be missing: number | undefined and boolean |
// undefined are each a pair of present and the value, and a reference is itself, missing as null.
func Maybe(t Type) Type {
	switch t {
	case Number:
		return MaybeNumber
	case Boolean:
		return MaybeBoolean
	}
	return t
}

// IsMaybe reports whether a value of the type is one of Maybe's pairs.
func (t Type) IsMaybe() bool {
	return t == MaybeNumber || t == MaybeBoolean
}

// Present is the type a Maybe pair holds when it's present, and any other type itself.
func (t Type) Present() Type {
	switch t {
	case MaybeNumber:
		return Number
	case MaybeBoolean:
		return Boolean
	}
	return t
}

// IsReference reports whether a value of the type lives on the heap and is counted.
func (t Type) IsReference() bool {
	return t == String || t == Object || t == Array || t == Map || t == Closure || t == Union || t == Weak || t.IsTypedArray()
}

// Local is a variable: its name as written, for reading the output, and its type.
type Local struct {
	Name string
	Type Type

	// Global is a variable declared at the module's top level, which functions can read and write.
	Global bool

	// Function is the function that declares it, -1 for the module's top level.
	Function int

	// Captured is a variable some closure reads or writes: it lives in a cell, shared by reference.
	Captured bool

	// Borrowed is a reference parameter the function only looks at: its caller keeps the value alive
	// for the whole call, so the function neither retains it on entry nor releases it on the way out
	// (docs/memory.md, "Borrowed parameters"). Nothing ever assigns a borrowed parameter.
	Borrowed bool

	// ConstantClosure is the function index plus one for a const initialized directly
	// with a closure literal. Zero means no such proof.
	ConstantClosure int

	// Counter is a for loop's counter proven to hold only whole numbers no larger than 2^53, each a
	// double exactly, so the native backend keeps it in an integer and reads it as the same double
	// (internal/lower/counters.go). Only the loop's update ever writes it, by a whole constant step.
	Counter bool
}

// Expression is a value. Evaluating a String expression yields a reference its consumer owns: the
// backend releases it once used, or hands it to a variable that does.
type Expression interface {
	Type() Type
}

type (
	NumberConstant  struct{ Value float64 }
	BooleanConstant struct{ Value bool }
	StringConstant  struct{ Index int }

	// Read reads a local. Of is the local's type, so a Read is typed without the program in hand.
	// Checked is a global read whose initialization is not proven, including function
	// bodies and cyclic module evaluation. JavaScript throws in the temporal dead zone;
	// Adamic checks it out loud.
	Read struct {
		Local   int
		Of      Type
		Checked bool
	}

	// Call calls a function. Returns is its result type, 0 for void.
	Call struct {
		Function  int
		Arguments []Expression
		Returns   Type

		// Virtual is a one-based method slot. Function supplies its static signature.
		Virtual  int
		Accessor string
		Setter   bool
	}

	HasAccessor struct {
		Object Expression
		Name   string
	}

	// InstanceOf tests nominal identity along a class ancestry chain.
	InstanceOf struct {
		Value Expression
		Class int
		Exact bool
	}

	// Unary is -, +, ! and ~ on its operand.
	Unary struct {
		Operator Operator
		Operand  Expression
	}

	// Binary is an operator whose operands are already of the types it takes (the checker and
	// lowering saw to that): arithmetic on numbers, comparison of numbers, equality of like types,
	// bitwise operations on numbers, and && and || on booleans or maybe booleans, which short-circuit.
	Binary struct {
		Operator    Operator
		Left, Right Expression
	}

	// NumberToString is String(value), as a template literal and string + number write it.
	NumberToString struct{ Value Expression }

	// BooleanToString is String(value): "true" or "false".
	BooleanToString struct{ Value Expression }

	// Concat joins strings, as + and template literals do.
	Concat struct{ Parts []Expression }

	// Conditional is the ?: operator.
	Conditional struct {
		Condition         Expression
		WhenTrue, WhenNot Expression

		// Of, when set, is what the conditional gives where its branches' own types don't say:
		// flag ? text : undefined is a string that may be missing, though one branch is undefined.
		Of Type
	}

	// ObjectLiteral makes an object. With Spread, it's { ...Spread, fields }: a copy of Spread's
	// object, whatever its shape, with Fields replaced (each one a field Spread's type has).
	//
	// SpreadMaybeUndefined says Spread may be undefined, and then JavaScript's { ...undefined } is
	// {}: the object made is Empty, each of the source type's fields the literal doesn't give, as
	// undefined (what JavaScript reads from a field that isn't there), with Fields written into it.
	ObjectLiteral struct {
		// Class is the nominal class ID, or zero for a plain object.
		Class                int
		Spread               Expression
		Fields               []Field
		NoReuse              bool
		SpreadMaybeUndefined bool
		Empty                []Field

		// Tuple is a tuple written out, [key, value]: natively an object whose fields are named "0",
		// "1" and on, as every tuple is, and in JavaScript an array, as the source's is. (What 0.2
		// lowers reads a tuple only by its fields, which an object answers the same way, new Map's
		// pairs included: ECMA-262 reads each by "0" and "1".)
		Tuple bool

		// Methods are, for the object a class's constructor makes, the class's methods, which a call
		// through an interface the class implements finds by name (Property.Method).
		Methods []Method
	}

	// Property reads a field. Of is its type. Optional is ?., which is undefined when Object is: a
	// number field read that way is number | undefined.
	Property struct {
		Object   Expression
		Name     string
		Of       Type
		Optional bool
		// Absent is an optional own field: a shape without it reads as undefined.
		Absent bool
		// Class is, when the field is one of a class's, that class's constructor plus one, and 0
		// otherwise: the constructor's object has the class's layout, so the field's place in it is
		// known, for an object that has that layout.
		Class int
		// Method says the read is a call's callee, object.name(...), through a type that isn't a
		// class: an interface or an object type. The object may be a class's, whose methods aren't
		// fields, so the call takes the object's own function value if it has one and otherwise its
		// class's method, called with the object as this. A method can't be read any other way
		// (docs/0.1.md), so only a callee is one.
		Method bool
	}

	// ArrayLiteral makes an array. Where Spread is set, the element at that position is an array of the
	// same elements, spread into this one at that point in the evaluation, as JavaScript does.
	ArrayLiteral struct {
		Element  Type
		Elements []Expression
		Spread   []bool
	}

	// Length is array.length.
	Length struct {
		Array Expression

		// Optional is array?.length: undefined, a number | undefined, where the array is.
		Optional bool
	}

	// StringFromCodes is String.fromCharCode(...Codes), or String.fromCodePoint when CodePoints is
	// set: a string of the UTF-16 units, or of the code points, the numbers name.
	StringFromCodes struct {
		Codes      []Expression
		CodePoints bool

		// Spread, when set, is the codes instead, as an array of numbers: fromCharCode(...codes).
		Spread Expression
	}

	// MathCall is Math.<Function>(...), on numbers.
	MathCall struct {
		Function  string
		Arguments []Expression

		// Spread, when set, is the arguments instead, as an array of numbers: Math.max(...values).
		Spread Expression
	}

	// NumberCall is Number.<Function>(...): parseInt (text, and a radix perhaps left out), parseFloat
	// (text), and isNaN, isFinite, isInteger and isSafeInteger (a number).
	NumberCall struct {
		Function  string
		Arguments []Expression
	}

	// ToFixed is Value.toFixed(Digits).
	ToFixed struct{ Value, Digits Expression }

	// NumberFormat is Value.toExponential(Argument), Value.toPrecision(Argument) or
	// Value.toString(Argument), the radix, as Method says. Argument is nil when the call has none,
	// which means something of its own to each (the shortest digits, String(Value), and radix 10).
	NumberFormat struct {
		Method   string
		Value    Expression
		Argument Expression
	}

	// Undefined is undefined where a reference goes: an object, an array or a string that may be
	// missing (Tree | undefined), held as a null pointer. Of is the reference it stands in for, a
	// string where a string | undefined is returned, and an object when it isn't set.
	Undefined struct{ Of Type }

	// IsUndefined is Value === undefined, for a reference that may be missing.
	IsUndefined struct{ Value Expression }

	// ArrayPush is Array.push(Value): it appends and is the new length.
	ArrayPush struct {
		Array   Expression
		Value   Expression
		Element Type
		// Site is which write of the program this is, for the cycle finder (lowering keeps the type of
		// what it writes into), or 0 when nothing recorded one.
		Site int
	}

	// Unwrap is a Maybe pair the checker has proven present (narrowed), as what it holds. A narrowing
	// outlives a call that assigns the variable again (the checker doesn't look inside the call), so
	// it's checked, in both backends: undefined there panics.
	Unwrap struct{ Value Expression }

	// Defined is a reference the checker narrowed undefined out of, checked for the same reason as
	// Unwrap: undefined there panics with Message. Where the value is about to be read through a
	// property, Message is the TypeError JavaScript throws there, so the check is what Node does.
	Defined struct {
		Null    bool
		Value   Expression
		Message string
	}

	// MaybeOf is a number or a boolean where Of, its Maybe pair, goes: Value, present, or undefined
	// when Value is nil.
	MaybeOf struct {
		Value Expression
		Of    Type
	}

	// MaybeToString is String(Value) for a Maybe pair: what it holds, written as String() writes it,
	// or "undefined".
	MaybeToString struct{ Value Expression }

	// Box is Value where a Union goes: a number boxed, a boolean as its box, a reference as itself.
	Box struct{ Value Expression }

	// Narrow is a Union the checker has proven to be one member (by typeof, ===, or assignment), as
	// that member's type To, which may be a Maybe pair (number | undefined, out of string | number |
	// undefined).
	Narrow struct {
		Value Expression
		To    Type
	}

	// TypeOf is typeof Value: "number", "string", "boolean", "undefined", "object" or "function".
	TypeOf struct {
		Value Expression
		// Null says a present value's NULL pointer is null. An absent lookup slot is still undefined.
		Null bool
	}

	// KindIs checks a represented object kind before a union is narrowed. Value is a Union.
	KindIs struct {
		Value Expression
		Of    Type
	}

	// MakeError is an Error with Message and an optional Name (nil means "Error").
	MakeError struct{ Message, Name Expression }

	// WeakOf is Value, a reference, kept weakly: the handle to it, made if it has none yet, or
	// undefined when Value is.
	WeakOf struct{ Value Expression }

	// WeakTarget is what a Weak value points to, as To: undefined once the target is freed. Present
	// is a read the checker narrowed to present, which panics natively if the target was freed since
	// (JavaScript would still have it; docs/memory.md).
	WeakTarget struct {
		Value   Expression
		To      Type
		Present bool
	}

	// UnionToString is String(Value) for a Union whose members are numbers, booleans, strings and
	// undefined, each written as String() writes it.
	UnionToString struct{ Value Expression }

	// Coalesce is Value ?? Fallback: Value when it's present, and otherwise Fallback, evaluated only
	// then. With Panic set instead of Fallback, it's Value ?? panic(Panic).
	Coalesce struct {
		Value    Expression
		Fallback Expression
		Panic    Expression
		Of       Type
	}

	// StringLength is string.length, in UTF-16 code units.
	StringLength struct {
		Value Expression

		// Optional is text?.length: undefined, a number | undefined, where the string is.
		Optional bool
	}

	// CharCodeAt is string.charCodeAt(Index): a UTF-16 code unit, or NaN.
	CharCodeAt struct{ Value, Index Expression }

	// Trim is string.trim().
	Trim struct{ Value Expression }

	// StringCall is one of a string's methods with JavaScript's meaning: slice, codePointAt,
	// padStart, padEnd, repeat, indexOf, lastIndexOf, includes, startsWith, endsWith, split,
	// trimStart, trimEnd, at (a string, or undefined: a null reference), replace and replaceAll
	// with a string pattern, toUpperCase and toLowerCase, and normalize (its form filled in). Arguments are as written; lowering filled in any default.
	StringCall struct {
		Method    string
		Value     Expression
		Arguments []Expression
	}

	// CodePoints is [...string]: an array of its code points, each a string.
	CodePoints struct{ Value Expression }

	// CheckedCast is value as Member, where value is a discriminated union and Member some of its
	// members: the discriminant Field must hold one of Allowed, or the program panics with Message,
	// in both backends (docs/0.1.md, decision 5).
	CheckedCast struct {
		Value     Expression
		Field     string
		FieldType Type
		Allowed   []Expression
		Message   string
	}

	// ArrayIndex is array[index]: the element, or undefined when index isn't one of the array's (a
	// null reference, or a Maybe pair). Relative is array.at(index), where a negative index counts
	// from the end and a fraction truncates.
	ArrayIndex struct {
		Array, Index Expression
		Element      Type
		Relative     bool
	}

	// ArraySearch is array.indexOf(Value), with ===, and array.includes(Value), with SameValueZero,
	// which finds NaN.
	ArraySearch struct {
		Array, Value Expression
		From         Expression
		Element      Type
		Includes     bool
		Last         bool
	}

	// ArraySplice is array.splice(Start, Count, ...Items): Count perhaps left out (everything after
	// Start), and what's removed, a new array.
	ArraySplice struct {
		Array, Start, Count Expression
		Items               []Expression
		Element             Type
		// Site is which write of the program this is, for the cycle finder (lowering keeps the type of
		// what it writes into), or 0 when nothing recorded one.
		Site int
	}

	// ArrayFill is array.fill(Value, Start, End), Start and End perhaps nil (left out); in place, and
	// the array. With Array nil, it's new Array(Length).fill(Value): a new array, every element Value.
	ArrayFill struct {
		Array, Length, Value, Start, End Expression
		Element                          Type
		// Site is which write of the program this is, for the cycle finder (lowering keeps the type of
		// what it writes into), or 0 when nothing recorded one.
		Site int
	}

	// ArrayFrom is Array.from({ length: Length }, Callback): a new array of Length elements (ToLength,
	// and more than 2^32 - 1 panics as JavaScript throws), each the callback's result called with
	// undefined and its index, in order. First is the type of the callback's first parameter, which
	// undefined is passed as (0 when it has none).
	ArrayFrom struct {
		Length, Callback Expression
		Element, First   Type
	}

	// ArrayReverse is array.reverse(): in place, and the array.
	ArrayReverse struct{ Array Expression }

	// ArrayConcat is array.concat(Others...): a new array of every one's elements, in order. Each of
	// Others is an array of the same elements.
	ArrayConcat struct {
		Array  Expression
		Others []Expression
	}

	// ArrayReduce is array.reduce(Callback, Initial): the callback called per element with what it
	// last returned (Initial the first time), the element, its index and the array, read and skipped
	// as ArrayVisit does. Result is Initial's type, and the callback's.
	ArrayReduce struct {
		Array, Callback, Initial Expression
		Element, Result          Type
	}

	// StringIndex is string[index]: the UTF-16 code unit there, as a string, or undefined when index
	// isn't one of the string's (a null reference).
	StringIndex struct{ Value, Index Expression }

	// ArrayPop is array.pop(): the last element, removed, or undefined when there's none (a null
	// reference, or a Maybe pair).
	ArrayPop struct {
		Array   Expression
		Element Type
	}

	// MakeClosure makes a closure of a function, capturing the cells of its Environment.
	MakeClosure struct{ Function int }

	// CallClosure calls a function value. Returns is its result type, 0 for void.
	CallClosure struct {
		Closure   Expression
		Arguments []Expression
		Returns   Type
	}

	// ArrayMap is array.map(callback): a new array of the callback's results, each called with the
	// element, its index and the array.
	ArrayMap struct {
		Array    Expression
		Callback Expression
		Element  Type
		Result   Type
	}

	// ArrayVisit is one of the array methods that call a function per element, in order, with the
	// element, its index and the array: forEach, filter, some, every, find and findIndex. The length
	// is read once, before the first call, and an index the array no longer has when its turn comes
	// is skipped, both as JavaScript does. Returns is what the callback returns, 0 for nothing; every
	// method but forEach requires a boolean.
	ArrayVisit struct {
		Method   string
		Array    Expression
		Callback Expression
		Element  Type
		Returns  Type
	}

	// MapEntries is [...map]: an array of [key, value] pairs, each a tuple, an object whose fields
	// are named "0" and "1".
	MapEntries struct {
		Map                Expression
		KeyType, ValueType Type
	}

	// ArraySlice is array.slice(start, end), either argument perhaps left out.
	ArraySlice struct {
		Array     Expression
		Arguments []Expression
	}

	// ArraySort is array.sort(comparator): one of the module's functions (Comparator), or a function
	// value (Callback, when it isn't nil). It sorts in place, stably, and is the array.
	ArraySort struct {
		Array      Expression
		Comparator int
		Callback   Expression
		Element    Type
	}

	// MapNew is new Map(), or new Map([[key, value], ...]) with the pairs written out.
	MapNew struct {
		Key, Value Type
		Entries    [][2]Expression

		// Pairs, when it's set, is an array of [key, value] tuples the map is made from instead, each
		// set in order: new Map(pairs), or new Map(otherMap) through its entries.
		Pairs Expression
	}

	// MapKeys and MapValues are [...map.keys()] and [...map.values()]: new arrays, in insertion order.
	MapKeys struct {
		Map Expression
		Key Type
	}
	MapValues struct {
		Map   Expression
		Value Type
	}

	// MapClear is map.clear() and set.clear(), which is void.
	MapClear struct{ Map Expression }

	// MapForEach is map.forEach(Callback), called with each value, its key and the map, and
	// set.forEach(Callback) (Set), with each element twice and the set, in insertion order and live as
	// for...of is. Returns is what the callback returns, 0 for nothing; forEach itself is void.
	MapForEach struct {
		Map, Callback Expression
		Key, Value    Type
		Set           bool
		Returns       Type
	}

	// MapGet is map.get(Key): the value, or undefined (a null reference, or a Maybe pair).
	MapGet struct {
		Map, Key  Expression
		KeyType   Type
		ValueType Type
	}

	// MapSet is map.set(Key, Value), which is the map.
	MapSet struct {
		Map, Key, Value Expression
		KeyType         Type
		ValueType       Type
		// Site is which write of the program this is, for the cycle finder (lowering keeps the type of
		// what it writes into), or 0 when nothing recorded one.
		Site int
	}

	// MapHas is map.has(Key), and MapDelete map.delete(Key).
	MapHas struct {
		Map, Key Expression
		KeyType  Type
	}
	MapDelete struct {
		Map, Key Expression
		KeyType  Type
	}

	// SetNew is new Set(), or new Set(Values), an array of the elements, each added in order.
	SetNew struct {
		Element Type
		Values  Expression
	}

	// SetAdd is set.add(Value), which is the set. An element it already has keeps its place. has,
	// delete and size are a Map's (MapHas, MapDelete, MapSize).
	SetAdd struct {
		Set, Value Expression
		Element    Type
		// Site is which write of the program this is, for the cycle finder (lowering keeps the type of
		// what it writes into), or 0 when nothing recorded one.
		Site int
	}

	// SetValues is [...set]: a new array of its elements, in order.
	SetValues struct {
		Set     Expression
		Element Type
	}

	// MapSize is map.size.
	MapSize struct{ Map Expression }

	// HasOwn is object.hasOwnProperty(Key): whether one of the object's own fields has that name.
	// The method lives on Object's prototype, which a shape does not store, so it is not a read of a
	// field named hasOwnProperty.
	HasOwn struct {
		Object Expression
		Key    Expression
	}

	// ArrayJoin is Array.join(Separator), writing each element as String() would.
	ArrayJoin struct {
		Array     Expression
		Separator Expression
		Element   Type
		Depth     int
	}

	// ReadTextFile is readTextFile(Path) from 'adamic': the file's bytes decoded as UTF-8 the way
	// Node's readFileSync(path, 'utf8') decodes them, in { kind: 'Ok', text }, or what went wrong in
	// { kind: 'Error', message }, a message in Adamic's own words.
	ReadTextFile struct{ Path Expression }

	// ProgramArguments is programArguments() from 'adamic': a new array of the arguments after the
	// program, as process.argv.slice(2) is.
	ProgramArguments struct{}

	// Utf8Length is utf8Length(Text) from 'adamic', and Utf8At utf8At(Text, Index): the text's UTF-8,
	// read in place.
	Utf8Length struct{ Text Expression }
	Utf8At     struct{ Text, Index Expression }

	// WriteTextFile is writeTextFile(Path, Text) from 'adamic': the file made or emptied, then Text
	// written as UTF-8 the way Node's writeFileSync(path, text) writes it (a lone surrogate as U+FFFD),
	// in { kind: 'Ok' }, or what went wrong in { kind: 'Error', message }.
	WriteTextFile struct{ Path, Text Expression }

	// ReadDirectory is readDirectory(Path) from 'adamic': { kind: 'Ok', names }, the names as Node's
	// readdirSync gives them on the same machine (sorted by their bytes, without . and ..), or
	// { kind: 'Error', message }.
	ReadDirectory struct{ Path Expression }

	// FileStatus is fileStatus(Path) from 'adamic': { kind: 'Ok', type, size, symbolicLink }, the type
	// and size of what Path names, a symbolic link followed, and whether Path is itself one, or
	// { kind: 'Error', message }.
	FileStatus struct{ Path Expression }
)

// Field is one field of an object literal.
type Field struct {
	Name    string
	Value   Expression
	Private bool
}

// Method is one of a class's methods: its name, and the function that is it, whose first parameter
// is this.
type Method struct {
	Name     string
	Function int
}

func (InstanceOf) Type() Type  { return Boolean }
func (HasAccessor) Type() Type { return Boolean }

func (NumberConstant) Type() Type  { return Number }
func (BooleanConstant) Type() Type { return Boolean }
func (StringConstant) Type() Type  { return String }
func (NumberToString) Type() Type  { return String }
func (BooleanToString) Type() Type { return String }
func (Concat) Type() Type          { return String }
func (c Conditional) Type() Type {
	if c.Of != 0 {
		return c.Of
	}
	return c.WhenTrue.Type()
}
func (ObjectLiteral) Type() Type { return Object }
func (p Property) Type() Type {
	if p.Optional {
		return Maybe(p.Of)
	}
	return p.Of
}
func (ArrayLiteral) Type() Type { return Array }
func (l Length) Type() Type {
	if l.Optional {
		return MaybeNumber
	}
	return Number
}
func (MathCall) Type() Type        { return Number }
func (StringFromCodes) Type() Type { return String }

func (c NumberCall) Type() Type {
	if c.Function == "parseInt" || c.Function == "parseFloat" || c.Function == "convert" {
		return Number
	}
	return Boolean
}
func (ToFixed) Type() Type      { return String }
func (NumberFormat) Type() Type { return String }
func (u Undefined) Type() Type {
	if u.Of != 0 {
		return u.Of
	}
	return Object
}
func (IsUndefined) Type() Type   { return Boolean }
func (ArrayPush) Type() Type     { return Number }
func (ArrayJoin) Type() Type     { return String }
func (u Unwrap) Type() Type      { return u.Value.Type().Present() }
func (d Defined) Type() Type     { return d.Value.Type() }
func (m MaybeOf) Type() Type     { return m.Of }
func (MaybeToString) Type() Type { return String }
func (Box) Type() Type           { return Union }
func (n Narrow) Type() Type      { return n.To }
func (KindIs) Type() Type        { return Boolean }
func (TypeOf) Type() Type        { return String }
func (UnionToString) Type() Type { return String }
func (WeakOf) Type() Type        { return Weak }
func (MakeError) Type() Type     { return Object }
func (w WeakTarget) Type() Type  { return w.To }
func (c Coalesce) Type() Type    { return c.Of }
func (l StringLength) Type() Type {
	if l.Optional {
		return MaybeNumber
	}
	return Number
}
func (CharCodeAt) Type() Type    { return Number }
func (Trim) Type() Type          { return String }
func (CodePoints) Type() Type    { return Array }
func (StringIndex) Type() Type   { return String }
func (MapEntries) Type() Type    { return Array }
func (MakeClosure) Type() Type   { return Closure }
func (CheckedCast) Type() Type   { return Object }
func (c CallClosure) Type() Type { return c.Returns }
func (ArrayMap) Type() Type      { return Array }

func (v ArrayVisit) Type() Type {
	switch v.Method {
	case "filter":
		return Array
	case "some", "every":
		return Boolean
	case "findIndex", "findLastIndex":
		return Number
	case "find", "findLast":
		return Maybe(v.Element)
	}
	return 0
}

func (i ArrayIndex) Type() Type { return Maybe(i.Element) }
func (p ArrayPop) Type() Type   { return Maybe(p.Element) }
func (ArraySlice) Type() Type   { return Array }

func (s ArraySearch) Type() Type {
	if s.Includes {
		return Boolean
	}
	return Number
}
func (ArrayReverse) Type() Type  { return Array }
func (ArraySplice) Type() Type   { return Array }
func (ArrayFill) Type() Type     { return Array }
func (ArrayFrom) Type() Type     { return Array }
func (ArrayConcat) Type() Type   { return Array }
func (r ArrayReduce) Type() Type { return r.Result }
func (ArraySort) Type() Type     { return Array }

func (c StringCall) Type() Type {
	switch c.Method {
	case "codePointAt":
		return MaybeNumber
	case "indexOf", "lastIndexOf":
		return Number
	case "includes", "startsWith", "endsWith":
		return Boolean
	case "split":
		return Array
	}
	return String
}
func (ReadTextFile) Type() Type     { return Object }
func (ProgramArguments) Type() Type { return Array }
func (Utf8Length) Type() Type       { return Number }
func (Utf8At) Type() Type           { return Number }
func (WriteTextFile) Type() Type    { return Object }
func (ReadDirectory) Type() Type    { return Object }
func (FileStatus) Type() Type       { return Object }

func (MapNew) Type() Type     { return Map }
func (MapKeys) Type() Type    { return Array }
func (MapValues) Type() Type  { return Array }
func (MapClear) Type() Type   { return 0 }
func (MapForEach) Type() Type { return 0 }
func (SetNew) Type() Type     { return Map }
func (SetAdd) Type() Type     { return Map }
func (SetValues) Type() Type  { return Array }
func (MapSet) Type() Type     { return Map }
func (MapHas) Type() Type     { return Boolean }
func (MapDelete) Type() Type  { return Boolean }
func (MapSize) Type() Type    { return Number }
func (HasOwn) Type() Type     { return Boolean }

func (g MapGet) Type() Type { return Maybe(g.ValueType) }

func (r Read) Type() Type { return r.Of }
func (c Call) Type() Type { return c.Returns }

func (u Unary) Type() Type {
	if u.Operator == Not {
		return Boolean
	}
	return Number
}

func (b Binary) Type() Type {
	if (b.Operator == And || b.Operator == Or) && b.Left.Type() == MaybeBoolean && b.Right.Type() == MaybeBoolean {
		return MaybeBoolean
	}
	switch b.Operator {
	case Add, Subtract, Multiply, Divide, Remainder, Power, BitAnd, BitOr, BitXor, ShiftLeft, ShiftRight, ShiftRightUnsigned:
		return Number
	case Less, LessOrEqual, Greater, GreaterOrEqual, Equal, NotEqual, And, Or:
		return Boolean
	}
	// An operator neither list names would be typed by a guess, and a wrong guess compiles wrong.
	panic(fmt.Sprintf("ir: a binary operator %d with no type", b.Operator))
}

// Operator is an arithmetic, comparison, equality or logical operator.
type Operator int

const (
	Add Operator = iota + 1
	Subtract
	Multiply
	Divide
	Remainder
	Power
	Less
	LessOrEqual
	Greater
	GreaterOrEqual
	Equal
	NotEqual
	And
	Or
	Not
	Negate
	Plus

	// The bitwise operators, on numbers, each through ToInt32 or ToUint32 as JavaScript's are: &, |,
	// ^, <<, >>, >>> and ~.
	BitAnd
	BitOr
	BitXor
	ShiftLeft
	ShiftRight
	ShiftRightUnsigned
	BitNot
)

// Statement is one step of a program.
type Statement interface {
	statement()
}

// Stream is where a line is written.
type Stream int

const (
	Stdout Stream = 1
	Stderr Stream = 2
)

type (
	// WriteLine writes a string and a newline: console.log and console.error with one string.
	WriteLine struct {
		Stream Stream
		Value  Expression
	}

	// Declare introduces a local with its first value.
	Declare struct {
		Local int
		Value Expression
	}

	// Assign gives a local a new value, releasing the old one if it's a string. Checked is as for
	// Read: a write to a global from inside a function.
	Assign struct {
		Local   int
		Value   Expression
		Checked bool
	}

	// Evaluate evaluates an expression for its effects and discards the value: a call as a statement.
	Evaluate struct{ Value Expression }

	// Panic is the prelude's panic(message): write "adamic: panic: <message>" and exit 70.
	Panic struct{ Message Expression }

	// SetIndex is array[index] = value, at an index the array has: anywhere else it panics, in both
	// backends, where JavaScript would grow the array or leave a hole.
	SetIndex struct {
		Array, Index, Value Expression
		Element             Type
		// Site is which write of the program this is, for the cycle finder (lowering keeps the type of
		// what it writes into), or 0 when nothing recorded one.
		Site int
	}

	// SetProperty is object.name = value: the field takes the value, and lets go of what it held.
	SetProperty struct {
		Object Expression
		Name   string
		Value  Expression
		// Class is as Property's.
		Class int
		// Site is which write of the program this is, for the cycle finder (lowering keeps the type of
		// what it writes into), or 0 when nothing recorded one.
		Site   int
		Define bool // class field initialization defines an own data property
	}

	// Return leaves the function, with Value unless it returns void.
	Return struct{ Value Expression }

	// If runs Then when Condition holds, and Else (perhaps empty) when it doesn't.
	If struct {
		Condition  Expression
		Then, Else []Statement
	}

	// Loop is every loop: for, while and do...while. Body runs while Condition holds (checked before
	// the body, or after it when CheckAfter is set), and Update runs after each pass, continue
	// included.
	Loop struct {
		Condition  Expression
		Body       []Statement
		Update     []Statement
		CheckAfter bool

		// PerIteration are the for (let ...) variables, each its own copy in every iteration, as
		// JavaScript makes them: a closure made in one iteration keeps that iteration's value.
		PerIteration []int
	}

	// Block is a scope: locals declared in it are released when it ends.
	Block struct{ Body []Statement }

	// ForOf runs Body once per element of Iterable, with the element in Local. Over an array, its
	// length is read again before each pass, as JavaScript's array iterator does; over a string, the
	// elements are its code points, each a string.
	ForOf struct {
		Iterable Expression
		Element  Type
		Local    int
		Body     []Statement

		// Pattern, when set, destructures each element, a tuple, into locals: for (const [a, b] of
		// pairs). Local is unused then.
		Pattern []Binding

		// RegexIterator consumes lazy RegExp matchAll results.
		RegexIterator bool

		// Over a Map, MapPart is what each step gives: "entries" ([key, value], destructured by
		// Pattern), "keys" or "values" (in Local), with Key and Value the map's types.
		MapPart    string
		Key, Value Type
	}

	// Switch matches Value against each case's tests in order with ===, and runs the first match's
	// Body, or Default's when none matches. A break inside a case leaves the switch. Source
	// fallthrough is expressed by lowering as entry dispatch followed by sequential guarded bodies.
	Switch struct {
		Value   Expression
		Cases   []Case
		Default []Statement
	}

	// Depth counts the enclosing loops and switches skipped by a labeled break.
	// Zero is the innermost breakable, as for an unlabeled break.
	Break    struct{ Depth int }
	Continue struct{}

	// Throw throws Value, an Error: to the innermost Try around it, or out of the function, whose
	// caller passes it on the same way, or, out of every function, as a panic of String(Value).
	Throw struct{ Value Expression }

	// Try runs Body; if a throw leaves it, Catch runs with the error in CatchLocal (-1 when the catch
	// binds nothing). Finally runs after either, however they're left, and a throw neither caught nor
	// thrown by Finally goes on after it. HasCatch and HasFinally say which clauses there are.
	Try struct {
		Body, Catch, Finally []Statement
		CatchLocal           int
		HasCatch, HasFinally bool
	}
)

// Binding is one name in a destructuring pattern: the local it declares, and the field it reads.
type Binding struct {
	Local int
	Field string
}

// Case is one switch case: the constants it matches, and what it runs.
type Case struct {
	Tests []Expression
	Body  []Statement
}

func (WriteLine) statement()   {}
func (Declare) statement()     {}
func (Assign) statement()      {}
func (Evaluate) statement()    {}
func (Panic) statement()       {}
func (SetProperty) statement() {}
func (SetIndex) statement()    {}
func (Return) statement()      {}
func (If) statement()          {}
func (Loop) statement()        {}
func (Block) statement()       {}
func (ForOf) statement()       {}
func (Switch) statement()      {}
func (Break) statement()       {}
func (Continue) statement()    {}
func (Throw) statement()       {}
func (Try) statement()         {}

func (p *Program) HasInheritance() bool {
	for _, class := range p.Classes {
		if class.Base != 0 {
			return true
		}
	}
	return false
}
