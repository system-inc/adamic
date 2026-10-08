package ir

// ClosureLength reads a function value's source arity, holding its receiver once.
// A nonoptional receiver is proven present or carries lowering's Defined check.
type ClosureLength struct {
	Value    Expression
	Optional bool
}

func (v ClosureLength) Type() Type {
	if v.Optional {
		return MaybeNumber
	}
	return Number
}
