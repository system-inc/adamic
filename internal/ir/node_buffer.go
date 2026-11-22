package ir

// NodeBufferCall is the census's Buffer and SHA-256 host surface. Arguments are
// evaluated once in source order. Encoding is a proven literal, not runtime dispatch.
// A Buffer is held as a Uint8Array, one byte an element, as on Node; Hash uses a
// counted runtime object.
type NodeBufferCall struct {
	Function  string
	Arguments []Expression
	Encoding  int
	Returns   Type
}

func (c NodeBufferCall) Type() Type { return c.Returns }

// MayThrow reports whether the call can throw a Node error a catch takes: Buffer.allocUnsafe, whose
// size can be out of range or one that can't be allocated (runtime node_buffer.c sets it on the pending
// word). Lowering counts it, the flow graph gives it an exception edge, and the emitter tests the word
// after it. Allocation can fail for any size, so a constant one doesn't spare the test.
func (c NodeBufferCall) MayThrow() bool {
	return c.Function == "buffer_alloc_unsafe"
}
