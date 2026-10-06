package ir

// ObjectKeys reflects the actual object's shape, rather than its possibly narrower static type.
// Lowering admits only fixed plain objects whose fields are all own enumerable string keys.
type ObjectKeys struct{ Object Expression }

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
