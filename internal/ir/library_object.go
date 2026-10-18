package ir

// ObjectCall is a static Object method whose arguments have proven representations.
// Element is the homogeneous value type for values and entries; Returns is tsc's result.
type ObjectCall struct {
	Method    string
	Arguments []Expression
	// PrimitiveOperands is checker proof for Object.is, before union boxing.
	PrimitiveOperands [2]bool
	Element           Type
	Returns           Type
	Checked           bool
	Allowed           []Expression
	ElementName       string
}

func (c ObjectCall) Type() Type { return c.Returns }
