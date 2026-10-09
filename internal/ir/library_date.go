package ir

// DateCall preserves the receiver before its arguments, including a setter argument that mutates it.
type DateCall struct {
	Receiver  Expression
	Method    string
	Arguments []Expression
	Returns   Type
}

func (d DateCall) Type() Type { return d.Returns }
