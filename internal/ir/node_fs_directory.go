package ir

// NodeHostCall is a selected synchronous host operation. Arguments are evaluated in order.
type NodeHostCall struct {
	Module, Member string
	Arguments      []Expression
	Returns        Type
	Throws         bool
}

func (n NodeHostCall) Type() Type { return n.Returns }
