package ir

// Truthy applies JavaScript's ToBoolean to one evaluated value.
type Truthy struct{ Value Expression }

func (Truthy) Type() Type { return Boolean }

// Logical returns an operand, evaluating Right only when Left does not decide the result.
type Logical struct {
	Left, Right Expression
	Of          Type
	KeepTruthy  bool
}

func (logical Logical) Type() Type { return logical.Of }

// LogicalAssignment holds a reference to its target across the conditional write.
// Write remains a statement so write and ownership analyses can see the assignment.
type LogicalAssignment struct {
	Read     Expression
	Write    Statement
	Key      Expression
	KeyNames []string
	Of       Type
	Operator string
}

func (assignment LogicalAssignment) Type() Type { return assignment.Of }

// Comma evaluates Left, discards its value, then evaluates and returns Right.
type Comma struct{ Left, Right Expression }

func (comma Comma) Type() Type { return comma.Right.Type() }

// Effects permits a statement's effects inside a value expression.
type Effects struct {
	Body   []Statement
	Result Expression
}

func (effects Effects) Type() Type { return effects.Result.Type() }

// Labeled is a statement with a named break destination.
type Labeled struct {
	Name string
	Body []Statement
}

func (Labeled) statement() {}

// Void evaluates Value and yields undefined, fitted to its destination when necessary.
type Void struct {
	Value Expression
	Of    Type
}

func (void Void) Type() Type { return (Undefined{Of: void.Of}).Type() }
