package ir

// RegExpProgram contains constants compiled and checked before native emission.
type RegExpProgram struct{ Pattern, Flags, Declarations string }
type RegExpNew struct {
	Index, Source, Flags int
	Arguments            []Expression
}

func (RegExpNew) Type() Type { return Object }

// RegExpCall covers both RegExp and String methods. Value is evaluated first.
type RegExpCall struct {
	Value     Expression
	Arguments []Expression
	Method    string
	Returns   Type
}

func (r RegExpCall) Type() Type { return r.Returns }

// RegExpProperty reads result-array metadata, preserving optional access.
type RegExpProperty struct {
	Array    Expression
	Name     string
	Of       Type
	Optional bool
}

func (r RegExpProperty) Type() Type {
	if r.Optional {
		return Maybe(r.Of)
	}
	return r.Of
}

// Null is a reference kind's null value: a sentinel for migrated kinds, NULL for legacy ones.
type Null struct{ Of Type }

func (n Null) Type() Type {
	if n.Of != 0 {
		return n.Of
	}
	return Object
}

type IsNull struct {
	Value       Expression
	AlwaysFalse bool
	// IncludeUndefined is == null, a nullish test, rather than === null.
	IncludeUndefined bool
}

func (IsNull) Type() Type { return Boolean }

// RegExpGroup reads a declared named-group dictionary; absent keys are undefined.
type RegExpGroup struct {
	Object   Expression
	Name     string
	Of       Type
	Optional bool
}

func (r RegExpGroup) Type() Type { return r.Of }
