package ir

// ArrayHoles makes new Array(length), keeping absence independently of payload.
// Length is a statically validated integer; array operations with dense-only
// semantics are refused until their sparse behavior is implemented.
type ArrayHoles struct {
	Length  Expression
	Element Type
}

func (ArrayHoles) Type() Type { return Array }
