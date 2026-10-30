package ir

// NodeHostCall is a selected synchronous host operation. Arguments are evaluated in order.
// Results are owned and may share an argument's string storage through a counted share
// (node:path.basename uses adamic_string_share). No operation otherwise retains an
// argument, calls back, or writes a program value.
type NodeHostCall struct {
	Module, Member string
	Arguments      []Expression
	Returns        Type
	Throws         bool
}

func (n NodeHostCall) Type() Type { return n.Returns }
