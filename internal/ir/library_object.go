package ir

// ObjectCall is a static Object method whose arguments have proven representations.
// Element is the homogeneous value type for values and entries; Returns is tsc's result.
type ObjectCall struct {
	// IntegrityShape: 0 plain data slots, 1 no own properties, 2 RegExp lastIndex.
	IntegrityShape int
	Method         string
	Arguments      []Expression
	Element        Type
	Returns        Type
}

func (c ObjectCall) Type() Type { return c.Returns }
