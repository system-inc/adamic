package ir

// HasProperty observes presence, including inherited names, without reading a value.
type HasProperty struct {
	Object Expression
	Name   string
}

func (HasProperty) Type() Type { return Boolean }

// DynamicProperty reads a runtime-tagged value. Its type remains unknown until narrowed.
type DynamicProperty struct {
	Optional bool
	Object   Expression
	Name     string
}

func (DynamicProperty) Type() Type { return Union }
