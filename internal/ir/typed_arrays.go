package ir

// IsTypedArray reports whether the counted value holds an unboxed numeric buffer.
func (t Type) IsTypedArray() bool {
	return t == Uint8Array || t == Int32Array || t == Float64Array
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

// NumericTypedArrayNew admits zero-initialized numeric typed arrays. The native holder
// owns its numeric storage through ordinary counted object cleanup. Operations
// that expose buffers, coerce writes or change views remain explicitly refused.
type NumericTypedArrayNew struct {
	Name   string
	Length Expression
}

func (NumericTypedArrayNew) Type() Type { return Object }

// TypedArrayData is an internal borrowed storage view, only used by numeric
// indexed reads and length. It never escapes as a JavaScript Array.
type TypedArrayData struct{ Value Expression }

func (TypedArrayData) Type() Type { return Array }
