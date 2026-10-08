package ir

// BuiltinError preserves constructor identity independently of the writable name.
// Kind is Error, TypeError, SyntaxError, RangeError, ReferenceError, EvalError,
// or URIError, in that order.
type BuiltinError struct {
	Message Expression
	Kind    int
}

func (BuiltinError) Type() Type { return Object }

// ErrorIs checks constructor identity, or builtin Error ancestry.
type ErrorIs struct {
	Value Expression
	Kind  int
	Exact bool
}

func (ErrorIs) Type() Type { return Boolean }
