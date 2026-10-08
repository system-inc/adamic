package ir

// PhantomMember reads a checker-only member from a primitive. Present primitives produce
// undefined. Missing ones throw Failure unless Optional is set, exactly as an ordinary JS read.
// Undefined uses Object's null pointer representation; no brand object is allocated.
type PhantomMember struct {
	Value    Expression
	Name     string
	Optional bool
	Failure  MakeError
}

func (PhantomMember) Type() Type { return Object }
