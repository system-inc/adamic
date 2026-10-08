package ir

// ObjectKeys reflects the actual object's shape, rather than its possibly narrower static type.
// Enumeration also handles runtime arrays, synthetic absence and inherited constructor fields.
type ObjectKeys struct {
	Object Expression
	// Enumeration snapshots for-in keys and uses the runtime value, including hidden arrays.
	Enumeration bool
}

func (ObjectKeys) Type() Type { return Array }

// ClosureSelf is a named function expression's immutable inner name, bound to the current
// closure without capturing a cell that would make a reference cycle.
type ClosureSelf struct{}

func (ClosureSelf) Type() Type { return Closure }

// LibraryGlobal is a built-in identity. Lowering permits equality and typeof, and refuses other
// observations until that built-in's value interface has been implemented.
type LibraryGlobal struct{ Name string }

func (g LibraryGlobal) Type() Type {
	if g.Name == "JSON" {
		return Object
	}
	return Closure
}

// ForInOwn checks that a snapshotted key still exists before visiting it.
type ForInOwn struct{ Object, Key Expression }

func (ForInOwn) Type() Type { return Boolean }
