package ir

// NodeBufferCall is the census's Buffer and SHA-256 host surface. Arguments are
// evaluated once in source order. Encoding is a proven literal, not runtime dispatch.
// Buffer bytes use numeric array slots; Hash uses a counted runtime object.
type NodeBufferCall struct {
	Function  string
	Arguments []Expression
	Encoding  int
	Returns   Type
}

func (c NodeBufferCall) Type() Type { return c.Returns }
