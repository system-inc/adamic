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
)

// IsReference reports whether a value of the type lives on the heap and is counted.
func (t Type) IsReference() bool {
	return t == String || t == Object || t == Array || t == Map
}

// Local is a variable: its name as written, for reading the output, and its type.
type Local struct {
	Name string
	Type Type

	// Global is a variable declared at the module's top level, which functions can read and write.
	Global bool
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

	// Property reads a field. Of is its type. Optional is ?., which is undefined when Object is.
	Property struct {
		Object   Expression
		Name     string
		Of       Type
		Optional bool
	}

	// ArrayLiteral makes an array.
	ArrayLiteral struct {
		Element  Type
		Elements []Expression
	}

	// Length is array.length.
	Length struct{ Array Expression }

	// MathCall is Math.<Function>(...), on numbers.
	MathCall struct {
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

	// Unwrap is a MaybeNumber the checker has proven present (narrowed), as a Number.
	Unwrap struct{ Value Expression }

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
	// padStart, padEnd, repeat, indexOf, includes, startsWith, endsWith. Arguments are as written;
	// lowering filled in any default.
	StringCall struct {
		Method    string
		Value     Expression
		Arguments []Expression
	}

	// CodePoints is [...string]: an array of its code points, each a string.
	CodePoints struct{ Value Expression }

	// MapNew is new Map(), or new Map([[key, value], ...]) with the pairs written out.
	MapNew struct {
		Key, Value Type
		Entries    [][2]Expression
	}

	// MapGet is map.get(Key): the value, or undefined (a null reference, or a MaybeNumber).
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
func (p Property) Type() Type      { return p.Of }
func (ArrayLiteral) Type() Type    { return Array }
func (Length) Type() Type          { return Number }
func (MathCall) Type() Type        { return Number }
func (ToFixed) Type() Type         { return String }
func (Undefined) Type() Type       { return Object }
func (IsUndefined) Type() Type     { return Boolean }
func (ArrayPush) Type() Type       { return Number }
func (ArrayJoin) Type() Type       { return String }
func (Unwrap) Type() Type          { return Number }
func (c Coalesce) Type() Type      { return c.Of }
func (StringLength) Type() Type    { return Number }
func (CharCodeAt) Type() Type      { return Number }
func (Trim) Type() Type            { return String }
func (CodePoints) Type() Type      { return Array }

func (c StringCall) Type() Type {
	switch c.Method {
	case "codePointAt":
		return MaybeNumber
	case "indexOf":
		return Number
	case "includes", "startsWith", "endsWith":
		return Boolean
	}
	return String
}
func (MapNew) Type() Type    { return Map }
func (MapSet) Type() Type    { return Map }
func (MapHas) Type() Type    { return Boolean }
func (MapDelete) Type() Type { return Boolean }
func (MapSize) Type() Type   { return Number }

func (g MapGet) Type() Type {
	if g.ValueType == Number {
		return MaybeNumber
	}
	return g.ValueType
}

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

// Case is one switch case: the constants it matches, and what it runs.
type Case struct {
	Tests []Expression
	Body  []Statement
}

func (WriteLine) statement() {}
func (Declare) statement()   {}
func (Assign) statement()    {}
func (Evaluate) statement()  {}
func (Panic) statement()     {}
func (Return) statement()    {}
func (If) statement()        {}
func (Loop) statement()      {}
func (Block) statement()     {}
func (ForOf) statement()     {}
func (Switch) statement()    {}
func (Break) statement()     {}
func (Continue) statement()  {}
