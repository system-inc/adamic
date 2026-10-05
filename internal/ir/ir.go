// Package ir is Adamic's intermediate representation: what the backends compile, with every type
// already proven by the checker.
//
// It holds exactly what the programs stage 0 can compile need, and grows one program at a time
// (docs/0.1.md). Nothing in it is speculative: an instruction exists because a fixture lowers to it
// and the oracle has checked what it prints.
package ir

// Program is one compiled Adamic program.
type Program struct {
	// Source is the entry file's base name, as written, for the header of what the backends emit.
	Source string

	// Strings are the program's string constants, as UTF-8, in first-use order.
	Strings []string

	// Locals are every variable the program declares, each its own, so shadowing is already resolved.
	Locals []Local

	// Functions are the module's function declarations, callable from anywhere in it.
	Functions []Function

	// Main is what the program does, in order.
	Main []Statement
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

	// Map is a Map; its key and value types travel on the expressions that read it.
	Map

	// MaybeNumber is number | undefined: a number that may be missing, as Map.get gives it.
	MaybeNumber

	// Closure is a function value: code, and the variables it captured.
	Closure

	// MaybeBoolean is boolean | undefined: a boolean that may be missing, as an array of booleans'
	// element is.
	MaybeBoolean
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
	return t == String || t == Object || t == Array || t == Map || t == Closure
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
	// Checked is a read of a global from inside a function, which may run before the global's
	// declaration has: JavaScript throws there (the temporal dead zone), and so does Adamic, out loud.
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
	}

	// Unary is -, + and ! on its operand.
	Unary struct {
		Operator Operator
		Operand  Expression
	}

	// Binary is an operator whose operands are already of the types it takes (the checker and
	// lowering saw to that): arithmetic on numbers, comparison of numbers, equality of like types,
	// and && and || on booleans, which short-circuit.
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
	}

	// ObjectLiteral makes an object. With Spread, it's { ...Spread, fields }: a copy of Spread's
	// object, whatever its shape, with Fields replaced (each one a field Spread's type has).
	ObjectLiteral struct {
		Spread Expression
		Fields []Field
	}

	// Property reads a field. Of is its type. Optional is ?., which is undefined when Object is: a
	// number field read that way is number | undefined.
	Property struct {
		Object   Expression
		Name     string
		Of       Type
		Optional bool
	}

	// ArrayLiteral makes an array. Where Spread is set, the element at that position is an array of the
	// same elements, spread into this one at that point in the evaluation, as JavaScript does.
	ArrayLiteral struct {
		Element  Type
		Elements []Expression
		Spread   []bool
	}

	// Length is array.length.
	Length struct{ Array Expression }

	// MathCall is Math.<Function>(...), on numbers.
	MathCall struct {
		Function  string
		Arguments []Expression
	}

	// NumberCall is Number.<Function>(...): parseInt (text, and a radix perhaps left out), parseFloat
	// (text), and isNaN, isFinite, isInteger and isSafeInteger (a number).
	NumberCall struct {
		Function  string
		Arguments []Expression
	}

	// ToFixed is Value.toFixed(Digits).
	ToFixed struct{ Value, Digits Expression }

	// Undefined is undefined where a reference goes: an object, an array or a string that may be
	// missing (Tree | undefined), held as a null pointer.
	Undefined struct{}

	// IsUndefined is Value === undefined, for a reference that may be missing.
	IsUndefined struct{ Value Expression }

	// ArrayPush is Array.push(Value): it appends and is the new length.
	ArrayPush struct {
		Array   Expression
		Value   Expression
		Element Type
	}

	// Unwrap is a Maybe pair the checker has proven present (narrowed), as what it holds.
	Unwrap struct{ Value Expression }

	// MaybeOf is a number or a boolean where Of, its Maybe pair, goes: Value, present, or undefined
	// when Value is nil.
	MaybeOf struct {
		Value Expression
		Of    Type
	}

	// MaybeToString is String(Value) for a Maybe pair: what it holds, written as String() writes it,
	// or "undefined".
	MaybeToString struct{ Value Expression }

	// Coalesce is Value ?? Fallback: Value when it's present, and otherwise Fallback, evaluated only
	// then. With Panic set instead of Fallback, it's Value ?? panic(Panic).
	Coalesce struct {
		Value    Expression
		Fallback Expression
		Panic    Expression
		Of       Type
	}

	// StringLength is string.length, in UTF-16 code units.
	StringLength struct{ Value Expression }

	// CharCodeAt is string.charCodeAt(Index): a UTF-16 code unit, or NaN.
	CharCodeAt struct{ Value, Index Expression }

	// Trim is string.trim().
	Trim struct{ Value Expression }

	// StringCall is one of a string's methods with JavaScript's meaning: slice, codePointAt,
	// padStart, padEnd, repeat, indexOf, lastIndexOf, includes, startsWith, endsWith, split,
	// trimStart, trimEnd, at (a string, or undefined: a null reference), replace and replaceAll
	// with a string pattern, and toUpperCase and toLowerCase. Arguments are as written; lowering filled in any default.
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
		Element      Type
		Includes     bool
	}

	// ArraySplice is array.splice(Start, Count, ...Items): Count perhaps left out (everything after
	// Start), and what's removed, a new array.
	ArraySplice struct {
		Array, Start, Count Expression
		Items               []Expression
		Element             Type
	}

	// ArrayFill is array.fill(Value, Start, End), Start and End perhaps nil (left out); in place, and
	// the array. With Array nil, it's new Array(Length).fill(Value): a new array, every element Value.
	ArrayFill struct {
		Array, Length, Value, Start, End Expression
		Element                          Type
	}

	// ArrayFrom is Array.from({ length: Length }, Callback): a new array of Length elements (ToLength,
	// and more than 2^32 - 1 panics as JavaScript throws), each the callback's result called with
	// undefined and its index, in order.
	ArrayFrom struct {
		Length, Callback Expression
		Element          Type
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

	// MapSize is map.size.
	MapSize struct{ Map Expression }

	// ArrayJoin is Array.join(Separator), writing each element as String() would.
	ArrayJoin struct {
		Array     Expression
		Separator Expression
		Element   Type
	}
)

// Field is one field of an object literal.
type Field struct {
	Name  string
	Value Expression
}

func (NumberConstant) Type() Type  { return Number }
func (BooleanConstant) Type() Type { return Boolean }
func (StringConstant) Type() Type  { return String }
func (NumberToString) Type() Type  { return String }
func (BooleanToString) Type() Type { return String }
func (Concat) Type() Type          { return String }
func (c Conditional) Type() Type   { return c.WhenTrue.Type() }
func (ObjectLiteral) Type() Type   { return Object }
func (p Property) Type() Type {
	if p.Optional {
		return Maybe(p.Of)
	}
	return p.Of
}
func (ArrayLiteral) Type() Type { return Array }
func (Length) Type() Type       { return Number }
func (MathCall) Type() Type     { return Number }

func (c NumberCall) Type() Type {
	if c.Function == "parseInt" || c.Function == "parseFloat" {
		return Number
	}
	return Boolean
}
func (ToFixed) Type() Type       { return String }
func (Undefined) Type() Type     { return Object }
func (IsUndefined) Type() Type   { return Boolean }
func (ArrayPush) Type() Type     { return Number }
func (ArrayJoin) Type() Type     { return String }
func (u Unwrap) Type() Type      { return u.Value.Type().Present() }
func (m MaybeOf) Type() Type     { return m.Of }
func (MaybeToString) Type() Type { return String }
func (c Coalesce) Type() Type    { return c.Of }
func (StringLength) Type() Type  { return Number }
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
	case "findIndex":
		return Number
	case "find":
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
func (MapNew) Type() Type    { return Map }
func (MapSet) Type() Type    { return Map }
func (MapHas) Type() Type    { return Boolean }
func (MapDelete) Type() Type { return Boolean }
func (MapSize) Type() Type   { return Number }

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
	switch b.Operator {
	case Add, Subtract, Multiply, Divide, Remainder, Power:
		return Number
	}
	return Boolean
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
	}

	// SetProperty is object.name = value: the field takes the value, and lets go of what it held.
	SetProperty struct {
		Object Expression
		Name   string
		Value  Expression
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

		// Over a Map, MapPart is what each step gives: "entries" ([key, value], destructured by
		// Pattern), "keys" or "values" (in Local), with Key and Value the map's types.
		MapPart    string
		Key, Value Type
	}

	// Switch matches Value against each case's tests in order with ===, and runs the first match's
	// Body, or Default's when none matches. A break inside a case leaves the switch. (0.1 has no
	// fallthrough: the checker refuses it.)
	Switch struct {
		Value   Expression
		Cases   []Case
		Default []Statement
	}

	Break    struct{}
	Continue struct{}
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
