package ir

// NodeHostCall is a selected host operation, synchronous unless its Returns is Promise, which it
// settles to. Arguments are evaluated in order, and nothing else of the program's runs: no operation
// calls back or writes a program value. A Promise form is only ever the direct operand of an Await
// (lowering's one road, library_workers.go), and while it is pending it holds its arguments until it
// settles. A synchronous form's results are owned and may share an argument's string storage through
// a counted share (node:path.basename uses adamic_string_share); it otherwise retains no argument.
// node:path and node:fs are both backends'. A workers: module (workers:KvNamespace.get) is
// JavaScript's alone, and native refuses it.
type NodeHostCall struct {
	Module, Member string
	Arguments      []Expression
	Returns        Type
	Throws         bool
}

func (n NodeHostCall) Type() Type { return n.Returns }
