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

	// Main is what the program does, in order.
	Main []Statement
}

// Type is a value's representation. The checker proved the TypeScript type; this is what's left of
// it at runtime.
type Type int

const (
	Number Type = iota + 1
	Boolean
	String
)

// Local is a variable: its name as written, for reading the output, and its type.
type Local struct {
	Name string
	Type Type
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
	Read struct {
		Local int
		Of    Type
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
)

func (NumberConstant) Type() Type  { return Number }
func (BooleanConstant) Type() Type { return Boolean }
func (StringConstant) Type() Type  { return String }
func (NumberToString) Type() Type  { return String }
func (BooleanToString) Type() Type { return String }
func (Concat) Type() Type          { return String }
func (c Conditional) Type() Type   { return c.WhenTrue.Type() }

func (r Read) Type() Type { return r.Of }

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

	// Assign gives a local a new value, releasing the old one if it's a string.
	Assign struct {
		Local int
		Value Expression
	}

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

	Break    struct{}
	Continue struct{}
)

func (WriteLine) statement() {}
func (Declare) statement()   {}
func (Assign) statement()    {}
func (If) statement()        {}
func (Loop) statement()      {}
func (Block) statement()     {}
func (Break) statement()     {}
func (Continue) statement()  {}
