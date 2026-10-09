package ir

// ObjectCall is a static Object method whose arguments have proven representations.
// Element is the homogeneous value type for values and entries; Returns is tsc's result.
type ObjectCall struct {
	Method      string
	Arguments   []Expression
	Element     Type
	Returns     Type
	Checked     bool
	Allowed     []Expression
	ElementName string
	Readiness   string // runtime optional-contract diagnostic site
}

func (c ObjectCall) Type() Type { return c.Returns }
