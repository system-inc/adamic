package ir

// IteratorMethod caches a protocol function and its receiver together. Optional
// is used only for IteratorClose: a missing return method needs no call.
type IteratorMethod struct {
	Object   Expression
	Name     string
	Optional bool
}

func (IteratorMethod) Type() Type { return Closure }

// IteratorField reads one result property, including accessors. Missing done
// means false; a missing yielded value is allowed only by its proven type.
type IteratorField struct {
	Object Expression
	Name   string
	Of     Type
	Absent bool
}

func (f IteratorField) Type() Type { return f.Of }
