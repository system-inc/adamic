package high_level_intermediate_representation

// Pattern is a destructuring target.
//
// Kept structured rather than expanded into loads and stores at lowering time. The reason is that
// the shape is observable and several consumers want it: `const {a, b} = props` is the canonical
// dependency shape React's memoization rules reason about, and recovering "these two bindings came
// from one destructure of one object" from a flattened sequence of property loads means
// re-inferring what the syntax already said.
//
// The flattening is available to any pass that wants it - walk the pattern and emit loads - and is
// not available in reverse. So the structured form is what lowering produces.
type Pattern interface {
	pattern()
}

// PlacePattern binds one value: the `a` in `const [a] = xs`.
type PlacePattern struct{ Place Place }

// ObjectPattern destructures an object.
type ObjectPattern struct {
	Properties []ObjectPatternProperty
	// Rest is `...rest`, nil when absent.
	Rest *Place
}

// ObjectPatternProperty is one binding in an object pattern.
type ObjectPatternProperty struct {
	// Key is the property read. Empty when ComputedKey is set.
	Key string
	// ComputedKey is the key expression for `{[k]: v}`.
	ComputedKey *Place
	// Value is the nested pattern this property binds to.
	Value Pattern
	// Default is the value used when the property is undefined, nil when absent.
	//
	// The branch that selects it is control flow and is a terminal; this names the value the
	// branch produces, so a pass reading the pattern alone knows a default exists.
	Default *Place
}

// ArrayPattern destructures an array.
type ArrayPattern struct {
	// Elements are the positional bindings. A nil element is a hole from `[a, , b]`.
	Elements []ArrayPatternElement
	// Rest is `...rest`, nil when absent.
	Rest *Place
}

// ArrayPatternElement is one position in an array pattern.
type ArrayPatternElement struct {
	// Value is the pattern bound at this position, nil for a hole.
	Value Pattern
	// Default is the value used when the element is undefined, nil when absent.
	Default *Place
}

func (*PlacePattern) pattern()  {}
func (*ObjectPattern) pattern() {}
func (*ArrayPattern) pattern()  {}
