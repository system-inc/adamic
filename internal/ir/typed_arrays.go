package ir

// TypedArrayNew admits zero-initialized numeric typed arrays. The native holder
// owns its numeric storage through ordinary counted object cleanup. Operations
// that expose buffers, coerce writes or change views remain explicitly refused.
type TypedArrayNew struct {
	Name   string
	Length Expression
}

func (TypedArrayNew) Type() Type { return Object }

// TypedArrayData is an internal borrowed storage view, only used by numeric
// indexed reads and length. It never escapes as a JavaScript Array.
type TypedArrayData struct{ Value Expression }

func (TypedArrayData) Type() Type { return Array }
