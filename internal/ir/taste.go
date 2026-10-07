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

func (l Logical) Type() Type { return l.Of }
