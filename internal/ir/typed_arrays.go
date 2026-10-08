package ir

// IsTypedArray reports whether the counted value holds an unboxed numeric buffer.
func (t Type) IsTypedArray() bool {
	return t == Uint8Array || t == Uint16Array || t == Int32Array || t == Float64Array
}

// TypedArrayNew converts Source, a length or number[], to an owned typed array.
type TypedArrayNew struct {
	Of        Type
	Source    Expression
	FromArray bool
}

// TypedArrayFill mutates the receiver and returns it. Arguments preserve omissions.
type TypedArrayFill struct {
	Array     Expression
	Arguments []Expression
}

// TypedArraySet copies a same-kind source into the receiver, safely across overlap.
type TypedArraySet struct {
	Array     Expression
	Arguments []Expression
}

// TypedArraySubarray makes an owned view sharing the receiver's buffer.
type TypedArraySubarray struct {
	Array     Expression
	Arguments []Expression
}

func (v TypedArrayNew) Type() Type      { return v.Of }
func (v TypedArrayFill) Type() Type     { return v.Array.Type() }
func (TypedArraySet) Type() Type        { return 0 }
func (v TypedArraySubarray) Type() Type { return v.Array.Type() }
