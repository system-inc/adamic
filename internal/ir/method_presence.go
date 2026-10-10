package ir

// MethodPresence observes a method without making a detached function value.
// Own closure fields shadow prototype methods, including when the field is undefined.
type MethodPresence struct {
	Object   Expression
	Name     string
	Optional bool
}

func (MethodPresence) Type() Type { return Boolean }
