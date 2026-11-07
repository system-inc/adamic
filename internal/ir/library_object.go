package ir

import "strings"

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
}

func (call ObjectCall) Type() Type { return call.Returns }

// FrozenWriteMayThrow says a node can throw Node's TypeError because it writes an object that may be
// frozen (runtime/library_object.c's adamic_object_check_write sets it on the pending word): in a
// program that calls Object.freeze (ObjectsMayFreeze), an object write under a name that isn't
// private, and Object.assign. Which objects a freeze reaches isn't known across aliases and calls, so
// in such a program every write can; in one without, none does. Lowering counts the nodes that can,
// the flow graph gives them an exception edge, and the emitter tests the word after each check.
func (program *Program) FrozenWriteMayThrow(node any) bool {
	if !program.ObjectsMayFreeze {
		return false
	}
	switch node := node.(type) {
	case SetProperty:
		// A private field is an internal slot, which Object.freeze leaves writable.
		return !strings.HasPrefix(node.Name, "#")
	case ObjectCall:
		return node.Method == "assign"
	}
	return false
}
